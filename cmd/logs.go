package cmd

import (
	"github.com/spf13/cobra"
)

const (
	logsAddr  = "address"
	logsQuery = "query"
	logsLimit = "limit"
)

// vaultCmd represents the vault command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.PersistentFlags().StringP(logsAddr, "a", "", "Victorialogs address.")

}
