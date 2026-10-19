## sc tf

Run terragrunt on the stacks of the soeren.cloud repository

### Synopsis

Run terragrunt on the units below terraform/stacks/<stack>/envs/<env> of a
checkout of the soeren.cloud repository, with the unit's state encryption
passphrase decrypted from its tf-encryption.sops.yaml.

Without a subcommand, and with a terminal, a wizard asks for the stack, the
environment, the operation and whatever else the operation needs. Values given
as flags are not asked for again.

The checkout is given by --repo, a profile ("tf: {repo: ...}") or SC_TF_REPO.
Its terraform/.sc-contract must declare a version this sc supports. Operations
that change the infrastructure (apply, destroy, refresh, raw and the passphrase
rotation) refuse to run while the checkout is behind its upstream or the
upstream cannot be fetched, unless --allow-stale is given.

See terraform/README.md in the repository for state encryption and
passphrase rotation.

```
sc tf [flags]
```

### Examples

```
  sc tf
  sc tf list
  sc tf plan -s aws-s3
  sc tf apply -s aws-s3 -p latest
  sc tf output -s aws-s3 -n user_credentials | jq -r '.someuser.secret_key'
  sc tf plan -s aws-s3 -- -target=module.buckets
  sc tf raw -s aws-s3 -- state list
  sc tf rotate-passphrase-start -s aws-s3
```

### Options

```
      --allow-stale    Change the infrastructure although the checkout is behind its upstream or it cannot be fetched
  -e, --env string     Environment, terraform/stacks/<stack>/envs/<env> (default "prod")
  -h, --help           help for tf
  -i, --interactive    Ask for everything not given as a flag, even when --stack is given
      --repo string    Checkout of the soeren.cloud repository (default: $SC_TF_REPO)
  -s, --stack string   Stack below terraform/stacks/ that has a unit for the environment
  -y, --yes            Do not ask for confirmation, run non-interactively
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc](sc.md)	 - Universal Command Line Interface for soeren.cloud
* [sc tf apply](sc_tf_apply.md)	 - apply a saved plan, or plan and apply
* [sc tf destroy](sc_tf_destroy.md)	 - destroy everything
* [sc tf fmt](sc_tf_fmt.md)	 - tofu fmt -recursive -check -diff
* [sc tf init](sc_tf_init.md)	 - initialise backend, providers and modules
* [sc tf init-upgrade](sc_tf_init-upgrade.md)	 - init with -upgrade
* [sc tf list](sc_tf_list.md)	 - List the environments and their stacks
* [sc tf output](sc_tf_output.md)	 - print the outputs
* [sc tf plan](sc_tf_plan.md)	 - show what would change and write a plan file
* [sc tf plan-destroy](sc_tf_plan-destroy.md)	 - plan the destruction of everything
* [sc tf raw](sc_tf_raw.md)	 - any tofu command, with encryption and lock
* [sc tf refresh](sc_tf_refresh.md)	 - refresh the state
* [sc tf rotate-passphrase-finish](sc_tf_rotate-passphrase-finish.md)	 - finish a passphrase rotation of the stack
* [sc tf rotate-passphrase-start](sc_tf_rotate-passphrase-start.md)	 - start a passphrase rotation of the stack
* [sc tf show](sc_tf_show.md)	 - show a saved plan
* [sc tf test](sc_tf_test.md)	 - tofu test, no backend needed
* [sc tf validate](sc_tf_validate.md)	 - validate the configuration

