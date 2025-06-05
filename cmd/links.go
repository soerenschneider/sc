package cmd

import (
	"github.com/spf13/cobra"
)

const (
	linksAddr  = "address"
	linksQuery = "query"
	linksTags  = "tags"
	linksUrl   = "url"
	linksToken = "token"
)

var linksCmd = &cobra.Command{
	Use:   "links",
	Short: "Interact with Linkding for querying and managing links",
	Long: `The 'links' command is the entry point for interacting with Linkding.

Use subcommands to query, tail, or manage links stored in Linkding.

Examples:
  sc links query -q='#example-tag'   # Query links with specific filters`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(linksCmd)

	linksCmd.PersistentFlags().StringP(linksAddr, "a", "", "Linkding address.")
	_ = linksCmd.MarkPersistentFlagRequired(linksAddr)

	linksCmd.PersistentFlags().StringP(linksToken, "t", "", "Linkding token.")
	_ = linksCmd.MarkPersistentFlagRequired(linksToken)
}
