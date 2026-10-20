package linkding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// ErrNotFound is returned if a bookmark does not exist.
var ErrNotFound = errors.New("bookmark not found")

type LinkdingClient struct {
	client *http.Client
	url    string
	token  string
}

func NewLinkdingClient(url string, token string) (*LinkdingClient, error) {
	return &LinkdingClient{
		client: &http.Client{Timeout: 3 * time.Second},
		url:    url,
		token:  token,
	}, nil
}

func buildURL(baseURL string, q string, limit int) (string, error) {
	u, err := url.Parse(baseURL + "/api/bookmarks/")
	if err != nil {
		return "", err
	}

	query := url.Values{}
	if q != "" {
		query.Set("q", q)
	}

	if limit < 50 || limit > 1000 {
		limit = 50
	}
	query.Set("limit", strconv.Itoa(limit))

	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (c *LinkdingClient) AddBookmark(ctx context.Context, bookmark Bookmark) (int64, error) {
	u, err := url.Parse(c.url + "/api/bookmarks/")
	if err != nil {
		return -1, err
	}

	endpoint := u.String()

	data, _ := json.Marshal(bookmark)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+c.token)

	// #nosec G704
	resp, err := c.client.Do(req)
	if err != nil {
		return -1, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		return -1, fmt.Errorf("expected http code 201, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	var created Bookmark
	err = json.Unmarshal(body, &created)
	if err != nil {
		return -1, err
	}

	return created.ID, nil
}

func (c *LinkdingClient) ListBookmarks(ctx context.Context, query string, limit int) ([]Bookmark, error) {
	var allBookmarks []Bookmark
	startURL, err := buildURL(c.url, query, limit)
	if err != nil {
		return nil, err
	}
	url := startURL

	fetchPage := func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Token "+c.token)

		// #nosec G704
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch bookmarks: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var data BookmarksResponse
		err = json.Unmarshal(body, &data)
		if err != nil {
			return err
		}

		allBookmarks = append(allBookmarks, data.Results...)
		url = data.Next // continue to next page if available
		return nil
	}

	for url != "" {
		if err := fetchPage(); err != nil {
			return nil, err
		}
	}

	return allBookmarks, nil
}

func (c *LinkdingClient) GetAllTags(ctx context.Context) ([]Tag, error) {
	var allTags []Tag

	u, err := url.Parse(c.url + "/api/tags/")
	if err != nil {
		return nil, err
	}

	url := u.String()

	fetchPage := func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Token "+c.token)

		// #nosec G704
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch tags: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var data TagsResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}

		allTags = append(allTags, data.Results...)
		url = data.Next // Follow pagination
		return nil
	}

	for url != "" {
		if err := fetchPage(); err != nil {
			return nil, err
		}

	}

	sort.Slice(allTags, func(i, j int) bool {
		return allTags[i].Name < allTags[j].Name
	})

	return allTags, nil
}

// BookmarkUpdate holds the fields to change on a bookmark, nil fields are kept.
type BookmarkUpdate struct {
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	TagNames    *[]string `json:"tag_names,omitempty"`
}

// GetBookmark returns the bookmark with the given id.
func (c *LinkdingClient) GetBookmark(ctx context.Context, id int64) (*Bookmark, error) {
	var bookmark Bookmark
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/bookmarks/%d/", id), nil, http.StatusOK, &bookmark); err != nil {
		return nil, err
	}
	return &bookmark, nil
}

// CheckBookmark returns the bookmark for the given URL, nil if it is not bookmarked.
func (c *LinkdingClient) CheckBookmark(ctx context.Context, bookmarkURL string) (*Bookmark, error) {
	var resp struct {
		Bookmark *Bookmark `json:"bookmark"`
	}
	path := "/api/bookmarks/check/?" + url.Values{"url": {bookmarkURL}}.Encode()
	if err := c.do(ctx, http.MethodGet, path, nil, http.StatusOK, &resp); err != nil {
		return nil, err
	}
	return resp.Bookmark, nil
}

// UpdateBookmark changes the fields of update on the bookmark with the given id
// and returns the updated bookmark.
func (c *LinkdingClient) UpdateBookmark(ctx context.Context, id int64, update BookmarkUpdate) (*Bookmark, error) {
	var bookmark Bookmark
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/api/bookmarks/%d/", id), update, http.StatusOK, &bookmark); err != nil {
		return nil, err
	}
	return &bookmark, nil
}

// DeleteBookmark deletes the bookmark with the given id.
func (c *LinkdingClient) DeleteBookmark(ctx context.Context, id int64) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/bookmarks/%d/", id), nil, http.StatusNoContent, nil)
}

// do sends a request with body encoded as JSON, if not nil, and decodes the
// response into out, if not nil.
func (c *LinkdingClient) do(ctx context.Context, method, path string, body any, wantStatus int, out any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.url+path, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Token "+c.token)

	// #nosec G704
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != wantStatus {
		return fmt.Errorf("%s %s: expected http code %d, got %s", method, path, wantStatus, resp.Status)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
