package internal

import (
	"reflect"
	"testing"
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
