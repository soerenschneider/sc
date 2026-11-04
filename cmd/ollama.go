package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"charm.land/huh/v2"
	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/ollama"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	ollamaAddress = "address"
	ollamaYes     = "yes"

	ollamaDefaultTimeout = 10 * time.Second
)

var ollamaCmd = &cobra.Command{
	Use:   "ollama",
	Short: "Manage the models of an Ollama instance",
	Long: `Download, delete, inspect and browse the models of an Ollama instance, and
ask them quick questions.

The address defaults to OLLAMA_HOST like for the ollama CLI, and then to
` + ollama.DefaultAddress + `.

Examples:
  sc ollama list
  sc ollama pull llama3.2 qwen3:8b
  sc ollama pull --all
  sc ollama delete llama3.1
  sc ollama show qwen3:8b
  sc ollama browse
  sc ollama ps
  sc ollama ask -m qwen3:8b "how do I undo the last git commit"
  git diff | sc ollama ask -c "and what about this change?"
  sc ollama -a gpu.lan list`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(ollamaCmd)

	ollamaCmd.PersistentFlags().StringP(ollamaAddress, "a", "", "Ollama address, host[:port] or URL")
}

func mustOllamaClient(cmd *cobra.Command) *ollama.Client {
	client, err := ollama.NewClient(pkg.GetString(cmd, ollamaAddress))
	if err != nil {
		log.Fatal().Err(err).Msg("could not build ollama client")
	}
	return client
}

func mustListOllamaModels(ctx context.Context, client *ollama.Client) []ollama.Model {
	ctx, cancel := context.WithTimeout(ctx, ollamaDefaultTimeout)
	defer cancel()

	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatal().Err(err).Msgf("could not list models of %s", client.Address())
	}
	sort.Slice(models, func(i, j int) bool { return models[i].Name < models[j].Name })
	return models
}

// pickOllamaModels lets the user pick models interactively. It exits if
// there is no terminal or no model to pick.
func pickOllamaModels(ctx context.Context, client *ollama.Client, title string, multi bool) []string {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		log.Fatal().Msg("No model given")
	}

	models := mustListOllamaModels(ctx, client)
	if len(models) == 0 {
		log.Info().Msgf("No models available on %s", client.Address())
		os.Exit(0)
	}

	options := make([]huh.Option[string], len(models))
	for i, m := range models {
		options[i] = huh.NewOption(fmt.Sprintf("%s · %s", m.Name, ollamaModelSummary(m)), m.Name)
	}

	var selected []string
	var err error
	if multi {
		err = huh.NewMultiSelect[string]().
			Title(title).
			Options(options...).
			Filterable(true).
			Value(&selected).
			Run()
	} else {
		var name string
		err = huh.NewSelect[string]().
			Title(title).
			Options(options...).
			Filtering(true).
			Value(&name).
			Run()
		selected = []string{name}
	}
	if err != nil {
		os.Exit(0)
	}
	return selected
}

func ollamaModelSummary(m ollama.Model) string {
	summary := humanize.Bytes(uint64(max(m.Size, 0)))
	if m.Details.ParameterSize != "" {
		summary += ", " + m.Details.ParameterSize
	}
	return summary
}
