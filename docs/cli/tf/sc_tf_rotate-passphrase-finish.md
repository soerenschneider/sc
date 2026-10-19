## sc tf rotate-passphrase-finish

finish a passphrase rotation of the stack

### Synopsis

Checks that the stack reads its state with the new passphrase alone, then drops the previous one. Commit and push the secrets file afterwards.

```
sc tf rotate-passphrase-finish [-- <extra arguments for tofu>] [flags]
```

### Examples

```
  sc tf rotate-passphrase-finish -s aws-s3
```

### Options

```
  -h, --help   help for rotate-passphrase-finish
```

### Options inherited from parent commands

```
      --allow-stale      Change the infrastructure although the checkout is behind its upstream or it cannot be fetched
  -e, --env string       Environment, terraform/stacks/<stack>/envs/<env> (default "prod")
  -i, --interactive      Ask for everything not given as a flag, even when --stack is given
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --repo string      Checkout of the soeren.cloud repository (default: $SC_TF_REPO)
  -s, --stack string     Stack below terraform/stacks/ that has a unit for the environment
  -v, --verbose          Print debug logs
  -y, --yes              Do not ask for confirmation, run non-interactively
```

### SEE ALSO

* [sc tf](sc_tf.md)	 - Run terragrunt on the stacks of the soeren.cloud repository

