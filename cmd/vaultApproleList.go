package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/internal/vault/prompt"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var vaultApproleListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the names of all AppRoles",
	Long: `Lists the names of all AppRoles of the mount, sorted alphabetically and printed one per line.

Use --output json to print them as JSON array instead. To list the secret_ids of the AppRoles,
use 'list-accessors'.`,
	Example: `  sc vault approle list
  sc vault approle list -m my-approle-mount
  sc vault approle list -o json`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		output := mustGetNameListOutput(cmd)

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		roles, err := prompt.Load(cmd.Context(), "Loading AppRoles...", func(ctx context.Context) ([]string, error) {
			return client.ListAppRoles(ctx, mount)
		})
		if err != nil {
			log.Fatal().Err(err).Msg("could not list AppRoles")
		}

		mustPrintNames(roles, output)
	},
}

func init() {
	vaultApproleCmd.AddCommand(vaultApproleListCmd)
	addNameListOutputFlag(vaultApproleListCmd)
}
