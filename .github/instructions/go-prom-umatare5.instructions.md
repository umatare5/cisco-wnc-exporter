---
description: "Go Library Development Instructions"
applyTo: "**/*.go,**/go.mod,**/go.sum"
---

# Go Prometheus Exporter Development Instructions

GitHub Copilot **MUST** follow these instructions when generating or modifying **Go code** in this repository.

## Scope & Metadata

- **Last Updated**: 2025-09-28
- **Precedence**: 1. `copilot-instructions.md` (Global) → 2. `go.instructions.md` (Community) → 3. `go-prom-umatare5.instructions.md` (This)
- **Compatibility**: Go **1.25+** cross-platform
- **Style Base**: [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- **Goal**: Build a **reliable, operator-friendly Prometheus exporter** for Cisco WNC: **low-cardinality, stable metrics**, minimal public surface, idiomatic Go, robust error handling, and **high-performance optimized collectors**.
- **Expert Persona**: Operate as a **Go 1.25+ expert**, **Prometheus exporter author** (naming, types, labels, `/metrics` HTTP, graceful shutdown), and **performance optimization specialist**.
- **Current Design**: **New collector module architecture** with General/Radio/Traffic/Errors/Info structure (2025-09-28 redesign).
- **Optimization Status**: Client Collector fully optimized (18 patterns), AP Collector fully optimized (18 patterns), WLAN Collector pending.

---

## 1. Architecture

- **AR-001 (MUST)** Keep a clear package layout:

  - `cmd/` – main entrypoint, calls `internal/cli` for CLI application setup.
  - `internal/cli/` – CLI implementation with flag definitions and application wiring.
  - `internal/config/` – flag/env parsing, defaults, validation.
  - `internal/server/` – HTTP server (`/metrics`, health), graceful shutdown.
  - `internal/collector/` – collectors for all domains (ap, wlan, client).
  - `internal/wnc/` – thin interfaces/adapters to the WNC SDK.
  - `internal/log/` – slog setup; structured logging helpers.
  - `pkg/testutil/` – public test utilities (mock/fakes, `CollectAndCompare`).

- **AR-002 (MUST)** Keep collectors **thin**. Move transport/model mapping into `internal/wnc` or per‑collector fetch helpers. Apply **18-pattern optimization framework** for performance.
- **AR-003 (MUST)** Use a **custom `prometheus.Registry`** in the server; register exporter collectors explicitly. Register `process`/`go` collectors **behind flags** (opt‑in) when appropriate.
- **AR-004 (SHOULD)** Provide a **build info** gauge (`wnc_build_info{version,revision}` = 1).
- **AR-005 (MUST)** **Follow optimized collector architecture**: Single-loop processing O(n), map-based lookups O(1), Slice of Structs patterns, and per-entity collection methods.

---

## 2. Collector Pattern

- **CL-001 (MUST)** Implement the `prometheus.Collector` interface with **value semantics** where possible.
- **CL-002 (MUST)** Avoid heavy I/O inside `Collect`. Prefer a **prefetch** step with cache + TTL; serve cached values within the scrape budget.
- **CL-003 (MUST)** Make collectors **concurrency‑safe** (mutexes or lock‑free snapshots). Do not mutate shared descriptors.
- **CL-004 (MUST)** Keep label sets **minimal and bounded**. Provide label selection via flags (e.g., AP label = MAC/IP/Hostname).
- **CL-005 (SHOULD)** Use **predeclared descriptors** and **typed const labels** when appropriate. Avoid dynamic descriptor churn.
- **CL-006 (SHOULD)** Provide **`New...Collector(cfg)` constructors** returning concrete types; attach `With...` options if needed.
- **CL-007 (MUST)** **Apply 18-pattern optimization framework**: Loop consolidation, map-based lookups, Slice of Structs, determine functions, and efficient single-loop processing.
- **CL-008 (MUST)** **Use optimized collection architecture**: Build lookup maps once in Collect(), pass to per-entity collection methods, avoid duplicate processing.

**Optimized Skeleton**

```go
// internal/collector/ap.go (optimized pattern)
package collector

import (
    "context"
    "strconv"
    "github.com/prometheus/client_golang/prometheus"
    ap "github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
)

// Label constants to avoid repetition
const (
    labelMAC  = "mac"
    labelName = "name"
    labelBand = "band"
)

type APCollector struct {
    // Descriptors organized by module
    infoDesc       *prometheus.Desc
    channelDesc    *prometheus.Desc
    // ... other descriptors

    src     wnc.APSource
    metrics ModuleConfig
}

func NewAPCollector(src wnc.APSource, metrics ModuleConfig) *APCollector {
    return &APCollector{
        infoDesc: prometheus.NewDesc(
            "wnc_ap_info", "AP information", []string{"mac", "name"}, nil,
        ),
        // ... other descriptors
        src:     src,
        metrics: metrics,
    }
}

func (c *APCollector) Describe(ch chan<- *prometheus.Desc) {
    // Unified describe - no separate describe* functions
    if c.metrics.General {
        ch <- c.infoDesc
        // ... other general descriptors
    }
    if c.metrics.Radio {
        ch <- c.channelDesc
        // ... other radio descriptors
    }
}

func (c *APCollector) Collect(ch chan<- prometheus.Metric) {
    // Single data fetch
    radioData, err := c.src.GetRadioData(context.Background())
    if err != nil {
        return
    }

    // Build lookup maps once (O(1) access)
    capwapMap := buildCAPWAPMap(capwapData)
    radioStatsMap := buildRadioOperStatsMap(radioStatsData)

    // Single loop optimization (O(n) not O(n×modules))
    for _, radio := range radioData {
        baseLabels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}

        // Per-entity collection methods
        if c.metrics.General {
            c.collectGeneralMetricsForRadio(ch, &radio, baseLabels, capwapMap)
        }
        if c.metrics.Radio {
            c.collectRadioMetricsForRadio(ch, &radio, baseLabels, radioStatsMap)
        }
    }
}

// Per-entity collection methods
func (c *APCollector) collectGeneralMetricsForRadio(
    ch chan<- prometheus.Metric,
    radio *ap.RadioOperData,
    labels []string,
    capwapMap map[string]ap.CAPWAPData,
) {
    // Optimized metric collection using Slice of Structs pattern
    metrics := []struct {
        desc  *prometheus.Desc
        value float64
    }{
        {c.channelDesc, float64(radio.PhyHtCfg.CfgData.CurrFreq)},
        // ... other metrics
    }

    for _, metric := range metrics {
        ch <- prometheus.MustNewConstMetric(
            metric.desc, prometheus.GaugeValue, metric.value, labels...,
        )
    }
}

// Helper functions for map building
func buildCAPWAPMap(data []ap.CAPWAPData) map[string]ap.CAPWAPData {
    result := make(map[string]ap.CAPWAPData)
    for _, item := range data {
        result[item.WtpMAC] = item
    }
    return result
}

// Example: Current optimized pattern from ClientCollector
func (c *ClientCollector) collectRadioMetrics(
    ch chan<- prometheus.Metric,
    commonData client.CommonOperData,
    traffic client.TrafficStats,
    dot11 client.Dot11OperData,
) {
    baseLabels := []string{commonData.ClientMAC}

    // Slice of Structs pattern (optimized metric registration)
    metrics := []struct {
        desc  *prometheus.Desc
        value float64
    }{
        {c.protocolDesc, float64(MapWirelessProtocol(dot11.EwlcMsPhyType, dot11.RadioType, dot11.Is11GClient))},
        {c.speedDesc, float64(traffic.Speed)},
        {c.spatialStreamsDesc, float64(traffic.SpatialStream)},
        {c.mcsIndexDesc, float64(parseMCSIndex(traffic.CurrentRate))},
        {c.rssiDesc, float64(traffic.MostRecentRSSI)},
        {c.snrDesc, float64(traffic.MostRecentSNR)},
    }

    for _, metric := range metrics {
        ch <- prometheus.MustNewConstMetric(
            metric.desc, prometheus.GaugeValue, metric.value, baseLabels...,
        )
    }
}
```

---

## 3. Data Validation Architecture

- **DV-001 (MUST)** **Data validation is the exclusive responsibility of the WNC layer (`internal/wnc`)**. The collector layer (`internal/collector`) must focus solely on business logic (metrics transformation and exposition).

- **DV-002 (MUST)** **Eliminate all type checks, nil checks, and data structure validations from collectors**. Examples of prohibited patterns in collectors:

  ```go
  // ❌ PROHIBITED in collectors:
  if measurement.Load == nil { return }
  if measurement.Noise != nil && len(measurement.Noise.Noise.NoiseData) > 0 { ... }
  if apOper.ApSysStats == nil { continue }
  if len(radio.RadioBandInfo) == 0 { return }
  ```

- **DV-003 (MUST)** **WNC layer must ensure data completeness and validity**. Return sanitized, complete data structures or empty safe defaults. Never return nil pointers that would require defensive programming in collectors.

- **DV-004 (MUST)** **Use consistent data validation patterns in WNC layer**:

  ```go
  // ✅ CORRECT pattern in internal/wnc:
  func (s *rrmSource) GetValidatedRRMMeasurement(ctx context.Context) (*rrm.RRMOperRRMMeasurement, error) {
      data, err := s.sharedDataSource.GetCachedData(ctx)
      if err != nil {
          return nil, err
      }

      // Validate and sanitize data structure
      validated := validateRRMData(data.RRMData)
      return validated, nil
  }
  ```

- **DV-005 (SHOULD)** **Create validation helper functions** for complex data structures to ensure consistency across the WNC layer.

- **DV-006 (MUST)** **Collectors assume valid data**. If WNC layer returns data, collectors must be able to use it directly without additional checks.

- **DV-007 (SHOULD)** **Prepare for SDK omitempty removal**. Fields marked with "Live: IOS-XE 17.12.5" in the SDK will have `omitempty` tags removed in future versions. Current validation should be minimal and include TODO comments for future simplification.

---

## 4. Metrics Naming & Labels

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

## 5. Config & Flags

- **CF-001 (MUST)** Define defaults as **named constants**; mirror to env vars with clear precedence (env → flag default or vice versa; document).
- **CF-002 (MUST)** Validate flag combinations; fail fast with actionable errors.
- **CF-003 (SHOULD)** Group flags logically (server, logging, WNC auth, collectors). Provide `--dry-run` to validate config without starting.

---

## 6. HTTP Server

- **HS-001 (MUST)** Use `promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{EnableOpenMetrics: true, MaxRequestsInFlight: N})`.
- **HS-002 (MUST)** Implement **graceful shutdown** (context + timeout) and log start/stop with bind address.
- **HS-003 (SHOULD)** Provide `/healthz` and `/readyz` lightweight endpoints.
- **HS-004 (SHOULD)** Optionally add `X-Prometheus-Scrape-Timeout-Seconds` awareness to align caches.

---

## 7. Context, Transport & SDK

- **CT-001 (MUST)** Thread `context.Context` through all outbound calls; respect deadlines and cancellation.
- **CT-002 (MUST)** Reuse a single `http.Client` with timeouts; **always** close `resp.Body`.
- **CT-003 (MUST)** Use the **WNC SDK** for all controller access behind tiny interfaces. **No speculative parsing**; base mapping on verified fields.
- **CT-004 (SHOULD)** Normalize identifiers (e.g., MAC format) before requests.

---

## 8. Errors & Logging

- **EL-001 (MUST)** Wrap errors with `%w` and include identifiers (never secrets). Provide typed/sentinel errors where useful.
- **EL-002 (MUST)** Use `log/slog` with structured key/value fields. Keep messages concise.
- **EL-003 (SHOULD)** Classify transient vs. permanent errors; in collectors, **emit nothing** on transient errors instead of exporting partial/wrong values. Consider an `wnc_scrape_errors_total` counter per collector.

---

## 9. Testing Framework

### 8.A Unit Tests

- **TU-001 (MUST)** Use table‑driven tests and `t.Run`. Keep tests deterministic.
- **TU-002 (MUST)** For metrics surface, use `prometheus/testutil.CollectAndCompare` with **golden files** under `testdata/`.
- **TU-003 (SHOULD)** Fake the WNC SDK via interfaces; return realistic payloads. Place shared fakes/mocks under `pkg/testutil`.

### 8.B Integration Tests

- **TI-001 (SHOULD)** Guard with env vars (e.g., `WNC_CONTROLLER`, `WNC_ACCESS_TOKEN`). Skip if missing.
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

## 10. Performance & Memory

- **PM-001 (MUST)** Avoid per‑scrape allocations (prebuild descriptors/const labels, reuse buffers).
- **PM-002 (MUST)** Avoid unbounded maps keyed by user input or names.
- **PM-003 (SHOULD)** Consider a **poller + cache** when backend latency exceeds typical scrape timeouts.
- **PM-004 (MUST)** **Apply loop consolidation**: Use single loop O(n) instead of multiple loops O(n×modules). Build lookup maps once, pass to per-entity methods.
- **PM-005 (MUST)** **Use map-based lookups**: Convert linear O(n) searches to hash-based O(1) lookups with `build*Map()` helper functions.
- **PM-006 (MUST)** **Apply Slice of Structs pattern**: Replace repetitive metric registration with structured arrays for 25-30% code reduction.
- **PM-007 (SHOULD)** **Remove unnecessary validations**: Eliminate conditions guaranteed by Cisco C9800 design (association times, protocol data, etc.).

---

## 11. Collector Optimization Framework (18 Patterns)

### 11.1 Core Performance Patterns

- **OP-001 (MUST)** **Loop Consolidation**: Replace O(n×modules) with O(n) single-loop processing:

  ```go
  // ❌ Multiple loops (inefficient)
  for _, item := range data { /* General */ }
  for _, item := range data { /* Radio */ }
  for _, item := range data { /* Traffic */ }

  // ✅ Single loop (optimized)
  for _, item := range data {
      if c.metrics.General { c.collectGeneralForItem(...) }
      if c.metrics.Radio { c.collectRadioForItem(...) }
      if c.metrics.Traffic { c.collectTrafficForItem(...) }
  }
  ```

- **OP-002 (MUST)** **Map-based Lookups**: Convert O(n) linear search to O(1) hash lookup:

  ```go
  // Build maps once in Collect()
  deviceMap := buildDeviceMap(deviceData)  // O(n) build
  dot11Map := buildDot11Map(dot11Data)     // O(n) build

  // O(1) lookups in per-entity methods
  device := deviceMap[clientMAC]  // O(1) access
  ```

- **OP-003 (MUST)** **Slice of Structs Pattern**: Reduce repetitive metric registration:

  ```go
  metrics := []struct {
      desc  *prometheus.Desc
      value float64
  }{
      {c.speedDesc, float64(traffic.Speed)},
      {c.rssiDesc, float64(traffic.RSSI)},
  }

  for _, metric := range metrics {
      ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, metric.value, labels...)
  }
  ```

## 5.A New Configuration Design (2025-09-28 Redesign)

- **CF-004 (MUST)** **Use module-based collector configuration** following the new design:

  ```go
  // New collector module structure
  type APCollectorModules struct {
      General    bool     `json:"general"`    // admin_state, oper_state, radio_state, config_state, uptime, CPU, memory
      Radio      bool     `json:"radio"`      // channel, power, noise, utilization
      Traffic    bool     `json:"traffic"`    // clients, bytes, packets, frames
      Errors     bool     `json:"errors"`     // errors, drops, retries, failures
      Info       bool     `json:"info"`       // info metric with labels
      InfoLabels []string `json:"info_labels"`
  }

  type ClientCollectorModules struct {
      General    bool     `json:"general"`    // state, uptime, power_save_state
      Radio      bool     `json:"radio"`      // protocol, mcs, streams, speed, rssi, snr
      Traffic    bool     `json:"traffic"`    // bytes, packets
      Errors     bool     `json:"errors"`     // retries, drops, failures
      Info       bool     `json:"info"`       // info metric with labels
      InfoLabels []string `json:"info_labels"`
  }

  type WLANCollectorModules struct {
      General    bool     `json:"general"`    // enabled
      Traffic    bool     `json:"traffic"`    // clients, bytes
      Config     bool     `json:"config"`     // auth, security, networking settings
      Info       bool     `json:"info"`       // info metric with labels
      InfoLabels []string `json:"info_labels"`
  }
  ```

- **CF-005 (MUST)** **Use the new flag structure**:

  - AP: `--collector.ap.general`, `--collector.ap.radio`, `--collector.ap.traffic`, `--collector.ap.errors`, `--collector.ap.info`
  - Client: `--collector.client.general`, `--collector.client.radio`, `--collector.client.traffic`, `--collector.client.errors`, `--collector.client.info`
  - WLAN: `--collector.wlan.general`, `--collector.wlan.traffic`, `--collector.wlan.config`, `--collector.wlan.info`

- **CF-006 (MUST)** **No backward compatibility** with old flags (inventory, state, cpu, memory, phy, rf, security, networking). The new design is breaking-change compliant.

### 11.2 Code Quality Patterns

- **OP-004 (MUST)** **Determine Functions**: Extract complex data logic into testable functions:

  ```go
  func determineBandFromRadioInfo(radioSlot int, radioType string) string {
      if radioSlot == 0 { return "2.4GHz" }
      return "5GHz"
  }
  ```

- **OP-005 (MUST)** **Remove Unnecessary Conditions**: Eliminate validations guaranteed by system design:

  ```go
  // ❌ Unnecessary validation (C9800 guarantees association time for RUN clients)
  if !dot11.MsAssocTime.IsZero() {
      uptime := time.Since(dot11.MsAssocTime).Seconds()
  }

  // ✅ Direct usage (design-guaranteed)
  uptime := time.Since(dot11.MsAssocTime).Seconds()
  ```

- **OP-006 (MUST)** **Unified Describe Method**: Consolidate individual describe\* functions:
  ```go
  func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
      if c.metrics.General { ch <- c.countDesc; ch <- c.stateDesc }
      if c.metrics.Radio { ch <- c.channelDesc; ch <- c.powerDesc }
  }
  ```

### 11.3 Architecture Patterns

- **OP-007 (MUST)** **Per-Entity Collection Methods**: Replace module-wide loops with per-entity processing:

  ```go
  func (c *Collector) collectGeneralMetricsForClient(
      ch chan<- prometheus.Metric,
      client ClientData,
      baseLabels []string,
      lookupMaps map[string]Data,
  ) { /* focused processing */ }
  ```

- **OP-008 (SHOULD)** **Consistent Function Arguments**: Standardize argument order across collect methods:

  1. `ch chan<- prometheus.Metric` (output)
  2. Primary entity data
  3. Secondary data (by usage frequency)
  4. Lookup maps (last)

- **OP-009 (MUST)** **String Constants**: Define repeated strings as constants:
  ```go
  const (
      labelMAC  = "mac"
      labelName = "name"
      labelBand = "band"
  )
  ```

### 11.4 Implementation Checklist

**Phase 1: Structure**

- [ ] Prometheus method order (New → Describe → Collect)
- [ ] Unified Describe method (eliminate describe\* functions)
- [ ] Per-entity collection methods

**Phase 2: Performance**

- [ ] Single-loop processing in Collect()
- [ ] Map-based lookups (build\*Map helpers)
- [ ] Slice of Structs for metric registration

**Phase 3: Quality**

- [ ] Remove unnecessary conditions (design guarantees)
- [ ] Extract determine\* functions
- [ ] Standardize argument order
- [ ] String constants for labels

**Phase 4: Validation**

- [ ] `make lint` passes
- [ ] `make build` succeeds
- [ ] `make test-unit` passes
- [ ] Performance improvement verified

---

## 12. Naming & API Conventions

- **NA-001 (MUST)** Use **imperative verbs** for actions (`Run`, `Start`, `Shutdown`).
- **NA-002 (MUST)** Keep package names singular and stutter‑free (`collector/ap`, not `collectors` or `apcollector`).
- **NA-003 (SHOULD)** Export only what the `cmd/` needs; keep internals unexported.
- **NA-004 (SHOULD)** **Minimize function scope**. Functions should be defined in the most specific scope where they are used:
  - Functions used by a single file → private functions within that file
  - Functions shared by multiple files within a package → package-level unexported functions
  - Functions shared across packages → exported functions in the most specific common package
  - Avoid global utility functions that create unnecessary coupling between packages

---

## 13. Documentation

- **DC-001 (MUST)** Start each exported type/function comment with its identifier.
- **DC-002 (MUST)** Document each metric family in code (help text) and in `README.md` table (name, type, labels).
- **DC-003 (SHOULD)** Provide minimal code examples in package comments.
- **DC-004 (SHOULD)** Document optimization patterns applied and performance improvements achieved.
- **DC-005 (MUST)** Update `.copilot_reports/` with optimization results and implementation notes.

---

## 14. Quality Gate

- **QG-001 (MUST)** CI must run: `go vet`, `staticcheck`, `golangci-lint`, `go test ./...`, and (if configured) `promtool` checks for examples.
- **QG-002 (MUST)** `go.mod` is tidy; pinned minimal dependencies (prefer stdlib + `client_golang`).
- **QG-003 (MUST)** Zero lint violations; no flaky tests; deterministic metric order in tests (use `testutil.ToFloat64` or golden with stable ordering).
- **QG-004 (MUST)** **Optimization compliance**: All collectors must implement the 18-pattern framework before production deployment.
- **QG-005 (SHOULD)** **Performance benchmarks**: Measure and document performance improvements (loop reduction, memory efficiency).

---

## 15. Example: Server Wiring

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
// cmd/main.go calls internal/cli which handles application setup
// internal/cli/main.go (excerpt)
func (cmd *cli.Command) Action(ctx context.Context, cmd *cli.Command) error {
    cfg, err := config.Parse(cmd)
    if err != nil {
        return errors.New("configuration error")
    }

    slog.SetDefault(log.Setup(cfg.Log))

    if cfg.DryRun {
        slog.Info("Configuration validation successful", "dry_run", true)
        return nil
    }

    return server.StartAndServe(ctx, cfg, getVersion())
}

// internal/server/server.go handles collector registration
func StartAndServe(ctx context.Context, cfg *config.Config, version string) error {
    collector := collector.NewCollector(cfg)
    collector.Setup(version)

    srv := server.New(collector.Registry(), net.JoinHostPort(cfg.Web.ListenAddress, strconv.Itoa(cfg.Web.ListenPort)))
    // graceful shutdown and server startup
    return srv.ListenAndServe()
}
```

---

## 16. Optimization Implementation Status

### Completed Optimizations

#### Client Collector ✅ (100% Complete - 18/18 patterns)

- **Status**: Production ready
- **Code reduction**: 68 lines (7.9% reduction)
- **Performance**: O(n×modules) → O(n), 5× theoretical speedup
- **Quality**: All tests passing, zero lint violations
- **Patterns applied**: All 18 optimization patterns successfully implemented

#### AP Collector ✅ (100% Complete - 18/18 patterns)

- **Status**: Production ready
- **Code reduction**: 530 lines (35.6% reduction)
- **Performance**: O(n×modules) → O(n), map-based O(1) lookups
- **Quality**: All tests passing, zero lint violations
- **Patterns applied**: All 18 optimization patterns successfully implemented
- **New flag compliance**: Updated to use new collector.ap.general/radio/traffic/errors/info structure

#### WLAN Collector ⏳ (Pending)

- **Status**: Requires optimization
- **Target**: Apply all 18 optimization patterns
- **Priority**: Next for implementation

### Key Optimization Achievements

#### Performance Improvements

- **Loop consolidation**: 80% reduction in iteration cycles
- **Map-based lookups**: O(n) → O(1) for data access
- **Memory efficiency**: 64% reduction in duplicate map building
- **CPU utilization**: Eliminated unnecessary validation checks

#### Code Quality Improvements

- **Slice of Structs**: 25-30% reduction in metric registration code
- **Determine functions**: Extracted complex logic into testable units
- **String constants**: Eliminated repeated string literals
- **Unified architecture**: Consistent patterns across all collectors

#### Technical Debt Reduction

- **Removed deprecated methods**: 13 unused legacy functions eliminated
- **Simplified conditionals**: Removed design-guaranteed validations
- **Consistent naming**: Unified \*Map pattern for all lookup structures
- **Argument standardization**: Logical order across all collection methods

### Implementation Template

When optimizing new collectors, follow this proven sequence:

1. **Structure**: Prometheus method order, unified Describe
2. **Performance**: Single-loop, map-based lookups, Slice of Structs
3. **Quality**: Remove unnecessary conditions, extract determine functions
4. **Validation**: Lint, build, test, performance verification

### Success Metrics

- **Client Collector**: 864→796 lines (7.9% reduction), O(n×5)→O(n) loops, 5 map builds→1
- **AP Collector**: 1488→958 lines (35.6% reduction), complete legacy method removal, unified architecture
- **Configuration**: Complete redesign with new module structure (breaking changes)
- **Combined**: 600+ lines removed, 2 collectors production-optimized, new config architecture deployed

### Current Implementation Status (2025-09-28)

- ✅ **Client Collector**: 100% optimized (18/18 patterns)
- ✅ **AP Collector**: 100% optimized (18/18 patterns)
- ✅ **Configuration Architecture**: New module-based design deployed
- ✅ **Breaking Changes**: Complete migration from old flag structure
- ⏳ **WLAN Collector**: Awaiting optimization (target: Q4 2025)

### Recent Achievements

- **Prometheus Community Standards**: Full compliance with collector patterns
- **Cisco C9800 Expertise**: Wireless protocol knowledge integrated for validation optimization
- **Performance Engineering**: Algorithmic improvements (O(n×modules)→O(n))
- **Code Quality**: Self-documenting code through determine\* functions and Slice of Structs patterns

This optimization framework has been **field-tested and proven** across multiple collectors, delivering measurable performance improvements while maintaining code quality and test coverage.

### Breaking Changes Implemented (2025-09-28)

#### Configuration Redesign

- **BC-001** **Complete flag structure redesign**: Migrated from old inventory/state/cpu/memory/phy/rf flags to new general/radio/traffic/errors/info structure
- **BC-002** **No backward compatibility**: Old flags are completely removed to maintain design consistency
- **BC-003** **Module consolidation**: General module consolidates admin/oper state + CPU + memory; Radio module consolidates channel + power + RF metrics
- **BC-004** **CLI simplification**: Removed verbose flag descriptions with parenthetical details to avoid dual maintenance

#### Implementation Impact

- **Files updated**: `internal/config/config.go`, `internal/cli/main.go`, `internal/collector/main.go`, `.air.toml`
- **Quality assurance**: All changes verified with `make lint`, `make build`, `make test-unit`
- **Documentation**: README.md remains the single source of truth for detailed module descriptions

#### Migration Guide

```bash
# Old flags (REMOVED)
--collector.ap.inventory --collector.ap.state --collector.ap.cpu --collector.ap.memory
--collector.ap.phy --collector.ap.rf
--collector.client.inventory --collector.client.phy --collector.client.rf
--collector.wlan.inventory --collector.wlan.state --collector.wlan.security --collector.wlan.networking

# New flags (CURRENT)
--collector.ap.general --collector.ap.radio --collector.ap.traffic --collector.ap.errors --collector.ap.info
--collector.client.general --collector.client.radio --collector.client.traffic --collector.client.errors --collector.client.info
--collector.wlan.general --collector.wlan.traffic --collector.wlan.config --collector.wlan.info
```

````
```
````
