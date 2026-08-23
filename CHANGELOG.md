# Changelog

Notable changes to the metric surface, one section per release — a short preamble, then the flag changes, then the metric changes. Each version heading links its GitHub release, which carries the date, the downloads and the full commit list. The measurement prose these sections carried before v0.13.0 is preserved in the [CHANGELOG at v0.12.0](https://github.com/umatare5/cisco-wnc-exporter/blob/v0.12.0/CHANGELOG.md).

A minor release may rename or remove a metric, because the controller owns the schema every series reads. Read the section for the version you are upgrading to before you upgrade.

## [v0.12.0]

> [!IMPORTANT]
> **BREAKING CHANGE** — a leaf the controller omits is withheld instead of published as `0` or `false`, so twenty-eight families can be absent where they were previously always present. **No name, type or label changes.** The [Absence](docs/README.md#a-leaf-the-controller-omits-is-withheld-not-published-as-zero) section carries the semantics, and `wnc_refresh_defaults_fallback_total` says whether a controller is on the affected path.

The AP collector learns where the controller thinks each AP is, and `wnc_client_info` gains two opt-in labels. Two panels of the bundled admin dashboard that returned nothing when a WLAN was selected now filter on the new `wlan_id` label.

### Flags

- **Added**: `--collector.ap.geolocation` — the two coordinate families, off by default, one request per refresh.
- **Changed**: `--collector.client.info-labels` accepts `wlan_id` and `device_type` — the default label set is unchanged.

### Metrics

- **Added**: `wnc_ap_longitude_degrees{mac}` and `wnc_ap_latitude_degrees{mac}` — published together or withheld together, absent rather than `0` on an estate with no coordinates, and a failed read counts under `wnc_refresh_errors_total{data="ap_geo_loc_data"}`.
- **Changed**: `wnc_client_info` gains the opt-in labels `wlan_id` and `device_type` — the [Client](docs/collector.client.md) page carries the join expression and the churn hazard.
- **Changed**: twenty-eight families go absent where the controller omits the leaf they read — nine WLAN config families, the RRM profile verdict, the AP configuration state, the two transmit power levels, the join state, and the fourteen CAPWAP and DTLS counters.
- **Changed**: ten HELP strings drop a `0 = disabled or not reported` conflation — six superficially similar families still decode an omitted leaf as `0` on purpose, and their HELP says so.
- **Changed**: `wnc_controller_boot_time_seconds` is decoded by the SDK's own accessor — the series is unchanged, and a malformed instant now counts under `wnc_refresh_errors_total{data="controller_boot_time"}` instead of passing as a silent absence.

## [v0.11.0]

> [!IMPORTANT]
> **BREAKING CHANGE** — twelve gauges drop their `state` label and publish the number the controller's own enumeration assigns that spelling. Every name, type and remaining label is unchanged, so a query keeps working and returns something different — [docs/enums.md](docs/enums.md) lists every spelling with its value.

Three timestamp families arrive behind existing flags, and the twelve state-labelled gauges become value-encoded. Six documented claims about counter resets, DTLS zeros and the WLAN byte total are corrected.

### Flags

None.

### Metrics

- **Added**: `wnc_ap_last_air_quality_timestamp_seconds{mac,radio}`, `wnc_rrm_last_rf_grouping_run_timestamp_seconds{band}` and `wnc_rrm_last_dca_run_timestamp_seconds{band}` — one new read whose cost does not grow with the number of APs.
- **Changed**: the HELP of `wnc_ap_radio_resets_total` now records that the count can fall.
- **Changed**: twelve families drop `state` and publish the enumeration value — a negated `state` matcher now matches everything, and a `sum` over one of them has to become a `count`.
- **Changed**: a spelling this release does not number is withheld rather than published with a wrong value, so one of the twelve can be absent for a subject it always carried before — the spelling reaches the log at `--log.level=debug` and nowhere else.

| Metric                                  | v0.10.0                           | v0.11.0              |
| :-------------------------------------- | :-------------------------------- | :------------------- |
| `wnc_ap_oper_state`                     | `{mac,state}`, always `1`         | `{mac}`, 1-6         |
| `wnc_ap_last_discovery_failure_reason`  | `{mac,state}`, always `1`         | `{mac}`, 0-16        |
| `wnc_ap_last_join_failure_reason`       | `{mac,state}`, always `1`         | `{mac}`, 0-41        |
| `wnc_ap_last_config_failure_reason`     | `{mac,state}`, always `1`         | `{mac}`, 0-13        |
| `wnc_ap_last_error_phase`               | `{mac,state}`, always `1`         | `{mac}`, 0-6         |
| `wnc_ap_last_dtls_failure_reason`       | `{mac,channel,state}`, always `1` | `{mac,channel}`, 0-9 |
| `wnc_ap_last_reboot_reason`             | `{mac,state}`, always `1`         | `{mac}`, 0-58        |
| `wnc_ap_last_disconnect_reason`         | `{mac,state}`, always `1`         | `{mac}`, 0-40        |
| `wnc_client_state`                      | `{mac,state}`, always `1`         | `{mac}`, 0-13        |
| `wnc_client_roam_type`                  | `{mac,state}`, always `1`         | `{mac}`, 0-4         |
| `wnc_wlan_pmf_state`                    | `{id,state}`, always `1`          | `{id}`, 0-2          |
| `wnc_wlan_ft_state`                     | `{id,state}`, always `1`          | `{id}`, 0-2          |

## [v0.10.0]

> [!IMPORTANT]
> **BREAKING CHANGE** — this release renames one gauge and withholds twenty-one per-radio series that previously reported a fabricated `0` for a slot that is not a radio.

Nine families arrive behind existing flags, and the per-radio counters stop reporting a remote-LAN port as a radio that carries no traffic and never errs.

### Flags

None.

### Metrics

- **Renamed**: `wnc_ap_air_quality_index` is now `wnc_ap_air_quality_index_avg` — same type, labels and value.
- **Added**: `wnc_ap_air_quality_index_min{mac,radio}`, `wnc_ap_interferers{mac,radio}`, `wnc_ap_channel_energy_dbm{mac,radio}`, `wnc_ap_association_uptime_seconds{mac}` and `wnc_client_roam_type{mac,state}`.
- **Added**: `wnc_rrm_worst_channel_air_quality_index_avg{band}`, `wnc_rrm_worst_channel_air_quality_index_min{band}`, `wnc_rrm_worst_channel_interferers{band}` and `wnc_rrm_worst_channel_number{band}` — one new read.
- **Changed**: the AP `traffic` and `errors` counters and `wnc_ap_clients` are absent for a slot that is not a radio — twenty-one series that reported `0` on a remote-LAN port.
- **Changed**: `wnc_client_state_transition_seconds` is absent in the two shapes the controller uses to say it measured no transition, instead of publishing `0`.
- **Changed**: the HELP of `wnc_ap_air_quality_index_avg` names the reporting period the controller declares and the direction of the scale, where it had called the window undeclared.
- **Changed**: the HELP of `wnc_controller_client_ap_auth_dot11i_fast_roams_total` drops a withdrawn alert shape — a roam the total counts need not fall into either sub-counter.

## [v0.9.2]

No metric, label, flag or HELP string changes — this release carries a test fix only.

## [v0.9.1]

No metric, label, flag or HELP string changes — this release carries a test fix only.

## [v0.9.0]

> [!IMPORTANT]
> **BREAKING CHANGE** — this release renames two counters and withholds two series that previously reported a fabricated `0`.

Four families arrive, one behind the new `spectrum` flag, and two AP state series stop reporting a slot whose state leaf the controller omits.

### Flags

- **Added**: `--collector.ap.spectrum` — the CleanAir air quality read, off by default, fetched last because its table grows with the number of CleanAir APs.
- **Changed**: `--collector.ap.radio` gains one request per refresh, for the RRM verdicts and the channel-change counter that join it.
- **Changed**: `--collector.controller.general` describes itself like the other module flags — no behavior change.

### Metrics

- **Renamed**: `wnc_ap_fragmentation_rx_total` is now `wnc_ap_rx_fragments_total` and `wnc_ap_fragmentation_tx_total` is now `wnc_ap_tx_fragments_total` — same type, labels and value.
- **Added**: `wnc_wlan_onboarding_clients{id,phase}`, `wnc_ap_rrm_profile_passed{mac,radio,profile}`, `wnc_ap_channel_changes_total{mac,radio}` and `wnc_ap_air_quality_index{mac,radio}`.
- **Changed**: `wnc_ap_radio_state` and `wnc_ap_admin_state` are absent for a slot whose state leaf the controller omits, instead of reporting a permanently failing radio.

## [v0.8.0]

Two counters take the plural the controller's own naming implies, and a fabricated client uptime is withheld. Five panels of the bundled admin dashboard drop a `radio` filter that matched nothing.

### Flags

None.

### Metrics

- **Renamed**: `wnc_ap_rts_success_total` is now `wnc_ap_rts_successes_total` and `wnc_ap_radio_reset_total` is now `wnc_ap_radio_resets_total` — same type, labels and value.
- **Changed**: `wnc_client_uptime_seconds` is absent for a client whose record carries no association time or an epoch timestamp, instead of reading as a session some 292 years long.
- **Changed**: four HELP strings of the controller collector no longer call its counters cumulative since the controller booted.

## [v0.7.0]

The AP collector gains a `join` module of thirty-two metrics, a controller collector arrives with five, and the WLAN collector gains two families. One fabricated AP uptime is withheld.

### Flags

- **Added**: `--collector.ap.join` — the CAPWAP join statistics, off by default, one request per refresh.
- **Added**: `--collector.controller.general` — the controller-wide series, off by default.

### Metrics

- **Added**: `wnc_ap_joined`, `wnc_ap_join_info`, and the discovery, join, configuration and DTLS counters, timestamps and reason enumerations — thirty-two metrics, listed individually in [docs/collector.ap.md](docs/collector.ap.md).
- **Added**: `wnc_controller_boot_time_seconds`, `wnc_controller_client_deletes_total` and the three `wnc_controller_client_ap_auth_*roams_total` counters — documented in [docs/collector.controller.md](docs/collector.controller.md).
- **Added**: `wnc_wlan_policy_binding{id,policy_profile,policy_tag}` and `wnc_wlan_data_usage_bytes_total`.
- **Changed**: `wnc_ap_uptime_seconds` is absent for a boot time the controller does not report, reports in an unparsable form, or reports at the Unix epoch, instead of publishing `0`.

## [v0.6.0]

No metric, label, flag or HELP string changes. The `linux/amd64` image can now execute — its binary was linked against a loader a `FROM scratch` image does not carry — and `docker run` with no argument starts the exporter instead of printing the flag list.

## [v0.5.0]

Three WLAN config families arrive, and a refresh reads only the `data` types the enabled modules need — the module flags now bound what the controller is asked for as well as what Prometheus stores.

### Flags

- **Changed**: every collector module flag now bounds the RESTCONF reads — with a module off, the `data` types only it needs are not fetched.

### Metrics

- **Added**: `wnc_wlan_pmf_state`, `wnc_wlan_ft_state` and `wnc_wlan_policy_enabled` — all in the WLAN `config` module, no new RESTCONF route.
- **Changed**: `wnc_refresh_errors_total` and `wnc_refresh_items` carry series only for the `data` types the enabled modules read, so their label sets depend on the collector flags.
- **Changed**: `wnc_up` reports `0` when every `data` type the enabled modules need failed, rather than when all eighteen failed.
- **Changed**: the HELP of `wnc_refresh_defaults_fallback_total` covers both decode paths — a valueless leaf reads as `0`, and a `state`-labelled one drops its series.

## [v0.4.0]

This release asks the controller for the values in force, which restores the two series v0.3.0 removed. The floating container tag `:v0` is no longer published — pin `:v0.4` to follow the minor line, or `:v0.4.0` for an immutable tag.

### Flags

- **Changed**: `--wnc.controller` and `--wnc.access-token` are trimmed of surrounding whitespace, and a whitespace-only value is rejected at validation instead of panicking at start-up.

### Metrics

- **Added**: `wnc_wlan_wpa2_enabled` and `wnc_wlan_11k_neighbor_list_enabled`, restored — v0.3.0 removed them because a plain read reported `0` on a WLAN where the feature was in force, and asking for the values in force is what they needed.
- **Added**: `wnc_refresh_defaults_fallback_total` — rises while a controller rejects the request for the values in force and the exporter falls back to a plain read.
- **Changed**: `wnc_wlan_enabled`, `wnc_wlan_session_timeout_seconds` and the four `wnc_wlan_central_*_enabled` series are emitted only when the controller returned the container they are read from, so a `0` no longer stands in for a container that was never sent.
- **Changed**: the `config` series report the values in force rather than only the leaves that differ from the default, so a series can change value with no configuration change.
- **Changed**: the HELP of `wnc_ap_admin_state` and `wnc_wlan_enabled` no longer equates a `0` with disabled, and the HELP of `wnc_client_power_save_state` no longer calls a `0` awake.

## [v0.3.0]

The metric surface is audited against what the controller actually sends — twelve series are removed, eleven are renamed, and five refresh meta-series arrive.

### Flags

- **Changed**: `--wnc.cache-ttl` is the minimum interval between refresh completions, not a snapshot expiry — the first scrape after start-up reports `wnc_up 0` and carries no data series.
- **Changed**: `--web.telemetry-path` now moves the endpoint, values that were accepted and ignored are rejected at start-up, and `/` replaces the landing page.

### Metrics

Removed — where a substitute exists, the reason names it.

| Metric                               | Reason                                                                         |
| :----------------------------------- | :----------------------------------------------------------------------------- |
| `wnc_ap_rx_bytes_total`              | A frame counter times a constant 1500; use `wnc_ap_data_rx_frames_total`       |
| `wnc_ap_tx_bytes_total`              | Same, against `wnc_ap_data_tx_frames_total`                                    |
| `wnc_ap_rx_packets_total`            | Read the leaf `wnc_ap_data_rx_frames_total` reads; substitute that name        |
| `wnc_ap_tx_packets_total`            | Same, against `wnc_ap_data_tx_frames_total`                                    |
| `wnc_ap_tx_drops_total`              | Named a drop count but read `ack-failure-count`                                |
| `wnc_ap_tx_errors_total`             | Read `failed-count`, already published as `wnc_ap_transmission_failures_total` |
| `wnc_ap_wep_undecryptable_total`     | Cannot leave zero unless a WLAN is configured for static WEP                   |
| `wnc_client_retry_ratio_percent`     | Summed two retry counters over a packet counter, all three still published raw |
| `wnc_wlan_rx_bytes_total`            | A sum over a changing client set is not a counter                              |
| `wnc_wlan_tx_bytes_total`            | Same                                                                           |
| `wnc_wlan_wpa2_enabled`              | Reported `0` on WLANs whose operative value is enabled — restored in v0.4.0    |
| `wnc_wlan_11k_neighbor_list_enabled` | Same                                                                           |

Renamed and rescaled — values are now ratios in `0-1`, so divide a threshold written for the old percentages by 100.

| v0.2.0                               | v0.3.0                             |
| :----------------------------------- | :--------------------------------- |
| `wnc_ap_channel_utilization_percent` | `wnc_ap_channel_utilization_ratio` |
| `wnc_ap_cpu_utilization_percent`     | `wnc_ap_cpu_utilization_ratio`     |
| `wnc_ap_memory_utilization_percent`  | `wnc_ap_memory_utilization_ratio`  |
| `wnc_ap_noise_utilization_percent`   | `wnc_ap_noise_utilization_ratio`   |
| `wnc_ap_rx_utilization_percent`      | `wnc_ap_rx_utilization_ratio`      |
| `wnc_ap_tx_utilization_percent`      | `wnc_ap_tx_utilization_ratio`      |

Renamed:

| v0.2.0                              | v0.3.0                                             |
| :---------------------------------- | :------------------------------------------------- |
| `wnc_ap_clients_total`              | `wnc_ap_clients`                                   |
| `wnc_ap_coverage_hole_events_total` | `wnc_ap_coverage_failed_clients`, now a Gauge      |
| `wnc_ap_last_radar_on_radio_at`     | `wnc_ap_last_radar_timestamp_seconds`, now a Gauge |
| `wnc_client_rx_group_counter_total` | `wnc_client_rx_group_total`                        |
| `wnc_wlan_clients_total`            | `wnc_wlan_clients`                                 |

Added — five series report the health of the WNC data refresh itself.

| Metric                                   | Reports                                                    |
| :--------------------------------------- | :--------------------------------------------------------- |
| `wnc_up`                                 | Whether the last completed refresh reached the controller  |
| `wnc_refresh_duration_seconds`           | Duration of the last refresh attempt                       |
| `wnc_refresh_success_timestamp_seconds`  | Start time of the refresh behind the served snapshot       |
| `wnc_refresh_errors_total`               | Fetch failures per `data` type since process start         |
| `wnc_refresh_items`                      | Items the last refresh returned per `data` type            |

Changed, under unchanged names:

- **Absence**: a series whose source record is missing is absent instead of `0` — a fetch failure and an unlisted client or radio both withhold.
- **`wnc_ap_radio_reset_total`**: sums every reset cause per radio instead of tracking the cause the controller listed last.
- **`band`**: the label on `wnc_ap_info` and `wnc_client_info` changes value and gains `unknown`.
- **`wnc_ap_tx_power_dbm` and `wnc_ap_tx_power_max_dbm`**: read from the record matching the operating band, so they can change value and can be absent.
- **`wnc_ap_noise_floor_dbm`**: read from the entry matching the radio's channel, so it can change value and can be absent.
- **`wnc_client_protocol`**: returns 802.11b where it returned `unknown` or 802.11g.
- **`wnc_ap_oper_state` and `wnc_client_state`**: always `1` with the state in a `state` label, and `wnc_ap_oper_state` drops its `radio` label — rewrite `== 0` alerts as label matches.

[v0.12.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.12.0
[v0.11.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.11.0
[v0.10.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.10.0
[v0.9.2]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.9.2
[v0.9.1]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.9.1
[v0.9.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.9.0
[v0.8.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.8.0
[v0.7.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.7.0
[v0.6.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.6.0
[v0.5.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.5.0
[v0.4.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.4.0
[v0.3.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.3.0
