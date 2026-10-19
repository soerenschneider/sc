package cmd

import (
	"context"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/spf13/cobra"
)

var vaultIdentityListEntitiesCmd = &cobra.Command{
	Use: "list-entities",
	Aliases: []string{
		"list",
	},
	Short: "List identity entities in Vault",
	Long: `Lists the names of all entities managed by the Vault identity system.

Note: Appropriate Vault permissions are required to access identity entity data.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		keys, err := client.IdentityListEntities(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not list identity entities")
		}

		sort.Strings(keys)
		tui.WriteListOutput(keys)
	},
}

func init() {
	vaultIdentityCmd.AddCommand(vaultIdentityListEntitiesCmd)
}
