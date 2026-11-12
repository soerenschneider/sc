package cmd

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/soerenschneider/sc/internal/vault"
)

func TestEditorSaveKeepsJSONValues(t *testing.T) {
	store := &fakeKv2Store{}
	m := newModel(store, vaultKv2EditOptions{Mount: "secret", Path: "app", Timeout: vaultDefaultTimeout})

	sec := &vault.Secret{Version: 3}
	sec.Data, sec.JSONKeys = vault.EncodeValues(map[string]any{
		"user":  "admin",
		"hosts": []any{"a", "b"},
		"port":  json.Number("5432"),
		"tls":   map[string]any{"enabled": true},
	})
	m.applyServerSecret(sec, true)

	// Edit two values like the pair form does, one into invalid JSON.
	for i, p := range m.pairs {
		switch p.key {
		case "hosts":
			m.pairs[i].value = `["a","b","c"]`
		case "port":
			m.pairs[i].value = "not a number"
		}
	}
	m.pairs = append(m.pairs, pair{key: "new", value: `["looks","like","json"]`})

	msg := m.save()()
	if errMsg, ok := msg.(errMsg); ok {
		t.Fatal(errMsg.err)
	}
	want := map[string]any{
		"user":  "admin",
		"hosts": []any{"a", "b", "c"},
		"port":  "not a number",
		"tls":   map[string]any{"enabled": true},
		"new":   `["looks","like","json"]`,
	}
	if !reflect.DeepEqual(store.put, want) {
		t.Errorf("expected %#v, got %#v", want, store.put)
	}
	if store.putCas != 3 {
		t.Errorf("expected cas 3, got %d", store.putCas)
	}
}

func TestMarshalSecretJSONKeepsTypes(t *testing.T) {
	body, err := marshalSecretJSON(vault.DecodeValues(
		map[string]string{"user": "admin", "hosts": `["a","b"]`, "port": "5432"},
		map[string]bool{"hosts": true, "port": true},
	))
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"hosts\": [\n    \"a\",\n    \"b\"\n  ],\n  \"port\": 5432,\n  \"user\": \"admin\"\n}"
	if string(body) != want {
		t.Errorf("expected\n%s\ngot\n%s", want, body)
	}
}
