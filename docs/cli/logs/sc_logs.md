## sc logs

Interact with VictoriaLogs for querying and managing logs

### Synopsis

The 'logs' command is the entry point for interacting with VictoriaLogs.

Use subcommands to query logs stored in VictoriaLogs, a high-performance log database.

Examples:
  sc logs query -q='level="error"'   # Query logs with specific filters

```
sc logs [flags]
```

### Options

```
  -a, --address string   Victorialogs address.
  -h, --help             help for logs
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc](sc.md)	 - Universal Command Line Interface for soeren.cloud
* [sc logs query](sc_logs_query.md)	 - Query logs from VictoriaLogs using a logQL-like expression

