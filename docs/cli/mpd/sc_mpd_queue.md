## sc mpd queue

Show and edit the queue

### Synopsis

Show the queue in a TUI, where enter plays the selected song, d removes it,
J/K move it down and up, pressing c twice clears the queue and tab switches
to the library.

If the output is not a terminal or with --print, the queue is printed instead,
marking the current song.

```
sc mpd queue [flags]
```

### Options

```
  -h, --help    help for queue
  -p, --print   Print the queue instead of opening the TUI
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

