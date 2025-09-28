// Package collector provides collectors for cisco-wnc-exporter.
package collector

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

const (
	labelMAC  = "mac"
	labelName = "name"
	labelBand = "band"
)

// APMetrics represents which AP metrics are enabled.
type APMetrics struct {
	General    bool
	Radio      bool
	Traffic    bool
	Errors     bool
	Info       bool
	InfoLabels []string
}

// APCollector implements prometheus.Collector for AP metrics from WNC.
type APCollector struct {
	metrics        APMetrics
	infoDesc       *prometheus.Desc
	infoLabelNames []string
	src            wnc.APSource
	rrmSrc         wnc.RRMSource
	clientSrc      wnc.ClientSource

	channelUtilizationDesc      *prometheus.Desc
	rxUtilizationDesc           *prometheus.Desc
	txUtilizationDesc           *prometheus.Desc
	noiseUtilizationDesc        *prometheus.Desc
	noiseFloorDesc              *prometheus.Desc
	txPowerDesc                 *prometheus.Desc
	channelDesc                 *prometheus.Desc
	channelWidthDesc            *prometheus.Desc
	associatedClientsDesc       *prometheus.Desc
	radioStateDesc              *prometheus.Desc
	adminStateDesc              *prometheus.Desc
	operStateDesc               *prometheus.Desc
	configStateDesc             *prometheus.Desc
	txPowerMaxDesc              *prometheus.Desc
	rxPacketsTotalDesc          *prometheus.Desc
	txPacketsTotalDesc          *prometheus.Desc
	rxBytesTotalDesc            *prometheus.Desc
	txBytesTotalDesc            *prometheus.Desc
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
	rxErrorsTotalDesc           *prometheus.Desc
	txErrorsTotalDesc           *prometheus.Desc
	txDropsTotalDesc            *prometheus.Desc
	txRetriesTotalDesc          *prometheus.Desc
	ackFailuresTotalDesc        *prometheus.Desc
	duplicateFramesTotalDesc    *prometheus.Desc
	fcsErrorsTotalDesc          *prometheus.Desc
	fragmentationRxTotalDesc    *prometheus.Desc
	fragmentationTxTotalDesc    *prometheus.Desc
	rtsFailuresTotalDesc        *prometheus.Desc
	decryptionErrorsTotalDesc   *prometheus.Desc
	micErrorsTotalDesc          *prometheus.Desc
	wepUndecryptableTotalDesc   *prometheus.Desc
	coverageHoleEventsDesc      *prometheus.Desc
	lastRadarOnRadioAtDesc      *prometheus.Desc
	radioResetTotalDesc         *prometheus.Desc
	cpuUsageCurrentDesc         *prometheus.Desc
	cpuUsageAverageDesc         *prometheus.Desc
	memoryUsageCurrentDesc      *prometheus.Desc
	memoryUsageAverageDesc      *prometheus.Desc
	uptimeSecondsDesc           *prometheus.Desc
}

// NewAPCollector creates a new AP collector.
func NewAPCollector(
	src wnc.APSource, rrmSrc wnc.RRMSource, clientSrc wnc.ClientSource, metrics APMetrics,
) *APCollector {
	baseRadioLabels := []string{labelMAC, "radio"}
	baseAPLabels := []string{labelMAC}

	collector := &APCollector{
		metrics:   metrics,
		src:       src,
		rrmSrc:    rrmSrc,
		clientSrc: clientSrc,
	}

	if metrics.Info {
		requiredLabels := []string{labelMAC, "radio"}
		availableLabels := []string{labelName, "ip", labelBand, "model", "serial", "sw_version", "eth_mac"}
		infoLabels := buildInfoLabels(requiredLabels, metrics.InfoLabels, availableLabels)
		collector.infoDesc = prometheus.NewDesc(
			"wnc_ap_info",
			"AP information labels for joining with other metrics, including device details",
			infoLabels, nil,
		)
		collector.infoLabelNames = infoLabels
	}

	if metrics.General {
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
		collector.configStateDesc = prometheus.NewDesc(
			"wnc_ap_config_state",
			"Configuration state (0=valid, 1=invalid) from IsApMisconfigured",
			baseAPLabels,
			nil,
		)
		collector.uptimeSecondsDesc = prometheus.NewDesc(
			"wnc_ap_uptime_seconds",
			"AP uptime in seconds",
			baseAPLabels,
			nil,
		)
	}

	if metrics.Radio {
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
		collector.txPowerMaxDesc = prometheus.NewDesc(
			"wnc_ap_tx_power_max_dbm",
			"Maximum TX power capability (dBm)",
			baseRadioLabels,
			nil,
		)
	}

	if metrics.Radio {
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
			"Channel noise floor (dBm)",
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

	if metrics.General {
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

	if metrics.General {
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
			"Total received packets",
			baseRadioLabels,
			nil,
		)
		collector.txPacketsTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_packets_total",
			"Total transmitted packets",
			baseRadioLabels,
			nil,
		)
		collector.rxBytesTotalDesc = prometheus.NewDesc(
			"wnc_ap_rx_bytes_total",
			"Total received bytes",
			baseRadioLabels,
			nil,
		)
		collector.txBytesTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_bytes_total",
			"Total transmitted bytes",
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
			"Total TX errors",
			baseRadioLabels,
			nil,
		)
		collector.txDropsTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_drops_total",
			"Total TX drops",
			baseRadioLabels,
			nil,
		)
		collector.txRetriesTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_retries_total",
			"Total TX retries",
			baseRadioLabels,
			nil,
		)
		collector.ackFailuresTotalDesc = prometheus.NewDesc(
			"wnc_ap_transmission_failures_total",
			"Failed transmission attempts",
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
			"Coverage hole events",
			baseRadioLabels,
			nil,
		)
		collector.lastRadarOnRadioAtDesc = prometheus.NewDesc(
			"wnc_ap_last_radar_on_radio_at",
			"Last radar detection timestamp (unix time, 0=never detected)",
			baseRadioLabels,
			nil,
		)
		collector.radioResetTotalDesc = prometheus.NewDesc(
			"wnc_ap_radio_reset_total",
			"Radio reset count",
			baseRadioLabels,
			nil,
		)
	}
	return collector
}

// Describe implements prometheus.Collector.
func (c *APCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.General {
		ch <- c.radioStateDesc
		ch <- c.adminStateDesc
		ch <- c.operStateDesc
		ch <- c.configStateDesc
		ch <- c.uptimeSecondsDesc
		ch <- c.cpuUsageCurrentDesc
		ch <- c.cpuUsageAverageDesc
		ch <- c.memoryUsageCurrentDesc
		ch <- c.memoryUsageAverageDesc
	}
	if c.metrics.Radio {
		ch <- c.channelDesc
		ch <- c.channelWidthDesc
		ch <- c.txPowerDesc
		ch <- c.txPowerMaxDesc
		ch <- c.noiseFloorDesc
		ch <- c.channelUtilizationDesc
		ch <- c.rxUtilizationDesc
		ch <- c.txUtilizationDesc
		ch <- c.noiseUtilizationDesc
		ch <- c.associatedClientsDesc
	}
	if c.metrics.Traffic {
		ch <- c.rxPacketsTotalDesc
		ch <- c.txPacketsTotalDesc
		ch <- c.rxBytesTotalDesc
		ch <- c.txBytesTotalDesc
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
	if c.metrics.Errors {
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
		ch <- c.lastRadarOnRadioAtDesc
		ch <- c.radioResetTotalDesc
	}
	if c.metrics.Info {
		ch <- c.infoDesc
	}
}

// Collect implements prometheus.Collector by retrieving AP data from WNC.
func (c *APCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if !c.isAnyMetricFlagEnabled() {
		return
	}

	var capwapMap map[string]ap.CAPWAPData
	capwapData, err := c.src.GetCAPWAPData(ctx)
	if err != nil {
		slog.Error("Failed to get CAPWAP data", "error", err)
	}
	capwapMap = buildCAPWAPMap(capwapData)

	var radioDataMap map[string]*ap.RadioOperData
	radioDataSlice, err := c.src.GetRadioData(ctx)
	if err != nil {
		slog.Error("Failed to get radio data", "error", err)
	}
	radioDataMap = buildRadioDataMap(radioDataSlice)

	var radioOperStatsMap map[string]map[int]ap.RadioOperStats
	if IsEnabled(c.metrics.Traffic, c.metrics.Errors) {
		radioOperStats, err := c.src.GetRadioOperStats(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve radio operational stats", "error", err)
		}
		radioOperStatsMap = buildRadioOperStatsMap(radioOperStats)
	}

	var apOperDataMap map[string]ap.OperData
	if IsEnabled(c.metrics.General) {
		apOperData, err := c.src.GetAPOperData(ctx)
		if err != nil {
			slog.Warn("Failed to retrieve AP operational data", "error", err)
		}
		apOperDataMap = buildAPOperDataMap(apOperData)
	}

	var rrmMeasurementsMap map[string]*rrm.RRMMeasurement
	var rrmCoverageMap map[string]*rrm.RRMCoverage
	var apDot11RadarMap map[string]*rrm.ApDot11RadarData
	if IsEnabled(c.metrics.Radio) {
		rrmMeasurements, err := c.rrmSrc.GetRRMMeasurements(ctx)
		if err != nil {
			slog.Warn("Failed to get RRM data for radio metrics", "error", err)
		}
		rrmMeasurementsMap = buildRRMMeasurementsMap(rrmMeasurements)
	}

	var radioResetStatsMap map[string]map[int]*ap.RadioResetStats
	if IsEnabled(c.metrics.Errors) {
		radioResetStats, err := c.src.GetRadioResetStats(ctx)
		if err != nil {
			slog.Warn("Failed to get radio reset stats for error metrics", "error", err)
		} else {
			radioResetStatsMap = buildRadioResetStatsMap(radioResetStats)
		}

		rrmCoverage, err := c.rrmSrc.GetRRMCoverage(ctx)
		if err != nil {
			slog.Warn("Failed to get RRM coverage for error metrics", "error", err)
		} else {
			rrmCoverageMap = buildRRMCoverageMap(rrmCoverage)
		}

		apDot11Radar, err := c.rrmSrc.GetApDot11RadarData(ctx)
		if err != nil {
			slog.Warn("Failed to get radar data for error metrics", "error", err)
		} else {
			apDot11RadarMap = buildApDot11RadarMap(apDot11Radar)
		}
	}

	if IsEnabled(c.metrics.General) {
		for wtpMAC := range capwapMap {
			c.collectSystemMetrics(ch, wtpMAC, capwapMap, apOperDataMap)
		}
	}

	for _, radio := range radioDataMap {
		if c.metrics.General {
			c.collectGeneralMetrics(ch, radio)
		}
		if c.metrics.Radio {
			c.collectRadioMetrics(ch, radio, rrmMeasurementsMap)
		}
		if c.metrics.Traffic {
			c.collectTrafficMetrics(ch, radio, radioOperStatsMap)
		}
		if c.metrics.Errors {
			c.collectErrorMetrics(
				ch, radio, radioOperStatsMap, radioResetStatsMap,
				rrmCoverageMap, apDot11RadarMap,
			)
		}
		if c.metrics.Info {
			c.collectInfoMetrics(ch, radio, capwapMap)
		}
	}
}

// collectSystemMetrics collects AP system metrics.
func (c *APCollector) collectSystemMetrics(
	ch chan<- prometheus.Metric,
	wtpMAC string,
	capwapMap map[string]ap.CAPWAPData,
	apOperDataMap map[string]ap.OperData,
) {
	labels := []string{wtpMAC}

	metrics := []Float64Metric{
		{c.configStateDesc, boolToFloat64(capwapMap[wtpMAC].TagInfo.IsApMisconfigured)},
		{c.uptimeSecondsDesc, float64(determineUptimeFromBootTime(capwapMap[wtpMAC].ApTimeInfo.BootTime))},
		{c.cpuUsageCurrentDesc, float64(apOperDataMap[wtpMAC].ApSysStats.CPUUsage)},
		{c.cpuUsageAverageDesc, float64(apOperDataMap[wtpMAC].ApSysStats.AvgCPUUsage)},
		{c.memoryUsageCurrentDesc, float64(apOperDataMap[wtpMAC].ApSysStats.MemoryUsage)},
		{c.memoryUsageAverageDesc, float64(apOperDataMap[wtpMAC].ApSysStats.AvgMemoryUsage)},
	}

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.GaugeValue, metric.Value, labels...)
	}
}

// collectGeneralMetrics collects general radio metrics.
func (c *APCollector) collectGeneralMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
) {
	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}

	metrics := []Float64Metric{
		{c.radioStateDesc, boolToFloat64(radio.OperState == APRadioStateUp)},
		{c.adminStateDesc, boolToFloat64(radio.AdminState == APAdminStateEnabled)},
		{c.operStateDesc, boolToFloat64(radio.OperState == APRadioStateUp)},
	}

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.GaugeValue, metric.Value, labels...)
	}
}

// collectRadioMetrics collects radio-specific metrics.
func (c *APCollector) collectRadioMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	rrmMeasurementsMap map[string]*rrm.RRMMeasurement,
) {
	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	radioID := radio.WtpMAC + ":" + strconv.Itoa(radio.RadioSlotID)

	var clientCount float64
	clientCountsMap := make(map[string]map[int]int)
	if apCounts, ok := clientCountsMap[radio.WtpMAC]; ok {
		if count, ok := apCounts[radio.RadioSlotID]; ok {
			clientCount = float64(count)
		}
	}

	metrics := []Float64Metric{
		{c.txPowerDesc, float64(radio.RadioBandInfo[0].PhyTxPwrLvlCfg.CfgData.CurrTxPowerInDbm)},
		{c.channelDesc, float64(radio.PhyHtCfg.CfgData.CurrFreq)},
		{c.channelWidthDesc, float64(radio.PhyHtCfg.CfgData.ChanWidth)},
		{c.txPowerMaxDesc, float64(radio.RadioBandInfo[0].PhyTxPwrLvlCfg.CfgData.TxPowerLevel1)},
		{c.channelUtilizationDesc, float64(rrmMeasurementsMap[radioID].Load.CcaUtilPercentage)},
		{c.rxUtilizationDesc, float64(rrmMeasurementsMap[radioID].Load.RxUtilPercentage)},
		{c.txUtilizationDesc, float64(rrmMeasurementsMap[radioID].Load.TxUtilPercentage)},
		{c.noiseUtilizationDesc, float64(rrmMeasurementsMap[radioID].Load.RxNoiseChannelUtilization)},
		{c.noiseFloorDesc, float64(rrmMeasurementsMap[radioID].Noise.Noise.NoiseData[0].Noise)},
		{c.associatedClientsDesc, clientCount},
	}

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.GaugeValue, metric.Value, labels...)
	}
}

func (c *APCollector) collectTrafficMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	radioOperStatsMap map[string]map[int]ap.RadioOperStats,
) {
	stats, ok := radioOperStatsMap[radio.WtpMAC][radio.RadioSlotID]
	if !ok {
		return
	}

	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	const averageFrameSize = 1500

	trafficMetrics := []Float64Metric{
		{c.rxPacketsTotalDesc, float64(stats.RxDataFrameCount)},
		{c.txPacketsTotalDesc, float64(stats.TxDataFrameCount)},
		{c.rxBytesTotalDesc, float64(stats.RxDataFrameCount * averageFrameSize)},
		{c.txBytesTotalDesc, float64(stats.TxDataFrameCount * averageFrameSize)},
		{c.dataRxFramesTotalDesc, float64(stats.RxDataFrameCount)},
		{c.dataTxFramesTotalDesc, float64(stats.TxDataFrameCount)},
		{c.managementRxFramesTotalDesc, float64(stats.RxMgmtFrameCount)},
		{c.managementTxFramesTotalDesc, float64(stats.TxMgmtFrameCount)},
		{c.controlRxFramesTotalDesc, float64(stats.RxCtrlFrameCount)},
		{c.controlTxFramesTotalDesc, float64(stats.TxCtrlFrameCount)},
		{c.multicastRxFramesTotalDesc, float64(stats.MulticastRxFrameCnt)},
		{c.multicastTxFramesTotalDesc, float64(stats.MulticastTxFrameCnt)},
		{c.totalTxFramesTotalDesc, float64(stats.TxFrameCount)},
		{c.rtsSuccessTotalDesc, float64(stats.RtsSuccessCount)},
	}

	for _, metric := range trafficMetrics {
		ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.CounterValue, metric.Value, labels...)
	}
}

func (c *APCollector) collectErrorMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	radioOperStatsMap map[string]map[int]ap.RadioOperStats,
	radioResetStatsMap map[string]map[int]*ap.RadioResetStats,
	rrmCoverageMap map[string]*rrm.RRMCoverage,
	apDot11RadarMap map[string]*rrm.ApDot11RadarData,
) {
	stats, ok := radioOperStatsMap[radio.WtpMAC][radio.RadioSlotID]
	if !ok {
		return
	}

	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	radioID := radio.WtpMAC + ":" + strconv.Itoa(radio.RadioSlotID)

	var radioResetCount float64
	if resetStats, exists := radioResetStatsMap[radio.WtpMAC][radio.RadioSlotID]; exists {
		radioResetCount = float64(resetStats.Count)
	}

	var coverageHoleEvents float64
	if coverage, exists := rrmCoverageMap[radioID]; exists {
		coverageHoleEvents = float64(coverage.FailedClientCount)
	}

	var lastRadarOnRadioAt float64
	if radar, exists := apDot11RadarMap[radioID]; exists {
		if !radar.LastRadarOnRadio.IsZero() && radar.LastRadarOnRadio.Year() > 1970 {
			lastRadarOnRadioAt = float64(radar.LastRadarOnRadio.Unix())
		}
	}

	errorMetrics := []Float64Metric{
		{c.rxErrorsTotalDesc, float64(stats.RxErrorFrameCount)},
		{c.txErrorsTotalDesc, float64(stats.FailedCount)},
		{c.txDropsTotalDesc, float64(stats.AckFailureCount)},
		{c.txRetriesTotalDesc, float64(stats.RetryCount)},
		{c.ackFailuresTotalDesc, float64(stats.FailedCount)},
		{c.duplicateFramesTotalDesc, float64(stats.FrameDuplicateCount)},
		{c.fcsErrorsTotalDesc, float64(stats.FcsErrorCount)},
		{c.fragmentationRxTotalDesc, float64(stats.RxFragmentCount)},
		{c.fragmentationTxTotalDesc, float64(stats.TxFragmentCount)},
		{c.rtsFailuresTotalDesc, float64(stats.RtsFailureCount)},
		{c.decryptionErrorsTotalDesc, float64(stats.MACDecryErrFrameCount)},
		{c.micErrorsTotalDesc, float64(stats.MACMicErrFrameCount)},
		{c.wepUndecryptableTotalDesc, float64(stats.WepUndecryptableCount)},
		{c.coverageHoleEventsDesc, coverageHoleEvents},
		{c.lastRadarOnRadioAtDesc, lastRadarOnRadioAt},
		{c.radioResetTotalDesc, radioResetCount},
	}

	for _, metric := range errorMetrics {
		ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.CounterValue, metric.Value, labels...)
	}
}

func (c *APCollector) collectInfoMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	capwapMap map[string]ap.CAPWAPData,
) {
	if c.infoDesc == nil {
		return
	}

	capwap, exists := capwapMap[radio.WtpMAC]
	if !exists {
		return
	}

	band := DetermineBandFromRadioInfo(radio.RadioSlotID, radio.RadioType)
	radioSlot := strconv.Itoa(radio.RadioSlotID)

	model := capwap.DeviceDetail.StaticInfo.ApModels.Model
	serial := capwap.DeviceDetail.StaticInfo.BoardData.WtpSerialNum
	swVersion := capwap.DeviceDetail.WtpVersion.SwVersion
	ethMAC := capwap.DeviceDetail.StaticInfo.BoardData.WtpEnetMAC

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
			values[i] = ""
		}
	}

	ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1, values...)
}

func buildCAPWAPMap(capwapData []ap.CAPWAPData) map[string]ap.CAPWAPData {
	capwapMap := make(map[string]ap.CAPWAPData)
	for _, capwap := range capwapData {
		capwapMap[capwap.WtpMAC] = capwap
	}
	return capwapMap
}

func buildRadioDataMap(radioData []ap.RadioOperData) map[string]*ap.RadioOperData {
	radioMap := make(map[string]*ap.RadioOperData)
	for i, radio := range radioData {
		key := radio.WtpMAC + ":" + strconv.Itoa(radio.RadioSlotID)
		radioMap[key] = &radioData[i]
	}
	return radioMap
}

func buildRadioOperStatsMap(radioOperStats []ap.RadioOperStats) map[string]map[int]ap.RadioOperStats {
	statsMap := make(map[string]map[int]ap.RadioOperStats)
	for _, stats := range radioOperStats {
		if statsMap[stats.ApMAC] == nil {
			statsMap[stats.ApMAC] = make(map[int]ap.RadioOperStats)
		}
		statsMap[stats.ApMAC][stats.SlotID] = stats
	}
	return statsMap
}

func buildAPOperDataMap(apOperData []ap.OperData) map[string]ap.OperData {
	operMap := make(map[string]ap.OperData)
	for _, oper := range apOperData {
		operMap[oper.WtpMAC] = oper
	}
	return operMap
}

func buildRRMMeasurementsMap(measurements []rrm.RRMMeasurement) map[string]*rrm.RRMMeasurement {
	measurementMap := make(map[string]*rrm.RRMMeasurement)
	for i, measurement := range measurements {
		key := measurement.WtpMAC + ":" + strconv.Itoa(measurement.RadioSlotID)
		measurementMap[key] = &measurements[i]
	}
	return measurementMap
}

func buildRadioResetStatsMap(radioResetStats []ap.RadioResetStats) map[string]map[int]*ap.RadioResetStats {
	statsMap := make(map[string]map[int]*ap.RadioResetStats)
	for i, stats := range radioResetStats {
		if statsMap[stats.ApMAC] == nil {
			statsMap[stats.ApMAC] = make(map[int]*ap.RadioResetStats)
		}
		statsMap[stats.ApMAC][stats.RadioID] = &radioResetStats[i]
	}
	return statsMap
}

func buildRRMCoverageMap(rrmCoverage []rrm.RRMCoverage) map[string]*rrm.RRMCoverage {
	coverageMap := make(map[string]*rrm.RRMCoverage)
	for i, coverage := range rrmCoverage {
		key := coverage.WtpMAC + ":" + strconv.Itoa(coverage.RadioSlotID)
		coverageMap[key] = &rrmCoverage[i]
	}
	return coverageMap
}

func buildApDot11RadarMap(apDot11Radar []rrm.ApDot11RadarData) map[string]*rrm.ApDot11RadarData {
	radarMap := make(map[string]*rrm.ApDot11RadarData)
	for i, radar := range apDot11Radar {
		key := radar.WtpMAC + ":" + strconv.Itoa(radar.RadioSlotID)
		radarMap[key] = &apDot11Radar[i]
	}
	return radarMap
}

// determineUptimeFromBootTime determines uptime from boot time timestamp.
func determineUptimeFromBootTime(bootTimeStr string) int64 {
	if bootTimeStr == "" {
		return 0
	}

	bootTime, err := time.Parse(time.RFC3339, bootTimeStr)
	if err != nil {
		return 0
	}

	uptime := time.Since(bootTime)
	return int64(uptime.Seconds())
}

func (c *APCollector) isAnyMetricFlagEnabled() bool {
	return IsEnabled(c.metrics.General, c.metrics.Radio, c.metrics.Traffic, c.metrics.Errors, c.metrics.Info)
}

// CalculateUptimeFromBootTime calculates uptime in seconds from boot time ISO 8601 timestamp.
func CalculateUptimeFromBootTime(bootTimeStr string) (int64, error) {
	bootTime, err := time.Parse(time.RFC3339, bootTimeStr)
	if err != nil {
		return 0, err
	}

	uptime := time.Since(bootTime)
	return int64(uptime.Seconds()), nil
}
