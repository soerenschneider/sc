## sc healthcheck

Display a dashboard of the health of soeren.cloud

### Synopsis

Display a dashboard of the health of soeren.cloud.

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

Exits with status 1 if any probe failed, unless --watch is given.

```
sc healthcheck [flags]
```

### Options

```
      --domain string       Domain to strip from host names
  -h, --help                help for healthcheck
      --interval duration   Refresh interval in watch mode (default 30s)
  -w, --watch               Keep refreshing the dashboard
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc](sc.md)	 - Universal Command Line Interface for soeren.cloud

