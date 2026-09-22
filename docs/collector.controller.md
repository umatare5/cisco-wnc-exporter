# Controller Collector

## Metrics

All five rows register behind `--collector.controller.general`, which is the only flag this collector takes.

| Group   | Metric                                                  | Type    | Description                        |
| :------ | :------------------------------------------------------ | :------ | :--------------------------------- |
| general | `wnc_controller_boot_time_seconds`                      | Gauge   | Unix time of the last boot         |
| general | `wnc_controller_client_deletes_total`                   | Counter | Client deletions per `reason`      |
| general | `wnc_controller_client_ap_auth_roams_total`             | Counter | Roams on the AP-authenticated path |
| general | `wnc_controller_client_ap_auth_dot11i_fast_roams_total` | Counter | 802.11i fast roams on that path    |
| general | `wnc_controller_client_ap_auth_dot11i_slow_roams_total` | Counter | 802.11i slow roams on that path    |

## Labels

Only one series here takes a label, and none of the five identifies a device. The `wnc_controller_client_deletes_total` series carries `reason`, whose values are the raw leaf names of `co-client-del-reason`.

The `reason` label is open rather than closed. The exporter iterates the map without whitelisting any specific values.

The collector publishes no `info` series and therefore offers nothing to join on.

## Annotations

The five series come from three data types: `controller_boot_time`, `co_client_del_reason` and `client_roaming_stats`. Each block below records what the series' HELP text leaves unstated.

**`wnc_controller_boot_time_seconds`**

It is withheld on the 1970 sentinel as well as on an omitted leaf. The HELP text names only half the silence. [Absence](architecture.md#absence) carries the sentinel rule itself.

An instant the wire form cannot express fails the read rather than arriving as a sentinel. Only `wnc_refresh_errors_total{data="controller_boot_time"}` separates it from a leaf the controller omitted. [Exporter Health](health.md#annotations) explains that series.

Neither counter container on this page reports an epoch of its own. This series is the only reset anchor the four counters below have.

It reads the controller's native leaf rather than the derived copy a second model reports. That copy agreed on four of five samples and read a second earlier on the fifth. Do not cross-check the two by equality.

**`wnc_controller_client_deletes_total`**

Every reason leaf is published, including the ones reading zero. The series count is fixed per controller rather than growing with the AP or client count. A reason's first non-zero read is therefore a rise on a series that was already there.

The operational data carries no per-client, per-AP or per-WLAN equivalent. Correlate a rise with the client and WLAN series by time rather than by a join.

**`Three wnc_controller_client_ap_auth_*_roams_total series`**

Which policy-profile setting excludes a WLAN from these counters is unsettled. Local authentication and central association are both candidates. Nothing measured separates the two.

The container carries ten further leaves, all reading zero where these three carried a total. Two of them duplicate `wnc_wlan_central_association_enabled` and `wnc_wlan_ft_state`. None of the ten is published.

**`Two wnc_controller_client_ap_auth_dot11i_*_roams_total series`**

Their ratio is not a key-cache hit rate. Neither counter partitions the total. A roam the total counted need not fall into either of them.

**`wnc_refresh_errors_total`**

A controller or an image that does not carry one of this page's three data types answers `404`. [Absence](architecture.md#absence) explains why a `404` counts as a failure rather than an absence.

## Technical Notes

**Controller API Errors**: Enabling this collector against such a controller raises the counter for that data type indefinitely. Leave it disabled there or exclude all three data types, `controller_boot_time` included:

```bash
increase(wnc_refresh_errors_total{data!~"controller_boot_time|co_client_del_reason|client_roaming_stats"}[15m]) > 0
```
