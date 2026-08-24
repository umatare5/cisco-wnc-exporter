# Changelog

Notable changes to the metric surface, one section per release — a short preamble, the breaking change where the release has one, then the metric changes by collector and the flag changes. A section whose table spans collectors keeps one metric list instead.

A minor release may rename or remove a metric, because the controller owns the schema every series reads. Every collector module flag is off by default, so a release that adds one adds no series until you set it. A rename carries the type, the labels and the value of the old name unless the entry says otherwise.

## [v0.13.0]

This release stops fabricating a channel `0` for a radio whose leaf the controller omits, and stops fetching the CAPWAP inventory for modules that never read it.

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - `wnc_ap_channel_number` and `wnc_ap_channel_width_mhz` — either can now be absent where the leaf it reads is omitted or zero. Names, types and labels are unchanged, and a rule comparing either to `0` stops matching.

### AP Metrics

#### Changed

- `wnc_ap_channel_number` — absent on a radio in monitor mode, where the controller omits the channel leaf.
- `wnc_ap_channel_width_mhz` — absent where the width leaf reads zero, a case no live controller produced.

### Exporter Metrics

#### Changed

- `wnc_refresh_items{data="ap_capwap_data"}` — absent unless the AP `general` or `info` module is enabled.
- `wnc_refresh_errors_total{data="ap_capwap_data"}` — absent unless the AP `general` or `info` module is enabled.

### Flags

None.

## [v0.12.0]

This release reads the coordinates the controller holds for each AP, and gives `wnc_client_info` two opt-in labels. Two panels of the bundled admin dashboard that returned nothing when a WLAN was selected now filter on the new `wlan_id` label.

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - **Twenty-eight families** — any can now be absent where the controller omits the leaf it reads, and six similar families still decode an omitted leaf as `0` on purpose. Names, types and labels are unchanged.
> - `wnc_refresh_defaults_fallback_total` — says whether a controller is on the affected path, and [Absence](docs/README.md#absence) carries the semantics.

### AP Metrics

#### Added

- `wnc_ap_longitude_degrees{mac}` — the longitude the controller holds for the AP, absent rather than `0` where it holds none.
- `wnc_ap_latitude_degrees{mac}` — the matching latitude, published with the longitude or withheld with it.

#### Changed

- `wnc_ap_rrm_profile_passed` — absent where the controller reports no verdict, and its HELP no longer calls an unreported verdict a failure.
- **Eighteen other AP families** — the configuration state, the two transmit power levels, the join state and the fourteen CAPWAP and DTLS counters go absent where the controller omits their leaf.

### Client Metrics

#### Changed

- `wnc_client_info` — gains the opt-in labels `wlan_id` and `device_type`.

### WLAN Metrics

#### Changed

- **Nine config families**: `wnc_wlan_wpa2_enabled` and eight others go absent where the controller omits their leaf, and their HELP no longer equates a `0` with disabled or unreported.

### Controller Metrics

#### Changed

- `wnc_controller_boot_time_seconds` — now counts under `wnc_refresh_errors_total{data="controller_boot_time"}`.

### Flags

#### Added

- `--collector.ap.geolocation` — the two coordinate families, longitude and latitude, are behind this flag.

#### Changed

- `--collector.client.info-labels` — accepts `wlan_id` and `device_type`, and the default label set is unchanged.

## [v0.11.0]

This release adds three timestamp families behind existing flags, and turns the twelve state-labelled gauges into value-encoded ones.

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - **Twelve gauges** — drop the `state` label and publish the number the controller's own enumeration assigns that spelling, which [docs/enums.md](docs/enums.md) lists. Every name, type and remaining label is unchanged.
> - **A query over those twelve** — a negated `state` matcher now matches everything, and a `sum` over one of them has to become a `count`.
> - **An unnumbered spelling** — withheld rather than published with a wrong value, so one of the twelve can be absent for a subject it always carried. The spelling reaches only the `--log.level=debug` log.

### Metrics

#### Added

- `wnc_ap_last_air_quality_timestamp_seconds{mac,radio}` — the instant of the CleanAir row the air quality series read, withheld where the controller carries none.
- `wnc_rrm_last_rf_grouping_run_timestamp_seconds{band}` — the last RF grouping run, from one new read keyed per band rather than per AP.
- `wnc_rrm_last_dca_run_timestamp_seconds{band}` — the last DCA run, from that same per-band read.

#### Changed

- `wnc_ap_radio_resets_total` — its HELP now records that the count can fall.
- **The twelve gauges** — the table below gives the labels and range each now publishes, across the AP, Client and WLAN collectors.

| Metric                                 | v0.10.0                           | v0.11.0              |
| :------------------------------------- | :-------------------------------- | :------------------- |
| `wnc_ap_oper_state`                    | `{mac,state}`, always `1`         | `{mac}`, 1-6         |
| `wnc_ap_last_discovery_failure_reason` | `{mac,state}`, always `1`         | `{mac}`, 0-16        |
| `wnc_ap_last_join_failure_reason`      | `{mac,state}`, always `1`         | `{mac}`, 0-41        |
| `wnc_ap_last_config_failure_reason`    | `{mac,state}`, always `1`         | `{mac}`, 0-13        |
| `wnc_ap_last_error_phase`              | `{mac,state}`, always `1`         | `{mac}`, 0-6         |
| `wnc_ap_last_dtls_failure_reason`      | `{mac,channel,state}`, always `1` | `{mac,channel}`, 0-9 |
| `wnc_ap_last_reboot_reason`            | `{mac,state}`, always `1`         | `{mac}`, 0-58        |
| `wnc_ap_last_disconnect_reason`        | `{mac,state}`, always `1`         | `{mac}`, 0-40        |
| `wnc_client_state`                     | `{mac,state}`, always `1`         | `{mac}`, 0-13        |
| `wnc_client_roam_type`                 | `{mac,state}`, always `1`         | `{mac}`, 0-4         |
| `wnc_wlan_pmf_state`                   | `{id,state}`, always `1`          | `{id}`, 0-2          |
| `wnc_wlan_ft_state`                    | `{id,state}`, always `1`          | `{id}`, 0-2          |

### Flags

None.

## [v0.10.0]

This release adds nine families behind existing flags, and stops the per-radio counters from reporting a remote-LAN port as a radio that carries no traffic and never errs.

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - `wnc_ap_air_quality_index` — renamed, so a query naming it returns nothing.
> - **Twenty-one per-radio series** — a rule that expected one series per slot now sees fewer, where a remote-LAN port had reported a fabricated `0`.

### AP Metrics

#### Renamed

- `wnc_ap_air_quality_index` — now `wnc_ap_air_quality_index_avg`.

#### Added

- `wnc_ap_air_quality_index_min{mac,radio}` — the lowest index over the same reporting period as the average, which it never exceeds.
- `wnc_ap_interferers{mac,radio}` — interference devices CleanAir attributes to the channel the radio operates on.
- `wnc_ap_channel_energy_dbm{mac,radio}` — energy measured on the assigned channel, from the controller's DCA statistics.
- `wnc_ap_association_uptime_seconds{mac}` — seconds since the CAPWAP association the AP currently holds began.
- `wnc_rrm_worst_channel_air_quality_index_avg{band}` — the average index of the channel the controller ranks worst, from one new read.
- `wnc_rrm_worst_channel_air_quality_index_min{band}` — the lowest index the controller saw on that channel.
- `wnc_rrm_worst_channel_interferers{band}` — interference devices CleanAir counts on that channel.
- `wnc_rrm_worst_channel_number{band}` — which channel that is, as a reading rather than a label.

#### Changed

- `wnc_ap_air_quality_index_avg` — its HELP names the reporting period the controller declares, and says a higher index is cleaner.
- **The `traffic` and `errors` counters and `wnc_ap_clients`** — absent for a slot that is not a radio, twenty-one series in all.

### Client Metrics

#### Added

- `wnc_client_roam_type{mac,state}` — how the client reached the association it currently holds.

#### Changed

- `wnc_client_state_transition_seconds` — absent in the two shapes the controller uses to say it measured no transition, instead of publishing `0`.

### Controller Metrics

#### Changed

- `wnc_controller_client_ap_auth_dot11i_fast_roams_total` — its HELP drops a withdrawn alert shape, as a roam the total counts may fall into neither sub-counter.

### Flags

None.

## [v0.9.2]

This release carries a test fix only, with no metric, label, flag or HELP string change.

## [v0.9.1]

This release carries a test fix only, with no metric, label, flag or HELP string change.

## [v0.9.0]

This release adds four families, one behind the new `spectrum` flag, and stops two AP state series from reporting a slot whose state leaf the controller omits.

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - `wnc_ap_fragmentation_rx_total` and `wnc_ap_fragmentation_tx_total` — renamed, so a query naming either returns nothing.
> - `wnc_ap_radio_state` and `wnc_ap_admin_state` — withheld where they had reported a fabricated `0`, so a rule reading a permanently failing radio now sees nothing.

### AP Metrics

#### Renamed

- `wnc_ap_fragmentation_rx_total` — now `wnc_ap_rx_fragments_total`.
- `wnc_ap_fragmentation_tx_total` — now `wnc_ap_tx_fragments_total`.

#### Added

- `wnc_ap_rrm_profile_passed{mac,radio,profile}` — whether the radio passes each RRM profile the controller reports a verdict for.
- `wnc_ap_channel_changes_total{mac,radio}` — channel changes from the controller's DCA assignment statistics.
- `wnc_ap_air_quality_index{mac,radio}` — the CleanAir air quality index of the channel the radio operates on.

#### Changed

- `wnc_ap_radio_state` and `wnc_ap_admin_state` — absent for a slot whose state leaf the controller omits.

### WLAN Metrics

#### Added

- `wnc_wlan_onboarding_clients{id,phase}` — clients held in one onboarding phase, short of the run state `wnc_wlan_clients` counts.

### Flags

#### Added

- `--collector.ap.spectrum` — the CleanAir air quality read, fetched last because its table grows with the number of CleanAir APs.

#### Changed

- `--collector.ap.radio` — gains one request per refresh, for the RRM verdicts and the channel-change counter that join it.
- `--collector.controller.general` — describes itself like the other module flags, with no behavior change.

## [v0.8.0]

This release gives two counters the plural the controller's own naming implies, and withholds a fabricated client uptime. Five panels of the bundled admin dashboard drop a `radio` filter that matched nothing.

### AP Metrics

#### Renamed

- `wnc_ap_rts_success_total` — now `wnc_ap_rts_successes_total`.
- `wnc_ap_radio_reset_total` — now `wnc_ap_radio_resets_total`.

### Client Metrics

#### Changed

- `wnc_client_uptime_seconds` — absent for a client whose record carries no association time or an epoch timestamp, instead of a session some 292 years long.

### Controller Metrics

#### Changed

- **Four HELP strings** — no longer call the controller collector's counters cumulative since the controller booted.

### Flags

None.

## [v0.7.0]

This release adds an AP `join` module of thirty-two metrics, a controller collector of five, and two WLAN families, and withholds one fabricated AP uptime.

### AP Metrics

#### Added

- `wnc_ap_joined` — whether the AP holds a CAPWAP session with this controller now.
- `wnc_ap_join_info` — the AP name as its CAPWAP join record reports it, a record that outlives the session.
- **The thirty other join metrics** — the discovery, join, configuration and DTLS counters, timestamps and reason enumerations, listed in [docs/collector.ap.md](docs/collector.ap.md).

#### Changed

- `wnc_ap_uptime_seconds` — absent for a boot time the controller omits, reports in an unparsable form, or reports at the Unix epoch, instead of publishing `0`.

### WLAN Metrics

#### Added

- `wnc_wlan_policy_binding{id,policy_profile,policy_tag}` — the policy tag binding for the WLAN, one series per binding the exporter can resolve.
- `wnc_wlan_data_usage_bytes_total` — bytes in both directions as the controller totals them, including clients that have since disconnected.

### Controller Metrics

#### Added

- `wnc_controller_boot_time_seconds` — the Unix time the controller last booted.
- `wnc_controller_client_deletes_total` — client deletions counted per reason the controller reports.
- **The three `wnc_controller_client_ap_auth_*roams_total` counters** — documented in [docs/collector.controller.md](docs/collector.controller.md).

### Flags

#### Added

- `--collector.ap.join` — the CAPWAP join statistics, one request per refresh.
- `--collector.controller.general` — the controller-wide series.

## [v0.6.0]

This release changes no metric, label, flag or HELP string. The `linux/amd64` image can now execute — its binary was linked against a loader a `FROM scratch` image does not carry — and `docker run` with no argument starts the exporter instead of printing the flag list.

## [v0.5.0]

This release adds three WLAN config families on no new RESTCONF route, and reads only the `data` types the enabled modules need. The module flags now bound what the controller is asked for as well as what Prometheus stores.

### WLAN Metrics

#### Added

- `wnc_wlan_pmf_state` — the protected management frames setting the WLAN carries.
- `wnc_wlan_ft_state` — the 802.11r fast transition mode.
- `wnc_wlan_policy_enabled` — whether the policy profile bound to the WLAN is active.

### Exporter Metrics

#### Changed

- `wnc_refresh_errors_total` and `wnc_refresh_items` — carry series only for the `data` types the enabled modules read, so their label sets depend on the collector flags.
- `wnc_up` — reports `0` when every `data` type the enabled modules need failed, rather than when all eighteen failed.
- `wnc_refresh_defaults_fallback_total` — its HELP covers both decode paths, a valueless leaf reading as `0` and a `state`-labelled one dropping its series.

### Flags

#### Changed

- **Every collector module flag** — now bounds the RESTCONF reads, so with a module off the `data` types only it needs are not fetched.

## [v0.4.0]

This release asks the controller for the values in force rather than only the leaves that differ from the default, which restores the two series v0.3.0 removed. A `config` series can therefore change value with no configuration change. The floating container tag `:v0` is no longer published — pin `:v0.4` to follow the minor line, or `:v0.4.0` for an immutable tag.

### AP Metrics

#### Changed

- `wnc_ap_admin_state` — its HELP no longer equates a `0` with disabled.

### Client Metrics

#### Changed

- `wnc_client_power_save_state` — its HELP no longer calls a `0` awake.

### WLAN Metrics

#### Added

- `wnc_wlan_wpa2_enabled` — restored, now that the exporter asks for the values in force.
- `wnc_wlan_11k_neighbor_list_enabled` — restored on the same read.

#### Changed

- `wnc_wlan_enabled` — emitted only when the controller returned the container it is read from, and its HELP no longer equates a `0` with disabled.
- `wnc_wlan_session_timeout_seconds` — emitted only when that container arrived.
- **The four `wnc_wlan_central_*_enabled` series** — emitted only when that container arrived, so a `0` no longer stands in for a container that was never sent.

### Exporter Metrics

#### Added

- `wnc_refresh_defaults_fallback_total` — rises while a controller rejects the request for the values in force and the exporter falls back to a plain read.

### Flags

#### Changed

- `--wnc.controller` and `--wnc.access-token` — trimmed of surrounding whitespace, and a whitespace-only value is rejected at validation instead of panicking at start-up.

## [v0.3.0]

This release audits the metric surface against what the controller actually sends — twelve series are removed, eleven are renamed, and five refresh meta-series arrive.

### Metrics

#### Removed

Where a substitute exists, the reason names it.

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

#### Renamed

Six utilization series are rescaled as well — values are now ratios in `0-1`, so divide a threshold written for the old percentages by 100.

| v0.2.0                               | v0.3.0                             |
| :----------------------------------- | :--------------------------------- |
| `wnc_ap_channel_utilization_percent` | `wnc_ap_channel_utilization_ratio` |
| `wnc_ap_cpu_utilization_percent`     | `wnc_ap_cpu_utilization_ratio`     |
| `wnc_ap_memory_utilization_percent`  | `wnc_ap_memory_utilization_ratio`  |
| `wnc_ap_noise_utilization_percent`   | `wnc_ap_noise_utilization_ratio`   |
| `wnc_ap_rx_utilization_percent`      | `wnc_ap_rx_utilization_ratio`      |
| `wnc_ap_tx_utilization_percent`      | `wnc_ap_tx_utilization_ratio`      |

The other five change name, and two of them become Gauges.

| v0.2.0                              | v0.3.0                                             |
| :---------------------------------- | :------------------------------------------------- |
| `wnc_ap_clients_total`              | `wnc_ap_clients`                                   |
| `wnc_ap_coverage_hole_events_total` | `wnc_ap_coverage_failed_clients`, now a Gauge      |
| `wnc_ap_last_radar_on_radio_at`     | `wnc_ap_last_radar_timestamp_seconds`, now a Gauge |
| `wnc_client_rx_group_counter_total` | `wnc_client_rx_group_total`                        |
| `wnc_wlan_clients_total`            | `wnc_wlan_clients`                                 |

#### Added

Five series report the health of the WNC data refresh itself.

| Metric                                  | Reports                                                   |
| :-------------------------------------- | :-------------------------------------------------------- |
| `wnc_up`                                | Whether the last completed refresh reached the controller |
| `wnc_refresh_duration_seconds`          | Duration of the last refresh attempt                      |
| `wnc_refresh_success_timestamp_seconds` | Start time of the refresh behind the served snapshot      |
| `wnc_refresh_errors_total`              | Fetch failures per `data` type since process start        |
| `wnc_refresh_items`                     | Items the last refresh returned per `data` type           |

#### Changed

- **Every per-subject series** — absent instead of `0` where its source record is missing, so a fetch failure and an unlisted client or radio both withhold.
- `wnc_ap_radio_reset_total` — sums every reset cause per radio instead of tracking the cause the controller listed last.
- `wnc_ap_info` and `wnc_client_info` — their `band` label changes value and gains `unknown`.
- `wnc_ap_tx_power_dbm` and `wnc_ap_tx_power_max_dbm` — read from the record matching the operating band, so they can change value and can be absent.
- `wnc_ap_noise_floor_dbm` — read from the entry matching the radio's channel, so it can change value and can be absent.
- `wnc_client_protocol` — returns 802.11b where it returned `unknown` or 802.11g.
- `wnc_ap_oper_state` and `wnc_client_state` — always `1` with the state in a `state` label, and `wnc_ap_oper_state` drops its `radio` label, so rewrite a `== 0` alert as a label match.

### Flags

#### Changed

- `--wnc.cache-ttl` — the minimum interval between refresh completions rather than a snapshot expiry, so the first scrape reports `wnc_up 0` with no data series.
- `--web.telemetry-path` — now moves the endpoint, values that were accepted and ignored are rejected at start-up, and `/` replaces the landing page.

[v0.13.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.13.0
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
