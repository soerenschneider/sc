//go:build docgen

package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docsDir = "./docs/cli"

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate CLI docs",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.RemoveAll(docsDir); err != nil {
			log.Fatalf("Failed to remove docs directory: %v", err)
		}

		if err := os.Mkdir(docsDir, 0755); err != nil {
			log.Fatalf("Failed to recreate docs directory: %v", err)
		}

		// every top-level command gets its own directory
		for _, c := range rootCmd.Commands() {
			if !c.IsAvailableCommand() || c.Name() == "completion" {
				continue
			}

			dir := filepath.Join(docsDir, c.Name())
			if err := os.Mkdir(dir, 0755); err != nil {
				log.Fatalf("Failed to create dir: %v", err)
			}

			if err := doc.GenMarkdownTree(c, dir); err != nil {
				log.Fatalf("Could not generate docs for %s: %v", c.Name(), err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
