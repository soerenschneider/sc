package cmd

import (
	"context"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultIdentityListGroupsCmd = &cobra.Command{
	Use: "list-groups",
	Aliases: []string{
		"groups",
	},
	Short: "List identity groups in Vault",
	Long: `Lists the names of all groups managed by the Vault identity system.

Note: Appropriate Vault permissions are required to access identity group data.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		keys, err := client.IdentityListGroups(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not list identity groups")
		}

		sort.Strings(keys)
		tui.WriteListOutput(keys)
	},
}

func init() {
	vaultIdentityCmd.AddCommand(vaultIdentityListGroupsCmd)
}
