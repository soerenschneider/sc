## sc tf rotate-passphrase-start

start a passphrase rotation of the stack

### Synopsis

Generates a new passphrase and keeps the current one as the previous one, which the stack falls back to until it was applied once. Commit and push the secrets file, apply the stack, then run rotate-passphrase-finish.

```
sc tf rotate-passphrase-start [-- <extra arguments for tofu>] [flags]
```

### Examples

```
  sc tf rotate-passphrase-start -s aws-s3
```

### Options

```
  -h, --help   help for rotate-passphrase-start
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

