# Changelog

Notable changes to the metric surface, one section per release. Release dates, downloads and the full commit list live on the [releases page](https://github.com/umatare5/cisco-wnc-exporter/releases).

This project is pre-1.0, so a minor release may rename or remove a metric. Read the section for the version you are upgrading to before you upgrade.

## Unreleased

### Added

| Metric                      | Reports                                            |
| :-------------------------- | :------------------------------------------------- |
| `wnc_wlan_pmf_state`        | Protected management frames setting, in `state`    |
| `wnc_wlan_ft_state`         | 802.11r fast transition setting, in `state`        |
| `wnc_wlan_policy_enabled`   | Whether the bound policy profile is active         |

All three belong to the `config` module, which is disabled by default, and none of them adds a RESTCONF route. A default scrape is unchanged.

The first two carry the controller's own spelling and always have the value `1`, so `== 0` never fires and an equality match on the healthy spelling selects the healthy WLANs rather than revealing the rest. Alert on any other spelling with `state` aggregated away, as in `group by (id) (wnc_wlan_pmf_state{state!="apf-vap-pmf-required"})`. Neither is published for a WLAN whose response omits the leaf, because the leaf decodes to an empty string and an empty `state` label reads as no label at all.

`wnc_wlan_pmf_state` reports the setting that applies to a WLAN's 2.4 GHz and 5 GHz BSSes. A 6 GHz BSS requires the protection whatever this series reports, so a rule that pages on anything other than required raises a false alarm on a WLAN advertised on 6 GHz — exclude those by `id`. The error runs one way only: the series can under-report 6 GHz protection and never over-report it. The setting has three values rather than two, and the middle one admits an unprotected association, which is why the spelling is published instead of a boolean.

`wnc_wlan_policy_enabled` reads a leaf that carries no absent value, so an omitted leaf reports as shut down. Every profile read on the controller this was measured against carried the leaf on a plain read as well as on a read for the values in force — but each of them read as active, so a profile that was never switched either way was never read, and that is the one a fallback could turn into a false shutdown. Keep `unless on(job, instance) increase(wnc_refresh_defaults_fallback_total[15m]) > 0` on any rule. The series is absent, and any rule on it silent, for a WLAN that resolves to no policy profile and whenever the `wlan_cfg_entries`, `wlan_policies` or `wlan_policy_list_entries` fetch fails, so pair it with `absent()` or watch `wnc_refresh_items` for those three `data` types. Whether a shut policy profile also stops the SSID being advertised is not established, so treat the series as change detection rather than an outage signal.

### Changed

The HELP text of `wnc_refresh_defaults_fallback_total` no longer says an omitted config leaf reads as `0`. That holds for a leaf whose Go type has no absent value, while a leaf read into a `state` label drops its series instead, and the HELP now covers both.

A refresh now reads only the `data` types the enabled modules need, so the module flags bound what the controller is asked for as well as what Prometheus stores. `wnc_refresh_errors_total` is seeded for those types alone, which makes its `data` label set depend on the collector flags. With no AP module enabled, `ap_capwap_data`, `ap_oper_data`, `ap_radio_oper_data`, `ap_name_mac_map`, `ap_radio_oper_stats`, `ap_radio_reset_stats`, `rrm_measurement`, `rrm_coverage` and `rrm_ap_dot11_radar_data` have no series in either refresh metric, and a panel or rule naming one of them goes empty.

`wnc_refresh_items` is still recorded on success only, so an absent series now has two causes. A `data` type carrying a series in `wnc_refresh_errors_total` and none in `wnc_refresh_items` failed its fetch; a type carrying a series in neither is one no enabled module reads. Guard a rule that names a single `data` type with `and on(job, instance) wnc_refresh_errors_total{data="..."}`, which holds it silent where that type is never fetched.

`wnc_up` now reports `0` when every `data` type the enabled modules need failed, rather than when all eighteen failed. It read `1` before whenever any other fetch succeeded, so a deployment running one module could lose every series it asked for while `wnc_up` read `1`, and because the refresh reported no error the withhold-after-three-failures path never armed. It is still not a completeness signal: with two collectors enabled, one can fail entirely and `wnc_up` stays `1`. Alert on `increase(wnc_refresh_errors_total[15m]) > 0` for that.

`examples/prometheus_alert_rules.yml` gains that guard on `WNCAPInventoryEmpty`, which named `ap_capwap_data`. An `unless` against an absent series excludes nothing, so without the guard the rule would fire continuously wherever no AP module is enabled. The `WNCRefreshFailing` summary no longer says the controller is unreachable, because `wnc_up` now also covers a reachable controller that refuses the models one module needs.

The README now names the two mechanisms that bound what the controller is asked for. One shared refresh serves every enabled collector and runs no more often than `--wnc.cache-ttl`, which is what makes the request rate independent of the scrape interval and of how many collectors are enabled; that has always been so. What is new is the second: the refresh reads only the `data` types the enabled modules need, where before it read all of them whatever was enabled.

## v0.4.0

### Added

| Metric                                | Reports                                            |
| :------------------------------------ | :------------------------------------------------- |
| `wnc_wlan_wpa2_enabled`               | WPA2 support enabled, restored                     |
| `wnc_wlan_11k_neighbor_list_enabled`  | 802.11k neighbor list enabled, restored            |
| `wnc_refresh_defaults_fallback_total` | WLAN config fetches that fell back to a plain read |

v0.3.0 removed the first two because the controller omits a config leaf left at its default and an omitted leaf decodes to `0`, so both reported `0` on a WLAN where the feature was in force. This release asks the controller for the values in force, which is what those two series needed. A rule written against them on v0.2.0 resolves again, and its value flips from `0` to `1` wherever the default was in force.

Both restored series read a leaf on the WLAN entry itself, so no container check can guard them. A controller that answers the request for the values in force with success and then ignores it is indistinguishable on the wire: these two report `0` again while the feature is in force. On IOS-XE 17.12 both leaves can be arbitrated from the controller itself, because `show running-config all` prints the negated form for a WLAN that has the feature off, so a WLAN with no such line has it on. Compare that against these two series once per controller before writing a rule on them.

### Removed

- The floating container tag `ghcr.io/umatare5/cisco-wnc-exporter:v0` is no longer published. It moved on every 0.x release, so a deployment pinned to `:v0` adopted the metric renames and removals of v0.3.0 without any notice.
- Pin `:v0.4` to follow the current minor line, or `:v0.4.0` for an immutable tag. A deployment still pinned to `:v0` keeps running the image that tag last pointed at and stops receiving updates.

### Changed

Six WLAN series are now emitted only when the controller returned the container they are read from. A `0` from one of them no longer stands in for a container the controller never sent.

| Metric                                    | Emitted only when this container is present |
| :---------------------------------------- | :------------------------------------------ |
| `wnc_wlan_enabled`                        | `apf-vap-id-data`                           |
| `wnc_wlan_session_timeout_seconds`        | `wlan-timeout`                              |
| `wnc_wlan_central_switching_enabled`      | `wlan-switching-policy`                     |
| `wnc_wlan_central_authentication_enabled` | `wlan-switching-policy`                     |
| `wnc_wlan_central_dhcp_enabled`           | `wlan-switching-policy`                     |
| `wnc_wlan_central_association_enabled`    | `wlan-switching-policy`                     |

A rule that treats one of these series as always present needs `absent()` or `or vector(0)`, the same guard v0.3.0 introduced for a failed fetch.

The `config` series now report the values in force rather than only the values the profile set explicitly. `wnc_wlan_session_timeout_seconds`, `wnc_wlan_central_authentication_enabled` and `wnc_wlan_central_association_enabled` can therefore change value with no configuration change. On the controller this was measured against, no published series moved, because the leaves that were missing happened to hold the value already published. Do not read that as a guarantee for another controller.

`wnc_refresh_defaults_fallback_total` rises when a controller rejects that request and the exporter falls back to a plain read. While it rises the `config` series report only what the profile set explicitly, which is the v0.3.0 behaviour. Alert on `increase(wnc_refresh_defaults_fallback_total[15m]) > 0`. A zero on that counter is not proof that the request was applied, because a controller may accept it and ignore it, and nothing on the wire distinguishes that case.

Separately, the HELP text of `wnc_ap_admin_state` and `wnc_wlan_enabled` no longer equates a `0` with disabled. A `0` means the controller reported something other than the enabled spelling, or reported nothing at all. The HELP text of `wnc_client_power_save_state` no longer calls a `0` awake either. That series publishes the leaf as the controller reports it, and its value domain is undocumented.

### Fixed

- A whitespace-only `--wnc.controller` or `--wnc.access-token` is now rejected during validation instead of panicking at start-up.
- Surrounding whitespace is trimmed from `--wnc.controller` and `--wnc.access-token`, so a value read from a file or a secret no longer breaks every RESTCONF request.
- Release archives now carry `docs/*.md`, so the README links that moved into `docs/` in v0.3.0 resolve inside the tarball. The unused `*.go` entry, which matched nothing in the repository root, is gone.

## v0.3.0

### Removed

These series are gone. Where a substitute exists, the reason names it.

| Metric                               | Reason                                                                         |
| :----------------------------------- | :----------------------------------------------------------------------------- |
| `wnc_ap_rx_bytes_total`              | A frame counter times a constant 1500; use `wnc_ap_data_rx_frames_total`       |
| `wnc_ap_tx_bytes_total`              | Same, against `wnc_ap_data_tx_frames_total`                                    |
| `wnc_ap_rx_packets_total`            | Read the leaf `wnc_ap_data_rx_frames_total` reads; substitute that name        |
| `wnc_ap_tx_packets_total`            | Same, against `wnc_ap_data_tx_frames_total`                                    |
| `wnc_ap_tx_drops_total`              | Named a drop count but read `ack-failure-count` — see AP note \*4              |
| `wnc_ap_tx_errors_total`             | Read `failed-count`, already published as `wnc_ap_transmission_failures_total` |
| `wnc_ap_wep_undecryptable_total`     | Cannot leave zero unless a WLAN is configured for static WEP                   |
| `wnc_client_retry_ratio_percent`     | Summed two retry counters over a packet counter, all three still published raw |
| `wnc_wlan_rx_bytes_total`            | A sum over a changing client set is not a counter — see the note below         |
| `wnc_wlan_tx_bytes_total`            | Same                                                                           |
| `wnc_wlan_wpa2_enabled`              | Reported `0` on WLANs whose operative value is enabled — see the note below    |
| `wnc_wlan_11k_neighbor_list_enabled` | Same                                                                           |

**Per-WLAN throughput.** No series replaces the two WLAN byte counters directly. Sum the rate of the per-client counters and carry the WLAN identity from the client info metric: `sum by (wlan) (rate(wnc_client_tx_bytes_total[15m]) * on (mac) group_left (wlan) wnc_client_info)`. That needs the client collector's traffic and info modules, and `wlan` in `--collector.client.info-labels`, which the default set omits. It groups by SSID name rather than the `id` the WLAN series carried, so joining back to a WLAN series takes a second hop through `wnc_wlan_info`.

### Renamed and rescaled

Values are now ratios in `0-1`. Divide any threshold written for the old `0-100` percentages by 100.

| v0.2.0                               | v0.3.0                             |
| :----------------------------------- | :--------------------------------- |
| `wnc_ap_channel_utilization_percent` | `wnc_ap_channel_utilization_ratio` |
| `wnc_ap_cpu_utilization_percent`     | `wnc_ap_cpu_utilization_ratio`     |
| `wnc_ap_memory_utilization_percent`  | `wnc_ap_memory_utilization_ratio`  |
| `wnc_ap_noise_utilization_percent`   | `wnc_ap_noise_utilization_ratio`   |
| `wnc_ap_rx_utilization_percent`      | `wnc_ap_rx_utilization_ratio`      |
| `wnc_ap_tx_utilization_percent`      | `wnc_ap_tx_utilization_ratio`      |

### Renamed

| v0.2.0                              | v0.3.0                                             |
| :---------------------------------- | :------------------------------------------------- |
| `wnc_ap_clients_total`              | `wnc_ap_clients`                                   |
| `wnc_ap_coverage_hole_events_total` | `wnc_ap_coverage_failed_clients`, now a Gauge      |
| `wnc_ap_last_radar_on_radio_at`     | `wnc_ap_last_radar_timestamp_seconds`, now a Gauge |
| `wnc_client_rx_group_counter_total` | `wnc_client_rx_group_total`                        |
| `wnc_wlan_clients_total`            | `wnc_wlan_clients`                                 |

`wnc_ap_clients_total` read `0` on every radio in v0.2.0, because the lookup table it consulted was built empty and never filled. The successor reports a real count, so a threshold migrated along with the name has never been exercised.

### Changed

These names are unchanged, so a query keeps working and returns something different.

- A series whose source record is missing is now absent instead of `0` — because the fetch failed, or because the controller has not listed that client or radio in that data type. Queries that relied on the zero being present need `or vector(0)` or an `absent()` guard.
- `--wnc.cache-ttl` is the minimum interval between refresh completions, not a snapshot expiry. The first scrape after start-up reports `wnc_up 0` and carries no data series.
- `wnc_ap_radio_reset_total` sums every reset cause per radio, so the value steps up instead of tracking whichever cause the controller listed last.
- The `band` label on `wnc_ap_info` and `wnc_client_info` changes value and gains `unknown`. A rule that filters or groups by `band` does not error, it covers a different set.
- `wnc_ap_tx_power_dbm` and `wnc_ap_tx_power_max_dbm` are read from the record matching the operating band, so they change value on a radio supporting more than one band and are absent when no record matches.
- `wnc_ap_noise_floor_dbm` is read from the entry matching the radio's channel, so it changes value and is absent when the channel has no measurement, including a radio in monitor or sniffer mode.
- `wnc_client_protocol` returns 802.11b where it returned `unknown` or 802.11g, and no longer guesses a generation from `is-11g-client`.
- `wnc_ap_oper_state` and `wnc_client_state` are always `1` and carry the state in a `state` label. An alert such as `wnc_ap_oper_state == 0` keeps returning samples and silently stops firing.
- `wnc_ap_oper_state` also drops its `radio` label and is one series per AP, tracking the AP registration state. Per-radio down detection is `wnc_ap_radio_state == 0`, which is unchanged.
- Rewrite label matches as `wnc_ap_oper_state{state!="registered"}` and `wnc_client_state{state!="client-status-run"}`.
- A non-default `--web.telemetry-path` now moves the endpoint, so a scrape config still pointing at `/metrics` gets the landing page instead of metrics.
- Values that were accepted and then ignored are now rejected at start-up: whitespace, `{`, `?`, `#`, `%`, a path needing cleaning such as a trailing slash, and the health path.
- Setting the path to `/` replaces the landing page rather than adding a second handler to the root.

### Added

Five series report the health of the WNC data refresh itself. Without them a refresh that reached nothing yields a successful scrape carrying no series, which no alert can detect.

| Metric                                  | Reports                                                   |
| :-------------------------------------- | :-------------------------------------------------------- |
| `wnc_up`                                | Whether the last completed refresh reached the controller |
| `wnc_refresh_duration_seconds`          | Duration of the last refresh attempt                      |
| `wnc_refresh_success_timestamp_seconds` | Start time of the refresh behind the served snapshot      |
| `wnc_refresh_errors_total`              | Fetch failures per `data` type since process start        |
| `wnc_refresh_items`                     | Items the last refresh returned per `data` type           |

The other eleven names that appear in v0.3.0 are the new names in the two Renamed tables above.

### Notes

The two WLAN config booleans were removed because the controller omits config leaves left at their default values, and an omitted leaf decodes to `0`. Both series therefore reported `0` on WLANs where the feature was in fact enabled. The remaining `config` series carry the same caveat, now stated in their HELP text and in [docs/collector.wlan.md](docs/collector.wlan.md).

Binaries are published as release assets. `go install` on the module root has never worked, because the main package lives in `cmd/`.
