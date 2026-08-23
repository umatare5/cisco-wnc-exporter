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
	roamTypeDesc               *prometheus.Desc
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

	labels := []string{labelMAC}

	if metrics.General {
		collector.stateDesc = prometheus.NewDesc(
			"wnc_client_state",
			"Client connection state, as the value the controller's own enumeration assigns "+
				"its spelling. 11 is client-status-run, the state a client holds while it "+
				"passes traffic. The numbering follows the onboarding sequence, so a value "+
				"below 11 has not reached it and a value above 11 is a deletion",
			labels, nil,
		)
		collector.roamTypeDesc = prometheus.NewDesc(
			"wnc_client_roam_type",
			"How the client reached the association it currently holds, as the value the "+
				"controller's own enumeration assigns its spelling (0=dot11-roam-type-none, "+
				"1=dot11-roam-type-slow-11i, 2=dot11-roam-type-fast-okc, 3=dot11-roam-type-cckm, "+
				"4=dot11-roam-type-fast-11r). It is a property of that association rather than "+
				"a count, so it does not move until the client associates again",
			labels, nil,
		)
		collector.associationUptimeDesc = prometheus.NewDesc(
			"wnc_client_uptime_seconds",
			"Client association uptime in seconds. Withheld rather than measured from a "+
				"zero timestamp when the controller reports no association time, so a "+
				"session-age check has no reading instead of a false one",
			labels, nil,
		)
		collector.stateTransitionSecondsDesc = prometheus.NewDesc(
			"wnc_client_state_transition_seconds",
			"Client state transition latency in seconds",
			labels, nil,
		)
		collector.powerSaveStateDesc = prometheus.NewDesc(
			"wnc_client_power_save_state",
			"Power save state as reported",
			labels, nil,
		)
	}

	if metrics.Radio {
		collector.protocolDesc = prometheus.NewDesc(
			"wnc_client_protocol",
			"Client wireless protocol (0=unknown, 1=802.11a, 2=802.11b, 3=802.11g, "+
				"4=802.11n, 5=802.11ac, 6=802.11ax, 7=802.11be)",
			labels, nil,
		)
		collector.mcsIndexDesc = prometheus.NewDesc(
			"wnc_client_mcs_index",
			"Client MCS index, -1 when the rate carries none",
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
			"wnc_client_rx_group_total",
			"RX group counter (rx-group-counter)",
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
			[]string{
				labelAP, labelBand, labelWLAN, labelWLANID, labelName, labelDeviceType,
				labelUsername, labelIPv4, labelIPv6,
			},
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
		ch <- c.roamTypeDesc
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
		slog.Debug("Failed to retrieve client data", "error", err)
		return
	}

	var deviceMap map[string]client.DcInfo
	if IsEnabled(c.metrics.Info) {
		deviceData, err := c.src.GetDeviceData(ctx)
		if err != nil {
			slog.Debug("Failed to retrieve device data", "error", err)
		}
		deviceMap = buildDeviceMap(deviceData)
	}

	var dot11Map map[string]client.Dot11OperData
	if IsEnabled(c.metrics.General, c.metrics.Radio, c.metrics.Info) {
		dot11Data, err := c.src.GetDot11Data(ctx)
		if err != nil {
			slog.Debug("Failed to retrieve dot11 data", "error", err)
		}
		dot11Map = buildDot11Map(dot11Data)
	}

	var sisfMap map[string]client.SisfDBMac
	if IsEnabled(c.metrics.Info) {
		sisfdbData, err := c.src.GetSISFDBData(ctx)
		if err != nil {
			slog.Debug("Failed to retrieve SISF database data", "error", err)
		}
		sisfMap = buildSISFMap(sisfdbData)
	}

	var trafficMap map[string]client.TrafficStats
	if IsEnabled(c.metrics.General, c.metrics.Radio, c.metrics.Traffic, c.metrics.Errors) {
		trafficStats, err := c.src.GetTrafficStats(ctx)
		if err != nil {
			slog.Debug("Failed to retrieve traffic stats", "error", err)
		}
		trafficMap = buildTrafficMap(trafficStats)
	}

	var mobilityMap map[string]client.MmIfClientHistory
	if IsEnabled(c.metrics.General) {
		mobilityData, err := c.src.GetMobilityHistory(ctx)
		if err != nil {
			slog.Debug("Failed to retrieve mobility history data", "error", err)
		}
		mobilityMap = buildMobilityMap(mobilityData)
	}

	for _, data := range clientData {
		// A client the controller holds short of the run state is the failure an
		// operator most needs to see, and the filter below would drop it.
		if c.metrics.General {
			emitEnumReading(ch, c.stateDesc, clientStates, data.CoState, data.ClientMAC)
		}

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

	// An absent key means either the fetch failed or the controller has not
	// reported this client in that data set yet. Emitting the zero value would
	// read as a measurement: an association uptime taken from a zero timestamp, or
	// a power save state the client never reported.
	//
	// The dot11 record can also arrive without ms-assoc-time, which decodes to the
	// zero time in year 1 rather than to an absent key, so the map guard alone lets
	// that uptime through: time.Since saturates the Duration there and reports some
	// 9.2e9 seconds. The comparison is the one emitTimestamp makes, so the epoch
	// sentinel is withheld here as well.
	var metrics []Float64Metric
	if dot11, ok := dot11Map[data.ClientMAC]; ok && dot11.MsAssocTime.Year() > epochYear {
		metrics = append(metrics,
			Float64Metric{c.associationUptimeDesc, time.Since(dot11.MsAssocTime).Seconds()})
	}
	if latency, ok := determineLastRunLatency(mobilityMap[data.ClientMAC]); ok {
		metrics = append(metrics, Float64Metric{c.stateTransitionSecondsDesc, latency})
	}
	emitEnumReading(ch, c.roamTypeDesc, clientRoamTypes,
		firstRoamType(mobilityMap[data.ClientMAC]), data.ClientMAC)
	if traffic, ok := trafficMap[data.ClientMAC]; ok {
		metrics = append(metrics,
			Float64Metric{c.powerSaveStateDesc, float64(traffic.PowerSaveState)})
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

	// Zero is a plausible reading for every one of these, which is why they must
	// be absent rather than zero when their source is missing: 0 dBm RSSI reads as
	// a perfect signal, and a zero protocol reads as a valid enum.
	var metrics []Float64Metric
	if dot11, ok := dot11Map[data.ClientMAC]; ok {
		metrics = append(metrics, Float64Metric{
			c.protocolDesc,
			float64(MapWirelessProtocol(dot11.EwlcMsPhyType, dot11.RadioType)),
		})
	}
	if traffic, ok := trafficMap[data.ClientMAC]; ok {
		metrics = append(metrics,
			Float64Metric{c.speedDesc, float64(traffic.Speed)},
			Float64Metric{c.spatialStreamsDesc, float64(traffic.SpatialStream)},
			Float64Metric{c.mcsIndexDesc, float64(parseMCSIndex(traffic.CurrentRate))},
			Float64Metric{c.rssiDesc, float64(traffic.MostRecentRSSI)},
			Float64Metric{c.snrDesc, float64(traffic.MostRecentSNR)},
		)
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
	// Emitting zero counters for a client whose traffic stats are missing makes
	// the next successful scrape look like a counter reset, which rate() reports
	// as a spike of the full counter value.
	traffic, ok := trafficMap[data.ClientMAC]
	if !ok {
		return
	}

	labels := []string{data.ClientMAC}

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
	traffic, ok := trafficMap[data.ClientMAC]
	if !ok {
		return
	}

	labels := []string{data.ClientMAC}

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

	band := ClientBand(data)
	name := determineDeviceNameFromDeviceMap(deviceMap, data.ClientMAC)
	deviceType := determineDeviceTypeFromDeviceMap(deviceMap, data.ClientMAC)
	wlan := dot11.VapSsid
	// The identifier comes from the record being iterated rather than from the dot11 map,
	// which can miss, and it is the same leaf wnc_wlan_clients buckets by, so a
	// sum by (wlan_id) over these reconciles with that series by construction.
	wlanID := determineWLANIDLabel(data.WlanID)
	ipv4 := determineIPv4FromSISF(sisfMap[data.ClientMAC])
	ipv6 := determineIPv6FromSISF(sisfMap[data.ClientMAC])

	// Build dynamic labels based on configured labels
	values := make([]string, len(c.infoLabelNames))
	for i, label := range c.infoLabelNames {
		switch label {
		case labelMAC:
			values[i] = data.ClientMAC
		case labelAP:
			values[i] = data.ApName
		case labelBand:
			values[i] = band
		case labelWLAN:
			values[i] = wlan
		case labelWLANID:
			values[i] = wlanID
		case labelDeviceType:
			values[i] = deviceType
		case labelName:
			values[i] = name
		case labelUsername:
			values[i] = data.Username
		case labelIPv4:
			values[i] = ipv4
		case labelIPv6:
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

// determineDeviceTypeFromDeviceMap extracts the device class the controller assigned.
//
// The class is the controller's own conclusion rather than something the client claims, and
// the confidence leaf beside it read 0 on most records measured, so read a change here as the
// controller reclassifying rather than as the device changing.
func determineDeviceTypeFromDeviceMap(deviceMap map[string]client.DcInfo, mac string) string {
	if device, ok := deviceMap[mac]; ok {
		return device.DeviceType
	}
	return ""
}

// determineWLANIDLabel renders the WLAN identifier as a label value, and empty for the zero
// the leaf decodes to when the controller omits it. No WLAN carries identifier 0, so a "0"
// here would name a WLAN that cannot exist.
func determineWLANIDLabel(wlanID int) string {
	if wlanID == 0 {
		return ""
	}
	return strconv.Itoa(wlanID)
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

// firstRoamType reports how the client reached the association it currently holds, and
// reports an empty spelling when the controller records none.
//
// The history is ordered by association time, newest first, so the first entry is the
// current association. The list carries no key and declares no order, which is why the
// ordering is measured rather than modeled.
func firstRoamType(mobility client.MmIfClientHistory) string {
	entries := mobility.MobilityHistory.Entry
	if len(entries) == 0 {
		return ""
	}

	return entries[0].Dot11RoamType
}

// determineLastRunLatency reports the run latency the controller recorded for the
// association the client currently holds, and reports false when it recorded none.
//
// Both shapes the controller uses for "no measurement" are withheld: an entry list that
// is empty, and a first entry whose latency reads zero. Zero would publish an instant
// transition, which is a measurement rather than the absence of one.
func determineLastRunLatency(mobility client.MmIfClientHistory) (float64, bool) {
	const millisecondsToSeconds = 1000.0

	entries := mobility.MobilityHistory.Entry
	if len(entries) == 0 || entries[0].RunLatency == 0 {
		return 0, false
	}

	return float64(entries[0].RunLatency) / millisecondsToSeconds, true
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
