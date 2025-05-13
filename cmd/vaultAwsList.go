package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	api "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
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

The command expects that the AWS secrets engine has been enabled and configured.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustGetVaultClient(cmd))

		mount := pkg.GetString(cmd, VaultMountPath)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		roles, err := vaultAwsListRoles(ctx, client, mount)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to list AWS roles for mount %q", mount)
		}

		var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("AWS roles")
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), strings.Join(roles, ", ")))
	},
}

func vaultAwsListRoles(ctx context.Context, client *api.Client, mount string) ([]string, error) {
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
	vaultAwsCmd.AddCommand(vaultAwsListCmd)

	vaultAwsListCmd.Flags().StringP(VaultMountPath, "m", defaultAwsMount, "Mount path for the AWS secret engine")
}
