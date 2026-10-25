## sc mpd

Control a Music Player Daemon

### Synopsis

Control playback and browse the database of a Music Player Daemon.

The address defaults to MPD_HOST and MPD_PORT like for mpc, and then to
localhost:6600. Paths are used as unix sockets, and a password may be given
as "password@host".

Examples:
  sc mpd status
  sc mpd queue
  sc mpd toggle
  sc mpd next
  sc mpd search -t artist radiohead
  sc mpd random off
  sc mpd update -w
  sc mpd browse Artist/Album
  sc mpd -a music.lan:6600 clear

```
sc mpd [flags]
```

### Options

```
  -a, --address string    MPD address, host[:port] or path of a unix socket
  -h, --help              help for mpd
      --password string   MPD password
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
* [sc mpd browse](sc_mpd_browse.md)	 - Browse the database in a TUI and control playback
* [sc mpd clear](sc_mpd_clear.md)	 - Remove all songs from the queue
* [sc mpd next](sc_mpd_next.md)	 - Play the next song in the queue
* [sc mpd play](sc_mpd_play.md)	 - Start or resume playback
* [sc mpd prev](sc_mpd_prev.md)	 - Play the previous song in the queue
* [sc mpd queue](sc_mpd_queue.md)	 - Show and edit the queue
* [sc mpd random](sc_mpd_random.md)	 - Toggle or set random mode
* [sc mpd repeat](sc_mpd_repeat.md)	 - Toggle or set repeat mode
* [sc mpd search](sc_mpd_search.md)	 - Search the database
* [sc mpd status](sc_mpd_status.md)	 - Show the current song and player state
* [sc mpd stop](sc_mpd_stop.md)	 - Stop playback
* [sc mpd toggle](sc_mpd_toggle.md)	 - Toggle between play and pause
* [sc mpd update](sc_mpd_update.md)	 - Rescan the music directory for changes

