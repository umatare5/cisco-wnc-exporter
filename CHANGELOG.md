# Changelog

Notable changes to the metric surface, one section per release. Release dates, downloads and the full commit list live on the [releases page](https://github.com/umatare5/cisco-wnc-exporter/releases).

This project is pre-1.0, so a minor release may rename or remove a metric. Read the section for the version you are upgrading to before you upgrade.

## v0.3.0

### Removed

These series are gone with no replacement.

| Metric                               | Reason                                                                         |
| :----------------------------------- | :----------------------------------------------------------------------------- |
| `wnc_ap_rx_bytes_total`              | Frame counter times a constant 1500; the frame counters remain                 |
| `wnc_ap_tx_bytes_total`              | Same                                                                           |
| `wnc_ap_rx_packets_total`            | Read the same leaf as the data-frame counter                                   |
| `wnc_ap_tx_packets_total`            | Same                                                                           |
| `wnc_ap_tx_drops_total`              | Read `ack-failure-count`, which Cisco bug CSCwn96363 says not to use           |
| `wnc_ap_tx_errors_total`             | Read `failed-count`, already published as `wnc_ap_transmission_failures_total` |
| `wnc_ap_wep_undecryptable_total`     | Cannot leave zero unless a WLAN is configured for static WEP                   |
| `wnc_client_retry_ratio_percent`     | A quotient of two counters the controller reports no retry rate for            |
| `wnc_wlan_rx_bytes_total`            | Summing a varying member set does not yield a counter                          |
| `wnc_wlan_tx_bytes_total`            | Same                                                                           |
| `wnc_wlan_wpa2_enabled`              | Reported `0` on WLANs whose operative value is enabled — see the note below    |
| `wnc_wlan_11k_neighbor_list_enabled` | Same                                                                           |

### Renamed and rescaled

Values are now ratios in `0-1`. Divide any threshold written for the old `0-100` percentages by 100.

| v0.2.0                               | v0.3.0                             |
| :----------------------------------- | :--------------------------------- |
| `wnc_ap_channel_utilization_percent` | `wnc_ap_channel_utilization_ratio` |
| `wnc_ap_cpu_utilization_percent`     | `wnc_ap_cpu_utilization_ratio`     |
| `wnc_ap_memory_utilization_percent`  | `wnc_ap_memory_utilization_ratio`  |
| `wnc_ap_noise_utilization_percent`   | `wnc_ap_noise_utilization_ratio`   |
| `wnc_ap_rx_utilization_percent`      | `wnc_ap_rx_utilization_ratio`      |
| `wnc_ap_tx_utilization_percent`      | `wnc_ap_tx_utilization_ratio`      |

### Renamed

| v0.2.0                              | v0.3.0                                             |
| :---------------------------------- | :------------------------------------------------- |
| `wnc_ap_clients_total`              | `wnc_ap_clients`                                   |
| `wnc_ap_coverage_hole_events_total` | `wnc_ap_coverage_failed_clients`, now a Gauge      |
| `wnc_ap_last_radar_on_radio_at`     | `wnc_ap_last_radar_timestamp_seconds`, now a Gauge |
| `wnc_client_rx_group_counter_total` | `wnc_client_rx_group_total`                        |
| `wnc_wlan_clients_total`            | `wnc_wlan_clients`                                 |

### Same name, new meaning

`wnc_ap_oper_state` and `wnc_client_state` are always `1` and carry the state in a `state` label. An alert such as `wnc_ap_oper_state == 0` keeps returning samples and silently stops firing.

- `wnc_ap_oper_state` also drops its `radio` label and is one series per AP, tracking the AP registration state. Per-radio down detection is `wnc_ap_radio_state == 0`, which is unchanged.
- Rewrite label matches as `wnc_ap_oper_state{state!="registered"}` and `wnc_client_state{state!="client-status-run"}`.

### Notes

The two WLAN config booleans were removed because the controller omits config leaves left at their default values, and an omitted leaf decodes to `0`. Both series therefore reported `0` on WLANs where the feature was in fact enabled. The remaining `config` series carry the same caveat, now stated in their HELP text and in [docs/collector.wlan.md](docs/collector.wlan.md).

Binaries are published as release assets. `go install` on the module root has never worked, because the main package lives in `cmd/`.
