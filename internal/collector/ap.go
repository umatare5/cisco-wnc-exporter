// Package collector provides collectors for cisco-wnc-exporter.
package collector

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
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
	cpuUtilizationDesc          *prometheus.Desc
	memoryUtilizationDesc       *prometheus.Desc
	uptimeSecondsDesc           *prometheus.Desc
}

// NewAPCollector creates a new AP collector.
func NewAPCollector(
	src wnc.APSource, rrmSrc wnc.RRMSource, clientSrc wnc.ClientSource, metrics APMetrics,
) *APCollector {
	baseRadioLabels := []string{labelMAC, labelRadio}
	baseAPLabels := []string{labelMAC}

	collector := &APCollector{
		metrics:   metrics,
		src:       src,
		rrmSrc:    rrmSrc,
		clientSrc: clientSrc,
	}

	if metrics.Info {
		requiredLabels := []string{labelMAC, labelRadio}
		availableLabels := []string{labelName, labelIP, labelBand, labelModel, labelSerial, labelSWVersion, labelEthMAC}
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
			"AP operational state reported in the state label, always 1",
			[]string{labelMAC, labelState},
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
			"wnc_ap_channel_utilization_ratio",
			"Channel utilization ratio (CCA-based, 0-1)",
			baseRadioLabels,
			nil,
		)
		collector.rxUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_rx_utilization_ratio",
			"RX utilization ratio (0-1)",
			baseRadioLabels,
			nil,
		)
		collector.txUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_tx_utilization_ratio",
			"TX utilization ratio (0-1)",
			baseRadioLabels,
			nil,
		)
		collector.noiseUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_noise_utilization_ratio",
			"Noise channel utilization ratio (0-1)",
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
			"wnc_ap_clients",
			"Number of clients in the run state on this radio",
			baseRadioLabels,
			nil,
		)
	}

	if metrics.General {
		collector.cpuUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_cpu_utilization_ratio",
			"CPU utilization ratio (0-1)",
			baseAPLabels,
			nil,
		)
		collector.memoryUtilizationDesc = prometheus.NewDesc(
			"wnc_ap_memory_utilization_ratio",
			"Memory utilization ratio (0-1)",
			baseAPLabels,
			nil,
		)
	}

	if metrics.Traffic {
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
			"TX frames as the controller counts them, not the sum of the per-type series",
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
			"wnc_ap_coverage_failed_clients",
			"RRM coverage failed client count (current value, not cumulative)",
			baseRadioLabels,
			nil,
		)
		collector.lastRadarOnRadioAtDesc = prometheus.NewDesc(
			"wnc_ap_last_radar_timestamp_seconds",
			"Unix timestamp of the last radar detection on this radio",
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
		ch <- c.cpuUtilizationDesc
		ch <- c.memoryUtilizationDesc
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
		slog.Debug("Failed to get CAPWAP data", "error", err)
	}
	capwapMap = buildCAPWAPMap(capwapData)

	var radioDataMap map[string]*ap.RadioOperData
	radioDataSlice, err := c.src.GetRadioData(ctx)
	if err != nil {
		slog.Debug("Failed to get radio data", "error", err)
	}
	radioDataMap = buildRadioDataMap(radioDataSlice)

	var radioOperStatsMap map[string]map[int]ap.RadioOperStats
	if IsEnabled(c.metrics.Traffic, c.metrics.Errors) {
		radioOperStats, err := c.src.GetRadioOperStats(ctx)
		if err != nil {
			slog.Debug("Failed to retrieve radio operational stats", "error", err)
		}
		radioOperStatsMap = buildRadioOperStatsMap(radioOperStats)
	}

	var apOperDataMap map[string]ap.OperData
	if IsEnabled(c.metrics.General) {
		apOperData, err := c.src.GetAPOperData(ctx)
		if err != nil {
			slog.Debug("Failed to retrieve AP operational data", "error", err)
		}
		apOperDataMap = buildAPOperDataMap(apOperData)
	}

	var rrmMeasurementsMap map[string]*rrm.RRMMeasurement
	var rrmCoverageMap map[string]*rrm.RRMCoverage
	var apDot11RadarMap map[string]*rrm.ApDot11RadarData
	var clientCountsMap map[string]map[int]int
	if IsEnabled(c.metrics.Radio) {
		rrmMeasurements, err := c.rrmSrc.GetRRMMeasurements(ctx)
		if err != nil {
			slog.Debug("Failed to get RRM data for radio metrics", "error", err)
		}
		rrmMeasurementsMap = buildRRMMeasurementsMap(rrmMeasurements)

		clientData, clientErr := c.clientSrc.GetClientData(ctx)
		if clientErr != nil {
			slog.Debug("Failed to get client data for radio client counts", "error", clientErr)
		}

		nameMACMaps, mapErr := c.src.ListNameMACMaps(ctx)
		if mapErr != nil {
			slog.Debug("Failed to get AP name to MAC mapping for radio client counts", "error", mapErr)
		}

		// The count needs both sources: an empty name-to-MAC map makes every client
		// unattributable, which would silently report zero associated clients on
		// every radio. Leave the map nil so the series is omitted instead.
		if clientErr == nil && mapErr == nil {
			clientCountsMap = buildRadioClientCountsMap(clientData, nameMACMaps)
		}
	}

	var radioResetStatsMap map[string]map[int]int
	if IsEnabled(c.metrics.Errors) {
		radioResetStats, err := c.src.GetRadioResetStats(ctx)
		if err != nil {
			slog.Debug("Failed to get radio reset stats for error metrics", "error", err)
		} else {
			radioResetStatsMap = buildRadioResetStatsMap(radioResetStats)
		}

		rrmCoverage, err := c.rrmSrc.GetRRMCoverage(ctx)
		if err != nil {
			slog.Debug("Failed to get RRM coverage for error metrics", "error", err)
		} else {
			rrmCoverageMap = buildRRMCoverageMap(rrmCoverage)
		}

		apDot11Radar, err := c.rrmSrc.GetApDot11RadarData(ctx)
		if err != nil {
			slog.Debug("Failed to get radar data for error metrics", "error", err)
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
			c.collectRadioMetrics(ch, radio, rrmMeasurementsMap, clientCountsMap)
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

	// The controller lists only APs that have joined, so an AP that leaves CAPWAP
	// loses this series rather than reporting a state. An empty leaf is not a state,
	// and an empty label reads as no label at all.
	if operState := capwapMap[wtpMAC].ApState.ApOperationState; operState != "" {
		ch <- prometheus.MustNewConstMetric(
			c.operStateDesc,
			prometheus.GaugeValue,
			1,
			wtpMAC,
			operState,
		)
	}

	metrics := []Float64Metric{
		{c.configStateDesc, boolToFloat64(capwapMap[wtpMAC].TagInfo.IsApMisconfigured)},
		{c.uptimeSecondsDesc, float64(determineUptimeFromBootTime(capwapMap[wtpMAC].ApTimeInfo.BootTime))},
	}

	if sysStats := apOperDataMap[wtpMAC].ApSysStats; sysStats != nil {
		metrics = append(metrics,
			Float64Metric{c.cpuUtilizationDesc, float64(sysStats.CPUUsage) / 100},
			Float64Metric{c.memoryUtilizationDesc, float64(sysStats.MemoryUsage) / 100},
		)
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
	clientCountsMap map[string]map[int]int,
) {
	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	radioID := radio.WtpMAC + ":" + strconv.Itoa(radio.RadioSlotID)

	metrics := []Float64Metric{}

	// Both series come from the record for the band the radio is operating on. A
	// fixed index reports another band's power table, and only tx-power-level-1 may
	// be read from that table: the entries past num-supp-power-levels are undefined
	// and are not required to be ordered.
	if bandInfo, ok := currentBandInfo(radio); ok {
		metrics = append(metrics,
			Float64Metric{c.txPowerDesc, float64(bandInfo.PhyTxPwrLvlCfg.CfgData.CurrTxPowerInDbm)},
			Float64Metric{c.txPowerMaxDesc, float64(bandInfo.PhyTxPwrLvlCfg.CfgData.TxPowerLevel1)},
		)
	}

	if radio.PhyHtCfg != nil {
		metrics = append(metrics,
			Float64Metric{c.channelDesc, float64(radio.PhyHtCfg.CfgData.CurrFreq)},
			Float64Metric{c.channelWidthDesc, float64(radio.PhyHtCfg.CfgData.ChanWidth)},
		)
	}

	if rrmData, ok := rrmMeasurementsMap[radioID]; ok {
		if rrmData.Load != nil {
			metrics = append(metrics,
				Float64Metric{c.channelUtilizationDesc, float64(rrmData.Load.CcaUtilPercentage) / 100},
				Float64Metric{c.rxUtilizationDesc, float64(rrmData.Load.RxUtilPercentage) / 100},
				Float64Metric{c.txUtilizationDesc, float64(rrmData.Load.TxUtilPercentage) / 100},
				Float64Metric{c.noiseUtilizationDesc, float64(rrmData.Load.RxNoiseChannelUtilization) / 100},
			)
		}
		if noise, found := noiseOnCurrentChannel(rrmData, radio); found {
			metrics = append(metrics, Float64Metric{c.noiseFloorDesc, float64(noise)})
		}
	}

	// A nil map means client data or the name-to-MAC mapping was unavailable.
	// Indexing it would yield zero, which reads as "no clients on this radio".
	if clientCountsMap != nil {
		metrics = append(metrics, Float64Metric{
			c.associatedClientsDesc,
			float64(clientCountsMap[radio.WtpMAC][radio.RadioSlotID]),
		})
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
	radioResetStatsMap map[string]map[int]int,
	rrmCoverageMap map[string]*rrm.RRMCoverage,
	apDot11RadarMap map[string]*rrm.ApDot11RadarData,
) {
	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	radioID := radio.WtpMAC + ":" + strconv.Itoa(radio.RadioSlotID)

	// These three come from data types other than radio-oper-stats, so they must
	// not share its early return below. Each is omitted when its own source has no
	// entry for this radio: a zero reset count reads as a healthy radio, and a zero
	// radar timestamp reads as a DFS detection at the Unix epoch.
	if resetCount, exists := radioResetStatsMap[radio.WtpMAC][radio.RadioSlotID]; exists {
		ch <- prometheus.MustNewConstMetric(
			c.radioResetTotalDesc, prometheus.CounterValue, float64(resetCount), labels...)
	}
	if coverage, exists := rrmCoverageMap[radioID]; exists {
		ch <- prometheus.MustNewConstMetric(
			c.coverageHoleEventsDesc, prometheus.GaugeValue,
			float64(coverage.FailedClientCount), labels...)
	}
	if radar, exists := apDot11RadarMap[radioID]; exists &&
		!radar.LastRadarOnRadio.IsZero() && radar.LastRadarOnRadio.Year() > 1970 {
		ch <- prometheus.MustNewConstMetric(
			c.lastRadarOnRadioAtDesc, prometheus.GaugeValue,
			float64(radar.LastRadarOnRadio.Unix()), labels...)
	}

	stats, ok := radioOperStatsMap[radio.WtpMAC][radio.RadioSlotID]
	if !ok {
		return
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

	band := APRadioBand(radio)
	radioSlot := strconv.Itoa(radio.RadioSlotID)

	model := capwap.DeviceDetail.StaticInfo.ApModels.Model
	serial := capwap.DeviceDetail.StaticInfo.BoardData.WtpSerialNum
	swVersion := capwap.DeviceDetail.WtpVersion.SwVersion
	ethMAC := capwap.DeviceDetail.StaticInfo.BoardData.WtpEnetMAC

	values := make([]string, len(c.infoLabelNames))
	for i, label := range c.infoLabelNames {
		switch label {
		case labelMAC:
			values[i] = radio.WtpMAC
		case labelName:
			values[i] = capwap.Name
		case labelIP:
			values[i] = capwap.IPAddr
		case labelRadio:
			values[i] = radioSlot
		case labelBand:
			values[i] = band
		case labelModel:
			values[i] = model
		case labelSerial:
			values[i] = serial
		case labelSWVersion:
			values[i] = swVersion
		case labelEthMAC:
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

// buildRadioResetStatsMap totals the reset count per radio. The YANG list is
// keyed by ap-mac, radio-id, cause and detail-cause, so one radio legitimately
// has several entries; keying only on the first two and overwriting would report
// one arbitrary cause's count and would decrease whenever the entry set changes,
// which Prometheus reads as a counter reset.
func buildRadioResetStatsMap(radioResetStats []ap.RadioResetStats) map[string]map[int]int {
	statsMap := make(map[string]map[int]int)
	for _, stats := range radioResetStats {
		if statsMap[stats.ApMAC] == nil {
			statsMap[stats.ApMAC] = make(map[int]int)
		}
		statsMap[stats.ApMAC][stats.RadioID] += stats.Count
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

// buildRadioClientCountsMap counts associated clients per AP radio slot.
// Client data carries the AP name only, so it is resolved to the WTP MAC used by radio metrics.
func buildRadioClientCountsMap(
	clientData []client.CommonOperData,
	nameMACMaps []ap.ApNameMACMap,
) map[string]map[int]int {
	countsMap := make(map[string]map[int]int)

	nameToMAC := make(map[string]string, len(nameMACMaps))
	for _, entry := range nameMACMaps {
		nameToMAC[entry.WtpName] = entry.WtpMAC
	}

	for _, data := range clientData {
		if data.CoState != ClientStatusRun {
			continue
		}

		wtpMAC, ok := nameToMAC[data.ApName]
		if !ok {
			continue
		}

		if countsMap[wtpMAC] == nil {
			countsMap[wtpMAC] = make(map[int]int)
		}
		countsMap[wtpMAC][data.MsApSlotID]++
	}

	return countsMap
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

// noiseOnCurrentChannel returns the RRM noise for the channel the radio operates on,
// and reports whether the channel was found.
//
// noise-data is a per-channel list spanning the band's channel set, so a fixed index
// reports a channel the radio is not on. The list carries no YANG key either, so its
// order is not specified. The channel to match is the primary one, which is absent on
// a radio in monitor or sniffer mode; because the SDK types it as a plain integer,
// zero is treated as absent rather than as a channel.
func noiseOnCurrentChannel(rrmData *rrm.RRMMeasurement, radio *ap.RadioOperData) (int, bool) {
	if radio.PhyHtCfg == nil || rrmData.Noise == nil {
		return 0, false
	}

	channel := radio.PhyHtCfg.CfgData.CurrFreq
	if channel == 0 {
		return 0, false
	}

	for _, item := range rrmData.Noise.Noise.NoiseData {
		if item.Chan == channel {
			return item.Noise, true
		}
	}
	return 0, false
}
