## sc task groom

Interactively groom inbox tasks (assign project, tags, …)

### Synopsis

Walk every pending task with project:inbox and groom it: pick a real
project (which removes the inbox project), add tags/location/priority/size/
due, then apply with 'task <id> modify'.

Uses the same taxonomy YAML as 'task add' (default ~/.sc/taskwarrior.yaml,
override with --config). Tasks are processed oldest-first. For each one you
can Apply the changes, Skip it (leave it in the inbox), or Quit the sweep.

Like 'task add', the project and tags are pre-selected based on the task's
description, marked with ✨ and the reason. See 'sc task add --help'.

```
sc task groom [flags]
```

### Options

```
  -c, --config string   Path to taxonomy YAML (default ~/.sc/taskwarrior.yaml; falls back to generic GTD)
  -h, --help            help for groom
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc task](sc_task.md)	 - Taskwarrior helpers

