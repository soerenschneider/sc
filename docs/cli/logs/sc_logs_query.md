## sc logs query

Query logs from VictoriaLogs using a logQL-like expression

### Synopsis

Query logs stored in VictoriaLogs by providing filter expressions and optional time ranges.

This command allows you to search logs with filters, labels, and time constraints. It supports various output formats
for integration with scripts or for human-readable views.

```
sc logs query [flags]
```

### Options

```
  -f, --follow                      Whether to follow logs
      --format string               Output format: text, json, raw (default "text")
  -h, --help                        help for query
  -i, --id-fields stringArray       Identity fields highlighted before the message (joined with '/'; use '_stream' to include the VictoriaLogs stream) (default [host,hostname,service,svc,app])
  -l, --limit int                   Maximum entries (bounded mode and replay phase) (default 100)
      --no-color                    Disable ANSI color in text output
  -q, --query string                Query (default "error AND _time:45m")
  -r, --refresh-interval duration   Tail refresh interval (must be > 0) (default 2s)
  -s, --since string                Start time: duration (1h), RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or Unix epoch
      --timeout duration            Timeout (default 10s)
  -u, --until string                End time (same formats as --since)
```

### Options inherited from parent commands

```
  -a, --address string   Victorialogs address.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc logs](sc_logs.md)	 - Interact with VictoriaLogs for querying and managing logs

