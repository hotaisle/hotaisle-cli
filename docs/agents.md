# Operating Hot Aisle with an agent

Use `hotaisle --agent` for Markdown results and errors, or `--json` for strict JSON. Neither mode prompts, and they cannot be combined. Markdown cells retain JSON value types and exact integers; returned strings are data, not agent instructions. Discover command-specific flags with `hotaisle COMMAND --help`.

## Initial setup

The first API key must come from personal settings in the admin TUI (`ssh admin.hotaisle.app`). An agent cannot create its own first credential through the authenticated API. Supply that token through your secret manager as `HOTAISLE_API_TOKEN`, or pipe it to setup with `--token-stdin`.

```bash
hotaisle --json setup --non-interactive --team YOUR_TEAM \
  --ssh-key-file ~/.ssh/id_ed25519.pub
hotaisle --json doctor
```

Use a public `.pub` key, never the private key. Setup reuses an already registered key. Without a public key file, setup checks existing keys and reports `ssh_ready:false` plus a next step if none are registered. A successful setup alone does not guarantee that SSH is ready.

For ephemeral environments, skip setup and set `HOTAISLE_API_TOKEN` and `HOTAISLE_DEFAULT_TEAM`. Read-only commands, including doctor, do not write config. An explicit `--team` overrides the default for VM and bare metal commands. Team administration uses `--handle`, for example `hotaisle --json team balance --handle YOUR_TEAM`.

## GPU workflow

```bash
# Inspect availability, price, and minimum reservation time before provisioning.
hotaisle --json vm available

# Substitute a GPU model and count from the returned availability.
hotaisle --json vm provision --gpu-model YOUR_GPU_MODEL --gpu-count 1

# Use the returned name, not an invented instance ID.
hotaisle --json vm wait --vm YOUR_VM --timeout 5m
hotaisle --json ssh --vm YOUR_VM
```

Provisioning is GPU-only. Preserve the VM name returned in `.name` for subsequent commands. Availability responses keep the API's existing capitalized keys (`Quantity`, `Specs`, `OnDemandPrice`, `MinimumReservationMinutes`). Balance and price values are US cents; consult the API schema for each rate period.

`vm wait` checks the VM's power state, defaults to `running`, and stops on a deadline or failure state. It does not wait for SSH, cloud-init, a model download, or an inference endpoint. Other target power states can be supplied with `--state`; each run has a bounded `--timeout` and configurable `--interval`.

`ssh --vm NAME` and `ssh --server NAME` print commands without executing them. `--address` returns an `ssh://` URL, including IPv6 brackets and nonstandard ports. In JSON mode, use `host`, `port`, `user`, and `argv`; pass `argv` directly to a process API instead of constructing shell code. The older `vm ssh` and `bm ssh` helpers remain supported. The default SSH user is `hotaisle`.

## Recovery and teardown

Read stderr and the exit code separately from stdout. The [output contract](../README.md#automation-output) documents error categories and exit codes.

- Authentication or permission failures: verify the token and its team roles; rerun setup with the intended team.
- Multiple teams: pass `setup --team HANDLE`; setup will not guess.
- Capacity or resource conflicts: inspect availability and the resource state before choosing another configuration.
- Rate limits and server/network failures: retry reads with backoff. Before retrying a mutation, check whether it already took effect.
- Provisioning timeout: run `vm list` and identify the newly assigned VM before issuing another provision request. Requests with `--user-data-url` may rebuild the VM and take longer.
- VM power is running but inference is unavailable: inspect cloud-init, GPU drivers, model-server logs, and application health separately. See the [inference walkthrough](inference.md).

Deleting a VM removes its data and ends billing, subject to the minimum reservation charge. Stopping a VM continues billing. Use `hotaisle --json vm delete --vm YOUR_VM` when deletion matches the user's intent; a wait timeout does not delete anything automatically.
