package vault

import (
	"cmp"
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

type AwsCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

type VaultClient struct {
	client *vault.Client
}

func MustBuildClient(command *cobra.Command) *VaultClient {
	vaultAddr, err := GetVaultAddress(command)
	if err != nil {
		log.Fatal().Err(err).Msg("could not get vault address")
	}

	config := vault.DefaultConfig()
	config.Address = vaultAddr

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatal().Msgf("Error creating Vault client: %v", err)
	}

	return &VaultClient{
		client: client,
	}
}

func MustAuthenticateClient(client *VaultClient) *VaultClient {
	if client == nil {
		log.Fatal().Msg("nil client passed")
	}

	token, err := GetVaultToken()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get Vault token")
	}

	client.client.SetToken(token)
	return client
}

func GetVaultAddress(cmd *cobra.Command) (string, error) {
	address, err := cmd.Flags().GetString("addr")
	if err == nil && len(address) > 0 {
		return address, nil
	}

	address, found := os.LookupEnv(vaultAddrEnvVarKey)
	if !found {
		return "", errors.New("no vault address specified")
	}

	log.Info().Msgf("No vault address supplied explicitly, using value of env var %s=%s", vaultAddrEnvVarKey, address)
	return address, nil
}

func GetVaultToken() (string, error) {
	token, ok := os.LookupEnv(vaultTokenEnvVarKey)
	if ok {
		return token, nil
	}

	tokenFilePath, err := GetVaultTokenFilePath()
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
func GetVaultTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("could not get user home directory")
	}
	return filepath.Join(homeDir, ".vault-token"), nil
}

func (c *VaultClient) AwsListRoles(ctx context.Context, mount string) ([]string, error) {
	path := fmt.Sprintf("%s/roles", mount)
	secret, err := c.client.Logical().ListWithContext(ctx, path)
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

func (c *VaultClient) LookupToken(ctx context.Context) (*vault.Secret, error) {
	return c.client.Auth().Token().LookupSelfWithContext(ctx)
}

func (c *VaultClient) AwsGenCreds(ctx context.Context, mount string, role string, ttl string) (*AwsCredentials, error) {
	path := fmt.Sprintf("%s/creds/%s", mount, role)

	options := map[string][]string{
		"ttl": {cmp.Or(ttl, "3600s")},
	}

	secret, err := c.client.Logical().ReadWithDataWithContext(ctx, path, options)
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

func (c *VaultClient) SshListRoles(ctx context.Context, mount string) ([]string, error) {
	path := fmt.Sprintf("%s/roles", mount)
	secret, err := c.client.Logical().ListWithContext(ctx, path)
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

func (c *VaultClient) SendPassword(ctx context.Context, mount, username, password string) (*vault.Secret, error) {
	data := map[string]any{"password": password}
	path := fmt.Sprintf("auth/%s/login/%s", mount, username)
	return c.client.Logical().WriteWithContext(ctx, path, data)
}

func (c *VaultClient) UpdatePassword(ctx context.Context, mount, username, password string) error {
	userPath := fmt.Sprintf("auth/%s/users/%s", mount, username)
	data := map[string]interface{}{
		"password": password,
	}

	_, err := c.client.Logical().WriteWithContext(ctx, userPath, data)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func GetVaultMfaRequirement(secret *vault.Secret) (*vault.MFARequirement, string, error) {
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

func (c *VaultClient) SendOtp(ctx context.Context, mfaReq *vault.MFARequirement, otp string) (*vault.Secret, error) {
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

	return c.client.Logical().WriteWithContext(ctx, "sys/mfa/validate", mfaPayload)
}

func (c *VaultClient) LoginSinglePhase(ctx context.Context, username, password, mfaId, otp string) (*vault.Secret, error) {
	var vaultMfaHeaderValue = mfaId
	if otp != "" {
		vaultMfaHeaderValue = fmt.Sprintf("%s: %s", mfaId, otp)
	}
	c.client.SetHeaders(map[string][]string{
		"X-Vault-MFA": {
			vaultMfaHeaderValue,
		},
	})
	defer c.client.SetHeaders(nil)

	data := map[string]any{"password": password}
	path := fmt.Sprintf("auth/userpass/login/%s", username)
	return c.client.Logical().WriteWithContext(ctx, path, data)
}
