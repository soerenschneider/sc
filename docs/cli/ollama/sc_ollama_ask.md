## sc ollama ask

Ask a model a quick question

### Synopsis

Ask a model a question and stream the answer to stdout. Input piped to the
command is passed to the model along with the question, or is the question
itself if no other is given. Without either, the question is read in a form
that accepts multiple lines and pasted text: enter starts a new line, ctrl+d
sends the question and ctrl+e opens $EDITOR.

The model is given by --model, which is best set in a profile. Without it, the
model of the continued conversation is used or, in a terminal, picked
interactively.

Every exchange is saved, and --continue asks a follow-up question in the last
conversation. Without --continue, a new conversation is started.

In a terminal, the markdown of the answer is rendered as it streams, one
paragraph, list or code block at a time. --raw prints the markdown as is,
which is also what is printed if the output is piped.

Thinking models use their default unless --think is given. With --think, the
reasoning is printed to stderr, --think=false turns it off for models that
support it.

Examples:
  sc ollama ask -m qwen3:8b "what's the difference between SIGTERM and SIGKILL"
  sc ollama ask -c "and SIGHUP?"
  git diff --staged | sc ollama ask "write a conventional commit message"
  kubectl logs pod/foo | sc ollama ask "why is this crashing"

```
sc ollama ask [question]... [flags]
```

### Options

```
  -c, --continue        Continue the last conversation
  -h, --help            help for ask
  -m, --model string    Model to ask
      --raw             Print the answer as raw markdown
  -s, --system string   System prompt for a new conversation
      --think           Enable thinking and print the reasoning to stderr, or disable it with --think=false
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

