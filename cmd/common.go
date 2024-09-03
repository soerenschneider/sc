package cmd

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const vaultAddrEnvVarKey = "VAULT_ADDR"

func getVaultAddress(cmd *cobra.Command) (string, error) {
	address, err := cmd.Flags().GetString(sshCmdFlagsVaultAddress)
	if err == nil && len(address) > 0 {
		return address, nil
	}
	address = os.Getenv(vaultAddrEnvVarKey)
	if len(address) == 0 {
		return "", errors.New("no vault address specified")
	}
	log.Info().Msgf("No vault address supplied explicitly, using value of env var %s=%s", vaultAddrEnvVarKey, address)
	return address, nil
}

func GetExpandedFile(filename string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if strings.HasPrefix(filename, "~/") {
		return filepath.Join(dir, filename[2:])
	}

	if strings.HasPrefix(filename, "$HOME/") {
		return filepath.Join(dir, filename[6:])
	}

	return filename
}

func DieOnErr(err error, msg string) {
	if err == nil {
		return
	}

	log.Fatal().Err(err).Msg(msg)
}
