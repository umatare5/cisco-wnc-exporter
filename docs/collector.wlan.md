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
| config  | `wnc_wlan_wpa3_enabled`                   | Gauge | WPA3 support enabled                 |
| config  | `wnc_wlan_session_timeout_seconds`        | Gauge | Session timeout duration             |
| config  | `wnc_wlan_load_balance_enabled`           | Gauge | Load balancing enabled               |
| config  | `wnc_wlan_client_steering_enabled`        | Gauge | 6GHz client steering enabled         |
| config  | `wnc_wlan_central_switching_enabled`      | Gauge | Central switching enabled            |
| config  | `wnc_wlan_central_authentication_enabled` | Gauge | Central authentication enabled       |
| config  | `wnc_wlan_central_dhcp_enabled`           | Gauge | Central DHCP enabled                 |
| config  | `wnc_wlan_central_association_enabled`    | Gauge | Central association enabled          |

## Notes

The `general` and `config` modules read `wlan-cfg-entries`, and `config` also reads `wlan-policies`. The controller omits every leaf the profile never set from these responses — observed on IOS-XE 17.12, which answers RESTCONF with the `with-defaults` basic mode `explicit`. The same leaf can therefore arrive for one policy profile and be missing for another on one controller.

`wnc_wlan_enabled`, `wnc_wlan_session_timeout_seconds` and the four `wnc_wlan_central_*` series read their leaf from an optional container, and are not published for a WLAN whose response omits that container. A container the controller does send may still omit individual leaves, which decode to `0`, so a `0` on these series can still mean that the controller did not report the leaf.

The remaining `config` boolean series read a leaf on the entry itself, where an omitted leaf and a configured `false` decode alike. They report `0` for any leaf the controller did not send, so a leaf whose default is enabled would report `0` while the feature is in force. Every default measured behind these series on IOS-XE 17.12 was disabled — the two series that did read a default-enabled leaf were removed in v0.3.0.

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
