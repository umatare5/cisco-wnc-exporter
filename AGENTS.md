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
- `internal/collector/` — AP, Client, WLAN and Controller `prometheus.Collector` implementations
- `internal/wnc/` — Thin interfaces (`APSource`, `ClientSource`, `WLANSource`, `RRMSource`, `ControllerSource`) and the adapters that implement them over the SDK client
- `internal/cache/` — Generic TTL cache, used only for `*_info` metrics (`--collector.info-cache-ttl`)
- `internal/log/` — `log/slog` setup; structured logging helpers
- `docs/` — One page per collector, the shared rules, and the health, enumeration and help references; [`CONTRIBUTING.md`](CONTRIBUTING.md) names the owner of each `docs/` page
- `examples/` — Prometheus job, alert rules and Grafana dashboards; tunable values live here

## Setup and Commands

Install required tools (one-time):

- `go install gotest.tools/gotestsum@latest`
- `golangci-lint` - See <https://golangci-lint.run/docs/welcome/install/local/>
- `goreleaser` release builds (see [`.goreleaser.yml`](.goreleaser.yml))
- `make pre-commit-install` wires `no-commit-to-main`, `golangci-lint`, `actionlint`, `gitleaks` and `markdownlint-cli2` (see [`.pre-commit-config.yaml`](.pre-commit-config.yaml))

Make targets ([`Makefile`](Makefile)):

- `make help` — List the targets below, then the four tools they require
- `make build` — Build binary into `tmp/cisco-wnc-exporter`
- `make lint` — `golangci-lint run` + `go mod tidy`
- `make test-unit` — Run unit tests via `gotestsum` with coverage
- `make test-unit-coverage` — Generate HTML report at `coverage/report.html`
- `make clean` — Remove build artifacts and `.bak*` files
- `make image` — Build Docker image (`$USER/cisco-wnc-exporter`)
- `make pre-commit-install` / `pre-commit-test` / `pre-commit-uninstall` — Manage the pre-commit hooks

The `markdownlint-cli2` hook runs with `--fix`, so a rewritten Markdown file has to be re-staged.

## Code Style

- Linting and formatting are enforced by `golangci-lint` in the pre-commit hook ([`.golangci.yml`](.golangci.yml)).
- Comments record only what the code cannot say, and never address the reader.
- A `--collector.<name>.<group>` flag switches a collector's group; `module` means a YANG module.

## Testing

- Run `make test-unit` before committing.
- Place tests next to code under test (`*_test.go`).
- Coverage threshold is enforced by [`.github/workflows/go-test-coverage.yml`](.github/workflows/go-test-coverage.yml).
- One snapshot backs every collector test — [`CONTRIBUTING.md`](CONTRIBUTING.md) says why a private one hides absence.

## Commits and PRs

- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore(deps):`, etc.).
- Sign off commits with `Signed-off-by:` (DCO).
- Open PRs against `main`. CI runs Format and Lint, Test and Build, Coverage, Prometheus Rules and CodeQL on every pull request, and gates markdownlint, Link Check, actionlint and govulncheck on the paths they read.

## Domain Knowledge

### Verifying Values

- **A YANG model is a design document, not the implementation.** Units, ranges, enum spellings and even the presence of a leaf can differ on a live controller, so confirm each against a RESTCONF response before relying on it.
- **Arbitrate configuration on the device with `show running-config all`.** It prints the negated form for a feature that is off, so a WLAN with no such line has it on.
- **Never ask an operational read for the values in force.** Materialising defaults there would defeat the absence guards, and [`CONTRIBUTING.md`](CONTRIBUTING.md) carries the RESTCONF request forms — two GETs and the one RPC this exporter never issues.
- **A claim in `docs/` is a measurement.** Cite the Go file and line, or the controller reading it came from, rather than restating what a sibling page says.

### Controller Behaviour

- **An omitted configuration leaf means its default is in force, not that nothing set it.** The default is often `true`, so decoding it as `false` inverts the reading — [Absence](docs/README.md#absence) carries the rule.
- **Omission is per leaf rather than per container.** A container can arrive with two of its four leaves present and the other two omitted at `true`, so a present sibling proves nothing.
- **A controller that rejects `with-defaults=report-all` answers `400`.** The read then falls back to a plain one and counts it — [Exporter Health](docs/health.md#specifications) carries the counter.
- **Operational routes carry no hidden defaults.** Every one this exporter reads was byte-identical plain and with `report-all`, which is why the parameter is a configuration-read tool alone.
- **The controller owns the numbering of every enumeration.** [Enumeration Values](docs/enums.md) records the YANG revision each was read at, because nothing else makes a renumbering detectable.
