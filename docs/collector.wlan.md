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

The `config` module reads `wlan-cfg-entries` and `wlan-policies`. The controller
omits leaves left at their default values from these responses — observed on
IOS-XE 17.12, which answers RESTCONF with the `with-defaults` basic mode
`explicit`. A boolean series therefore reports `0`, and
`wnc_wlan_session_timeout_seconds` reports `0`, for any leaf the controller did
not send, even when the operative default is enabled or non-zero.

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
