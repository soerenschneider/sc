package cmd

import (
	"context"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultIdentityGetEntityCmd = &cobra.Command{
	Use: "get-entity",
	Aliases: []string{
		"get",
	},
	Short: "Manage Vault totp",
	Long: `The 'token' command group contains subcommands for interacting with Vault tokens.

This command itself does not perform any actions. Instead, use one of its subcommands
to inspect or manage tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		name := pkg.GetString(cmd, "name")
		if name == "" {
			var suggestions []string

			ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
			defer cancel()

			if err := spinner.New().
				Type(spinner.Line).
				ActionWithErr(func(ctx context.Context) error {
					entities, err := client.IdentityListEntities(ctx)
					if err == nil {
						suggestions = entities
					}
					return nil
				}).
				Title("Loading available entities...").
				Accessible(false).
				Context(ctx).
				Type(spinner.Dots).
				Run(); err != nil {
				log.Fatal().Err(err).Msg("sending login to request to Vault failed")
			}

			name = tui.ReadInput("Enter entity name", suggestions)
		}

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		data, err := client.IdentityGetEntityByName(ctx, name)
		if err != nil {
			log.Fatal().Err(err).Msg("could not list identity entities")
		}

		writeOutput(data)
	},
}

func init() {
	vaultIdentityCmd.AddCommand(vaultIdentityGetEntityCmd)

	vaultIdentityGetEntityCmd.Flags().StringP(vaulEntityName, "n", "", "Name of the entity")
}
