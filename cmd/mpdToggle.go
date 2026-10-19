package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/mpd"
	"github.com/spf13/cobra"
)

var mpdToggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle between play and pause",
	Long: `Pause playback if playing, otherwise resume it or, if stopped, start
playing the queue.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), mpdDefaultTimeout)
		defer cancel()

		state, err := client.Toggle(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not toggle playback")
		}
		if state == mpd.StatePause {
			log.Info().Msg("Paused")
			return
		}
		song, err := client.CurrentSong(ctx)
		if err != nil || song == nil {
			log.Info().Msg("Playing")
			return
		}
		log.Info().Msgf("Playing %s", song.DisplayName())
	},
}

func init() {
	mpdCmd.AddCommand(mpdToggleCmd)
}
