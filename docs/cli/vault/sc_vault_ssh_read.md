## sc vault ssh read

Reads ca data for the SSH secret engine

```
sc vault ssh read [flags]
```

### Options

```
  -h, --help   help for read
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

