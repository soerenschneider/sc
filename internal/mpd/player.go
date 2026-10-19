package mpd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
)

type State string

const (
	StatePlay  State = "play"
	StatePause State = "pause"
	StateStop  State = "stop"
)

// Status is the player state as reported by the "status" command.
type Status struct {
	State State
	// Volume is -1 if mpd has no mixer.
	Volume int
	// Song is the position of the current song in the queue, -1 if none.
	Song        int
	QueueLength int
	// QueueVersion changes whenever the queue is modified.
	QueueVersion uint64
	Elapsed      time.Duration
	Duration     time.Duration
	Repeat       bool
	Random       bool
	// UpdatingDB is the id of the running database update, 0 if none.
	UpdatingDB int
	// Error is the last player error, e.g. an unplayable file.
	Error string
}

// Song is a song in the database or the queue.
type Song struct {
	File     string
	Artist   string
	Title    string
	Album    string
	Name     string // name of a stream
	Duration time.Duration
	// ID identifies the song in the queue and Pos is its position, both 0
	// for database songs.
	ID  int
	Pos int
}

// DisplayName returns "Artist - Title", falling back to the stream name or
// the file name for untagged songs.
func (s Song) DisplayName() string {
	switch {
	case s.Artist != "" && s.Title != "":
		return s.Artist + " - " + s.Title
	case s.Title != "":
		return s.Title
	case s.Name != "":
		return s.Name
	}
	return path.Base(s.File)
}

func (c *Client) Status(ctx context.Context) (*Status, error) {
	pairs, err := c.Command(ctx, "status")
	if err != nil {
		return nil, err
	}
	a := attrs(pairs)
	s := &Status{
		State:        State(a["state"]),
		Volume:       atoi(a["volume"], -1),
		Song:         atoi(a["song"], -1),
		QueueLength:  atoi(a["playlistlength"], 0),
		QueueVersion: atou(a["playlist"]),
		Elapsed:      parseSeconds(a["elapsed"]),
		Duration:     parseSeconds(a["duration"]),
		Repeat:       a["repeat"] == "1",
		Random:       a["random"] == "1",
		UpdatingDB:   atoi(a["updating_db"], 0),
		Error:        a["error"],
	}
	if s.Duration == 0 {
		// older servers only report "time: elapsed:total"
		if _, total, ok := strings.Cut(a["time"], ":"); ok {
			s.Duration = parseSeconds(total)
		}
	}
	return s, nil
}

// CurrentSong returns the current song, nil if there is none.
func (c *Client) CurrentSong(ctx context.Context) (*Song, error) {
	pairs, err := c.Command(ctx, "currentsong")
	if err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return nil, nil
	}
	song := songFromAttrs(attrs(pairs))
	return &song, nil
}

// Toggle pauses playback if playing and resumes or starts it otherwise. It
// returns the new state.
func (c *Client) Toggle(ctx context.Context) (State, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return "", err
	}
	switch status.State {
	case StatePlay:
		_, err = c.Command(ctx, "pause", "1")
		return StatePause, err
	case StatePause:
		_, err = c.Command(ctx, "pause", "0")
		return StatePlay, err
	}
	if status.QueueLength == 0 {
		return StateStop, errors.New("the queue is empty")
	}
	_, err = c.Command(ctx, "play")
	return StatePlay, err
}

// Play starts playback at the given queue position or, if pos is negative,
// resumes it or starts it at the current song.
func (c *Client) Play(ctx context.Context, pos int) error {
	args := []string{}
	if pos >= 0 {
		args = append(args, strconv.Itoa(pos))
	}
	_, err := c.Command(ctx, "play", args...)
	return err
}

// Stop stops playback.
func (c *Client) Stop(ctx context.Context) error {
	_, err := c.Command(ctx, "stop")
	return err
}

// Next plays the next song in the queue.
func (c *Client) Next(ctx context.Context) error {
	_, err := c.Command(ctx, "next")
	return err
}

// Previous plays the previous song in the queue.
func (c *Client) Previous(ctx context.Context) error {
	_, err := c.Command(ctx, "previous")
	return err
}

// SetRandom enables or disables random mode.
func (c *Client) SetRandom(ctx context.Context, on bool) error {
	_, err := c.Command(ctx, "random", boolArg(on))
	return err
}

// SetRepeat enables or disables repeat mode.
func (c *Client) SetRepeat(ctx context.Context, on bool) error {
	_, err := c.Command(ctx, "repeat", boolArg(on))
	return err
}

// Update starts a scan of dir in the music directory for changes, the whole
// directory if empty, and returns the id of the update job.
func (c *Client) Update(ctx context.Context, dir string) (int, error) {
	var args []string
	if dir != "" {
		args = append(args, dir)
	}
	pairs, err := c.Command(ctx, "update", args...)
	if err != nil {
		return 0, err
	}
	return atoi(attrs(pairs)["updating_db"], 0), nil
}

// SearchTagAny matches a query against all tags and the file name.
const SearchTagAny = "any"

// Search returns the songs in the database whose tag contains query, case
// insensitively. Tag is e.g. "artist", "album", "title", "file" or
// SearchTagAny.
func (c *Client) Search(ctx context.Context, tag, query string) ([]Song, error) {
	pairs, err := c.Command(ctx, "search", tag, query)
	if err != nil {
		return nil, err
	}
	return songsFromPairs(pairs), nil
}

// SearchAdd appends the songs matching like Search to the queue.
func (c *Client) SearchAdd(ctx context.Context, tag, query string) error {
	_, err := c.Command(ctx, "searchadd", tag, query)
	return err
}

func boolArg(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// Clear removes all songs from the queue.
func (c *Client) Clear(ctx context.Context) error {
	_, err := c.Command(ctx, "clear")
	return err
}

// Add appends a song or, recursively, a directory to the queue.
func (c *Client) Add(ctx context.Context, uri string) error {
	_, err := c.Command(ctx, "add", uri)
	return err
}

// AddAndPlay appends a song to the queue and plays it.
func (c *Client) AddAndPlay(ctx context.Context, uri string) error {
	pairs, err := c.Command(ctx, "addid", uri)
	if err != nil {
		return err
	}
	id, err := strconv.Atoi(attrs(pairs)["Id"])
	if err != nil {
		return fmt.Errorf("mpd did not return the id of the added song")
	}
	return c.PlayID(ctx, id)
}

// Entry is an item of a database directory.
type Entry struct {
	// Path is the full path of the entry in the database.
	Path  string
	IsDir bool
	// Song is set for songs.
	Song *Song
}

// Name returns the name to display for the entry.
func (e Entry) Name() string {
	if e.IsDir {
		return path.Base(e.Path)
	}
	return e.Song.DisplayName()
}

// List returns the directories and songs in dir of the database, "" being the
// root. Playlist files are skipped.
func (c *Client) List(ctx context.Context, dir string) ([]Entry, error) {
	pairs, err := c.Command(ctx, "lsinfo", dir)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, item := range splitItems(pairs) {
		switch {
		case item["directory"] != "":
			entries = append(entries, Entry{Path: item["directory"], IsDir: true})
		case item["file"] != "":
			song := songFromAttrs(item)
			entries = append(entries, Entry{Path: song.File, Song: &song})
		}
	}
	return entries, nil
}

// Queue returns the songs in the queue.
func (c *Client) Queue(ctx context.Context) ([]Song, error) {
	pairs, err := c.Command(ctx, "playlistinfo")
	if err != nil {
		return nil, err
	}
	return songsFromPairs(pairs), nil
}

// songsFromPairs returns the songs of a listing, skipping other items.
func songsFromPairs(pairs []Pair) []Song {
	var songs []Song
	for _, item := range splitItems(pairs) {
		if item["file"] != "" {
			songs = append(songs, songFromAttrs(item))
		}
	}
	return songs
}

// PlayID plays the song with the given queue id.
func (c *Client) PlayID(ctx context.Context, id int) error {
	_, err := c.Command(ctx, "playid", strconv.Itoa(id))
	return err
}

// DeleteID removes the song with the given queue id from the queue.
func (c *Client) DeleteID(ctx context.Context, id int) error {
	_, err := c.Command(ctx, "deleteid", strconv.Itoa(id))
	return err
}

// MoveID moves the song with the given queue id to position to.
func (c *Client) MoveID(ctx context.Context, id, to int) error {
	_, err := c.Command(ctx, "moveid", strconv.Itoa(id), strconv.Itoa(to))
	return err
}

// splitItems splits a listing into its items, each starting with a
// "directory", "file" or "playlist" line followed by its attributes. For
// repeated attributes, e.g. several artists, the first value is kept.
func splitItems(pairs []Pair) []map[string]string {
	var items []map[string]string
	var item map[string]string
	for _, p := range pairs {
		switch p.Key {
		case "directory", "file", "playlist":
			item = map[string]string{}
			items = append(items, item)
		}
		if item == nil {
			continue
		}
		if _, seen := item[p.Key]; !seen {
			item[p.Key] = p.Value
		}
	}
	return items
}

func songFromAttrs(a map[string]string) Song {
	s := Song{
		File:     a["file"],
		Artist:   a["Artist"],
		Title:    a["Title"],
		Album:    a["Album"],
		Name:     a["Name"],
		Duration: parseSeconds(a["duration"]),
		ID:       atoi(a["Id"], 0),
		Pos:      atoi(a["Pos"], 0),
	}
	if s.Duration == 0 {
		s.Duration = parseSeconds(a["Time"])
	}
	return s
}

func atou(s string) uint64 {
	u, _ := strconv.ParseUint(s, 10, 64)
	return u
}

func atoi(s string, fallback int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return i
}

// FormatDuration formats d as m:ss, or h:mm:ss for an hour or more.
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h, m, s := int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
