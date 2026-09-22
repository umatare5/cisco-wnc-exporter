# Exporter Health

These series describe the exporter itself rather than the wireless network.

## Metrics

No collector flag means no `refresh` series. The six register only once a collector has.

| Subsystem | Metric                                  | Type    | Description                        |
| :-------- | :-------------------------------------- | :------ | :--------------------------------- |
| build     | `wnc_build_info`                        | Gauge   | Version in `version`, always 1     |
| refresh   | `wnc_up`                                | Gauge   | Last completed refresh reached WNC |
| refresh   | `wnc_refresh_duration_seconds`          | Gauge   | Duration of the last attempt       |
| refresh   | `wnc_refresh_success_timestamp_seconds` | Gauge   | Start of the served refresh        |
| refresh   | `wnc_refresh_errors_total`              | Counter | Failed fetches per `data` type     |
| refresh   | `wnc_refresh_items`                     | Gauge   | Items the last refresh returned    |
| refresh   | `wnc_refresh_defaults_fallback_total`   | Counter | Reads that fell back to plain      |

## Labels

Every refresh family is labeled with `data` (except `wnc_build_info` and the three scalar `refresh` series).

| Label     | Value domain                                 |
| :-------- | :------------------------------------------- |
| `version` | The build version string (defaults to `dev`) |
| `data`    | One of 28 WNC data types                     |

The `data` label takes one of the 28 identifiers below, and only the types the enabled collectors read appear.

The AP collector reads `ap_capwap_data`, `ap_oper_data`, `ap_radio_oper_data`, `ap_name_mac_map`, `ap_radio_oper_stats`, `ap_radio_reset_stats`, `ap_join_stats` and `ap_geo_loc_data`. It also reads `rrm_measurement`, `rrm_coverage`, `rrm_ap_dot11_radar_data`, `rrm_radio_slot`, `rrm_main_data`, `rrm_spectrum_aq_worst_table` and `rrm_spectrum_aq_table`.

The client collector reads `client_common_oper_data`, `client_dot11_oper_data`, `client_traffic_stats`, `client_dc_info`, `client_sisf_db_mac` and `client_mm_if_client_history`. The WLAN collector reads `wlan_cfg_entries`, `wlan_policies`, `wlan_policy_list_entries` and `wlan_client_stats`. The controller collector reads `controller_boot_time`, `co_client_del_reason` and `client_roaming_stats`.

The `--collector.ap.radio` and `--collector.wlan.traffic` flags also read `client_common_oper_data`, because both count clients, so it appears with no client collector enabled.

## Annotations

Each entry adds what the HELP text does not say.

**`wnc_build_info`**

It registers before every other collector and no flag gates it. Every scrape carries it.

The two `--collector.internal.*` flags register upstream families. It is not always alone.

**`wnc_up`**

It reads 1 unless every data type the enabled collectors need failed. A refresh that lost one type of several still reports 1 while that type's series are gone.

It reads 0 before the first refresh completes. The value is a zero rather than an absence.

A panic recovered inside the refresh goroutine marks every data type failed, which forces 0.

Pair it with `wnc_refresh_errors_total`. This is the only one of the two a partial failure moves.

**`wnc_refresh_duration_seconds`**

It is absent until the first refresh records an outcome. Its absence means none has run.

It times the whole fetch loop rather than one request. It is not what `--wnc.timeout` bounds.

**`wnc_refresh_success_timestamp_seconds`**

It dates the start of the refresh behind the served snapshot rather than its completion. Taking `time()` minus this overstates the age of the data by one refresh duration.

It is absent until a refresh has published a snapshot. It keeps advancing while the staleness rule withholds the data series, which makes a withheld interval observable.

**`wnc_refresh_errors_total`**

It carries a zero series for every data type the enabled collectors need from the first scrape. A `rate()` over it needs no `absent()` guard.

It never carries a series for a data type no enabled collector reads. Silence is not failure.

It counts a data type the refresh deadline never reached as a failure. Publishing the types that arrived and fabricating the rest would read as a healthy scrape.

**`wnc_refresh_items`**

It reports the last refresh only and the map is replaced wholesale. A type that failed this time loses its series rather than holding the previous count.

It reports presence as 1 or 0 rather than a length where the data type reads a single leaf.

**`wnc_refresh_defaults_fallback_total`**

It rises only where the controller answers `400` to a read carrying `with-defaults=report-all`. A controller that answers `200` while ignoring it stays flat.

It rises by at most 2 per refresh. Only `wlan_cfg_entries` and `wlan_policies` ask for defaults.

A flat counter is therefore no proof that every WLAN config series read a leaf the WNC sent.

## Technical Notes

**The refresh series during an outage**: Every series on this page keeps publishing because the statistics ignore the withhold.

The `wnc_refresh_success_timestamp_seconds` stops advancing while `wnc_refresh_duration_seconds` keeps moving. This dates the outage and separates it from a process that never refreshed.

> [!TIP]
> The withhold is a rule about the data series. [Scrape Path](architecture.md#scrape-path) carries its count.
