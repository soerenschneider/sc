package tui

import (
	"encoding/json"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderMap(t *testing.T) {
	data := map[string]any{
		"token_ttl":             json.Number("3600"),
		"token_policies":        []any{"default", "my-policy"},
		"secret_id_num_uses":    json.Number("0"),
		"secret_id_bound_cidrs": []any{},
		"meta":                  map[string]any{"b": "2", "a": "1"},
		"role_id":               "abc",
		"token_type":            "",
		"bind_secret_id":        true,
	}

	want := `bind_secret_id         true
meta                   a=1, b=2
role_id                abc
secret_id_bound_cidrs  -
secret_id_num_uses     0
token_policies         default, my-policy
token_ttl              1h0m0s
token_type             -`
	if got := ansi.Strip(RenderMap(data)); got != want {
		t.Errorf("RenderMap() =\n%s\nwant\n%s", got, want)
	}
}
