package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var ollamaBrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse the models of the Ollama instance in a TUI",
	Long: `Browse the models of the Ollama instance in a filterable list. Enter shows
the details of the selected model, esc goes back to the list.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			log.Fatal().Msg("browse needs a terminal, use 'sc ollama list' instead")
		}

		client := mustOllamaClient(cmd)
		models := mustListOllamaModels(cmd.Context(), client)
		if len(models) == 0 {
			log.Info().Msgf("No models available on %s", client.Address())
			return
		}

		names := make([]string, len(models))
		for i, m := range models {
			names[i] = m.Name
		}

		load := func(ctx context.Context, name string) (string, error) {
			ctx, cancel := context.WithTimeout(ctx, ollamaDefaultTimeout)
			defer cancel()
			info, err := client.ShowModel(ctx, name)
			if err != nil {
				return "", err
			}
			return info.Describe(name), nil
		}

		result, err := tui.Browse(cmd.Context(), "Models on "+client.Address(), names, load)
		if err != nil {
			log.Fatal().Err(err).Msg("could not browse models")
		}
		if result.Name != "" {
			fmt.Print(result.Content)
		}
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaBrowseCmd)
}
