## sc mpd browse

Browse the database in a TUI and control playback

### Synopsis

Browse the music database starting at path, the root by default.

Enter opens a folder or plays a song, a adds the selected folder or song to the
queue, space toggles play/pause and pressing c twice clears the queue. Tab
switches to the queue, see "sc mpd queue".

```
sc mpd browse [path] [flags]
```

### Options

```
  -h, --help   help for browse
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

