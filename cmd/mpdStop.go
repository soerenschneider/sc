package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mpdStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop playback",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		if err := client.Stop(ctx); err != nil {
			log.Fatal().Err(err).Msg("could not stop playback")
		}
		log.Info().Msg("Stopped")
	},
}

func init() {
	mpdCmd.AddCommand(mpdStopCmd)
}
