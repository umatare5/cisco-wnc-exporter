# Documentation

Reference pages for cisco-wnc-exporter. The [README](../README.md) covers getting a scrape working, and these pages carry the full metric catalogue and the behaviour the collectors share.

## Collectors

| Collector                             | Focus                                              |
| :------------------------------------ | :------------------------------------------------- |
| [AP](collector.ap.md)                 | RF foundation and radio performance                |
| [Client](collector.client.md)         | User experience quality and connection performance |
| [WLAN](collector.wlan.md)             | Logical SSID performance and parameter checks      |
| [Controller](collector.controller.md) | The controller itself, with no per-device label    |

## Data refresh and caching

### WNC data refresh (`--wnc.cache-ttl`)

- A scrape is served from the last snapshot and never waits for the controller
- The flag sets the minimum idle time between refresh completions, not a snapshot expiry
- The first scrape after start-up therefore reports `wnc_up 0` and carries no data series
- A refresh is bounded at twice the flag value — data types the deadline never reached count as failures
- A refresh reads only the data types the enabled modules need, so a narrower flag set leaves more of that budget per data type
- `wnc_refresh_errors_total` names the data types a configuration reads — a type absent from both refresh series is one no enabled module reads
- Data series are withheld after three consecutive failed refreshes, so Prometheus can mark them stale
- Every read is a registered data type, so it is gated by a module flag, bounded by the refresh deadline and counted in both refresh series alike — twenty-two of the twenty-five go through a typed SDK accessor, and the three the SDK has no route for build their path directly and check the container they were answered with, as [Controller](collector.controller.md) note *4 describes

### Request timeout (`--wnc.timeout`)

- The flag bounds a whole RESTCONF request, from the dial to the last byte of the body
- It does not bound the wait for the response headers, nor the TLS handshake: the SDK pins both at 5 seconds and exposes no option for them, so raising the flag past that does not buy more patience with a controller that is slow to begin answering
- That failure raises `wnc_refresh_errors_total` for the data type and withholds its series, so it is visible
- It does **not** raise `wnc_refresh_defaults_fallback_total`, which counts only a controller answering `400` to the request for the values in force — a header timeout carries no HTTP status at all
- On the controller this was measured against, the first byte of a WLAN config read arrived within 0.21 seconds including the request for the values in force, a factor of twenty below the pinned limit

### Info metric caching (`--collector.info-cache-ttl`)

- Info metrics are served from a snapshot up to the flag value old, and the collector behind them still runs on every scrape, so no controller request is saved
- A client that roamed keeps its previous `ap` label until the cache expires
- A newly associated client is missing from the info metric for up to that long, so `group_left` joins on it return nothing
- Caching does not reduce cardinality: every `ap` label value a client has held remains its own series

## Reading counters

### Controller-side update schedule

- The controller updates AP and client counters, RSSI and SNR on its own schedule rather than on scrape
- The AP profile `stats-timer` governs that schedule and defaults to 180 seconds
- RRM coverage updates every 180 seconds and RRM load every 60 seconds
- The RRM measurement interval is per band — 180 seconds on 2.4 and 5 GHz and 600 seconds on 6 GHz on the controller this exporter was measured against
- `show ap dot11 {24ghz | 5ghz | 6ghz} monitor` reports the coverage, load, measurement and reporting intervals in force
- Use a range of **15 minutes or more** for `rate()` and `increase()`
- A shorter range spans too few controller updates to be meaningful

### Counter reset timing

- Counters also reset when an AP re-joins CAPWAP, because the controller allocates fresh statistics
- A client's counters reset the same way when it re-associates, because the statistics belong to the association rather than to the device
- Across 750 reads of the client traffic container, the byte and packet counters fell to zero together three times, each time in the same read that carried a new association timestamp for that client, matching the instant of the fall
- Query them with a range long enough to absorb a re-join or a re-association, such as `increase(...[1h])`

## States

### A state is a label, not a number

- A series carrying a `state` label always has the value `1`, so the label is the reading
- The controller's spelling passes through unmapped, so a value this file does not name can appear
- Only the current state has a series, so `== 0` never fires, and an equality match on the healthy spelling selects the healthy devices rather than revealing the unhealthy ones
- Alert on any spelling other than the healthy one, with `state` aggregated away:

```bash
group by (mac) (wnc_client_state{state!="client-status-run"})
```

- The query is silent for a device the controller no longer lists, which has no series in any state — watch `wnc_refresh_items` per `data` type for that case
- Keeping `state` in the result restarts `for:` on every state change, so a stuck device never fires
- Pair the query with a `for:` longer than a legitimate transition takes
- `wnc_client_state` also covers a client held short of `client-status-run`, which no other client series does
- `wnc_ap_oper_state` is one series per AP, healthy at `registered`, and carries no `radio` label
- `wnc_wlan_pmf_state` and `wnc_wlan_ft_state` are one series per WLAN and report a configured setting rather than an operational state
- The reason and phase series of the AP `join` module report the **last recorded** event rather than a current one, and freeze with the record — so they keep their reading for an AP that has left CAPWAP. All are one series per AP except the DTLS reason, which is one per tunnel channel
- Every other `_state` metric keeps its numeric `0` or `1`

## Labels

### Utilization is reported as a ratio

- Metric names ending in `_ratio` carry 0 to 1, the Prometheus base unit for a percentage
- Multiply by 100 for a percentage axis, as in `wnc_ap_channel_utilization_ratio * 100`
- Grafana renders these series without that multiplication under the `percentunit` unit

### The `band` label

- AP band comes from the radio's operating band, not from the `radio` slot
- A dual band radio changes band without changing slot, so the slot is not a band
- Client band comes from the PHY generation the client associated with
- `unknown` means the controller reported a value this exporter does not map
- `unknown` is a label value, not a missing series: the other labels stay joinable
- A band change takes up to `--collector.info-cache-ttl` to appear

### Multi-link clients

- An 802.11be client may hold links on more than one band at once
- The controller reports one PHY generation per client, so `band` names one link
- Aggregations such as `count by (band)` therefore undercount the other links
