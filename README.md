# Hot Aisle CLI

A command-line interface tool for [Hot Aisle operations](https://admin.hotaisle.app/api/docs/). After installation, you can run `hotaisle` from the command line for a list of commands.

## Quickstart

After [installation](#installation), run:

```bash
hotaisle setup
hotaisle doctor
hotaisle vm available
```

Setup guides you through authentication and team selection, verifies access, and offers to register an SSH public key. Get your initial API key using the [admin TUI](#getting-an-api-key). Token entry in a terminal is hidden. Setup never provisions compute or purchases credits.

For agents and CI, supply `HOTAISLE_API_TOKEN` through your secret manager, then run:

```bash
hotaisle --json setup --non-interactive --team YOUR_TEAM \
  --ssh-key-file ~/.ssh/id_ed25519.pub
hotaisle --json doctor
hotaisle --json vm available
```

Setup automatically selects a team when exactly one is accessible. Multiple teams require an explicit choice. Prompts are disabled for piped input, `--non-interactive`, and `--json`. To read a token from a secret manager's stdout instead, pipe it into `hotaisle --json setup --token-stdin --non-interactive --team YOUR_TEAM`.

Setup saves credentials and the default team to the selected configuration file. For ephemeral credentials without writing a file, set both `HOTAISLE_API_TOKEN` and `HOTAISLE_DEFAULT_TEAM`, and use resource commands directly.

# Installation

## Binary Releases

Download the latest release for your platform from the [Releases page](https://github.com/hotaisle/hotaisle-cli/releases).

### Debian/ubuntu apt

```bash
# Add repository
echo "deb [signed-by=/usr/share/keyrings/hotaisle-gpg.key] https://hotaisle.github.io/apt-repo stable main" | sudo tee /etc/apt/sources.list.d/hotaisle.list

# Add GPG key
curl -fsSL https://hotaisle.github.io/apt-repo/gpg.key | sudo tee /usr/share/keyrings/hotaisle-gpg.key > /dev/null

# Update and install
sudo apt update
sudo apt install hotaisle
```

### By hand

Download the appropriate deb package from the [Releases page](https://github.com/hotaisle/hotaisle-cli/releases) and install it with `dpkg -i`.

### RPM-based distros (Fedora, CentOS, RHEL)

```bash
# Add repository
sudo tee /etc/yum.repos.d/hotaisle.repo > /dev/null <<'EOF'
[hotaisle]
name=Hot Aisle
baseurl=https://hotaisle.github.io/rpm-repo
enabled=1
repo_gpgcheck=1
gpgcheck=0
gpgkey=https://hotaisle.github.io/rpm-repo/gpg.key
EOF

# Install
sudo dnf install hotaisle
```

### By hand (RPM)

Download the appropriate rpm package from the [Releases page](https://github.com/hotaisle/hotaisle-cli/releases) and install it with `dnf install`.

### Homebrew

* `brew install hotaisle/tap/hotaisle`

## Source

### Prerequisites

- Go 1.27.3 or later
- [Just](https://github.com/casey/just) command runner
- [act](https://github.com/nektos/act) (optional)

# Getting an API key

When you log in to the admin TUI via `ssh admin.hotaisle.app`, check the breadcrumbs at the top; you’ll likely start in the team settings. Press Esc, then use the arrow keys to move up to your name to edit your personal settings, including API keys.

## Configuration and compatibility

Existing commands, GPU provisioning flags, config keys, and default result formats remain supported. Reads continue to return their existing JSON shapes; mutations retain their existing text messages unless `--json` is supplied. Normal failures retain exit code 1. No major version bump is needed for these additions; future incompatible changes must use the repository's Conventional Commit breaking-change mechanism.

The default config is `~/.hotaisle/config.json`. `--config-file PATH` (or `-c PATH`) overrides `HOTAISLE_CONFIG_FILE`, which overrides the default path. Both reads and writes use the selected file. Missing files are created only when saving configuration; help and version work even when configuration is invalid.

`HOTAISLE_API_TOKEN` overrides the saved token for API requests without persisting it. `HOTAISLE_DEFAULT_TEAM` overrides the saved default team; an explicit resource `--team` overrides both. The existing `config set token` still saves the environment token, and `config get token` still prints the saved secret. Avoid including that output in agent transcripts or logs.

## Automation output

Put `--json` before the command, for example `hotaisle --json vm start --vm YOUR_VM`. Reads retain their existing shapes. Mutations that previously printed text return `{"success":true,"message":"..."}`. Config getters return an object keyed by the configuration name. JSON documents end with a newline.

Failures emit a single JSON document to stderr and leave stdout empty:

```json
{"error":{"code":"authentication","message":"...","hint":"...","http_status":401,"retryable":false}}
```

`hint` and `http_status` are present when applicable. Exit codes in JSON mode are:

| Exit | Meaning |
| --- | --- |
| 0 | Success; doctor warnings appear in its `warnings` array |
| 1 | Other failure, including an unexpected resource state |
| 2 | Invalid arguments or configuration |
| 3 | Missing credentials, authentication failure, or denied access |
| 4 | Resource not found (HTTP 404) |
| 5 | Resource conflict (HTTP 409) |
| 6 | Temporary network/server failure or rate limiting |
| 7 | Deadline exceeded |
| 130 | Canceled |

The CLI does not automatically retry mutations. Provisioning or rebuild requests can continue after an HTTP timeout; inspect `vm list` and `vm state` before retrying. `vm stop` continues billing; `vm delete` ends it, subject to minimum reservation charges.

# Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

Our release process is automated with GitHub Actions ensuring that all tests pass before a release is published and that the CLI is built for all supported platforms. We've removed the human factor from the release process, so you don't have to worry about the security of it. The main branch can only accept PRs from maintainers with verified commits.

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
