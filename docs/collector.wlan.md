# WLAN collector

WLAN collector focuses on logical SSID performance and parameter checks.

## Metrics

| Module  | Metric                                    | Type    | Description                          |
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

## Specifications

Each entry carries what the series' HELP text and the shared [Absence](README.md#absence) rules do not.

**`wnc_wlan_enabled`**

- Reads `wlan-status` from the optional `apf-vap-id-data` container, so a WLAN whose entry carries neither the container nor the leaf reports nothing rather than `0`.

**`wnc_wlan_clients`**

- Counts only the run state, so during an onboarding failure it **falls** while clients pile up in an earlier phase — the opposite of what a rule written against a client-count rise expects.
- Reports `0` for a WLAN with no run-state client and goes absent only when the client-list fetch fails, unlike the two series below, which are keyed to a per-WLAN statistics record.

**`wnc_wlan_data_usage_bytes_total`**

- An administrative shutdown of the WLAN returned it to zero and re-enabling started the count again, so `increase()` and `rate()` across that shutdown lose everything counted before it. Whether any other event zeroes it was not measured.
- A record whose leaf is missing or unparsable is skipped rather than read as zero, because a zero on a counter cannot be told from a reset.

**`wnc_wlan_onboarding_clients`**

- The four `phase` values — `l2auth`, `mobility`, `iplearn` and `webauth_pending` — are this exporter's own names for four separate leaves, not spellings the controller assigns.
- Whether the counts partition a WLAN's clients was not measured, so do not add them to `wnc_wlan_clients`, which counts the run state alone.
- They detect a stall rather than a failure rate: normal onboarding occupies a phase for milliseconds and a failure that ended in a disconnect leaves them at zero, so `wnc_controller_client_deletes_total` is what reports whether onboarding is failing.

**`wnc_wlan_auth_dot1x_enabled`, `wnc_wlan_wpa2_enabled`, `wnc_wlan_11k_neighbor_list_enabled`**

- Measured on IOS-XE 17.15 to be omitted from exactly the WLANs where the setting is **on**, which is why all three withhold the series rather than read an omitted leaf as `false`.

**`wnc_refresh_defaults_fallback_total`**

- Rises only where the controller answers `400` to the request for the values in force. A controller that accepts that request and ignores it answers `200`, so a flat counter is no proof that no `0` on this page stands for a leaf never sent.

**`wnc_wlan_policy_enabled`, `wnc_wlan_session_timeout_seconds` and the four `wnc_wlan_central_*_enabled`**

- All six report a property of the policy profile the WLAN resolves to through a policy tag rather than of the WLAN profile itself.
- None of them, and no `wnc_wlan_policy_binding`, is published for a WLAN that resolves to no policy profile or for any WLAN while either the `wlan-policies` or the `policy-list-entries` fetch fails — the remaining `config` series are unaffected.

**`wnc_wlan_policy_enabled`**

- What a shut policy profile does to a client — refuse new associations, drop existing ones, or stop advertising the SSID — is not established, so read it as change detection rather than as an outage signal.

**`wnc_wlan_pmf_state`**

- Match it by equality: at `>= 1` it still admits an unprotected association, because the middle value is `apf-vap-pmf-optional` — [Enumeration values](enums.md) maps both this and `wnc_wlan_ft_state`.
- A rule paging on anything other than `2` raises a false alarm on a WLAN advertised on 6 GHz, and the error runs one way only — under-reporting 6 GHz protection and never over-reporting it.

**`wnc_wlan_policy_binding`**

- Published only where both ends resolve, so a tag naming a WLAN the controller does not define and a binding whose profile is absent from `wlan-policies` are both skipped: it is not a complete inventory of the controller's tags.
- Where it shows more than one profile for an `id`, the six policy series report only one of them, so a `0` on `wnc_wlan_policy_enabled` can mean one binding's profile is shut while the others are active, and a `1` can hide a shut one.
- Alert on the count of **distinct profiles** rather than of series, because one WLAN bound to one profile through several tags is not ambiguous and the inner `count` is what excludes it:

> ```bash
> count by (id) (count by (id, policy_profile) (wnc_wlan_policy_binding)) > 1
> ```

## Info Labels

The `info` module publishes `wnc_wlan_info` with the following labels to join with other metrics:

| Labels | Description     | Example Value  | Default | Required |
| :----- | :-------------- | :------------- | :-----: | :------: |
| `id`   | WLAN identifier | `1`, `5`, `10` | **Yes** | **Yes**  |
| `name` | WLAN ESSID name | `labo-wifi`    | **Yes** |    No    |

Use this info metric to add contextual labels to other metrics in PromQL queries:

```bash
wnc_wlan_enabled * on(id) group_left(name) wnc_wlan_info
```
