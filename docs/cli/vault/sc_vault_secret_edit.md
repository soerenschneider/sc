## sc vault secret edit

Interactively edit a Vault KV2 secret

```
sc vault secret edit <path> [flags]
```

### Options

```
  -h, --help   help for edit
```

### Options inherited from parent commands

```
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
  -m, --mount string     Path where the KV2 secret engine is mounted (default "secret")
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault secret](sc_vault_secret.md)	 - Manages the Vault KV2 secret engine

