package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mpdPrevCmd = &cobra.Command{
	Use:     "prev",
	Aliases: []string{"previous"},
	Short:   "Play the previous song in the queue",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		if err := client.Previous(ctx); err != nil {
			log.Fatal().Err(err).Msg("could not go back to the previous song")
		}
		mpdLogNowPlaying(ctx, client)
	},
}

func init() {
	mpdCmd.AddCommand(mpdPrevCmd)
}
