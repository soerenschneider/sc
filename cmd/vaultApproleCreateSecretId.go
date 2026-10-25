package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

type approleSecretIDCreation struct {
	RoleName         string   `json:"role_name"`
	SecretID         string   `json:"secret_id"`
	SecretIDAccessor string   `json:"secret_id_accessor"`
	Cidrs            []string `json:"cidrs,omitempty"`
}

var vaultApproleCreateSecretIdCmd = &cobra.Command{
	Use:     "create-secret-id",
	Aliases: []string{"create"},
	Short:   "Create a new secret_id for an AppRole",
	Long: `Create a new secret_id for an AppRole. Unlike 'rotate-secret-id', existing secret_ids are kept.

The role is given using --role or --role-file. If it is not given and the command is run in a
terminal, the AppRoles are listed for selection.

Using --cidr, the new secret_id and the tokens issued with it are bound to the given CIDR blocks.
Using --cidr-host, the CIDR blocks are looked up dynamically: the host is resolved using DNS and
each of its IP addresses is added as a single host CIDR (/32 for IPv4, /128 for IPv6). Both
flags can be combined.

Use --json to print the new secret_id as JSON, e.g. for further processing with jq.`,
	Example: `  sc vault approle create-secret-id
  sc vault approle create-secret-id -r my-role --metadata host=foo --ttl 720h
  sc vault approle create-secret-id -r my-role --cidr-host my-host.example.com --json | jq -r .secret_id`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		roleName := mustGetApproleRoleName(cmd)
		opts := mustParseApproleSecretIDOptions(cmd)

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		if roleName == "" {
			roleName = selectApprole(cmd.Context(), client, mount)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), vaultDefaultTimeout)
		defer cancel()

		creation := approleSecretIDCreation{RoleName: roleName, Cidrs: opts.Cidrs}
		err := tui.RunWithSpinner(ctx, "Creating secret_id...", func(ctx context.Context) error {
			var err error
			creation.SecretID, creation.SecretIDAccessor, err = client.CreateAppRoleSecretID(ctx, mount, roleName, opts.Metadata, opts.Cidrs, opts.Cidrs, opts.Ttl)
			return err
		})
		if err != nil {
			log.Fatal().Err(err).Msgf("could not create secret_id for role %q", roleName)
		}

		if pkg.MustGetBool(cmd, vaultApproleJson) {
			printJson(creation)
			return
		}

		tui.PrintMapOutput(map[string]any{
			"role_name":          creation.RoleName,
			"secret_id":          creation.SecretID,
			"secret_id_accessor": creation.SecretIDAccessor,
			"cidrs":              creation.Cidrs,
		})
	},
}

func init() {
	vaultApproleCmd.AddCommand(vaultApproleCreateSecretIdCmd)

	addApproleSecretIDFlags(vaultApproleCreateSecretIdCmd)
}
