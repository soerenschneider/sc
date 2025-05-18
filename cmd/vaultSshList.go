package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var sshListCmd = &cobra.Command{
	Use:     "list-roles",
	Aliases: []string{"list"},
	Short:   "Lists all roles for the SSH secrets engine",
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		mount := pkg.GetString(cmd, VaultMountPath)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		roles, err := client.SshListRoles(ctx, mount)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to list SSH roles for mount %q", mount)
		}

		var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("SSH roles")
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), strings.Join(roles, ", ")))
	},
}

func init() {
	sshCmd.AddCommand(sshListCmd)

	sshListCmd.Flags().StringP(VaultMountPath, "m", "ssh", "Mount path for the SSH secret engine")
}
