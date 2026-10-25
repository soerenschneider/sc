## sc vault approle list-accessors

List the secret_id accessors of AppRoles

### Synopsis

Lists the secret_id accessors of all AppRoles, or of a single AppRole if it is passed as argument
or via --role / --role-file, along with whether the secret_ids are expired.

Roles without any secret_id are listed as well. If some roles or accessors can not be read,
the remaining ones are printed nonetheless and the command exits with an error.

The output format is chosen with --output:
  pretty  a styled table (default)
  simple  plain, tab-aligned columns without styling, suited for grep and awk
  json    JSON, e.g. for further processing with jq

```
sc vault approle list-accessors [role] [flags]
```

### Examples

```
  sc vault approle list-accessors
  sc vault approle list-accessors my-role
  sc vault approle accessors -o simple | grep expired
  sc vault approle list-accessors -o json | jq '.[] | select(.secret_ids | length == 0) | .role'
```

### Options

```
  -h, --help            help for list-accessors
  -o, --output string   Output format: pretty, simple, json (default "pretty")
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

