package selfupdate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		local, remote string
		want          bool
		wantErr       bool
	}{
		{"v1.10.0", "v1.11.0", true, false},
		{"v1.11.0", "v1.11.0", false, false},
		{"v1.12.0", "v1.11.0", false, false},
		{"v1.9.0", "v1.10.0", true, false},
		{"1.9.0", "v1.10.0", true, false},
		{"v1.11.0-rc1", "v1.11.0", true, false},
		{"dev", "v1.11.0", false, true},
		{"v1.0.0", "", false, true},
	}
	for _, tt := range tests {
		got, err := IsNewer(tt.local, tt.remote)
		if (err != nil) != tt.wantErr {
			t.Errorf("IsNewer(%q, %q) error = %v, wantErr %v", tt.local, tt.remote, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.local, tt.remote, got, tt.want)
		}
	}
}

func TestParseChecksum(t *testing.T) {
	sums := `78ce0f193213f5b69c64a9a437ccf7945f07c7e4b8da4296e305ec8741f2ff46  builds/sc-linux-aarch64
404b1ae1d5829fc62981eecf76e1fe6cd819df04dc2d940a54f78c19afed42fc  builds/sc-linux-amd64
5a58c9593040f5ad283274c2bb7fea06e57c024664ed94f92b8e987d9bc5bcf8 *sc-linux-armv6
`
	got, err := parseChecksum(strings.NewReader(sums), "sc-linux-amd64")
	if err != nil || got != "404b1ae1d5829fc62981eecf76e1fe6cd819df04dc2d940a54f78c19afed42fc" {
		t.Errorf("parseChecksum() = %q, %v", got, err)
	}
	got, err = parseChecksum(strings.NewReader(sums), "sc-linux-armv6")
	if err != nil || got != "5a58c9593040f5ad283274c2bb7fea06e57c024664ed94f92b8e987d9bc5bcf8" {
		t.Errorf("parseChecksum() = %q, %v", got, err)
	}
	if _, err := parseChecksum(strings.NewReader(sums), "sc-darwin-arm64"); err == nil {
		t.Error("expected error for missing asset")
	}
}

func TestAssetName(t *testing.T) {
	tests := map[string]string{
		"linux/amd64":  "sc-linux-amd64",
		"linux/arm64":  "sc-linux-aarch64",
		"linux/arm":    "sc-linux-armv6",
		"darwin/arm64": "sc-darwin-arm64",
	}
	for platform, want := range tests {
		goos, goarch, _ := strings.Cut(platform, "/")
		got, err := AssetName(goos, goarch)
		if err != nil || got != want {
			t.Errorf("AssetName(%s) = %q, %v, want %q", platform, got, err, want)
		}
	}
	if _, err := AssetName("darwin", "amd64"); err == nil {
		t.Error("expected error for darwin/amd64")
	}
}

func TestParseVersion(t *testing.T) {
	got, err := parseVersion(strings.NewReader("# version: v1.12.0\nabc  builds/sc-linux-amd64\n"))
	if err != nil || got != "v1.12.0" {
		t.Errorf("parseVersion() = %q, %v", got, err)
	}
	if _, err := parseVersion(strings.NewReader("abc  builds/sc-linux-amd64\n")); err == nil {
		t.Error("expected error for missing version")
	}
}

type fakeClient map[string][]byte

func (c fakeClient) Do(r *http.Request) (*http.Response, error) {
	body, ok := c[r.URL.String()]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not Found", Body: io.NopCloser(bytes.NewReader(nil))}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(body))}, nil
}

func TestUpdate(t *testing.T) {
	name, err := AssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skip(err)
	}

	verifiers := []Verifier{hashVerifier{}}

	newBinary := []byte("new binary")
	sum := sha256.Sum256(newBinary)
	release := func(signedVersion, tag string) (*Release, fakeClient) {
		sums := []byte(fmt.Sprintf("%s%s\n%s  builds/%s\n", versionPrefix, signedVersion, hex.EncodeToString(sum[:]), name))
		client := fakeClient{
			"https://example.com/bin": newBinary,
			"https://example.com/sum": sums,
			"https://example.com/sig": hashVerifier{}.sign(sums),
		}
		return &Release{TagName: tag, Assets: []Asset{
			{Name: name, BrowserDownloadURL: "https://example.com/bin"},
			{Name: checksumAsset, BrowserDownloadURL: "https://example.com/sum"},
			{Name: hashVerifier{}.SignatureAsset(), BrowserDownloadURL: "https://example.com/sig"},
		}}, client
	}

	setup := func(t *testing.T) string {
		exe := filepath.Join(t.TempDir(), "sc")
		if err := os.WriteFile(exe, []byte("old binary"), 0o600); err != nil {
			t.Fatal(err)
		}
		return exe
	}

	t.Run("valid", func(t *testing.T) {
		exe := setup(t)
		r, client := release("v1.12.0", "v1.12.0")
		if err := Update(context.Background(), client, r, exe, verifiers); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.ReadFile(exe); !bytes.Equal(got, newBinary) {
			t.Errorf("binary not replaced, got %q", got)
		}
	})

	t.Run("replayed under newer tag", func(t *testing.T) {
		exe := setup(t)
		r, client := release("v1.11.0", "v1.12.0")
		if err := Update(context.Background(), client, r, exe, verifiers); err == nil || !strings.Contains(err.Error(), "not v1.12.0") {
			t.Fatalf("expected version mismatch, got %v", err)
		}
	})

	t.Run("tampered checksums", func(t *testing.T) {
		exe := setup(t)
		r, client := release("v1.12.0", "v1.12.0")
		client["https://example.com/sum"] = append(client["https://example.com/sum"], '\n')
		if err := Update(context.Background(), client, r, exe, verifiers); err == nil || !strings.Contains(err.Error(), "could not verify") {
			t.Fatalf("expected verification failure, got %v", err)
		}
		if got, _ := os.ReadFile(exe); string(got) != "old binary" {
			t.Errorf("binary was replaced despite failed verification")
		}
	})

	t.Run("missing signature", func(t *testing.T) {
		exe := setup(t)
		r, client := release("v1.12.0", "v1.12.0")
		r.Assets = r.Assets[:2]
		if err := Update(context.Background(), client, r, exe, verifiers); err == nil || !strings.Contains(err.Error(), "no signature") {
			t.Fatalf("expected error for missing signature, got %v", err)
		}
	})

	t.Run("no verifiers", func(t *testing.T) {
		exe := setup(t)
		r, client := release("v1.12.0", "v1.12.0")
		if err := Update(context.Background(), client, r, exe, nil); err == nil {
			t.Fatal("expected error without verifiers")
		}
	})

	t.Run("falls back to next verifier", func(t *testing.T) {
		exe := setup(t)
		r, client := release("v1.12.0", "v1.12.0")
		failing := fakeVerifier{asset: "checksum.sha256.other", err: errors.New("bad signature")}
		r.Assets = append(r.Assets, Asset{Name: failing.asset, BrowserDownloadURL: "https://example.com/other"})
		client["https://example.com/other"] = []byte("sig")
		skipped := fakeVerifier{asset: "checksum.sha256.missing"}
		if err := Update(context.Background(), client, r, exe, []Verifier{skipped, failing, verifiers[0]}); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.ReadFile(exe); !bytes.Equal(got, newBinary) {
			t.Errorf("binary not replaced, got %q", got)
		}
	})

	t.Run("all verifiers fail", func(t *testing.T) {
		exe := setup(t)
		r, client := release("v1.12.0", "v1.12.0")
		failing := fakeVerifier{asset: hashVerifier{}.SignatureAsset(), err: errors.New("bad signature")}
		err := Update(context.Background(), client, r, exe, []Verifier{failing})
		if err == nil || !strings.Contains(err.Error(), "fake: bad signature") {
			t.Fatalf("expected verification failure, got %v", err)
		}
	})
}

// hashVerifier accepts the sha256 of the message as its signature, so tests of
// Update notice tampered checksum files without needing real signatures.
type hashVerifier struct{}

func (hashVerifier) Name() string { return "hash" }

func (hashVerifier) SignatureAsset() string { return checksumAsset + ".hash" }

func (hashVerifier) sign(msg []byte) []byte {
	sum := sha256.Sum256(msg)
	return []byte(hex.EncodeToString(sum[:]))
}

func (h hashVerifier) Verify(_ context.Context, msg, sig []byte) (*Signer, error) {
	if !bytes.Equal(h.sign(msg), sig) {
		return nil, errors.New("signature verification failed")
	}
	return &Signer{Identity: "hash"}, nil
}

type fakeVerifier struct {
	asset string
	err   error
}

func (fakeVerifier) Name() string { return "fake" }

func (f fakeVerifier) SignatureAsset() string { return f.asset }

func (f fakeVerifier) Verify(context.Context, []byte, []byte) (*Signer, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &Signer{Identity: "fake"}, nil
}

func TestDefaultVerifiers(t *testing.T) {
	keyless := DefaultVerifiers()[len(DefaultVerifiers())-1].(CosignKeyless)
	re := regexp.MustCompile(keyless.SubjectRegexp)
	for subject, want := range map[string]bool{
		"https://github.com/soerenschneider/sc/.github/workflows/release-container.yaml@refs/heads/main":    true,
		"https://github.com/soerenschneider/sc/.github/workflows/release-container.yaml@refs/heads/mainx":   false,
		"https://github.com/soerenschneider/sc/.github/workflows/release-container.yaml@refs/heads/feature": false,
		"https://github.com/soerenschneider/sc/.github/workflows/release-container.yaml@refs/pull/1/merge":  false,
		"https://github.com/soerenschneider/sc/.github/workflows/test.yaml@refs/heads/main":                 false,
		"https://github.com/soerenschneider/scx/.github/workflows/release-container.yaml@refs/heads/main":   false,
		"https://github.com/evil/sc/.github/workflows/release-container.yaml@refs/heads/main":               false,
	} {
		if got := re.MatchString(subject); got != want {
			t.Errorf("match(%s) = %v, want %v", subject, got, want)
		}
	}
}
