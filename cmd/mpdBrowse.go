package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/mpd"
	"github.com/spf13/cobra"
)

var mpdBrowseCmd = &cobra.Command{
	Use:     "browse [path]",
	Aliases: []string{"ls"},
	Short:   "Browse the database in a TUI and control playback",
	Long: `Browse the music database starting at path, the root by default.

Enter opens a folder or plays a song, a adds the selected folder or song to the
queue, space toggles play/pause and pressing c twice clears the queue. Tab
switches to the queue, see "sc mpd queue".`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := ""
		if len(args) == 1 {
			dir = args[0]
		}

		client := mustDialMpd(cmd)
		defer client.Close()

		if err := mpd.Browse(cmd.Context(), client, dir); err != nil {
			log.Fatal().Err(err).Msg("could not browse mpd database")
		}
	},
}

func init() {
	mpdCmd.AddCommand(mpdBrowseCmd)
}
