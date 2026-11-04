## sc vault totp delete-secret

Delete an entity's TOTP secret

### Synopsis

Destroys the TOTP secret of an identity entity for a given TOTP MFA method.

Once the secret has been destroyed, the entity can no longer satisfy this MFA method
until a new secret has been generated using the 'generate-secret' subcommand.

Method id and entity name may be passed as flags. Anything that is not passed is
prompted for interactively.

```
sc vault totp delete-secret [flags]
```

### Options

```
  -e, --entity-id string   Identity Entity ID
  -f, --force              Skip the confirmation prompt
  -h, --help               help for delete-secret
  -m, --method-id string   TOTP method ID
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

* [sc vault totp](sc_vault_totp.md)	 - Manage Vault TOTP MFA methods and secrets

