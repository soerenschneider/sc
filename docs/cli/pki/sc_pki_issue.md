## sc pki issue

Issues a x509 certificate

```
sc pki issue [flags]
```

### Options

```
      --alt-names stringArray   Alternative names
      --ca-file string          File to save the ca to
  -c, --cert-file string        File to save the certificate to
  -n, --common-name string      The CN for the certificate
  -h, --help                    help for issue
  -s, --ip-sans stringArray     IP Sans
  -k, --key-file string         File to save the private key to
  -r, --role string             Vault role
  -t, --ttl string              TTL of the certificate (default "24h")
```

### Options inherited from parent commands

```
  -m, --mount string           Path where the PKI secret engine is mounted (default "pki")
      --no-telemetry           Do not perform check for updated version
  -a, --vault-address string   Vault address
  -v, --verbose                Print debug logs
```

### SEE ALSO

* [sc pki](sc_pki.md)	 - Sign, issue and revoke x509 certificates and retrieve x509 CA data

