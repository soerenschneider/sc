package cmd

import (
	"bytes"
	"testing"
)

func TestPrintNames(t *testing.T) {
	tests := []struct {
		name   string
		roles  []string
		output string
		want   string
	}{
		{"simple", []string{"b", "a"}, approleOutputSimple, "a\nb\n"},
		{"simple empty", nil, approleOutputSimple, ""},
		{"json", []string{"b", "a"}, approleOutputJson, "[\"a\",\"b\"]\n"},
		{"json empty", nil, approleOutputJson, "[]\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := printNames(&buf, tt.roles, tt.output); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
