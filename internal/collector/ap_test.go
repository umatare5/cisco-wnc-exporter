package collector

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

func TestNewAPCollector(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		metrics APMetrics
		wantNil bool
	}{
		{
			"All modules disabled",
			APMetrics{},
			false,
		},
		{
			"General module enabled",
			APMetrics{General: true},
			false,
		},
		{
			"Radio module enabled",
			APMetrics{Radio: true},
			false,
		},
		{
			"Traffic module enabled",
			APMetrics{Traffic: true},
			false,
		},
		{
			"Errors module enabled",
			APMetrics{Errors: true},
			false,
		},
		{
			"Info module enabled",
			APMetrics{Info: true},
			false,
		},
		{
			"All modules enabled",
			APMetrics{
				General: true,
				Radio:   true,
				Traffic: true,
				Errors:  true,
				Info:    true,
			},
			false,
		},
		{
			"Info module with labels",
			APMetrics{
				Info:       true,
				InfoLabels: []string{"name", "ip", "band"},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := NewAPCollector(nil, nil, nil, tt.metrics)

			if tt.wantNil {
				if collector != nil {
					t.Errorf("NewAPCollector() = %v, want nil", collector)
				}
			} else {
				if collector == nil {
					t.Errorf("NewAPCollector() = nil, want non-nil")
					return
				}
				if collector.metrics.General != tt.metrics.General {
					t.Errorf(
						"NewAPCollector() General = %v, want %v",
						collector.metrics.General,
						tt.metrics.General,
					)
				}
				if collector.metrics.Info != tt.metrics.Info {
					t.Errorf("NewAPCollector() Info = %v, want %v", collector.metrics.Info, tt.metrics.Info)
				}
			}
		})
	}
}

func TestAPCollector_isAnyMetricFlagEnabled(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		metrics  APMetrics
		expected bool
	}{
		{
			"All disabled",
			APMetrics{},
			false,
		},
		{
			"General enabled",
			APMetrics{General: true},
			true,
		},
		{
			"Radio enabled",
			APMetrics{Radio: true},
			true,
		},
		{
			"Traffic enabled",
			APMetrics{Traffic: true},
			true,
		},
		{
			"Errors enabled",
			APMetrics{Errors: true},
			true,
		},
		{
			"Info enabled",
			APMetrics{Info: true},
			true,
		},
		{
			"Multiple enabled",
			APMetrics{General: true, Radio: true},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := &APCollector{metrics: tt.metrics}
			if got := collector.isAnyMetricFlagEnabled(); got != tt.expected {
				t.Errorf("isAnyMetricFlagEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPCollector_Describe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		metrics     APMetrics
		expectDescs int
	}{
		{
			"No modules enabled",
			APMetrics{},
			0,
		},
		{
			"General module only",
			APMetrics{General: true},
			9, // radio_state, admin_state, oper_state, config_state, uptime, cpu_current, cpu_avg, mem_current, mem_avg
		},
		{
			"Radio module only",
			APMetrics{Radio: true},
			10, // channel, channel_width, tx_power, tx_power_max, noise_floor, channel_util, rx_util, tx_util, noise_util, clients
		},
		{
			"Traffic module only",
			APMetrics{Traffic: true},
			14, // rx/tx packets/bytes, data/mgmt/ctrl/multicast rx/tx frames, total_tx_frames, rts_success
		},
		{
			"Errors module only",
			APMetrics{Errors: true},
			16, // rx/tx errors, drops, retries, ack_failures, duplicates, fcs, frag rx/tx, rts_failures, decrypt, mic, wep, coverage_hole, radar, radio_reset
		},
		{
			"Info module only",
			APMetrics{Info: true},
			1, // info metric
		},
		{
			"All modules enabled",
			APMetrics{
				General: true,
				Radio:   true,
				Traffic: true,
				Errors:  true,
				Info:    true,
			},
			50, // 9+10+14+16+1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := NewAPCollector(nil, nil, nil, tt.metrics)

			ch := make(chan *prometheus.Desc, 100)
			collector.Describe(ch)
			close(ch)

			count := 0
			for range ch {
				count++
			}

			if count != tt.expectDescs {
				t.Errorf("Describe() sent %d descriptors, want %d", count, tt.expectDescs)
			}
		})
	}
}

func TestBuildCAPWAPMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []ap.CAPWAPData
		expected map[string]ap.CAPWAPData
	}{
		{
			"Empty slice",
			[]ap.CAPWAPData{},
			map[string]ap.CAPWAPData{},
		},
		{
			"Single AP",
			[]ap.CAPWAPData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", Name: "AP1"},
			},
			map[string]ap.CAPWAPData{
				"aa:bb:cc:dd:ee:ff": {WtpMAC: "aa:bb:cc:dd:ee:ff", Name: "AP1"},
			},
		},
		{
			"Multiple APs",
			[]ap.CAPWAPData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", Name: "AP1"},
				{WtpMAC: "11:22:33:44:55:66", Name: "AP2"},
			},
			map[string]ap.CAPWAPData{
				"aa:bb:cc:dd:ee:ff": {WtpMAC: "aa:bb:cc:dd:ee:ff", Name: "AP1"},
				"11:22:33:44:55:66": {WtpMAC: "11:22:33:44:55:66", Name: "AP2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildCAPWAPMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildCAPWAPMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for mac, expectedData := range tt.expected {
				if gotData, exists := got[mac]; !exists {
					t.Errorf("buildCAPWAPMap() missing MAC %q", mac)
				} else if gotData.WtpMAC != expectedData.WtpMAC || gotData.Name != expectedData.Name {
					t.Errorf("buildCAPWAPMap()[%q] = %+v, want %+v", mac, gotData, expectedData)
				}
			}
		})
	}
}

func TestBuildRadioDataMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []ap.RadioOperData
		expected map[string]*ap.RadioOperData
	}{
		{
			"Empty slice",
			[]ap.RadioOperData{},
			map[string]*ap.RadioOperData{},
		},
		{
			"Single radio",
			[]ap.RadioOperData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
			},
			map[string]*ap.RadioOperData{
				"aa:bb:cc:dd:ee:ff:0": {},
			},
		},
		{
			"Multiple radios same AP",
			[]ap.RadioOperData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 1},
			},
			map[string]*ap.RadioOperData{
				"aa:bb:cc:dd:ee:ff:0": {},
				"aa:bb:cc:dd:ee:ff:1": {},
			},
		},
		{
			"Multiple radios different APs",
			[]ap.RadioOperData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
				{WtpMAC: "11:22:33:44:55:66", RadioSlotID: 0},
			},
			map[string]*ap.RadioOperData{
				"aa:bb:cc:dd:ee:ff:0": {},
				"11:22:33:44:55:66:0": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildRadioDataMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildRadioDataMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for key := range tt.expected {
				if _, exists := got[key]; !exists {
					t.Errorf("buildRadioDataMap() missing key %q", key)
				}
			}
		})
	}
}

func TestBuildRadioOperStatsMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []ap.RadioOperStats
		expected map[string]map[int]ap.RadioOperStats
	}{
		{
			"Empty slice",
			[]ap.RadioOperStats{},
			map[string]map[int]ap.RadioOperStats{},
		},
		{
			"Single radio",
			[]ap.RadioOperStats{
				{ApMAC: "aa:bb:cc:dd:ee:ff", SlotID: 0, RxDataFrameCount: 100},
			},
			map[string]map[int]ap.RadioOperStats{
				"aa:bb:cc:dd:ee:ff": {
					0: {ApMAC: "aa:bb:cc:dd:ee:ff", SlotID: 0, RxDataFrameCount: 100},
				},
			},
		},
		{
			"Multiple radios same AP",
			[]ap.RadioOperStats{
				{ApMAC: "aa:bb:cc:dd:ee:ff", SlotID: 0, RxDataFrameCount: 100},
				{ApMAC: "aa:bb:cc:dd:ee:ff", SlotID: 1, RxDataFrameCount: 200},
			},
			map[string]map[int]ap.RadioOperStats{
				"aa:bb:cc:dd:ee:ff": {
					0: {ApMAC: "aa:bb:cc:dd:ee:ff", SlotID: 0, RxDataFrameCount: 100},
					1: {ApMAC: "aa:bb:cc:dd:ee:ff", SlotID: 1, RxDataFrameCount: 200},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildRadioOperStatsMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildRadioOperStatsMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for apMAC, expectedSlots := range tt.expected {
				if gotSlots, exists := got[apMAC]; !exists {
					t.Errorf("buildRadioOperStatsMap() missing AP MAC %q", apMAC)
				} else if len(gotSlots) != len(expectedSlots) {
					t.Errorf(
						"buildRadioOperStatsMap()[%q] has %d slots, want %d",
						apMAC,
						len(gotSlots),
						len(expectedSlots),
					)
				}
			}
		})
	}
}

func TestBuildAPOperDataMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []ap.OperData
		expected map[string]ap.OperData
	}{
		{
			"Empty slice",
			[]ap.OperData{},
			map[string]ap.OperData{},
		},
		{
			"Single AP",
			[]ap.OperData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff"},
			},
			map[string]ap.OperData{
				"aa:bb:cc:dd:ee:ff": {WtpMAC: "aa:bb:cc:dd:ee:ff"},
			},
		},
		{
			"Multiple APs",
			[]ap.OperData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff"},
				{WtpMAC: "11:22:33:44:55:66"},
			},
			map[string]ap.OperData{
				"aa:bb:cc:dd:ee:ff": {WtpMAC: "aa:bb:cc:dd:ee:ff"},
				"11:22:33:44:55:66": {WtpMAC: "11:22:33:44:55:66"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildAPOperDataMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildAPOperDataMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for mac := range tt.expected {
				if _, exists := got[mac]; !exists {
					t.Errorf("buildAPOperDataMap() missing MAC %q", mac)
				}
			}
		})
	}
}

func TestBuildRRMMeasurementsMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []rrm.RRMMeasurement
		expected map[string]*rrm.RRMMeasurement
	}{
		{
			"Empty slice",
			[]rrm.RRMMeasurement{},
			map[string]*rrm.RRMMeasurement{},
		},
		{
			"Single measurement",
			[]rrm.RRMMeasurement{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
			},
			map[string]*rrm.RRMMeasurement{
				"aa:bb:cc:dd:ee:ff:0": {},
			},
		},
		{
			"Multiple measurements",
			[]rrm.RRMMeasurement{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 1},
			},
			map[string]*rrm.RRMMeasurement{
				"aa:bb:cc:dd:ee:ff:0": {},
				"aa:bb:cc:dd:ee:ff:1": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildRRMMeasurementsMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf(
					"buildRRMMeasurementsMap() returned map with %d entries, want %d",
					len(got),
					len(tt.expected),
				)
			}

			for key := range tt.expected {
				if _, exists := got[key]; !exists {
					t.Errorf("buildRRMMeasurementsMap() missing key %q", key)
				}
			}
		})
	}
}

func TestBuildRadioResetStatsMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []ap.RadioResetStats
		expected map[string]map[int]*ap.RadioResetStats
	}{
		{
			"Empty slice",
			[]ap.RadioResetStats{},
			map[string]map[int]*ap.RadioResetStats{},
		},
		{
			"Single reset stat",
			[]ap.RadioResetStats{
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 0, Count: 5},
			},
			map[string]map[int]*ap.RadioResetStats{
				"aa:bb:cc:dd:ee:ff": {
					0: {},
				},
			},
		},
		{
			"Multiple reset stats",
			[]ap.RadioResetStats{
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 0, Count: 5},
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 1, Count: 3},
			},
			map[string]map[int]*ap.RadioResetStats{
				"aa:bb:cc:dd:ee:ff": {
					0: {},
					1: {},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildRadioResetStatsMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf(
					"buildRadioResetStatsMap() returned map with %d entries, want %d",
					len(got),
					len(tt.expected),
				)
			}

			for apMAC, expectedRadios := range tt.expected {
				if gotRadios, exists := got[apMAC]; !exists {
					t.Errorf("buildRadioResetStatsMap() missing AP MAC %q", apMAC)
				} else if len(gotRadios) != len(expectedRadios) {
					t.Errorf(
						"buildRadioResetStatsMap()[%q] has %d radios, want %d",
						apMAC,
						len(gotRadios),
						len(expectedRadios),
					)
				}
			}
		})
	}
}

func TestBuildRRMCoverageMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []rrm.RRMCoverage
		expected map[string]*rrm.RRMCoverage
	}{
		{
			"Empty slice",
			[]rrm.RRMCoverage{},
			map[string]*rrm.RRMCoverage{},
		},
		{
			"Single coverage",
			[]rrm.RRMCoverage{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
			},
			map[string]*rrm.RRMCoverage{
				"aa:bb:cc:dd:ee:ff:0": {},
			},
		},
		{
			"Multiple coverage entries",
			[]rrm.RRMCoverage{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 1},
			},
			map[string]*rrm.RRMCoverage{
				"aa:bb:cc:dd:ee:ff:0": {},
				"aa:bb:cc:dd:ee:ff:1": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildRRMCoverageMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildRRMCoverageMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for key := range tt.expected {
				if _, exists := got[key]; !exists {
					t.Errorf("buildRRMCoverageMap() missing key %q", key)
				}
			}
		})
	}
}

func TestBuildApDot11RadarMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []rrm.ApDot11RadarData
		expected map[string]*rrm.ApDot11RadarData
	}{
		{
			"Empty slice",
			[]rrm.ApDot11RadarData{},
			map[string]*rrm.ApDot11RadarData{},
		},
		{
			"Single radar entry",
			[]rrm.ApDot11RadarData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
			},
			map[string]*rrm.ApDot11RadarData{
				"aa:bb:cc:dd:ee:ff:0": {},
			},
		},
		{
			"Multiple radar entries",
			[]rrm.ApDot11RadarData{
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
				{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 1},
			},
			map[string]*rrm.ApDot11RadarData{
				"aa:bb:cc:dd:ee:ff:0": {},
				"aa:bb:cc:dd:ee:ff:1": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildApDot11RadarMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf(
					"buildApDot11RadarMap() returned map with %d entries, want %d",
					len(got),
					len(tt.expected),
				)
			}

			for key := range tt.expected {
				if _, exists := got[key]; !exists {
					t.Errorf("buildApDot11RadarMap() missing key %q", key)
				}
			}
		})
	}
}

func TestDetermineUptimeFromBootTime(t *testing.T) {
	t.Parallel()
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour).Format(time.RFC3339)
	oneDayAgo := now.Add(-24 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name        string
		bootTimeStr string
		minExpected int64
		maxExpected int64
		expectZero  bool
		expectError bool
	}{
		{
			"Empty string",
			"",
			0,
			0,
			true,
			false,
		},
		{
			"Invalid format",
			"2024-01-01",
			0,
			0,
			true,
			false,
		},
		{
			"One hour ago",
			oneHourAgo,
			3599,
			3601,
			false,
			false,
		},
		{
			"One day ago",
			oneDayAgo,
			86399,
			86401,
			false,
			false,
		},
		{
			"Malformed timestamp",
			"not-a-timestamp",
			0,
			0,
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := determineUptimeFromBootTime(tt.bootTimeStr)

			if tt.expectZero {
				if got != 0 {
					t.Errorf("determineUptimeFromBootTime(%q) = %d, want 0", tt.bootTimeStr, got)
				}
			} else {
				if got < tt.minExpected || got > tt.maxExpected {
					t.Errorf(
						"determineUptimeFromBootTime(%q) = %d, want between %d and %d",
						tt.bootTimeStr,
						got,
						tt.minExpected,
						tt.maxExpected,
					)
				}
			}
		})
	}
}

func TestAPCollector_MetricNames(t *testing.T) {
	t.Parallel()

	collector := NewAPCollector(nil, nil, nil, APMetrics{
		General: true,
		Radio:   true,
		Traffic: true,
		Errors:  true,
		Info:    true,
	})

	tests := []struct {
		desc     *prometheus.Desc
		expected string
	}{
		{collector.radioStateDesc, "wnc_ap_radio_state"},
		{collector.adminStateDesc, "wnc_ap_admin_state"},
		{collector.operStateDesc, "wnc_ap_oper_state"},
		{collector.configStateDesc, "wnc_ap_config_state"},
		{collector.uptimeSecondsDesc, "wnc_ap_uptime_seconds"},
		{collector.channelDesc, "wnc_ap_channel_number"},
		{collector.txPowerDesc, "wnc_ap_tx_power_dbm"},
		{collector.noiseFloorDesc, "wnc_ap_noise_floor_dbm"},
		{collector.rxPacketsTotalDesc, "wnc_ap_rx_packets_total"},
		{collector.txPacketsTotalDesc, "wnc_ap_tx_packets_total"},
		{collector.rxErrorsTotalDesc, "wnc_ap_rx_errors_total"},
		{collector.txErrorsTotalDesc, "wnc_ap_tx_errors_total"},
		{collector.infoDesc, "wnc_ap_info"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			if tt.desc == nil {
				t.Errorf("Descriptor for %s is nil", tt.expected)
				return
			}

			descStr := tt.desc.String()
			if !strings.Contains(descStr, tt.expected) {
				t.Errorf("Descriptor %s does not contain expected metric name %s", descStr, tt.expected)
			}

			if !strings.Contains(descStr, "wnc_ap_") {
				t.Errorf("Descriptor %s does not follow wnc_ap_ naming convention", descStr)
			}
		})
	}
}

func TestNewAPCollector_InfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		infoLabels []string
		wantLabels []string
	}{
		{
			"No info labels",
			nil,
			[]string{"mac", "radio"},
		},
		{
			"Custom info labels",
			[]string{"name", "ip", "band"},
			[]string{"mac", "radio", "name", "ip", "band"},
		},
		{
			"Info labels with required labels included",
			[]string{"mac", "radio", "name"},
			[]string{"mac", "radio", "name"},
		},
		{
			"Invalid labels filtered",
			[]string{"name", "invalid_label", "model"},
			[]string{"mac", "radio", "name", "model"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			metrics := APMetrics{
				Info:       true,
				InfoLabels: tt.infoLabels,
			}
			collector := NewAPCollector(nil, nil, nil, metrics)

			if collector.infoDesc == nil {
				t.Error("NewAPCollector() info descriptor should not be nil when Info is enabled")
				return
			}

			if len(collector.infoLabelNames) != len(tt.wantLabels) {
				t.Errorf(
					"NewAPCollector() infoLabelNames length = %d, want %d",
					len(collector.infoLabelNames),
					len(tt.wantLabels),
				)
				return
			}

			for _, expectedLabel := range tt.wantLabels {
				found := false
				for _, actualLabel := range collector.infoLabelNames {
					if actualLabel == expectedLabel {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(
						"NewAPCollector() missing expected label %q in infoLabelNames %v",
						expectedLabel,
						collector.infoLabelNames,
					)
				}
			}
		})
	}
}

func TestAPCollector_Integration(t *testing.T) {
	t.Parallel()

	metrics := APMetrics{
		General:    true,
		Radio:      true,
		Traffic:    true,
		Errors:     true,
		Info:       true,
		InfoLabels: []string{"name", "ip", "band", "model", "serial", "sw_version", "eth_mac"},
	}

	collector := NewAPCollector(nil, nil, nil, metrics)
	if collector == nil {
		t.Fatal("NewAPCollector() returned nil")
	}

	var _ prometheus.Collector = collector

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		defer close(ch)
		collector.Describe(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	if count == 0 {
		t.Error("Collector did not emit any descriptors")
	}

	expectedDescs := 50
	if count != expectedDescs {
		t.Errorf("Collector emitted %d descriptors, want %d", count, expectedDescs)
	}
}

func TestFloat64Metric_Type(t *testing.T) {
	t.Parallel()
	desc := prometheus.NewDesc("test_metric", "Test metric", nil, nil)
	metric := Float64Metric{
		Desc:  desc,
		Value: 42.5,
	}

	if metric.Desc == nil {
		t.Error("Float64Metric.Desc should not be nil")
	}
	if metric.Value != 42.5 {
		t.Errorf("Float64Metric.Value = %f, want 42.5", metric.Value)
	}
}

func TestStringMetric_Type(t *testing.T) {
	t.Parallel()
	desc := prometheus.NewDesc("test_metric", "Test metric", nil, nil)
	metric := StringMetric{
		Desc:  desc,
		Value: "test_value",
	}

	if metric.Desc == nil {
		t.Error("StringMetric.Desc should not be nil")
	}
	if metric.Value != "test_value" {
		t.Errorf("StringMetric.Value = %q, want \"test_value\"", metric.Value)
	}
}

func TestBuildRadioDataMap_KeyGeneration(t *testing.T) {
	t.Parallel()

	input := []ap.RadioOperData{
		{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0},
		{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 15},
	}

	got := buildRadioDataMap(input)

	expectedKey0 := "aa:bb:cc:dd:ee:ff:" + strconv.Itoa(0)
	expectedKey15 := "aa:bb:cc:dd:ee:ff:" + strconv.Itoa(15)

	if _, exists := got[expectedKey0]; !exists {
		t.Errorf("buildRadioDataMap() missing key %q", expectedKey0)
	}
	if _, exists := got[expectedKey15]; !exists {
		t.Errorf("buildRadioDataMap() missing key %q", expectedKey15)
	}
}

func TestAPCollector_Collect_EarlyReturn(t *testing.T) {
	t.Parallel()

	collector := &APCollector{
		metrics: APMetrics{},
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.Collect(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("Collect() emitted %d metrics when all flags disabled, want 0", count)
	}
}

func TestAPCollector_LabelConstants(t *testing.T) {
	t.Parallel()

	if labelMAC != "mac" {
		t.Errorf("labelMAC = %q, want \"mac\"", labelMAC)
	}
	if labelName != "name" {
		t.Errorf("labelName = %q, want \"name\"", labelName)
	}
	if labelBand != "band" {
		t.Errorf("labelBand = %q, want \"band\"", labelBand)
	}
}

func TestAPCollector_collectInfoMetrics_LabelValues(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "dot11-24ghz-radio",
	}

	capwapData := ap.CAPWAPData{
		WtpMAC: "aa:bb:cc:dd:ee:ff",
		Name:   "AP-01",
		IPAddr: "192.168.1.10",
	}
	capwapData.DeviceDetail.StaticInfo.ApModels.Model = "C9115AXI-B"
	capwapData.DeviceDetail.StaticInfo.BoardData.WtpSerialNum = "FOC12345678"
	capwapData.DeviceDetail.StaticInfo.BoardData.WtpEnetMAC = "aa:bb:cc:dd:ee:00"
	capwapData.DeviceDetail.WtpVersion.SwVersion = "17.12.1"

	capwapMap := map[string]ap.CAPWAPData{
		"aa:bb:cc:dd:ee:ff": capwapData,
	}

	metrics := APMetrics{
		Info:       true,
		InfoLabels: []string{"mac", "name", "ip", "radio", "band", "model", "serial", "sw_version", "eth_mac"},
	}

	collector := NewAPCollector(nil, nil, nil, metrics)

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectInfoMetrics(ch, radio, capwapMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount != 1 {
		t.Errorf("collectInfoMetrics() emitted %d metrics, want 1", metricCount)
	}
}

func TestAPCollector_collectGeneralMetrics(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "dot11-5ghz-radio",
		OperState:   "radio-up",
		AdminState:  "admin-enabled",
	}

	collector := &APCollector{
		metrics:        APMetrics{General: true},
		radioStateDesc: prometheus.NewDesc("test_radio_state", "test", []string{"mac", "radio"}, nil),
		adminStateDesc: prometheus.NewDesc("test_admin_state", "test", []string{"mac", "radio"}, nil),
		operStateDesc:  prometheus.NewDesc("test_oper_state", "test", []string{"mac", "radio"}, nil),
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectGeneralMetrics(ch, radio)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectGeneralMetrics() emitted 0 metrics, want > 0")
	}
}

func TestAPCollector_collectRadioMetrics(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "dot11-5ghz-radio",
	}

	rrmMap := map[string]*rrm.RRMMeasurement{}

	collector := &APCollector{
		metrics:                APMetrics{Radio: true},
		associatedClientsDesc:  prometheus.NewDesc("test_clients", "test", []string{"mac", "radio"}, nil),
		txPowerDesc:            prometheus.NewDesc("test_tx_power", "test", []string{"mac", "radio"}, nil),
		txPowerMaxDesc:         prometheus.NewDesc("test_tx_power_max", "test", []string{"mac", "radio"}, nil),
		channelDesc:            prometheus.NewDesc("test_channel", "test", []string{"mac", "radio"}, nil),
		channelWidthDesc:       prometheus.NewDesc("test_channel_width", "test", []string{"mac", "radio"}, nil),
		channelUtilizationDesc: prometheus.NewDesc("test_channel_util", "test", []string{"mac", "radio"}, nil),
		rxUtilizationDesc:      prometheus.NewDesc("test_rx_util", "test", []string{"mac", "radio"}, nil),
		txUtilizationDesc:      prometheus.NewDesc("test_tx_util", "test", []string{"mac", "radio"}, nil),
		noiseUtilizationDesc:   prometheus.NewDesc("test_noise_util", "test", []string{"mac", "radio"}, nil),
		noiseFloorDesc:         prometheus.NewDesc("test_noise_floor", "test", []string{"mac", "radio"}, nil),
	}

	ch := make(chan prometheus.Metric, 20)
	go func() {
		defer close(ch)
		collector.collectRadioMetrics(ch, radio, rrmMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectRadioMetrics() emitted 0 metrics, want > 0")
	}
}

func TestAPCollector_collectTrafficMetrics(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "dot11-5ghz-radio",
	}

	statsMap := map[string]map[int]ap.RadioOperStats{
		"aa:bb:cc:dd:ee:ff": {
			0: {
				RxDataFrameCount: 1000,
				TxDataFrameCount: 2000,
				RxMgmtFrameCount: 100,
				TxMgmtFrameCount: 200,
			},
		},
	}

	collector := &APCollector{
		metrics:                     APMetrics{Traffic: true},
		rxPacketsTotalDesc:          prometheus.NewDesc("test_rx_packets", "test", []string{"mac", "radio"}, nil),
		txPacketsTotalDesc:          prometheus.NewDesc("test_tx_packets", "test", []string{"mac", "radio"}, nil),
		rxBytesTotalDesc:            prometheus.NewDesc("test_rx_bytes", "test", []string{"mac", "radio"}, nil),
		txBytesTotalDesc:            prometheus.NewDesc("test_tx_bytes", "test", []string{"mac", "radio"}, nil),
		dataRxFramesTotalDesc:       prometheus.NewDesc("test_data_rx", "test", []string{"mac", "radio"}, nil),
		dataTxFramesTotalDesc:       prometheus.NewDesc("test_data_tx", "test", []string{"mac", "radio"}, nil),
		managementRxFramesTotalDesc: prometheus.NewDesc("test_mgmt_rx", "test", []string{"mac", "radio"}, nil),
		managementTxFramesTotalDesc: prometheus.NewDesc("test_mgmt_tx", "test", []string{"mac", "radio"}, nil),
		controlRxFramesTotalDesc:    prometheus.NewDesc("test_ctrl_rx", "test", []string{"mac", "radio"}, nil),
		controlTxFramesTotalDesc:    prometheus.NewDesc("test_ctrl_tx", "test", []string{"mac", "radio"}, nil),
		multicastRxFramesTotalDesc:  prometheus.NewDesc("test_mcast_rx", "test", []string{"mac", "radio"}, nil),
		multicastTxFramesTotalDesc:  prometheus.NewDesc("test_mcast_tx", "test", []string{"mac", "radio"}, nil),
		totalTxFramesTotalDesc:      prometheus.NewDesc("test_total_tx", "test", []string{"mac", "radio"}, nil),
		rtsSuccessTotalDesc:         prometheus.NewDesc("test_rts_success", "test", []string{"mac", "radio"}, nil),
	}

	ch := make(chan prometheus.Metric, 20)
	go func() {
		defer close(ch)
		collector.collectTrafficMetrics(ch, radio, statsMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectTrafficMetrics() emitted 0 metrics, want > 0")
	}
}

func TestAPCollector_collectErrorMetrics(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "dot11-5ghz-radio",
	}

	statsMap := map[string]map[int]ap.RadioOperStats{
		"aa:bb:cc:dd:ee:ff": {
			0: {
				RxErrorFrameCount:     10,
				FailedCount:           5,
				AckFailureCount:       3,
				RetryCount:            20,
				FrameDuplicateCount:   2,
				FcsErrorCount:         1,
				RxFragmentCount:       0,
				TxFragmentCount:       0,
				RtsFailureCount:       0,
				MACDecryErrFrameCount: 0,
				MACMicErrFrameCount:   0,
				WepUndecryptableCount: 0,
			},
		},
	}
	resetStatsMap := map[string]map[int]*ap.RadioResetStats{}
	rrmCoverageMap := map[string]*rrm.RRMCoverage{}
	apDot11RadarMap := map[string]*rrm.ApDot11RadarData{}

	collector := &APCollector{
		metrics:                   APMetrics{Errors: true},
		rxErrorsTotalDesc:         prometheus.NewDesc("test_rx_errors", "test", []string{"mac", "radio"}, nil),
		txErrorsTotalDesc:         prometheus.NewDesc("test_tx_errors", "test", []string{"mac", "radio"}, nil),
		txDropsTotalDesc:          prometheus.NewDesc("test_tx_drops", "test", []string{"mac", "radio"}, nil),
		txRetriesTotalDesc:        prometheus.NewDesc("test_tx_retries", "test", []string{"mac", "radio"}, nil),
		ackFailuresTotalDesc:      prometheus.NewDesc("test_ack_failures", "test", []string{"mac", "radio"}, nil),
		duplicateFramesTotalDesc:  prometheus.NewDesc("test_duplicates", "test", []string{"mac", "radio"}, nil),
		fcsErrorsTotalDesc:        prometheus.NewDesc("test_fcs_errors", "test", []string{"mac", "radio"}, nil),
		fragmentationRxTotalDesc:  prometheus.NewDesc("test_frag_rx", "test", []string{"mac", "radio"}, nil),
		fragmentationTxTotalDesc:  prometheus.NewDesc("test_frag_tx", "test", []string{"mac", "radio"}, nil),
		rtsFailuresTotalDesc:      prometheus.NewDesc("test_rts_failures", "test", []string{"mac", "radio"}, nil),
		decryptionErrorsTotalDesc: prometheus.NewDesc("test_decrypt_errors", "test", []string{"mac", "radio"}, nil),
		micErrorsTotalDesc:        prometheus.NewDesc("test_mic_errors", "test", []string{"mac", "radio"}, nil),
		wepUndecryptableTotalDesc: prometheus.NewDesc("test_wep_undecrypt", "test", []string{"mac", "radio"}, nil),
		coverageHoleEventsDesc:    prometheus.NewDesc("test_coverage_holes", "test", []string{"mac", "radio"}, nil),
		lastRadarOnRadioAtDesc:    prometheus.NewDesc("test_last_radar", "test", []string{"mac", "radio"}, nil),
		radioResetTotalDesc:       prometheus.NewDesc("test_radio_resets", "test", []string{"mac", "radio"}, nil),
	}

	ch := make(chan prometheus.Metric, 30)
	go func() {
		defer close(ch)
		collector.collectErrorMetrics(ch, radio, statsMap, resetStatsMap, rrmCoverageMap, apDot11RadarMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectErrorMetrics() emitted 0 metrics, want > 0")
	}
}

func TestAPCollector_collectMetrics_NilSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		testFunc   func(*testing.T)
		wantsPanic bool
	}{
		{
			name: "collectGeneralMetrics with nil radio",
			testFunc: func(t *testing.T) {
				t.Helper()
				collector := &APCollector{
					metrics:        APMetrics{General: true},
					radioStateDesc: prometheus.NewDesc("test", "test", []string{"mac", "radio"}, nil),
					adminStateDesc: prometheus.NewDesc("test", "test", []string{"mac", "radio"}, nil),
					operStateDesc:  prometheus.NewDesc("test", "test", []string{"mac", "radio"}, nil),
				}
				ch := make(chan prometheus.Metric, 10)
				panicked := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicked = true
						}
						close(ch)
						for range ch {
						}
					}()
					collector.collectGeneralMetrics(ch, nil)
				}()
				if panicked {
					t.Log("collectGeneralMetrics() panicked with nil radio (expected)")
				}
			},
			wantsPanic: true,
		},
		{
			name: "collectRadioMetrics with nil radio",
			testFunc: func(t *testing.T) {
				t.Helper()
				collector := &APCollector{
					metrics:               APMetrics{Radio: true},
					associatedClientsDesc: prometheus.NewDesc("test", "test", []string{"mac", "radio"}, nil),
				}
				ch := make(chan prometheus.Metric, 10)
				panicked := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicked = true
						}
						close(ch)
						for range ch {
						}
					}()
					collector.collectRadioMetrics(ch, nil, nil)
				}()
				if panicked {
					t.Log("collectRadioMetrics() panicked with nil radio (expected)")
				}
			},
			wantsPanic: true,
		},
		{
			name: "collectTrafficMetrics with nil radio",
			testFunc: func(t *testing.T) {
				t.Helper()
				collector := &APCollector{
					metrics: APMetrics{Traffic: true},
				}
				ch := make(chan prometheus.Metric, 10)
				panicked := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicked = true
						}
						close(ch)
						for range ch {
						}
					}()
					collector.collectTrafficMetrics(ch, nil, nil)
				}()
				if panicked {
					t.Log("collectTrafficMetrics() panicked with nil radio (expected)")
				}
			},
			wantsPanic: true,
		},
		{
			name: "collectErrorMetrics with nil radio",
			testFunc: func(t *testing.T) {
				t.Helper()
				collector := &APCollector{
					metrics: APMetrics{Errors: true},
				}
				ch := make(chan prometheus.Metric, 10)
				panicked := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicked = true
						}
						close(ch)
						for range ch {
						}
					}()
					collector.collectErrorMetrics(ch, nil, nil, nil, nil, nil)
				}()
				if panicked {
					t.Log("collectErrorMetrics() panicked with nil radio (expected)")
				}
			},
			wantsPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}
