package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const (
	vaultUnwrapField = "field"
	vaultUnwrapJson  = "json"
)

var vaultUnwrapCmd = &cobra.Command{
	Use:   "unwrap [token]",
	Short: "Unwrap a response wrapping token",
	Long: `Unwrap a response wrapping token and print the data it wraps.

The wrapping token is used to authenticate the request, so no login is required. Wrapping tokens
can only be unwrapped once.

The token is read from the argument. If it is not given, it is read from stdin if data is piped
to it, and prompted for otherwise.

Use --field to print the value of a single field only, or --json to print the data as JSON.`,
	Example: `  sc vault unwrap
  sc vault unwrap s.Xyz --field secret_id
  sc vault approle rotate-secret-id -r my-role --wrap-ttl 5m -f --json | jq -r .wrapping_token | sc vault unwrap --json`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token, err := readWrappingToken(args)
		if err != nil {
			log.Fatal().Err(err).Msg("could not read wrapping token")
		}
		if token == "" {
			log.Fatal().Msg("no wrapping token given")
		}

		client := vault.MustBuildClient(cmd)

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		secret, err := client.Unwrap(ctx, token)
		if err != nil {
			log.Fatal().Err(err).Msg("could not unwrap token")
		}

		// unwrapping a token wrapping an auth response, e.g. a token, does not return data
		data := secret.Data
		if data == nil && secret.Auth != nil {
			data = map[string]any{
				"token":          secret.Auth.ClientToken,
				"token_accessor": secret.Auth.Accessor,
				"token_policies": secret.Auth.Policies,
				"lease_duration": secret.Auth.LeaseDuration,
				"renewable":      secret.Auth.Renewable,
			}
		}

		if field := pkg.GetString(cmd, vaultUnwrapField); field != "" {
			value, ok := data[field]
			if !ok {
				log.Fatal().Msgf("field %q not found in wrapped data", field)
			}
			fmt.Println(value)
			return
		}

		if pkg.MustGetBool(cmd, vaultUnwrapJson) {
			printJson(data)
			return
		}

		tui.PrintMapOutput(data)
	},
}

// readWrappingToken reads the wrapping token from the args, from stdin if data is piped to it, and prompts for it
// interactively otherwise.
func readWrappingToken(args []string) (string, error) {
	switch {
	case len(args) > 0:
		return strings.TrimSpace(args[0]), nil
	case !isTerminal(os.Stdin):
		data, err := io.ReadAll(os.Stdin)
		return strings.TrimSpace(string(data)), err
	default:
		return strings.TrimSpace(tui.ReadSensitiveInput("Wrapping token")), nil
	}
}

func init() {
	vaultCmd.AddCommand(vaultUnwrapCmd)

	vaultUnwrapCmd.Flags().String(vaultUnwrapField, "", "Print the value of this field only")
	vaultUnwrapCmd.Flags().Bool(vaultUnwrapJson, false, "Print the wrapped data as JSON")
	vaultUnwrapCmd.MarkFlagsMutuallyExclusive(vaultUnwrapField, vaultUnwrapJson)
}
