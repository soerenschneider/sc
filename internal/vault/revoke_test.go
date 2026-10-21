package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

func TestRevokeLeaseUsesPath(t *testing.T) {
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	config := vault.DefaultConfig()
	config.Address = srv.URL
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	c := &VaultClient{client: client}
	if err := c.RevokeLease(context.Background(), "aws/prod-default/creds/developer/abc123"); err != nil {
		t.Fatal(err)
	}

	if want := "/v1/sys/leases/revoke/aws/prod-default/creds/developer/abc123"; method != http.MethodPut || path != want {
		t.Errorf("expected PUT %s, got %s %s", want, method, path)
	}
}
