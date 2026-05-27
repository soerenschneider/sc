package cmd

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var vaultSecretEditCmd = &cobra.Command{
	Use:   "edit <path>",
	Short: "Interactively edit a Vault KV2 secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)
		return RunWithStore(client, vaultKv2EditOptions{
			Mount:   mount,
			Path:    args[0],
			Timeout: vaultDefaultTimeout,
		})
	},
}

func init() {
	vaultSecretCmd.AddCommand(vaultSecretEditCmd)
}

type vaultKv2EditOptions struct {
	Mount   string
	Path    string
	Timeout time.Duration
}

func RunWithStore(store vault.SecretStore, opts vaultKv2EditOptions) error {
	if store == nil {
		return errors.New("nil SecretStore")
	}
	if opts.Mount == "" || opts.Path == "" {
		return errors.New("mount and path are required")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}

	m := newModel(store, opts)
	if err := m.load(); err != nil {
		return err
	}
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

// ---------------------------------------------------------------------------
// model
// ---------------------------------------------------------------------------

type mode int

const (
	modeList mode = iota
	modeFormPair
	modeConfirmDeletePair
	modeConfirmSoftDelete
	modeConfirmQuit
	modeConfirmReload
	modeVersionPicker
	modeConfirmLoadVersion
)

type pair struct {
	key   string
	value string
}

type model struct {
	store vault.SecretStore
	opts  vaultKv2EditOptions

	pairs     []pair
	cursor    int
	revealed  map[string]bool
	revealAll bool

	// Versioning. `version` is what we're showing; `currentVersion` is
	// what the server says is latest (used as the CAS expectation). When
	// they differ, we're in historical view.
	version        int
	updatedAt      time.Time
	currentVersion int
	exists         bool
	metadata       []vault.VersionInfo // cached, refreshed each time the picker opens
	pendingVersion int                 // selected in the picker, waiting to load

	mode mode
	form *huh.Form

	editKey      string
	editVal      string
	editingIndex int
	confirm      bool

	dirty  bool
	saving bool
	status string
	errMsg string

	width, height int
}

func newModel(store vault.SecretStore, opts vaultKv2EditOptions) *model {
	return &model{
		store:    store,
		opts:     opts,
		revealed: map[string]bool{},
	}
}

func (m *model) historical() bool {
	return m.exists && m.currentVersion > 0 && m.version != m.currentVersion
}

// ---------------------------------------------------------------------------
// store I/O
// ---------------------------------------------------------------------------

// loadResult is the outcome of resolveSecret: either a readable secret (sec
// set), a metadata-only fallback (sec nil, md set, secret has history but no
// readable version), or a genuinely-missing secret (both nil).
type loadResult struct {
	sec        *vault.Secret
	md         *vault.SecretMetadata
	historical bool // sec is set, but is not the current version
}

// resolveSecret reads the latest version of a secret. If the latest version
// is missing because it has been soft-deleted (or destroyed), it falls back
// to GetMetadata and returns the most recent readable historical version.
// This is what makes the editor useful for browsing back to a soft-deleted
// secret: instead of showing "(new)" we show the last good version with a
// clear note, and allow saving to roll it forward.
func resolveSecret(ctx context.Context, store vault.SecretStore, mount, path string) (*loadResult, error) {
	sec, err := store.Get(ctx, mount, path)
	if err == nil {
		return &loadResult{sec: sec}, nil
	}
	if !vault.IsNotFound(err) {
		return nil, err
	}

	md, mderr := store.GetMetadata(ctx, mount, path)
	if mderr != nil {
		if vault.IsNotFound(mderr) {
			// No metadata either — genuinely a fresh path.
			return &loadResult{}, nil
		}
		return nil, mderr
	}

	// Latest version is unreadable but the secret has history. Walk
	// newest-first looking for a readable version.
	for _, v := range md.Versions {
		if !v.Readable() {
			continue
		}
		hist, herr := store.GetVersion(ctx, mount, path, v.Version)
		if herr != nil {
			if vault.IsNotFound(herr) {
				continue // raced with another delete; try the next one
			}
			return nil, herr
		}
		return &loadResult{sec: hist, md: md, historical: true}, nil
	}

	// Metadata exists but nothing is readable.
	return &loadResult{md: md}, nil
}

func (m *model) load() error {
	ctx, cancel := context.WithTimeout(context.Background(), m.opts.Timeout)
	defer cancel()
	res, err := resolveSecret(ctx, m.store, m.opts.Mount, m.opts.Path)
	if err != nil {
		return fmt.Errorf("reading %s/%s: %w", m.opts.Mount, m.opts.Path, err)
	}
	m.applyLoadResult(res)
	return nil
}

// applyLoadResult writes a loadResult into the model and sets an appropriate
// status message for the special cases (new / all-deleted / latest-deleted).
// For the normal case it leaves status untouched so the caller can decide
// (e.g. "reloaded · v5" for the r-keypress path).
func (m *model) applyLoadResult(res *loadResult) {
	m.dirty = false

	switch {
	case res.sec == nil && res.md == nil:
		// Path doesn't exist at all.
		m.pairs = nil
		m.version = 0
		m.currentVersion = 0
		m.updatedAt = time.Time{}
		m.exists = false
		m.status = "new secret — nothing saved yet"

	case res.sec == nil:
		// Metadata exists but every version is deleted or destroyed.
		m.pairs = nil
		m.version = 0
		m.updatedAt = time.Time{}
		m.currentVersion = res.md.CurrentVersion
		m.exists = true
		m.metadata = res.md.Versions
		m.status = fmt.Sprintf("all versions through v%d are deleted or destroyed — saving will create v%d",
			res.md.CurrentVersion, res.md.CurrentVersion+1)

	case res.historical:
		// Latest is soft-deleted; we loaded the most recent readable version.
		m.currentVersion = res.md.CurrentVersion
		m.metadata = res.md.Versions
		m.applyServerSecret(res.sec, false)
		m.status = fmt.Sprintf("latest v%d is deleted — showing v%d (save to roll forward to v%d)",
			res.md.CurrentVersion, res.sec.Version, res.md.CurrentVersion+1)

	default:
		// Normal: latest version, readable.
		m.applyServerSecret(res.sec, true)
	}
}

// applyServerSecret loads a *Secret into the model. When updateCurrent is
// true, the secret is treated as the latest version (Get/Put result); when
// false, it's treated as a historical read that must not move currentVersion.
func (m *model) applyServerSecret(sec *vault.Secret, updateCurrent bool) {
	m.pairs = m.pairs[:0]
	for k, v := range sec.Data {
		m.pairs = append(m.pairs, pair{key: k, value: v})
	}
	sort.Slice(m.pairs, func(i, j int) bool { return m.pairs[i].key < m.pairs[j].key })
	m.version = sec.Version
	m.updatedAt = sec.UpdatedAt
	if updateCurrent {
		m.currentVersion = sec.Version
	}
	m.exists = true
	m.dirty = false
}

func (m *model) save() tea.Cmd {
	data := make(map[string]string, len(m.pairs))
	for _, p := range m.pairs {
		if p.key == "" {
			continue
		}
		data[p.key] = p.value
	}
	// CAS uses currentVersion, never the viewed version. Saving from a
	// historical view creates a fresh version on top of the latest.
	store, mount, path, cas, timeout := m.store, m.opts.Mount, m.opts.Path, m.currentVersion, m.opts.Timeout
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		sec, err := store.Put(ctx, mount, path, data, cas)
		if err != nil {
			return errMsg{err: err, op: "save"}
		}
		return savedMsg{sec: sec}
	}
}

func (m *model) softDelete() tea.Cmd {
	store, mount, path, timeout := m.store, m.opts.Mount, m.opts.Path, m.opts.Timeout
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := store.SoftDelete(ctx, mount, path); err != nil {
			return errMsg{err: err, op: "delete"}
		}
		return deletedMsg{}
	}
}

func (m *model) reload() tea.Cmd {
	store, mount, path, timeout := m.store, m.opts.Mount, m.opts.Path, m.opts.Timeout
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		res, err := resolveSecret(ctx, store, mount, path)
		if err != nil {
			return errMsg{err: err, op: "reload"}
		}
		return reloadedMsg{res: res}
	}
}

func (m *model) fetchMetadata() tea.Cmd {
	store, mount, path, timeout := m.store, m.opts.Mount, m.opts.Path, m.opts.Timeout
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		md, err := store.GetMetadata(ctx, mount, path)
		if err != nil {
			return errMsg{err: err, op: "get-metadata"}
		}
		return metadataMsg{md: md}
	}
}

func (m *model) loadVersion(version int) tea.Cmd {
	store, mount, path, timeout := m.store, m.opts.Mount, m.opts.Path, m.opts.Timeout
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		sec, err := store.GetVersion(ctx, mount, path, version)
		if err != nil {
			return errMsg{err: err, op: fmt.Sprintf("load v%d", version)}
		}
		return versionLoadedMsg{sec: sec}
	}
}

type savedMsg struct{ sec *vault.Secret }
type deletedMsg struct{}
type reloadedMsg struct{ res *loadResult }
type metadataMsg struct{ md *vault.SecretMetadata }
type versionLoadedMsg struct{ sec *vault.Secret }
type errMsg struct {
	err error
	op  string
}

// ---------------------------------------------------------------------------
// bubbletea
// ---------------------------------------------------------------------------

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case errMsg:
		m.saving = false
		m.errMsg = humanizeError(msg.err, msg.op)
		m.status = ""
		return m, nil

	case savedMsg:
		m.saving = false
		m.errMsg = ""
		if msg.sec != nil {
			m.applyServerSecret(msg.sec, true)
		}
		m.metadata = nil // history is stale
		m.status = fmt.Sprintf("saved · now at v%d", m.version)
		return m, nil

	case deletedMsg:
		m.status = "secret soft-deleted (recoverable out-of-band)"
		return m, tea.Quit

	case reloadedMsg:
		m.saving = false
		m.errMsg = ""
		m.metadata = nil
		m.revealed = map[string]bool{}
		m.applyLoadResult(msg.res)
		if m.cursor >= len(m.pairs) {
			m.cursor = max0(len(m.pairs) - 1)
		}
		// applyLoadResult sets a useful status for the special cases
		// (new / all-deleted / latest-deleted). For a vanilla reload of
		// a live secret we overwrite with a confirmation.
		if msg.res.sec != nil && !msg.res.historical {
			m.status = fmt.Sprintf("reloaded · v%d", m.version)
		}
		return m, nil

	case metadataMsg:
		m.saving = false
		m.errMsg = ""
		m.status = ""
		m.metadata = msg.md.Versions
		// keep currentVersion in sync with the server's view
		if msg.md.CurrentVersion > m.currentVersion {
			m.currentVersion = msg.md.CurrentVersion
		}
		return m, m.openVersionPicker()

	case versionLoadedMsg:
		m.saving = false
		m.errMsg = ""
		if msg.sec != nil {
			m.applyServerSecret(msg.sec, false)
		}
		m.revealed = map[string]bool{}
		if m.cursor >= len(m.pairs) {
			m.cursor = max0(len(m.pairs) - 1)
		}
		if m.historical() {
			m.status = fmt.Sprintf("viewing v%d (current is v%d) · saving creates v%d",
				m.version, m.currentVersion, m.currentVersion+1)
		} else {
			m.status = fmt.Sprintf("loaded · v%d", m.version)
		}
		return m, nil
	}

	if m.form != nil {
		// Esc cancels any open form — pair editor, confirm, version
		// picker — without running handleFormDone. In-progress edits
		// are discarded. (huh handles ctrl+c via the StateAborted path
		// below; this just adds esc as the natural cancel key.)
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc" {
			m.form = nil
			m.mode = modeList
			return m, nil
		}
		f, cmd := m.form.Update(msg)
		if ff, ok := f.(*huh.Form); ok {
			m.form = ff
		}
		switch m.form.State {
		case huh.StateCompleted:
			finishedMode := m.mode
			m.form = nil
			m.mode = modeList
			next := m.handleFormDone(finishedMode)
			return m, tea.Batch(cmd, next)
		case huh.StateAborted:
			m.form = nil
			m.mode = modeList
			return m, cmd
		}
		return m, cmd
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		return m.handleListKey(km)
	}
	return m, nil
}

func (m *model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		if m.cursor < len(m.pairs)-1 {
			m.cursor++
		}
		return m, nil
	case "g", "home":
		m.cursor = 0
		return m, nil
	case "G", "end":
		if len(m.pairs) > 0 {
			m.cursor = len(m.pairs) - 1
		}
		return m, nil
	case " ":
		if len(m.pairs) > 0 {
			k := m.pairs[m.cursor].key
			m.revealed[k] = !m.revealed[k]
		}
		return m, nil
	case "T":
		m.revealAll = !m.revealAll
		return m, nil
	}

	if m.saving {
		m.status = "busy — wait for the in-flight request to finish"
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		if m.dirty {
			return m, m.openConfirm(modeConfirmQuit,
				"Discard unsaved changes?",
				"You have unsaved changes that will be lost.",
				"Discard", "Cancel")
		}
		return m, tea.Quit

	case "enter", "e":
		if len(m.pairs) > 0 {
			return m, m.openPairForm(m.cursor)
		}

	case "a", "n":
		return m, m.openPairForm(-1)

	case "d":
		if len(m.pairs) > 0 {
			return m, m.openConfirm(modeConfirmDeletePair,
				fmt.Sprintf("Delete key %q?", m.pairs[m.cursor].key),
				"Removes the pair locally — press 's' afterwards to persist.",
				"Delete", "Cancel")
		}

	case "s":
		// A save makes sense when either:
		//   - the user has unsaved edits, OR
		//   - the user is viewing a historical version that still has
		//     data — pressing 's' there rolls that content forward into
		//     a new version, which is the natural recovery path for a
		//     soft-deleted secret.
		canSave := m.dirty || (m.historical() && len(m.pairs) > 0)
		if !canSave {
			m.status = "nothing to save"
			return m, nil
		}
		m.saving = true
		m.status = "saving…"
		m.errMsg = ""
		return m, m.save()

	case "r":
		if m.dirty {
			return m, m.openConfirm(modeConfirmReload,
				"Reload from server?",
				"Your unsaved local changes will be discarded.",
				"Reload", "Cancel")
		}
		m.saving = true
		m.status = "reloading…"
		m.errMsg = ""
		return m, m.reload()

	case "v":
		if !m.exists {
			m.status = "no history yet — save first"
			return m, nil
		}
		m.saving = true
		m.status = "loading history…"
		m.errMsg = ""
		return m, m.fetchMetadata()

	case "X":
		if !m.exists {
			m.status = "nothing to delete — secret has never been saved"
			return m, nil
		}
		return m, m.openConfirm(modeConfirmSoftDelete,
			"Soft-delete this secret?",
			fmt.Sprintf("Marks v%d of %s/%s as deleted. Recoverable out-of-band.",
				m.currentVersion, m.opts.Mount, m.opts.Path),
			"Soft-delete", "Cancel")
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// forms
// ---------------------------------------------------------------------------

func (m *model) openPairForm(index int) tea.Cmd {
	if index >= 0 {
		m.editKey = m.pairs[index].key
		m.editVal = m.pairs[index].value
	} else {
		m.editKey = ""
		m.editVal = ""
	}
	m.editingIndex = index

	heading := "Edit pair"
	if index < 0 {
		heading = "Add pair"
	}
	desc := fmt.Sprintf("%s/%s @ v%d  ·  esc to cancel",
		m.opts.Mount, m.opts.Path, m.version)
	if m.historical() {
		desc += fmt.Sprintf("  ·  historical — saving creates v%d", m.currentVersion+1)
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(heading).Description(desc),
			huh.NewInput().
				Title("Key").
				Value(&m.editKey).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("key cannot be empty")
					}
					for i, p := range m.pairs {
						if i == m.editingIndex {
							continue
						}
						if p.key == s {
							return fmt.Errorf("key %q already exists", s)
						}
					}
					return nil
				}),
			huh.NewText().
				Title("Value").
				Value(&m.editVal),
		),
	).WithShowHelp(true).WithTheme(huh.ThemeCharm())

	m.mode = modeFormPair
	return m.form.Init()
}

func (m *model) openConfirm(target mode, title, desc, yes, no string) tea.Cmd {
	m.confirm = false
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(desc).
				Affirmative(yes).
				Negative(no).
				Value(&m.confirm),
		),
	).WithShowHelp(true).WithTheme(huh.ThemeCharm())
	m.mode = target
	return m.form.Init()
}

// openVersionPicker turns the cached metadata into a huh.Select form.
func (m *model) openVersionPicker() tea.Cmd {
	if len(m.metadata) == 0 {
		m.status = "no versions to show"
		return nil
	}
	options := make([]huh.Option[int], 0, len(m.metadata))
	for _, v := range m.metadata {
		label := fmt.Sprintf("v%-4d  %s  (%s)",
			v.Version,
			v.CreatedAt.Local().Format("2006-01-02 15:04"),
			humanizeAge(v.CreatedAt))
		switch {
		case v.Destroyed:
			label += "  · destroyed"
		case !v.DeletedAt.IsZero():
			label += "  · deleted"
		case v.Version == m.currentVersion:
			label += "  · current"
		case v.Version == m.version:
			label += "  · viewing"
		}
		options = append(options, huh.NewOption(label, v.Version))
	}

	m.pendingVersion = m.version
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Version history").
				Description(fmt.Sprintf("%s/%s — pick a version to load  ·  esc to cancel", m.opts.Mount, m.opts.Path)).
				Options(options...).
				Height(min(15, len(options)+2)).
				Value(&m.pendingVersion),
		),
	).WithShowHelp(true).WithTheme(huh.ThemeCharm())

	m.mode = modeVersionPicker
	return m.form.Init()
}

func (m *model) handleFormDone(prev mode) tea.Cmd {
	switch prev {
	case modeFormPair:
		k := strings.TrimSpace(m.editKey)
		if k == "" {
			return nil
		}
		// True no-op detection: when editing an existing pair, if neither
		// the key nor the value changed, leave the dirty flag untouched so
		// we don't claim "modified" for a pair the user only inspected.
		// The dirty flag is global ("any unsaved change"), so we never
		// clear it here — we only avoid setting it spuriously.
		if m.editingIndex >= 0 {
			cur := m.pairs[m.editingIndex]
			if cur.key == k && cur.value == m.editVal {
				m.status = "no changes"
				return nil
			}
		}
		if m.editingIndex < 0 {
			m.pairs = append(m.pairs, pair{key: k, value: m.editVal})
		} else {
			oldKey := m.pairs[m.editingIndex].key
			if oldKey != k {
				delete(m.revealed, oldKey)
			}
			m.pairs[m.editingIndex].key = k
			m.pairs[m.editingIndex].value = m.editVal
		}
		sort.Slice(m.pairs, func(i, j int) bool { return m.pairs[i].key < m.pairs[j].key })
		for i, p := range m.pairs {
			if p.key == k {
				m.cursor = i
				break
			}
		}
		m.dirty = true
		m.status = "modified (unsaved)"

	case modeConfirmDeletePair:
		if m.confirm && len(m.pairs) > 0 {
			delete(m.revealed, m.pairs[m.cursor].key)
			m.pairs = append(m.pairs[:m.cursor], m.pairs[m.cursor+1:]...)
			if m.cursor >= len(m.pairs) && m.cursor > 0 {
				m.cursor--
			}
			m.dirty = true
			m.status = "pair removed (unsaved)"
		}

	case modeConfirmSoftDelete:
		if m.confirm {
			m.saving = true
			m.status = "deleting…"
			m.errMsg = ""
			return m.softDelete()
		}

	case modeConfirmReload:
		if m.confirm {
			m.saving = true
			m.status = "reloading…"
			m.errMsg = ""
			return m.reload()
		}

	case modeConfirmQuit:
		if m.confirm {
			return tea.Quit
		}

	case modeVersionPicker:
		// pendingVersion holds the selected version.
		if m.pendingVersion == 0 || m.pendingVersion == m.version {
			return nil
		}
		var sel *vault.VersionInfo
		for i := range m.metadata {
			if m.metadata[i].Version == m.pendingVersion {
				sel = &m.metadata[i]
				break
			}
		}
		if sel == nil {
			return nil
		}
		if !sel.Readable() {
			if sel.Destroyed {
				m.errMsg = fmt.Sprintf("v%d is destroyed and cannot be loaded", sel.Version)
			} else {
				m.errMsg = fmt.Sprintf("v%d is soft-deleted and cannot be loaded", sel.Version)
			}
			return nil
		}
		if m.dirty {
			return m.openConfirm(modeConfirmLoadVersion,
				fmt.Sprintf("Load v%d?", sel.Version),
				"Your unsaved local changes will be discarded.",
				"Load", "Cancel")
		}
		m.saving = true
		m.status = fmt.Sprintf("loading v%d…", sel.Version)
		m.errMsg = ""
		return m.loadVersion(sel.Version)

	case modeConfirmLoadVersion:
		if m.confirm && m.pendingVersion > 0 {
			m.saving = true
			m.status = fmt.Sprintf("loading v%d…", m.pendingVersion)
			m.errMsg = ""
			return m.loadVersion(m.pendingVersion)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// error humanization
// ---------------------------------------------------------------------------

func humanizeError(err error, op string) string {
	if err == nil {
		return ""
	}
	var se *vault.StoreError
	if errors.As(err, &se) {
		switch se.Kind {
		case vault.ErrKindCASConflict:
			return "version conflict — the secret changed on the server since you loaded it. " +
				"Press 'r' to reload (your local edits will be lost)."
		case vault.ErrKindNotFound:
			return fmt.Sprintf("%s: not found on server", op)
		case vault.ErrKindPermissionDenied:
			msg := se.Message
			if msg == "" {
				msg = "your token lacks permission for this operation"
			}
			return fmt.Sprintf("%s denied: %s", op, msg)
		case vault.ErrKindRateLimited:
			return fmt.Sprintf("%s rate-limited — retry in a moment", op)
		case vault.ErrKindServerError:
			return fmt.Sprintf("%s failed: backend error. %s", op, se.Message)
		case vault.ErrKindTimeout:
			return fmt.Sprintf("%s timed out", op)
		case vault.ErrKindNetwork:
			return fmt.Sprintf("%s network error: %s", op, se.Message)
		case vault.ErrKindBadRequest:
			return fmt.Sprintf("%s rejected: %s", op, se.Message)
		case vault.ErrKindUnknown:
			return fmt.Sprintf("%s failed: %s", op, se.Message)
		}
	}
	return fmt.Sprintf("%s failed: %v", op, err)
}

// ---------------------------------------------------------------------------
// view
// ---------------------------------------------------------------------------

func (m *model) View() string {
	if m.form != nil {
		return m.form.View()
	}

	var b strings.Builder

	// Title + path
	b.WriteString(titleStyle.Render(" vault edit "))
	b.WriteString(" ")
	b.WriteString(pathStyle.Render(fmt.Sprintf("%s/%s", m.opts.Mount, m.opts.Path)))
	if m.dirty {
		b.WriteString(" ")
		b.WriteString(dirtyStyle.Render("●"))
	}
	if m.saving {
		b.WriteString(" ")
		b.WriteString(metaStyle.Render("· busy"))
	}
	b.WriteString("\n")

	// Status bar: version + date. This is the line the user asked for.
	b.WriteString(m.versionStatusLine())
	b.WriteString("\n\n")

	if len(m.pairs) == 0 {
		b.WriteString(helpStyle.Render("  (empty — press 'a' to add a key/value pair)\n"))
	}
	for i, p := range m.pairs {
		marker := "  "
		k := keyStyle.Render(p.key)
		if i == m.cursor {
			marker = selectedStyle.Render("▸ ")
			k = selectedStyle.Render(p.key)
		}
		b.WriteString(fmt.Sprintf("%s%s %s %s\n",
			marker, k, helpStyle.Render("="), m.renderValue(p)))
	}

	b.WriteString("\n")
	switch {
	case m.errMsg != "":
		b.WriteString(errorStyle.Render("✗ " + m.errMsg))
		b.WriteString("\n")
	case m.status != "":
		b.WriteString(statusStyle.Render("· " + m.status))
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render(
		"\n  ↑/↓ move  ·  space reveal  ·  T reveal all  ·  enter edit  ·  a add" +
			"  ·  d delete  ·  s save  ·  r reload  ·  v history  ·  X soft-delete  ·  q quit",
	))
	return b.String()
}

// versionStatusLine renders the dedicated status bar that always shows the
// version + absolute date + relative age. When in historical view it also
// shows the current (latest) version for context.
func (m *model) versionStatusLine() string {
	if !m.exists {
		return "  " + dirtyStyle.Render("(new — not yet saved)")
	}
	if m.version == 0 {
		// Metadata exists but every version is unreadable.
		body := fmt.Sprintf("v1–v%d all deleted or destroyed", m.currentVersion)
		return "  " + historicalStyle.Render(body) + "  " + historicalTagStyle.Render("[no readable version]")
	}
	when := m.updatedAt.Local().Format("2006-01-02 15:04")
	age := humanizeAge(m.updatedAt)
	if m.historical() {
		body := fmt.Sprintf("v%d of %d  ·  %s  (%s)",
			m.version, m.currentVersion, when, age)
		return "  " + historicalStyle.Render(body) + "  " + historicalTagStyle.Render("[historical]")
	}
	body := fmt.Sprintf("v%d  ·  %s  (%s)", m.version, when, age)
	return "  " + metaStyle.Render(body)
}

func (m *model) renderValue(p pair) string {
	if m.revealAll || m.revealed[p.key] {
		return valueStyle.Render(p.value)
	}
	if p.value == "" {
		return maskedStyle.Render("(empty)")
	}
	n := len(p.value)
	if n > 12 {
		n = 12
	}
	return maskedStyle.Render(strings.Repeat("•", n))
}

func humanizeAge(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// ---------------------------------------------------------------------------
// shared styles
// ---------------------------------------------------------------------------

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
