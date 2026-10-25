## sc update

Update sc to the latest release

### Synopsis

Check github.com/soerenschneider/sc for a newer release and replace the
running binary with it.

The binary is downloaded next to the current one, verified against the
release's sha256 checksum, whose keyless cosign signature must have been made
by the release workflow of github.com/soerenschneider/sc, and then atomically
moved into place, so the directory containing the binary must be writable.

```
sc update [flags]
```

### Options

```
  -c, --check   Only check whether an update is available
  -f, --force   Install the latest release even if it is not newer than the running version
  -h, --help    help for update
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

