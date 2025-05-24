## sc vault

A brief description of your command

### Synopsis

A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.

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
* [sc vault aws](sc_vault_aws.md)	 - Manage AWS secret engine
* [sc vault identity](sc_vault_identity.md)	 - Manage Vault identities
* [sc vault login](sc_vault_login.md)	 - Authenticate with a Vault server using username and password
* [sc vault pki](sc_vault_pki.md)	 - Manages the Vault pki secret engine
* [sc vault ssh](sc_vault_ssh.md)	 - Sign SSH certificates or retrieve SSH CA data
* [sc vault token](sc_vault_token.md)	 - Manage Vault tokens
* [sc vault totp](sc_vault_totp.md)	 - Manage Vault totp
* [sc vault update-password](sc_vault_update-password.md)	 - Update the password for Vault userpass authentication method

