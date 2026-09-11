# Documentation

Reference pages for cisco-wnc-exporter. The [README](../README.md) covers getting a scrape working; these pages carry the catalogue and the rules every collector obeys.

| Page                                  | Focus                                              |
| :------------------------------------ | :------------------------------------------------- |
| [AP](collector.ap.md)                 | RF foundation and radio performance                |
| [Client](collector.client.md)         | User experience quality and connection performance |
| [WLAN](collector.wlan.md)             | Logical SSID performance and parameter checks      |
| [Controller](collector.controller.md) | The controller itself, such as its boot time       |
| [Exporter Health](health.md)          | The exporter's own build and refresh series        |
| [Enumeration Values](enums.md)        | The number each enumerated family reports          |
| [Help](help.md)                       | The verbatim `--help` transcript and its notes     |

## Technical Information

### Scrape Path

Every scrape is served from the last completed refresh, so no scrape waits on the controller.

- **Interval** — `--wnc.cache-ttl` is the minimum idle between refresh completions, not an expiry on the snapshot, so a faster scrape rate cannot shorten it.
- **Scope** — one refresh reads only the data types the enabled collectors need, so a client-only deployment never fetches `ap_capwap_data` and leaves more of the deadline for the rest.
- **Deadline** — a refresh is bounded at twice `--wnc.cache-ttl`, and a data type it never reached is recorded as a failure rather than published as a zero.
- **Timeout** — `--wnc.timeout` bounds a whole request, while the SDK caps response headers and the TLS handshake at five seconds each with no flag.
- **Staleness** — three consecutive refreshes in which **every** fetch failed withhold the data series; a partial failure publishes a snapshot and resets that count.

The refresh starts on the first scrape arriving after the idle elapses, so the period a dashboard sees is a multiple of the scrape interval:

```text
P = scrape_interval * ceil((cache-ttl + R) / scrape_interval)
```

- `R` — the refresh duration, which `wnc_refresh_duration_seconds` reports.
- `P` — 120s for `R` above 5s and up to 65s, at the default 55s TTL and a 60s `scrape_interval`.

### Endpoints

The exporter serves `/metrics`, `/healthz` and a landing page at `/`, and none inspects the method.

- **No 404 and no 405** — `/` matches every unclaimed path, so an unknown one returns the landing page.
- **One route or two** — setting `--web.telemetry-path` to `/` suppresses the landing page.
- **`/healthz` checks nothing** — it returns a static 200 without reading the registry, the snapshot or the controller, which makes it a liveness probe and never a readiness one.
- **`/metrics` stays 200** — an unreachable controller withholds series rather than failing the scrape; only a registry error gives 500, and an eleventh concurrent scrape gives 503.

> [!NOTE]
> Reflecting the WNC state in `/healthz` would let an orchestrator kill the exporter during a controller outage, taking the stale snapshot and the [Exporter Health](health.md) series down with it.

### Absence

A C9800 omits a leaf whose value equals its schema default, so absence on the wire is not a reading and publishing `0` for it invents one.

- **Withheld or decoded** — a series is withheld where the SDK types the leaf as a pointer, where an enum leaf arrives empty or unnumbered, and where a value-typed leaf reads a value that cannot be a measurement, as `wnc_ap_channel_energy_dbm` does for `0` and `-128`. Every remaining value-typed leaf decodes an omitted leaf as `0`, and the HELP string says so.
- **Direction** — a leaf omitted because the feature is **on** reads as its inverse when taken for `0`, and `wpa2-enabled` is absent from exactly the WLANs that enable WPA2.
- **Granularity** — absence is per leaf rather than per container, so a sibling series being present is no evidence that this one's leaf was sent.
- **Not clean air, not zero traffic** — a failed fetch withholds the series rather than publishing `0`, so a zero is never a failed fetch, though it can still be the decoded zero of an omitted value-typed leaf.
- **Sentinels** — the controller writes the 1970 epoch for an event that has not happened, and every timestamp series withholds that value rather than publishing it.
- **A `404` is a failure** — a controller or image carrying no such data type answers `404`, which counts as a failure because a path the exporter got wrong answers `404` too.

> [!IMPORTANT]
> `wnc_refresh_defaults_fallback_total` rises only where the controller refuses the request for the values in force, so a flat counter is no proof that no `0` stands for a leaf never sent.

### Info Caching

`--collector.info-cache-ttl` serves the `_info` series from a snapshot up to that old, while every other series is collected on the scrape itself.

- **No saving** — the collector behind the info series still runs on every scrape, so the cache spares the controller no request and reduces no cardinality: every label value a series has held remains its own series.
- **Stale labels** — a client that roamed keeps its previous `ap` label and a newly associated client is missing altogether, because the snapshot predates both.
- **Collapse before joining** — a `group_left` fails outright where the info side holds two series for one key, because the match group is ambiguous, and returns nothing where it lacks the join label.

### Update Schedule

The counters, RSSI and SNR are read on the controller's own schedule, not at scrape time.

- **Give a range** — `rate()` and `increase()` need a range spanning several of those updates.
- **The intervals** — `show ap dot11 {24ghz | 5ghz | 6ghz} monitor` reports the RRM coverage, load and measurement intervals in force, and the AP profile `stats-timer` sets the statistics period.
- **No record instant** — no series dates a per-radio record, so a scrape cannot tell one that just refreshed from one untouched since the previous scrape.
- **Model and release** — whether a counter leaf is maintained at all depends on the AP model and the image, so a counter reading zero is an observation rather than a platform property.

### Counter Semantics

A per-radio counter is anchored at its own AP's boot rather than at its CAPWAP join, so an AP re-joining does not reset it and a reboot does.

- **Per radio** — a reset touches that AP's series alone, and its two radios can anchor separately.
- **Per association** — a client's counters restart when it re-associates, because the statistics belong to the association rather than to the device.
- **Range** — query over a range long enough to absorb a re-join or a re-association, because a reset window left inside the range charges the whole counter as one increase.
- **Gate** — gate on the age of the association rather than on the counter, matched on `mac` because `wnc_ap_association_uptime_seconds` carries no `radio` label, and set the age above the range by a margin:

> ```text
> rate(wnc_ap_fcs_errors_total[15m]) > 0 and on(mac) wnc_ap_association_uptime_seconds > 960
> ```

### Enumerated States

Twelve families publish the number the controller's own enumeration assigns the spelling it sent, so the reading is in the value and none carries a `state` label.

- **The mapping** — [Enumeration Values](enums.md) lists every spelling with its number and names each family's `0`, which means a different thing in each.
- **Withheld spellings** — a spelling that page does not carry is withheld rather than published, so one subject's series can disappear while the rest keep publishing.
- **Alerting** — match the healthy value by equality rather than by threshold, and pair it with a `for:` longer than a legitimate transition takes.

### Dashboards

The bundled dashboards join the data series onto an `_info` series to recover a readable name.

- **Join labels** — they join on `band`, `ap`, `wlan` and `username`, which the default `--collector.*.info-labels` sets omit, so a flag has to name them.
- **The shape** — a join multiplies the data series by the info series and pulls the label across with `group_left`, so the `info` group flag behind whichever `_info` the join names has to be set as well as the data one.
