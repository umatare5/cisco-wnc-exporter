# WLAN collector

WLAN collector focuses on logical SSID performance and parameter checks.

## Metrics

| Module  | Metric                                    | Type  | Description                          |
| :------ | :---------------------------------------- | :---- | :----------------------------------- |
| general | `wnc_wlan_enabled`                        | Gauge | WLAN status                          |
| traffic | `wnc_wlan_clients`                        | Gauge | Run-state clients count (calculated) |
| config  | `wnc_wlan_auth_psk_enabled`               | Gauge | PSK authentication enabled           |
| config  | `wnc_wlan_auth_dot1x_enabled`             | Gauge | 802.1x authentication enabled        |
| config  | `wnc_wlan_auth_dot1x_sha256_enabled`      | Gauge | 802.1x SHA256 auth enabled           |
| config  | `wnc_wlan_wpa2_enabled`                   | Gauge | WPA2 support enabled                 |
| config  | `wnc_wlan_wpa3_enabled`                   | Gauge | WPA3 support enabled                 |
| config  | `wnc_wlan_session_timeout_seconds`        | Gauge | Session timeout duration             |
| config  | `wnc_wlan_load_balance_enabled`           | Gauge | Load balancing enabled               |
| config  | `wnc_wlan_11k_neighbor_list_enabled`      | Gauge | 802.11k neighbor list enabled        |
| config  | `wnc_wlan_client_steering_enabled`        | Gauge | 6GHz client steering enabled         |
| config  | `wnc_wlan_central_switching_enabled`      | Gauge | Central switching enabled            |
| config  | `wnc_wlan_central_authentication_enabled` | Gauge | Central authentication enabled       |
| config  | `wnc_wlan_central_dhcp_enabled`           | Gauge | Central DHCP enabled                 |
| config  | `wnc_wlan_central_association_enabled`    | Gauge | Central association enabled          |
| config  | `wnc_wlan_policy_enabled`                 | Gauge | Bound policy profile is active       |
| config  | `wnc_wlan_pmf_state`                      | Gauge | PMF setting **(\*1)**                |
| config  | `wnc_wlan_ft_state`                       | Gauge | 802.11r fast transition setting      |

## Notes

The `general` and `config` modules read `wlan-cfg-entries`, and `config` also reads `wlan-policies`. The exporter asks the controller for the values in force on both, because a controller answering with the `with-defaults` basic mode `explicit` omits every leaf the profile never set — observed on IOS-XE 17.12. A controller that rejects the request is read plainly instead, and `wnc_refresh_defaults_fallback_total` rises for as long as that lasts.

`wnc_wlan_pmf_state` and `wnc_wlan_ft_state` report the controller's own spelling in the `state` label and always have the value `1`, so `== 0` never fires — see [States](README.md#a-state-is-a-label-not-a-number). Neither is published for a WLAN whose response omits the leaf, because the leaf decodes to an empty string and an empty `state` label reads as no label at all.

`wnc_wlan_policy_enabled` reads the status of the policy profile the WLAN resolves to through its policy tag, and is not published for a WLAN that resolves to none. A WLAN whose policy profile is shut down does not admit clients even while `wnc_wlan_enabled` reads `1`, because that series reads the WLAN profile's own administrative state. Whether the SSID also stops being advertised is not established here, so treat the series as change detection rather than an outage signal. Where one policy profile is bound to several WLANs, each WLAN reports it separately; where one WLAN is bound through more than one policy tag, only one binding is reported.

<details><summary><b>*1</b> What the PMF setting covers</summary><br/>

The leaf reports the setting that applies to the WLAN's 2.4 GHz and 5 GHz BSSes. A 6 GHz BSS requires PMF whichever value this series reports, and the controller reports that requirement separately in a form no leaf carries. So a rule that pages on anything other than the required spelling raises a false alarm on a WLAN advertised on 6 GHz. The error runs one way only — the series can under-report 6 GHz protection and never over-report it.

The setting has three values rather than two, and the middle one admits an unprotected association, which is why the spelling is published rather than a boolean.

</details>

`wnc_wlan_enabled`, `wnc_wlan_session_timeout_seconds` and the four `wnc_wlan_central_*` series read their leaf from an optional container, and are not published for a WLAN whose response omits that container. A container the controller does send may still omit individual leaves, which decode to `0`.

The remaining `config` boolean series read a leaf on the entry itself, where an omitted leaf and a configured `false` decode alike, and no container check can tell them apart. While the fallback counter is rising, or on a controller that accepts the request and ignores it, a `0` on those series can mean a leaf the controller did not send rather than a feature that is off.

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
