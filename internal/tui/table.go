package tui

import (
	"cmp"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

type TableOpts struct {
	Wrap      bool
	FullWidth bool
	Style     *func(row, col int) lipgloss.Style
}

var defaultStyle = func(row, col int) lipgloss.Style {
	headerTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("213")).
		Bold(true).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 2)

	//highlightStyle := lipgloss.NewStyle().
	//	Foreground(lipgloss.Color("231")). // bright white
	//	Background(lipgloss.Color("161")). // rich red
	//	Padding(0, 2)

	if row == table.HeaderRow {
		return headerTextStyle
	}

	return cellStyle
}

func PrintTable(tableHeader string, headers []string, data [][]string, opts TableOpts) {
	// Define styles

	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	tableTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).  // magenta-like
		Background(lipgloss.Color("236")). // dark gray
		Padding(0, 2).
		Align(lipgloss.Center).
		MarginBottom(1).
		MarginTop(1)

	// Get terminal width if needed
	var width int
	if opts.FullWidth {
		var err error
		width, _, err = term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			fmt.Println("Error getting terminal size:", err)
			return
		}
	}

	styleFunc := cmp.Or(opts.Style, &defaultStyle)
	// Build the table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		Rows(data...).
		Width(width).
		StyleFunc(*styleFunc).
		Wrap(opts.Wrap)

	// Print header and table
	fmt.Println(tableTitleStyle.Render(" " + tableHeader + " "))
	fmt.Println(t.String())
}
