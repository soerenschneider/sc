package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/soerenschneider/sc/internal/vault"
)

func TestFormatApproleTtl(t *testing.T) {
	tests := []struct {
		seconds int64
		want    string
	}{
		{0, "fallback"},
		{30, "30s"},
		{3600, "1h"},
		{5400, "1h30m"},
		{36 * 3600, "1d12h"},
		{90061, "1d1h1m1s"},
	}

	for _, tt := range tests {
		if got := formatApproleTtl(tt.seconds, "fallback"); got != tt.want {
			t.Errorf("formatApproleTtl(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestPrintApprolesJson(t *testing.T) {
	var buf bytes.Buffer
	if err := printApprolesJson(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "[]\n" {
		t.Errorf("expected empty array, got %q", got)
	}

	buf.Reset()
	roles := []approleInfo{toApproleInfo(&vault.AppRole{Name: "my-role", TokenTtl: time.Hour})}
	if err := printApprolesJson(&buf, roles); err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded[0]["token_ttl_seconds"] != float64(3600) {
		t.Errorf("expected token_ttl_seconds 3600, got %v", decoded[0]["token_ttl_seconds"])
	}
	if policies, ok := decoded[0]["token_policies"].([]any); !ok || len(policies) != 0 {
		t.Errorf("expected empty token_policies array, got %v", decoded[0]["token_policies"])
	}
}

func TestPrintApprolesSimple(t *testing.T) {
	roles := []approleInfo{
		toApproleInfo(&vault.AppRole{Name: "defaults"}),
		toApproleInfo(&vault.AppRole{
			Name:               "my-role",
			TokenPolicies:      []string{"a", "b"},
			TokenTtl:           time.Hour,
			TokenMaxTtl:        4 * time.Hour,
			SecretIDTtl:        30 * 24 * time.Hour,
			SecretIDNumUses:    3,
			SecretIDBoundCidrs: []string{"10.0.0.0/8"},
		}),
	}

	var buf bytes.Buffer
	if err := printApprolesSimple(&buf, roles); err != nil {
		t.Fatal(err)
	}

	want := `ROLE      POLICIES  TOKEN_TTL  TOKEN_MAX_TTL  SECRET_ID_TTL  SECRET_ID_USES  SECRET_ID_CIDRS
defaults  -         default    default        never          unlimited       -
my-role   a,b       1h         4h             30d            3               10.0.0.0/8
`
	if got := buf.String(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
