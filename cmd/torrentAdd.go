package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/transmission"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var torrentAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new torrent to the Transmission client",
	Long: `This command adds a new torrent to the Transmission client using the Transmission API.

You can provide a magnet link or a torrent file URL to initiate a download.

Examples:
  sc torrent add -m "magnet:?xt=urn:btih:..."
  sc torrent add -m "https://example.com/my.torrent"`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := pkg.GetString(cmd, torrentAddr)
		client, err := transmission.NewClient(addr)
		if err != nil {
			log.Fatal().Err(err).Msg("could not build transmission client")
		}

		magnet := pkg.GetString(cmd, torrentMagnet)
		if magnet == "" {
			validation := func(input string) error {
				if strings.HasPrefix(input, "magnet:") {
					return nil
				}

				return errors.New("this magnet seems to be broken")
			}
			magnet = tui.ReadInputWithValidation("Please enter magnet link", nil, validation)
		}

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		if err := client.AddTorrent(ctx, magnet); err != nil {
			log.Fatal().Err(err).Msg("could not add torrent")
		}
	},
}

func init() {
	torrentCmd.AddCommand(torrentAddCmd)

	torrentAddCmd.Flags().StringP(torrentMagnet, "m", "", "Magnet to add")
}
