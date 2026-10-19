## sc links add

Add a bookmark to Linkding

### Synopsis

Add a bookmark to Linkding. If the URL already exists, the bookmark is updated
instead. Without --url, the URL is prompted for interactively.

Examples:
  sc links add -u https://example.com --tags go --tags cli

```
sc links add [flags]
```

### Options

```
  -h, --help               help for add
      --tags stringArray   Tags for the URL
  -u, --url string         URL to add
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

