package selfupdate

import (
	"context"
	"regexp"
	"strings"
)

// releaseWorkflow is the GitHub Actions workflow whose keyless signatures are
// trusted. Renaming it breaks updates of binaries built before the rename.
const releaseWorkflow = ".github/workflows/release-container.yaml"

// Verifier authenticates the checksum file of a release. Every implementation
// expects its signature in a release asset of its own, so a release can carry
// signatures of several schemes at once, e.g. while migrating between them.
type Verifier interface {
	// Name identifies the signature scheme in errors.
	Name() string
	// SignatureAsset returns the name of the release asset that holds the
	// signature of the checksum file.
	SignatureAsset() string
	// Verify checks that sig is a valid signature of msg by a trusted signer
	// and describes that signer.
	Verify(ctx context.Context, msg, sig []byte) (*Signer, error)
}

// Signer describes who made a signature that a Verifier accepted.
type Signer struct {
	// Identity is the trusted key or certificate identity that signed.
	Identity string
	// Details holds further scheme specific facts, e.g. when the signature
	// was logged in a transparency log.
	Details map[string]string
}

// DefaultVerifiers returns verifiers for the signers trusted by this build.
// Binaries only ever trust the signers they were built with, so keep the
// previous signer in here for a release when rotating keys or schemes.
func DefaultVerifiers() []Verifier {
	var verifiers []Verifier
	verifiers = append(verifiers, CosignKeyless{
		Issuer: githubActionsIssuer,
		// Only the release workflow running on main, which is where
		// release-please creates releases, so neither other workflows,
		// branches nor pull requests can sign releases.
		SubjectRegexp: "^https://" + regexp.QuoteMeta(Repository+"/"+releaseWorkflow+"@refs/heads/main") + "$",
	})
	return verifiers
}

func verifierNames(verifiers []Verifier) string {
	names := make([]string, 0, len(verifiers))
	for _, v := range verifiers {
		names = append(names, v.Name())
	}
	return strings.Join(names, ", ")
}
