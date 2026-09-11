# WLAN Collector

WLAN collector reports logical SSID performance and the configuration each WLAN resolves to.

## Metrics

| Group   | Metric                                    | Type    | Description                          |
| :------ | :---------------------------------------- | :------ | :----------------------------------- |
| general | `wnc_wlan_enabled`                        | Gauge   | WLAN status                          |
| traffic | `wnc_wlan_clients`                        | Gauge   | Run-state clients count (calculated) |
| traffic | `wnc_wlan_data_usage_bytes_total`         | Counter | Bytes in both directions             |
| traffic | `wnc_wlan_onboarding_clients`             | Gauge   | Clients held in a phase              |
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
| config  | `wnc_wlan_pmf_state`                      | Gauge   | PMF setting                          |
| config  | `wnc_wlan_ft_state`                       | Gauge   | 802.11r fast transition setting      |
| config  | `wnc_wlan_policy_binding`                 | Gauge   | Policy tag binding                   |

## Labels

Every series above carries `id`, `wnc_wlan_onboarding_clients` carries a phase this exporter names, and only `wnc_wlan_info` takes labels from a flag.

| Label            | Series                        | Values                          |
| :--------------- | :---------------------------- | :------------------------------ |
| `id`             | Every WLAN series             | `1`, `5`, `10`                  |
| `phase`          | `wnc_wlan_onboarding_clients` | Closed at the four names below  |
| `policy_profile` | `wnc_wlan_policy_binding`     | An operator-chosen profile name |
| `policy_tag`     | `wnc_wlan_policy_binding`     | An operator-chosen tag name     |
| `name`           | `wnc_wlan_info`               | `labo-wifi` or the profile name |

**`id`** identifies the WLAN on every series this collector publishes, `wnc_wlan_info` included.

**`phase`** is closed at `l2auth`, `mobility`, `iplearn` and `webauth_pending`, which are this exporter's names for four separate leaves rather than the spellings of one controller enumeration.

**`policy_profile` and `policy_tag`** are both always present on `wnc_wlan_policy_binding` and both open, because each is read off a policy-list entry and carries whatever name the operator gave it.

**`name`** is the only optional label on `wnc_wlan_info` and also its default, so `--collector.wlan.info-labels` toggles exactly one label and `id` is appended whatever that flag names.

`--collector.wlan.info` publishes `wnc_wlan_info`, which carries the ESSID and joins on `id`. Where the entry carries no `ssid`, the label falls back to the WLAN profile name rather than going absent:

```bash
wnc_wlan_enabled * on(id) group_left(name) wnc_wlan_info
```

## Specifications

**`wnc_wlan_enabled`**

- Reads `wlan-status` from the optional `apf-vap-id-data` container, so a WLAN whose entry carries neither the container nor the leaf reports nothing rather than `0`.

**`wnc_wlan_clients`**

- Counts only the run state, so during an onboarding failure it **falls** while clients pile up in an earlier phase — the opposite of what a rule written against a client-count rise expects.
- Reports `0` for a WLAN with no run-state client and goes absent only when the client-list fetch fails, unlike the two series below, which are keyed to a per-WLAN statistics record.

**`wnc_wlan_data_usage_bytes_total`**

- An administrative shutdown of the WLAN returned it to zero and re-enabling started the count again, so `increase()` and `rate()` across that shutdown lose everything counted before it. Whether any other event zeroes it was not measured.
- A record whose leaf is missing or unparsable is skipped rather than read as zero, because a zero on a counter cannot be told from a reset.

**`wnc_wlan_onboarding_clients`**

- Whether the counts partition a WLAN's clients was not measured, so do not add them to `wnc_wlan_clients`, which counts the run state alone.
- They detect a stall rather than a failure rate: normal onboarding occupies a phase for milliseconds and a failure that ended in a disconnect leaves them at zero. `wnc_controller_client_deletes_total` is what reports whether onboarding is failing.

**`wnc_wlan_auth_psk_enabled`, `wnc_wlan_auth_dot1x_sha256_enabled`, `wnc_wlan_wpa3_enabled`, `wnc_wlan_load_balance_enabled`, `wnc_wlan_client_steering_enabled`**

- These five decode a value-typed leaf rather than a pointer, so an omitted leaf reads as a configured `false` and publishes `0` instead of withholding the series. They are this page's exception to [Absence](README.md#absence), which carries the general form of the rule.

**`wnc_wlan_auth_dot1x_enabled`, `wnc_wlan_wpa2_enabled`, `wnc_wlan_11k_neighbor_list_enabled`**

- Measured on IOS-XE 17.15 to be omitted from exactly the WLANs where the setting is **on**, which is why all three withhold the series rather than read an omitted leaf as `false`.

**`wnc_refresh_defaults_fallback_total`**

- Only `wlan_cfg_entries` and `wlan_policies` ask for the values in force, so this collector is the one the counter moves for. [Exporter Health](health.md#specifications) carries what a flat counter does and does not prove about the `config` zeros above.

**`wnc_wlan_policy_enabled`, `wnc_wlan_session_timeout_seconds` and `wnc_wlan_central_*_enabled`**

- All six report a property of the policy profile the WLAN resolves to through a policy tag rather than a property of the WLAN profile itself.
- None of them, and no `wnc_wlan_policy_binding`, is published for a WLAN that resolves to no policy profile or for any WLAN while either the `wlan-policies` or the `policy-list-entries` fetch fails. The remaining `config` series are unaffected.

**`wnc_wlan_policy_enabled`**

- What a shut policy profile does to a client — refuse new associations, drop existing ones, or stop advertising the SSID — is not established. Read it as change detection rather than as an outage signal.

**`wnc_wlan_pmf_state`**

- The middle value `apf-vap-pmf-optional` still admits an unprotected association, so `>= 1` is not protection — [Enumeration Values](enums.md) maps both this and `wnc_wlan_ft_state`.
- A rule paging on anything other than `2` raises a false alarm on a WLAN advertised on 6 GHz, and the error runs one way only — under-reporting 6 GHz protection and never over-reporting it.

**`wnc_wlan_policy_binding`**

- Published only where both ends resolve, so a tag naming a WLAN the controller does not define and a binding whose profile is absent from `wlan-policies` are both skipped. It is not a complete inventory of the controller's tags.
- Where it shows more than one profile for an `id`, the six policy series report only one of them. A `0` on `wnc_wlan_policy_enabled` can then mean one binding's profile is shut while the others are active, and a `1` can hide a shut one.
- Alert on the count of **distinct profiles** rather than of series, because one WLAN bound to one profile through several tags is not ambiguous and the inner `count` is what excludes it:

> ```bash
> count by (id) (count by (id, policy_profile) (wnc_wlan_policy_binding)) > 1
> ```
