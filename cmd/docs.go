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
		err := os.RemoveAll(docsDir)
		if err != nil {
			log.Fatalf("Failed to remove docs directory: %v", err)
		}

		err = os.Mkdir(docsDir, 0755)
		if err != nil {
			log.Fatalf("Failed to recreate docs directory: %v", err)
		}

		agentDocsDir := filepath.Join(docsDir, "agent")
		if err = os.Mkdir(agentDocsDir, 0755); err != nil {
			log.Fatalf("Failed to create dir: %v", err)
		}

		if err = doc.GenMarkdownTree(agentCmd, agentDocsDir); err != nil {
			log.Fatalf("Could not generate docs: %v", err)
		}

		sshDocsDir := filepath.Join(docsDir, "ssh")
		if err = os.Mkdir(sshDocsDir, 0755); err != nil {
			log.Fatalf("Failed to create dir: %v", err)
		}

		if err = doc.GenMarkdownTree(vaultSshCmd, sshDocsDir); err != nil {
			log.Fatalf("Could not generate docs: %v", err)
		}

		pkiDocsDir := filepath.Join(docsDir, "pki")
		if err = os.Mkdir(pkiDocsDir, 0755); err != nil {
			log.Fatalf("Failed to create dir: %v", err)
		}

		if err = doc.GenMarkdownTree(pkiCmd, pkiDocsDir); err != nil {
			log.Fatalf("Could not generate docs: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
