package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/pki"
	"github.com/soerenschneider/sc/internal/vault"
	vaultSecretFormatter "github.com/soerenschneider/sc/internal/vault/formatter"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type secretFormatter interface {
	Format(data map[string]string) ([]byte, error)
}

type config struct {
	Items map[string]vault.Kv2SyncConfig `validate:"required,dive"`
}

const (
	vaultSecretSyncFormatterYamlKey                    = "yaml"
	vaultSecretSyncFormatterJsonKey                    = "json"
	vaultSecretSyncFormatterEnvKey                     = "env"
	vaultSecretSyncFormatterEnvOptionUppercaseKeys     = "uppercase_keys"
	vaultSecretSyncFormatterEnvOptionValueOnly         = "value_only"
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

		config, err := readSecretsConfiguration(syncConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("could not read secrets configuration")
		}

		if len(config.Items) == 0 {
			log.Fatal().Msg("no secrets configured")
		}

		if secretName == "" && !syncAllSecrets {
			keys := make([]string, 0, len(config.Items))
			for k := range config.Items {
				keys = append(keys, k)
			}

			sort.Strings(keys)

			log.Fatal().
				Strs("available_secrets", keys).
				Msg("No secret name given")
		}

		if err := validator.New().Struct(config); err != nil {
			log.Fatal().Err(err).Msg("could not validate secrets configuration")
		}

		names := []string{secretName}
		if syncAllSecrets {
			names = make([]string, 0, len(config.Items))
			for name := range config.Items {
				names = append(names, name)
			}
			sort.Strings(names)
		} else if _, found := config.Items[secretName]; !found {
			log.Fatal().Msgf("secret %q not found in configuration", secretName)
		}

		failed := 0
		for _, name := range names {
			item := config.Items[name]
			changed, err := syncItemWithTimeout(cmd.Context(), client, mount, item)
			switch {
			case vault.IsNotFound(err):
				failed++
				log.Error().Msgf("Could not sync %q: secret %q does not exist in mount %q", name, item.SecretPath, mount)
			case err != nil:
				failed++
				log.Error().Err(err).Msgf("Could not sync %q from %q", name, item.SecretPath)
			case changed:
				log.Info().Msgf("Synced %q to %s", name, item.DestUri)
			default:
				log.Info().Msgf("%s is up to date with %q", item.DestUri, name)
			}
		}

		if failed > 0 {
			log.Fatal().Msgf("Could not sync %d of %d secret(s)", failed, len(names))
		}
	},
}

func syncItemWithTimeout(ctx context.Context, client *vault.VaultClient, mount string, item vault.Kv2SyncConfig) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return syncItem(ctx, client, mount, item)
}

// syncItem writes the secret to its destination and reports whether the
// destination changed.
func syncItem(ctx context.Context, client *vault.VaultClient, mount string, item vault.Kv2SyncConfig) (bool, error) {
	formatImpl, err := buildSecretFormatter(item.Formatter, item.FormatterArgs)
	if err != nil {
		return false, fmt.Errorf("could not build secret formatter: %w", err)
	}

	storage, err := pki.NewFilesystemStorageFromUri(pkg.GetExpandedFile(item.DestUri))
	if err != nil {
		return false, fmt.Errorf("could not create storage from uri %s: %w", item.DestUri, err)
	}

	read, err := client.Get(ctx, mount, item.SecretPath)
	if err != nil {
		return false, fmt.Errorf("could not read secret from vault: %w", err)
	}

	formatted, err := formatImpl.Format(read.Data)
	if err != nil {
		return false, fmt.Errorf("formatting secret failed: %w", err)
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

// updateFile writes data unless the storage already holds it and reports
// whether it was written.
func updateFile(data []byte, storageImpl vault.StorageImplementation) (bool, error) {
	hash := hashContent(data)

	diskContent, err := storageImpl.Read()
	if err == nil {
		diskHash := hashContent(diskContent)
		if diskHash == hash {
			// file exists locally and is identical to the item we downloaded, we're done
			return false, nil
		}
	}

	if err := storageImpl.Write(data); err != nil {
		return false, err
	}
	return true, nil
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
		valueOnly := false
		if arguments != nil {
			val, found := arguments[vaultSecretSyncFormatterEnvOptionUppercaseKeys]
			if found {
				convertedVal, success := val.(bool)
				if success {
					uppercaseKeys = convertedVal
				}
			}

			val, found = arguments[vaultSecretSyncFormatterEnvOptionValueOnly]
			if found {
				convertedVal, success := val.(bool)
				if success {
					valueOnly = convertedVal
				}
			}
		}

		return vaultSecretFormatter.NewEnvVarFormatter(uppercaseKeys, valueOnly), nil
	case vaultSecretSyncFormatterYamlKey:
		return &vaultSecretFormatter.YamlFormatter{}, nil
	case vaultSecretSyncFormatterJsonKey:
		return &vaultSecretFormatter.JsonFormatter{}, nil
	case vaultSecretSyncFormatterTemplateKey:
		if arguments == nil {
			return nil, errors.New("no logsFormatter arguments found")
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
			return vaultSecretFormatter.NewTemplateFormatterFromFile(pkg.GetExpandedFile(templateFile))
		}

		return vaultSecretFormatter.NewTemplateFormatterFromTemplate(template)
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
