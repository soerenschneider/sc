package cmd

import (
	"context"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/healthcheck"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

// vaultCmd represents the vault command
var logsQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		queryArgs := healthcheck.VictorialogsQuery{
			Address: pkg.GetString(cmd, logsAddr),
			Query:   pkg.GetString(cmd, logsQuery),
			Limit:   pkg.GetInt(cmd, logsLimit),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var logs []healthcheck.LogEntry

		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				var err error
				logs, err = healthcheck.QueryVictorialogs(ctx, queryArgs)
				return err
			}).
			Title("Receiving logs data...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("could not get logs")
		}

		tableHeaders, tableData := healthcheck.TransformLogs(logs)
		tui.PrintTable("Logs", tableHeaders, tableData)
	},
}

func init() {
	logsCmd.AddCommand(logsQueryCmd)

	logsQueryCmd.Flags().StringP(logsQuery, "q", "error AND _time:45m", "Query")
	logsQueryCmd.Flags().IntP(logsLimit, "l", 25, "Limit")
}
