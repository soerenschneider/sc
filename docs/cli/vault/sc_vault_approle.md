## sc vault approle

Manage Vault approle auth

### Synopsis

The 'approle' command group contains subcommands for the Vault AppRole auth method.

This command itself does not perform any actions. Instead, use one of its subcommands
to log in or inspect AppRoles.

```
sc vault approle [flags]
```

### Options

```
  -h, --help                    help for approle
  -m, --mount string            Vault mount of the approle auth method (default "approle")
  -r, --role string             AppRole role name
      --role-file string        AppRole role name file
  -i, --role-id string          AppRole role id
      --role-id–file string   AppRole role id file
  -s, --secret-id string        AppRole secret id
      --secret-id-file string   AppRole secret id file
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault](sc_vault.md)	 - Commands for interacting with HashiCorp Vault
* [sc vault approle list-accessors](sc_vault_approle_list-accessors.md)	 - List approle accessors
* [sc vault approle login](sc_vault_approle_login.md)	 - Login to Vault using approle

