<div align="center">

  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/logo_dark.png" width="180px" />
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/logo.png" width="180px" />
    <img src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/logo.png" width="180px" />
  </picture>

  <h1>cisco-wnc-exporter</h1>

  <p>A third-party Prometheus Exporter for Cisco C9800 Wireless Network Controller.</p>

  <p>
    <img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/umatare5/cisco-wnc-exporter?label=Latest%20version" />
    <a href="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml"><img alt="Test and Build" src="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml/badge.svg?branch=main" /></a>
    <img alt="Test Coverage" src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/coverage.svg" /><br>
    <a href="https://www.bestpractices.dev/projects/11293"><img alt="OpenSSF Best Practices" src="https://www.bestpractices.dev/projects/11293/badge" /></a>
    <a href="./LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" /></a>
    <a href="https://developer.cisco.com/codeexchange/github/repo/umatare5/cisco-wnc-exporter"><img alt="Published" src="https://static.production.devnetcloud.com/codeexchange/assets/images/devnet-published.svg" /></a>
  </p>

</div>

## Overview

This exporter allows a Prometheus instance to scrape metrics from [Cisco Catalyst 9800 Wireless Controllers](https://www.cisco.com/site/us/en/products/networking/wireless/wireless-lan-controllers/catalyst-9800-series/index.html).

- 🛡️ **Critical State Monitoring**: Detects changes such as AP mis-configurations or WLAN enable/disable
- 🌐 **Client Connectivity Tracking**: Monitors client signal strength, speed, protocols, traffic and latency
- 📊 **Long-Term Observability**: Extends metric retention for historical analysis and wireless trend tracking
- ↩️ **Pull-Based Telemetry**: Alternative to the [Streaming Telemetry](https://www.cisco.com/c/en/us/td/docs/wireless/controller/9800/17-12/config-guide/b_wl_17_12_cg/streaming-telemetry-on-Cisco-Catalyst-9800-series-wireless-controller.html) feature using RESTCONF APIs

## Quick Start

Please enable RESTCONF and HTTPS on the C9800 before using this exporter. Please see:

- [Cisco IOS XE 17.12 Programmability Configuration Guide — RESTCONF](https://www.cisco.com/c/en/us/td/docs/ios-xml/ios/prog/configuration/1712/b_1712_programmability_cg/m_1712_prog_restconf.html#id_70432)

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
  ghcr.io/umatare5/cisco-wnc-exporter:latest
```

> [!Tip]
> If you prefer using binaries, download them from the [release page](https://github.com/umatare5/cisco-wnc-exporter/releases).
>
> Supported Platforms are: `linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64` and `windows_amd64`

## Syntax

`cisco-wnc-exporter --help` prints every flag, and [docs/configuration.md](docs/configuration.md) carries the same list.

Each collector is enabled per module:

| Module       | Flags                                                                                                    |
| :----------- | :------------------------------------------------------------------------------------------------------- |
| `ap`         | `--collector.ap.general`, `.radio`, `.traffic`, `.errors`, `.join`, `.geolocation`, `.spectrum`, `.info` |
| `client`     | `--collector.client.general`, `.radio`, `.traffic`, `.errors`, `.info`                                   |
| `wlan`       | `--collector.wlan.general`, `.traffic`, `.config`, `.info`                                               |
| `controller` | `--collector.controller.general`                                                                         |

> [!CAUTION]
> `--wnc.tls-skip-verify` disables TLS certificate verification. **Never use it in production.**

## Configuration

This exporter reads two environment variables:

| Environment Variable | Description                                      |
| :------------------- | :----------------------------------------------- |
| `WNC_CONTROLLER`     | WNC controller hostname or IP address (required) |
| `WNC_ACCESS_TOKEN`   | WNC API access token (required)                  |

## Metrics

This exporter collects wireless network metrics from Cisco C9800 WNC using four collectors:

| Collector                                      | Focus                                              |
| :--------------------------------------------- | :------------------------------------------------- |
| **[AP](docs/collector.ap.md)**                 | RF foundation and radio performance                |
| **[Client](docs/collector.client.md)**         | User experience quality and connection performance |
| **[WLAN](docs/collector.wlan.md)**             | Logical SSID performance and parameter checks      |
| **[Controller](docs/collector.controller.md)** | The controller itself metrics such as boot time    |

The series a dashboard usually starts from:

| Collector  | Metric                             | Type  | Description                          |
| :--------- | :--------------------------------- | :---- | :----------------------------------- |
| AP         | `wnc_ap_oper_state`                | Gauge | Operational state (4=registered)     |
| AP         | `wnc_ap_channel_number`            | Gauge | Operating channel number             |
| AP         | `wnc_ap_tx_power_dbm`              | Gauge | Current transmit power (dBm)         |
| AP         | `wnc_ap_noise_floor_dbm`           | Gauge | Noise on the operating channel (dBm) |
| AP         | `wnc_ap_channel_utilization_ratio` | Gauge | Channel utilization ratio (CCA), 0-1 |
| AP         | `wnc_ap_clients`                   | Gauge | Run-state clients count (calculated) |
| Client     | `wnc_client_state`                 | Gauge | Connection state (11=run state)      |
| Client     | `wnc_client_protocol`              | Gauge | 802.11 protocol (0=unknown, 1..7)    |
| Client     | `wnc_client_speed_mbps`            | Gauge | Negotiated PHY rate (Mbps)           |
| Client     | `wnc_client_rssi_dbm`              | Gauge | Signal strength (dBm)                |
| Client     | `wnc_client_snr_decibels`          | Gauge | Signal-to-noise ratio (dB)           |
| WLAN       | `wnc_wlan_enabled`                 | Gauge | WLAN status                          |
| WLAN       | `wnc_wlan_clients`                 | Gauge | Run-state clients count (calculated) |
| Controller | `wnc_controller_boot_time_seconds` | Gauge | Unix time of the last boot           |

See [docs/README.md](docs/README.md) for the refresh, caching, counter-reset and state semantics every collector shares.

> [!Important]
>
> All collectors are **disabled by default** to reduce load on both Prometheus and the controller, and an exporter with no collector enabled never contacts the controller at all.
>
> - Every enabled collector is served from one refresh, which runs at most once per `--wnc.cache-ttl`, so the controller sees one pass of requests per interval.
> - That refresh reads only the `data` types the enabled modules need, so a client-only deployment never fetches `ap_capwap_data`.

### Exporter Health Metrics

These series describe the exporter itself rather than the wireless network. They have no module and no collector flag. Without the refresh series a failed refresh produces a successful scrape carrying no series, which no alert can detect.

| Metric                                  | Type    | Description                                            |
| :-------------------------------------- | :------ | :----------------------------------------------------- |
| `wnc_build_info`                        | Gauge   | Exporter version in the `version` label, always 1      |
| `wnc_up`                                | Gauge   | Whether last **completed** refresh reached WNC         |
| `wnc_refresh_duration_seconds`          | Gauge   | Duration of the last refresh **attempt**               |
| `wnc_refresh_success_timestamp_seconds` | Gauge   | Start time of the refresh behind the served snapshot   |
| `wnc_refresh_errors_total`              | Counter | Fetch failures per `data` type since start-up          |
| `wnc_refresh_items`                     | Gauge   | Items the last refresh returned per `data` type        |
| `wnc_refresh_defaults_fallback_total`   | Counter | WLAN config fetches that fell back to a plain read     |

> [!Important]
>
> `wnc_up == 1` is not a claim that the data series are present, and `up == 1` is not a claim that the controller is reachable. A scrape always returns 200 because it is served from the cached snapshot.

> [!Note]
> `wnc_build_info` is registered before any collector, so it is the only series a scrape carries when every collector is disabled.

## Use Cases

### Exporter Configuration

The exporter serves three endpoints:

- `/` — landing page, which confirms the exporter is running when reached at <http://localhost:10039/>
- `/metrics` — metrics endpoint, which `--web.telemetry-path` moves to any path, including `/`, where it replaces the landing page
- `/healthz` — liveness probe, which returns a static 200 and deliberately ignores WNC reachability

> [!Note]
>
> Reflecting the WNC state in `/healthz` would let an orchestrator kill the exporter during a controller outage, taking the stale snapshot and the [Exporter Health Metrics](#exporter-health-metrics) series down with it.

#### Basic Usage - No Collectors

```bash
$ WNC_CONTROLLER="wnc1.example.internal"
$ WNC_ACCESS_TOKEN="$(echo -n 'admin:your-password' | base64)"
$ ./cisco-wnc-exporter
time="2025-04-13T18:50:54Z" level=info msg="Starting the cisco-wnc-exporter on port 10039."
```

#### Essential Usage

```bash
$ WNC_CONTROLLER="wnc1.example.internal"
$ WNC_ACCESS_TOKEN="$(echo -n 'admin:your-password' | base64)"
$ ./cisco-wnc-exporter \
    --collector.ap.general --collector.client.general --collector.wlan.general
```

#### Complete Usage

For complete monitoring, see [`.air.toml`](https://github.com/umatare5/cisco-wnc-exporter/blob/main/.air.toml) which enables every collector module with maximum info-labels.

### Prometheus Configuration

#### Job Configuration Example

Add the job config to your Prometheus YAML file using [examples/prometheus.yml](./examples/prometheus.yml) as a reference.

> [!Note]
>
> A refresh starts on the first scrape that arrives after `--wnc.cache-ttl` has elapsed since the previous refresh finished, so the effective refresh period is:
>
> ```text
> P = scrape_interval * ceil((cache-ttl + R) / scrape_interval)
> ```
>
> - `R` — the refresh duration, which `wnc_refresh_duration_seconds` reports.
> - `P` — 120s for `R` over 5s up to 65s, with default `--wnc.cache-ttl` 55s and `scrape_interval` 60s.

#### Alerting Rules Configuration Example

Add the alerting rules to your Prometheus YAML file using [examples/prometheus_alert_rules.yml](./examples/prometheus_alert_rules.yml) as a reference.

### Grafana Admin-level Dashboard Example

Import [examples/grafana_cisco-wnc-admin-dashboard.json](https://github.com/umatare5/cisco-wnc-exporter/blob/main/examples/grafana_cisco-wnc-admin-dashboard.json) to add the example Grafana Dashboard for administrators.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-admin-dashboard_dark.png">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-admin-dashboard.png">
  <img src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-admin-dashboard.png">
</picture>

> [!Tip]
> See [cisco-wnc-admin-dashboard_full.png](https://github.com/umatare5/cisco-wnc-exporter/blob/main/docs/assets/cisco-wnc-admin-dashboard_full.png) for the full capture image of the example.

### Grafana User-level Dashboard Example

Import [examples/grafana_cisco-wnc-user-dashboard.json](https://github.com/umatare5/cisco-wnc-exporter/blob/main/examples/grafana_cisco-wnc-user-dashboard.json) to add the example Grafana Dashboard for users.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-user-dashboard_dark.png">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-user-dashboard.png">
  <img src="https://raw.githubusercontent.com/umatare5/cisco-wnc-exporter/main/docs/assets/cisco-wnc-user-dashboard.png">
</picture>

> [!Tip]
> See [cisco-wnc-user-dashboard_full.png](https://github.com/umatare5/cisco-wnc-exporter/blob/main/docs/assets/cisco-wnc-user-dashboard_full.png) for the full capture image of the example.

## Contributing

See [CONTRIBUTING.md](https://github.com/umatare5/cisco-wnc-exporter/blob/main/CONTRIBUTING.md) for the `make` targets, the Docker build, the release process and how to open a pull request.

## Acknowledgement

I launched this project with the help of **GitHub Copilot Coding Agent**, and I am grateful to the global developer community for their contributions to open source projects and public repositories.

## Licence

[MIT](LICENSE). The binary statically links Apache-2.0, MIT and BSD 3-Clause dependencies, whose notices are reproduced in [NOTICE](NOTICE) and shipped alongside `LICENSE` in every release archive and container image.
