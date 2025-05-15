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
	internetAccess map[string]healthcheck.ProbeResult
	vpn            map[string]healthcheck.ProbeResult
	kubernetes     map[string]healthcheck.ProbeResult
	dnsServers     map[string]healthcheck.ProbeResult
}

var healthcheckCmd = &cobra.Command{
	Use: "healthcheck",
	Aliases: []string{
		"health",
	},
	Short: "Sign, issue and revoke x509 certificates and retrieve x509 CA data",
	Run: func(cmd *cobra.Command, args []string) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var data *Data

		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				data = fetchData(ctx)
				return nil
			}).
			Title("Collection health data...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("collecting health data failed")
		}

		tableHeaders, tableData := transform(data.internetAccess)
		tui.PrintTable("Internet Connectivity", tableHeaders, tableData)

		tableHeaders, tableData = transform(data.dnsServers)
		tui.PrintTable("Dns Servers", tableHeaders, tableData)

		// Print in sorted order
		tableHeaders, tableData = transform(data.vpn)
		tui.PrintTable("VPN", tableHeaders, tableData)

		tableHeaders, tableData = transform(data.kubernetes)
		tui.PrintTable("Kubernetes Clusters", tableHeaders, tableData)

		tui.PrintTable("Prometheus", []string{"Instance", "Avg ", "Max"}, data.prometheus)
	},
}

func fetchData(ctx context.Context) *Data {
	endpoints := []string{
		"router.dd.soeren.cloud:22",
		"router.ez.soeren.cloud:22",
		"router.pt.soeren.cloud:22",
		"rs.soeren.cloud:22",
		"swiss.soeren.cloud:22",
	}

	kubernetes := []string{
		"rs.soeren.cloud:6443",
		"k8s.dd.soeren.cloud:6443",
		"k8s.ez.soeren.cloud:6443",
		"k8s.pt.soeren.cloud:6443",
	}

	dnsServers := []string{
		"192.168.65.1",
		"192.168.2.3",
		"192.168.73.1",
		"192.168.200.1",
	}

	data := &Data{}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		data.vpn = healthcheck.ProbeTcp(endpoints, 2*time.Second)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		data.internetAccess = healthcheck.CheckInternetConnection(ctx)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		data.kubernetes = healthcheck.ProbeTcp(kubernetes, 2*time.Second)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		data.dnsServers = healthcheck.CheckDnsServers(dnsServers, "google.com", 2*time.Second)
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
