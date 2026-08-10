package collector

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

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
	typeClientCommonOperData  = "client_common_oper_data"
	typeClientDCInfo          = "client_dc_info"
	typeClientDot11OperData   = "client_dot11_oper_data"
	typeClientSISFDBMac       = "client_sisf_db_mac"
	typeClientTrafficStats    = "client_traffic_stats"
	typeClientMMIFHistory     = "client_mm_if_client_history"
	typeRRMMeasurement        = "rrm_measurement"
	typeRRMCoverage           = "rrm_coverage"
	typeRRMAPDot11RadarData   = "rrm_ap_dot11_radar_data"
	typeWLANCfgEntries        = "wlan_cfg_entries"
	typeWLANPolicies          = "wlan_policies"
	typeWLANPolicyListEntries = "wlan_policy_list_entries"
)

var allDataTypes = []string{
	typeAPCAPWAPData, typeAPOperData, typeAPRadioOperData, typeAPNameMACMap,
	typeAPRadioOperStats, typeAPRadioResetStats,
	typeClientCommonOperData, typeClientDCInfo, typeClientDot11OperData,
	typeClientSISFDBMac, typeClientTrafficStats, typeClientMMIFHistory,
	typeRRMMeasurement, typeRRMCoverage, typeRRMAPDot11RadarData,
	typeWLANCfgEntries, typeWLANPolicies, typeWLANPolicyListEntries,
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
)

// fixtureSource serves one snapshot to every adapter in internal/wnc.
type fixtureSource struct {
	data *wnc.WNCDataCache
}

func (f fixtureSource) GetCachedData(context.Context) (*wnc.WNCDataCache, error) {
	return f.data, nil
}

// TestCollectors_OmitSeriesWhenDataTypeFails is the regression test for the whole
// change: a data type that failed to fetch must make its derived series disappear
// rather than report a zero, a NaN or a timestamp near the Unix epoch. Prometheus
// cannot distinguish a fabricated zero from a measured one.
func TestCollectors_OmitSeriesWhenDataTypeFails(t *testing.T) {
	t.Parallel()

	policyDerived := []string{
		"wnc_wlan_session_timeout_seconds",
		"wnc_wlan_central_switching_enabled",
		"wnc_wlan_central_authentication_enabled",
		"wnc_wlan_central_dhcp_enabled",
		"wnc_wlan_central_association_enabled",
	}
	clientTrafficDerived := []string{
		"wnc_client_rssi_dbm",
		"wnc_client_snr_decibels",
		"wnc_client_speed_mbps",
		"wnc_client_spatial_streams",
		"wnc_client_mcs_index",
		"wnc_client_power_save_state",
		"wnc_client_retry_ratio_percent",
		"wnc_client_rx_bytes_total",
		"wnc_client_tx_bytes_total",
		"wnc_client_rx_packets_total",
		"wnc_client_tx_packets_total",
		"wnc_wlan_rx_bytes_total",
		"wnc_wlan_tx_bytes_total",
	}

	tests := []struct {
		dataType string
		absent   []string
	}{
		{typeAPCAPWAPData, []string{"wnc_ap_config_state", "wnc_ap_uptime_seconds"}},
		{typeAPOperData, []string{"wnc_ap_cpu_utilization_percent", "wnc_ap_memory_utilization_percent"}},
		{typeAPRadioOperData, []string{
			"wnc_ap_radio_state", "wnc_ap_channel_number", "wnc_ap_clients", "wnc_ap_info",
		}},
		{typeAPNameMACMap, []string{"wnc_ap_clients"}},
		{typeAPRadioOperStats, []string{
			"wnc_ap_rx_packets_total", "wnc_ap_tx_bytes_total",
			"wnc_ap_rx_errors_total", "wnc_ap_fcs_errors_total",
		}},
		{typeAPRadioResetStats, []string{"wnc_ap_radio_reset_total"}},
		{typeClientCommonOperData, []string{
			"wnc_client_state", "wnc_client_info", "wnc_ap_clients",
			"wnc_wlan_clients_total", "wnc_wlan_rx_bytes_total", "wnc_wlan_tx_bytes_total",
		}},
		{typeClientDot11OperData, []string{"wnc_client_protocol", "wnc_client_uptime_seconds"}},
		{typeClientTrafficStats, clientTrafficDerived},
		{typeClientMMIFHistory, []string{"wnc_client_state_transition_seconds"}},
		{typeRRMMeasurement, []string{
			"wnc_ap_channel_utilization_percent", "wnc_ap_rx_utilization_percent",
			"wnc_ap_tx_utilization_percent", "wnc_ap_noise_utilization_percent",
			"wnc_ap_noise_floor_dbm",
		}},
		{typeRRMCoverage, []string{"wnc_ap_coverage_hole_events_total"}},
		{typeRRMAPDot11RadarData, []string{"wnc_ap_last_radar_on_radio_at"}},
		{typeWLANCfgEntries, []string{
			"wnc_wlan_enabled", "wnc_wlan_clients_total", "wnc_wlan_auth_psk_enabled", "wnc_wlan_info",
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

// TestCollectors_FailedDataTypeNeverAddsSeries covers all eighteen data types,
// including the ones that only contribute info labels.
func TestCollectors_FailedDataTypeNeverAddsSeries(t *testing.T) {
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

// gatherAllCollectors registers the three data collectors over one snapshot in
// which the named data type is marked as failed, and reports which metric names
// carry at least one series. An empty failedDataType gathers the baseline.
func gatherAllCollectors(t *testing.T, failedDataType string) map[string]bool {
	t.Helper()

	data := fullFixtureSnapshot()
	if failedDataType != "" {
		data.FetchErrors[failedDataType] = errors.New("fetch failed")
	}
	src := fixtureSource{data: data}

	apMetrics := APMetrics{General: true, Radio: true, Traffic: true, Errors: true, Info: true}
	clientMetrics := ClientMetrics{General: true, Radio: true, Traffic: true, Errors: true, Info: true}
	wlanMetrics := WLANMetrics{General: true, Traffic: true, Config: true, Info: true}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		NewAPCollector(
			wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src), apMetrics,
		),
		NewClientCollector(wnc.NewClientSource(src), clientMetrics),
		NewWLANCollector(wnc.NewWLANSource(src), wnc.NewClientSource(src), wlanMetrics),
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	present := make(map[string]bool, len(families))
	for _, family := range families {
		present[family.GetName()] = len(family.GetMetric()) > 0
	}
	return present
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
		}},
		ApOperData: []ap.OperData{{
			WtpMAC:     fixtureAPMAC,
			RadioID:    0,
			ApSysStats: &ap.ApSystemStats{CPUUsage: 12, MemoryUsage: 34},
		}},
		RadioOperData: []ap.RadioOperData{{
			WtpMAC:            fixtureAPMAC,
			RadioSlotID:       0,
			RadioType:         "radio-80211bg",
			CurrentBandID:     fixtureBandID,
			CurrentActiveBand: "dot11-2-dot-4-ghz-band",
			AdminState:        APAdminStateEnabled,
			OperState:         APRadioStateUp,
			PhyHtCfg:          &ap.PhyHtCfg{CfgData: ap.PhyHtCfgData{CurrFreq: fixtureChannel, ChanWidth: 20}},
			RadioBandInfo: []ap.RadioBandInfo{{
				BandID: fixtureBandID,
				PhyTxPwrLvlCfg: ap.PhyTxPwrLvlCfg{
					CfgData: ap.PhyTxPwrLvlCfgData{CurrTxPowerInDbm: 14, TxPowerLevel1: 20},
				},
			}},
		}},
		RadioOperStats: []ap.RadioOperStats{{
			ApMAC:            fixtureAPMAC,
			SlotID:           0,
			RxDataFrameCount: 100,
			TxDataFrameCount: 200,
			FcsErrorCount:    1,
		}},
		RadioResetStats: []ap.RadioResetStats{{ApMAC: fixtureAPMAC, RadioID: 0, Count: 2}},
		NameMACMaps:     []ap.ApNameMACMap{{WtpName: fixtureAPName, WtpMAC: fixtureAPMAC, EthMAC: fixtureAPMAC}},

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
			MsMACAddress:   fixtureClientMAC,
			BytesRx:        "1000",
			BytesTx:        "2000",
			PktsRx:         "10",
			PktsTx:         "20",
			DataRetries:    "1",
			TxRetries:      "2",
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
			Load:        &rrm.Load{CcaUtilPercentage: 30, RxUtilPercentage: 10, TxUtilPercentage: 5},
			Noise: &rrm.Noise{
				Noise: rrm.NoiseData{NoiseData: []rrm.NoiseDataItem{
					{Chan: 1, Noise: -95},
					{Chan: fixtureChannel, Noise: -90},
				}},
			},
		}},
		RRMCoverage: []rrm.RRMCoverage{{
			WtpMAC: fixtureAPMAC, RadioSlotID: 0, FailedClientCount: 1,
		}},
		ApDot11RadarData: []rrm.ApDot11RadarData{{
			WtpMAC:           fixtureAPMAC,
			RadioSlotID:      0,
			LastRadarOnRadio: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		}},

		WLANConfigEntries: []wlan.WlanCfgEntry{{
			WlanID:         1,
			ProfileName:    fixtureProfile,
			AuthKeyMgmtPsk: true,
			WPA2Enabled:    true,
			APFVapIDData:   &wlan.APFVapIDData{SSID: "TestWLAN", WlanStatus: true},
		}},
		WLANPolicies: []wlan.WlanPolicy{{
			PolicyProfileName:   fixturePolicy,
			WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{CentralSwitching: true},
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
