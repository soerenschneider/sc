package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/spf13/cobra"
)

// vaultLoginCmd represents the vaultLogin command
var vaultTokenLookupCmd = &cobra.Command{
	Use:     "lookup",
	Aliases: []string{"ls"},
	Short:   "Lookup and display information about the current Vault token",
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

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		secret, err := client.LookupToken(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("could not lookup")
		}

		// don't leak token
		delete(secret.Data, "id")

		// re-format times
		var problematicColumns []string
		expiry, ok := secret.Data["expire_time"]
		if ok {
			updatedValue, parsedTime := reformatTimestamp(expiry.(string))
			if parsedTime != nil && time.Now().After(*parsedTime) {
				problematicColumns = append(problematicColumns, "expire_time")
			}
			secret.Data["expire_time"] = updatedValue
		}

		issue, ok := secret.Data["issue_time"]
		if ok {
			updatedValue, _ := reformatTimestamp(issue.(string))
			secret.Data["issue_time"] = updatedValue
		}

		tui.PrintMapOutput(secret.Data, problematicColumns...)
	},
}

func reformatTimestamp(timestamp string) (string, *time.Time) {
	parsedTime, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return timestamp, nil
	}

	format := "15:04:05 MST"
	if !isToday(parsedTime) {
		format = "06-01-02T15:04 MST"
	}

	var relative string
	if parsedTime.Before(time.Now()) {
		relative = fmt.Sprintf("%s ago", time.Since(parsedTime).Round(time.Minute))
	} else {
		relative = fmt.Sprintf("in %s", time.Until(parsedTime).Round(time.Minute))
	}

	return fmt.Sprintf("%s (%s)", relative, parsedTime.Format(format)), &parsedTime
}

func isToday(t time.Time) bool {
	now := time.Now().UTC()

	return t.Year() == now.Year() &&
		t.Month() == now.Month() &&
		t.Day() == now.Day()
}

func init() {
	vaultTokenCmd.AddCommand(vaultTokenLookupCmd)
}
