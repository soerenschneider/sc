package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mpdNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Play the next song in the queue",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		if err := client.Next(ctx); err != nil {
			log.Fatal().Err(err).Msg("could not skip to the next song")
		}
		mpdLogNowPlaying(ctx, client)
	},
}

func init() {
	mpdCmd.AddCommand(mpdNextCmd)
}
