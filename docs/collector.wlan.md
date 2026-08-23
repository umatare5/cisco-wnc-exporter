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

Every WLAN module reads `wlan-cfg-entries`, and `config` also reads `wlan-policies` and `policy-list-entries`. The exporter asks the controller for the values in force on the first two, and a controller that answers `400` is read plainly instead while `wnc_refresh_defaults_fallback_total` rises — [Absence](README.md#a-leaf-the-controller-omits-is-withheld-not-published-as-zero) carries why that matters. The `traffic` module also reads the client list, and `wnc_wlan_clients` is withheld for every WLAN when that fetch fails.

`wnc_wlan_pmf_state` and `wnc_wlan_ft_state` publish the number the controller's own enumeration assigns, `0` to `2` on both — [Enumeration values](enums.md) is the mapping and [States](README.md#a-state-is-a-number-not-a-label) the query shape. Match both by equality rather than by threshold: `wnc_wlan_pmf_state` at `>= 1` still admits an unprotected association, and `wnc_wlan_ft_state` `2`, `dot11r-adaptive-enabled`, is a compatibility mode for clients that cannot use the fast-transition AKM rather than a stronger form of its `1`. Neither is published for a WLAN whose response omits the leaf.

`wnc_wlan_policy_enabled`, `wnc_wlan_session_timeout_seconds` and the four `wnc_wlan_central_*_enabled` series read the policy profile the WLAN resolves to through a policy tag, so all six report a property of that profile rather than of the WLAN profile. A `0` on `wnc_wlan_policy_enabled` can therefore coexist with `wnc_wlan_enabled` reading `1`, and what a shut policy profile does to a client is not established here — treat the six as change detection rather than an outage signal.

None of the six is published for a WLAN that resolves to no policy profile, or for any WLAN while the `wlan-policies` or `policy-list-entries` fetch fails. Where one WLAN is bound through more than one policy tag, they report the last binding that resolves, and no label names the tag or the profile — so a `0` can mean one binding is shut while the others are active, and a `1` can hide a shut one.

<details><summary><b>*1</b> What the byte counter totals, and what it is not</summary><br/>

The controller keeps one byte total per WLAN, and this series publishes it unchanged. **The unit is bytes**, measured against the summed per-client byte counters over intervals where the WLAN's client set did not change.

It counts **both directions together** and keeps the bytes of clients that have since disconnected, so it is **not** the sum of the per-client counters — a WLAN with no clients at all can carry a large and unmoving value. For a rate use `rate()` over it directly, and to attribute traffic to a client use the per-client counters instead.

**An administrative shutdown of the WLAN returned this counter to zero, and re-enabling the WLAN started the count again from zero.** So `increase()` over a range spanning the shutdown loses everything counted before it, and the value is not a lifetime total for the WLAN.

The series is absent for a WLAN the controller lists no statistics record for, and for every WLAN while the `wlan_client_stats` fetch fails.

</details>

<details><summary><b>*2</b> What the phase counts detect, and what they do not</summary><br/>

The controller keeps one count per phase rather than one enumerated leaf, so the four `phase` values — `l2auth`, `mobility`, `iplearn` and `webauth_pending` — are this exporter's own names for those four leaves.

**They are current counts, not cumulative ones**, established against the run-state count in the same record. Do not add them to `wnc_wlan_clients`, which counts the run state only.

**What they detect is a stall, not a failure rate.** A client that onboards normally occupies a phase for milliseconds, so a scrape lands on it only by coincidence: all four leaves read zero in every one of ninety consecutive reads taken ten seconds apart. A client held in a phase, by contrast, stays there and is what the series is for. Alert on a count that persists rather than on any non-zero reading:

```bash
min_over_time(wnc_wlan_onboarding_clients[5m]) > 0
```

**A failure that has already completed is counted elsewhere.** These gauges hold a client only while it is stuck, so a failure that ended in a disconnect leaves them at zero — `wnc_controller_client_deletes_total{reason}` on the [Controller](collector.controller.md) page keeps the cumulative count per reason. Read that counter for whether onboarding is failing and these gauges for which WLAN is holding a client now.

The four series are absent for a WLAN the controller lists no statistics record for, and for every WLAN while the `wlan_client_stats` fetch fails. A record present with a phase leaf omitted cannot be told from a zero, so the error runs one way only — a stall can be under-reported and never invented.

</details>

<details><summary><b>*3</b> What the PMF setting covers</summary><br/>

The leaf reports the setting that applies to the WLAN's 2.4 GHz and 5 GHz BSSes, while a 6 GHz BSS requires PMF whichever value it reports. So a rule that pages on anything other than `2` raises a false alarm on a WLAN advertised on 6 GHz — the series can under-report 6 GHz protection and never over-report it.

The setting has three values rather than two, and the middle one admits an unprotected association, which is why the controller's own value is published rather than a boolean.

</details>

<details><summary><b>*4</b> Reading which policy profile the six config series report</summary><br/>

`wnc_wlan_policy_binding` publishes one series per binding, carrying the `id` of the WLAN, the `policy_profile` it is bound to and the `policy_tag` carrying the binding, always with the value `1`. The six policy-derived series above name none of those, so where a WLAN is bound through more than one tag they report one of the profiles and nothing says which — this series is what makes that state visible.

Alert on it with the count of **distinct profiles**, not of series:

```bash
count by (id) (count by (id, policy_profile) (wnc_wlan_policy_binding)) > 1
```

Binding one WLAN to one profile through several tags is not ambiguous, and the inner `count` is what excludes it — counting series instead fires on it.

A binding is published only when the exporter can resolve both ends. A tag naming a WLAN the controller does not define carries no `id` to key a series by, and a binding whose policy profile is absent from `wlan-policies` is skipped by the six series as well, so publishing either would show a binding they are not reporting. Both cases are real: a controller can carry a policy tag that names WLANs which no longer exist.

It comes from the same two reads as the six series and adds no request of its own, so it is absent exactly when they are.

</details>

Absence divides the `config` series. `wnc_wlan_enabled`, `wnc_wlan_session_timeout_seconds` and the four `wnc_wlan_central_*_enabled` series read optional containers, so none is published when its container is absent — and every one but `wnc_wlan_enabled` carries that absence per leaf as well, so a sibling being present is no evidence that theirs was sent. `wnc_wlan_auth_dot1x_enabled`, `wnc_wlan_wpa2_enabled` and `wnc_wlan_11k_neighbor_list_enabled` guard no container but go absent on an omitted leaf of their own, which is what makes them safe to read.

The remaining six decode an omitted leaf and a configured `false` alike — `wnc_wlan_policy_enabled`, `wnc_wlan_auth_psk_enabled`, `wnc_wlan_auth_dot1x_sha256_enabled`, `wnc_wlan_wpa3_enabled`, `wnc_wlan_load_balance_enabled` and `wnc_wlan_client_steering_enabled`. So while the fallback counter is rising, a `0` on those six can mean a leaf the controller did not send rather than a feature that is off.

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
