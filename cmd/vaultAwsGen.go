package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var vaultAwsGenerateCredentialsCmd = &cobra.Command{
	Use:     "generate-credentials",
	Aliases: []string{"gen-credentials", "gen"},
	Short:   "Generate AWS credentials using Vault",
	Long: `The 'generate-credentials' command is part of the 'aws' command group, which provides tools
for working with the Vault AWS secrets engine.

This command is used to generate dynamic AWS credentials by interacting with Vault. 
It requires appropriate configuration and permissions set in Vault to access the AWS secrets engine.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		mount := pkg.GetString(cmd, vaultMountPath)
		profile := pkg.GetString(cmd, vaultAwsProfile)
		role := pkg.GetString(cmd, vaultRoleName)
		ttl := pkg.GetInt(cmd, vaultTtl)

		if role == "" {
			var availableRoles []string

			ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
			defer cancel()

			if err := spinner.New().
				Type(spinner.Line).
				ActionWithErr(func(ctx context.Context) error {
					roles, err := client.AwsListRoles(ctx, mount)
					if err == nil {
						availableRoles = roles
					}
					return nil
				}).
				Title("Loading available roles...").
				Accessible(false).
				Context(ctx).
				Type(spinner.Dots).
				Run(); err != nil {
				log.Fatal().Err(err).Msg("sending login to request to Vault failed")
			}

			if len(availableRoles) > 0 {
				sort.Strings(availableRoles)
				role = tui.SelectInput("Enter role", availableRoles)
			} else {
				role = tui.ReadInput("Enter role", nil)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		var creds *vault.AwsCredentials
		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				var err error
				creds, err = client.AwsGenCreds(ctx, mount, role, strconv.Itoa(ttl))
				return err
			}).
			Title(fmt.Sprintf("Generating credentials for role %s", role)).
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("failed to generate credentials")
		}

		if err := updateAwsCredentialsFile(profile, *creds); err != nil {
			var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("access_key")
			fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), creds.AccessKeyID))

			outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("secret_key")
			fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), creds.SecretAccessKey))

			log.Fatal().Err(err).Msg("could not write credentials to file")
		}

		if err := spinner.New().
			Type(spinner.Line).
			Action(func() {
				time.Sleep(5 * time.Second)
			}).
			Title("Waiting for credentials to become effective").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Warn().Err(err).Msg("could not display spinner")
		}

		var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("Wrote credentials")
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), vaultAwsDefaultCredentialsFilename))
	},
}

func updateAwsCredentialsFile(profile string, creds vault.AwsCredentials) error {
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
	vaultAwsCmd.AddCommand(vaultAwsGenerateCredentialsCmd)

	vaultAwsGenerateCredentialsCmd.Flags().StringP(vaultMountPath, "m", vaultAwsDefaultMount, "Mount path for the AWS secret engine")
	vaultAwsGenerateCredentialsCmd.Flags().IntP(vaultTtl, "t", vaultAwsDefaultTtl, "Specify how long the credentials should be valid for in seconds")
	vaultAwsGenerateCredentialsCmd.Flags().StringP(vaultRoleName, "r", "", "Specifies the name of the role to generate credentials for")
	vaultAwsGenerateCredentialsCmd.Flags().StringP(vaultAwsProfile, "p", vaultAwsDefaultProfile, "Specifies the name of the AWS credentials profile section")
}
