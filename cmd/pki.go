package cmd

import (
	"github.com/spf13/cobra"
)

const (
	pkiCmdFlagsPkiMount        = "mount"
	defaultPkiCmdFlagsPkiMount = "pki"

	pkiCmdFlagsVaultAddress = "vault-address"
)

var pkiCmd = &cobra.Command{
	Use:   "pki",
	Short: "Sign, issue and revoke x509 certificates and retrieve x509 CA data",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(pkiCmd)
	pkiCmd.PersistentFlags().StringP(pkiCmdFlagsPkiMount, "m", defaultPkiCmdFlagsPkiMount, "Path where the PKI secret engine is mounted")
	pkiCmd.PersistentFlags().StringP(pkiCmdFlagsVaultAddress, "a", "", "Vault address")
}
