# Hot Aisle CLI

> ⚠️ **Alpha Status**: This project is currently in early active development. APIs and functionality are subject to change. Features are incomplete and may change significantly.

A command-line interface tool for [Hot Aisle operations](https://admin.hotaisle.app/api/docs/).

## Installation

### From Releases

### From Homebrew

### From Source

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Prerequisites

- Go 1.25 or later
- [Just](https://github.com/casey/just) command runner
- [act](https://github.com/nektos/act) (optional)

# License

See LICENSE file for details.

# Maintainer

For questions or issues, contact: hello@hotaisle.ai

## Project Structure
```
hotaisle-cli/
├── client/           # API Client implementation (AI generated)
├── cmd/cli/          # CLI commands and application logic
├── internal/         # Internal packages
│   ├── api/          # API client
│   ├── config/       # Configuration management
│   └── log/          # Logging utilities
├── test/             # Test files and fixtures
├── bin/              # Built binaries (generated)
├── dist/             # Distribution builds (generated)
├── swagger.json      # Copy of our swagger file. https://admin.hotaisle.app/api/docs/swagger.json
└── main.go           # Application entry point
```
