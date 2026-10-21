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

	secretID, accessor, gotDestroyed, err := c.RotateAppRoleSecretID(context.Background(), "approle", "my-role", nil, []string{"10.0.0.0/8"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if secretID != "new-secret" || accessor != "new-1" {
		t.Errorf("unexpected secret_id %q / accessor %q", secretID, accessor)
	}
	if want := []string{"old-1", "old-2"}; !slices.Equal(destroyed, want) || !slices.Equal(gotDestroyed, want) {
		t.Errorf("expected %v to be destroyed, got %v (reported %v)", want, destroyed, gotDestroyed)
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

	secretID, _, destroyed, err := c.RotateAppRoleSecretID(context.Background(), "approle", "my-role", nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if secretID != "new-secret" || len(destroyed) != 0 {
		t.Errorf("unexpected result %q, %v", secretID, destroyed)
	}
}
