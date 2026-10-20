package cmd

import (
	"slices"
	"testing"

	"github.com/soerenschneider/sc/internal/linkding"
	"github.com/spf13/cobra"
)

func TestParseTags(t *testing.T) {
	got := parseTags("go #cli,go", "", " web ")
	if want := []string{"go", "cli", "web"}; !slices.Equal(got, want) {
		t.Errorf("parseTags() = %v, want %v", got, want)
	}
}

func TestLinksUpdateFromFlags(t *testing.T) {
	bookmark := &linkding.Bookmark{Title: "Old", Notes: "keep", TagNames: []string{"inbox", "go"}}

	tests := map[string]struct {
		args      []string
		wantTitle *string
		wantTags  []string // nil if the tags must not change
	}{
		"tags": {
			args:     []string{"--add-tag", "cli,go", "--remove-tag", "#inbox"},
			wantTags: []string{"go", "cli"},
		},
		"unchanged": {
			args: []string{"--add-tag", "go", "--title", "Old", "--notes", "keep"},
		},
		"clear title": {
			args:      []string{"--title", ""},
			wantTitle: new(string),
		},
		"remove all tags": {
			args:     []string{"-r", "inbox", "-r", "go"},
			wantTags: []string{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String(linksTitle, "", "")
			cmd.Flags().String(linksDescription, "", "")
			cmd.Flags().String(linksNotes, "", "")
			cmd.Flags().StringArrayP(linksAddTag, "g", nil, "")
			cmd.Flags().StringArrayP(linksRemoveTag, "r", nil, "")
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatal(err)
			}

			update := linksUpdateFromFlags(cmd, bookmark)
			if (update.Title == nil) != (tt.wantTitle == nil) || (update.Title != nil && *update.Title != *tt.wantTitle) {
				t.Errorf("Title = %v, want %v", update.Title, tt.wantTitle)
			}
			if update.Notes != nil {
				t.Errorf("Notes = %q, want unchanged", *update.Notes)
			}
			switch {
			case tt.wantTags == nil && update.TagNames != nil:
				t.Errorf("TagNames = %v, want unchanged", *update.TagNames)
			case tt.wantTags != nil && (update.TagNames == nil || !slices.Equal(*update.TagNames, tt.wantTags)):
				t.Errorf("TagNames = %v, want %v", update.TagNames, tt.wantTags)
			}
		})
	}
	if !slices.Equal(bookmark.TagNames, []string{"inbox", "go"}) {
		t.Errorf("bookmark tags were modified: %v", bookmark.TagNames)
	}
}
