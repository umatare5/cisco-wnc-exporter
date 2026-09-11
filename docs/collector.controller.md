# Controller Collector

Controller collector focuses on the controller itself rather than on an AP, a client or a WLAN.

## Metrics

| Group   | Metric                                                  | Type    | Description                        |
| :------ | :------------------------------------------------------ | :------ | :--------------------------------- |
| general | `wnc_controller_boot_time_seconds`                      | Gauge   | Unix time of the last boot         |
| general | `wnc_controller_client_deletes_total`                   | Counter | Client deletions per `reason`      |
| general | `wnc_controller_client_ap_auth_roams_total`             | Counter | Roams on the AP-authenticated path |
| general | `wnc_controller_client_ap_auth_dot11i_fast_roams_total` | Counter | 802.11i fast roams on that path    |
| general | `wnc_controller_client_ap_auth_dot11i_slow_roams_total` | Counter | 802.11i slow roams on that path    |

## Labels

- `wnc_controller_client_deletes_total{reason}` is the only labelled series on this page.
- `wnc_controller_boot_time_seconds` and the three roam counters carry no label at all.
- `reason` is open rather than closed: its values are the raw leaf names of the `co-client-del-reason` container, iterated straight out of a map with no whitelist and no translation in the exporter.
- The collector publishes no `info` series and therefore offers nothing to join on.
- No series here identifies a device, so nothing on this page can be attributed to one.

## Specifications

The five series come from three data types — `controller_boot_time`, `co_client_del_reason` and `client_roaming_stats` — and each block below records what the series' HELP text leaves unstated.

**`wnc_controller_boot_time_seconds`**

- Withheld on the 1970 sentinel as well as on an omitted leaf, so the HELP text names only half the silence — [Absence](README.md#absence) carries the sentinel rule itself.
- An instant the wire form cannot express fails the read rather than arriving as a sentinel. Only `wnc_refresh_errors_total{data="controller_boot_time"}` then separates it from a leaf the controller omitted — [Health](health.md) carries that series.
- Neither counter container on this page reports an epoch of its own, so this series is the only reset anchor the four counters below have.
- It reads the controller's native leaf rather than the derived copy a second model reports. That copy agreed on four of five samples and read a second earlier on the fifth, so do not cross-check the two by equality.

**`wnc_controller_client_deletes_total`**

- Every reason leaf is published, the ones reading zero included, so the series count is fixed per controller rather than growing with the AP or client count. A reason's first non-zero read is therefore a rise on a series that was already there.
- The operational data carries no per-client, per-AP or per-WLAN equivalent, so correlate a rise with the client and WLAN series by time rather than by a join.

**The three `wnc_controller_client_ap_auth_*_roams_total`**

- Which policy-profile setting excludes a WLAN from these counters is unsettled — local authentication and central association are both candidates, and nothing measured separates the two.
- The container carries ten further leaves, all reading zero where these three carried a total and two of them duplicating `wnc_wlan_central_association_enabled` and `wnc_wlan_ft_state`, so none of the ten is published.

**The two `wnc_controller_client_ap_auth_dot11i_*_roams_total`**

- Their ratio is **not a key-cache hit rate**: neither counter partitions the total, so a roam the total counted need not fall into either of them.

**`wnc_refresh_errors_total`**

- A controller or an image that does not carry one of this page's three data types answers `404` — [Absence](README.md#absence) carries why a `404` counts as a failure rather than an absence.
- Enabling this collector against such a controller raises the counter for that data type indefinitely, so leave it disabled there or exclude all three data types, `controller_boot_time` included:

```bash
increase(wnc_refresh_errors_total{data!~"controller_boot_time|co_client_del_reason|client_roaming_stats"}[15m]) > 0
```
