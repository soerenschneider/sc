## sc vault aws generate-credentials

Generate AWS credentials using Vault

### Synopsis

The 'generate-credentials' command is part of the 'aws' command group, which provides tools
for working with the Vault AWS secrets engine.

This command is used to generate dynamic AWS credentials by interacting with Vault. 
It requires appropriate configuration and permissions set in Vault to access the AWS secrets engine.

```
sc vault aws generate-credentials [flags]
```

### Options

```
  -p, --aws-profile string      Specifies the name of the AWS credentials profile section (default "default")
  -h, --help                    help for generate-credentials
      --no-wait                 Do not wait for the credentials to become effective
  -r, --role string             Specifies the name of the role to generate credentials for
  -t, --ttl int                 Specify how long the credentials should be valid for in seconds (default 3600)
      --wait-timeout duration   How long to wait for the credentials to become effective (default 30s)
```

### Options inherited from parent commands

```
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
  -m, --mount string     Path where the AWS secret engine is mounted. If not specified, available mounts are discovered.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault aws](sc_vault_aws.md)	 - Manage AWS secret engine

