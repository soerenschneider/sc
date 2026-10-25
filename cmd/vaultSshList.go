package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/internal/vault/prompt"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var sshListCmd = &cobra.Command{
	Use:     "list-roles",
	Aliases: []string{"list"},
	Short:   "Lists all roles for the SSH secrets engine",
	Long: `Lists the names of all roles of the SSH secrets engine, sorted alphabetically and printed one
per line. Use --output json to print them as JSON array instead.`,
	Example: `  sc vault ssh list-roles
  sc vault ssh list-roles -m my-ssh-mount -o json`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		output := mustGetNameListOutput(cmd)

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		mount := pkg.GetString(cmd, vaultMountPath)
		if mount == "" {
			mount = prompt.Mount(cmd.Context(), client, "ssh")
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), vaultDefaultTimeout)
		defer cancel()

		roles, err := client.SshListRoles(ctx, mount)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to list SSH roles for mount %q", mount)
		}

		mustPrintNames(roles, output)
	},
}

func init() {
	vaultSshCmd.AddCommand(sshListCmd)
	addNameListOutputFlag(sshListCmd)
}
