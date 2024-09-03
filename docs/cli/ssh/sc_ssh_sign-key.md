## sc ssh sign-key

Signs a SSH public key

```
sc ssh sign-key [flags]
```

### Options

```
  -c, --cert-file string         Where to save the certificate to
  -h, --help                     help for sign-key
      --principals stringArray   Principals
  -p, --pub-key-file string      Location of the public key to sign
  -r, --role string              Vault role
  -t, --ttl string               TTL of the certificate (default "24h")
```

### Options inherited from parent commands

```
  -m, --mount string           Path where the SSH secret engine is mounted (default "ssh")
      --no-telemetry           Do not perform check for updated version
  -a, --vault-address string   Vault address
  -v, --verbose                Print debug logs
```

### SEE ALSO

* [sc ssh](sc_ssh.md)	 - Sign SSH certificates or retrieve SSH CA data

