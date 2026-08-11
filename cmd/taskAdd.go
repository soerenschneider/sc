package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var taskAddConfigPath string

var taskAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactively create a taskwarrior task",
	Long: `Interactive alternative to 'task add' with picker-driven project,
tag, and attribute selection.

Loads the taxonomy of known projects and tag categories from a YAML file
(default: ~/.sc/taskwarrior.yaml, override with --config). When the file
is absent, a small generic GTD taxonomy is used.

The Project step is a filterable picker with the configured
'default_project' pre-selected (falls back to "inbox" when the projects
list contains it). Press "/" to filter, esc/enter to exit filter mode,
↑↓ to navigate, enter to pick. To create a new project (or subproject
like "myproject.sub"), pick "＋ create new…" at the bottom of the list
and type the full dotted name in the follow-up input.

An optional Depends-on step lists pending tasks; "(none)" is the default,
so enter skips it. Pick a task to append depends:<ID> to the invocation.`,
	RunE: runTaskAdd,
}

func init() {
	taskAddCmd.Flags().StringVarP(&taskAddConfigPath, "config", "c", "",
		"Path to taxonomy YAML (default ~/.sc/taskwarrior.yaml; falls back to generic GTD)")
	// Wire taskCmd into your root command:
	//   rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskAddCmd)
}

// The label of the category whose members become the Location Select.
const locationCategoryLabel = "loc"

// sentinelNewProject is the Select value for the "＋ create new…" option
// in the Project picker. NUL-wrapped so it can't collide with a real
// project name. When selected, the follow-up Input group is revealed so
// the user can type a new project name.
const sentinelNewProject = "\x00__new__\x00"

// ---------------------------------------------------------------------------
// config
// ---------------------------------------------------------------------------

type taskConfig struct {
	Projects          []string            `yaml:"projects"`
	DefaultProject    string              `yaml:"default_project"`
	AllowEmptyProject bool                `yaml:"allow_empty_project"`
	TagCategories     []tagCategoryConfig `yaml:"tag_categories"`
}

type tagCategoryConfig struct {
	Label       string   `yaml:"label"`
	Members     []string `yaml:"members,omitempty"`
	MatchPrefix string   `yaml:"match_prefix,omitempty"`
}

func (c *taskConfig) categorize(tag string) string {
	for _, tc := range c.TagCategories {
		for _, m := range tc.Members {
			if tag == m {
				return tc.Label
			}
		}
		if tc.MatchPrefix != "" && strings.HasPrefix(tag, tc.MatchPrefix) {
			return tc.Label
		}
	}
	return "other"
}

func (c *taskConfig) seedTags() []string {
	var out []string
	for _, tc := range c.TagCategories {
		out = append(out, tc.Members...)
	}
	return out
}

// defaultProject returns the project the Project field is pre-filled with.
// Priority: explicit `default_project` from YAML → "inbox" if it's in the
// projects list → empty (user starts with a blank field).
func (c *taskConfig) defaultProject() string {
	if c.DefaultProject != "" {
		return c.DefaultProject
	}
	if slices.Contains(c.Projects, "inbox") {
		return "inbox"
	}
	return ""
}

func loadTaskConfig(explicit string) (*taskConfig, error) {
	path := explicit
	usingDefault := explicit == ""
	if usingDefault {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolving home directory: %w", err)
		}
		path = filepath.Join(home, ".sc", "taskwarrior.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && usingDefault {
			return defaultTaskConfig(), nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg taskConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

func defaultTaskConfig() *taskConfig {
	return &taskConfig{
		Projects: []string{
			"inbox", "personal", "work", "home", "errands",
		},
		DefaultProject:    "inbox",
		AllowEmptyProject: true,
		TagCategories: []tagCategoryConfig{
			{Label: "state", Members: []string{"next", "waiting", "someday", "quick", "deep"}},
			{
				Label:       "context",
				Members:     []string{"@computer", "@phone", "@home", "@online"},
				MatchPrefix: "@",
			},
			{Label: "energy", Members: []string{"lowenergy", "highenergy"}},
		},
	}
}

// ---------------------------------------------------------------------------
// entry point
// ---------------------------------------------------------------------------

type addOptions struct {
	Description  string
	Project      string
	Location     string
	Tags         []string
	Priority     string
	Size         string
	Due          string
	DependsID    string // task ID as string; "" means no dependency
	DependsLabel string // human label like "[5] Fix login bug", for preview
}

func runTaskAdd(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cfg, err := loadTaskConfig(taskAddConfigPath)
	if err != nil {
		return err
	}

	projects, err := taskProjects(ctx, cfg)
	if err != nil {
		return fmt.Errorf("fetching taskwarrior projects: %w", err)
	}
	tags, err := taskTags(ctx, cfg)
	if err != nil {
		return fmt.Errorf("fetching taskwarrior tags: %w", err)
	}
	tasks, err := taskList(ctx)
	if err != nil {
		return fmt.Errorf("fetching taskwarrior tasks for depends-on picker: %w", err)
	}

	opts, ok, err := runAddForm(projects, tags, tasks, cfg)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return confirmAndCreate(ctx, opts)
}

// ---------------------------------------------------------------------------
// taskwarrior I/O
// ---------------------------------------------------------------------------

func taskProjects(ctx context.Context, cfg *taskConfig) ([]string, error) {
	out, err := taskExec(ctx, "_projects")
	if err != nil {
		return nil, err
	}
	seed := append([]string{}, cfg.Projects...)
	return dedupe(append(seed, splitLines(out)...)), nil
}

func taskTags(ctx context.Context, cfg *taskConfig) ([]string, error) {
	out, err := taskExec(ctx, "_tags")
	if err != nil {
		return nil, err
	}
	fetched := splitLines(out)
	userTags := make([]string, 0, len(fetched))
	for _, t := range fetched {
		if t != strings.ToLower(t) {
			continue // virtual tag
		}
		userTags = append(userTags, t)
	}
	return dedupe(append(cfg.seedTags(), userTags...)), nil
}

// taskRef is a minimal projection of a taskwarrior task, enough to render
// a picker entry and shell out with depends:ID.
type taskRef struct {
	ID          int
	Description string
	Project     string
}

// taskList returns the pending tasks by shelling out to `task export
// status:pending`. Used to populate the Depends-on picker. Returned in
// ID-descending order (newest first) since recent tasks are more likely
// dependency candidates.
func taskList(ctx context.Context) ([]taskRef, error) {
	out, err := taskExec(ctx, "rc.verbose=nothing", "status:pending", "export")
	if err != nil {
		return nil, err
	}
	var raw []struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
		Project     string `json:"project"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing task export output: %w", err)
	}
	tasks := make([]taskRef, 0, len(raw))
	for _, r := range raw {
		if r.ID == 0 {
			continue // skip completed/deleted with id=0
		}
		tasks = append(tasks, taskRef{ID: r.ID, Description: r.Description, Project: r.Project})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID > tasks[j].ID })
	return tasks, nil
}

func taskExec(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "task", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		return out, nil
	}
	if execErr := new(exec.Error); errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
		return nil, errors.New("'task' not found in PATH, install taskwarrior first")
	}
	msg := strings.TrimSpace(stderr.String())
	if msg == "" {
		msg = err.Error()
	}
	return nil, fmt.Errorf("task %s: %s", strings.Join(args, " "), msg)
}

func splitLines(b []byte) []string {
	raw := strings.Split(string(b), "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// form
// ---------------------------------------------------------------------------

func runAddForm(projects, tags []string, tasks []taskRef, cfg *taskConfig) (addOptions, bool, error) {
	var (
		opts       addOptions
		newProject string
		newTagsRaw string
	)

	// Pre-fill the Project field with the configured default (or "inbox"
	// if the projects list contains it, or empty otherwise). When the
	// picker opens, the cursor lands on this pre-selection.
	opts.Project = cfg.defaultProject()

	locationTags, otherTags := splitLocationTags(tags, cfg)

	// Sort projects for a stable suggestion order.
	sort.Strings(projects)

	locationOpts := buildLocationOptions(locationTags)
	tagOpts := buildCategorizedTagOptions(otherTags, cfg)

	// Picker order:
	//   1. ＋ create new…                 — always first (escape hatch to
	//                                        typing a new project name)
	//   2. Default project                — from cfg.defaultProject().
	//                                        If the default resolves to
	//                                        empty AND allow_empty_project
	//                                        is true, (none) takes this
	//                                        slot as the default itself.
	//   3. (none)                         — only if allow_empty_project
	//                                        AND (none) isn't already at
	//                                        position 2 (avoid duplicate).
	//   4. Everything else alphabetically — the actual project list.
	//
	// opts.Project is pre-set to cfg.defaultProject() before the form
	// runs, so huh's cursor lands on whatever option holds that value —
	// either the real default at position 2, or the (none) at position 2
	// when empty is the default.
	def := cfg.defaultProject()
	defInList := def != "" && slices.Contains(projects, def)

	projectOpts := make([]huh.Option[string], 0, len(projects)+2)
	projectOpts = append(projectOpts, huh.NewOption("＋ create new…", sentinelNewProject))

	switch {
	case defInList:
		projectOpts = append(projectOpts, huh.NewOption(def, def))
	case cfg.AllowEmptyProject && def == "":
		// Empty is the default AND allowed — (none) is the default slot.
		projectOpts = append(projectOpts, huh.NewOption("(none)", ""))
	}

	if cfg.AllowEmptyProject && def != "" {
		projectOpts = append(projectOpts, huh.NewOption("(none)", ""))
	}

	for _, p := range projects {
		if p != def {
			projectOpts = append(projectOpts, huh.NewOption(p, p))
		}
	}

	// Depends-on picker options: "(none)" sentinel first (value ""),
	// then each pending task labelled "[ID] description (project)".
	depOpts := make([]huh.Option[string], 0, len(tasks)+1)
	depOpts = append(depOpts, huh.NewOption("(none)", ""))
	for _, t := range tasks {
		label := fmt.Sprintf("[%d] %s", t.ID, t.Description)
		if t.Project != "" {
			label += fmt.Sprintf(" (%s)", t.Project)
		}
		depOpts = append(depOpts, huh.NewOption(label, strconv.Itoa(t.ID)))
	}

	groups := []*huh.Group{
		// 1. Description.
		huh.NewGroup(
			huh.NewInput().
				Title("Description").
				Description("Required — the task itself").
				Placeholder("e.g. Ship the quarterly report").
				Value(&opts.Description).
				Validate(nonEmpty("description")),
		).Title("✎  New task"),

		// 2. Project picker. Filterable Select over the project list with
		// the configured default pre-selected; huh's built-in "/" key
		// activates filter mode (esc/enter exits filter). The last
		// option is a sentinel that reveals group 2b for typing a new
		// project name.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Project").
				Description("↑↓ navigate · enter picks · / filters (esc exits) · pick ＋ to create new").
				Options(projectOpts...).
				Value(&opts.Project).
				Height(pickerHeight(len(projectOpts))),
		).Title("📁  Project"),

		// 2b. New project name. Conditional Input — shown only when the
		// user picks "＋ create new…" in the picker above. Autocomplete
		// suggestions from the existing project list help with typing
		// subprojects (e.g. autocomplete "selfhost.k8s." then add a new
		// leaf). Post-processing below replaces the sentinel in
		// opts.Project with whatever the user types here.
		huh.NewGroup(
			huh.NewInput().
				Title("New project name").
				Description("Type · ↓↑ cycle suggestions · ctrl+e completes · . = subproject").
				Placeholder("e.g. work.acme · selfhost.k8s.cluster1").
				Suggestions(projects).
				Value(&newProject).
				Validate(nonEmpty("new project name")),
		).Title("📁  New project").WithHideFunc(func() bool {
			return opts.Project != sentinelNewProject
		}),

		// 3. Context & sizing.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Location").
				Description("Physical-access constraint · leave (none) if it doesn't matter").
				Options(locationOpts...).
				Value(&opts.Location),
			huh.NewSelect[string]().
				Title("Priority").
				Description("Importance").
				Options(priorityOptions()...).
				Value(&opts.Priority),
			huh.NewSelect[string]().
				Title("Size").
				Description("Effort estimate (requires uda.size in .taskrc)").
				Options(sizeOptions()...).
				Value(&opts.Size),
			huh.NewInput().
				Title("Due").
				Description("Anything taskwarrior parses — leave empty for none").
				Placeholder("e.g. tomorrow · friday · 2026-06-15 · +2d").
				Value(&opts.Due).
				Validate(validateDue),
		).Title("🎯  Context & sizing"),

		// 3b. Depends on. Optional single-task picker over pending
		// tasks. First option is a "(none)" sentinel with value "",
		// which is what opts.DependsID starts as — so the cursor lands
		// on "(none)" by default and enter accepts it with zero extra
		// keystrokes. Hidden entirely if the task database has no
		// pending tasks to depend on.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Depends on").
				Description("↑↓ navigate · enter picks · / filters · (none) for no dependency").
				Options(depOpts...).
				Value(&opts.DependsID).
				Height(pickerHeight(len(depOpts))),
		).Title("🔗  Depends on").WithHideFunc(func() bool {
			return len(tasks) == 0
		}),

		// 4. Tags.
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Tags").
				Description(`Grouped by purpose · space to toggle · "/" to filter`).
				Options(tagOpts...).
				Value(&opts.Tags).
				Height(pickerHeight(len(tagOpts))),
			huh.NewInput().
				Title("Additional tags").
				Description("Space-separated · for anything not in the list").
				Placeholder("e.g. urgent q3-review").
				Value(&newTagsRaw),
		).Title("🏷  Tags"),
	}

	form := huh.NewForm(groups...).
		WithShowHelp(true).
		WithTheme(huh.ThemeFunc(taskAddTheme))

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return opts, false, nil
		}
		return opts, false, err
	}

	// Post-process.
	if opts.Project == sentinelNewProject {
		opts.Project = strings.TrimSpace(newProject)
	}
	opts.Project = strings.TrimSpace(opts.Project)

	// Resolve DependsID to a human label for the preview.
	if opts.DependsID != "" {
		for _, t := range tasks {
			if strconv.Itoa(t.ID) == opts.DependsID {
				label := fmt.Sprintf("[%d] %s", t.ID, t.Description)
				if t.Project != "" {
					label += fmt.Sprintf(" (%s)", t.Project)
				}
				opts.DependsLabel = label
				break
			}
		}
	}
	if opts.Location != "" {
		opts.Tags = append([]string{opts.Location}, opts.Tags...)
	}
	for _, t := range strings.Fields(newTagsRaw) {
		t = strings.TrimPrefix(t, "+")
		if t != "" {
			opts.Tags = append(opts.Tags, t)
		}
	}
	opts.Tags = dedupe(opts.Tags)
	return opts, true, nil
}

func splitLocationTags(all []string, cfg *taskConfig) (locs, rest []string) {
	for _, t := range all {
		if cfg.categorize(t) == locationCategoryLabel {
			locs = append(locs, t)
		} else {
			rest = append(rest, t)
		}
	}
	return locs, rest
}

func buildLocationOptions(locs []string) []huh.Option[string] {
	sort.Strings(locs)
	opts := make([]huh.Option[string], 0, len(locs)+1)
	opts = append(opts, huh.NewOption("(none)", ""))
	for _, l := range locs {
		opts = append(opts, huh.NewOption(l, l))
	}
	return opts
}

func buildCategorizedTagOptions(tags []string, cfg *taskConfig) []huh.Option[string] {
	byCategory := map[string][]string{}
	for _, t := range tags {
		byCategory[cfg.categorize(t)] = append(byCategory[cfg.categorize(t)], t)
	}
	for _, ts := range byCategory {
		sort.Strings(ts)
	}

	const prefixWidth = 10
	pad := func(label string) string {
		return fmt.Sprintf("%-*s", prefixWidth, "["+label+"]")
	}

	opts := make([]huh.Option[string], 0, len(tags))
	for _, tc := range cfg.TagCategories {
		for _, t := range byCategory[tc.Label] {
			opts = append(opts, huh.NewOption(pad(tc.Label)+" "+t, t))
		}
	}
	for _, t := range byCategory["other"] {
		opts = append(opts, huh.NewOption(pad("other")+" "+t, t))
	}
	return opts
}

func priorityOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("(none)", ""),
		huh.NewOption("High", "H"),
		huh.NewOption("Medium", "M"),
		huh.NewOption("Low", "L"),
	}
}

func sizeOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("(none)", ""),
		huh.NewOption("XS · tiny", "xs"),
		huh.NewOption("S · small", "s"),
		huh.NewOption("M · medium", "m"),
		huh.NewOption("L · large", "l"),
		huh.NewOption("XL · huge", "xl"),
	}
}

func nonEmpty(name string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s cannot be empty", name)
		}
		return nil
	}
}

// validateDue checks whether taskwarrior can parse the input as a due
// date by running "task count due:INPUT" and inspecting the exit status.
// An empty string is allowed (no due date). Called by huh's Input.Validate
// on every attempt to advance the field — on error, huh shows the message
// inline and keeps focus so the user can fix it.
func validateDue(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "task", "rc.verbose=nothing",
		"rc.confirmation=no", "due:"+input, "count")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err == nil {
		return nil
	}

	// task's stderr is the useful diagnostic. Strip whitespace and take
	// the first line so it renders cleanly in a form validator inline.
	msg := strings.TrimSpace(stderr.String())
	if i := strings.IndexByte(msg, '\n'); i > 0 {
		msg = msg[:i]
	}
	if msg == "" {
		return fmt.Errorf("taskwarrior rejected %q as a due date", input)
	}
	return errors.New(msg)
}

func pickerHeight(items int) int {
	h := items + 2
	if h > 15 {
		h = 15
	}
	if h < 5 {
		h = 5
	}
	return h
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if seen[x] {
			continue
		}
		seen[x] = true
		out = append(out, x)
	}
	return out
}

// ---------------------------------------------------------------------------
// preview + confirm
// ---------------------------------------------------------------------------

var (
	taskHasDarkBg = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	taskLightDark = lipgloss.LightDark(taskHasDarkBg)

	previewAccentFg = taskLightDark(lipgloss.Color("#5f4dc2"), lipgloss.Color("#c0a3ff"))
	previewMutedFg  = taskLightDark(lipgloss.Color("#737373"), lipgloss.Color("#8a8a8a"))
	previewTagFg    = taskLightDark(lipgloss.Color("#0277bd"), lipgloss.Color("#5bc0eb"))

	previewIconStyle  = lipgloss.NewStyle().Foreground(previewAccentFg)
	previewLabelStyle = lipgloss.NewStyle().Foreground(previewMutedFg)
	previewTagStyle   = lipgloss.NewStyle().Foreground(previewTagFg)
)

// taskAddTheme derives from huh.ThemeCharm and adds a colored thick left
// border to the focused field. Blurred fields get a hidden border of the
// same width so cycling focus doesn't cause layout shift. Without this,
// the two-field Tags page is hard to read — huh's default focused vs.
// blurred colors are similar enough on many terminals that the active
// widget isn't obviously distinguishable.
func taskAddTheme(isDark bool) *huh.Styles {
	s := huh.ThemeCharm(isDark)

	// Charm's brand purple, adjusted for background contrast.
	accent := lipgloss.Color("#7571F9")
	if !isDark {
		accent = lipgloss.Color("#5A56E0")
	}

	s.Focused.Base = s.Focused.Base.
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(accent).
		PaddingLeft(1)

	s.Blurred.Base = s.Blurred.Base.
		BorderStyle(lipgloss.HiddenBorder()).
		BorderLeft(true).
		PaddingLeft(1)

	return s
}

func renderPreview(opts addOptions) string {
	const labelWidth = 9

	row := func(icon, label, value string) string {
		return previewIconStyle.Render(icon) + "  " +
			previewLabelStyle.Render(fmt.Sprintf("%-*s", labelWidth, label)) + "  " +
			value
	}

	var b strings.Builder
	b.WriteString(row("✎", "task", opts.Description) + "\n")
	if opts.Project != "" {
		b.WriteString(row("📁", "project", opts.Project) + "\n")
	}
	if len(opts.Tags) > 0 {
		parts := make([]string, len(opts.Tags))
		for i, t := range opts.Tags {
			parts[i] = previewTagStyle.Render("+" + t)
		}
		b.WriteString(row("🏷", "tags", strings.Join(parts, " ")) + "\n")
	}
	if opts.Priority != "" {
		b.WriteString(row("⚡", "priority", opts.Priority) + "\n")
	}
	if opts.Size != "" {
		b.WriteString(row("📏", "size", opts.Size) + "\n")
	}
	if opts.Due != "" {
		b.WriteString(row("⏰", "due", opts.Due) + "\n")
	}
	if opts.DependsLabel != "" {
		b.WriteString(row("🔗", "depends", opts.DependsLabel) + "\n")
	}
	return b.String()
}

func confirmAndCreate(ctx context.Context, opts addOptions) error {
	confirm := false
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create this task?").
				Description(renderPreview(opts)).
				Affirmative("Create").
				Negative("Cancel").
				Value(&confirm),
		).Title("✓  Confirm"),
	).WithTheme(huh.ThemeFunc(taskAddTheme))

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}
	if !confirm {
		return nil
	}
	return runTaskAddExec(ctx, buildTaskArgs(opts))
}

func buildTaskArgs(opts addOptions) []string {
	args := []string{"add", opts.Description}
	if opts.Project != "" {
		args = append(args, "project:"+opts.Project)
	}
	for _, t := range opts.Tags {
		args = append(args, "+"+t)
	}
	if opts.Priority != "" {
		args = append(args, "priority:"+opts.Priority)
	}
	if opts.Size != "" {
		args = append(args, "size:"+opts.Size)
	}
	if opts.Due != "" {
		args = append(args, "due:"+opts.Due)
	}
	if opts.DependsID != "" {
		args = append(args, "depends:"+opts.DependsID)
	}
	return args
}

func runTaskAddExec(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "task", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
