package cmd

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/linkding"
	"github.com/spf13/cobra"
)

const (
	linksTitle       = "title"
	linksDescription = "description"
	linksNotes       = "notes"
	linksAddTag      = "add-tag"
	linksRemoveTag   = "remove-tag"
)

var linksEditCmd = &cobra.Command{
	Use:     "edit [id|url|query]...",
	Aliases: []string{"tag", "e"},
	Short:   "Edit the title, description, notes or tags of a bookmark",
	Long: `Edit a bookmark in Linkding. The bookmark is given by its id, its URL or a
search query. If the query matches several bookmarks, or none is given, one is
picked interactively.

The fields to change are given as flags. Without any of them, a form to edit
all fields is shown. Tags may be separated by spaces or commas.

Examples:
  sc links edit 42                                  # Edit in a form
  sc links tag 42 --add-tag go --remove-tag inbox   # Change tags only
  sc links edit https://example.com --title "Example"`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustLinkdingClient(cmd)
		bookmark := mustResolveBookmark(cmd, client, args)

		var update linkding.BookmarkUpdate
		if cmd.Flags().Changed(linksTitle) || cmd.Flags().Changed(linksDescription) || cmd.Flags().Changed(linksNotes) ||
			cmd.Flags().Changed(linksAddTag) || cmd.Flags().Changed(linksRemoveTag) {
			update = linksUpdateFromFlags(cmd, bookmark)
		} else {
			var err error
			update, err = linksUpdateFromForm(bookmark)
			if errors.Is(err, huh.ErrUserAborted) {
				os.Exit(0)
			}
			if err != nil {
				log.Fatal().Err(err).Msg("could not display form")
			}
		}

		if update == (linkding.BookmarkUpdate{}) {
			log.Info().Msg("Nothing to change")
			return
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), linksDefaultTimeout)
		defer cancel()
		updated, err := client.UpdateBookmark(ctx, bookmark.ID, update)
		if err != nil {
			log.Fatal().Err(err).Msg("could not update bookmark")
		}
		log.Info().Msgf("Updated bookmark %d, %s, tags: %s", updated.ID, updated.DisplayName(), strings.Join(updated.TagNames, " "))
	},
}

// linksUpdateFromFlags returns the changes given as flags, leaving out fields
// that would not change.
func linksUpdateFromFlags(cmd *cobra.Command, bookmark *linkding.Bookmark) linkding.BookmarkUpdate {
	var update linkding.BookmarkUpdate
	setIfChanged := func(flag, current string, field **string) {
		if !cmd.Flags().Changed(flag) {
			return
		}
		if value, _ := cmd.Flags().GetString(flag); value != current {
			*field = &value
		}
	}
	setIfChanged(linksTitle, bookmark.Title, &update.Title)
	setIfChanged(linksDescription, bookmark.Description, &update.Description)
	setIfChanged(linksNotes, bookmark.Notes, &update.Notes)

	add, _ := cmd.Flags().GetStringArray(linksAddTag)
	remove, _ := cmd.Flags().GetStringArray(linksRemoveTag)
	removeTags := parseTags(remove...)
	tags := slices.DeleteFunc(parseTags(append(slices.Clone(bookmark.TagNames), add...)...), func(tag string) bool {
		return slices.Contains(removeTags, tag)
	})
	if !slices.Equal(tags, bookmark.TagNames) {
		update.TagNames = &tags
	}
	return update
}

// linksUpdateFromForm lets the user edit all fields and returns the changed ones.
func linksUpdateFromForm(bookmark *linkding.Bookmark) (linkding.BookmarkUpdate, error) {
	title, description, notes := bookmark.Title, bookmark.Description, bookmark.Notes
	tags := strings.Join(bookmark.TagNames, " ")

	if err := huh.NewForm(huh.NewGroup(
		huh.NewNote().Title(bookmark.URL),
		huh.NewInput().Title("Title").Placeholder(bookmark.WebsiteTitle).Value(&title),
		huh.NewText().Title("Description").Placeholder(bookmark.WebsiteDescription).Value(&description),
		huh.NewText().Title("Notes").Value(&notes),
		huh.NewInput().Title("Tags").Description("Separated by spaces").Value(&tags),
	)).Run(); err != nil {
		return linkding.BookmarkUpdate{}, err
	}

	var update linkding.BookmarkUpdate
	if title != bookmark.Title {
		update.Title = &title
	}
	if description != bookmark.Description {
		update.Description = &description
	}
	if notes != bookmark.Notes {
		update.Notes = &notes
	}
	if newTags := parseTags(tags); !slices.Equal(newTags, bookmark.TagNames) {
		update.TagNames = &newTags
	}
	return update, nil
}

func init() {
	linksCmd.AddCommand(linksEditCmd)

	linksEditCmd.Flags().String(linksTitle, "", "New title")
	linksEditCmd.Flags().String(linksDescription, "", "New description")
	linksEditCmd.Flags().String(linksNotes, "", "New notes")
	linksEditCmd.Flags().StringArrayP(linksAddTag, "g", nil, "Tag to add, may be given multiple times")
	linksEditCmd.Flags().StringArrayP(linksRemoveTag, "r", nil, "Tag to remove, may be given multiple times")
}
