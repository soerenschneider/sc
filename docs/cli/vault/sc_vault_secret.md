## sc vault secret

Manages the Vault KV2 secret engine

```
sc vault secret [flags]
```

### Options

```
  -h, --help           help for secret
  -m, --mount string   Path where the KV2 secret engine is mounted (default "secret")
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault](sc_vault.md)	 - Commands for interacting with HashiCorp Vault
* [sc vault secret edit](sc_vault_secret_edit.md)	 - Interactively edit a Vault KV2 secret
* [sc vault secret ls](sc_vault_secret_ls.md)	 - Browse Vault KV v2 secrets in a TUI
* [sc vault secret patch](sc_vault_secret_patch.md)	 - Set or delete single keys of a Vault KV2 secret
* [sc vault secret sync](sc_vault_secret_sync.md)	 - Syncs a KV2 secret to disk

