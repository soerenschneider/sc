package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

func PrintTable(tableHeader string, headers []string, data [][]string, useFullWidth bool) {
	// Define styles
	headerTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("213")).
		Bold(true).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 2)

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")). // bright white
		Background(lipgloss.Color("161")). // rich red
		Padding(0, 2)

	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	tableTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")). // magenta-like
		Background(lipgloss.Color("236")). // dark gray
		Padding(0, 2).
		Align(lipgloss.Center).
		MarginBottom(1).
		MarginTop(1)

	// Get terminal width if needed
	var width int
	if useFullWidth {
		var err error
		width, _, err = term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			fmt.Println("Error getting terminal size:", err)
			return
		}
	}

	// Build the table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		Rows(data...).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerTextStyle
			}
			cell := data[row][col]
			if num, err := strconv.ParseFloat(cell, 32); err == nil && num >= 2000 {
				return highlightStyle
			} else if strings.HasPrefix(strings.ToLower(cell), "down") {
				return highlightStyle
			}
			return cellStyle
		})

	t.Wrap(true)

	// Print header and table
	fmt.Println(tableTitleStyle.Render(" " + tableHeader + " "))
	fmt.Println(t.String())
}
