## sc vault totp

Manage Vault TOTP MFA methods and secrets

### Synopsis

The 'totp' command group contains subcommands for managing TOTP MFA methods and
the TOTP secrets of identity entities.

This command itself does not perform any actions. Instead, use one of its subcommands
to inspect or manage TOTP methods and secrets.

```
sc vault totp [flags]
```

### Options

```
  -h, --help   help for totp
```

### Options inherited from parent commands

```
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault](sc_vault.md)	 - Commands for interacting with HashiCorp Vault
* [sc vault totp delete-secret](sc_vault_totp_delete-secret.md)	 - Delete an entity's TOTP secret
* [sc vault totp generate-method](sc_vault_totp_generate-method.md)	 - Create a TOTP MFA method
* [sc vault totp generate-secret](sc_vault_totp_generate-secret.md)	 - Generate a TOTP secret for an entity
* [sc vault totp list-methods](sc_vault_totp_list-methods.md)	 - List TOTP MFA methods

