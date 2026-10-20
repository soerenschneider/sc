package linkding

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type request struct {
	method, uri, body string
}

func newTestClient(t *testing.T, handler func(w http.ResponseWriter, r request)) (*LinkdingClient, *[]request) {
	t.Helper()
	var requests []request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Token secret" {
			t.Errorf("Authorization = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		req := request{r.Method, r.URL.RequestURI(), string(body)}
		requests = append(requests, req)
		handler(w, req)
	}))
	t.Cleanup(srv.Close)
	client, _ := NewLinkdingClient(srv.URL, "secret")
	return client, &requests
}

func TestGetBookmark(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r request) {
		switch r.uri {
		case "/api/bookmarks/7/":
			_, _ = io.WriteString(w, `{"id": 7, "url": "https://example.com", "website_title": "Example", "tag_names": ["a"]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	b, err := client.GetBookmark(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != 7 || b.DisplayName() != "Example" || len(b.TagNames) != 1 {
		t.Errorf("GetBookmark() = %+v", b)
	}

	if _, err := client.GetBookmark(context.Background(), 8); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetBookmark(missing) error = %v, want ErrNotFound", err)
	}
}

func TestCheckBookmark(t *testing.T) {
	client, requests := newTestClient(t, func(w http.ResponseWriter, r request) {
		if r.uri == "/api/bookmarks/check/?url=https%3A%2F%2Fexample.com%2F%3Fa%3D1" {
			_, _ = io.WriteString(w, `{"bookmark": {"id": 3, "url": "https://example.com/?a=1"}, "auto_tags": []}`)
			return
		}
		_, _ = io.WriteString(w, `{"bookmark": null, "auto_tags": []}`)
	})

	b, err := client.CheckBookmark(context.Background(), "https://example.com/?a=1")
	if err != nil || b == nil || b.ID != 3 {
		t.Errorf("CheckBookmark() = %+v, %v (requests %v)", b, err, *requests)
	}
	b, err = client.CheckBookmark(context.Background(), "https://other.example.com")
	if err != nil || b != nil {
		t.Errorf("CheckBookmark(unknown) = %+v, %v", b, err)
	}
}

func TestUpdateBookmark(t *testing.T) {
	client, requests := newTestClient(t, func(w http.ResponseWriter, r request) {
		_, _ = io.WriteString(w, `{"id": 7, "title": "New", "tag_names": []}`)
	})

	title := "New"
	tags := []string{}
	b, err := client.UpdateBookmark(context.Background(), 7, BookmarkUpdate{Title: &title, TagNames: &tags})
	if err != nil || b.Title != "New" {
		t.Fatalf("UpdateBookmark() = %+v, %v", b, err)
	}

	r := (*requests)[0]
	if r.method != http.MethodPatch || r.uri != "/api/bookmarks/7/" {
		t.Errorf("request = %s %s", r.method, r.uri)
	}
	// only the given fields are sent, an empty tag list clears the tags
	var body map[string]any
	if err := json.Unmarshal([]byte(r.body), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 2 || body["title"] != "New" || len(body["tag_names"].([]any)) != 0 {
		t.Errorf("body = %s", r.body)
	}
}

func TestDeleteBookmark(t *testing.T) {
	client, requests := newTestClient(t, func(w http.ResponseWriter, r request) {
		if r.uri == "/api/bookmarks/7/" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := client.DeleteBookmark(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if r := (*requests)[0]; r.method != http.MethodDelete {
		t.Errorf("method = %s", r.method)
	}
	if err := client.DeleteBookmark(context.Background(), 8); err == nil {
		t.Error("DeleteBookmark() succeeded on a server error")
	}
}
