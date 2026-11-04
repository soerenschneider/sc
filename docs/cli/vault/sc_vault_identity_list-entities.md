## sc vault identity list-entities

List identity entities in Vault

### Synopsis

Lists the names of all entities managed by the Vault identity system.

Note: Appropriate Vault permissions are required to access identity entity data.

```
sc vault identity list-entities [flags]
```

### Options

```
  -h, --help   help for list-entities
```

### Options inherited from parent commands

```
  -a, --address string   Vault address. If not specified, tries to read env variable VAULT_ADDR.
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault identity](sc_vault_identity.md)	 - Manage Vault identities

