package cmd

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	mpdUpdateWait = "wait"

	mpdUpdatePollInterval = 500 * time.Millisecond
)

var mpdUpdateCmd = &cobra.Command{
	Use:   "update [path]",
	Short: "Rescan the music directory for changes",
	Long: `Scan the music directory, or only path within it, for new, changed and
removed files. The scan runs in the background unless --wait is given.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := ""
		if len(args) == 1 {
			dir = args[0]
		}

		client := mustDialMpd(cmd)
		defer client.Close()

		ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
		defer cancel()

		id, err := client.Update(ctx, dir)
		if err != nil {
			log.Fatal().Err(err).Msg("could not start the database update")
		}

		if wait, _ := cmd.Flags().GetBool(mpdUpdateWait); !wait {
			log.Info().Msgf("Started database update #%d", id)
			return
		}

		log.Info().Msgf("Waiting for database update #%d", id)
		start := time.Now()
		ticker := time.NewTicker(mpdUpdatePollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-cmd.Context().Done():
				log.Fatal().Msg("Interrupted, the update continues in the background")
			case <-ticker.C:
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), mpdDefaultTimeout)
			status, err := client.Status(ctx)
			cancel()
			if err != nil {
				log.Fatal().Err(err).Msg("could not get mpd status")
			}
			// updates are queued, ours is done once a later one or none runs
			if status.UpdatingDB == 0 || status.UpdatingDB > id {
				log.Info().Msgf("Database update #%d finished after %s", id, time.Since(start).Round(time.Second))
				return
			}
		}
	},
}

func init() {
	mpdCmd.AddCommand(mpdUpdateCmd)

	mpdUpdateCmd.Flags().BoolP(mpdUpdateWait, "w", false, "Wait for the update to finish")
}
