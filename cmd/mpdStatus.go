package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/mpd"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/spf13/cobra"
)

var mpdStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show the current song and player state",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustDialMpd(cmd)
		defer func() {
			_ = client.Close()
		}()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		status, err := client.Status(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get mpd status")
		}
		song, err := client.CurrentSong(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get current song")
		}

		fmt.Println(mpd.NowPlaying(status, song))
		if song != nil && song.Album != "" {
			fmt.Println(tui.MetaStyle.Render("  " + song.Album))
		}

		var details []string
		if status.Volume >= 0 {
			details = append(details, fmt.Sprintf("volume %d%%", status.Volume))
		}
		details = append(details,
			"repeat "+mpdOnOff(status.Repeat),
			"random "+mpdOnOff(status.Random),
		)
		fmt.Println(tui.MetaStyle.Render("  " + strings.Join(details, " · ")))
		if status.Error != "" {
			fmt.Println(tui.ErrorStyle.Render("  " + status.Error))
		}
	},
}

func init() {
	mpdCmd.AddCommand(mpdStatusCmd)
}
