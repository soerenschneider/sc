package cmd

import (
	"context"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const linksYes = "yes"

var linksDeleteCmd = &cobra.Command{
	Use:     "delete [id|url|query]...",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a bookmark",
	Long: `Delete a bookmark from Linkding. The bookmark is given by its id, its URL or a
search query. If the query matches several bookmarks, or none is given, one is
picked interactively.

The deletion has to be confirmed, unless --yes is given.

Examples:
  sc links delete 42
  sc links delete https://example.com --yes`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustLinkdingClient(cmd)
		bookmark := mustResolveBookmark(cmd, client, args)

		if yes, _ := cmd.Flags().GetBool(linksYes); !yes {
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				log.Fatal().Msgf("Not deleting %s without confirmation, use --%s", bookmark.URL, linksYes)
			}
			confirmed := false
			if err := huh.NewConfirm().
				Title(fmt.Sprintf("Delete %q?", bookmark.DisplayName())).
				Description(bookmark.URL).
				Value(&confirmed).
				Run(); err != nil || !confirmed {
				log.Info().Msg("Not deleting the bookmark")
				return
			}
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), linksDefaultTimeout)
		defer cancel()
		if err := client.DeleteBookmark(ctx, bookmark.ID); err != nil {
			log.Fatal().Err(err).Msg("could not delete bookmark")
		}
		log.Info().Msgf("Deleted bookmark %d, %s", bookmark.ID, bookmark.URL)
	},
}

func init() {
	linksCmd.AddCommand(linksDeleteCmd)

	linksDeleteCmd.Flags().BoolP(linksYes, "y", false, "Delete without asking for confirmation")
}
