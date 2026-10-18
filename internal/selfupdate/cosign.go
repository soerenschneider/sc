package selfupdate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

const (
	githubActionsIssuer = "https://token.actions.githubusercontent.com"
	sigstoreBundleAsset = checksumAsset + ".sigstore.json"
)

// CosignKeyless verifies Sigstore bundles created with "cosign sign-blob
// --bundle" without a key, i.e. signed with a short-lived Fulcio certificate
// for an OIDC identity. The signature must be logged in Rekor and the
// certificate must be issued by Issuer for a subject matching SubjectRegexp.
type CosignKeyless struct {
	Issuer        string
	SubjectRegexp string
	// TrustedMaterial replaces the Sigstore public good instance's trust
	// root, which is fetched via TUF otherwise.
	TrustedMaterial root.TrustedMaterial

	// withoutSCTs disables checking the certificate transparency log, for
	// tests with certificates that do not embed timestamps of one.
	withoutSCTs bool
}

func (CosignKeyless) Name() string { return "cosign keyless" }

func (CosignKeyless) SignatureAsset() string { return sigstoreBundleAsset }

func (c CosignKeyless) Verify(ctx context.Context, msg, sig []byte) (*Signer, error) {
	b, err := parseBundle(sig)
	if err != nil {
		return nil, err
	}
	// checked before fetching the trust root to fail fast and clearly
	if content, err := b.VerificationContent(); err == nil && content.Certificate() == nil {
		return nil, errors.New("bundle is signed with a key, not keyless")
	}
	trusted, err := trustedMaterial(ctx, c.TrustedMaterial)
	if err != nil {
		return nil, err
	}
	return c.verifyEntity(b, msg, trusted)
}

func (c CosignKeyless) verifyEntity(entity verify.SignedEntity, msg []byte, trusted root.TrustedMaterial) (*Signer, error) {
	if c.Issuer == "" || c.SubjectRegexp == "" {
		return nil, errors.New("no trusted signer identity configured")
	}
	identity, err := verify.NewShortCertificateIdentity(c.Issuer, "", "", c.SubjectRegexp)
	if err != nil {
		return nil, err
	}
	opts := []verify.VerifierOption{verify.WithTransparencyLog(1), verify.WithObserverTimestamps(1)}
	if !c.withoutSCTs {
		opts = append(opts, verify.WithSignedCertificateTimestamps(1))
	}
	result, err := verifyBundle(entity, msg, trusted, verify.WithCertificateIdentity(identity), opts...)
	if err != nil {
		return nil, err
	}

	signer := &Signer{Details: bundleDetails(entity, result)}
	if cert := result.Signature.Certificate; cert != nil {
		signer.Identity = cert.SubjectAlternativeName
		for key, value := range map[string]string{
			"issuer":       cert.Issuer,
			"commit":       cert.SourceRepositoryDigest,
			"ref":          cert.SourceRepositoryRef,
			"trigger":      cert.BuildTrigger,
			"workflow_run": cert.RunInvocationURI,
		} {
			if value != "" {
				signer.Details[key] = value
			}
		}
	}
	return signer, nil
}

func parseBundle(data []byte) (*bundle.Bundle, error) {
	var b bundle.Bundle
	if err := b.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("invalid sigstore bundle: %w", err)
	}
	return &b, nil
}

func verifyBundle(entity verify.SignedEntity, msg []byte, trusted root.TrustedMaterial, identity verify.PolicyOption, opts ...verify.VerifierOption) (*verify.VerificationResult, error) {
	verifier, err := verify.NewVerifier(trusted, opts...)
	if err != nil {
		return nil, err
	}
	return verifier.Verify(entity, verify.NewPolicy(verify.WithArtifact(bytes.NewReader(msg)), identity))
}

// bundleDetails describes where and when a verified signature was logged.
func bundleDetails(entity verify.SignedEntity, result *verify.VerificationResult) map[string]string {
	details := map[string]string{}
	for _, ts := range result.VerifiedTimestamps {
		details["signed_at"] = ts.Timestamp.UTC().Format(time.RFC3339)
		details["timestamp_by"] = ts.URI
	}
	if entries, err := entity.TlogEntries(); err == nil && len(entries) > 0 {
		details["rekor_log_index"] = strconv.FormatInt(entries[0].LogIndex(), 10)
	}
	return details
}

func trustedMaterial(ctx context.Context, override root.TrustedMaterial) (root.TrustedMaterial, error) {
	if override != nil {
		return override, nil
	}
	opts := tuf.DefaultOptions().WithContext(ctx)
	if dir, err := os.UserCacheDir(); err == nil {
		opts = opts.WithCachePath(filepath.Join(dir, "sc", "sigstore-tuf"))
	}
	trusted, err := root.FetchTrustedRootWithOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("could not fetch sigstore trust root: %w", err)
	}
	return trusted, nil
}
