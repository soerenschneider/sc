package cmd

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/ollama"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/spf13/cobra"
)

var ollamaPsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List the models currently loaded into memory",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustOllamaClient(cmd)

		ctx, cancel := context.WithTimeout(cmd.Context(), ollamaDefaultTimeout)
		defer cancel()
		models, err := client.ListRunning(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not list running models")
		}

		headers, rows := ollama.RunningModelsAsTable(models)
		tui.Table{
			Title:        "Running models on " + client.Address(),
			Headers:      headers,
			Rows:         rows,
			Aligns:       []tui.Align{0, tui.AlignRight, 0, tui.AlignRight},
			Caption:      fmt.Sprintf("%d models", len(rows)),
			EmptyMessage: "no models loaded",
		}.Print()
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaPsCmd)
}
