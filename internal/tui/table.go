package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func PrintTable(tableHeader string, headers []string, data [][]string) {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 1)
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("210")).Padding(0, 1)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(headers...).
		Rows(data...).
		//Width(160).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			cell := data[row][col]
			if num, err := strconv.ParseFloat(cell, 32); err == nil && num >= 2000 {
				return redStyle
			} else if strings.HasPrefix(cell, "down") {
				return redStyle
			}

			return cellStyle
		}).
		Wrap(false)

	t.Wrap(false)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1).
		Align(lipgloss.Center).
		MarginBottom(0).MarginTop(1)
	fmt.Println(header.Render(tableHeader))
	fmt.Println(t.String())
}
