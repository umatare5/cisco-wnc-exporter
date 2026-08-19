// Package collector provides collectors for cisco-wnc-exporter.
// This file holds the CAPWAP join statistics module of the AP collector.
package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
)

// epochYear is the year of the sentinel the controller writes into a timestamp leaf
// for an event that has not happened. It sends 1970-01-01T00:00:00+00:00, which
// parses to a real instant, so IsZero does not recognize it.
const epochYear = 1970

// The controller keeps one set of DTLS counters per tunnel channel in a single
// container, so the two sets are folded into one series carrying this label.
const (
	dtlsChannelControl = "control"
	dtlsChannelData    = "data"
)

// apJoinDescs holds the descriptors of the join module. A nil value means the module
// is disabled, which is what keeps every series of it out of a default scrape.
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
			"Whether the AP holds a CAPWAP session with this controller now "+
				"(0=not joined, 1=joined). The record outlives the session, so the join, "+
				"configuration and DTLS series freeze while this reports 0, while the "+
				"discovery series keep advancing for as long as the AP still reaches the controller",
			apLabels, nil,
		),
		name: prometheus.NewDesc(
			"wnc_ap_join_info",
			"AP name as its CAPWAP join record reports it, always 1. The record outlives "+
				"the session, so this names an AP the AP inventory no longer carries",
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
			"Unix timestamp of the last CAPWAP connection error recorded for this AP, "+
				"in the phase wnc_ap_last_error_phase reports",
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
			"Reason for this AP's last CAPWAP discovery failure, as the value the controller's "+
				"own enumeration assigns its spelling. 0 is disc-fail-none, which reports that "+
				"no discovery has failed",
			apLabels, nil,
		),
		lastJoinFailureReason: prometheus.NewDesc(
			"wnc_ap_last_join_failure_reason",
			"Reason for this AP's last CAPWAP join failure, as the value the controller's own "+
				"enumeration assigns its spelling. 0 is jf-none, which reports that no join "+
				"has failed",
			apLabels, nil,
		),
		lastConfigFailureReason: prometheus.NewDesc(
			"wnc_ap_last_config_failure_reason",
			"Reason for this AP's last CAPWAP configuration failure, as the value the "+
				"controller's own enumeration assigns its spelling. 0 is cf-none, which reports "+
				"that no configuration has failed",
			apLabels, nil,
		),
		lastErrorPhase: prometheus.NewDesc(
			"wnc_ap_last_error_phase",
			"CAPWAP phase of this AP's last connection error, as the value the controller's "+
				"own enumeration assigns its spelling (0=ap-con-failure-unknown, "+
				"1=ap-con-failure-discovery, 2=ap-con-failure-dtls, 3=ap-con-failure-join, "+
				"4=ap-con-failure-config, 5=ap-con-failure-imgdwnld, 6=ap-con-failure-run). "+
				"0 reports that the phase is unknown rather than that nothing failed. It "+
				"freezes with the record, and an AP that is not joined reports the same 6 as "+
				"one that is",
			apLabels, nil,
		),
		lastDTLSFailureReason: prometheus.NewDesc(
			"wnc_ap_last_dtls_failure_reason",
			"Reason for the last DTLS handshake outcome on the channel label, as the value "+
				"the controller's own enumeration assigns its spelling. 0 is dtls-hs-success, "+
				"which is also what a channel carrying no session reports",
			channelLabels, nil,
		),
		lastRebootReason: prometheus.NewDesc(
			"wnc_ap_last_reboot_reason",
			"Reason this AP last rebooted as the AP reported it, as the value the controller's "+
				"own enumeration assigns its spelling. 0 is ap-reboot-reason-none",
			apLabels, nil,
		),
		lastDisconnectReason: prometheus.NewDesc(
			"wnc_ap_last_disconnect_reason",
			"Reason this AP last left CAPWAP, as the value the controller's own enumeration "+
				"assigns its spelling. 0 is the enumeration's own unknown member rather than "+
				"the absence of a disconnect",
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
	ch <- prometheus.MustNewConstMetric(
		d.joined, prometheus.GaugeValue, boolToFloat64(record.ApJoinInfo.IsJoined), record.WtpMAC,
	)

	// An empty name reads as no label at all, and this is the only container that
	// still names an AP the inventory has dropped, so there is nothing to fall back on.
	if name := record.ApJoinInfo.ApName; name != "" {
		ch <- prometheus.MustNewConstMetric(d.name, prometheus.GaugeValue, 1, record.WtpMAC, name)
	}
}

// collectCounters publishes the phase counters, including one series per DTLS channel.
func (d *apJoinDescs) collectCounters(ch chan<- prometheus.Metric, record *ap.ApJoinStats) {
	join := record.ApJoinInfo
	discovery := record.ApDiscoveryInfo
	dtls := record.DTLSSessInfo

	for _, metric := range []Float64Metric{
		{d.discoveryRequests, float64(discovery.NumDiscoveryReqRecvd)},
		{d.discoveryResponses, float64(discovery.NumSuccDiscRespSent)},
		{d.discoveryErrors, float64(discovery.NumErrDiscReq)},
		{d.joinRequests, float64(join.NumJoinReqRecvd)},
		{d.joinResponses, float64(join.NumSuccJoinRespSent)},
		{d.joinFailures, float64(join.NumUnsuccJoinReqProcn)},
		{d.configRequests, float64(join.NumConfigReqRecvd)},
		{d.configResponses, float64(join.NumSuccConfRespSent)},
		{d.configFailures, float64(join.NumUnsuccConfReqProcn)},
	} {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc, prometheus.CounterValue, metric.Value, record.WtpMAC,
		)
	}

	for _, channel := range []struct {
		label   string
		metrics []Float64Metric
	}{
		{dtlsChannelControl, []Float64Metric{
			{d.dtlsRequests, float64(dtls.CtrlDTLSSetupReq)},
			{d.dtlsSuccesses, float64(dtls.CtrlDTLSSuccess)},
			{d.dtlsFailures, float64(dtls.CtrlDTLSFailure)},
			{d.dtlsDecryptErrors, float64(dtls.CtrlDTLSDecryptErr)},
			{d.dtlsAntiReplayError, float64(dtls.CtrlDTLSAntiReplayErr)},
		}},
		{dtlsChannelData, []Float64Metric{
			{d.dtlsRequests, float64(dtls.DataDTLSSetupReq)},
			{d.dtlsSuccesses, float64(dtls.DataDTLSSuccess)},
			{d.dtlsFailures, float64(dtls.DataDTLSFailure)},
			{d.dtlsDecryptErrors, float64(dtls.DataDTLSDecryptErr)},
			{d.dtlsAntiReplayError, float64(dtls.DataDTLSAntiReplayErr)},
		}},
	} {
		for _, metric := range channel.metrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc, prometheus.CounterValue, metric.Value, record.WtpMAC, channel.label,
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

// emitTimestamp publishes the instant as a Unix timestamp gauge, and publishes
// nothing for the epoch sentinel. A zero there would read as an event in 1970, and
// time() minus the series as five decades.
func emitTimestamp(ch chan<- prometheus.Metric, desc *prometheus.Desc, at time.Time, labels ...string) {
	if at.Year() <= epochYear {
		return
	}

	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(at.Unix()), labels...)
}
