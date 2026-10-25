## sc torrent delete

Delete torrents from the Transmission client

### Synopsis

This command deletes one or more torrents from the Transmission client using the Transmission API.

You can specify one or more torrent IDs to delete. Optionally, you can also delete the associated downloaded data.

Examples:
  sc torrent delete --ids 1,2,3         # Delete torrents with IDs 1, 2, and 3
  sc torrent delete --ids 4 -d          # Delete torrent with ID 4 and its data

```
sc torrent delete [flags]
```

### Options

```
  -d, --delete-data     Delete all data of a torrent
  -h, --help            help for delete
      --id int64Slice   ID(s) to delete (default [])
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

