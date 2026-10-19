package mpd

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func key(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	}
	r := []rune(s)[0]
	return tea.KeyPressMsg{Code: r, Text: s}
}

// press sends a key and runs the resulting commands to completion. Only the
// browser's own messages are fed back, as e.g. the filter input's cursor
// blinks forever.
func press(t *testing.T, m *browser, keys ...string) {
	t.Helper()
	for _, k := range keys {
		_, cmd := m.Update(key(k))
		run(t, m, cmd)
	}
}

func run(t *testing.T, m *browser, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			run(t, m, c)
		}
	case listedMsg, queuedMsg, nowPlayingMsg, actionMsg:
		_, next := m.Update(msg)
		run(t, m, next)
	}
}

func newTestBrowser(t *testing.T, responses map[string]string) (*browser, *fakeServer) {
	t.Helper()
	s := newFakeServer(t, responses)
	m := newBrowser(context.Background(), dial(t, s), "")
	if err := m.loadInitial(); err != nil {
		t.Fatal(err)
	}
	m.height = 40
	return m, s
}

var library = map[string]string{
	`lsinfo ""`:       "file: loose.mp3\ndirectory: Artist\ndirectory: Other",
	`lsinfo "Artist"`: "directory: Artist/Album\nfile: Artist/song.flac\nArtist: Artist\nTitle: Song\nduration: 61",
	"status":          "state: play\nsong: 0\nplaylistlength: 2\nelapsed: 1\nduration: 61",
	"currentsong":     "file: Artist/song.flac\nArtist: Artist\nTitle: Song",
}

func withResponses(extra map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range library {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func names(entries []Entry) string {
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return strings.Join(out, ",")
}

func TestBrowseNavigation(t *testing.T) {
	m, _ := newTestBrowser(t, library)
	if got := names(m.entries); got != "Artist,Other,loose.mp3" {
		t.Fatalf("root = %s, want folders first", got)
	}

	press(t, m, "enter")
	if m.dir != "Artist" || names(m.entries) != "Album,Artist - Song" {
		t.Fatalf("after enter: dir %q, entries %s", m.dir, names(m.entries))
	}

	press(t, m, "h")
	if m.dir != "" {
		t.Fatalf("after h: dir %q", m.dir)
	}
	if e, _ := m.selected(); e.Path != "Artist" {
		t.Errorf("cursor on %q after going up, want the folder we came from", e.Path)
	}

	press(t, m, "j", "j", "j", "j")
	if e, _ := m.selected(); e.Path != "loose.mp3" {
		t.Errorf("cursor on %q, want it to stop at the last entry", e.Path)
	}
}

func TestBrowsePlayback(t *testing.T) {
	m, s := newTestBrowser(t, withResponses(map[string]string{
		`addid "Artist/song.flac"`: "Id: 5",
		`playid "5"`:               "",
		`add "Other"`:              "",
		`pause "1"`:                "",
		"clear":                    "",
	}))

	press(t, m, "enter", "j", "enter")
	if !strings.Contains(m.status, "playing Artist - Song") || m.err != "" {
		t.Errorf("after playing: status %q, err %q", m.status, m.err)
	}
	if m.song == nil || m.song.Title != "Song" {
		t.Errorf("now playing not refreshed: %+v", m.song)
	}

	press(t, m, "h", "j", "a")
	if m.status != "added Other to the queue" {
		t.Errorf("after add: status %q, err %q", m.status, m.err)
	}

	press(t, m, "space")
	if m.status != "paused" {
		t.Errorf("after space: status %q, err %q", m.status, m.err)
	}

	press(t, m, "c", "c")
	if m.status != "cleared the queue" {
		t.Errorf("after clear: status %q, err %q", m.status, m.err)
	}

	want := []string{`lsinfo ""`, `lsinfo "Artist"`, `addid "Artist/song.flac"`, `playid "5"`, "status", "currentsong",
		`lsinfo ""`, `add "Other"`, "status", "currentsong",
		"status", `pause "1"`, "status", "currentsong",
		"clear", "status", "currentsong"}
	if got := s.sent(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("sent:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestBrowseFilter(t *testing.T) {
	m, _ := newTestBrowser(t, library)
	// the filter input only returns cursor blinks, which take a while to run
	for _, k := range []string{"/", "o", "t", "h", "enter"} {
		m.Update(key(k))
	}
	if got := names(m.filteredEntries()); got != "Other" {
		t.Errorf("filtered = %s", got)
	}
	// keys go to the list again after enter
	press(t, m, "enter")
	if m.err == "" {
		t.Errorf("opening Other should fail as the fake server does not know it")
	}
	press(t, m, "esc")
	if got := names(m.filteredEntries()); got != "Artist,Other,loose.mp3" {
		t.Errorf("after esc = %s", got)
	}
}

func TestBrowseError(t *testing.T) {
	m, _ := newTestBrowser(t, withResponses(map[string]string{
		`add "Artist"`: "ACK [4@0] {add} Permission denied",
	}))
	press(t, m, "a")
	if m.err != "mpd: add: Permission denied" {
		t.Errorf("err = %q", m.err)
	}
	if !strings.Contains(m.View().Content, "Permission denied") {
		t.Error("error not shown")
	}
}

func TestNowPlaying(t *testing.T) {
	song := &Song{Artist: "Artist", Title: "Song"}
	for _, tt := range []struct {
		status *Status
		song   *Song
		want   string
	}{
		{nil, nil, ""},
		{&Status{State: StateStop, QueueLength: 3}, nil, "■ stopped · queue 3"},
		{&Status{State: StatePlay, Song: 1, QueueLength: 3, Elapsed: 5e9, Duration: 61e9}, song, "▶ Artist - Song · 0:05 / 1:01 · 2/3"},
		{&Status{State: StatePause, Song: 0, QueueLength: 1, Elapsed: 5e9}, song, "⏸ Artist - Song · 0:05 · 1/1"},
	} {
		if got := NowPlaying(tt.status, tt.song); got != tt.want {
			t.Errorf("NowPlaying(%+v) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

const queueListing = "file: a.flac\nTitle: A\nduration: 60\nPos: 0\nId: 1\n" +
	"file: b.flac\nTitle: B\nduration: 30\nPos: 1\nId: 2\n" +
	"file: c.flac\nTitle: C\nPos: 2\nId: 3"

func queueTitles(songs []Song) string {
	var out []string
	for _, s := range songs {
		out = append(out, s.Title)
	}
	return strings.Join(out, ",")
}

func TestBrowseQueue(t *testing.T) {
	m, s := newTestBrowser(t, withResponses(map[string]string{
		"status":         "state: play\nsong: 0\nplaylist: 7\nplaylistlength: 3",
		"playlistinfo":   queueListing,
		`playid "2"`:     "",
		`moveid "2" "2"`: "",
		`deleteid "3"`:   "",
	}))
	run(t, m, m.refreshCmd())

	press(t, m, "tab")
	if m.view != viewQueue || queueTitles(m.queue) != "A,B,C" || m.queueVersion != 7 {
		t.Fatalf("after tab: view %d, queue %s, version %d", m.view, queueTitles(m.queue), m.queueVersion)
	}
	if content := m.View().Content; !strings.Contains(content, "3 song(s) · 1:30") || !strings.Contains(content, "▶   1  A") {
		t.Errorf("summary missing:\n%s", m.View().Content)
	}

	press(t, m, "j", "enter")
	if m.status != "playing B" {
		t.Errorf("after enter: status %q, err %q", m.status, m.err)
	}

	press(t, m, "J")
	if got := s.sent(); !contains(got, `moveid "2" "2"`) {
		t.Errorf("move not sent: %v", got)
	}
	// the fake queue does not change, so the cursor stays on B by its id
	if song, _ := m.selectedSong(); song.ID != 2 {
		t.Errorf("cursor on %+v after moving, want B", song)
	}

	press(t, m, "G", "d")
	if m.status != "removed C from the queue" {
		t.Errorf("after d: status %q, err %q", m.status, m.err)
	}

	// moving past the end is a no-op
	before := len(s.sent())
	press(t, m, "J")
	if len(s.sent()) != before {
		t.Errorf("moved the last song down: %v", s.sent()[before:])
	}

	press(t, m, "tab")
	if m.view != viewLibrary || names(m.entries) != "Artist,Other,loose.mp3" {
		t.Errorf("after second tab: view %d, entries %s", m.view, names(m.entries))
	}
}

func TestBrowseQueueReloadsOnExternalChange(t *testing.T) {
	m, s := newTestBrowser(t, withResponses(map[string]string{
		"status":       "state: play\nsong: 0\nplaylist: 7\nplaylistlength: 3",
		"playlistinfo": queueListing,
	}))
	press(t, m, "tab")

	// another client appends a song
	s.mu.Lock()
	s.responses["status"] = "state: play\nsong: 0\nplaylist: 8\nplaylistlength: 4"
	s.responses["playlistinfo"] = queueListing + "\nfile: d.flac\nTitle: D\nPos: 3\nId: 4"
	s.mu.Unlock()

	_, cmd := m.Update(tickMsg{})
	// run the refresh, not the next tick
	run(t, m, cmd().(tea.BatchMsg)[0])
	if queueTitles(m.queue) != "A,B,C,D" || m.queueVersion != 8 {
		t.Errorf("queue %s, version %d", queueTitles(m.queue), m.queueVersion)
	}
}

func TestBrowseQueueFilterBlocksMoving(t *testing.T) {
	m, s := newTestBrowser(t, withResponses(map[string]string{
		"status":       "state: stop\nplaylist: 1\nplaylistlength: 3",
		"playlistinfo": queueListing,
	}))
	press(t, m, "tab")
	for _, k := range []string{"/", "b", "enter"} {
		m.Update(key(k))
	}
	if queueTitles(m.filteredQueue()) != "B" {
		t.Fatalf("filtered %s", queueTitles(m.filteredQueue()))
	}
	before := len(s.sent())
	press(t, m, "K")
	if len(s.sent()) != before || !strings.Contains(m.status, "clear the filter") {
		t.Errorf("moved while filtered: status %q, sent %v", m.status, s.sent()[before:])
	}
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

func TestBrowseQueueStartsInQueue(t *testing.T) {
	s := newFakeServer(t, withResponses(map[string]string{
		"status":       "state: play\nsong: 1\nplaylist: 3\nplaylistlength: 3",
		"playlistinfo": queueListing,
		"clear":        "",
	}))
	m := newBrowser(context.Background(), dial(t, s), "")
	m.view = viewQueue
	if err := m.loadInitial(); err != nil {
		t.Fatal(err)
	}
	m.height = 40
	// run everything Init starts but the tick
	for i, cmd := range m.Init()().(tea.BatchMsg) {
		if i != 1 {
			run(t, m, cmd)
		}
	}
	if queueTitles(m.queue) != "A,B,C" || !strings.Contains(m.View().Content, "▶   2  B") {
		t.Fatalf("queue %s, view:\n%s", queueTitles(m.queue), m.View().Content)
	}

	press(t, m, "c", "c")
	if m.status != "cleared the queue" {
		t.Errorf("after c c: status %q, err %q", m.status, m.err)
	}
}

func TestBrowseClearNeedsConfirmation(t *testing.T) {
	m, s := newTestBrowser(t, withResponses(map[string]string{
		"status": "state: play\nsong: 0\nplaylist: 1\nplaylistlength: 3",
		"clear":  "",
	}))
	run(t, m, m.refreshCmd())

	press(t, m, "c")
	if m.status != "press c again to clear the queue (3 song(s))" || contains(s.sent(), "clear") {
		t.Fatalf("after c: status %q, sent %v", m.status, s.sent())
	}
	// a tick in between does not cancel
	m.Update(nowPlayingMsg{status: m.playing, song: m.song})

	press(t, m, "j")
	if m.status != "" || m.confirmClear {
		t.Errorf("other key did not cancel: status %q", m.status)
	}
	press(t, m, "c")
	if contains(s.sent(), "clear") {
		t.Fatal("cleared after cancelling and a single c")
	}

	press(t, m, "c")
	if m.status != "cleared the queue" || !contains(s.sent(), "clear") {
		t.Errorf("after c c: status %q, sent %v", m.status, s.sent())
	}
	// the next c asks again
	press(t, m, "c")
	if !m.confirmClear {
		t.Error("confirmation was not reset after clearing")
	}
}
