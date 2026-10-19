package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mpdRandomCmd = &cobra.Command{
	Use:       "random [on|off]",
	Short:     "Toggle or set random mode",
	Long:      `Turn random mode, which plays the queue in random order, on or off. Without an argument it is toggled.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"on", "off"},
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := mpdParseSwitch(args, false); err != nil {
			log.Fatal().Err(err).Msg("could not set random mode")
		}

		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		status, err := client.Status(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get mpd status")
		}
		on, err := mpdParseSwitch(args, status.Random)
		if err != nil {
			log.Fatal().Err(err).Msg("could not set random mode")
		}
		if err := client.SetRandom(ctx, on); err != nil {
			log.Fatal().Err(err).Msg("could not set random mode")
		}
		log.Info().Msgf("Random mode is %s", mpdOnOff(on))
	},
}

func init() {
	mpdCmd.AddCommand(mpdRandomCmd)
}
