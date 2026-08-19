# Enumeration values

Twelve series report a state, a reason or a mode as a number, and the number is the one the controller's own schema assigns the spelling it sent. This page carries every number each of them can take. What each series measures is on the [AP](collector.ap.md), [Client](collector.client.md) and [WLAN](collector.wlan.md) pages, and [States](README.md#a-state-is-a-number-not-a-label) carries the query shapes these numbers need.

## Reading a value

- The number is the controller's rather than this exporter's: every member of all twelve enumerations carries an explicit `value` statement in the module that declares it, so these tables transcribe the device's numbering
- **Compare against a member, never against a threshold.** A larger number is not more of anything unless the family's HELP says so — `wnc_client_state` is the one whose numbering follows the onboarding sequence, while `wnc_wlan_pmf_state` at `>= 1` still admits an unprotected association and `wnc_wlan_ft_state` `2` is a compatibility mode for clients that cannot use the fast-transition AKM rather than a stronger form of `1`
- **`0` is a real member of eleven of the twelve, and it does not mean the same thing in each.** Six number a nothing-on-record member there — `disc-fail-none`, `jf-none`, `cf-none`, `dtls-hs-success`, `ap-reboot-reason-none` and `dot11-roam-type-none`. Two number a disabled setting, `apf-vap-pmf-disabled` and `dot11r-disabled`. Two report that the controller does not know: `ap-con-failure-unknown` is an unknown phase rather than an absence of failure, and `unkown` is the disconnect enumeration's own unknown member. The eleventh is `client-status-idle`, a state a client really holds
- `wnc_ap_oper_state` is the twelfth, and its enumeration declares no member at `0`, so that series never reads `0` and a rule written against one never fires
- **Two spellings are misspelled in the schema itself**, both in `spam-ap-disconnect-reason`: the unknown member at `0` is `unkown`, and `38` is `wtp-reboot-dimished-pwr-change`. Both are what the controller sends, so both are carried below verbatim — the correctly spelled `ap-reboot-reason-diminished-pwr-change` belongs to a different enumeration
- A spelling no table below lists is **withheld**: no series for that subject at all, rather than a number this release cannot name. No value is free to stand for one, because `0` is taken in eleven of the twelve and declared in none of the twelfth. Each member carries its own `value`, so an image that adds a member does not shift the numbers already assigned, and no member of the twelve was renumbered or removed across the releases compared — a newer controller therefore loses a series rather than reporting a wrong number. The same comparison counted what a newer image adds: **33 spellings these tables do not carry**, 31 of them in `spam-ap-reboot-reason` and 2 in `ap-discovery-failure-reason`, so those two families are where a controller outside the verified range goes silent first
- Run with `--log.level=debug` to read the spelling behind a withheld series. It is the one datum no query recovers, and the default level is `info`, so nothing is logged without the flag. An empty reading is withheld as well and logs nothing, because a leaf the controller omits is ordinary
- `wnc_client_protocol` is **not** one of these twelve. Its `0` to `7` are this exporter's own numbering, derived from the PHY-type spelling the controller sends rather than assigned by the controller's schema, and its HELP names all eight — so two numbering conventions coexist in one scrape and only the twelve below carry the controller's

## Where the numbers come from

RESTCONF carries the spelling and the CLI prints a rendered word, so the number itself exists only in the schema. These are the modules that declare the twelve enumerations, at the revision the controller's own `ietf-yang-library:modules-state` reported for each. **The revisions are here because nothing else makes a renumbering detectable.**

| Module                                    | Revision     |
| :---------------------------------------- | :----------- |
| `Cisco-IOS-XE-wireless-ap-global-oper`    | `2022-11-01` |
| `Cisco-IOS-XE-wireless-types`             | `2023-08-20` |
| `Cisco-IOS-XE-wireless-access-point-oper` | `2023-08-01` |
| `Cisco-IOS-XE-wireless-client-types`      | `2023-07-01` |
| `Cisco-IOS-XE-wireless-mobility-types`    | `2022-11-01` |
| `Cisco-IOS-XE-wireless-enum-types`        | `2023-07-20` |

Nothing compares this page against the exporter's own tables automatically. It transcribes the twelve tables in `internal/collector/enum.go`, which hold **221 spellings** between them, and every block below states how many members its enumeration has — so a table whose rows do not match its stated count has drifted. A count cannot catch two spellings whose values are exchanged, which is why the transcription is reviewed against the source rather than counted.

## AP collector

### `wnc_ap_oper_state`

Reads `ap-operation-state` — typedef `enum-ap-state` of `Cisco-IOS-XE-wireless-access-point-oper`, 6 members.

| Value | Spelling          |
| :---- | :---------------- |
| 1     | `ap-down`         |
| 2     | `ap-up`           |
| 3     | `unregistered`    |
| 4     | `registered`      |
| 5     | `downloading`     |
| 6     | `pre-downloading` |

### `wnc_ap_last_discovery_failure_reason`

Reads `last-disc-failure-type` — typedef `ap-discovery-failure-reason` of `Cisco-IOS-XE-wireless-ap-global-oper`, 17 members.

| Value | Spelling                               |
| :---- | :------------------------------------- |
| 0     | `disc-fail-none`                       |
| 1     | `disc-fail-req-dec-board-data`         |
| 2     | `disc-fail-req-dec-rad-info`           |
| 3     | `disc-fail-req-dec-wtp-dscrptr`        |
| 4     | `disc-fail-req-max-conc-wtp-dwnlds`    |
| 5     | `disc-fail-req-high-prity-max-apjoin`  |
| 6     | `disc-fail-req-max-wtp-joined`         |
| 7     | `disc-fail-req-max-conc-wtp-joins`     |
| 8     | `disc-fail-resp-enc-dscrptr`           |
| 9     | `disc-fail-resp-enc-acname`            |
| 10    | `disc-fail-resp-enc-ipv4-addr`         |
| 11    | `disc-fail-resp-enc-ipv6-addr`         |
| 12    | `disc-fail-resp-enc-mwar-payld`        |
| 13    | `disc-fail-resp-enc-wtp-rad-info`      |
| 14    | `disc-fail-resp-send-fail`             |
| 15    | `disc-fail-req-non-wireless-mgmt-intf` |
| 16    | `disc-fail-req-un-reg-license-mgr`     |

### `wnc_ap_last_join_failure_reason`

Reads `last-join-failure-type` — typedef `enm-ap-join-failure-reason` of `Cisco-IOS-XE-wireless-ap-global-oper`, 42 members.

| Value | Spelling                         |
| :---- | :------------------------------- |
| 0     | `jf-none`                        |
| 1     | `jf-reqrej-swver`                |
| 2     | `jf-reqrej-hwver`                |
| 3     | `jf-reqrej-bootver`              |
| 4     | `jf-reqrej-wtpdescrptr`          |
| 5     | `jf-reqrej-unsupportedwtp`       |
| 6     | `jf-reqrej-notfabric`            |
| 7     | `jf-reqrej-modelnum`             |
| 8     | `jf-reqrej-serialnum`            |
| 9     | `jf-reqrej-boardid`              |
| 10    | `jf-reqrej-boardrev`             |
| 11    | `jf-reqrej-basemacaddr`          |
| 12    | `jf-reqrej-locationdata`         |
| 13    | `jf-reqrej-wtpname`              |
| 14    | `jf-reqrej-wtpipv4addr`          |
| 15    | `jf-reqrej-boarddataopt`         |
| 16    | `jf-reqrej-invalid-radio`        |
| 17    | `jf-reqrej-maxmsgsize`           |
| 18    | `jf-reqrej-sessionid`            |
| 19    | `jf-resp-wtpradioinfo`           |
| 20    | `jf-resp-maxmsglen`              |
| 21    | `jf-resp-acdscrptr`              |
| 22    | `jf-resp-acname`                 |
| 23    | `jf-resp-cntrlipv4addr`          |
| 24    | `jf-resp-mwartypepayload`        |
| 25    | `jf-resp-authtokenpayload`       |
| 26    | `jf-resp-dudplite`               |
| 27    | `jf-delete-progress`             |
| 28    | `jf-resp-respsendf`              |
| 29    | `jf-ap-auth-pending`             |
| 30    | `jf-reqrej-capwapcapab`          |
| 31    | `jf-dtls-alert-from-peer`        |
| 32    | `jf-internal-error`              |
| 33    | `jf-idb-creation-failed`         |
| 34    | `jf-resp-cntrlipv6addr`          |
| 35    | `jf-resp-efficientimagedownload` |
| 36    | `jf-maxrexmitreached`            |
| 37    | `jf-heartbeattimer`              |
| 38    | `jf-hwfailure`                   |
| 39    | `jf-ap-auth-failure`             |
| 40    | `jf-invalid-mtu`                 |
| 41    | `jf-dtls-version`                |

### `wnc_ap_last_config_failure_reason`

Reads `last-config-failure-type` — typedef `enm-ap-config-failure-reason` of `Cisco-IOS-XE-wireless-ap-global-oper`, 14 members.

| Value | Spelling                     |
| :---- | :--------------------------- |
| 0     | `cf-none`                    |
| 1     | `cf-reqrej-unknown-ap`       |
| 2     | `cf-reqrej-reg-domain-check` |
| 3     | `cf-req-rej-capwap-data`     |
| 4     | `cf-reqrej-inval-reg-domain` |
| 5     | `cf-resp-build-fail`         |
| 6     | `cf-resp-send-fail`          |
| 7     | `cf-dtls-close-alert`        |
| 8     | `cf-internal-error`          |
| 9     | `cf-process-fail`            |
| 10    | `cf-max-rexmit`              |
| 11    | `cf-heartbeat-timer`         |
| 12    | `cf-hw-fail`                 |
| 13    | `cf-echo-req-fail`           |

### `wnc_ap_last_error_phase`

Reads `last-error-type` — typedef `last-failure-phase` of `Cisco-IOS-XE-wireless-ap-global-oper`, 7 members.

| Value | Spelling                   |
| :---- | :------------------------- |
| 0     | `ap-con-failure-unknown`   |
| 1     | `ap-con-failure-discovery` |
| 2     | `ap-con-failure-dtls`      |
| 3     | `ap-con-failure-join`      |
| 4     | `ap-con-failure-config`    |
| 5     | `ap-con-failure-imgdwnld`  |
| 6     | `ap-con-failure-run`       |

### `wnc_ap_last_dtls_failure_reason`

Reads `ctrl-dtls-failure-type` and `data-dtls-failure-type`, one per `channel` — typedef `enm-dtls-handshake-failure-reason` of `Cisco-IOS-XE-wireless-ap-global-oper`, 10 members.

| Value | Spelling                   |
| :---- | :------------------------- |
| 0     | `dtls-hs-success`          |
| 1     | `dtls-hs-err`              |
| 2     | `dtls-hs-cert-auth`        |
| 3     | `dtls-hs-aaa-auth`         |
| 4     | `dtls-hs-timer-exp`        |
| 5     | `dtls-hs-peer-alert`       |
| 6     | `dtls-hs-server-shut`      |
| 7     | `dtls-hs-config-not-done`  |
| 8     | `dtls-hs-unsupp-protocol`  |
| 9     | `dtls-hs-no-shared-cipher` |

### `wnc_ap_last_reboot_reason`

Reads `reboot-reason` — typedef `spam-ap-reboot-reason` of `Cisco-IOS-XE-wireless-types`, 59 members.

| Value | Spelling                                                |
| :---- | :------------------------------------------------------ |
| 0     | `ap-reboot-reason-none`                                 |
| 1     | `ap-reboot-reason-11-g-mode`                            |
| 2     | `ap-reboot-reason-ip-addr-set`                          |
| 3     | `ap-reboot-reason-ip-addr-reset`                        |
| 4     | `ap-reboot-reason-reboot-cmd`                           |
| 5     | `ap-reboot-reason-dhcp-fallback`                        |
| 6     | `ap-reboot-reason-discovery`                            |
| 7     | `ap-reboot-reason-join-resp`                            |
| 8     | `ap-reboot-reason-deny-join`                            |
| 9     | `ap-reboot-reason-config-resp`                          |
| 10    | `ap-reboot-reason-config-mwar`                          |
| 11    | `ap-reboot-reason-img-upgrade`                          |
| 12    | `ap-reboot-reason-img-opcode`                           |
| 13    | `ap-reboot-reason-img-chksum`                           |
| 14    | `ap-reboot-reason-img-data`                             |
| 15    | `ap-reboot-reason-cfgfile`                              |
| 16    | `ap-reboot-reason-img-error`                            |
| 17    | `ap-reboot-reason-ap-reboot-cmd`                        |
| 18    | `ap-reboot-reason-rap-ota-map`                          |
| 19    | `ap-reboot-reason-power-low`                            |
| 20    | `ap-reboot-reason-power-high`                           |
| 21    | `ap-reboot-reason-power-loss`                           |
| 22    | `ap-reboot-reason-power-chg`                            |
| 23    | `ap-reboot-reason-comp-fail`                            |
| 24    | `ap-reboot-reason-watchdog`                             |
| 25    | `ap-reboot-reason-lsc-enabled`                          |
| 26    | `ap-reboot-reason-lsc-disabled`                         |
| 27    | `ap-reboot-reason-lsc-provision-timeout`                |
| 28    | `ap-reboot-reason-lsc-max-prov-retry`                   |
| 29    | `ap-reboot-reason-lsc-load-failure`                     |
| 30    | `ap-reboot-reason-lsc-join-failure`                     |
| 31    | `ap-reboot-reason-capwap-timer-failure`                 |
| 32    | `ap-reboot-reason-fail-disc-with-dhcp-ip`               |
| 33    | `ap-reboot-reason-vlan-tag-failover`                    |
| 34    | `ap-reboot-reason-vlan-tag-retry`                       |
| 35    | `ap-reboot-reason-ipv6-addr-set`                        |
| 36    | `ap-reboot-reason-mode-change`                          |
| 37    | `ap-reboot-reason-ap-type-changed-to-capwap`            |
| 38    | `ap-reboot-reason-ap-type-changed-to-me`                |
| 39    | `ap-reboot-reason-erase-cfg-cmd`                        |
| 40    | `ap-reboot-reason-oeap-mode-cfg-upload`                 |
| 41    | `ap-reboot-reason-lag-cfg`                              |
| 42    | `ap-reboot-reason-fips-mode-change`                     |
| 43    | `ap-reboot-reason-diminished-pwr-change`                |
| 44    | `ap-reboot-reason-slub-debug`                           |
| 45    | `ap-reboot-reason-lsc-mode-capwap`                      |
| 46    | `ap-reboot-reason-lsc-mode-dot1x`                       |
| 47    | `ap-reboot-reason-lsc-mode-all`                         |
| 48    | `ap-reboot-reason-ap-type-changed-to-cloud`             |
| 49    | `ap-reboot-reason-dtls-init-failure`                    |
| 50    | `ap-reboot-reason-pnp-no-capwap-backoff`                |
| 51    | `ap-reboot-reason-day0-config-failure`                  |
| 52    | `ap-reboot-reason-day1-config-failure`                  |
| 53    | `ap-reboot-reason-pnp-triggered-reload`                 |
| 54    | `ap-reboot-reason-tri-radio-support`                    |
| 55    | `ap-reboot-reason-indoor-deployment`                    |
| 56    | `ap-reboot-reason-ap-type-changed-from-wgb-to-capwap`   |
| 57    | `ap-reboot-reason-ap-type-changed-from-cloud-to-capwap` |
| 58    | `ap-reboot-reason-ap-type-changed-to-wgb`               |

### `wnc_ap_last_disconnect_reason`

Reads `disconnect-reason` — typedef `spam-ap-disconnect-reason` of `Cisco-IOS-XE-wireless-types`, 41 members.

| Value | Spelling                               |
| :---- | :------------------------------------- |
| 0     | `unkown`                               |
| 1     | `wtp-post-join-timer-expired`          |
| 2     | `wtp-wait-dtls-timer-expired`          |
| 3     | `wtp-join-response-decode-failed`      |
| 4     | `wtp-img-data-resp-decode-failed`      |
| 5     | `wtp-config-status-decode-failed`      |
| 6     | `wtp-change-state-report-send-failed`  |
| 7     | `wtp-udi-info-send-failed`             |
| 8     | `wtp-data-dtls-init-failed`            |
| 9     | `wtp-heartbeat-timer-start-failed`     |
| 10    | `wtp-echo-timer-start-failed`          |
| 11    | `wtp-max-retransmission-reached`       |
| 12    | `wtp-found-master-mwar`                |
| 13    | `wtp-found-configured-primary-mwar`    |
| 14    | `wtp-found-configured-secondary-mwar`  |
| 15    | `wtp-found-configured-tertiary-mwar`   |
| 16    | `wtp-ip-addr-set-to-static`            |
| 17    | `wtp-ip-addr-reset`                    |
| 18    | `wtp-image-error`                      |
| 19    | `wtp-capwap-sm-restart`                |
| 20    | `wtp-controller-initiated-reason`      |
| 21    | `wtp-dtls-session-est-fail`            |
| 22    | `wtp-wait-dtls-no-join-response`       |
| 23    | `wtp-img-resp-error-image-rejected`    |
| 24    | `wtp-img-resp-err-db-entry-fetch-fail` |
| 25    | `wtp-img-req-err-db-entry-fetch-fail`  |
| 26    | `wtp-img-req-err-decode-fail`          |
| 27    | `wtp-img-req-err-img-data-resp-fail`   |
| 28    | `wtp-img-req-err-predownload-fail`     |
| 29    | `wtp-img-req-err-activate-fail`        |
| 30    | `wtp-reboot-mode-change-11g`           |
| 31    | `wtp-reboot-mode-change-wgb`           |
| 32    | `wtp-reboot-mode-change-me`            |
| 33    | `wtp-reboot-mode-change-cloud`         |
| 34    | `wtp-reboot-mode-change-capwap`        |
| 35    | `wtp-reboot-image-upgrade`             |
| 36    | `wtp-reboot-user-cmd`                  |
| 37    | `wtp-reboot-erase-cfg-cmd`             |
| 38    | `wtp-reboot-dimished-pwr-change`       |
| 39    | `wtp-capwap-cli-restart`               |
| 40    | `wtp-reboot-mode-change-site-survey`   |

## Client collector

### `wnc_client_state`

Reads `co-state` — typedef `client-co-state` of `Cisco-IOS-XE-wireless-client-types`, 14 members.

| Value | Spelling                                   |
| :---- | :----------------------------------------- |
| 0     | `client-status-idle`                       |
| 1     | `client-status-associating`                |
| 2     | `client-status-associated`                 |
| 3     | `client-status-authenticating`             |
| 4     | `client-status-authenticated`              |
| 5     | `client-status-mobility-discovery`         |
| 6     | `client-status-mobility-complete`          |
| 7     | `client-status-ip-learning`                |
| 8     | `client-status-ip-learn-complete`          |
| 9     | `client-status-webauth-required`           |
| 10    | `client-status-static-ip-anchor-discovery` |
| 11    | `client-status-run`                        |
| 12    | `client-status-delete-in-progress`         |
| 13    | `client-status-deleted`                    |

### `wnc_client_roam_type`

Reads `dot11-roam-type` — typedef `dot11-client-roam-type` of `Cisco-IOS-XE-wireless-mobility-types`, 5 members.

| Value | Spelling                   |
| :---- | :------------------------- |
| 0     | `dot11-roam-type-none`     |
| 1     | `dot11-roam-type-slow-11i` |
| 2     | `dot11-roam-type-fast-okc` |
| 3     | `dot11-roam-type-cckm`     |
| 4     | `dot11-roam-type-fast-11r` |

## WLAN collector

### `wnc_wlan_pmf_state`

Reads `pmf-options` — typedef `apf-vap-pmf-policies` of `Cisco-IOS-XE-wireless-enum-types`, 3 members.

| Value | Spelling               |
| :---- | :--------------------- |
| 0     | `apf-vap-pmf-disabled` |
| 1     | `apf-vap-pmf-optional` |
| 2     | `apf-vap-pmf-required` |

### `wnc_wlan_ft_state`

Reads `ft-mode` — typedef `ft-dot11r-mode` of `Cisco-IOS-XE-wireless-enum-types`, 3 members.

| Value | Spelling                  |
| :---- | :------------------------ |
| 0     | `dot11r-disabled`         |
| 1     | `dot11r-enabled`          |
| 2     | `dot11r-adaptive-enabled` |
