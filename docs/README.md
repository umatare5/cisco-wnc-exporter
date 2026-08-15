# Documentation

Reference pages for cisco-wnc-exporter. The [README](../README.md) covers getting a scrape working, and these pages carry the full metric catalogue and the behaviour the collectors share.

## Collectors

| Collector                     | Focus                                              |
| :---------------------------- | :------------------------------------------------- |
| [AP](collector.ap.md)         | RF foundation and radio performance                |
| [Client](collector.client.md) | User experience quality and connection performance |
| [WLAN](collector.wlan.md)     | Logical SSID performance and parameter checks      |

## Data refresh and caching

### WNC data refresh (`--wnc.cache-ttl`)

- A scrape is served from the last snapshot and never waits for the controller
- The flag sets the minimum idle time between refresh completions, not a snapshot expiry
- The first scrape after start-up therefore reports `wnc_up 0` and carries no data series
- A refresh is bounded at twice the flag value — data types the deadline never reached count as failures
- Data series are withheld after three consecutive failed refreshes, so Prometheus can mark them stale

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
- Query them with a range long enough to absorb a re-join, such as `increase(...[1h])`

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
