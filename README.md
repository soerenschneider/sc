# sc


[![Go Report Card](https://goreportcard.com/badge/github.com/soerenschneider/sc)](https://goreportcard.com/report/github.com/soerenschneider/sc-agent)
![test-workflow](https://github.com/soerenschneider/sc/actions/workflows/test.yaml/badge.svg)
![release-workflow](https://github.com/soerenschneider/sc/actions/workflows/release-container.yaml/badge.svg)
![golangci-lint-workflow](https://github.com/soerenschneider/sc/actions/workflows/golangci-lint.yaml/badge.svg)

sc is a command-line interface (CLI) tool, similar to the AWS CLI, designed as an interface for [soeren.cloud](https://github.com/soerenschneider/soeren.cloud).

## Installation

```shell
go install github.com/soerenschneider/sc@latest
```

```shell
Universal Command Line Interface for soeren.cloud

Usage:
  sc [command]

Available Commands:
  agent       Interact with a remote sc-agent instance
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  pki         Sign, issue and revoke x509 certificates and retrieve x509 CA data
  ssh         Sign SSH certificates or retrieve SSH CA data
  version     Print version and exit

Flags:
  -h, --help           help for sc
      --no-telemetry   Do not perform check for updated version
  -v, --verbose        Print debug logs

Use "sc [command] --help" for more information about a command.
```

## Documentation

Detailed documentation for all CLI sub commands is available below
- [agent subcommand](./docs/cli/agent/sc_agent.md)
- [ssh subcommand](./docs/cli/ssh/sc_ssh.md)

## Code Generation

The majority of the functionality is auto-generated using the `cobra-cli` and `oapi-codegen` using [sc-agent's OpenAPI spec](https://github.com/soerenschneider/sc-agent/blob/main/openapi.yaml). It leverages the auto-generated libraries from [github.com/soerenschneider/sc-agent/pkg/api](https://github.com/soerenschneider/sc-agent/tree/main/pkg/api) to interact with the `sc-agent` REST API.
