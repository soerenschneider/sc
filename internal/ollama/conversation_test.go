package ollama

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConversation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sc", "conversation.json")

	if _, err := LoadConversation(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("LoadConversation(missing) = %v", err)
	}

	conv := &Conversation{Model: "qwen3", Messages: []Message{{Role: "user", Content: "hi"}, {Role: "assistant", Content: "hello"}}}
	if err := conv.Save(path); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("conversation file mode = %v, %v", info.Mode(), err)
	}

	loaded, err := LoadConversation(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Model != "qwen3" || len(loaded.Messages) != 2 || loaded.Messages[1].Content != "hello" || loaded.Updated.IsZero() {
		t.Errorf("LoadConversation() = %+v", loaded)
	}

	if err := (&Conversation{}).Save(path); err == nil {
		t.Error("Save() without model succeeded")
	}
}
