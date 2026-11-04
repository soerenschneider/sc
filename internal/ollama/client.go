package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	DefaultAddress = "http://localhost:11434"
	defaultPort    = "11434"
)

// ErrNotFound is returned if a model does not exist.
var ErrNotFound = errors.New("model not found")

type Client struct {
	client *http.Client
	url    string
}

// NewClient returns a client for the Ollama instance at address. If address is
// empty, it defaults to OLLAMA_HOST and then to DefaultAddress.
func NewClient(address string) (*Client, error) {
	if address == "" {
		address = os.Getenv("OLLAMA_HOST")
	}
	if address == "" {
		address = DefaultAddress
	}

	base, err := normalizeAddress(address)
	if err != nil {
		return nil, err
	}

	// No client timeout, pulls stream for a long time. Requests are bound by
	// their context instead.
	return &Client{client: &http.Client{}, url: base}, nil
}

// normalizeAddress accepts addresses like ollama does for OLLAMA_HOST, i.e.
// "host", "host:port" or a full URL, and returns the base URL.
func normalizeAddress(address string) (string, error) {
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}
	u, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("invalid address %q: %w", address, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid address %q: no host", address)
	}
	if u.Port() == "" && u.Scheme == "http" {
		u.Host += ":" + defaultPort
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func (c *Client) Address() string {
	return c.url
}

func (c *Client) Version(ctx context.Context) (string, error) {
	var resp struct {
		Version string `json:"version"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/version", nil, &resp); err != nil {
		return "", err
	}
	return resp.Version, nil
}

// ListModels returns the models available locally on the instance.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	var resp struct {
		Models []Model `json:"models"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/tags", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}

// ListRunning returns the models currently loaded into memory.
func (c *Client) ListRunning(ctx context.Context) ([]RunningModel, error) {
	var resp struct {
		Models []RunningModel `json:"models"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/ps", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}

func (c *Client) ShowModel(ctx context.Context, name string) (*ModelInfo, error) {
	var resp ModelInfo
	if err := c.do(ctx, http.MethodPost, "/api/show", map[string]string{"model": name}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteModel(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/delete", map[string]string{"model": name}, nil)
}

// UnloadModel evicts a model from memory.
func (c *Client) UnloadModel(ctx context.Context, name string) error {
	body := map[string]any{"model": name, "keep_alive": 0}
	return c.do(ctx, http.MethodPost, "/api/generate", body, nil)
}

// PullModel downloads a model, calling progress for every status update the
// server streams.
func (c *Client) PullModel(ctx context.Context, name string, progress func(PullProgress)) error {
	resp, err := c.request(ctx, http.MethodPost, "/api/pull", map[string]any{"model": name, "stream": true})
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	err = decodeStream(resp.Body, func(p PullProgress) bool {
		if progress != nil {
			progress(p)
		}
		return p.Status == "success"
	})
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return errors.New("pull ended without success")
	}
	return err
}

// Chat sends the conversation to the model and streams the answer, calling
// onChunk for every part the server sends. It returns the complete answer.
func (c *Client) Chat(ctx context.Context, req ChatRequest, onChunk func(Message)) (Message, error) {
	req.Stream = true
	resp, err := c.request(ctx, http.MethodPost, "/api/chat", req)
	if err != nil {
		return Message{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	answer := Message{Role: "assistant"}
	var content, thinking strings.Builder
	err = decodeStream(resp.Body, func(chunk ChatResponse) bool {
		content.WriteString(chunk.Message.Content)
		thinking.WriteString(chunk.Message.Thinking)
		if onChunk != nil {
			onChunk(chunk.Message)
		}
		return chunk.Done
	})
	if err != nil {
		return Message{}, err
	}
	answer.Content = content.String()
	answer.Thinking = thinking.String()
	return answer, nil
}

// decodeStream decodes the newline delimited JSON objects of a streaming
// response, calling handle for each until it returns true. Errors sent by the
// server are returned.
func decodeStream[T any](body io.Reader, handle func(T) bool) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var apiErr struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(line, &apiErr); err != nil {
			return fmt.Errorf("decoding stream: %w", err)
		}
		if apiErr.Error != "" {
			return errors.New(apiErr.Error)
		}
		var v T
		if err := json.Unmarshal(line, &v); err != nil {
			return fmt.Errorf("decoding stream: %w", err)
		}
		if handle(v) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return io.ErrUnexpectedEOF
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	resp, err := c.request(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding response of %s: %w", path, err)
	}
	return nil
}

// request sends a request and returns the response if its status is 2xx.
func (c *Client) request(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.url+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// #nosec G704
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var apiErr struct {
		Error string `json:"error"`
	}
	msg := strings.TrimSpace(string(data))
	if json.Unmarshal(data, &apiErr) == nil && apiErr.Error != "" {
		msg = apiErr.Error
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, msg)
	}
	return nil, fmt.Errorf("%s %s: %s: %s", method, path, resp.Status, msg)
}
