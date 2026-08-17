# WLAN collector

WLAN collector focuses on logical SSID performance and parameter checks.

## Metrics

| Module  | Metric                                    | Type    | Description                          |
| :------ | :---------------------------------------- | :------ | :----------------------------------- |
| general | `wnc_wlan_enabled`                        | Gauge   | WLAN status                          |
| traffic | `wnc_wlan_clients`                        | Gauge   | Run-state clients count (calculated) |
| traffic | `wnc_wlan_data_usage_bytes_total`         | Counter | Bytes in both directions **(\*1)**   |
| traffic | `wnc_wlan_onboarding_clients`             | Gauge   | Clients held in a phase **(\*2)**    |
| config  | `wnc_wlan_auth_psk_enabled`               | Gauge   | PSK authentication enabled           |
| config  | `wnc_wlan_auth_dot1x_enabled`             | Gauge   | 802.1x authentication enabled        |
| config  | `wnc_wlan_auth_dot1x_sha256_enabled`      | Gauge   | 802.1x SHA256 auth enabled           |
| config  | `wnc_wlan_wpa2_enabled`                   | Gauge   | WPA2 support enabled                 |
| config  | `wnc_wlan_wpa3_enabled`                   | Gauge   | WPA3 support enabled                 |
| config  | `wnc_wlan_session_timeout_seconds`        | Gauge   | Session timeout duration             |
| config  | `wnc_wlan_load_balance_enabled`           | Gauge   | Load balancing enabled               |
| config  | `wnc_wlan_11k_neighbor_list_enabled`      | Gauge   | 802.11k neighbor list enabled        |
| config  | `wnc_wlan_client_steering_enabled`        | Gauge   | 6GHz client steering enabled         |
| config  | `wnc_wlan_central_switching_enabled`      | Gauge   | Central switching enabled            |
| config  | `wnc_wlan_central_authentication_enabled` | Gauge   | Central authentication enabled       |
| config  | `wnc_wlan_central_dhcp_enabled`           | Gauge   | Central DHCP enabled                 |
| config  | `wnc_wlan_central_association_enabled`    | Gauge   | Central association enabled          |
| config  | `wnc_wlan_policy_enabled`                 | Gauge   | Bound policy profile is active       |
| config  | `wnc_wlan_pmf_state`                      | Gauge   | PMF setting **(\*3)**                |
| config  | `wnc_wlan_ft_state`                       | Gauge   | 802.11r fast transition setting      |
| config  | `wnc_wlan_policy_binding`                 | Gauge   | Policy tag binding **(\*4)**         |

## Notes

`wnc_wlan_clients` counts only the clients the controller reports in the run state, so it does not count a client held short of it. During an onboarding failure the count therefore **falls** while clients pile up in an earlier phase, which is the opposite of what a rule written against a client-count rise expects. `wnc_wlan_onboarding_clients` counts those clients per phase, and `wnc_client_state` names them individually.

Every WLAN module reads `wlan-cfg-entries`, and `config` also reads `wlan-policies` and `policy-list-entries`. The exporter asks the controller for the values in force on the first two, because a plain read omits every leaf whose value equals its default, whether the profile set that value or never touched it — observed on IOS-XE 17.12. The same criterion reaches a whole container, so a container missing from a plain read can mean every leaf in it is at its default rather than that the feature is off. A controller that answers `400` is read plainly instead, and `wnc_refresh_defaults_fallback_total` rises for as long as that lasts. The `traffic` module also reads the client list, and `wnc_wlan_clients` is withheld for every WLAN when that fetch fails.

`wnc_wlan_pmf_state` and `wnc_wlan_ft_state` report the controller's own spelling in the `state` label and always have the value `1`, so `== 0` never fires — see [States](README.md#a-state-is-a-label-not-a-number). Neither is published for a WLAN whose response omits the leaf, because the leaf decodes to an empty string and an empty `state` label reads as no label at all.

`wnc_wlan_policy_enabled` reads the `status` leaf of the policy profile the WLAN resolves to through a policy tag, and `wnc_wlan_session_timeout_seconds` and the four `wnc_wlan_central_*_enabled` series read leaves of that same profile, so all six report a property of the policy profile per WLAN rather than of the WLAN profile. None of these six is published for a WLAN that resolves to no policy profile, or for any WLAN at all when either the `wlan-policies` or the `policy-list-entries` fetch fails. A `0` on `wnc_wlan_policy_enabled` can coexist with `wnc_wlan_enabled` reading `1`, because that series reads the WLAN profile's own administrative state and not the profile bound to it. What a shut policy profile does to a client — refuse new associations, drop existing ones, or stop the SSID being advertised — is not established here, so treat the series as change detection rather than an outage signal. Where one policy profile is bound to several WLANs, each WLAN reports it separately. Where one WLAN is bound through more than one policy tag, the six report the last binding the exporter can resolve, and a binding naming a policy profile absent from `wlan-policies` is skipped rather than reported, so it cannot displace an earlier one. No label on these series names the tag or the profile, so a `0` can mean the profile on one tag's binding is shut while the others are active, and a `1` can hide a shut binding.

<details><summary><b>*1</b> What the byte counter totals, and what it is not</summary><br/>

The controller keeps one byte total per WLAN, and this series publishes it unchanged. **The unit is bytes**, established by comparing the delta of this leaf against the summed delta of the per-client receive and transmit byte counters over intervals where the WLAN's client set did not change: the two matched exactly, on every WLAN, in three independent runs. A single interval on a busy WLAN deviates by a sampling skew that cancels over the pool, so compare pooled deltas rather than one interval if you repeat the measurement.

It counts **both directions together**, and it is a controller-side accumulator rather than a re-sum of the clients present now: it keeps the bytes of clients that have since disconnected. So it is **not** the sum of the per-client counters, and a WLAN with no clients at all can carry a large and unmoving value. For a rate use `rate()` over it directly; to attribute traffic to a client, use the per-client counters instead.

The leaf is a string on the wire. A record whose leaf is missing or unparsable is skipped rather than read as zero, because a counter that drops to zero is read as a reset and extrapolated from. The series is likewise absent for a WLAN the controller lists no statistics record for, and for every WLAN while the `wlan_client_stats` fetch fails.

</details>

<details><summary><b>*2</b> What the phase counts detect, and what they do not</summary><br/>

The controller keeps one count per phase rather than one enumerated leaf, so the four `phase` values — `l2auth`, `mobility`, `iplearn` and `webauth_pending` — are this exporter's own names for those four leaves.

**They are current counts, not cumulative ones.** The fifth count in the same record, the clients in the run state, equalled the per-WLAN client records exactly on every WLAN and in total, which is what types all five as gauges. They are **not additive** with `wnc_wlan_clients`, which counts the run state only.

**What they detect is a stall, not a failure rate.** A client that onboards normally occupies a phase for milliseconds, so a scrape lands on it only by coincidence: all four leaves read zero in every one of ninety consecutive reads taken ten seconds apart. A client held in a phase, by contrast, stays there and is what the series is for. Alert on a count that persists rather than on any non-zero reading:

```bash
min_over_time(wnc_wlan_onboarding_clients[5m]) > 0
```

The four series are absent for a WLAN the controller lists no statistics record for, and for every WLAN while the `wlan_client_stats` fetch fails. A record present with a phase leaf omitted cannot be told from a zero, so the error runs one way only — a stall can be under-reported and never invented.

</details>

<details><summary><b>*3</b> What the PMF setting covers</summary><br/>

The leaf reports the setting that applies to the WLAN's 2.4 GHz and 5 GHz BSSes. A 6 GHz BSS requires PMF whichever value this series reports, and the controller reports that requirement separately in a form no leaf carries. So a rule that pages on anything other than the required spelling raises a false alarm on a WLAN advertised on 6 GHz. The error runs one way only — the series can under-report 6 GHz protection and never over-report it.

The setting has three values rather than two, and the middle one admits an unprotected association, which is why the spelling is published rather than a boolean.

</details>

<details><summary><b>*4</b> Reading which policy profile the six config series report</summary><br/>

`wnc_wlan_policy_binding` publishes one series per binding, carrying the `id` of the WLAN, the `policy_profile` it is bound to and the `policy_tag` carrying the binding, always with the value `1`. The six policy-derived series above name none of those, so where a WLAN is bound through more than one tag they report one of the profiles and nothing says which — this series is what makes that state visible.

Alert on it with the count of **distinct profiles**, not of series:

```bash
count by (id) (count by (id, policy_profile) (wnc_wlan_policy_binding)) > 1
```

Binding one WLAN to one profile through several tags is not ambiguous, and the inner `count` is what excludes it; counting series instead fires on it.

A binding is published only when the exporter can resolve both ends. A tag naming a WLAN the controller does not define carries no `id` to key a series by, and a binding whose policy profile is absent from `wlan-policies` is skipped by the six series as well, so publishing either would show a binding they are not reporting. Both cases are real: a controller can carry a policy tag that names WLANs which no longer exist.

It comes from the same two reads as the six series and adds no request of its own, so it is absent exactly when they are.

</details>

`wnc_wlan_enabled` reads `wlan-status` from the optional `apf-vap-id-data` container on the WLAN entry, while `wnc_wlan_session_timeout_seconds` reads `session-timeout` from `wlan-timeout` and the four `wnc_wlan_central_*_enabled` series read their `central-*` leaf from `wlan-switching-policy`, two optional containers on the policy profile. None of the six named here is published when the container it reads is absent, and a container the controller does send may still omit individual leaves, which decode to `0`. While the fallback counter is rising, or on a controller that accepts the request and ignores it, a profile that left every leaf in `wlan-switching-policy` at its default carries no such container, so the four `wnc_wlan_central_*_enabled` series go absent.

The remaining `config` boolean series read a leaf that no container guards, where an omitted leaf and a configured `false` decode alike, and no container check can tell them apart. While the fallback counter is rising, or on a controller that accepts the request and ignores it, a `0` on those series can mean a leaf the controller did not send rather than a feature that is off.

## Labels

`info` module provides `wnc_wlan_info` contains following labels to join with other metrics:

| Labels | Description     | Example Value  | Default | Required |
| :----- | :-------------- | :------------- | :-----: | :------: |
| `id`   | WLAN identifier | `1`, `5`, `10` | **Yes** | **Yes**  |
| `name` | WLAN ESSID name | `labo-wifi`    | **Yes** |    No    |

Use this info metric to add contextual labels to other metrics in PromQL queries:

```bash
wnc_wlan_enabled * on(id) group_left(name) wnc_wlan_info
```
