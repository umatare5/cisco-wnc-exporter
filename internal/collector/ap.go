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
	General     bool
	Radio       bool
	Traffic     bool
	Errors      bool
	Join        bool
	Geolocation bool
	Spectrum    bool
	Info        bool
	InfoLabels  []string
}

// APCollector implements prometheus.Collector for AP metrics from WNC.
type APCollector struct {
	metrics        APMetrics
	infoDesc       *prometheus.Desc
	infoLabelNames []string
	join           *apJoinDescs
	geo            *apGeoDescs
	band           *apBandDescs
	rrmRuns        *apRRMDescs
	src            wnc.APSource
	rrmSrc         wnc.RRMSource
	clientSrc      wnc.ClientSource

	channelUtilizationDesc        *prometheus.Desc
	rxUtilizationDesc             *prometheus.Desc
	txUtilizationDesc             *prometheus.Desc
	noiseUtilizationDesc          *prometheus.Desc
	noiseFloorDesc                *prometheus.Desc
	txPowerDesc                   *prometheus.Desc
	rrmProfilePassedDesc          *prometheus.Desc
	channelChangesTotalDesc       *prometheus.Desc
	channelEnergyDesc             *prometheus.Desc
	airQualityDesc                *prometheus.Desc
	airQualityMinDesc             *prometheus.Desc
	interferersDesc               *prometheus.Desc
	lastAirQualityAtDesc          *prometheus.Desc
	channelDesc                   *prometheus.Desc
	channelWidthDesc              *prometheus.Desc
	associatedClientsDesc         *prometheus.Desc
	radioStateDesc                *prometheus.Desc
	adminStateDesc                *prometheus.Desc
	operStateDesc                 *prometheus.Desc
	configStateDesc               *prometheus.Desc
	txPowerMaxDesc                *prometheus.Desc
	dataRxFramesTotalDesc         *prometheus.Desc
	dataTxFramesTotalDesc         *prometheus.Desc
	managementRxFramesTotalDesc   *prometheus.Desc
	managementTxFramesTotalDesc   *prometheus.Desc
	controlRxFramesTotalDesc      *prometheus.Desc
	controlTxFramesTotalDesc      *prometheus.Desc
	multicastRxFramesTotalDesc    *prometheus.Desc
	multicastTxFramesTotalDesc    *prometheus.Desc
	totalTxFramesTotalDesc        *prometheus.Desc
	rtsSuccessesTotalDesc         *prometheus.Desc
	rxErrorsTotalDesc             *prometheus.Desc
	txRetriesTotalDesc            *prometheus.Desc
	transmissionFailuresTotalDesc *prometheus.Desc
	duplicateFramesTotalDesc      *prometheus.Desc
	fcsErrorsTotalDesc            *prometheus.Desc
	rxFragmentsTotalDesc          *prometheus.Desc
	txFragmentsTotalDesc          *prometheus.Desc
	rtsFailuresTotalDesc          *prometheus.Desc
	decryptionErrorsTotalDesc     *prometheus.Desc
	micErrorsTotalDesc            *prometheus.Desc
	coverageFailedClientsDesc     *prometheus.Desc
	lastRadarOnRadioAtDesc        *prometheus.Desc
	radioResetsTotalDesc          *prometheus.Desc
	cpuUtilizationDesc            *prometheus.Desc
	memoryUtilizationDesc         *prometheus.Desc
	uptimeSecondsDesc             *prometheus.Desc
	associationUptimeSecondsDesc  *prometheus.Desc
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

	if metrics.Join {
		collector.join = newAPJoinDescs()
	}

	if metrics.Geolocation {
		collector.geo = newAPGeoDescs()
	}

	if metrics.General {
		collector.radioStateDesc = prometheus.NewDesc(
			"wnc_ap_radio_state",
			"Radio state (1=up, 0=any other value), absent if unreported",
			baseRadioLabels,
			nil,
		)
		collector.adminStateDesc = prometheus.NewDesc(
			"wnc_ap_admin_state",
			"Admin state (1=enabled, 0=any other value), absent if unreported",
			baseRadioLabels,
			nil,
		)
		collector.operStateDesc = prometheus.NewDesc(
			"wnc_ap_oper_state",
			"AP operational state (4=registered). The enumeration declares no 0",
			baseAPLabels,
			nil,
		)
		collector.configStateDesc = prometheus.NewDesc(
			"wnc_ap_config_state",
			"Tag configuration state (0=valid, 1=invalid), absent if unreported",
			baseAPLabels,
			nil,
		)
		collector.uptimeSecondsDesc = prometheus.NewDesc(
			"wnc_ap_uptime_seconds",
			"AP uptime since boot, absent rather than 0 without a usable boot instant",
			baseAPLabels,
			nil,
		)
		collector.associationUptimeSecondsDesc = prometheus.NewDesc(
			"wnc_ap_association_uptime_seconds",
			"Age of this AP's current CAPWAP association, absent rather than 0 without a usable join instant",
			baseAPLabels,
			nil,
		)
	}

	if metrics.Radio {
		collector.channelDesc = prometheus.NewDesc(
			"wnc_ap_channel_number",
			"Operating channel number, absent if unreported",
			baseRadioLabels,
			nil,
		)
		collector.channelWidthDesc = prometheus.NewDesc(
			"wnc_ap_channel_width_mhz",
			"Channel bandwidth (MHz), absent if unreported",
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
		collector.rrmProfilePassedDesc = prometheus.NewDesc(
			"wnc_ap_rrm_profile_passed",
			"Whether the radio passes this RRM profile (1=passed, 0=failed), absent if unreported",
			[]string{labelMAC, labelRadio, labelProfile},
			nil,
		)
		collector.channelEnergyDesc = prometheus.NewDesc(
			"wnc_ap_channel_energy_dbm",
			"Energy measured on the channel DCA assigned this radio, absent when unmeasured",
			baseRadioLabels,
			nil,
		)
		collector.channelChangesTotalDesc = prometheus.NewDesc(
			"wnc_ap_channel_changes_total",
			"Channel changes from the DCA statistics of this radio, absent if unreported. It resets",
			baseRadioLabels,
			nil,
		)
		collector.rrmRuns = newAPRRMDescs()
	}

	if metrics.Spectrum {
		collector.airQualityDesc = prometheus.NewDesc(
			"wnc_ap_air_quality_index_avg",
			"Average CleanAir air quality index on this radio's channel; higher is cleaner, absent if unreported",
			baseRadioLabels,
			nil,
		)
		collector.airQualityMinDesc = prometheus.NewDesc(
			"wnc_ap_air_quality_index_min",
			"Lowest CleanAir air quality index on this radio's channel; higher is cleaner, absent if unreported",
			baseRadioLabels,
			nil,
		)
		collector.interferersDesc = prometheus.NewDesc(
			"wnc_ap_interferers",
			"CleanAir interference devices on this radio's channel; 0 is a real reading, absent if unreported",
			baseRadioLabels,
			nil,
		)
		collector.lastAirQualityAtDesc = prometheus.NewDesc(
			"wnc_ap_last_air_quality_timestamp_seconds",
			"Instant stamped on the CleanAir row this radio's spectrum series read; absent if unreported",
			baseRadioLabels,
			nil,
		)
		collector.band = newAPBandDescs()
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
		collector.rtsSuccessesTotalDesc = prometheus.NewDesc(
			"wnc_ap_rts_successes_total",
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
		collector.txRetriesTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_retries_total",
			"Total TX retries",
			baseRadioLabels,
			nil,
		)
		collector.transmissionFailuresTotalDesc = prometheus.NewDesc(
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
		collector.rxFragmentsTotalDesc = prometheus.NewDesc(
			"wnc_ap_rx_fragments_total",
			"RX fragments",
			baseRadioLabels,
			nil,
		)
		collector.txFragmentsTotalDesc = prometheus.NewDesc(
			"wnc_ap_tx_fragments_total",
			"TX fragments",
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
		collector.coverageFailedClientsDesc = prometheus.NewDesc(
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
		collector.radioResetsTotalDesc = prometheus.NewDesc(
			"wnc_ap_radio_resets_total",
			"Radio resets summed over cause entries, absent if unreported. The total falls when one is dropped",
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
		ch <- c.associationUptimeSecondsDesc
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
		ch <- c.rrmProfilePassedDesc
		ch <- c.channelChangesTotalDesc
		ch <- c.channelEnergyDesc
		c.rrmRuns.describe(ch)
	}
	if c.metrics.Traffic {
		ch <- c.dataRxFramesTotalDesc
		ch <- c.dataTxFramesTotalDesc
		ch <- c.managementRxFramesTotalDesc
		ch <- c.managementTxFramesTotalDesc
		ch <- c.controlRxFramesTotalDesc
		ch <- c.controlTxFramesTotalDesc
		ch <- c.multicastRxFramesTotalDesc
		ch <- c.multicastTxFramesTotalDesc
		ch <- c.totalTxFramesTotalDesc
		ch <- c.rtsSuccessesTotalDesc
	}
	if c.metrics.Errors {
		ch <- c.rxErrorsTotalDesc
		ch <- c.txRetriesTotalDesc
		ch <- c.transmissionFailuresTotalDesc
		ch <- c.duplicateFramesTotalDesc
		ch <- c.fcsErrorsTotalDesc
		ch <- c.rxFragmentsTotalDesc
		ch <- c.txFragmentsTotalDesc
		ch <- c.rtsFailuresTotalDesc
		ch <- c.decryptionErrorsTotalDesc
		ch <- c.micErrorsTotalDesc
		ch <- c.coverageFailedClientsDesc
		ch <- c.lastRadarOnRadioAtDesc
		ch <- c.radioResetsTotalDesc
	}
	if c.metrics.Join {
		c.join.describe(ch)
	}
	if c.metrics.Geolocation {
		c.geo.describe(ch)
	}
	if c.metrics.Spectrum {
		ch <- c.airQualityDesc
		ch <- c.airQualityMinDesc
		ch <- c.interferersDesc
		ch <- c.lastAirQualityAtDesc
		c.band.describe(ch)
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

	if c.metrics.Join {
		joinStats, err := c.src.GetAPJoinStats(ctx)
		if err != nil {
			slog.Debug("Failed to get AP join statistics", "error", err)
		} else {
			c.join.collect(ch, joinStats)
		}
	}

	// The coordinate list carries its own ap-mac, so this reads no other AP data type and
	// runs here rather than in the per-radio loop below, where the AP-keyed label set would
	// repeat on a multi-radio AP.
	if c.metrics.Geolocation {
		geoData, err := c.src.GetAPGeoLocData(ctx)
		if err != nil {
			slog.Debug("Failed to get AP geolocation data", "error", err)
		} else {
			c.geo.collect(ch, geoData)
		}
	}

	// Every module below reads the radio list, and the general and info modules read the
	// AP inventory as well. The join and coordinate modules read neither, so a deployment
	// enabling only those must not go on to ask for data types no enabled module declared.
	if !c.isAnyRadioKeyedFlagEnabled() {
		return
	}

	var capwapMap map[string]ap.CAPWAPData
	if IsEnabled(c.metrics.General, c.metrics.Info) {
		capwapData, err := c.src.GetCAPWAPData(ctx)
		if err != nil {
			slog.Debug("Failed to get CAPWAP data", "error", err)
		}
		capwapMap = buildCAPWAPMap(capwapData)
	}

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

	var rrmCoverageMap map[string]*rrm.RRMCoverage
	var apDot11RadarMap map[string]*rrm.ApDot11RadarData
	var radioSources radioJoins
	if IsEnabled(c.metrics.Radio) {
		radioSources = c.readRadioJoins(ctx)
	}

	var spectrumSources spectrumReads
	if IsEnabled(c.metrics.Spectrum) {
		spectrumSources = c.readSpectrum(ctx)
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
			c.collectRadioMetrics(
				ch, radio, radioSources.measurements, radioSources.clientCounts, radioSources.slots,
			)
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
		if c.metrics.Spectrum {
			c.collectSpectrumMetrics(ch, radio, spectrumSources.aqTable)
		}
		if c.metrics.Info {
			c.collectInfoMetrics(ch, radio, capwapMap)
		}
	}

	// Outside the loop above: the band label is the whole identifier of these series, so
	// emitting them per radio would repeat one label set and fail the whole scrape.
	if IsEnabled(c.metrics.Spectrum) {
		c.band.collect(ch, spectrumSources.worst)
	}
	if IsEnabled(c.metrics.Radio) {
		c.rrmRuns.collect(ch, radioSources.mainData)
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

	// The controller lists only APs that have joined. A record is replaced rather than
	// removed while an AP rejoins, so this series can carry the state from before it left,
	// and it disappears only for an AP the controller drops from the list.
	emitEnumReading(ch, c.operStateDesc, apOperationStates,
		capwapMap[wtpMAC].ApState.ApOperationState, wtpMAC)

	// Zero is the valid verdict here, so publishing it for an omitted leaf would assert a
	// configuration the controller never confirmed.
	metrics := appendFlag(nil, c.configStateDesc, capwapMap[wtpMAC].TagInfo.IsApMisconfigured)

	timeInfo := capwapMap[wtpMAC].ApTimeInfo
	if uptime, ok := determineUptimeFromTimestamp(timeInfo.JoinTime); ok {
		ch <- prometheus.MustNewConstMetric(
			c.associationUptimeSecondsDesc,
			prometheus.GaugeValue,
			float64(uptime),
			wtpMAC,
		)
	}

	if uptime, ok := determineUptimeFromTimestamp(timeInfo.BootTime); ok {
		metrics = append(metrics, Float64Metric{c.uptimeSecondsDesc, float64(uptime)})
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

// isRadio reports whether an entry of the slot list is a radio. The list carries
// entries that are not: a remote-LAN port arrives with three leaves and neither state,
// measured, and that absence is the only thing identifying it. The controller does send
// a counter record for such an entry, and every counter in it is zero, so a reading
// taken from it would report a radio that never carries traffic.
func isRadio(radio *ap.RadioOperData) bool {
	return radio.OperState != ""
}

// collectGeneralMetrics collects general radio metrics.
func (c *APCollector) collectGeneralMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
) {
	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}

	metrics := []Float64Metric{}

	// An absent leaf is not a state, and comparing it against the up spelling reports
	// the radio down. The AP-level state above applies the same rule.
	if isRadio(radio) {
		metrics = append(metrics,
			Float64Metric{c.radioStateDesc, boolToFloat64(radio.OperState == APRadioStateUp)},
		)
	}

	if radio.AdminState != "" {
		metrics = append(metrics,
			Float64Metric{c.adminStateDesc, boolToFloat64(radio.AdminState == APAdminStateEnabled)},
		)
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
	radioSlotMap map[string]*rrm.RadioSlot,
) {
	if !isRadio(radio) {
		return
	}

	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	radioID := radio.WtpMAC + ":" + strconv.Itoa(radio.RadioSlotID)

	metrics := []Float64Metric{}

	// Both series come from the record for the band the radio is operating on. A
	// fixed index reports another band's power table, and only tx-power-level-1 may
	// be read from that table: the entries past num-supp-power-levels are undefined
	// and are not required to be ordered.
	// Each level is withheld on its own: the current power is a reading in its own right
	// whether or not the table also carries the maximum.
	if bandInfo, ok := currentBandInfo(radio); ok {
		cfgData := bandInfo.PhyTxPwrLvlCfg.CfgData
		metrics = appendNumber(metrics, c.txPowerDesc, cfgData.CurrTxPowerInDbm)
		metrics = appendNumber(metrics, c.txPowerMaxDesc, cfgData.TxPowerLevel1)
	}

	// The controller omits curr-freq on a radio in monitor mode, measured on 17.12, and the
	// SDK types both leaves as plain integers, so an omitted one decodes to a zero that is
	// neither a channel on any band nor a width a radio can use. The schema the controller
	// serves declares the same absence for sniffer mode, which is unmeasured. Each leaf is
	// withheld on its own zero: the measured radio kept its width while its channel was gone.
	if radio.PhyHtCfg != nil {
		if channel := radio.PhyHtCfg.CfgData.CurrFreq; channel != 0 {
			metrics = append(metrics, Float64Metric{c.channelDesc, float64(channel)})
		}
		if width := radio.PhyHtCfg.CfgData.ChanWidth; width != 0 {
			metrics = append(metrics, Float64Metric{c.channelWidthDesc, float64(width)})
		}
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

	// The verdicts carry a third label, so they are emitted here rather than through
	// the two-label slice above. A radio the slot list has no record for, one whose
	// record carries no radio-data container, and one profile whose verdict leaf is
	// omitted all report no verdict rather than a failure.
	if slot, found := radioSlotMap[radioID]; found && slot.RadioData != nil {
		for _, profile := range rrmProfiles {
			verdict := profile.passed(slot.RadioData)
			if verdict == nil {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.rrmProfilePassedDesc,
				prometheus.GaugeValue,
				boolToFloat64(*verdict),
				radio.WtpMAC, strconv.Itoa(radio.RadioSlotID), profile.name,
			)
		}

		// The channel-change counter lives one container deeper, which the controller
		// omits on a radio it reports no assignment statistics for. A zero there would
		// read as a radio DCA has never moved.
		if dca := slot.RadioData.DCAStats; dca != nil {
			ch <- prometheus.MustNewConstMetric(
				c.channelChangesTotalDesc,
				prometheus.CounterValue,
				float64(dca.ChanChanges),
				labels...,
			)

			if hasChannelEnergy(dca.CurrentChanEnergy) {
				metrics = append(metrics,
					Float64Metric{c.channelEnergyDesc, float64(dca.CurrentChanEnergy)})
			}
		}
	}

	for _, metric := range metrics {
		ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.GaugeValue, metric.Value, labels...)
	}
}

// The two readings the energy leaf carries that are not measurements. Neither can be one:
// the sentinel is the lower bound of the leaf's own signed type and sits far below the
// thermal noise floor of any channel width, and zero sits far above every measured
// energy and above the controller's own lower limit for the assignment. The sentinel is
// what a radio reads until DCA next runs for its band, measured on several radios with an
// untouched radio as a control; zero has not been observed and is guarded because the
// leaf is a plain integer, so an omitted one would decode to it.
const (
	channelEnergyAbsent   = 0
	channelEnergySentinel = -128
)

// hasChannelEnergy reports whether the energy leaf carries a measurement.
func hasChannelEnergy(energy int) bool {
	return energy != channelEnergyAbsent && energy != channelEnergySentinel
}

// rrmProfiles pairs each profile label value with the verdict leaf it is read from.
// The controller keeps one leaf per profile rather than an enumeration, so the label
// values are named here and are this exporter's own.
var rrmProfiles = []struct {
	name   string
	passed func(*rrm.RadioData) *bool
}{
	{"coverage", func(d *rrm.RadioData) *bool { return d.CoverageProfilePassed }},
	{"load", func(d *rrm.RadioData) *bool { return d.LoadProfPassed }},
	{"interference", func(d *rrm.RadioData) *bool { return d.InterferenceProfilePassed }},
	{"noise", func(d *rrm.RadioData) *bool { return d.NoiseProfilePassed }},
}

// radioJoins holds the four reads of the radio module. A map with no entry for a radio
// withholds that radio's series rather than reporting a zero; clientCounts is left nil
// outright, because a partial count reads as a radio with no clients. mainData joins
// against no radio — it is keyed by band — and is carried here so that the module makes
// its reads in one place.
type radioJoins struct {
	measurements map[string]*rrm.RRMMeasurement
	slots        map[string]*rrm.RadioSlot
	clientCounts map[string]map[int]int
	mainData     []rrm.MainData
}

// spectrumReads holds the two air quality reads of the spectrum module. They key on
// different things — one on the radio, one on the band — so one failing leaves the other
// published rather than withholding both.
type spectrumReads struct {
	aqTable []rrm.SpectrumAqTable
	worst   []rrm.SpectrumAqWorstTable
}

// readSpectrum reads the two data types the spectrum module publishes from.
func (c *APCollector) readSpectrum(ctx context.Context) spectrumReads {
	var reads spectrumReads

	table, err := c.rrmSrc.GetSpectrumAqTable(ctx)
	if err != nil {
		slog.Debug("Failed to get the air quality table for spectrum metrics", "error", err)
	}
	reads.aqTable = table

	worst, worstErr := c.rrmSrc.GetSpectrumAqWorstTable(ctx)
	if worstErr != nil {
		slog.Debug("Failed to get the worst air quality table for spectrum metrics", "error", worstErr)
	}
	reads.worst = worst

	return reads
}

// readRadioJoins reads the four data types the radio module publishes from. Each keeps
// its own absence rule, so one failing does not withhold the others.
func (c *APCollector) readRadioJoins(ctx context.Context) radioJoins {
	var joins radioJoins

	measurements, err := c.rrmSrc.GetRRMMeasurements(ctx)
	if err != nil {
		slog.Debug("Failed to get RRM data for radio metrics", "error", err)
	}
	joins.measurements = buildRRMMeasurementsMap(measurements)

	slots, slotErr := c.rrmSrc.GetRadioSlots(ctx)
	if slotErr != nil {
		slog.Debug("Failed to get RRM radio slot data for radio metrics", "error", slotErr)
	}
	joins.slots = buildRadioSlotMap(slots)

	clientData, clientErr := c.clientSrc.GetClientData(ctx)
	if clientErr != nil {
		slog.Debug("Failed to get client data for radio client counts", "error", clientErr)
	}

	nameMACMaps, mapErr := c.src.ListNameMACMaps(ctx)
	if mapErr != nil {
		slog.Debug("Failed to get AP name to MAC mapping for radio client counts", "error", mapErr)
	}

	// The count needs both sources: an empty name-to-MAC map makes every client
	// unattributable, which would silently report zero associated clients on every
	// radio. Leave the map nil so the series is omitted instead.
	if clientErr == nil && mapErr == nil {
		joins.clientCounts = buildRadioClientCountsMap(clientData, nameMACMaps)
	}

	mainData, mainErr := c.rrmSrc.GetRRMMainData(ctx)
	if mainErr != nil {
		slog.Debug("Failed to get RRM main data for the band-keyed run instants", "error", mainErr)
	}
	joins.mainData = mainData

	return joins
}

// collectSpectrumMetrics publishes the air quality of the channel the radio operates
// on. The reading is absent for a radio the table has no record for, which covers an AP
// without CleanAir and a radio whose spectrum operation is down, and for every radio
// while the fetch fails.
//
// emitTimestamp guards the instant once more, so a radio can publish the three readings
// while their instant is withheld.
func (c *APCollector) collectSpectrumMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	spectrumAqTable []rrm.SpectrumAqTable,
) {
	row, found := airQualityOnCurrentChannel(spectrumAqTable, radio)
	if !found {
		return
	}

	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	for _, metric := range []Float64Metric{
		{c.airQualityDesc, float64(row.Aqi)},
		{c.airQualityMinDesc, float64(row.MinAqi)},
		{c.interferersDesc, float64(row.TotalIntfDeviceCount)},
	} {
		ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.GaugeValue, metric.Value, labels...)
	}

	emitTimestamp(ch, c.lastAirQualityAtDesc, row.SpectrumTimestamp, labels...)
}

func (c *APCollector) collectTrafficMetrics(
	ch chan<- prometheus.Metric,
	radio *ap.RadioOperData,
	radioOperStatsMap map[string]map[int]ap.RadioOperStats,
) {
	if !isRadio(radio) {
		return
	}

	stats, ok := radioOperStatsMap[radio.WtpMAC][radio.RadioSlotID]
	if !ok {
		return
	}

	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	trafficMetrics := []Float64Metric{
		{c.dataRxFramesTotalDesc, float64(stats.RxDataFrameCount)},
		{c.dataTxFramesTotalDesc, float64(stats.TxDataFrameCount)},
		{c.managementRxFramesTotalDesc, float64(stats.RxMgmtFrameCount)},
		{c.managementTxFramesTotalDesc, float64(stats.TxMgmtFrameCount)},
		{c.controlRxFramesTotalDesc, float64(stats.RxCtrlFrameCount)},
		{c.controlTxFramesTotalDesc, float64(stats.TxCtrlFrameCount)},
		{c.multicastRxFramesTotalDesc, float64(stats.MulticastRxFrameCnt)},
		{c.multicastTxFramesTotalDesc, float64(stats.MulticastTxFrameCnt)},
		{c.totalTxFramesTotalDesc, float64(stats.TxFrameCount)},
		{c.rtsSuccessesTotalDesc, float64(stats.RtsSuccessCount)},
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
	if !isRadio(radio) {
		return
	}

	labels := []string{radio.WtpMAC, strconv.Itoa(radio.RadioSlotID)}
	radioID := radio.WtpMAC + ":" + strconv.Itoa(radio.RadioSlotID)

	// These three come from data types other than radio-oper-stats, so they must
	// not share its early return below. Each is omitted when its own source has no
	// entry for this radio: a zero reset count reads as a healthy radio, and a zero
	// radar timestamp reads as a DFS detection at the Unix epoch.
	if resetCount, exists := radioResetStatsMap[radio.WtpMAC][radio.RadioSlotID]; exists {
		ch <- prometheus.MustNewConstMetric(
			c.radioResetsTotalDesc, prometheus.CounterValue, float64(resetCount), labels...)
	}
	if coverage, exists := rrmCoverageMap[radioID]; exists {
		ch <- prometheus.MustNewConstMetric(
			c.coverageFailedClientsDesc, prometheus.GaugeValue,
			float64(coverage.FailedClientCount), labels...)
	}
	if radar, exists := apDot11RadarMap[radioID]; exists && radar.LastRadarOnRadio.Year() > epochYear {
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
		{c.txRetriesTotalDesc, float64(stats.RetryCount)},
		{c.transmissionFailuresTotalDesc, float64(stats.FailedCount)},
		{c.duplicateFramesTotalDesc, float64(stats.FrameDuplicateCount)},
		{c.fcsErrorsTotalDesc, float64(stats.FcsErrorCount)},
		{c.rxFragmentsTotalDesc, float64(stats.RxFragmentCount)},
		{c.txFragmentsTotalDesc, float64(stats.TxFragmentCount)},
		{c.rtsFailuresTotalDesc, float64(stats.RtsFailureCount)},
		{c.decryptionErrorsTotalDesc, float64(stats.MACDecryErrFrameCount)},
		{c.micErrorsTotalDesc, float64(stats.MACMicErrFrameCount)},
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

// buildRadioSlotMap indexes the RRM slot record by the same key the radio series are
// labeled with, so a verdict is published against the radio it was measured on.
func buildRadioSlotMap(radioSlots []rrm.RadioSlot) map[string]*rrm.RadioSlot {
	slotMap := make(map[string]*rrm.RadioSlot, len(radioSlots))
	for i, slot := range radioSlots {
		key := slot.WtpMAC + ":" + strconv.Itoa(slot.RadioSlotID)
		slotMap[key] = &radioSlots[i]
	}
	return slotMap
}

// buildRadioResetStatsMap totals the reset count per radio. The YANG list is
// keyed by ap-mac, radio-id, cause and detail-cause, so one radio legitimately
// has several entries; keying only on the first two and overwriting would report
// one arbitrary cause's count and would decrease whenever the entry set changes.
// The fold removes that decrease and not every one: the total itself falls when
// the controller deletes entries, observed together with an AP boot or re-join on
// every occasion.
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

// determineUptimeFromTimestamp derives elapsed seconds from a timestamp leaf, and reports
// false when the leaf is absent, unparsable, or at the Unix epoch. Neither zero nor five
// decades is a usable substitute: the first reads as an event that happened this instant,
// which is what a reboot rule fires on, and the second silences one.
//
// Nothing here happened in 1970, so an instant there is a placeholder whatever the
// controller meant by it. This is the same guard the join module applies to its own
// timestamps.
func determineUptimeFromTimestamp(timestamp string) (int64, bool) {
	instant, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0, false
	}
	if instant.Year() <= epochYear {
		return 0, false
	}

	return int64(time.Since(instant).Seconds()), true
}

func (c *APCollector) isAnyMetricFlagEnabled() bool {
	return c.isAnyRadioKeyedFlagEnabled() || c.metrics.Join || c.metrics.Geolocation
}

// isAnyRadioKeyedFlagEnabled reports whether a module keyed by the AP inventory or
// the radio list is enabled.
func (c *APCollector) isAnyRadioKeyedFlagEnabled() bool {
	return IsEnabled(
		c.metrics.General, c.metrics.Radio, c.metrics.Traffic, c.metrics.Errors,
		c.metrics.Spectrum, c.metrics.Info,
	)
}

// airQualityOnCurrentChannel returns the CleanAir row for the channel the radio operates
// on, and reports whether it was found. The row is returned rather than one of its leaves
// because the series published from it must agree on which row they read.
//
// The table is keyed by AP and band, and its per-channel list spans the band's channel
// set, so the reading is reached by matching the band first and the primary channel
// second. Matching the channel alone would cross bands, because 6 GHz channel numbers
// restart at 1 and collide with 2.4 GHz. The list is a fixed-size array padded with
// rows whose channel number is zero, and the primary channel is likewise absent as a
// zero on a radio in monitor or sniffer mode, so rejecting the zero channel excludes
// both without a second test.
func airQualityOnCurrentChannel(
	table []rrm.SpectrumAqTable, radio *ap.RadioOperData,
) (*rrm.PerChannelAqList, bool) {
	if radio.PhyHtCfg == nil {
		return nil, false
	}

	channel := radio.PhyHtCfg.CfgData.CurrFreq
	if channel == 0 {
		return nil, false
	}

	for i := range table {
		record := &table[i]
		if record.WtpMAC != radio.WtpMAC || record.Band != radio.CurrentActiveBand {
			continue
		}
		if record.PerRadioAqData == nil {
			return nil, false
		}

		rows := record.PerRadioAqData.PerChannelAqList
		for j := range rows {
			if rows[j].ChannelNum == channel {
				return &rows[j], true
			}
		}
	}
	return nil, false
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

// The controller keeps one set of DTLS counters per tunnel channel in a single
// container, so the two sets are folded into one series carrying this label.
const (
	dtlsChannelControl = "control"
	dtlsChannelData    = "data"
)

// apJoinDescs holds the descriptors of the join module. It is nil when the join module is off,
// and the metrics.Join guards in the AP collector's Describe and Collect are what keep every
// series of it out of a scrape and the nil out of a dereference.
type apJoinDescs struct {
	joined *prometheus.Desc
	name   *prometheus.Desc

	discoveryRequests  *prometheus.Desc
	discoveryResponses *prometheus.Desc
	discoveryErrors    *prometheus.Desc
	joinRequests       *prometheus.Desc
	joinResponses      *prometheus.Desc
	joinFailures       *prometheus.Desc
	configRequests     *prometheus.Desc
	configResponses    *prometheus.Desc
	configFailures     *prometheus.Desc

	dtlsRequests        *prometheus.Desc
	dtlsSuccesses       *prometheus.Desc
	dtlsFailures        *prometheus.Desc
	dtlsDecryptErrors   *prometheus.Desc
	dtlsAntiReplayError *prometheus.Desc

	lastErrorAt            *prometheus.Desc
	lastJoinSuccessAt      *prometheus.Desc
	lastJoinFailureAt      *prometheus.Desc
	lastConfigSuccessAt    *prometheus.Desc
	lastConfigFailureAt    *prometheus.Desc
	lastDiscoverySuccessAt *prometheus.Desc
	lastDiscoveryFailureAt *prometheus.Desc
	lastDTLSSuccessAt      *prometheus.Desc
	lastDTLSFailureAt      *prometheus.Desc

	lastDiscoveryFailureReason *prometheus.Desc
	lastJoinFailureReason      *prometheus.Desc
	lastConfigFailureReason    *prometheus.Desc
	lastErrorPhase             *prometheus.Desc
	lastDTLSFailureReason      *prometheus.Desc
	lastRebootReason           *prometheus.Desc
	lastDisconnectReason       *prometheus.Desc
}

// newAPJoinDescs builds the descriptors of the join module.
//
// Every series is keyed by the AP radio MAC alone. The AP name is published as its
// own series instead of as a label, because a bare and between a counter of this
// module and wnc_ap_joined requires the two to carry identical label sets, and
// because renaming an AP would otherwise start a fresh counter series.
func newAPJoinDescs() *apJoinDescs {
	apLabels := []string{labelMAC}
	nameLabels := []string{labelMAC, labelName}
	channelLabels := []string{labelMAC, labelChannel}

	return &apJoinDescs{
		joined: prometheus.NewDesc(
			"wnc_ap_joined",
			"Whether the AP holds a CAPWAP session with this controller now, absent if unreported",
			apLabels, nil,
		),
		name: prometheus.NewDesc(
			"wnc_ap_join_info",
			"AP name as its CAPWAP join record reports it, always 1; absent if the record carries no name",
			nameLabels, nil,
		),

		discoveryRequests: prometheus.NewDesc(
			"wnc_ap_discovery_requests_total",
			"CAPWAP discovery requests received from this AP",
			apLabels, nil,
		),
		discoveryResponses: prometheus.NewDesc(
			"wnc_ap_discovery_responses_total",
			"Successful CAPWAP discovery responses sent to this AP",
			apLabels, nil,
		),
		discoveryErrors: prometheus.NewDesc(
			"wnc_ap_discovery_errors_total",
			"CAPWAP discovery requests from this AP the controller found in error",
			apLabels, nil,
		),
		joinRequests: prometheus.NewDesc(
			"wnc_ap_join_requests_total",
			"CAPWAP join requests received from this AP",
			apLabels, nil,
		),
		joinResponses: prometheus.NewDesc(
			"wnc_ap_join_responses_total",
			"Successful CAPWAP join responses sent to this AP",
			apLabels, nil,
		),
		joinFailures: prometheus.NewDesc(
			"wnc_ap_join_failures_total",
			"CAPWAP join requests from this AP the controller failed to process",
			apLabels, nil,
		),
		configRequests: prometheus.NewDesc(
			"wnc_ap_config_requests_total",
			"CAPWAP configuration requests received from this AP",
			apLabels, nil,
		),
		configResponses: prometheus.NewDesc(
			"wnc_ap_config_responses_total",
			"Successful CAPWAP configuration responses sent to this AP",
			apLabels, nil,
		),
		configFailures: prometheus.NewDesc(
			"wnc_ap_config_failures_total",
			"CAPWAP configuration requests from this AP the controller failed to process",
			apLabels, nil,
		),

		dtlsRequests: prometheus.NewDesc(
			"wnc_ap_dtls_session_requests_total",
			"DTLS session setup requests received from this AP on the channel label",
			channelLabels, nil,
		),
		dtlsSuccesses: prometheus.NewDesc(
			"wnc_ap_dtls_session_successes_total",
			"DTLS sessions established with this AP on the channel label",
			channelLabels, nil,
		),
		dtlsFailures: prometheus.NewDesc(
			"wnc_ap_dtls_session_failures_total",
			"DTLS sessions with this AP that failed to establish on the channel label",
			channelLabels, nil,
		),
		dtlsDecryptErrors: prometheus.NewDesc(
			"wnc_ap_dtls_decrypt_errors_total",
			"DTLS decrypt errors on the channel label of this AP's tunnel",
			channelLabels, nil,
		),
		dtlsAntiReplayError: prometheus.NewDesc(
			"wnc_ap_dtls_anti_replay_errors_total",
			"DTLS anti-replay errors on the channel label of this AP's tunnel",
			channelLabels, nil,
		),

		lastErrorAt: prometheus.NewDesc(
			"wnc_ap_last_error_timestamp_seconds",
			"Unix timestamp of this AP's last CAPWAP connection error",
			apLabels, nil,
		),
		lastJoinSuccessAt: prometheus.NewDesc(
			"wnc_ap_last_join_success_timestamp_seconds",
			"Unix timestamp of this AP's last successful CAPWAP join attempt",
			apLabels, nil,
		),
		lastJoinFailureAt: prometheus.NewDesc(
			"wnc_ap_last_join_failure_timestamp_seconds",
			"Unix timestamp of this AP's last failed CAPWAP join attempt",
			apLabels, nil,
		),
		lastConfigSuccessAt: prometheus.NewDesc(
			"wnc_ap_last_config_success_timestamp_seconds",
			"Unix timestamp of this AP's last successful CAPWAP configuration attempt",
			apLabels, nil,
		),
		lastConfigFailureAt: prometheus.NewDesc(
			"wnc_ap_last_config_failure_timestamp_seconds",
			"Unix timestamp of this AP's last failed CAPWAP configuration attempt",
			apLabels, nil,
		),
		lastDiscoverySuccessAt: prometheus.NewDesc(
			"wnc_ap_last_discovery_success_timestamp_seconds",
			"Unix timestamp of this AP's last successful CAPWAP discovery attempt",
			apLabels, nil,
		),
		lastDiscoveryFailureAt: prometheus.NewDesc(
			"wnc_ap_last_discovery_failure_timestamp_seconds",
			"Unix timestamp of this AP's last failed CAPWAP discovery attempt",
			apLabels, nil,
		),
		lastDTLSSuccessAt: prometheus.NewDesc(
			"wnc_ap_last_dtls_success_timestamp_seconds",
			"Unix timestamp of this AP's last established DTLS session on the channel label",
			channelLabels, nil,
		),
		lastDTLSFailureAt: prometheus.NewDesc(
			"wnc_ap_last_dtls_failure_timestamp_seconds",
			"Unix timestamp of this AP's last failed DTLS session on the channel label",
			channelLabels, nil,
		),

		lastDiscoveryFailureReason: prometheus.NewDesc(
			"wnc_ap_last_discovery_failure_reason",
			"Reason for this AP's last CAPWAP discovery failure; 0 is disc-fail-none, meaning none has failed",
			apLabels, nil,
		),
		lastJoinFailureReason: prometheus.NewDesc(
			"wnc_ap_last_join_failure_reason",
			"Reason for this AP's last CAPWAP join failure; 0 is jf-none, meaning no join has failed",
			apLabels, nil,
		),
		lastConfigFailureReason: prometheus.NewDesc(
			"wnc_ap_last_config_failure_reason",
			"Reason for this AP's last CAPWAP configuration failure; 0 is cf-none, meaning none has failed",
			apLabels, nil,
		),
		lastErrorPhase: prometheus.NewDesc(
			"wnc_ap_last_error_phase",
			"CAPWAP phase of this AP's last connection error; 0 is unknown, and a joined AP reads 6 too",
			apLabels, nil,
		),
		lastDTLSFailureReason: prometheus.NewDesc(
			"wnc_ap_last_dtls_failure_reason",
			"Outcome of the last DTLS handshake per channel; 0 is dtls-hs-success, also read with no session",
			channelLabels, nil,
		),
		lastRebootReason: prometheus.NewDesc(
			"wnc_ap_last_reboot_reason",
			"Reason this AP last rebooted, as the AP reported it; 0 is ap-reboot-reason-none",
			apLabels, nil,
		),
		lastDisconnectReason: prometheus.NewDesc(
			"wnc_ap_last_disconnect_reason",
			"Reason this AP last left CAPWAP; 0 is the unknown member, not the absence of a disconnect",
			apLabels, nil,
		),
	}
}

// describe sends every descriptor of the join module.
func (d *apJoinDescs) describe(ch chan<- *prometheus.Desc) {
	for _, desc := range []*prometheus.Desc{
		d.joined, d.name,
		d.discoveryRequests, d.discoveryResponses, d.discoveryErrors,
		d.joinRequests, d.joinResponses, d.joinFailures,
		d.configRequests, d.configResponses, d.configFailures,
		d.dtlsRequests, d.dtlsSuccesses, d.dtlsFailures,
		d.dtlsDecryptErrors, d.dtlsAntiReplayError,
		d.lastErrorAt, d.lastJoinSuccessAt, d.lastJoinFailureAt,
		d.lastConfigSuccessAt, d.lastConfigFailureAt,
		d.lastDiscoverySuccessAt, d.lastDiscoveryFailureAt,
		d.lastDTLSSuccessAt, d.lastDTLSFailureAt,
		d.lastDiscoveryFailureReason, d.lastJoinFailureReason, d.lastConfigFailureReason,
		d.lastErrorPhase, d.lastDTLSFailureReason,
		d.lastRebootReason, d.lastDisconnectReason,
	} {
		ch <- desc
	}
}

// collect publishes the join module for every record the controller returned. The
// list is keyed by the AP radio MAC and keeps a record for an AP that has left
// CAPWAP, so the record set is wider than the AP inventory.
func (d *apJoinDescs) collect(ch chan<- prometheus.Metric, records []ap.ApJoinStats) {
	for i := range records {
		record := &records[i]

		d.collectSession(ch, record)
		d.collectCounters(ch, record)
		d.collectTimestamps(ch, record)
		d.collectReasons(ch, record)
	}
}

// collectSession publishes the join state and the name the record carries.
func (d *apJoinDescs) collectSession(ch chan<- prometheus.Metric, record *ap.ApJoinStats) {
	if joined := record.ApJoinInfo.IsJoined; joined != nil {
		ch <- prometheus.MustNewConstMetric(
			d.joined, prometheus.GaugeValue, boolToFloat64(*joined), record.WtpMAC,
		)
	}

	// An empty name reads as no label at all, and this is the only container that
	// still names an AP the inventory has dropped, so there is nothing to fall back on.
	if name := record.ApJoinInfo.ApName; name != "" {
		ch <- prometheus.MustNewConstMetric(d.name, prometheus.GaugeValue, 1, record.WtpMAC, name)
	}
}

// counterLeaf pairs a descriptor with the counter leaf behind it. The leaf is a pointer so that
// a leaf the controller omitted stays distinguishable from one it reported as zero, which is a
// legitimate reading for every counter here.
type counterLeaf struct {
	desc  *prometheus.Desc
	value *int
}

// collectCounters publishes the phase counters, including one series per DTLS channel.
// Each counter is withheld on its own: absence is per leaf, so one omitted leaf must not cost
// the rest of the record its series.
func (d *apJoinDescs) collectCounters(ch chan<- prometheus.Metric, record *ap.ApJoinStats) {
	join := record.ApJoinInfo
	discovery := record.ApDiscoveryInfo
	dtls := record.DTLSSessInfo

	for _, leaf := range []counterLeaf{
		{d.discoveryRequests, discovery.NumDiscoveryReqRecvd},
		{d.discoveryResponses, discovery.NumSuccDiscRespSent},
		{d.discoveryErrors, discovery.NumErrDiscReq},
		{d.joinRequests, join.NumJoinReqRecvd},
		{d.joinResponses, join.NumSuccJoinRespSent},
		{d.joinFailures, join.NumUnsuccJoinReqProcn},
		{d.configRequests, join.NumConfigReqRecvd},
		{d.configResponses, join.NumSuccConfRespSent},
		{d.configFailures, join.NumUnsuccConfReqProcn},
	} {
		if leaf.value == nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			leaf.desc, prometheus.CounterValue, float64(*leaf.value), record.WtpMAC,
		)
	}

	for _, channel := range []struct {
		label  string
		leaves []counterLeaf
	}{
		{dtlsChannelControl, []counterLeaf{
			{d.dtlsRequests, dtls.CtrlDTLSSetupReq},
			{d.dtlsSuccesses, dtls.CtrlDTLSSuccess},
			{d.dtlsFailures, dtls.CtrlDTLSFailure},
			{d.dtlsDecryptErrors, dtls.CtrlDTLSDecryptErr},
			{d.dtlsAntiReplayError, dtls.CtrlDTLSAntiReplayErr},
		}},
		{dtlsChannelData, []counterLeaf{
			{d.dtlsRequests, dtls.DataDTLSSetupReq},
			{d.dtlsSuccesses, dtls.DataDTLSSuccess},
			{d.dtlsFailures, dtls.DataDTLSFailure},
			{d.dtlsDecryptErrors, dtls.DataDTLSDecryptErr},
			{d.dtlsAntiReplayError, dtls.DataDTLSAntiReplayErr},
		}},
	} {
		for _, leaf := range channel.leaves {
			if leaf.value == nil {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				leaf.desc, prometheus.CounterValue, float64(*leaf.value), record.WtpMAC, channel.label,
			)
		}
	}
}

// collectTimestamps publishes the last-event timestamps, each withheld when its leaf
// carries the epoch sentinel.
func (d *apJoinDescs) collectTimestamps(ch chan<- prometheus.Metric, record *ap.ApJoinStats) {
	join := record.ApJoinInfo
	discovery := record.ApDiscoveryInfo
	dtls := record.DTLSSessInfo

	for _, stamp := range []struct {
		desc *prometheus.Desc
		at   time.Time
	}{
		{d.lastErrorAt, join.LastErrorTime},
		{d.lastJoinSuccessAt, join.LastSuccJoinAtmptTime},
		{d.lastJoinFailureAt, join.LastFailJoinAtmptTime},
		{d.lastConfigSuccessAt, join.LastSuccConfAtmptTime},
		{d.lastConfigFailureAt, join.LastFailConfAtmptTime},
		{d.lastDiscoverySuccessAt, discovery.LastSuccessDiscTime},
		{d.lastDiscoveryFailureAt, discovery.LastFailedDiscTime},
	} {
		emitTimestamp(ch, stamp.desc, stamp.at, record.WtpMAC)
	}

	emitTimestamp(ch, d.lastDTLSSuccessAt, dtls.CtrlDTLSSuccessTime, record.WtpMAC, dtlsChannelControl)
	emitTimestamp(ch, d.lastDTLSSuccessAt, dtls.DataDTLSSuccessTime, record.WtpMAC, dtlsChannelData)
	emitTimestamp(ch, d.lastDTLSFailureAt, dtls.CtrlDTLSFailureTime, record.WtpMAC, dtlsChannelControl)
	emitTimestamp(ch, d.lastDTLSFailureAt, dtls.DataDTLSFailureTime, record.WtpMAC, dtlsChannelData)
}

// collectReasons publishes the enum leaves, each as the value the controller's own
// enumeration assigns the spelling it sent.
func (d *apJoinDescs) collectReasons(ch chan<- prometheus.Metric, record *ap.ApJoinStats) {
	join := record.ApJoinInfo
	dtls := record.DTLSSessInfo

	for _, reason := range []struct {
		desc    *prometheus.Desc
		table   enumTable
		reading string
	}{
		{d.lastDiscoveryFailureReason, apDiscoveryFailureReasons, record.ApDiscoveryInfo.LastDiscFailureType},
		{d.lastJoinFailureReason, apJoinFailureReasons, join.LastJoinFailureType},
		{d.lastConfigFailureReason, apConfigFailureReasons, join.LastConfigFailureType},
		{d.lastErrorPhase, apFailurePhases, join.LastErrorType},
		{d.lastRebootReason, apRebootReasons, record.RebootReason},
		{d.lastDisconnectReason, apDisconnectReasons, record.DisconnectReason},
	} {
		emitEnumReading(ch, reason.desc, reason.table, reason.reading, record.WtpMAC)
	}

	emitEnumReading(ch, d.lastDTLSFailureReason, apDTLSFailureReasons,
		dtls.CtrlDTLSFailureType, record.WtpMAC, dtlsChannelControl)
	emitEnumReading(ch, d.lastDTLSFailureReason, apDTLSFailureReasons,
		dtls.DataDTLSFailureType, record.WtpMAC, dtlsChannelData)
}

// apBandDescs holds the descriptors of the band-keyed air quality series. It is nil when the
// spectrum module is off, and the metrics.Spectrum guards in the AP collector's Describe and
// Collect are what keep the series out of a scrape and the nil out of a dereference.
type apBandDescs struct {
	worstAirQuality    *prometheus.Desc
	worstAirQualityMin *prometheus.Desc
	worstInterferers   *prometheus.Desc
	worstChannel       *prometheus.Desc
}

// newAPBandDescs builds the descriptors of the band-keyed air quality series.
//
// The band is the whole identifier. The controller ranks the channels of a band across
// every AP that scans it and keeps one row per band, so neither the AP nor the radio the
// reading was taken on keys the row.
func newAPBandDescs() *apBandDescs {
	bandLabels := []string{labelBand}

	return &apBandDescs{
		worstAirQuality: prometheus.NewDesc(
			"wnc_rrm_worst_channel_air_quality_index_avg",
			"Average CleanAir air quality index of the worst channel; higher is cleaner, absent if unranked",
			bandLabels, nil,
		),
		worstAirQualityMin: prometheus.NewDesc(
			"wnc_rrm_worst_channel_air_quality_index_min",
			"Lowest CleanAir air quality index of the worst channel; higher is cleaner, absent if unranked",
			bandLabels, nil,
		),
		worstInterferers: prometheus.NewDesc(
			"wnc_rrm_worst_channel_interferers",
			"CleanAir interference devices on the worst channel; 0 is a real reading, absent if unranked",
			bandLabels, nil,
		),
		worstChannel: prometheus.NewDesc(
			"wnc_rrm_worst_channel_number",
			"Channel the controller ranks worst in this band, as a value; absent if unranked",
			bandLabels, nil,
		),
	}
}

// describe sends every descriptor of the band-keyed air quality series.
func (d *apBandDescs) describe(ch chan<- *prometheus.Desc) {
	for _, desc := range []*prometheus.Desc{
		d.worstAirQuality, d.worstAirQualityMin, d.worstInterferers, d.worstChannel,
	} {
		ch <- desc
	}
}

// collect publishes the four readings of every band the controller ranked.
//
// It runs once per scrape rather than inside the per-radio loop. The band label is the
// whole identifier, so emitting from that loop would repeat every series with the same
// label set, which Gather rejects by failing the entire endpoint.
//
// Both guards withhold a row rather than publish part of it. A row whose band-id cannot
// be named is skipped instead of labeled unknown, and a row reporting no channel is a
// slot the controller has not ranked: none of the four leaves can be told apart from a
// reading there, and the worst reading of each is what a zero would publish.
func (d *apBandDescs) collect(ch chan<- prometheus.Metric, rows []rrm.SpectrumAqWorstTable) {
	for i := range rows {
		row := &rows[i]

		band, ok := RRMWorstBand(row)
		if !ok {
			continue
		}

		if row.ChannelNum == 0 {
			continue
		}

		for _, metric := range []Float64Metric{
			{d.worstAirQuality, float64(row.Aqi)},
			{d.worstAirQualityMin, float64(row.MinAqi)},
			{d.worstInterferers, float64(row.TotalIntfDeviceCount)},
			{d.worstChannel, float64(row.ChannelNum)},
		} {
			ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.GaugeValue, metric.Value, band)
		}
	}
}

// apRRMDescs holds the descriptors of the band-keyed RRM run instants. It is nil when the
// radio module is off, and the metrics.Radio guards in the AP collector's Describe and Collect
// are what keep the series out of a scrape and the nil out of a dereference.
type apRRMDescs struct {
	lastRFGroupingRunAt *prometheus.Desc
	lastDCARunAt        *prometheus.Desc
}

// newAPRRMDescs builds the descriptors of the band-keyed RRM run instants.
//
// The band is the whole identifier. The controller keeps one record per band rather than
// one per radio, so neither the AP nor the radio an assignment lands on keys the series.
func newAPRRMDescs() *apRRMDescs {
	bandLabels := []string{labelBand}

	return &apRRMDescs{
		lastRFGroupingRunAt: prometheus.NewDesc(
			"wnc_rrm_last_rf_grouping_run_timestamp_seconds",
			"Instant of the last RF grouping run for this band; absent if unreported",
			bandLabels, nil,
		),
		lastDCARunAt: prometheus.NewDesc(
			"wnc_rrm_last_dca_run_timestamp_seconds",
			"Instant of the last DCA run for this band; absent if unreported",
			bandLabels, nil,
		),
	}
}

// describe sends every descriptor of the band-keyed RRM run instants.
func (d *apRRMDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.lastRFGroupingRunAt
	ch <- d.lastDCARunAt
}

// collect publishes the run instants of every band the controller reports one for.
//
// It runs once per scrape rather than inside the per-radio loop. The band label is the
// whole identifier, so emitting from that loop would repeat every series with the same
// label set, which Gather rejects by failing the entire endpoint.
//
// Each guard withholds only what it guards. A record whose PHY type cannot be named
// carries no identity, for the reason RRMWorstBand records. The group container and the
// channel assignment container inside it are both optional on the wire, and reading
// through an absent one ends this collector's scrape where it stands, which costs every
// series after it with nothing but a log line to say so. The channel assignment guard
// comes last so that a band the controller sent a grouping instant for keeps it.
func (d *apRRMDescs) collect(ch chan<- prometheus.Metric, records []rrm.MainData) {
	for i := range records {
		record := &records[i]

		band, ok := RRMMainBand(record)
		if !ok {
			continue
		}

		if record.Grp == nil {
			continue
		}

		emitTimestamp(ch, d.lastRFGroupingRunAt, record.Grp.LastRun, band)

		if record.Grp.DCA == nil {
			continue
		}

		emitTimestamp(ch, d.lastDCARunAt, record.Grp.DCA.DCALastRun, band)
	}
}
