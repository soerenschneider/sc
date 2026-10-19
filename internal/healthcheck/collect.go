package healthcheck

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"sync"
	"time"
)

const probeTimeout = 2 * time.Second

// TcpGroup is a named set of host:port endpoints that are probed by opening a TCP connection.
type TcpGroup struct {
	Name      string   `yaml:"name"`
	Endpoints []string `yaml:"endpoints"`
}

// DnsGroup is a named set of DNS servers that are each asked to resolve Record.
type DnsGroup struct {
	Name    string   `yaml:"name"`
	Servers []string `yaml:"servers"`
	Record  string   `yaml:"record"`
}

// PrometheusLatency describes the queries for the average and maximum latency in milliseconds per instance.
type PrometheusLatency struct {
	Title    string `yaml:"title"`
	Endpoint string `yaml:"endpoint"`
	AvgQuery string `yaml:"avg-query"`
	MaxQuery string `yaml:"max-query"`
}

// Options configures the checks. The latency and logs sections are skipped if their address is not set.
type Options struct {
	Internet  InternetConnectivityOpts `yaml:"internet"`
	TcpGroups []TcpGroup               `yaml:"tcp"`
	DnsGroups []DnsGroup               `yaml:"dns"`
	Latency   PrometheusLatency        `yaml:"latency"`
	Logs      VictorialogsQuery        `yaml:"logs"`
}

type Probe struct {
	Name string
	ProbeResult
}

type Group struct {
	Name   string
	Probes []Probe
}

func (g Group) Up() int {
	up := 0
	for _, p := range g.Probes {
		if p.Err == nil {
			up++
		}
	}
	return up
}

type LatencyStat struct {
	Instance string
	Avg      float64
	Max      float64
}

type Report struct {
	Time   time.Time
	Took   time.Duration
	Groups []Group

	// LatencyTitle is the title of the latency card, empty if latency is not configured.
	LatencyTitle string
	Latency      []LatencyStat
	LatencyErr   error

	LogsEnabled bool
	Logs        []LogEntry
	LogsErr     error
}

// Counts returns the number of successful probes and the total number of probes.
func (r *Report) Counts() (up, total int) {
	for _, g := range r.Groups {
		up += g.Up()
		total += len(g.Probes)
	}
	return up, total
}

// Collect runs all checks concurrently and returns a report preserving the order of the configured groups.
func Collect(ctx context.Context, opts Options) *Report {
	start := time.Now()
	report := &Report{
		Time:        start,
		Groups:      make([]Group, 1+len(opts.TcpGroups)+len(opts.DnsGroups)),
		LogsEnabled: opts.Logs.Address != "",
	}
	if opts.Latency.Endpoint != "" {
		report.LatencyTitle = cmp.Or(opts.Latency.Title, "Latency")
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		results := CheckInternetConnection(ctx, opts.Internet)
		report.Groups[0] = Group{
			Name: "Internet",
			Probes: []Probe{
				{Name: "dns resolution", ProbeResult: results["DNS Check"]},
				{Name: "http", ProbeResult: results["HTTP Check"]},
				{Name: "public resolvers", ProbeResult: results["TCP Check"]},
			},
		}
	})

	for i, group := range opts.TcpGroups {
		wg.Go(func() {
			report.Groups[1+i] = toGroup(group.Name, group.Endpoints, ProbeTcp(group.Endpoints, probeTimeout))
		})
	}

	for i, group := range opts.DnsGroups {
		wg.Go(func() {
			results := CheckDnsServers(group.Servers, group.Record, probeTimeout)
			report.Groups[1+len(opts.TcpGroups)+i] = toGroup(group.Name, group.Servers, results)
		})
	}

	if report.LogsEnabled {
		wg.Go(func() {
			report.Logs, report.LogsErr = QueryVictorialogs(ctx, opts.Logs)
		})
	}

	if report.LatencyTitle != "" {
		wg.Go(func() {
			report.Latency, report.LatencyErr = queryLatency(ctx, opts.Latency)
		})
	}

	wg.Wait()
	report.Took = time.Since(start)
	return report
}

func toGroup(name string, order []string, results map[string]ProbeResult) Group {
	probes := make([]Probe, 0, len(order))
	for _, key := range order {
		probes = append(probes, Probe{Name: key, ProbeResult: results[key]})
	}
	return Group{Name: name, Probes: probes}
}

func queryLatency(ctx context.Context, opts PrometheusLatency) ([]LatencyStat, error) {
	var avg, maxima map[string]float64
	var avgErr, maxErr error

	var wg sync.WaitGroup
	wg.Go(func() { avg, avgErr = QueryPrometheus(ctx, opts.AvgQuery, opts.Endpoint) })
	wg.Go(func() { maxima, maxErr = QueryPrometheus(ctx, opts.MaxQuery, opts.Endpoint) })
	wg.Wait()

	if avgErr != nil {
		return nil, avgErr
	}
	if maxErr != nil {
		return nil, maxErr
	}

	stats := make([]LatencyStat, 0, len(avg))
	for instance, value := range avg {
		stats = append(stats, LatencyStat{Instance: instance, Avg: value, Max: maxima[instance]})
	}
	slices.SortFunc(stats, func(a, b LatencyStat) int {
		return strings.Compare(a.Instance, b.Instance)
	})
	return stats, nil
}
