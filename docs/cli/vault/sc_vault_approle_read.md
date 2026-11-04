## sc vault approle read

Display the configuration of a Vault AppRole

### Synopsis

Displays the configuration of a Vault AppRole, including its role_id if it can be read.

The role may be passed as argument or via --role / --role-file. If it is not specified, all
AppRoles are listed in an interactive browser: select a role to display its configuration and
go back to the list with esc. The role displayed when quitting with q is printed. If the
AppRoles can not be listed, the name is prompted for instead.

Examples:
  sc vault approle read
  sc vault approle read my-role
  sc vault approle read -r my-role -m my-approle-mount

```
sc vault approle read [role] [flags]
```

### Options

```
  -h, --help   help for read
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

