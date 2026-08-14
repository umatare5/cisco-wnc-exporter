// Package collector provides collectors for cisco-wnc-exporter.
package collector

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

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
	authPskDesc               *prometheus.Desc
	authDot1xDesc             *prometheus.Desc
	authDot1xSha256Desc       *prometheus.Desc
	wpa3EnabledDesc           *prometheus.Desc
	sessionTimeoutDesc        *prometheus.Desc
	loadBalanceDesc           *prometheus.Desc
	clientSteeringDesc        *prometheus.Desc
	centralSwitchingDesc      *prometheus.Desc
	centralAuthenticationDesc *prometheus.Desc
	centralDHCPDesc           *prometheus.Desc
	centralAssocEnableDesc    *prometheus.Desc
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
			"WLAN status (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
	}

	if metrics.Traffic {
		collector.clientCountDesc = prometheus.NewDesc(
			"wnc_wlan_clients",
			"Number of clients in the run state on this WLAN",
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
			"802.1x authentication enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.authDot1xSha256Desc = prometheus.NewDesc(
			"wnc_wlan_auth_dot1x_sha256_enabled",
			"802.1x SHA256 authentication enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.wpa3EnabledDesc = prometheus.NewDesc(
			"wnc_wlan_wpa3_enabled",
			"WPA3 support enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.sessionTimeoutDesc = prometheus.NewDesc(
			"wnc_wlan_session_timeout_seconds",
			"Session timeout duration in seconds, 0 when the controller omits the leaf",
			labels, nil,
		)
		collector.loadBalanceDesc = prometheus.NewDesc(
			"wnc_wlan_load_balance_enabled",
			"Load balancing enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.clientSteeringDesc = prometheus.NewDesc(
			"wnc_wlan_client_steering_enabled",
			"6GHz client steering enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.centralSwitchingDesc = prometheus.NewDesc(
			"wnc_wlan_central_switching_enabled",
			"Central switching enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.centralAuthenticationDesc = prometheus.NewDesc(
			"wnc_wlan_central_authentication_enabled",
			"Central authentication enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.centralDHCPDesc = prometheus.NewDesc(
			"wnc_wlan_central_dhcp_enabled",
			"Central DHCP enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
		)
		collector.centralAssocEnableDesc = prometheus.NewDesc(
			"wnc_wlan_central_association_enabled",
			"Central association enabled (0=disabled or not reported, 1=enabled)",
			labels, nil,
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
	}
	if c.metrics.Config {
		ch <- c.authPskDesc
		ch <- c.authDot1xDesc
		ch <- c.authDot1xSha256Desc
		ch <- c.wpa3EnabledDesc
		ch <- c.sessionTimeoutDesc
		ch <- c.loadBalanceDesc
		ch <- c.clientSteeringDesc
		ch <- c.centralSwitchingDesc
		ch <- c.centralAuthenticationDesc
		ch <- c.centralDHCPDesc
		ch <- c.centralAssocEnableDesc
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
		}
	}

	var wlanStatsMap map[int]wlanStats
	if IsEnabled(c.metrics.Traffic) {
		clientData, clientErr := c.clientSrc.GetClientData(ctx)
		if clientErr != nil {
			slog.Debug("Failed to get client data for WLAN traffic metrics", "error", clientErr)
		}

		if clientErr == nil {
			wlanStatsMap = c.buildWLANStats(clientData)
		}
	}

	for _, entry := range wlanConfigEntries {
		if c.metrics.General {
			c.collectGeneralMetrics(ch, entry)
		}
		if c.metrics.Traffic {
			c.collectTrafficMetrics(ch, entry, wlanStatsMap)
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

	metrics := []Float64Metric{
		{c.enabledDesc, float64(determineWLANEnabledStatus(entry))},
	}

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
) {
	// A nil map means client data was unavailable. Indexing it would publish a
	// zero client count, which reads as "no clients on this SSID".
	if wlanStatsMap == nil {
		return
	}

	labels := []string{strconv.Itoa(entry.WlanID)}
	stats := wlanStatsMap[entry.WlanID]

	ch <- prometheus.MustNewConstMetric(
		c.clientCountDesc,
		prometheus.GaugeValue,
		float64(stats.clientCount),
		labels...,
	)
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

	// The controller omits entry leaves left at their default values, and the
	// value-typed decode turns an omitted leaf into zero, which no query can
	// tell apart from a configured false.
	metrics := []Float64Metric{
		{c.authPskDesc, boolToFloat64(entry.AuthKeyMgmtPsk)},
		{c.authDot1xDesc, boolToFloat64(entry.AuthKeyMgmtDot1x)},
		{c.authDot1xSha256Desc, boolToFloat64(entry.AuthKeyMgmtDot1xSha256)},
		{c.wpa3EnabledDesc, boolToFloat64(entry.WPA3Enabled)},
		{c.loadBalanceDesc, boolToFloat64(entry.LoadBalance)},
		{c.clientSteeringDesc, boolToFloat64(entry.ClientSteering)},
	}

	// The policy-derived series need a resolved policy profile and the container
	// each leaf lives in. A WLAN not bound to a policy tag has no policy, an absent
	// container is not a disabled feature, and reporting zero would assert that
	// central switching is deliberately disabled and that no session timeout is set.
	if profile, ok := policyMap[entry.ProfileName]; ok {
		if profile.WlanTimeout != nil {
			metrics = append(metrics,
				Float64Metric{c.sessionTimeoutDesc, float64(determineSessionTimeout(profile))},
			)
		}

		if profile.WlanSwitchingPolicy != nil {
			metrics = append(metrics,
				Float64Metric{c.centralSwitchingDesc, determineCentralSwitchingValue(profile)},
				Float64Metric{c.centralAuthenticationDesc, determineCentralAuthenticationValue(profile)},
				Float64Metric{c.centralDHCPDesc, determineCentralDHCPValue(profile)},
				Float64Metric{c.centralAssocEnableDesc, determineCentralAssocEnableValue(profile)},
			)
		}
	}

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			prometheus.GaugeValue,
			metric.Value,
			labels...,
		)
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
	// Step 1: Build policy map by policy-profile-name for O(1) lookups
	policyMap := make(map[string]*wlan.WlanPolicy)
	for i := range wlanPolicies {
		policy := &wlanPolicies[i]
		policyMap[policy.PolicyProfileName] = policy
	}

	// Step 2: Build WLAN profile to policy mapping using policy-list-entries
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

// determineWLANEnabledStatus extracts WLAN enabled status.
func determineWLANEnabledStatus(wlanEntry wlan.WlanCfgEntry) int {
	if wlanEntry.APFVapIDData != nil && wlanEntry.APFVapIDData.WlanStatus {
		return 1
	}
	return 0
}

// determineWLANName extracts WLAN name.
func determineWLANName(wlanEntry wlan.WlanCfgEntry) string {
	if wlanEntry.APFVapIDData != nil && wlanEntry.APFVapIDData.SSID != "" {
		return wlanEntry.APFVapIDData.SSID
	}
	return wlanEntry.ProfileName
}

// determineSessionTimeout extracts session timeout from policy.
func determineSessionTimeout(policy *wlan.WlanPolicy) int {
	if policy != nil && policy.WlanTimeout != nil {
		return policy.WlanTimeout.SessionTimeout
	}
	return 0
}

func determineCentralSwitchingValue(policy *wlan.WlanPolicy) float64 {
	if policy != nil && policy.WlanSwitchingPolicy != nil {
		return boolToFloat64(policy.WlanSwitchingPolicy.CentralSwitching)
	}
	return 0
}

func determineCentralAuthenticationValue(policy *wlan.WlanPolicy) float64 {
	if policy != nil && policy.WlanSwitchingPolicy != nil {
		return boolToFloat64(policy.WlanSwitchingPolicy.CentralAuthentication)
	}
	return 0
}

func determineCentralDHCPValue(policy *wlan.WlanPolicy) float64 {
	if policy != nil && policy.WlanSwitchingPolicy != nil {
		return boolToFloat64(policy.WlanSwitchingPolicy.CentralDHCP)
	}
	return 0
}

func determineCentralAssocEnableValue(policy *wlan.WlanPolicy) float64 {
	if policy != nil && policy.WlanSwitchingPolicy != nil {
		return boolToFloat64(policy.WlanSwitchingPolicy.CentralAssocEnable)
	}
	return 0
}

func (c *WLANCollector) isAnyMetricFlagEnabled() bool {
	return IsEnabled(c.metrics.General, c.metrics.Traffic, c.metrics.Config, c.metrics.Info)
}
