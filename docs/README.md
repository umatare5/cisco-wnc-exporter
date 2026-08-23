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
- The direction of the error is what makes this matter: on IOS-XE 17.12, a plain read of `wlan-cfg-entries` omitted `wpa2-enabled` and `wlan-11k-neigh-list` from exactly the WLANs where they were **on**, so a zero there reported the inverse of the setting
- Absence is per leaf, not per container: a `wlan-switching-policy` was observed carrying two of its four `central-*` leaves with the other two omitted, so a sibling series being present is no evidence that this one's leaf was sent
- Where the SDK preserves the absence, the series goes absent on its own and its HELP says so — where it does not, an omitted leaf still decodes to `0` and the series' HELP says that instead
- `wnc_refresh_defaults_fallback_total` is what tells the two apart in practice: while it is flat the exporter is reading the values in force and few leaves are omitted at all, and while it is rising the controller is refusing that request and the omissions are back

## Data refresh and caching

### WNC data refresh (`--wnc.cache-ttl`)

- A scrape is served from the last snapshot and never waits for the controller
- The flag sets the minimum idle time between refresh completions, not a snapshot expiry
- The first scrape after start-up therefore reports `wnc_up 0` and carries no data series
- A refresh is bounded at twice the flag value — data types the deadline never reached count as failures
- A refresh reads only the data types the enabled modules need, so a narrower flag set leaves more of that budget per data type
- `wnc_refresh_errors_total` names the data types a configuration reads — a type absent from both refresh series is one no enabled module reads
- Data series are withheld after three consecutive failed refreshes, so Prometheus can mark them stale
- Every read is a registered data type, so it is gated by a module flag, bounded by the refresh deadline and counted in both refresh series alike — twenty-six of the twenty-eight go through a typed SDK accessor, and the two the SDK has no route for build their path directly and check the container they were answered with, as [Controller](collector.controller.md) note *4 describes

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
- The per-radio statistics carry the instant the controller last updated them and no series publishes it, so a scrape cannot tell a record just refreshed from one untouched since before the previous scrape — a second reason for the range above
- The CleanAir readings — `wnc_ap_air_quality_index_avg` and `_min` and `wnc_ap_interferers` — are refreshed on the reporting period the controller declares for air quality and hold their value between reports, and `wnc_ap_last_air_quality_timestamp_seconds` carries the instant the controller stamped on the row all three read
- The four `wnc_rrm_worst_channel_*` series are refreshed on the same period and no leaf in their row dates it, so a value held between reports cannot be told from one just taken
- `wnc_rrm_last_rf_grouping_run_timestamp_seconds` and `wnc_rrm_last_dca_run_timestamp_seconds` date the RRM runs themselves, and the two need not advance together, so a range that catches one can miss the other — see note \*16 on the [AP](collector.ap.md) page
- Whether a reading is one the controller aggregated cannot be decided from the metric name: `_avg` and `_min` name the statistic where it is established, their absence elsewhere is not a claim that a reading is instantaneous, and where a series is one the controller aggregated its HELP says so — where neither the name nor the HELP says, this exporter has not established it

### Counter reset timing

- An AP re-joining CAPWAP does not reset the twenty per-radio counters the `traffic` and `errors` modules read from the per-radio statistics. The controller rebuilds its own copy of that record and serves every counter in it as `0` for a window well under a minute, and the record then returns to the value the AP kept counting through the event — note \*17 on the [AP](collector.ap.md) page names those twenty series and the three it does not reach
- A client's counters reset when it re-associates, because the statistics belong to the association rather than to the device
- Across 750 reads of the client traffic container, the byte and packet counters fell to zero together three times, each time in the same read that carried a new association timestamp for that client, matching the instant of the fall
- **A reset of a counter read from the per-radio statistics is anchored at its own AP's boot rather than at its CAPWAP join.** An access point left untouched while another rebooted kept every leaf of its own records and every counter anchor, so a reset that reaches one AP does not reach another's series
- Three events were separated on one access point: toggling a radio and re-joining CAPWAP each left the counters running, and only a reboot restarted them from zero, so read a reboot rather than a re-join as the reset
- Two radios of one AP can anchor at different instants. A radio on a DFS channel anchored later than the instant it was enabled while its sibling on a non-DFS channel did not, so treat the anchor as per radio even within one AP
- Only `wnc_ap_management_tx_frames_total` supports recovering an anchor by dividing the counter by its own rate. The other per-radio leaves put the anchor before the AP had booted or implausibly far back, so a reproduction that reads them will disagree
- Query them with a range long enough to absorb a re-join or a re-association, such as `increase(...[1h])`
- Gate a rule on the age of the association rather than on the counter. `wnc_ap_association_uptime_seconds` carries `mac` alone while these counters carry `mac` and `radio`, so match on `mac` and require an association older than the range plus a margin — 900 seconds for the range above and 60 for a window measured well under a minute:

```bash
rate(wnc_ap_fcs_errors_total[15m]) > 0 and on(mac) wnc_ap_association_uptime_seconds > 960
```

- Matching the range exactly is not enough: an association that has only just passed it still has the window inside the range, and the whole counter is charged as one increase. Three cautions. `wnc_ap_association_uptime_seconds` belongs to the `general` module, so a configuration enabling `traffic` or `errors` alone has nothing to gate on; it is withheld for an AP whose join time this exporter cannot use; and the two readings come from different data types, so a read taken right after a re-join can still serve the previous association and pass a spike through — note \*14 on the [AP](collector.ap.md) page records that case. The first two fail towards silence, the third towards noise

## States

### A state is a number, not a label

- Twelve families publish the number the controller's own enumeration assigns the spelling it sent, so the value is the reading and there is no `state` label to match on
- [Enumeration values](enums.md) lists every spelling and its number. A spelling absent from that page is withheld rather than published, so one subject's series can disappear while the rest publish
- `== 0` is a real comparison now, and it means a different thing per family: `client-status-idle` on `wnc_client_state`, the healthy sentinel on the four AP failure reasons, an unknown phase rather than an absence of failure on `wnc_ap_last_error_phase`, and nothing at all on `wnc_ap_oper_state`, whose enumeration declares no `0`
- Alert on any value other than the healthy one, with nothing to aggregate away:

```bash
wnc_client_state != 11
```

- The query is silent for a device the controller no longer lists, which has no series in any state — watch `wnc_refresh_items` per `data` type for that case
- The series identity no longer moves when the reading does, so `for:` accumulates across a transition — under the label shape it restarted on every state change and a stuck device never fired
- Pair the query with a `for:` longer than a legitimate transition takes
- `wnc_client_state` also covers a client held short of `client-status-run`, which no other client series does
- `wnc_ap_oper_state` is one series per AP, healthy at `4`, and carries no `radio` label
- `wnc_wlan_pmf_state` and `wnc_wlan_ft_state` are one series per WLAN and report a configured setting rather than an operational state
- The reason and phase series of the AP `join` module report the **last recorded** event rather than a current one, and freeze with the record — so they keep their reading for an AP that has left CAPWAP. All are one series per AP except the DTLS reason, which is one per tunnel channel
- The `_state` metrics outside this list carry no controller enumeration: `wnc_ap_radio_state`, `wnc_ap_admin_state` and `wnc_ap_config_state` are this exporter's own `0` or `1`, and `wnc_client_power_save_state` publishes the controller's integer unchanged

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
- Two such rows would carry the same label set, and a duplicate fails the whole `/metrics` endpoint instead of one series, so that exception protects every other metric in the scrape
- A band change takes up to `--collector.info-cache-ttl` to appear

### Multi-link clients

- An 802.11be client may hold links on more than one band at once
- The controller reports one PHY generation per client, so `band` names one link
- Aggregations such as `count by (band)` therefore undercount the other links
