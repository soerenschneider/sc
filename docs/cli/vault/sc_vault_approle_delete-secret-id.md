## sc vault approle delete-secret-id

Delete a secret_id of an AppRole

### Synopsis

Deletes a secret_id of an AppRole, identified by its accessor.

The accessor may be passed as argument or via --accessor, the role via --role or --role-file.
If the role is not given, it is looked up using the accessor. If neither is given, an
interactive wizard lets you select the role first, offering only roles that have secret_ids,
and then the secret_id to delete, showing its status, expiration and metadata. The selection of the secret_id is skipped if the role
has only a single one.

Unless --force is given, the deletion has to be confirmed.

```
sc vault approle delete-secret-id [accessor] [flags]
```

### Examples

```
  sc vault approle delete-secret-id
  sc vault approle delete-secret-id -r my-role
  sc vault approle delete-secret-id 2f4e1c3a-0f9b-4d2e-9a6c-1b2c3d4e5f60 -f
```

### Options

```
      --accessor string   Accessor of the secret_id to delete
  -f, --force             Skip the confirmation prompt
  -h, --help              help for delete-secret-id
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

