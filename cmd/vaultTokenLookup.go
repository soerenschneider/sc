package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultTokenLookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup and display information about the current Vault token",
	Long: `Retrieve and display information about the currently active Vault token.

This command queries the Vault API to show metadata about the current token,
such as its creation time, expiration, policies, and identity.

It attempts to authenticate using the following sources (in order):
  1. The VAULT_TOKEN environment variable
  2. A token loaded from the local configuration file (e.g. ~/.config/mycli/token)
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustBuildClient(cmd)
		vault.MustAuthenticateClient(client)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		secret, err := client.LookupToken(ctx)
		if err != nil {
			log.Fatal().Msgf("could not lookup: %v", err)
		}

		writeOutput(secret.Data)
	},
}

func writeOutput(data map[string]any) {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")) // Blue

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // Gray

	// Get and sort keys
	keys := make([]string, 0, len(data))
	maxKeyLen := 0
	for k := range data {
		keys = append(keys, k)
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}
	sort.Strings(keys)

	// Build output
	indent := 2
	var builder strings.Builder
	for _, k := range keys {
		v := data[k]
		paddedKey := fmt.Sprintf("%-*s", maxKeyLen, k)
		key := keyStyle.Render(paddedKey)
		val := valueStyle.Render(fmt.Sprintf("%v", v))
		spaces := strings.Repeat(" ", indent)
		builder.WriteString(fmt.Sprintf("%s%s%s\n", key, spaces, val))
	}

	fmt.Println(builder.String())
}

func init() {
	vaultTokenCmd.AddCommand(vaultTokenLookupCmd)
}
