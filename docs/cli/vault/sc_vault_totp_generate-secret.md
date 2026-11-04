## sc vault totp generate-secret

Generate a TOTP secret for an entity

### Synopsis

Generates the TOTP secret of an identity entity for a given TOTP MFA method. The
resulting otpauth URI is printed and displayed as QR code to be scanned with an
authenticator app.

Method id and entity name may be passed as flags. Anything that is not passed is
prompted for interactively. Use --force to replace an existing secret.

```
sc vault totp generate-secret [flags]
```

### Options

```
  -e, --entity-id string   Identity Entity ID
  -f, --force              Force overwriting of existing TOTP secrets
  -h, --help               help for generate-secret
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

