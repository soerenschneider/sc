## sc vault transit decrypt

Decrypt data using a transit key

### Synopsis

Decrypts a ciphertext using a key of the Vault transit secret engine and prints the resulting plaintext.

The ciphertext is read from the file given by --input ('-' for stdin). If no input file is
given and data is piped to stdin, stdin is read. Otherwise, the ciphertext is prompted for
interactively.

Mount and key may be passed as flags. Anything that is not passed is discovered and
offered for selection, falling back to free-form input if discovery is not possible.

Examples:
  sc vault transit decrypt
  sc vault transit decrypt -m transit -k my-key -i secret.enc -o secret.txt
  sc vault transit encrypt -m transit -k my-key | sc vault transit decrypt -m transit -k my-key

```
sc vault transit decrypt [flags]
```

### Options

```
  -c, --context string   Context for key derivation, required if the key has derivation enabled
  -h, --help             help for decrypt
  -i, --input string     File to read the ciphertext from, '-' for stdin. If not specified, reads piped stdin or prompts.
  -k, --key string       Name of the transit key. If not specified, available keys are discovered.
  -o, --output string    File to write the plaintext to. If not specified, prints to stdout.
```

### Options inherited from parent commands

```
  -m, --mount string     Path where the transit secret engine is mounted. If not specified, available mounts are discovered.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault transit](sc_vault_transit.md)	 - Manage Vault transit secret engine

