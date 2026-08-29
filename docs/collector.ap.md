# AP collector

AP collector focuses on RF foundation and radio performance.

## Metrics

| Module      | Metric                                            | Type    | Description                                 |
| :---------- | :------------------------------------------------ | :------ | :------------------------------------------ |
| general     | `wnc_ap_admin_state`                              | Gauge   | Admin state, absent if unreported           |
| general     | `wnc_ap_oper_state`                               | Gauge   | Operational state (4=registered)            |
| general     | `wnc_ap_radio_state`                              | Gauge   | Radio state, absent if unreported           |
| general     | `wnc_ap_config_state`                             | Gauge   | Tag config state, absent if unreported      |
| general     | `wnc_ap_uptime_seconds`                           | Gauge   | Uptime in seconds, absent without boot time |
| general     | `wnc_ap_association_uptime_seconds`               | Gauge   | Age of the current association              |
| general     | `wnc_ap_cpu_utilization_ratio`                    | Gauge   | CPU utilization ratio (0-1)                 |
| general     | `wnc_ap_memory_utilization_ratio`                 | Gauge   | Memory utilization ratio (0-1)              |
| radio       | `wnc_ap_channel_number`                           | Gauge   | Channel number, absent if unreported        |
| radio       | `wnc_ap_channel_width_mhz`                        | Gauge   | Channel bandwidth, absent if unreported     |
| radio       | `wnc_ap_tx_power_dbm`                             | Gauge   | TX power (dBm), absent if unreported        |
| radio       | `wnc_ap_tx_power_max_dbm`                         | Gauge   | Max TX power (dBm), absent if unreported    |
| radio       | `wnc_ap_noise_floor_dbm`                          | Gauge   | Noise on the operating channel (dBm)        |
| radio       | `wnc_ap_channel_utilization_ratio`                | Gauge   | Channel utilization ratio (CCA), 0-1        |
| radio       | `wnc_ap_rx_utilization_ratio`                     | Gauge   | RX utilization ratio (0-1)                  |
| radio       | `wnc_ap_tx_utilization_ratio`                     | Gauge   | TX utilization ratio (0-1)                  |
| radio       | `wnc_ap_noise_utilization_ratio`                  | Gauge   | Noise channel utilization ratio (0-1)       |
| radio       | `wnc_ap_clients`                                  | Gauge   | Run-state clients count (calculated)        |
| radio       | `wnc_ap_rrm_profile_passed`                       | Gauge   | RRM profile verdict per `profile`           |
| radio       | `wnc_ap_channel_changes_total`                    | Counter | Channel changes, DCA statistics             |
| radio       | `wnc_ap_channel_energy_dbm`                       | Gauge   | Energy DCA measured on the channel          |
| radio       | `wnc_rrm_last_rf_grouping_run_timestamp_seconds`  | Gauge   | Last RF grouping run per band               |
| radio       | `wnc_rrm_last_dca_run_timestamp_seconds`          | Gauge   | Last DCA run per band                       |
| traffic     | `wnc_ap_total_tx_frames_total`                    | Counter | TX frames, not a sum of the frame series    |
| traffic     | `wnc_ap_data_rx_frames_total`                     | Counter | Data RX frames                              |
| traffic     | `wnc_ap_data_tx_frames_total`                     | Counter | Data TX frames                              |
| traffic     | `wnc_ap_management_rx_frames_total`               | Counter | Management RX frames                        |
| traffic     | `wnc_ap_management_tx_frames_total`               | Counter | Management TX frames                        |
| traffic     | `wnc_ap_control_rx_frames_total`                  | Counter | Control RX frames                           |
| traffic     | `wnc_ap_control_tx_frames_total`                  | Counter | Control TX frames                           |
| traffic     | `wnc_ap_multicast_rx_frames_total`                | Counter | Multicast RX frames                         |
| traffic     | `wnc_ap_multicast_tx_frames_total`                | Counter | Multicast TX frames                         |
| traffic     | `wnc_ap_rts_successes_total`                      | Counter | Successful RTS transmissions                |
| errors      | `wnc_ap_rx_errors_total`                          | Counter | Total RX errors                             |
| errors      | `wnc_ap_tx_retries_total`                         | Counter | Total TX retries                            |
| errors      | `wnc_ap_transmission_failures_total`              | Counter | Failed transmission attempts                |
| errors      | `wnc_ap_duplicate_frames_total`                   | Counter | Duplicate frames received                   |
| errors      | `wnc_ap_fcs_errors_total`                         | Counter | Frame Check Sequence errors                 |
| errors      | `wnc_ap_rx_fragments_total`                       | Counter | RX fragments                                |
| errors      | `wnc_ap_tx_fragments_total`                       | Counter | TX fragments                                |
| errors      | `wnc_ap_rts_failures_total`                       | Counter | RTS failures                                |
| errors      | `wnc_ap_decryption_errors_total`                  | Counter | Decryption errors                           |
| errors      | `wnc_ap_mic_errors_total`                         | Counter | MIC errors                                  |
| errors      | `wnc_ap_coverage_failed_clients`                  | Gauge   | Clients failing the RRM coverage check      |
| errors      | `wnc_ap_last_radar_timestamp_seconds`             | Gauge   | Unix timestamp of the last radar            |
| errors      | `wnc_ap_radio_resets_total`                       | Counter | Radio reset count                           |
| join        | `wnc_ap_joined`                                   | Gauge   | CAPWAP session held now (0=no, 1=yes)       |
| join        | `wnc_ap_join_info`                                | Gauge   | AP name from the join record, always 1      |
| join        | `wnc_ap_discovery_requests_total`                 | Counter | CAPWAP discovery requests received          |
| join        | `wnc_ap_discovery_responses_total`                | Counter | Successful discovery responses sent         |
| join        | `wnc_ap_discovery_errors_total`                   | Counter | Discovery requests found in error           |
| join        | `wnc_ap_join_requests_total`                      | Counter | CAPWAP join requests received               |
| join        | `wnc_ap_join_responses_total`                     | Counter | Successful join responses sent              |
| join        | `wnc_ap_join_failures_total`                      | Counter | Join requests that failed to process        |
| join        | `wnc_ap_config_requests_total`                    | Counter | Configuration requests received             |
| join        | `wnc_ap_config_responses_total`                   | Counter | Successful configuration responses sent     |
| join        | `wnc_ap_config_failures_total`                    | Counter | Configuration requests that failed          |
| join        | `wnc_ap_dtls_session_requests_total`              | Counter | DTLS setup requests, per `channel`          |
| join        | `wnc_ap_dtls_session_successes_total`             | Counter | DTLS sessions established, per `channel`    |
| join        | `wnc_ap_dtls_session_failures_total`              | Counter | DTLS sessions that failed, per `channel`    |
| join        | `wnc_ap_dtls_decrypt_errors_total`                | Counter | DTLS decrypt errors, per `channel`          |
| join        | `wnc_ap_dtls_anti_replay_errors_total`            | Counter | DTLS anti-replay errors, per `channel`      |
| join        | `wnc_ap_last_error_timestamp_seconds`             | Gauge   | Last connection error                       |
| join        | `wnc_ap_last_join_success_timestamp_seconds`      | Gauge   | Last successful join                        |
| join        | `wnc_ap_last_join_failure_timestamp_seconds`      | Gauge   | Last failed join                            |
| join        | `wnc_ap_last_config_success_timestamp_seconds`    | Gauge   | Last successful configuration               |
| join        | `wnc_ap_last_config_failure_timestamp_seconds`    | Gauge   | Last failed configuration                   |
| join        | `wnc_ap_last_discovery_success_timestamp_seconds` | Gauge   | Last successful discovery                   |
| join        | `wnc_ap_last_discovery_failure_timestamp_seconds` | Gauge   | Last failed discovery                       |
| join        | `wnc_ap_last_dtls_success_timestamp_seconds`      | Gauge   | Last DTLS session, per `channel`            |
| join        | `wnc_ap_last_dtls_failure_timestamp_seconds`      | Gauge   | Last failed DTLS, per `channel`             |
| join        | `wnc_ap_last_discovery_failure_reason`            | Gauge   | Discovery failure reason                    |
| join        | `wnc_ap_last_join_failure_reason`                 | Gauge   | Join failure reason                         |
| join        | `wnc_ap_last_config_failure_reason`               | Gauge   | Configuration failure reason                |
| join        | `wnc_ap_last_error_phase`                         | Gauge   | Phase of the last error                     |
| join        | `wnc_ap_last_dtls_failure_reason`                 | Gauge   | DTLS outcome per `channel`                  |
| join        | `wnc_ap_last_reboot_reason`                       | Gauge   | Reboot reason                               |
| join        | `wnc_ap_last_disconnect_reason`                   | Gauge   | Disconnect reason                           |
| geolocation | `wnc_ap_longitude_degrees`                        | Gauge   | WGS 84 longitude of the AP                  |
| geolocation | `wnc_ap_latitude_degrees`                         | Gauge   | WGS 84 latitude of the AP                   |
| spectrum    | `wnc_ap_air_quality_index_avg`                    | Gauge   | CleanAir air quality of the channel         |
| spectrum    | `wnc_ap_air_quality_index_min`                    | Gauge   | CleanAir air quality minimum                |
| spectrum    | `wnc_ap_interferers`                              | Gauge   | Interference devices on that channel        |
| spectrum    | `wnc_ap_last_air_quality_timestamp_seconds`       | Gauge   | Instant stamped on the air quality row      |
| spectrum    | `wnc_rrm_worst_channel_air_quality_index_avg`     | Gauge   | Worst channel air quality per band          |
| spectrum    | `wnc_rrm_worst_channel_air_quality_index_min`     | Gauge   | Worst channel minimum per band              |
| spectrum    | `wnc_rrm_worst_channel_interferers`               | Gauge   | Interference devices on that channel        |
| spectrum    | `wnc_rrm_worst_channel_number`                    | Gauge   | Which channel that is, as a value           |

## Specifications

Each entry carries what the series' HELP text and the shared [Absence](README.md#absence) rules do not.

**`wnc_ap_admin_state`, `wnc_ap_radio_state` and every other per-radio series**

- The slot list is not a list of radios — a remote-LAN port arrives as a slot whose state leaves the controller omits entirely, and every per-radio series is withheld for it, the `traffic` and `errors` counters and `wnc_ap_clients` included.
- A rule treating any of those as always present therefore needs `absent()` or `or vector(0)`, and a `sum()` over a controller carrying such a port reads lower than the radio count suggests.

**`wnc_ap_uptime_seconds` and `wnc_ap_association_uptime_seconds`**

- The pair diverges for an AP that re-joined without rebooting and agrees for one that joined straight after booting, so the two answer different questions.
- Neither measures how long an AP was gone: a reboot replaces the record in place rather than deleting it, so a read right after one can still serve the association held before it.
- Whether a silent AP leaves the inventory at all, and how long its record keeps its reading, was not the same on every model measured, so do not build an outage rule on the absence of either.

**`wnc_ap_cpu_utilization_ratio` and `wnc_ap_memory_utilization_ratio`**

- Both read `0` rather than going absent while AP system monitoring is off — the response still carries the statistics block with its leaf at zero, so a zero cannot be told from an idle access point. Enable the collection on the AP join profile:

> ```plaintext
> configure terminal
>  ap profile <profile-name>
>   statistics ap-system-monitoring enable
>  end
> write memory
> ```

**`wnc_ap_channel_number`**

- It is absent on a radio in monitor mode, because the controller omits the channel leaf there rather than reporting a channel the radio does not serve. A rule comparing the series to `0` therefore stops matching on such a radio, and `wnc_ap_channel_width_mhz` is withheld on its own zero independently of it.
- It reports the number without a band, and 6 GHz numbering restarts at 1, so a 6 GHz number collides with a 2.4 GHz one — join `wnc_ap_info` to disambiguate it, which needs `band` named in `--collector.ap.info-labels` because that label is off by default:

> ```bash
> wnc_ap_channel_number * on(mac,radio) group_left(band) wnc_ap_info
> ```

**`wnc_ap_noise_floor_dbm`**

- The controller reports noise per channel across the whole band, so the reading is selected by matching the radio's operating channel and the series is absent where no entry matches it — the case for a radio in monitor or sniffer mode.

**`wnc_ap_channel_utilization_ratio`**

- It is the measured quantity behind the `load` verdict, as `wnc_ap_noise_floor_dbm` is behind `noise` and `wnc_ap_coverage_failed_clients` behind `coverage`.
- Read capacity from this ratio and from `wnc_ap_clients` rather than from the `load` verdict, which a radio carrying a crowd fails while nothing about the RF is wrong.

**`wnc_ap_rx_utilization_ratio` and the thirteen `wnc_ap_*_total` counters below**

- Every series in the table read zero on every radio measured while its neighbours in the same container advanced, and the controller CLI agreed — a failed fetch withholds a series instead, so a zero here is in the data the controller holds.
- Whether a leaf is maintained depends on the access point model and the release — `wnc_ap_fcs_errors_total` advanced on one model and read zero on another, with multicast transmit frames the reverse — so confirm the table against your own access points.

  | Metric                                   | What the zero means here                                            |
  | :--------------------------------------- | :------------------------------------------------------------------ |
  | `wnc_ap_rx_utilization_ratio`            | Channel and noise utilization read non-zero. Cause not established. |
  | `wnc_ap_control_(rx\|tx)_frames_total`   | Data and management frames advanced.                                |
  | `wnc_ap_multicast_(rx\|tx)_frames_total` | Receive zero. Transmit advanced on one model, not on another.       |
  | `wnc_ap_rx_errors_total`                 | FCS errors advanced on the same radio.                              |
  | `wnc_ap_transmission_failures_total`     | Retries advanced on the same radio.                                 |
  | `wnc_ap_duplicate_frames_total`          | Counted on receive, so client retransmissions drive it.             |
  | `wnc_ap_rts_(successes\|failures)_total` | The RTS threshold sits at its maximum, so RTS never triggers.       |
  | `wnc_ap_(rx\|tx)_fragments_total`        | Fragmentation threshold at its maximum. Receive count unverified.   |
  | `wnc_ap_decryption_errors_total`         | Zero is the healthy reading. Failure reporting is unconfirmed.      |
  | `wnc_ap_mic_errors_total`                | Zero is the healthy reading, with the same caveat.                  |

**`wnc_ap_clients`**

- Clients are attributed to a radio through the AP **name** the record carries, that record having no AP MAC, so one whose AP name is absent from `ap-name-mac-map` is left out silently and `sum(wnc_ap_clients)` reads low rather than high.
- The series is withheld for every radio when either the client list or that mapping fails to fetch, rather than counting the clients it could still attribute.

**`wnc_ap_rrm_profile_passed`**

- The four `profile` values — `coverage`, `load`, `interference` and `noise` — are this exporter's own names for four separate verdict leaves, not spellings the controller assigns.
- Just after a radio re-joined all four read `1` while the channel energy the controller reported for that radio still carried its unmeasured sentinel, so read a verdict taken shortly after a re-join as unmeasured in either direction.
- A failure is not by itself an incident — `interference`, `load` and `noise` have all been observed failing in ordinary conditions and only `coverage` never has — so alert on a verdict stuck failed rather than on the first scrape:

> ```bash
> max_over_time(wnc_ap_rrm_profile_passed[30m]) == 0
> ```

**`wnc_ap_channel_changes_total` and `wnc_ap_channel_energy_dbm`**

- Both read the DCA assignment statistics of the record the four verdicts come from, so they add no request, and both go absent for a record carrying no DCA statistics and for every radio while the `rrm_radio_slot` read fails.
- What `wnc_ap_channel_changes_total` counts is not established — the CLI prints it beside a radar-driven count no leaf carries, and `wnc_ap_last_radar_timestamp_seconds` being absent is what says a move it counted was not radar-driven.
- `wnc_ap_channel_energy_dbm` withholds `-128` and `0`, neither of which can be a measurement: `-128` is the lower bound of the leaf's own signed type and what a radio reads until DCA next runs for its band.

**`wnc_rrm_last_rf_grouping_run_timestamp_seconds` and `wnc_rrm_last_dca_run_timestamp_seconds`**

- Both carry `band` alone, so neither joins a per-AP series, and both disappear while the `rrm_main_data` read fails — a grouping container carrying no channel assignment withholds the DCA instant alone.
- The grouping instant is not exclusive to grouping: the controller prints the same instant under its transmit-power heading, and the two leaves agreed to the second on every band and across a run.
- `show ap dot11 <band> group` and `show ap dot11 <band> channel` print the two intervals, so read those in force before choosing a range — the DCA instant is what tells a held `wnc_ap_channel_energy_dbm` reading from a fresh one.

**`wnc_ap_transmission_failures_total`**

- It reads `failed-count`, and a retry that eventually succeeds produces no terminal failure, so read a zero as unconfirmed rather than as a healthy radio.
- The CLI's per-radio transmit-drop counter was non-zero where this read zero, and the operational tree carries drop leaves for the Ethernet interface alone, so no series can reach it.

**`wnc_ap_last_radar_timestamp_seconds`**

- It is published only for a radio whose last-radar leaf carries a real instant, and every record measured carried the epoch instead — so on a controller that has seen no radar it is absent everywhere, the ordinary reading rather than a fault.

**`wnc_ap_radio_resets_total`**

- The cause is deliberately not a label, since it would multiply the series by the key domain and outlive the entry the controller dropped, and no leaf names or orders the entries — so the most recent reset's cause cannot be recovered here.

**`wnc_ap_joined`**

- Its record set is wider than the AP inventory — leaving CAPWAP evicts nothing and what does evict a record is not established, so a `0` here can name an AP months out of the inventory with its counters frozen.
- It is a snapshot, so an AP flapping in and out of CAPWAP reads joined at every evaluation and never fires — count the transitions instead:

> ```bash
> changes(wnc_ap_joined[1h]) > 2
> ```

**`wnc_ap_join_info`**

- Its name puts it in the info cache, which wraps this collector only when `--collector.ap.info` is enabled, so with the join module alone the name is read fresh on every scrape and with both modules enabled it is up to `--collector.info-cache-ttl` old.

**`wnc_ap_discovery_requests_total`**

- An AP that holds this controller as its secondary, or that discovers by broadcast, DHCP option 43 or DNS, sends discovery requests here while joining another controller, so exclude those by `mac` and use a `for:` longer than the rate window.
- Paired with `wnc_ap_joined` it carries the signal the module exists for, an AP that reaches the controller and cannot complete a join:

> ```bash
> rate(wnc_ap_discovery_requests_total[15m]) > 0 and wnc_ap_joined == 0
> ```

**`wnc_ap_last_join_success_timestamp_seconds`**

- An outage leaves no mark in either uptime series, so this is what an outage check watches — count the re-joins inside a range:

> ```bash
> changes(wnc_ap_last_join_success_timestamp_seconds[1h]) > 0
> ```

**`wnc_ap_last_reboot_reason`**

- It dates nothing, so two reboots recorded with the same reason cannot be told from one, which is why `wnc_ap_last_join_success_timestamp_seconds` is what counts them.

**`wnc_ap_longitude_degrees` and `wnc_ap_latitude_degrees`**

- An estate with no coordinates configured leaves both absent without raising `wnc_refresh_errors_total`, because the container answers a successful `204` rather than an error, so absence here is not a failed read.
- A `reset capwap` was observed losing the position for an AP that was still joined, and the pair is withheld for a value outside `±180` or `±90` as well — a bound this exporter applies because the schema declares none.

**`wnc_ap_air_quality_index_avg`, `wnc_ap_air_quality_index_min` and `wnc_ap_interferers`**

- All three read one row of the per-AP-and-band table and are published or withheld together — absent for an AP without CleanAir, a radio whose spectrum operation is down or in monitor mode, and every radio while the `rrm_spectrum_aq_table` read fails.
- Silence never means clean air — a refresh can publish every other per-radio series for a newly joined AP while the table carries no row for it, and whether a Spectrum Intelligence radio reports once enabled was not measured.
- That table is the last read of a refresh, so a refresh cut short by its deadline drops these three and their instant before any other series.

**`wnc_ap_air_quality_index_min`**

- Interference elsewhere in the band moves neither this nor the average, because both report the radio's primary channel alone even on a bonded radio — `show ap dot11 <band> cleanair config` reports the period they cover, which an operator can change.

**`wnc_ap_last_air_quality_timestamp_seconds`**

- The instant is stamped per AP rather than per radio, so an AP reporting on several bands repeats one instant and a difference between two APs is the controller's report boundary rather than one radio going stale.

**the twenty `wnc_ap_*_total` counters that read the per-radio statistics record**

- The read after the window that follows an AP re-joining CAPWAP carries a value above the one from before it, so Prometheus sees `X → 0 → X + Δ` and charges the whole of `X + Δ` as an increase over any range spanning the return.
- The bundled admin dashboard plots `rate()` over all twenty and so does show that spike, while no bundled alert rule reads any of them.
- `wnc_ap_coverage_failed_clients`, `wnc_ap_radio_resets_total` and `wnc_ap_last_radar_timestamp_seconds` read other data types and are published before that record is consulted, so the window does not reach them.

**the nine `wnc_ap_last_*_timestamp_seconds` series of the `join` module**

- The controller writes the Unix epoch for an event that has not happened and this exporter withholds that sentinel, so the four failure timestamps are absent for every AP on a controller where nothing has failed rather than reading 1970.

**the five `wnc_ap_dtls_*_total` counters**

- `channel` is valued `control` or `data`, this exporter's own names for the two sets the controller keeps in one container, and it is the CAPWAP tunnel channel rather than the RF one `wnc_ap_channel_number` reports, so an `and` needs `on(mac)`.
- DTLS on the data channel is off by default, and every `channel="data"` counter read zero while the `control` ones advanced — disabling it again left the counters at their count, so a zero reports the AP join profile and what has already run on it.
- The `channel="data"` series appear once a handshake on that channel has succeeded and stay until the controller rebuilds the record, `wnc_ap_last_dtls_success_timestamp_seconds` included, so an `absent()` rule reading the data channel as unused stops matching after the first success and does not recover when it is disabled again.
- `wnc_ap_dtls_session_failures_total` and the two `wnc_ap_dtls_*_errors_total` counters stayed at zero with the data channel enabled, so whether they can advance at all was not established — read a zero as unconfirmed rather than as nothing configured.

**the four `wnc_ap_last_*_failure_reason` series**

- These four are the reason series whose `0` is a healthy sentinel, so `== 0` selects the healthy APs and `!= 0` is the alertable set — `wnc_ap_last_error_phase`, `wnc_ap_last_reboot_reason` and `wnc_ap_last_disconnect_reason` number something else at `0`.

**the four `wnc_rrm_worst_channel_*` series**

- They carry `band` alone — the whole identifier the controller gives the row — so none joins a per-AP series, and a band this exporter cannot name or one the controller has not ranked is withheld as a whole row rather than zeroed.
- The row set follows the controller's table rather than the radios that have joined, so a band no CleanAir-capable radio scans never appears and a band's absence reads as no ranking rather than as clean air.
- A controller or an image not carrying the `rrm_spectrum_aq_worst_table` these four alone read answers `404` — a failure rather than an absence, so all four disappear while that type's `wnc_refresh_errors_total` rises on every refresh and `wnc_up` stays `1`.

## Info Labels

The `info` module publishes `wnc_ap_info` with the following labels to join with other metrics:

| Labels       | Description             | Example Value              | Default | Required |
| :----------- | :---------------------- | :------------------------- | :-----: | :------: |
| `mac`        | AP wireless MAC address | `aa:bb:cc:dd:ee:f0`        | **Yes** | **Yes**  |
| `name`       | AP hostname             | `TEST-AP01`                | **Yes** |    No    |
| `ip`         | AP IP address           | `192.168.1.10`             | **Yes** |    No    |
| `radio`      | Radio identifier        | `0`, `1`, `2`              | **Yes** | **Yes**  |
| `band`       | Radio band              | `2.4`, `5`, `6`, `unknown` |   No    |    No    |
| `model`      | AP model                | `AIR-AP1815I-Q-K9`         |   No    |    No    |
| `serial`     | AP serial number        | `FGL1234ABCD`              |   No    |    No    |
| `sw_version` | Software version        | `17.15.6.12`               |   No    |    No    |
| `eth_mac`    | Ethernet MAC address    | `aa:bb:cc:00:11:22`        |   No    |    No    |

Use this info metric to add contextual labels to other metrics in PromQL queries:

```bash
wnc_ap_radio_state * on(mac,radio) group_left(name,ip) wnc_ap_info
```

> [!NOTE]
>
> ### About the Labels
>
> **`band`:** `unknown` is the value this exporter publishes for a band it cannot name — `dot11-invalid-band`, the leaf a remote-LAN radio omits, and any value a later release adds. It is a label value rather than a withheld series, so the other labels on that row still join. The band-keyed `wnc_rrm_*` series do the opposite and withhold the whole row, because the band is their whole identifier.
>
> **`radio`:** No substitute for `band`, because a dual band radio keeps its slot while it moves between bands. A join against a series that carries no `radio` label fails with a duplicate match group rather than returning nothing, so collapse the info metric first:
>
> ```bash
> wnc_ap_uptime_seconds * on(mac) group_left(name) max by (mac,name) (wnc_ap_info)
> ```
