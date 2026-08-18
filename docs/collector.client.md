# Client collector

Client collector focuses on user experience quality and connection performance.

## Metrics

| Module  | Metric                                | Type    | Description                          |
| :------ | :------------------------------------ | :------ | :----------------------------------- |
| general | `wnc_client_state`                    | Gauge   | Connection state in `state` label    |
| general | `wnc_client_roam_type`                | Gauge   | Roam type in `state` **(\*4)**       |
| general | `wnc_client_state_transition_seconds` | Gauge   | State transition latency             |
| general | `wnc_client_power_save_state`         | Gauge   | Power save state **(\*1)**           |
| general | `wnc_client_uptime_seconds`           | Gauge   | Connection duration                  |
| radio   | `wnc_client_protocol`                 | Gauge   | 802.11 protocol (0=unknown, 1..7)    |
| radio   | `wnc_client_mcs_index`                | Gauge   | MCS index **(\*2)**                  |
| radio   | `wnc_client_spatial_streams`          | Gauge   | Spatial streams count                |
| radio   | `wnc_client_speed_mbps`               | Gauge   | Negotiated PHY rate, not throughput  |
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

## Labels

`info` module provides `wnc_client_info` contains following labels to join with other metrics:

| Labels     | Description                 | Example Value                 | Default | Required |
| :--------- | :-------------------------- | :---------------------------- | :-----: | :------: |
| `mac`      | MAC address                 | `aa:bb:cc:12:34:56`           | **Yes** | **Yes**  |
| `ap`       | Access point identifier     | `TEST-AP01`                   |   No    |    No    |
| `band`     | Radio band                  | `2.4`, `5`, `6`, `unknown`    |   No    |    No    |
| `wlan`     | WLAN ESSID name             | `labo-wifi`                   |   No    |    No    |
| `name`     | Device Classification Name  | `MacBook Pro (14-inch, 2021)` | **Yes** |    No    |
| `username` | EAP authentication identity | `john.doe@example.com`        |   No    |    No    |
| `ipv4`     | Client IPv4 address         | `192.168.1.100`               | **Yes** |    No    |
| `ipv6`     | Client IPv6 address         | `2001:db8::1`                 |   No    |    No    |

Use this info metric to add contextual labels to other metrics in PromQL queries:

```bash
wnc_client_state * on(mac) group_left(ap,wlan,name) wnc_client_info
```

The example joins `ap` and `wlan`, which are not in the default label set, so `--collector.client.info-labels` has to name them for those labels to carry a value.

`wnc_client_info` exists only for clients that reached `client-status-run`, so this join silently drops a client held short of it. Keep an alert on stuck clients join-free, as shown in [States](README.md#states).

## Notes

`wnc_client_speed_mbps` is the rate the client negotiated for the link, reported alongside the MCS index and the stream count it is derived from. It does not measure how much the client is sending or receiving, so it stays at its negotiated value on an idle client; for throughput use `rate()` over the byte counters in the `traffic` module.

`wnc_client_state_transition_seconds` reports the run latency the controller recorded for the association this client currently holds, taken from the first entry of its mobility history. It is a property of that association rather than a live measurement, so it does not move until the client associates again. It is withheld in the two shapes the controller uses to say it measured no transition — a mobility history with no entry, and a first entry whose latency reads zero — because a zero reports an instant transition, which is a measurement rather than the absence of one. It is also published only for a client in the run state, which is the state the transition ends at, so a transition that is slow, stuck or failed has no series here at all — use `wnc_client_state` for those, as shown in [States](README.md#a-state-is-a-label-not-a-number).

`wnc_client_uptime_seconds` is withheld for a client whose record carries no association time, and for the epoch the controller writes where an event has not happened — the same rule the AP uptime and the AP timestamp series follow. Measuring from either would report a session centuries long, so absence is the ordinary reading for a record the controller has not filled in rather than a fault.

No series here counts roams. The controller maintains a roam count for itself rather than per client, and the exporter publishes it on the [Controller](collector.controller.md) page. The controller reports two per-client roam-type leaves and they are not the same reading: `wnc_client_roam_type` publishes `dot11-roam-type` from the association the client currently holds, while the mobility manager's own `mm-client-roam-type` stays unpublished because it read the same value for every client while the controller counted tens of thousands of roams. Neither is a count.

Recomputing a retry rate needs both `--collector.client.errors` and `--collector.client.traffic`. `wnc_client_data_retries_total` and `wnc_client_tx_retries_total` come from the errors module while `wnc_client_tx_packets_total` comes from the traffic module, and the ratio series removed in v0.3.0 sat in the errors module alone. Both flags default off.

<details><summary><b>*1</b> Power save state value domain</summary><br/>

The exporter decodes this leaf as an integer and publishes it unchanged, so a fractional reading would fail the fetch that carries it rather than arrive rounded. Zero was observed, and so was a non-zero reading of at most 1 whose exact value the measurement did not record. The full domain is not documented, so treat any other value as a state this exporter has not seen rather than as an error.

</details>

<details><summary><b>*2</b> MCS index range, and what -1 means</summary><br/>

The index is parsed out of the rate string the controller reports for the client, which spells it as `m<index>` followed by the stream count. The value is not bounded at 11:

- 802.11n encodes the stream count in the index itself, so a two-stream client reports 8 through 15, and observed values already exceed 11.
- 802.11ac indexes 0 through 9 and 802.11ax 0 through 11, both with the stream count in a separate leaf, so the same number means a different rate depending on the protocol. Read this metric together with `wnc_client_protocol` and `wnc_client_spatial_streams`.
- 802.11be adds indexes 12 and 13 for the standard rate set.

`-1` is reported whenever no index can be parsed. That covers a legacy client whose rate carries none, and equally a rate string that is empty or spelled in a form the parser does not recognise. The two are not distinguished.

</details>

<details><summary><b>*3</b> Client error metrics observed to stay at zero on the access points this exporter was measured against</summary><br/>

The client error metrics below were observed at zero on every client of the access points this exporter was measured against. Whether a counter is maintained depends on the access point model and the release, so read the list as an observation rather than as a property of the platform. That applies while the fetch succeeds: a data type whose fetch failed makes its series absent rather than zero.

- `wnc_client_duplicate_received_total`
- `wnc_client_excessive_retries_total`
- `wnc_client_mic_mismatch_total`
- `wnc_client_mic_missing_total`
- `wnc_client_policy_errors_total`
- `wnc_client_rts_retries_total`
- `wnc_client_rx_group_total`
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

<details><summary><b>*4</b> The roam type belongs to the current association, and it is not a count</summary><br/>

`wnc_client_roam_type` reports how the client reached the association it currently holds, in a `state` label with the value `1` — see [States](README.md#a-state-is-a-label-not-a-number). It is read from the first entry of the mobility history, the record `wnc_client_state_transition_seconds` also reads, so it adds no request and both are absent while that fetch fails.

It is a property of that association rather than a count, so it does not move until the client associates again and a client that has roamed many times reports one spelling. The spelling is the controller's own and passes through unmapped.

Two shapes withhold it: a mobility history with no entry, and an entry whose roam type the controller left empty, because an empty label reads as no label at all. It is published for a client in the run state only, like the rest of this module, so a client held short of that state has no series here — `wnc_client_state` covers those.

</details>
