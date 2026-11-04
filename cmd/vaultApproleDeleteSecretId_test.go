package cmd

import (
	"slices"
	"testing"
	"time"
)

func TestApproleSecretIDLabels(t *testing.T) {
	expiration := time.Date(2026, 10, 21, 12, 0, 0, 0, time.Local)
	secretIDs := []approleSecretID{
		{Accessor: "short", Status: approleStatusValid, ExpirationTime: &expiration, Metadata: map[string]string{"host": "a"}},
		{Accessor: "longer-accessor", Status: approleStatusNonExpiring},
	}

	want := []string{
		"short            valid         expires 2026-10-21 12:00:00  host=a",
		"longer-accessor  non-expiring  expires never                -",
	}
	if got := approleSecretIDLabels(secretIDs); !slices.Equal(got, want) {
		t.Errorf("got\n%q\nwant\n%q", got, want)
	}
}
