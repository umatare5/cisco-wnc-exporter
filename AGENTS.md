# Repository Instructions

> [!IMPORTANT]
> Read [`README.md`](README.md) for project overview.

## Tech Stack

- Go 1.27+ (see [`go.mod`](go.mod))
- [`prometheus/client_golang`](https://github.com/prometheus/client_golang) v1.24+ – metric registration and HTTP handler
- [`umatare5/cisco-ios-xe-wireless-go`](https://github.com/umatare5/cisco-ios-xe-wireless-go) v0.11+ – sole RESTCONF SDK for Cisco C9800 WNC
- [`urfave/cli/v3`](https://github.com/urfave/cli) v3.11+ – CLI flags and application lifecycle
- [`goreleaser`](https://goreleaser.com/) v2 – cross-platform release builds (see [`.goreleaser.yml`](.goreleaser.yml))

## Repository Structure

Read from `cmd/main.go`. Each package is named for what it owns.

- [`cmd/main.go`](cmd/main.go) – application entry point
- [`internal/cli/`](internal/cli) – command-line flags, defaults and app wiring
- [`internal/config/`](internal/config) – flag reads and RESTCONF validation
- [`internal/collector/`](internal/collector) – metric descriptions and collection logic
- [`internal/cache/`](internal/cache) – cache abstraction
- [`internal/wnc/`](internal/wnc) – upstream API client and data refresh
- [`internal/server/`](internal/server) – HTTP server configuration and routing
- [`internal/log/`](internal/log) – logger setup
- [`docs/`](docs) – reference pages behind the README
- [`scripts/`](scripts) – helper scripts the pre-commit hooks run
- [`examples/`](examples) – Prometheus configuration, alert rules and the Grafana dashboard

## Setup and Commands

Run `make pre-commit-install` first.

- Read [`Makefile`](Makefile) which lists all available make targets and their descriptions.
- Read [`CONTRIBUTING.md`](CONTRIBUTING.md) which provides guidelines for contributing to the project.

## Code Style

Follow [Effective Go](https://go.dev/doc/effective_go) conventions and the software development principles DRY/YAGNI/SRP.

- Keep code simple and readable, avoiding clever tricks that obscure intent.
- Keep minimal for all changes, coding, testing, commenting, and documentation.
- Write simple comments that explain the reasoning behind the code, not just what it does.

## Testing

Follow [`CONTRIBUTING.md`](CONTRIBUTING.md).

- Run `make lint` and `make test-unit` before creating a commit.
- One snapshot backs every collector test – see [`CONTRIBUTING.md`](CONTRIBUTING.md) for why a private one hides absence.

## Commits and PRs

Follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore(deps):`, etc.).

- Run pre-commit and ensure all hooks pass before committing.
- Must Sign off all commits with `Signed-off-by:` (DCO).
- Open PRs against `main`. Create Draft PR as default.

## Domain Knowledge

Learn the constraints related to external entities, such as network devices, RESTCONF endpoints, and YANG models.

### About the Controller

How the Catalyst 9800 controller behaves and interacts with the collectors. See also [Architecture](docs/architecture.md).

- **An omitted configuration leaf means its default is in force, not that nothing set it.** The default is often `true`, so decoding it as `false` inverts the reading – [Absence](docs/architecture.md#absence) carries the rule.
- **Omission is per leaf rather than per container.** A container can arrive with two of its four leaves present and the other two omitted at `true`, so a present sibling proves nothing.
- **A controller that rejects `with-defaults=report-all` answers `400`.** The read then falls back to a plain one and counts it – [Exporter Health](docs/health.md#technical-notes) carries the counter.
- **Operational routes carry no hidden defaults.** Every one this exporter reads was byte-identical plain and with `report-all`, which is why the parameter is a configuration-read tool alone.
- **The controller owns the numbering of every enumeration.** [Enumeration Values](docs/enums.md) records the YANG revision each was read at, because nothing else makes a renumbering detectable.

### About the YANG Models

How the YANG models are interpreted and used in this project.

- **A YANG model is a design document, not the implementation.** Units, ranges, enum spellings and even the presence of a leaf can differ on a live controller, so confirm each against a RESTCONF response before relying on it.
- **Arbitrate configuration on the device with `show running-config all`.** It prints the negated form for a feature that is off, so a WLAN with no such line has it on.
- **Never ask an operational read for the values in force.** Materialising defaults there would defeat the absence guards, and [`CONTRIBUTING.md`](CONTRIBUTING.md) carries the RESTCONF request forms.
- **A claim in `docs/` is a measurement.** Cite the Go file and line, or the controller reading it came from, rather than restating what a sibling page says.
