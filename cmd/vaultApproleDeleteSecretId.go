package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const vaultApproleAccessor = "accessor"

var vaultApproleDeleteSecretIdCmd = &cobra.Command{
	Use:     "delete-secret-id [accessor]",
	Aliases: []string{"delete", "rm"},
	Short:   "Delete a secret_id of an AppRole",
	Long: `Deletes a secret_id of an AppRole, identified by its accessor.

The accessor may be passed as argument or via --accessor, the role via --role or --role-file.
If the role is not given, it is looked up using the accessor. If neither is given, an
interactive wizard lets you select the role first, offering only roles that have secret_ids,
and then the secret_id to delete, showing its status, expiration and metadata. The selection of the secret_id is skipped if the role
has only a single one.

Unless --force is given, the deletion has to be confirmed.`,
	Example: `  sc vault approle delete-secret-id
  sc vault approle delete-secret-id -r my-role
  sc vault approle delete-secret-id 2f4e1c3a-0f9b-4d2e-9a6c-1b2c3d4e5f60 -f`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		roleName, err := readValueOrFile(pkg.GetString(cmd, vaultApproleRoleName), pkg.GetString(cmd, vaultApproleRoleNameFile), "role")
		if err != nil {
			log.Fatal().Err(err).Msg("could not read role name")
		}

		accessor := pkg.GetString(cmd, vaultApproleAccessor)
		if len(args) > 0 {
			accessor = args[0]
		}

		interactive := isTerminal(os.Stdin)
		if accessor == "" && !interactive {
			log.Fatal().Msgf("the accessor must be given as argument or using --%s", vaultApproleAccessor)
		}

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		switch {
		case roleName == "" && accessor != "":
			roleName, err = tui.LoadWithSpinner(cmd.Context(), "Looking up role of accessor...", func(ctx context.Context) (string, error) {
				return findApproleOfAccessor(ctx, client, mount, accessor)
			})
			if err != nil {
				log.Fatal().Err(err).Msg("could not look up role of accessor")
			}
			log.Info().Msgf("Accessor belongs to role %q", roleName)
		case roleName == "":
			roleName = selectApproleWithSecretIDs(cmd.Context(), client, mount)
			if roleName == "" {
				log.Info().Msg("No AppRole has any secret_ids, nothing to delete")
				return
			}
		}

		if accessor == "" {
			accessor = selectApproleSecretID(cmd.Context(), client, mount, roleName)
			if accessor == "" {
				log.Info().Msgf("Role %q has no secret_ids, nothing to delete", roleName)
				return
			}
		}

		if !pkg.MustGetBool(cmd, vaultForce) {
			if !interactive {
				log.Fatal().Msgf("deletion can not be confirmed without terminal, use --%s", vaultForce)
			}
			question := fmt.Sprintf("Really delete secret_id with accessor %q of role %q?", accessor, roleName)
			if tui.SelectInput(question, []string{"no", "yes"}) != "yes" {
				log.Info().Msg("Aborted, secret_id has not been deleted")
				return
			}
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), vaultDefaultTimeout)
		defer cancel()

		if err := client.DeleteAppRoleAccessor(ctx, mount, roleName, accessor); err != nil {
			log.Fatal().Err(err).Msg("could not delete secret_id")
		}

		log.Info().Msgf("Deleted secret_id with accessor %q of role %q", accessor, roleName)
	},
}

// findApproleOfAccessor returns the name of the AppRole the given secret_id accessor belongs to.
func findApproleOfAccessor(ctx context.Context, client *vault.VaultClient, mount, accessor string) (string, error) {
	listCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
	roles, err := client.ListAppRoles(listCtx, mount)
	cancel()
	if err != nil {
		return "", err
	}

	for _, role := range roles {
		roleCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
		accessors, err := client.GetAppRoleAccessors(roleCtx, mount, role)
		cancel()
		if err != nil {
			log.Warn().Err(err).Msgf("could not list accessors of role %q", role)
			continue
		}
		if slices.Contains(accessors, accessor) {
			return role, nil
		}
	}

	return "", fmt.Errorf("no role found for accessor %q", accessor)
}

// selectApproleWithSecretIDs lets the user select one of the AppRoles of the mount that have at least one
// secret_id. Returns an empty string if no role has a secret_id. Falls back to free-form input if the AppRoles can
// not be listed.
func selectApproleWithSecretIDs(ctx context.Context, client *vault.VaultClient, mount string) string {
	roles, err := tui.LoadWithSpinner(ctx, "Loading AppRoles with secret_ids...", func(ctx context.Context) ([]string, error) {
		return listApprolesWithSecretIDs(ctx, client, mount)
	})
	if err != nil {
		log.Warn().Err(err).Msg("could not list AppRoles, falling back to manual input")
		return tui.ReadInput("Enter AppRole", nil)
	}

	if len(roles) == 0 {
		return ""
	}

	return tui.SelectInput("Select AppRole", roles)
}

// listApprolesWithSecretIDs returns the sorted names of the AppRoles that have at least one secret_id. Roles whose
// secret_ids can not be listed are skipped.
func listApprolesWithSecretIDs(ctx context.Context, client *vault.VaultClient, mount string) ([]string, error) {
	listCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
	roles, err := client.ListAppRoles(listCtx, mount)
	cancel()
	if err != nil {
		return nil, err
	}
	sort.Strings(roles)

	var ret []string
	for _, role := range roles {
		roleCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
		accessors, err := client.GetAppRoleAccessors(roleCtx, mount, role)
		cancel()
		if err != nil {
			log.Warn().Err(err).Msgf("could not list secret_ids of role %q", role)
			continue
		}
		if len(accessors) > 0 {
			ret = append(ret, role)
		}
	}
	return ret, nil
}

// selectApproleSecretID lets the user select one of the secret_ids of the role and returns its accessor. If the
// role has only a single secret_id, it is returned without asking. Returns an empty string if the role has no
// secret_ids.
func selectApproleSecretID(ctx context.Context, client *vault.VaultClient, mount, roleName string) string {
	var roles []approleSecretIDs
	var errs error
	err := tui.RunWithSpinner(ctx, "Loading secret_ids...", func(ctx context.Context) error {
		roles, errs = listApproleSecretIDs(ctx, client, mount, roleName)
		return nil
	})
	if err == nil && len(roles) == 0 {
		err = errs
	}
	if err != nil {
		log.Fatal().Err(err).Msgf("could not list secret_ids of role %q", roleName)
	}
	if errs != nil {
		log.Warn().Err(errs).Msg("could not look up all secret_ids")
	}

	secretIDs := roles[0].SecretIDs
	switch len(secretIDs) {
	case 0:
		return ""
	case 1:
		log.Info().Msgf("Using only secret_id of role %q with accessor %q", roleName, secretIDs[0].Accessor)
		return secretIDs[0].Accessor
	}

	accessors := make([]string, 0, len(secretIDs))
	for _, secretID := range secretIDs {
		accessors = append(accessors, secretID.Accessor)
	}
	return tui.SelectLabeledInput("Select secret_id to delete", approleSecretIDLabels(secretIDs), accessors)
}

// approleSecretIDLabels returns aligned labels describing the secret_ids, in the same order.
func approleSecretIDLabels(secretIDs []approleSecretID) []string {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for _, s := range secretIDs {
		_, _ = fmt.Fprintf(tw, "%s\t%s\texpires %s\t%s\n", s.Accessor, s.Status, formatApproleTime(s.ExpirationTime, "never"), formatApproleMetadata(s.Metadata))
	}
	_ = tw.Flush()

	return strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
}

func init() {
	vaultApproleCmd.AddCommand(vaultApproleDeleteSecretIdCmd)

	vaultApproleDeleteSecretIdCmd.Flags().String(vaultApproleAccessor, "", "Accessor of the secret_id to delete")
	vaultApproleDeleteSecretIdCmd.Flags().BoolP(vaultForce, "f", false, "Skip the confirmation prompt")
}
