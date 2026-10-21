package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/soerenschneider/sc/internal/vault"
)

func TestAwsCredentialsFileRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	credsPath := filepath.Join(home, ".aws", "credentials")
	if err := os.MkdirAll(filepath.Dir(credsPath), 0700); err != nil {
		t.Fatal(err)
	}
	existing := "[default]\nregion = eu-central-1\naws_session_token = stale\n\n[other]\naws_access_key_id = OTHER\n"
	if err := os.WriteFile(credsPath, []byte(existing), 0600); err != nil {
		t.Fatal(err)
	}

	expiration := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	creds := vault.AwsCredentials{
		AccessKeyID:     "AKIA",
		SecretAccessKey: "secret",
		LeaseID:         "aws/creds/dev/abc",
		Expiration:      expiration,
	}
	if err := updateAwsCredentialsFile("default", creds); err != nil {
		t.Fatal(err)
	}

	profiles, err := readAwsCredentialsProfiles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected only the profile issued by Vault, got %v", profiles)
	}
	got := profiles[0]
	if got.Name != "default" || got.AccessKeyID != "AKIA" || got.LeaseID != creds.LeaseID || got.HasSession {
		t.Errorf("unexpected profile %+v", got)
	}
	if !got.ExpiresKnown || !got.Expiration.Equal(expiration) {
		t.Errorf("expected expiration %v, got %v", expiration, got.Expiration)
	}

	if err := removeAwsCredentials("default"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(credsPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "AKIA\n") || strings.Contains(content, "x_vault_lease_id") {
		t.Errorf("credentials have not been removed:\n%s", content)
	}
	if !strings.Contains(content, "region") || !strings.Contains(content, "OTHER") {
		t.Errorf("unrelated settings have been removed:\n%s", content)
	}

	if err := removeAwsCredentials("other"); err != nil {
		t.Fatal(err)
	}
	if _, err := readAwsCredentialsProfiles("other"); err == nil {
		t.Error("expected empty profile to be removed")
	}
}
