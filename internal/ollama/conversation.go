package ollama

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Conversation is the history of an exchange with a model that can be
// continued later.
type Conversation struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Updated  time.Time `json:"updated"`
}

// DefaultConversationFile returns the file the last conversation is kept in.
func DefaultConversationFile() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sc", "ollama-conversation.json"), nil
}

// LoadConversation reads a conversation, returning os.ErrNotExist if there is
// none.
func LoadConversation(path string) (*Conversation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("decoding conversation %s: %w", path, err)
	}
	if len(conv.Messages) == 0 {
		return nil, os.ErrNotExist
	}
	return &conv, nil
}

// Save writes the conversation readable only by the user, as it may contain
// whatever was piped into it.
func (c *Conversation) Save(path string) error {
	if c.Model == "" {
		return errors.New("conversation has no model")
	}
	c.Updated = time.Now()
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
