# Hot Aisle CLI

> ⚠️ **Alpha Status**: This project is currently in early active development. APIs and functionality are subject to change. Features are incomplete and may change significantly.

A command-line interface tool for [Hot Aisle operations](https://admin.hotaisle.app/api/docs/). After installation, you can run `hotaisle` from the command line for list of commands.

# Installation

## Binary Releases

Download the latest release for your platform from the [Releases page](https://github.com/hotaisle/hotaisle-cli/releases).

### Debian/ubuntu apt

Download the appropriate deb package from the [Releases page](https://github.com/hotaisle/hotaisle-cli/releases) and install it with `dpkg -i`.

> ⚠️ **Alpha Status**: Below isn't implemented yet. Will want to use something like this: https://github.com/morph027/apt-repo-action/

```
wget -qO - https://hotaisle.github.io/hotaisle-cli/public.key | sudo apt-key add -
echo "deb https://hotaisle.github.io/hotaisle-cli/deb/ /" | sudo tee /etc/apt/sources.list.d/hotaisle.list
sudo apt update && sudo apt install hotaisle
```

### RPM-based distros (Fedora, CentOS, RHEL)

Download the appropriate rpm package from the [Releases page](https://github.com/hotaisle/hotaisle-cli/releases) and install it with  `dnf install`.

> ⚠️ **Alpha Status**: This isn't implemented yet. Need some help with this.

### Homebrew

* `brew install hotaisle/tap/hotaisle`

## Source

### Prerequisites

- Go 1.25 or later
- [Just](https://github.com/casey/just) command runner
- [act](https://github.com/nektos/act) (optional)

# Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

Our release process is automated with Github Actions ensuring that all tests pass before a release is published and that the CLI is built for all supported platforms. We've removed the human factor the release process, so you don't have to worry about the security of it. The main branch can only accept PRs from maintainers with verified commits.

PRs run tests.
Merge PR to main, runs a binary build, tag and GH release.

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
├── package/          # OS packaging
├── swagger.json      # Copy of our swagger file. https://admin.hotaisle.app/api/docs/swagger.json
└── main.go           # Application entry point
```

# License

See the LICENSE file for details.

# Maintainer

For questions or issues, contact: hello@hotaisle.ai
