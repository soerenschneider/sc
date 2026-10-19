package mpd

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/soerenschneider/sc/internal/tui"
)

const (
	commandTimeout  = 10 * time.Second
	refreshInterval = time.Second
)

// Browse opens a TUI to browse the database starting at dir, to manage the
// queue and to control playback.
func Browse(ctx context.Context, client *Client, dir string) error {
	return runBrowser(ctx, newBrowser(ctx, client, strings.Trim(dir, "/")))
}

// BrowseQueue opens the same TUI as Browse, starting with the queue.
func BrowseQueue(ctx context.Context, client *Client) error {
	m := newBrowser(ctx, client, "")
	m.view = viewQueue
	m.loading = true
	return runBrowser(ctx, m)
}

func runBrowser(ctx context.Context, m *browser) error {
	if err := m.loadInitial(); err != nil {
		return err
	}
	_, err := tea.NewProgram(m, tea.WithContext(ctx)).Run()
	return err
}

type view int

const (
	viewLibrary view = iota
	viewQueue
)

type (
	listedMsg struct {
		dir     string
		entries []Entry
		// focus is the path of the entry to put the cursor on
		focus string
		err   error
	}
	queuedMsg struct {
		songs   []Song
		version uint64
		err     error
	}
	nowPlayingMsg struct {
		status *Status
		song   *Song
		err    error
	}
	actionMsg struct {
		status string
		err    error
	}
	tickMsg struct{}
)

// listCursor is the cursor position and scroll offset of a list.
type listCursor struct {
	pos, offset int
}

func (c *listCursor) up() {
	c.pos = max(0, c.pos-1)
}

func (c *listCursor) down(n int) {
	c.pos = max(0, min(c.pos+1, n-1))
}

func (c *listCursor) clamp(n int) {
	c.pos = max(0, min(c.pos, n-1))
}

// scroll keeps the cursor within the visible rows.
func (c *listCursor) scroll(rows, n int) {
	if c.pos < c.offset {
		c.offset = c.pos
	}
	if c.pos >= c.offset+rows {
		c.offset = c.pos - rows + 1
	}
	c.offset = max(0, min(c.offset, tui.Max0(n-rows)))
}

type browser struct {
	ctx    context.Context
	client *Client

	view view

	dir     string
	entries []Entry
	lib     listCursor

	queue        []Song
	queueVersion uint64
	queueLoaded  bool
	qc           listCursor
	// focusID is the queue id of the song to keep the cursor on after the
	// queue was reloaded, e.g. after moving it
	focusID int
	// confirmClear is set after the first c, the second one clears the queue
	confirmClear bool

	filter   textinput.Model
	filterOn bool

	playing *Status
	song    *Song

	loading bool
	err     string
	status  string

	width, height int
}

func newBrowser(ctx context.Context, client *Client, dir string) *browser {
	ti := textinput.New()
	ti.Placeholder = "filter"
	ti.Prompt = "/ "
	ti.CharLimit = 256
	return &browser{ctx: ctx, client: client, dir: dir, filter: ti}
}

func (m *browser) loadInitial() error {
	ctx, cancel := context.WithTimeout(m.ctx, commandTimeout)
	defer cancel()
	entries, err := m.client.List(ctx, m.dir)
	if err != nil {
		return err
	}
	m.entries = sortEntries(entries)
	return nil
}

func sortEntries(entries []Entry) []Entry {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].IsDir && !entries[j].IsDir
	})
	return entries
}

// ---------------------------------------------------------------------------
// commands
// ---------------------------------------------------------------------------

func (m *browser) withTimeout(f func(ctx context.Context) tea.Msg) tea.Cmd {
	parent := m.ctx
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(parent, commandTimeout)
		defer cancel()
		return f(ctx)
	}
}

func (m *browser) listCmd(dir, focus string) tea.Cmd {
	client := m.client
	return m.withTimeout(func(ctx context.Context) tea.Msg {
		entries, err := client.List(ctx, dir)
		return listedMsg{dir: dir, entries: sortEntries(entries), focus: focus, err: err}
	})
}

func (m *browser) queueCmd() tea.Cmd {
	client := m.client
	return m.withTimeout(func(ctx context.Context) tea.Msg {
		// read the version first, so a change while listing triggers
		// another reload
		status, err := client.Status(ctx)
		if err != nil {
			return queuedMsg{err: err}
		}
		songs, err := client.Queue(ctx)
		return queuedMsg{songs: songs, version: status.QueueVersion, err: err}
	})
}

func (m *browser) refreshCmd() tea.Cmd {
	client := m.client
	return m.withTimeout(func(ctx context.Context) tea.Msg {
		status, err := client.Status(ctx)
		if err != nil {
			return nowPlayingMsg{err: err}
		}
		song, err := client.CurrentSong(ctx)
		return nowPlayingMsg{status: status, song: song, err: err}
	})
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *browser) actionCmd(f func(ctx context.Context, c *Client) (string, error)) tea.Cmd {
	client := m.client
	return m.withTimeout(func(ctx context.Context) tea.Msg {
		status, err := f(ctx, client)
		return actionMsg{status: status, err: err}
	})
}

// ---------------------------------------------------------------------------
// bubbletea
// ---------------------------------------------------------------------------

func (m *browser) Init() tea.Cmd {
	cmds := []tea.Cmd{m.refreshCmd(), tickCmd()}
	if m.view == viewQueue {
		cmds = append(cmds, m.queueCmd())
	}
	return tea.Batch(cmds...)
}

func (m *browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		return m, tea.Batch(m.refreshCmd(), tickCmd())

	case nowPlayingMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.playing, m.song = msg.status, msg.song
		// also picks up changes by other clients
		if m.view == viewQueue && m.queueLoaded && msg.status.QueueVersion != m.queueVersion {
			return m, m.queueCmd()
		}
		return m, nil

	case listedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.err, m.status = "", ""
		m.dir, m.entries = msg.dir, msg.entries
		m.lib = listCursor{}
		m.filter.SetValue("")
		for i, e := range m.entries {
			if e.Path == msg.focus {
				m.lib.pos = i
			}
		}
		return m, nil

	case queuedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.queue, m.queueVersion, m.queueLoaded = msg.songs, msg.version, true
		m.focusQueue()
		return m, nil

	case actionMsg:
		if msg.err != nil {
			m.err, m.status = msg.err.Error(), ""
			m.focusID = 0
		} else {
			m.err, m.status = "", msg.status
		}
		return m, m.refreshCmd()
	}

	if m.filterOn {
		if km, ok := msg.(tea.KeyPressMsg); ok {
			switch km.String() {
			case "esc":
				m.filterOn = false
				m.filter.SetValue("")
				m.filter.Blur()
				m.resetCursor()
				return m, nil
			case "enter":
				m.filterOn = false
				m.filter.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		m.resetCursor()
		return m, cmd
	}

	if km, ok := msg.(tea.KeyPressMsg); ok {
		if m.confirmClear && km.String() != "c" {
			// any other key cancels clearing the queue
			m.confirmClear = false
			m.status = ""
		}
		if m.view == viewQueue {
			return m.handleQueueKey(km)
		}
		return m.handleKey(km)
	}
	return m, nil
}

func (m *browser) resetCursor() {
	if m.view == viewQueue {
		m.qc = listCursor{}
	} else {
		m.lib = listCursor{}
	}
}

// focusQueue puts the cursor on the song with focusID if it is still queued,
// otherwise keeps it in range.
func (m *browser) focusQueue() {
	songs := m.filteredQueue()
	if m.focusID != 0 {
		for i, s := range songs {
			if s.ID == m.focusID {
				m.qc.pos = i
			}
		}
		m.focusID = 0
	}
	m.qc.clamp(len(songs))
}

func (m *browser) matches(names ...string) bool {
	q := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	for _, name := range names {
		if strings.Contains(strings.ToLower(name), q) {
			return true
		}
	}
	return false
}

func (m *browser) filteredEntries() []Entry {
	if strings.TrimSpace(m.filter.Value()) == "" {
		return m.entries
	}
	var out []Entry
	for _, e := range m.entries {
		if m.matches(e.Name(), path.Base(e.Path)) {
			out = append(out, e)
		}
	}
	return out
}

func (m *browser) filteredQueue() []Song {
	if strings.TrimSpace(m.filter.Value()) == "" {
		return m.queue
	}
	var out []Song
	for _, s := range m.queue {
		if m.matches(s.DisplayName(), path.Base(s.File)) {
			out = append(out, s)
		}
	}
	return out
}

func (m *browser) selected() (Entry, bool) {
	entries := m.filteredEntries()
	if m.lib.pos >= len(entries) {
		return Entry{}, false
	}
	return entries[m.lib.pos], true
}

func (m *browser) selectedSong() (Song, bool) {
	songs := m.filteredQueue()
	if m.qc.pos >= len(songs) {
		return Song{}, false
	}
	return songs[m.qc.pos], true
}

// handleCommonKey handles the keys of both views and reports whether it did.
func (m *browser) handleCommonKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c", "q":
		return tea.Quit, true
	case "tab":
		m.filter.SetValue("")
		m.err, m.status = "", ""
		if m.view == viewQueue {
			m.view = viewLibrary
			return nil, true
		}
		m.view = viewQueue
		m.loading = true
		return m.queueCmd(), true
	case "esc":
		if m.filter.Value() != "" {
			m.filter.SetValue("")
			m.resetCursor()
		}
		return nil, true
	case "/":
		m.filterOn = true
		return m.filter.Focus(), true
	case "space":
		return m.actionCmd(func(ctx context.Context, c *Client) (string, error) {
			state, err := c.Toggle(ctx)
			return stateLabel(state), err
		}), true
	case "c":
		if !m.confirmClear {
			m.confirmClear = true
			m.err, m.status = "", "press c again to clear the queue"
			if m.playing != nil {
				m.status += fmt.Sprintf(" (%d song(s))", m.playing.QueueLength)
			}
			return nil, true
		}
		m.confirmClear = false
		return m.actionCmd(func(ctx context.Context, c *Client) (string, error) {
			return "cleared the queue", c.Clear(ctx)
		}), true
	}
	return nil, false
}

func (m *browser) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if cmd, ok := m.handleCommonKey(msg); ok {
		return m, cmd
	}
	entries := m.filteredEntries()

	switch msg.String() {
	case "up", "k":
		m.lib.up()
	case "down", "j":
		m.lib.down(len(entries))
	case "g", "home":
		m.lib = listCursor{}
	case "G", "end":
		m.lib.pos = tui.Max0(len(entries) - 1)
	case "r":
		m.loading = true
		return m, m.listCmd(m.dir, "")
	case "backspace", "h", "left", "-":
		if m.dir == "" {
			return m, nil
		}
		m.loading = true
		parent := path.Dir(m.dir)
		if parent == "." {
			parent = ""
		}
		return m, m.listCmd(parent, m.dir)
	case "enter", "right", "l":
		e, ok := m.selected()
		if !ok {
			return m, nil
		}
		if e.IsDir {
			m.loading = true
			return m, m.listCmd(e.Path, "")
		}
		return m, m.actionCmd(func(ctx context.Context, c *Client) (string, error) {
			return "playing " + e.Name(), c.AddAndPlay(ctx, e.Path)
		})
	case "a":
		e, ok := m.selected()
		if !ok {
			return m, nil
		}
		return m, m.actionCmd(func(ctx context.Context, c *Client) (string, error) {
			return "added " + e.Name() + " to the queue", c.Add(ctx, e.Path)
		})
	}
	return m, nil
}

func (m *browser) handleQueueKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if cmd, ok := m.handleCommonKey(msg); ok {
		return m, cmd
	}
	songs := m.filteredQueue()

	switch msg.String() {
	case "up", "k":
		m.qc.up()
	case "down", "j":
		m.qc.down(len(songs))
	case "g", "home":
		m.qc = listCursor{}
	case "G", "end":
		m.qc.pos = tui.Max0(len(songs) - 1)
	case "r":
		m.loading = true
		return m, m.queueCmd()
	case "enter", "right", "l":
		s, ok := m.selectedSong()
		if !ok {
			return m, nil
		}
		return m, m.actionCmd(func(ctx context.Context, c *Client) (string, error) {
			return "playing " + s.DisplayName(), c.PlayID(ctx, s.ID)
		})
	case "d", "x", "delete":
		s, ok := m.selectedSong()
		if !ok {
			return m, nil
		}
		return m, m.actionCmd(func(ctx context.Context, c *Client) (string, error) {
			return "removed " + s.DisplayName() + " from the queue", c.DeleteID(ctx, s.ID)
		})
	case "K", "shift+up":
		return m, m.moveCmd(-1)
	case "J", "shift+down":
		return m, m.moveCmd(1)
	}
	return m, nil
}

// moveCmd moves the selected song by delta positions in the queue.
func (m *browser) moveCmd(delta int) tea.Cmd {
	if m.filter.Value() != "" {
		m.status = "clear the filter to reorder the queue"
		return nil
	}
	s, ok := m.selectedSong()
	if !ok {
		return nil
	}
	to := s.Pos + delta
	if to < 0 || to >= len(m.queue) {
		return nil
	}
	m.focusID = s.ID
	return m.actionCmd(func(ctx context.Context, c *Client) (string, error) {
		return "", c.MoveID(ctx, s.ID, to)
	})
}

// ---------------------------------------------------------------------------
// view
// ---------------------------------------------------------------------------

func stateLabel(state State) string {
	switch state {
	case StatePlay:
		return "playing"
	case StatePause:
		return "paused"
	}
	return "stopped"
}

// StateIcon returns the icon for state, empty if stopped.
func StateIcon(state State) string {
	return map[State]string{StatePlay: "▶", StatePause: "⏸"}[state]
}

// NowPlaying renders a one line summary of the player state.
func NowPlaying(status *Status, song *Song) string {
	if status == nil {
		return ""
	}
	icon := StateIcon(status.State)
	if icon == "" || song == nil {
		return fmt.Sprintf("■ stopped · queue %d", status.QueueLength)
	}
	progress := FormatDuration(status.Elapsed)
	if status.Duration > 0 {
		progress += " / " + FormatDuration(status.Duration)
	}
	return fmt.Sprintf("%s %s · %s · %d/%d", icon, song.DisplayName(), progress, status.Song+1, status.QueueLength)
}

// row is a rendered list entry.
type row struct {
	text  string
	style func(...string) string
}

func (m *browser) View() tea.View {
	var b strings.Builder

	b.WriteString(tui.TitleStyle.Render(" mpd "))
	b.WriteString(" ")
	if m.view == viewQueue {
		b.WriteString(tui.MetaStyle.Render("library · ") + tui.PathStyle.Render("queue"))
	} else {
		b.WriteString(tui.PathStyle.Render("/"+m.dir) + tui.MetaStyle.Render(" · queue"))
	}
	if m.loading {
		b.WriteString(" ")
		b.WriteString(tui.MetaStyle.Render("· loading"))
	}
	b.WriteString("\n")
	b.WriteString(tui.InfoStyle.Render(NowPlaying(m.playing, m.song)))
	b.WriteString("\n")

	if m.filterOn || m.filter.Value() != "" {
		b.WriteString(m.filter.View())
		b.WriteString("\n")
	}
	b.WriteString("\n")

	visible := max(5, m.height-8)
	var summary, help string
	if m.view == viewQueue {
		songs := m.filteredQueue()
		rows := make([]row, len(songs))
		current, icon := -1, ""
		if m.playing != nil {
			current, icon = m.playing.Song, StateIcon(m.playing.State)
		}
		for i, s := range songs {
			marker := "  "
			style := tui.LeafStyle.Render
			if s.Pos == current && icon != "" {
				marker = icon + " "
				style = tui.InfoStyle.Render
			}
			text := fmt.Sprintf("%s%3d  %s", marker, s.Pos+1, s.DisplayName())
			if s.Duration > 0 {
				text += tui.MetaStyle.Render(" " + FormatDuration(s.Duration))
			}
			rows[i] = row{text: text, style: style}
		}
		m.renderRows(&b, rows, &m.qc, visible, "  (the queue is empty)\n")
		summary = fmt.Sprintf("  %d song(s) · %s", len(songs), FormatDuration(totalDuration(songs)))
		help = "\n  ↑/↓ move  ·  enter play  ·  d remove  ·  J/K move down/up  ·  space play/pause  ·  c c clear  ·  tab library  ·  / filter  ·  q quit"
	} else {
		entries := m.filteredEntries()
		rows := make([]row, len(entries))
		for i, e := range entries {
			text := e.Name()
			style := tui.LeafStyle.Render
			if e.IsDir {
				text += "/"
				style = tui.DirStyle.Render
			} else if e.Song.Duration > 0 {
				text += tui.MetaStyle.Render(" " + FormatDuration(e.Song.Duration))
			}
			rows[i] = row{text: text, style: style}
		}
		m.renderRows(&b, rows, &m.lib, visible, "  (empty)\n")
		dirs, songs := countEntries(entries)
		summary = fmt.Sprintf("  %d folder(s) · %d song(s)", dirs, songs)
		help = "\n  ↑/↓ move  ·  enter open/play  ·  a add  ·  space play/pause  ·  c c clear queue  ·  h/← up  ·  tab queue  ·  / filter  ·  q quit"
	}

	b.WriteString("\n")
	b.WriteString(tui.HelpStyle.Render(summary))
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString(tui.ErrorStyle.Render("✗ " + m.err))
		b.WriteString("\n")
	} else if m.status != "" {
		b.WriteString(tui.StatusStyle.Render("· " + m.status))
		b.WriteString("\n")
	}
	b.WriteString(tui.HelpStyle.Render(help))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m *browser) renderRows(b *strings.Builder, rows []row, c *listCursor, visible int, empty string) {
	if len(rows) == 0 {
		if m.filter.Value() != "" {
			empty = "  (no matches)\n"
		}
		b.WriteString(tui.HelpStyle.Render(empty))
		return
	}
	c.scroll(visible, len(rows))
	end := min(c.offset+visible, len(rows))
	if c.offset > 0 {
		b.WriteString(tui.HelpStyle.Render(fmt.Sprintf("  ↑ %d more\n", c.offset)))
	}
	for i := c.offset; i < end; i++ {
		if i == c.pos {
			b.WriteString(tui.SelectedStyle.Render("▸ " + rows[i].text))
		} else {
			b.WriteString("  " + rows[i].style(rows[i].text))
		}
		b.WriteString("\n")
	}
	if end < len(rows) {
		b.WriteString(tui.HelpStyle.Render(fmt.Sprintf("  ↓ %d more\n", len(rows)-end)))
	}
}

func countEntries(entries []Entry) (dirs, songs int) {
	for _, e := range entries {
		if e.IsDir {
			dirs++
		} else {
			songs++
		}
	}
	return
}

func totalDuration(songs []Song) time.Duration {
	var total time.Duration
	for _, s := range songs {
		total += s.Duration
	}
	return total
}
