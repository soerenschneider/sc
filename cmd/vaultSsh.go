package cmd

import (
	"github.com/spf13/cobra"
)

var vaultSshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Sign SSH certificates or retrieve SSH CA data",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	vaultCmd.AddCommand(vaultSshCmd)
	vaultSshCmd.PersistentFlags().StringP(vaultMountPath, "m", "ssh", "Path where the SSH secret engine is mounted")
}
