package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/internal/vault/prompt"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultAwsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all existing roles in the AWS secrets engine.",
	Long: `List all configured roles in the Vault AWS secrets engine.

This command queries the AWS secrets engine mounted in Vault to retrieve a list of
all available role names. These roles define the IAM permissions that Vault can
generate for temporary AWS credentials.

The role names are sorted alphabetically and printed one per line. Use --output json to print
them as JSON array instead.

The command expects that the AWS secrets engine has been enabled and configured.`,
	Example: `  sc vault aws list
  sc vault aws list -m my-aws-mount -o json`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		output := mustGetNameListOutput(cmd)

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		mount := pkg.GetString(cmd, vaultMountPath)
		if mount == "" {
			mount = prompt.Mount(cmd.Context(), client, "aws")
		}

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		roles, err := client.AwsListRoles(ctx, mount)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to list AWS roles for mount %q", mount)
		}

		mustPrintNames(roles, output)
	},
}

func init() {
	vaultAwsCmd.AddCommand(vaultAwsListCmd)
	addNameListOutputFlag(vaultAwsListCmd)
}
