package cmd

import (
	"context"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultIdentityListGroupsCmd = &cobra.Command{
	Use: "list-groups",
	Aliases: []string{
		"groups",
	},
	Short: "Manage Vault totp",
	Long: `The 'token' command group contains subcommands for interacting with Vault tokens.

This command itself does not perform any actions. Instead, use one of its subcommands
to inspect or manage tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		keys, err := client.IdentityListGroups(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not list identity groups")
		}

		sort.Strings(keys)
		writeListOutput(keys)
	},
}

func init() {
	vaultIdentityCmd.AddCommand(vaultIdentityListGroupsCmd)
}
