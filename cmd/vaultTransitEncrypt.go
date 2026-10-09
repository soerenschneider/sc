package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"charm.land/huh/v2/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const (
	vaultTransitKey     = "key"
	vaultTransitContext = "context"
	vaultTransitInput   = "input"
	vaultTransitOutput  = "output"
)

// transitMountLister is the subset of the vault client needed to discover transit mounts.
type transitMountLister interface {
	TransitListMounts(ctx context.Context) ([]string, error)
}

// transitKeyLister is the subset of the vault client needed to discover transit keys.
type transitKeyLister interface {
	TransitListKeys(ctx context.Context, mount string) ([]string, error)
}

var vaultTransitEncryptCmd = &cobra.Command{
	Use:     "encrypt",
	Aliases: []string{"enc"},
	Short:   "Encrypt data using a transit key",
	Long: `Encrypts data using a key of the Vault transit secret engine and prints the resulting ciphertext.

The plaintext is read from the file given by --input ('-' for stdin). If no input file is
given and data is piped to stdin, stdin is read. Otherwise, the plaintext is prompted for
interactively.

Mount and key may be passed as flags. Anything that is not passed is discovered and
offered for selection, falling back to free-form input if discovery is not possible.

Examples:
  sc vault transit encrypt
  sc vault transit encrypt -m transit -k my-key -i secret.txt
  echo -n "hunter2" | sc vault transit encrypt -m transit -k my-key`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		mount := pkg.GetString(cmd, vaultMountPath)
		keyName := pkg.GetString(cmd, vaultTransitKey)
		keyContext := pkg.GetString(cmd, vaultTransitContext)
		inputFile := pkg.GetString(cmd, vaultTransitInput)
		outputFile := pkg.GetString(cmd, vaultTransitOutput)

		plaintext, err := readTransitInput(inputFile, func() string {
			return tui.ReadSensitiveInput("Enter plaintext")
		})
		if err != nil {
			log.Fatal().Err(err).Msg("could not read plaintext")
		}

		if mount == "" {
			mount = promptTransitMount(cmd.Context(), client)
		}

		if keyName == "" {
			keyName = promptTransitKey(cmd.Context(), client, mount)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), vaultDefaultTimeout)
		defer cancel()

		var ciphertext string
		title := fmt.Sprintf("Encrypting data using key %q on mount %q...", keyName, mount)
		if err := runWithSpinner(ctx, title, func(ctx context.Context) error {
			var err error
			ciphertext, err = client.TransitEncrypt(ctx, mount, keyName, plaintext, []byte(keyContext))
			return err
		}); err != nil {
			log.Fatal().Err(err).Msg("could not encrypt data")
		}

		if outputFile != "" {
			if err := os.WriteFile(pkg.GetExpandedFile(outputFile), []byte(ciphertext+"\n"), 0600); err != nil {
				log.Fatal().Err(err).Msg("could not write ciphertext")
			}
			log.Info().Msgf("Ciphertext written to %q", outputFile)
			return
		}

		fmt.Println(ciphertext)
	},
}

// readTransitInput reads the input from the given file, from stdin if the file is '-' or
// data is piped to stdin, and prompts for it interactively otherwise.
func readTransitInput(inputFile string, prompt func() string) ([]byte, error) {
	switch {
	case inputFile == "-":
		return io.ReadAll(os.Stdin)
	case inputFile != "":
		return os.ReadFile(pkg.GetExpandedFile(inputFile))
	case !isTerminal(os.Stdin):
		return io.ReadAll(os.Stdin)
	default:
		return []byte(prompt()), nil
	}
}

// promptTransitMount tries to offer a selection of the mounted transit secret engines and
// falls back to free-form input if the mounts can not be listed.
func promptTransitMount(ctx context.Context, client transitMountLister) string {
	var availableMounts []string

	ctx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
	defer cancel()

	if err := runWithSpinner(ctx, "Discovering transit mounts...", func(ctx context.Context) error {
		mounts, err := client.TransitListMounts(ctx)
		if err != nil {
			return err
		}
		availableMounts = mounts
		return nil
	}); err != nil {
		log.Warn().Err(err).Msg("could not discover transit mounts, falling back to manual input")
	}

	switch len(availableMounts) {
	case 0:
		return tui.ReadInput("Enter transit mount", []string{"transit"})
	case 1:
		log.Info().Msgf("Using only available transit mount %q", availableMounts[0])
		return availableMounts[0]
	default:
		return tui.SelectInput("Select transit mount", availableMounts)
	}
}

// promptTransitKey tries to offer a selection of the keys of the given transit mount and
// falls back to free-form input if the keys can not be listed.
func promptTransitKey(ctx context.Context, client transitKeyLister, mount string) string {
	var availableKeys []string

	ctx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
	defer cancel()

	title := fmt.Sprintf("Loading keys of transit mount %q...", mount)
	if err := runWithSpinner(ctx, title, func(ctx context.Context) error {
		keys, err := client.TransitListKeys(ctx, mount)
		if err != nil {
			return err
		}
		availableKeys = keys
		return nil
	}); err != nil {
		log.Warn().Err(err).Msg("could not list transit keys, falling back to manual input")
	}

	if len(availableKeys) > 0 {
		return tui.SelectInput("Select transit key", availableKeys)
	}

	return tui.ReadInput("Enter transit key", nil)
}

const (
	// spinnerDelay is how long an action may take before a spinner is displayed.
	spinnerDelay = 200 * time.Millisecond
	// spinnerMinVisible is how long a spinner is displayed at least, once shown.
	spinnerMinVisible = 300 * time.Millisecond
)

// runWithSpinner runs action while displaying a spinner. As the spinner requires a terminal,
// action is run without a spinner if stderr is not a terminal, e.g. when used in scripts.
//
// The spinner is only displayed if action does not finish within spinnerDelay and is then
// displayed for at least spinnerMinVisible. Besides avoiding flickering, this makes sure the
// spinner lives long enough to consume the terminal's replies to the queries bubbletea sends
// on startup, which would otherwise leak into the shell.
func runWithSpinner(ctx context.Context, title string, action func(ctx context.Context) error) error {
	if !isTerminal(os.Stderr) {
		return action(ctx)
	}

	done := make(chan error, 1)
	go func() {
		done <- action(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(spinnerDelay):
	}

	return spinner.New().
		ActionWithErr(func(_ context.Context) error {
			minVisible := time.After(spinnerMinVisible)
			err := <-done
			<-minVisible
			return err
		}).
		Title(title).
		Context(ctx).
		Type(spinner.Dots).
		WithOutput(os.Stderr).
		Run()
}

func init() {
	vaultTransitCmd.AddCommand(vaultTransitEncryptCmd)

	vaultTransitEncryptCmd.Flags().StringP(vaultTransitKey, "k", "", "Name of the transit key. If not specified, available keys are discovered.")
	vaultTransitEncryptCmd.Flags().StringP(vaultTransitContext, "c", "", "Context for key derivation, required if the key has derivation enabled")
	vaultTransitEncryptCmd.Flags().StringP(vaultTransitInput, "i", "", "File to read the plaintext from, '-' for stdin. If not specified, reads piped stdin or prompts.")
	vaultTransitEncryptCmd.Flags().StringP(vaultTransitOutput, "o", "", "File to write the ciphertext to. If not specified, prints to stdout.")
}
