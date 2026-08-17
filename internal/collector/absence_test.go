package collector

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// Data type identifiers as published in the data label. They are string literals
// on purpose: this test asserts the contract the README documents, not the
// constants, and a mistyped one shows up as every series staying present.
const (
	typeAPCAPWAPData          = "ap_capwap_data"
	typeAPOperData            = "ap_oper_data"
	typeAPRadioOperData       = "ap_radio_oper_data"
	typeAPNameMACMap          = "ap_name_mac_map"
	typeAPRadioOperStats      = "ap_radio_oper_stats"
	typeAPRadioResetStats     = "ap_radio_reset_stats"
	typeAPJoinStats           = "ap_join_stats"
	typeControllerBootTime    = "controller_boot_time"
	typeCoClientDelReason     = "co_client_del_reason"
	typeClientRoamingStats    = "client_roaming_stats"
	typeClientCommonOperData  = "client_common_oper_data"
	typeClientDCInfo          = "client_dc_info"
	typeClientDot11OperData   = "client_dot11_oper_data"
	typeClientSISFDBMac       = "client_sisf_db_mac"
	typeClientTrafficStats    = "client_traffic_stats"
	typeClientMMIFHistory     = "client_mm_if_client_history"
	typeRRMMeasurement        = "rrm_measurement"
	typeRRMCoverage           = "rrm_coverage"
	typeRRMAPDot11RadarData   = "rrm_ap_dot11_radar_data"
	typeRRMRadioSlot          = "rrm_radio_slot"
	typeRRMSpectrumAqTable    = "rrm_spectrum_aq_table"
	typeWLANCfgEntries        = "wlan_cfg_entries"
	typeWLANPolicies          = "wlan_policies"
	typeWLANPolicyListEntries = "wlan_policy_list_entries"
	typeWLANClientStats       = "wlan_client_stats"
)

var allDataTypes = []string{
	typeAPCAPWAPData, typeAPOperData, typeAPRadioOperData, typeAPNameMACMap,
	typeAPRadioOperStats, typeAPRadioResetStats, typeAPJoinStats,
	typeControllerBootTime, typeCoClientDelReason, typeClientRoamingStats,
	typeClientCommonOperData, typeClientDCInfo, typeClientDot11OperData,
	typeClientSISFDBMac, typeClientTrafficStats, typeClientMMIFHistory,
	typeRRMMeasurement, typeRRMCoverage, typeRRMAPDot11RadarData, typeRRMRadioSlot,
	typeRRMSpectrumAqTable,
	typeWLANCfgEntries, typeWLANPolicies, typeWLANPolicyListEntries, typeWLANClientStats,
}

const (
	fixtureAPMAC     = "aa:bb:cc:dd:ee:ff"
	fixtureAPName    = "TEST-AP01"
	fixtureClientMAC = "11:22:33:44:55:66"
	fixtureProfile   = "test-profile"
	fixturePolicy    = "test-policy"

	// fixtureChannel is the channel the AP radio operates on. The RRM noise list
	// must carry an entry for it, and the exporter must not read another entry.
	fixtureChannel = 6
	// fixtureBandID ties radio-band-info to current-band-id.
	fixtureBandID = 0
	// The two state leaves carry distinct spellings so that a descriptor reading
	// the wrong one publishes a label value the assertions do not expect.
	fixturePMFOptions = "apf-vap-pmf-required"
	fixtureFTMode     = "dot11r-disabled"
)

// The eleven timestamps of the join record, one day apart so that every pair is
// distinguishable and a descriptor reading a neighboring leaf reports another day.
var (
	fixtureJoinErrorAt        = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	fixtureJoinSuccessAt      = time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	fixtureJoinFailureAt      = time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)
	fixtureConfigSuccessAt    = time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	fixtureConfigFailureAt    = time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	fixtureDiscoverySuccessAt = time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	fixtureDiscoveryFailureAt = time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	fixtureCtrlDTLSSuccessAt  = time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	fixtureCtrlDTLSFailureAt  = time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	fixtureDataDTLSSuccessAt  = time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
	fixtureDataDTLSFailureAt  = time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	// fixtureEpochSentinel is what the controller writes into a timestamp leaf for an
	// event that has not happened. It is not the zero time, so IsZero reports false.
	fixtureEpochSentinel = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	// fixtureBootTime is the controller boot instant, a day past the join timestamps so
	// that a descriptor reading one of those reports another day.
	fixtureBootTime = "2026-01-13T00:00:00Z"

	// The two delete reasons carry distinct values, and the second is there because one
	// entry cannot tell a per-reason loop from a single emit.
	fixtureDeleteReason      = "ap-delete"
	fixtureOtherDeleteReason = "bssid-down"
)

// fixtureSource serves one snapshot to every adapter in internal/wnc.
type fixtureSource struct {
	data *wnc.WNCDataCache
}

func (f fixtureSource) GetCachedData(context.Context) (*wnc.WNCDataCache, error) {
	return f.data, nil
}

// TestAllCollectors_OmitSeriesWhenDataTypeFails is the regression test for the whole
// change: a data type that failed to fetch must make its derived series disappear
// rather than report a zero, a NaN or a timestamp near the Unix epoch. Prometheus
// cannot distinguish a fabricated zero from a measured one.
func TestAllCollectors_OmitSeriesWhenDataTypeFails(t *testing.T) {
	t.Parallel()

	policyDerived := []string{
		"wnc_wlan_session_timeout_seconds",
		"wnc_wlan_central_switching_enabled",
		"wnc_wlan_central_authentication_enabled",
		"wnc_wlan_central_dhcp_enabled",
		"wnc_wlan_central_association_enabled",
		"wnc_wlan_policy_enabled",
		"wnc_wlan_policy_binding",
	}
	clientTrafficDerived := []string{
		"wnc_client_rssi_dbm",
		"wnc_client_snr_decibels",
		"wnc_client_speed_mbps",
		"wnc_client_spatial_streams",
		"wnc_client_mcs_index",
		"wnc_client_power_save_state",
		"wnc_client_rx_bytes_total",
		"wnc_client_tx_bytes_total",
		"wnc_client_rx_packets_total",
		"wnc_client_tx_packets_total",
		// The error counters share one omission guard, so two of them witness all
		// eleven. wnc_client_rx_group_total also pins the name against a rename.
		"wnc_client_policy_errors_total",
		"wnc_client_rx_group_total",
	}

	tests := []struct {
		dataType string
		absent   []string
	}{
		{typeAPCAPWAPData, []string{
			"wnc_ap_config_state", "wnc_ap_uptime_seconds", "wnc_ap_oper_state",
		}},
		{typeAPOperData, []string{"wnc_ap_cpu_utilization_ratio", "wnc_ap_memory_utilization_ratio"}},
		{typeAPRadioOperData, []string{
			"wnc_ap_radio_state", "wnc_ap_channel_number", "wnc_ap_clients", "wnc_ap_info",
		}},
		{typeAPNameMACMap, []string{"wnc_ap_clients"}},
		{typeAPRadioOperStats, []string{
			"wnc_ap_data_rx_frames_total", "wnc_ap_data_tx_frames_total",
			"wnc_ap_rx_errors_total", "wnc_ap_fcs_errors_total",
		}},
		{typeAPRadioResetStats, []string{"wnc_ap_radio_resets_total"}},
		{typeAPJoinStats, []string{
			"wnc_ap_joined", "wnc_ap_join_info", "wnc_ap_discovery_requests_total",
			"wnc_ap_dtls_session_requests_total", "wnc_ap_last_join_success_timestamp_seconds",
			"wnc_ap_last_dtls_success_timestamp_seconds", "wnc_ap_last_reboot_reason",
		}},
		{typeControllerBootTime, []string{"wnc_controller_boot_time_seconds"}},
		{typeCoClientDelReason, []string{"wnc_controller_client_deletes_total"}},
		{typeClientRoamingStats, []string{
			"wnc_controller_client_ap_auth_roams_total",
			"wnc_controller_client_ap_auth_dot11i_fast_roams_total",
			"wnc_controller_client_ap_auth_dot11i_slow_roams_total",
		}},
		{typeClientCommonOperData, []string{
			"wnc_client_state", "wnc_client_info", "wnc_ap_clients",
			"wnc_wlan_clients",
		}},
		{typeClientDot11OperData, []string{"wnc_client_protocol", "wnc_client_uptime_seconds"}},
		{typeClientTrafficStats, clientTrafficDerived},
		{typeClientMMIFHistory, []string{"wnc_client_state_transition_seconds"}},
		{typeRRMMeasurement, []string{
			"wnc_ap_channel_utilization_ratio", "wnc_ap_rx_utilization_ratio",
			"wnc_ap_tx_utilization_ratio", "wnc_ap_noise_utilization_ratio",
			"wnc_ap_noise_floor_dbm",
		}},
		{typeRRMCoverage, []string{"wnc_ap_coverage_failed_clients"}},
		{typeRRMAPDot11RadarData, []string{"wnc_ap_last_radar_timestamp_seconds"}},
		{typeRRMRadioSlot, []string{"wnc_ap_rrm_profile_passed"}},
		{typeRRMSpectrumAqTable, []string{"wnc_ap_air_quality_index"}},
		{typeWLANCfgEntries, []string{
			"wnc_wlan_enabled", "wnc_wlan_clients", "wnc_wlan_auth_psk_enabled", "wnc_wlan_info",
			"wnc_wlan_pmf_state", "wnc_wlan_ft_state", "wnc_wlan_onboarding_clients",
		}},
		{typeWLANClientStats, []string{
			"wnc_wlan_data_usage_bytes_total", "wnc_wlan_onboarding_clients",
		}},
		{typeWLANPolicies, policyDerived},
		{typeWLANPolicyListEntries, policyDerived},
	}

	baseline := gatherAllCollectors(t, "")

	for _, tt := range tests {
		t.Run(tt.dataType, func(t *testing.T) {
			t.Parallel()

			if !slices.Contains(allDataTypes, tt.dataType) {
				t.Fatalf("%s is not a published data type", tt.dataType)
			}

			// Without this the test would pass on a collector that never emitted the
			// series at all.
			for _, name := range tt.absent {
				if !baseline[name] {
					t.Fatalf("%s is absent from the all-succeed baseline, so its omission proves nothing", name)
				}
			}

			present := gatherAllCollectors(t, tt.dataType)
			for _, name := range tt.absent {
				if present[name] {
					t.Errorf("%s is present while %s failed, want it absent instead of a fabricated value",
						name, tt.dataType)
				}
			}
		})
	}
}

// TestAllCollectors_FailedDataTypeNeverAddsSeries iterates every data type.
// Two of them gate no series, only info label values, and a collector keeps the
// series and empties the label when their fetch fails, so those two pass here
// whatever the collector does. Nothing in this file covers label provenance.
func TestAllCollectors_FailedDataTypeNeverAddsSeries(t *testing.T) {
	t.Parallel()

	baseline := gatherAllCollectors(t, "")

	for _, dataType := range allDataTypes {
		t.Run(dataType, func(t *testing.T) {
			t.Parallel()

			for name := range gatherAllCollectors(t, dataType) {
				if !baseline[name] {
					t.Errorf("%s appears only when %s fails, want the failure to remove series only",
						name, dataType)
				}
			}
		})
	}
}

// TestAllCollectors_MetricNamesMatchTypes binds every family's type to its name.
// Nothing else in the suite asserts a type, and Prometheus reads a counter's drop
// as a reset and extrapolates, so a type regression silently invents data. This
// exporter gives every counter a _total suffix and no other metric one, which is
// the correspondence promlint checks, so the whole surface is covered rather than
// a hand-kept list of the families that regressed once.
func TestAllCollectors_MetricNamesMatchTypes(t *testing.T) {
	t.Parallel()

	families, err := fixtureRegistry(t, "").Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}
	if len(families) == 0 {
		t.Fatal("no families were gathered, so the lint below proves nothing")
	}

	problems, err := promlint.NewWithMetricFamilies(families).Lint()
	if err != nil {
		t.Fatalf("Lint() error = %v, want nil", err)
	}
	for _, problem := range problems {
		t.Errorf("%s: %s", problem.Metric, problem.Text)
	}
}

// gatherAllCollectors reports which metric names carry at least one series when
// the named data type is marked as failed. An empty failedDataType gathers the
// baseline.
func gatherAllCollectors(t *testing.T, failedDataType string) map[string]bool {
	t.Helper()

	families, err := fixtureRegistry(t, failedDataType).Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	present := make(map[string]bool, len(families))
	for _, family := range families {
		present[family.GetName()] = len(family.GetMetric()) > 0
	}
	return present
}

// fixtureRegistry registers the three data collectors over one snapshot in which
// the named data type is marked as failed. An empty failedDataType leaves every
// fetch succeeding.
func fixtureRegistry(t *testing.T, failedDataType string) *prometheus.Registry {
	t.Helper()

	data := fullFixtureSnapshot()
	if failedDataType != "" {
		data.FetchErrors[failedDataType] = errors.New("fetch failed")
	}

	registry := prometheus.NewRegistry()
	for _, collector := range fixtureCollectors(t, data) {
		registry.MustRegister(collector)
	}
	return registry
}

// fixtureCollectors returns every data collector over the given snapshot, with every
// module enabled. fixtureRegistry gathers them; a test that needs the metrics before
// the registry encodes them collects from them directly.
func fixtureCollectors(t *testing.T, data *wnc.WNCDataCache) []prometheus.Collector {
	t.Helper()

	src := fixtureSource{data: data}

	apMetrics := APMetrics{
		General: true, Radio: true, Traffic: true, Errors: true, Join: true,
		Spectrum: true, Info: true,
	}
	clientMetrics := ClientMetrics{General: true, Radio: true, Traffic: true, Errors: true, Info: true}
	wlanMetrics := WLANMetrics{General: true, Traffic: true, Config: true, Info: true}

	return []prometheus.Collector{
		NewControllerCollector(wnc.NewControllerSource(src), ControllerMetrics{General: true}),
		NewAPCollector(
			wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src), apMetrics,
		),
		NewClientCollector(wnc.NewClientSource(src), clientMetrics),
		NewWLANCollector(wnc.NewWLANSource(src), wnc.NewClientSource(src), wlanMetrics),
	}
}

// fullFixtureSnapshot returns a snapshot in which every data type carries one
// entry, so every series exists in the baseline.
func fullFixtureSnapshot() *wnc.WNCDataCache {
	return &wnc.WNCDataCache{
		FetchErrors: map[string]error{},
		RefreshedAt: time.Now(),

		CAPWAPData: []ap.CAPWAPData{{
			WtpMAC:     fixtureAPMAC,
			IPAddr:     "192.0.2.11",
			Name:       fixtureAPName,
			ApTimeInfo: ap.ApTimeInfo{BootTime: "2026-01-01T00:00:00Z"},
			ApState:    ap.ApState{ApOperationState: "registered"},
		}},
		ApOperData: []ap.OperData{{
			WtpMAC:  fixtureAPMAC,
			RadioID: 0,
			// The averaged siblings carry their own values so that a descriptor
			// reading them, or summing them into the current reading, reports a
			// number the value assertions do not expect.
			ApSysStats: &ap.ApSystemStats{
				CPUUsage: 12, MemoryUsage: 34,
				AvgCPUUsage: 56, AvgMemoryUsage: 78,
			},
		}},
		RadioOperData: []ap.RadioOperData{{
			WtpMAC:            fixtureAPMAC,
			RadioSlotID:       0,
			RadioType:         "radio-80211bg",
			CurrentBandID:     fixtureBandID,
			CurrentActiveBand: "dot11-2-dot-4-ghz-band",
			AdminState:        APAdminStateEnabled,
			OperState:         APRadioStateUp,
			// The width differs from every transmit power level below so that a
			// descriptor reading another leaf of this radio reports another number.
			PhyHtCfg: &ap.PhyHtCfg{CfgData: ap.PhyHtCfgData{CurrFreq: fixtureChannel, ChanWidth: 40}},
			RadioBandInfo: []ap.RadioBandInfo{{
				BandID: fixtureBandID,
				PhyTxPwrLvlCfg: ap.PhyTxPwrLvlCfg{
					CfgData: ap.PhyTxPwrLvlCfgData{CurrTxPowerInDbm: 14, TxPowerLevel1: 20},
				},
			}},
		}},
		// Every counter leaf carries a distinct value, including the ones no
		// descriptor reads: those are the numbers a mis-wired or a summing
		// descriptor would surface. wep-undecryptable-count stays for that reason
		// after its series was withdrawn.
		RadioOperStats: []ap.RadioOperStats{{
			ApMAC:  fixtureAPMAC,
			SlotID: 0,

			RxDataFrameCount:    3101,
			TxDataFrameCount:    3102,
			RxMgmtFrameCount:    3103,
			TxMgmtFrameCount:    3104,
			RxCtrlFrameCount:    3105,
			TxCtrlFrameCount:    3106,
			MulticastRxFrameCnt: 3107,
			MulticastTxFrameCnt: 3108,
			TxFrameCount:        3109,
			RtsSuccessCount:     3110,

			RxErrorFrameCount:     3201,
			RetryCount:            3202,
			FailedCount:           3203,
			FrameDuplicateCount:   3204,
			FcsErrorCount:         3205,
			RxFragmentCount:       3206,
			TxFragmentCount:       3207,
			RtsFailureCount:       3208,
			MACDecryErrFrameCount: 3209,
			MACMicErrFrameCount:   3210,
			WepUndecryptableCount: 3211,

			AckFailureCount:    3901,
			MultipleRetryCount: 3902,
			RxDataPktCount:     3903,
			TxDataPktCount:     3904,
		}},
		// A controller returns several entries for one radio, which the collector
		// totals. A single entry cannot tell a total from an overwrite.
		RadioResetStats: []ap.RadioResetStats{
			{ApMAC: fixtureAPMAC, RadioID: 0, Cause: "test-cause-1", Count: 3},
			{ApMAC: fixtureAPMAC, RadioID: 0, Cause: "test-cause-2", Count: 5},
		},
		NameMACMaps: []ap.ApNameMACMap{{WtpName: fixtureAPName, WtpMAC: fixtureAPMAC, EthMAC: fixtureAPMAC}},
		JoinStats:   []ap.ApJoinStats{newFixtureJoinStats()},

		ControllerBootTime: fixtureBootTime,
		ClientDeleteReasons: map[string]float64{
			fixtureDeleteReason:      6101,
			fixtureOtherDeleteReason: 6102,
		},
		// Three of the container's thirteen leaves are published. The rest carry values
		// here as well, so a descriptor reading one of those reports a number no
		// assertion expects.
		ClientRoamingStats: map[string]float64{
			"ap-auth-roams":            6201,
			"ap-auth-dot11i-fast-roam": 6202,
			"ap-auth-dot11i-slow-roam": 6203,
			"total-roam":               6204,
			"roam-fail":                6205,
			"dot11r-roam":              6206,
		},

		CommonOperData: []client.CommonOperData{{
			ClientMAC:   fixtureClientMAC,
			ApName:      fixtureAPName,
			MsApSlotID:  0,
			WlanID:      1,
			CoState:     ClientStatusRun,
			MsRadioType: "client-dot11ax-24ghz-prot",
		}},
		DCInfo: []client.DcInfo{{ClientMAC: fixtureClientMAC, DeviceType: "Un-Classified Device"}},
		Dot11OperData: []client.Dot11OperData{{
			MsMACAddress:  fixtureClientMAC,
			Dot11State:    "associated",
			VapSsid:       "TestWLAN",
			RadioType:     "dot11-radio-type-bg",
			EwlcMsPhyType: "client-dot11ax-24ghz-prot",
			MsAssocTime:   time.Now().Add(-time.Hour),
		}},
		SisfDBMac: []client.SisfDBMac{{MACAddr: fixtureClientMAC}},
		TrafficStats: []client.TrafficStats{{
			MsMACAddress: fixtureClientMAC,

			BytesRx: "4101",
			BytesTx: "4102",
			PktsRx:  "4103",
			PktsTx:  "4104",

			PolicyErrs:         "4201",
			DuplicateRcv:       "4202",
			DecryptFailed:      "4203",
			MicMismatch:        "4204",
			MicMissing:         "4205",
			TxExcessiveRetries: "4206",
			RxGroupCounter:     "4207",
			TxTotalDrops:       "4208",
			DataRetries:        "4209",
			RtsRetries:         "4210",
			TxRetries:          "4211",

			MostRecentRSSI: -55,
			MostRecentSNR:  35,
			Speed:          866,
			SpatialStream:  2,
			CurrentRate:    "m9 ss2",
			PowerSaveState: 1,
		}},
		MmIfClientHistory: []client.MmIfClientHistory{newFixtureMobilityHistory()},

		RRMMeasurements: []rrm.RRMMeasurement{{
			WtpMAC:      fixtureAPMAC,
			RadioSlotID: 0,
			Load: &rrm.Load{
				CcaUtilPercentage: 30, RxUtilPercentage: 10,
				TxUtilPercentage: 5, RxNoiseChannelUtilization: 22,
			},
			Noise: &rrm.Noise{
				Noise: rrm.NoiseData{NoiseData: []rrm.NoiseDataItem{
					{Chan: 1, Noise: -95},
					{Chan: fixtureChannel, Noise: -90},
				}},
			},
		}},
		RRMCoverage: []rrm.RRMCoverage{{
			WtpMAC: fixtureAPMAC, RadioSlotID: 0, FailedClientCount: 7,
		}},
		RadioSlots: []rrm.RadioSlot{{
			WtpMAC:      fixtureAPMAC,
			RadioSlotID: 0,
			RadioData: &rrm.RadioData{
				CoverageProfilePassed:     true,
				LoadProfPassed:            false,
				InterferenceProfilePassed: false,
				NoiseProfilePassed:        true,
			},
		}},
		SpectrumAqTable: []rrm.SpectrumAqTable{{
			WtpMAC: fixtureAPMAC,
			Band:   "dot11-2-dot-4-ghz-band",
			PerRadioAqData: &rrm.PerRadioAqData{
				ChannelCount: 3,
				PerChannelAqList: []rrm.PerChannelAqList{
					{ChannelNum: 0, Aqi: 0, MinAqi: 0},
					{ChannelNum: fixtureChannel + 1, Aqi: 91, MinAqi: 90},
					{ChannelNum: fixtureChannel, Aqi: 93, MinAqi: 92},
				},
			},
		}},
		ApDot11RadarData: []rrm.ApDot11RadarData{{
			WtpMAC:           fixtureAPMAC,
			RadioSlotID:      0,
			LastRadarOnRadio: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		}},

		// The record carries every leaf of the container. The four onboarding phases each
		// carry a distinct number, so a descriptor pointed at the wrong phase reports a
		// value the assertions do not expect while every count and label stays intact.
		// The run count and the random-MAC count remain unread by any series.
		WLANClientStats: []ap.WlanClientStats{{
			WlanID:                  1,
			WlanProfileName:         fixtureProfile,
			DataUsage:               "7101",
			TotalRandomMACClients:   7102,
			ClientCurrStateL2Auth:   7103,
			ClientCurrStateMobility: 7104,
			ClientCurrStateIplearn:  7105,
			CurrStateWebauthPending: 7106,
			ClientCurrStateRun:      7107,
		}},
		WLANConfigEntries: []wlan.WlanCfgEntry{{
			WlanID:         1,
			ProfileName:    fixtureProfile,
			AuthKeyMgmtPsk: true,
			WPA2Enabled:    true,
			PMFOptions:     fixturePMFOptions,
			FTMode:         fixtureFTMode,
			APFVapIDData:   &wlan.APFVapIDData{SSID: "TestWLAN", WlanStatus: true},
		}},
		WLANPolicies: []wlan.WlanPolicy{{
			PolicyProfileName:   fixturePolicy,
			WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{CentralSwitching: true},
			WlanTimeout:         &wlan.WlanTimeout{SessionTimeout: 1800},
		}},
		WLANPolicyListEntries: []wlan.PolicyListEntry{{
			TagName: "test-tag",
			WLANPolicies: &wlan.WLANPolicies{WLANPolicy: []wlan.WLANPolicyMap{{
				WLANProfileName:   fixtureProfile,
				PolicyProfileName: fixturePolicy,
			}}},
		}},
	}
}

// newFixtureJoinStats fills one join record. Every numeric leaf carries a distinct
// value and every timestamp a distinct instant, including the ones no descriptor
// reads, so a descriptor wired to a neighboring leaf reports a number the value
// assertions do not expect. All eleven timestamps are real here: the epoch sentinel
// the controller writes for an event that has not happened is exercised in
// TestAPJoinModule_WithholdsTheEpochSentinel, which needs the series present in this
// baseline to prove its absence there.
func newFixtureJoinStats() ap.ApJoinStats {
	return ap.ApJoinStats{
		WtpMAC: fixtureAPMAC,
		ApJoinInfo: ap.ApJoinInfo{
			ApName:   fixtureAPName,
			IsJoined: true,

			// Neither address is published: the record is keyed by wtp-mac, and an IP in a
			// counter's labels would turn a DHCP lease change into a counter reset.
			ApIPAddr:      "192.0.2.99",
			ApEthernetMAC: "aa:bb:cc:dd:ee:00",

			NumJoinReqRecvd:       5101,
			NumSuccJoinRespSent:   5102,
			NumUnsuccJoinReqProcn: 5103,
			NumConfigReqRecvd:     5104,
			NumSuccConfRespSent:   5105,
			NumUnsuccConfReqProcn: 5106,

			LastJoinFailureType:   "jf-none",
			LastConfigFailureType: "cf-none",
			LastErrorType:         "ap-con-failure-run",
			// Free text with no value domain, so no series reads it.
			LastMsgDecrFailReason: "fixture decryption failure text",

			LastErrorTime:         fixtureJoinErrorAt,
			LastSuccJoinAtmptTime: fixtureJoinSuccessAt,
			LastFailJoinAtmptTime: fixtureJoinFailureAt,
			LastSuccConfAtmptTime: fixtureConfigSuccessAt,
			LastFailConfAtmptTime: fixtureConfigFailureAt,
		},
		ApDiscoveryInfo: ap.ApDiscoveryInfo{
			WtpMAC:      fixtureAPMAC,
			EthernetMAC: "aa:bb:cc:dd:ee:00",
			ApIPAddress: "192.0.2.99",

			NumDiscoveryReqRecvd: 5201,
			NumSuccDiscRespSent:  5202,
			NumErrDiscReq:        5203,
			LastDiscFailureType:  "disc-fail-none",
			LastSuccessDiscTime:  fixtureDiscoverySuccessAt,
			LastFailedDiscTime:   fixtureDiscoveryFailureAt,
		},
		DTLSSessInfo: ap.DTLSSessInfo{
			MACAddr: fixtureAPMAC,

			CtrlDTLSSetupReq:      5301,
			CtrlDTLSSuccess:       5302,
			CtrlDTLSFailure:       5303,
			CtrlDTLSDecryptErr:    5304,
			CtrlDTLSAntiReplayErr: 5305,

			DataDTLSSetupReq:      5401,
			DataDTLSSuccess:       5402,
			DataDTLSFailure:       5403,
			DataDTLSDecryptErr:    5404,
			DataDTLSAntiReplayErr: 5405,

			CtrlDTLSFailureType: "dtls-hs-success",
			DataDTLSFailureType: "dtls-hs-fragment-error",

			CtrlDTLSSuccessTime: fixtureCtrlDTLSSuccessAt,
			CtrlDTLSFailureTime: fixtureCtrlDTLSFailureAt,
			DataDTLSSuccessTime: fixtureDataDTLSSuccessAt,
			DataDTLSFailureTime: fixtureDataDTLSFailureAt,
		},
		RebootReason:     "ap-reboot-reason-reboot-cmd",
		DisconnectReason: "wtp-controller-initiated-reason",
		// Free English prose, so no series reads it.
		ApDisconnectReason: "Tag modified",
	}
}

// newFixtureMobilityHistory fills the anonymous nested entry the SDK declares.
func newFixtureMobilityHistory() client.MmIfClientHistory {
	history := client.MmIfClientHistory{ClientMAC: fixtureClientMAC}
	history.MobilityHistory.Entry = append(history.MobilityHistory.Entry, struct {
		InstanceID    int       `json:"instance-id"`
		MsApSlotID    int       `json:"ms-ap-slot-id"`
		MsAssocTime   time.Time `json:"ms-assoc-time"`
		Role          string    `json:"role"`
		Bssid         string    `json:"bssid"`
		ApName        string    `json:"ap-name"`
		RunLatency    int       `json:"run-latency"`
		Dot11RoamType string    `json:"dot11-roam-type"`
	}{
		ApName:     fixtureAPName,
		Role:       "mm-client-role-local",
		RunLatency: 120,
	})
	return history
}
