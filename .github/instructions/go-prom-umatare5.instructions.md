---
description: "Go Library Development Instructions"
applyTo: "**/*.go,**/go.mod,**/go.sum"
---

# Go Prometheus Exporter Development Instructions

GitHub Copilot **MUST** follow these instructions when generating or modifying **Go code** in this repository.

## Scope & Metadata

- **Last Updated**: 2025-09-20
- **Precedence**: 1. `copilot-instructions.md` (Global) → 2. `go.instructions.md` (Community) → 3. `go-prom-umatare5.instructions.md` (This)
- **Compatibility**: Go **1.25+** cross-platform
- **Style Base**: [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- **Goal**: Build a **reliable, operator-friendly Prometheus exporter** for Cisco WNC: **low-cardinality, stable metrics**, minimal public surface, idiomatic Go, and robust error handling.
- **Expert Persona**: Operate as a **Go 1.25+ expert** and a **Prometheus exporter author** (naming, types, labels, `/metrics` HTTP, graceful shutdown).

---

## 1. Architecture

- **AR-001 (MUST)** Keep a clear package layout:

  - `cmd/cisco-wnc-exporter/` – main entrypoint, flag/env wiring only.
  - `internal/config/` – flag/env parsing, defaults, validation.
  - `internal/server/` – HTTP server (`/metrics`, health), graceful shutdown.
  - `internal/collector/{ap,wlan,client}/` – collectors per domain.
  - `internal/wnc/` – thin interfaces/adapters to the WNC SDK.
  - `internal/log/` – slog setup; structured logging helpers.
  - `pkg/testutil/` – public test utilities (mock/fakes, `CollectAndCompare`).

- **AR-002 (MUST)** Keep collectors **thin**. Move transport/model mapping into `internal/wnc` or per‑collector fetch helpers.
- **AR-003 (MUST)** Use a **custom `prometheus.Registry`** in the server; register exporter collectors explicitly. Register `process`/`go` collectors **behind flags** (opt‑in) when appropriate.
- **AR-004 (SHOULD)** Provide a **build info** gauge (`wnc_build_info{version,revision}` = 1).

---

## 2. Collector Pattern

- **CL-001 (MUST)** Implement the `prometheus.Collector` interface with **value semantics** where possible.
- **CL-002 (MUST)** Avoid heavy I/O inside `Collect`. Prefer a **prefetch** step with cache + TTL; serve cached values within the scrape budget.
- **CL-003 (MUST)** Make collectors **concurrency‑safe** (mutexes or lock‑free snapshots). Do not mutate shared descriptors.
- **CL-004 (MUST)** Keep label sets **minimal and bounded**. Provide label selection via flags (e.g., AP label = MAC/IP/Hostname).
- **CL-005 (SHOULD)** Use **predeclared descriptors** and **typed const labels** when appropriate. Avoid dynamic descriptor churn.
- **CL-006 (SHOULD)** Provide **`New...Collector(cfg)` constructors** returning concrete types; attach `With...` options if needed.

**Skeleton**

```go
// internal/collector/ap/ap.go
package ap

import (
    "sync"
    "github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
    mu   sync.RWMutex
    desc *prometheus.Desc
    src  Source // interface backed by WNC SDK adapter
}

func NewCollector(src Source) *Collector {
    return &Collector{
        desc: prometheus.NewDesc(
            "wnc_ap_count", "Number of access points", nil, nil,
        ),
        src: src,
    }
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) { ch <- c.desc }

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
    // Consider: prefetch outside Collect; here we keep it simple
    n, err := c.src.CountAPs()
    if err == nil {
        ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(n))
    }
}
```

---

## 3. Metrics Naming & Labels

- **MN-001 (MUST)** Prefix names with `wnc_`. Use `snake_case` and **unit suffixes** (`_bytes`, `_seconds`, `_ratio`, `_percent`, `_count`).
- **MN-002 (MUST)** **Counters** for monotonic totals; **Gauges** for instantaneous values; **Histograms/Summaries** only with strong justification.
- **MN-003 (MUST)** Labels must be **bounded** and **documented**. For AP/WLAN/Client collectors, expose **one** selector label as configured by flags (e.g., AP = MAC/IP/Hostname; WLAN = ID/Name; Client = MAC/IP/Hostname/Username).
- **MN-004 (SHOULD)** Prefer IDs over names. Avoid free‑form text and high‑cardinality labels.
- **MN-005 (MUST)** Keep help strings concise and operator‑friendly; include unit context.

### 3.A Label Naming Conventions (Prometheus Community Standards)

- **LN-001 (MUST)** Use **simplified label names** following Prometheus community conventions. Avoid redundant suffixes when the metric context makes the label meaning clear:

  - ✅ **Good**: `wnc_client_rssi_dbm{mac="xx:yy:zz", ap="ap-name", radio="0", wlan="guest"}`
  - ❌ **Avoid**: `wnc_client_rssi_dbm{client_mac="xx:yy:zz", ap_name="ap-name", radio_id="0", wlan_name="guest"}`

- **LN-002 (MUST)** **Context-based label simplification rules**:

  - **Client Metrics**: Use `mac` (not `client_mac`), `ap` (not `ap_name`), `radio` (not `radio_id`), `wlan` (not `wlan_name`), `name` (client hostname from DcInfo.DeviceName), `username` (EAP authentication username)
  - **AP Metrics**: Use `mac` (not `ap_mac`), `radio` (not `radio_id`), `wlan` (not `wlan_name`)
  - **WLAN Metrics**: Use `id` (not `wlan_id`) for base labels, `name` (not `wlan_name`) for additional labels

- **LN-003 (MUST)** **Rationale**: Metric prefixes (`wnc_client_`, `wnc_ap_`, etc.) already establish context, making verbose label suffixes redundant. This follows the same pattern used by node_exporter (`device`, `interface`), blackbox_exporter (`module`, `target`), and kube-state-metrics (`namespace`, `pod`, `container`).

- **LN-004 (SHOULD)** **Conflict resolution**: Only add descriptive suffixes when actual label conflicts occur within the same metric. Use kube-state-metrics' `_conflict1`, `_conflict2` pattern if automatic resolution is needed.

- **LN-005 (MUST)** **Documentation consistency**: Update all README.md tables, help text, and examples to reflect the simplified naming scheme.

### 3.A Label Naming Conventions (Prometheus Community Standards)

- **LN-001 (MUST)** Use **simplified label names** following Prometheus community conventions. Avoid redundant suffixes when the metric context makes the label meaning clear:

  - ✅ **Good**: `wnc_client_rssi_dbm{mac="xx:yy:zz", ap="ap-name", radio="0", wlan="guest"}`
  - ❌ **Avoid**: `wnc_client_rssi_dbm{client_mac="xx:yy:zz", ap_name="ap-name", radio_id="0", wlan_name="guest"}`

- **LN-002 (MUST)** **Context-based label simplification rules**:

  - **Client Metrics**: Use `mac` (not `client_mac`), `ap` (not `ap_name`), `radio` (not `radio_id`), `wlan` (not `wlan_name`), `name` (client hostname from DcInfo.DeviceName), `username` (EAP authentication username)
  - **AP Metrics**: Use `mac` (not `ap_mac`), `radio` (not `radio_id`), `wlan` (not `wlan_name`)
  - **WLAN Metrics**: Use `id` (not `wlan_id`) for base labels, `name` (not `wlan_name`) for additional labels

- **LN-003 (MUST)** **Rationale**: Metric prefixes (`wnc_client_`, `wnc_ap_`, etc.) already establish context, making verbose label suffixes redundant. This follows the same pattern used by node_exporter (`device`, `interface`), blackbox_exporter (`module`, `target`), and kube-state-metrics (`namespace`, `pod`, `container`).

- **LN-004 (SHOULD)** **Conflict resolution**: Only add descriptive suffixes when actual label conflicts occur within the same metric. Use kube-state-metrics' `_conflict1`, `_conflict2` pattern if automatic resolution is needed.

- **LN-005 (MUST)** **Documentation consistency**: Update all README.md tables, help text, and examples to reflect the simplified naming scheme.

---

## 4. Config & Flags

- **CF-001 (MUST)** Define defaults as **named constants**; mirror to env vars with clear precedence (env → flag default or vice versa; document).
- **CF-002 (MUST)** Validate flag combinations; fail fast with actionable errors.
- **CF-003 (SHOULD)** Group flags logically (server, logging, WNC auth, collectors). Provide `--dry-run` to validate config without starting.

---

## 5. HTTP Server

- **HS-001 (MUST)** Use `promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{EnableOpenMetrics: true, MaxRequestsInFlight: N})`.
- **HS-002 (MUST)** Implement **graceful shutdown** (context + timeout) and log start/stop with bind address.
- **HS-003 (SHOULD)** Provide `/healthz` and `/readyz` lightweight endpoints.
- **HS-004 (SHOULD)** Optionally add `X-Prometheus-Scrape-Timeout-Seconds` awareness to align caches.

---

## 6. Context, Transport & SDK

- **CT-001 (MUST)** Thread `context.Context` through all outbound calls; respect deadlines and cancellation.
- **CT-002 (MUST)** Reuse a single `http.Client` with timeouts; **always** close `resp.Body`.
- **CT-003 (MUST)** Use the **WNC SDK** for all controller access behind tiny interfaces. **No speculative parsing**; base mapping on verified fields.
- **CT-004 (SHOULD)** Normalize identifiers (e.g., MAC format) before requests.

---

## 7. Errors & Logging

- **EL-001 (MUST)** Wrap errors with `%w` and include identifiers (never secrets). Provide typed/sentinel errors where useful.
- **EL-002 (MUST)** Use `log/slog` with structured key/value fields. Keep messages concise.
- **EL-003 (SHOULD)** Classify transient vs. permanent errors; in collectors, **emit nothing** on transient errors instead of exporting partial/wrong values. Consider an `wnc_scrape_errors_total` counter per collector.

---

## 8. Testing Framework

### 8.A Unit Tests

- **TU-001 (MUST)** Use table‑driven tests and `t.Run`. Keep tests deterministic.
- **TU-002 (MUST)** For metrics surface, use `prometheus/testutil.CollectAndCompare` with **golden files** under `testdata/`.
- **TU-003 (SHOULD)** Fake the WNC SDK via interfaces; return realistic payloads. Place shared fakes/mocks under `pkg/testutil`.

### 8.B Integration Tests

- **TI-001 (SHOULD)** Guard with env vars (e.g., `WNC_HOST`, `WNC_TOKEN`). Skip if missing.
- **TI-002 (SHOULD)** Validate that `/metrics` exposes expected families and no unexpected label cardinality.

### 8.C Test Server Management

- **TS-001 (MUST)** **Always start test servers in background** using `&` operator and capture PID for proper cleanup:

  ```bash
  # ✅ Correct: Background start with PID capture
  ./cisco-wnc-exporter --web.listen-port 10050 > /dev/null 2>&1 &
  SERVER_PID=$!

  # Later: Targeted cleanup
  kill $SERVER_PID
  ```

- **TS-002 (MUST)** **Never use `pkill cisco-wnc-exporter`** for test cleanup as it may kill other processes. Always use specific PID termination.

- **TS-003 (MUST)** **Always wait for server readiness** before testing:

  ```bash
  sleep 3  # Allow server startup time
  curl -s http://localhost:10050/metrics | head -5
  ```

- **TS-004 (MUST)** **Use unique ports** (10050+) for test servers to avoid conflicts with development instances.

### 8.D Coverage & CI

- **TC-001 (MUST)** Target **≥ 85%** unit coverage overall (**≥ 90%** in collectors). Keep CI green.

**Collector Test Skeleton**

```go
func TestAPCollector_Collect_MetricsSurface(t *testing.T) {
    reg := prometheus.NewRegistry()
    fake := testutil.NewFakeWNCApi().WithAPCount(42)
    c := ap.NewCollector(fake)
    reg.MustRegister(c)

    // Compare with golden text-format metrics
    expected := filepath.Join("testdata", "ap_metrics.txt")
    if err := testutil.CollectAndCompare(c, os.ReadFile(expected)); err != nil {
        t.Fatal(err)
    }
}
```

---

## 9. Performance & Memory

- **PM-001 (MUST)** Avoid per‑scrape allocations (prebuild descriptors/const labels, reuse buffers).
- **PM-002 (MUST)** Avoid unbounded maps keyed by user input or names.
- **PM-003 (SHOULD)** Consider a **poller + cache** when backend latency exceeds typical scrape timeouts.

---

## 10. Naming & API Conventions

- **NA-001 (MUST)** Use **imperative verbs** for actions (`Run`, `Start`, `Shutdown`).
- **NA-002 (MUST)** Keep package names singular and stutter‑free (`collector/ap`, not `collectors` or `apcollector`).
- **NA-003 (SHOULD)** Export only what the `cmd/` needs; keep internals unexported.
- **NA-004 (SHOULD)** **Minimize function scope**. Functions should be defined in the most specific scope where they are used:
  - Functions used by a single file → private functions within that file
  - Functions shared by multiple files within a package → package-level unexported functions
  - Functions shared across packages → exported functions in the most specific common package
  - Avoid global utility functions that create unnecessary coupling between packages

---

## 11. Documentation

- **DC-001 (MUST)** Start each exported type/function comment with its identifier.
- **DC-002 (MUST)** Document each metric family in code (help text) and in `README.md` table (name, type, labels).
- **DC-003 (SHOULD)** Provide minimal code examples in package comments.

---

## 12. Quality Gate

- **QG-001 (MUST)** CI must run: `go vet`, `staticcheck`, `golangci-lint`, `go test ./...`, and (if configured) `promtool` checks for examples.
- **QG-002 (MUST)** `go.mod` is tidy; pinned minimal dependencies (prefer stdlib + `client_golang`).
- **QG-003 (MUST)** Zero lint violations; no flaky tests; deterministic metric order in tests (use `testutil.ToFloat64` or golden with stable ordering).

---

## 13. Example: Server Wiring

```go
// internal/server/server.go
func New(reg *prometheus.Registry, addr string) *http.Server {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
    mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
    return &http.Server{Addr: addr, Handler: mux}
}
```

```go
// cmd/cisco-wnc-exporter/main.go (excerpt)
func run(ctx context.Context, cfg config.Config) error {
    reg := prometheus.NewRegistry()
    if cfg.Metrics.EnableGoCollector { reg.MustRegister(collectors.NewGoCollector()) }
    if cfg.Metrics.EnableProcessCollector { reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})) }

    apc := ap.NewCollector(wnc.NewAPSource(cfg.WNC))
    wlan := wlan.NewCollector(wnc.NewWLANSource(cfg.WNC))
    cl := client.NewCollector(wnc.NewClientSource(cfg.WNC))

    if cfg.Collectors.APEnabled { reg.MustRegister(apc) }
    if cfg.Collectors.WLANEnabled { reg.MustRegister(wlan) }
    if cfg.Collectors.ClientEnabled { reg.MustRegister(cl) }

    srv := server.New(reg, net.JoinHostPort(cfg.Web.ListenAddress, strconv.Itoa(cfg.Web.ListenPort)))
    // graceful shutdown
    go func() { <-ctx.Done(); shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second); defer cancel(); _ = srv.Shutdown(shutdownCtx) }()
    return srv.ListenAndServe()
}
```
