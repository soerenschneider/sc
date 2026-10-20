## sc links open

Open a bookmark in the browser

### Synopsis

Open a bookmark in the default browser. The bookmark is given by its id, its URL
or a search query. If the query matches several bookmarks, or none is given, one
is picked interactively.

Examples:
  sc links open 42
  sc links open '#golang' generics

```
sc links open [id|url|query]... [flags]
```

### Options

```
  -h, --help   help for open
```

### Options inherited from parent commands

```
  -a, --address string      Linkding address.
      --no-telemetry        Do not perform check for updated version
      --profile string      Profile to use
  -t, --token string        Linkding token.
  -T, --token-file string   Path to a file containing the Linkding token.
  -v, --verbose             Print debug logs
```

### SEE ALSO

* [sc links](sc_links.md)	 - Manage bookmarks in Linkding

