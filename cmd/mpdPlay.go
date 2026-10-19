package cmd

import (
	"context"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mpdPlayCmd = &cobra.Command{
	Use:   "play [position]",
	Short: "Start or resume playback",
	Long: `Resume playback if paused, otherwise start playing the queue. With a
position, play the song at that position of the queue, starting at 1 as
shown by "sc mpd queue".`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pos := -1
		if len(args) == 1 {
			n, err := strconv.Atoi(args[0])
			if err != nil || n < 1 {
				log.Fatal().Msgf("invalid position %q, expected a number from 1", args[0])
			}
			pos = n - 1
		}

		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		status, err := client.Status(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get mpd status")
		}
		if status.QueueLength == 0 {
			log.Fatal().Msg("The queue is empty")
		}
		if pos >= status.QueueLength {
			log.Fatal().Msgf("The queue has only %d song(s)", status.QueueLength)
		}
		if err := client.Play(ctx, pos); err != nil {
			log.Fatal().Err(err).Msg("could not start playback")
		}
		mpdLogNowPlaying(ctx, client)
	},
}

func init() {
	mpdCmd.AddCommand(mpdPlayCmd)
}
