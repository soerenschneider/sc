package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/soerenschneider/sc/internal/vault"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// vaultLoginCmd represents the vaultLogin command
var vaultAwsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Manage AWS secret engine",
	Long: `The 'aws' command group contains subcommands for interacting with Vault AWS secret engine.

This command itself does not perform any actions. Instead, use one of its subcommands
to list roles, generate or revoke credentials and check when they expire.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	vaultCmd.AddCommand(vaultAwsCmd)
	vaultAwsCmd.PersistentFlags().StringP(vaultMountPath, "m", "", "Path where the AWS secret engine is mounted. If not specified, available mounts are discovered.")
}

const (
	awsCredsKeyAccessKeyID     = "aws_access_key_id"
	awsCredsKeySecretAccessKey = "aws_secret_access_key" //nolint G101
	awsCredsKeySessionToken    = "aws_session_token"     //nolint G101
	// The AWS SDKs ignore unknown keys, the x_ prefix follows the convention of saml2aws and others.
	awsCredsKeyVaultLeaseID = "x_vault_lease_id"
	awsCredsKeyExpires      = "x_security_token_expires"
)

var awsCredsKeys = []string{
	awsCredsKeyAccessKeyID,
	awsCredsKeySecretAccessKey,
	awsCredsKeySessionToken,
	awsCredsKeyVaultLeaseID,
	awsCredsKeyExpires,
}

// awsCredentialsProfile is a profile of the AWS credentials file as written by generate-credentials.
type awsCredentialsProfile struct {
	Name         string
	AccessKeyID  string
	HasSession   bool
	LeaseID      string
	Expiration   time.Time
	ExpiresKnown bool
}

func awsCredentialsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".aws", "credentials"), nil
}

// loadAwsCredentialsFile returns the parsed credentials file, or an empty one if it does not exist yet.
func loadAwsCredentialsFile() (*ini.File, string, error) {
	credsPath, err := awsCredentialsFilePath()
	if err != nil {
		return nil, "", err
	}

	cfg := ini.Empty()
	if _, err := os.Stat(credsPath); err == nil {
		if err := cfg.Append(credsPath); err != nil {
			return nil, "", fmt.Errorf("failed to load existing credentials file: %w", err)
		}
	}

	return cfg, credsPath, nil
}

func updateAwsCredentialsFile(profile string, creds vault.AwsCredentials) error {
	cfg, credsPath, err := loadAwsCredentialsFile()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(credsPath), 0700); err != nil {
		return fmt.Errorf("failed to create .aws directory: %w", err)
	}

	section, err := cfg.GetSection(profile)
	if err != nil {
		section, err = cfg.NewSection(profile)
		if err != nil {
			return fmt.Errorf("failed to create profile section: %w", err)
		}
	}

	// Don't leave stale values from earlier credentials behind.
	setOrDelete := func(key, value string) {
		if value != "" {
			section.Key(key).SetValue(value)
		} else {
			section.DeleteKey(key)
		}
	}

	setOrDelete(awsCredsKeyAccessKeyID, creds.AccessKeyID)
	setOrDelete(awsCredsKeySecretAccessKey, creds.SecretAccessKey)
	setOrDelete(awsCredsKeySessionToken, creds.SessionToken)
	setOrDelete(awsCredsKeyVaultLeaseID, creds.LeaseID)
	var expires string
	if !creds.Expiration.IsZero() {
		expires = creds.Expiration.UTC().Format(time.RFC3339)
	}
	setOrDelete(awsCredsKeyExpires, expires)

	if err := cfg.SaveTo(credsPath); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// readAwsCredentialsProfiles returns the requested profile, or all profiles holding credentials
// issued by Vault if profile is empty.
func readAwsCredentialsProfiles(profile string) ([]awsCredentialsProfile, error) {
	cfg, credsPath, err := loadAwsCredentialsFile()
	if err != nil {
		return nil, err
	}

	var sections []*ini.Section
	if profile != "" {
		section, err := cfg.GetSection(profile)
		if err != nil {
			return nil, fmt.Errorf("profile %q not found in %s", profile, credsPath)
		}
		sections = append(sections, section)
	} else {
		for _, section := range cfg.Sections() {
			if section.HasKey(awsCredsKeyVaultLeaseID) {
				sections = append(sections, section)
			}
		}
	}

	ret := make([]awsCredentialsProfile, 0, len(sections))
	for _, section := range sections {
		p := awsCredentialsProfile{
			Name:        section.Name(),
			AccessKeyID: section.Key(awsCredsKeyAccessKeyID).String(),
			HasSession:  section.Key(awsCredsKeySessionToken).String() != "",
			LeaseID:     section.Key(awsCredsKeyVaultLeaseID).String(),
		}
		if expires := section.Key(awsCredsKeyExpires).String(); expires != "" {
			if t, err := time.Parse(time.RFC3339, expires); err == nil {
				p.Expiration = t
				p.ExpiresKnown = true
			}
		}
		ret = append(ret, p)
	}

	return ret, nil
}

// removeAwsCredentials removes the credentials from the given profile. Other settings of the
// profile are kept, the profile is removed entirely if nothing else is left.
func removeAwsCredentials(profile string) error {
	cfg, credsPath, err := loadAwsCredentialsFile()
	if err != nil {
		return err
	}

	section, err := cfg.GetSection(profile)
	if err != nil {
		return nil
	}

	for _, key := range awsCredsKeys {
		section.DeleteKey(key)
	}
	if len(section.Keys()) == 0 {
		cfg.DeleteSection(profile)
	}

	if err := cfg.SaveTo(credsPath); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}
