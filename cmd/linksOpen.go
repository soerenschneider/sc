package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var linksOpenCmd = &cobra.Command{
	Use:     "open [id|url|query]...",
	Aliases: []string{"o"},
	Short:   "Open a bookmark in the browser",
	Long: `Open a bookmark in the default browser. The bookmark is given by its id, its URL
or a search query. If the query matches several bookmarks, or none is given, one
is picked interactively.

Examples:
  sc links open 42
  sc links open '#golang' generics`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustLinkdingClient(cmd)
		bookmark := mustResolveBookmark(cmd, client, args)

		if err := pkg.OpenBrowser(bookmark.URL); err != nil {
			log.Fatal().Err(err).Msg("could not open browser")
		}
		log.Info().Msgf("Opened %s", bookmark.URL)
	},
}

func init() {
	linksCmd.AddCommand(linksOpenCmd)
}
