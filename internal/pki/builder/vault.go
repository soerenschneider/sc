package builder

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	auth "github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/vault-pki-cli/pkg/pki"
	"github.com/soerenschneider/vault-pki-cli/pkg/renew_strategy"
	"github.com/soerenschneider/vault-pki-cli/pkg/vault"
)

func BuildPki(address string, sshMount, role string) (*pki.PkiService, error) {
	vaultPkiClient, err := newVaultPkiClient(address, sshMount, role)
	if err != nil {
		return nil, fmt.Errorf("could not build client for signatures: %w", err)
	}

	issueStrategy, err := renew_strategy.NewPercentage(50)
	if err != nil {
		return nil, fmt.Errorf("could not build issue strategy implementation: %w", err)
	}

	return pki.NewPkiService(vaultPkiClient, issueStrategy)
}

func newVaultPkiClient(address string, pkiMount, role string) (*vault.VaultPki, error) {
	conf := api.DefaultConfig()

	conf.Address = address
	conf.MaxRetries = 5

	vaultClient, err := api.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("could not build vault client: %w", err)
	}

	auth := auth.NewTokenImplicitAuth()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = vaultClient.Auth().Login(ctx, auth)
	if err != nil {
		return nil, err
	}

	var opts []vault.VaultOpts
	opts = append(opts, vault.WithPkiMount(pkiMount))

	return vault.NewVaultPki(vaultClient.Logical(), role, opts...)
}
