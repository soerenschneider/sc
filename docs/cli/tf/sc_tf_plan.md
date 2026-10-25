## sc tf plan

show what would change and write a plan file

```
sc tf plan [-- <extra arguments for tofu>] [flags]
```

### Examples

```
  sc tf plan -s aws-s3
  sc tf plan -s aws-s3 -- -target=module.buckets
```

### Options

```
  -d, --detailed-exitcode   Exit 2 if there are changes, 0 if not
  -h, --help                help for plan
  -p, --plan-file string    Where to write the plan file, 'none' disables writing one (default: plans/<env>-<stack>-<timestamp>.tfplan)
```

### Options inherited from parent commands

```
      --allow-stale      Change the infrastructure although the checkout is behind its upstream or it cannot be fetched
  -e, --env string       Environment, terraform/stacks/<stack>/envs/<env> (default "prod")
  -i, --interactive      Ask for everything not given as a flag, even when --stack is given
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
      --repo string      Checkout of the soeren.cloud repository (default: $SC_TF_REPO)
  -s, --stack string     Stack below terraform/stacks/ that has a unit for the environment
  -v, --verbose          Print debug logs
  -y, --yes              Do not ask for confirmation, run non-interactively
```

### SEE ALSO

* [sc tf](sc_tf.md)	 - Run terragrunt on the stacks of the soeren.cloud repository

