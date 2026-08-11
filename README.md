<div align="center">

  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo_dark.png" width="180px" />
    <source media="(prefers-color-scheme: light)" srcset="docs/assets/logo.png" width="180px" />
    <img src="docs/assets/logo.png" width="180px" />
  </picture>

  <h1>cisco-wnc-exporter</h1>

  <p>A third-party Prometheus Exporter for Cisco C9800 Wireless Network Controller.</p>

  <p>
    <img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/umatare5/cisco-wnc-exporter?label=Latest%20version" />
    <a href="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml"><img alt="Test and Build" src="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml/badge.svg?branch=main" /></a>
    <img alt="Test Coverage" src="docs/assets/coverage.svg" />
    <a href="https://goreportcard.com/badge/github.com/umatare5/cisco-wnc-exporter"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/umatare5/cisco-wnc-exporter" /></a><br/>
    <a href="https://www.bestpractices.dev/projects/11293"><img alt="OpenSSF Best Practices" src="https://www.bestpractices.dev/projects/11293/badge" /></a>
    <a href="./LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" /></a>
    <a href="https://developer.cisco.com/codeexchange/github/repo/umatare5/cisco-wnc-exporter"><img alt="Published" src="https://static.production.devnetcloud.com/codeexchange/assets/images/devnet-published.svg" /></a>
  </p>

</div>

## Overview

This exporter allows a prometheus instance to scrape metrics from [Cisco Catalyst 9800 Wireless Controllers](https://www.cisco.com/site/us/en/products/networking/wireless/wireless-lan-controllers/catalyst-9800-series/index.html).

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
  ghcr.io/umatare5/cisco-wnc-exporter
```

- `-p`: Maps container port `10039/tcp` to host port `10039/tcp`.
- `-e`: Passes the environment variables into the container.

> [!Tip]
> If you prefer using binaries, download them from the [release page](https://github.com/umatare5/cisco-wnc-exporter/releases).
>
> - Supported Platforms: `linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64` and `windows_amd64`

## Syntax

```bash
NAME:
   cisco-wnc-exporter - Prometheus exporter for Cisco WNC

USAGE:
   cisco-wnc-exporter [global options]

VERSION:
   0.1.0

OPTIONS:
   --dry-run                    Validate configuration without starting the server
   --help, -h                   show help
   --log.format string          Log format (json, text) (default: "json")
   --log.level string           Log level (debug, info, warn, error) (default: "info")
   --version, -v                print the version
   --web.listen-address string  Address to bind the HTTP server to (default: "0.0.0.0")
   --web.listen-port int        Port number to bind the HTTP server to (default: 10039)
   --web.telemetry-path string  Path for the metrics endpoint (default: "/metrics")
   --wnc.access-token string    WNC API access token [$WNC_ACCESS_TOKEN]
   --wnc.cache-ttl duration     Minimum interval between WNC data refreshes (default: 55s)
   --wnc.controller string      WNC controller hostname or IP address [$WNC_CONTROLLER]
   --wnc.timeout duration       WNC API request timeout (default: 55s)
   --wnc.tls-skip-verify        Skip TLS certificate verification (default: false)

   # AP Collector Options

   --collector.ap.errors              Enable AP error metrics
   --collector.ap.general             Enable AP general metrics
   --collector.ap.info                Enable AP info metrics
   --collector.ap.info-labels string  Comma-separated list of AP info labels (default: "name,ip")
   --collector.ap.radio               Enable AP radio metrics
   --collector.ap.traffic             Enable AP traffic metrics

   # Client Collector Options

   --collector.client.errors              Enable Client error metrics
   --collector.client.general             Enable Client general metrics
   --collector.client.info                Enable Client info metrics
   --collector.client.info-labels string  Comma-separated list of Client info labels (default: "name,ipv4")
   --collector.client.radio               Enable Client radio metrics
   --collector.client.traffic             Enable Client traffic metrics

   # WLAN Collector Options

   --collector.wlan.config              Enable WLAN config metrics
   --collector.wlan.general             Enable WLAN general metrics
   --collector.wlan.info                Enable WLAN info metrics
   --collector.wlan.info-labels string  Comma-separated list of WLAN info labels (default: "name")
   --collector.wlan.traffic             Enable WLAN traffic metrics

   * Collector Wide Options

   --collector.info-cache-ttl duration  Cache TTL for collector info metrics (default: 30m0s)

   * Internal Collector Options

   --collector.internal.go-runtime  Enable Go runtime metrics collector
   --collector.internal.process     Enable process metrics collector
```

> [!CAUTION]
> The `--wnc.tls-skip-verify` flag disables TLS certificate verification. This should only be used in development environments or when connecting to controllers with self-signed certificates. **Never use this option in production environments** as it compromises security.

## Configuration

This exporter supports following environment variables:

| Environment Variable | Description                                      |
| :------------------- | :----------------------------------------------- |
| `WNC_CONTROLLER`     | WNC controller hostname or IP address (required) |
| `WNC_ACCESS_TOKEN`   | WNC API access token (required)                  |

## Metrics

This exporter collects wireless network metrics from Cisco C9800 WNC using following collectors:

- **AP Collector** - For RF foundation and radio performance
- **Client Collector** - For user experience quality and connection performance
- **WLAN Collector** - For logical SSID performance and parameter checks

The `Module` column names the flag suffix that enables a metric, as in `--collector.ap.radio`.

The exporter also exposes the **[Exporter Health Metrics](#exporter-health-metrics)** series.

- These describe the exporter's own data refresh rather than the wireless network
- These are exposed whenever at least one of the collectors above is enabled

> [!Important]
>
> All collectors are **disabled by default** to reduce load on both Prometheus and the Cisco C9800 WNC. Because a Cisco C9800 WNC typically manages hundreds or even thousands of APs and clients, selective monitoring is essential to maintain performance and stability.
>
> **WNC data refresh (`--wnc.cache-ttl`)**
>
> - A scrape is served from the last snapshot and never waits for the controller
> - The flag sets the minimum idle time between refresh completions, not a snapshot expiry
> - The first scrape after start-up therefore reports `wnc_up 0` and carries no data series
> - A refresh is bounded at twice the flag value; data types the deadline never reached count as failures
> - Data series are withheld after three consecutive failed refreshes, so Prometheus can mark them stale
>
> **Prometheus info metrics caching (`--collector.info-cache-ttl`)**
>
> - Info metrics are served from a snapshot up to 1800 seconds (30 minutes) old, and the collector behind them still runs on every scrape, so no controller request is saved
> - A client that roamed keeps its previous `ap` label until the cache expires
> - A newly associated client is missing from the info metric for up to that long, so `group_left` joins on it return nothing
> - Caching does not reduce cardinality: every `ap` label value a client has held remains its own series

> [!Note]
>
> **WNC-side metrics management**
>
> - The controller updates AP and client counters, RSSI and SNR on its own schedule rather than on scrape
> - The AP profile `stats-timer` governs that schedule and defaults to 180 seconds
> - RRM noise and coverage update every 180 seconds, and RRM load every 60 seconds
> - Use a range of **15 minutes or more** for `rate()` and `increase()`
> - A shorter range spans too few controller updates to be meaningful
>
> **Counter reset timing**
>
> - Counters also reset when an AP re-joins CAPWAP, because the controller allocates fresh statistics
> - Query them with a range long enough to absorb a re-join, such as `increase(...[1h])`

> [!Note]
>
> **A state is a label, not a number**
>
> - A series carrying a `state` label always has the value `1`, so the label is the reading
> - The controller's spelling passes through unmapped, so a value this file does not name can appear
> - Only the current state has a series, so `== 0` never fires, and an equality match on the healthy spelling selects the healthy devices rather than revealing the unhealthy ones
> - Alert on the healthy spelling's absence, with `state` aggregated away:
>
> ```bash
> group by (mac) (wnc_client_state{state!="client-status-run"})
> ```
>
> - Keeping `state` in the result restarts `for:` on every state change, so a stuck device never fires
> - Pair the query with a `for:` longer than a legitimate transition takes
> - `wnc_client_state` also covers a client held short of `client-status-run`, which no other client series does
> - `wnc_ap_oper_state` is one series per AP, healthy at `registered`, and carries no `radio` label
> - Every other `_state` metric keeps its numeric `0` or `1`

### AP Collector

AP collector focuses on RF foundation and radio performance. See [docs/collector.ap.md](docs/collector.ap.md) for every metric it publishes and the labels `wnc_ap_info` carries.

> [!Note]
>
> **Utilization is reported as a ratio**
>
> - Metric names ending in `_ratio` carry 0 to 1, the Prometheus base unit for a percentage
> - Multiply by 100 for a percentage axis, as in `wnc_ap_channel_utilization_ratio * 100`
> - Grafana renders these series without that multiplication under the `percentunit` unit

> [!Note]
>
> **The `band` label**
>
> - AP band comes from the radio's operating band, not from the `radio` slot
> - A dual band radio changes band without changing slot, so the slot is not a band
> - Client band comes from the PHY generation the client associated with
> - `unknown` means the controller reported a value this exporter does not map
> - `unknown` is a label value, not a missing series: the other labels stay joinable
> - A band change takes up to `--collector.info-cache-ttl` to appear

> [!Warning]
>
> **Multi-link clients**
>
> - An 802.11be client may hold links on more than one band at once
> - The controller reports one PHY generation per client, so `band` names one link
> - Aggregations such as `count by (band)` therefore undercount the other links

### Client Collector

Client collector focuses on user experience quality and connection performance.

| Module  | Metric                                | Type    | Description                          |
| :------ | :------------------------------------ | :------ | :----------------------------------- |
| general | `wnc_client_state`                    | Gauge   | Connection state in `state` label    |
| general | `wnc_client_state_transition_seconds` | Gauge   | State transition latency             |
| general | `wnc_client_power_save_state`         | Gauge   | Power save state (0=active, 1=save)  |
| radio   | `wnc_client_protocol`                 | Gauge   | 802.11 protocol (0=unknown, 1..7)    |
| general | `wnc_client_uptime_seconds`           | Gauge   | Connection duration                  |
| radio   | `wnc_client_mcs_index`                | Gauge   | MCS index (-1=legacy, 0-11)          |
| radio   | `wnc_client_spatial_streams`          | Gauge   | Spatial streams count                |
| radio   | `wnc_client_speed_mbps`               | Gauge   | Connection throughput                |
| radio   | `wnc_client_rssi_dbm`                 | Gauge   | Signal strength (dBm)                |
| radio   | `wnc_client_snr_decibels`             | Gauge   | Signal-to-noise ratio (dB)           |
| traffic | `wnc_client_rx_bytes_total`           | Counter | Received bytes                       |
| traffic | `wnc_client_tx_bytes_total`           | Counter | Transmitted bytes                    |
| traffic | `wnc_client_rx_packets_total`         | Counter | Received packets                     |
| traffic | `wnc_client_tx_packets_total`         | Counter | Transmitted packets                  |
| errors  | `wnc_client_tx_retries_total`         | Counter | TX retries count **(\*3)**           |
| errors  | `wnc_client_data_retries_total`       | Counter | Data retries by mobile station       |
| errors  | `wnc_client_excessive_retries_total`  | Counter | Excessive retries count **(\*3)**    |
| errors  | `wnc_client_rts_retries_total`        | Counter | RTS retries count **(\*3)**          |
| errors  | `wnc_client_duplicate_received_total` | Counter | Duplicate packets received **(\*3)** |
| errors  | `wnc_client_tx_drops_total`           | Counter | TX drops count                       |
| errors  | `wnc_client_decryption_failed_total`  | Counter | Decryption failures                  |
| errors  | `wnc_client_mic_mismatch_total`       | Counter | MIC mismatch errors **(\*3)**        |
| errors  | `wnc_client_mic_missing_total`        | Counter | MIC missing errors **(\*3)**         |
| errors  | `wnc_client_policy_errors_total`      | Counter | Policy errors **(\*3)**              |
| errors  | `wnc_client_rx_group_total`           | Counter | RX group counter                     |

<details><summary><b>*3</b> Client error metrics consistently returning zero values on Cisco IOS-XE 17.12.6a with FlexConnect AP</summary><br/>

The following client error metrics consistently return zero values due to implementation limitations. That applies while the fetch succeeds: a data type whose fetch failed makes its series absent rather than zero.

- `wnc_client_duplicate_received_total`
- `wnc_client_excessive_retries_total`
- `wnc_client_mic_mismatch_total`
- `wnc_client_mic_missing_total`
- `wnc_client_policy_errors_total`
- `wnc_client_rts_retries_total`
- `wnc_client_tx_retries_total`

This was verified through direct RESTCONF API access to the live WNC environment:

```bash
❯ curl -sS -k -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
           -H "Accept: application/yang-data+json" \
           "https://$WNC_CONTROLLER/restconf/data/Cisco-IOS-XE-wireless-client-oper:client-oper-data/traffic-stats" \
           | jq '.["Cisco-IOS-XE-wireless-client-oper:traffic-stats"][0]' | \
           jq '{duplicate_rcv, tx_excessive_retries, mic_mismatch, mic_missing, policy_errs, rts_retries, rx_group_counter, tx_retries}'
{
  "duplicate_rcv": "0",
  "tx_excessive_retries": "0",
  "mic_mismatch": "0",
  "mic_missing": "0",
  "policy_errs": "0",
  "rts_retries": "0",
  "rx_group_counter": "0",
  "tx_retries": "0"
}
```

</details>

> [!Tip]
>
> `info` module provides `wnc_client_info` contains following labels to join with other metrics:
>
> | Labels     | Description                 | Example Value                 | Default | Required |
> | :--------- | :-------------------------- | :---------------------------- | :-----: | :------: |
> | `mac`      | MAC address                 | `aa:bb:cc:12:34:56`           | **Yes** | **Yes**  |
> | `ap`       | Access point identifier     | `TEST-AP01`                   |   No    |    No    |
> | `band`     | Radio band                  | `2.4`, `5`, `6`, `unknown`    |   No    |    No    |
> | `wlan`     | WLAN ESSID name             | `labo-wifi`                   |   No    |    No    |
> | `name`     | Device Classification Name  | `MacBook Pro (14-inch, 2021)` | **Yes** |    No    |
> | `username` | EAP authentication identity | `john.doe@example.com`        |   No    |    No    |
> | `ipv4`     | Client IPv4 address         | `192.168.1.100`               | **Yes** |    No    |
> | `ipv6`     | Client IPv6 address         | `2001:db8::1`                 |   No    |    No    |
>
> Use this info metric to add contextual labels to other metrics in PromQL queries:
>
> ```bash
> wnc_client_state * on(mac) group_left(ap,wlan,name) wnc_client_info
> ```

### WLAN Collector

WLAN collector focuses on logical SSID performance and parameter checks.

| Module  | Metric                                    | Type  | Description                          |
| :------ | :---------------------------------------- | :---- | :----------------------------------- |
| general | `wnc_wlan_enabled`                        | Gauge | WLAN status                          |
| traffic | `wnc_wlan_clients`                        | Gauge | Run-state clients count (calculated) |
| config  | `wnc_wlan_auth_psk_enabled`               | Gauge | PSK authentication enabled           |
| config  | `wnc_wlan_auth_dot1x_enabled`             | Gauge | 802.1x authentication enabled        |
| config  | `wnc_wlan_auth_dot1x_sha256_enabled`      | Gauge | 802.1x SHA256 auth enabled           |
| config  | `wnc_wlan_wpa2_enabled`                   | Gauge | WPA2 support enabled                 |
| config  | `wnc_wlan_wpa3_enabled`                   | Gauge | WPA3 support enabled                 |
| config  | `wnc_wlan_session_timeout_seconds`        | Gauge | Session timeout duration             |
| config  | `wnc_wlan_load_balance_enabled`           | Gauge | Load balancing enabled               |
| config  | `wnc_wlan_11k_neighbor_list_enabled`      | Gauge | 802.11k neighbor list enabled        |
| config  | `wnc_wlan_client_steering_enabled`        | Gauge | 6GHz client steering enabled         |
| config  | `wnc_wlan_central_switching_enabled`      | Gauge | Central switching enabled            |
| config  | `wnc_wlan_central_authentication_enabled` | Gauge | Central authentication enabled       |
| config  | `wnc_wlan_central_dhcp_enabled`           | Gauge | Central DHCP enabled                 |
| config  | `wnc_wlan_central_association_enabled`    | Gauge | Central association enabled          |

> [!Tip]
>
> `info` module provides `wnc_wlan_info` contains following labels to join with other metrics:
>
> | Labels | Description     | Example Value  | Default | Required |
> | :----- | :-------------- | :------------- | :-----: | :------: |
> | `id`   | WLAN identifier | `1`, `5`, `10` | **Yes** | **Yes**  |
> | `name` | WLAN ESSID name | `labo-wifi`    | **Yes** |    No    |
>
> Use this info metric to add contextual labels to other metrics in PromQL queries:
>
> ```bash
> wnc_wlan_enabled * on(id) group_left(name) wnc_wlan_info
> ```

### Exporter Health Metrics

These series describe the exporter's own WNC data refresh. They have no module and no collector flag: enabling any collector above exposes them all. Without them a failed refresh produces a successful scrape carrying no series, which no alert can detect.

| Metric                                  | Type    | Description                                            |
| :-------------------------------------- | :------ | :----------------------------------------------------- |
| `wnc_up`                                | Gauge   | Whether the last **completed** refresh reached the WNC |
| `wnc_refresh_duration_seconds`          | Gauge   | Duration of the last refresh **attempt**               |
| `wnc_refresh_success_timestamp_seconds` | Gauge   | Start time of the refresh behind the served snapshot   |
| `wnc_refresh_errors_total`              | Counter | Fetch failures per `data` type since start-up          |
| `wnc_refresh_items`                     | Gauge   | Items the last refresh returned per `data` type        |

> [!Important]
>
> `wnc_up == 1` is not a claim that the data series are present, and `up == 1` is not a claim that the controller is reachable. A scrape always returns 200 because it is served from the cached snapshot, so the target's `up` reports only that the exporter's HTTP server answered.

> [!Note]
>
> `wnc_refresh_items` is recorded on success only, so an absent series means that fetch failed while a zero series means the controller returned nothing. `wnc_refresh_errors_total` is seeded to zero for every `data` type at start-up, which makes it the authoritative list of `data` label values.

## Use Cases

There are multiple ways to run the exporter, including direct binary execution and Docker containerization.

### Exporter Configuration

The exporter serves three endpoints:

- `/` - Landing page. Visit http://localhost:10039/ to verify the exporter is running
- `/metrics` - Metrics endpoint, moved by `--web.telemetry-path`. Pointing it at `/` replaces the landing page
- `/healthz` - Liveness probe. Returns a static 200 and deliberately ignores WNC reachability

> [!Note]
>
> `/healthz` stays liveness-only on purpose. Reflecting the WNC state there would let an orchestrator kill the exporter during a controller outage, taking the stale snapshot and the [Exporter Health Metrics](#exporter-health-metrics) series down with it — exactly when they are needed.

#### Basic Usage - No Collectors

The exporter starts without any collectors enabled by default:

```bash
$ WNC_CONTROLLER="wnc1.example.internal"
$ WNC_ACCESS_TOKEN="foobarbaz"
$ ./cisco-wnc-exporter
time="2025-04-13T18:50:54Z" level=info msg="Starting the cisco-wnc-exporter on port 10039."
```

#### Essential Usage

Enable essential collectors for basic monitoring:

```bash
$ WNC_CONTROLLER="wnc1.example.internal"
$ WNC_ACCESS_TOKEN="foobarbaz"
$ ./cisco-wnc-exporter \
    --collector.ap.general --collector.client.general --collector.wlan.general
```

#### Complete Usage

For complete monitoring, see [`.air.toml`](./.air.toml) which enables all collectors with maximum info-labels.

### Prometheus Configuration

This section describes how to configure Prometheus to scrape metrics from the cisco-wnc-exporter.

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
> - `R`: The refresh duration, which `wnc_refresh_duration_seconds` reports.
> - `P`: 120 seconds for `R` of 5 to 65 seconds at a 60s `scrape_interval`.

#### Alerting Rules Configuration Example

Add the alerting rules to your Prometheus YAML file using [examples/prometheus_alert_rules.yml](./examples/prometheus_alert_rules.yml) as a reference.

### Example Grafana Dashboard

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/cisco-wnc-exporter-dashboard_dark.png">
  <source media="(prefers-color-scheme: light)" srcset="docs/assets/cisco-wnc-exporter-dashboard.png">
  <img src="docs/assets/cisco-wnc-exporter-dashboard.png">
</picture>

## Development

### Commands

The following `make` commands are available for development and testing:

| Command                   | Description                                    |
| :------------------------ | :--------------------------------------------- |
| `make help`               | Display available targets and requirements     |
| `make build`              | Build the binary to `./tmp/cisco-wnc-exporter` |
| `make lint`               | Run golangci-lint and tidy go.mod              |
| `make test-unit`          | Run unit tests with coverage using gotestsum   |
| `make test-unit-coverage` | Generate HTML coverage report                  |
| `make clean`              | Remove build artifacts and backup files        |
| `make image`              | Build Docker image                             |

### Build

The repository includes a ready to use `Dockerfile`. To build a new Docker image:

```bash
make image
```

This creates an image named `ghcr.io/$USER/cisco-wnc-exporter` and exposes `10039/tcp`.

### Release

To release a new version, follow these steps:

1. Update the version in the `VERSION` file.
2. Submit a pull request with the updated `VERSION` file.

Once the pull request is merged, the GitHub Workflow automatically creates and pushes a new tag, after which I manually publish a release using the [GitHub Actions release workflow](https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-release.yml).

## Contribution

1. Fork ([https://github.com/umatare5/cisco-wnc-exporter/fork](https://github.com/umatare5/cisco-wnc-exporter/fork))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Create a new Pull Request

## Acknowledgement

I launched this project with the help of **GitHub Copilot Coding Agent**, and I am grateful to the global developer community for their contributions to open source projects and public repositories.

## Licence

[MIT](LICENSE)
