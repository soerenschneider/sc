package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal"
	"github.com/soerenschneider/sc/internal/healthcheck"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	healthcheckWatch    = "watch"
	healthcheckInterval = "interval"
	healthcheckDomain   = "domain"

	// defaultWidth is used when stdout is not a terminal.
	defaultWidth = 120
)

var healthcheckCmd = &cobra.Command{
	Use: "healthcheck",
	Aliases: []string{
		"health",
	},
	Short: "Display a dashboard of the health of soeren.cloud",
	Long: `Display a dashboard of the health of soeren.cloud.

The hosts to check are read from the "healthcheck" section of the profile. Apart from the
internet connectivity checks, which default to public endpoints, every section is optional.

  prod:
    healthcheck:
      domain: soeren.cloud
      interval: 30s
      internet:
        dns-record: google.com
        http-url: https://www.google.com/generate_204
        dns-servers: [1.1.1.1:53, 8.8.8.8:53]
      tcp:
        - name: SSH
          endpoints: [router.dd.soeren.cloud:22, rs.soeren.cloud:22]
      dns:
        - name: Router DNS
          servers: [192.168.65.1, 192.168.200.1]
          record: google.com
      latency:
        title: VPN latency 30m
        endpoint: https://victoriametrics.example.com/api/v1/query
        avg-query: avg_over_time(probe_duration_seconds{job="blackbox-icmp-vpn"}[30m]) * 1000
        max-query: max_over_time(probe_duration_seconds{job="blackbox-icmp-vpn"}[30m]) * 1000
      logs:
        address: https://logs.example.com
        query: error AND _time:15m
        limit: 5

Exits with status 1 if any probe failed, unless --watch is given.`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := mustHealthcheckOptions(cmd)
		domain := pkg.GetString(cmd, healthcheckDomain)

		if pkg.MustGetBool(cmd, healthcheckWatch) {
			interval := pkg.GetDuration(cmd, healthcheckInterval)
			if err := healthcheck.Watch(cmd.Context(), opts, domain, interval); err != nil {
				log.Fatal().Err(err).Msg("running dashboard failed")
			}
			return
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		report, err := tui.LoadWithSpinner(ctx, "Collecting health data…", func(ctx context.Context) (*healthcheck.Report, error) {
			return healthcheck.Collect(ctx, opts), nil
		})
		if err != nil {
			log.Fatal().Err(err).Msg("collecting health data failed")
		}

		width := defaultWidth
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil { //#nosec:G115
			width = w
		}
		fmt.Println(healthcheck.Render(report, healthcheck.RenderOptions{Width: width, Domain: domain}))

		if up, total := report.Counts(); up < total {
			os.Exit(1)
		}
	},
}

// mustHealthcheckOptions reads the checks from the profile, falling back to public endpoints for the internet checks.
func mustHealthcheckOptions(cmd *cobra.Command) healthcheck.Options {
	opts := healthcheck.Options{
		Internet: healthcheck.InternetConnectivityOpts{
			DnsRecord:    "google.com",
			HttpCheckUrl: "https://www.google.com/generate_204",
			DnsServers: []string{
				"1.1.1.1:53",
				"8.8.8.8:53",
				"8.8.4.4:53",
			},
		},
	}

	found, err := internal.Decode(getCommandName(cmd), profile, &opts)
	if err != nil {
		log.Fatal().Err(err).Msg("could not read healthcheck settings")
	}
	if !found {
		log.Warn().Msgf("No healthcheck settings found in profile %q, only checking internet connectivity. See 'sc healthcheck --help'.", profile)
	}
	return opts
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
	healthcheckCmd.Flags().BoolP(healthcheckWatch, "w", false, "Keep refreshing the dashboard")
	healthcheckCmd.Flags().Duration(healthcheckInterval, 30*time.Second, "Refresh interval in watch mode")
	healthcheckCmd.Flags().String(healthcheckDomain, "", "Domain to strip from host names")
}
