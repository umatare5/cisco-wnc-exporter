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
	General    bool
	Radio      bool
	Traffic    bool
	Errors     bool
	Info       bool
	InfoLabels []string
}

// ClientCollector implements prometheus.Collector for Client metrics.
type ClientCollector struct {
	metrics        ClientMetrics
	infoDesc       *prometheus.Desc
	infoLabelNames []string
	src            wnc.ClientSource

	stateDesc                  *prometheus.Desc
	associationUptimeDesc      *prometheus.Desc
	stateTransitionSecondsDesc *prometheus.Desc
	protocolDesc               *prometheus.Desc
	mcsIndexDesc               *prometheus.Desc
	spatialStreamsDesc         *prometheus.Desc
	speedDesc                  *prometheus.Desc
	rssiDesc                   *prometheus.Desc
	snrDesc                    *prometheus.Desc
	bytesRxDesc                *prometheus.Desc
	bytesTxDesc                *prometheus.Desc
	packetsRxDesc              *prometheus.Desc
	packetsTxDesc              *prometheus.Desc
	retryRatioDesc             *prometheus.Desc
	policyErrorsDesc           *prometheus.Desc
	duplicateReceivedDesc      *prometheus.Desc
	decryptionFailedDesc       *prometheus.Desc
	micMismatchDesc            *prometheus.Desc
	micMissingDesc             *prometheus.Desc
	excessiveRetriesDesc       *prometheus.Desc
	rxGroupCounterDesc         *prometheus.Desc
	txTotalDropsDesc           *prometheus.Desc
	dataRetriesDesc            *prometheus.Desc
	rtsRetriesDesc             *prometheus.Desc
	txRetriesDesc              *prometheus.Desc
	powerSaveStateDesc         *prometheus.Desc
}

// NewClientCollector creates a new ClientCollector for retrieving metrics from WNC.
func NewClientCollector(src wnc.ClientSource, metrics ClientMetrics) *ClientCollector {
	collector := &ClientCollector{
		src:     src,
		metrics: metrics,
	}

	labels := []string{"mac"}

	if metrics.General {
		collector.stateDesc = prometheus.NewDesc(
			"wnc_client_state",
			"Client connection state",
			labels, nil,
		)
		collector.associationUptimeDesc = prometheus.NewDesc(
			"wnc_client_uptime_seconds",
			"Client association uptime in seconds",
			labels, nil,
		)
		collector.stateTransitionSecondsDesc = prometheus.NewDesc(
			"wnc_client_state_transition_seconds",
			"Client state transition latency in seconds",
			labels, nil,
		)
		collector.powerSaveStateDesc = prometheus.NewDesc(
			"wnc_client_power_save_state",
			"Power save state (0=active, 1=power-save)",
			labels, nil,
		)
	}

	if metrics.Radio {
		collector.protocolDesc = prometheus.NewDesc(
			"wnc_client_protocol",
			"Client wireless protocol (1=802.11a, 2=802.11b, 3=802.11g, 4=802.11n, 5=802.11ac, 6=802.11ax)",
			labels, nil,
		)
		collector.mcsIndexDesc = prometheus.NewDesc(
			"wnc_client_mcs_index",
			"Client MCS index (-1=legacy, 0-11=802.11n/ac/ax MCS values)",
			labels, nil,
		)
		collector.spatialStreamsDesc = prometheus.NewDesc(
			"wnc_client_spatial_streams",
			"Number of spatial streams",
			labels, nil,
		)
		collector.speedDesc = prometheus.NewDesc(
			"wnc_client_speed_mbps",
			"Connection speed in Mbps",
			labels, nil,
		)
		collector.rssiDesc = prometheus.NewDesc(
			"wnc_client_rssi_dbm",
			"Received signal strength in dBm",
			labels, nil,
		)
		collector.snrDesc = prometheus.NewDesc(
			"wnc_client_snr_decibels",
			"Signal-to-noise ratio in dB",
			labels, nil,
		)
	}

	if metrics.Traffic {
		collector.bytesRxDesc = prometheus.NewDesc(
			"wnc_client_rx_bytes_total",
			"Total bytes received",
			labels, nil,
		)
		collector.bytesTxDesc = prometheus.NewDesc(
			"wnc_client_tx_bytes_total",
			"Total bytes transmitted",
			labels, nil,
		)
		collector.packetsRxDesc = prometheus.NewDesc(
			"wnc_client_rx_packets_total",
			"Total packets received",
			labels, nil,
		)
		collector.packetsTxDesc = prometheus.NewDesc(
			"wnc_client_tx_packets_total",
			"Total packets transmitted",
			labels, nil,
		)
	}

	if metrics.Errors {
		collector.retryRatioDesc = prometheus.NewDesc(
			"wnc_client_retry_ratio_percent",
			"Retry rate percentage",
			labels, nil,
		)
		collector.policyErrorsDesc = prometheus.NewDesc(
			"wnc_client_policy_errors_total",
			"Policy errors",
			labels, nil,
		)
		collector.duplicateReceivedDesc = prometheus.NewDesc(
			"wnc_client_duplicate_received_total",
			"Duplicate packets received",
			labels, nil,
		)
		collector.decryptionFailedDesc = prometheus.NewDesc(
			"wnc_client_decryption_failed_total",
			"Decryption failed packets",
			labels, nil,
		)
		collector.micMismatchDesc = prometheus.NewDesc(
			"wnc_client_mic_mismatch_total",
			"MIC mismatch errors",
			labels, nil,
		)
		collector.micMissingDesc = prometheus.NewDesc(
			"wnc_client_mic_missing_total",
			"MIC missing errors",
			labels, nil,
		)
		collector.excessiveRetriesDesc = prometheus.NewDesc(
			"wnc_client_excessive_retries_total",
			"Excessive retries",
			labels, nil,
		)
		collector.rxGroupCounterDesc = prometheus.NewDesc(
			"wnc_client_rx_group_counter_total",
			"RX group counter",
			labels, nil,
		)
		collector.txTotalDropsDesc = prometheus.NewDesc(
			"wnc_client_tx_drops_total",
			"Total TX drops",
			labels, nil,
		)
		collector.dataRetriesDesc = prometheus.NewDesc(
			"wnc_client_data_retries_total",
			"Data retries",
			labels, nil,
		)
		collector.rtsRetriesDesc = prometheus.NewDesc(
			"wnc_client_rts_retries_total",
			"RTS retries",
			labels, nil,
		)
		collector.txRetriesDesc = prometheus.NewDesc(
			"wnc_client_tx_retries_total",
			"TX retries",
			labels, nil,
		)
	}

	if metrics.Info {
		infoLabels := buildInfoLabels(
			labels,
			metrics.InfoLabels,
			[]string{"ap", "band", "wlan", "name", "username", "ipv4", "ipv6"},
		)
		collector.infoDesc = prometheus.NewDesc(
			"wnc_client_info",
			"Client information labels for joining with other metrics",
			infoLabels, nil,
		)
		collector.infoLabelNames = infoLabels
	}

	return collector
}

func (c *ClientCollector) isAnyMetricFlagEnabled() bool {
	return IsEnabled(c.metrics.General, c.metrics.Radio, c.metrics.Traffic, c.metrics.Errors, c.metrics.Info)
}

// Describe sends the descriptors of all metrics to the provided channel.
func (c *ClientCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.General {
		ch <- c.stateDesc
		ch <- c.associationUptimeDesc
		ch <- c.stateTransitionSecondsDesc
		ch <- c.powerSaveStateDesc
	}
	if c.metrics.Info {
		ch <- c.infoDesc
	}
	if c.metrics.Radio {
		ch <- c.protocolDesc
		ch <- c.mcsIndexDesc
		ch <- c.spatialStreamsDesc
		ch <- c.speedDesc
		ch <- c.rssiDesc
		ch <- c.snrDesc
	}
	if c.metrics.Traffic {
		ch <- c.bytesRxDesc
		ch <- c.bytesTxDesc
		ch <- c.packetsRxDesc
		ch <- c.packetsTxDesc
	}
	if c.metrics.Errors {
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
}

// Collect implements the prometheus.Collector interface.
func (c *ClientCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if !c.isAnyMetricFlagEnabled() {
		return
	}

	clientData, err := c.src.GetClientData(ctx)
	if err != nil {
		slog.Warn("Failed to retrieve client data", "error", err)
		return
	}

	var deviceMap map[string]client.DcInfo
	if IsEnabled(c.metrics.Info) {
		deviceData, err := c.src.GetDeviceData(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve device data", "error", err)
		}
		deviceMap = buildDeviceMap(deviceData)
	}

	var dot11Map map[string]client.Dot11OperData
	if IsEnabled(c.metrics.General, c.metrics.Radio, c.metrics.Info) {
		dot11Data, err := c.src.GetDot11Data(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve dot11 data", "error", err)
		}
		dot11Map = buildDot11Map(dot11Data)
	}

	var sisfMap map[string]client.SisfDBMac
	if IsEnabled(c.metrics.Info) {
		sisfdbData, err := c.src.GetSISFDBData(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve SISF database data", "error", err)
		}
		sisfMap = buildSISFMap(sisfdbData)
	}

	var trafficMap map[string]client.TrafficStats
	if IsEnabled(c.metrics.General, c.metrics.Radio, c.metrics.Traffic, c.metrics.Errors) {
		trafficStats, err := c.src.GetTrafficStats(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve traffic stats", "error", err)
		}
		trafficMap = buildTrafficMap(trafficStats)
	}

	var mobilityMap map[string]client.MmIfClientHistory
	if IsEnabled(c.metrics.General) {
		mobilityData, err := c.src.GetMobilityHistory(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve mobility history data", "error", err)
		}
		mobilityMap = buildMobilityMap(mobilityData)
	}

	for _, data := range clientData {
		if data.CoState != ClientStatusRun {
			continue
		}
		if c.metrics.General {
			c.collectGeneralMetrics(ch, data, trafficMap, dot11Map, mobilityMap)
		}
		if c.metrics.Radio {
			c.collectRadioMetrics(ch, data, trafficMap, dot11Map)
		}
		if c.metrics.Traffic {
			c.collectTrafficMetrics(ch, data, trafficMap)
		}
		if c.metrics.Errors {
			c.collectErrorMetrics(ch, data, trafficMap)
		}
		if c.metrics.Info {
			c.collectInfoMetrics(ch, data, dot11Map, deviceMap, sisfMap)
		}
	}
}

// collectGeneralMetrics collects general client metrics.
func (c *ClientCollector) collectGeneralMetrics(
	ch chan<- prometheus.Metric,
	data client.CommonOperData,
	trafficMap map[string]client.TrafficStats,
	dot11Map map[string]client.Dot11OperData,
	mobilityMap map[string]client.MmIfClientHistory,
) {
	labels := []string{data.ClientMAC}
	traffic := trafficMap[data.ClientMAC]
	dot11 := dot11Map[data.ClientMAC]

	metrics := []Float64Metric{
		{c.stateDesc, float64(MapClientState(data.CoState))},
		{c.associationUptimeDesc, time.Since(dot11.MsAssocTime).Seconds()},
		{c.stateTransitionSecondsDesc, determineLastRunLatency(mobilityMap, data.ClientMAC)},
		{c.powerSaveStateDesc, float64(traffic.PowerSaveState)},
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

// collectRadioMetrics collects radio-related metrics.
func (c *ClientCollector) collectRadioMetrics(
	ch chan<- prometheus.Metric,
	data client.CommonOperData,
	trafficMap map[string]client.TrafficStats,
	dot11Map map[string]client.Dot11OperData,
) {
	labels := []string{data.ClientMAC}
	traffic := trafficMap[data.ClientMAC]
	dot11 := dot11Map[data.ClientMAC]

	metrics := []Float64Metric{
		{c.protocolDesc, float64(MapWirelessProtocol(dot11.EwlcMsPhyType, dot11.RadioType, dot11.Is11GClient))},
		{c.speedDesc, float64(traffic.Speed)},
		{c.spatialStreamsDesc, float64(traffic.SpatialStream)},
		{c.mcsIndexDesc, float64(parseMCSIndex(traffic.CurrentRate))},
		{c.rssiDesc, float64(traffic.MostRecentRSSI)},
		{c.snrDesc, float64(traffic.MostRecentSNR)},
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

// collectTrafficMetrics collects traffic-related metrics.
func (c *ClientCollector) collectTrafficMetrics(
	ch chan<- prometheus.Metric,
	data client.CommonOperData,
	trafficMap map[string]client.TrafficStats,
) {
	labels := []string{data.ClientMAC}
	traffic := trafficMap[data.ClientMAC]

	metrics := []StringMetric{
		{c.bytesRxDesc, traffic.BytesRx},
		{c.bytesTxDesc, traffic.BytesTx},
		{c.packetsRxDesc, traffic.PktsRx},
		{c.packetsTxDesc, traffic.PktsTx},
	}

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			prometheus.CounterValue,
			float64(stringToUint64(metric.Value)),
			labels...,
		)
	}
}

// collectErrorMetrics collects error-related metrics.
func (c *ClientCollector) collectErrorMetrics(
	ch chan<- prometheus.Metric,
	data client.CommonOperData,
	trafficMap map[string]client.TrafficStats,
) {
	labels := []string{data.ClientMAC}
	traffic := trafficMap[data.ClientMAC]

	dataRetries := stringToUint64(traffic.DataRetries)
	txRetries := stringToUint64(traffic.TxRetries)
	packetsTx := stringToUint64(traffic.PktsTx)
	retryRatio := (float64(dataRetries+txRetries) / float64(packetsTx)) * 100.0

	ch <- prometheus.MustNewConstMetric(
		c.retryRatioDesc,
		prometheus.GaugeValue,
		retryRatio,
		labels...,
	)

	metrics := []StringMetric{
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

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			prometheus.CounterValue,
			float64(stringToUint64(metric.Value)),
			labels...,
		)
	}
}

// collectInfoMetrics collects client info metrics.
func (c *ClientCollector) collectInfoMetrics(
	ch chan<- prometheus.Metric,
	data client.CommonOperData,
	dot11Map map[string]client.Dot11OperData,
	deviceMap map[string]client.DcInfo,
	sisfMap map[string]client.SisfDBMac,
) {
	dot11 := dot11Map[data.ClientMAC]

	band := DetermineBandFromRadioInfo(-1, data.MsRadioType)
	name := determineDeviceNameFromDeviceMap(deviceMap, data.ClientMAC)
	wlan := dot11.VapSsid
	ipv4 := determineIPv4FromSISF(sisfMap[data.ClientMAC])
	ipv6 := determineIPv6FromSISF(sisfMap[data.ClientMAC])

	// Build dynamic labels based on configured labels
	values := make([]string, len(c.infoLabelNames))
	for i, labelName := range c.infoLabelNames {
		switch labelName {
		case "mac":
			values[i] = data.ClientMAC
		case "ap":
			values[i] = data.ApName
		case "band":
			values[i] = band
		case "wlan":
			values[i] = wlan
		case "name":
			values[i] = name
		case "username":
			values[i] = data.Username
		case "ipv4":
			values[i] = ipv4
		case "ipv6":
			values[i] = ipv6
		default:
			values[i] = ""
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.infoDesc,
		prometheus.GaugeValue,
		1, // Emit info metric (always value 1)
		values...,
	)
}

func buildDeviceMap(deviceData []client.DcInfo) map[string]client.DcInfo {
	deviceMap := make(map[string]client.DcInfo)
	for _, device := range deviceData {
		if device.ClientMAC != "" {
			deviceMap[device.ClientMAC] = device
		}
	}
	return deviceMap
}

func buildDot11Map(dot11Data []client.Dot11OperData) map[string]client.Dot11OperData {
	dot11Map := make(map[string]client.Dot11OperData)
	for _, dot11 := range dot11Data {
		if dot11.MsMACAddress != "" {
			dot11Map[dot11.MsMACAddress] = dot11
		}
	}
	return dot11Map
}

func buildSISFMap(sisfdbData []client.SisfDBMac) map[string]client.SisfDBMac {
	sisfMap := make(map[string]client.SisfDBMac)
	for _, sisf := range sisfdbData {
		if sisf.MACAddr != "" {
			sisfMap[sisf.MACAddr] = sisf
		}
	}
	return sisfMap
}

func buildTrafficMap(trafficStats []client.TrafficStats) map[string]client.TrafficStats {
	trafficMap := make(map[string]client.TrafficStats)
	for _, traffic := range trafficStats {
		if traffic.MsMACAddress != "" {
			trafficMap[traffic.MsMACAddress] = traffic
		}
	}
	return trafficMap
}

func buildMobilityMap(mobilityData []client.MmIfClientHistory) map[string]client.MmIfClientHistory {
	mobilityMap := make(map[string]client.MmIfClientHistory)
	for _, mobility := range mobilityData {
		if mobility.ClientMAC != "" {
			mobilityMap[mobility.ClientMAC] = mobility
		}
	}
	return mobilityMap
}

// determineDeviceNameFromDeviceMap extracts device name from device classification data.
func determineDeviceNameFromDeviceMap(deviceMap map[string]client.DcInfo, mac string) string {
	if device, ok := deviceMap[mac]; ok {
		return device.DeviceName
	}
	return ""
}

// determineIPv4FromSISF extracts IPv4 address from SISF data.
func determineIPv4FromSISF(sisf client.SisfDBMac) string {
	if sisf.Ipv4Binding.IPKey.IPAddr != "" {
		return sisf.Ipv4Binding.IPKey.IPAddr
	}
	return ""
}

// determineIPv6FromSISF extracts the first non-link-local IPv6 address from SISF data.
func determineIPv6FromSISF(sisf client.SisfDBMac) string {
	for _, ipv6Binding := range sisf.Ipv6Binding {
		if ipv6Binding.Ipv6BindingIPKey.IPAddr != "" &&
			!strings.HasPrefix(ipv6Binding.Ipv6BindingIPKey.IPAddr, "fe80:") {
			return ipv6Binding.Ipv6BindingIPKey.IPAddr
		}
	}
	return ""
}

// determineLastRunLatency extracts state transition latency from mobility map.
func determineLastRunLatency(mobilityMap map[string]client.MmIfClientHistory, clientMAC string) float64 {
	const millisecondsToSeconds = 1000.0

	if mobility, ok := mobilityMap[clientMAC]; ok {
		latencyMs := mobility.MobilityHistory.Entry[0].RunLatency

		return float64(latencyMs) / millisecondsToSeconds
	}
	return 0.0
}

// parseMCSIndex extracts MCS index from current-rate string.
func parseMCSIndex(currentRate string) int {
	if currentRate == "" {
		return -1
	}

	if strings.HasPrefix(currentRate, "m") {
		// Extract number after 'm'
		parts := strings.Fields(currentRate)
		mcsStr := strings.TrimPrefix(parts[0], "m")
		if mcs, err := strconv.Atoi(mcsStr); err == nil {
			return mcs
		}
	}

	return -1
}
