package healthcheck

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// collectTimeout bounds a single collection run.
const collectTimeout = 10 * time.Second

// Watch displays the dashboard in the alt screen and refreshes it every interval until the user quits.
func Watch(ctx context.Context, opts Options, domain string, interval time.Duration) error {
	m := &watchModel{ctx: ctx, opts: opts, domain: domain, interval: interval}
	_, err := tea.NewProgram(m, tea.WithContext(ctx)).Run()
	return err
}

type reportMsg *Report

type tickMsg time.Time

type watchModel struct {
	ctx      context.Context
	opts     Options
	domain   string
	interval time.Duration

	report  *Report
	loading bool
	next    time.Time
	width   int
}

func (m *watchModel) collect() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, collectTimeout)
		defer cancel()
		return reportMsg(Collect(ctx, m.opts))
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *watchModel) Init() tea.Cmd {
	return tea.Batch(m.collect(), tick())
}

func (m *watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case reportMsg:
		m.report = msg
		m.loading = false
		m.next = time.Now().Add(m.interval)
	case tickMsg:
		if !m.loading && !m.next.IsZero() && time.Now().After(m.next) {
			return m, tea.Batch(m.collect(), tick())
		}
		return m, tick()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "r":
			if !m.loading {
				return m, m.collect()
			}
		}
	}
	return m, nil
}

func (m *watchModel) View() tea.View {
	var content string
	if m.report == nil {
		content = lipgloss.NewStyle().Padding(1, 2).Render(dimStyle.Render("Collecting health data…"))
	} else {
		status := ""
		if m.loading {
			status = "refreshing…"
		}
		next := max(0, time.Until(m.next).Round(time.Second))
		content = Render(m.report, RenderOptions{
			Width:  m.width,
			Domain: m.domain,
			Status: status,
			Footer: fmt.Sprintf("r refresh · q quit · next refresh in %s", next),
		})
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
