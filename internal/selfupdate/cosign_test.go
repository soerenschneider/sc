package selfupdate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sigstore/sigstore-go/pkg/testing/ca"
)

// The cosign-key fixture is a bundle signed with a throwaway key pair instead
// of keyless, created with cosign v3.1.3:
//
//	cosign generate-key-pair
//	cosign sign-blob --key cosign.key --bundle checksum.sha256.sigstore.json checksum.sha256
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCosignKeyless(t *testing.T) {
	sigstore, err := ca.NewVirtualSigstore()
	if err != nil {
		t.Fatal(err)
	}
	const (
		subject = "https://github.com/soerenschneider/sc/.github/workflows/release-container.yaml@refs/heads/main"
		branch  = "https://github.com/soerenschneider/sc/.github/workflows/release-container.yaml@refs/heads/feature"
	)
	msg := []byte("# version: v1.2.3\nabc  builds/sc-linux-amd64\n")
	v := DefaultVerifiers()[len(DefaultVerifiers())-1].(CosignKeyless)
	// The virtual sigstore's certificates carry no SCTs.
	v.withoutSCTs = true

	signed, err := sigstore.Sign(subject, githubActionsIssuer, msg)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := v.verifyEntity(signed, msg, sigstore)
	if err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	if signer.Identity != subject || signer.Details["signed_at"] == "" {
		t.Errorf("signer = %+v", signer)
	}
	if _, err := v.verifyEntity(signed, append(msg, 'x'), sigstore); err == nil {
		t.Error("tampered message accepted")
	}

	fromBranch, err := sigstore.Sign(branch, githubActionsIssuer, msg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.verifyEntity(fromBranch, msg, sigstore); err == nil {
		t.Error("signature from branch accepted")
	}

	otherIssuer, err := sigstore.Sign(subject, "https://accounts.google.com", msg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.verifyEntity(otherIssuer, msg, sigstore); err == nil {
		t.Error("signature from other issuer accepted")
	}

	untrusted, err := ca.NewVirtualSigstore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.verifyEntity(signed, msg, untrusted); err == nil {
		t.Error("signature from untrusted sigstore instance accepted")
	}

	keySigned := readFixture(t, "cosign-key/checksum.sha256.sigstore.json")
	if _, err := v.Verify(context.Background(), readFixture(t, "cosign-key/checksum.sha256"), keySigned); err == nil || !strings.Contains(err.Error(), "signed with a key") {
		t.Errorf("key signed bundle: got %v", err)
	}
	if _, err := v.Verify(context.Background(), msg, []byte("garbage")); err == nil {
		t.Error("accepted malformed bundle")
	}
	v.withoutSCTs = false
	if _, err := v.verifyEntity(signed, msg, sigstore); err == nil {
		t.Error("accepted certificate without SCT")
	}
	if _, err := (CosignKeyless{withoutSCTs: true}).verifyEntity(signed, msg, sigstore); err == nil {
		t.Error("accepted signature without trusted identity")
	}
}
