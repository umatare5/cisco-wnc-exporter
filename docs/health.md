# Exporter Health

These series describe the exporter itself rather than the wireless network.

## Metrics

| Subsystem | Metric                                  | Type    | Description                        |
| :-------- | :-------------------------------------- | :------ | :--------------------------------- |
| build     | `wnc_build_info`                        | Gauge   | Version in `version`, always 1     |
| refresh   | `wnc_up`                                | Gauge   | Last completed refresh reached WNC |
| refresh   | `wnc_refresh_duration_seconds`          | Gauge   | Duration of the last attempt       |
| refresh   | `wnc_refresh_success_timestamp_seconds` | Gauge   | Start of the served refresh        |
| refresh   | `wnc_refresh_errors_total`              | Counter | Failed fetches per `data` type     |
| refresh   | `wnc_refresh_items`                     | Gauge   | Items the last refresh returned    |
| refresh   | `wnc_refresh_defaults_fallback_total`   | Counter | Reads that fell back to plain      |

> [!IMPORTANT]
> No collector flag means no `refresh` series: the six register only once a collector has.

## Labels

Three series carry a label, and none of the three identifies a device.

| Label     | Series                                          | Values                   |
| :-------- | :---------------------------------------------- | :----------------------- |
| `version` | `wnc_build_info`                                | The build version string |
| `data`    | `wnc_refresh_errors_total`, `wnc_refresh_items` | One of 28 data types     |

**`version`** carries whatever the build stamped, and an unstamped build reports `dev`.

**`data`** is closed at these 28 identifiers, and only the types the enabled collectors read appear.

- **AP** — `ap_capwap_data`, `ap_oper_data`, `ap_radio_oper_data`, `ap_name_mac_map`, `ap_radio_oper_stats`, `ap_radio_reset_stats`, `ap_join_stats`, `ap_geo_loc_data`, `rrm_measurement`, `rrm_coverage`, `rrm_ap_dot11_radar_data`, `rrm_radio_slot`, `rrm_main_data`, `rrm_spectrum_aq_worst_table` and `rrm_spectrum_aq_table`
- **Client** — `client_common_oper_data`, `client_dot11_oper_data`, `client_traffic_stats`, `client_dc_info`, `client_sisf_db_mac` and `client_mm_if_client_history`
- **WLAN** — `wlan_cfg_entries`, `wlan_policies`, `wlan_policy_list_entries` and `wlan_client_stats`
- **Controller** — `controller_boot_time`, `co_client_del_reason` and `client_roaming_stats`

> [!NOTE]
> `client_common_oper_data` is also read by `--collector.ap.radio` and `--collector.wlan.traffic`, both of which count clients, so it appears with no client collector enabled.

## Specifications

Each entry adds what the HELP text does not say.

**`wnc_build_info`**

- Registers before every other collector and no flag gates it, so every scrape carries it.
- The two `--collector.internal.*` flags register upstream families, so it is not always alone.

**`wnc_up`**

- Reads 1 unless **every** data type the enabled collectors need failed, so a refresh that lost one type of several still reports 1 while that type's series are gone.
- Reads 0 before the first refresh completes, where the value is a zero rather than an absence.
- A panic recovered inside the refresh goroutine marks every data type failed, which forces 0.
- Pair it with `wnc_refresh_errors_total`, the only one of the two a partial failure moves.

**`wnc_refresh_duration_seconds`**

- Absent until the first refresh records an outcome, so its absence means none has run.
- Times the whole fetch loop rather than one request, so it is not what `--wnc.timeout` bounds.

**`wnc_refresh_success_timestamp_seconds`**

- Dates the **start** of the refresh behind the served snapshot rather than its completion, so `time()` minus this overstates the age of the data by one refresh duration.
- Absent until a refresh has published a snapshot, and it keeps advancing while the staleness rule withholds the data series, which is what makes a withheld interval observable.

**`wnc_refresh_errors_total`**

- Carries a zero series for every data type the enabled collectors need from the first scrape, so a `rate()` over it needs no `absent()` guard.
- Never carries a series for a data type no enabled collector reads, so silence is not failure.
- Counts a data type the refresh deadline never reached as a failure, because publishing the types that arrived and fabricating the rest would read as a healthy scrape.

**`wnc_refresh_items`**

- Reports the last refresh only and the map is replaced wholesale, so a type that failed this time loses its series rather than holding the previous count.
- Reports presence as 1 or 0 rather than a length where the data type reads a single leaf.

**`wnc_refresh_defaults_fallback_total`**

- Rises only where the controller answers `400` to a read carrying `with-defaults=report-all`, so a controller that answers `200` while ignoring it stays flat.
- Rises by at most 2 per refresh: only `wlan_cfg_entries` and `wlan_policies` ask for defaults.
- A flat counter is therefore no proof that every WLAN config series read a leaf the WNC sent.

**The `refresh` series during an outage**

- Every series on this page keeps publishing, because the statistics ignore the withhold.
- `wnc_refresh_success_timestamp_seconds` stops advancing while `wnc_refresh_duration_seconds` keeps moving, which is what dates the outage and separates it from a process that never refreshed.

> [!TIP]
> The withhold is a rule about the data series — [Scrape Path](README.md#scrape-path) carries its count.
