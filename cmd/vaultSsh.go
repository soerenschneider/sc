package cmd

import (
	"github.com/spf13/cobra"
)

const (
	sshCmdFlagsSshMount     = "mount"
	sshCmdFlagsVaultAddress = "vault-address"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Sign SSH certificates or retrieve SSH CA data",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	vaultCmd.AddCommand(sshCmd)
	sshCmd.PersistentFlags().StringP(VaultMountPath, "m", "ssh", "Path where the SSH secret engine is mounted")
	sshCmd.PersistentFlags().StringP(sshCmdFlagsVaultAddress, "a", "", "Vault address")
}
