## sc vault policy read

Display the rules of a Vault ACL policy

### Synopsis

Displays the rules of a Vault ACL policy.

The policy may be passed as argument or via --name. If it is not specified, all policies
are listed in an interactive browser: select a policy to display its rules and go back to
the list with esc. The policy displayed when quitting with q is printed. If the policies
can not be listed, the name is prompted for instead.

Examples:
  sc vault policy read
  sc vault policy read default
  sc vault policy read -n my-policy

```
sc vault policy read [name] [flags]
```

### Options

```
  -h, --help          help for read
  -n, --name string   Name of the policy. If not specified, available policies are offered for selection.
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc vault policy](sc_vault_policy.md)	 - Inspect Vault ACL policies

