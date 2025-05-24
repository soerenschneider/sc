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

var vaultPkiReadCaCmd = &cobra.Command{
	Use:     "read",
	Aliases: []string{"read-ca"},
	Short:   "Reads ca data for the PKI secret engine",
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		mount := pkg.GetString(cmd, vaultMountPath)

		fullChain, _ := cmd.Flags().GetBool(vaultFullChain)

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		var caData []byte
		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				var err error
				if fullChain {
					caData, err = client.PkiGetCaChain(ctx, mount)
				} else {
					caData, err = client.PkiGetCa(ctx, mount, false)
				}
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
	vaultPkiCmd.AddCommand(vaultPkiReadCaCmd)

	vaultPkiReadCaCmd.Flags().BoolP(vaultFullChain, "f", true, "Read the full chain")
}
