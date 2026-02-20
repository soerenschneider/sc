package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/healthcheck"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/spf13/cobra"
)

type Data struct {
	prometheus     [][]string
	TcpChecks      map[string]map[string]healthcheck.ProbeResult
	internetAccess map[string]healthcheck.ProbeResult
	logs           []healthcheck.LogEntry
	DnsChecks      map[string]map[string]healthcheck.ProbeResult
}

type DnsCheck struct {
	DnsServers []string
	DnsRecord  string
}

type HealthcheckOptions struct {
	TcpChecks            map[string][]string
	DnsChecks            map[string]DnsCheck
	InternetConnectivity healthcheck.InternetConnectivityOpts
}

var healthcheckCmd = &cobra.Command{
	Use: "healthcheck",
	Aliases: []string{
		"health",
	},
	Short: "Performs a healthcheck for internet connectivity",
	Run: func(cmd *cobra.Command, args []string) {

		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		var data *Data

		opts := HealthcheckOptions{
			TcpChecks: map[string][]string{
				"local": []string{
					"router.dd.soeren.cloud:22",
					"router.ez.soeren.cloud:22",
					"router.pt.soeren.cloud:22",
					"rs.soeren.cloud:22",
					"swiss.soeren.cloud:22",
				},
				"kubernetes": []string{
					"rs.soeren.cloud:6443",
					"k8s.dd.soeren.cloud:6443",
					"k8s.ez.soeren.cloud:6443",
					"k8s.pt.soeren.cloud:6443",
				},
			},
			DnsChecks: map[string]DnsCheck{
				"routers": {
					DnsServers: []string{
						"192.168.65.1",
						"192.168.2.3",
						"192.168.73.1",
						"192.168.200.1",
					},
					DnsRecord: "google.com",
				},
			},
			InternetConnectivity: healthcheck.InternetConnectivityOpts{
				DnsRecord:    "google.com",
				HttpCheckUrl: "https://www.google.com/generate_204",
				DnsServers: []string{
					"1.1.1.1:53",
					"8.8.8.8:53",
					"8.8.4.4:53",
				},
			},
		}
		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				data = fetchData(ctx, opts)
				return nil
			}).
			Title("Collection health data...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("collecting health data failed")
		}

		tableOpts := tui.TableOpts{
			Wrap:      false,
			FullWidth: false,
			Style:     nil,
		}
		tableHeaders, tableData := transform(data.internetAccess)
		tui.PrintTable("Internet Connectivity", tableHeaders, tableData, tableOpts)

		// Print in sorted order
		for name, results := range data.TcpChecks {
			tableHeaders, tableData = transform(results)
			tui.PrintTable(name, tableHeaders, tableData, tableOpts)
		}

		for name, results := range data.DnsChecks {
			tableHeaders, tableData = transform(results)
			tui.PrintTable(name, tableHeaders, tableData, tableOpts)
		}

		tableHeaders, tableData = healthcheck.TransformLogs(data.logs)
		tui.PrintTable("Logs", tableHeaders, tableData, tableOpts)

		tui.PrintTable("Prometheus", []string{"Instance", "Avg ", "Max"}, data.prometheus, tableOpts)
	},
}

func fetchData(ctx context.Context, opts HealthcheckOptions) *Data {

	data := &Data{}

	wg := sync.WaitGroup{}
	for name, probes := range opts.TcpChecks {
		data.TcpChecks = make(map[string]map[string]healthcheck.ProbeResult, len(opts.TcpChecks))
		wg.Add(1)
		go func() {
			data.TcpChecks[name] = healthcheck.ProbeTcp(probes, 2*time.Second)
			wg.Done()
		}()
	}

	for name, probes := range opts.DnsChecks {
		data.DnsChecks = make(map[string]map[string]healthcheck.ProbeResult, len(opts.DnsChecks))
		wg.Add(1)
		go func() {
			data.DnsChecks[name] = healthcheck.CheckDnsServers(probes.DnsServers, probes.DnsRecord, 2*time.Second)
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		data.internetAccess = healthcheck.CheckInternetConnection(ctx, opts.InternetConnectivity)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		var err error
		queryArgs := healthcheck.VictorialogsQuery{
			Address: "https://logs.rs.soeren.cloud",
			Query:   "error AND _time:15m",
			Limit:   5,
		}
		data.logs, err = healthcheck.QueryVictorialogs(ctx, queryArgs)
		if err != nil {
			log.Error().Err(err).Msg("could not get logs")
		}
		wg.Done()
	}()

	wg.Add(1)
	var avg, mot map[string]float64
	go func() {
		query := `avg_over_time(probe_duration_seconds{job="blackbox-icmp-vpn"}[30m]) * 1000`
		endpoint := "https://victoriametrics.rs.soeren.cloud/api/v1/query"
		avg, _ = healthcheck.QueryPrometheus(ctx, query, endpoint)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		query := `max_over_time(probe_duration_seconds{job="blackbox-icmp-vpn"}[30m]) * 1000`
		endpoint := "https://victoriametrics.rs.soeren.cloud/api/v1/query"
		mot, _ = healthcheck.QueryPrometheus(ctx, query, endpoint)
		wg.Done()
	}()

	wg.Wait()

	data.prometheus = healthcheck.MergePrometheusResults(avg, mot)
	return data
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
}

func transform(data map[string]healthcheck.ProbeResult) ([]string, [][]string) {
	headers := []string{
		"Host",
		"Status",
		"Duration",
	}

	ret := make([][]string, 0, len(data))
	for host, result := range data {
		status := "up"
		if result.Err != nil {
			status = fmt.Sprintf("down: %v", result.Err)
		}
		ret = append(ret, []string{host, status, fmt.Sprintf("%v", result.Duration.Round(time.Millisecond))})
	}

	return headers, ret
}
