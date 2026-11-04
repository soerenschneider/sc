package cmd

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var ollamaShowCmd = &cobra.Command{
	Use:     "show [model]",
	Aliases: []string{"info"},
	Short:   "Show the details of a model",
	Long: `Show the details of a model, e.g. its parameters, capabilities, template and
license. Without an argument, the model is picked interactively.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := mustOllamaClient(cmd)

		if len(args) == 0 {
			args = pickOllamaModels(cmd.Context(), client, "Model", false)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), ollamaDefaultTimeout)
		defer cancel()
		info, err := client.ShowModel(ctx, args[0])
		if err != nil {
			log.Fatal().Err(err).Msgf("could not show %s", args[0])
		}
		fmt.Print(info.Describe(args[0]))
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaShowCmd)
}
