## sc vault approle create-secret-id

Create a new secret_id for an AppRole

### Synopsis

Create a new secret_id for an AppRole. Unlike 'rotate-secret-id', existing secret_ids are kept.

The role is given using --role or --role-file. If it is not given and the command is run in a
terminal, the AppRoles are listed for selection.

Using --cidr, the new secret_id and the tokens issued with it are bound to the given CIDR blocks.
Using --cidr-host, the CIDR blocks are looked up dynamically: the host is resolved using DNS and
each of its IP addresses is added as a single host CIDR (/32 for IPv4, /128 for IPv6). Both
flags can be combined.

Use --json to print the new secret_id as JSON, e.g. for further processing with jq.

```
sc vault approle create-secret-id [flags]
```

### Examples

```
  sc vault approle create-secret-id
  sc vault approle create-secret-id -r my-role --metadata host=foo --ttl 720h
  sc vault approle create-secret-id -r my-role --cidr-host my-host.example.com --json | jq -r .secret_id
```

### Options

```
      --cidr stringArray          Bind the secret_id and the tokens issued with it to this CIDR block, can be given multiple times
      --cidr-host stringArray     Resolve this host using DNS and bind the secret_id and its tokens to its IP addresses, can be given multiple times
  -h, --help                      help for create-secret-id
      --json                      Print the new secret_id as JSON
      --metadata stringToString   Metadata to attach to the secret_id, e.g. --metadata host=foo,env=prod (default [])
      --ttl string                TTL of the secret_id, defaults to the secret_id_ttl of the role
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

