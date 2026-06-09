package tui

import "charm.land/lipgloss/v2"

func Max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// ---------------------------------------------------------------------------
// shared styles
// ---------------------------------------------------------------------------

var (
	DirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Bold(true)
	LeafStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)

	PathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	MetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	HistoricalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	HistoricalTagStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("214")).
				Bold(true).
				Padding(0, 1)

	KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110")).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	MaskedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("76"))

	DirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
)
