package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	vaultAddrEnvVarKey  = "VAULT_ADDR"
	vaultTokenEnvVarKey = "VAULT_TOKEN"
)

func getVaultAddress(cmd *cobra.Command) (string, error) {
	address, err := cmd.Flags().GetString(sshCmdFlagsVaultAddress)
	if err == nil && len(address) > 0 {
		return address, nil
	}
	address = os.Getenv(vaultAddrEnvVarKey)
	if len(address) == 0 {
		return "", errors.New("no vault address specified")
	}
	log.Info().Msgf("No vault address supplied explicitly, using value of env var %s=%s", vaultAddrEnvVarKey, address)
	return address, nil
}

// getVaultAddr detects the Vault address from the environment or defaults to localhost
func getVaultAddr() string {
	addr, _ := os.LookupEnv(vaultAddrEnvVarKey)
	return addr
}

func getVaultToken() (string, error) {
	token, ok := os.LookupEnv(vaultTokenEnvVarKey)
	if ok {
		return token, nil
	}

	tokenFilePath, err := getVaultTokenFilePath()
	if err != nil {
		return "", fmt.Errorf("could not get default Vault token file: %w", err)
	}

	tokenData, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", fmt.Errorf("could not read Vault token from file: %w", err)
	}

	return string(tokenData), nil
}

// getVaultTokenFilePath returns the default Vault token file path
func getVaultTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("could not get user home directory")
	}
	return filepath.Join(homeDir, ".vault-token"), nil
}

func vaultSendPassword(ctx context.Context, client *vault.Client, username, password string) (*vault.Secret, error) {
	data := map[string]any{"password": password}
	path := fmt.Sprintf("auth/userpass/login/%s", username)
	return client.Logical().WriteWithContext(ctx, path, data)
}

func getVaultMfaRequirement(secret *vault.Secret) (*vault.MFARequirement, string, error) {
	if secret == nil {
		return nil, "", errors.New("empty secret supplied")
	}

	mfaRequirement := secret.Auth.MFARequirement
	defaultMfa, found := mfaRequirement.MFAConstraints["default"]
	if !found {
		return nil, "", errors.New("no default mfa requirement available")
	}

	if len(defaultMfa.Any) == 0 {
		return nil, "", errors.New("no mfa requirements provided")
	}

	method := defaultMfa.Any[0]
	mfaType := method.Type
	if mfaType == "totp" {
		return mfaRequirement, method.ID, nil
	}

	return nil, "", errors.New("no totp available")
}

func vaultSendOtp(ctx context.Context, client *vault.Client, mfaReq *vault.MFARequirement, otp string) (*vault.Secret, error) {
	defaultMfa, found := mfaReq.MFAConstraints["default"]
	if !found {
		return nil, errors.New("no default mfa requirement available")
	}

	if len(defaultMfa.Any) == 0 {
		return nil, errors.New("no mfa requirements provided")
	}

	method := defaultMfa.Any[0]
	mfaId := method.ID

	mfaPayload := map[string]any{
		"mfa_request_id": mfaReq.MFARequestID,
		"mfa_payload": map[string]any{
			mfaId: []string{otp},
		},
	}

	return client.Logical().WriteWithContext(ctx, "sys/mfa/validate", mfaPayload)
}

func vaultLoginSinglePhase(ctx context.Context, client *vault.Client, username, password, mfaId, otp string) (*vault.Secret, error) {
	var vaultMfaHeaderValue = mfaId
	if otp != "" {
		vaultMfaHeaderValue = fmt.Sprintf("%s: %s", mfaId, otp)
	}
	client.SetHeaders(map[string][]string{
		"X-Vault-MFA": {
			vaultMfaHeaderValue,
		},
	})
	defer client.SetHeaders(nil)

	data := map[string]any{"password": password}
	path := fmt.Sprintf("auth/userpass/login/%s", username)
	return client.Logical().WriteWithContext(ctx, path, data)
}
