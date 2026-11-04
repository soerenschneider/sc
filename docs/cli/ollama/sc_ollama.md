## sc ollama

Manage the models of an Ollama instance

### Synopsis

Download, delete, inspect and browse the models of an Ollama instance, and
ask them quick questions.

The address defaults to OLLAMA_HOST like for the ollama CLI, and then to
http://localhost:11434.

Examples:
  sc ollama list
  sc ollama pull llama3.2 qwen3:8b
  sc ollama pull --all
  sc ollama delete llama3.1
  sc ollama show qwen3:8b
  sc ollama browse
  sc ollama ps
  sc ollama ask -m qwen3:8b "how do I undo the last git commit"
  git diff | sc ollama ask -c "and what about this change?"
  sc ollama -a gpu.lan list

```
sc ollama [flags]
```

### Options

```
  -a, --address string   Ollama address, host[:port] or URL
  -h, --help             help for ollama
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
      --quiet            Only print errors
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc](sc.md)	 - Universal Command Line Interface for soeren.cloud
* [sc ollama ask](sc_ollama_ask.md)	 - Ask a model a quick question
* [sc ollama browse](sc_ollama_browse.md)	 - Browse the models of the Ollama instance in a TUI
* [sc ollama delete](sc_ollama_delete.md)	 - Delete models from the Ollama instance
* [sc ollama list](sc_ollama_list.md)	 - List the models available on the Ollama instance
* [sc ollama ps](sc_ollama_ps.md)	 - List the models currently loaded into memory
* [sc ollama pull](sc_ollama_pull.md)	 - Download models or update them to their latest version
* [sc ollama show](sc_ollama_show.md)	 - Show the details of a model
* [sc ollama stop](sc_ollama_stop.md)	 - Unload models from memory

