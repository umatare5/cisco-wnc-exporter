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
	Inventory  bool
	Info       bool
	InfoLabels []string
	State      bool
	Traffic    bool
	Security   bool
	Networking bool
}

// WLANCollector implements prometheus.Collector for WLAN metrics using SDK-direct approach.
type WLANCollector struct {
	src       wnc.WLANSource
	clientSrc wnc.ClientSource
	metrics   WLANMetrics

	// Inventory metrics
	countDesc *prometheus.Desc

	// Info metrics
	infoDesc       *prometheus.Desc
	infoLabelNames []string // Store configured label names

	// State metrics
	enabledDesc *prometheus.Desc

	// Traffic metrics
	clientCountDesc *prometheus.Desc
	bytesRxDesc     *prometheus.Desc
	bytesTxDesc     *prometheus.Desc

	// Security module metrics
	authPskDesc         *prometheus.Desc
	authDot1xDesc       *prometheus.Desc
	authDot1xSha256Desc *prometheus.Desc
	wpa2EnabledDesc     *prometheus.Desc
	wpa3EnabledDesc     *prometheus.Desc
	sessionTimeoutDesc  *prometheus.Desc

	// Networking module metrics
	loadBalanceDesc           *prometheus.Desc
	wlan11kNeighDesc          *prometheus.Desc
	clientSteeringDesc        *prometheus.Desc
	centralSwitchingDesc      *prometheus.Desc
	centralAuthenticationDesc *prometheus.Desc
	centralDHCPDesc           *prometheus.Desc
	centralAssocEnableDesc    *prometheus.Desc
}

// NewWLANCollector creates a new WLAN collector using SDK-direct approach.
func NewWLANCollector(
	src wnc.WLANSource,
	clientSrc wnc.ClientSource,
	metrics WLANMetrics,
) *WLANCollector {
	collector := &WLANCollector{
		src:       src,
		clientSrc: clientSrc,
		metrics:   metrics,
	}

	// Define base labels - all metrics now use minimal labels
	baseWLANLabels := []string{"id"}

	// Inventory module metrics
	if metrics.Inventory {
		collector.countDesc = prometheus.NewDesc(
			"wnc_wlans_total",
			"Total number of configured WLANs",
			[]string{}, nil,
		)
	}

	// Info module - provides all additional labels for joining
	if metrics.Info {
		// Build dynamic labels based on configuration
		infoLabels := buildWLANInfoLabels(metrics.InfoLabels)
		collector.infoDesc = prometheus.NewDesc(
			"wnc_wlan_info",
			"WLAN information labels for joining with other metrics",
			infoLabels, nil,
		)
		collector.infoLabelNames = infoLabels // Store for later use
	}

	// State module metrics
	if metrics.State {
		collector.enabledDesc = prometheus.NewDesc(
			"wnc_wlan_enabled",
			"WLAN operational status (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
	}

	// Traffic module metrics
	if metrics.Traffic {
		collector.clientCountDesc = prometheus.NewDesc(
			"wnc_wlan_clients_total",
			"Number of connected clients",
			baseWLANLabels, nil,
		)
		collector.bytesRxDesc = prometheus.NewDesc(
			"wnc_wlan_rx_bytes_total",
			"WLAN received bytes",
			baseWLANLabels, nil,
		)
		collector.bytesTxDesc = prometheus.NewDesc(
			"wnc_wlan_tx_bytes_total",
			"WLAN transmitted bytes",
			baseWLANLabels, nil,
		)
	}

	// Security module metrics
	if metrics.Security {
		collector.authPskDesc = prometheus.NewDesc(
			"wnc_wlan_auth_psk_enabled",
			"PSK authentication enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.authDot1xDesc = prometheus.NewDesc(
			"wnc_wlan_auth_dot1x_enabled",
			"802.1x authentication enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.authDot1xSha256Desc = prometheus.NewDesc(
			"wnc_wlan_auth_dot1x_sha256_enabled",
			"802.1x SHA256 authentication enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.wpa2EnabledDesc = prometheus.NewDesc(
			"wnc_wlan_wpa2_enabled",
			"WPA2 support enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.wpa3EnabledDesc = prometheus.NewDesc(
			"wnc_wlan_wpa3_enabled",
			"WPA3 support enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.sessionTimeoutDesc = prometheus.NewDesc(
			"wnc_wlan_session_timeout_seconds",
			"Session timeout duration in seconds",
			baseWLANLabels, nil,
		)
	}

	// Networking module metrics
	if metrics.Networking {
		collector.loadBalanceDesc = prometheus.NewDesc(
			"wnc_wlan_load_balance_enabled",
			"Load balancing enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.wlan11kNeighDesc = prometheus.NewDesc(
			"wnc_wlan_11k_neighbor_list_enabled",
			"802.11k neighbor list enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.clientSteeringDesc = prometheus.NewDesc(
			"wnc_wlan_client_steering_enabled",
			"6GHz client steering enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.centralSwitchingDesc = prometheus.NewDesc(
			"wnc_wlan_central_switching_enabled",
			"Central switching enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.centralAuthenticationDesc = prometheus.NewDesc(
			"wnc_wlan_central_authentication_enabled",
			"Central authentication enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.centralDHCPDesc = prometheus.NewDesc(
			"wnc_wlan_central_dhcp_enabled",
			"Central DHCP enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
		collector.centralAssocEnableDesc = prometheus.NewDesc(
			"wnc_wlan_central_association_enabled",
			"Central association enabled (0=disabled, 1=enabled)",
			baseWLANLabels, nil,
		)
	}

	return collector
}

// Describe sends the descriptors of all metrics to the provided channel.
func (c *WLANCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.Inventory {
		c.describeInventoryMetrics(ch)
	}
	if c.metrics.Info {
		c.describeInfoMetrics(ch)
	}
	if c.metrics.State {
		c.describeStateMetrics(ch)
	}
	if c.metrics.Traffic {
		c.describeTrafficMetrics(ch)
	}
	if c.metrics.Security {
		c.describeSecurityMetrics(ch)
	}
	if c.metrics.Networking {
		c.describeNetworkingMetrics(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (c *WLANCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Get WLAN configuration data
	wlanConfigEntries, err := c.src.ListConfigEntries(ctx)
	if err != nil {
		slog.Warn("Failed to retrieve WLAN configuration entries", "error", err)
		return
	}
	// Get WLAN policies for security and networking modules
	var wlanPolicies *wlan.WlanCfgWlanPolicies
	if c.metrics.Security || c.metrics.Networking {
		wlanPolicies, err = c.src.ListPolicies(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve WLAN policies", "error", err)
			// Don't return here - continue with other metrics that don't need policies
		}
	}

	// Extract profile configs from response
	var profileConfigs *wlan.WlanCfgEntries
	if wlanConfigEntries != nil {
		profileConfigs = wlanConfigEntries.WlanCfgEntries
	}

	// Debug: Log profile names
	if profileConfigs != nil {
		slog.Debug("Profile configurations:", "profiles", func() []string {
			var profiles []string
			for _, entry := range profileConfigs.WlanCfgEntry {
				profiles = append(profiles, entry.ProfileName)
			}
			return profiles
		}())
	}

	// Collect metrics according to enabled metrics - 1:1 with describe methods
	if c.metrics.Inventory {
		c.collectInventoryMetrics(ch, profileConfigs)
	}
	if c.metrics.Info {
		c.collectInfoMetrics(ch, profileConfigs)
	}
	if c.metrics.State {
		c.collectStateMetrics(ch, profileConfigs)
	}
	if c.metrics.Traffic {
		c.collectTrafficMetrics(ch, profileConfigs)
	}
	if c.metrics.Security {
		c.collectSecurityMetrics(ch, profileConfigs, wlanPolicies)
	}
	if c.metrics.Networking {
		c.collectNetworkingMetrics(ch, profileConfigs, wlanPolicies)
	}
}

// describeInventoryMetrics describes inventory-related metric descriptors.
func (c *WLANCollector) describeInventoryMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.countDesc
}

// describeInfoMetrics describes info-related metric descriptors.
func (c *WLANCollector) describeInfoMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
}

// describeStateMetrics describes state-related metric descriptors.
func (c *WLANCollector) describeStateMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.enabledDesc
}

// describeTrafficMetrics describes traffic-related metric descriptors.
func (c *WLANCollector) describeTrafficMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.clientCountDesc
	ch <- c.bytesRxDesc
	ch <- c.bytesTxDesc
}

// describeSecurityMetrics describes security-related metric descriptors.
func (c *WLANCollector) describeSecurityMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.authPskDesc
	ch <- c.authDot1xDesc
	ch <- c.authDot1xSha256Desc
	ch <- c.wpa2EnabledDesc
	ch <- c.wpa3EnabledDesc
	ch <- c.sessionTimeoutDesc
}

// describeNetworkingMetrics describes networking-related metric descriptors.
func (c *WLANCollector) describeNetworkingMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.loadBalanceDesc
	ch <- c.wlan11kNeighDesc
	ch <- c.clientSteeringDesc
	ch <- c.centralSwitchingDesc
	ch <- c.centralAuthenticationDesc
	ch <- c.centralDHCPDesc
	ch <- c.centralAssocEnableDesc
}

// collectInventoryMetrics collects inventory-related metrics.
func (c *WLANCollector) collectInventoryMetrics(ch chan<- prometheus.Metric, profileConfigs *wlan.WlanCfgEntries) {
	if profileConfigs == nil {
		return
	}

	// Count configured WLANs
	wlanCount := len(profileConfigs.WlanCfgEntry)

	if c.countDesc != nil {
		ch <- prometheus.MustNewConstMetric(
			c.countDesc,
			prometheus.GaugeValue,
			float64(wlanCount),
		)
	}
}

// collectStateMetrics collects state-related metrics.
func (c *WLANCollector) collectStateMetrics(ch chan<- prometheus.Metric, profileConfigs *wlan.WlanCfgEntries) {
	if profileConfigs == nil {
		return
	}

	for _, wlanEntry := range profileConfigs.WlanCfgEntry {
		if wlanEntry.ProfileName == "" {
			continue
		}

		// Build base labels
		wlanIDStr := strconv.Itoa(wlanEntry.WlanID)
		labels := []string{wlanIDStr}

		if c.enabledDesc != nil {
			enabled := 0
			if wlanEntry.APFVapIDData != nil && wlanEntry.APFVapIDData.WlanStatus {
				enabled = 1
			}

			ch <- prometheus.MustNewConstMetric(
				c.enabledDesc,
				prometheus.GaugeValue,
				float64(enabled),
				labels...,
			)
		}
	}
}

// collectTrafficMetrics collects traffic-related metrics.
func (c *WLANCollector) collectTrafficMetrics(
	ch chan<- prometheus.Metric,
	profileConfigs *wlan.WlanCfgEntries,
) {
	if profileConfigs == nil {
		return
	}

	// Get client traffic statistics for accurate aggregation
	clientData, err := c.clientSrc.GetClientData(context.Background())
	if err != nil {
		slog.Warn("Failed to get client data for WLAN metrics", "error", err)
		return
	}

	trafficStats, err := c.clientSrc.GetTrafficStats(context.Background())
	if err != nil {
		slog.Warn("Failed to get client traffic stats for WLAN metrics", "error", err)
		return
	}

	// Aggregate client statistics by WLAN
	wlanStats := c.aggregateClientStatsByWLAN(clientData, trafficStats)

	for _, wlanEntry := range profileConfigs.WlanCfgEntry {
		if wlanEntry.ProfileName == "" {
			continue
		}

		// Build base labels
		wlanIDStr := strconv.Itoa(wlanEntry.WlanID)
		labels := []string{wlanIDStr}

		// Get aggregated statistics for this WLAN
		stats := wlanStats[wlanEntry.WlanID]

		// Client count - use aggregated count for accuracy
		if c.clientCountDesc != nil {
			ch <- prometheus.MustNewConstMetric(
				c.clientCountDesc,
				prometheus.GaugeValue,
				float64(stats.clientCount),
				labels...,
			)
		}

		// Bytes RX - aggregated from client traffic statistics
		if c.bytesRxDesc != nil {
			ch <- prometheus.MustNewConstMetric(
				c.bytesRxDesc,
				prometheus.CounterValue,
				float64(stats.bytesRx),
				labels...,
			)
		}

		// Bytes TX - aggregated from client traffic statistics
		if c.bytesTxDesc != nil {
			ch <- prometheus.MustNewConstMetric(
				c.bytesTxDesc,
				prometheus.CounterValue,
				float64(stats.bytesTx),
				labels...,
			)
		}
	}
}

// Internal helper structures for WLAN data aggregation.

// wlanAggregatedStats represents aggregated statistics for a WLAN.
// This internal structure is used by aggregateClientStatsByWLAN to collect
// client statistics grouped by WLAN ID.
type wlanAggregatedStats struct {
	clientCount int
	bytesRx     uint64
	bytesTx     uint64
	packetsRx   uint64
	packetsTx   uint64
}

// aggregateClientStatsByWLAN aggregates client traffic statistics by WLAN ID.
func (c *WLANCollector) aggregateClientStatsByWLAN(
	clientData *client.ClientOperCommonOperData,
	trafficStats *client.ClientOperTrafficStats,
) map[int]wlanAggregatedStats {
	wlanStats := make(map[int]wlanAggregatedStats)

	if clientData == nil || trafficStats == nil {
		return wlanStats
	}

	// Build traffic stats lookup map
	trafficMap := make(map[string]client.TrafficStats)
	for _, traffic := range trafficStats.TrafficStats {
		if traffic.MsMACAddress != "" {
			trafficMap[traffic.MsMACAddress] = traffic
		}
	}

	// Aggregate by WLAN ID
	for _, commonData := range clientData.CommonOperData {
		// Only count RUN state clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Get WLAN ID
		wlanID := commonData.WlanID
		if wlanID == 0 {
			continue
		}

		// Initialize stats if needed
		stats := wlanStats[wlanID]
		stats.clientCount++

		// Add traffic statistics if available
		if traffic, ok := trafficMap[commonData.ClientMAC]; ok {
			stats.bytesRx += parseUint64Safely(traffic.BytesRx)
			stats.bytesTx += parseUint64Safely(traffic.BytesTx)
			stats.packetsRx += parseUint64Safely(traffic.PktsRx)
			stats.packetsTx += parseUint64Safely(traffic.PktsTx)
		}

		wlanStats[wlanID] = stats
	}

	return wlanStats
}

// collectInfoMetrics collects WLAN info metrics from WNC data.
func (c *WLANCollector) collectInfoMetrics(ch chan<- prometheus.Metric, profileConfigs *wlan.WlanCfgEntries) {
	if c.infoDesc == nil || profileConfigs == nil {
		return
	}

	for _, wlanEntry := range profileConfigs.WlanCfgEntry {
		wlanIDStr := strconv.Itoa(wlanEntry.WlanID)

		// Get WLAN name from SSID or profile name
		wlanName := wlanEntry.APFVapIDData.SSID
		if wlanName == "" {
			wlanName = wlanEntry.ProfileName
		}

		// Build dynamic labels based on configured labels
		labels := c.buildWLANInfoLabelValues(wlanIDStr, wlanName)

		// Emit info metric (always value 1)
		ch <- prometheus.MustNewConstMetric(
			c.infoDesc,
			prometheus.GaugeValue,
			1,
			labels...,
		)
	}
}

// buildWLANInfoLabelValues constructs the label values array based on configured labels.
func (c *WLANCollector) buildWLANInfoLabelValues(id, name string) []string {
	labelNames := c.infoLabelNames
	values := make([]string, len(labelNames))

	// Map field names to values
	valueMap := map[string]string{
		"id":   id,
		"name": name,
	}

	// Build values array in the same order as label names
	for i, labelName := range labelNames {
		if value, exists := valueMap[labelName]; exists {
			values[i] = value
		} else {
			values[i] = "" // fallback for unknown labels
		}
	}

	return values
}

// buildWLANInfoLabels constructs the WLAN info labels slice based on configuration.
func buildWLANInfoLabels(configuredLabels []string) []string {
	// id is always required as first label
	labels := []string{"id"}

	// Add other configured labels in consistent order
	labelOrder := []string{"name"}
	for _, label := range labelOrder {
		if contains(configuredLabels, label) && !contains(labels, label) {
			labels = append(labels, label)
		}
	}

	return labels
}

// collectSecurityMetrics collects security-related metrics from WLAN configuration entries and policies.
func (c *WLANCollector) collectSecurityMetrics(
	ch chan<- prometheus.Metric,
	profileConfigs *wlan.WlanCfgEntries,
	wlanPolicies *wlan.WlanCfgWlanPolicies,
) {
	if profileConfigs == nil {
		return
	}

	// Build policy profile lookup map for timeout information
	policyMap := make(map[string]*wlan.WlanPolicy)
	if wlanPolicies != nil && wlanPolicies.CiscoIOSXEWirelessWlanCfgData.WlanPolicies != nil {
		for i := range wlanPolicies.CiscoIOSXEWirelessWlanCfgData.WlanPolicies.WlanPolicy {
			policy := &wlanPolicies.CiscoIOSXEWirelessWlanCfgData.WlanPolicies.WlanPolicy[i]
			policyMap[policy.PolicyProfileName] = policy
		}
	}

	for _, wlanEntry := range profileConfigs.WlanCfgEntry {
		if wlanEntry.ProfileName == "" {
			continue
		}

		// Build base labels
		wlanIDStr := strconv.Itoa(wlanEntry.WlanID)
		labels := []string{wlanIDStr}

		// Authentication metrics
		if c.authPskDesc != nil {
			pskEnabled := 0
			if wlanEntry.AuthKeyMgmtPsk {
				pskEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.authPskDesc,
				prometheus.GaugeValue,
				float64(pskEnabled),
				labels...,
			)
		}

		if c.authDot1xDesc != nil {
			dot1xEnabled := 0
			if wlanEntry.AuthKeyMgmtDot1x {
				dot1xEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.authDot1xDesc,
				prometheus.GaugeValue,
				float64(dot1xEnabled),
				labels...,
			)
		}

		if c.authDot1xSha256Desc != nil {
			dot1xSha256Enabled := 0
			if wlanEntry.AuthKeyMgmtDot1xSha256 {
				dot1xSha256Enabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.authDot1xSha256Desc,
				prometheus.GaugeValue,
				float64(dot1xSha256Enabled),
				labels...,
			)
		}

		if c.wpa2EnabledDesc != nil {
			wpa2Enabled := 0
			if wlanEntry.WPA2Enabled {
				wpa2Enabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.wpa2EnabledDesc,
				prometheus.GaugeValue,
				float64(wpa2Enabled),
				labels...,
			)
		}

		if c.wpa3EnabledDesc != nil {
			wpa3Enabled := 0
			if wlanEntry.WPA3Enabled {
				wpa3Enabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.wpa3EnabledDesc,
				prometheus.GaugeValue,
				float64(wpa3Enabled),
				labels...,
			)
		}

		// Session timeout from policies
		if c.sessionTimeoutDesc != nil {
			sessionTimeout := 0
			if policy, exists := policyMap[wlanEntry.ProfileName]; exists && policy.WlanTimeout != nil {
				sessionTimeout = policy.WlanTimeout.SessionTimeout
			}
			ch <- prometheus.MustNewConstMetric(
				c.sessionTimeoutDesc,
				prometheus.GaugeValue,
				float64(sessionTimeout),
				labels...,
			)
		}
	}
}

// collectNetworkingMetrics collects networking-related metrics from WLAN configuration entries and policies.
func (c *WLANCollector) collectNetworkingMetrics(
	ch chan<- prometheus.Metric,
	profileConfigs *wlan.WlanCfgEntries,
	wlanPolicies *wlan.WlanCfgWlanPolicies,
) {
	if profileConfigs == nil {
		return
	}

	// Build policy profile lookup map for switching information
	policyMap := make(map[string]*wlan.WlanPolicy)
	if wlanPolicies != nil && wlanPolicies.CiscoIOSXEWirelessWlanCfgData.WlanPolicies != nil {
		for i := range wlanPolicies.CiscoIOSXEWirelessWlanCfgData.WlanPolicies.WlanPolicy {
			policy := &wlanPolicies.CiscoIOSXEWirelessWlanCfgData.WlanPolicies.WlanPolicy[i]
			policyMap[policy.PolicyProfileName] = policy
		}
	}

	for _, wlanEntry := range profileConfigs.WlanCfgEntry {
		if wlanEntry.ProfileName == "" {
			continue
		}

		// Build base labels
		wlanIDStr := strconv.Itoa(wlanEntry.WlanID)
		labels := []string{wlanIDStr}

		// Configuration metrics from WLAN entries
		if c.loadBalanceDesc != nil {
			loadBalanceEnabled := 0
			if wlanEntry.LoadBalance {
				loadBalanceEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.loadBalanceDesc,
				prometheus.GaugeValue,
				float64(loadBalanceEnabled),
				labels...,
			)
		}

		if c.wlan11kNeighDesc != nil {
			wlan11kNeighEnabled := 0
			if wlanEntry.Wlan11kNeighList {
				wlan11kNeighEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.wlan11kNeighDesc,
				prometheus.GaugeValue,
				float64(wlan11kNeighEnabled),
				labels...,
			)
		}

		if c.clientSteeringDesc != nil {
			clientSteeringEnabled := 0
			if wlanEntry.ClientSteering {
				clientSteeringEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.clientSteeringDesc,
				prometheus.GaugeValue,
				float64(clientSteeringEnabled),
				labels...,
			)
		}

		// Switching metrics from policies
		policy, exists := policyMap[wlanEntry.ProfileName]
		if !exists {
			// If no matching policy found, emit 0 values
			policy = &wlan.WlanPolicy{}
		}

		if c.centralSwitchingDesc != nil {
			centralSwitchingEnabled := 0
			if policy.WlanSwitchingPolicy != nil && policy.WlanSwitchingPolicy.CentralSwitching {
				centralSwitchingEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.centralSwitchingDesc,
				prometheus.GaugeValue,
				float64(centralSwitchingEnabled),
				labels...,
			)
		}

		if c.centralAuthenticationDesc != nil {
			centralAuthenticationEnabled := 0
			if policy.WlanSwitchingPolicy != nil && policy.WlanSwitchingPolicy.CentralAuthentication {
				centralAuthenticationEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.centralAuthenticationDesc,
				prometheus.GaugeValue,
				float64(centralAuthenticationEnabled),
				labels...,
			)
		}

		if c.centralDHCPDesc != nil {
			centralDHCPEnabled := 0
			if policy.WlanSwitchingPolicy != nil && policy.WlanSwitchingPolicy.CentralDHCP {
				centralDHCPEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.centralDHCPDesc,
				prometheus.GaugeValue,
				float64(centralDHCPEnabled),
				labels...,
			)
		}

		if c.centralAssocEnableDesc != nil {
			centralAssocEnabled := 0
			if policy.WlanSwitchingPolicy != nil && policy.WlanSwitchingPolicy.CentralAssocEnable {
				centralAssocEnabled = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.centralAssocEnableDesc,
				prometheus.GaugeValue,
				float64(centralAssocEnabled),
				labels...,
			)
		}
	}
}
