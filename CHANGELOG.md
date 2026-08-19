# Changelog

Notable changes to the metric surface, one section per release. Release dates, downloads and the full commit list live on the [releases page](https://github.com/umatare5/cisco-wnc-exporter/releases).

This project is pre-1.0, so a minor release may rename or remove a metric. Read the section for the version you are upgrading to before you upgrade.

## Unreleased

> [!IMPORTANT]
> **BREAKING CHANGE** — twelve gauges drop their `state` label and publish the number the controller's own enumeration assigns the spelling that label carried. All twelve are listed below; **every name, type and remaining label is unchanged**, so a query keeps working and returns something different — a selector on `state` matches nothing, a negated selector matches everything, and a `sum()` that counted series now adds enumeration values. [docs/enums.md](docs/enums.md) lists every spelling with its value.

### Added

Three families arrive behind flags that already exist: one joins the AP `spectrum` module and two the AP `radio` module, so a configuration that already enables those gains series without a flag change. One new read comes with them, and its cost does not grow with the number of APs.

- `wnc_ap_last_air_quality_timestamp_seconds{mac,radio}` reports the instant the controller stamped on the CleanAir row that `wnc_ap_air_quality_index_avg`, `wnc_ap_air_quality_index_min` and `wnc_ap_interferers` already read, so it adds no request and it is one series per CleanAir-reporting radio — the same cardinality as the average. An instant that does not advance means the reading is held from an earlier report, which no value can show: **the readings and the instant froze together** over a window of about twelve minutes. It is withheld where the controller carries no instant this exporter can use, so this one series can be absent while the three readings publish — note \*11 on the [AP](docs/collector.ap.md) page carries the rest.
- `wnc_rrm_last_rf_grouping_run_timestamp_seconds{band}` and `wnc_rrm_last_dca_run_timestamp_seconds{band}` date the RF grouping run and the channel assignment the controller last ran for a band. They carry `band` alone because that is the whole identifier the controller gives the record, they join the AP `radio` module, and each is one series per band. They add the `rrm_main_data` read, whose cost does not grow with the number of APs. **The two instants need not advance together** — the grouping instant advanced between two reads while the DCA instant did not move on any band — so a range that catches one can miss the other, and **a band whose PHY type this exporter cannot name is withheld as a row rather than labelled `unknown`**.

### Changed

Twelve families drop the `state` label and publish the number the controller's own enumeration assigns each spelling. Every name, type and subject is unchanged, and the table gives each family's label set and value range — [docs/enums.md](docs/enums.md) carries all 221 spellings, the six schema modules that declare them and the revision each was read at. Cardinality is unchanged for every spelling this release carries: nine series per AP, up to two per client and up to two per WLAN, each one label pair smaller.

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

- **A negated label matcher inverts, and it does so silently.** `{state!="..."}` selected the unhealthy subjects; with the label gone it matches every series the metric publishes, because a matcher of that form also matches a series that carries no such label. An alert of that shape fires on the whole estate at its next evaluation and a panel of that shape draws every subject. Compare the value instead — `wnc_client_state != 11` for a client short of the run state, `wnc_ap_oper_state != 4` for an AP that is not registered — and drop the `group by` that aggregated `state` away: the series identity no longer moves when the reading does, so `for:` accumulates across a transition rather than restarting on each one.
- **A `sum` over one of the twelve has to become a `count`.** Each series was `1`, so `sum by (...)` counted subjects; it now adds enumeration values. An expression of the form `sum by (...) (wnc_client_state * on(mac) group_left(...) wnc_client_info)` counted clients while both sides were `1`, and because the info series exists only for a client in the run state the total is now exactly eleven times the count. `count by (...)` is what counts them, and under `max_over_time` the multiplier is not even constant — a client that passed through `client-status-delete-in-progress` or `client-status-deleted` inside the range contributes `12` or `13`.
- **A `== 1` guard and a presence join both change what they mean.** `== 1` was always true, so `and on(mac) wnc_ap_oper_state == 1` meant "this AP reports a state"; it now means `ap-down`, and the guard that survives is the bare `and on(mac) wnc_ap_oper_state`. A `* on(...) group_left(...)` join still returns its series, but the product now carries the enumeration value rather than a `1`, so a panel whose value mapping covers only `0` and `1` no longer names what the series reports — on `wnc_ap_oper_state` such a mapping names `ap-down` and nothing else, because the healthy `registered` is `4`.
- **`== 0` becomes meaningful, and not uniformly.** It never fired before. It is now the healthy sentinel on the discovery, join, configuration and DTLS reasons, `ap-reboot-reason-none` on the reboot reason, `dot11-roam-type-none` on `wnc_client_roam_type`, `client-status-idle` on `wnc_client_state` and the disabled setting on the two WLAN families — but an unknown phase rather than an absence of failure on `wnc_ap_last_error_phase`, the enumeration's own unknown member on `wnc_ap_last_disconnect_reason`, and no match at all on `wnc_ap_oper_state`, whose enumeration declares no `0`. A rule copied from one family to another is wrong on arrival.
- **A spelling this release does not number is withheld.** No value is free to stand for one, so the series is absent for that subject rather than carrying a wrong number, and the spelling reaches the log once at `--log.level=debug` and nowhere else. `absent()` does not detect that loss for a per-subject family, because one publishing subject keeps it silent — cross-check against a series of the same module that is published unconditionally, as in `wnc_ap_config_state unless wnc_ap_oper_state` or `wnc_ap_joined unless wnc_ap_last_reboot_reason`. The client and WLAN families have no exact equivalent, so treat an absent series there as a reading that has to be recovered from the log.

### Fixed

- The [Overview](docs/README.md) page said an AP re-joining CAPWAP resets its counters, and told the reader not to read the AP's boot time as the anchor. Both are corrected. Measured on one access point with a CAPWAP reset and then a reboot: the re-join left the per-radio counters running and only the reboot restarted them from zero, so **the anchor is the AP's own boot**. The v0.10.0 entry below carried the same claim and is corrected in place.
- The same page and note \*17 on the [AP](docs/collector.ap.md) page now record a gap a rule can hit. For a window well under a minute after a re-join the controller serves the per-radio statistics record with **every counter in it at zero**, and the return to a real value is charged as one increase over any range spanning it. Twenty series per radio are reached — the AP `traffic` and `errors` counters — and three of the `errors` module are not. The page carries a gate on `wnc_ap_association_uptime_seconds` for the rules that are.
- The HELP of `wnc_ap_radio_resets_total` was "Radio reset count" and said nothing about the fall this counter can take. The controller keeps the count per cause and deletes cause entries, and the total then falls; every fall observed happened together with the access point booting or re-joining. The HELP now says so, and note \*18 on the [AP](docs/collector.ap.md) page records why the cause is not a label and that no leaf carries the last cause.
- Note \*8 on the [AP](docs/collector.ap.md) page said to read every `channel="data"` DTLS zero as not configured. Enabling DTLS on the data channel advanced two of the five counters and left three at zero, and it published the `channel="data"` entry of `wnc_ap_last_dtls_success_timestamp_seconds`, which disabling it again did not withdraw. A zero on the three that never moved is now called **unconfirmed** rather than not configured.
- Note \*1 on the [WLAN](docs/collector.wlan.md) page named no path that returns `wnc_wlan_data_usage_bytes_total` to zero. An administrative shutdown of the WLAN did, and re-enabling it started the count again from zero, so `increase()` over a range spanning the shutdown loses everything counted before it. Whether any other event zeroes it was not measured.
- Note \*15 on the [AP](docs/collector.ap.md) page and the [Overview](docs/README.md) page called the band-worst series the release's one new read and the one exception to the `band` rule. This release adds a second read and two more band-keyed series, so both now read as a class rather than as an exception of four.

## v0.10.0

> [!IMPORTANT]
> **BREAKING CHANGE** — this release renames one gauge and withholds twenty-one per-radio series that previously reported a fabricated `0` for a slot that is not a radio. Both are listed below; a rule or panel touching `wnc_ap_air_quality_index`, or reading the AP `traffic` and `errors` counters or `wnc_ap_clients` as one series per slot, needs a change.

### Renamed

The series does not change type, labels or value, so a rule or panel needs only the new
name. The controller renders this reading as an average over its own reporting period and
names the minimum of that period beside it, so the name now carries which statistic it is.

| v0.9.2                      | v0.10.0                         |
| :-------------------------- | :------------------------------ |
| `wnc_ap_air_quality_index`  | `wnc_ap_air_quality_index_avg`  |

### Added

Nine families arrive behind flags that already exist: six join the AP `spectrum` module, one the AP `radio` module, one the AP `general` module and one the client `general` module, so a configuration that already enables those gains series without a flag change. One new read comes with them, and its cost does not grow with the number of APs.

- `wnc_ap_air_quality_index_min{mac,radio}` and `wnc_ap_interferers{mac,radio}` report the lowest air quality of the reporting period and the interference devices CleanAir attributes to the channel the radio operates on. Both read the row `wnc_ap_air_quality_index_avg` already reads, so they add no request and the three are published and withheld together — note \*11 on the [AP](docs/collector.ap.md) page lists the cases. **A zero from `wnc_ap_interferers` is a reading rather than a fabricated one**: neither series is guarded on its value.
- `wnc_ap_channel_energy_dbm{mac,radio}` reports the energy the controller measured on the channel it assigned a radio, from the DCA statistics `wnc_ap_channel_changes_total` already reads, so it joins the AP `radio` module and adds no request. It reaches radios the air quality series cannot, being reported for a radio the CleanAir table has no record for at all. **It is a step**, recomputed when DCA next runs for the band, and two values the leaf can carry are withheld as impossible measurements.
- `wnc_ap_association_uptime_seconds{mac}` counts from the start of the CAPWAP association an AP currently holds, beside `wnc_ap_uptime_seconds`, which counts from its boot. It joins the AP `general` module and adds no request. **Neither series measures how long an AP was gone** — the controller replaces the record in place across a reboot — so the AP page gives the re-join detection to use instead.
- `wnc_client_roam_type{mac,state}` reports how a client reached the association it currently holds, in a `state` label with the value `1`. It joins the client `general` module and reads the record `wnc_client_state_transition_seconds` already fetches, so it adds no request. It is a property of that association rather than a count, and **no series counts roams per client**.
- `wnc_rrm_worst_channel_air_quality_index_avg{band}`, `wnc_rrm_worst_channel_air_quality_index_min{band}`, `wnc_rrm_worst_channel_interferers{band}` and `wnc_rrm_worst_channel_number{band}` report the channel the controller ranks worst in each band. They carry `band` alone because no AP keys the controller's table, they join the AP `spectrum` module, and they add the one new read. **A band this exporter cannot name is withheld as a row rather than labelled `unknown`**, the one exception to the `band` rule on the [Overview](docs/README.md) page, and **a controller that does not carry the table answers `404`, which is a failure rather than an absence**.

### Fixed

- The AP `traffic` and `errors` counters and `wnc_ap_clients` are now **absent** for a slot that is not a radio, instead of reporting `0` on twenty-one series. The slot list carries entries that are not radios — a remote-LAN port arrives with neither state leaf — and the controller does send a counter record for such an entry with every counter in it zero, so the series reported a radio that carries no traffic and never errs. This extends to the whole per-radio set the rule v0.9.0 applied to `wnc_ap_radio_state` and `wnc_ap_admin_state`. A rule or panel that expected one series per slot needs `absent()` or `or vector(0)`, and a `sum()` over a controller carrying such a port now reads lower.
- `wnc_client_state_transition_seconds` is now **absent** in the two shapes the controller uses to say it measured no transition — a mobility history with no entry, and a first entry whose latency reads zero — instead of publishing `0`. A zero reports an instant transition, which is a measurement rather than the absence of one, so a rule reading the series as a latency saw the best possible value where there was none. The [Client](docs/collector.client.md) page records both shapes.

- The HELP of the air quality average and note \*11 on the [AP](docs/collector.ap.md) page said the value is an average over a window the controller does not declare. The controller declares it: `show ap dot11 <band> cleanair config` reports the air quality reporting period, no configuration leaf carries it, and an operator can change it. Both now name it, and both now give the direction of the scale — a higher index is cleaner, and the controller's own alarm threshold is a lower bound on it. The v0.9.0 entry below carried the same claim and is corrected in place.
- The [Overview](docs/README.md) page said a counter reset can reach an AP that went nowhere, and note \*4 on the [AP](docs/collector.ap.md) page said the channel-change counter falls to zero on every radio at once. Re-measured against an access point left untouched as a control, neither holds: the untouched AP kept every counter and every anchor, and the count fell only on the radio that had one to lose. **A per-radio counter is anchored at its own AP**, so a rule watching one AP does not see a reset caused elsewhere. A reset is anchored at that AP's boot rather than at its CAPWAP join: this entry originally put it at the join and said the boot time is not the anchor, and a CAPWAP reset measured against a reboot on the same access point separated the two. Two radios of one AP can anchor at different instants. The v0.9.0 entry below drew the withdrawn conclusion and is corrected in place.
- Note \*3 on the [Controller](docs/collector.controller.md) page recommended alerting on a fast 802.11i counter at zero while the slow one advances. A roam the total counts need not fall into either sub-counter, so that shape is withdrawn from the note and from the HELP of the fast counter. What replaces it is narrower: the total advances on a roam and not on a fresh association.
- Note \*4 on the [AP](docs/collector.ap.md) page said the omission of a verdict leaf runs the error one way only. Just after a radio re-joins, all four verdicts can be present and passing while the channel energy in the same record reads unmeasured, so a verdict can also report a pass that was never measured. The note's `>= 2` recipe now excludes the `load` profile, whose verdict is judged against a configured client count, and points at `wnc_ap_clients` and `wnc_ap_channel_utilization_ratio` for capacity instead.
- The [Client](docs/collector.client.md) page said the per-client roam type is unpublished. The controller reports two roam-type leaves: the one this release publishes as `wnc_client_roam_type`, and the mobility manager's own, which stays unpublished because it read the same value for every client.

## v0.9.2

No metric, label, flag or HELP string changes: this release carries a test fix only.

### Fixed

- The test that pinned the refresher's snapshot ordering bounded its own gate, which bounds a race rather than removing it: a stall longer than the bound reopens it, and the assertion it then fails names the wrong cause. The racy assertion moves out of the test named for starting a refresh, and `TestRefresher_Get_DoesNotWaitForTheRefresh` now owns the contract `get()` documents — that a scrape never waits for the controller. A `get()` that runs the refresh itself is reported as the blocking it is rather than as an early publication. No production code changed.

## v0.9.1

No metric, label, flag or HELP string changes: this release carries a test fix only.

### Fixed

- A unit test for the background refresher asserted an ordering the code does not provide, so it failed at a low rate in CI and passed on a re-run. `get()` starts the refresh in a goroutine and then reads the published snapshot, and nothing orders the two, so a refresh that completed first made the test see the snapshot it asserts is absent. Under the flags the coverage job uses it failed 8 times in 20000 iterations before the change and none after. The test now holds the refresh until the first `get()` has returned, bounded so a `get()` regressed to refresh synchronously fails the assertion rather than hanging. No production code changed.

## v0.9.0

> [!IMPORTANT]
> **BREAKING CHANGE** — this release renames two counters and withholds two series that previously reported a fabricated `0`. Both are listed below; a rule or panel touching `wnc_ap_fragmentation_*`, `wnc_ap_radio_state` or `wnc_ap_admin_state` needs a change.

### Renamed

Neither counter changes type, labels or value, so a rule or panel needs only the new name. Both now name the noun the leaf counts and put the direction first, which is what every other directional series here does.

| v0.8.0                          | v0.9.0                      |
| :------------------------------ | :-------------------------- |
| `wnc_ap_fragmentation_rx_total` | `wnc_ap_rx_fragments_total` |
| `wnc_ap_fragmentation_tx_total` | `wnc_ap_tx_fragments_total` |

> [!NOTE]
> The receive side carries a caveat the new name does not resolve. The leaves are `rx-fragment-count` and `tx-fragment-count`, but note \*3 on the [AP](docs/collector.ap.md) page records the controller calling the receive one an incomplete-fragment counter, and that wording could not be re-checked: no `show` command was found that reports the per-radio frame counters. Both counters read zero on every radio measured, so the distinction has no observable consequence today.

### Added

- `wnc_wlan_onboarding_clients{id,phase}` reports how many clients each WLAN currently holds in an onboarding phase, over the four phases `l2auth`, `mobility`, `iplearn` and `webauth_pending`. It joins the WLAN `traffic` module and adds no request: the record it reads is the one `wnc_wlan_data_usage_bytes_total` already fetches. It closes the gap the WLAN page documents against `wnc_wlan_clients`, which counts the run state only and therefore **falls** while clients pile up short of it. The counts are current rather than cumulative — see note \*2 on that page for what they detect, what they do not, and where a failure that has already completed is counted instead.
- `wnc_ap_rrm_profile_passed{mac,radio,profile}` reports the controller's RRM verdict per radio over the four profiles `coverage`, `load`, `interference` and `noise`, where `1` is a pass. It joins the existing AP `radio` module — its key is the same one every radio series already carries — and adds one request per refresh. **Three of the four have been observed failing in ordinary conditions**, so a rule of the form `== 0` fires from the first scrape; note \*4 on the [AP](docs/collector.ap.md) page gives two shapes that do not. The controller's own CLI agrees verdict for verdict on every radio checked.
- `wnc_ap_channel_changes_total{mac,radio}` counts the channel changes the controller records for a radio, from its DCA assignment statistics. It comes from the same record as the RRM verdicts, so it joins the AP `radio` module and adds no request. **What it counts is not established** — the controller keeps a separate count of radar-driven changes that no leaf carries, and both read zero here. Where `wnc_ap_last_radar_timestamp_seconds` is absent no radar is on record, which is the most the pair supports — note \*6 on the [AP](docs/collector.ap.md) page. **It resets**, so read it with `rate()` rather than as a lifetime total.
- `wnc_ap_air_quality_index{mac,radio}` reports the CleanAir air quality of the channel the radio operates on, behind the new `--collector.ap.spectrum` flag, off by default. **The value is an average over the air quality reporting period the controller declares, and a higher index is cleaner.** The table it reads is the largest of the RRM reads and grows with the number of CleanAir APs, which is why it has its own flag and is fetched last. **The series is absent, not zero, wherever the reading cannot be reached** — the cases are listed in note \*11 on the AP page — so silence does not mean clean air.

### Changed

- `--collector.ap.spectrum` is the only new flag. The two other new readings join existing modules, so a configuration that already enables `--collector.ap.radio` or `--collector.wlan.traffic` gains series without a flag change; `--collector.ap.radio` also gains one request per refresh.
- `--collector.controller.general` describes itself as `Enable Controller general metrics`, which is how the same flag reads on the other three modules. No behaviour change.

### Fixed

- `wnc_ap_radio_state` and `wnc_ap_admin_state` are now **absent** for a slot whose state leaf the controller omits, instead of reporting `0` — "radio down" and "admin disabled". The slot list carries entries that are not radios: a remote-LAN port arrives with both leaves omitted, measured on a controller, and the string comparison published a permanently failing radio for it. A rule of the form `wnc_ap_radio_state == 0` therefore fired on every AP carrying such a port. This applies to these two series the rule `wnc_ap_oper_state` has always followed. A rule that treated either series as always present needs `absent()` or `or vector(0)`.
- The [Overview](docs/README.md) page records a counter reset that reaches an AP which went nowhere. Both APs stayed joined and the controller's boot time did not change, yet the per-radio reset totals and the channel-change counts fell on every radio together. The existing guidance only covered the counters of the AP or client that re-joined. **The conclusion drawn here — that the controller rebuilds the per-radio statistics tables rather than one AP's — is withdrawn in v0.10.0**, which re-measured it against an access point left untouched as a control.
- Three cross-references on the [AP](docs/collector.ap.md) page, and one in the v0.7.0 section below, pointed at the wrong note: this release inserts a note into that page and every number after it moved by one.
- The [Client](docs/collector.client.md) page lists `wnc_client_rx_group_total` among the error counters observed at zero. The page's own evidence block already showed it at zero and the list omitted it; a second measurement on every client of the controller agreed.

## v0.8.0

### Renamed

Neither counter changes type, labels or value, so a rule or panel needs only the new name. Both take the plural the controller's own naming implies: it spells a counter carrying an explicit count noun in the singular (`rts-success-count`) and one carrying none in the plural (`resets`), and this exporter drops the count noun.

| v0.7.0                     | v0.8.0                       |
| :------------------------- | :--------------------------- |
| `wnc_ap_rts_success_total` | `wnc_ap_rts_successes_total` |
| `wnc_ap_radio_reset_total` | `wnc_ap_radio_resets_total`  |

### Changed

- Four HELP strings on the [Controller](docs/collector.controller.md) page no longer call the counters cumulative since the controller booted. Neither container reports an epoch leaf, and no read caught the three roam counters moving, so `wnc_controller_boot_time_seconds` is a reset anchor rather than the instant the counts run from. The v0.7.0 section below states the stronger claim — read this one instead. No name, type, label or value changes.
- Five panels of the example admin dashboard no longer filter AP-level series by `radio`. Those series carry `mac` alone, so the filter matched nothing whenever a single radio was selected and matched everything only on `All`.

### Fixed

- `wnc_client_uptime_seconds` is now absent for a client whose record carries no association time, and for the epoch the controller writes where an event has not happened. It measured from a zero timestamp instead, which saturates and reads as a session some 292 years long, so a rule of the form `wnc_client_uptime_seconds > 86400` fired on it. This applies to the client collector the rule `wnc_ap_uptime_seconds` already follows. A rule that treated the series as always present needs `absent()` or `or vector(0)`. The HELP now says so.

## v0.7.0

### Added

The AP collector gains a `join` module behind `--collector.ap.join`, disabled by default like every other module. It reads one new `data` type, `ap_join_stats`, and reads nothing else — a deployment enabling only this module makes one request per refresh.

| Metric                                               | Reports                                            |
| :--------------------------------------------------- | :------------------------------------------------- |
| `wnc_ap_joined`                                      | Whether a CAPWAP session is held now               |
| `wnc_ap_join_info`                                   | AP name from the join record, always `1`           |
| `wnc_ap_{discovery,join,config}_*_total`             | Requests, responses and failures of each phase     |
| `wnc_ap_dtls_*_total`                                | DTLS setup, decrypt and anti-replay, per `channel` |
| `wnc_ap_last_*_timestamp_seconds`                    | Last success and failure of each phase             |
| `wnc_ap_last_*_reason` and `wnc_ap_last_error_phase` | Each enumeration in a `state` label                |

A new **controller collector** behind `--collector.controller.general` reports the controller rather than a device, so none of its series carries an identifying label. It is documented in [docs/collector.controller.md](docs/collector.controller.md).

| Metric                                       | Reports                                            |
| :------------------------------------------- | :------------------------------------------------- |
| `wnc_controller_boot_time_seconds`           | Unix time of the last boot                         |
| `wnc_controller_client_deletes_total`        | Client deletions per `reason`, several hundred     |
| `wnc_controller_client_ap_auth_*roams_total` | Roams on the AP-authenticated path, three counters |

The boot time is the epoch the other counters are read against, which is why one flag enables all five: behind a second flag, a rule of the form `and on() (time() - wnc_controller_boot_time_seconds > 3600)` would return nothing rather than fire. It is withheld rather than reported as `0` when the controller does not carry the leaf, and it moves by a second between reads, so compare it against a threshold and not with `changes()`.

The delete reasons are controller-wide with no per-client, per-AP or per-WLAN equivalent anywhere in the operational data, so a rise says what happened and never to whom. The three roam counters cover the FlexConnect local-authentication path only, and the two `dot11i` counters are not a partition of the total.

None of the three containers behind this collector has a route in the SDK, so all three are read by building the RESTCONF path directly. **A controller or image that does not carry one answers `404`, which is a failure rather than an absence** — note *4 on that page gives the rule expression that excludes them.

`wnc_wlan_policy_binding` joins the WLAN `config` module, one series per binding carrying `id`, `policy_profile` and `policy_tag`. The six policy-derived config series name none of those, so where a WLAN is bound through more than one policy tag they report one of the profiles and nothing said which; this makes that state observable. It adds no request, coming from the two reads those six already use.

`wnc_wlan_data_usage_bytes_total` joins the WLAN `traffic` module, reporting the controller's own byte total per WLAN in both directions. The unit was established by measurement rather than by the model, which reports none. It keeps the bytes of clients that have since disconnected, so it is not the sum of the per-client counters and a WLAN with no clients can carry a large unmoving value.

Thirty-two metrics in all — more series than that per AP, because the five DTLS counters and the DTLS reason carry one series per tunnel channel — listed individually in [docs/collector.ap.md](docs/collector.ap.md). **The statistics list keeps a record for an AP that has left CAPWAP**, which is what the module is for: every other AP series is read from the AP inventory and disappears with it, so nothing distinguished a departed AP from a failed fetch. The signal that becomes available is `rate(wnc_ap_discovery_requests_total[15m]) > 0 and wnc_ap_joined == 0`, an AP that reaches the controller and cannot complete a join, and it works because the discovery counters keep advancing while the session is gone.

No series carries the AP name as a label. The counters and the join state carry `mac` alone, so a bare `and` between them matches — it requires identical label sets on both sides — and renaming an AP cannot start a fresh counter series. The DTLS series add `channel`, the reason series add `state`, and the name is published as `wnc_ap_join_info{mac,name}`, because the AP inventory no longer names a departed AP. Read note *7 on that page before writing a rule: the query has a known false positive for an AP that holds this controller as its secondary.

A timestamp leaf carrying the controller's epoch sentinel is withheld rather than published as an instant in 1970, so the failure timestamps are absent on a controller where nothing has failed.

`examples/prometheus_alert_rules.yml` gains three rules for the series above — `WNCAPNotJoining`, `WNCWLANPolicyBindingAmbiguous` and `WNCControllerBootTimeMissing` — each with a fixture case that fails if the expression is written the obvious wrong way. The `WNCAPInventoryEmpty` description no longer claims that every AP series is labeled from the AP inventory, which the join module makes false.

### Fixed

- `wnc_ap_uptime_seconds` is now absent for an AP whose boot time the controller does not report, reports in a form this exporter cannot parse, or reports at the Unix epoch. It published `0` for both, which reads as an AP that booted at the instant of the scrape, so a rule of the form `wnc_ap_uptime_seconds < 600` fired on it. `wnc_ap_oper_state` is already withheld when the leaf it reads is empty, so this applies the rule the same collector already followed, and the epoch case matches what the same collector does with the radar timestamp. A rule that treated the series as always present needs `absent()` or `or vector(0)`. The HELP now says so.

## v0.6.0

No metric was added, renamed or removed, and no series changed what it reads from the controller.

### Fixed

- The `linux/amd64` image can execute. Its binary was linked against a host loader that a `FROM scratch` image does not carry, so no `linux/amd64` image this project published before this release could start. The `linux/arm64` image was unaffected.
- `docker run` with no argument now starts the exporter. The image carried `CMD ["--help"]`, so the quickstart form in the README printed the flag list and exited instead of serving `/metrics`.

## v0.5.0

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

The `config` series now report the values in force rather than only the leaves whose value differs from the default. `wnc_wlan_session_timeout_seconds`, `wnc_wlan_central_authentication_enabled` and `wnc_wlan_central_association_enabled` can therefore change value with no configuration change. On the controller this was measured against, no published series moved, because the leaves that were missing happened to hold the value already published. Do not read that as a guarantee for another controller.

`wnc_refresh_defaults_fallback_total` rises when a controller rejects that request and the exporter falls back to a plain read. While it rises the `config` series report only the leaves whose value differs from the default, which is the v0.3.0 behaviour. Alert on `increase(wnc_refresh_defaults_fallback_total[15m]) > 0`. A zero on that counter is not proof that the request was applied, because a controller may accept it and ignore it, and nothing on the wire distinguishes that case.

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
| `wnc_ap_tx_drops_total`              | Named a drop count but read `ack-failure-count` — see AP note \*5              |
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
