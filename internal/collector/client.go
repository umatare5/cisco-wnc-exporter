// Package collector provides collectors for cisco-wnc-exporter.
package collector

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// ClientMetrics represents the configuration for client metrics.
type ClientMetrics struct {
	Inventory  bool
	Info       bool
	InfoLabels []string
	Session    bool
	Phy        bool
	RF         bool
	Traffic    bool
	Errors     bool
	Power      bool
}

// ClientCollector implements prometheus.Collector for Client metrics.
type ClientCollector struct {
	src     wnc.ClientSource
	metrics ClientMetrics

	// Inventory metrics
	countDesc *prometheus.Desc

	// Info metrics
	infoDesc       *prometheus.Desc
	infoLabelNames []string // Store configured label names

	// Session metrics
	stateDesc             *prometheus.Desc
	associationUptimeDesc *prometheus.Desc

	// PHY metrics
	protocolDesc       *prometheus.Desc
	mcsIndexDesc       *prometheus.Desc
	spatialStreamsDesc *prometheus.Desc
	speedDesc          *prometheus.Desc

	// RF metrics
	rssiDesc *prometheus.Desc
	snrDesc  *prometheus.Desc

	// Traffic metrics
	bytesRxDesc   *prometheus.Desc
	bytesTxDesc   *prometheus.Desc
	packetsRxDesc *prometheus.Desc
	packetsTxDesc *prometheus.Desc

	// Error metrics
	retryRatioDesc        *prometheus.Desc
	policyErrorsDesc      *prometheus.Desc
	duplicateReceivedDesc *prometheus.Desc
	decryptionFailedDesc  *prometheus.Desc
	micMismatchDesc       *prometheus.Desc
	micMissingDesc        *prometheus.Desc
	excessiveRetriesDesc  *prometheus.Desc
	rxGroupCounterDesc    *prometheus.Desc
	txTotalDropsDesc      *prometheus.Desc
	dataRetriesDesc       *prometheus.Desc
	rtsRetriesDesc        *prometheus.Desc
	txRetriesDesc         *prometheus.Desc

	// Power metrics
	powerSaveStateDesc *prometheus.Desc
}

// NewClientCollector creates a new ClientCollector for retrieving metrics from WNC.
func NewClientCollector(
	src wnc.ClientSource,
	metrics ClientMetrics,
) *ClientCollector {
	collector := &ClientCollector{
		src:     src,
		metrics: metrics,
	}

	// Define base labels - all metrics now use only MAC address
	baseLabels := []string{"mac"}

	// Inventory module metrics
	if metrics.Inventory {
		collector.countDesc = prometheus.NewDesc(
			"wnc_clients_total",
			"Total number of connected clients",
			[]string{"state"}, nil,
		)
	}

	// Info module metrics - provides all additional labels for joining
	if metrics.Info {
		// Build dynamic labels based on configuration
		infoLabels := buildClientInfoLabelsConfig(metrics.InfoLabels)
		collector.infoDesc = prometheus.NewDesc(
			"wnc_client_info",
			"Client information labels for joining with other metrics",
			infoLabels, nil,
		)
		collector.infoLabelNames = infoLabels // Store for later use
	}

	// Session module metrics
	if metrics.Session {
		collector.stateDesc = prometheus.NewDesc(
			"wnc_client_state",
			"Client connection state",
			baseLabels, nil,
		)
		collector.associationUptimeDesc = prometheus.NewDesc(
			"wnc_client_uptime_seconds",
			"Client association uptime in seconds",
			baseLabels, nil,
		)
	}

	// PHY module metrics
	if metrics.Phy {
		collector.protocolDesc = prometheus.NewDesc(
			"wnc_client_protocol",
			"Client wireless protocol (1=802.11a, 2=802.11b, 3=802.11g, 4=802.11n, 5=802.11ac, 6=802.11ax)",
			baseLabels, nil,
		)
		collector.mcsIndexDesc = prometheus.NewDesc(
			"wnc_client_mcs_index",
			"Client MCS index (-1=legacy, 0-11=802.11n/ac/ax MCS values)",
			baseLabels, nil,
		)
		collector.spatialStreamsDesc = prometheus.NewDesc(
			"wnc_client_spatial_streams",
			"Number of spatial streams",
			baseLabels, nil,
		)
		collector.speedDesc = prometheus.NewDesc(
			"wnc_client_speed_mbps",
			"Connection speed in Mbps",
			baseLabels, nil,
		)
	}

	// RF module metrics
	if metrics.RF {
		collector.rssiDesc = prometheus.NewDesc(
			"wnc_client_rssi_dbm",
			"Received signal strength in dBm",
			baseLabels, nil,
		)
		collector.snrDesc = prometheus.NewDesc(
			"wnc_client_snr_decibels",
			"Signal-to-noise ratio in dB",
			baseLabels, nil,
		)
	}

	// Traffic module metrics
	if metrics.Traffic {
		collector.bytesRxDesc = prometheus.NewDesc(
			"wnc_client_rx_bytes_total",
			"Total bytes received",
			baseLabels, nil,
		)
		collector.bytesTxDesc = prometheus.NewDesc(
			"wnc_client_tx_bytes_total",
			"Total bytes transmitted",
			baseLabels, nil,
		)
		collector.packetsRxDesc = prometheus.NewDesc(
			"wnc_client_rx_packets_total",
			"Total packets received",
			baseLabels, nil,
		)
		collector.packetsTxDesc = prometheus.NewDesc(
			"wnc_client_tx_packets_total",
			"Total packets transmitted",
			baseLabels, nil,
		)
	}

	// Error module metrics
	if metrics.Errors {
		collector.retryRatioDesc = prometheus.NewDesc(
			"wnc_client_retry_ratio_percent",
			"Retry rate percentage",
			baseLabels, nil,
		)
		collector.policyErrorsDesc = prometheus.NewDesc(
			"wnc_client_policy_errors_total",
			"Policy errors",
			baseLabels, nil,
		)
		collector.duplicateReceivedDesc = prometheus.NewDesc(
			"wnc_client_duplicate_received_total",
			"Duplicate packets received",
			baseLabels, nil,
		)
		collector.decryptionFailedDesc = prometheus.NewDesc(
			"wnc_client_decryption_failed_total",
			"Decryption failed packets",
			baseLabels, nil,
		)
		collector.micMismatchDesc = prometheus.NewDesc(
			"wnc_client_mic_mismatch_total",
			"MIC mismatch errors",
			baseLabels, nil,
		)
		collector.micMissingDesc = prometheus.NewDesc(
			"wnc_client_mic_missing_total",
			"MIC missing errors",
			baseLabels, nil,
		)
		collector.excessiveRetriesDesc = prometheus.NewDesc(
			"wnc_client_excessive_retries_total",
			"Excessive retries",
			baseLabels, nil,
		)
		collector.rxGroupCounterDesc = prometheus.NewDesc(
			"wnc_client_rx_group_counter_total",
			"RX group counter",
			baseLabels, nil,
		)
		collector.txTotalDropsDesc = prometheus.NewDesc(
			"wnc_client_tx_drops_total",
			"Total TX drops",
			baseLabels, nil,
		)
		collector.dataRetriesDesc = prometheus.NewDesc(
			"wnc_client_data_retries_total",
			"Data retries",
			baseLabels, nil,
		)
		collector.rtsRetriesDesc = prometheus.NewDesc(
			"wnc_client_rts_retries_total",
			"RTS retries",
			baseLabels, nil,
		)
		collector.txRetriesDesc = prometheus.NewDesc(
			"wnc_client_tx_retries_total",
			"TX retries",
			baseLabels, nil,
		)
	}

	// Power module metrics
	if metrics.Power {
		collector.powerSaveStateDesc = prometheus.NewDesc(
			"wnc_client_power_save_state",
			"Power save state (0=active, 1=power-save)",
			baseLabels, nil,
		)
	}

	return collector
}

// Describe sends the descriptors of all metrics to the provided channel.
func (c *ClientCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.Inventory {
		c.describeInventoryMetrics(ch)
	}
	if c.metrics.Info {
		c.describeInfoMetrics(ch)
	}
	if c.metrics.Session {
		c.describeSessionMetrics(ch)
	}
	if c.metrics.Phy {
		c.describePhyMetrics(ch)
	}
	if c.metrics.RF {
		c.describeRFMetrics(ch)
	}
	if c.metrics.Traffic {
		c.describeTrafficMetrics(ch)
	}
	if c.metrics.Errors {
		c.describeErrorMetrics(ch)
	}
	if c.metrics.Power {
		c.describePowerMetrics(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (c *ClientCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Get client data from WNC
	clientData, err := c.src.GetClientData(ctx)
	if err != nil {
		slog.Warn("Failed to retrieve client data", "error", err)
		return
	}

	deviceData, err := c.src.GetDeviceData(ctx)
	if err != nil {
		slog.Warn("Failed to retrieve device data", "error", err)
		return
	}

	dot11Data, err := c.src.GetDot11Data(ctx)
	if err != nil {
		slog.Warn("Failed to retrieve dot11 data", "error", err)
		return
	}

	sisfdbData, err := c.src.GetSISFDBData(ctx)
	if err != nil {
		slog.Warn("Failed to retrieve SISF database data", "error", err)
		return
	}

	trafficStats, err := c.src.GetTrafficStats(ctx)
	if err != nil {
		slog.Warn("Failed to retrieve traffic stats", "error", err)
		return
	}

	// Collect metrics according to enabled metrics - 1:1 with describe methods
	if c.metrics.Inventory {
		c.collectInventoryMetrics(ch, clientData)
	}
	if c.metrics.Info {
		c.collectInfoMetrics(ch, clientData, deviceData, dot11Data, sisfdbData)
	}
	if c.metrics.Session {
		c.collectSessionMetrics(ch, clientData, dot11Data)
	}
	if c.metrics.Phy {
		c.collectPhyMetrics(ch, clientData, dot11Data, trafficStats)
	}
	if c.metrics.RF {
		c.collectRFMetrics(ch, clientData, trafficStats)
	}
	if c.metrics.Traffic {
		c.collectTrafficMetrics(ch, clientData, trafficStats)
	}
	if c.metrics.Errors {
		c.collectErrorMetrics(ch, clientData, trafficStats)
	}
	if c.metrics.Power {
		c.collectPowerMetrics(ch, clientData, trafficStats)
	}
}

// describeInventoryMetrics describes inventory-related metric descriptors.
func (c *ClientCollector) describeInventoryMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.countDesc
}

// describeInfoMetrics describes info-related metric descriptors.
func (c *ClientCollector) describeInfoMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
}

// describeSessionMetrics describes session-related metric descriptors.
func (c *ClientCollector) describeSessionMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.stateDesc
	ch <- c.associationUptimeDesc
}

// describePhyMetrics describes PHY-related metric descriptors.
func (c *ClientCollector) describePhyMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.protocolDesc
	ch <- c.mcsIndexDesc
	ch <- c.spatialStreamsDesc
	ch <- c.speedDesc
}

// describeRFMetrics describes RF-related metric descriptors.
func (c *ClientCollector) describeRFMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.rssiDesc
	ch <- c.snrDesc
}

// describeTrafficMetrics describes traffic-related metric descriptors.
func (c *ClientCollector) describeTrafficMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.bytesRxDesc
	ch <- c.bytesTxDesc
	ch <- c.packetsRxDesc
	ch <- c.packetsTxDesc
}

// describeErrorMetrics describes error-related metric descriptors.
func (c *ClientCollector) describeErrorMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.retryRatioDesc
	ch <- c.policyErrorsDesc
	ch <- c.duplicateReceivedDesc
	ch <- c.decryptionFailedDesc
	ch <- c.micMismatchDesc
	ch <- c.micMissingDesc
	ch <- c.excessiveRetriesDesc
	ch <- c.rxGroupCounterDesc
	ch <- c.txTotalDropsDesc
	ch <- c.dataRetriesDesc
	ch <- c.rtsRetriesDesc
	ch <- c.txRetriesDesc
}

// describePowerMetrics describes power-related metric descriptors.
func (c *ClientCollector) describePowerMetrics(ch chan<- *prometheus.Desc) {
	ch <- c.powerSaveStateDesc
}

// collectInventoryMetrics collects inventory-related metrics.
func (c *ClientCollector) collectInventoryMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
) {
	if clientData == nil {
		return
	}

	// Count clients by state
	stateCounts := make(map[string]int)
	for _, client := range clientData.CommonOperData {
		state := client.CoState
		if state == "" {
			state = "unknown"
		}
		stateCounts[state]++
	}

	// Emit metrics for each state
	for state, count := range stateCounts {
		if c.countDesc != nil {
			ch <- prometheus.MustNewConstMetric(
				c.countDesc,
				prometheus.GaugeValue,
				float64(count),
				state,
			)
		}
	}
}

// collectInfoMetrics collects client info metrics from WNC data.
func (c *ClientCollector) collectInfoMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
	deviceData *client.ClientOperDcInfo,
	dot11Data *client.ClientOperDot11OperData,
	sisfdbData *client.ClientOperSisfDBMac,
) {
	if clientData == nil {
		return
	}

	// Build lookup maps for efficient processing
	deviceMap := buildDeviceMap(deviceData)
	dot11Map := buildDot11Map(dot11Data)
	sisfMap := buildSISFMap(sisfdbData)

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		if !isValidMACAddress(commonData.ClientMAC) {
			continue
		}

		// Only process associated clients for info metrics
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Build full info labels
		labels := c.buildClientInfoLabels(commonData, deviceMap, dot11Map, sisfMap)

		// Emit info metric (always value 1)
		if c.infoDesc != nil {
			ch <- prometheus.MustNewConstMetric(
				c.infoDesc,
				prometheus.GaugeValue,
				1,
				labels...,
			)
		}
	}
}

// collectSessionMetrics collects session-related metrics.
func (c *ClientCollector) collectSessionMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
	dot11Data *client.ClientOperDot11OperData,
) {
	if clientData == nil {
		return
	}

	// Build lookup maps
	dot11Map := buildDot11Map(dot11Data)

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		if !isValidMACAddress(commonData.ClientMAC) {
			continue
		}

		// Only process associated clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Build base labels
		baseLabels := []string{commonData.ClientMAC}

		// Client state
		if c.stateDesc != nil {
			state := mapClientState(commonData.CoState)
			ch <- prometheus.MustNewConstMetric(
				c.stateDesc,
				prometheus.GaugeValue,
				float64(state),
				baseLabels...,
			)
		}

		// Association uptime
		if c.associationUptimeDesc != nil {
			if dot11, ok := dot11Map[commonData.ClientMAC]; ok {
				if !dot11.MsAssocTime.IsZero() {
					uptime := time.Since(dot11.MsAssocTime).Seconds()
					ch <- prometheus.MustNewConstMetric(
						c.associationUptimeDesc,
						prometheus.GaugeValue,
						uptime,
						baseLabels...,
					)
				}
			}
		}
	}
}

// collectPhyMetrics collects PHY-related metrics.
func (c *ClientCollector) collectPhyMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
	dot11Data *client.ClientOperDot11OperData,
	trafficStats *client.ClientOperTrafficStats,
) {
	if clientData == nil {
		return
	}

	// Build lookup maps
	dot11Map := buildDot11Map(dot11Data)
	trafficMap := buildTrafficMap(trafficStats)

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		if !isValidMACAddress(commonData.ClientMAC) {
			continue
		}

		// Only process associated clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Build base labels
		baseLabels := []string{commonData.ClientMAC}

		// Protocol information from dot11 data
		if dot11, ok := dot11Map[commonData.ClientMAC]; ok {
			if c.protocolDesc != nil {
				protocol := mapWirelessProtocol(dot11.EwlcMsPhyType, dot11.RadioType, dot11.Is11GClient)
				ch <- prometheus.MustNewConstMetric(
					c.protocolDesc,
					prometheus.GaugeValue,
					float64(protocol),
					baseLabels...,
				)
			}
		}

		// Speed and streams from traffic stats
		if traffic, ok := trafficMap[commonData.ClientMAC]; ok {
			if c.speedDesc != nil && traffic.Speed != 0 {
				ch <- prometheus.MustNewConstMetric(
					c.speedDesc,
					prometheus.GaugeValue,
					float64(traffic.Speed),
					baseLabels...,
				)
			}

			if c.spatialStreamsDesc != nil && traffic.SpatialStream != 0 {
				ch <- prometheus.MustNewConstMetric(
					c.spatialStreamsDesc,
					prometheus.GaugeValue,
					float64(traffic.SpatialStream),
					baseLabels...,
				)
			}

			if c.mcsIndexDesc != nil {
				mcsIndex := parseMCSIndex(traffic.CurrentRate)
				ch <- prometheus.MustNewConstMetric(
					c.mcsIndexDesc,
					prometheus.GaugeValue,
					float64(mcsIndex),
					baseLabels...,
				)
			}
		}
	}
}

// collectRFMetrics collects RF-related metrics.
func (c *ClientCollector) collectRFMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
	trafficStats *client.ClientOperTrafficStats,
) {
	if clientData == nil {
		return
	}

	// Build lookup maps
	trafficMap := buildTrafficMap(trafficStats)

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		if !isValidMACAddress(commonData.ClientMAC) {
			continue
		}

		// Only process associated clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Build base labels
		baseLabels := []string{commonData.ClientMAC}

		if traffic, ok := trafficMap[commonData.ClientMAC]; ok {
			if c.rssiDesc != nil {
				ch <- prometheus.MustNewConstMetric(
					c.rssiDesc,
					prometheus.GaugeValue,
					float64(traffic.MostRecentRSSI),
					baseLabels...,
				)
			}

			if c.snrDesc != nil {
				ch <- prometheus.MustNewConstMetric(
					c.snrDesc,
					prometheus.GaugeValue,
					float64(traffic.MostRecentSNR),
					baseLabels...,
				)
			}
		}
	}
}

// collectTrafficMetrics collects traffic-related metrics.
func (c *ClientCollector) collectTrafficMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
	trafficStats *client.ClientOperTrafficStats,
) {
	if clientData == nil {
		return
	}

	// Build lookup maps
	trafficMap := buildTrafficMap(trafficStats)

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		if !isValidMACAddress(commonData.ClientMAC) {
			continue
		}

		// Only process associated clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Build base labels
		baseLabels := []string{commonData.ClientMAC}

		if traffic, ok := trafficMap[commonData.ClientMAC]; ok {
			if c.bytesRxDesc != nil {
				bytesRx := parseUint64Safely(traffic.BytesRx)
				ch <- prometheus.MustNewConstMetric(
					c.bytesRxDesc,
					prometheus.CounterValue,
					float64(bytesRx),
					baseLabels...,
				)
			}

			if c.bytesTxDesc != nil {
				bytesTx := parseUint64Safely(traffic.BytesTx)
				ch <- prometheus.MustNewConstMetric(
					c.bytesTxDesc,
					prometheus.CounterValue,
					float64(bytesTx),
					baseLabels...,
				)
			}

			if c.packetsRxDesc != nil {
				packetsRx := parseUint64Safely(traffic.PktsRx)
				ch <- prometheus.MustNewConstMetric(
					c.packetsRxDesc,
					prometheus.CounterValue,
					float64(packetsRx),
					baseLabels...,
				)
			}

			if c.packetsTxDesc != nil {
				packetsTx := parseUint64Safely(traffic.PktsTx)
				ch <- prometheus.MustNewConstMetric(
					c.packetsTxDesc,
					prometheus.CounterValue,
					float64(packetsTx),
					baseLabels...,
				)
			}
		}
	}
}

// collectErrorMetrics collects error-related metrics.
func (c *ClientCollector) collectErrorMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
	trafficStats *client.ClientOperTrafficStats,
) {
	if clientData == nil {
		return
	}

	// Build lookup maps
	trafficMap := buildTrafficMap(trafficStats)

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		if !isValidMACAddress(commonData.ClientMAC) {
			continue
		}

		// Only process associated clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Build base labels
		baseLabels := []string{commonData.ClientMAC}

		if traffic, ok := trafficMap[commonData.ClientMAC]; ok {
			// Calculate retry ratio
			if c.retryRatioDesc != nil {
				dataRetries := parseUint64Safely(traffic.DataRetries)
				txRetries := parseUint64Safely(traffic.TxRetries)
				packetsTx := parseUint64Safely(traffic.PktsTx)

				if packetsTx > 0 {
					retryRatio := (float64(dataRetries+txRetries) / float64(packetsTx)) * 100.0
					ch <- prometheus.MustNewConstMetric(
						c.retryRatioDesc,
						prometheus.GaugeValue,
						retryRatio,
						baseLabels...,
					)
				}
			}

			// Error counters
			errorMetrics := []struct {
				desc  *prometheus.Desc
				value string
			}{
				{c.policyErrorsDesc, traffic.PolicyErrs},
				{c.duplicateReceivedDesc, traffic.DuplicateRcv},
				{c.decryptionFailedDesc, traffic.DecryptFailed},
				{c.micMismatchDesc, traffic.MicMismatch},
				{c.micMissingDesc, traffic.MicMissing},
				{c.excessiveRetriesDesc, traffic.TxExcessiveRetries},
				{c.rxGroupCounterDesc, traffic.RxGroupCounter},
				{c.txTotalDropsDesc, traffic.TxTotalDrops},
				{c.dataRetriesDesc, traffic.DataRetries},
				{c.rtsRetriesDesc, traffic.RtsRetries},
				{c.txRetriesDesc, traffic.TxRetries},
			}

			for _, metric := range errorMetrics {
				if metric.desc != nil {
					value := parseUint64Safely(metric.value)
					ch <- prometheus.MustNewConstMetric(
						metric.desc,
						prometheus.CounterValue,
						float64(value),
						baseLabels...,
					)
				}
			}
		}
	}
}

// collectPowerMetrics collects power-related metrics.
func (c *ClientCollector) collectPowerMetrics(
	ch chan<- prometheus.Metric,
	clientData *client.ClientOperCommonOperData,
	trafficStats *client.ClientOperTrafficStats,
) {
	if clientData == nil {
		return
	}

	// Build lookup maps
	trafficMap := buildTrafficMap(trafficStats)

	// Process each client
	for _, commonData := range clientData.CommonOperData {
		if !isValidMACAddress(commonData.ClientMAC) {
			continue
		}

		// Only process associated clients
		if commonData.CoState != ClientStatusRun {
			continue
		}

		// Build base labels
		baseLabels := []string{commonData.ClientMAC}

		if traffic, ok := trafficMap[commonData.ClientMAC]; ok {
			if c.powerSaveStateDesc != nil {
				ch <- prometheus.MustNewConstMetric(
					c.powerSaveStateDesc,
					prometheus.GaugeValue,
					float64(traffic.PowerSaveState),
					baseLabels...,
				)
			}
		}
	}
}

// Helper functions to build lookup maps.
func buildDeviceMap(deviceData *client.ClientOperDcInfo) map[string]client.DcInfo {
	deviceMap := make(map[string]client.DcInfo)
	if deviceData != nil {
		for _, device := range deviceData.DcInfo {
			if device.ClientMAC != "" {
				deviceMap[device.ClientMAC] = device
			}
		}
	}
	return deviceMap
}

func buildDot11Map(dot11Data *client.ClientOperDot11OperData) map[string]client.Dot11OperData {
	dot11Map := make(map[string]client.Dot11OperData)
	if dot11Data != nil {
		for _, dot11 := range dot11Data.Dot11OperData {
			if dot11.MsMACAddress != "" {
				dot11Map[dot11.MsMACAddress] = dot11
			}
		}
	}
	return dot11Map
}

func buildSISFMap(sisfdbData *client.ClientOperSisfDBMac) map[string]client.SisfDBMac {
	sisfMap := make(map[string]client.SisfDBMac)
	if sisfdbData != nil {
		for _, sisf := range sisfdbData.SisfDBMac {
			if sisf.MACAddr != "" {
				sisfMap[sisf.MACAddr] = sisf
			}
		}
	}
	return sisfMap
}

func buildTrafficMap(trafficStats *client.ClientOperTrafficStats) map[string]client.TrafficStats {
	trafficMap := make(map[string]client.TrafficStats)
	if trafficStats != nil {
		for _, traffic := range trafficStats.TrafficStats {
			if traffic.MsMACAddress != "" {
				trafficMap[traffic.MsMACAddress] = traffic
			}
		}
	}
	return trafficMap
}

// buildClientInfoLabels builds full info labels for the client.
func (c *ClientCollector) buildClientInfoLabels(
	commonData client.CommonOperData,
	deviceMap map[string]client.DcInfo,
	dot11Map map[string]client.Dot11OperData,
	sisfMap map[string]client.SisfDBMac,
) []string {
	var name, wlan, ipv4, ipv6 string

	// Device name from device classification
	if device, ok := deviceMap[commonData.ClientMAC]; ok {
		name = device.DeviceName
	}

	// WLAN name from dot11 data
	if dot11, ok := dot11Map[commonData.ClientMAC]; ok {
		wlan = dot11.VapSsid
	}

	// IP addresses from SISF
	if sisf, ok := sisfMap[commonData.ClientMAC]; ok {
		// IPv4
		if sisf.Ipv4Binding.IPKey.IPAddr != "" {
			ipv4 = sisf.Ipv4Binding.IPKey.IPAddr
		}
		// IPv6
		for _, ipv6Binding := range sisf.Ipv6Binding {
			if ipv6Binding.Ipv6BindingIPKey.IPAddr != "" &&
				!strings.HasPrefix(ipv6Binding.Ipv6BindingIPKey.IPAddr, "fe80:") {
				ipv6 = ipv6Binding.Ipv6BindingIPKey.IPAddr
				break
			}
		}
	}

	// Determine band
	band := DetermineBandFromRadioInfo(-1, commonData.MsRadioType)

	// Build dynamic labels based on configured labels
	labelNames := c.infoLabelNames
	values := make([]string, len(labelNames))

	// Map field names to values
	valueMap := map[string]string{
		"mac":      commonData.ClientMAC,
		"ap":       commonData.ApName,
		"band":     band,
		"wlan":     wlan,
		"name":     name,
		"username": commonData.Username,
		"ipv4":     ipv4,
		"ipv6":     ipv6,
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

// parseMCSIndex extracts MCS index from current-rate string.
// Examples: "m7" -> 7, "m9 ss2" -> 9, "6.0" -> -1 (legacy rate).
func parseMCSIndex(currentRate string) int {
	if currentRate == "" {
		return -1
	}

	// Handle MCS format: "m7", "m9 ss2", etc.
	if strings.HasPrefix(currentRate, "m") {
		// Extract number after 'm'
		parts := strings.Fields(currentRate)
		mcsStr := strings.TrimPrefix(parts[0], "m")
		if mcs, err := strconv.Atoi(mcsStr); err == nil {
			return mcs
		}
	}

	// Legacy rates (e.g., "6.0", "54.0") return -1
	return -1
}

// parseUint64Safely safely parses a string to uint64, returning 0 for empty strings or parse errors.
func parseUint64Safely(value string) uint64 {
	if value == "" {
		return 0
	}
	if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
		return parsed
	}
	return 0
}

// isValidMACAddress performs basic validation on MAC address format.
func isValidMACAddress(mac string) bool {
	return mac != "" && len(mac) >= 17 // Basic length check for MAC address format
}

// mapClientState maps client operational state to numeric value.
func mapClientState(state string) int {
	switch state {
	case ClientStatusRun:
		return ClientStateAssociated // associated
	case "client-status-authenticated":
		return 1 // authenticated
	default:
		return 0 // disconnected
	}
}

// mapWirelessProtocol maps WNC PHY type strings and radio information to WirelessProtocol enum values.
func mapWirelessProtocol(phyType, radioType string, is11GClient bool) WirelessProtocol {
	switch {
	case strings.Contains(phyType, "dot11n"):
		return ProtocolN
	case strings.Contains(phyType, "dot11ac"):
		return ProtocolAC
	case strings.Contains(phyType, "dot11ax"):
		return ProtocolAX
	case strings.Contains(phyType, "dot11be"), strings.Contains(phyType, "eht"):
		return ProtocolBE
	case strings.Contains(phyType, "dot11bg"):
		// 802.11b/g mixed mode - determine by is-11g-client flag
		if is11GClient {
			return Protocol11G
		}
		return Protocol11B
	case strings.Contains(phyType, "dot11a") || radioType == "dot11-radio-type-a":
		return Protocol11A
	case strings.Contains(phyType, "dot11g") || is11GClient:
		return Protocol11G
	default:
		return ProtocolUnknown
	}
}

// buildClientInfoLabelsConfig constructs the Client info labels slice based on configuration.
func buildClientInfoLabelsConfig(configuredLabels []string) []string {
	// mac is always required as first label
	labels := []string{"mac"}

	// Add other configured labels in consistent order
	labelOrder := []string{"ap", "band", "wlan", "name", "username", "ipv4", "ipv6"}
	for _, label := range labelOrder {
		if contains(configuredLabels, label) && !contains(labels, label) {
			labels = append(labels, label)
		}
	}

	return labels
}
