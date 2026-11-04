## sc vault token lookup

Lookup and display information about the current Vault token

### Synopsis

Retrieve and display information about the currently active Vault token.

This command queries the Vault API to show metadata about the current token,
such as its creation time, expiration, policies, and identity.

It attempts to authenticate using the following sources (in order):
  1. The VAULT_TOKEN environment variable
  2. A token loaded from the local configuration file (e.g. ~/.config/mycli/token)


```
sc vault token lookup [flags]
```

### Options

```
  -h, --help   help for lookup
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

