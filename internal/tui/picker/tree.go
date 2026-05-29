package picker

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
)

// Generic prefix-tree TUI
// Two use modes:
//   - Standalone: TreeBrowser.Browse(ctx, startPrefix) runs the TUI in its
//     own tea.Program and returns when the user acts. Used by the Vault `ls`
//     command.
//   - Embedded: TreeBrowser.NewModel(startPrefix) returns a *TreeModel that
//     a parent bubbletea model can host as a sub-screen. The model never
//     calls tea.Quit itself; the parent polls TreeModel.Done() each tick and
//     swaps it out when the user is finished. Used by the editor's `E` key,
//     which keeps the editor running underneath while the user picks an fs
//     destination.

// TreeNode is one item in a tree.
type TreeNode struct {
	Name  string
	IsDir bool
}

// TreeProvider is the data source the browser walks.
type TreeProvider interface {
	Children(ctx context.Context, prefix string) ([]TreeNode, error)
}

// TreeAction is what the user did when they exited the browser.
type TreeAction int

const (
	// ActionQuit is used when user pressed q/esc/ctrl+c. Prefix/Name are not meaningful.
	ActionQuit TreeAction = iota
	// ActionOpen is used when user selected an existing leaf with enter.
	ActionOpen
	// ActionSaveAs is used when user picked a new name via 's' in a directory.
	ActionSaveAs
	// ActionCreate is used when user picked a new name via 'n' in a directory.
	ActionCreate
)

// TreeResult is the user's choice on exit.
type TreeResult struct {
	Action TreeAction
	Prefix string // directory the user was at
	Name   string // leaf they selected or typed (empty for ActionQuit)
}

type TreeBrowser struct {
	Provider    TreeProvider
	Title       string // shown in title bar, e.g. "vault ls", "export"
	LeafLabel   string // singular noun, e.g. "secret", "file"
	DirLabel    string // singular noun, e.g. "folder", "directory"
	AllowCreate bool   // enables 'n' key — create new leaf in current prefix
	AllowSaveAs bool   // enables 's' key — pick a destination name
	Timeout     time.Duration
}

// Browse runs the TUI as a standalone bubbletea program. Returns when the
// user acts. The caller can call Browse again with the returned Prefix to
// resume at the same location.
func (b *TreeBrowser) Browse(ctx context.Context, startPrefix string) (*TreeResult, error) {
	m, err := b.NewModel(startPrefix)
	if err != nil {
		return nil, err
	}
	shim := &standaloneShim{m: m}
	out, err := tea.NewProgram(shim, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}
	final := out.(*standaloneShim).m
	return final.Result(), nil
}

// NewModel constructs a TreeModel for embedding inside another bubbletea
// model. It loads the initial prefix synchronously so the first frame is
// consistent.
func (b *TreeBrowser) NewModel(startPrefix string) (*TreeModel, error) {
	m := newTreeModel(b, startPrefix)
	if err := m.loadInitial(); err != nil {
		return nil, err
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// model
// ---------------------------------------------------------------------------

// TreeModel is the bubbletea model. Exported so it can be embedded.
type TreeModel struct {
	cfg    *TreeBrowser
	prefix string

	items  []TreeNode
	cursor int
	offset int

	filter   textinput.Model
	filterOn bool

	loading bool
	err     string
	status  string

	form       *huh.Form
	newName    string
	formAction TreeAction

	// Output state — read by parent (embedded) or by the standalone shim.
	done   bool
	action TreeAction
	name   string

	width, height int
}

func newTreeModel(cfg *TreeBrowser, prefix string) *TreeModel {
	ti := textinput.New()
	ti.Placeholder = "filter"
	ti.Prompt = "/ "
	ti.CharLimit = 256
	return &TreeModel{
		cfg:    cfg,
		prefix: prefix,
		filter: ti,
	}
}

// Done reports whether the user has finished. Once true, Result() is valid
// and the model will not transition further on input.
func (m *TreeModel) Done() bool { return m.done }

// Result returns the user's choice. Only call when Done() is true.
func (m *TreeModel) Result() *TreeResult {
	return &TreeResult{
		Action: m.action,
		Prefix: m.prefix,
		Name:   m.name,
	}
}

func (m *TreeModel) loadInitial() error {
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.Timeout)
	defer cancel()
	items, err := m.cfg.Provider.Children(ctx, m.prefix)
	if err != nil {
		return err
	}
	m.items = sortTreeNodes(items)
	return nil
}

type treeListedMsg struct {
	prefix string
	items  []TreeNode
	err    error
}

func (m *TreeModel) loadCmd(prefix string) tea.Cmd {
	provider, timeout := m.cfg.Provider, m.cfg.Timeout
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		items, err := provider.Children(ctx, prefix)
		return treeListedMsg{prefix: prefix, items: sortTreeNodes(items), err: err}
	}
}

func (m *TreeModel) Init() tea.Cmd { return nil }

func (m *TreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		// Already finished — ignore further input so the parent can
		// safely defer the swap-out to its own tick.
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case treeListedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = vault.HumanizeError(msg.err, "list")
			return m, nil
		}
		m.err = ""
		m.prefix = msg.prefix
		m.items = msg.items
		m.cursor, m.offset = 0, 0
		m.filter.SetValue("")
		return m, nil
	}

	// Name-prompt form (create / save-as) owns input while open.
	if m.form != nil {
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc" {
			m.form = nil
			return m, nil
		}
		f, cmd := m.form.Update(msg)
		if ff, ok := f.(*huh.Form); ok {
			m.form = ff
		}
		switch m.form.State {
		case huh.StateCompleted:
			name := strings.TrimSpace(m.newName)
			m.form = nil
			if name != "" {
				m.action = m.formAction
				m.name = name
				m.done = true
			}
			return m, cmd
		case huh.StateAborted:
			m.form = nil
			return m, cmd
		}
		return m, cmd
	}

	// Filter input takes most keys when active.
	if m.filterOn {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "esc":
				m.filterOn = false
				m.filter.SetValue("")
				m.filter.Blur()
				m.cursor, m.offset = 0, 0
				return m, nil
			case "enter":
				m.filterOn = false
				m.filter.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		m.cursor, m.offset = 0, 0
		return m, cmd
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		return m.handleKey(km)
	}
	return m, nil
}

func (m *TreeModel) filteredItems() []TreeNode {
	q := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if q == "" {
		return m.items
	}
	out := make([]TreeNode, 0, len(m.items))
	for _, it := range m.items {
		if strings.Contains(strings.ToLower(it.Name), q) {
			out = append(out, it)
		}
	}
	return out
}

func (m *TreeModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.filteredItems()
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.action = ActionQuit
		m.done = true
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case "g", "home":
		m.cursor, m.offset = 0, 0
	case "G", "end":
		m.cursor = tui.Max0(len(items) - 1)
	case "/":
		m.filterOn = true
		return m, m.filter.Focus()
	case "r":
		m.loading = true
		m.status = "reloading…"
		return m, m.loadCmd(m.prefix)
	case "backspace", "h", "left", "-":
		if m.prefix == "" {
			return m, nil
		}
		m.loading = true
		return m, m.loadCmd(parentPrefix(m.prefix))
	case "enter", "right", "l":
		if len(items) == 0 {
			return m, nil
		}
		it := items[m.cursor]
		if it.IsDir {
			m.loading = true
			return m, m.loadCmd(m.prefix + it.Name + "/")
		}
		m.action = ActionOpen
		m.name = it.Name
		m.done = true
		return m, nil
	case "n":
		if m.cfg.AllowCreate {
			return m, m.openNamePrompt(ActionCreate, "New "+m.cfg.LeafLabel)
		}
	case "s":
		if m.cfg.AllowSaveAs {
			return m, m.openNamePrompt(ActionSaveAs, "Save here as")
		}
	}
	return m, nil
}

func (m *TreeModel) openNamePrompt(action TreeAction, title string) tea.Cmd {
	m.newName = ""
	m.formAction = action

	existing := map[string]bool{}
	for _, it := range m.items {
		if !it.IsDir {
			existing[it.Name] = true
		}
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(title).
				Description(fmt.Sprintf("in %s — esc to cancel", displayPrefix(m.prefix))),
			huh.NewInput().
				Title("Name").
				Value(&m.newName).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return errors.New("name cannot be empty")
					}
					if strings.Contains(s, "/") {
						return errors.New("name cannot contain '/'")
					}
					// ActionCreate forbids overwriting existing leaves;
					// ActionSaveAs permits it (caller will confirm).
					if action == ActionCreate && existing[s] {
						return fmt.Errorf("a %s named %q already exists here", m.cfg.LeafLabel, s)
					}
					return nil
				}),
		),
	).WithShowHelp(true).WithTheme(huh.ThemeCharm())
	return m.form.Init()
}

// ---------------------------------------------------------------------------
// view
// ---------------------------------------------------------------------------

func (m *TreeModel) View() string {
	if m.form != nil {
		return m.form.View()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(" " + m.cfg.Title + " "))
	b.WriteString(" ")
	b.WriteString(pathStyle.Render(displayPrefix(m.prefix)))
	if m.loading {
		b.WriteString(" ")
		b.WriteString(metaStyle.Render("· loading"))
	}
	b.WriteString("\n")

	if m.filterOn || m.filter.Value() != "" {
		b.WriteString(m.filter.View())
		b.WriteString("\n")
	}
	b.WriteString("\n")

	items := m.filteredItems()

	view := m.height - 6
	if view < 5 {
		view = 5
	}
	m.adjustOffset(view, len(items))

	if len(items) == 0 {
		if m.filter.Value() != "" {
			b.WriteString(helpStyle.Render("  (no matches)\n"))
		} else {
			b.WriteString(helpStyle.Render("  (empty)\n"))
		}
	}

	end := m.offset + view
	if end > len(items) {
		end = len(items)
	}
	if m.offset > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ↑ %d more\n", m.offset)))
	}
	for i := m.offset; i < end; i++ {
		it := items[i]
		marker := "  "
		var name string
		if it.IsDir {
			name = dirStyle.Render(it.Name + "/")
		} else {
			name = leafStyle.Render(it.Name)
		}
		if i == m.cursor {
			marker = selectedStyle.Render("▸ ")
			if it.IsDir {
				name = selectedStyle.Render(it.Name + "/")
			} else {
				name = selectedStyle.Render(it.Name)
			}
		}
		b.WriteString(marker)
		b.WriteString(name)
		b.WriteString("\n")
	}
	if end < len(items) {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ↓ %d more\n", len(items)-end)))
	}

	b.WriteString("\n")
	dirs, leaves := countTreeNodes(items)
	b.WriteString(helpStyle.Render(fmt.Sprintf("  %d %s · %d %s",
		dirs, plural(dirs, m.cfg.DirLabel),
		leaves, plural(leaves, m.cfg.LeafLabel))))
	if m.filter.Value() != "" {
		td, tl := countTreeNodes(m.items)
		b.WriteString(helpStyle.Render(fmt.Sprintf(" of %d/%d", td, tl)))
	}
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString(errorStyle.Render("✗ " + m.err))
		b.WriteString("\n")
	} else if m.status != "" {
		b.WriteString(statusStyle.Render("· " + m.status))
		b.WriteString("\n")
	}

	foot := "  ↑/↓ move  ·  enter open  ·  h/← up  ·  / filter  ·  r reload"
	if m.cfg.AllowCreate {
		foot += "  ·  n new"
	}
	if m.cfg.AllowSaveAs {
		foot += "  ·  s save as"
	}
	foot += "  ·  esc cancel"
	b.WriteString(helpStyle.Render("\n" + foot))
	return b.String()
}

func (m *TreeModel) adjustOffset(view, n int) {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+view {
		m.offset = m.cursor - view + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	if m.offset > tui.Max0(n-view) {
		m.offset = tui.Max0(n - view)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// standaloneShim wraps a TreeModel as a top-level bubbletea program, issuing
// tea.Quit when the model signals done.
type standaloneShim struct{ m *TreeModel }

func (s *standaloneShim) Init() tea.Cmd { return s.m.Init() }
func (s *standaloneShim) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newM, cmd := s.m.Update(msg)
	if tm, ok := newM.(*TreeModel); ok {
		s.m = tm
	}
	if s.m.Done() {
		return s, tea.Quit
	}
	return s, cmd
}
func (s *standaloneShim) View() string { return s.m.View() }

func sortTreeNodes(items []TreeNode) []TreeNode {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return items[i].Name < items[j].Name
	})
	return items
}

func countTreeNodes(items []TreeNode) (dirs, leaves int) {
	for _, it := range items {
		if it.IsDir {
			dirs++
		} else {
			leaves++
		}
	}
	return
}

func plural(n int, label string) string {
	if n == 1 {
		return label
	}
	return label + "s"
}

// parentPrefix returns the parent of p, or "" for root.
func parentPrefix(p string) string {
	p = strings.TrimSuffix(p, "/")
	if !strings.Contains(p, "/") {
		return ""
	}
	return p[:strings.LastIndex(p, "/")+1]
}

// normalizePrefix ensures a non-empty prefix ends with "/".
func normalizePrefix(p string) string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return ""
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

// displayPrefix returns a user-facing rendering of a prefix. Empty becomes
// "/" so the user always sees a slash.
func displayPrefix(p string) string {
	if p == "" {
		return "/"
	}
	return p
}

// ---------------------------------------------------------------------------
// styles (shared with the editor view)
// ---------------------------------------------------------------------------

var (
	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Bold(true)
	leafStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	historicalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	historicalTagStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("214")).
				Bold(true).
				Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	maskedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("76"))

	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
)
