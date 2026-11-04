package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

var (
	keyStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))  // Blue
	warnStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("202")) // Orange/red
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))            // Gray

	// BadgeStyle is used for headings of command output, e.g. the name of the displayed resource.
	BadgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1)
)

// PrintHeader prints title as badge followed by an empty line. Nothing is printed if stdout is not a
// terminal, so the output stays parseable when piped.
func PrintHeader(title string) {
	if !term.IsTerminal(int(os.Stdout.Fd())) { //#nosec:G115
		return
	}
	_, _ = lipgloss.Println(BadgeStyle.Render(title))
	fmt.Println()
}

// PrintMapOutput prints data as aligned "key  value" lines, sorted by key, see RenderMap.
func PrintMapOutput(data map[string]any, problematicColumns ...string) {
	_, _ = lipgloss.Println(RenderMap(data, problematicColumns...))
}

// RenderMap renders data as aligned "key  value" lines, sorted by key and without trailing newline. The keys
// given as problematicColumns are highlighted. Values are formatted using FormatValue.
func RenderMap(data map[string]any, problematicColumns ...string) string {
	keys := make([]string, 0, len(data))
	maxKeyLen := 0
	for k := range data {
		keys = append(keys, k)
		maxKeyLen = max(maxKeyLen, len(k))
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		style := keyStyle
		if slices.Contains(problematicColumns, k) {
			style = warnStyle
		}
		paddedKey := fmt.Sprintf("%-*s", maxKeyLen, k)
		lines = append(lines, style.Render(paddedKey)+"  "+valueStyle.Render(FormatValue(k, data[k])))
	}

	return strings.Join(lines, "\n")
}

// FormatValue formats a single value as returned by Vault for display. Empty values are displayed as "-",
// lists are joined and maps are displayed as sorted "key=value" pairs. Numbers whose key denotes a TTL or
// period are returned by Vault in seconds and are displayed as durations.
func FormatValue(key string, val any) string {
	switch v := val.(type) {
	case nil:
		return "-"
	case string:
		if v == "" {
			return "-"
		}
		return v
	case []string:
		if len(v) == 0 {
			return "-"
		}
		return strings.Join(v, ", ")
	case []any:
		if len(v) == 0 {
			return "-"
		}
		items := make([]string, 0, len(v))
		for _, item := range v {
			items = append(items, FormatValue("", item))
		}
		return strings.Join(items, ", ")
	case map[string]any:
		if len(v) == 0 {
			return "-"
		}
		pairs := make([]string, 0, len(v))
		for k, item := range v {
			pairs = append(pairs, k+"="+FormatValue(k, item))
		}
		sort.Strings(pairs)
		return strings.Join(pairs, ", ")
	case json.Number:
		if strings.HasSuffix(key, "ttl") || strings.HasSuffix(key, "period") {
			if seconds, err := v.Int64(); err == nil {
				return (time.Duration(seconds) * time.Second).String()
			}
		}
	}
	return fmt.Sprintf("%v", val)
}

func WriteListOutput[T any](items []T) {
	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // Gray

	var builder strings.Builder
	for _, item := range items {
		builder.WriteString(itemStyle.Render(fmt.Sprintf("- %v", item)) + "\n")
	}
	fmt.Println(builder.String())
}
