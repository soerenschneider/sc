package cmd

import (
	"cmp"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultTokenLookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hi!")
		vaultAddr := cmp.Or(GetString(cmd, vaultAddr), getVaultAddr())

		config := vault.DefaultConfig()
		config.Address = vaultAddr

		client, err := vault.NewClient(config)
		if err != nil {
			log.Fatal().Msgf("Error creating Vault client: %v", err)
		}

		token, err := getVaultToken()
		if err != nil {
			log.Fatal().Err(err).Msg("unable to get Vault token")
		}
		client.SetToken(token)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		secret, err := client.Auth().Token().LookupSelfWithContext(ctx)
		if err != nil {
			log.Fatal().Msgf("could not lookup: %v", err)
		}

		/*
			var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("WROTE")
			fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), fmt.Sprintf("%v", secret.Data)))
		*/

		writeOutput(secret)
	},
}

func writeOutput(secret *vault.Secret) {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")) // Blue

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // Gray

	// Get and sort keys
	keys := make([]string, 0, len(secret.Data))
	maxKeyLen := 0
	for k := range secret.Data {
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
		v := secret.Data[k]
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
