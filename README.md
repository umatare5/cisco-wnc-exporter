<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo_dark.png" width="180px">
  <img src="docs/assets/logo.png" width="180px">
</| errors    | `wnc_ap_transmission_failures_total`  | Counter | Failed transmission attempts **(*1)** **(*2)**                                                                |icture>

  <h1>cisco-wnc-exporter</h1>

  <p>A third-party Prometheus Exporter for Cisco C9800 Wireless Network Controller.</p>

  <p>
    <img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/umatare5/cisco-wnc-exporter?label=Latest%20version" />
    <a href="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml"><img alt="Test and Build" src="https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-test-build.yml/badge.svg?branch=main" /></a>
    <a href="https://goreportcard.com/badge/github.com/umatare5/cisco-wnc-exporter"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/umatare5/cisco-wnc-exporter" /></a>
    <a href="https://pkg.go.dev/github.com/umatare5/cisco-wnc-exporter@main"><img alt="Go Reference" src="https://pkg.go.dev/badge/umatare5/cisco-wnc-exporter.svg" /></a>
    <a href="./LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" /></a>
  </p>

</div>

## Overview

This exporter allows a Prometheus to scrape metrics from [Cisco Catalyst 9800 Wireless Network Controller](https://www.cisco.com/site/us/en/products/networking/wireless/wireless-lan-controllers/catalyst-9800-series/index.html).

- 🔧 **Leverages RESTCONF** for more flexibility than traditional SNMP monitoring
- 🛡️ **Monitors config drift** to detect parameter changes and prevent modifications
- 📊 **Enables long-term observability** with extended metrics retention beyond appliances

> [!Important]
>
> Please enable RESTCONF and HTTPS on the C9800 before deploy this exporter and generate an access token. See [umatare5/cisco-ios-xe-wireless-go - 🚀 Quick Start](https://github.com/umatare5/cisco-ios-xe-wireless-go?tab=readme-ov-file#-quick-start) section.

## Quick Start

```bash
docker run -p 10040:10040 -e WNC_HOST -e WNC_TOKEN ghcr.io/umatare5/cisco-wnc-exporter
```

- `-p`: Maps container port `10040/tcp` to host port `10040/tcp`.
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

GLOBAL OPTIONS:
   --dry-run                    Validate configuration without starting the server
   --help, -h                   show help
   --log.format string          Log format (json, text) (default: "json")
   --log.level string           Log level (debug, info, warn, error) (default: "info")
   --version, -v                print the version
   --web.listen-address string  Address to bind the HTTP server to (default: "0.0.0.0")
   --web.listen-port int        Port number to bind the HTTP server to (default: 10040)
   --web.telemetry-path string  Path for the metrics endpoint (default: "/metrics")
   --wnc.access-token string    WNC API access token [$WNC_TOKEN]
   --wnc.cache-ttl duration     WNC API response cache TTL (default: 55s)
   --wnc.controller string      WNC controller hostname or IP address [$WNC_HOST]
   --wnc.timeout duration       WNC API request timeout (default: 55s)
   --wnc.tls-skip-verify        Skip TLS certificate verification (default: false)

   # AP Collector Options

   --collector.ap.cpu                 Enable AP CPU usage metrics
   --collector.ap.errors              Enable AP error metrics
   --collector.ap.info                Enable AP info metrics
   --collector.ap.info-labels string  Comma-separated list of AP info labels (default: "name,ip")
   --collector.ap.inventory           Enable AP inventory metrics
   --collector.ap.memory              Enable AP memory usage metrics
   --collector.ap.phy                 Enable AP physical layer metrics
   --collector.ap.rf                  Enable AP RF environment metrics
   --collector.ap.state               Enable AP state metrics
   --collector.ap.traffic             Enable AP traffic metrics

   # Client Collector Options

   --collector.client.errors              Enable Client errors collector
   --collector.client.info                Enable Client info collector
   --collector.client.info-labels string  Comma-separated list of Client info labels (default: "name,ipv4")
   --collector.client.inventory           Enable Client inventory collector
   --collector.client.phy                 Enable Client PHY collector
   --collector.client.power               Enable Client power collector
   --collector.client.rf                  Enable Client RF collector
   --collector.client.session             Enable Client session collector
   --collector.client.traffic             Enable Client traffic collector

   # WLAN Collector Options

   --collector.wlan.info                Enable WLAN info metrics
   --collector.wlan.info-labels string  Comma-separated list of WLAN info labels (default: "name")
   --collector.wlan.inventory           Enable WLAN inventory collector
   --collector.wlan.networking          Enable WLAN networking collector
   --collector.wlan.security            Enable WLAN security collector
   --collector.wlan.state               Enable WLAN state collector
   --collector.wlan.traffic             Enable WLAN traffic collector

   # Collector Wide Options

   --collector.cache-ttl duration  Cache TTL for collector metrics (default: 30m0s)

   # Internal Collector Options

   --collector.internal.go-runtime  Enable Go runtime metrics collector
   --collector.internal.process     Enable process metrics collector
```

## Configuration

This exporter supports following environment variables:

| Environment Variable | Description                                      |
| :------------------- | :----------------------------------------------- |
| `WNC_HOST`           | WNC controller hostname or IP address (required) |
| `WNC_TOKEN`          | WNC API access token (required)                  |

## Metrics

This exporter collects wireless network metrics from Cisco C9800 WNC using following collectors:

- **AP Collector** - For RF foundation and radio performance
- **Client Collector** - For user experience quality and connection performance
- **WLAN Collector** - For logical SSID performance and parameter checks

All collectors have multiple modules to allow fine-grained control over which metrics to collect.

> [!Important]
>
> All collectors are **disabled by default** to minimize load on both Prometheus and Cisco C9800 WNC. Cisco C9800 WNC typically manage hundreds or thousands of APs and clients, making selective monitoring essential for both Prometheus and Cisco C9800 WNC performance and stability. Addition to that:
>
> - All WNC API calls execute **sequentially** to minimize concurrent load on the controller
> - API responses cache for **55 seconds** to reduce repeated requests from multiple Prometheus
>   - Cache TTL optimizes for standard Prometheus scrape and AP metrics reporting intervals (60s)
>   - Use `--wnc.cache-ttl` flag to adjust cache TTL (default: 55s)
> - **Info metrics cache for 30 minutes** to mitigate cardinality explosion from client roaming
>   - Wireless clients frequently roam between access points, causing `ap` label values to change
>   - Use `--collector.cache-ttl` flag to adjust info cache TTL (default: 30m0s)

### AP Collector

AP collector focus on RF foundation and radio performance.

| Module    | Metric                                | Type    | Description                                      |
| :-------- | :------------------------------------ | :------ | :----------------------------------------------- |
| inventory | `wnc_ap_count`                        | Gauge   | Total access points                              |
| state     | `wnc_ap_admin_state`                  | Gauge   | Admin state (0=disabled, 1=enabled)              |
| state     | `wnc_ap_oper_state`                   | Gauge   | Operational state (0=down, 1=up)                 |
| state     | `wnc_ap_uptime_seconds`               | Gauge   | AP uptime in seconds                             |
| state     | `wnc_ap_radio_state`                  | Gauge   | Radio state (0=down, 1=up)                       |
| phy       | `wnc_ap_channel_number`               | Gauge   | Operating channel number                         |
| phy       | `wnc_ap_channel_width_mhz`            | Gauge   | Channel bandwidth (MHz)                          |
| phy       | `wnc_ap_tx_power_dbm`                 | Gauge   | Current transmit power (dBm)                     |
| phy       | `wnc_ap_tx_power_level`               | Gauge   | TX power level (0-8)                             |
| phy       | `wnc_ap_tx_power_max_dbm`             | Gauge   | Maximum TX power (dBm)                           |
| rf        | `wnc_ap_noise_floor_dbm`              | Gauge   | Channel noise floor (dBm)                        |
| rf        | `wnc_ap_channel_utilization_percent`  | Gauge   | Channel utilization percentage                   |
| traffic   | `wnc_ap_clients_total`                | Gauge   | Associated clients count (calclulated)           |
| traffic   | `wnc_ap_rx_bytes_total`               | Counter | Total received bytes (calculated)                |
| traffic   | `wnc_ap_tx_bytes_total`               | Counter | Total transmitted bytes (calculated)             |
| traffic   | `wnc_ap_rx_packets_total`             | Counter | Total received packets                           |
| traffic   | `wnc_ap_tx_packets_total`             | Counter | Total transmitted packets                        |
| traffic   | `wnc_ap_total_tx_frames_total`        | Counter | Total TX frames                                  |
| traffic   | `wnc_ap_data_rx_frames_total`         | Counter | Data RX frames                                   |
| traffic   | `wnc_ap_data_tx_frames_total`         | Counter | Data TX frames                                   |
| traffic   | `wnc_ap_management_rx_frames_total`   | Counter | Management RX frames                             |
| traffic   | `wnc_ap_management_tx_frames_total`   | Counter | Management TX frames                             |
| traffic   | `wnc_ap_control_rx_frames_total`      | Counter | Control RX frames **(\*1)**                      |
| traffic   | `wnc_ap_control_tx_frames_total`      | Counter | Control TX frames **(\*1)**                      |
| traffic   | `wnc_ap_multicast_rx_frames_total`    | Counter | Multicast RX frames **(\*1)**                    |
| traffic   | `wnc_ap_multicast_tx_frames_total`    | Counter | Multicast TX frames **(\*1)**                    |
| traffic   | `wnc_ap_rts_success_total`            | Counter | Successful RTS transmissions **(\*1)**           |
| errors    | `wnc_ap_tx_errors_total`              | Counter | Total TX errors **(\*1)**                        |
| errors    | `wnc_ap_rx_errors_total`              | Counter | Total RX errors **(\*1)**                        |
| errors    | `wnc_ap_tx_drops_total`               | Counter | Total TX drops (calculated)                      |
| errors    | `wnc_ap_tx_retries_total`             | Counter | Total TX retries (calculated)                    |
| errors    | `wnc_ap_transmission_failures_total`  | Counter | Failed transmission attempts **(\*1)** **(\*2)** |
| errors    | `wnc_ap_duplicate_frames_total`       | Counter | Duplicate frames received                        |
| errors    | `wnc_ap_fcs_errors_total`             | Counter | Frame Check Sequence errors                      |
| errors    | `wnc_ap_fragmentation_rx_total`       | Counter | RX fragmented packets **(\*1)**                  |
| errors    | `wnc_ap_fragmentation_tx_total`       | Counter | TX fragmented packets **(\*1)**                  |
| errors    | `wnc_ap_rts_failures_total`           | Counter | RTS failures **(\*1)**                           |
| errors    | `wnc_ap_decryption_errors_total`      | Counter | Decryption errors **(\*1)**                      |
| errors    | `wnc_ap_mic_errors_total`             | Counter | MIC errors **(\*1)**                             |
| errors    | `wnc_ap_wep_undecryptable_total`      | Counter | WEP undecryptable frames **(\*1)**               |
| errors    | `wnc_ap_coverage_hole_events_total`   | Counter | Coverage hole events                             |
| errors    | `wnc_ap_interference_events_total`    | Counter | Interference events                              |
| errors    | `wnc_ap_dfs_channel_changes_total`    | Counter | DFS channel change events                        |
| errors    | `wnc_ap_radar_detected_events_total`  | Counter | Radar detection events                           |
| errors    | `wnc_ap_radio_reset_total`            | Counter | Radio reset count                                |
| cpu       | `wnc_ap_cpu_usage_current_percent`    | Gauge   | Current CPU utilization %                        |
| cpu       | `wnc_ap_cpu_usage_average_percent`    | Gauge   | Average CPU utilization %                        |
| memory    | `wnc_ap_memory_usage_current_percent` | Gauge   | Current memory utilization %                     |
| memory    | `wnc_ap_memory_usage_average_percent` | Gauge   | Average memory utilization %                     |

<details>
<summary>*1 Metrics consistently returning zero values on Cisco IOS-XE 17.12.5 with FlexConnect AP</summary>

The following metrics consistently return zero values due to implementation limitations:

- `wnc_ap_control_rx_frames_total` / `wnc_ap_control_tx_frames_total`
- `wnc_ap_fragmentation_rx_total` / `wnc_ap_fragmentation_tx_total`
- `wnc_ap_multicast_rx_frames_total` / `wnc_ap_multicast_tx_frames_total`
- `wnc_ap_rts_success_total` / `wnc_ap_rts_failures_total`
- `wnc_ap_tx_errors_total` / `wnc_ap_rx_errors_total`
- `wnc_ap_transmission_failures_total`
- `wnc_ap_decryption_errors_total`
- `wnc_ap_mic_errors_total`
- `wnc_ap_wep_undecryptable_total`

This was verified through direct RESTCONF API access to the live WNC environment:

```bash
❯ curl -sS -k -H "Authorization: Basic $WNC_TOKEN" \
           -H "Accept: application/yang-data+json" \
           "https://$WNC_HOST/restconf/data/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-stats" \
           | jq '.["Cisco-IOS-XE-wireless-access-point-oper:radio-oper-stats"]'
[
  {
    "ap-mac": "aa:bb:cc:11:22:30",
    "slot-id": 0,
    "aid-user-list": 2,
    "tx-fragment-count": 0,
    "multicast-tx-frame-cnt": 0,
    "failed-count": 0,
    "retry-count": 79738,
    "multiple-retry-count": 0,
    "frame-duplicate-count": 0,
    "rts-success-count": 0,
    "rts-failure-count": 0,
    "ack-failure-count": 0,
    "rx-fragment-count": 0,
    "multicast-rx-frame-cnt": 0,
    "fcs-error-count": 324978,
    "tx-frame-count": 2457172,
    "wep-undecryptable-count": 0,
    "rx-error-frame-count": 0,
    "mac-mic-err-frame-count": 0,
    "rx-mgmt-frame-count": 2084196,
    "rx-ctrl-frame-count": 0,
    "rx-data-frame-count": 3121831,
    "tx-mgmt-frame-count": 2345422,
    "tx-ctrl-frame-count": 0,
    "tx-data-frame-count": 2457172,
    "rx-data-pkt-count": 0,
    "tx-data-pkt-count": 0,
    "noise-floor": 0,
    "ap-radio-stats": {
      "stuck-ts": "1970-01-01T00:00:00+00:00",
      "last-ts": "2025-09-23T10:25:19+00:00",
      "num-radio-stuck-reset": 0
    },
    "mac-decry-err-frame-count": 0
  },
  {
    "ap-mac": "aa:bb:cc:11:22:30",
    "slot-id": 1,
    "aid-user-list": 3,
    "tx-fragment-count": 0,
    "multicast-tx-frame-cnt": 0,
    "failed-count": 0,
    "retry-count": 216205,
    "multiple-retry-count": 0,
    "frame-duplicate-count": 0,
    "rts-success-count": 0,
    "rts-failure-count": 0,
    "ack-failure-count": 0,
    "rx-fragment-count": 0,
    "multicast-rx-frame-cnt": 0,
    "fcs-error-count": 15398676,
    "tx-frame-count": 17767278,
    "wep-undecryptable-count": 0,
    "rx-error-frame-count": 0,
    "mac-mic-err-frame-count": 0,
    "rx-mgmt-frame-count": 19155,
    "rx-ctrl-frame-count": 0,
    "rx-data-frame-count": 6552508,
    "tx-mgmt-frame-count": 4628753,
    "tx-ctrl-frame-count": 0,
    "tx-data-frame-count": 17767278,
    "rx-data-pkt-count": 0,
    "tx-data-pkt-count": 0,
    "noise-floor": 0,
    "ap-radio-stats": {
      "stuck-ts": "1970-01-01T00:00:00+00:00",
      "last-ts": "2025-09-23T10:25:19+00:00",
      "num-radio-stuck-reset": 0
    },
    "mac-decry-err-frame-count": 0
  },
  <snip>
]
```

</details>

<details>
<summary>*2 Cisco Bug CSCwn96363 - AckFailureCount vs FailedCount</summary>

According to [Cisco Bug CSCwn96363](https://bst.cloudapps.cisco.com/bugsearch/bug/CSCwn96363), there are redundant counters in the wireless statistics:

- **Issue**: `AckFailureCount` always returns 0 and does not increment
- **Root Cause**: `AckFailureCount` and `FailedCount` represent the same counter
- **Solution**: Use `FailedCount` instead of `AckFailureCount` for accurate transmission failure statistics

This exporter implements the recommended workaround by using `failed-count` from the RESTCONF API for the `wnc_ap_transmission_failures_total` metric.

</details>

> [!Tip]
>
> `info` module provides `wnc_ap_info` contains following labels to join with other metrics:
>
> | Labels       | Description             | Example Value       | Required | Default |
> | :----------- | :---------------------- | :------------------ | :------: | :-----: |
> | `mac`        | AP wireless MAC address | `aa:bb:cc:dd:ee:f0` | **Yes**  | **Yes** |
> | `name`       | AP hostname             | `TEST-AP01`         |    No    | **Yes** |
> | `ip`         | AP IP address           | `192.168.1.10`      |    No    | **Yes** |
> | `radio`      | Radio identifier        | `0`, `1`, `2`       | **Yes**  | **Yes** |
> | `band`       | Radio band              | `2.4`, `5`, `6`     |    No    |   No    |
> | `model`      | AP model                | `AIR-AP1815I-Q-K9`  |    No    |   No    |
> | `serial`     | AP serial number        | `FGL1234ABCD`       |    No    |   No    |
> | `sw_version` | Software version        | `17.12.5.41`        |    No    |   No    |
> | `eth_mac`    | Ethernet MAC address    | `aa:bb:cc:00:11:22` |    No    |   No    |
>
> Use this info metric to add contextual labels to other metrics in PromQL queries:
>
> ```bash
> wnc_ap_radio_admin_state * on(mac,radio) group_left(name,ip) wnc_ap_info{name="TEST-AP01"}
> ```

### Client Collector

Client collector focus on user experience quality and connection performance.

| Module    | Metric                                | Type    | Description                          |
| :-------- | :------------------------------------ | :------ | :----------------------------------- |
| inventory | `wnc_clients_total`                   | Gauge   | Total connected clients              |
| session   | `wnc_client_state`                    | Gauge   | Client state (0-2)                   |
| session   | `wnc_client_uptime_seconds`           | Gauge   | Connection duration                  |
| phy       | `wnc_client_protocol`                 | Gauge   | 802.11 protocol (5=ac, 6=ax)         |
| phy       | `wnc_client_mcs_index`                | Gauge   | MCS index (-1=legacy, 0-11)          |
| phy       | `wnc_client_spatial_streams`          | Gauge   | Spatial streams count                |
| phy       | `wnc_client_speed_mbps`               | Gauge   | Connection throughput                |
| rf        | `wnc_client_rssi_dbm`                 | Gauge   | Signal strength (dBm)                |
| rf        | `wnc_client_snr_decibels`             | Gauge   | Signal-to-noise ratio (dB)           |
| traffic   | `wnc_client_rx_bytes_total`           | Counter | Received bytes                       |
| traffic   | `wnc_client_tx_bytes_total`           | Counter | Transmitted bytes                    |
| traffic   | `wnc_client_rx_packets_total`         | Counter | Received packets                     |
| traffic   | `wnc_client_tx_packets_total`         | Counter | Transmitted packets                  |
| errors    | `wnc_client_retry_ratio_percent`      | Gauge   | Retry rate percentage                |
| errors    | `wnc_client_tx_retries_total`         | Counter | TX retries count **(\*3)**           |
| errors    | `wnc_client_data_retries_total`       | Counter | Data retries by mobile station       |
| errors    | `wnc_client_excessive_retries_total`  | Counter | Excessive retries count **(\*3)**    |
| errors    | `wnc_client_rts_retries_total`        | Counter | RTS retries count **(\*3)**          |
| errors    | `wnc_client_duplicate_received_total` | Counter | Duplicate packets received **(\*3)** |
| errors    | `wnc_client_tx_drops_total`           | Counter | TX drops count                       |
| errors    | `wnc_client_decryption_failed_total`  | Counter | Decryption failures                  |
| errors    | `wnc_client_mic_mismatch_total`       | Counter | MIC mismatch errors **(\*3)**        |
| errors    | `wnc_client_mic_missing_total`        | Counter | MIC missing errors **(\*3)**         |
| errors    | `wnc_client_policy_errors_total`      | Counter | Policy errors **(\*3)**              |
| errors    | `wnc_client_rx_group_counter_total`   | Counter | RX group counter **(\*3)**           |
| power     | `wnc_client_power_save_state`         | Gauge   | Power save state (0=active, 1=save)  |

<details>
<summary>*3 Client error metrics consistently returning zero values on Cisco IOS-XE 17.12.5 with FlexConnect AP</summary>

The following client error metrics consistently return zero values due to implementation limitations:

- `wnc_client_duplicate_received_total`
- `wnc_client_excessive_retries_total`
- `wnc_client_mic_mismatch_total`
- `wnc_client_mic_missing_total`
- `wnc_client_policy_errors_total`
- `wnc_client_rts_retries_total`
- `wnc_client_rx_group_counter_total`
- `wnc_client_tx_retries_total`

This was verified through direct RESTCONF API access to the live WNC environment:

```bash
❯ curl -sS -k -H "Authorization: Basic $WNC_TOKEN" \
           -H "Accept: application/yang-data+json" \
           "https://$WNC_HOST/restconf/data/Cisco-IOS-XE-wireless-client-oper:client-oper-data/traffic-stats" \
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
> | Labels     | Description                 | Example Value                 | Required | Default |
> | :--------- | :-------------------------- | :---------------------------- | :------: | :-----: |
> | `mac`      | MAC address                 | `aa:bb:cc:12:34:56`           | **Yes**  | **Yes** |
> | `ap`       | Access point identifier     | `TEST-AP01`                   |    No    |   No    |
> | `band`     | Radio band                  | `2.4`, `5`, `6`               |    No    |   No    |
> | `wlan`     | WLAN ESSID name             | `labo-wifi`                   |    No    |   No    |
> | `name`     | Device Classification Name  | `MacBook Pro (14-inch, 2021)` |    No    | **Yes** |
> | `username` | EAP authentication identity | `john.doe@example.com`        |    No    |   No    |
> | `ipv4`     | Client IPv4 address         | `192.168.1.100`               |    No    | **Yes** |
> | `ipv6`     | Client IPv6 address         | `2001:db8::1`                 |    No    |   No    |
>
> Use this info metric to add contextual labels to other metrics in PromQL queries:
>
> ```bash
> wnc_client_state * on(mac) group_left(ap,wlan,name) wnc_client_info{ap="TEST-AP01"}
> ```

### WLAN Collector

WLAN collector focus on logical SSID performance and parameter checks.

| Module     | Metric                                    | Type    | Description                                               |
| :--------- | :---------------------------------------- | :------ | :-------------------------------------------------------- |
| inventory  | `wnc_wlans_total`                         | Gauge   | Total configured WLANs                                    |
| state      | `wnc_wlan_enabled`                        | Gauge   | WLAN status<br>(0=disabled, 1=enabled)                    |
| traffic    | `wnc_wlan_clients_total`                  | Gauge   | Connected clients count<br>(calclulated by client)        |
| traffic    | `wnc_wlan_rx_bytes_total`                 | Counter | WLAN received bytes<br>(calclulated by client)            |
| traffic    | `wnc_wlan_tx_bytes_total`                 | Counter | WLAN transmitted bytes<br>(calclulated by client)         |
| security   | `wnc_wlan_auth_psk_enabled`               | Gauge   | PSK authentication enabled<br>(0=disabled, 1=enabled)     |
| security   | `wnc_wlan_auth_dot1x_enabled`             | Gauge   | 802.1x authentication enabled<br>(0=disabled, 1=enabled)  |
| security   | `wnc_wlan_auth_dot1x_sha256_enabled`      | Gauge   | 802.1x SHA256 auth enabled<br>(0=disabled, 1=enabled)     |
| security   | `wnc_wlan_wpa2_enabled`                   | Gauge   | WPA2 support enabled<br>(0=disabled, 1=enabled)           |
| security   | `wnc_wlan_wpa3_enabled`                   | Gauge   | WPA3 support enabled<br>(0=disabled, 1=enabled)           |
| security   | `wnc_wlan_session_timeout_seconds`        | Gauge   | Session timeout duration<br>in seconds                    |
| networking | `wnc_wlan_load_balance_enabled`           | Gauge   | Load balancing enabled<br>(0=disabled, 1=enabled)         |
| networking | `wnc_wlan_11k_neighbor_list_enabled`      | Gauge   | 802.11k neighbor list enabled<br>(0=disabled, 1=enabled)  |
| networking | `wnc_wlan_client_steering_enabled`        | Gauge   | 6GHz client steering enabled<br>(0=disabled, 1=enabled)   |
| networking | `wnc_wlan_central_switching_enabled`      | Gauge   | Central switching enabled<br>(0=disabled, 1=enabled)      |
| networking | `wnc_wlan_central_authentication_enabled` | Gauge   | Central authentication enabled<br>(0=disabled, 1=enabled) |
| networking | `wnc_wlan_central_dhcp_enabled`           | Gauge   | Central DHCP enabled<br>(0=disabled, 1=enabled)           |
| networking | `wnc_wlan_central_association_enabled`    | Gauge   | Central association enabled<br>(0=disabled, 1=enabled)    |

> [!Tip]
>
> `info` module provides `wnc_wlan_info` contains following labels to join with other metrics:
>
> | Labels | Description     | Example Value  | Required | Default |
> | :----- | :-------------- | :------------- | :------: | :-----: |
> | `id`   | WLAN identifier | `1`, `5`, `10` | **Yes**  | **Yes** |
> | `name` | WLAN ESSID name | `labo-wifi`    |    No    | **Yes** |
>
> Use this info metric to add contextual labels to other metrics in PromQL queries:
>
> ```bash
> wnc_wlan_enabled * on(id) group_left(name) wnc_wlan_info{name="labo-wifi"}
> ```

## Usage

There are multiple ways to run the exporter, including direct binary execution and Docker containerization.

### Exporter

Visit http://localhost:10040/ to verify the exporter is running.

#### Basic Usage - No Collectors

The exporter starts without any collectors enabled by default:

```bash
$ WNC_HOST="wnc1.example.internal"
$ WNC_TOKEN="foobarbaz"
$ ./cisco-wnc-exporter
time="2025-04-13T18:50:54Z" level=info msg="Starting the cisco-wnc-exporter on port 10040."
```

#### Essential Usage

Enable essential collectors for basic monitoring:

```bash
$ WNC_HOST="wnc1.example.internal"
$ WNC_TOKEN="foobarbaz"
$ ./cisco-wnc-exporter \
    --collector.ap.inventory --collector.ap.info --collector.ap.state \
    --collector.client.inventory --collector.client.info --collector.client.session \
    --collector.wlan.inventory --collector.wlan.info --collector.wlan.state
```

#### Complete Usage

For complete monitoring, see [`.air.toml`](./.air.toml) which enables all collectors with maximum info-labels.

#### Using Docker

All the above examples can be run with Docker by adding the appropriate flags:

```bash
$ WNC_HOST="wnc1.example.internal"
$ WNC_TOKEN="foobarbaz"
$ docker run -p 10040:10040 -e WNC_HOST -e WNC_TOKEN ghcr.io/umatare5/cisco-wnc-exporter \
    --collector.ap.inventory --collector.ap.info --collector.ap.state
```

## Development

### Build

The repository includes a ready to use `Dockerfile`. To build a new Docker image:

```bash
make image
```

This creates an image named `ghcr.io/$USER/cisco-wnc-exporter` and exposes `10040/tcp`.

### Release

To release a new version, follow these steps:

1. Update the version in the `VERSION` file.
2. Submit a pull request with the updated `VERSION` file.

Once the pull request is merged, the GitHub Workflow will automatically:

- Create and push a new tag based on the `VERSION` file.

After that, I will manually release using [GitHub Actions: release workflow](https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/release.yaml).

## Acknowledgement

I maintain this project with the help of **GitHub Copilot Coding Agent**, and I'm grateful to the global developer community for their contributions to open source projects and public repositories.

## Licence

[MIT](LICENSE)
