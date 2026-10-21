package cmd

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/soerenschneider/sc/internal/vault"
)

type fakeKv2Store struct {
	vaultKv2Provider
	secret   *vault.Secret
	metadata bool
	puts     int
	putCas   int

	patchErr error
	patched  map[string]any
}

func (f *fakeKv2Store) Patch(_ context.Context, _, _ string, data map[string]any) (*vault.Secret, error) {
	if f.patchErr != nil {
		return nil, f.patchErr
	}
	f.patched = data
	return &vault.Secret{Version: 7}, nil
}

func (f *fakeKv2Store) Get(_ context.Context, _, _ string) (*vault.Secret, error) {
	if f.secret == nil {
		return nil, &vault.VaultError{Kind: vault.ErrKindNotFound}
	}
	return &vault.Secret{Data: maps.Clone(f.secret.Data), Version: f.secret.Version}, nil
}

func (f *fakeKv2Store) GetMetadata(_ context.Context, _, _ string) (*vault.SecretMetadata, error) {
	if !f.metadata {
		return nil, &vault.VaultError{Kind: vault.ErrKindNotFound}
	}
	return &vault.SecretMetadata{}, nil
}

func (f *fakeKv2Store) Put(_ context.Context, _, _ string, data map[string]string, cas int) (*vault.Secret, error) {
	f.puts++
	f.putCas = cas
	f.secret = &vault.Secret{Data: data, Version: cas + 1}
	return f.secret, nil
}

func TestPatchSecretNative(t *testing.T) {
	store := &fakeKv2Store{}
	result, err := patchSecret(context.Background(), store, "secret", "app", map[string]string{"pass": "new"}, []string{"legacy"})
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]any{"pass": "new", "legacy": nil}; !maps.Equal(store.patched, want) {
		t.Errorf("expected patch %v, got %v", want, store.patched)
	}
	if store.puts != 0 {
		t.Errorf("expected no fallback write, got %d", store.puts)
	}
	if result.secret.Version != 7 || !slices.Equal(result.changed, []string{"legacy", "pass"}) {
		t.Errorf("unexpected result %+v", result)
	}
}

func TestPatchSecretFallback(t *testing.T) {
	for _, kind := range []vault.ErrorKind{vault.ErrKindNotFound, vault.ErrKindPermissionDenied, vault.ErrKindUnsupported} {
		t.Run(kind.String(), func(t *testing.T) {
			store := &fakeKv2Store{
				patchErr: &vault.VaultError{Kind: kind},
				secret:   &vault.Secret{Data: map[string]string{"user": "admin"}, Version: 2},
			}
			if _, err := patchSecret(context.Background(), store, "secret", "app", map[string]string{"pass": "new"}, nil); err != nil {
				t.Fatal(err)
			}
			if store.puts != 1 || store.putCas != 2 {
				t.Errorf("expected a fallback write with cas 2, got %d writes with cas %d", store.puts, store.putCas)
			}
			if want := map[string]string{"user": "admin", "pass": "new"}; !maps.Equal(store.secret.Data, want) {
				t.Errorf("expected %v, got %v", want, store.secret.Data)
			}
		})
	}

	store := &fakeKv2Store{patchErr: &vault.VaultError{Kind: vault.ErrKindServerError}}
	if _, err := patchSecret(context.Background(), store, "secret", "app", map[string]string{"pass": "new"}, nil); err == nil || store.puts != 0 {
		t.Errorf("expected server errors to fail without fallback, got err %v and %d writes", err, store.puts)
	}
}

func TestPatchSecretReadWrite(t *testing.T) {
	store := &fakeKv2Store{secret: &vault.Secret{Data: map[string]string{"user": "admin", "pass": "old", "legacy": "x"}, Version: 3}}

	result, err := patchSecretReadWrite(context.Background(), store, "secret", "app", map[string]string{"pass": "new", "user": "admin", "port": "5432"}, []string{"legacy", "missing"})
	if err != nil {
		t.Fatal(err)
	}
	if store.puts != 1 || store.putCas != 3 {
		t.Errorf("expected one write with cas 3, got %d writes with cas %d", store.puts, store.putCas)
	}
	if want := map[string]string{"user": "admin", "pass": "new", "port": "5432"}; !maps.Equal(store.secret.Data, want) {
		t.Errorf("expected %v, got %v", want, store.secret.Data)
	}
	if want := []string{"pass", "port", "legacy"}; !slices.Equal(result.changed, want) {
		t.Errorf("expected changed keys %v, got %v", want, result.changed)
	}
	if want := []string{"missing"}; !slices.Equal(result.missing, want) {
		t.Errorf("expected missing keys %v, got %v", want, result.missing)
	}

	// Writing the same values again must not create a new version.
	result, err = patchSecretReadWrite(context.Background(), store, "secret", "app", map[string]string{"pass": "new"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.secret != nil || store.puts != 1 {
		t.Errorf("expected no write for unchanged values, got %d writes", store.puts)
	}
}

func TestPatchSecretNew(t *testing.T) {
	store := &fakeKv2Store{patchErr: &vault.VaultError{Kind: vault.ErrKindNotFound}}
	if _, err := patchSecret(context.Background(), store, "secret", "app", map[string]string{"user": "admin"}, nil); err != nil {
		t.Fatal(err)
	}
	if store.puts != 1 || store.putCas != 0 {
		t.Errorf("expected the secret to be created with cas 0, got %d writes with cas %d", store.puts, store.putCas)
	}
}

func TestPatchSecretRefusesDeletedLatestVersion(t *testing.T) {
	store := &fakeKv2Store{metadata: true, patchErr: &vault.VaultError{Kind: vault.ErrKindNotFound}}
	if _, err := patchSecret(context.Background(), store, "secret", "app", map[string]string{"user": "admin"}, nil); err == nil {
		t.Error("expected an error when the latest version is deleted")
	}
	if store.puts != 0 {
		t.Error("expected no write")
	}
}

func TestPatchSecretSetAndDelete(t *testing.T) {
	if _, err := patchSecret(context.Background(), &fakeKv2Store{}, "secret", "app", map[string]string{"a": "1"}, []string{"a"}); err == nil {
		t.Error("expected an error when a key is set and deleted")
	}
}

func TestParseSecretPatchPairs(t *testing.T) {
	got, err := parseSecretPatchPairs([]string{"user=admin", "url=https://x?a=b", "empty=", "pass=-"}, strings.NewReader("line1\nline2\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"user": "admin", "url": "https://x?a=b", "empty": "", "pass": "line1\nline2"}
	if !maps.Equal(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}

	for _, args := range [][]string{
		{"novalue"},
		{"=value"},
		{"a=1", "a=2"},
		{"a=-", "b=-"},
	} {
		if _, err := parseSecretPatchPairs(args, strings.NewReader("x")); err == nil {
			t.Errorf("expected an error for %v", args)
		}
	}
}
