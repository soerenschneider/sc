package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/mpd"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/spf13/cobra"
)

const (
	mpdSearchTag = "tag"
	mpdSearchAdd = "add"
)

var mpdSearchCmd = &cobra.Command{
	Use:   "search <query>...",
	Short: "Search the database",
	Long: `Search the database for songs whose tag contains the query, ignoring case.
All arguments are joined to a single query. By default all tags and the file
name are searched, --tag restricts the search to a single tag such as artist,
album, title or file.

With --add, the songs found are appended to the queue.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		tag, _ := cmd.Flags().GetString(mpdSearchTag)
		add, _ := cmd.Flags().GetBool(mpdSearchAdd)

		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		songs, err := client.Search(ctx, tag, query)
		if err != nil {
			log.Fatal().Err(err).Msg("could not search the database")
		}
		if len(songs) == 0 {
			log.Info().Msgf("No songs found for %q", query)
			return
		}

		var total time.Duration
		for _, s := range songs {
			total += s.Duration
			line := "  " + s.DisplayName()
			if s.Album != "" {
				line += tui.MetaStyle.Render(" · " + s.Album)
			}
			if s.Duration > 0 {
				line += tui.MetaStyle.Render(" " + mpd.FormatDuration(s.Duration))
			}
			fmt.Println(line)
		}
		fmt.Println(tui.MetaStyle.Render(fmt.Sprintf("%d song(s) · %s", len(songs), mpd.FormatDuration(total))))

		if add {
			if err := client.SearchAdd(ctx, tag, query); err != nil {
				log.Fatal().Err(err).Msg("could not add the songs to the queue")
			}
			log.Info().Msgf("Added %d song(s) to the queue", len(songs))
		}
	},
}

func init() {
	mpdCmd.AddCommand(mpdSearchCmd)

	mpdSearchCmd.Flags().StringP(mpdSearchTag, "t", mpd.SearchTagAny, "Tag to search, e.g. artist, album, title or file")
	mpdSearchCmd.Flags().Bool(mpdSearchAdd, false, "Append the songs found to the queue")
	_ = mpdSearchCmd.RegisterFlagCompletionFunc(mpdSearchTag, cobra.FixedCompletions(
		[]string{mpd.SearchTagAny, "artist", "albumartist", "album", "title", "genre", "date", "composer", "file"},
		cobra.ShellCompDirectiveNoFileComp,
	))
}
