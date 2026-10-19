## sc mpd update

Rescan the music directory for changes

### Synopsis

Scan the music directory, or only path within it, for new, changed and
removed files. The scan runs in the background unless --wait is given.

```
sc mpd update [path] [flags]
```

### Options

```
  -h, --help   help for update
  -w, --wait   Wait for the update to finish
```

### Options inherited from parent commands

```
  -a, --address string    MPD address, host[:port] or path of a unix socket
      --no-telemetry      Do not perform check for updated version
      --password string   MPD password
      --profile string    Profile to use
  -v, --verbose           Print debug logs
```

### SEE ALSO

* [sc mpd](sc_mpd.md)	 - Control a Music Player Daemon

