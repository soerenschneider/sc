package cmd

import (
	"github.com/spf13/cobra"
)

const (
	vaultAddr      = "address"
	vaultTokenFile = "token-file"
)

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
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
	rootCmd.AddCommand(vaultCmd)

	vaultCmd.Flags().StringP(vaultAddr, "a", "", "Vault address. If not specified, tries to read env variable VAULT_ADDR.")
	vaultCmd.Flags().StringP(vaultTokenFile, "t", "~/.vault-token", "Vault token file.")
}
