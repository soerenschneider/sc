## sc tf raw

any tofu command, with encryption and lock

### Synopsis

Passes everything after '--' to tofu unchanged, for the commands sc tf does not wrap. It still sets up state encryption and the lock, which is why it exists at all.

```
sc tf raw [-- <extra arguments for tofu>] [flags]
```

### Examples

```
  sc tf raw -s aws-s3 -- import 'module.buckets["x"].aws_s3_bucket.this' x
  sc tf raw -s aws-s3 -- state list
```

### Options

```
  -h, --help   help for raw
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

