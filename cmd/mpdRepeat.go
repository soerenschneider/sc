package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mpdRepeatCmd = &cobra.Command{
	Use:       "repeat [on|off]",
	Short:     "Toggle or set repeat mode",
	Long:      `Turn repeat mode, which starts the queue over after the last song, on or off. Without an argument it is toggled.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"on", "off"},
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := mpdParseSwitch(args, false); err != nil {
			log.Fatal().Err(err).Msg("could not set repeat mode")
		}

		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		status, err := client.Status(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get mpd status")
		}
		on, err := mpdParseSwitch(args, status.Repeat)
		if err != nil {
			log.Fatal().Err(err).Msg("could not set repeat mode")
		}
		if err := client.SetRepeat(ctx, on); err != nil {
			log.Fatal().Err(err).Msg("could not set repeat mode")
		}
		log.Info().Msgf("Repeat mode is %s", mpdOnOff(on))
	},
}

func init() {
	mpdCmd.AddCommand(mpdRepeatCmd)
}
