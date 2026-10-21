## sc vault secret patch

Set or delete single keys of a Vault KV2 secret

### Synopsis

Set or delete single keys of a Vault KV2 secret, while keeping all other keys untouched.

Keys are set using key=value pairs. If the value is '-', it is read from stdin instead, which
keeps it out of the shell history and the process list. A single trailing newline is removed
from the value read from stdin. Only one key can read its value from stdin.

Keys are deleted using --delete, which can be given multiple times.

The secret is changed using Vault's native patch operation, which merges the changes on the
server without reading the secret. This requires the 'patch' capability on the secret and
Vault 1.9 or newer.

If patching is not possible, because the capability is missing, Vault is too old or the
secret does not exist yet, the secret is read and written instead. This requires the 'read'
and 'update' (or 'create' for a new secret) capabilities. The secret is written using
check-and-set, so changes that were made in the meantime are not overwritten. If the latest
version of the secret is deleted, it is not changed.

```
sc vault secret patch <path> [key=value...] [flags]
```

### Examples

```
  sc vault secret patch my/app username=admin port=5432
  pass generate -n 32 | sc vault secret patch my/app password=-
  sc vault secret patch my/app --delete old_password --delete legacy_token
```

### Options

```
  -d, --delete stringArray   Delete this key from the secret, can be given multiple times
  -h, --help                 help for patch
```

### Options inherited from parent commands

```
  -m, --mount string     Path where the KV2 secret engine is mounted (default "secret")
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault secret](sc_vault_secret.md)	 - Manages the Vault KV2 secret engine

