## sc pw gen

Generates a random password

```
sc pw gen [flags]
```

### Options

```
  -h, --help                                   help for gen
  -l, --language string                        Wordlist language, e.g. 'en', 'de' (default "de")
      --lowercase                              Converts all words to lowercase
  -p, --print                                  Print password to stdout, do not copy to clipboard
  -s, --separator string                       Separator between words (default " ")
      --special                                Include special character
      --special-pos SpecialCharacterPosition   Position of special characters (default append)
  -w, --words int                              Number of words in the password (default 4)
```

### Options inherited from parent commands

```
      --no-telemetry     Do not perform check for updated version
      --profile string   Profile to use
  -v, --verbose          Print debug logs
```

### SEE ALSO

* [sc pw](sc_pw.md)	 - Generate secure passwords

