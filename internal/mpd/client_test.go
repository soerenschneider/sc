package mpd

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeServer speaks enough of the MPD protocol for tests. Responses maps a
// command line to the response lines before "OK", or to an "ACK ..." line.
type fakeServer struct {
	t         *testing.T
	ln        net.Listener
	responses map[string]string

	mu       sync.Mutex
	commands []string
	// dropAfter closes a connection after it handled that many commands,
	// like mpd does with idle connections.
	dropAfter int
	conns     int
}

func newFakeServer(t *testing.T, responses map[string]string) *fakeServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeServer{t: t, ln: ln, responses: responses}
	t.Cleanup(func() { ln.Close() })
	go s.serve()
	return s
}

func (s *fakeServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.conns++
		s.mu.Unlock()
		go s.handle(conn)
	}
}

func (s *fakeServer) handle(conn net.Conn) {
	defer conn.Close()
	_, _ = conn.Write([]byte("OK MPD 0.24.0\n"))
	r := bufio.NewReader(conn)
	for handled := 0; ; handled++ {
		s.mu.Lock()
		drop := s.dropAfter > 0 && handled >= s.dropAfter
		s.mu.Unlock()
		if drop {
			return
		}

		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSuffix(line, "\n")
		if line == "close" {
			return
		}
		s.mu.Lock()
		s.commands = append(s.commands, line)
		resp, ok := s.responses[line]
		s.mu.Unlock()

		switch {
		case !ok:
			resp = "ACK [5@0] {} unknown command " + line + "\n"
		case strings.HasPrefix(resp, "ACK "):
			resp += "\n"
		case resp != "":
			resp += "\nOK\n"
		default:
			resp = "OK\n"
		}
		_, _ = conn.Write([]byte(resp))
	}
}

func (s *fakeServer) sent() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.commands...)
}

func dial(t *testing.T, s *fakeServer) *Client {
	t.Helper()
	c, err := Dial(context.Background(), s.ln.Addr().String(), "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestResolveAddress(t *testing.T) {
	tests := []struct {
		addr, host, port               string
		network, address, wantPassword string
	}{
		{"", "", "", "tcp", "localhost:6600", ""},
		{"", "music", "", "tcp", "music:6600", ""},
		{"", "music", "6601", "tcp", "music:6601", ""},
		{"", "", "6601", "tcp", "localhost:6601", ""},
		{"", "secret@music", "", "tcp", "music:6600", "secret"},
		{"", "/run/mpd/socket", "", "unix", "/run/mpd/socket", ""},
		{"", "secret@@mpd", "", "unix", "@mpd", "secret"},
		{"music:6602", "ignored", "1", "tcp", "music:6602", ""},
		{"secret@music", "", "", "tcp", "music:6600", "secret"},
		{"::1", "", "", "tcp", "[::1]:6600", ""},
	}
	for _, tt := range tests {
		t.Setenv(envHost, tt.host)
		t.Setenv(envPort, tt.port)
		network, address, password := ResolveAddress(tt.addr)
		if network != tt.network || address != tt.address || password != tt.wantPassword {
			t.Errorf("ResolveAddress(%q) with MPD_HOST=%q MPD_PORT=%q = %s, %s, %q", tt.addr, tt.host, tt.port, network, address, password)
		}
	}
}

func TestFormatCommand(t *testing.T) {
	got := formatCommand("add", `Artist/My "Best" Song\1.flac`)
	if want := `add "Artist/My \"Best\" Song\\1.flac"` + "\n"; got != want {
		t.Errorf("formatCommand() = %q, want %q", got, want)
	}
}

func TestDialPassword(t *testing.T) {
	s := newFakeServer(t, map[string]string{ // #nosec:G101
		`password "secret"`: "",
		`password "wrong"`:  "ACK [3@0] {password} incorrect password",
	})
	if _, err := Dial(context.Background(), "secret@"+s.ln.Addr().String(), ""); err != nil {
		t.Errorf("valid password: %v", err)
	}
	_, err := Dial(context.Background(), s.ln.Addr().String(), "wrong")
	var mpdErr *Error
	if !errors.As(err, &mpdErr) || mpdErr.Message != "incorrect password" || mpdErr.Command != "password" {
		t.Errorf("wrong password: got %v", err)
	}
}

func TestDialNotMPD(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_, _ = conn.Write([]byte("SSH-2.0-OpenSSH\n"))
			conn.Close()
		}
	}()
	if _, err := Dial(context.Background(), ln.Addr().String(), ""); err == nil || !strings.Contains(err.Error(), "not an mpd server") {
		t.Errorf("got %v", err)
	}
}

func TestDialTimeout(t *testing.T) {
	// accepts but never greets
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			defer conn.Close()
			time.Sleep(time.Second)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := Dial(ctx, ln.Addr().String(), ""); err == nil {
		t.Fatal("expected timeout")
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Errorf("dial did not honor the context, took %s", time.Since(start))
	}
}

func TestReconnectsDroppedConnection(t *testing.T) {
	s := newFakeServer(t, map[string]string{"clear": ""})
	s.mu.Lock()
	s.dropAfter = 1
	s.mu.Unlock()
	c := dial(t, s)
	for i := 0; i < 3; i++ {
		if err := c.Clear(context.Background()); err != nil {
			t.Fatalf("command %d: %v", i, err)
		}
	}
	if got := s.sent(); len(got) != 3 {
		t.Errorf("sent %v, want each command once", got)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns != 3 {
		t.Errorf("connections = %d, want 3", s.conns)
	}
}

func TestStatusAndToggle(t *testing.T) {
	tests := map[string]struct {
		status   string
		want     State
		wantSent string
		wantErr  bool
	}{
		"pauses":  {status: "state: play\nplaylistlength: 3", want: StatePause, wantSent: `pause "1"`},
		"resumes": {status: "state: pause\nplaylistlength: 3", want: StatePlay, wantSent: `pause "0"`},
		"starts":  {status: "state: stop\nplaylistlength: 3", want: StatePlay, wantSent: "play"},
		"empty":   {status: "state: stop\nplaylistlength: 0", want: StateStop, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := newFakeServer(t, map[string]string{"status": tt.status, `pause "1"`: "", `pause "0"`: "", "play": ""})
			got, err := dial(t, s).Toggle(context.Background())
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("Toggle() = %s, %v", got, err)
			}
			sent := s.sent()
			if tt.wantSent != "" && sent[len(sent)-1] != tt.wantSent {
				t.Errorf("sent %v, want last %q", sent, tt.wantSent)
			}
			if tt.wantSent == "" && len(sent) != 1 {
				t.Errorf("sent %v, want only status", sent)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	s := newFakeServer(t, map[string]string{
		"status":      "volume: 42\nrepeat: 1\nrandom: 0\nplaylistlength: 12\nstate: play\nsong: 2\nsongid: 3\nelapsed: 83.512\nduration: 296.000\nerror: Failed to decode",
		"currentsong": "file: Artist/Album/03 Song.flac\nArtist: Artist\nArtist: Featured\nTitle: Song\nAlbum: Album\nTime: 296\nduration: 296.000\nPos: 2\nId: 3",
	})
	c := dial(t, s)
	status, err := c.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := Status{State: StatePlay, Volume: 42, Song: 2, QueueLength: 12, Elapsed: 83512 * time.Millisecond, Duration: 296 * time.Second, Repeat: true, Error: "Failed to decode"}
	if *status != want {
		t.Errorf("Status() = %+v, want %+v", *status, want)
	}
	song, err := c.CurrentSong(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if song.DisplayName() != "Artist - Song" || song.Album != "Album" || song.ID != 3 || song.Duration != 296*time.Second {
		t.Errorf("CurrentSong() = %+v", song)
	}
}

func TestCurrentSongNone(t *testing.T) {
	s := newFakeServer(t, map[string]string{"currentsong": ""})
	song, err := dial(t, s).CurrentSong(context.Background())
	if err != nil || song != nil {
		t.Errorf("CurrentSong() = %+v, %v", song, err)
	}
}

func TestList(t *testing.T) {
	s := newFakeServer(t, map[string]string{
		`lsinfo "Artist"`: strings.Join([]string{
			"directory: Artist/Album",
			"Last-Modified: 2024-01-01T00:00:00Z",
			"file: Artist/untagged.mp3",
			"Last-Modified: 2024-01-01T00:00:00Z",
			"playlist: Artist/mix.m3u",
			"Last-Modified: 2024-01-01T00:00:00Z",
			"file: Artist/song.flac",
			"Artist: Artist",
			"Title: Song",
			"duration: 61.5",
		}, "\n"),
		`lsinfo "missing"`: "ACK [50@0] {lsinfo} No such directory",
	})
	c := dial(t, s)
	entries, err := c.List(context.Background(), "Artist")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range entries {
		got = append(got, e.Path+"="+e.Name())
	}
	want := []string{"Artist/Album=Album", "Artist/untagged.mp3=untagged.mp3", "Artist/song.flac=Artist - Song"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("List() = %v, want %v", got, want)
	}
	if !entries[0].IsDir || entries[2].Song.Duration != 61500*time.Millisecond {
		t.Errorf("List() = %+v", entries)
	}

	_, err = c.List(context.Background(), "missing")
	var mpdErr *Error
	if !errors.As(err, &mpdErr) || mpdErr.Message != "No such directory" {
		t.Errorf("missing directory: got %v", err)
	}
	// the connection stays usable after an ACK
	if _, err := c.List(context.Background(), "Artist"); err != nil {
		t.Errorf("after ACK: %v", err)
	}
}

func TestAddAndPlay(t *testing.T) {
	s := newFakeServer(t, map[string]string{
		`addid "Artist/song.flac"`: "Id: 17",
		`playid "17"`:              "",
		`add "Artist"`:             "",
	})
	c := dial(t, s)
	if err := c.AddAndPlay(context.Background(), "Artist/song.flac"); err != nil {
		t.Fatal(err)
	}
	if err := c.Add(context.Background(), "Artist"); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(s.sent(), ","); got != `addid "Artist/song.flac",playid "17",add "Artist"` {
		t.Errorf("sent %s", got)
	}
}

func TestFormatDuration(t *testing.T) {
	for d, want := range map[time.Duration]string{
		0:                               "0:00",
		83512 * time.Millisecond:        "1:24",
		time.Hour + 2*time.Minute + 3e9: "1:02:03",
		59*time.Minute + 59*time.Second: "59:59",
	} {
		if got := FormatDuration(d); got != want {
			t.Errorf("FormatDuration(%s) = %q, want %q", d, got, want)
		}
	}
}

func TestQueue(t *testing.T) {
	s := newFakeServer(t, map[string]string{
		"playlistinfo": strings.Join([]string{
			"file: Artist/one.flac", "Artist: Artist", "Title: One", "duration: 60", "Pos: 0", "Id: 11",
			"file: http://radio.example/stream", "Name: Radio", "Pos: 1", "Id: 12",
		}, "\n"),
		"status":          "state: play\nplaylist: 42\nplaylistlength: 2",
		`playid "12"`:     "",
		`deleteid "11"`:   "",
		`moveid "12" "0"`: "",
	})
	c := dial(t, s)
	songs, err := c.Queue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) != 2 || songs[0].ID != 11 || songs[0].Pos != 0 || songs[0].Duration != time.Minute ||
		songs[1].DisplayName() != "Radio" || songs[1].Pos != 1 || songs[1].ID != 12 {
		t.Errorf("Queue() = %+v", songs)
	}
	status, err := c.Status(context.Background())
	if err != nil || status.QueueVersion != 42 {
		t.Errorf("QueueVersion = %+v, %v", status, err)
	}

	ctx := context.Background()
	if err := c.PlayID(ctx, 12); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteID(ctx, 11); err != nil {
		t.Fatal(err)
	}
	if err := c.MoveID(ctx, 12, 0); err != nil {
		t.Fatal(err)
	}
	sent := s.sent()
	if got := strings.Join(sent[len(sent)-3:], ","); got != `playid "12",deleteid "11",moveid "12" "0"` {
		t.Errorf("sent %s", got)
	}
}

func TestPlayback(t *testing.T) {
	s := newFakeServer(t, map[string]string{
		"play": "", `play "4"`: "", "stop": "", "next": "", "previous": "",
		`random "1"`: "", `repeat "0"`: "",
	})
	c := dial(t, s)
	ctx := context.Background()
	for _, f := range []func() error{
		func() error { return c.Play(ctx, -1) },
		func() error { return c.Play(ctx, 4) },
		func() error { return c.Stop(ctx) },
		func() error { return c.Next(ctx) },
		func() error { return c.Previous(ctx) },
		func() error { return c.SetRandom(ctx, true) },
		func() error { return c.SetRepeat(ctx, false) },
	} {
		if err := f(); err != nil {
			t.Fatal(err)
		}
	}
	want := `play,play "4",stop,next,previous,random "1",repeat "0"`
	if got := strings.Join(s.sent(), ","); got != want {
		t.Errorf("sent %s, want %s", got, want)
	}
}

func TestUpdate(t *testing.T) {
	s := newFakeServer(t, map[string]string{
		"update":          "updating_db: 3",
		`update "Artist"`: "updating_db: 4",
		"status":          "state: stop\nupdating_db: 4",
	})
	c := dial(t, s)
	ctx := context.Background()
	if id, err := c.Update(ctx, ""); err != nil || id != 3 {
		t.Errorf("Update() = %d, %v", id, err)
	}
	if id, err := c.Update(ctx, "Artist"); err != nil || id != 4 {
		t.Errorf("Update(Artist) = %d, %v", id, err)
	}
	if status, err := c.Status(ctx); err != nil || status.UpdatingDB != 4 {
		t.Errorf("Status() = %+v, %v", status, err)
	}
}

func TestSearch(t *testing.T) {
	s := newFakeServer(t, map[string]string{
		`search "any" "song"`: strings.Join([]string{
			"file: Artist/one.flac", "Artist: Artist", "Title: Song One", "duration: 60",
			"file: Other/song.mp3",
		}, "\n"),
		`searchadd "artist" "Artist"`: "",
	})
	c := dial(t, s)
	ctx := context.Background()
	songs, err := c.Search(ctx, SearchTagAny, "song")
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) != 2 || songs[0].DisplayName() != "Artist - Song One" || songs[1].DisplayName() != "song.mp3" {
		t.Errorf("Search() = %+v", songs)
	}
	if err := c.SearchAdd(ctx, "artist", "Artist"); err != nil {
		t.Fatal(err)
	}
}
