## sc vault unwrap

Unwrap a response wrapping token

### Synopsis

Unwrap a response wrapping token and print the data it wraps.

The wrapping token is used to authenticate the request, so no login is required. Wrapping tokens
can only be unwrapped once.

The token is read from the argument. If it is not given, it is read from stdin if data is piped
to it, and prompted for otherwise.

Use --field to print the value of a single field only, or --json to print the data as JSON.

```
sc vault unwrap [token] [flags]
```

### Examples

```
  sc vault unwrap
  sc vault unwrap s.Xyz --field secret_id
  sc vault approle rotate-secret-id -r my-role --wrap-ttl 5m -f --json | jq -r .wrapping_token | sc vault unwrap --json
```

### Options

```
      --field string   Print the value of this field only
  -h, --help           help for unwrap
      --json           Print the wrapped data as JSON
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

