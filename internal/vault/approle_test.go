package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

func TestRotateAppRoleSecretID(t *testing.T) {
	var created map[string]any
	var destroyed []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/auth/approle/role/my-role/secret-id" && r.URL.Query().Get("list") == "true":
			_, _ = w.Write([]byte(`{"data":{"keys":["old-1","old-2"]}}`))
		case r.URL.Path == "/v1/auth/approle/role/my-role/secret-id" && r.Method == http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&created)
			_, _ = w.Write([]byte(`{"data":{"secret_id":"new-secret","secret_id_accessor":"new-1"}}`))
		case r.URL.Path == "/v1/auth/approle/role/my-role/secret-id-accessor/destroy":
			if created == nil {
				t.Error("old secret_id destroyed before the new one has been created")
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			destroyed = append(destroyed, body["secret_id_accessor"])
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	config := vault.DefaultConfig()
	config.Address = srv.URL
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	c := &VaultClient{client: client}

	rotation, err := c.RotateAppRoleSecretID(context.Background(), "approle", "my-role", nil, []string{"10.0.0.0/8"}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if rotation.SecretID != "new-secret" || rotation.SecretIDAccessor != "new-1" || rotation.WrapInfo != nil {
		t.Errorf("unexpected rotation %+v", rotation)
	}
	if want := []string{"old-1", "old-2"}; !slices.Equal(destroyed, want) || !slices.Equal(rotation.Destroyed, want) {
		t.Errorf("expected %v to be destroyed, got %v (reported %v)", want, destroyed, rotation.Destroyed)
	}
	for _, key := range []string{"cidr_list", "token_bound_cidrs"} {
		if cidrs, ok := created[key].([]any); !ok || len(cidrs) != 1 || cidrs[0] != "10.0.0.0/8" {
			t.Errorf("expected %s to be set, got %v", key, created[key])
		}
	}
}

func TestRotateAppRoleSecretIDWithoutExisting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":[]}`))
		case http.MethodPut:
			_, _ = w.Write([]byte(`{"data":{"secret_id":"new-secret","secret_id_accessor":"new-1"}}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	config := vault.DefaultConfig()
	config.Address = srv.URL
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	c := &VaultClient{client: client}

	rotation, err := c.RotateAppRoleSecretID(context.Background(), "approle", "my-role", nil, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if rotation.SecretID != "new-secret" || len(rotation.Destroyed) != 0 {
		t.Errorf("unexpected result %+v", rotation)
	}
}

func TestRotateAppRoleSecretIDWrapped(t *testing.T) {
	var destroyed []string
	var listWrapped bool
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		wrapTtl := r.Header.Get("X-Vault-Wrap-TTL")
		switch {
		case r.URL.Path == "/v1/auth/approle/role/my-role/secret-id" && r.URL.Query().Get("list") == "true":
			listWrapped = wrapTtl != ""
			_, _ = w.Write([]byte(`{"data":{"keys":["old-1"]}}`))
		case r.URL.Path == "/v1/auth/approle/role/my-role/secret-id" && r.Method == http.MethodPut:
			if wrapTtl != "5m" {
				t.Errorf("expected wrap ttl 5m, got %q", wrapTtl)
			}
			if token := r.Header.Get("X-Vault-Token"); token != "my-token" {
				t.Errorf("expected token to be passed on, got %q", token)
			}
			_, _ = w.Write([]byte(`{"wrap_info":{"token":"wrapping-token","accessor":"wrapping-accessor","ttl":300,"creation_path":"auth/approle/role/my-role/secret-id"}}`))
		case r.URL.Path == "/v1/auth/approle/role/my-role/secret-id-accessor/destroy":
			if wrapTtl != "" {
				t.Error("destroy request must not be wrapped")
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			destroyed = append(destroyed, body["secret_id_accessor"])
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	c.client.SetToken("my-token")

	rotation, err := c.RotateAppRoleSecretID(context.Background(), "approle", "my-role", nil, nil, "", "5m")
	if err != nil {
		t.Fatal(err)
	}
	if listWrapped {
		t.Error("list request must not be wrapped")
	}
	if rotation.SecretID != "" || rotation.WrapInfo == nil || rotation.WrapInfo.Token != "wrapping-token" || rotation.WrapInfo.TTL != 300 {
		t.Errorf("unexpected rotation %+v", rotation)
	}
	if want := []string{"old-1"}; !slices.Equal(destroyed, want) || !slices.Equal(rotation.Destroyed, want) {
		t.Errorf("expected %v to be destroyed, got %v (reported %v)", want, destroyed, rotation.Destroyed)
	}
}

func TestUnwrap(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/wrapping/unwrap" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if token := r.Header.Get("X-Vault-Token"); token != "wrapping-token" {
			t.Errorf("expected request to be authenticated using the wrapping token, got %q", token)
		}
		_, _ = w.Write([]byte(`{"data":{"secret_id":"new-secret"}}`))
	})
	c.client.SetToken("my-token")

	secret, err := c.Unwrap(context.Background(), "wrapping-token")
	if err != nil {
		t.Fatal(err)
	}
	if secret.Data["secret_id"] != "new-secret" {
		t.Errorf("unexpected data %v", secret.Data)
	}
	if c.client.Token() != "my-token" {
		t.Errorf("token of the client has been replaced")
	}
}
