package cmd

import (
	"context"
	"net/url"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/linkding"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var linksAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Interact with VictoriaLogs for querying and managing links",
	Long: `The 'links' command is the entry point for interacting with VictoriaLogs.

Use subcommands to query, tail, or manage links stored in VictoriaLogs, a high-performance log database.

Examples:
  sc links query -q='level="error"'   # Query links with specific filters`,
	Run: func(cmd *cobra.Command, args []string) {
		address := pkg.GetString(cmd, linksAddr)

		bookmark := linkding.Bookmark{
			URL:      pkg.GetString(cmd, linksUrl),
			TagNames: pkg.GetStringArray(cmd, linksTags),
		}

		token := pkg.GetString(cmd, linksToken)

		client, err := linkding.NewLinkdingClient(address, token)
		if err != nil {
			log.Fatal().Err(err).Msg("could not build client")
		}

		if bookmark.URL == "" {
			urlValidation := func(input string) error {
				_, err := url.Parse(input)
				return err
			}
			bookmark.URL = tui.ReadInputWithValidation("URL", nil, urlValidation)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var id int64

		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				var err error
				id, err = client.AddBookmark(ctx, bookmark)
				return err
			}).
			Title("Adding bookmark...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("could not add bookmark")
		}

		log.Info().Msgf("Upserted bookmark with ID %d", id)
	},
}

func init() {
	linksCmd.AddCommand(linksAddCmd)

	linksAddCmd.PersistentFlags().StringP(linksUrl, "u", "", "URL to add")
	linksAddCmd.PersistentFlags().StringArrayP(linksTags, "", nil, "Tags for the URL")
}
