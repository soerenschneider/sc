## sc mpd play

Start or resume playback

### Synopsis

Resume playback if paused, otherwise start playing the queue. With a
position, play the song at that position of the queue, starting at 1 as
shown by "sc mpd queue".

```
sc mpd play [position] [flags]
```

### Options

```
  -h, --help   help for play
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

