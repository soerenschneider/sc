## sc links query

Search bookmarks in Linkding

### Synopsis

Search bookmarks in Linkding and display them in a table. The query uses
Linkding's search syntax, e.g. #tag to filter by tag. Without --query, the
query is prompted for interactively, offering the existing tags as suggestions.

Examples:
  sc links query -q '#golang'

```
sc links query [flags]
```

### Options

```
  -h, --help           help for query
  -q, --query string   Query. Can be empty to return all bookmarks. To search for tags, tags must be preceded by a '#'.
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

