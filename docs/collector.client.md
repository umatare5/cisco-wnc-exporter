# Client Collector

Client collector reports per-client connection quality, from association state to PHY rate.

## Metrics

Each row registers only behind the `--collector.client.*` flag its `Group` column names.

| Group   | Metric                                | Type    | Description                         |
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
| info    | `wnc_client_info`                     | Gauge   | Client identity labels, always 1    |

## Labels

Every data series above carries `mac` and nothing else. A reading identifies a client and no AP, band or WLAN, and a join with `wnc_client_info` adds them.

The `--collector.client.info` flag publishes `wnc_client_info`. The join below yields nothing without it. The `--collector.client.info-labels` flag selects which of the following labels it carries:

| Label         | Description                 | Example Value                 |
| :------------ | :-------------------------- | :---------------------------- |
| `mac`         | MAC address                 | `aa:bb:cc:12:34:56`           |
| `ap`          | Access point identifier     | `TEST-AP01`                   |
| `band`        | Radio band                  | `2.4`, `5`, `6`, `unknown`    |
| `wlan`        | WLAN ESSID name             | `labo-wifi`                   |
| `wlan_id`     | WLAN identifier             | `5`                           |
| `name`        | Device classification name  | `MacBook Pro (14-inch, 2021)` |
| `device_type` | Device classification type  | `Un-Classified Device`        |
| `username`    | EAP authentication identity | `john.doe@example.com`        |
| `ipv4`        | Client IPv4 address         | `192.168.1.100`               |
| `ipv6`        | Client IPv6 address         | `2001:db8::1`                 |

The `mac` label is on `wnc_client_info` whatever `--collector.client.info-labels` names. It is the join key every other series in this collector offers.

The flag selects from `ap`, `band`, `wlan`, `wlan_id`, `name`, `device_type`, `username`, `ipv4` and `ipv6`. It defaults to `name,ipv4`. A label it omits is missing from the series rather than present and empty.

```bash
wnc_client_state * on(mac) group_left(ap,wlan,name) wnc_client_info
```

The join names `ap` and `wlan`. The default set omits them, so the flag has to carry both.

The `wlan_id` label is an open integer string, read off the same record `wnc_client_state` is built from. It is the same leaf `wnc_wlan_clients` buckets by.

The empty string is reachable and reserved for an identifier the controller omitted. No WLAN carries `0`.

The `device_type` label is open and copied verbatim from the controller's own classification record. It reads `""` where no such record names the client. It changes when the controller reclassifies rather than when the device does.

The `band` label is the PHY generation the client associated on rather than the band its AP radio reports. The `unknown` value covers a generation naming no band, such as a wired client. It also covers a spelling this exporter does not map.

## Annotations

Each entry carries what the series' HELP text and the shared [Absence](architecture.md#absence) rules do not.

**`wnc_client_state`**

Every other series in this collector needs the client in `client-status-run`. A client whose onboarding is slow, stuck or failed carries this series alone.

**`wnc_client_roam_type`**

It is read from the first entry of the mobility history, which is the record `wnc_client_state_transition_seconds` also reads. It therefore adds no request. A failed fetch or a history with no entry takes both series away at once.

No series in this collector counts roams. The controller keeps its roam counters for itself rather than per client. The [Controller](collector.controller.md) page publishes the three counters it maintains.

**`wnc_client_state_transition_seconds`**

It belongs to the association the client currently holds, read from that same first entry. It does not move until the client associates again.

It is withheld in the two shapes the controller uses to say it measured no transition. The first is a mobility history with no entry. The second is a first entry whose latency reads zero, which would report an instant transition rather than none.

**`wnc_client_power_save_state`**

Zero was observed, and one non-zero reading of at most 1, with no documented domain behind either. Read any other value as a state this exporter has not seen rather than as an error.

**`wnc_client_mcs_index`**

It is parsed out of the rate string the controller spells as `m<index>` followed by the stream count. A `-1` value covers a legacy rate carrying no index, an empty string, and a spelling the parser does not recognise.

It is not bounded at 11, and readings above it were observed. Pair it with `wnc_client_protocol` and `wnc_client_spatial_streams`. The index alone fixes neither the protocol's rate table nor whether the stream count is already inside it.

**`wnc_client_speed_mbps`**

It holds the rate the client negotiated for the link, so an idle client keeps it. Throughput requires taking `rate()` over the byte counters in the `traffic` group.

**`wnc_client_tx_retries_total and wnc_client_data_retries_total`**

A retry rate over either series also needs `wnc_client_tx_packets_total` from the `traffic` group. Recomputing one takes both `--collector.client.errors` and `--collector.client.traffic`. [Help](help.md#flags) records these flags as off by default.

**`Eight errors counters`**

They read zero on every client of the access points this exporter was measured against. This is an observation of those access points rather than a property of the platform. [Update Schedule](architecture.md#update-schedule) explains what decides whether a counter leaf is maintained.

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

## Technical Notes

**Label Churn**: The `ap` label follows the association. Naming it beside `device_type` puts two labels that churn on their own schedules onto one series. Each of them can leave a `mac` holding two series.

A `group_left` over them fails with a duplicate match rather than returning nothing. A reclassification breaks the query outright instead of thinning its result.

**WLAN Info Join**: Joining `wnc_wlan_info` takes a `label_replace` first, because it spells the identifier `id`:

```bash
wnc_client_info * on(wlan_id) group_left(name) label_replace(wnc_wlan_info, "wlan_id", "$1", "id", "(.*)")
```

> [!NOTE]
> The `wnc_client_info` series is served from the info snapshot while `wnc_wlan_clients` is collected on the scrape itself. The two reconcile only over the WLANs holding at least one run-state client, and only while that snapshot is current. The query `count by (wlan_id) (wnc_client_info)` carries no series where `wnc_wlan_clients{id}` reads `0`. It diverges from it for up to `--collector.info-cache-ttl`. [Info Caching](architecture.md#info-caching) explains what the cache holds and its operational costs.
