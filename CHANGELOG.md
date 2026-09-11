# Changelog

Notable changes to the metric surface, one section per release, listing the pull requests that release carries.

## [Unreleased]

- [#81](https://github.com/umatare5/cisco-wnc-exporter/pull/81) — Sync the CI and pre-commit stack with xflow
- [#82](https://github.com/umatare5/cisco-wnc-exporter/pull/82) — Adopt the xflow documentation conventions
- [#83](https://github.com/umatare5/cisco-wnc-exporter/pull/83) — Raise Endpoints to a section of its own
- [#84](https://github.com/umatare5/cisco-wnc-exporter/pull/84) — Update dependency golangci/golangci-lint to v2.13.2
- [#85](https://github.com/umatare5/cisco-wnc-exporter/pull/85) — Replace configuration.md with a help.md transcript
- [#86](https://github.com/umatare5/cisco-wnc-exporter/pull/86) — Extend the shared Renovate profile and pin the Alpine tag
- [#88](https://github.com/umatare5/cisco-wnc-exporter/pull/88) — Run the build and tests weekly
- [#89](https://github.com/umatare5/cisco-wnc-exporter/pull/89) — Report failed scheduled runs and build a weekly release snapshot
- [#90](https://github.com/umatare5/cisco-wnc-exporter/pull/90) — Align headings and license spelling with controld-exporter
- [#91](https://github.com/umatare5/cisco-wnc-exporter/pull/91) — Add govulncheck badge to README
- [#93](https://github.com/umatare5/cisco-wnc-exporter/pull/93) — Link the shared baseline and narrow the release archive

## [v0.14.1]

- [#79](https://github.com/umatare5/cisco-wnc-exporter/pull/79) — Sync the release, hook and linter toolchain configuration
- [#80](https://github.com/umatare5/cisco-wnc-exporter/pull/80) — Raise the SDK to v0.11.1 and release v0.14.1

## [v0.14.0]

- [#78](https://github.com/umatare5/cisco-wnc-exporter/pull/78) — Adopt SDK v0.10.1 and fix two band-index reading defects

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - `wnc_ap_tx_power_dbm` and `wnc_ap_tx_power_max_dbm` — either can now be absent on a multi-band radio whose band index is omitted, so a rule expecting one series per radio sees fewer.
> - `wnc_client_protocol` — reports `0` where it reported `1` for a client the controller has not classified, so a rule matching `== 1` stops counting it.

## [v0.13.3]

- [#77](https://github.com/umatare5/cisco-wnc-exporter/pull/77) — Raise the toolchain to Go 1.27 and release v0.13.3

## [v0.13.2]

- [#75](https://github.com/umatare5/cisco-wnc-exporter/pull/75) — Update grafana dashboards
- [#76](https://github.com/umatare5/cisco-wnc-exporter/pull/76) — Retarget IOS-XE version references to 17.15, release v0.13.2

## [v0.13.1]

- [#70](https://github.com/umatare5/cisco-wnc-exporter/pull/70) — Update description for wnc_up metric in README
- [#71](https://github.com/umatare5/cisco-wnc-exporter/pull/71) — Fit every HELP string in the exporter to 100 characters
- [#72](https://github.com/umatare5/cisco-wnc-exporter/pull/72) — Drop gemini cli support and remove copilot instructions
- [#73](https://github.com/umatare5/cisco-wnc-exporter/pull/73) — Merge the AP coordinate module and relocate test helpers
- [#74](https://github.com/umatare5/cisco-wnc-exporter/pull/74) — Fit alert descriptions in 240 characters, release v0.13.1

## [v0.13.0]

- [#65](https://github.com/umatare5/cisco-wnc-exporter/pull/65) — Update module github.com/urfave/cli/v3 to v3.11.0
- [#66](https://github.com/umatare5/cisco-wnc-exporter/pull/66) — Withhold the channel series on a zero leaf
- [#67](https://github.com/umatare5/cisco-wnc-exporter/pull/67) — Correct the comments and claims an audit found false
- [#68](https://github.com/umatare5/cisco-wnc-exporter/pull/68) — Merge the AP satellite files into ap.go
- [#69](https://github.com/umatare5/cisco-wnc-exporter/pull/69) — Restructure every page per metric, release v0.13.0

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - `wnc_ap_channel_number` and `wnc_ap_channel_width_mhz` — either can now be absent where the leaf it reads is omitted or zero. Names, types and labels are unchanged, and a rule comparing either to `0` stops matching.

## [v0.12.0]

- [#59](https://github.com/umatare5/cisco-wnc-exporter/pull/59) — Bump the SDK and withhold an omitted leaf
- [#60](https://github.com/umatare5/cisco-wnc-exporter/pull/60) — Add AP coordinates, client labels, release v0.12.0

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - **Twenty-eight families** — any can now be absent where the controller omits the leaf it reads, and six similar families still decode an omitted leaf as `0` on purpose. Names, types and labels are unchanged.
> - `wnc_refresh_defaults_fallback_total` — says whether a controller is on the affected path, and [Absence](docs/README.md#absence) carries the semantics.

## [v0.11.0]

- [#57](https://github.com/umatare5/cisco-wnc-exporter/pull/57) — Add three timestamp families and correct six documented claims
- [#58](https://github.com/umatare5/cisco-wnc-exporter/pull/58) — Value-encode twelve gauges, release v0.11.0

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - **Twelve gauges** — drop the `state` label and publish the number the controller's own enumeration assigns that spelling, which [`docs/enums.md`](docs/enums.md) lists. Every name, type and remaining label is unchanged.
> - **A query over those twelve** — a negated `state` matcher now matches everything, and a `sum` over one of them has to become a `count`.
> - **An unnumbered spelling** — withheld rather than published with a wrong value, so one of the twelve can be absent for a subject it always carried. The spelling reaches only the `--log.level=debug` log.

## [v0.10.0]

- [#56](https://github.com/umatare5/cisco-wnc-exporter/pull/56) — Add nine metric families, close two fabricated zeros, and release v0.10.0

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - `wnc_ap_air_quality_index` — renamed, so a query naming it returns nothing.
> - **Twenty-one per-radio series** — a rule that expected one series per slot now sees fewer, where a remote-LAN port had reported a fabricated `0`.

## [v0.9.2]

- [#55](https://github.com/umatare5/cisco-wnc-exporter/pull/55) — Name the contract the refresher promises and release v0.9.2

## [v0.9.1]

- [#54](https://github.com/umatare5/cisco-wnc-exporter/pull/54) — Make the refresher test deterministic and release v0.9.1

## [v0.9.0]

- [#53](https://github.com/umatare5/cisco-wnc-exporter/pull/53) — Add four metric families, correct two AP series, and release v0.9.0

> [!IMPORTANT]
>
> ### BREAKING CHANGE
>
> - `wnc_ap_fragmentation_rx_total` and `wnc_ap_fragmentation_tx_total` — renamed, so a query naming either returns nothing.
> - `wnc_ap_radio_state` and `wnc_ap_admin_state` — withheld where they had reported a fabricated `0`, so a rule reading a permanently failing radio now sees nothing.

## [v0.8.0]

- [#44](https://github.com/umatare5/cisco-wnc-exporter/pull/44) — Record device-side verification rules for YANG and RESTCONF
- [#52](https://github.com/umatare5/cisco-wnc-exporter/pull/52) — Ship third-party licence notices and prepare the v0.8.0 release

## [v0.7.0]

- [#51](https://github.com/umatare5/cisco-wnc-exporter/pull/51) — Add the CAPWAP join and controller-wide metrics for v0.7.0

## [v0.6.0]

- [#48](https://github.com/umatare5/cisco-wnc-exporter/pull/48) — Update example Grafana dashboards
- [#49](https://github.com/umatare5/cisco-wnc-exporter/pull/49) — Correct dark mode image source in README
- [#50](https://github.com/umatare5/cisco-wnc-exporter/pull/50) — Fix the container image, cover the shipped alerts and release v0.6.0

## [v0.5.0]

- [#47](https://github.com/umatare5/cisco-wnc-exporter/pull/47) — Report three WLAN config states, fix wnc_up and release v0.5.0

## [v0.4.0]

- [#45](https://github.com/umatare5/cisco-wnc-exporter/pull/45) — Omit WLAN series for an absent container and harden the release path
- [#46](https://github.com/umatare5/cisco-wnc-exporter/pull/46) — Report the WLAN config values in force and release v0.4.0

## [v0.3.0]

- [#17](https://github.com/umatare5/cisco-wnc-exporter/pull/17) — Update module github.com/urfave/cli/v3 to v3.6.2
- [#18](https://github.com/umatare5/cisco-wnc-exporter/pull/18) — Update alpine:latest docker digest to 2510918
- [#19](https://github.com/umatare5/cisco-wnc-exporter/pull/19) — Update module github.com/urfave/cli/v3 to v3.7.0
- [#20](https://github.com/umatare5/cisco-wnc-exporter/pull/20) — Update module github.com/urfave/cli/v3 to v3.8.0
- [#21](https://github.com/umatare5/cisco-wnc-exporter/pull/21) — Update alpine:latest docker digest to 5b10f43
- [#22](https://github.com/umatare5/cisco-wnc-exporter/pull/22) — Update AI instructions, pre-commit, and docs
- [#23](https://github.com/umatare5/cisco-wnc-exporter/pull/23) — Update module github.com/urfave/cli/v3 to v3.9.0
- [#24](https://github.com/umatare5/cisco-wnc-exporter/pull/24) — Update umatare5/common action to v0.14.0
- [#25](https://github.com/umatare5/cisco-wnc-exporter/pull/25) — Update alpine:latest docker digest to 28bd5fe
- [#26](https://github.com/umatare5/cisco-wnc-exporter/pull/26) — Update module github.com/urfave/cli/v3 to v3.9.1
- [#27](https://github.com/umatare5/cisco-wnc-exporter/pull/27) — Update module github.com/urfave/cli/v3 to v3.10.0
- [#28](https://github.com/umatare5/cisco-wnc-exporter/pull/28) — Update module github.com/urfave/cli/v3 to v3.10.1
- [#29](https://github.com/umatare5/cisco-wnc-exporter/pull/29) — Update module github.com/prometheus/client_golang to v1.24.0
- [#30](https://github.com/umatare5/cisco-wnc-exporter/pull/30) — Update module github.com/prometheus/client_golang to v1.24.1
- [#31](https://github.com/umatare5/cisco-wnc-exporter/pull/31) — Prevent panic on partial WNC oper data
- [#32](https://github.com/umatare5/cisco-wnc-exporter/pull/32) — Count clients per radio and recover from panics
- [#33](https://github.com/umatare5/cisco-wnc-exporter/pull/33) — Serve scrapes from a background WNC data refresh and stop fabricating zeros
- [#34](https://github.com/umatare5/cisco-wnc-exporter/pull/34) — Report the radio band, power and noise for the band and channel in use
- [#35](https://github.com/umatare5/cisco-wnc-exporter/pull/35) — Rename, rescale and drop metrics that misstated what the WNC reports
- [#36](https://github.com/umatare5/cisco-wnc-exporter/pull/36) — Rename the absence contract test and close its unverified assertions
- [#37](https://github.com/umatare5/cisco-wnc-exporter/pull/37) — Serve metrics at the configured telemetry path
- [#38](https://github.com/umatare5/cisco-wnc-exporter/pull/38) — Correct the false documentation claims and refresh the coverage badge
- [#39](https://github.com/umatare5/cisco-wnc-exporter/pull/39) — Drop the cache methods nothing calls
- [#40](https://github.com/umatare5/cisco-wnc-exporter/pull/40) — Shrink the README into per-collector reference pages, fix make image
- [#41](https://github.com/umatare5/cisco-wnc-exporter/pull/41) — Pin every published metric value to the leaf it is read from
- [#42](https://github.com/umatare5/cisco-wnc-exporter/pull/42) — Withdraw two misreported WLAN config series and add a CHANGELOG
- [#43](https://github.com/umatare5/cisco-wnc-exporter/pull/43) — Bump version from 0.2.0 to 0.3.0

[Unreleased]: https://github.com/umatare5/cisco-wnc-exporter/compare/v0.14.1...HEAD
[v0.14.1]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.14.1
[v0.14.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.14.0
[v0.13.3]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.13.3
[v0.13.2]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.13.2
[v0.13.1]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.13.1
[v0.13.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.13.0
[v0.12.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.12.0
[v0.11.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.11.0
[v0.10.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.10.0
[v0.9.2]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.9.2
[v0.9.1]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.9.1
[v0.9.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.9.0
[v0.8.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.8.0
[v0.7.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.7.0
[v0.6.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.6.0
[v0.5.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.5.0
[v0.4.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.4.0
[v0.3.0]: https://github.com/umatare5/cisco-wnc-exporter/releases/tag/v0.3.0
