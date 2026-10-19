## sc vault pki issue

Issues a x509 certificate

### Synopsis

Issues a x509 certificate using the Vault PKI secret engine.

Anything that is not passed as a flag is deduced if possible:
  - mount: the only PKI mount, otherwise a selection of all PKI mounts is offered
  - role: the only role, otherwise a selection is offered
  - common name, alt names and ip sans: taken from the existing certificate when renewing,
    otherwise the common name is prompted for with suggestions based on the role
  - ttl: 24h, capped to the max_ttl of the role

```
sc vault pki issue [flags]
```

### Options

```
      --alt-names stringArray   Alternative names
      --ca-file string          File to save the ca to
  -c, --cert-file string        File to save the certificate to
  -n, --common-name string      The CN for the certificate. If not specified, taken from the existing certificate or prompted for.
  -h, --help                    help for issue
  -s, --ip-sans stringArray     IP Sans
  -k, --key-file string         File to save the private key to. If not specified, prompted for.
  -d, --min-duration string     Minimum duration of the cert before forcing issuing a new certificate (default "15m")
  -l, --min-lifetime float32    Minimum percentage of cert lifetime left before forcing issuing a new certificate (default 10)
  -r, --role string             Vault role. If not specified, available roles are discovered.
  -t, --ttl string              TTL of the certificate, capped to the max_ttl of the role if not specified (default "24h")
```

### Options inherited from parent commands

```
  -m, --mount string     Path where the PKI secret engine is mounted. If not specified, available mounts are discovered.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault pki](sc_vault_pki.md)	 - Manages the Vault PKI secret engine

