# AP collector

AP collector focuses on RF foundation and radio performance.

## Metrics

| Module      | Metric                                            | Type    | Description                                                 |
| :---------- | :------------------------------------------------ | :------ | :---------------------------------------------------------- |
| general     | `wnc_ap_admin_state`                              | Gauge   | Admin state, absent if unreported **(\*12)**                |
| general     | `wnc_ap_oper_state`                               | Gauge   | Operational state (4=registered)                            |
| general     | `wnc_ap_radio_state`                              | Gauge   | Radio state, absent if unreported **(\*12)**                |
| general     | `wnc_ap_config_state`                             | Gauge   | Tag config state, absent if unreported (0=valid, 1=invalid) |
| general     | `wnc_ap_uptime_seconds`                           | Gauge   | AP uptime in seconds, absent without boot time              |
| general     | `wnc_ap_association_uptime_seconds`               | Gauge   | Age of the current association **(\*14)**                   |
| general     | `wnc_ap_cpu_utilization_ratio`                    | Gauge   | CPU utilization ratio (0-1) **(\*1)**                       |
| general     | `wnc_ap_memory_utilization_ratio`                 | Gauge   | Memory utilization ratio (0-1) **(\*1)**                    |
| radio       | `wnc_ap_channel_number`                           | Gauge   | Operating channel number **(\*2)**                          |
| radio       | `wnc_ap_channel_width_mhz`                        | Gauge   | Channel bandwidth (MHz)                                     |
| radio       | `wnc_ap_tx_power_dbm`                             | Gauge   | Current transmit power (dBm)                                |
| radio       | `wnc_ap_tx_power_max_dbm`                         | Gauge   | Maximum TX power capability (dBm)                           |
| radio       | `wnc_ap_noise_floor_dbm`                          | Gauge   | Noise on the operating channel (dBm) **(\*2)**              |
| radio       | `wnc_ap_channel_utilization_ratio`                | Gauge   | Channel utilization ratio (CCA), 0-1                        |
| radio       | `wnc_ap_rx_utilization_ratio`                     | Gauge   | RX utilization ratio (0-1) **(\*3)**                        |
| radio       | `wnc_ap_tx_utilization_ratio`                     | Gauge   | TX utilization ratio (0-1)                                  |
| radio       | `wnc_ap_noise_utilization_ratio`                  | Gauge   | Noise channel utilization ratio (0-1)                       |
| radio       | `wnc_ap_clients`                                  | Gauge   | Run-state clients count (calculated)                        |
| radio       | `wnc_ap_rrm_profile_passed`                       | Gauge   | RRM profile verdict per `profile` **(\*4)**                 |
| radio       | `wnc_ap_channel_changes_total`                    | Counter | Channel changes, DCA statistics **(\*4)**                   |
| radio       | `wnc_ap_channel_energy_dbm`                       | Gauge   | Energy DCA measured on the channel **(\*13)**               |
| radio       | `wnc_rrm_last_rf_grouping_run_timestamp_seconds`  | Gauge   | Last RF grouping run per band **(\*16)**                    |
| radio       | `wnc_rrm_last_dca_run_timestamp_seconds`          | Gauge   | Last DCA run per band **(\*16)**                            |
| traffic     | `wnc_ap_total_tx_frames_total`                    | Counter | TX frames, not a sum of the frame series **(\*17)**         |
| traffic     | `wnc_ap_data_rx_frames_total`                     | Counter | Data RX frames **(\*17)**                                   |
| traffic     | `wnc_ap_data_tx_frames_total`                     | Counter | Data TX frames **(\*17)**                                   |
| traffic     | `wnc_ap_management_rx_frames_total`               | Counter | Management RX frames **(\*17)**                             |
| traffic     | `wnc_ap_management_tx_frames_total`               | Counter | Management TX frames **(\*17)**                             |
| traffic     | `wnc_ap_control_rx_frames_total`                  | Counter | Control RX frames **(\*3)** **(\*17)**                      |
| traffic     | `wnc_ap_control_tx_frames_total`                  | Counter | Control TX frames **(\*3)** **(\*17)**                      |
| traffic     | `wnc_ap_multicast_rx_frames_total`                | Counter | Multicast RX frames **(\*3)** **(\*17)**                    |
| traffic     | `wnc_ap_multicast_tx_frames_total`                | Counter | Multicast TX frames **(\*3)** **(\*17)**                    |
| traffic     | `wnc_ap_rts_successes_total`                      | Counter | Successful RTS transmissions **(\*3)** **(\*17)**           |
| errors      | `wnc_ap_rx_errors_total`                          | Counter | Total RX errors **(\*3)** **(\*17)**                        |
| errors      | `wnc_ap_tx_retries_total`                         | Counter | Total TX retries **(\*17)**                                 |
| errors      | `wnc_ap_transmission_failures_total`              | Counter | Failed transmission attempts **(\*3)** **(\*5)** **(\*17)** |
| errors      | `wnc_ap_duplicate_frames_total`                   | Counter | Duplicate frames received **(\*17)**                        |
| errors      | `wnc_ap_fcs_errors_total`                         | Counter | Frame Check Sequence errors **(\*17)**                      |
| errors      | `wnc_ap_rx_fragments_total`                       | Counter | RX fragments **(\*3)** **(\*17)**                           |
| errors      | `wnc_ap_tx_fragments_total`                       | Counter | TX fragments **(\*3)** **(\*17)**                           |
| errors      | `wnc_ap_rts_failures_total`                       | Counter | RTS failures **(\*3)** **(\*17)**                           |
| errors      | `wnc_ap_decryption_errors_total`                  | Counter | Decryption errors **(\*3)** **(\*17)**                      |
| errors      | `wnc_ap_mic_errors_total`                         | Counter | MIC errors **(\*3)** **(\*17)**                             |
| errors      | `wnc_ap_coverage_failed_clients`                  | Gauge   | Clients failing the RRM coverage check                      |
| errors      | `wnc_ap_last_radar_timestamp_seconds`             | Gauge   | Unix timestamp of the last radar **(\*6)**                  |
| errors      | `wnc_ap_radio_resets_total`                       | Counter | Radio reset count **(\*18)**                                |
| join        | `wnc_ap_joined`                                   | Gauge   | CAPWAP session held now (0=no, 1=yes) **(\*7)**             |
| join        | `wnc_ap_join_info`                                | Gauge   | AP name from the join record, always 1                      |
| join        | `wnc_ap_discovery_requests_total`                 | Counter | CAPWAP discovery requests received                          |
| join        | `wnc_ap_discovery_responses_total`                | Counter | Successful discovery responses sent                         |
| join        | `wnc_ap_discovery_errors_total`                   | Counter | Discovery requests found in error                           |
| join        | `wnc_ap_join_requests_total`                      | Counter | CAPWAP join requests received                               |
| join        | `wnc_ap_join_responses_total`                     | Counter | Successful join responses sent                              |
| join        | `wnc_ap_join_failures_total`                      | Counter | Join requests that failed to process                        |
| join        | `wnc_ap_config_requests_total`                    | Counter | Configuration requests received                             |
| join        | `wnc_ap_config_responses_total`                   | Counter | Successful configuration responses sent                     |
| join        | `wnc_ap_config_failures_total`                    | Counter | Configuration requests that failed                          |
| join        | `wnc_ap_dtls_session_requests_total`              | Counter | DTLS setup requests, per `channel` **(\*8)**                |
| join        | `wnc_ap_dtls_session_successes_total`             | Counter | DTLS sessions established, per `channel`                    |
| join        | `wnc_ap_dtls_session_failures_total`              | Counter | DTLS sessions that failed, per `channel`                    |
| join        | `wnc_ap_dtls_decrypt_errors_total`                | Counter | DTLS decrypt errors, per `channel`                          |
| join        | `wnc_ap_dtls_anti_replay_errors_total`            | Counter | DTLS anti-replay errors, per `channel`                      |
| join        | `wnc_ap_last_error_timestamp_seconds`             | Gauge   | Last connection error **(\*9)**                             |
| join        | `wnc_ap_last_join_success_timestamp_seconds`      | Gauge   | Last successful join **(\*9)**                              |
| join        | `wnc_ap_last_join_failure_timestamp_seconds`      | Gauge   | Last failed join **(\*9)**                                  |
| join        | `wnc_ap_last_config_success_timestamp_seconds`    | Gauge   | Last successful configuration **(\*9)**                     |
| join        | `wnc_ap_last_config_failure_timestamp_seconds`    | Gauge   | Last failed configuration **(\*9)**                         |
| join        | `wnc_ap_last_discovery_success_timestamp_seconds` | Gauge   | Last successful discovery **(\*9)**                         |
| join        | `wnc_ap_last_discovery_failure_timestamp_seconds` | Gauge   | Last failed discovery **(\*9)**                             |
| join        | `wnc_ap_last_dtls_success_timestamp_seconds`      | Gauge   | Last DTLS session, per `channel` **(\*9)**                  |
| join        | `wnc_ap_last_dtls_failure_timestamp_seconds`      | Gauge   | Last failed DTLS, per `channel` **(\*9)**                   |
| join        | `wnc_ap_last_discovery_failure_reason`            | Gauge   | Discovery failure reason **(\*10)**                         |
| join        | `wnc_ap_last_join_failure_reason`                 | Gauge   | Join failure reason **(\*10)**                              |
| join        | `wnc_ap_last_config_failure_reason`               | Gauge   | Configuration failure reason **(\*10)**                     |
| join        | `wnc_ap_last_error_phase`                         | Gauge   | Phase of the last error **(\*10)**                          |
| join        | `wnc_ap_last_dtls_failure_reason`                 | Gauge   | DTLS outcome per `channel` **(\*10)**                       |
| join        | `wnc_ap_last_reboot_reason`                       | Gauge   | Reboot reason **(\*10)**                                    |
| join        | `wnc_ap_last_disconnect_reason`                   | Gauge   | Disconnect reason **(\*10)**                                |
| geolocation | `wnc_ap_longitude_degrees`                        | Gauge   | WGS 84 longitude of the AP **(\*19)**                       |
| geolocation | `wnc_ap_latitude_degrees`                         | Gauge   | WGS 84 latitude of the AP **(\*19)**                        |
| spectrum    | `wnc_ap_air_quality_index_avg`                    | Gauge   | CleanAir air quality of the channel **(\*11)**              |
| spectrum    | `wnc_ap_air_quality_index_min`                    | Gauge   | CleanAir air quality minimum **(\*11)**                     |
| spectrum    | `wnc_ap_interferers`                              | Gauge   | Interference devices on that channel **(\*11)**             |
| spectrum    | `wnc_ap_last_air_quality_timestamp_seconds`       | Gauge   | Instant stamped on the air quality row **(\*11)**           |
| spectrum    | `wnc_rrm_worst_channel_air_quality_index_avg`     | Gauge   | Worst channel air quality per band **(\*15)**               |
| spectrum    | `wnc_rrm_worst_channel_air_quality_index_min`     | Gauge   | Worst channel minimum per band **(\*15)**                   |
| spectrum    | `wnc_rrm_worst_channel_interferers`               | Gauge   | Interference devices on that channel **(\*15)**             |
| spectrum    | `wnc_rrm_worst_channel_number`                    | Gauge   | Which channel that is, as a value **(\*15)**                |

## Labels

`info` module provides `wnc_ap_info` contains following labels to join with other metrics:

| Labels       | Description             | Example Value              | Default | Required |
| :----------- | :---------------------- | :------------------------- | :-----: | :------: |
| `mac`        | AP wireless MAC address | `aa:bb:cc:dd:ee:f0`        | **Yes** | **Yes**  |
| `name`       | AP hostname             | `TEST-AP01`                | **Yes** | No       |
| `ip`         | AP IP address           | `192.168.1.10`             | **Yes** | No       |
| `radio`      | Radio identifier        | `0`, `1`, `2`              | **Yes** | **Yes**  |
| `band`       | Radio band              | `2.4`, `5`, `6`, `unknown` | No      | No       |
| `model`      | AP model                | `AIR-AP1815I-Q-K9`         | No      | No       |
| `serial`     | AP serial number        | `FGL1234ABCD`              | No      | No       |
| `sw_version` | Software version        | `17.12.5.41`               | No      | No       |
| `eth_mac`    | Ethernet MAC address    | `aa:bb:cc:00:11:22`        | No      | No       |

Use this info metric to add contextual labels to other metrics in PromQL queries:

```bash
wnc_ap_radio_state * on(mac,radio) group_left(name,ip) wnc_ap_info
```

`radio` is a required label, so this metric carries one series per radio. Joining a series that has no `radio` label fails with a duplicate match group instead of returning nothing, so collapse the info metric first:

```bash
wnc_ap_uptime_seconds * on(mac) group_left(name) max by (mac,name) (wnc_ap_info)
```

## Notes

`wnc_ap_clients` counts the clients the controller reports in the run state, attributed to a radio through the AP **name** the client record carries, because that record carries no AP MAC. A client whose AP name has no entry in `ap-name-mac-map` is therefore left out silently, and the series is withheld for every radio when either the client list or that mapping fails to fetch. `sum(wnc_ap_clients)` can consequently read lower than `sum(wnc_wlan_clients)`, which counts the same clients without needing the mapping.

<details><summary><b>*1</b> CPU and memory need AP system monitoring enabled, and read zero until it is</summary><br/>

Both values come from the AP system statistics the controller collects, and that collection is disabled by default. Enable it on the AP join profile the access point uses:

```plaintext
configure terminal
 ap profile <profile-name>
  statistics ap-system-monitoring enable
 end
write memory
```

While it is disabled the statistics block is still present in the controller's response and both leaves read zero, so the exporter publishes `0` rather than omitting the series. A zero therefore cannot be told apart from an idle access point. Both leaves were observed at zero on every access point before the collection was enabled, and reporting a value afterwards.

</details>

<details><summary><b>*2</b> Channel numbers do not identify the band, and noise is reported per channel</summary><br/>

6 GHz channel numbering restarts at 1, so a 6 GHz channel number collides with a 2.4 GHz one and overlaps the 5 GHz range as well. `wnc_ap_channel_number` reports the number the controller gives, without a band.

Join `wnc_ap_info` to disambiguate, which requires `band` in `--collector.ap.info-labels` because it is not enabled by default:

```bash
wnc_ap_channel_number * on(mac,radio) group_left(band) wnc_ap_info
```

The `radio` label is not a substitute. A dual band radio keeps its slot while it moves between bands.

`wnc_ap_noise_floor_dbm` is the noise the controller measured on that same channel. The controller reports noise per channel across the whole band, so it is selected by matching the radio's operating channel — the series is absent when no entry matches it, which is the case for a radio in monitor or sniffer mode.

</details>

<details><summary><b>*3</b> Metrics observed to stay at zero on the AP models this exporter was measured against</summary><br/>

The metrics below were observed at zero on every radio of the access points this exporter was measured against, while their neighbours in the same container were not. That applies while the fetch succeeds: a data type whose fetch failed makes its series absent rather than zero.

**Whether a leaf is maintained depends on the access point model and the release.** One model reported FCS errors while another returned zero for them on every band, and the reverse held for multicast transmit frames. Read the list as an observation, not as a property of the platform, and confirm it against your own access points before building an alert on the absence of a value.

| Metric                                   | What the zero means here                                                                                                   |
| :--------------------------------------- | :------------------------------------------------------------------------------------------------------------------------- |
| `wnc_ap_rx_utilization_ratio`            | Observed at zero while channel and noise utilization on the same radio read non-zero. Cause not established.               |
| `wnc_ap_control_(rx\|tx)_frames_total`   | Observed at zero while data and management frames advanced.                                                                |
| `wnc_ap_multicast_(rx\|tx)_frames_total` | Receive observed at zero; transmit advanced on one model and not on another.                                               |
| `wnc_ap_rx_errors_total`                 | Observed at zero while FCS errors advanced on the same radio.                                                              |
| `wnc_ap_transmission_failures_total`     | Observed at zero while retries advanced. See note *5.                                                                      |
| `wnc_ap_duplicate_frames_total`          | Observed at zero. A duplicate is counted on receive, so client retransmissions drive it.                                   |
| `wnc_ap_rts_(successes\|failures)_total` | The RTS threshold sits at its maximum, so length-triggered RTS never happens.                                              |
| `wnc_ap_(rx\|tx)_fragments_total`        | The fragmentation threshold sits at its maximum. A report that the receive side counts incomplete fragments is unverified. |
| `wnc_ap_decryption_errors_total`         | Zero is the healthy reading. Whether the counter would report a failure has not been confirmed.                            |
| `wnc_ap_mic_errors_total`                | Zero is the healthy reading, with the same caveat.                                                                         |

Sampling the container twice separated by an interval showed the same leaves at zero while their neighbours advanced, and the controller CLI reported the same values, so the zeros are in the data the controller holds rather than in this exporter.

The container also carries `rx-data-pkt-count` and `tx-data-pkt-count`, which no series reads. Both were observed at zero on every radio measured, through RESTCONF and through the CLI, including radios whose data-frame counters advanced over the same interval. Re-pointing a series at these leaves would publish a constant zero on a radio carrying traffic.

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

<details><summary><b>*4</b> Reading the RRM profile verdicts and the channel-change counter</summary><br/>

The controller judges each radio against four profiles and reports one verdict leaf per profile, so the four `profile` values — `coverage`, `load`, `interference` and `noise` — are this exporter's own names for those leaves. `1` is a pass.

**A failure is not by itself an incident.** Three of the four have been observed failing in ordinary conditions — `interference`, `load` and `noise` — and only `coverage` never has. The controller's own CLI agrees verdict for verdict on every radio checked, printing `Passed` and `Failed` where these series read `1` and `0`. A rule of the form `wnc_ap_rrm_profile_passed == 0` therefore fires from the first scrape. Two shapes do not. A verdict stuck failed for half an hour:

```bash
max_over_time(wnc_ap_rrm_profile_passed[30m]) == 0
```

Or several verdicts failing at once on one radio, which is the shape that survives a profile flapping:

```bash
count by (job, instance, mac, radio) (wnc_ap_rrm_profile_passed{profile!="load"} == 0) >= 2
```

`load` is excluded because its verdict is the controller's judgement against a client-count threshold configured on the controller, so a radio carrying a crowd fails it while nothing about the RF is wrong. Read capacity from `wnc_ap_clients` and `wnc_ap_channel_utilization_ratio` instead. The recipe above keeps every profile on purpose: it reports which one is failing in the `profile` label, so nothing is conflated there — the exclusion matters only where verdicts are counted together.

The thresholds each profile is judged against are configured on the controller and are not read here, so the series says a profile failed and never by how much. `wnc_ap_channel_utilization_ratio` and `wnc_ap_noise_floor_dbm` are the measured quantities behind the `load` and `noise` verdicts, and `wnc_ap_coverage_failed_clients`, which the `errors` module publishes, counts the clients behind the `coverage` one.

The four series are absent for a radio the slot list has no record for, for a record that carries no radio data, for a verdict leaf the controller omits from a record it did send, and for every radio while the `rrm_radio_slot` fetch fails. Each verdict goes absent on its own, so three siblings reading `1` is no evidence that the fourth was measured. **The controller fabricates a verdict of its own, though.** Just after a radio re-joined, all four verdicts were present and reading `1` while the channel energy the controller reports for that radio still read its unmeasured sentinel, and the radio had been failing interference before the re-join and failed it again afterwards — so the controller serves verdicts it has not yet measured, and the fabrication is the controller's rather than this exporter's decode. Read a verdict taken shortly after a re-join as unmeasured in either direction.

`wnc_ap_channel_changes_total` comes from the same record and adds no request. **It is read from the controller's DCA assignment statistics, and what it counts is not established.** The controller's CLI prints the same count under a DCA heading and keeps a separate count of radar-driven changes, which no leaf carries, so whether a radar move also advances this count could not be measured — both read zero on every radio here. `wnc_ap_last_radar_timestamp_seconds`, which the `errors` module publishes, dates a DFS event, so where it is absent no radar is on record and a move was not radar-driven. That inference is the most the pair supports, and on a controller that has seen no radar it is absent everywhere — see note *6. The counter was monotonic non-decreasing on every radio across fifteen consecutive reads two minutes apart, with one radio observed stepping by one as its best channel moved — and it was later observed **falling to zero on the one radio that had a count to lose**, with the controller's boot time unchanged, while every radio already reading zero stayed there. **The anchor is the radio's own AP joining CAPWAP rather than a controller-wide rebuild of the statistics tables** — an access point left untouched through the same interval kept its counters and its anchor, so a reset reaching one AP does not reach another's series, as [Counter reset timing](README.md#counter-reset-timing) describes. `rate()` and `increase()` absorb that; a rule reading the raw value as a lifetime total does not. Keep the range well above the controller's channel-assignment interval, or a healthy radio and a churning one both read zero. It sits one container deeper than the verdicts and is absent on its own when the controller reports no assignment statistics for a radio, which a zero would misreport as a radio that has never moved.

</details>

<details><summary><b>*5</b> Why <code>failed-count</code> is read and <code>ack-failure-count</code> is not</summary><br/>

[Cisco Bug CSCwn96363](https://bst.cloudapps.cisco.com/bugsearch/bug/CSCwn96363) reports `ack-failure-count` as a counter that never increments and calls it redundant with `failed-count`. That record is the vendor's and this repository has not read it, so the choice below does not rest on it.

What was measured on every radio measured, through RESTCONF and through the controller CLI:

- `ack-failure-count` held at zero, including on radios whose `retry-count` advanced over the same interval.
- `failed-count` held at zero as well. A retry that eventually succeeds produces no terminal failure, so zero is also the reading a healthy radio gives.
- The CLI reports a per-radio transmit-drop counter that was non-zero where both leaves above read zero. Enumerating the whole access point operational tree found drop leaves only for the Ethernet interface, so **no series can read that counter**.

`wnc_ap_transmission_failures_total` reads `failed-count`, and nothing reads `ack-failure-count`. Read a zero on the surviving series as unconfirmed rather than as a healthy radio. That a maintained ack-failure counter would have advanced alongside the retries is an inference from the protocol, not a measurement.

</details>

<details><summary><b>*6</b> When the radar timestamp series is absent</summary><br/>

The series is published only for a radio whose last-radar leaf carries a real instant. A radio that has recorded no radar carries the epoch, which is withheld rather than published as a timestamp in 1970. The container carries a record for every radio, on every band, and every record measured carried the epoch — so on a controller that has seen no radar, no series is published at all, and absence is the ordinary reading rather than a fault. Use `time() - series` for the age, and treat absence as no radar on record.

</details>
<details><summary><b>*7</b> Why the join module exists, and the one signal it makes possible</summary><br/>

The join statistics list is keyed by the AP radio MAC and **keeps a record for an AP that has left CAPWAP**: a record was observed for an AP absent from the AP inventory for months, with its counters frozen and its join state reporting `0`. Every other AP series is read from the inventory and disappears with it, so before this module nothing distinguished an AP that had gone from a fetch that had failed.

The join state and the phase counters carry `mac` and nothing else, deliberately. A bare `and` requires both sides to carry identical label sets, so one extra label on either side would make the query below return an empty result rather than an error. The reason series carry `mac` alone as well, so only the DTLS series add `channel`, and an `and` written against one of those needs `on(mac)` or `ignoring(...)`:

```bash
rate(wnc_ap_discovery_requests_total[15m]) > 0 and wnc_ap_joined == 0
```

That is the signal this module exists for: **the AP reaches the controller and cannot complete a join.** It works because the discovery counters keep advancing while the session is gone — measured on an AP whose join, configuration and DTLS counters were frozen while its discovery counters were the highest of any AP on the controller, its last successful discovery minutes after its last successful join.

Three cautions on that query. An AP that holds this controller as its secondary, or that discovers by broadcast, DHCP option 43 or DNS, sends discovery requests here while joining another controller, so pair the rule with a `for:` longer than the rate window and exclude those APs by `mac`. A counter here is cumulative since the controller allocated the record, so `increase()` over a range shorter than the discovery interval reports nothing. And the query is a snapshot of the join state, so an AP flapping in and out of CAPWAP can be joined at every evaluation and never fire — `changes(wnc_ap_joined[1h]) > 2` catches that case instead.

`wnc_ap_join_info` carries the name because the AP inventory no longer does, which is why the name is a series of its own rather than a label on the other thirty-one. The join record is not the only place a departed AP is still named — the AP history list names one too, measured — but this module reads the join record, and that record is keyed by the same radio MAC every series here carries. Join it with `* on(mac) group_left(name) wnc_ap_join_info`, which needs no collapsing because it carries no `radio` label. Its name puts it in the info cache, and that cache is only wrapped around the collector when `--collector.ap.info` is enabled, so with the join module alone the name is read fresh on every scrape and with both modules enabled it is up to `--collector.info-cache-ttl` old.

The record set is wider than the inventory. Compare like with like to see that — `wnc_ap_info` is one series per radio, so the count to compare against is `count(count by (mac) (wnc_ap_info))`. What evicts a record is not established — leaving CAPWAP does not, and no controller reload has been observed against this list.

</details>

<details><summary><b>*8</b> The DTLS channel label, and what a zero on the data channel means</summary><br/>

The controller keeps one set of DTLS counters per CAPWAP tunnel channel in a single container, and the exporter folds the pair into one series carrying `channel`, valued `control` or `data`. This is the tunnel channel; the RF channel is `wnc_ap_channel_number`.

DTLS on the data channel is not enabled by default, and with it disabled every `channel="data"` counter read zero on every AP measured while the `control` counters advanced. Enabling it moved two of the five: the session request and the session success counters each advanced, and the failure, decrypt-error and anti-replay-error counters stayed at zero even then. So a zero on those three is **unconfirmed rather than not configured** — whether they can advance at all was not established here.

The same change published the `channel="data"` entry of `wnc_ap_last_dtls_success_timestamp_seconds`, which had been absent because the controller carried the epoch in that leaf. Disabling the data channel again restored neither state: the two counters held their count and the timestamp kept its instant. **Once a data-channel handshake has succeeded, that series set exists until the controller rebuilds the record**, so what you see here is a property of the AP join profile and of what has already run on it rather than of the release.

Confirm against your own AP join profile before alerting on any of these zeros.

</details>

<details><summary><b>*9</b> When a join timestamp series is absent</summary><br/>

The controller writes `1970-01-01T00:00:00+00:00` into a timestamp leaf for an event that has not happened, and that sentinel is withheld rather than published as an instant in 1970 — the same rule as note *6. On a controller where nothing has failed, the join, configuration and discovery failure timestamps and the DTLS failure timestamp are therefore absent for every AP, which is the ordinary reading rather than a fault. Use `time() - series` for the age.

</details>

<details><summary><b>*10</b> The reason series freeze with the record, and the controller misspells a value</summary><br/>

Each of these publishes the number the controller's own enumeration assigns the spelling it sent — [Enumeration values](enums.md) is the mapping and [States](README.md#a-state-is-a-number-not-a-label) is the query shape. They report the **last recorded** event rather than a current state, and they freeze with the record, so an AP that has left CAPWAP keeps reporting the reason it recorded while it was joined.

`wnc_ap_last_error_phase` is the sharpest example: an AP that is not joined reports the same `ap-con-failure-run` phase, `6`, as one that is, on every record measured. Read it as the phase of the last error, never as a health check.

On the controller measured, `wnc_ap_last_join_failure_reason` reported `jf-none`, `wnc_ap_last_config_failure_reason` `cf-none`, `wnc_ap_last_discovery_failure_reason` `disc-fail-none` and `wnc_ap_last_dtls_failure_reason` `dtls-hs-success` on every AP. Each of those four is the healthy sentinel of its own enumeration and each is numbered `0`, so `== 0` selects the healthy APs and `!= 0` is the alertable set.

**The recipe covers those four and no others.** `wnc_ap_last_error_phase` numbers an unknown phase `0` and `wnc_ap_last_disconnect_reason` numbers the enumeration's own unknown member `0`, so a `0` on either reports that the controller does not know rather than that nothing went wrong. `wnc_ap_last_reboot_reason` numbers `ap-reboot-reason-none` `0` and classifies a past cause rather than reporting health, as note \*14 describes.

**The disconnect reason enumeration declares its own unknown member misspelled, as `unkown`**, and that misspelling is on the wire. No query needs it now that the value is the reading — it is `0` — but [Enumeration values](enums.md) carries the spelling verbatim, because the lookup that resolves a reading has to match what the controller sends.

The record also carries two free-text leaves — a prose disconnect description and a message-decryption failure reason — and neither is published, because neither has a bounded value domain to number.

</details>

<details><summary><b>*11</b> Which radios report air quality, when it was stamped, and why one goes silent</summary><br/>

The controller publishes air quality per AP and band, and the reading here is the one for the channel the radio operates on. **The series is absent rather than zero wherever that reading cannot be reached**, chiefly an AP without CleanAir, a radio whose spectrum operation is down, a radio in monitor or sniffer mode with no primary channel, a slot that is not a radio at all and so has no channel to match — see note \*12 — and every radio while the `rrm_spectrum_aq_table` fetch fails. Silence therefore does not mean clean air — on the controller measured here most radios reported a reading and the rest did not, and which ones moved as APs joined.

One of those cases was narrowed. A radio whose spectrum operation reads down belonged, on every such radio measured, to an access point carrying Spectrum Intelligence rather than full CleanAir, with spectrum intelligence itself left disabled — **whether such a radio reports a reading once spectrum intelligence is enabled could not be measured**, because none measured had it enabled. Silence can also be early rather than settled: after an AP joins, a refresh can publish every other per-radio series for it while the air quality table still carries no row for it, so this reading appears later than the rest of the radio. And **a reading present is not a reading now** — across an AP reboot the controller served the value from before the reboot until the AP reported again, which is what `wnc_ap_last_air_quality_timestamp_seconds` makes visible. The controller keeps a second and unrelated air quality as well, `ap-sensor-air-quality`, which a sensor radio reports and no series here reads.

**The value is an average**, over the air quality reporting period the controller declares — `show ap dot11 <band> cleanair config` reports its length. No configuration leaf carries it, so the CLI is the only place to read it, and an operator can change it rather than relying on this page. **A higher index is cleaner**: the controller's own air-quality alarm threshold is a lower bound on the index, and so is the sensitivity its event-driven RRM acts on, though which of the two readings that sensitivity is compared against could not be read. The CLI names the same per-channel figures the average and the minimum, and the two matched this exporter's source leaves row for row on every channel checked. `wnc_ap_air_quality_index_min` publishes the minimum of the same reporting period and `wnc_ap_interferers` the number of interference devices CleanAir attributes to that channel; both read the row this average reads, so the three are published and withheld together. `wnc_ap_last_air_quality_timestamp_seconds` reads the same row and reports the instant the controller stamped on it, with one guard the readings do not have — an instant at the epoch is withheld — so that series alone can be absent for a radio the three readings publish for. The minimum carries a signal the average does not — it was measured well below the average during an interference event — which is why it is published rather than left as a note. None of the three is guarded on its value, so a zero from `wnc_ap_interferers` is the controller's reading rather than a fabricated one.

**The instant is stamped per AP rather than per radio.** The bands of one access point carried the same second while another access point's differed, so an AP reporting on several bands repeats one instant across them and a difference between two APs is the controller's own report boundary rather than a difference in when each radio measured. Over a window of about twelve minutes the instant did not advance on any row and neither did any reading, so **a frozen instant means a frozen reading**; the reporting period is longer than that window and its length is not established here. The epoch guard above is a guard rather than a state anyone has observed — no row carrying a real channel was seen with the epoch in it, in either pass.

The reading covers the **operating channel only**, whatever the channel width. Interference elsewhere in the same band does not move it: measured, an interferer drove five of the thirteen channels of one band to an average of 91 and a minimum of 43 while this series held at 98 on the 20 MHz radio of that band. A radio on a bonded channel likewise reports its primary and not the rest of its width.

The table is the largest of the RRM reads — some seven times the whole RRM slot list — and it grows with the number of CleanAir APs, which is why it has its own flag and is off by default. It also arrives last, so a refresh cut short by its deadline drops this series before any other.

</details>

<details><summary><b>*12</b> Why a slot can carry no state at all</summary><br/>

The slot list is not a list of radios. A remote-LAN port arrives as a slot whose type names it as such and whose state leaves the controller omits entirely, measured on a controller. Every per-radio reading is therefore **absent** for such a slot rather than reading `0` — the two state series, the `traffic` and `errors` counters and `wnc_ap_clients` alike. A `0` there would say "radio down, admin disabled, no traffic, no errors, no clients" of a port that is none of those, and the controller does send a counter record for such a slot with every counter in it zero, so the reading came from the controller rather than from a lookup that missed.

A rule that treated any of them as always present needs `absent()` or `or vector(0)`, and a `sum()` over a controller carrying such a port now reads lower.

</details>

<details><summary><b>*13</b> Channel energy comes from DCA, and two of its values are not measurements</summary><br/>

`wnc_ap_channel_energy_dbm` is the energy the controller measured on the channel it assigned the radio, read from the DCA assignment statistics — the record `wnc_ap_channel_changes_total` comes from, so it adds no request. It is not a CleanAir reading: it is reported for a radio the air quality table has no record for at all, so it reaches radios `wnc_ap_air_quality_index_avg` cannot, which is why it sits in the `radio` module rather than behind `--collector.ap.spectrum`.

**It is a step rather than a sample.** The controller recomputes it when DCA next runs for the band, so it holds its value between runs and a range shorter than the channel-assignment interval sees no change at all — the same caution note \*4 gives for `wnc_ap_channel_changes_total`. After a radio joins, the series is absent until that band's next run, and the wait is a property of the band rather than of the radio.

**Two values are withheld rather than published, because neither can be a measurement.** One is `-128`, the lower bound of the leaf's own signed type, which sits far below the thermal noise floor of any channel width this reports; the other is `0`, which sits far above the energies the controller assigns channels on and above the lower limit it declares for the assignment. The first is what a radio reads until DCA next runs for its band, measured with an untouched radio as a control. The second has not been observed, and it is guarded because the leaf is a plain integer: one the controller omits decodes to zero, and no absence guard would catch it.

The series is absent as well for a radio the slot list has no record for, for a record carrying no radio data or no DCA statistics, and for every radio while the `rrm_radio_slot` fetch fails.

</details>

<details><summary><b>*14</b> Association uptime, and why no series here measures an outage</summary><br/>

`wnc_ap_association_uptime_seconds` counts from the start of the CAPWAP association the AP currently holds, read from the join time in the record `wnc_ap_uptime_seconds` already fetches, so it adds no request. The controller prints the two under separate headings of its own, and they answer different questions: the pair diverges for an AP that re-joined without rebooting and agrees for one that joined straight after booting. Both are withheld rather than reported as `0` where the controller carries no instant this exporter can use, the epoch included — the rule every timestamp-derived series here follows.

**Neither series measures how long an AP was gone.** A reboot does not delete the AP's record: the controller replaces it in place, so a read taken right after one can still serve the association the AP held before it, and once the new association is in place the count starts from that instant. The outage leaves no mark in either series. Watch the re-join instead:

```bash
changes(wnc_ap_last_join_success_timestamp_seconds[1h]) > 0
```

Three cautions on the alternatives. Whether an AP that has gone silent leaves the inventory at all, and how long its record keeps the reading it had, was not the same on every model measured, so do not build an outage rule on the absence of these series. `wnc_ap_last_reboot_reason` classifies a cause and dates nothing: it reports the number the controller's enumeration assigns the cause, so two reboots recorded with the same reason cannot be told from one. And a state that lasts less than the interval between refreshes, which `--wnc.cache-ttl` sets, may never be sampled at all.

</details>

<details><summary><b>*15</b> The band-worst series are keyed by band alone, and one read carries all four</summary><br/>

The controller ranks the channels of a band across every AP that scans it and keeps one row per band for the worst of them; a higher index is cleaner, so the worst channel is the one with the lowest average. The four `wnc_rrm_worst_channel_*` series publish that row — the average, the minimum, the interference device count and the channel number — behind `--collector.ap.spectrum`, and they cost one read per refresh whatever the number of APs, because the table is per band rather than per AP.

The row is chosen on the average, so its minimum is the minimum of that one channel and not the deepest of the band: a channel with a better average can carry a worse minimum, and the per-radio pair is what reports the channels radios actually use.

**They carry `band` and nothing else, which is the whole identifier the controller gives the row.** Neither an AP nor a radio keys it, so none of these series can be joined with the per-AP series. The controller does name the AP that detected the channel, and that name is not published: it changes while the reading does not, and it could not be joined against `wnc_ap_info`, which is keyed by MAC.

`wnc_rrm_worst_channel_number` is a **value rather than a label**, because the worst channel of a band moves — a label would start a new series on every move and leave the previous one to go stale. Read the other three together with it: they describe whichever channel it names at that moment.

**A band this exporter cannot name is withheld as a whole row rather than published as `unknown`.** That is the exception to the rule on the [Overview](README.md#the-band-label) page, which every band-keyed `wnc_rrm_*` series shares, and it holds for two reasons: `band` is the whole identifier here, so an `unknown` row would name nothing to act on, and two such rows would carry the same label set, which fails the whole `/metrics` endpoint instead of one series. A row reporting no channel is withheld the same way — the controller returns that shape for a band it has not ranked, and a zero there cannot be told from a reading: it would report the worst possible air quality, no interference at all, and a channel that does not exist.

**The row set follows the controller's table rather than the radios that have joined**, so fewer bands than the controller supports can appear, and a band no CleanAir-capable radio scans never does. Treat one, two or three bands as ordinary, and read a band's absence as no ranking rather than as clean air.

These four are the only series behind the `rrm_spectrum_aq_worst_table` read, and that read is a failure path of its own. **A controller or an image that does not carry the table answers `404`, and a `404` is a failure rather than an absence** — the rule note \*4 on the [Controller](collector.controller.md) page describes, with the expression that excludes such a data type. There, `wnc_refresh_errors_total{data="rrm_spectrum_aq_worst_table"}` rises on every refresh, no `wnc_refresh_items` series appears for it, and all four families disappear from the scrape rather than reading zero, while `wnc_up` stays `1`. The error series is there either way — it is seeded at `0` for every data type the enabled modules read — so read its value and not its presence.

The reading is refreshed on the reporting period the controller declares for air quality and no leaf in the row dates it, so a value held between reports cannot be told from one just taken. `wnc_ap_last_air_quality_timestamp_seconds` dates the per-radio readings and nothing dates this one, and the range guidance on the [Overview](README.md#controller-side-update-schedule) page covers the rest.

</details>

<details><summary><b>*16</b> The RRM run instants are keyed by band, and the two need not advance together</summary><br/>

The controller runs its radio resource management per band rather than per radio and reports when each run last happened. `wnc_rrm_last_rf_grouping_run_timestamp_seconds` dates the RF grouping run and `wnc_rrm_last_dca_run_timestamp_seconds` the channel assignment. **Both carry `band` and nothing else**, which is the whole identifier the controller gives the record, so neither joins with a per-AP or a per-radio series. They come from one read behind `--collector.ap.radio` whose cost does not grow with the number of APs.

**The two runs are configured with separate intervals.** On the controller measured, grouping ran on the 600-second interval the controller declares for it, and the controller declared an hour for channel assignment on the band read: between two reads the grouping instant advanced by about its own interval on every band while the DCA instant did not move on any of them. On the two bands where DCA ran during a measured window, it reported the same instant as that band's grouping run, about an hour after the instant it replaced. `show ap dot11 <band> group` prints the grouping interval and `show ap dot11 <band> channel` the channel-assignment one, each beside the age of the last run. Read the intervals in force before choosing a range — the 15 minutes the [Overview](README.md#controller-side-update-schedule) page recommends is shorter than a channel-assignment interval of an hour, so a range that catches one run can miss the other.

**The grouping instant is not exclusive to grouping.** The controller prints the same instant under its transmit-power heading as well, and the two leaves behind those headings agreed to the second on every band and across a run, so read this series as the instant of the run that carries both rather than of grouping alone. The controller prints an age where these series carry an absolute instant, so use `time() - series` to compare the two.

These are what tells a held reading from a fresh one for the two `radio` module series the runs feed. `wnc_ap_channel_energy_dbm` is recomputed when DCA next runs for the band and `wnc_ap_channel_changes_total` is read from that run's assignment statistics, so a flat pair over a range containing no DCA run says nothing about the radio — see notes \*4 and \*13.

A band whose PHY type this exporter cannot name is withheld as a whole row rather than published as `unknown`, for the reason note \*15 gives. An instant at the epoch is withheld as everywhere else here. A record the controller sends without its grouping container withholds both series for that band, and a grouping container carrying no channel assignment withholds the DCA series alone. While the `rrm_main_data` read fails both families disappear from the scrape rather than reading zero, and `wnc_refresh_errors_total{data="rrm_main_data"}` is what reports that.

</details>

<details><summary><b>*17</b> The per-radio counters read zero for a window after their AP re-joins</summary><br/>

The twenty series marked here — the ten `traffic` counters and the ten `errors` counters above — all read one per-radio statistics record. After an AP re-joins CAPWAP the controller serves that record with **every counter in it at zero**, for a window well under a minute, and the next read carries a value **above** the one from before the event. The access point kept counting throughout; what was rebuilt is the controller's copy. Why the zeros appear was not established, so read them as a transient of that rebuild rather than as a reset.

Prometheus sees `X → 0 → X + Δ` and compensates for a reset that did not happen, so the whole of `X + Δ` is charged as an increase over any range spanning the return — on a counter with a long history, a spike far above anything the radio did. The gate is on the [Overview](README.md#counter-reset-timing) page, which also records what a re-join does and does not reset.

Three series of the `errors` module are not reached: `wnc_ap_radio_resets_total`, `wnc_ap_coverage_failed_clients` and `wnc_ap_last_radar_timestamp_seconds` read other data types and are published before this record is consulted. No shipped alert rule reads a series that is reached, but the shipped admin dashboard plots `rate()` over all twenty, so those panels do show the spike, and so does a rule of your own over the `traffic` or `errors` counters.

</details>

<details><summary><b>*18</b> Why the reset count is one series per radio, and when it falls</summary><br/>

The controller keeps this count in a list keyed by AP MAC, radio, cause and detail cause, so one radio legitimately carries several entries and this series is their total. **The cause is deliberately not a label**: it would multiply the series by the key domain, and every cause a radio had ever recorded would keep a series of its own long after the controller dropped the entry.

**The total falls when the controller drops entries.** Every fall observed happened together with the access point booting or re-joining, and nothing here generalises beyond that. Read it with `rate()` or `increase()` rather than as a lifetime total, and gate a rule on it the way the [Overview](README.md#counter-reset-timing) page gates the per-radio counters, since the event is the same one.

No leaf names the last cause and none orders the entries. Reading the container plainly, with `explicit`, with `report-all` and with `report-all-tagged` returned the same bytes every time, against a control the controller rejected with `400`, so nothing is being omitted: the cause of the most recent reset cannot be recovered from this container at all.

</details>

<details><summary><b>*19</b> Reading the AP coordinates</summary><br/>

The two series come from the geolocation operational container, which is a module of its own on the controller rather than part of the AP inventory, so the `geolocation` module reads it and nothing else. That is why the flag is separate: enabling it adds one request whose cost does not grow with the number of APs, and no other module pays for it. The container answered `200` on IOS-XE 17.12 and `204` on 17.15, 17.18 and 26.01 — the three had no AP joined, and a `204` is a successful read of nothing, so an estate with no coordinates configured leaves both series absent without raising `wnc_refresh_errors_total`. It answered `404` on none of the four.

**The unit is the platform's own.** `show ap geolocation summary` heads both columns with degrees, and the schema the controller serves declares `units "degree"` on both leaves. They are two families rather than one carrying an axis label because longitude and latitude are different quantities over different domains, `±180` against `±90`, so an average or a sum across such a label denotes nothing and a `sum()` spanning both mixes them silently.

**The pair is published together or not at all.** A longitude without its latitude puts the AP on the prime meridian at whatever latitude a query joins in, which reads as a location rather than as the gap it is, so both are withheld when either is missing, unparsable, or outside its own domain. The domain check is this exporter's: the schema types both leaves as `decimal64` with no `range` statement, so a value beyond `±180` is well formed on the wire and nothing else refuses it.

**It is a configured position rather than a measurement.** On the estate this was read, all three APs reported `Location Source` and `Height Source` as `Manual`, and the statistics container counted zero APs with a GNSS module and zero with a derived position — so the value changes when someone changes it, not on its own. A `reset capwap` was observed losing it, which is why the series can disappear for an AP that is still joined.

</details>
