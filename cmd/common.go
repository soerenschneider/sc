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
