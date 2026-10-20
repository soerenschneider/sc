## sc links delete

Delete a bookmark

### Synopsis

Delete a bookmark from Linkding. The bookmark is given by its id, its URL or a
search query. If the query matches several bookmarks, or none is given, one is
picked interactively.

The deletion has to be confirmed, unless --yes is given.

Examples:
  sc links delete 42
  sc links delete https://example.com --yes

```
sc links delete [id|url|query]... [flags]
```

### Options

```
  -h, --help   help for delete
  -y, --yes    Delete without asking for confirmation
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

