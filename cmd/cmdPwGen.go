package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/pw"
	"github.com/soerenschneider/sc/pkg/clipboard"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const passwordClearTimeout = 30 * time.Second

var passwordOpts pw.PasswordOptions

var pwGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generates a random password",
	Run: func(cmd *cobra.Command, args []string) {
		generatedPw, err := pw.GenerateXkcd(passwordOpts)
		if err != nil {
			log.Fatal().Err(err).Msg("could not generate password")
		}

		printPassword, _ := cmd.Flags().GetBool("print")
		if printPassword {
			printPasswordTemporarily(cmd.Context(), generatedPw)
		} else {
			if err := clipboard.CopyClipboard(generatedPw); err != nil {
				log.Fatal().Err(err).Msg("could not copy password to clipboard")
			}
			log.Info().Msgf("Generated password copied to clipboard")

			if err := spinner.New().
				Context(cmd.Context()).
				Action(func() {
					clearPassword(cmd.Context(), generatedPw)
				}).
				Title(fmt.Sprintf("Waiting %0f to wipe password", passwordClearTimeout.Seconds())).
				Type(spinner.Dots).
				Run(); err != nil {
				clearPassword(cmd.Context(), generatedPw)
				log.Warn().Err(err).Msg("could not display spinner")
			}
		}
	},
}

func clearPassword(ctx context.Context, generatedPw string) {
	select {
	case <-ctx.Done():
		log.Info().Msg("Context canceled, clearing clipboard early")
	case <-time.After(passwordClearTimeout):
	}

	current, err := clipboard.PasteClipboard()
	if err == nil && current == generatedPw {
		_ = clipboard.CopyClipboard("")
	}
}

func printPasswordTemporarily(ctx context.Context, pw string) {
	fd := int(os.Stdin.Fd())

	// put terminal in raw mode (no Enter key, no echo)
	oldState, err := term.MakeRaw(fd)
	if err == nil {
		defer term.Restore(fd, oldState)
	}

	fmt.Print(pw)
	select {
	case <-ctx.Done():
	case <-time.After(passwordClearTimeout):
	}

	fmt.Print("\r\033[K")
}

func init() {
	pwCmd.AddCommand(pwGenCmd)
	pwGenCmd.Flags().BoolVarP(&passwordOpts.Lowercase, "lowercase", "", false, "Converts all words to lowercase")
	pwGenCmd.Flags().BoolVarP(&passwordOpts.Special, "special", "", false, "Include special character")
	pwGenCmd.Flags().BoolP("print", "p", false, "Print password to stdout, do not copy to clipboard")
	pwGenCmd.Flags().IntVarP(&passwordOpts.NumWords, "words", "w", 4, "Number of words in the password")
	pwGenCmd.Flags().VarP(&passwordOpts.SpecialPos, "special-pos", "", "Position of special characters")
	pwGenCmd.Flags().StringVarP(&passwordOpts.Language, "language", "l", "de", "Wordlist language, e.g. 'en', 'de'")
	pwGenCmd.Flags().StringVarP(&passwordOpts.Separator, "separator", "s", " ", "Separator between words")
}
