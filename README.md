# Hot Aisle CLI

A command-line interface tool for [Hot Aisle operations](https://admin.hotaisle.app/api/docs/). After installation, you can run `hotaisle` from the command line for a list of commands.

## Quickstart

After [installation](#installation), run:

```bash
hotaisle quickstart
hotaisle login
hotaisle doctor
hotaisle offers
```

Setup guides you through authentication and team selection, verifies access, and offers to register an SSH public key. Get your initial API key using the [admin TUI](#getting-an-api-key). Token entry in a terminal is hidden. Setup never provisions compute or purchases credits.

`login` is an alias for `setup`. `quickstart` (alias `skill`) prints the Markdown skill embedded in your installed binary with its version and commit, without network access or loading credentials. Save that output as `SKILL.md` in your agent's supported skill directory, or pass it directly to the agent.

| Task | Direct command |
| --- | --- |
| Inspect identity and teams | `hotaisle whoami` |
| Manage API credentials | `hotaisle api-keys --help` |
| Add an SSH public key | `hotaisle ssh-keys add ~/.ssh/id_ed25519.pub` |
| Add a key and check an existing VM | `hotaisle ssh-keys add ~/.ssh/id_ed25519.pub --vm YOUR_VM` |
| List VM and bare metal rental options | `hotaisle offers` or `hotaisle rentals list` |
| Get an SSH command | `hotaisle ssh --vm YOUR_VM` |
| Get an SSH address | `hotaisle ssh --server YOUR_SERVER --address` |

For agents and CI, supply `HOTAISLE_API_TOKEN` through your secret manager, then run:

```bash
hotaisle --json setup --non-interactive --team YOUR_TEAM \
  --ssh-key-file ~/.ssh/id_ed25519.pub
hotaisle --json doctor
hotaisle --json vm available
```

Setup automatically selects a team when exactly one is accessible. Multiple teams require an explicit choice. Prompts are disabled for piped input, `--non-interactive`, `--json`, and `--agent`. To read a token from a secret manager's stdout instead, pipe it into `hotaisle --json setup --token-stdin --non-interactive --team YOUR_TEAM`.

Setup saves credentials and the default team to the selected configuration file. For ephemeral credentials without writing a file, set both `HOTAISLE_API_TOKEN` and `HOTAISLE_DEFAULT_TEAM`, and use resource commands directly.

See the [agent guide](docs/agents.md), [GPU inference walkthrough](docs/inference.md), and distributable [agent skill](skills/hotaisle/SKILL.md).

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

Existing commands, GPU provisioning flags, and config keys remain supported. Non-empty reads retain their existing JSON shapes; empty availability queries print an explicit message for terminal users and return `[]` with `--json`. Mutations retain their existing text messages unless `--json` or `--agent` is supplied. Normal failures retain exit code 1. Future incompatible changes must use the repository's Conventional Commit breaking-change mechanism.

The default config is `~/.hotaisle/config.json`. `--config-file PATH` (or `-c PATH`) overrides `HOTAISLE_CONFIG_FILE`, which overrides the default path. Both reads and writes use the selected file. Missing files are created only when saving configuration; help and version work even when configuration is invalid.

`HOTAISLE_API_TOKEN` overrides the saved token for API requests without persisting it. `HOTAISLE_DEFAULT_TEAM` overrides the saved default team; an explicit resource `--team` overrides both. The existing `config set token` still saves the environment token, and `config get token` still prints the saved secret. Avoid including that output in agent transcripts or logs.

## Automation output

Use `hotaisle --agent COMMAND` for Markdown output intended for models. Results use tables, exact JSON literals within cells, and escaped multiline text. Errors use Markdown on stderr with the same fields and exit codes as JSON mode. This mode never prompts. `--agent` and `--json` are mutually exclusive.

Put `--json` before the command, for example `hotaisle --json vm start --vm YOUR_VM`. Reads retain their existing shapes. Mutations that previously printed text return `{"success":true,"message":"..."}`. Config getters return an object keyed by the configuration name. JSON documents end with a newline.

Empty lists of resources, teams, members, invitations, and keys return `[]` in JSON mode. Terminal output explains what was not found; agent output includes a zero count and message. Populated lists keep their existing shapes.

An empty, `null`, malformed, or unusable resource object is an `invalid_response` error (exit 1), rather than a successful object filled with default values. Zero balances and other valid zero values remain data. Optional fields such as missing rental price quotes retain their meaningful `null` values.

`team purchase-credits` returns checkout details when payment can proceed. If the API instead requests payment approval (HTTP 202), the CLI explicitly reports `approval_requested:true` in structured output and explains that credits have not been added.

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

## SSH keys for running instances

`ssh-keys add` accepts a positional public key file, `--file PATH` (alias `--key-file`), `--key 'ssh-ed25519 ...'`, or `--stdin`. It reuses an existing account key. Key registration requires an API token with personal owner permission. A key is account-wide, even when `--vm`, `--server`, or `--team` is supplied; those flags check access and identify a target, rather than limiting the key's scope.

Managed instances fetch authorized keys at the next SSH login as `hotaisle`, including running instances. The account must have the team's effective `user` role. No rebuild is needed. Custom SSH configurations and key-service outages can affect propagation; a guest can fall back to cached keys. Targeted registration checks team permissions and resource existence, then refreshes SSH connection details. Its `guest_login_verified:false` field distinguishes API registration from an actual guest login.

## Rental options

`hotaisle offers` combines all currently available VM and bare metal configurations for the selected team. Each row shows the complete GPU bundle (for example, `8x MI300X`), number of available bundles, CPU cores, RAM/disk in GiB, hourly USD price, and minimum reservation minutes. A missing price is shown as `not quoted`. The API does not expose a catalog of sold-out configurations, so the CLI does not invent those options.

Use `--kind vm` or `--kind bm` to query one resource type. `hotaisle --json offers` returns normalized `options` with prices in cents and memory/storage in bytes; the older `vm available` and `bm available` commands keep their existing response shapes. An eight-GPU row is one indivisible offered configuration, not eight independent rentals.

When no matching capacity is returned, `bm available` and `vm available` explain that no resources are currently available for the selected team. `--json` returns an empty array (`[]`), and `--agent` includes the team, zero count, and explanatory message in Markdown. An empty result is a successful query; authentication and API failures remain errors.

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
