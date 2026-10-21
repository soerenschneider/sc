package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *VaultClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	config := vault.DefaultConfig()
	config.Address = srv.URL
	config.MaxRetries = 0
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	return &VaultClient{client: client}
}

func TestPatch(t *testing.T) {
	var body map[string]map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/secret/data/my/app" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/merge-patch+json" {
			t.Errorf("unexpected content type %q", ct)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"data":{"version":4,"created_time":"2026-10-21T10:00:00.123Z"}}`))
	})

	sec, err := c.Patch(context.Background(), "secret", "my/app", map[string]any{"pass": "new", "legacy": nil})
	if err != nil {
		t.Fatal(err)
	}
	if sec.Version != 4 || sec.UpdatedAt.IsZero() {
		t.Errorf("unexpected secret %+v", sec)
	}
	if v, ok := body["data"]["legacy"]; !ok || v != nil {
		t.Errorf("expected deleted key to be sent as null, got %v", body)
	}
	if body["data"]["pass"] != "new" {
		t.Errorf("unexpected body %v", body)
	}
}

func TestPatchErrors(t *testing.T) {
	tests := map[int]ErrorKind{
		http.StatusNotFound:         ErrKindNotFound,
		http.StatusForbidden:        ErrKindPermissionDenied,
		http.StatusMethodNotAllowed: ErrKindUnsupported,
	}
	for status, kind := range tests {
		t.Run(kind.String(), func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"errors":[]}`))
			})
			if _, err := c.Patch(context.Background(), "secret", "app", map[string]any{"a": "b"}); !IsKind(err, kind) {
				t.Errorf("expected error of kind %s, got %v", kind, err)
			}
		})
	}
}
