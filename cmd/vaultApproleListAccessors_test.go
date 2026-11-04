package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/soerenschneider/sc/internal/vault"
)

func TestToApproleSecretID(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		expiration time.Time
		want       string
	}{
		{"valid", now.Add(time.Hour), approleStatusValid},
		{"expired", now.Add(-time.Hour), approleStatusExpired},
		{"non-expiring", time.Time{}, approleStatusNonExpiring},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toApproleSecretID(&vault.AppRoleSecretID{Accessor: "a", ExpirationTime: tt.expiration}, now)
			if got.Status != tt.want {
				t.Errorf("status = %q, want %q", got.Status, tt.want)
			}
			if got.CidrList == nil || got.TokenBoundCidrs == nil || got.Metadata == nil {
				t.Errorf("expected empty, non-nil collections, got %+v", got)
			}
		})
	}
}

func TestPrintApproleSecretIDsJson(t *testing.T) {
	var buf bytes.Buffer
	if err := printApproleSecretIDsJson(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(buf.String()); got != "[]" {
		t.Errorf("expected empty array, got %q", got)
	}

	buf.Reset()
	roles := []approleSecretIDs{{Role: "empty", SecretIDs: []approleSecretID{}}}
	if err := printApproleSecretIDsJson(&buf, roles); err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if ids, ok := decoded[0]["secret_ids"].([]any); !ok || len(ids) != 0 {
		t.Errorf("expected empty secret_ids array, got %v", decoded[0]["secret_ids"])
	}
}

func TestPrintApproleSecretIDsSimple(t *testing.T) {
	expiration := time.Date(2026, 10, 21, 12, 0, 0, 0, time.Local)
	roles := []approleSecretIDs{
		{Role: "empty", SecretIDs: []approleSecretID{}},
		{Role: "my-role", SecretIDs: []approleSecretID{{
			Accessor:       "acc",
			Status:         approleStatusValid,
			ExpirationTime: &expiration,
			NumUses:        3,
			CidrList:       []string{"10.0.0.0/8"},
			Metadata:       map[string]string{"b": "2", "a": "1"},
		}}},
	}

	var buf bytes.Buffer
	if err := printApproleSecretIDsSimple(&buf, roles); err != nil {
		t.Fatal(err)
	}

	want := `ROLE     ACCESSOR  STATUS  CREATED  EXPIRES              USES  CIDRS       METADATA
empty    -         -       -        -                    -     -           -
my-role  acc       valid   -        2026-10-21 12:00:00  3     10.0.0.0/8  a=1,b=2
`
	if got := buf.String(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
