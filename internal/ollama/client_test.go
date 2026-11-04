package ollama

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestNormalizeAddress(t *testing.T) {
	tests := map[string]string{
		"localhost":               "http://localhost:11434",
		"gpu.lan:8080":            "http://gpu.lan:8080",
		"http://gpu.lan":          "http://gpu.lan:11434",
		"https://ollama.example/": "https://ollama.example",
		"http://10.0.0.1:11434":   "http://10.0.0.1:11434",
		"https://example/ollama/": "https://example/ollama",
		"0.0.0.0":                 "http://0.0.0.0:11434",
	}
	for in, want := range tests {
		got, err := normalizeAddress(in)
		if err != nil || got != want {
			t.Errorf("normalizeAddress(%q) = %q, %v, want %q", in, got, err, want)
		}
	}
}

func TestNewClientUsesEnv(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "gpu.lan")
	client, err := NewClient("")
	if err != nil || client.Address() != "http://gpu.lan:11434" {
		t.Errorf("NewClient() = %v, %v", client.Address(), err)
	}
}

func TestListModels(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/tags" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"models": [{"name": "llama3.2:latest", "size": 2019393189, "digest": "sha256:a80c4f17acd55265feec403c7aef86be0c25983ab279d83f3bcd3abbcb5b8b72", "details": {"parameter_size": "3.2B", "quantization_level": "Q4_K_M", "family": "llama"}}]}`)
	})

	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].Name != "llama3.2:latest" || models[0].Details.ParameterSize != "3.2B" {
		t.Errorf("ListModels() = %+v", models)
	}

	_, rows := ModelsAsTable(models)
	if rows[0][5] != "a80c4f17acd5" {
		t.Errorf("ID = %q", rows[0][5])
	}
}

func TestDeleteModel(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != http.MethodDelete || r.URL.Path != "/api/delete" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if string(body) == `{"model":"missing"}` {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"error": "model 'missing' not found"}`)
		}
	})

	if err := client.DeleteModel(context.Background(), "llama3.2"); err != nil {
		t.Errorf("DeleteModel() = %v", err)
	}
	err := client.DeleteModel(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), "model 'missing' not found") {
		t.Errorf("DeleteModel(missing) = %v", err)
	}
}

func TestPullModel(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"model":"llama3.2"`) {
			_, _ = io.WriteString(w, `{"status":"pulling manifest"}`+"\n"+`{"error":"pull model manifest: file does not exist"}`+"\n")
			return
		}
		_, _ = io.WriteString(w, `{"status":"pulling manifest"}
{"status":"pulling dde5aa3fc5ff","digest":"sha256:dde5aa3fc5ff","total":100,"completed":50}
{"status":"pulling dde5aa3fc5ff","digest":"sha256:dde5aa3fc5ff","total":100,"completed":100}
{"status":"verifying sha256 digest"}
{"status":"success"}
`)
	})

	var updates []PullProgress
	if err := client.PullModel(context.Background(), "llama3.2", func(p PullProgress) { updates = append(updates, p) }); err != nil {
		t.Fatal(err)
	}
	if len(updates) != 5 || updates[1].Completed != 50 || updates[4].Status != "success" {
		t.Errorf("updates = %+v", updates)
	}

	if err := client.PullModel(context.Background(), "nope", nil); err == nil || !strings.Contains(err.Error(), "file does not exist") {
		t.Errorf("PullModel(nope) = %v", err)
	}
}

func TestPullModelTruncated(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"pulling manifest"}`+"\n")
	})
	if err := client.PullModel(context.Background(), "llama3.2", nil); err == nil {
		t.Error("PullModel() succeeded on a truncated stream")
	}
}

func TestShowModelDescribe(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"parameters": "stop \"<|eot_id|>\"", "details": {"family": "llama", "parameter_size": "3.2B"}, "model_info": {"general.architecture": "llama", "llama.context_length": 131072}, "capabilities": ["completion", "tools"]}`)
	})

	info, err := client.ShowModel(context.Background(), "llama3.2")
	if err != nil {
		t.Fatal(err)
	}
	out := info.Describe("llama3.2")
	for _, want := range []string{"llama3.2", "3.2B", "131072", "tools", "Parameters", `stop "<|eot_id|>"`} {
		if !strings.Contains(out, want) {
			t.Errorf("Describe() misses %q:\n%s", want, out)
		}
	}
}

func TestProcessor(t *testing.T) {
	tests := []struct {
		m    RunningModel
		want string
	}{
		{RunningModel{Size: 100}, "100% CPU"},
		{RunningModel{Size: 100, SizeVRAM: 100}, "100% GPU"},
		{RunningModel{Size: 100, SizeVRAM: 25}, "75%/25% CPU/GPU"},
	}
	for _, tt := range tests {
		if got := tt.m.Processor(); got != tt.want {
			t.Errorf("Processor(%+v) = %q, want %q", tt.m, got, tt.want)
		}
	}
}

func TestChat(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.URL.Path != "/api/chat" || !strings.Contains(string(body), `"stream":true`) || !strings.Contains(string(body), `"think":true`) {
			t.Errorf("unexpected request %s %s", r.URL.Path, body)
		}
		_, _ = io.WriteString(w, `{"message":{"role":"assistant","content":"","thinking":"hmm"},"done":false}
{"message":{"role":"assistant","content":"Hello"},"done":false}
{"message":{"role":"assistant","content":" world"},"done":false}
{"message":{"role":"assistant","content":""},"done":true}
`)
	})

	think := true
	var chunks int
	answer, err := client.Chat(context.Background(), ChatRequest{Model: "qwen3", Messages: []Message{{Role: "user", Content: "hi"}}, Think: &think}, func(Message) { chunks++ })
	if err != nil {
		t.Fatal(err)
	}
	if answer.Role != "assistant" || answer.Content != "Hello world" || answer.Thinking != "hmm" || chunks != 4 {
		t.Errorf("Chat() = %+v after %d chunks", answer, chunks)
	}
}

func TestChatErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"message":{"role":"assistant","content":"Hel"},"done":false}`+"\n")
	})
	if _, err := client.Chat(context.Background(), ChatRequest{Model: "qwen3"}, nil); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("Chat(truncated) = %v", err)
	}

	client = newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"\"llama3.2\" does not support thinking"}`)
	})
	if _, err := client.Chat(context.Background(), ChatRequest{Model: "llama3.2"}, nil); err == nil || !strings.Contains(err.Error(), "does not support thinking") {
		t.Errorf("Chat(bad request) = %v", err)
	}
}
