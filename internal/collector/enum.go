// Package collector provides utilities for WNC collectors.
// This file holds the value the controller's own enumeration assigns each spelling it
// sends in the twelve enum leaves this exporter publishes, and the emit that resolves
// one against the other.
//
// Every member of all twelve enumerations carries an explicit value statement in the
// schema the controller implements, so these are the controller's numbers rather than
// an ordering this exporter invented. They were read from these modules, at the
// revision the controller reported for each:
//
//   - Cisco-IOS-XE-wireless-ap-global-oper 2022-11-01
//   - Cisco-IOS-XE-wireless-types 2023-08-20
//   - Cisco-IOS-XE-wireless-access-point-oper 2023-08-01
//   - Cisco-IOS-XE-wireless-client-types 2023-07-01
//   - Cisco-IOS-XE-wireless-mobility-types 2022-11-01
//   - Cisco-IOS-XE-wireless-enum-types 2023-07-20
//
// The 221 spellings are unique across the twelve tables, which is what makes one
// shared type safe: a reading resolved against the wrong table finds nothing and is
// withheld rather than published as another enumeration's number.
package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// enumTable maps the spelling the controller sends to the value its own enumeration
// assigns that spelling.
type enumTable map[string]float64

// apDiscoveryFailureReasons holds ap-discovery-failure-reason of Cisco-IOS-XE-wireless-ap-global-oper.
var apDiscoveryFailureReasons = enumTable{
	"disc-fail-none":                       0,
	"disc-fail-req-dec-board-data":         1,
	"disc-fail-req-dec-rad-info":           2,
	"disc-fail-req-dec-wtp-dscrptr":        3,
	"disc-fail-req-max-conc-wtp-dwnlds":    4,
	"disc-fail-req-high-prity-max-apjoin":  5,
	"disc-fail-req-max-wtp-joined":         6,
	"disc-fail-req-max-conc-wtp-joins":     7,
	"disc-fail-resp-enc-dscrptr":           8,
	"disc-fail-resp-enc-acname":            9,
	"disc-fail-resp-enc-ipv4-addr":         10,
	"disc-fail-resp-enc-ipv6-addr":         11,
	"disc-fail-resp-enc-mwar-payld":        12,
	"disc-fail-resp-enc-wtp-rad-info":      13,
	"disc-fail-resp-send-fail":             14,
	"disc-fail-req-non-wireless-mgmt-intf": 15,
	"disc-fail-req-un-reg-license-mgr":     16,
}

// apJoinFailureReasons holds enm-ap-join-failure-reason of Cisco-IOS-XE-wireless-ap-global-oper.
var apJoinFailureReasons = enumTable{
	"jf-none":                        0,
	"jf-reqrej-swver":                1,
	"jf-reqrej-hwver":                2,
	"jf-reqrej-bootver":              3,
	"jf-reqrej-wtpdescrptr":          4,
	"jf-reqrej-unsupportedwtp":       5,
	"jf-reqrej-notfabric":            6,
	"jf-reqrej-modelnum":             7,
	"jf-reqrej-serialnum":            8,
	"jf-reqrej-boardid":              9,
	"jf-reqrej-boardrev":             10,
	"jf-reqrej-basemacaddr":          11,
	"jf-reqrej-locationdata":         12,
	"jf-reqrej-wtpname":              13,
	"jf-reqrej-wtpipv4addr":          14,
	"jf-reqrej-boarddataopt":         15,
	"jf-reqrej-invalid-radio":        16,
	"jf-reqrej-maxmsgsize":           17,
	"jf-reqrej-sessionid":            18,
	"jf-resp-wtpradioinfo":           19,
	"jf-resp-maxmsglen":              20,
	"jf-resp-acdscrptr":              21,
	"jf-resp-acname":                 22,
	"jf-resp-cntrlipv4addr":          23,
	"jf-resp-mwartypepayload":        24,
	"jf-resp-authtokenpayload":       25,
	"jf-resp-dudplite":               26,
	"jf-delete-progress":             27,
	"jf-resp-respsendf":              28,
	"jf-ap-auth-pending":             29,
	"jf-reqrej-capwapcapab":          30,
	"jf-dtls-alert-from-peer":        31,
	"jf-internal-error":              32,
	"jf-idb-creation-failed":         33,
	"jf-resp-cntrlipv6addr":          34,
	"jf-resp-efficientimagedownload": 35,
	"jf-maxrexmitreached":            36,
	"jf-heartbeattimer":              37,
	"jf-hwfailure":                   38,
	"jf-ap-auth-failure":             39,
	"jf-invalid-mtu":                 40,
	"jf-dtls-version":                41,
}

// apConfigFailureReasons holds enm-ap-config-failure-reason of Cisco-IOS-XE-wireless-ap-global-oper.
var apConfigFailureReasons = enumTable{
	"cf-none":                    0,
	"cf-reqrej-unknown-ap":       1,
	"cf-reqrej-reg-domain-check": 2,
	"cf-req-rej-capwap-data":     3,
	"cf-reqrej-inval-reg-domain": 4,
	"cf-resp-build-fail":         5,
	"cf-resp-send-fail":          6,
	"cf-dtls-close-alert":        7,
	"cf-internal-error":          8,
	"cf-process-fail":            9,
	"cf-max-rexmit":              10,
	"cf-heartbeat-timer":         11,
	"cf-hw-fail":                 12,
	"cf-echo-req-fail":           13,
}

// apFailurePhases holds last-failure-phase of Cisco-IOS-XE-wireless-ap-global-oper.
var apFailurePhases = enumTable{
	"ap-con-failure-unknown":   0,
	"ap-con-failure-discovery": 1,
	"ap-con-failure-dtls":      2,
	"ap-con-failure-join":      3,
	"ap-con-failure-config":    4,
	"ap-con-failure-imgdwnld":  5,
	"ap-con-failure-run":       6,
}

// apDTLSFailureReasons holds enm-dtls-handshake-failure-reason of Cisco-IOS-XE-wireless-ap-global-oper.
var apDTLSFailureReasons = enumTable{
	"dtls-hs-success":          0,
	"dtls-hs-err":              1,
	"dtls-hs-cert-auth":        2,
	"dtls-hs-aaa-auth":         3,
	"dtls-hs-timer-exp":        4,
	"dtls-hs-peer-alert":       5,
	"dtls-hs-server-shut":      6,
	"dtls-hs-config-not-done":  7,
	"dtls-hs-unsupp-protocol":  8,
	"dtls-hs-no-shared-cipher": 9,
}

// apRebootReasons holds spam-ap-reboot-reason of Cisco-IOS-XE-wireless-types.
var apRebootReasons = enumTable{
	"ap-reboot-reason-none":                                 0,
	"ap-reboot-reason-11-g-mode":                            1,
	"ap-reboot-reason-ip-addr-set":                          2,
	"ap-reboot-reason-ip-addr-reset":                        3,
	"ap-reboot-reason-reboot-cmd":                           4,
	"ap-reboot-reason-dhcp-fallback":                        5,
	"ap-reboot-reason-discovery":                            6,
	"ap-reboot-reason-join-resp":                            7,
	"ap-reboot-reason-deny-join":                            8,
	"ap-reboot-reason-config-resp":                          9,
	"ap-reboot-reason-config-mwar":                          10,
	"ap-reboot-reason-img-upgrade":                          11,
	"ap-reboot-reason-img-opcode":                           12,
	"ap-reboot-reason-img-chksum":                           13,
	"ap-reboot-reason-img-data":                             14,
	"ap-reboot-reason-cfgfile":                              15,
	"ap-reboot-reason-img-error":                            16,
	"ap-reboot-reason-ap-reboot-cmd":                        17,
	"ap-reboot-reason-rap-ota-map":                          18,
	"ap-reboot-reason-power-low":                            19,
	"ap-reboot-reason-power-high":                           20,
	"ap-reboot-reason-power-loss":                           21,
	"ap-reboot-reason-power-chg":                            22,
	"ap-reboot-reason-comp-fail":                            23,
	"ap-reboot-reason-watchdog":                             24,
	"ap-reboot-reason-lsc-enabled":                          25,
	"ap-reboot-reason-lsc-disabled":                         26,
	"ap-reboot-reason-lsc-provision-timeout":                27,
	"ap-reboot-reason-lsc-max-prov-retry":                   28,
	"ap-reboot-reason-lsc-load-failure":                     29,
	"ap-reboot-reason-lsc-join-failure":                     30,
	"ap-reboot-reason-capwap-timer-failure":                 31,
	"ap-reboot-reason-fail-disc-with-dhcp-ip":               32,
	"ap-reboot-reason-vlan-tag-failover":                    33,
	"ap-reboot-reason-vlan-tag-retry":                       34,
	"ap-reboot-reason-ipv6-addr-set":                        35,
	"ap-reboot-reason-mode-change":                          36,
	"ap-reboot-reason-ap-type-changed-to-capwap":            37,
	"ap-reboot-reason-ap-type-changed-to-me":                38,
	"ap-reboot-reason-erase-cfg-cmd":                        39,
	"ap-reboot-reason-oeap-mode-cfg-upload":                 40,
	"ap-reboot-reason-lag-cfg":                              41,
	"ap-reboot-reason-fips-mode-change":                     42,
	"ap-reboot-reason-diminished-pwr-change":                43,
	"ap-reboot-reason-slub-debug":                           44,
	"ap-reboot-reason-lsc-mode-capwap":                      45,
	"ap-reboot-reason-lsc-mode-dot1x":                       46,
	"ap-reboot-reason-lsc-mode-all":                         47,
	"ap-reboot-reason-ap-type-changed-to-cloud":             48,
	"ap-reboot-reason-dtls-init-failure":                    49,
	"ap-reboot-reason-pnp-no-capwap-backoff":                50,
	"ap-reboot-reason-day0-config-failure":                  51,
	"ap-reboot-reason-day1-config-failure":                  52,
	"ap-reboot-reason-pnp-triggered-reload":                 53,
	"ap-reboot-reason-tri-radio-support":                    54,
	"ap-reboot-reason-indoor-deployment":                    55,
	"ap-reboot-reason-ap-type-changed-from-wgb-to-capwap":   56,
	"ap-reboot-reason-ap-type-changed-from-cloud-to-capwap": 57,
	"ap-reboot-reason-ap-type-changed-to-wgb":               58,
}

// apDisconnectReasons holds spam-ap-disconnect-reason of Cisco-IOS-XE-wireless-types.
var apDisconnectReasons = enumTable{
	"unkown":                               0, //nolint:misspell
	"wtp-post-join-timer-expired":          1,
	"wtp-wait-dtls-timer-expired":          2,
	"wtp-join-response-decode-failed":      3,
	"wtp-img-data-resp-decode-failed":      4,
	"wtp-config-status-decode-failed":      5,
	"wtp-change-state-report-send-failed":  6,
	"wtp-udi-info-send-failed":             7,
	"wtp-data-dtls-init-failed":            8,
	"wtp-heartbeat-timer-start-failed":     9,
	"wtp-echo-timer-start-failed":          10,
	"wtp-max-retransmission-reached":       11,
	"wtp-found-master-mwar":                12,
	"wtp-found-configured-primary-mwar":    13,
	"wtp-found-configured-secondary-mwar":  14,
	"wtp-found-configured-tertiary-mwar":   15,
	"wtp-ip-addr-set-to-static":            16,
	"wtp-ip-addr-reset":                    17,
	"wtp-image-error":                      18,
	"wtp-capwap-sm-restart":                19,
	"wtp-controller-initiated-reason":      20,
	"wtp-dtls-session-est-fail":            21,
	"wtp-wait-dtls-no-join-response":       22,
	"wtp-img-resp-error-image-rejected":    23,
	"wtp-img-resp-err-db-entry-fetch-fail": 24,
	"wtp-img-req-err-db-entry-fetch-fail":  25,
	"wtp-img-req-err-decode-fail":          26,
	"wtp-img-req-err-img-data-resp-fail":   27,
	"wtp-img-req-err-predownload-fail":     28,
	"wtp-img-req-err-activate-fail":        29,
	"wtp-reboot-mode-change-11g":           30,
	"wtp-reboot-mode-change-wgb":           31,
	"wtp-reboot-mode-change-me":            32,
	"wtp-reboot-mode-change-cloud":         33,
	"wtp-reboot-mode-change-capwap":        34,
	"wtp-reboot-image-upgrade":             35,
	"wtp-reboot-user-cmd":                  36,
	"wtp-reboot-erase-cfg-cmd":             37,
	"wtp-reboot-dimished-pwr-change":       38,
	"wtp-capwap-cli-restart":               39,
	"wtp-reboot-mode-change-site-survey":   40,
}

// apOperationStates holds enum-ap-state of Cisco-IOS-XE-wireless-access-point-oper.
// The enumeration declares no member at 0.
var apOperationStates = enumTable{
	"ap-down":         1,
	"ap-up":           2,
	"unregistered":    3,
	"registered":      4,
	"downloading":     5,
	"pre-downloading": 6,
}

// clientStates holds client-co-state of Cisco-IOS-XE-wireless-client-types.
var clientStates = enumTable{
	"client-status-idle":                       0,
	"client-status-associating":                1,
	"client-status-associated":                 2,
	"client-status-authenticating":             3,
	"client-status-authenticated":              4,
	"client-status-mobility-discovery":         5,
	"client-status-mobility-complete":          6,
	"client-status-ip-learning":                7,
	"client-status-ip-learn-complete":          8,
	"client-status-webauth-required":           9,
	"client-status-static-ip-anchor-discovery": 10,
	"client-status-run":                        11,
	"client-status-delete-in-progress":         12,
	"client-status-deleted":                    13,
}

// clientRoamTypes holds dot11-client-roam-type of Cisco-IOS-XE-wireless-mobility-types.
var clientRoamTypes = enumTable{
	"dot11-roam-type-none":     0,
	"dot11-roam-type-slow-11i": 1,
	"dot11-roam-type-fast-okc": 2,
	"dot11-roam-type-cckm":     3,
	"dot11-roam-type-fast-11r": 4,
}

// wlanPMFPolicies holds apf-vap-pmf-policies of Cisco-IOS-XE-wireless-enum-types.
var wlanPMFPolicies = enumTable{
	"apf-vap-pmf-disabled": 0,
	"apf-vap-pmf-optional": 1,
	"apf-vap-pmf-required": 2,
}

// wlanFTModes holds ft-dot11r-mode of Cisco-IOS-XE-wireless-enum-types.
var wlanFTModes = enumTable{
	"dot11r-disabled":         0,
	"dot11r-enabled":          1,
	"dot11r-adaptive-enabled": 2,
}

// emitEnumReading publishes the value the controller's enumeration assigns the reading.
//
// A reading no table numbers is withheld rather than published as some other number: 0
// is a real member of eleven of the twelve enumerations, and in the twelfth it names no
// member at all, so no value is free to stand for a reading this release cannot name.
//
// The empty reading is withheld before the lookup so that it is not logged. An absent
// leaf is ordinary, and it is what a controller that rejects a request for the values in
// force returns for every leaf whose default is in force, while a spelling no table
// numbers is an anomaly whose spelling no query recovers.
func emitEnumReading(
	ch chan<- prometheus.Metric, desc *prometheus.Desc,
	table enumTable, reading string, labels ...string,
) {
	if reading == "" {
		return
	}

	value, ok := table[reading]
	if !ok {
		slog.Debug("Withheld a series for a spelling no enumeration of this release numbers",
			"spelling", reading)

		return
	}

	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, labels...)
}
