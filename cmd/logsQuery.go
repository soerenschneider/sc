package cmd

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/healthcheck"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/userdata"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const defaultLogsQuery = "error AND _time:45m"

var logsQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query logs from VictoriaLogs using a logQL-like expression",
	Long: `Query logs stored in VictoriaLogs by providing filter expressions and optional time ranges.

This command allows you to search logs with filters, labels, and time constraints. It supports various output formats
for integration with scripts or for human-readable views.`,
	Run: func(cmd *cobra.Command, args []string) {
		queryArgs := healthcheck.VictorialogsQuery{
			Address: pkg.GetString(cmd, logsAddr),
			Query:   pkg.GetString(cmd, logsQuery),
			Limit:   pkg.GetInt(cmd, logsLimit),
		}

		commandName := getCommandName(cmd)
		userData, err := userdata.LoadCommandData[logsQueryUserdata](cmp.Or(profile, defaultProfileName), commandName)
		if err != nil {
			log.Warn().Err(err).Msg("could not load userdata")
		}

		var fields []huh.Field
		if queryArgs.Query == defaultLogsQuery {
			if len(userData.LastQueries) != 0 {
				queryArgs.Query = userData.LastQueries[0]
			}
			suggestions := userData.getSuggestions()
			fields = append(fields, huh.NewInput().Title("Query").Suggestions(suggestions).Value(&queryArgs.Query).Validate(huh.ValidateNotEmpty()))
		} else {
			queryArgs.Query = defaultLogsQuery
			userData.addQuery(queryArgs.Query)
		}

		if len(fields) > 0 {
			if err := huh.NewForm(huh.NewGroup(fields...)).Run(); err != nil {
				log.Fatal().Err(err).Msg("could not get info from user")
			}
			userData.addQuery(queryArgs.Query)
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
		tableOpts := tui.TableOpts{
			Wrap:      true,
			FullWidth: true,
		}
		tui.PrintTable("Logs", tableHeaders, tableData, tableOpts)

		if err := userdata.Upsert[logsQueryUserdata](cmp.Or(profile, defaultProfileName), commandName, userData); err != nil {
			log.Warn().Err(err).Msg("could not save userdata")
		}
	},
}

func init() {
	logsCmd.AddCommand(logsQueryCmd)

	logsQueryCmd.Flags().StringP(logsQuery, "q", defaultLogsQuery, "Query")
	logsQueryCmd.Flags().IntP(logsLimit, "l", 25, "Limit")
}

type logsQueryUserdata struct {
	LastQueries []string `json:"last_queries"`
}

func (u *logsQueryUserdata) addQuery(query string) {
	if len(u.LastQueries) == 0 {
		u.LastQueries = []string{query}
		return
	}

	// Remove duplicates
	for i, q := range u.LastQueries {
		if q == query {
			u.LastQueries = append(u.LastQueries[:i], u.LastQueries[i+1:]...)
			break
		}
	}

	// extend capacity if needed
	if len(u.LastQueries) < 5 {
		u.LastQueries = append(u.LastQueries, "")
	}

	for i := len(u.LastQueries) - 1; i >= 1; i-- {
		u.LastQueries[i] = u.LastQueries[i-1]
	}
	u.LastQueries[0] = query
}

func (u *logsQueryUserdata) getSuggestions() []string {
	if slices.Contains(u.LastQueries, defaultLogsQuery) {
		return u.LastQueries
	}

	suggestions := make([]string, len(u.LastQueries)+1)
	suggestions = append(suggestions, u.LastQueries...)
	return append(suggestions, defaultLogsQuery)
}
