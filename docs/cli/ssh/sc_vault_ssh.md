## sc vault ssh

Sign SSH certificates or retrieve SSH CA data

```
sc vault ssh [flags]
```

### Options

```
  -h, --help                   help for ssh
  -m, --mount string           Path where the SSH secret engine is mounted (default "ssh")
  -a, --vault-address string   Vault address
```

### Options inherited from parent commands

```
      --no-telemetry   Do not perform check for updated version
  -v, --verbose        Print debug logs
```

### SEE ALSO

* [sc vault](sc_vault.md)	 - A brief description of your command
* [sc vault ssh list-roles](sc_vault_ssh_list-roles.md)	 - Lists all roles for the SSH secrets engine
* [sc vault ssh sign-key](sc_vault_ssh_sign-key.md)	 - Signs a SSH public key

