## sc pass show

Pick an entry and copy or show its secret

### Synopsis

Pick an entry of the pass password store in a filterable list. By default the
first line of the entry is copied to the clipboard, with --show the whole entry
is displayed instead. With --otp, only OTP entries are offered and a one-time
code is generated.

```
sc pass show [flags]
```

### Options

```
  -h, --help                help for show
  -o, --otp                 Get an OTP code. The entry picker will only show OTP entries.
      --otp-prefix string   Sets the otp path prefix. (default "otp/")
  -s, --show                Show secret data on the output. If set to false (default), the first line is copied to the clipboard.
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc pass](sc_pass.md)	 - Interacts with the pass unix password store

