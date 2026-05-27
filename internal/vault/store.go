// Package vaultedit provides a TUI for editing versioned key/value secrets.
//
// The TUI is decoupled from any specific backend through the SecretStore
// interface defined in this file. A Vault KV v2 implementation ships in
// vaultstore.go; you can substitute your own (e.g. a fake for tests, or
// an adapter for a different secrets backend).
package vault

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// SecretStore is the backend abstraction used by the editor and browser.
//
// Implementations MUST return errors wrapped in *StoreError with the
// correct Kind so the UI can react appropriately. In particular, the UI
// depends on ErrKindNotFound (to start in "new secret" mode) and
// ErrKindCASConflict (to surface a clear "version changed" message).
type SecretStore interface {
	// Get reads the current version of a secret. Returns a StoreError with
	// Kind == ErrKindNotFound when no readable version exists.
	Get(ctx context.Context, mount, path string) (*Secret, error)

	// GetVersion reads a specific historical version. Returns
	// ErrKindNotFound when the version is missing, soft-deleted, or
	// destroyed.
	GetVersion(ctx context.Context, mount, path string, version int) (*Secret, error)

	// GetMetadata returns the version history (without secret values).
	GetMetadata(ctx context.Context, mount, path string) (*SecretMetadata, error)

	// Put writes a new version with a compare-and-set check.
	// cas is the version the caller expects to be current: pass 0 to
	// assert the secret does not exist yet, otherwise pass the version
	// returned by the most recent Get/Put. Returns a StoreError with
	// Kind == ErrKindCASConflict when the server's version has advanced.
	Put(ctx context.Context, mount, path string, data map[string]string, cas int) (*Secret, error)

	// SoftDelete marks the latest version as deleted while preserving
	// history (so it can be undeleted out-of-band).
	SoftDelete(ctx context.Context, mount, path string) error

	// List returns the immediate children of prefix. Prefix is "" for the
	// mount root; otherwise it ends with "/". Children that end in "/" are
	// represented with IsDir == true.
	List(ctx context.Context, mount, prefix string) ([]Entry, error)
}

// Secret is the storage-neutral view of a versioned KV secret.
type Secret struct {
	Data      map[string]string
	Version   int
	UpdatedAt time.Time
}

// SecretMetadata is the storage-neutral view of a secret's version history.
type SecretMetadata struct {
	CurrentVersion int
	OldestVersion  int
	Versions       []VersionInfo // newest first
}

// VersionInfo describes one historical version. DeletedAt is zero if the
// version is still readable; Destroyed is true if the version's data has
// been permanently removed.
type VersionInfo struct {
	Version   int
	CreatedAt time.Time
	DeletedAt time.Time
	Destroyed bool
}

// Readable reports whether the version's data can still be fetched.
func (v VersionInfo) Readable() bool {
	return !v.Destroyed && v.DeletedAt.IsZero()
}

// Entry is one item returned by SecretStore.List.
type Entry struct {
	Name  string
	IsDir bool
}

// ErrorKind classifies StoreError so the UI can decide how to react.
type ErrorKind int

const (
	ErrKindUnknown ErrorKind = iota
	ErrKindNotFound
	ErrKindPermissionDenied
	ErrKindCASConflict
	ErrKindRateLimited
	ErrKindServerError
	ErrKindBadRequest
	ErrKindTimeout
	ErrKindNetwork
)

func (k ErrorKind) String() string {
	switch k {
	case ErrKindNotFound:
		return "not-found"
	case ErrKindPermissionDenied:
		return "permission-denied"
	case ErrKindCASConflict:
		return "cas-conflict"
	case ErrKindRateLimited:
		return "rate-limited"
	case ErrKindServerError:
		return "server-error"
	case ErrKindBadRequest:
		return "bad-request"
	case ErrKindTimeout:
		return "timeout"
	case ErrKindNetwork:
		return "network"
	default:
		return "unknown"
	}
}

// StoreError is the typed error every SecretStore implementation must return.
type StoreError struct {
	Kind    ErrorKind
	Op      string // "get", "get-version", "get-metadata", "put", "delete", "list"
	Message string // human-readable detail (safe to display; must not contain secret values)
	Cause   error  // underlying error, if any
}

func (e *StoreError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %s", e.Op, e.Kind, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Kind)
}

func (e *StoreError) Unwrap() error { return e.Cause }

// IsKind reports whether err is (or wraps) a *StoreError with the given Kind.
func IsKind(err error, k ErrorKind) bool {
	var se *StoreError
	return errors.As(err, &se) && se.Kind == k
}

// Convenience predicates for the two kinds the UI cares about most.
func IsNotFound(err error) bool    { return IsKind(err, ErrKindNotFound) }
func IsCASConflict(err error) bool { return IsKind(err, ErrKindCASConflict) }
