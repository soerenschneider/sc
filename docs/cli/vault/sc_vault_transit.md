## sc vault transit

Manage Vault transit secret engine

### Synopsis

The 'transit' command group contains subcommands for interacting with the Vault transit secret engine.

This command itself does not perform any actions. Instead, use one of its subcommands.

```
sc vault transit [flags]
```

### Options

```
  -h, --help           help for transit
  -m, --mount string   Path where the transit secret engine is mounted. If not specified, available mounts are discovered.
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault](sc_vault.md)	 - Commands for interacting with HashiCorp Vault
* [sc vault transit decrypt](sc_vault_transit_decrypt.md)	 - Decrypt data using a transit key
* [sc vault transit encrypt](sc_vault_transit_encrypt.md)	 - Encrypt data using a transit key

