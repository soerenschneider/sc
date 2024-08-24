package cmd

import (
	"errors"
	"os"
	"os/user"

	"github.com/rs/zerolog/log"
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
	sshCmd.PersistentFlags().StringP(sshCmdFlagsSshMount, "m", "ssh", "Path where the SSH secret engine is mounted")
	sshCmd.PersistentFlags().StringP(sshCmdFlagsVaultAddress, "a", "", "Vault address")
}

func getVaultAddress(cmd *cobra.Command) (string, error) {
	address, err := cmd.Flags().GetString(sshCmdFlagsVaultAddress)
	if err == nil && len(address) > 0 {
		return address, nil
	}
	log.Info().Msg("No vault address supplied, trying env var VAULT_ADDR")
	address = os.Getenv("VAULT_ADDR")
	if len(address) == 0 {
		return "", errors.New("no vault address specified")
	}

	return address, nil
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
