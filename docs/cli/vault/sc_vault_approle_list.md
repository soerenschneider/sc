## sc vault approle list

List the names of all AppRoles

### Synopsis

Lists the names of all AppRoles of the mount, sorted alphabetically and printed one per line.

Use --output json to print them as JSON array instead. To list the secret_ids of the AppRoles,
use 'list-accessors'.

```
sc vault approle list [flags]
```

### Examples

```
  sc vault approle list
  sc vault approle list -m my-approle-mount
  sc vault approle list -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format: simple, json (default "simple")
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

