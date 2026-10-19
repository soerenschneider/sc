package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/mpd"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const mpdQueuePrint = "print"

var mpdQueueCmd = &cobra.Command{
	Use:     "queue",
	Aliases: []string{"q"},
	Short:   "Show and edit the queue",
	Long: `Show the queue in a TUI, where enter plays the selected song, d removes it,
J/K move it down and up, pressing c twice clears the queue and tab switches
to the library.

If the output is not a terminal or with --print, the queue is printed instead,
marking the current song.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustDialMpd(cmd)
		defer client.Close()

		if print, _ := cmd.Flags().GetBool(mpdQueuePrint); !print && term.IsTerminal(int(os.Stdout.Fd())) {
			if err := mpd.BrowseQueue(cmd.Context(), client); err != nil {
				log.Fatal().Err(err).Msg("could not show the queue")
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), mpdDefaultTimeout)
		defer cancel()

		status, err := client.Status(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get mpd status")
		}
		songs, err := client.Queue(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get the queue")
		}
		if len(songs) == 0 {
			log.Info().Msg("The queue is empty")
			return
		}

		var total time.Duration
		for _, s := range songs {
			total += s.Duration
			line := fmt.Sprintf("%3d  %s", s.Pos+1, s.DisplayName())
			if s.Duration > 0 {
				line += tui.MetaStyle.Render(" " + mpd.FormatDuration(s.Duration))
			}
			if icon := mpd.StateIcon(status.State); icon != "" && s.Pos == status.Song {
				fmt.Println(tui.InfoStyle.Render(icon + " " + line))
			} else {
				fmt.Println("  " + line)
			}
		}
		fmt.Println(tui.MetaStyle.Render(fmt.Sprintf("%d song(s) · %s", len(songs), mpd.FormatDuration(total))))
	},
}

func init() {
	mpdCmd.AddCommand(mpdQueueCmd)

	mpdQueueCmd.Flags().BoolP(mpdQueuePrint, "p", false, "Print the queue instead of opening the TUI")
}
