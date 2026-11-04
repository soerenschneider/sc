## sc vault aws list

Lists all existing roles in the AWS secrets engine.

### Synopsis

List all configured roles in the Vault AWS secrets engine.

This command queries the AWS secrets engine mounted in Vault to retrieve a list of
all available role names. These roles define the IAM permissions that Vault can
generate for temporary AWS credentials.

The role names are sorted alphabetically and printed one per line. Use --output json to print
them as JSON array instead.

The command expects that the AWS secrets engine has been enabled and configured.

```
sc vault aws list [flags]
```

### Examples

```
  sc vault aws list
  sc vault aws list -m my-aws-mount -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format: simple, json (default "simple")
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

