package tui

import (
	"io"
	"regexp"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
)

// MarkdownStream renders markdown that arrives in chunks, e.g. the streamed
// answer of a language model. As markdown can only be rendered reliably as a
// whole, it renders every top level block, such as a paragraph, list or code
// block, as soon as the block is complete.
type MarkdownStream struct {
	out    io.Writer
	render func(string) (string, error)

	partial  string   // incomplete last line
	block    []string // lines of the current block
	inFence  string   // the opening fence while in a fenced code block
	sawBlank bool     // the last line was blank
	written  bool
}

// NewMarkdownStream returns a stream that renders to out, wrapping at width.
func NewMarkdownStream(out io.Writer, width int) (*MarkdownStream, error) {
	style := styles.LightStyleConfig
	if HasDarkBg {
		style = styles.DarkStyleConfig
	}
	r, err := glamour.NewTermRenderer(glamour.WithStyles(style), glamour.WithWordWrap(width))
	if err != nil {
		return nil, err
	}
	return &MarkdownStream{out: out, render: r.Render}, nil
}

// Write adds a chunk of markdown, rendering the blocks it completes.
func (m *MarkdownStream) Write(chunk string) error {
	m.partial += chunk
	for {
		i := strings.IndexByte(m.partial, '\n')
		if i < 0 {
			return nil
		}
		line := m.partial[:i]
		m.partial = m.partial[i+1:]
		if err := m.addLine(line); err != nil {
			return err
		}
	}
}

// Close renders what is left.
func (m *MarkdownStream) Close() error {
	if m.partial != "" {
		if err := m.addLine(m.partial); err != nil {
			return err
		}
		m.partial = ""
	}
	return m.flush()
}

func (m *MarkdownStream) addLine(line string) error {
	trimmed := strings.TrimSpace(line)

	if m.inFence != "" {
		m.block = append(m.block, line)
		if strings.HasPrefix(trimmed, m.inFence) && strings.Trim(trimmed, m.inFence[:1]) == "" {
			m.inFence = ""
		}
		return nil
	}

	if trimmed == "" {
		m.sawBlank = len(m.block) > 0
		if m.sawBlank {
			m.block = append(m.block, line)
		}
		return nil
	}

	// A block ends with a blank line, unless the next line is indented and
	// therefore continues it, e.g. a paragraph of a list item.
	if m.sawBlank && !startsIndented(line) {
		if err := m.flush(); err != nil {
			return err
		}
	}
	m.sawBlank = false

	if fence := fenceOf(trimmed); fence != "" {
		m.inFence = fence
	}
	m.block = append(m.block, line)
	return nil
}

func (m *MarkdownStream) flush() error {
	source := strings.TrimSpace(strings.Join(m.block, "\n"))
	m.block = m.block[:0]
	m.sawBlank = false
	if source == "" {
		return nil
	}

	rendered, err := m.render(source)
	if err != nil {
		// show the markdown rather than nothing
		rendered = source
	}

	lines := strings.Split(rendered, "\n")
	for i, l := range lines {
		lines[i] = trimTrailingPadding(l)
	}
	rendered = strings.Trim(strings.Join(lines, "\n"), "\n")

	if m.written {
		rendered = "\n\n" + rendered
	}
	m.written = true
	_, err = io.WriteString(m.out, rendered)
	return err
}

// trailingPadding matches the spaces glamour pads lines with, which may be
// interleaved with styling escape sequences.
var trailingPadding = regexp.MustCompile(`(?:[ \t]|\x1b\[[0-9;]*m)+$`)

// trimTrailingPadding removes the padding of a rendered line, resetting the
// styling if it contained any.
func trimTrailingPadding(line string) string {
	trimmed := trailingPadding.ReplaceAllString(line, "")
	if trimmed != "" && strings.Contains(trimmed, "\x1b[") {
		trimmed += "\x1b[0m"
	}
	return trimmed
}

func startsIndented(line string) bool {
	return strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")
}

// fenceOf returns the fence if line opens a fenced code block.
func fenceOf(line string) string {
	for _, c := range []string{"`", "~"} {
		n := len(line) - len(strings.TrimLeft(line, c))
		if n >= 3 {
			return line[:n]
		}
	}
	return ""
}
