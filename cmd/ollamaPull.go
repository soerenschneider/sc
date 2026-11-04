package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/ollama"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const ollamaPullAll = "all"

var ollamaPullCmd = &cobra.Command{
	Use:     "pull model...",
	Aliases: []string{"download", "get"},
	Short:   "Download models or update them to their latest version",
	Long: `Download models from the Ollama library to the instance. Pulling a model
that already exists updates it to the latest version, only changed layers are
downloaded. With --all, every model on the instance is updated.

Examples:
  sc ollama pull llama3.2
  sc ollama pull qwen3:8b gemma3:12b
  sc ollama pull --all`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustOllamaClient(cmd)

		if pkg.MustGetBool(cmd, ollamaPullAll) {
			for _, m := range mustListOllamaModels(cmd.Context(), client) {
				args = append(args, m.Name)
			}
		}
		if len(args) == 0 {
			log.Fatal().Msgf("No model given, pass model names or --%s", ollamaPullAll)
		}

		var failed []string
		for _, name := range args {
			p := newOllamaPullPrinter(name)
			err := client.PullModel(cmd.Context(), name, p.update)
			p.done()
			if err != nil {
				log.Error().Err(err).Msgf("could not pull %s", name)
				failed = append(failed, name)
				continue
			}
			log.Info().Msgf("Pulled %s", name)
		}

		if len(failed) > 0 {
			log.Fatal().Msgf("Could not pull %s", strings.Join(failed, ", "))
		}
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaPullCmd)

	ollamaPullCmd.Flags().Bool(ollamaPullAll, false, "Update all models on the instance")
}

// ollamaPullPrinter shows the progress of a pull on a single, updated line if
// stderr is a terminal and logs every new status otherwise.
type ollamaPullPrinter struct {
	name       string
	tty        bool
	lastStatus string
	printed    bool
}

func newOllamaPullPrinter(name string) *ollamaPullPrinter {
	return &ollamaPullPrinter{name: name, tty: term.IsTerminal(int(os.Stderr.Fd()))}
}

func (p *ollamaPullPrinter) update(progress ollama.PullProgress) {
	if !p.tty {
		if progress.Status != p.lastStatus {
			log.Debug().Msgf("%s: %s", p.name, progress.Status)
		}
		p.lastStatus = progress.Status
		return
	}

	line := fmt.Sprintf("%s: %s", p.name, progress.Status)
	if progress.Total > 0 {
		line += fmt.Sprintf(" %3d%% (%s/%s)",
			progress.Completed*100/progress.Total,
			humanize.Bytes(uint64(max(progress.Completed, 0))),
			humanize.Bytes(uint64(progress.Total)))
	}
	if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 1 && len(line) >= w {
		line = line[:w-1]
	}
	_, _ = fmt.Fprintf(os.Stderr, "\r\033[K%s", line)
	p.printed = true
}

func (p *ollamaPullPrinter) done() {
	if p.printed {
		_, _ = fmt.Fprint(os.Stderr, "\r\033[K")
	}
}
