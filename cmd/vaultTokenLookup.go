package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultTokenLookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup and display information about the current Vault token",
	Long: `Retrieve and display information about the currently active Vault token.

This command queries the Vault API to show metadata about the current token,
such as its creation time, expiration, policies, and identity.

It attempts to authenticate using the following sources (in order):
  1. The VAULT_TOKEN environment variable
  2. A token loaded from the local configuration file (e.g. ~/.config/mycli/token)
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustBuildClient(cmd)
		vault.MustAuthenticateClient(client)

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		secret, err := client.LookupToken(ctx)
		if err != nil {
			log.Fatal().Msgf("could not lookup: %v", err)
		}

		tui.PrintMapOutput(secret.Data)
	},
}

func init() {
	vaultTokenCmd.AddCommand(vaultTokenLookupCmd)
}
