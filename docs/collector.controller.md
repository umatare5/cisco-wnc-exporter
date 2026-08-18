# Controller collector

Controller collector focuses on the controller itself rather than on an AP, a client or a WLAN.

Every series here describes the whole controller, so none of them carries an identifying label. There is no `info` metric to join with, and nothing on this page can be attributed to a device.

## Metrics

| Module  | Metric                                                  | Type    | Description                                  |
| :------ | :------------------------------------------------------ | :------ | :------------------------------------------- |
| general | `wnc_controller_boot_time_seconds`                      | Gauge   | Unix time of the last boot **(\*1)**         |
| general | `wnc_controller_client_deletes_total`                   | Counter | Client deletions per `reason` **(\*2)**      |
| general | `wnc_controller_client_ap_auth_roams_total`             | Counter | Roams on the AP-authenticated path **(\*3)** |
| general | `wnc_controller_client_ap_auth_dot11i_fast_roams_total` | Counter | 802.11i fast roams on that path **(\*3)**    |
| general | `wnc_controller_client_ap_auth_dot11i_slow_roams_total` | Counter | 802.11i slow roams on that path **(\*3)**    |

One flag, `--collector.controller.general`, enables all five, and all three of its reads bypass the SDK's typed accessors — see note **(\*4)**. Neither counter container on this page reports an epoch of its own, so the boot time is the only reset anchor available, and putting it behind a second flag would let an operator enable the counters and lose the anchor they need — a rule of the form `and on() (time() - wnc_controller_boot_time_seconds > 3600)` returns nothing when the right-hand side is absent, silently and forever.

## Notes

<details><summary><b>*1</b> The boot time moves by a second, and is withheld rather than reported as zero</summary><br/>

The leaf was observed alternating between two adjacent seconds across repeated reads, so **a rule must not use `changes()` or an equality on this series.** Compare a delta against a threshold instead, as in `time() - wnc_controller_boot_time_seconds < 600` for a controller that has just restarted.

The series is absent when the controller does not carry the leaf, when the value cannot be parsed, and when it reads as the Unix epoch. A zero there would report a boot in 1970, which makes an uptime derived from it five decades long and a counter-reset check quietly wrong.

The controller reports the same instant in a second model as well, derived rather than native, and that copy was observed a second away from this one — this series reads the native leaf.

</details>

<details><summary><b>*2</b> The delete reasons cannot be attributed to a client, an AP or a WLAN</summary><br/>

The controller keeps one counter per reason for the whole controller. There is no per-client, per-AP or per-WLAN equivalent in the operational data, the CLI reports the same single table, and the log records a client joining but not a client being deleted. So a rise here says what happened and never to whom: correlate it with the client and WLAN series by time, not by a join.

Every reason leaf the controller reports is published, including the ones reading zero. Withholding those would make a leaf's first non-zero read look like a counter created at that moment, which Prometheus cannot tell from a reset. Expect several hundred series, fixed per controller rather than growing with the AP or client count.

The label carries the controller's own spelling of the reason.

</details>

<details><summary><b>*3</b> The roam counters cover one path only, and do not sum</summary><br/>

All three count roams on the **FlexConnect local-authentication path**, which is how the controller itself labels them. **A WLAN whose roams do not take that path is not counted here at all**, so read a flat zero as "no roam took this path" and not as "nobody roamed". Exactly which policy-profile setting excludes a WLAN was not separated by measurement — the counters carried a non-zero total while every WLAN measured had local authentication and central association both switched off, so nothing measured tells the two knobs apart. One of the three was later caught advancing between two reads, which types it as a counter; the other two have never been seen to move.

The two `dot11i` counters are **not a partition** of the total. Where this was measured the total stood at some two hundred and fifty times their sum, so their ratio is not a key-cache hit rate. **This note recommended alerting on the shape — a fast counter at zero while the slow one advances — and that recommendation is withdrawn.** A roam the total counts need not fall into either sub-counter, so a fast counter at zero says nothing about how the roams behind the total authenticated. What the total itself supports is narrower: it advances on a roam and does not advance on a fresh association, so read it as a roam counter and not as an association counter.

Both are cumulative, and the container reports no epoch of its own, so read `wnc_controller_boot_time_seconds` as a reset anchor rather than as the instant the counts run from. The container carries ten further leaves that read zero where these three carried a total, and two of those duplicate what `wnc_wlan_central_association_enabled` and `wnc_wlan_ft_state` already publish, so none of the ten is published here.

</details>

<details><summary><b>*4</b> These reads do not go through a typed SDK accessor, and what that changes</summary><br/>

Three of this exporter's data types are read by building the RESTCONF path directly, because the SDK carries no route and no type for any of the three containers behind this page: `controller_boot_time`, `co_client_del_reason` and `client_roaming_stats`. They reuse the SDK client, so the credentials, the TLS settings, the request timeout, the connection pool and the error typing are the same as everywhere else, and each is a registered data type like any other — gated by its flag, bounded by the refresh deadline, and counted in `wnc_refresh_items` and `wnc_refresh_errors_total`.

Two consequences are worth knowing.

A response is decoded without struct tags, and the exporter checks that the container the controller answered with is the one the path asked for. A decode that trusted a tag would turn a container renamed between releases into an empty family with no error at all, which is the failure this check exists to prevent.

**A controller or an image that does not carry one of these containers answers `404`, and a `404` is a failure rather than an absence.** That is deliberate: a path this exporter got wrong answers `404` as well, and making it silent would hide the mistake. The cost is that enabling this module against a controller without the container raises `wnc_refresh_errors_total` for that data type indefinitely. Leave the module disabled there, or exclude the three data types from the rule:

```bash
increase(wnc_refresh_errors_total{data!~"controller_boot_time|co_client_del_reason|client_roaming_stats"}[15m]) > 0
```

A container that is present but empty is a different case: the controller answers with no body, the read counts as a successful fetch of nothing, and the series are simply absent.

</details>
