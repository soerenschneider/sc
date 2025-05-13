package cmd

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

const (
	awsProfile = "aws-profile"

	defaultAwsTtl              = "3600s"
	defaultCredentialsFilename = "~/.aws/credentials"
	defaultProfile             = "default"
)

type AwsCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

// vaultLoginCmd represents the vaultLogin command
var vaultAwsGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Manage AWS secret engine",
	Long: `The 'aws' command group contains subcommands for interacting with Vault AWS secret engine.

This command itself does not perform any actions. Instead, use one of its subcommands
to inspect or manage tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustGetVaultClient(cmd))

		mount := pkg.GetString(cmd, VaultMountPath)
		profile := pkg.GetString(cmd, awsProfile)
		role := pkg.GetString(cmd, VaultRoleName)
		ttl := pkg.GetInt(cmd, VaultTtl)

		if role == "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			availableRoles, err := vaultAwsListRoles(ctx, client, mount)
			if err == nil {
				role = huhSelectInput("Enter role", availableRoles)
			} else {
				role = huhReadInput("Enter role", nil)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		creds, err := vaultAwsGenCreds(ctx, client, mount, role, strconv.Itoa(ttl))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to generate credentials")
		}

		if err := updateAwsCredentialsFile(profile, *creds); err != nil {
			var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("access_key")
			fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), creds.AccessKeyID))

			outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("secret_key")
			fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), creds.SecretAccessKey))

			log.Fatal().Err(err).Msg("could not write credentials to file")
		}

		var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("Wrote credentials")
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), defaultCredentialsFilename))
	},
}

func vaultAwsGenCreds(ctx context.Context, client *api.Client, mount string, role string, ttl string) (*AwsCredentials, error) {
	path := fmt.Sprintf("%s/creds/%s", mount, role)

	options := map[string][]string{
		"ttl": []string{cmp.Or(ttl, defaultAwsTtl)},
	}

	secret, err := client.Logical().ReadWithDataWithContext(ctx, path, options)
	if err != nil {
		return nil, fmt.Errorf("failed to read AWS credentials: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data returned from Vault")
	}

	accessKey, ok1 := secret.Data["access_key"].(string)
	secretKey, ok2 := secret.Data["secret_key"].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("unexpected response structure: %#v", secret.Data)
	}

	return &AwsCredentials{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}, nil
}

func updateAwsCredentialsFile(profile string, creds AwsCredentials) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	awsDir := filepath.Join(homeDir, ".aws")
	credsPath := filepath.Join(awsDir, "credentials")

	if err := os.MkdirAll(awsDir, 0700); err != nil {
		return fmt.Errorf("failed to create .aws directory: %w", err)
	}

	cfg := ini.Empty()
	if _, err := os.Stat(credsPath); err == nil {
		if err := cfg.Append(credsPath); err != nil {
			return fmt.Errorf("failed to load existing credentials file: %w", err)
		}
	}

	section, err := cfg.GetSection(profile)
	if err != nil {
		section, err = cfg.NewSection(profile)
		if err != nil {
			return fmt.Errorf("failed to create profile section: %w", err)
		}
	}
	section.Key("aws_access_key_id").SetValue(creds.AccessKeyID)
	section.Key("aws_secret_access_key").SetValue(creds.SecretAccessKey)

	if err := cfg.SaveTo(credsPath); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

func init() {
	vaultAwsCmd.AddCommand(vaultAwsGenCmd)

	vaultAwsGenCmd.Flags().StringP(VaultMountPath, "m", defaultAwsMount, "Mount path for the AWS secret engine")
	vaultAwsGenCmd.Flags().IntP(VaultTtl, "t", 3600, "Specify how long the credentials should be valid for in seconds")
	vaultAwsGenCmd.Flags().StringP(VaultRoleName, "r", "", "Specifies the name of the role to generate credentials for")
	vaultAwsGenCmd.Flags().StringP(awsProfile, "p", defaultProfile, "Specifies the name of the AWS credentials profile section")
}
