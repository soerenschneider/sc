package internal

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func Test_getCommandHierarchy(t *testing.T) {
	type args struct {
		cmd string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "happy case",
			args: args{
				cmd: "a-b-c",
			},
			want: []string{
				"a-b-c",
				"a-b",
				"a",
			},
		},
		{
			name: "edge case empty",
			args: args{
				cmd: "",
			},
			want: []string{
				"",
			},
		},
		{
			name: "edge case",
			args: args{
				cmd: "a",
			},
			want: []string{
				"a",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCommandHierarchy(tt.args.cmd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCommandHierarchy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type group struct {
		Name      string   `yaml:"name"`
		Endpoints []string `yaml:"endpoints"`
	}
	type settings struct {
		Record string  `yaml:"record"`
		Groups []group `yaml:"groups"`
	}

	profileData = Profile{
		"prod": {
			"health": {
				"groups": []any{
					map[string]any{"name": "ssh", "endpoints": []any{"a:22", "b:22"}},
				},
			},
		},
	}
	t.Cleanup(func() { profileData = nil })

	got := settings{Record: "default"}
	found, err := Decode("health-sub", "prod", &got)
	if err != nil || !found {
		t.Fatalf("Decode() = %v, %v, want true, nil", found, err)
	}
	want := settings{Record: "default", Groups: []group{{Name: "ssh", Endpoints: []string{"a:22", "b:22"}}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Decode() decoded %+v, want %+v", got, want)
	}

	if found, err := Decode("health", "missing", &got); found || err != nil {
		t.Errorf("Decode() for missing profile = %v, %v, want false, nil", found, err)
	}
}

func TestApplyFlags(t *testing.T) {
	profileData = Profile{
		"prod": {
			"vault": {
				"address":  "https://vault.example.com",
				"username": "parent",
			},
			"vault-login": {
				"username": "jdoe",
				"tags":     []any{"a", "b,c"},
				"ids":      []any{1, 2},
				"verbose":  true,
				"bad":      map[string]any{"nested": "value"},
			},
		},
	}
	t.Cleanup(func() { profileData = nil })

	vault := &cobra.Command{Use: "vault"}
	vault.PersistentFlags().String("address", "", "")
	login := &cobra.Command{Use: "login", Run: func(*cobra.Command, []string) {}}
	login.Flags().String("username", "", "")
	login.Flags().String("mount", "userpass", "")
	login.Flags().StringArray("tags", nil, "")
	login.Flags().IntSlice("ids", nil, "")
	login.Flags().Bool("verbose", false, "")
	login.Flags().String("bad", "", "")
	vault.AddCommand(login)
	if err := login.ParseFlags([]string{"--mount", "ldap"}); err != nil {
		t.Fatal(err)
	}

	if err := ApplyFlags("vault-login", login, "prod"); err != nil {
		t.Fatal(err)
	}

	str := func(name string) string {
		v, _ := login.Flags().GetString(name)
		return v
	}
	if got := str("username"); got != "jdoe" {
		t.Errorf("username = %q, want the value of the most specific section", got)
	}
	if got := str("address"); got != "https://vault.example.com" {
		t.Errorf("address = %q, want the value of the parent section", got)
	}
	if got := str("mount"); got != "ldap" {
		t.Errorf("mount = %q, want the value given on the command line", got)
	}
	if got, _ := login.Flags().GetStringArray("tags"); !reflect.DeepEqual(got, []string{"a", "b,c"}) {
		t.Errorf("tags = %q", got)
	}
	if got, _ := login.Flags().GetIntSlice("ids"); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("ids = %v", got)
	}
	if got, _ := login.Flags().GetBool("verbose"); !got {
		t.Error("verbose not set")
	}
	if got := str("bad"); got != "" {
		t.Errorf("bad = %q, want unsupported values to be ignored", got)
	}
	if !login.Flags().Changed("tags") || !login.Flags().Changed("username") {
		t.Error("flags set from the profile are not marked as changed")
	}

	if err := ApplyFlags("vault-login", login, "missing"); !errors.Is(err, ErrProfileNotFound) {
		t.Errorf("missing profile: got %v", err)
	}
}

func TestGet(t *testing.T) {
	profileData = Profile{"prod": {
		"logs":       {"address": "https://logs.example.com", "id-fields": []any{"host", "app"}},
		"logs-query": {"limit": 50},
	}}
	t.Cleanup(func() { profileData = nil })

	for key, want := range map[string]string{"address": "https://logs.example.com", "id-fields": "host,app", "limit": "50"} {
		if got, ok := Get("logs-query", "prod", key); !ok || got != want {
			t.Errorf("Get(%q) = %q, %v, want %q", key, got, ok, want)
		}
	}
	if _, ok := Get("logs-query", "prod", "missing"); ok {
		t.Error("Get(missing) found a value")
	}
}
