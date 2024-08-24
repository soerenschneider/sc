package builder

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/vault-ssh-cli/pkg/signature"
)

func BuildSshSigner(address string, sshMount string) (*signature.SignatureService, error) {
	signatureClient, err := newSignatureClient(address, sshMount)
	if err != nil {
		return nil, fmt.Errorf("could not build client for signatures: %w", err)
	}

	issueStrategy, err := signature.NewPercentageStrategy(50)
	if err != nil {
		return nil, fmt.Errorf("could not build issue strategy implementation: %w", err)
	}

	return signature.NewSignatureService(signatureClient, issueStrategy)
}

func newSignatureClient(address string, sshMount string) (*signature.SignatureClient, error) {
	conf := api.DefaultConfig()

	conf.Address = address
	conf.MaxRetries = 5

	vaultClient, err := api.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("could not build vault client: %w", err)
	}

	auth := vault.NewTokenImplicitAuth()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = vaultClient.Auth().Login(ctx, auth)
	if err != nil {
		return nil, err
	}

	var opts []signature.VaultOpts
	opts = append(opts, signature.WithSshMountPath(sshMount))

	return signature.NewVaultSigner(vaultClient.Logical(), opts...)
}
