# Documentation

Reference pages for cisco-wnc-exporter. The [README](../README.md) covers getting a scrape working, and these pages carry the full metric catalogue and the behaviour the collectors share.

## Collectors

| Collector                             | Focus                                              |
| :------------------------------------ | :------------------------------------------------- |
| [AP](collector.ap.md)                 | RF foundation and radio performance                |
| [Client](collector.client.md)         | User experience quality and connection performance |
| [WLAN](collector.wlan.md)             | Logical SSID performance and parameter checks      |
| [Controller](collector.controller.md) | The controller itself, with no per-device label    |

## Absence

### A leaf the controller omits is withheld, not published as zero

- A C9800 omits a leaf whose value equals its schema default, so absence on the wire is not a reading and publishing `0` or `false` for it invents one
- The direction of the error is what makes this matter: a leaf omitted because the feature is **on** reports the inverse of the setting when it is read as `0`
- Absence is per leaf, not per container, so a sibling series being present is no evidence that this one's leaf was sent
- Where the series goes absent its HELP says so, and where an omitted leaf still decodes to `0` the HELP says that instead
- `wnc_refresh_defaults_fallback_total` tells the two apart: while it is flat the exporter is reading the values in force, and while it is rising the controller is refusing that request and the omissions are back

## Data refresh and caching

### WNC data refresh (`--wnc.cache-ttl`)

- A scrape is served from the last snapshot and never waits for the controller
- The flag sets the minimum idle time between refresh completions, not a snapshot expiry
- The first scrape after start-up therefore reports `wnc_up 0` and carries no data series
- A refresh is bounded at twice the flag value — data types the deadline never reached count as failures
- A refresh reads only the data types the enabled modules need, so a narrower flag set leaves more of that budget per data type
- `wnc_refresh_errors_total` names the data types a configuration reads — a type absent from both refresh series is one no enabled module reads
- Data series are withheld after three consecutive failed refreshes, so Prometheus can mark them stale
- Every read is a registered data type, so it is gated by a module flag, bounded by the refresh deadline and counted in both refresh series alike

### Request timeout (`--wnc.timeout`)

- The flag bounds a whole RESTCONF request, from the dial to the last byte of the body
- It bounds neither the wait for the response headers nor the TLS handshake, which the SDK pins and exposes no option for
- A timeout raises `wnc_refresh_errors_total` for the data type and withholds its series, so it is visible
- It does **not** raise `wnc_refresh_defaults_fallback_total`, which counts only a controller answering `400` to the request for the values in force

### Info metric caching (`--collector.info-cache-ttl`)

- Info metrics are served from a snapshot up to the flag value old, and the collector behind them still runs on every scrape, so no controller request is saved
- A client that roamed keeps its previous `ap` label until the cache expires
- A newly associated client is missing from the info metric for up to that long, so `group_left` joins on it return nothing
- Caching does not reduce cardinality: every `ap` label value a client has held remains its own series

## Reading counters

### Controller-side update schedule

- The controller updates AP and client counters, RSSI and SNR on its own schedule rather than on scrape
- The AP profile `stats-timer` governs that schedule, and the RRM coverage, load and measurement intervals are separate from it and per band
- `show ap dot11 {24ghz | 5ghz | 6ghz} monitor` reports every one of those intervals in force
- Use a `rate()` or `increase()` range that spans several of those updates, since a shorter one carries too few to be meaningful
- No series publishes the instant a per-radio record was last updated, so a scrape cannot tell one just refreshed from one untouched since the previous scrape — a second reason for the range above
- The CleanAir readings hold their value between reports, and `wnc_ap_last_air_quality_timestamp_seconds` dates the row they read
- The four `wnc_rrm_worst_channel_*` series are refreshed on the same period with no leaf dating their row
- `wnc_rrm_last_rf_grouping_run_timestamp_seconds` and `wnc_rrm_last_dca_run_timestamp_seconds` date the RRM runs themselves, and need not advance together — see note \*16 on the [AP](collector.ap.md) page
- Only the name or the HELP says whether a reading is one the controller aggregated — `_avg` and `_min` name it, and the rest are silent on the question

### Counter reset timing

- An AP re-joining CAPWAP does not reset the per-radio counters the `traffic` and `errors` modules read — the controller serves them as `0` for a window under a minute, then returns them to the value the AP kept counting through, as note \*17 on the [AP](collector.ap.md) page describes
- A client's counters reset when it re-associates, because the statistics belong to the association rather than to the device
- **A per-radio counter is anchored at its own AP's boot rather than at its CAPWAP join**, so read a reboot rather than a re-join as the reset, and a reset reaching one AP does not reach another's series
- Two radios of one AP can anchor at different instants, so treat the anchor as per radio
- Query them with a range long enough to absorb a re-join or a re-association, such as `increase(...[1h])`
- Gate a rule on the age of the association rather than on the counter, matching on `mac` because `wnc_ap_association_uptime_seconds` carries no `radio`:

```bash
rate(wnc_ap_fcs_errors_total[15m]) > 0 and on(mac) wnc_ap_association_uptime_seconds > 960
```

- Require an association older than the range plus a margin — matching it exactly leaves the reset window inside the range, charging the whole counter as one increase
- Note \*14 on the [AP](collector.ap.md) page records the three ways that gate falls short

## States

### A state is a number, not a label

- Twelve families publish the number the controller's own enumeration assigns the spelling it sent, so the value is the reading and there is no `state` label to match on
- [Enumeration values](enums.md) lists every spelling and its number. A spelling absent from that page is withheld rather than published, so one subject's series can disappear while the rest publish
- `== 0` means a different thing per family, and [Enumeration values](enums.md) is where each `0` is named
- Alert on any value other than the healthy one, with nothing to aggregate away:

```bash
wnc_client_state != 11
```

- The query is silent for a device the controller no longer lists, which has no series in any state — watch `wnc_refresh_items` per `data` type for that case
- The series identity does not move when the reading does, so pair the query with a `for:` longer than a legitimate transition takes and it accumulates across one
- `wnc_client_state` also covers a client held short of `client-status-run`, which no other client series does
- `wnc_ap_oper_state` is one series per AP, healthy at `4`, and carries no `radio` label
- `wnc_wlan_pmf_state` and `wnc_wlan_ft_state` are one series per WLAN and report a configured setting rather than an operational state
- The reason and phase series of the AP `join` module report the **last recorded** event rather than a current one, so they keep their reading for an AP that has left CAPWAP
- The `_state` metrics outside this list carry no controller enumeration — the three AP state series are this exporter's own `0` or `1`, and `wnc_client_power_save_state` publishes the controller's integer unchanged

## Labels

### Utilization is reported as a ratio

- Metric names ending in `_ratio` carry 0 to 1, the Prometheus base unit for a percentage

### The `band` label

- AP band comes from the radio's operating band, not from the `radio` slot
- A dual band radio changes band without changing slot, so the slot is not a band
- Client band comes from the PHY generation the client associated with
- `unknown` means the controller reported a value this exporter does not map
- `unknown` is a label value, not a missing series: the other labels stay joinable
- Every `wnc_rrm_*` series is the exception: `band` is its whole identifier, so a band this exporter cannot name is withheld as a row rather than published as `unknown`
- Two such rows would carry one label set, and a duplicate fails the whole `/metrics` endpoint rather than one series
- A band change takes up to `--collector.info-cache-ttl` to appear

### Multi-link clients

- An 802.11be client may hold links on more than one band at once
- The controller reports one PHY generation per client, so `band` names one link
- Aggregations such as `count by (band)` therefore undercount the other links
