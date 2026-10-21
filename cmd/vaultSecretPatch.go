package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const (
	vaultSecretPatchDelete = "delete"
	vaultSecretPatchStdin  = "-"
)

var vaultSecretPatchCmd = &cobra.Command{
	Use:   "patch <path> [key=value...]",
	Short: "Set or delete single keys of a Vault KV2 secret",
	Long: `Set or delete single keys of a Vault KV2 secret, while keeping all other keys untouched.

Keys are set using key=value pairs. If the value is '-', it is read from stdin instead, which
keeps it out of the shell history and the process list. A single trailing newline is removed
from the value read from stdin. Only one key can read its value from stdin.

Keys are deleted using --delete, which can be given multiple times.

The secret is changed using Vault's native patch operation, which merges the changes on the
server without reading the secret. This requires the 'patch' capability on the secret and
Vault 1.9 or newer.

If patching is not possible, because the capability is missing, Vault is too old or the
secret does not exist yet, the secret is read and written instead. This requires the 'read'
and 'update' (or 'create' for a new secret) capabilities. The secret is written using
check-and-set, so changes that were made in the meantime are not overwritten. If the latest
version of the secret is deleted, it is not changed.`,
	Example: `  sc vault secret patch my/app username=admin port=5432
  pass generate -n 32 | sc vault secret patch my/app password=-
  sc vault secret patch my/app --delete old_password --delete legacy_token`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		deletes := pkg.GetStringArray(cmd, vaultSecretPatchDelete)

		sets, err := parseSecretPatchPairs(args[1:], os.Stdin)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid arguments")
		}
		if len(sets) == 0 && len(deletes) == 0 {
			log.Fatal().Msgf("nothing to do, either key=value pairs or --%s must be given", vaultSecretPatchDelete)
		}

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		result, err := patchSecret(ctx, client, mount, path, sets, deletes)
		if vault.IsCASConflict(err) {
			log.Fatal().Msgf("secret %q has been changed in the meantime, nothing has been written, please try again", path)
		}
		if err != nil {
			log.Fatal().Msg(vault.HumanizeError(err, fmt.Sprintf("patching secret %q", path)))
		}

		for _, key := range result.missing {
			log.Warn().Msgf("Key %q does not exist, nothing to delete", key)
		}
		if result.secret == nil {
			log.Info().Msgf("Secret %q is unchanged, nothing has been written", path)
			return
		}
		if result.secret.Version > 0 {
			log.Info().Strs("changed", result.changed).Msgf("Wrote version %d of secret %q", result.secret.Version, path)
		} else {
			log.Info().Strs("changed", result.changed).Msgf("Wrote secret %q", path)
		}
	},
}

type vaultKv2Patcher interface {
	vaultKv2Provider

	// Patch changes single keys of a secret without reading it, keys with a nil value are deleted.
	Patch(ctx context.Context, mount, path string, data map[string]any) (*vault.Secret, error)
}

type secretPatchResult struct {
	// secret is the written secret, nil if nothing needed to be written.
	secret  *vault.Secret
	changed []string
	missing []string
}

// patchSecret changes the secret using a native patch. If that is not possible, it falls back to
// reading and writing the secret.
func patchSecret(ctx context.Context, store vaultKv2Patcher, mount, path string, sets map[string]string, deletes []string) (*secretPatchResult, error) {
	for _, key := range deletes {
		if _, ok := sets[key]; ok {
			return nil, fmt.Errorf("key %q can not be set and deleted at the same time", key)
		}
	}

	data := make(map[string]any, len(sets)+len(deletes))
	for key, value := range sets {
		data[key] = value
	}
	for _, key := range deletes {
		data[key] = nil
	}

	secret, err := store.Patch(ctx, mount, path, data)
	switch {
	case err == nil:
		return &secretPatchResult{secret: secret, changed: slices.Sorted(maps.Keys(data))}, nil
	case vault.IsNotFound(err), vault.IsPermissionDenied(err), vault.IsUnsupported(err):
		log.Debug().Err(err).Msg("Native patch not possible, reading and writing the secret instead")
		return patchSecretReadWrite(ctx, store, mount, path, sets, deletes)
	default:
		return nil, err
	}
}

// patchSecretReadWrite applies the changes to the current version of the secret and writes it using check-and-set.
func patchSecretReadWrite(ctx context.Context, store vaultKv2Provider, mount, path string, sets map[string]string, deletes []string) (*secretPatchResult, error) {
	data := map[string]string{}
	cas := 0
	current, err := store.Get(ctx, mount, path)
	switch {
	case err == nil:
		data = maps.Clone(current.Data)
		cas = current.Version
	case vault.IsNotFound(err):
		// The latest version may be deleted while older ones exist, patching an empty secret would lose their keys.
		if _, err := store.GetMetadata(ctx, mount, path); err == nil {
			return nil, errors.New("the latest version of the secret is deleted, use 'sc vault secret edit' to restore it")
		} else if !vault.IsNotFound(err) {
			return nil, err
		}
	default:
		return nil, err
	}

	result := &secretPatchResult{}
	for _, key := range slices.Sorted(maps.Keys(sets)) {
		if old, ok := data[key]; !ok || old != sets[key] {
			data[key] = sets[key]
			result.changed = append(result.changed, key)
		}
	}
	for _, key := range deletes {
		if _, ok := data[key]; !ok {
			result.missing = append(result.missing, key)
			continue
		}
		delete(data, key)
		result.changed = append(result.changed, key)
	}

	if len(result.changed) == 0 {
		return result, nil
	}

	result.secret, err = store.Put(ctx, mount, path, data, cas)
	return result, err
}

// parseSecretPatchPairs parses key=value pairs. A value of '-' is read from stdin.
func parseSecretPatchPairs(args []string, stdin io.Reader) (map[string]string, error) {
	sets := make(map[string]string, len(args))
	stdinKey := ""
	for _, arg := range args {
		key, value, found := strings.Cut(arg, "=")
		if !found || key == "" {
			return nil, fmt.Errorf("expected key=value, got %q", arg)
		}
		if _, ok := sets[key]; ok {
			return nil, fmt.Errorf("key %q is given multiple times", key)
		}

		if value == vaultSecretPatchStdin {
			if stdinKey != "" {
				return nil, fmt.Errorf("only one key can read its value from stdin, got %q and %q", stdinKey, key)
			}
			stdinKey = key

			raw, err := io.ReadAll(stdin)
			if err != nil {
				return nil, fmt.Errorf("could not read value of %q from stdin: %w", key, err)
			}
			value = strings.TrimSuffix(strings.TrimSuffix(string(raw), "\n"), "\r")
		}
		sets[key] = value
	}
	return sets, nil
}

func init() {
	vaultSecretCmd.AddCommand(vaultSecretPatchCmd)

	vaultSecretPatchCmd.Flags().StringArrayP(vaultSecretPatchDelete, "d", nil, "Delete this key from the secret, can be given multiple times")
}
