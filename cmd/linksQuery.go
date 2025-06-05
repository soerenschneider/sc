package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/linkding"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

var linksQueryCmd = &cobra.Command{
	Use:     "query",
	Aliases: []string{"q"},
	Short:   "Interact with VictoriaLogs for querying and managing links",
	Long: `The 'links' command is the entry point for interacting with VictoriaLogs.

Use subcommands to query, tail, or manage links stored in VictoriaLogs, a high-performance log database.

Examples:
  sc links query -q='level="error"'   # Query links with specific filters`,
	Run: func(cmd *cobra.Command, args []string) {
		address := pkg.GetString(cmd, linksAddr)
		query := pkg.GetString(cmd, linksQuery)
		token := pkg.GetString(cmd, linksToken)
		limit := 500

		client, err := linkding.NewLinkdingClient(address, token)
		if err != nil {
			log.Fatal().Err(err).Msg("could not build client")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if query == "" && !cmd.Flags().Changed(linksQuery) {
			var tags []linkding.Tag
			if err := spinner.New().
				Type(spinner.Line).
				ActionWithErr(func(ctx context.Context) error {
					var err error
					tags, err = client.GetAllTags(ctx)
					return err
				}).
				Title("Receiving tags...").
				Accessible(false).
				Context(ctx).
				Type(spinner.Dots).
				Run(); err != nil {
				log.Fatal().Err(err).Msg("could not get tags")
			}

			normalizedTags := make([]string, len(tags))
			for idx, tag := range tags {
				normalizedTags[idx] = fmt.Sprintf("#%s", tag.Name)
			}

			noValidation := func(input string) error {
				return nil
			}
			query = tui.ReadInputWithValidation("Query", normalizedTags, noValidation)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var bookmarks []linkding.Bookmark

		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				var err error
				bookmarks, err = client.ListBookmarks(ctx, query, limit)
				return err
			}).
			Title("Receiving links data...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("could not get links")
		}

		headers, tableData := linkding.BookmarksAsTable(bookmarks)
		tableOpts := tui.TableOpts{
			Wrap: true,
		}
		tui.PrintTable("Bookmarks", headers, tableData, tableOpts)
	},
}

func init() {
	linksCmd.AddCommand(linksQueryCmd)

	linksQueryCmd.PersistentFlags().StringP(linksQuery, "q", "", "Query. Can be empty to return all bookmarks. To search for tags, tags must be preceded by a '#'.")
}
