## sc torrent add

Add a new torrent to the Transmission client

### Synopsis

This command adds a new torrent to the Transmission client using the Transmission API.

You can provide a magnet link or a torrent file URL to initiate a download.

Examples:
  sc torrent add -m "magnet:?xt=urn:btih:..."
  sc torrent add -m "https://example.com/my.torrent"

```
sc torrent add [flags]
```

### Options

```
  -h, --help            help for add
  -m, --magnet string   Magnet to add
  -n, --no-confirm      Try to parse magnet from the clipboard and do not ask for confirmation
```

### Options inherited from parent commands

```
  -a, --address string   Transmission address.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc torrent](sc_torrent.md)	 - Commands for interacting with the Transmission torrent client

