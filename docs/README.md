# Documentation

Reference pages for cisco-wnc-exporter. The [README](../README.md) covers getting a scrape working, and these pages carry the full metric catalogue and the behaviour the collectors share.

## Collectors

The four collectors focus on different aspects of the controller's operation, and each has its own module flag.

| Collector                             | Focus                                              |
| :------------------------------------ | :------------------------------------------------- |
| [AP](collector.ap.md)                 | RF foundation and radio performance                |
| [Client](collector.client.md)         | User experience quality and connection performance |
| [WLAN](collector.wlan.md)             | Logical SSID performance and parameter checks      |
| [Controller](collector.controller.md) | The controller itself metrics such as boot time    |

Additional pages supplement the metric catalogue and the README's brief command-line reference.

| Page                              | Focus                                         |
| :-------------------------------- | :-------------------------------------------- |
| [Enumeration values](enums.md)    | Numbers the twelve enumerated families report |
| [Configuration](configuration.md) | Flags, defaults and environment variables     |

## Technical Information

### Absence

A C9800 omits a leaf whose value equals its schema default, so absence on the wire is not a reading and publishing `0` or `false` for it invents one. This exporter withholds the series instead.

- **Direction** — a leaf omitted because the feature is **on** reads as its inverse when taken for `0`, and `wpa2-enabled` is absent from exactly the WLANs that enable WPA2.
- **Granularity** — absence is per leaf rather than per container, so a sibling series being present is no evidence that this one's leaf was sent.
- **Which applies** — the HELP says whether a series goes absent or still decodes an omitted leaf as `0`, and `wnc_refresh_defaults_fallback_total` rises while the controller refuses the request for the values in force.

### Data refresh and caching

Every scrape is served from the last completed refresh rather than from the controller, so no scrape waits on one. `--wnc.cache-ttl` sets the minimum idle between refresh completions, not an expiry on the snapshot.

- **Scope** — one refresh reads only the data types the enabled modules need, so a narrower flag set leaves more of the deadline for each of them.
- **Deadline** — a refresh is bounded at twice the flag value, and a data type it never reached is recorded as a failure like any other, raising `wnc_refresh_errors_total` and withholding that type's series.
- **Staleness** — after three consecutive failed refreshes the data series are withheld rather than served from a snapshot the exporter can no longer confirm.

### Info metric caching

`--collector.info-cache-ttl` serves the `_info` series from a snapshot up to that old, while every other series is collected on the scrape itself.

- **No saving** — the collector behind the info series still runs on every scrape, so the cache spares the controller no request and reduces no cardinality, because every label value a series has held remains its own series.
- **Stale labels** — a client that roamed keeps its previous `ap` label and a newly associated client is missing altogether, both for up to the flag value, so a `group_left` on one returns nothing.
- **Join labels** — the bundled dashboards join on `band`, `ap`, `wlan` and `username`, which the default label set omits, so `--collector.*.info-labels` has to name them.

### Controller-side update schedule

The AP and client counters, RSSI and SNR are read on the controller's own schedule rather than at scrape time, so give `rate()` and `increase()` a range spanning several of those updates.

- **The intervals** — the AP profile `stats-timer` sets the statistics period, the RRM coverage, load and measurement intervals are separate and per band, and `show ap dot11 {24ghz | 5ghz | 6ghz} monitor` reports every one of them.
- **No record instant** — no series dates a per-radio record, so a scrape cannot tell one that just refreshed from one untouched since the previous scrape.
- **What is dated** — `wnc_ap_last_air_quality_timestamp_seconds` dates the CleanAir row it reads, and the two `wnc_rrm_last_*_run_timestamp_seconds` series date the RRM runs themselves.

### Counter reset timing

A per-radio counter is anchored at its own AP's boot rather than at its CAPWAP join, so an AP re-joining does not reset it and a reboot does. A client's counters reset when it re-associates, because the statistics belong to the association rather than to the device.

- **Reach** — a reset touches that AP's series alone, and the two radios of one AP can anchor at different instants, so treat the anchor as per radio.
- **Range** — query over a range long enough to absorb a re-join or a re-association, because a reset window left inside the range charges the whole counter as one increase.
- **Gate** — gate on the age of the association rather than on the counter, matched on `mac` because `wnc_ap_association_uptime_seconds` carries no `radio` label, and set the age above the range by a margin:

> ```bash
> rate(wnc_ap_fcs_errors_total[15m]) > 0 and on(mac) wnc_ap_association_uptime_seconds > 960
> ```

### A state is a number, not a label

Twelve enumerated families publish the number the controller's own enumeration assigns the spelling it sent, so the reading is in the value and none of them carries a `state` label to match on.

- **The mapping** — [Enumeration values](enums.md) lists every spelling with its number and names each family's `0`, which means a different thing in each.
- **Withheld spellings** — a spelling that page does not carry is withheld rather than published, so one subject's series can disappear while the rest keep publishing.
- **Alerting** — match the healthy value by equality rather than by threshold, and pair it with a `for:` longer than a legitimate transition takes, because the identity holds while the reading moves.
