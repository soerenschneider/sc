## sc torrent

Commands for interacting with the Transmission torrent client

### Synopsis

This command serves as the entry point for managing torrents via the Transmission API.

Use one of the available subcommands to interact with the Transmission client for tasks
such as adding new torrents, deleting existing ones, or listing current torrent activity.

Examples:
  sc torrent add -m "magnet:?xt=urn:btih:..."
  sc torrent list               # List all current torrents
  sc torrent delete <id>        # Delete a torrent by ID

```
sc torrent [flags]
```

### Options

```
  -a, --address string   Transmission address.
  -h, --help             help for torrent
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
* [sc torrent add](sc_torrent_add.md)	 - Add a new torrent to the Transmission client
* [sc torrent clean](sc_torrent_clean.md)	 - Remove finished torrents from the Transmission client
* [sc torrent delete](sc_torrent_delete.md)	 - Delete torrents from the Transmission client
* [sc torrent list](sc_torrent_list.md)	 - List torrents in the Transmission client

