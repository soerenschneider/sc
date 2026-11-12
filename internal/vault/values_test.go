package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

// decodeJSON decodes like the Vault client does, keeping numbers as json.Number.
func decodeJSON(t *testing.T, s string) map[string]any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewBufferString(s))
	dec.UseNumber()
	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

const mixedSecret = `{
	"str": "s3cret",
	"json_str": "{\"a\":1}",
	"num": 42,
	"big": 12345678901234567890,
	"float": 1.5,
	"bool": true,
	"null": null,
	"list": ["a", "b"],
	"obj": {"user": "x", "port": 5432}
}`

func TestEncodeValues(t *testing.T) {
	data, jsonKeys := EncodeValues(decodeJSON(t, mixedSecret))

	wantData := map[string]string{
		"str":      "s3cret",
		"json_str": `{"a":1}`,
		"num":      "42",
		"big":      "12345678901234567890",
		"float":    "1.5",
		"bool":     "true",
		"null":     "null",
		"list":     `["a","b"]`,
		"obj":      `{"port":5432,"user":"x"}`,
	}
	if !maps.Equal(data, wantData) {
		t.Errorf("expected data %v, got %v", wantData, data)
	}
	wantKeys := map[string]bool{"num": true, "big": true, "float": true, "bool": true, "null": true, "list": true, "obj": true}
	if !maps.Equal(jsonKeys, wantKeys) {
		t.Errorf("expected json keys %v, got %v", wantKeys, jsonKeys)
	}
}

func TestDecodeValuesRoundTrip(t *testing.T) {
	raw := decodeJSON(t, mixedSecret)
	if got := DecodeValues(EncodeValues(raw)); !reflect.DeepEqual(got, raw) {
		t.Errorf("round trip changed the secret\nwant %#v\ngot  %#v", raw, got)
	}
}

func TestDecodeValues(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		jsonKey  bool
		expected any
	}{
		{"string stays string", "s3cret", false, "s3cret"},
		{"string looking like json stays string", `["a"]`, false, `["a"]`},
		{"edited list", `["a","b","c"]`, true, []any{"a", "b", "c"}},
		{"edited object", `{"port":5433}`, true, map[string]any{"port": json.Number("5433")}},
		{"number", "42", true, json.Number("42")},
		{"json string", `"quoted"`, true, "quoted"},
		{"invalid json becomes string", "[a b]", true, "[a b]"},
		{"trailing data becomes string", `{"a":1} x`, true, `{"a":1} x`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeValues(map[string]string{"k": tt.value}, map[string]bool{"k": tt.jsonKey})
			if !reflect.DeepEqual(got["k"], tt.expected) {
				t.Errorf("expected %#v, got %#v", tt.expected, got["k"])
			}
		})
	}
}

// TestGetPutRoundTrip reads a secret with non-string values and writes it back unchanged, the
// written payload must be identical to what has been read.
func TestGetPutRoundTrip(t *testing.T) {
	var written map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/data/app" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"data":{"data":` + mixedSecret + `,"metadata":{"version":3,"created_time":"2026-11-04T10:00:00Z"}}}`))
		case http.MethodPut, http.MethodPost:
			var body struct {
				Data    map[string]any `json:"data"`
				Options map[string]any `json:"options"`
			}
			dec := json.NewDecoder(r.Body)
			dec.UseNumber()
			_ = dec.Decode(&body)
			written = body.Data
			if cas := body.Options["cas"]; cas != json.Number("3") {
				t.Errorf("expected cas 3, got %v", cas)
			}
			_, _ = w.Write([]byte(`{"data":{"version":4,"created_time":"2026-11-04T10:01:00Z"}}`))
		}
	}))
	defer srv.Close()

	config := vault.DefaultConfig()
	config.Address = srv.URL
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	c := &VaultClient{client: client}

	sec, err := c.Get(context.Background(), "secret", "app")
	if err != nil {
		t.Fatal(err)
	}
	if sec.Data["list"] != `["a","b"]` || sec.Data["obj"] != `{"port":5432,"user":"x"}` {
		t.Errorf("expected JSON-encoded values, got %v", sec.Data)
	}

	if _, err := c.Put(context.Background(), "secret", "app", DecodeValues(sec.Data, sec.JSONKeys), sec.Version); err != nil {
		t.Fatal(err)
	}
	if want := decodeJSON(t, mixedSecret); !reflect.DeepEqual(written, want) {
		t.Errorf("written secret differs from the read one\nwant %#v\ngot  %#v", want, written)
	}
}
