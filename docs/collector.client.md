# Client collector

Client collector focuses on user experience quality and connection performance.

## Metrics

| Module  | Metric                                | Type    | Description                         |
| :------ | :------------------------------------ | :------ | :---------------------------------- |
| general | `wnc_client_state`                    | Gauge   | Connection state (11=run state)     |
| general | `wnc_client_roam_type`                | Gauge   | Roam type                           |
| general | `wnc_client_state_transition_seconds` | Gauge   | State transition latency            |
| general | `wnc_client_power_save_state`         | Gauge   | Power save state                    |
| general | `wnc_client_uptime_seconds`           | Gauge   | Connection duration                 |
| radio   | `wnc_client_protocol`                 | Gauge   | 802.11 protocol (0=unknown, 1..7)   |
| radio   | `wnc_client_mcs_index`                | Gauge   | MCS index                           |
| radio   | `wnc_client_spatial_streams`          | Gauge   | Spatial streams count               |
| radio   | `wnc_client_speed_mbps`               | Gauge   | Negotiated PHY rate, not throughput |
| radio   | `wnc_client_rssi_dbm`                 | Gauge   | Signal strength (dBm)               |
| radio   | `wnc_client_snr_decibels`             | Gauge   | Signal-to-noise ratio (dB)          |
| traffic | `wnc_client_rx_bytes_total`           | Counter | Received bytes                      |
| traffic | `wnc_client_tx_bytes_total`           | Counter | Transmitted bytes                   |
| traffic | `wnc_client_rx_packets_total`         | Counter | Received packets                    |
| traffic | `wnc_client_tx_packets_total`         | Counter | Transmitted packets                 |
| errors  | `wnc_client_tx_retries_total`         | Counter | TX retries count                    |
| errors  | `wnc_client_data_retries_total`       | Counter | Data retries by mobile station      |
| errors  | `wnc_client_excessive_retries_total`  | Counter | Excessive retries count             |
| errors  | `wnc_client_rts_retries_total`        | Counter | RTS retries count                   |
| errors  | `wnc_client_duplicate_received_total` | Counter | Duplicate packets received          |
| errors  | `wnc_client_tx_drops_total`           | Counter | TX drops count                      |
| errors  | `wnc_client_decryption_failed_total`  | Counter | Decryption failures                 |
| errors  | `wnc_client_mic_mismatch_total`       | Counter | MIC mismatch errors                 |
| errors  | `wnc_client_mic_missing_total`        | Counter | MIC missing errors                  |
| errors  | `wnc_client_policy_errors_total`      | Counter | Policy errors                       |
| errors  | `wnc_client_rx_group_total`           | Counter | RX group counter                    |

## Specifications

Each entry carries what the series' HELP text and the shared [Absence](README.md#absence) rules do not.

**`wnc_client_state`**

- Every other series in this collector needs the client in `client-status-run`, so a client whose onboarding is slow, stuck or failed carries this one alone.

**`wnc_client_roam_type`**

- Read from the first entry of the mobility history, the record `wnc_client_state_transition_seconds` also reads, so it adds no request and a failed fetch or a history with no entry takes both series away at once.
- No series in this collector counts roams — the controller keeps its roam counters for itself rather than per client, and the [Controller](collector.controller.md) page publishes the three it maintains.

**`wnc_client_state_transition_seconds`**

- Belongs to the association the client currently holds, read from that same first entry, so it does not move until the client associates again.
- Withheld in the two shapes the controller uses to say it measured no transition — a mobility history with no entry, and a first entry whose latency reads zero, which would report an instant transition rather than none.

**`wnc_client_power_save_state`**

- Zero was observed, and one non-zero reading of at most 1, with no documented domain behind either — read any other value as a state this exporter has not seen rather than as an error.

**`wnc_client_mcs_index`**

- Parsed out of the rate string the controller spells as `m<index>` followed by the stream count, so `-1` covers a legacy rate carrying no index, an empty string and a spelling the parser does not recognise alike.
- Not bounded at 11, and readings above it were observed, so pair it with `wnc_client_protocol` and `wnc_client_spatial_streams` — the index alone fixes neither the protocol's rate table nor whether the stream count is already inside it.

**`wnc_client_speed_mbps`**

- Holds the rate the client negotiated for the link, so an idle client keeps it — throughput takes `rate()` over the byte counters in the `traffic` module.

**`wnc_client_tx_retries_total` and `wnc_client_data_retries_total`**

- A retry rate over either also needs `wnc_client_tx_packets_total` from the `traffic` module, so recomputing one takes both `--collector.client.errors` and `--collector.client.traffic`, and both default off.

**The eight `errors` counters below**

- Read zero on every client of the access points this exporter was measured against, and the access point model and the release decide whether each is maintained, so take the list as an observation rather than as a property of the platform.

  | Metric                                | Leaf                   |
  | :------------------------------------ | :--------------------- |
  | `wnc_client_duplicate_received_total` | `duplicate-rcv`        |
  | `wnc_client_excessive_retries_total`  | `tx-excessive-retries` |
  | `wnc_client_mic_mismatch_total`       | `mic-mismatch`         |
  | `wnc_client_mic_missing_total`        | `mic-missing`          |
  | `wnc_client_policy_errors_total`      | `policy-errs`          |
  | `wnc_client_rts_retries_total`        | `rts-retries`          |
  | `wnc_client_rx_group_total`           | `rx-group-counter`     |
  | `wnc_client_tx_retries_total`         | `tx-retries`           |

## Info Labels

The `info` module publishes `wnc_client_info` with the following labels to join with other metrics:

| Labels        | Description                 | Example Value                 | Default | Required |
| :------------ | :-------------------------- | :---------------------------- | :-----: | :------: |
| `mac`         | MAC address                 | `aa:bb:cc:12:34:56`           | **Yes** | **Yes**  |
| `ap`          | Access point identifier     | `TEST-AP01`                   |   No    |    No    |
| `band`        | Radio band                  | `2.4`, `5`, `6`, `unknown`    |   No    |    No    |
| `wlan`        | WLAN ESSID name             | `labo-wifi`                   |   No    |    No    |
| `wlan_id`     | WLAN identifier             | `5`                           |   No    |    No    |
| `name`        | Device Classification Name  | `MacBook Pro (14-inch, 2021)` | **Yes** |    No    |
| `device_type` | Device Classification Type  | `Un-Classified Device`        |   No    |    No    |
| `username`    | EAP authentication identity | `john.doe@example.com`        |   No    |    No    |
| `ipv4`        | Client IPv4 address         | `192.168.1.100`               | **Yes** |    No    |
| `ipv6`        | Client IPv6 address         | `2001:db8::1`                 |   No    |    No    |

Use this info metric to add contextual labels to other metrics in PromQL queries:

```bash
wnc_client_state * on(mac) group_left(ap,wlan,name) wnc_client_info
```

The example above names `ap` and `wlan`, which the default label set omits, so `--collector.client.info-labels` has to carry both for those labels to hold a value.

> [!NOTE]
>
> ### About the Labels
>
> **`wlan_id`:** Reads the same leaf `wnc_wlan_clients` buckets by, so `count by (wlan_id) (wnc_client_info)` reconciles with `wnc_wlan_clients{id}`, and it is empty rather than `0` where the controller omitted the identifier.
>
> **`device_type`:** The controller's own classification rather than something the client states, so a reclassification leaves two `wnc_client_info` series for one `mac` and a `group_left` over them fails with a duplicate-series error.
>
> **`band`:** Reads the PHY generation the client associated on rather than the band its AP radio reports, and `unknown` covers a generation naming no band, a wired client for instance, as well as a spelling this exporter does not map.
>
> **`ap`:** Carries that same hazard whenever a client roams rather than when the controller reclassifies it, so naming both labels puts two that churn on one series.
>
> ```bash
> wnc_client_info * on(wlan_id) group_left(name) label_replace(wnc_wlan_info, "wlan_id", "$1", "id", "(.*)")
> ```
