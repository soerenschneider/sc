## sc ollama pull

Download models or update them to their latest version

### Synopsis

Download models from the Ollama library to the instance. Pulling a model
that already exists updates it to the latest version, only changed layers are
downloaded. With --all, every model on the instance is updated.

Examples:
  sc ollama pull llama3.2
  sc ollama pull qwen3:8b gemma3:12b
  sc ollama pull --all

```
sc ollama pull model... [flags]
```

### Options

```
      --all    Update all models on the instance
  -h, --help   help for pull
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

