## sc vault approle rotate-secret-id

Rotate the secret_id of an AppRole

### Synopsis

Create a new secret_id for an AppRole and destroy all of its previous secret_ids afterwards.

The old secret_ids are only destroyed after the new one has been created successfully. If
destroying an old secret_id fails, the new secret_id is printed nonetheless and the command
exits with an error.

The role is given using --role or --role-file. If it is not given and the command is run in a
terminal, the AppRoles are listed for selection.

Using --cidr, the new secret_id and the tokens issued with it are bound to the given CIDR blocks.
Using --cidr-host, the CIDR blocks are looked up dynamically: the host is resolved using DNS and
each of its IP addresses is added as a single host CIDR (/32 for IPv4, /128 for IPv6). Both
flags can be combined.

Using --wrap-ttl, the new secret_id is response wrapped: instead of the secret_id, a single-use
wrapping token valid for the given TTL is printed, which can be unwrapped using
"vault unwrap <token>" or handed to Vault Agent.

When run without a terminal, e.g. by cron, --role (or --role-file) and --force are required.

Use --json to print the new secret_id as JSON, e.g. for further processing with jq.

```
sc vault approle rotate-secret-id [flags]
```

### Examples

```
  sc vault approle rotate-secret-id
  sc vault approle rotate-secret-id -r my-role
  sc vault approle rotate-secret-id -r my-role --cidr 10.0.0.0/8 --cidr 192.168.1.10/32 -f --json
  sc vault approle rotate-secret-id -r my-role --cidr-host my-host.example.com -f
  sc vault approle rotate-secret-id -r my-role --wrap-ttl 5m -f --json | jq -r .wrapping_token
```

### Options

```
      --cidr stringArray          Bind the secret_id and the tokens issued with it to this CIDR block, can be given multiple times
      --cidr-host stringArray     Resolve this host using DNS and bind the secret_id and its tokens to its IP addresses, can be given multiple times
  -f, --force                     Skip the confirmation prompt
  -h, --help                      help for rotate-secret-id
      --json                      Print the new secret_id as JSON
      --metadata stringToString   Metadata to attach to the secret_id, e.g. --metadata host=foo,env=prod (default [])
      --ttl string                TTL of the secret_id, defaults to the secret_id_ttl of the role
      --wrap-ttl string           Response wrap the new secret_id with the given TTL (e.g. 5m) and print the wrapping token instead
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

