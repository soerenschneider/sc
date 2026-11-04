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
		{"simple", []string{"b", "a"}, listOutputSimple, "a\nb\n"},
		{"simple empty", nil, listOutputSimple, ""},
		{"json", []string{"b", "a"}, listOutputJson, "[\n  \"a\",\n  \"b\"\n]\n"},
		{"json empty", nil, listOutputJson, "[]\n"},
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

func TestDefaultListOutput(t *testing.T) {
	tests := []struct {
		name     string
		formats  []string
		terminal bool
		want     string
	}{
		{"table terminal", tableOutputFormats, true, listOutputPretty},
		{"table pipe", tableOutputFormats, false, listOutputSimple},
		{"names terminal", nameListOutputFormats, true, listOutputSimple},
		{"names pipe", nameListOutputFormats, false, listOutputSimple},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultListOutput(tt.formats, tt.terminal); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
