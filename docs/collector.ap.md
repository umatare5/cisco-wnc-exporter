# AP collector

AP collector focuses on RF foundation and radio performance.

## Metrics

| Module  | Metric                                | Type    | Description                                      |
| :------ | :------------------------------------ | :------ | :----------------------------------------------- |
| general | `wnc_ap_admin_state`                  | Gauge   | Admin state (0=disabled, 1=enabled)              |
| general | `wnc_ap_oper_state`                   | Gauge   | Operational state in `state` label               |
| general | `wnc_ap_radio_state`                  | Gauge   | Radio state (0=down, 1=up)                       |
| general | `wnc_ap_config_state`                 | Gauge   | Tag config state (0=valid, 1=invalid)            |
| general | `wnc_ap_uptime_seconds`               | Gauge   | AP uptime in seconds                             |
| general | `wnc_ap_cpu_utilization_ratio`        | Gauge   | CPU utilization ratio (0-1)                      |
| general | `wnc_ap_memory_utilization_ratio`     | Gauge   | Memory utilization ratio (0-1)                   |
| radio   | `wnc_ap_channel_number`               | Gauge   | Operating channel number **(\*4)**               |
| radio   | `wnc_ap_channel_width_mhz`            | Gauge   | Channel bandwidth (MHz)                          |
| radio   | `wnc_ap_tx_power_dbm`                 | Gauge   | Current transmit power (dBm)                     |
| radio   | `wnc_ap_tx_power_max_dbm`             | Gauge   | Maximum TX power capability (dBm)                |
| radio   | `wnc_ap_noise_floor_dbm`              | Gauge   | Noise on the operating channel (dBm)             |
| radio   | `wnc_ap_channel_utilization_ratio`    | Gauge   | Channel utilization ratio (CCA), 0-1             |
| radio   | `wnc_ap_rx_utilization_ratio`         | Gauge   | RX utilization ratio (0-1)                       |
| radio   | `wnc_ap_tx_utilization_ratio`         | Gauge   | TX utilization ratio (0-1)                       |
| radio   | `wnc_ap_noise_utilization_ratio`      | Gauge   | Noise channel utilization ratio (0-1)            |
| radio   | `wnc_ap_clients`                      | Gauge   | Run-state clients count (calculated)             |
| traffic | `wnc_ap_total_tx_frames_total`        | Counter | TX frames, not a sum of the frame series         |
| traffic | `wnc_ap_data_rx_frames_total`         | Counter | Data RX frames                                   |
| traffic | `wnc_ap_data_tx_frames_total`         | Counter | Data TX frames                                   |
| traffic | `wnc_ap_management_rx_frames_total`   | Counter | Management RX frames                             |
| traffic | `wnc_ap_management_tx_frames_total`   | Counter | Management TX frames                             |
| traffic | `wnc_ap_control_rx_frames_total`      | Counter | Control RX frames **(\*1)**                      |
| traffic | `wnc_ap_control_tx_frames_total`      | Counter | Control TX frames **(\*1)**                      |
| traffic | `wnc_ap_multicast_rx_frames_total`    | Counter | Multicast RX frames **(\*1)**                    |
| traffic | `wnc_ap_multicast_tx_frames_total`    | Counter | Multicast TX frames **(\*1)**                    |
| traffic | `wnc_ap_rts_success_total`            | Counter | Successful RTS transmissions **(\*1)**           |
| errors  | `wnc_ap_rx_errors_total`              | Counter | Total RX errors **(\*1)**                        |
| errors  | `wnc_ap_tx_retries_total`             | Counter | Total TX retries (calculated)                    |
| errors  | `wnc_ap_transmission_failures_total`  | Counter | Failed transmission attempts **(\*1)** **(\*2)** |
| errors  | `wnc_ap_duplicate_frames_total`       | Counter | Duplicate frames received                        |
| errors  | `wnc_ap_fcs_errors_total`             | Counter | Frame Check Sequence errors                      |
| errors  | `wnc_ap_fragmentation_rx_total`       | Counter | RX fragmented packets **(\*1)**                  |
| errors  | `wnc_ap_fragmentation_tx_total`       | Counter | TX fragmented packets **(\*1)**                  |
| errors  | `wnc_ap_rts_failures_total`           | Counter | RTS failures **(\*1)**                           |
| errors  | `wnc_ap_decryption_errors_total`      | Counter | Decryption errors **(\*1)**                      |
| errors  | `wnc_ap_mic_errors_total`             | Counter | MIC errors **(\*1)**                             |
| errors  | `wnc_ap_wep_undecryptable_total`      | Counter | WEP undecryptable frames **(\*1)**               |
| errors  | `wnc_ap_coverage_failed_clients`      | Gauge   | Clients failing the RRM coverage check           |
| errors  | `wnc_ap_last_radar_timestamp_seconds` | Gauge   | Last radar detection unix timestamp              |
| errors  | `wnc_ap_radio_reset_total`            | Counter | Radio reset count                                |

## Labels

`info` module provides `wnc_ap_info` contains following labels to join with other metrics:

| Labels       | Description             | Example Value              | Default | Required |
| :----------- | :---------------------- | :------------------------- | :-----: | :------: |
| `mac`        | AP wireless MAC address | `aa:bb:cc:dd:ee:f0`        | **Yes** | **Yes**  |
| `name`       | AP hostname             | `TEST-AP01`                | **Yes** |    No    |
| `ip`         | AP IP address           | `192.168.1.10`             | **Yes** |    No    |
| `radio`      | Radio identifier        | `0`, `1`, `2`              | **Yes** | **Yes**  |
| `band`       | Radio band              | `2.4`, `5`, `6`, `unknown` |   No    |    No    |
| `model`      | AP model                | `AIR-AP1815I-Q-K9`         |   No    |    No    |
| `serial`     | AP serial number        | `FGL1234ABCD`              |   No    |    No    |
| `sw_version` | Software version        | `17.12.5.41`               |   No    |    No    |
| `eth_mac`    | Ethernet MAC address    | `aa:bb:cc:00:11:22`        |   No    |    No    |

Use this info metric to add contextual labels to other metrics in PromQL queries:

```bash
wnc_ap_radio_state * on(mac,radio) group_left(name,ip) wnc_ap_info
```

`radio` is a required label, so this metric carries one series per radio. Joining a series that has no `radio` label fails with a duplicate match group instead of returning nothing, so collapse the info metric first:

```bash
wnc_ap_uptime_seconds * on(mac) group_left(name) max by (mac,name) (wnc_ap_info)
```

## Notes

<details><summary><b>*1</b> Metrics consistently returning zero values on Cisco IOS-XE 17.12.6a with FlexConnect AP</summary><br/>

The following metrics consistently return zero values due to implementation limitations. That applies while the fetch succeeds: a data type whose fetch failed makes its series absent rather than zero.

- `wnc_ap_control_(rx|tx)_frames_total`
- `wnc_ap_decryption_errors_total`
- `wnc_ap_fragmentation_(rx|tx)_total`
- `wnc_ap_mic_errors_total`
- `wnc_ap_multicast_(rx|tx)_frames_total`
- `wnc_ap_rts_(success|failures)_total`
- `wnc_ap_rx_errors_total`
- `wnc_ap_transmission_failures_total`
- `wnc_ap_wep_undecryptable_total`

This was verified through direct RESTCONF API access to the live WNC environment:

```bash
❯ curl -sS -k -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
           -H "Accept: application/yang-data+json" \
           "https://$WNC_CONTROLLER/restconf/data/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-stats" \
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

<details><summary><b>*2</b> Cisco Bug CSCwn96363 - AckFailureCount vs FailedCount</summary><br/>

According to [Cisco Bug CSCwn96363](https://bst.cloudapps.cisco.com/bugsearch/bug/CSCwn96363), there are redundant counters in the wireless statistics:

- **Issue**: `AckFailureCount` always returns 0 and does not increment
- **Root Cause**: `AckFailureCount` and `FailedCount` represent the same counter
- **Solution**: Use `FailedCount` instead of `AckFailureCount` for accurate transmission failure statistics

This exporter implements the recommended workaround by using `failed-count` from the RESTCONF API for the `wnc_ap_transmission_failures_total` metric.

</details>

<details><summary><b>*4</b> Channel numbers do not identify the band on their own</summary><br/>

6 GHz channel numbering restarts at 1, so a 6 GHz channel number collides with a 2.4 GHz one and overlaps the 5 GHz range as well. `wnc_ap_channel_number` reports the number the controller gives, without a band.

Join `wnc_ap_info` to disambiguate, which requires `band` in `--collector.ap.info-labels` because it is not enabled by default:

```bash
wnc_ap_channel_number * on(mac,radio) group_left(band) wnc_ap_info
```

The `radio` label is not a substitute. A dual band radio keeps its slot while it moves between bands.

`wnc_ap_noise_floor_dbm` is the noise the controller measured on that same channel. The controller reports noise per channel across the whole band, so it is selected by matching the radio's operating channel; the series is absent when no entry matches it, which is the case for a radio in monitor or sniffer mode.

</details>
