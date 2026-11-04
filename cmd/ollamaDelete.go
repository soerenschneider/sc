package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var ollamaDeleteCmd = &cobra.Command{
	Use:     "delete [model]...",
	Aliases: []string{"del", "rm"},
	Short:   "Delete models from the Ollama instance",
	Long: `Delete models from the Ollama instance. Without arguments, the models are
picked interactively.

The deletion has to be confirmed, unless --yes is given.

Examples:
  sc ollama delete
  sc ollama delete llama3.1:8b --yes`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustOllamaClient(cmd)

		if len(args) == 0 {
			args = pickOllamaModels(cmd.Context(), client, "Models to delete", true)
			if len(args) == 0 {
				return
			}
		}

		if !pkg.MustGetBool(cmd, ollamaYes) {
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				log.Fatal().Msgf("Not deleting without confirmation, use --%s", ollamaYes)
			}
			confirmed := false
			if err := huh.NewConfirm().
				Title(fmt.Sprintf("Delete %d model(s) from %s?", len(args), client.Address())).
				Description(strings.Join(args, "\n")).
				Value(&confirmed).
				Run(); err != nil || !confirmed {
				log.Info().Msg("Not deleting any model")
				return
			}
		}

		var failed bool
		for _, name := range args {
			ctx, cancel := context.WithTimeout(cmd.Context(), ollamaDefaultTimeout)
			err := client.DeleteModel(ctx, name)
			cancel()
			if err != nil {
				log.Error().Err(err).Msgf("could not delete %s", name)
				failed = true
				continue
			}
			log.Info().Msgf("Deleted %s", name)
		}
		if failed {
			os.Exit(1)
		}
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaDeleteCmd)

	ollamaDeleteCmd.Flags().BoolP(ollamaYes, "y", false, "Delete without asking for confirmation")
}
