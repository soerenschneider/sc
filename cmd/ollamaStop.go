package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var ollamaStopCmd = &cobra.Command{
	Use:   "stop [model]...",
	Short: "Unload models from memory",
	Long: `Unload models from memory. Without arguments, all loaded models are
unloaded.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustOllamaClient(cmd)

		if len(args) == 0 {
			ctx, cancel := context.WithTimeout(cmd.Context(), ollamaDefaultTimeout)
			running, err := client.ListRunning(ctx)
			cancel()
			if err != nil {
				log.Fatal().Err(err).Msg("could not list running models")
			}
			for _, m := range running {
				args = append(args, m.Name)
			}
			if len(args) == 0 {
				log.Info().Msg("No models loaded")
				return
			}
		}

		for _, name := range args {
			ctx, cancel := context.WithTimeout(cmd.Context(), ollamaDefaultTimeout)
			err := client.UnloadModel(ctx, name)
			cancel()
			if err != nil {
				log.Fatal().Err(err).Msgf("could not unload %s", name)
			}
			log.Info().Msgf("Unloaded %s", name)
		}
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaStopCmd)
}
