## sc logs tail

Follow logs from VictoriaLogs as they arrive

### Synopsis

Follow logs matching a LogsQL query live until interrupted. The query is given as
arguments, which are joined with spaces, or with --query, and defaults to all logs.

With --since, matching entries of that time window are printed first, capped by
--limit, before following new ones.

Pipes that need the complete input, such as sort or stats, are not supported
by VictoriaLogs when tailing.

Examples:
  sc logs tail                            # Follow all logs
  sc logs tail error                      # Follow logs containing "error"
  sc logs tail 'service:nginx' -s 10m     # Show the last 10 minutes, then follow

```
sc logs tail [query]... [flags]
```

### Options

```
      --format string               Output format: text, json, raw (default "text")
  -h, --help                        help for tail
  -i, --id-fields stringArray       Identity fields highlighted before the message (joined with '/'; use '_stream' to include the VictoriaLogs stream) (default [host,hostname,service,svc,app])
  -l, --limit int                   Maximum entries printed for --since (default 100)
      --no-color                    Disable ANSI color in text output
  -q, --query string                Query, overridden by arguments (default "*")
  -r, --refresh-interval duration   How often the server checks for new entries (default 2s)
  -s, --since string                Print entries since this time before following: duration (1h), RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or Unix epoch
      --timeout duration            Timeout of the --since query (default 10s)
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

