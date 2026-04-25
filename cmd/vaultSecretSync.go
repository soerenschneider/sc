package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/pki"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/internal/vault/formatter"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

type secretFormatter interface {
	Format(data map[string]any) ([]byte, error)
}

type config struct {
	Items map[string]vault.Kv2SyncConfig `validate:"required,dive"`
}

const (
	vaultSecretSyncFormatterYamlKey                    = "yaml"
	vaultSecretSyncFormatterJsonKey                    = "json"
	vaultSecretSyncFormatterEnvKey                     = "env"
	vaultSecretSyncFormatterEnvOptionUppercaseKeys     = "uppercase_keys"
	vaultSecretSyncFormatterTemplateKey                = "template"
	vaultSecretSyncFormatterTemplateOptionTemplateFile = "file"
	vaultSecretSyncFormatterTemplateOptionTemplate     = "template"
)

var vaultSecretSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a KV2 secret to disk",
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.MustGetString(cmd, vaultMountPath)
		syncConfig := pkg.MustGetString(cmd, vaultSecretSyncConfig)
		secretName := pkg.GetString(cmd, vaultSecretSyncName)
		syncAllSecrets, _ := pkg.GetBool(cmd, vaultSecretSyncAll)

		if secretName == "" && !syncAllSecrets {
			log.Fatal().Msg("No secret name given")
		}

		config, err := readSecretsConfiguration(syncConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("could not read secrets configuration")
		}

		if err := validator.New().Struct(config); err != nil {
			log.Fatal().Err(err).Msg("could not validate secrets configuration")
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		if !syncAllSecrets {
			item, found := config.Items[secretName]
			if !found {
				log.Fatal().Msgf("secret %q not found in configuration", secretName)
			}

			if err := syncItem(ctx, client, mount, item); err != nil {
				log.Fatal().Err(err).Msg("could not sync secret")
			}
		}

		var errs error
		for _, replicationConfig := range config.Items {
			errs = multierr.Append(errs, syncItem(ctx, client, mount, replicationConfig))
		}

		if errs != nil {
			log.Fatal().Err(errs).Msg("could not sync all secrets")
		}
	},
}

func syncItem(ctx context.Context, client *vault.VaultClient, mount string, item vault.Kv2SyncConfig) error {
	formatImpl, err := buildSecretFormatter(item.Formatter, item.FormatterArgs)
	if err != nil {
		return fmt.Errorf("could not build secret formatter: %w", err)
	}

	storage, err := pki.NewFilesystemStorageFromUri(pkg.GetExpandedFile(item.DestUri))
	if err != nil {
		return fmt.Errorf("could not create storage from uri %s: %w", item.DestUri, err)
	}

	read, err := client.ReadSecret(ctx, mount, item.SecretPath)
	if err != nil {
		return fmt.Errorf("could not read secret from vault: %w", err)
	}

	formatted, err := formatImpl.Format(read.Data)
	if err != nil {
		return fmt.Errorf("formatting secret failed: %w", err)
	}

	return updateFile(formatted, storage)
}

func readSecretsConfiguration(file string) (*config, error) {
	data, err := os.ReadFile(pkg.GetExpandedFile(file))
	if err != nil {
		return nil, fmt.Errorf("reading configuration for secrets to sync failed: %w", err)
	}

	var secretsConfig map[string]vault.Kv2SyncConfig
	if err := yaml.Unmarshal(data, &secretsConfig); err != nil {
		return nil, fmt.Errorf("parsing configuration for secrets to sync failed: %w", err)
	}

	return &config{
		Items: secretsConfig,
	}, nil
}

func updateFile(data []byte, storageImpl vault.StorageImplementation) error {
	hash := hashContent(data)

	diskContent, err := storageImpl.Read()
	if err == nil {
		diskHash := hashContent(diskContent)
		if diskHash == hash {
			// file exists locally and is identical to the item we downloaded, we're done
			return nil
		}
	}

	return storageImpl.Write(data)
}

func hashContent(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func buildSecretFormatter(name string, arguments map[string]any) (secretFormatter, error) {
	switch name {
	case vaultSecretSyncFormatterEnvKey:
		uppercaseKeys := false
		if arguments != nil {
			val, found := arguments[vaultSecretSyncFormatterEnvOptionUppercaseKeys]
			if found {
				convertedVal, success := val.(bool)
				if success {
					uppercaseKeys = convertedVal
				}
			}
		}

		return formatter.NewEnvVarFormatter(uppercaseKeys), nil
	case vaultSecretSyncFormatterYamlKey:
		return &formatter.YamlFormatter{}, nil
	case vaultSecretSyncFormatterJsonKey:
		return &formatter.JsonFormatter{}, nil
	case vaultSecretSyncFormatterTemplateKey:
		if arguments == nil {
			return nil, errors.New("no formatter arguments found")
		}
		var templateFile, template string
		val, found := arguments[vaultSecretSyncFormatterTemplateOptionTemplateFile]
		if found {
			convertedVal, success := val.(string)
			if success {
				templateFile = convertedVal
			}
		}

		val, found = arguments[vaultSecretSyncFormatterTemplateOptionTemplate]
		if found {
			convertedVal, success := val.(string)
			if success {
				template = convertedVal
			}
		}

		if template != "" && templateFile != "" {
			return nil, errors.New("both template and templateFile specified")
		}

		if templateFile != "" {
			return formatter.NewTemplateFormatterFromFile(pkg.GetExpandedFile(templateFile))
		}

		return formatter.NewTemplateFormatterFromTemplate(template)
	default:
		return nil, errors.New("no implementation found")
	}
}

func init() {
	vaultSecretCmd.AddCommand(vaultSecretSyncCmd)

	vaultSecretSyncCmd.Flags().BoolP(vaultSecretSyncAll, "a", false, "Flag to sync all secrets")
	vaultSecretSyncCmd.Flags().StringP(vaultSecretSyncName, "n", "", "Name of the secret to sync")
	vaultSecretSyncCmd.MarkFlagsMutuallyExclusive(vaultSecretSyncAll, vaultSecretSyncName)

	vaultSecretSyncCmd.Flags().StringP(vaultSecretSyncConfig, "c", "", "Config file that holds secret sync configuration")
	if err := vaultSecretSyncCmd.MarkFlagRequired(vaultSecretSyncConfig); err != nil {
		log.Fatal().Err(err).Msg("could not mark flag required")
	}
}
