## sc vault token renew

Try to extend the TTL of the current Vault token

### Synopsis

Renew the currently active Vault token to extend its TTL.

The requested TTL is passed to Vault as increment and counts from now, not from
the current expiry. If no TTL is given, Vault applies the token's default TTL.
Vault caps the resulting TTL at the token's max TTL, so the granted TTL may be
lower than requested. Tokens that are not renewable can not be extended.

Information about the renewed token is displayed afterwards.

```
sc vault token renew [flags]
```

### Options

```
  -h, --help         help for renew
  -t, --ttl string   TTL to request, counted from now (e.g. 8h). Defaults to the token's default TTL
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

