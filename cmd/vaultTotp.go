package cmd

import (
	"github.com/spf13/cobra"
)

const ()

var vaultTotpCmd = &cobra.Command{
	Use:   "totp",
	Short: "Manage Vault TOTP MFA methods and secrets",
	Long: `The 'totp' command group contains subcommands for managing TOTP MFA methods and
the TOTP secrets of identity entities.

This command itself does not perform any actions. Instead, use one of its subcommands
to inspect or manage TOTP methods and secrets.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	vaultCmd.AddCommand(vaultTotpCmd)
}
