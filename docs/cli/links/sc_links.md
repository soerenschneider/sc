## sc links

Manage bookmarks in Linkding

### Synopsis

The 'links' command is the entry point for interacting with Linkding.

Use subcommands to add, search, open, edit and delete bookmarks and to list
the tags stored in Linkding.

Commands working on a single bookmark take its id, its URL or a search query.
If the query matches several bookmarks, one is picked interactively.

Examples:
  sc links query -q='#example-tag'   # Query links with specific filters
  sc links open golang               # Open the bookmark matching "golang"
  sc links tag 42 --add-tag go       # Add a tag to bookmark 42
  sc links delete https://example.com

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
* [sc links delete](sc_links_delete.md)	 - Delete a bookmark
* [sc links edit](sc_links_edit.md)	 - Edit the title, description, notes or tags of a bookmark
* [sc links open](sc_links_open.md)	 - Open a bookmark in the browser
* [sc links query](sc_links_query.md)	 - Search bookmarks in Linkding
* [sc links tags](sc_links_tags.md)	 - List all tags in Linkding

