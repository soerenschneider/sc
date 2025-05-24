package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"

	"github.com/spf13/cobra"
)

var vaultSshReadCaCmd = &cobra.Command{
	Use:     "read",
	Aliases: []string{"read-ca"},
	Short:   "Reads ca data for the SSH secret engine",
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		mount := pkg.GetString(cmd, vaultMountPath)

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		var caData []byte
		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				var err error
				caData, err = client.SshGetCa(ctx, mount)
				return err
			}).
			Title("Fetching ca data from Vault").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Warn().Err(err).Msg("could not fetch ca data")
		}

		fmt.Println(string(caData))
	},
}

func init() {
	vaultSshCmd.AddCommand(vaultSshReadCaCmd)
}
