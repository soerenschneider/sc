package cmd

import (
	"context"
	"os"
	"sort"

	"charm.land/lipgloss/v2"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/internal/vault/prompt"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var vaultApproleReadCmd = &cobra.Command{
	Use:     "read [role]",
	Aliases: []string{"get", "show"},
	Short:   "Display the configuration of a Vault AppRole",
	Long: `Displays the configuration of a Vault AppRole, including its role_id if it can be read.

The role may be passed as argument or via --role / --role-file. If it is not specified, all
AppRoles are listed in an interactive browser: select a role to display its configuration and
go back to the list with esc. The role displayed when quitting with q is printed. If the
AppRoles can not be listed, the name is prompted for instead.

Examples:
  sc vault approle read
  sc vault approle read my-role
  sc vault approle read -r my-role -m my-approle-mount`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		roleName, err := readValueOrFile(pkg.GetString(cmd, vaultApproleRoleName), pkg.GetString(cmd, vaultApproleRoleNameFile), "role")
		if err != nil {
			log.Fatal().Err(err).Msg("could not read role name")
		}
		if len(args) > 0 {
			roleName = args[0]
		}

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		readRole := func(ctx context.Context, name string) (string, error) {
			ctx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
			defer cancel()

			data, err := client.ReadAppRole(ctx, mount, name)
			if err != nil {
				return "", err
			}
			return tui.RenderMap(data), nil
		}

		if roleName == "" {
			roles, err := prompt.Load(cmd.Context(), "Loading AppRoles...", func(ctx context.Context) ([]string, error) {
				return client.ListAppRoles(ctx, mount)
			})
			if err != nil {
				log.Warn().Err(err).Msg("could not list AppRoles, falling back to manual input")
			}
			sort.Strings(roles)

			if len(roles) > 1 && isTerminal(os.Stdin) && isTerminal(os.Stdout) {
				result, err := tui.Browse(cmd.Context(), "Vault AppRoles", roles, readRole)
				if err != nil {
					log.Fatal().Err(err).Msg("could not browse AppRoles")
				}
				if result.Name != "" {
					printApprole(result.Name, result.Content)
				}
				return
			}

			roleName = tui.SelectOrReadInput("AppRole", roles, nil)
		}

		content, err := readRole(cmd.Context(), roleName)
		if err != nil {
			log.Fatal().Err(err).Msg("could not read AppRole")
		}

		printApprole(roleName, content)
	},
}

// printApprole prints the formatted configuration of an AppRole, preceded by a header with its name if stdout is
// a terminal.
func printApprole(name, content string) {
	tui.PrintHeader("AppRole " + name)
	_, _ = lipgloss.Println(content)
}

func init() {
	vaultApproleCmd.AddCommand(vaultApproleReadCmd)
}
