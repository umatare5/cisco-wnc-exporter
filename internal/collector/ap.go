// Package collector provides collectors for cisco-wnc-exporter.
package collector

import (
	"context"
	"log/slog"
	"slices"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// APCollector implements prometheus.Collector for AP metrics from WNC.
type APCollector struct {
	// Inventory metrics
	countDesc *prometheus.Desc

	// Info metrics
	infoDesc       *prometheus.Desc
	infoLabelNames []string // Store configured label names

	// Radio-level metrics
	channelUtilizationDesc *prometheus.Desc
	rxUtilizationDesc      *prometheus.Desc
	txUtilizationDesc      *prometheus.Desc
	noiseUtilizationDesc   *prometheus.Desc
	noiseFloorDesc         *prometheus.Desc
	txPowerDesc            *prometheus.Desc
	channelDesc            *prometheus.Desc
	channelWidthDesc       *prometheus.Desc
	associatedClientsDesc  *prometheus.Desc
	radioStateDesc         *prometheus.Desc
	adminStateDesc         *prometheus.Desc
	operStateDesc          *prometheus.Desc

	// Physical layer metrics
	txPowerLevelDesc *prometheus.Desc
	txPowerMaxDesc   *prometheus.Desc

	// Traffic metrics
	rxPacketsTotalDesc *prometheus.Desc
	txPacketsTotalDesc *prometheus.Desc
	rxBytesTotalDesc   *prometheus.Desc
	txBytesTotalDesc   *prometheus.Desc

	// Frame counters
	dataRxFramesTotalDesc       *prometheus.Desc
	dataTxFramesTotalDesc       *prometheus.Desc
	managementRxFramesTotalDesc *prometheus.Desc
	managementTxFramesTotalDesc *prometheus.Desc
	controlRxFramesTotalDesc    *prometheus.Desc
	controlTxFramesTotalDesc    *prometheus.Desc
	multicastRxFramesTotalDesc  *prometheus.Desc
	multicastTxFramesTotalDesc  *prometheus.Desc
	totalTxFramesTotalDesc      *prometheus.Desc
	rtsSuccessTotalDesc         *prometheus.Desc

	// Error metrics
	rxErrorsTotalDesc         *prometheus.Desc
	txErrorsTotalDesc         *prometheus.Desc
	txDropsTotalDesc          *prometheus.Desc
	txRetriesTotalDesc        *prometheus.Desc
	ackFailuresTotalDesc      *prometheus.Desc
	duplicateFramesTotalDesc  *prometheus.Desc
	fcsErrorsTotalDesc        *prometheus.Desc
	fragmentationRxTotalDesc  *prometheus.Desc
	fragmentationTxTotalDesc  *prometheus.Desc
	rtsFailuresTotalDesc      *prometheus.Desc
	decryptionErrorsTotalDesc *prometheus.Desc
	micErrorsTotalDesc        *prometheus.Desc
	wepUndecryptableTotalDesc *prometheus.Desc
	coverageHoleEventsDesc    *prometheus.Desc
	interferenceEventsDesc    *prometheus.Desc
	dfsChannelChangesDesc     *prometheus.Desc
	radarDetectedEventsDesc   *prometheus.Desc
	radioResetTotalDesc       *prometheus.Desc

	// AP-level system metrics
	cpuUsageCurrentDesc    *prometheus.Desc
	cpuUsageAverageDesc    *prometheus.Desc
	memoryUsageCurrentDesc *prometheus.Desc
	memoryUsageAverageDesc *prometheus.Desc
	uptimeSecondsDesc      *prometheus.Desc

	// Module configuration
	metrics   APMetrics
	src       wnc.APSource
	rrmSrc    wnc.RRMSource
	clientSrc wnc.ClientSource
}

// APMetrics represents which AP metrics are enabled.
type APMetrics struct {
	Inventory  bool
	Info       bool
	InfoLabels []string
	State      bool
	Phy        bool
	RF         bool
	Traffic    bool
	Errors     bool
	CPU        bool
	Memory     bool
}

// NewAPCollector creates a new AP collector with SDK-direct data access.
func NewAPCollector(
	src wnc.APSource, rrmSrc wnc.RRMSource, clientSrc wnc.ClientSource, metrics APMetrics,
) *APCollector {
	// Define base labels - all metrics now use minimal labels
	baseRadioLabels := []string{"mac", "radio"}
	baseAPLabels := []string{"mac"}

	collector := &APCollector{
		metrics:   metrics,
		src:       src,
		rrmSrc:    rrmSrc,
		clientSrc: clientSrc,
	}

	// Initialize metrics based on enabled metrics
	if metrics.Inventory {
		collector.countDesc = prometheus.NewDesc(
			"wnc_ap_count",
			"Number of access points",
			nil,
			nil,
		)
	}

	// Info module - provides all additional labels for joining including device info
	if metrics.Info {
		// Build dynamic labels based on configuration
		infoLabels := buildAPInfoLabels(metrics.InfoLabels)
		collector.infoDesc = prometheus.NewDesc(
			"wnc_ap_info",
			"AP information labels for joining with other metrics, including device details",
			infoLabels, nil,
		)
		collector.infoLabelNames = infoLabels // Store for later use
	}

	if metrics.State {
		collector.radioStateDesc = prometheus.NewDesc(
			"wnc_ap_radio_state",
			"Radio state (0=down, 1=up)",
			baseRadioLabels,
			nil,
		)
		collector.adminStateDesc = prometheus.NewDesc(
			"wnc_ap_admin_state",
			"Admin state (0=disabled, 1=enabled)",
			baseRadioLabels,
			nil,
		)
		collector.operStateDesc = prometheus.NewDesc(
			"wnc_ap_oper_state",
			"Operational state (0=down, 1=up)",
			baseRadioLabels,
			nil,
		)
		collector.uptimeSecondsDesc = prometheus.NewDesc(
			"wnc_ap_uptime_seconds",
			"AP uptime in seconds",
			baseAPLabels,
			nil,
		)
	}

	if metrics.Phy {
		collector.channelDesc = prometheus.NewDesc(
			"wnc_ap_channel_number",
			"Operating channel number",
			baseRadioLabels,
			nil,
		)
		collector.channelWidthDesc = prometheus.NewDesc(
			"wnc_ap_channel_width_mhz",
			"Channel bandwidth (MHz)",
			baseRadioLabels,
			nil,
		)
		collector.txPowerDesc = prometheus.NewDesc(
			"wnc_ap_tx_power_dbm",
			"Current transmit power (dBm)",
			baseRadioLabels,
			nil,
		)
		collector.txPowerLevelDesc = prometheus.NewDesc(
			"wnc_ap_tx_power_level",
			"TX power level setting (0-8)",
			baseRadioLabels,
			nil,
		)
		collector.txPowerMaxDesc = prometheus.NewDesc(
			"wnc_ap_tx_power_max_dbm",
			"Maximum TX power capability (dBm)",
			baseRadioLabels,
			nil,
		)
	}

	if metrics.RF {
		collector.channelUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_channel_utilization_percent",
			"Channel utilization percentage (CCA-based)",
			baseRadioLabels,
			nil,
		)
		collector.rxUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_rx_utilization_percent",
			"RX utilization percentage",
			baseRadioLabels,
			nil,
		)
		collector.txUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_tx_utilization_percent",
			"TX utilization percentage",
			baseRadioLabels,
			nil,
		)
		collector.noiseUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_noise_utilization_percent",
			"Noise channel utilization percentage",
			baseRadioLabels,
			nil,
		)
		collector.noiseFloorDesc = prometheus.NewDesc(
			"wnc_ap_noise_floor_dbm",
			"Channel noise floor (dBm, from RRM measurement data as radio-oper-stats noise-floor is not populated in IOS-XE 17.12.5)",
			baseRadioLabels,
			nil,
		)
		collector.associatedClientsDesc = prometheus.NewDesc(
			"wnc_ap_clients_total",
			"Number of associated clients",
			baseRadioLabels,
			nil,
		)
	}

	if metrics.CPU {
		collector.cpuUsageCurrentDesc = prometheus.NewDesc(
			"wnc_ap_cpu_usage_current_percent",
			"Current CPU utilization percentage",
			baseAPLabels,
			nil,
		)
		collector.cpuUsageAverageDesc = prometheus.NewDesc(
			"wnc_ap_cpu_usage_average_percent",
			"Average CPU utilization percentage",
			baseAPLabels,
			nil,
		)
	}

	if metrics.Memory {
		collector.memoryUsageCurrentDesc = prometheus.NewDesc(
			"wnc_ap_memory_usage_current_percent",
			"Current memory utilization percentage",
			baseAPLabels,
			nil,
		)
		collector.memoryUsageAverageDesc = prometheus.NewDesc(
			"wnc_ap_memory_usage_average_percent",
			"Average memory utilization percentage",
			baseAPLabels,
			nil,
		)
	}

	if metrics.Traffic {
		collector.rxPacketsTotalDesc = prometheus.NewDesc(
			"wnc_ap_rx_packets_total",
			"Total received packets (rx-data-frame-count, as rx-data-pkt-count is not populated in IOS-XE 17.12.5)",
			baseRadioLabels,
			nil,
		)
		collector.txPacketsTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_packets_total",
			"Total transmitted packets (tx-data-frame-count, as tx-data-pkt-count is not populated in IOS-XE 17.12.5)",
			baseRadioLabels,
			nil,
		)
		collector.rxBytesTotalDesc = prometheus.NewDesc(
			"wnc_ap_rx_bytes_total",
			"Total received bytes (calculated)",
			baseRadioLabels,
			nil,
		)
		collector.txBytesTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_bytes_total",
			"Total transmitted bytes (calculated)",
			baseRadioLabels,
			nil,
		)
		collector.dataRxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_data_rx_frames_total",
			"Data RX frames",
			baseRadioLabels,
			nil,
		)
		collector.dataTxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_data_tx_frames_total",
			"Data TX frames",
			baseRadioLabels,
			nil,
		)
		collector.managementRxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_management_rx_frames_total",
			"Management RX frames",
			baseRadioLabels,
			nil,
		)
		collector.managementTxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_management_tx_frames_total",
			"Management TX frames",
			baseRadioLabels,
			nil,
		)
		collector.controlRxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_control_rx_frames_total",
			"Control RX frames",
			baseRadioLabels,
			nil,
		)
		collector.controlTxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_control_tx_frames_total",
			"Control TX frames",
			baseRadioLabels,
			nil,
		)
		collector.multicastRxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_multicast_rx_frames_total",
			"Multicast RX frames",
			baseRadioLabels,
			nil,
		)
		collector.multicastTxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_multicast_tx_frames_total",
			"Multicast TX frames",
			baseRadioLabels,
			nil,
		)
		collector.totalTxFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_total_tx_frames_total",
			"Total TX frames",
			baseRadioLabels,
			nil,
		)
		collector.rtsSuccessTotalDesc = prometheus.NewDesc(
			"wnc_ap_rts_success_total",
			"Successful RTS transmissions",
			baseRadioLabels,
			nil,
		)
	}

	if metrics.Errors {
		collector.rxErrorsTotalDesc = prometheus.NewDesc(
			"wnc_ap_rx_errors_total",
			"Total RX errors (rx-error-frame-count)",
			baseRadioLabels,
			nil,
		)
		collector.txErrorsTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_errors_total",
			"Total TX errors (failed-count)",
			baseRadioLabels,
			nil,
		)
		collector.txDropsTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_drops_total",
			"Total TX drops (calculated from failures)",
			baseRadioLabels,
			nil,
		)
		collector.txRetriesTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_retries_total",
			"Total TX retries (retry-count + multiple-retry)",
			baseRadioLabels,
			nil,
		)
		collector.ackFailuresTotalDesc = prometheus.NewDesc(
			"wnc_ap_transmission_failures_total",
			"Failed transmission attempts (follow CSCwn96363)",
			baseRadioLabels,
			nil,
		)
		collector.duplicateFramesTotalDesc = prometheus.NewDesc(
			"wnc_ap_duplicate_frames_total",
			"Duplicate frames received",
			baseRadioLabels,
			nil,
		)
		collector.fcsErrorsTotalDesc = prometheus.NewDesc(
			"wnc_ap_fcs_errors_total",
			"Frame Check Sequence errors",
			baseRadioLabels,
			nil,
		)
		collector.fragmentationRxTotalDesc = prometheus.NewDesc(
			"wnc_ap_fragmentation_rx_total",
			"RX fragmented packets",
			baseRadioLabels,
			nil,
		)
		collector.fragmentationTxTotalDesc = prometheus.NewDesc(
			"wnc_ap_fragmentation_tx_total",
			"TX fragmented packets",
			baseRadioLabels,
			nil,
		)
		collector.rtsFailuresTotalDesc = prometheus.NewDesc(
			"wnc_ap_rts_failures_total",
			"RTS failures",
			baseRadioLabels,
			nil,
		)
		collector.decryptionErrorsTotalDesc = prometheus.NewDesc(
			"wnc_ap_decryption_errors_total",
			"Decryption errors",
			baseRadioLabels,
			nil,
		)
		collector.micErrorsTotalDesc = prometheus.NewDesc(
			"wnc_ap_mic_errors_total",
			"MIC errors",
			baseRadioLabels,
			nil,
		)
		collector.wepUndecryptableTotalDesc = prometheus.NewDesc(
			"wnc_ap_wep_undecryptable_total",
			"WEP undecryptable frames",
			baseRadioLabels,
			nil,
		)
		collector.coverageHoleEventsDesc = prometheus.NewDesc(
			"wnc_ap_coverage_hole_events_total",
			"Coverage hole events detected",
			baseRadioLabels,
			nil,
		)
		collector.interferenceEventsDesc = prometheus.NewDesc(
			"wnc_ap_interference_events_total",
			"Interference events detected",
			baseRadioLabels,
			nil,
		)
		collector.dfsChannelChangesDesc = prometheus.NewDesc(
			"wnc_ap_dfs_channel_changes_total",
			"DFS channel change events",
			baseRadioLabels,
			nil,
		)
		collector.radarDetectedEventsDesc = prometheus.NewDesc(
			"wnc_ap_radar_detected_events_total",
			"Radar detection events",
			baseRadioLabels,
			nil,
		)
		collector.radioResetTotalDesc = prometheus.NewDesc(
			"wnc_ap_radio_reset_total",
			"Total radio reset count since statistics clear",
			baseRadioLabels,
			nil,
		)
	}
	return collector
}

// Internal helper structures for AP data aggregation.

// rfMetricContext contains the common data needed for RF metrics collection.
type rfMetricContext struct {
	rrmData      *rrm.RRMOperRRMMeasurement
	clientCounts map[string]map[int]int
	// Pre-calculated measurement data per radio for efficiency
	measurements map[string]*rrm.RRMMeasurement
}

// buildAPInfoLabels constructs the AP info labels slice based on configuration.
func buildAPInfoLabels(configuredLabels []string) []string {
	// mac and radio are always required as first labels
	labels := []string{"mac", "radio"}

	// Add other configured labels in consistent order
	labelOrder := []string{"name", "ip", "band", "model", "serial", "sw_version", "eth_mac"}
	for _, label := range labelOrder {
		if slices.Contains(configuredLabels, label) && !slices.Contains(labels, label) {
			labels = append(labels, label)
		}
	}

	return labels
}

// Describe implements prometheus.Collector.
func (c *APCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.Inventory {
		c.describeInventoryMetrics(ch)
	}
	if c.metrics.Info {
		c.describeInfoMetrics(ch)
	}
	if c.metrics.State {
		c.describeStateMetrics(ch)
	}
	if c.metrics.RF {
		c.describeRFMetrics(ch)
	}
	if c.metrics.Phy {
		c.describePhysicalLayerMetrics(ch)
	}
	if c.metrics.Traffic {
		c.describeTrafficMetrics(ch)
	}
	if c.metrics.Errors {
		c.describeErrorMetrics(ch)
	}
	if c.metrics.CPU {
		c.describeCPUMetrics(ch)
	}
	if c.metrics.Memory {
		c.describeMemoryMetrics(ch)
	}
}

// Collect implements prometheus.Collector by retrieving AP data from WNC.
func (c *APCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	capwapData, err := c.src.GetCAPWAPData(ctx)
	if err != nil {
		slog.Error("Failed to get CAPWAP data", "error", err)
		return
	}

	radioData, err := c.src.GetRadioData(ctx)
	if err != nil {
		slog.Error("Failed to get radio data", "error", err)
		return
	}

	// Get radio operational statistics for traffic and error metrics
	var radioOperStats *ap.ApOperRadioOperStats
	if c.metrics.Traffic || c.metrics.Errors {
		radioOperStats, err = c.src.GetRadioOperStats(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve radio operational stats", "error", err)
		}
	}

	// Optional data (continue if these fail)
	var apOperData *ap.ApOperData
	if c.metrics.CPU || c.metrics.Memory {
		apOperData, err = c.src.GetAPOperData(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve AP operational data", "error", err)
		}
	}

	// Collect metrics according to enabled metrics - 1:1 with describe methods
	if c.metrics.Inventory {
		c.collectInventoryMetrics(ch, capwapData)
	}
	if c.metrics.Info {
		c.collectInfoMetrics(ch, capwapData, radioData)
	}
	// All radio-level metrics are handled by unified collection
	if c.metrics.State || c.metrics.RF || c.metrics.Phy || c.metrics.Traffic || c.metrics.Errors {
		// Get RF context if needed
		var rfCtx *rfMetricContext
		if c.metrics.RF {
			rfCtx = c.buildRFMetricContext()
		}

		c.collectRadioMetrics(ch, radioData, rfCtx, radioOperStats)
	}
	if c.metrics.CPU {
		c.collectCPUMetrics(ch, apOperData)
	}
	if c.metrics.Memory {
		c.collectMemoryMetrics(ch, apOperData)
	}
}

// describeInventoryMetrics describes inventory-related metric descriptors.
func (c *APCollector) describeInventoryMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.countDesc
}

// describeInfoMetrics describes info-related metric descriptors.
func (c *APCollector) describeInfoMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
}

// describeStateMetrics describes state-related metric descriptors.
func (c *APCollector) describeStateMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.radioStateDesc
	ch <- c.adminStateDesc
	ch <- c.operStateDesc
	ch <- c.uptimeSecondsDesc
}

// describeRFMetrics describes RF environment metric descriptors.
func (c *APCollector) describeRFMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.channelUtilizationDesc
	ch <- c.rxUtilizationDesc
	ch <- c.txUtilizationDesc
	ch <- c.noiseUtilizationDesc
	ch <- c.noiseFloorDesc
	ch <- c.associatedClientsDesc
}

// describePhysicalLayerMetrics describes physical layer metric descriptors.
func (c *APCollector) describePhysicalLayerMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.txPowerDesc
	ch <- c.channelDesc
	ch <- c.channelWidthDesc
	ch <- c.txPowerLevelDesc
	ch <- c.txPowerMaxDesc
}

// describeTrafficMetrics describes traffic-related metric descriptors including frame counters.
func (c *APCollector) describeTrafficMetrics(ch chan<- *prometheus.Desc) {
	// Basic traffic metrics
	ch <- c.rxPacketsTotalDesc
	ch <- c.txPacketsTotalDesc
	ch <- c.rxBytesTotalDesc
	ch <- c.txBytesTotalDesc

	// Frame counter metrics
	ch <- c.dataRxFramesTotalDesc
	ch <- c.dataTxFramesTotalDesc
	ch <- c.managementRxFramesTotalDesc
	ch <- c.managementTxFramesTotalDesc
	ch <- c.controlRxFramesTotalDesc
	ch <- c.controlTxFramesTotalDesc
	ch <- c.multicastRxFramesTotalDesc
	ch <- c.multicastTxFramesTotalDesc
	ch <- c.totalTxFramesTotalDesc
	ch <- c.rtsSuccessTotalDesc
}

// describeErrorMetrics describes error-related metric descriptors.
func (c *APCollector) describeErrorMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.rxErrorsTotalDesc
	ch <- c.txErrorsTotalDesc
	ch <- c.txDropsTotalDesc
	ch <- c.txRetriesTotalDesc
	ch <- c.ackFailuresTotalDesc
	ch <- c.duplicateFramesTotalDesc
	ch <- c.fcsErrorsTotalDesc
	ch <- c.fragmentationRxTotalDesc
	ch <- c.fragmentationTxTotalDesc
	ch <- c.rtsFailuresTotalDesc
	ch <- c.decryptionErrorsTotalDesc
	ch <- c.micErrorsTotalDesc
	ch <- c.wepUndecryptableTotalDesc
	ch <- c.coverageHoleEventsDesc
	ch <- c.interferenceEventsDesc
	ch <- c.dfsChannelChangesDesc
	ch <- c.radarDetectedEventsDesc
	ch <- c.radioResetTotalDesc
}

// describeCPUMetrics describes CPU-related metric descriptors.
func (c *APCollector) describeCPUMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.cpuUsageCurrentDesc
	ch <- c.cpuUsageAverageDesc
}

// describeMemoryMetrics describes memory-related metric descriptors.
func (c *APCollector) describeMemoryMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.memoryUsageCurrentDesc
	ch <- c.memoryUsageAverageDesc
}

// collectInventoryMetrics collects inventory-related metrics.
func (c *APCollector) collectInventoryMetrics(ch chan<- prometheus.Metric, capwapData *ap.ApOperCAPWAPData) {
	// AP count metric
	if c.countDesc != nil {
		count := len(capwapData.CAPWAPData)
		ch <- prometheus.MustNewConstMetric(c.countDesc, prometheus.GaugeValue, float64(count))
	}
}

// collectCPUMetrics collects CPU-related metrics.
func (c *APCollector) collectCPUMetrics(ch chan<- prometheus.Metric, apOperData *ap.ApOperData) {
	// Pre-allocate labels slice for reuse
	labels := make([]string, 1)

	for _, apOper := range apOperData.OperData {
		labels[0] = apOper.WtpMAC

		if c.cpuUsageCurrentDesc != nil {
			ch <- prometheus.MustNewConstMetric(c.cpuUsageCurrentDesc, prometheus.GaugeValue,
				float64(apOper.ApSysStats.CPUUsage), labels...)
		}
		if c.cpuUsageAverageDesc != nil {
			ch <- prometheus.MustNewConstMetric(c.cpuUsageAverageDesc, prometheus.GaugeValue,
				float64(apOper.ApSysStats.AvgCPUUsage), labels...)
		}
	}
}

// collectMemoryMetrics collects memory-related metrics.
func (c *APCollector) collectMemoryMetrics(ch chan<- prometheus.Metric, apOperData *ap.ApOperData) {
	// Pre-allocate labels slice for reuse
	labels := make([]string, 1)

	for _, apOper := range apOperData.OperData {
		labels[0] = apOper.WtpMAC

		if c.memoryUsageCurrentDesc != nil {
			ch <- prometheus.MustNewConstMetric(c.memoryUsageCurrentDesc, prometheus.GaugeValue,
				float64(apOper.ApSysStats.MemoryUsage), labels...)
		}
		if c.memoryUsageAverageDesc != nil {
			ch <- prometheus.MustNewConstMetric(c.memoryUsageAverageDesc, prometheus.GaugeValue,
				float64(apOper.ApSysStats.AvgMemoryUsage), labels...)
		}
	}
}

// collectInfoMetrics collects AP info metrics from WNC data with comprehensive device information.
func (c *APCollector) collectInfoMetrics(
	ch chan<- prometheus.Metric,
	capwapData *ap.ApOperCAPWAPData,
	radioData *ap.ApOperRadioOperData,
) {
	if c.infoDesc == nil {
		return
	}

	// Create a lookup map for CAPWAP data
	capwapMap := make(map[string]ap.CAPWAPData)
	for _, capwap := range capwapData.CAPWAPData {
		capwapMap[capwap.WtpMAC] = capwap
	}

	// Process each radio
	for _, radio := range radioData.RadioOperData {
		// Get corresponding CAPWAP data
		capwap, exists := capwapMap[radio.WtpMAC]
		if !exists {
			continue
		}

		// Determine band
		band := DetermineBandFromRadioInfo(radio.RadioSlotID, radio.RadioType)
		radioSlot := strconv.Itoa(radio.RadioSlotID)

		// Extract device information from CAPWAP data
		model := capwap.DeviceDetail.StaticInfo.ApModels.Model
		serial := capwap.DeviceDetail.StaticInfo.BoardData.WtpSerialNum
		swVersion := capwap.DeviceDetail.WtpVersion.SwVersion
		ethMAC := capwap.DeviceDetail.StaticInfo.BoardData.WtpEnetMAC

		// Build dynamic labels based on configured InfoLabels directly
		values := make([]string, len(c.infoLabelNames))
		for i, labelName := range c.infoLabelNames {
			switch labelName {
			case "mac":
				values[i] = radio.WtpMAC
			case "name":
				values[i] = capwap.Name
			case "ip":
				values[i] = capwap.IPAddr
			case "radio":
				values[i] = radioSlot
			case "band":
				values[i] = band
			case "model":
				values[i] = model
			case "serial":
				values[i] = serial
			case "sw_version":
				values[i] = swVersion
			case "eth_mac":
				values[i] = ethMAC
			default:
				values[i] = "" // fallback
			}
		}

		// Emit info metric (always value 1)
		ch <- prometheus.MustNewConstMetric(
			c.infoDesc,
			prometheus.GaugeValue,
			1,
			values...,
		)
	}
}

// collectRadioMetrics collects all radio-level metrics in a single efficient loop.
func (c *APCollector) collectRadioMetrics(
	ch chan<- prometheus.Metric,
	radioData *ap.ApOperRadioOperData,
	rfCtx *rfMetricContext,
	radioOperStats *ap.ApOperRadioOperStats,
) {
	// Pre-allocate labels slice for reuse (mac, radio)
	const radioLabelCount = 2
	labels := make([]string, radioLabelCount)

	for _, radio := range radioData.RadioOperData {
		// Build common labels once per radio
		radioSlot := strconv.Itoa(radio.RadioSlotID)
		labels[0] = radio.WtpMAC
		labels[1] = radioSlot

		// Find radio stats once per radio if needed for traffic/error metrics
		var radioStats *ap.RadioOperStats
		if c.metrics.Traffic || c.metrics.Errors {
			if radioOperStats != nil {
				radioStats = c.findRadioStats(&radio, radioOperStats)
			}
		}

		// Collect State metrics
		if c.metrics.State {
			c.collectStateMetrics(ch, &radio, labels)
		}

		// Collect RF metrics
		if c.metrics.RF {
			if rfCtx != nil {
				c.collectRFMetrics(ch, &radio, labels, rfCtx)
			}
		}

		// Collect Physical layer metrics
		if c.metrics.Phy {
			c.collectPhyMetrics(ch, &radio, labels)
		}

		// Collect Traffic metrics
		if c.metrics.Traffic {
			// Basic traffic counters
			ch <- prometheus.MustNewConstMetric(c.rxPacketsTotalDesc, prometheus.CounterValue, float64(radioStats.RxDataFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.txPacketsTotalDesc, prometheus.CounterValue, float64(radioStats.TxDataFrameCount), labels...)

			// Frame counters
			ch <- prometheus.MustNewConstMetric(c.dataRxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.RxDataFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.dataTxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.TxDataFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.managementRxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.RxMgmtFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.managementTxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.TxMgmtFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.controlRxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.RxCtrlFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.controlTxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.TxCtrlFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.multicastRxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.MulticastRxFrameCnt), labels...)
			ch <- prometheus.MustNewConstMetric(c.multicastTxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.MulticastTxFrameCnt), labels...)
			ch <- prometheus.MustNewConstMetric(c.totalTxFramesTotalDesc, prometheus.CounterValue, float64(radioStats.TxFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.rtsSuccessTotalDesc, prometheus.CounterValue, float64(radioStats.RtsSuccessCount), labels...)
		}

		// Collect Error metrics
		if c.metrics.Errors {
			// Radio operational errors
			ch <- prometheus.MustNewConstMetric(c.rxErrorsTotalDesc, prometheus.CounterValue, float64(radioStats.RxErrorFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.txRetriesTotalDesc, prometheus.CounterValue, float64(radioStats.RetryCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.txDropsTotalDesc, prometheus.CounterValue, float64(radioStats.AckFailureCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.txErrorsTotalDesc, prometheus.CounterValue, float64(radioStats.FailedCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.ackFailuresTotalDesc, prometheus.CounterValue, float64(radioStats.FailedCount), labels...)

			// Frame-level errors
			ch <- prometheus.MustNewConstMetric(c.duplicateFramesTotalDesc, prometheus.CounterValue, float64(radioStats.FrameDuplicateCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.fcsErrorsTotalDesc, prometheus.CounterValue, float64(radioStats.FcsErrorCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.fragmentationRxTotalDesc, prometheus.CounterValue, float64(radioStats.RxFragmentCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.fragmentationTxTotalDesc, prometheus.CounterValue, float64(radioStats.TxFragmentCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.rtsFailuresTotalDesc, prometheus.CounterValue, float64(radioStats.RtsFailureCount), labels...)

			// Security-related errors
			ch <- prometheus.MustNewConstMetric(c.decryptionErrorsTotalDesc, prometheus.CounterValue, float64(radioStats.MACDecryErrFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.micErrorsTotalDesc, prometheus.CounterValue, float64(radioStats.MACMicErrFrameCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.wepUndecryptableTotalDesc, prometheus.CounterValue, float64(radioStats.WepUndecryptableCount), labels...)

			// RF events (placeholder values - not verified in current IOS-XE version)
			ch <- prometheus.MustNewConstMetric(c.coverageHoleEventsDesc, prometheus.CounterValue, 0.0, labels...)
			ch <- prometheus.MustNewConstMetric(c.interferenceEventsDesc, prometheus.CounterValue, 0.0, labels...)
			ch <- prometheus.MustNewConstMetric(c.dfsChannelChangesDesc, prometheus.CounterValue, 0.0, labels...)
			ch <- prometheus.MustNewConstMetric(c.radarDetectedEventsDesc, prometheus.CounterValue, 0.0, labels...)
			ch <- prometheus.MustNewConstMetric(c.radioResetTotalDesc, prometheus.CounterValue, 0.0, labels...)
		}
	}
}

// collectRFMetrics collects all RF-related metrics efficiently.
func (c *APCollector) collectRFMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	labels []string,
	ctx *rfMetricContext,
) {
	// Get measurement data once
	measurement := ctx.measurements[radio.WtpMAC+":"+strconv.Itoa(radio.RadioSlotID)]

	// Early return if no measurement data
	if measurement == nil {
		return
	}

	// Noise floor
	if c.noiseFloorDesc != nil {
		value := float64(measurement.Noise.Noise.NoiseData[0].Noise)
		ch <- prometheus.MustNewConstMetric(c.noiseFloorDesc, prometheus.GaugeValue, value, labels...)
	}

	// Channel utilization
	if c.channelUtilizationDesc != nil {
		value := float64(measurement.Load.CcaUtilPercentage)
		ch <- prometheus.MustNewConstMetric(c.channelUtilizationDesc, prometheus.GaugeValue, value, labels...)
	}

	// RX utilization
	if c.rxUtilizationDesc != nil {
		value := float64(measurement.Load.RxUtilPercentage)
		ch <- prometheus.MustNewConstMetric(c.rxUtilizationDesc, prometheus.GaugeValue, value, labels...)
	}

	// TX utilization
	if c.txUtilizationDesc != nil {
		value := float64(measurement.Load.TxUtilPercentage)
		ch <- prometheus.MustNewConstMetric(c.txUtilizationDesc, prometheus.GaugeValue, value, labels...)
	}

	// Noise utilization
	if c.noiseUtilizationDesc != nil {
		value := float64(measurement.Load.RxNoiseChannelUtilization)
		ch <- prometheus.MustNewConstMetric(c.noiseUtilizationDesc, prometheus.GaugeValue, value, labels...)
	}

	// Associated clients
	if c.associatedClientsDesc != nil {
		var value float64
		if apCounts, ok := ctx.clientCounts[radio.WtpMAC]; ok {
			if count, ok := apCounts[radio.RadioSlotID]; ok {
				value = float64(count)
			}
		}
		ch <- prometheus.MustNewConstMetric(c.associatedClientsDesc, prometheus.GaugeValue, value, labels...)
	}
}

// collectPhyMetrics collects all physical layer metrics efficiently.
func (c *APCollector) collectPhyMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	labels []string,
) {
	// TX power
	if c.txPowerDesc != nil {
		value := float64(radio.RadioBandInfo[0].PhyTxPwrLvlCfg.CfgData.CurrTxPowerInDbm)
		ch <- prometheus.MustNewConstMetric(c.txPowerDesc, prometheus.GaugeValue, value, labels...)
	}

	// Channel number
	if c.channelDesc != nil {
		value := float64(radio.PhyHtCfg.CfgData.CurrFreq)
		ch <- prometheus.MustNewConstMetric(c.channelDesc, prometheus.GaugeValue, value, labels...)
	}

	// Channel width
	if c.channelWidthDesc != nil {
		value := float64(radio.PhyHtCfg.CfgData.ChanWidth)
		ch <- prometheus.MustNewConstMetric(c.channelWidthDesc, prometheus.GaugeValue, value, labels...)
	}

	// TX power level (default to 1)
	if c.txPowerLevelDesc != nil {
		ch <- prometheus.MustNewConstMetric(c.txPowerLevelDesc, prometheus.GaugeValue, 1.0, labels...)
	}

	// TX power max (use level-1 power)
	if c.txPowerMaxDesc != nil {
		value := float64(radio.RadioBandInfo[0].PhyTxPwrLvlCfg.CfgData.TxPowerLevel1)
		ch <- prometheus.MustNewConstMetric(c.txPowerMaxDesc, prometheus.GaugeValue, value, labels...)
	}
}

// collectStateMetrics collects all state-related metrics efficiently.
func (c *APCollector) collectStateMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	labels []string,
) {
	// Radio state
	if c.radioStateDesc != nil {
		value := boolToFloat64(radio.OperState == APRadioStateUp)
		ch <- prometheus.MustNewConstMetric(c.radioStateDesc, prometheus.GaugeValue, value, labels...)
	}

	// Admin state
	if c.adminStateDesc != nil {
		value := boolToFloat64(radio.AdminState == APAdminStateEnabled)
		ch <- prometheus.MustNewConstMetric(c.adminStateDesc, prometheus.GaugeValue, value, labels...)
	}

	// Oper state
	if c.operStateDesc != nil {
		value := boolToFloat64(radio.OperState == APRadioStateUp)
		ch <- prometheus.MustNewConstMetric(c.operStateDesc, prometheus.GaugeValue, value, labels...)
	}
}

// getClientDataForAPMetrics retrieves client data and AP name to MAC mapping for AP metrics.
// This method uses the dedicated clientSrc and src to get required data.
func (c *APCollector) getClientDataForAPMetrics() (*client.ClientOperCommonOperData, *ap.ApOperApNameMACMap, error) {
	// Get client data from dedicated client source
	clientData, err := c.clientSrc.GetClientData(context.Background())
	if err != nil {
		return nil, nil, err
	}

	// Get AP name to MAC mapping from AP source
	nameMACMaps, err := c.src.ListNameMACMaps(context.Background())
	if err != nil {
		return clientData, nil, err
	}

	return clientData, nameMACMaps, nil
}

// buildAPClientCounts builds client counts by AP MAC and radio slot from client data using name to MAC mapping.
func (c *APCollector) buildAPClientCounts(
	clientData *client.ClientOperCommonOperData,
	nameMACMaps *ap.ApOperApNameMACMap,
) map[string]map[int]int {
	clientCounts := make(map[string]map[int]int)

	if clientData == nil || nameMACMaps == nil {
		return clientCounts
	}

	// Build AP name to MAC mapping for efficient lookup
	nameToMAC := make(map[string]string)
	for _, mapping := range nameMACMaps.ApNameMACMap {
		nameToMAC[mapping.WtpName] = mapping.WtpMAC
	}

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		// Only count RUN state clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Get AP name and map to MAC address
		apName := commonData.ApName
		apMAC, exists := nameToMAC[apName]
		if !exists {
			continue
		}

		// Determine radio slot from radio type
		radioSlot := MapRadioTypeToSlot(commonData.MsRadioType)
		if radioSlot == -1 {
			continue
		}

		// Initialize maps if needed
		if clientCounts[apMAC] == nil {
			clientCounts[apMAC] = make(map[int]int)
		}

		// Increment count
		clientCounts[apMAC][radioSlot]++
	}

	return clientCounts
}

// getRRMDataForAPMetrics retrieves RRM measurement data for AP metrics.
func (c *APCollector) getRRMDataForAPMetrics() (*rrm.RRMOperRRMMeasurement, error) {
	ctx := context.Background()
	return c.rrmSrc.GetRRMMeasurement(ctx)
}

// buildRFMetricContext builds the common context needed for RF metrics collection.
func (c *APCollector) buildRFMetricContext() *rfMetricContext {
	// Get client data and AP name to MAC mapping for accurate client counts
	clientData, nameMACMaps, err := c.getClientDataForAPMetrics()
	if err != nil {
		slog.Warn("Failed to get client data for AP metrics", "error", err)
	}

	// Get RRM measurement data for channel utilization
	rrmData, err := c.getRRMDataForAPMetrics()
	if err != nil {
		slog.Warn("Failed to get RRM data for AP metrics", "error", err)
	}

	// Build client count by AP MAC and radio slot from client data and name mapping
	clientCounts := c.buildAPClientCounts(clientData, nameMACMaps)

	// Pre-calculate measurements per radio to avoid repeated loops
	measurements := make(map[string]*rrm.RRMMeasurement)
	if rrmData != nil {
		for _, measurement := range rrmData.RRMMeasurement {
			// Use radioSlot as key since we already have WtpMAC in outer loop
			key := strconv.Itoa(measurement.RadioSlotID)
			measurements[measurement.WtpMAC+":"+key] = &measurement
		}
	}

	return &rfMetricContext{
		rrmData:      rrmData,
		clientCounts: clientCounts,
		measurements: measurements,
	}
}

// findRadioStats finds the corresponding radio statistics for a given radio.
func (c *APCollector) findRadioStats(
	radio *ap.RadioOperData,
	radioOperStats *ap.ApOperRadioOperStats,
) *ap.RadioOperStats {
	for _, stats := range radioOperStats.RadioOperStats {
		if stats.ApMAC == radio.WtpMAC && stats.SlotID == radio.RadioSlotID {
			return &stats
		}
	}
	return nil
}
