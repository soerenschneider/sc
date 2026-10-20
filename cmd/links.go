package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/huh/v2"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/linkding"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	linksAddr      = "address"
	linksQuery     = "query"
	linksTags      = "tags"
	linksUrl       = "url"
	linksToken     = "token"
	linksTokenFile = "token-file"

	linksDefaultTimeout = 10 * time.Second
)

var linksCmd = &cobra.Command{
	Use:   "links",
	Short: "Manage bookmarks in Linkding",
	Long: `The 'links' command is the entry point for interacting with Linkding.

Use subcommands to add, search, open, edit and delete bookmarks and to list
the tags stored in Linkding.

Commands working on a single bookmark take its id, its URL or a search query.
If the query matches several bookmarks, one is picked interactively.

Examples:
  sc links query -q='#example-tag'   # Query links with specific filters
  sc links open golang               # Open the bookmark matching "golang"
  sc links tag 42 --add-tag go       # Add a tag to bookmark 42
  sc links delete https://example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(linksCmd)

	linksCmd.PersistentFlags().StringP(linksAddr, "a", "", "Linkding address.")
	_ = linksCmd.MarkPersistentFlagRequired(linksAddr)

	linksCmd.PersistentFlags().StringP(linksToken, "t", "", "Linkding token.")
	linksCmd.PersistentFlags().StringP(linksTokenFile, "T", "", "Path to a file containing the Linkding token.")

	// Exactly one source is allowed, and at least one is required.
	linksCmd.MarkFlagsMutuallyExclusive(linksToken, linksTokenFile)
	linksCmd.MarkFlagsOneRequired(linksToken, linksTokenFile)
}

func getLinkdingToken(cmd *cobra.Command) (string, error) {
	token, _ := cmd.Flags().GetString(linksToken)
	if tokenFile, _ := cmd.Flags().GetString(linksTokenFile); tokenFile != "" {
		tokenFile = pkg.GetExpandedFile(tokenFile)
		data, err := os.ReadFile(tokenFile)
		if err != nil {
			return "", fmt.Errorf("reading token file: %w", err)
		}
		token = strings.TrimSpace(string(data))
	}

	return token, nil
}

func mustLinkdingClient(cmd *cobra.Command) *linkding.LinkdingClient {
	token, err := getLinkdingToken(cmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting linkding token")
	}
	client, err := linkding.NewLinkdingClient(pkg.GetString(cmd, linksAddr), token)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build client")
	}
	return client
}

// resolveBookmark returns the bookmark identified by args, which is either the
// id or URL of a bookmark or a search query. If the query matches several
// bookmarks, the user picks one.
func resolveBookmark(ctx context.Context, client *linkding.LinkdingClient, args []string) (*linkding.Bookmark, error) {
	if len(args) == 1 {
		if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			ctx, cancel := context.WithTimeout(ctx, linksDefaultTimeout)
			defer cancel()
			return client.GetBookmark(ctx, id)
		}
		if u, err := url.Parse(args[0]); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
			ctx, cancel := context.WithTimeout(ctx, linksDefaultTimeout)
			defer cancel()
			bookmark, err := client.CheckBookmark(ctx, args[0])
			if err == nil && bookmark == nil {
				err = fmt.Errorf("%s is not bookmarked", args[0])
			}
			return bookmark, err
		}
	}

	query := strings.Join(args, " ")
	listCtx, cancel := context.WithTimeout(ctx, linksDefaultTimeout)
	defer cancel()
	bookmarks, err := client.ListBookmarks(listCtx, query, 100)
	if err != nil {
		return nil, err
	}

	switch {
	case len(bookmarks) == 0:
		return nil, fmt.Errorf("no bookmarks match %q", query)
	case len(bookmarks) == 1:
		return &bookmarks[0], nil
	case !term.IsTerminal(int(os.Stdin.Fd())):
		return nil, fmt.Errorf("%d bookmarks match %q, use a more specific query or the id", len(bookmarks), query)
	}

	options := make([]huh.Option[int], len(bookmarks))
	for i, b := range bookmarks {
		options[i] = huh.NewOption(fmt.Sprintf("%s · %s", b.DisplayName(), b.URL), i)
	}
	var selected int
	if err := huh.NewSelect[int]().
		Title(fmt.Sprintf("%d bookmarks match", len(bookmarks))).
		Options(options...).
		Filtering(true).
		Value(&selected).
		Run(); err != nil {
		return nil, err
	}
	return &bookmarks[selected], nil
}

// mustResolveBookmark is like resolveBookmark but exits on errors. It exits
// quietly if the user aborts the picker.
func mustResolveBookmark(cmd *cobra.Command, client *linkding.LinkdingClient, args []string) *linkding.Bookmark {
	bookmark, err := resolveBookmark(cmd.Context(), client, args)
	if errors.Is(err, huh.ErrUserAborted) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("could not find bookmark")
	}
	return bookmark
}

// parseTags splits tags separated by spaces or commas, drops a leading '#' and
// removes duplicates.
func parseTags(values ...string) []string {
	var tags []string
	for _, v := range values {
		for _, tag := range strings.FieldsFunc(v, func(r rune) bool { return r == ' ' || r == ',' }) {
			tag = strings.TrimPrefix(tag, "#")
			if tag != "" && !slices.Contains(tags, tag) {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}
