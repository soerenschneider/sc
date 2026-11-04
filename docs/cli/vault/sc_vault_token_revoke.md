## sc vault token revoke

Revoke the current Vault token

### Synopsis

Revoke the currently active Vault token, including all of its child tokens.

Before revoking, information about the token is displayed and confirmation is
requested. Use --force to skip the confirmation prompt.

If the revoked token was read from the local token file (~/.vault-token), the
file is removed afterwards as the token is no longer usable.

```
sc vault token revoke [flags]
```

### Options

```
  -f, --force   Skip displaying token info and the confirmation prompt
  -h, --help    help for revoke
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

* [sc vault token](sc_vault_token.md)	 - Manage Vault tokens

