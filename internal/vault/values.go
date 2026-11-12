package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// EncodeValues converts the values of a KV secret into text. Strings are kept as they are, all
// other JSON values (numbers, booleans, null, lists and objects) are JSON-encoded and their keys
// are returned in jsonKeys, so DecodeValues can restore their type.
func EncodeValues(raw map[string]any) (data map[string]string, jsonKeys map[string]bool) {
	data = make(map[string]string, len(raw))
	jsonKeys = map[string]bool{}
	for k, v := range raw {
		if s, ok := v.(string); ok {
			data[k] = s
			continue
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			// Can't happen for values decoded from JSON.
			encoded = []byte(fmt.Sprintf("%v", v))
		}
		data[k] = string(encoded)
		jsonKeys[k] = true
	}
	return data, jsonKeys
}

// DecodeValues reverses EncodeValues. Values of keys in jsonKeys are decoded from JSON, keeping
// numbers as json.Number to not lose precision. If such a value is not valid JSON anymore, it is
// returned as string.
func DecodeValues(data map[string]string, jsonKeys map[string]bool) map[string]any {
	out := make(map[string]any, len(data))
	for k, v := range data {
		out[k] = v
		if !jsonKeys[k] || !json.Valid([]byte(v)) {
			continue
		}
		dec := json.NewDecoder(bytes.NewBufferString(v))
		dec.UseNumber()
		var decoded any
		if err := dec.Decode(&decoded); err == nil {
			out[k] = decoded
		}
	}
	return out
}
