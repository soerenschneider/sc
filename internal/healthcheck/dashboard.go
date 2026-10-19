package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/soerenschneider/sc/internal/tui"
)

const (
	minCardWidth = 38
	maxWidth     = 140
	barWidth     = 10
)

var (
	okFg     = tui.LightDark(lipgloss.Color("#2e7d32"), lipgloss.Color("#5fd787"))
	warnFg   = tui.LightDark(lipgloss.Color("#b7791f"), lipgloss.Color("#f1c40f"))
	badFg    = tui.LightDark(lipgloss.Color("#c0392b"), lipgloss.Color("#fe5f86"))
	accentFg = tui.LightDark(lipgloss.Color("#7b1fa2"), lipgloss.Color("#d4b3ff"))
	textFg   = tui.LightDark(lipgloss.Color("#202020"), lipgloss.Color("#e4e4e4"))
	dimFg    = tui.LightDark(lipgloss.Color("#8a8a8a"), lipgloss.Color("#6c6c6c"))
	borderFg = tui.LightDark(lipgloss.Color("#c8c8c8"), lipgloss.Color("#444444"))
	pillFg   = tui.LightDark(lipgloss.Color("#ffffff"), lipgloss.Color("#1a1a1a"))

	okStyle     = lipgloss.NewStyle().Foreground(okFg)
	warnStyle   = lipgloss.NewStyle().Foreground(warnFg)
	badStyle    = lipgloss.NewStyle().Foreground(badFg)
	textStyle   = lipgloss.NewStyle().Foreground(textFg)
	dimStyle    = lipgloss.NewStyle().Foreground(dimFg)
	borderStyle = lipgloss.NewStyle().Foreground(borderFg)
	titleStyle  = lipgloss.NewStyle().Foreground(accentFg).Bold(true)
)

// RenderOptions controls how a report is rendered.
type RenderOptions struct {
	// Width is the available terminal width.
	Width int
	// Domain is stripped from host names to keep them short and displayed in the header instead.
	Domain string
	// Status replaces the time of the report in the header, e.g. while refreshing.
	Status string
	// Footer is rendered below the dashboard, e.g. key bindings.
	Footer string
}

// Render renders the report as a dashboard of cards that are arranged in as many columns as the width permits.
func Render(r *Report, opts RenderOptions) string {
	width := min(max(opts.Width, minCardWidth), maxWidth)

	cols := max(1, min(3, (width+1)/(minCardWidth+1)))
	cardWidth := (width - (cols - 1)) / cols
	// Let the cards span the full width, accounting for the rounding above.
	width = cardWidth*cols + cols - 1

	cards := make([]string, 0, len(r.Groups)+1)
	for _, g := range r.Groups {
		cards = append(cards, renderGroup(g, opts.Domain, cardWidth))
	}
	if r.LatencyTitle != "" {
		cards = append(cards, renderLatency(r, opts.Domain, cardWidth))
	}

	sections := []string{
		renderHeader(r, opts, width),
		"",
		masonry(cards, cols),
	}
	if r.LogsEnabled {
		sections = append(sections, renderLogs(r, width))
	}
	if opts.Footer != "" {
		sections = append(sections, dimStyle.Render(" "+opts.Footer))
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// masonry distributes the cards across the columns, always appending to the shortest column.
func masonry(cards []string, cols int) string {
	columns := make([][]string, cols)
	heights := make([]int, cols)
	for _, c := range cards {
		shortest := 0
		for i := range heights {
			if heights[i] < heights[shortest] {
				shortest = i
			}
		}
		columns[shortest] = append(columns[shortest], c)
		heights[shortest] += lipgloss.Height(c)
	}

	rendered := make([]string, 0, 2*cols-1)
	for i, col := range columns {
		if i > 0 {
			rendered = append(rendered, " ")
		}
		rendered = append(rendered, lipgloss.JoinVertical(lipgloss.Left, col...))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func renderHeader(r *Report, opts RenderOptions, width int) string {
	up, total := r.Counts()
	down := total - up

	label, bg := "HEALTHY", okFg
	switch {
	case len(r.Groups) > 0 && r.Groups[0].Up() == 0:
		label, bg = "OFFLINE", badFg
	case down > 0:
		label, bg = "DEGRADED", warnFg
	}
	pill := lipgloss.NewStyle().Background(bg).Foreground(pillFg).Bold(true).Padding(0, 1).Render(label)

	name := "health"
	if opts.Domain != "" {
		name = opts.Domain + " " + dimStyle.Render("health")
	}
	left := pill + "  " + textStyle.Bold(true).Render(name)

	counts := okStyle.Render(fmt.Sprintf("● %d up", up))
	if down > 0 {
		counts += "  " + badStyle.Render(fmt.Sprintf("✕ %d down", down))
	}
	status := opts.Status
	if status == "" {
		status = fmt.Sprintf("%s · took %s", r.Time.Format("15:04:05"), r.Took.Round(100*time.Millisecond))
	}
	right := counts + dimStyle.Render("   "+status)

	gap := max(2, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func renderGroup(g Group, domain string, width int) string {
	// Show a port shared by all endpoints once in the title rather than on every line.
	port := commonPort(g.Probes)
	title := titleStyle.Render(g.Name)
	if port != "" {
		title += dimStyle.Render(" :" + port)
	}

	inner := width - 4
	lines := make([]string, 0, len(g.Probes))
	for _, p := range g.Probes {
		name := shortHost(p.Name, domain, port != "")
		if p.Err != nil {
			lines = append(lines, row(badStyle.Render("✕"), textStyle.Render(name), badStyle.Render(shortErr(p.Err)), inner))
			continue
		}
		ms := float64(p.Duration) / float64(time.Millisecond)
		value := latencyBar(ms) + " " + latencyStyle(ms).Render(fmt.Sprintf("%6s", formatMs(ms)))
		lines = append(lines, row(okStyle.Render("●"), textStyle.Render(name), value, inner))
	}

	return card(title, healthBadge(g.Up(), len(g.Probes)), lines, width)
}

func renderLatency(r *Report, domain string, width int) string {
	inner := width - 4
	title := titleStyle.Render(r.LatencyTitle)

	var lines []string
	switch {
	case r.LatencyErr != nil:
		lines = []string{badStyle.Render("✕ unavailable · " + shortErr(r.LatencyErr))}
	case len(r.Latency) == 0:
		lines = []string{dimStyle.Render("no data")}
	}
	for _, s := range r.Latency {
		value := latencyBar(s.Avg) + " " + latencyStyle(s.Avg).Render(fmt.Sprintf("%6s", formatMs(s.Avg))) +
			dimStyle.Render(fmt.Sprintf(" ↑%6s", formatMs(s.Max)))
		lines = append(lines, row(latencyStyle(s.Max).Render("◆"), textStyle.Render(shortHost(s.Instance, domain, true)), value, inner))
	}

	badge := ""
	if r.LatencyErr == nil && len(r.Latency) > 0 {
		badge = dimStyle.Render("avg ↑max")
	}
	return card(title, badge, lines, width)
}

func renderLogs(r *Report, width int) string {
	title := titleStyle.Render("Recent errors")

	var lines []string
	var badge string
	switch {
	case r.LogsErr != nil:
		lines = []string{badStyle.Render("✕ unavailable · " + shortErr(r.LogsErr))}
	case len(r.Logs) == 0:
		lines = []string{okStyle.Render("✓ nothing to see here")}
		badge = okStyle.Render("0")
	default:
		badge = badStyle.Render(strconv.Itoa(len(r.Logs)))
	}
	for _, l := range r.Logs {
		msg := strings.Join(strings.Fields(l.Message), " ")
		lines = append(lines, dimStyle.Render(l.Timestamp)+"  "+textStyle.Render(msg))
	}
	return card(title, badge, lines, width)
}

// card renders lines in a rounded box with the title embedded in the top border and the badge right-aligned.
func card(title, badge string, lines []string, width int) string {
	inner := width - 4

	fill := width - 8 - lipgloss.Width(title) - lipgloss.Width(badge)
	if badge == "" {
		fill += 2
	}
	var b strings.Builder
	b.WriteString(borderStyle.Render("╭─ ") + title + " " + borderStyle.Render(strings.Repeat("─", max(0, fill))))
	if badge != "" {
		b.WriteString(" " + badge + " ")
	}
	b.WriteString(borderStyle.Render("─╮"))

	for _, line := range lines {
		line = ansi.Truncate(line, inner, "…")
		pad := strings.Repeat(" ", max(0, inner-lipgloss.Width(line)))
		b.WriteString("\n" + borderStyle.Render("│ ") + line + pad + borderStyle.Render(" │"))
	}
	b.WriteString("\n" + borderStyle.Render("╰"+strings.Repeat("─", width-2)+"╯"))
	return b.String()
}

// row renders an icon and a name on the left and a value on the right, truncating the name if required.
func row(icon, name, value string, width int) string {
	nameWidth := width - lipgloss.Width(value) - 3
	if nameWidth < 4 {
		value = ansi.Truncate(value, max(0, width-7), "…")
		nameWidth = width - lipgloss.Width(value) - 3
	}
	name = ansi.Truncate(name, nameWidth, "…")
	pad := strings.Repeat(" ", max(1, nameWidth-lipgloss.Width(name)+1))
	return icon + " " + name + pad + value
}

func healthBadge(up, total int) string {
	style := okStyle
	switch {
	case up == 0:
		style = badStyle
	case up < total:
		style = warnStyle
	}
	return style.Bold(true).Render(fmt.Sprintf("%d/%d", up, total))
}

// latencyBar renders a bar on a logarithmic scale from 1ms to 1s.
func latencyBar(ms float64) string {
	frac := math.Log10(max(ms, 1)) / 3
	filled := int(math.Round(min(max(frac, 0.1), 1) * barWidth))
	return latencyStyle(ms).Render(strings.Repeat("━", filled)) + borderStyle.Render(strings.Repeat("━", barWidth-filled))
}

func latencyStyle(ms float64) lipgloss.Style {
	switch {
	case ms < 50:
		return okStyle
	case ms < 250:
		return warnStyle
	default:
		return badStyle
	}
}

func formatMs(ms float64) string {
	switch {
	case ms >= 1000:
		return fmt.Sprintf("%.1fs", ms/1000)
	case ms >= 10:
		return fmt.Sprintf("%.0fms", ms)
	default:
		return fmt.Sprintf("%.1fms", ms)
	}
}

func commonPort(probes []Probe) string {
	port := ""
	for _, p := range probes {
		_, pp, err := net.SplitHostPort(p.Name)
		if err != nil || (port != "" && pp != port) {
			return ""
		}
		port = pp
	}
	return port
}

// shortHost strips the domain and, if dropPort is set, the port from a host name.
func shortHost(name, domain string, dropPort bool) string {
	host, port, err := net.SplitHostPort(name)
	if err != nil {
		host, port = name, ""
	}
	if domain != "" {
		host = strings.TrimSuffix(host, "."+domain)
	}
	if port == "" || dropPort {
		return host
	}
	return host + dimStyle.Render(":"+port)
}

// shortErr condenses the verbose errors returned by the net package into a few words.
func shortErr(err error) string {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		switch {
		case dnsErr.IsNotFound:
			return "no such host"
		case dnsErr.IsTimeout:
			return "dns timeout"
		}
	}

	var netErr net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &netErr) && netErr.Timeout():
		return "timeout"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "refused"
	case errors.Is(err, syscall.EHOSTUNREACH):
		return "host unreachable"
	case errors.Is(err, syscall.ENETUNREACH):
		return "network unreachable"
	case errors.Is(err, syscall.ECONNRESET):
		return "connection reset"
	}

	msg := err.Error()
	if i := strings.LastIndex(msg, ": "); i >= 0 {
		msg = msg[i+2:]
	}
	return msg
}
