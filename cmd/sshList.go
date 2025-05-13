package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var sshListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all roles for the SSH secrets engine",
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustGetVaultClient(cmd))

		mount := pkg.GetString(cmd, VaultMountPath)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		roles, err := vaultAwsListRoles(ctx, client, mount)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to list SSH roles for mount %q", mount)
		}

		var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("SSH roles")
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), strings.Join(roles, ", ")))
	},
}

func vaultSshListRoles(ctx context.Context, client *api.Client, mount string) ([]string, error) {
	path := fmt.Sprintf("%s/roles", mount)
	secret, err := client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read AWS credentials: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data returned from Vault")
	}

	var keys []string
	for _, v := range secret.Data["keys"].([]any) {
		if s, ok := v.(string); ok {
			keys = append(keys, s)
		}
	}

	return keys, nil
}

func init() {
	sshCmd.AddCommand(sshListCmd)

	sshListCmd.Flags().StringP(VaultMountPath, "m", "ssh", "Mount path for the SSH secret engine")
}
