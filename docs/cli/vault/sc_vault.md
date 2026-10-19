## sc vault

Commands for interacting with HashiCorp Vault

### Synopsis

This command serves as the entry point for Vault-related functionality.

Use one of the available subcommands to interact with HashiCorp Vault for tasks
such as reading and writing secrets, authentication, or policy management.

Examples:
  sc vault login             # Authenticate with Vault
  sc vault ssh sign-key      # Interact with SSH secret engine

```
sc vault [flags]
```

### Options

```
  -a, --address string      Vault address. If not specified, tries to read env variable VAULT_ADDR.
  -h, --help                help for vault
  -t, --token-file string   Vault token file. (default "~/.vault-token")
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc](sc.md)	 - Universal Command Line Interface for soeren.cloud
* [sc vault approle](sc_vault_approle.md)	 - Manage Vault approle auth
* [sc vault aws](sc_vault_aws.md)	 - Manage AWS secret engine
* [sc vault identity](sc_vault_identity.md)	 - Manage Vault identities
* [sc vault login](sc_vault_login.md)	 - Authenticate with a Vault server using username and password
* [sc vault pki](sc_vault_pki.md)	 - Manages the Vault PKI secret engine
* [sc vault policy](sc_vault_policy.md)	 - Inspect Vault ACL policies
* [sc vault secret](sc_vault_secret.md)	 - Manages the Vault KV2 secret engine
* [sc vault ssh](sc_vault_ssh.md)	 - Manages the Vault SSH secret engine
* [sc vault token](sc_vault_token.md)	 - Manage Vault tokens
* [sc vault totp](sc_vault_totp.md)	 - Manage Vault TOTP MFA methods and secrets
* [sc vault transit](sc_vault_transit.md)	 - Manage Vault transit secret engine
* [sc vault update-password](sc_vault_update-password.md)	 - Update the password for Vault userpass authentication method

