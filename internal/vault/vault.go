package vault

import (
	"cmp"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	backend "github.com/soerenschneider/sc/internal/pki"
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

type MfaCreateMethodRequest struct {
	MethodName string `json:"method_name"`
	Issuer     string `json:"issuer"`
	Algorithm  string `json:"algorithm"`
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
	if client == nil { //nolint SA5011
		log.Fatal().Msg("nil client passed")
	}

	token, err := GetVaultToken()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get Vault token")
	}

	client.client.SetToken(token) //nolint SA5011
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
		return nil, fmt.Errorf("failed to list roles: %w", err)
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

func (c *VaultClient) UserpassAuth(ctx context.Context, mount, username, password string) (*vault.Secret, error) {
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
	if mfaReq == nil {
		return nil, errors.New("nil mfarequirement passed")
	}

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

func (c *VaultClient) UserpassAuthMfa(ctx context.Context, username, password, mfaId, otp string) (*vault.Secret, error) {
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

func (c *VaultClient) TotpListMethods(ctx context.Context) ([]string, error) {
	path := "identity/mfa/method/totp"
	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, err
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

func structToMap(v any) (map[string]any, error) {
	var result map[string]any
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &result)
	return result, err
}

func (c *VaultClient) TotpGenerateMethod(ctx context.Context, req MfaCreateMethodRequest) (string, error) {
	path := "identity/mfa/method/totp"
	data, err := structToMap(req)
	if err != nil {
		return "", err
	}
	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return "", err
	}

	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("no data returned from Vault")
	}

	return secret.Data["method_id"].(string), nil
}

func (c *VaultClient) TotpDestroySecretAdmin(ctx context.Context, methodId, entityId string) error {
	path := "identity/mfa/method/totp/admin-destroy"
	data := map[string]any{
		"method_id": methodId,
		"entity_id": entityId,
	}
	_, err := c.client.Logical().WriteWithContext(ctx, path, data)
	return err
}

var ErrTotpAlreadyDefined = errors.New("TOTP already defined")
var ErrTotpDeleteFailed = errors.New("TOTP deletion failed")

func (c *VaultClient) TotpGenerateSecretAdmin(ctx context.Context, methodId, entityId string, force bool) (string, error) {
	path := "identity/mfa/method/totp/admin-generate"
	data := map[string]any{
		"method_id": methodId,
		"entity_id": entityId,
	}

	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return "", err
	}

	if secret == nil || secret.Data == nil || len(secret.Warnings) > 0 {
		log.Warn().Msgf("TOTP already defined for %s", entityId)
		if !force {
			return "", ErrTotpAlreadyDefined
		}

		if err := c.TotpDestroySecretAdmin(ctx, methodId, entityId); err != nil {
			return "", ErrTotpDeleteFailed
		}
		return c.TotpGenerateSecretAdmin(ctx, methodId, entityId, false)
	}

	return fmt.Sprintf("%s", secret.Data["url"]), nil
}

func (c *VaultClient) IdentityListEntities(ctx context.Context) ([]string, error) {
	path := "identity/entity/name"

	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, err
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

var ErrVaultEmptyDataReturned = errors.New("no data returned from Vault")

func (c *VaultClient) IdentityGetEntityIdByName(ctx context.Context, entityName string) (string, error) {
	entity, err := c.IdentityGetEntityByName(ctx, entityName)
	if err != nil {
		return "", err
	}

	if entity == nil {
		return "", ErrVaultEmptyDataReturned
	}

	raw, ok := entity["id"]
	if !ok {
		return "", errors.New("no id contained in response")
	}

	return raw.(string), nil
}

func (c *VaultClient) IdentityGetEntityByName(ctx context.Context, entityName string) (map[string]any, error) {
	path := fmt.Sprintf("identity/entity/name/%s", entityName)

	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, ErrVaultEmptyDataReturned
	}

	return secret.Data, nil
}

func (c *VaultClient) IdentityListGroups(ctx context.Context) ([]string, error) {
	path := "identity/group/name"

	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, err
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

func (c *VaultClient) SshSignKey(ctx context.Context, mount string, req SshSignatureRequest) (string, error) {
	data := convertUserKeyRequest(req)

	path := fmt.Sprintf("%s/sign/%s", mount, req.VaultRole)
	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		// TODO
		//var respErr *api.ResponseError
		//if errors.As(err, &respErr) && !shouldRetry(respErr.StatusCode) {
		//	return "", backoff.Permanent(err)
		//}
		return "", err
	}

	return fmt.Sprintf("%s", secret.Data["signed_key"]), nil
}

func convertUserKeyRequest(req SshSignatureRequest) map[string]any {
	data := map[string]interface{}{
		"public_key": req.PublicKey,
		"cert_type":  "user",
	}

	if len(req.Ttl) > 0 {
		data["ttl"] = req.Ttl
	}

	if len(req.Principals) > 0 {
		data["valid_principals"] = strings.Join(req.Principals, ",")
	}

	return data
}

func (c *VaultClient) PkiIssueCert(ctx context.Context, mount string, args PkiIssueArgs) (*vault.Secret, error) {
	path := fmt.Sprintf("%s/issue/%s", mount, args.Role)
	data := buildIssueRequestArgs(args)

	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		var respErr *vault.ResponseError
		if errors.As(err, &respErr) && !shouldRetry(respErr.StatusCode) {
			return nil, err
		}

		return nil, fmt.Errorf("could not issue certificate: %w", err)
	}

	return secret, nil
}

func (c *VaultClient) PkiVerifyCert(ctx context.Context, mount string, cert *x509.Certificate) error {
	caData, err := c.PkiGetCaChain(ctx, mount)
	if err != nil {
		return err
	}

	caBlock, _ := pem.Decode(caData)
	ca, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return err
	}

	return backend.VerifyCertAgainstCa(cert, ca)
}

func (c *VaultClient) PkiGetCa(ctx context.Context, mount string, binary bool) ([]byte, error) {
	path := fmt.Sprintf("%s/ca", mount)
	if !binary {
		path = path + "/pem"
	}

	return c.readRaw(ctx, path)
}

func (c *VaultClient) PkiGetCaChain(ctx context.Context, mount string) ([]byte, error) {
	path := fmt.Sprintf("/%s/ca_chain", mount)
	return c.readRaw(ctx, path)
}

func (c *VaultClient) PkiGetCrl(ctx context.Context, mount string, binary bool) ([]byte, error) {
	path := fmt.Sprintf("%s/crl", mount)
	if !binary {
		path += "/pem"
	}

	return c.readRaw(ctx, path)
}

func (c *VaultClient) readRaw(ctx context.Context, path string) ([]byte, error) {
	secret, err := c.client.Logical().ReadRawWithContext(ctx, path)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(secret.Body)
}

func (c *VaultClient) SshGetCa(ctx context.Context, mount string) ([]byte, error) {
	path := fmt.Sprintf("%s/public_key", mount)
	resp, err := c.client.Logical().ReadRawWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading cert failed: %v", err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body from response: %v", err)
	}

	return data, nil
}

func buildIssueRequestArgs(args PkiIssueArgs) map[string]any {
	data := map[string]any{
		"common_name": args.CommonName,
		"ttl":         args.Ttl,
		"format":      "pem",
		"ip_sans":     strings.Join(args.IpSans, ","),
		"alt_names":   strings.Join(args.AltNames, ","),
	}

	return data
}

func shouldRetry(statusCode int) bool {
	switch statusCode {
	case 400, // Bad Request
		401, // Unauthorized
		403, // Forbidden
		404, // Not Found
		405, // Method Not Allowed
		406, // Not Acceptable
		407, // Proxy Authentication Required
		409, // Conflict
		410, // Gone
		411, // Length Required
		412, // Precondition Failed
		413, // Payload Too Large
		414, // URI Too Long
		415, // Unsupported Media Type
		416, // Range Not Satisfiable
		417, // Expectation Failed
		418, // I'm a Teapot
		421, // Misdirected Request
		422, // Unprocessable Entity
		423, // Locked (WebDAV)
		424, // Failed Dependency (WebDAV)
		425, // Too Early
		426, // Upgrade Required
		428, // Precondition Required
		429, // Too Many Requests
		431, // Request Header Fields Too Large
		451: // Unavailable For Legal Reasons
		return false
	default:
		return true
	}
}
