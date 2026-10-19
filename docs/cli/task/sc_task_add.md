## sc task add

Interactively create a taskwarrior task

### Synopsis

Interactive alternative to 'task add' with picker-driven project,
tag, and attribute selection.

Loads the taxonomy of known projects and tag categories from a YAML file
(default: ~/.sc/taskwarrior.yaml, override with --config). When the file
is absent, a small generic GTD taxonomy is used.

The Project step is a filterable picker with the configured
'default_project' pre-selected (falls back to "inbox" when the projects
list contains it). Press "/" to filter, esc/enter to exit filter mode,
↑↓ to navigate, enter to pick. To create a new project (or subproject
like "myproject.sub"), pick "＋ create new…" at the bottom of the list
and type the full dotted name in the follow-up input.

An optional Depends-on step lists pending tasks; "(none)" is the default,
so enter skips it. Pick a task to append depends:<ID> to the invocation.

```
sc task add [flags]
```

### Options

```
  -c, --config string   Path to taxonomy YAML (default ~/.sc/taskwarrior.yaml; falls back to generic GTD)
  -h, --help            help for add
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc task](sc_task.md)	 - Taskwarrior helpers

