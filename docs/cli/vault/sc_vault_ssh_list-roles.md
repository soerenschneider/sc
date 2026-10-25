## sc vault ssh list-roles

Lists all roles for the SSH secrets engine

### Synopsis

Lists the names of all roles of the SSH secrets engine, sorted alphabetically and printed one
per line. Use --output json to print them as JSON array instead.

```
sc vault ssh list-roles [flags]
```

### Examples

```
  sc vault ssh list-roles
  sc vault ssh list-roles -m my-ssh-mount -o json
```

### Options

```
  -h, --help            help for list-roles
  -o, --output string   Output format: simple, json (default "simple")
```

### Options inherited from parent commands

```
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
  -m, --mount string     Path where the SSH secret engine is mounted. If not specified, available mounts are discovered.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault ssh](sc_vault_ssh.md)	 - Manages the Vault SSH secret engine

