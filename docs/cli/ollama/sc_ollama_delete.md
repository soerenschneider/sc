## sc ollama delete

Delete models from the Ollama instance

### Synopsis

Delete models from the Ollama instance. Without arguments, the models are
picked interactively.

The deletion has to be confirmed, unless --yes is given.

Examples:
  sc ollama delete
  sc ollama delete llama3.1:8b --yes

```
sc ollama delete [model]... [flags]
```

### Options

```
  -h, --help   help for delete
  -y, --yes    Delete without asking for confirmation
```

### Options inherited from parent commands

```
  -a, --address string   Ollama address, host[:port] or URL
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc ollama](sc_ollama.md)	 - Manage the models of an Ollama instance

