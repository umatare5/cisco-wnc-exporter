# Client collector

Client collector focuses on user experience quality and connection performance.

## Metrics

| Module  | Metric                                | Type    | Description                          |
| :------ | :------------------------------------ | :------ | :----------------------------------- |
| general | `wnc_client_state`                    | Gauge   | Connection state in `state` label    |
| general | `wnc_client_state_transition_seconds` | Gauge   | State transition latency             |
| general | `wnc_client_power_save_state`         | Gauge   | Power save state **(\*6)**           |
| radio   | `wnc_client_protocol`                 | Gauge   | 802.11 protocol (0=unknown, 1..7)    |
| general | `wnc_client_uptime_seconds`           | Gauge   | Connection duration                  |
| radio   | `wnc_client_mcs_index`                | Gauge   | MCS index **(\*5)**                  |
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

`wnc_client_info` exists only for clients that reached `client-status-run`, so
this join silently drops a client held short of it. Keep an alert on stuck
clients join-free, as shown in [States](README.md#states).

## Notes

<details><summary><b>*3</b> Client error metrics observed to stay at zero on the access points this exporter was measured against</summary><br/>

The client error metrics below were observed at zero on every client of the access
points this exporter was measured against. Whether a counter is maintained depends on
the access point model and the release, so read the list as an observation rather than
as a property of the platform. That applies while the fetch succeeds: a data type whose
fetch failed makes its series absent rather than zero.

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

<details><summary><b>*5</b> MCS index range, and what -1 means</summary><br/>

The index is parsed out of the rate string the controller reports for the client, which
spells it as `m<index>` followed by the stream count. The value is not bounded at 11:

- 802.11n encodes the stream count in the index itself, so a two-stream client reports
  8 through 15, and observed values already exceed 11.
- 802.11ac indexes 0 through 9 and 802.11ax 0 through 11, both with the stream count in
  a separate leaf, so the same number means a different rate depending on the protocol.
  Read this metric together with `wnc_client_protocol` and `wnc_client_spatial_streams`.
- 802.11be adds indexes 12 and 13 for the standard rate set.

`-1` is reported whenever no index can be parsed. That covers a legacy client whose
rate carries none, and equally a rate string that is empty or spelled in a form the
parser does not recognise; the two are not distinguished.

</details>

<details><summary><b>*6</b> Power save state value domain</summary><br/>

The controller reports this leaf as an integer and the exporter publishes it unchanged.
Values of 0 and 1 have been observed, with 0 on a client that was awake. The full domain
is not documented, so treat a value above 1 as a state this exporter has not seen rather
than as an error.

</details>
