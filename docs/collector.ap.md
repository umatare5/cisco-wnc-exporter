# AP collector

AP collector focuses on RF foundation and radio performance.

## Metrics

| Module   | Metric                                            | Type    | Description                                        |
| :------- | :------------------------------------------------ | :------ | :------------------------------------------------- |
| general  | `wnc_ap_admin_state`                              | Gauge   | Admin state, absent if unreported **(\*12)**       |
| general  | `wnc_ap_oper_state`                               | Gauge   | Operational state in `state` label                 |
| general  | `wnc_ap_radio_state`                              | Gauge   | Radio state, absent if unreported **(\*12)**       |
| general  | `wnc_ap_config_state`                             | Gauge   | Tag config state (0=valid, 1=invalid)              |
| general  | `wnc_ap_uptime_seconds`                           | Gauge   | AP uptime in seconds, absent without boot time     |
| general  | `wnc_ap_association_uptime_seconds`               | Gauge   | Age of the current association **(\*14)**          |
| general  | `wnc_ap_cpu_utilization_ratio`                    | Gauge   | CPU utilization ratio (0-1) **(\*1)**              |
| general  | `wnc_ap_memory_utilization_ratio`                 | Gauge   | Memory utilization ratio (0-1) **(\*1)**           |
| radio    | `wnc_ap_channel_number`                           | Gauge   | Operating channel number **(\*2)**                 |
| radio    | `wnc_ap_channel_width_mhz`                        | Gauge   | Channel bandwidth (MHz)                            |
| radio    | `wnc_ap_tx_power_dbm`                             | Gauge   | Current transmit power (dBm)                       |
| radio    | `wnc_ap_tx_power_max_dbm`                         | Gauge   | Maximum TX power capability (dBm)                  |
| radio    | `wnc_ap_noise_floor_dbm`                          | Gauge   | Noise on the operating channel (dBm) **(\*2)**     |
| radio    | `wnc_ap_channel_utilization_ratio`                | Gauge   | Channel utilization ratio (CCA), 0-1               |
| radio    | `wnc_ap_rx_utilization_ratio`                     | Gauge   | RX utilization ratio (0-1) **(\*3)**               |
| radio    | `wnc_ap_tx_utilization_ratio`                     | Gauge   | TX utilization ratio (0-1)                         |
| radio    | `wnc_ap_noise_utilization_ratio`                  | Gauge   | Noise channel utilization ratio (0-1)              |
| radio    | `wnc_ap_clients`                                  | Gauge   | Run-state clients count (calculated)               |
| radio    | `wnc_ap_rrm_profile_passed`                       | Gauge   | RRM profile verdict per `profile` **(\*4)**        |
| radio    | `wnc_ap_channel_changes_total`                    | Counter | Channel changes, DCA statistics **(\*4)**          |
| radio    | `wnc_ap_channel_energy_dbm`                       | Gauge   | Energy DCA measured on the channel **(\*13)**      |
| traffic  | `wnc_ap_total_tx_frames_total`                    | Counter | TX frames, not a sum of the frame series           |
| traffic  | `wnc_ap_data_rx_frames_total`                     | Counter | Data RX frames                                     |
| traffic  | `wnc_ap_data_tx_frames_total`                     | Counter | Data TX frames                                     |
| traffic  | `wnc_ap_management_rx_frames_total`               | Counter | Management RX frames                               |
| traffic  | `wnc_ap_management_tx_frames_total`               | Counter | Management TX frames                               |
| traffic  | `wnc_ap_control_rx_frames_total`                  | Counter | Control RX frames **(\*3)**                        |
| traffic  | `wnc_ap_control_tx_frames_total`                  | Counter | Control TX frames **(\*3)**                        |
| traffic  | `wnc_ap_multicast_rx_frames_total`                | Counter | Multicast RX frames **(\*3)**                      |
| traffic  | `wnc_ap_multicast_tx_frames_total`                | Counter | Multicast TX frames **(\*3)**                      |
| traffic  | `wnc_ap_rts_successes_total`                      | Counter | Successful RTS transmissions **(\*3)**             |
| errors   | `wnc_ap_rx_errors_total`                          | Counter | Total RX errors **(\*3)**                          |
| errors   | `wnc_ap_tx_retries_total`                         | Counter | Total TX retries                                   |
| errors   | `wnc_ap_transmission_failures_total`              | Counter | Failed transmission attempts **(\*3)** **(\*5)**   |
| errors   | `wnc_ap_duplicate_frames_total`                   | Counter | Duplicate frames received                          |
| errors   | `wnc_ap_fcs_errors_total`                         | Counter | Frame Check Sequence errors                        |
| errors   | `wnc_ap_rx_fragments_total`                       | Counter | RX fragments **(\*3)**                             |
| errors   | `wnc_ap_tx_fragments_total`                       | Counter | TX fragments **(\*3)**                             |
| errors   | `wnc_ap_rts_failures_total`                       | Counter | RTS failures **(\*3)**                             |
| errors   | `wnc_ap_decryption_errors_total`                  | Counter | Decryption errors **(\*3)**                        |
| errors   | `wnc_ap_mic_errors_total`                         | Counter | MIC errors **(\*3)**                               |
| errors   | `wnc_ap_coverage_failed_clients`                  | Gauge   | Clients failing the RRM coverage check             |
| errors   | `wnc_ap_last_radar_timestamp_seconds`             | Gauge   | Unix timestamp of the last radar **(\*6)**         |
| errors   | `wnc_ap_radio_resets_total`                       | Counter | Radio reset count                                  |
| join     | `wnc_ap_joined`                                   | Gauge   | CAPWAP session held now (0=no, 1=yes) **(\*7)**    |
| join     | `wnc_ap_join_info`                                | Gauge   | AP name from the join record, always 1             |
| join     | `wnc_ap_discovery_requests_total`                 | Counter | CAPWAP discovery requests received                 |
| join     | `wnc_ap_discovery_responses_total`                | Counter | Successful discovery responses sent                |
| join     | `wnc_ap_discovery_errors_total`                   | Counter | Discovery requests found in error                  |
| join     | `wnc_ap_join_requests_total`                      | Counter | CAPWAP join requests received                      |
| join     | `wnc_ap_join_responses_total`                     | Counter | Successful join responses sent                     |
| join     | `wnc_ap_join_failures_total`                      | Counter | Join requests that failed to process               |
| join     | `wnc_ap_config_requests_total`                    | Counter | Configuration requests received                    |
| join     | `wnc_ap_config_responses_total`                   | Counter | Successful configuration responses sent            |
| join     | `wnc_ap_config_failures_total`                    | Counter | Configuration requests that failed                 |
| join     | `wnc_ap_dtls_session_requests_total`              | Counter | DTLS setup requests, per `channel` **(\*8)**       |
| join     | `wnc_ap_dtls_session_successes_total`             | Counter | DTLS sessions established, per `channel`           |
| join     | `wnc_ap_dtls_session_failures_total`              | Counter | DTLS sessions that failed, per `channel`           |
| join     | `wnc_ap_dtls_decrypt_errors_total`                | Counter | DTLS decrypt errors, per `channel`                 |
| join     | `wnc_ap_dtls_anti_replay_errors_total`            | Counter | DTLS anti-replay errors, per `channel`             |
| join     | `wnc_ap_last_error_timestamp_seconds`             | Gauge   | Last connection error **(\*9)**                    |
| join     | `wnc_ap_last_join_success_timestamp_seconds`      | Gauge   | Last successful join **(\*9)**                     |
| join     | `wnc_ap_last_join_failure_timestamp_seconds`      | Gauge   | Last failed join **(\*9)**                         |
| join     | `wnc_ap_last_config_success_timestamp_seconds`    | Gauge   | Last successful configuration **(\*9)**            |
| join     | `wnc_ap_last_config_failure_timestamp_seconds`    | Gauge   | Last failed configuration **(\*9)**                |
| join     | `wnc_ap_last_discovery_success_timestamp_seconds` | Gauge   | Last successful discovery **(\*9)**                |
| join     | `wnc_ap_last_discovery_failure_timestamp_seconds` | Gauge   | Last failed discovery **(\*9)**                    |
| join     | `wnc_ap_last_dtls_success_timestamp_seconds`      | Gauge   | Last DTLS session, per `channel` **(\*9)**         |
| join     | `wnc_ap_last_dtls_failure_timestamp_seconds`      | Gauge   | Last failed DTLS, per `channel` **(\*9)**          |
| join     | `wnc_ap_last_discovery_failure_reason`            | Gauge   | Discovery failure reason in `state` **(\*10)**     |
| join     | `wnc_ap_last_join_failure_reason`                 | Gauge   | Join failure reason in `state` **(\*10)**          |
| join     | `wnc_ap_last_config_failure_reason`               | Gauge   | Configuration failure reason in `state` **(\*10)** |
| join     | `wnc_ap_last_error_phase`                         | Gauge   | Phase of the last error in `state` **(\*10)**      |
| join     | `wnc_ap_last_dtls_failure_reason`                 | Gauge   | DTLS outcome in `state`, per `channel` **(\*10)**  |
| join     | `wnc_ap_last_reboot_reason`                       | Gauge   | Reboot reason in `state` **(\*10)**                |
| join     | `wnc_ap_last_disconnect_reason`                   | Gauge   | Disconnect reason in `state` **(\*10)**            |
| spectrum | `wnc_ap_air_quality_index_avg`                    | Gauge   | CleanAir air quality of the channel **(\*11)**     |
| spectrum | `wnc_ap_air_quality_index_min`                    | Gauge   | CleanAir air quality minimum **(\*11)**            |
| spectrum | `wnc_ap_interferers`                              | Gauge   | Interference devices on that channel **(\*11)**    |
| spectrum | `wnc_rrm_worst_channel_air_quality_index_avg`     | Gauge   | Worst channel air quality per band **(\*15)**      |
| spectrum | `wnc_rrm_worst_channel_air_quality_index_min`     | Gauge   | Worst channel minimum per band **(\*15)**          |
| spectrum | `wnc_rrm_worst_channel_interferers`               | Gauge   | Interference devices on that channel **(\*15)**    |
| spectrum | `wnc_rrm_worst_channel_number`                    | Gauge   | Which channel that is, as a value **(\*15)**       |

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

The container also carries `rx-data-pkt-count` and `tx-data-pkt-count`, which no series reads. Both were observed at zero on every radio measured, through RESTCONF and through the CLI, including radios whose data-frame counters advanced over the same interval. v0.2.0 published a pair of packet series that read the data-frame leaves rather than these, and v0.3.0 removed those series as duplicates. Re-pointing a series at these leaves would publish a constant zero on a radio carrying traffic.

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

The four series are absent for a radio the slot list has no record for, for a record that carries no radio data, and for every radio while the `rrm_radio_slot` fetch fails. A verdict leaf the controller omits from a record it did send cannot be told from a reported failure, so a verdict can report a failure that was never measured. **The error runs the other way as well.** Just after a radio re-joined, all four verdicts were present and reading `1` while the channel energy the controller reports for that radio still read its unmeasured sentinel, and the radio had been failing interference before the re-join and failed it again afterwards — so the controller serves verdicts it has not yet measured, and the fabrication is the controller's rather than this exporter's decode. Read a verdict taken shortly after a re-join as unmeasured in either direction.

`wnc_ap_channel_changes_total` comes from the same record and adds no request. **It is read from the controller's DCA assignment statistics, and what it counts is not established.** The controller's CLI prints the same count under a DCA heading and keeps a separate count of radar-driven changes, which no leaf carries, so whether a radar move also advances this count could not be measured — both read zero on every radio here. `wnc_ap_last_radar_timestamp_seconds`, which the `errors` module publishes, dates a DFS event, so where it is absent no radar is on record and a move was not radar-driven. That inference is the most the pair supports, and on a controller that has seen no radar it is absent everywhere — see note *6. The counter was monotonic non-decreasing on every radio across fifteen consecutive reads two minutes apart, with one radio observed stepping by one as its best channel moved — and it was later observed **falling to zero on the one radio that had a count to lose**, with the controller's boot time unchanged, while every radio already reading zero stayed there. **The anchor is the radio's own AP joining CAPWAP rather than a controller-wide rebuild of the statistics tables** — an access point left untouched through the same interval kept its counters and its anchor, so a reset reaching one AP does not reach another's series, as [Counter reset timing](README.md#counter-reset-timing) describes. `rate()` and `increase()` absorb that; a rule reading the raw value as a lifetime total does not. Keep the range well above the controller's channel-assignment interval, ten minutes by default, or a healthy radio and a churning one both read zero. It sits one container deeper than the verdicts and is absent on its own when the controller reports no assignment statistics for a radio, which a zero would misreport as a radio that has never moved.

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

The join state and the phase counters carry `mac` and nothing else, deliberately. A bare `and` requires both sides to carry identical label sets, so one extra label on either side would make the query below return an empty result rather than an error. The DTLS series add `channel` and the reason series add `state`, so an `and` written against one of those needs `on(mac)` or `ignoring(...)`:

```bash
rate(wnc_ap_discovery_requests_total[15m]) > 0 and wnc_ap_joined == 0
```

That is the signal this module exists for: **the AP reaches the controller and cannot complete a join.** It works because the discovery counters keep advancing while the session is gone — measured on an AP whose join, configuration and DTLS counters were frozen while its discovery counters were the highest of any AP on the controller, its last successful discovery minutes after its last successful join.

Three cautions on that query. An AP that holds this controller as its secondary, or that discovers by broadcast, DHCP option 43 or DNS, sends discovery requests here while joining another controller, so pair the rule with a `for:` longer than the rate window and exclude those APs by `mac`. A counter here is cumulative since the controller allocated the record, so `increase()` over a range shorter than the discovery interval reports nothing. And the query is a snapshot of the join state, so an AP flapping in and out of CAPWAP can be joined at every evaluation and never fire — `changes(wnc_ap_joined[1h]) > 2` catches that case instead.

`wnc_ap_join_info` carries the name because the AP inventory no longer does, which is why the name is a series of its own rather than a label on the other thirty-one. The join record is not the only place a departed AP is still named — the AP history list names one too, measured — but this module reads the join record, and that record is keyed by the same radio MAC every series here carries. Join it with `* on(mac) group_left(name) wnc_ap_join_info`, which needs no collapsing because it carries no `radio` label. Its name puts it in the info cache, and that cache is only wrapped around the collector when `--collector.ap.info` is enabled, so with the join module alone the name is read fresh on every scrape and with both modules enabled it is up to `--collector.info-cache-ttl` old.

The record set is wider than the inventory. Compare like with like to see that — `wnc_ap_info` is one series per radio, so the count to compare against is `count(count by (mac) (wnc_ap_info))`. What evicts a record is not established — leaving CAPWAP does not, and no controller reload has been observed against this list.

</details>

<details><summary><b>*8</b> The DTLS channel label, and the zeros on the data channel</summary><br/>

The controller keeps one set of DTLS counters per CAPWAP tunnel channel in a single container, and the exporter folds the pair into one series carrying `channel`, valued `control` or `data`. This is the tunnel channel; the RF channel is `wnc_ap_channel_number`.

Every `channel="data"` counter was observed at zero on every AP measured while the `control` counters advanced, because DTLS on the data channel is not enabled by default. Read those zeros as not configured rather than as no traffic, and confirm against your own AP join profile before alerting on them.

</details>

<details><summary><b>*9</b> When a join timestamp series is absent</summary><br/>

The controller writes `1970-01-01T00:00:00+00:00` into a timestamp leaf for an event that has not happened, and that sentinel is withheld rather than published as an instant in 1970 — the same rule as note *6. On a controller where nothing has failed, the join, configuration and discovery failure timestamps and the DTLS failure timestamp are therefore absent for every AP, which is the ordinary reading rather than a fault. Use `time() - series` for the age.

</details>

<details><summary><b>*10</b> The reason series freeze with the record, and the controller misspells one value</summary><br/>

Each of these reports the controller's own spelling in the `state` label and always has the value `1`, so `== 0` never fires — see [States](README.md#a-state-is-a-label-not-a-number). They report the **last recorded** event rather than a current state, and they freeze with the record, so an AP that has left CAPWAP keeps reporting the reason it recorded while it was joined.

`wnc_ap_last_error_phase` is the sharpest example: an AP that is not joined reports the same `ap-con-failure-run` phase as one that is, on every record measured. Read it as the phase of the last error, never as a health check.

On the controller measured, `wnc_ap_last_join_failure_reason` reported `jf-none`, `wnc_ap_last_config_failure_reason` `cf-none`, `wnc_ap_last_discovery_failure_reason` `disc-fail-none` and `wnc_ap_last_dtls_failure_reason` `dtls-hs-success` on every AP. Those are the healthy sentinels of their enumerations, so an equality match on them selects the healthy APs and any other spelling is the alertable set.

**The disconnect reason enumeration spells its unknown value `unkown`**, and that misspelling is on the wire: a rule matching it has to use that spelling. It is recorded here rather than in the metric's HELP text because this repository's spell check rejects it in Go source.

The record also carries two free-text leaves — a prose disconnect description and a message-decryption failure reason — and neither is published, because neither has a value domain that a label can be matched against.

</details>

<details><summary><b>*11</b> Which radios report air quality, and why one goes silent</summary><br/>

The controller publishes air quality per AP and band, and the reading here is the one for the channel the radio operates on. **The series is absent rather than zero wherever that reading cannot be reached**, chiefly an AP without CleanAir, a radio whose spectrum operation is down, a radio in monitor or sniffer mode with no primary channel, a slot that is not a radio at all and so has no channel to match — see note \*12 — and every radio while the `rrm_spectrum_aq_table` fetch fails. Silence therefore does not mean clean air — on the controller measured here most radios reported a reading and the rest did not, and which ones moved as APs joined.

One of those cases was narrowed. A radio whose spectrum operation reads down belonged, on every such radio measured, to an access point carrying Spectrum Intelligence rather than full CleanAir, with spectrum intelligence itself left disabled — **whether such a radio reports a reading once spectrum intelligence is enabled could not be measured**, because none measured had it enabled. Silence can also be early rather than settled: after an AP joins, a refresh can publish every other per-radio series for it while the air quality table still carries no row for it, so this reading appears later than the rest of the radio. And **a reading present is not a reading now** — across an AP reboot the controller served the value from before the reboot, marked in no way, until the AP reported again. The controller keeps a second and unrelated air quality as well, `ap-sensor-air-quality`, which a sensor radio reports and no series here reads.

**The value is an average**, over the air quality reporting period the controller declares — `show ap dot11 <band> cleanair config` reports its length. No configuration leaf carries it, so the CLI is the only place to read it, and an operator can change it rather than relying on this page. **A higher index is cleaner**: the controller's own air-quality alarm threshold is a lower bound on the index, and so is the sensitivity its event-driven RRM acts on, though which of the two readings that sensitivity is compared against could not be read. The CLI names the same per-channel figures the average and the minimum, and the two matched this exporter's source leaves row for row on every channel checked. `wnc_ap_air_quality_index_min` publishes the minimum of the same reporting period and `wnc_ap_interferers` the number of interference devices CleanAir attributes to that channel; both read the row this average reads, so the three are published and withheld together. The minimum carries a signal the average does not — it was measured well below the average during an interference event — which is why it is published rather than left as a note. None of the three is guarded on its value, so a zero from `wnc_ap_interferers` is the controller's reading rather than a fabricated one.

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

Three cautions on the alternatives. Whether an AP that has gone silent leaves the inventory at all, and how long its record keeps the reading it had, was not the same on every model measured, so do not build an outage rule on the absence of these series. `wnc_ap_last_reboot_reason` classifies a cause and dates nothing: it reports the controller's spelling in a `state` label and always reads `1`, so two reboots recorded with the same reason cannot be told from one. And a state that lasts less than the interval between refreshes, which `--wnc.cache-ttl` sets, may never be sampled at all.

</details>

<details><summary><b>*15</b> The band-worst series are keyed by band alone, and one read carries all four</summary><br/>

The controller ranks the channels of a band across every AP that scans it and keeps one row per band for the worst of them; a higher index is cleaner, so the worst channel is the one with the lowest average. The four `wnc_rrm_worst_channel_*` series publish that row — the average, the minimum, the interference device count and the channel number — behind `--collector.ap.spectrum`, and they cost one read per refresh whatever the number of APs, because the table is per band rather than per AP.

The row is chosen on the average, so its minimum is the minimum of that one channel and not the deepest of the band: a channel with a better average can carry a worse minimum, and the per-radio pair is what reports the channels radios actually use.

**They carry `band` and nothing else, which is the whole identifier the controller gives the row.** Neither an AP nor a radio keys it, so none of these series can be joined with the per-AP series. The controller does name the AP that detected the channel, and that name is not published: it changes while the reading does not, and it could not be joined against `wnc_ap_info`, which is keyed by MAC.

`wnc_rrm_worst_channel_number` is a **value rather than a label**, because the worst channel of a band moves — a label would start a new series on every move and leave the previous one to go stale. Read the other three together with it: they describe whichever channel it names at that moment.

**A band this exporter cannot name is withheld as a whole row rather than published as `unknown`.** That is the one exception to the rule on the [Overview](README.md#the-band-label) page, and it holds for two reasons: `band` is the whole identifier here, so an `unknown` row would name nothing to act on, and two such rows would carry the same label set, which fails the whole `/metrics` endpoint instead of one series. A row reporting no channel is withheld the same way — the controller returns that shape for a band it has not ranked, and a zero there cannot be told from a reading: it would report the worst possible air quality, no interference at all, and a channel that does not exist.

**The row set follows the controller's table rather than the radios that have joined**, so fewer bands than the controller supports can appear, and a band no CleanAir-capable radio scans never does. Treat one, two or three bands as ordinary, and read a band's absence as no ranking rather than as clean air.

These four are the only series behind the release's one new read, `rrm_spectrum_aq_worst_table`, which is also its one new failure path. **A controller or an image that does not carry the table answers `404`, and a `404` is a failure rather than an absence** — the rule note \*4 on the [Controller](collector.controller.md) page describes, with the expression that excludes such a data type. There, `wnc_refresh_errors_total{data="rrm_spectrum_aq_worst_table"}` rises on every refresh, no `wnc_refresh_items` series appears for it, and all four families disappear from the scrape rather than reading zero, while `wnc_up` stays `1`. The error series is there either way — it is seeded at `0` for every data type the enabled modules read — so read its value and not its presence.

The reading is refreshed on the reporting period the controller declares for air quality and carries no timestamp, so a value held between reports cannot be told from one just taken; the range guidance on the [Overview](README.md#controller-side-update-schedule) page covers that.

</details>
