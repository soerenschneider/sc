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
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault](sc_vault.md)	 - Commands for interacting with HashiCorp Vault
* [sc vault approle create-secret-id](sc_vault_approle_create-secret-id.md)	 - Create a new secret_id for an AppRole
* [sc vault approle delete-secret-id](sc_vault_approle_delete-secret-id.md)	 - Delete a secret_id of an AppRole
* [sc vault approle list](sc_vault_approle_list.md)	 - List the names of all AppRoles
* [sc vault approle list-accessors](sc_vault_approle_list-accessors.md)	 - List the secret_id accessors of AppRoles
* [sc vault approle login](sc_vault_approle_login.md)	 - Login to Vault using approle
* [sc vault approle read](sc_vault_approle_read.md)	 - Display the configuration of a Vault AppRole
* [sc vault approle rotate-secret-id](sc_vault_approle_rotate-secret-id.md)	 - Rotate the secret_id of an AppRole

