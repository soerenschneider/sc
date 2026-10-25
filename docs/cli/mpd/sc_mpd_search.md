## sc mpd search

Search the database

### Synopsis

Search the database for songs whose tag contains the query, ignoring case.
All arguments are joined to a single query. By default all tags and the file
name are searched, --tag restricts the search to a single tag such as artist,
album, title or file.

With --add, the songs found are appended to the queue.

```
sc mpd search <query>... [flags]
```

### Options

```
      --add          Append the songs found to the queue
  -h, --help         help for search
  -t, --tag string   Tag to search, e.g. artist, album, title or file (default "any")
```

### Options inherited from parent commands

```
  -a, --address string    MPD address, host[:port] or path of a unix socket
      --no-telemetry      Do not perform check for updated version
      --password string   MPD password
      --profile string    Profile to use
      --quiet             Only print errors
  -v, --verbose           Print debug logs
```

### SEE ALSO

* [sc mpd](sc_mpd.md)	 - Control a Music Player Daemon

