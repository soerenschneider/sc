// Package mpd implements a minimal client for the Music Player Daemon
// protocol, see https://mpd.readthedocs.io/en/latest/protocol.html.
package mpd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultAddress = "localhost:6600"

	envHost = "MPD_HOST"
	envPort = "MPD_PORT"
)

// Error is an error reported by MPD ("ACK" response).
type Error struct {
	Command string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("mpd: %s: %s", e.Command, e.Message)
}

// Pair is a single "key: value" line of a response. Responses are kept as
// ordered pairs, as listings repeat keys for every entry.
type Pair struct {
	Key, Value string
}

// Client is a connection to MPD. It is safe for concurrent use, commands are
// sent one at a time.
type Client struct {
	network, addr, password string

	mu   sync.Mutex
	conn net.Conn
	r    *bufio.Reader
}

// ResolveAddress returns the network, address and password to connect to.
// An empty addr falls back to MPD_HOST and MPD_PORT like mpc does, and then
// to DefaultAddress. Addresses may carry a password ("secret@host:6600") and
// paths are used as unix sockets.
func ResolveAddress(addr string) (network, address, password string) {
	if addr == "" {
		host, port := os.Getenv(envHost), os.Getenv(envPort)
		switch {
		case host != "" && port != "":
			addr = host + ":" + port
		case host != "":
			addr = host
		case port != "":
			addr = "localhost:" + port
		default:
			addr = DefaultAddress
		}
	}

	// MPD_HOST may also be "password@@abstract-socket", so split at the
	// first @ only if something precedes it.
	if i := strings.Index(addr, "@"); i > 0 {
		password, addr = addr[:i], addr[i+1:]
	}
	if strings.HasPrefix(addr, "/") || strings.HasPrefix(addr, "@") || strings.HasPrefix(addr, "~") {
		if strings.HasPrefix(addr, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				addr = home + addr[1:]
			}
		}
		return "unix", addr, password
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "6600")
	}
	return "tcp", addr, password
}

// Dial connects to MPD at addr, see ResolveAddress, and authenticates with
// password if not empty. A password given here takes precedence over one in
// addr.
func Dial(ctx context.Context, addr, password string) (*Client, error) {
	network, address, addrPassword := ResolveAddress(addr)
	if password == "" {
		password = addrPassword
	}
	c := &Client{network: network, addr: address, password: password}
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// Address returns the address the client is connected to.
func (c *Client) Address() string {
	return c.addr
}

func (c *Client) connect(ctx context.Context) error {
	var d net.Dialer
	conn, err := d.DialContext(ctx, c.network, c.addr)
	if err != nil {
		return fmt.Errorf("could not connect to mpd at %s: %w", c.addr, err)
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()

	r := bufio.NewReader(conn)
	greeting, err := r.ReadString('\n')
	if err != nil {
		conn.Close()
		return fmt.Errorf("could not read greeting from %s: %w", c.addr, err)
	}
	if !strings.HasPrefix(greeting, "OK MPD ") {
		conn.Close()
		return fmt.Errorf("%s is not an mpd server, greeting: %q", c.addr, strings.TrimSpace(greeting))
	}
	c.conn, c.r = conn, r

	if c.password != "" {
		if _, err := c.roundTrip(ctx, "password", c.password); err != nil {
			func() {
				_ = c.closeConn()
			}()
			return err
		}
	}
	return nil
}

// Close closes the connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	_, _ = io.WriteString(c.conn, "close\n")
	return c.closeConn()
}

func (c *Client) closeConn() error {
	err := c.conn.Close()
	c.conn, c.r = nil, nil
	return err
}

// Command sends a command with the given arguments and returns the response.
func (c *Client) Command(ctx context.Context, name string, args ...string) ([]Pair, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		if err := c.connect(ctx); err != nil {
			return nil, err
		}
	}
	pairs, err := c.roundTrip(ctx, name, args...)
	if err != nil && isStale(err) {
		// MPD closes idle connections after its connection_timeout. It does
		// so before reading our command, so sending it again is safe.
		func() {
			_ = c.closeConn()
		}()
		if err := c.connect(ctx); err != nil {
			return nil, err
		}
		pairs, err = c.roundTrip(ctx, name, args...)
	}
	var mpdErr *Error
	if err != nil && !errors.As(err, &mpdErr) && c.conn != nil {
		// the connection is in an unknown state now
		func() {
			_ = c.closeConn()
		}()
	}
	return pairs, err
}

// staleError marks a connection that was closed before any response was read.
type staleError struct{ err error }

func (e staleError) Error() string { return e.err.Error() }
func (e staleError) Unwrap() error { return e.err }

func isStale(err error) bool {
	var stale staleError
	return errors.As(err, &stale)
}

func (c *Client) roundTrip(ctx context.Context, name string, args ...string) ([]Pair, error) {
	stop := context.AfterFunc(ctx, func() { _ = c.conn.SetDeadline(time.Now()) })
	defer stop()
	defer func() { _ = c.conn.SetDeadline(time.Time{}) }()

	if _, err := io.WriteString(c.conn, formatCommand(name, args...)); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, staleError{err}
	}

	var pairs []Pair
	for first := true; ; first = false {
		line, err := c.r.ReadString('\n')
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			if first && line == "" && (errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE)) {
				return nil, staleError{err}
			}
			return nil, fmt.Errorf("could not read response to %s: %w", name, err)
		}
		line = strings.TrimSuffix(line, "\n")

		switch {
		case line == "OK":
			return pairs, nil
		case strings.HasPrefix(line, "ACK "):
			return nil, parseAck(name, line)
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			return nil, fmt.Errorf("unexpected response to %s: %q", name, line)
		}
		pairs = append(pairs, Pair{Key: key, Value: value})
	}
}

// formatCommand quotes the arguments as required by the protocol.
func formatCommand(name string, args ...string) string {
	var b strings.Builder
	b.WriteString(name)
	for _, arg := range args {
		b.WriteString(` "`)
		for _, r := range arg {
			if r == '"' || r == '\\' {
				b.WriteByte('\\')
			}
			b.WriteRune(r)
		}
		b.WriteByte('"')
	}
	b.WriteByte('\n')
	return b.String()
}

// parseAck parses "ACK [error@command_listNum] {current_command} message".
func parseAck(name, line string) error {
	msg := strings.TrimPrefix(line, "ACK ")
	if _, rest, ok := strings.Cut(msg, "] "); ok {
		msg = rest
	}
	if cmd, rest, ok := strings.Cut(msg, "} "); ok && strings.HasPrefix(cmd, "{") {
		if cmd = strings.TrimPrefix(cmd, "{"); cmd != "" {
			name = cmd
		}
		msg = rest
	}
	return &Error{Command: name, Message: msg}
}

// attrs returns the pairs as a map. For repeated keys, e.g. a song with
// several artists, the first value is kept.
func attrs(pairs []Pair) map[string]string {
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		if _, seen := m[p.Key]; !seen {
			m[p.Key] = p.Value
		}
	}
	return m
}

func parseSeconds(s string) time.Duration {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return time.Duration(f * float64(time.Second))
}
