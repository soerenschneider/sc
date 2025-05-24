package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func PrintMapOutput(data map[string]any) {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")) // Blue

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // Gray

	// Get and sort keys
	keys := make([]string, 0, len(data))
	maxKeyLen := 0
	for k := range data {
		keys = append(keys, k)
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}
	sort.Strings(keys)

	// Build output
	indent := 2
	var builder strings.Builder
	for _, k := range keys {
		v := data[k]
		paddedKey := fmt.Sprintf("%-*s", maxKeyLen, k)
		key := keyStyle.Render(paddedKey)
		val := valueStyle.Render(fmt.Sprintf("%v", v))
		spaces := strings.Repeat(" ", indent)
		builder.WriteString(fmt.Sprintf("%s%s%s\n", key, spaces, val))
	}

	rendered := builder.String()
	fmt.Println(rendered)
	rendered = strings.ReplaceAll(rendered, "_", "\\_")

	_ = huh.NewNote().Title("Your data").Description(rendered).Run()
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
