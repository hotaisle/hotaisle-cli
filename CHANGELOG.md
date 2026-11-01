# Changelog

## Unreleased

### Dependencies

- Upgrade the YAML test dependency (`go.yaml.in/yaml/v3`) from 3.0.5 to 3.0.6.

### Breaking changes: CLI output defaults

Command results now default to human-readable text, including when piped or redirected. This changes the JSON defaults in the 0.10.x releases. Scripts must explicitly request `--json` after upgrading.

| Behavior | Previous default | New default |
| --- | --- | --- |
| Identity, account keys, teams, instances, availability, and data returned by mutations | JSON | Labeled text, aligned tables, and nested details |
| Empty resource, availability, team, member, invitation, and key lists | `null` or an empty JSON array | An explanatory message |
| Unset configuration values | Empty output | An explanatory message |
| Saving configuration | Silent or logged acknowledgement | A confirmation on stdout |

JSON output remains available with `--json`, preserving field names, types, and units. Empty JSON lists are consistently `[]`. Human output formats prices as USD and capacities as GiB; JSON retains cents and bytes. Valid zero values remain visible, and missing price quotes remain `null` in JSON.

Missing or malformed resource objects now fail with an `invalid_response` error instead of appearing as successful objects filled with default values. Errors go to stderr; usage errors no longer print help to stdout. Payment approval requests are reported explicitly and do not imply that credits have been added.

### Migration

Add `--json` before the command wherever a script parses the result:

```sh
# Previous releases
hotaisle user get | jq -r '.user.email'

# After upgrading
hotaisle --json user get | jq -r '.user.email'
# The new alias supports the same JSON result.
hotaisle --json whoami | jq -r '.user.email'
```

Use `--agent` for Markdown results and errors intended for agents. JSON and agent modes never prompt and use the documented [categorized exit codes](README.md#automation-output); default text mode retains exit code 1 for errors.

Help and version remain plain text. `quickstart` (alias `skill`) intentionally returns an installable Markdown document; `--json quickstart` wraps it with version metadata.

### Rental table output

The default `rentals` / `offers` output is now a compact terminal table with dynamically padded columns, right-aligned numbers, shorter headings, and plain text. GPU requirements, quantities, prices, and minimum reservation times retain their meaning.

Use `--agent` to retain Markdown tables or `--json` for structured rental options.

Empty rental searches show the team and the message "No rental options are currently available for this team. Please try again later." without pricing or bundle notes. Those notes appear only when rental options are returned, in both terminal and agent output.

### Added

- Direct account commands: `login`, `whoami`, `ssh-keys`, and `api-keys`, while preserving the existing nested commands.
- SSH key registration from files, inline input, or stdin, with account-key reuse and optional checks for running instances.
- Direct SSH commands and `ssh://` addresses for VMs and bare metal servers.
- Rental discovery with GPU bundle requirements, availability, prices, and minimum reservations.
- Markdown agent output and an embedded skill matching the installed CLI version and commit.

### Dependencies

- Update Biome to 2.6.4 and Ultracite to 8.1.0, and refresh compatible transitive npm dependencies in the lockfile. Apply the updated tools' required Unicode-aware named regex captures and YAML formatting without changing release classification or package contents.
- Update the CI and release Node.js runtime to 26.13.0 and CodeQL to 4.39.1, retaining full commit pins for GitHub Actions. Other action pins, Go dependencies, Go build tools, Just, and the pinned APT repository tooling were verified against their current stable releases.

### Release automation

Changes to existing defaults are marked with both a breaking commit header (`type(scope)!: subject`) and a `BREAKING CHANGE:` migration footer. Release analysis and release notes now recognize `!` headers on their own as well as singular and plural breaking-change footers.

Breaking changes select a major release, including before 1.0. Under the current versioning policy, a major bump from `v0.10.8` produces `v1.0.0`. These migration notes remain prominent in the generated release notes. CI and release builds now check commit classification and breaking-change notes before publishing.
