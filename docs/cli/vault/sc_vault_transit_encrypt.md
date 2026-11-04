## sc vault transit encrypt

Encrypt data using a transit key

### Synopsis

Encrypts data using a key of the Vault transit secret engine and prints the resulting ciphertext.

The plaintext is read from the file given by --input ('-' for stdin). If no input file is
given and data is piped to stdin, stdin is read. Otherwise, the plaintext is prompted for
interactively.

Mount and key may be passed as flags. Anything that is not passed is discovered and
offered for selection, falling back to free-form input if discovery is not possible.

Examples:
  sc vault transit encrypt
  sc vault transit encrypt -m transit -k my-key -i secret.txt
  echo -n "hunter2" | sc vault transit encrypt -m transit -k my-key

```
sc vault transit encrypt [flags]
```

### Options

```
  -c, --context string   Context for key derivation, required if the key has derivation enabled
  -h, --help             help for encrypt
  -i, --input string     File to read the plaintext from, '-' for stdin. If not specified, reads piped stdin or prompts.
  -k, --key string       Name of the transit key. If not specified, available keys are discovered.
  -o, --output string    File to write the ciphertext to. If not specified, prints to stdout.
```

### Options inherited from parent commands

```
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
  -m, --mount string     Path where the transit secret engine is mounted. If not specified, available mounts are discovered.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault transit](sc_vault_transit.md)	 - Manage Vault transit secret engine

