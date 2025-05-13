package cmd

import (
	"os/user"

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
	rootCmd.AddCommand(sshCmd)
	sshCmd.PersistentFlags().StringP(VaultMountPath, "m", "ssh", "Path where the SSH secret engine is mounted")
	sshCmd.PersistentFlags().StringP(sshCmdFlagsVaultAddress, "a", "", "Vault address")
}

func getPrincipals(cmd *cobra.Command) ([]string, error) {
	principals, err := cmd.Flags().GetStringArray(sshSignKeyCmdFlagsPrincipals)
	if err == nil && len(principals) > 0 {
		return principals, nil
	}

	return getDefaultPrincipal()
}

func getDefaultPrincipal() ([]string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	return []string{user.Name}, nil
}
