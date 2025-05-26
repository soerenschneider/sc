package cmd

import (
	"github.com/spf13/cobra"
)

const (
	logsAddr  = "address"
	logsQuery = "query"
	logsLimit = "limit"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Interact with VictoriaLogs for querying and managing logs",
	Long: `The 'logs' command is the entry point for interacting with VictoriaLogs.

Use subcommands to query, tail, or manage logs stored in VictoriaLogs, a high-performance log database.

Examples:
  sc logs query -q='level="error"'   # Query logs with specific filters`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.PersistentFlags().StringP(logsAddr, "a", "", "Victorialogs address.")

}
