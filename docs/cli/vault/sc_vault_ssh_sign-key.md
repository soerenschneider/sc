## sc vault ssh sign-key

Signs a SSH public key

### Synopsis

Signs a SSH public key using the Vault SSH secret engine and writes the certificate next to it.

Anything that is not passed as a flag is deduced if possible:
  - public key: the only public key in ~/.ssh, otherwise a selection is offered
  - mount: the only SSH mount, otherwise a selection of all SSH mounts is offered
  - role: the only role that is able to sign user keys, otherwise a selection is offered
  - principals: the users allowed by the role, a selection is offered if there are multiple
  - ttl: 24h, capped to the max_ttl of the role

```
sc vault ssh sign-key [flags]
```

### Options

```
  -c, --cert-file string         Where to save the certificate to
  -f, --force                    Request new certificate regardless of expiry date
  -h, --help                     help for sign-key
      --principals stringArray   Principals. If not specified, deduced from the users allowed by the role.
  -p, --pub-key-file string      Location of the public key to sign. If not specified, public keys in ~/.ssh are discovered.
  -r, --role string              Vault role. If not specified, available roles are discovered.
  -t, --ttl string               TTL of the certificate, capped to the max_ttl of the role if not specified (default "24h")
```

### Options inherited from parent commands

```
  -m, --mount string     Path where the SSH secret engine is mounted. If not specified, available mounts are discovered.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault ssh](sc_vault_ssh.md)	 - Manages the Vault SSH secret engine

