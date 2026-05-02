# Repository Instructions

> [!IMPORTANT]
> Read [README.md](README.md) for project overview and architecture.

## Tech Stack

- Go 1.25+ (see [go.mod](go.mod))
- [`prometheus/client_golang`](https://github.com/prometheus/client_golang) v1.23+ — metric registration and HTTP handler
- [`umatare5/cisco-ios-xe-wireless-go`](https://github.com/umatare5/cisco-ios-xe-wireless-go) v0.4+ — sole RESTCONF SDK for Cisco C9800 WNC
- [`urfave/cli/v3`](https://github.com/urfave/cli) v3.8+ — CLI flags and application lifecycle
- [`goreleaser`](https://goreleaser.com/) v2 — cross-platform release builds (see [.goreleaser.yml](.goreleaser.yml))

## Repository Structure

- `cmd/` — Entry point (`main.go`); calls `internal/cli` for app setup
- `internal/cli/` — CLI flag definitions and app wiring (urfave/cli/v3)
- `internal/config/` — flag/env parsing, defaults (`0.0.0.0:10039`), and validation
- `internal/server/` — HTTP server (`/metrics`, `/healthz`, `/`), graceful shutdown
- `internal/collector/` — AP, Client, WLAN collectors; `prometheus.Collector` implementations
- `internal/wnc/` — Thin interfaces (`APSource`, `ClientSource`, `WLANSource`, `RRMSource`) and adapters over the WNC SDK
- `internal/cache/` — Shared API response cache with configurable TTL
- `internal/log/` — `log/slog` setup; structured logging helpers

## Setup and Commands

Install required tools (one-time):

- `go install gotest.tools/gotestsum@latest`
- `golangci-lint` - See <https://golangci-lint.run/docs/welcome/install/local//>
- `goreleaser` release builds (see [.goreleaser.yml](.goreleaser.yml))
- `pre-commit install` wires `golangci-lint`, `gofmt`, `markdownlint-cli2`, `gitleaks` (see [.pre-commit-config.yaml](.pre-commit-config.yaml))

Make targets ([Makefile](Makefile)):

- `make build` — Build binary into `tmp/cisco-wnc-exporter`
- `make lint` — `golangci-lint run` + `go mod tidy`
- `make test-unit` — Run unit tests via `gotestsum` with coverage
- `make test-unit-coverage` — Generate HTML report at `coverage/report.html`
- `make clean` — Remove build artifacts and `.bak*` files
- `make image` — Build Docker image (`$USER/cisco-wnc-exporter`)

## Code Style

- `gofmt` and `golangci-lint` are enforced by the pre-commit hook (see [.pre-commit-config.yaml](.pre-commit-config.yaml)).
- Follow [.github/instructions/go.instructions.md](.github/instructions/go.instructions.md) for idiomatic Go.
- Follow [.github/instructions/go-prom-umatare5.instructions.md](.github/instructions/go-prom-umatare5.instructions.md) for the Prometheus Exporter.

## Testing Instructions

- Run `make test-unit` before committing.
- Place tests next to code under test (`*_test.go`).
- Coverage threshold is enforced by [.github/workflows/go-test-coverage.yml](.github/workflows/go-test-coverage.yml).

## Commits and PRs

- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore(deps):`, etc.).
- Sign off commits with `Signed-off-by:` (DCO).
- Open PRs against `main`. CI runs lint, tests, and CodeQL.

## Domain Knowledge

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

## References

- [.github/instructions/go.instructions.md](.github/instructions/go.instructions.md) — Idiomatic Go practices
- [.github/instructions/go-prom-umatare5.instructions.md](.github/instructions/go-prom-umatare5.instructions.md) — Prometheus Exporter patterns
- [.github/instructions/markdown.instructions.md](.github/instructions/markdown.instructions.md) - CommonMark (0.31.2)
- [.github/instructions/markdown-gfm.instructions.md](.github/instructions/markdown-gfm.instructions.md) - GitHub-flavored Markdown
- [.github/instructions/github-actions-ci-cd-best-practices.instructions.md](.github/instructions/github-actions-ci-cd-best-practices.instructions.md) - GitHub Actions CI/CD best practices
