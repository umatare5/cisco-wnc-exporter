# Controller collector

Controller collector focuses on the controller itself rather than on an AP, a client or a WLAN.

## Metrics

| Module  | Metric                                                  | Type    | Description                        |
| :------ | :------------------------------------------------------ | :------ | :--------------------------------- |
| general | `wnc_controller_boot_time_seconds`                      | Gauge   | Unix time of the last boot         |
| general | `wnc_controller_client_deletes_total`                   | Counter | Client deletions per `reason`      |
| general | `wnc_controller_client_ap_auth_roams_total`             | Counter | Roams on the AP-authenticated path |
| general | `wnc_controller_client_ap_auth_dot11i_fast_roams_total` | Counter | 802.11i fast roams on that path    |
| general | `wnc_controller_client_ap_auth_dot11i_slow_roams_total` | Counter | 802.11i slow roams on that path    |

## Specifications

Each entry carries what the series' HELP text and the shared [Absence](README.md#absence) rules do not.

**`wnc_controller_boot_time_seconds`**

- Withheld as well where the leaf carries the 1970 sentinel this controller writes for an event that has not happened, so the omitted leaf the HELP names is not the only silence.
- An instant the wire form cannot express fails the read rather than arriving as a sentinel, and `wnc_refresh_errors_total{data="controller_boot_time"}` is then the only thing separating it from a leaf the controller omitted.
- Neither counter container on this page reports an epoch of its own, so this series is the only reset anchor the four counters below have.
- It reads the controller's native leaf rather than the derived copy a second model reports, which agreed on four of five samples and read a second earlier on the fifth, so do not cross-check the two by equality.

**`wnc_controller_client_deletes_total`**

- Every reason leaf is published, the ones reading zero included, so the series count is fixed per controller rather than growing with the AP or client count, and a reason's first non-zero read is a rise on a series that was already there.
- The operational data carries no per-client, per-AP or per-WLAN equivalent, so correlate a rise with the client and WLAN series by time rather than by a join.

**`wnc_controller_client_ap_auth_roams_total` and the two `wnc_controller_client_ap_auth_dot11i_*_roams_total`**

- Which policy-profile setting excludes a WLAN from these counters is unsettled — local authentication and central association are both candidates, and nothing measured separates the two.
- The container carries ten further leaves, all reading zero where these three carried a total and two of them duplicating `wnc_wlan_central_association_enabled` and `wnc_wlan_ft_state`, so none of the ten is published.

**The two `wnc_controller_client_ap_auth_dot11i_*_roams_total`**

- Their ratio is **not a key-cache hit rate**: neither counter partitions the total, so a roam the total counted need not fall into either of them.

**`wnc_refresh_errors_total`**

- A controller or an image that does not carry one of this page's three data types answers `404`, and a `404` counts as **a failure rather than an absence**, because a path this exporter got wrong answers `404` as well.
- Enabling this module against such a controller raises the counter for that data type indefinitely, so leave it disabled there or exclude all three data types, `controller_boot_time` included:

> ```bash
> increase(wnc_refresh_errors_total{data!~"controller_boot_time|co_client_del_reason|client_roaming_stats"}[15m]) > 0
> ```

## Info Metrics

Every series here describes the whole controller, so none of them carries an identifying label. There is no `info` metric to join with, and nothing on this page can be attributed to a device.
