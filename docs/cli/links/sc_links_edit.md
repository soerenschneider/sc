## sc links edit

Edit the title, description, notes or tags of a bookmark

### Synopsis

Edit a bookmark in Linkding. The bookmark is given by its id, its URL or a
search query. If the query matches several bookmarks, or none is given, one is
picked interactively.

The fields to change are given as flags. Without any of them, a form to edit
all fields is shown. Tags may be separated by spaces or commas.

Examples:
  sc links edit 42                                  # Edit in a form
  sc links tag 42 --add-tag go --remove-tag inbox   # Change tags only
  sc links edit https://example.com --title "Example"

```
sc links edit [id|url|query]... [flags]
```

### Options

```
  -g, --add-tag stringArray      Tag to add, may be given multiple times
      --description string       New description
  -h, --help                     help for edit
      --notes string             New notes
  -r, --remove-tag stringArray   Tag to remove, may be given multiple times
      --title string             New title
```

### Options inherited from parent commands

```
  -a, --address string      Linkding address.
      --no-telemetry        Do not perform check for updated version
      --profile string      Profile to use
      --quiet               Only print errors
  -t, --token string        Linkding token.
  -T, --token-file string   Path to a file containing the Linkding token.
  -v, --verbose             Print debug logs
```

### SEE ALSO

* [sc links](sc_links.md)	 - Manage bookmarks in Linkding

