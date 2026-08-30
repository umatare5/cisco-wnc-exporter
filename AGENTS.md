# Repository Instructions

> [!IMPORTANT]
> Read [`README.md`](README.md) for project overview and architecture.

## Tech Stack

- Go 1.27+ (see [`go.mod`](go.mod))
- [`prometheus/client_golang`](https://github.com/prometheus/client_golang) v1.23+ — metric registration and HTTP handler
- [`umatare5/cisco-ios-xe-wireless-go`](https://github.com/umatare5/cisco-ios-xe-wireless-go) v0.11+ — sole RESTCONF SDK for Cisco C9800 WNC
- [`urfave/cli/v3`](https://github.com/urfave/cli) v3.8+ — CLI flags and application lifecycle
- [`goreleaser`](https://goreleaser.com/) v2 — cross-platform release builds (see [`.goreleaser.yml`](.goreleaser.yml))

## Repository Structure

- `cmd/` — Entry point (`main.go`); calls `internal/cli` for app setup
- `internal/cli/` — CLI flag definitions and app wiring (urfave/cli/v3)
- `internal/config/` — flag/env parsing, defaults (`0.0.0.0:10039`), and validation
- `internal/server/` — HTTP server (`/metrics`, `/healthz`, `/`), graceful shutdown
- `internal/collector/` — AP, Client, WLAN collectors; `prometheus.Collector` implementations
- `internal/wnc/` — Thin interfaces (`APSource`, `ClientSource`, `WLANSource`, `RRMSource`) and adapters over the WNC SDK
- `internal/cache/` — Generic TTL cache, used only for `*_info` metrics (`--collector.info-cache-ttl`)
- `internal/log/` — `log/slog` setup; structured logging helpers

## Setup and Commands

Install required tools (one-time):

- `go install gotest.tools/gotestsum@latest`
- `golangci-lint` - See <https://golangci-lint.run/docs/welcome/install/local/>
- `goreleaser` release builds (see [`.goreleaser.yml`](.goreleaser.yml))
- `make pre-commit-install` wires `no-commit-to-main`, `golangci-lint`, `actionlint`, `gitleaks` and `markdownlint-cli2` (see [`.pre-commit-config.yaml`](.pre-commit-config.yaml))

Make targets ([`Makefile`](Makefile)):

- `make build` — Build binary into `tmp/cisco-wnc-exporter`
- `make lint` — `golangci-lint run` + `go mod tidy`
- `make test-unit` — Run unit tests via `gotestsum` with coverage
- `make test-unit-coverage` — Generate HTML report at `coverage/report.html`
- `make clean` — Remove build artifacts and `.bak*` files
- `make image` — Build Docker image (`$USER/cisco-wnc-exporter`)
- `make pre-commit-install` / `pre-commit-test` / `pre-commit-uninstall` — Manage the pre-commit hooks

## Code Style

- Linting and formatting are enforced by `golangci-lint` in the pre-commit hook (see [`.golangci.yml`](.golangci.yml)).
- Comments record only what the code cannot say, and never address the reader.

## Testing Instructions

- Run `make test-unit` before committing.
- Place tests next to code under test (`*_test.go`).
- Coverage threshold is enforced by [`.github/workflows/go-test-coverage.yml`](.github/workflows/go-test-coverage.yml).

## Commits and PRs

- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore(deps):`, etc.).
- Sign off commits with `Signed-off-by:` (DCO).
- Open PRs against `main`. CI runs lint, tests, and CodeQL.

## Domain Knowledge

### Verifying Values

- **A YANG model is a design document, not the implementation.** Units, ranges, enum spellings, and even the presence of a leaf can differ on a live controller, so confirm every value against a RESTCONF response from a real WNC before relying on it.
- **A configuration leaf missing from a response means its default is in force, not that nothing set it.** The default is often `true`, so decoding an omitted boolean as `false` inverts the reading: on a plain read of `wlan-cfg-entries`, `wpa2-enabled` and `wlan-11k-neigh-list` are absent from exactly the WLANs where they are enabled, and present only where they were explicitly switched off.
- **Ask for the values in force, and expect omission to be per leaf rather than per container.** Appending `?with-defaults=report-all` returns the omitted leaves — a policy profile a plain read shows with a handful of them comes back with an order of magnitude more — and a `wlan-switching-policy` container can arrive with two of its four `central-*` leaves present and the other two omitted at `true`. A controller that rejects the parameter answers `400`, which is why a rejected read falls back to a plain one and counts that in `wnc_refresh_defaults_fallback_total`.
- **Arbitrate on the device with `show running-config all`, and only for configuration.** It prints the negated form for a feature that is off, so a WLAN with no such line has it on. Every operational route this exporter reads was byte-identical plain and with `report-all`, so do not add the parameter to an operational read: materialising defaults there would defeat the absence guards that keep a fabricated zero out of the metrics.

### RESTCONF Access Patterns

GET a collection:

```bash
curl -k -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
        -H "Accept: application/yang-data+json" \
        "https://$WNC_CONTROLLER/restconf/data/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data"
```

GET a single entry by list key (MAC address):

```bash
curl -k -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
        -H "Accept: application/yang-data+json" \
        "https://$WNC_CONTROLLER/restconf/data/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data=00:11:22:33:44:55"
```

POST an RPC operation (`/restconf/operations/`):

```bash
curl -k -X POST \
        -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
        -H "Content-Type: application/yang-data+json" \
        -d '{"input": {"ap-name": "TEST-AP01"}}' \
        "https://$WNC_CONTROLLER/restconf/operations/Cisco-IOS-XE-wireless-access-point-cmd-rpc:ap-reset"
```
