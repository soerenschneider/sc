package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/mpd"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const (
	mpdAddress  = "address"
	mpdPassword = "password"

	mpdDefaultTimeout = 10 * time.Second
)

var mpdCmd = &cobra.Command{
	Use:   "mpd",
	Short: "Control a Music Player Daemon",
	Long: `Control playback and browse the database of a Music Player Daemon.

The address defaults to MPD_HOST and MPD_PORT like for mpc, and then to
` + mpd.DefaultAddress + `. Paths are used as unix sockets, and a password may be given
as "password@host".

Examples:
  sc mpd status
  sc mpd queue
  sc mpd toggle
  sc mpd next
  sc mpd search -t artist radiohead
  sc mpd random off
  sc mpd update -w
  sc mpd browse Artist/Album
  sc mpd -a music.lan:6600 clear`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(mpdCmd)

	mpdCmd.PersistentFlags().StringP(mpdAddress, "a", "", "MPD address, host[:port] or path of a unix socket")
	mpdCmd.PersistentFlags().String(mpdPassword, "", "MPD password")
}

func mustDialMpd(cmd *cobra.Command) *mpd.Client {
	ctx, cancel := context.WithTimeout(context.Background(), mpdDefaultTimeout)
	defer cancel()

	client, err := mpd.Dial(ctx, pkg.GetString(cmd, mpdAddress), pkg.GetString(cmd, mpdPassword))
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to mpd")
	}
	return client
}

// mpdLogNowPlaying logs the player state and the current song.
func mpdLogNowPlaying(ctx context.Context, client *mpd.Client) {
	status, err := client.Status(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not get mpd status")
	}
	switch status.State {
	case mpd.StateStop:
		log.Info().Msg("Stopped")
		return
	case mpd.StatePause:
		log.Info().Msg("Paused")
		return
	}
	song, err := client.CurrentSong(ctx)
	if err != nil || song == nil {
		log.Info().Msg("Playing")
		return
	}
	log.Info().Msgf("Playing %s", song.DisplayName())
}

// mpdParseSwitch parses an optional on/off argument, toggling current if
// there is none.
func mpdParseSwitch(args []string, current bool) (bool, error) {
	if len(args) == 0 {
		return !current, nil
	}
	switch strings.ToLower(args[0]) {
	case "on", "1", "true", "yes":
		return true, nil
	case "off", "0", "false", "no":
		return false, nil
	}
	return false, fmt.Errorf("invalid value %q, expected on or off", args[0])
}

func mpdOnOff(on bool) string {
	if on {
		return "on"
	}
	return "off"
}
