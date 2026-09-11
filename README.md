<div align="center">

  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/logo_dark.png" width="115px" />
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/logo.png" width="115px" />
    <img alt="cisco-wnc-exporter" src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/logo.png" width="115px" />
  </picture>

  <h1>cisco-wnc-exporter</h1>

  <p>A third-party Prometheus Exporter for Cisco C9800 Wireless Network Controller.</p>

  <p>
    <img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/umatare5/cisco-wnc-exporter?label=Latest%20version" />
    <a href="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml"><img alt="Test and Build" src="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml/badge.svg?branch=main" /></a>
    <img alt="Test Coverage" src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/coverage.svg" />
    <a href="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-vulncheck.yml"><img alt="govulncheck" src="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-vulncheck.yml/badge.svg?branch=main" /></a><br>
    <a href="https://www.bestpractices.dev/projects/11293"><img alt="OpenSSF Best Practices" src="https://www.bestpractices.dev/projects/11293/badge" /></a>
    <a href="./LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" /></a>
    <a href="https://developer.cisco.com/codeexchange/github/repo/umatare5/cisco-wnc-exporter"><img alt="Published" src="https://static.production.devnetcloud.com/codeexchange/assets/images/devnet-published.svg" /></a>
  </p>

</div>

## Overview

This exporter lets Prometheus scrape metrics from [Cisco Catalyst 9800 Wireless Controllers](https://www.cisco.com/site/us/en/products/networking/wireless/wireless-lan-controllers/catalyst-9800-series/index.html).

- 🛡️ **Critical State Monitoring**: Detects AP mis-configuration and WLAN enable or disable
- 🌐 **Client Connectivity Tracking**: Follows client signal, speed, protocol, traffic and latency
- 📊 **Long-Term Observability**: Extends metric retention for historical and trend analysis
- ↩️ **Pull-Based Telemetry**: An alternative to [Streaming Telemetry](https://www.cisco.com/c/en/us/td/docs/wireless/controller/9800/17-15/config-guide/b_wl_17_15_cg/streaming-telemetry-on-Cisco-Catalyst-9800-series-wireless-controller.html), over RESTCONF

## Quick Start

Enable RESTCONF and HTTPS on the C9800 first, as the [Programmability Guide](https://www.cisco.com/c/en/us/td/docs/ios-xml/ios/prog/configuration/1715/b_1715_programmability_cg/restconf_protocol.html#id_125840) describes.

### 1. Generate a Basic Auth token

Encode your controller credentials as Base64.

```bash
# username:password → Base64
echo -n "admin:your-password" | base64
# Output: YWRtaW46eW91ci1wYXNzd29yZA==
```

### 2. Set required environment variables

```bash
export WNC_CONTROLLER="wnc1.example.internal"
export WNC_ACCESS_TOKEN="YWRtaW46eW91ci1wYXNzd29yZA=="
```

### 3. Run the exporter with Docker

```bash
docker run -p 10039:10039 -e WNC_CONTROLLER -e WNC_ACCESS_TOKEN \
  ghcr.io/umatare5/cisco-wnc-exporter:latest --collector.ap.general
```

> [!TIP]
> If you prefer using binaries, download them from the [Release](https://github.com/umatare5/cisco-wnc-exporter/releases).
>
> **Supported Platform:** `linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64` and `windows_amd64`

## Supported Versions

Every reading was measured on a Catalyst 9800 running IOS-XE 17.12 and 17.15.

- **No release gate** — nothing tests the version, so an older image serves the types it carries.
- **A missing data type is a failure** — a controller that answers `404` raises `wnc_refresh_errors_total`, because a path the exporter got wrong answers `404` too.
- **Renumbering is detectable** — [Enumeration Values](docs/enums.md) records the YANG revision behind each family.

## Collectors

Each collector is switched on per group of families rather than as a whole, all off by default.

| Collector                                      | Publishes                                          |
| :--------------------------------------------- | :------------------------------------------------- |
| **[AP](docs/collector.ap.md)**                 | RF foundation and radio performance                |
| **[Client](docs/collector.client.md)**         | User experience quality and connection performance |
| **[WLAN](docs/collector.wlan.md)**             | Logical SSID performance and parameter checks      |
| **[Controller](docs/collector.controller.md)** | The controller itself, such as its boot time       |

## Flags

`cisco-wnc-exporter --help` prints every flag, and [Help](docs/help.md) carries the same list with its notes.

- **`--collector.<collector>.<group>`** — switches one group of families on, e.g. `--collector.ap.radio`
- **`--collector.info-cache-ttl`** — ages the `_info` series, which `--collector.*.info-labels` labels
- **`--collector.internal.*`** — adds the Go runtime and process families of the exporter itself
- **`--wnc.*`** — the controller address, the token, the request timeout and the refresh interval
- **`--web.*`** — the listen address, the port and the telemetry path
- **`WNC_CONTROLLER`** — fills `--wnc.controller`, and the flag overrides it
- **`WNC_ACCESS_TOKEN`** — fills `--wnc.access-token`, which keeps the credential off the process table

> [!CAUTION]
> `--wnc.tls-skip-verify` disables TLS certificate verification. **Never use it in production.**

## Endpoints

The exporter serves three endpoints:

- `/` — landing page, which confirms the exporter is running at <http://localhost:10039/>
- `/metrics` — metrics endpoint, configurable via `--web.telemetry-path`
- `/healthz` — liveness probe, which returns a static 200 and deliberately ignores WNC reachability

> [!NOTE]
> See [Endpoints](docs/README.md#endpoints) for the method and status contract, and what `--web.telemetry-path` of `/` changes.

## Metrics

This exporter publishes 150 metric families, catalogued per collector rather than in one list.

| Page                                       | Covers                                      |
| :----------------------------------------- | :------------------------------------------ |
| [AP](docs/collector.ap.md)                 | Every AP and per-radio series               |
| [Client](docs/collector.client.md)         | Every per-client series                     |
| [WLAN](docs/collector.wlan.md)             | Every per-WLAN series                       |
| [Controller](docs/collector.controller.md) | Every controller-wide series                |
| [Exporter Health](docs/health.md)          | The exporter's own build and refresh series |

The series a dashboard usually starts from:

| Collector  | Metric                             | Type  | Description                          |
| :--------- | :--------------------------------- | :---- | :----------------------------------- |
| AP         | `wnc_ap_oper_state`                | Gauge | Operational state (4=registered)     |
| AP         | `wnc_ap_channel_utilization_ratio` | Gauge | Channel utilization ratio (CCA), 0-1 |
| Client     | `wnc_client_state`                 | Gauge | Connection state (11=run state)      |
| Client     | `wnc_client_rssi_dbm`              | Gauge | Signal strength (dBm)                |
| WLAN       | `wnc_wlan_clients`                 | Gauge | Run-state clients count (calculated) |
| Controller | `wnc_controller_boot_time_seconds` | Gauge | Unix time of the last boot           |

> [!NOTE]
> See [Technical Information](docs/README.md#technical-information) for the refresh, absence and counter-reset rules every series obeys.

> [!IMPORTANT]
>
> Every collector is **disabled by default** to spare both Prometheus and the controller, and an exporter with no collector enabled never contacts the controller at all.
>
> - A minor release may rename or remove a metric, because the controller owns the schema every series reads.
> - A rename carries the type, the labels and the value of the old name unless [`CHANGELOG.md`](CHANGELOG.md) says otherwise.

### Exporter Health Metrics

These series describe the exporter rather than the network, so no collector flag names them.

| Metric                                  | Type    | Description                             |
| :-------------------------------------- | :------ | :-------------------------------------- |
| `wnc_build_info`                        | Gauge   | Exporter version in `version`, always 1 |
| `wnc_up`                                | Gauge   | Whether the last refresh reached WNC    |
| `wnc_refresh_errors_total`              | Counter | Fetch failures per `data` type          |
| `wnc_refresh_success_timestamp_seconds` | Gauge   | Start of the refresh being served       |

> [!NOTE]
> See [Exporter Health](docs/health.md) for the other three series and what a partial failure leaves `wnc_up` reading, and for every `data` label value the refresh can key on.

## Examples

### Command Lines

Start with one group and add what a dashboard needs, because no group is on by default.

```bash
# Nothing but wnc_build_info: no collector flag, so the controller is never contacted
./cisco-wnc-exporter

# The three series a first dashboard needs
./cisco-wnc-exporter --collector.ap.general --collector.client.general --collector.wlan.general
```

Add the flags from [`.air.toml`](https://github.com/umatare5/cisco-wnc-exporter/blob/main/.air.toml) to enable every group with the widest label set.

### Prometheus Configuration

#### Job Configuration Example

Add the scrape job from [`examples/prometheus.yml`](./examples/prometheus.yml) to your configuration.

#### Alerting Rules Configuration Example

Add the alerting rules from [`examples/prometheus_alert_rules.yml`](./examples/prometheus_alert_rules.yml) to your configuration.

> [!NOTE]
> See [Scrape Path](docs/README.md#scrape-path) for the interval a scrape observes, a multiple of `scrape_interval`.

### Grafana Dashboard

#### Grafana Admin-level Dashboard Example

Import [`examples/grafana_cisco-wnc-exporter-admin-dashboard.json`](https://github.com/umatare5/cisco-wnc-exporter/blob/main/examples/grafana_cisco-wnc-exporter-admin-dashboard.json) to add admin dashboard.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-exporter-admin-dashboard_dark.png">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-exporter-admin-dashboard.png">
  <img alt="Grafana admin-level dashboard" src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-exporter-admin-dashboard.png">
</picture>

> [!TIP]
> See [`docs/assets/cisco-wnc-exporter-admin-dashboard_full.png`](https://github.com/umatare5/cisco-wnc-exporter/blob/main/docs/assets/cisco-wnc-exporter-admin-dashboard_full.png) for the full capture.

#### Grafana User-level Dashboard Example

Import [`examples/grafana_cisco-wnc-exporter-user-dashboard.json`](https://github.com/umatare5/cisco-wnc-exporter/blob/main/examples/grafana_cisco-wnc-exporter-user-dashboard.json) to add user dashboard.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-exporter-user-dashboard_dark.png">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-exporter-user-dashboard.png">
  <img alt="Grafana user-level dashboard" src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-exporter-user-dashboard.png">
</picture>

> [!TIP]
> See [`docs/assets/cisco-wnc-exporter-user-dashboard_full.png`](https://github.com/umatare5/cisco-wnc-exporter/blob/main/docs/assets/cisco-wnc-exporter-user-dashboard_full.png) for the full capture.

## Contributing

See [`CONTRIBUTING.md`](https://github.com/umatare5/cisco-wnc-exporter/blob/main/CONTRIBUTING.md) for the development setup, the tests and the documentation conventions.

## License

MIT. The binary statically links Apache-2.0, MIT and BSD 3-Clause dependencies, whose notices are reproduced in [`NOTICE`](NOTICE) and shipped alongside [`LICENSE`](LICENSE) in every release archive and container image.
