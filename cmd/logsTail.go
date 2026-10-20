package cmd

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const defaultLogsTailQuery = "*"

var logsTailCmd = &cobra.Command{
	Use:     "tail [query]...",
	Aliases: []string{"t", "follow"},
	Short:   "Follow logs from VictoriaLogs as they arrive",
	Long: `Follow logs matching a LogsQL query live until interrupted. The query is given as
arguments, which are joined with spaces, or with --query, and defaults to all logs.

With --since, matching entries of that time window are printed first, capped by
--limit, before following new ones.

Pipes that need the complete input, such as sort or stats, are not supported
by VictoriaLogs when tailing.

Examples:
  sc logs tail                            # Follow all logs
  sc logs tail error                      # Follow logs containing "error"
  sc logs tail 'service:nginx' -s 10m     # Show the last 10 minutes, then follow`,
	Run: func(cmd *cobra.Command, args []string) {
		query := pkg.GetString(cmd, logsQuery)
		if len(args) > 0 {
			query = strings.Join(args, " ")
		}

		logsOpts := &logsOpts{
			url:             pkg.GetString(cmd, logsAddr),
			query:           query,
			format:          pkg.GetString(cmd, logsFormat),
			idFields:        pkg.GetStringArray(cmd, logsIdFields),
			since:           pkg.GetString(cmd, logsSince),
			noColor:         pkg.MustGetBool(cmd, logsNoColor),
			limit:           pkg.GetInt(cmd, logsLimit),
			refreshInterval: pkg.GetDuration(cmd, logsRefreshInterval),
			follow:          true,
			timeout:         pkg.GetDuration(cmd, logsTimeout),
		}

		if err := runLogs(logsOpts); err != nil {
			log.Fatal().Err(err).Msg("could not tail logs")
		}
	},
}

func init() {
	logsCmd.AddCommand(logsTailCmd)

	logsTailCmd.Flags().StringP(logsQuery, "q", defaultLogsTailQuery, "Query, overridden by arguments")
	addLogsOutputFlags(logsTailCmd)
	logsTailCmd.Flags().StringP(logsSince, "s", "",
		"Print entries since this time before following: duration (1h), RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or Unix epoch")
	logsTailCmd.Flags().IntP(logsLimit, "l", 100, "Maximum entries printed for --since")
	logsTailCmd.Flags().DurationP(logsRefreshInterval, "r", 2*time.Second, "How often the server checks for new entries")
	logsTailCmd.Flags().DurationP(logsTimeout, "", 10*time.Second, "Timeout of the --since query")
}
