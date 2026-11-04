## sc torrent list

List torrents in the Transmission client

### Synopsis

This command lists all torrents currently managed by the Transmission client using the Transmission API.

It displays relevant information such as torrent ID, name, status, progress, and more.

Examples:
  sc torrent list     # List all torrents

```
sc torrent list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
  -a, --address string   Transmission address.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc torrent](sc_torrent.md)	 - Commands for interacting with the Transmission torrent client

