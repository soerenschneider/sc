## sc links

Manage bookmarks in Linkding

### Synopsis

The 'links' command is the entry point for interacting with Linkding.

Use subcommands to add and search bookmarks and list tags stored in Linkding.

Examples:
  sc links query -q='#example-tag'   # Query links with specific filters

```
sc links [flags]
```

### Options

```
  -a, --address string      Linkding address.
  -h, --help                help for links
  -t, --token string        Linkding token.
  -T, --token-file string   Path to a file containing the Linkding token.
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc](sc.md)	 - Universal Command Line Interface for soeren.cloud
* [sc links add](sc_links_add.md)	 - Add a bookmark to Linkding
* [sc links query](sc_links_query.md)	 - Search bookmarks in Linkding
* [sc links tags](sc_links_tags.md)	 - List all tags in Linkding

