package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mpdClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all songs from the queue",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), mpdDefaultTimeout)
		defer cancel()

		status, err := client.Status(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get mpd status")
		}
		if err := client.Clear(ctx); err != nil {
			log.Fatal().Err(err).Msg("could not clear the queue")
		}
		log.Info().Msgf("Cleared the queue (%d songs)", status.QueueLength)
	},
}

func init() {
	mpdCmd.AddCommand(mpdClearCmd)
}
