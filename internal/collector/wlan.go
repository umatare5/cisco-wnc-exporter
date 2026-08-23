// Package collector provides collectors for cisco-wnc-exporter.
package collector

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// WLANMetrics represents the configuration for WLAN metrics.
type WLANMetrics struct {
	General    bool
	Traffic    bool
	Config     bool
	Info       bool
	InfoLabels []string
}

// WLANCollector implements prometheus.Collector for WLAN metrics.
type WLANCollector struct {
	metrics        WLANMetrics
	infoDesc       *prometheus.Desc
	infoLabelNames []string
	src            wnc.WLANSource
	clientSrc      wnc.ClientSource

	enabledDesc               *prometheus.Desc
	clientCountDesc           *prometheus.Desc
	onboardingDesc            *prometheus.Desc
	dataUsageDesc             *prometheus.Desc
	authPskDesc               *prometheus.Desc
	authDot1xDesc             *prometheus.Desc
	authDot1xSha256Desc       *prometheus.Desc
	wpa2EnabledDesc           *prometheus.Desc
	wpa3EnabledDesc           *prometheus.Desc
	sessionTimeoutDesc        *prometheus.Desc
	loadBalanceDesc           *prometheus.Desc
	wlan11kNeighDesc          *prometheus.Desc
	clientSteeringDesc        *prometheus.Desc
	centralSwitchingDesc      *prometheus.Desc
	centralAuthenticationDesc *prometheus.Desc
	centralDHCPDesc           *prometheus.Desc
	centralAssocEnableDesc    *prometheus.Desc
	policyEnabledDesc         *prometheus.Desc
	pmfStateDesc              *prometheus.Desc
	ftStateDesc               *prometheus.Desc
	policyBindingDesc         *prometheus.Desc
}

// NewWLANCollector creates a new WLAN collector.
func NewWLANCollector(src wnc.WLANSource, clientSrc wnc.ClientSource, metrics WLANMetrics) *WLANCollector {
	collector := &WLANCollector{
		src:       src,
		clientSrc: clientSrc,
		metrics:   metrics,
	}

	labels := []string{labelID}

	if metrics.General {
		collector.enabledDesc = prometheus.NewDesc(
			"wnc_wlan_enabled",
			"WLAN status (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
	}

	if metrics.Traffic {
		collector.clientCountDesc = prometheus.NewDesc(
			"wnc_wlan_clients",
			"Number of clients in the run state on this WLAN",
			labels, nil,
		)
		collector.onboardingDesc = prometheus.NewDesc(
			"wnc_wlan_onboarding_clients",
			"Number of clients on this WLAN currently held in one onboarding phase, "+
				"short of the run state wnc_wlan_clients counts",
			[]string{labelID, labelPhase}, nil,
		)
		collector.dataUsageDesc = prometheus.NewDesc(
			"wnc_wlan_data_usage_bytes_total",
			"Bytes transferred on this WLAN in both directions, as the controller totals "+
				"them. It keeps the bytes of clients that have since disconnected, so it is "+
				"not the sum of the per-client byte counters",
			labels, nil,
		)
	}

	if metrics.Config {
		collector.authPskDesc = prometheus.NewDesc(
			"wnc_wlan_auth_psk_enabled",
			"PSK authentication enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.authDot1xDesc = prometheus.NewDesc(
			"wnc_wlan_auth_dot1x_enabled",
			"802.1x authentication enabled (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.authDot1xSha256Desc = prometheus.NewDesc(
			"wnc_wlan_auth_dot1x_sha256_enabled",
			"802.1x SHA256 authentication enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.wpa2EnabledDesc = prometheus.NewDesc(
			"wnc_wlan_wpa2_enabled",
			"WPA2 support enabled (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.wpa3EnabledDesc = prometheus.NewDesc(
			"wnc_wlan_wpa3_enabled",
			"WPA3 support enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.sessionTimeoutDesc = prometheus.NewDesc(
			"wnc_wlan_session_timeout_seconds",
			"Session timeout duration in seconds. Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.loadBalanceDesc = prometheus.NewDesc(
			"wnc_wlan_load_balance_enabled",
			"Load balancing enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.wlan11kNeighDesc = prometheus.NewDesc(
			"wnc_wlan_11k_neighbor_list_enabled",
			"802.11k neighbor list enabled (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.clientSteeringDesc = prometheus.NewDesc(
			"wnc_wlan_client_steering_enabled",
			"6GHz client steering enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.centralSwitchingDesc = prometheus.NewDesc(
			"wnc_wlan_central_switching_enabled",
			"Central switching enabled (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.centralAuthenticationDesc = prometheus.NewDesc(
			"wnc_wlan_central_authentication_enabled",
			"Central authentication enabled (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.centralDHCPDesc = prometheus.NewDesc(
			"wnc_wlan_central_dhcp_enabled",
			"Central DHCP enabled (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.centralAssocEnableDesc = prometheus.NewDesc(
			"wnc_wlan_central_association_enabled",
			"Central association enabled (0=disabled, 1=enabled). Absent when the controller omits the leaf",
			labels, nil,
		)
		collector.policyEnabledDesc = prometheus.NewDesc(
			"wnc_wlan_policy_enabled",
			"Policy profile bound to this WLAN is active (0=shut down or not reported, 1=active)",
			labels, nil,
		)
		// Protected management frames has three configurations, and the middle one
		// admits an unprotected association, so the setting is published as the value
		// the controller assigns it rather than collapsed to a boolean.
		collector.pmfStateDesc = prometheus.NewDesc(
			"wnc_wlan_pmf_state",
			"Protected management frames setting, as the value the controller's own "+
				"enumeration assigns its spelling (0=apf-vap-pmf-disabled, "+
				"1=apf-vap-pmf-optional, 2=apf-vap-pmf-required). It covers 2.4GHz and 5GHz "+
				"— a 6GHz BSS requires PMF whatever this reports",
			labels, nil,
		)
		collector.ftStateDesc = prometheus.NewDesc(
			"wnc_wlan_ft_state",
			"802.11r fast transition mode, as the value the controller's own enumeration "+
				"assigns its spelling (0=dot11r-disabled, 1=dot11r-enabled, "+
				"2=dot11r-adaptive-enabled). Match by equality: 2 is a third mode for clients "+
				"that cannot use the FT AKM, not a stronger form of 1",
			labels, nil,
		)
		// The six policy series above name neither the tag nor the profile they read, so
		// this is what makes a WLAN bound through several tags observable.
		collector.policyBindingDesc = prometheus.NewDesc(
			"wnc_wlan_policy_binding",
			"Policy tag binding for this WLAN, always 1. One series per binding the "+
				"exporter can resolve, so more than one policy_profile for an id means the "+
				"six policy series report only one of the bound profiles",
			[]string{labelID, labelPolicyProfile, labelPolicyTag}, nil,
		)
	}

	if metrics.Info {
		requiredLabels := []string{"id"}
		availableLabels := []string{"name"}
		infoLabels := buildInfoLabels(requiredLabels, metrics.InfoLabels, availableLabels)
		collector.infoDesc = prometheus.NewDesc(
			"wnc_wlan_info",
			"WLAN information labels for joining with other metrics",
			infoLabels, nil,
		)
		collector.infoLabelNames = infoLabels
	}

	return collector
}

// Describe sends the descriptors of all metrics to the provided channel.
func (c *WLANCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.General {
		ch <- c.enabledDesc
	}
	if c.metrics.Traffic {
		ch <- c.clientCountDesc
		ch <- c.onboardingDesc
		ch <- c.dataUsageDesc
	}
	if c.metrics.Config {
		ch <- c.authPskDesc
		ch <- c.authDot1xDesc
		ch <- c.authDot1xSha256Desc
		ch <- c.wpa2EnabledDesc
		ch <- c.wpa3EnabledDesc
		ch <- c.sessionTimeoutDesc
		ch <- c.loadBalanceDesc
		ch <- c.wlan11kNeighDesc
		ch <- c.clientSteeringDesc
		ch <- c.centralSwitchingDesc
		ch <- c.centralAuthenticationDesc
		ch <- c.centralDHCPDesc
		ch <- c.centralAssocEnableDesc
		ch <- c.policyEnabledDesc
		ch <- c.pmfStateDesc
		ch <- c.ftStateDesc
		ch <- c.policyBindingDesc
	}
	if c.metrics.Info {
		ch <- c.infoDesc
	}
}

// Collect implements the prometheus.Collector interface.
func (c *WLANCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if !c.isAnyMetricFlagEnabled() {
		return
	}

	wlanConfigEntries, err := c.src.ListConfigEntries(ctx)
	if err != nil {
		slog.Debug("Failed to retrieve WLAN configuration entries", "error", err)
		return
	}

	var wlanToPolicyMap map[string]*wlan.WlanPolicy
	if IsEnabled(c.metrics.Config) {
		wlanPolicies, policyErr := c.src.ListPolicies(ctx)
		if policyErr != nil {
			slog.Debug("Failed to retrieve WLAN policies", "error", policyErr)
		}

		policyListEntries, listErr := c.src.ListPolicyListEntries(ctx)
		if listErr != nil {
			slog.Debug("Failed to retrieve WLAN policy list entries", "error", listErr)
		}

		// The mapping needs both: without the policy list entries no WLAN resolves
		// to a policy, and the config series would report every WLAN as having
		// central switching disabled and no session timeout.
		if policyErr == nil && listErr == nil {
			wlanToPolicyMap = buildWLANToPolicyMap(policyListEntries, wlanPolicies)
			c.collectPolicyBindings(ch, wlanConfigEntries, policyListEntries, wlanPolicies)
		}
	}

	var wlanStatsMap map[int]wlanStats
	var dataUsageMap map[int]uint64
	var onboardingMap map[int]ap.WlanClientStats
	if IsEnabled(c.metrics.Traffic) {
		clientData, clientErr := c.clientSrc.GetClientData(ctx)
		if clientErr != nil {
			slog.Debug("Failed to get client data for WLAN traffic metrics", "error", clientErr)
		}

		if clientErr == nil {
			wlanStatsMap = c.buildWLANStats(clientData)
		}

		// The byte counter comes from a different data type than the client count, so
		// one failing must not withhold the other.
		clientStats, statsErr := c.src.ListClientStats(ctx)
		if statsErr != nil {
			slog.Debug("Failed to get WLAN client statistics for traffic metrics", "error", statsErr)
		} else {
			dataUsageMap = buildWLANDataUsageMap(clientStats)
			onboardingMap = buildWLANOnboardingMap(clientStats)
		}
	}

	for _, entry := range wlanConfigEntries {
		if c.metrics.General {
			c.collectGeneralMetrics(ch, entry)
		}
		if c.metrics.Traffic {
			c.collectTrafficMetrics(ch, entry, wlanStatsMap, dataUsageMap, onboardingMap)
		}
		if c.metrics.Config {
			c.collectConfigMetrics(ch, entry, wlanToPolicyMap)
		}
		if c.metrics.Info {
			c.collectInfoMetrics(ch, entry)
		}
	}
}

// collectGeneralMetrics collects general WLAN metrics.
func (c *WLANCollector) collectGeneralMetrics(
	ch chan<- prometheus.Metric,
	entry wlan.WlanCfgEntry,
) {
	// The status lives in an optional container. Reporting zero for an absent
	// container would assert that the operator disabled the WLAN.
	if entry.APFVapIDData == nil {
		return
	}

	labels := []string{strconv.Itoa(entry.WlanID)}

	metrics := appendFlag(nil, c.enabledDesc, determineWLANEnabledStatus(entry))

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			prometheus.GaugeValue,
			metric.Value,
			labels...,
		)
	}
}

// collectTrafficMetrics collects traffic statistics for a WLAN.
func (c *WLANCollector) collectTrafficMetrics(
	ch chan<- prometheus.Metric,
	entry wlan.WlanCfgEntry,
	wlanStatsMap map[int]wlanStats,
	dataUsageMap map[int]uint64,
	onboardingMap map[int]ap.WlanClientStats,
) {
	labels := []string{strconv.Itoa(entry.WlanID)}

	// A nil map means client data was unavailable. Indexing it would publish a
	// zero client count, which reads as "no clients on this SSID".
	if wlanStatsMap != nil {
		stats := wlanStatsMap[entry.WlanID]

		ch <- prometheus.MustNewConstMetric(
			c.clientCountDesc,
			prometheus.GaugeValue,
			float64(stats.clientCount),
			labels...,
		)
	}

	// The counter is absent for a WLAN the controller lists no statistics record for,
	// which includes the case of a failed fetch.
	if usage, ok := dataUsageMap[entry.WlanID]; ok {
		ch <- prometheus.MustNewConstMetric(
			c.dataUsageDesc,
			prometheus.CounterValue,
			float64(usage),
			labels...,
		)
	}

	// The four phases are absent for a WLAN with no statistics record, so a WLAN the
	// controller says nothing about reports no phase rather than an empty one.
	if stats, ok := onboardingMap[entry.WlanID]; ok {
		for _, phase := range onboardingPhases {
			ch <- prometheus.MustNewConstMetric(
				c.onboardingDesc,
				prometheus.GaugeValue,
				float64(phase.count(stats)),
				strconv.Itoa(entry.WlanID), phase.name,
			)
		}
	}
}

// onboardingPhases pairs each phase label value with the leaf that counts it. The
// controller keeps one leaf per phase rather than an enumeration, so the label values
// are named here and are this exporter's own.
var onboardingPhases = []struct {
	name  string
	count func(ap.WlanClientStats) int
}{
	{"l2auth", func(s ap.WlanClientStats) int { return s.ClientCurrStateL2Auth }},
	{"mobility", func(s ap.WlanClientStats) int { return s.ClientCurrStateMobility }},
	{"iplearn", func(s ap.WlanClientStats) int { return s.ClientCurrStateIplearn }},
	{"webauth_pending", func(s ap.WlanClientStats) int { return s.CurrStateWebauthPending }},
}

// buildWLANOnboardingMap indexes the whole statistics record by WLAN identifier.
//
// The four phase counts are current counts rather than cumulative ones: the fifth
// count in the same record, the clients in the run state, equalled the per-WLAN client
// records exactly on every WLAN and in total, which is what types them as gauges.
// Whether the five partition a WLAN's clients was not measured, so they are not summed
// with wnc_wlan_clients, which counts only the run state.
func buildWLANOnboardingMap(clientStats []ap.WlanClientStats) map[int]ap.WlanClientStats {
	stats := make(map[int]ap.WlanClientStats, len(clientStats))

	for _, entry := range clientStats {
		stats[entry.WlanID] = entry
	}

	return stats
}

// buildWLANDataUsageMap indexes the byte counter by WLAN identifier.
//
// A record whose leaf the controller omitted is left out rather than read as zero. The
// leaf is a string on the wire, and the shared conversion reads both an omitted leaf
// and an unparsable one as zero, which on a counter is indistinguishable from a reset,
// so the parse is done here where the failure can withhold the series instead.
func buildWLANDataUsageMap(clientStats []ap.WlanClientStats) map[int]uint64 {
	usage := make(map[int]uint64, len(clientStats))

	for _, stats := range clientStats {
		value, err := strconv.ParseUint(stats.DataUsage, 10, 64)
		if err != nil {
			slog.Debug("skipped a WLAN whose data usage leaf is unreadable", "id", stats.WlanID)
			continue
		}

		usage[stats.WlanID] = value
	}

	return usage
}

type wlanStats struct {
	clientCount int
}

// collectConfigMetrics collects config metrics.
func (c *WLANCollector) collectConfigMetrics(
	ch chan<- prometheus.Metric,
	entry wlan.WlanCfgEntry,
	policyMap map[string]*wlan.WlanPolicy,
) {
	labels := []string{strconv.Itoa(entry.WlanID)}

	// The read asks for the values in force, so an entry leaf is omitted here only
	// when the controller rejected that request or ignored it. These five still
	// decode by value, so an omitted one reads as a configured false.
	metrics := []Float64Metric{
		{c.authPskDesc, boolToFloat64(entry.AuthKeyMgmtPsk)},
		{c.authDot1xSha256Desc, boolToFloat64(entry.AuthKeyMgmtDot1xSha256)},
		{c.wpa3EnabledDesc, boolToFloat64(entry.WPA3Enabled)},
		{c.loadBalanceDesc, boolToFloat64(entry.LoadBalance)},
		{c.clientSteeringDesc, boolToFloat64(entry.ClientSteering)},
	}

	// These three are the leaves a plain read omits from exactly the WLANs where the
	// setting is on, which a value-typed decode published as its inverse.
	metrics = appendFlag(metrics, c.authDot1xDesc, entry.AuthKeyMgmtDot1x)
	metrics = appendFlag(metrics, c.wpa2EnabledDesc, entry.WPA2Enabled)
	metrics = appendFlag(metrics, c.wlan11kNeighDesc, entry.Wlan11kNeighList)

	// The policy-derived series need a resolved policy profile and the container
	// each leaf lives in. A WLAN not bound to a policy tag has no policy, an absent
	// container is not a disabled feature, and reporting zero would assert that
	// central switching is deliberately disabled and that no session timeout is set.
	if profile, ok := policyMap[entry.ProfileName]; ok {
		metrics = appendNumber(metrics, c.sessionTimeoutDesc, determineSessionTimeout(profile))

		// The controller omits these four per leaf rather than per container, so an
		// omitted one must not cost the other three their series. Each accessor owns
		// the container guard, so there is none here to drift from them.
		metrics = appendFlag(metrics, c.centralSwitchingDesc, determineCentralSwitchingValue(profile))
		metrics = appendFlag(metrics,
			c.centralAuthenticationDesc, determineCentralAuthenticationValue(profile))
		metrics = appendFlag(metrics, c.centralDHCPDesc, determineCentralDHCPValue(profile))
		metrics = appendFlag(metrics,
			c.centralAssocEnableDesc, determineCentralAssocEnableValue(profile))

		// Status sits on the profile itself, so no container guards it.
		metrics = append(metrics,
			Float64Metric{c.policyEnabledDesc, boolToFloat64(profile.Status)},
		)
	}

	// An omitted leaf means the default is in force rather than that the feature is
	// off, so a reading the controller did not send is withheld rather than published
	// as the value of the spelling that would have been its zero.
	emitEnumReading(ch, c.pmfStateDesc, wlanPMFPolicies, entry.PMFOptions, labels[0])
	emitEnumReading(ch, c.ftStateDesc, wlanFTModes, entry.FTMode, labels[0])

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			prometheus.GaugeValue,
			metric.Value,
			labels...,
		)
	}
}

// collectPolicyBindings publishes one series per binding the exporter can resolve.
// It is emitted here rather than in the per-WLAN loop because its label set changes
// per binding, and a WLAN can carry more than one.
func (c *WLANCollector) collectPolicyBindings(
	ch chan<- prometheus.Metric,
	wlanConfigEntries []wlan.WlanCfgEntry,
	policyListEntries []wlan.PolicyListEntry,
	wlanPolicies []wlan.WlanPolicy,
) {
	idByProfile := make(map[string]string, len(wlanConfigEntries))
	for _, entry := range wlanConfigEntries {
		idByProfile[entry.ProfileName] = strconv.Itoa(entry.WlanID)
	}

	resolvable := make(map[string]bool, len(wlanPolicies))
	for _, policy := range wlanPolicies {
		resolvable[policy.PolicyProfileName] = true
	}

	for _, entry := range policyListEntries {
		if entry.WLANPolicies == nil || entry.TagName == "" {
			continue
		}

		for _, mapping := range entry.WLANPolicies.WLANPolicy {
			// A tag naming a WLAN the controller does not define carries no identifier to
			// key the series by, and a binding whose policy profile is absent from
			// wlan-policies is skipped by the six series as well, so publishing either
			// would show a binding they are not reporting.
			id, defined := idByProfile[mapping.WLANProfileName]
			if !defined || !resolvable[mapping.PolicyProfileName] {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.policyBindingDesc, prometheus.GaugeValue, 1,
				id, mapping.PolicyProfileName, entry.TagName,
			)
		}
	}
}

// collectInfoMetrics collects info metrics.
func (c *WLANCollector) collectInfoMetrics(
	ch chan<- prometheus.Metric,
	entry wlan.WlanCfgEntry,
) {
	wlanIDStr := strconv.Itoa(entry.WlanID)

	wlanName := determineWLANName(entry)

	labelValues := c.buildWLANInfoLabelValues(wlanIDStr, wlanName)

	ch <- prometheus.MustNewConstMetric(
		c.infoDesc,
		prometheus.GaugeValue,
		1,
		labelValues...,
	)
}

// buildWLANStats counts the clients in the run state on each WLAN ID.
func (c *WLANCollector) buildWLANStats(clientData []client.CommonOperData) map[int]wlanStats {
	wlanStats := make(map[int]wlanStats)

	if clientData == nil {
		return wlanStats
	}

	for _, commonData := range clientData {
		if commonData.CoState != ClientStatusRun {
			continue
		}

		wlanID := commonData.WlanID

		stats := wlanStats[wlanID]
		stats.clientCount++

		wlanStats[wlanID] = stats
	}

	return wlanStats
}

// buildWLANInfoLabelValues constructs the label values array based on configured labels.
func (c *WLANCollector) buildWLANInfoLabelValues(id, name string) []string {
	labelNames := c.infoLabelNames
	values := make([]string, len(labelNames))

	valueMap := map[string]string{
		labelID:   id,
		labelName: name,
	}

	for i, label := range labelNames {
		if value, exists := valueMap[label]; exists {
			values[i] = value
		} else {
			values[i] = ""
		}
	}

	return values
}

// buildWLANToPolicyMap creates a map from WLAN profile names to policies using policy-list-entries.
func buildWLANToPolicyMap(
	policyListEntries []wlan.PolicyListEntry,
	wlanPolicies []wlan.WlanPolicy,
) map[string]*wlan.WlanPolicy {
	policyMap := make(map[string]*wlan.WlanPolicy)
	for i := range wlanPolicies {
		policy := &wlanPolicies[i]
		policyMap[policy.PolicyProfileName] = policy
	}

	wlanToPolicyMap := make(map[string]*wlan.WlanPolicy)
	for _, entry := range policyListEntries {
		if entry.WLANPolicies != nil {
			for _, mapping := range entry.WLANPolicies.WLANPolicy {
				if policy, exists := policyMap[mapping.PolicyProfileName]; exists {
					wlanToPolicyMap[mapping.WLANProfileName] = policy
				}
			}
		}
	}

	return wlanToPolicyMap
}

// determineWLANEnabledStatus returns the WLAN's administrative state, and nil when the
// controller sent neither the container nor the leaf.
func determineWLANEnabledStatus(wlanEntry wlan.WlanCfgEntry) *bool {
	if wlanEntry.APFVapIDData == nil {
		return nil
	}
	return wlanEntry.APFVapIDData.WlanStatus
}

// determineWLANName extracts WLAN name.
func determineWLANName(wlanEntry wlan.WlanCfgEntry) string {
	if wlanEntry.APFVapIDData != nil && wlanEntry.APFVapIDData.SSID != "" {
		return wlanEntry.APFVapIDData.SSID
	}
	return wlanEntry.ProfileName
}

// determineSessionTimeout returns the session timeout the policy sets, and nil when the
// controller sent neither the container nor the leaf. Zero is a valid timeout, so it
// cannot stand in for absence.
func determineSessionTimeout(policy *wlan.WlanPolicy) *int {
	if policy == nil || policy.WlanTimeout == nil {
		return nil
	}
	return policy.WlanTimeout.SessionTimeout
}

func determineCentralSwitchingValue(policy *wlan.WlanPolicy) *bool {
	if policy == nil || policy.WlanSwitchingPolicy == nil {
		return nil
	}
	return policy.WlanSwitchingPolicy.CentralSwitching
}

func determineCentralAuthenticationValue(policy *wlan.WlanPolicy) *bool {
	if policy == nil || policy.WlanSwitchingPolicy == nil {
		return nil
	}
	return policy.WlanSwitchingPolicy.CentralAuthentication
}

func determineCentralDHCPValue(policy *wlan.WlanPolicy) *bool {
	if policy == nil || policy.WlanSwitchingPolicy == nil {
		return nil
	}
	return policy.WlanSwitchingPolicy.CentralDHCP
}

func determineCentralAssocEnableValue(policy *wlan.WlanPolicy) *bool {
	if policy == nil || policy.WlanSwitchingPolicy == nil {
		return nil
	}
	return policy.WlanSwitchingPolicy.CentralAssocEnable
}

func (c *WLANCollector) isAnyMetricFlagEnabled() bool {
	return IsEnabled(c.metrics.General, c.metrics.Traffic, c.metrics.Config, c.metrics.Info)
}
