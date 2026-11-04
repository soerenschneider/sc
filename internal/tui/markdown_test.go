package tui

import (
	"strings"
	"testing"
)

func TestMarkdownStreamBlocks(t *testing.T) {
	var out strings.Builder
	var blocks []string
	m := &MarkdownStream{out: &out, render: func(s string) (string, error) {
		blocks = append(blocks, s)
		return "<" + s + ">  \n", nil
	}}

	input := "# Title\n\nFirst paragraph\nstill first.\n\n```go\nfunc main() {\n\n\tfmt.Println()\n}\n```\n\n- item one\n\n  continued item\n- item two\n\nLast"
	// feed in small chunks, like a model streams its tokens
	for i := 0; i < len(input); i += 3 {
		if err := m.Write(input[i:min(i+3, len(input))]); err != nil {
			t.Fatal(err)
		}
	}

	if len(blocks) != 3 {
		t.Errorf("rendered %d blocks before the end, want 3: %q", len(blocks), blocks)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"# Title",
		"First paragraph\nstill first.",
		"```go\nfunc main() {\n\n\tfmt.Println()\n}\n```",
		"- item one\n\n  continued item\n- item two",
		"Last",
	}
	if strings.Join(blocks, "|") != strings.Join(want, "|") {
		t.Errorf("blocks =\n%q\nwant\n%q", blocks, want)
	}
	if !strings.HasPrefix(out.String(), "<# Title>\n\n<First") || strings.Contains(out.String(), " \n") || !strings.HasSuffix(out.String(), "<Last>") {
		t.Errorf("output = %q", out.String())
	}
}

func TestMarkdownStreamFenceLengths(t *testing.T) {
	var blocks []string
	m := &MarkdownStream{out: &strings.Builder{}, render: func(s string) (string, error) {
		blocks = append(blocks, s)
		return s, nil
	}}
	_ = m.Write("````md\n```\n\nstill code\n````\n\nafter\n")
	_ = m.Close()
	if len(blocks) != 2 || blocks[1] != "after" {
		t.Errorf("blocks = %q", blocks)
	}
}

func TestMarkdownStreamRenders(t *testing.T) {
	var out strings.Builder
	m, err := NewMarkdownStream(&out, 80)
	if err != nil {
		t.Fatal(err)
	}
	_ = m.Write("Some **bold** text\n\n| a | b |\n|---|---|\n| 1 | 2 |\n")
	_ = m.Close()
	if strings.Contains(out.String(), "**") || !strings.Contains(out.String(), "bold") {
		t.Errorf("output = %q", out.String())
	}
}

func TestTrimTrailingPadding(t *testing.T) {
	tests := map[string]string{
		"text   ": "text",
		"\x1b[1mbold\x1b[0m  \x1b[38;5;252m  \x1b[0m": "\x1b[1mbold\x1b[0m",
		"\x1b[38;5;252m     \x1b[0m":                  "",
		"  indented":                                  "  indented",
	}
	for in, want := range tests {
		if got := trimTrailingPadding(in); got != want {
			t.Errorf("trimTrailingPadding(%q) = %q, want %q", in, got, want)
		}
	}
}
