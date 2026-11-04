## sc vault approle list

List all AppRoles along with their configuration

### Synopsis

Lists all AppRoles of the mount, sorted alphabetically, along with their token policies, token TTLs
and secret_id settings.

If some roles can not be read, the remaining ones are printed nonetheless and the command exits
with an error. To list the secret_ids of the AppRoles, use 'list-accessors'.

The output format is chosen with --output:
  pretty  a styled table (default if stdout is a terminal)
  simple  plain, tab-aligned columns without styling, suited for grep and awk (default otherwise)
  json    JSON, e.g. for further processing with jq

```
sc vault approle list [flags]
```

### Examples

```
  sc vault approle list
  sc vault approle list -m my-approle-mount
  sc vault approle list -o simple | awk 'NR > 1 { print $1 }'
  sc vault approle list -o json | jq -r '.[] | select(.secret_id_ttl_seconds == 0) | .role'
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format: pretty, simple, json (default pretty if stdout is a terminal, simple otherwise)
```

### Options inherited from parent commands

```
  -a, --address string          Vault address. If not specified, tries to read env variable VAULT_ADDR.
  -m, --mount string            Vault mount of the approle auth method (default "approle")
      --no-telemetry            Do not perform check for updated version
      --profile string          Profile to use
      --quiet                   Only print errors
  -r, --role string             AppRole role name
      --role-file string        AppRole role name file
  -i, --role-id string          AppRole role id
      --role-id–file string   AppRole role id file
  -s, --secret-id string        AppRole secret id
      --secret-id-file string   AppRole secret id file
  -v, --verbose                 Print debug logs
```

### SEE ALSO

* [sc vault approle](sc_vault_approle.md)	 - Manage Vault approle auth

