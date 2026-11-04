## sc vault pki read

Reads ca data for the PKI secret engine

```
sc vault pki read [flags]
```

### Options

```
  -f, --full-chain   Read the full chain (default true)
  -h, --help         help for read
```

### Options inherited from parent commands

```
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
  -m, --mount string     Path where the PKI secret engine is mounted. If not specified, available mounts are discovered.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault pki](sc_vault_pki.md)	 - Manages the Vault PKI secret engine

