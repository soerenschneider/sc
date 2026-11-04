package cmd

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/soerenschneider/sc/internal/ollama"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/spf13/cobra"
)

var ollamaListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List the models available on the Ollama instance",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustOllamaClient(cmd)
		models := mustListOllamaModels(cmd.Context(), client)

		var total int64
		for _, m := range models {
			total += m.Size
		}

		headers, rows := ollama.ModelsAsTable(models)
		tui.Table{
			Title:        "Models on " + client.Address(),
			Headers:      headers,
			Rows:         rows,
			Aligns:       []tui.Align{0, tui.AlignRight, tui.AlignRight},
			Caption:      fmt.Sprintf("%d models, %s", len(rows), humanize.Bytes(uint64(max(total, 0)))),
			Zebra:        len(rows) > 10,
			EmptyMessage: "no models",
		}.Print()
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaListCmd)
}
