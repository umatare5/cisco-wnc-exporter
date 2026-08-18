package collector

import (
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// TestAPCollector_StateReportsSpellingInLabel pins the AP-level encoding. The series
// previously duplicated wnc_ap_radio_state, so it must now carry the CAPWAP state in a
// label and must not carry a radio label.
func TestAPCollector_StateReportsSpellingInLabel(t *testing.T) {
	t.Parallel()

	data := fullFixtureSnapshot()
	data.CAPWAPData = append(data.CAPWAPData,
		ap.CAPWAPData{WtpMAC: "22:33:44:55:66:77", ApState: ap.ApState{ApOperationState: "downloading"}},
		ap.CAPWAPData{WtpMAC: "33:44:55:66:77:88"},
	)
	src := fixtureSource{data: data}

	registry := prometheus.NewRegistry()
	registry.MustRegister(NewAPCollector(
		wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src),
		APMetrics{General: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	states := make(map[string]float64)
	series := 0
	for _, family := range families {
		if family.GetName() != "wnc_ap_oper_state" {
			continue
		}
		for _, metric := range family.GetMetric() {
			series++
			for _, label := range metric.GetLabel() {
				if label.GetName() == labelRadio {
					t.Error("wnc_ap_oper_state carries a radio label, want one series per AP")
				}
				if label.GetName() == labelState {
					states[label.GetValue()] = metric.GetGauge().GetValue()
				}
			}
		}
	}

	// Two of the three fixture APs report a state. A third series would mean the
	// radio-level duplicate is back, since a per-radio emit reuses this descriptor
	// and passes a slot number where the state belongs.
	const wantSeries = 2
	if series != wantSeries {
		t.Errorf("wnc_ap_oper_state has %d series, want %d, one per AP reporting a state", series, wantSeries)
	}
	if got := states["downloading"]; got != 1 {
		t.Errorf("wnc_ap_oper_state{state=downloading} = %v, want 1 with the state in the label", got)
	}
	if _, ok := states[""]; ok {
		t.Error("wnc_ap_oper_state carries an empty state label, want that series omitted")
	}
}

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
			// radio_state, admin_state, oper_state, config_state, uptime,
			// association_uptime, cpu_utilization, memory_utilization
			8,
		},
		{
			"Radio module only",
			APMetrics{Radio: true},
			// channel, channel_width, tx_power, tx_power_max, noise_floor, channel_util,
			// rx_util, tx_util, noise_util, clients, rrm_profile_passed, channel_changes,
			// channel_energy
			13,
		},
		{
			"Traffic module only",
			APMetrics{Traffic: true},
			10, // data/mgmt/ctrl/multicast rx/tx frames, total_tx_frames, rts_successes
		},
		{
			"Errors module only",
			APMetrics{Errors: true},
			13, // rx errors, retries, transmission_failures, duplicates, fcs, frag rx/tx, rts_failures, decrypt, mic, coverage_hole, radar, radio_resets
		},
		{
			"Join module only",
			APMetrics{Join: true},
			// joined, join_info, 14 counters, 9 timestamps, 7 reasons
			32,
		},
		{
			"Spectrum module only",
			APMetrics{Spectrum: true},
			// The three per-radio air quality series and the four band-keyed ones
			7,
		},
		{
			"Info module only",
			APMetrics{Info: true},
			1, // info metric
		},
		{
			"All modules enabled",
			APMetrics{
				General:  true,
				Radio:    true,
				Traffic:  true,
				Errors:   true,
				Join:     true,
				Spectrum: true,
				Info:     true,
			},
			84, // 8+13+10+13+32+7+1
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
		expected map[string]map[int]int
	}{
		{
			"Empty slice",
			[]ap.RadioResetStats{},
			map[string]map[int]int{},
		},
		{
			"Single reset stat",
			[]ap.RadioResetStats{
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 0, Count: 5},
			},
			map[string]map[int]int{
				"aa:bb:cc:dd:ee:ff": {0: 5},
			},
		},
		{
			"One entry per radio",
			[]ap.RadioResetStats{
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 0, Count: 5},
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 1, Count: 3},
			},
			map[string]map[int]int{
				"aa:bb:cc:dd:ee:ff": {0: 5, 1: 3},
			},
		},
		{
			// The YANG list is keyed by ap-mac, radio-id, cause and detail-cause, so
			// one radio has one entry per cause. Keeping only the last one made the
			// counter drop whenever a cause appeared or disappeared.
			"Every cause on a radio is summed",
			[]ap.RadioResetStats{
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 0, Cause: "reset-cause-1", Count: 5},
				{ApMAC: "aa:bb:cc:dd:ee:ff", RadioID: 0, Cause: "reset-cause-2", Count: 3},
			},
			map[string]map[int]int{
				"aa:bb:cc:dd:ee:ff": {0: 8},
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
				gotRadios, exists := got[apMAC]
				if !exists {
					t.Errorf("buildRadioResetStatsMap() missing AP MAC %q", apMAC)
					continue
				}
				if len(gotRadios) != len(expectedRadios) {
					t.Errorf(
						"buildRadioResetStatsMap()[%q] has %d radios, want %d",
						apMAC,
						len(gotRadios),
						len(expectedRadios),
					)
				}
				for radioID, wantCount := range expectedRadios {
					if gotRadios[radioID] != wantCount {
						t.Errorf(
							"buildRadioResetStatsMap()[%q][%d] = %d, want %d",
							apMAC,
							radioID,
							gotRadios[radioID],
							wantCount,
						)
					}
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

func TestBuildRadioClientCountsMap(t *testing.T) {
	t.Parallel()

	nameMACMaps := []ap.ApNameMACMap{
		{WtpName: "AP01", WtpMAC: "aa:bb:cc:dd:ee:ff"},
		{WtpName: "AP02", WtpMAC: "11:22:33:44:55:66"},
	}

	tests := []struct {
		name       string
		clientData []client.CommonOperData
		expected   map[string]map[int]int
	}{
		{
			name: "Counts clients per AP and radio slot",
			clientData: []client.CommonOperData{
				{ApName: "AP01", MsApSlotID: 0, CoState: ClientStatusRun},
				{ApName: "AP01", MsApSlotID: 0, CoState: ClientStatusRun},
				{ApName: "AP01", MsApSlotID: 1, CoState: ClientStatusRun},
				{ApName: "AP02", MsApSlotID: 1, CoState: ClientStatusRun},
			},
			expected: map[string]map[int]int{
				"aa:bb:cc:dd:ee:ff": {0: 2, 1: 1},
				"11:22:33:44:55:66": {1: 1},
			},
		},
		{
			name: "Skips clients outside run state",
			clientData: []client.CommonOperData{
				{ApName: "AP01", MsApSlotID: 0, CoState: ClientStatusRun},
				{ApName: "AP01", MsApSlotID: 0, CoState: "client-status-associating"},
			},
			expected: map[string]map[int]int{
				"aa:bb:cc:dd:ee:ff": {0: 1},
			},
		},
		{
			name: "Skips clients on unresolvable AP names",
			clientData: []client.CommonOperData{
				{ApName: "AP99", MsApSlotID: 0, CoState: ClientStatusRun},
			},
			expected: map[string]map[int]int{},
		},
		{
			name:       "Returns empty map without client data",
			clientData: nil,
			expected:   map[string]map[int]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildRadioClientCountsMap(tt.clientData, nameMACMaps)

			if len(got) != len(tt.expected) {
				t.Fatalf("buildRadioClientCountsMap() returned %d APs, want %d", len(got), len(tt.expected))
			}

			for wtpMAC, slots := range tt.expected {
				if len(got[wtpMAC]) != len(slots) {
					t.Errorf("AP %s has %d slots, want %d", wtpMAC, len(got[wtpMAC]), len(slots))
				}
				for slotID, want := range slots {
					if got[wtpMAC][slotID] != want {
						t.Errorf("AP %s slot %d = %d, want %d", wtpMAC, slotID, got[wtpMAC][slotID], want)
					}
				}
			}
		})
	}
}

func TestDetermineUptimeFromTimestamp(t *testing.T) {
	t.Parallel()
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour).Format(time.RFC3339)
	oneDayAgo := now.Add(-24 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name        string
		bootTimeStr string
		minExpected int64
		maxExpected int64
		expectNotOK bool
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
		{
			// A placeholder rather than a boot: no AP booted in 1970. Whether this
			// controller renders one for this leaf is not established — the guard is
			// defensive, and it matches what the join timestamps do with the same value.
			"Unix epoch",
			"1970-01-01T00:00:00+00:00",
			0,
			0,
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := determineUptimeFromTimestamp(tt.bootTimeStr)

			if tt.expectNotOK {
				if ok {
					t.Errorf("determineUptimeFromTimestamp(%q) reported %d as usable, want it unusable",
						tt.bootTimeStr, got)
				}
			} else {
				if !ok {
					t.Errorf("determineUptimeFromTimestamp(%q) reported no usable uptime", tt.bootTimeStr)
				}
				if got < tt.minExpected || got > tt.maxExpected {
					t.Errorf(
						"determineUptimeFromTimestamp(%q) = %d, want between %d and %d",
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

// TestAPCollector_UptimeWithheldWhenBootTimeUnusable pins the emission side of the
// same contract. The helper test above proves only that the helper reports the leaf
// as unusable; nothing there stops the collector publishing the zero anyway.
func TestAPCollector_UptimeWithheldWhenBootTimeUnusable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		bootTime    string
		wantPresent bool
	}{
		{"absent leaf", "", false},
		{"unparsable leaf", "2026-01-01", false},
		{"epoch placeholder", "1970-01-01T00:00:00+00:00", false},
		{"usable leaf", "2026-01-01T00:00:00Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.CAPWAPData[0].ApTimeInfo.BootTime = tt.bootTime

			values := apSnapshotValues(t, data)
			if _, ok := values["wnc_ap_uptime_seconds"]; ok != tt.wantPresent {
				t.Errorf("wnc_ap_uptime_seconds present = %v for boot time %q, want %v",
					ok, tt.bootTime, tt.wantPresent)
			}

			// The withhold has to be scoped to this one series: it shares an emit loop
			// with the AP-level config state, which reads a leaf of its own.
			if _, ok := values["wnc_ap_config_state"]; !ok {
				t.Error("wnc_ap_config_state is absent, so the assertion above proves nothing")
			}
		})
	}
}

// TestAPCollector_AssociationUptimeReadsTheJoinTime pins which of the two instants of
// the time container this series reads. The AP boots before it joins, so a series reading
// the boot time reports an association older than the one the AP holds.
func TestAPCollector_AssociationUptimeReadsTheJoinTime(t *testing.T) {
	t.Parallel()

	values := apSnapshotValues(t, fullFixtureSnapshot())

	association, ok := values["wnc_ap_association_uptime_seconds"]
	if !ok {
		t.Fatal("wnc_ap_association_uptime_seconds is absent, so nothing below can be asserted")
	}

	boot, ok := values["wnc_ap_uptime_seconds"]
	if !ok {
		t.Fatal("wnc_ap_uptime_seconds is absent, so the comparison below proves nothing")
	}

	// Both are elapsed seconds against the same clock, so only their difference is
	// stable enough to assert. The fixture puts the two instants one day apart.
	const oneDay = 24 * 60 * 60
	if got := boot - association; got != oneDay {
		t.Errorf("wnc_ap_uptime_seconds - wnc_ap_association_uptime_seconds = %f, want %d",
			got, oneDay)
	}
}

// TestAPCollector_AssociationUptimeWithheldWhenJoinTimeUnusable pins the emission side of
// the withhold, which the helper test cannot reach.
func TestAPCollector_AssociationUptimeWithheldWhenJoinTimeUnusable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		joinTime    string
		wantPresent bool
	}{
		{"absent leaf", "", false},
		{"unparsable leaf", "2026-01-02", false},
		{"epoch placeholder", "1970-01-01T00:00:00+00:00", false},
		{"usable leaf", fixtureAPJoinTime, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.CAPWAPData[0].ApTimeInfo.JoinTime = tt.joinTime

			values := apSnapshotValues(t, data)
			if _, ok := values["wnc_ap_association_uptime_seconds"]; ok != tt.wantPresent {
				t.Errorf("wnc_ap_association_uptime_seconds present = %v for join time %q, want %v",
					ok, tt.joinTime, tt.wantPresent)
			}

			// The withhold is scoped to this series: the boot-time one reads its own leaf.
			if _, ok := values["wnc_ap_uptime_seconds"]; !ok {
				t.Error("wnc_ap_uptime_seconds is absent, so the assertion above proves nothing")
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
		{collector.rxErrorsTotalDesc, "wnc_ap_rx_errors_total"},
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
		Spectrum:   true,
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

	expectedDescs := 52
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
		// APRadioBand reads this leaf and nothing else, so leaving it unset makes the
		// band label read "unknown" and any assertion on it a tautology.
		CurrentActiveBand: "dot11-2-dot-4-ghz-band",
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

func TestAPCollector_collectSystemMetrics(t *testing.T) {
	t.Parallel()

	const wtpMAC = "aa:bb:cc:dd:ee:ff"

	collector := &APCollector{
		metrics:               APMetrics{General: true},
		configStateDesc:       prometheus.NewDesc("test_config_state", "test", []string{"mac"}, nil),
		uptimeSecondsDesc:     prometheus.NewDesc("test_uptime", "test", []string{"mac"}, nil),
		cpuUtilizationDesc:    prometheus.NewDesc("test_cpu", "test", []string{"mac"}, nil),
		memoryUtilizationDesc: prometheus.NewDesc("test_memory", "test", []string{"mac"}, nil),
		operStateDesc:         prometheus.NewDesc("test_oper_state", "test", []string{"mac", "state"}, nil),
	}

	capwapMap := map[string]ap.CAPWAPData{
		wtpMAC: {
			WtpMAC:     wtpMAC,
			ApState:    ap.ApState{ApOperationState: "registered"},
			ApTimeInfo: ap.ApTimeInfo{BootTime: "2026-01-01T00:00:00Z"},
		},
	}
	sysStats := &ap.ApSystemStats{CPUUsage: 20, MemoryUsage: 40}

	tests := []struct {
		name          string
		apOperDataMap map[string]ap.OperData
		// Oper state, config state and uptime are always emitted, CPU and memory only with ApSysStats.
		expected int
	}{
		{
			"ApSysStats present",
			map[string]ap.OperData{wtpMAC: {WtpMAC: wtpMAC, ApSysStats: sysStats}},
			5,
		},
		{
			"ApSysStats absent",
			map[string]ap.OperData{wtpMAC: {WtpMAC: wtpMAC}},
			3,
		},
		{
			"AP missing from oper data map",
			map[string]ap.OperData{},
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ch := make(chan prometheus.Metric, 10)
			go func() {
				defer close(ch)
				collector.collectSystemMetrics(ch, wtpMAC, capwapMap, tt.apOperDataMap)
			}()

			metricCount := 0
			for range ch {
				metricCount++
			}

			if metricCount != tt.expected {
				t.Errorf("collectSystemMetrics() emitted %d metrics, want %d", metricCount, tt.expected)
			}
		})
	}
}

func TestAPCollector_collectGeneralMetrics(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "dot11-5ghz-radio",
		OperState:   "radio-up",
		// A controller returns "disabled" here, never "admin-enabled". The two leaves
		// are kept deliberately unequal so that swapping the comparisons in
		// collectGeneralMetrics changes what this fixture publishes.
		AdminState: "disabled",
	}

	collector := &APCollector{
		metrics:        APMetrics{General: true},
		radioStateDesc: prometheus.NewDesc("test_radio_state", "test", []string{"mac", "radio"}, nil),
		adminStateDesc: prometheus.NewDesc("test_admin_state", "test", []string{"mac", "radio"}, nil),
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
		OperState:   APRadioStateUp,
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

	tests := []struct {
		name            string
		clientCountsMap map[string]map[int]int
		want            int
		reason          string
	}{
		{
			name:            "Client counts available",
			clientCountsMap: map[string]map[int]int{},
			want:            1,
			reason:          "an empty map is a radio with nobody on it, which is a real zero",
		},
		{
			name:            "Client counts unavailable",
			clientCountsMap: nil,
			want:            0,
			reason:          "a zero client count would read as an idle radio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ch := make(chan prometheus.Metric, 20)
			go func() {
				defer close(ch)
				collector.collectRadioMetrics(ch, radio, rrmMap, tt.clientCountsMap, nil)
			}()

			metricCount := 0
			for range ch {
				metricCount++
			}

			if metricCount != tt.want {
				t.Errorf("collectRadioMetrics() emitted %d metrics, want %d: %s",
					metricCount, tt.want, tt.reason)
			}
		})
	}
}

func TestAPCollector_collectRadioMetrics_NilRRMSubContainers(t *testing.T) {
	t.Parallel()

	const operatingChannel = 36

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "radio-80211a",
		OperState:   APRadioStateUp,
		PhyHtCfg:    &ap.PhyHtCfg{CfgData: ap.PhyHtCfgData{CurrFreq: operatingChannel}},
	}

	collector := &APCollector{
		metrics:                APMetrics{Radio: true},
		associatedClientsDesc:  prometheus.NewDesc("test_clients", "test", []string{"mac", "radio"}, nil),
		channelDesc:            prometheus.NewDesc("test_channel", "test", []string{"mac", "radio"}, nil),
		channelWidthDesc:       prometheus.NewDesc("test_channel_width", "test", []string{"mac", "radio"}, nil),
		channelUtilizationDesc: prometheus.NewDesc("test_channel_util", "test", []string{"mac", "radio"}, nil),
		rxUtilizationDesc:      prometheus.NewDesc("test_rx_util", "test", []string{"mac", "radio"}, nil),
		txUtilizationDesc:      prometheus.NewDesc("test_tx_util", "test", []string{"mac", "radio"}, nil),
		noiseUtilizationDesc:   prometheus.NewDesc("test_noise_util", "test", []string{"mac", "radio"}, nil),
		noiseFloorDesc:         prometheus.NewDesc("test_noise_floor", "test", []string{"mac", "radio"}, nil),
	}

	load := &rrm.Load{CcaUtilPercentage: 30, RxUtilPercentage: 10, TxUtilPercentage: 5}
	// The operating channel is deliberately not the first entry: reading a fixed
	// index reports a channel the radio is not on.
	noise := &rrm.Noise{Noise: rrm.NoiseData{NoiseData: []rrm.NoiseDataItem{
		{Chan: 40, Noise: -100},
		{Chan: operatingChannel, Noise: -95},
	}}}
	offChannel := &rrm.Noise{Noise: rrm.NoiseData{NoiseData: []rrm.NoiseDataItem{
		{Chan: 40, Noise: -100},
	}}}

	tests := []struct {
		name        string
		measurement *rrm.RRMMeasurement
		// PhyHtCfg yields channel and width, Load 4, noise 1, associated clients 1.
		expected int
	}{
		{"Load and Noise present", &rrm.RRMMeasurement{Load: load, Noise: noise}, 8},
		{"Noise absent", &rrm.RRMMeasurement{Load: load}, 7},
		{"Load absent", &rrm.RRMMeasurement{Noise: noise}, 4},
		{"Load and Noise absent", &rrm.RRMMeasurement{}, 3},
		{"Noise present without noise data", &rrm.RRMMeasurement{Load: load, Noise: &rrm.Noise{}}, 7},
		{"Noise present for another channel only", &rrm.RRMMeasurement{Load: load, Noise: offChannel}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rrmMap := map[string]*rrm.RRMMeasurement{"aa:bb:cc:dd:ee:ff:0": tt.measurement}

			ch := make(chan prometheus.Metric, 20)
			go func() {
				defer close(ch)
				collector.collectRadioMetrics(ch, radio, rrmMap, map[string]map[int]int{}, nil)
			}()

			metricCount := 0
			for range ch {
				metricCount++
			}

			if metricCount != tt.expected {
				t.Errorf("collectRadioMetrics() emitted %d metrics, want %d", metricCount, tt.expected)
			}
		})
	}
}

// radioMetricsOnly adapts collectRadioMetrics to prometheus.Collector for value assertions.
type radioMetricsOnly struct {
	collector       *APCollector
	radio           *ap.RadioOperData
	rrmMap          map[string]*rrm.RRMMeasurement
	clientCountsMap map[string]map[int]int
	radioSlotMap    map[string]*rrm.RadioSlot
}

func (r radioMetricsOnly) Describe(_ chan<- *prometheus.Desc) {}

func (r radioMetricsOnly) Collect(ch chan<- prometheus.Metric) {
	r.collector.collectRadioMetrics(ch, r.radio, r.rrmMap, r.clientCountsMap, r.radioSlotMap)
}

// gatherRadioValues indexes the samples collectRadioMetrics produced by metric name.
func gatherRadioValues(t *testing.T, source radioMetricsOnly) map[string]float64 {
	t.Helper()

	registry := prometheus.NewRegistry()
	registry.MustRegister(source)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	values := make(map[string]float64, len(families))
	for _, family := range families {
		values[family.GetName()] = family.GetMetric()[0].GetGauge().GetValue()
	}
	return values
}

// TestAPCollector_collectRadioMetrics_SelectsOperatingBand asserts the values rather
// than the count. A dual band radio carries a radio-band-info record for the band it
// is not using, and reading a fixed index reported that record instead.
func TestAPCollector_collectRadioMetrics_SelectsOperatingBand(t *testing.T) {
	t.Parallel()

	const (
		idleBandID      = 1
		operatingBandID = 2
		idleBandPower   = 22
		activePower     = 7
		activeMaxPower  = 19
	)

	source := radioMetricsOnly{
		collector: &APCollector{
			metrics:        APMetrics{Radio: true},
			txPowerDesc:    prometheus.NewDesc("wnc_ap_tx_power_dbm", "t", []string{"mac", "radio"}, nil),
			txPowerMaxDesc: prometheus.NewDesc("wnc_ap_tx_power_max_dbm", "t", []string{"mac", "radio"}, nil),
		},
		radio: &ap.RadioOperData{
			WtpMAC:            "aa:bb:cc:dd:ee:ff",
			RadioSlotID:       2,
			RadioType:         "radio-80211-xor-5-6ghz",
			OperState:         APRadioStateUp,
			CurrentBandID:     operatingBandID,
			CurrentActiveBand: "dot11-6-ghz-band",
			RadioBandInfo: []ap.RadioBandInfo{
				{
					BandID: idleBandID,
					PhyTxPwrLvlCfg: ap.PhyTxPwrLvlCfg{CfgData: ap.PhyTxPwrLvlCfgData{
						CurrTxPowerInDbm: idleBandPower, TxPowerLevel1: idleBandPower,
					}},
				},
				{
					BandID: operatingBandID,
					PhyTxPwrLvlCfg: ap.PhyTxPwrLvlCfg{CfgData: ap.PhyTxPwrLvlCfgData{
						CurrTxPowerInDbm: activePower, TxPowerLevel1: activeMaxPower,
					}},
				},
			},
		},
	}

	values := gatherRadioValues(t, source)

	if got := values["wnc_ap_tx_power_dbm"]; got != activePower {
		t.Errorf("wnc_ap_tx_power_dbm = %v, want %v from the operating band's record", got, activePower)
	}
	if got := values["wnc_ap_tx_power_max_dbm"]; got != activeMaxPower {
		t.Errorf("wnc_ap_tx_power_max_dbm = %v, want %v from the operating band's record",
			got, activeMaxPower)
	}
}

// TestAPCollector_collectRadioMetrics_SelectsOperatingChannel asserts the noise value
// comes from the channel in use rather than from a fixed index of the per-channel list.
func TestAPCollector_collectRadioMetrics_SelectsOperatingChannel(t *testing.T) {
	t.Parallel()

	const (
		operatingChannel = 69
		operatingNoise   = -84
		otherNoise       = -103
	)

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 2,
		OperState:   APRadioStateUp,
		PhyHtCfg:    &ap.PhyHtCfg{CfgData: ap.PhyHtCfgData{CurrFreq: operatingChannel}},
	}

	source := radioMetricsOnly{
		collector: &APCollector{
			metrics:          APMetrics{Radio: true},
			noiseFloorDesc:   prometheus.NewDesc("wnc_ap_noise_floor_dbm", "t", []string{"mac", "radio"}, nil),
			channelDesc:      prometheus.NewDesc("wnc_ap_channel_number", "t", []string{"mac", "radio"}, nil),
			channelWidthDesc: prometheus.NewDesc("wnc_ap_channel_width_mhz", "t", []string{"mac", "radio"}, nil),
		},
		radio: radio,
		rrmMap: map[string]*rrm.RRMMeasurement{
			"aa:bb:cc:dd:ee:ff:2": {Noise: &rrm.Noise{Noise: rrm.NoiseData{NoiseData: []rrm.NoiseDataItem{
				{Chan: 1, Noise: otherNoise},
				{Chan: operatingChannel, Noise: operatingNoise},
			}}}},
		},
	}

	if got := gatherRadioValues(t, source)["wnc_ap_noise_floor_dbm"]; got != operatingNoise {
		t.Errorf("wnc_ap_noise_floor_dbm = %v, want %v from the operating channel", got, operatingNoise)
	}
}

// TestAPCollector_collectRadioMetrics_ScalesUtilization pins the unit conversion.
// The controller reports these leaves in whole percent, so a name that promises a
// ratio without the division would report a value a hundred times too large, and
// no other assertion in this suite reads their value.
func TestAPCollector_collectRadioMetrics_ScalesUtilization(t *testing.T) {
	t.Parallel()

	source := radioMetricsOnly{
		collector: &APCollector{
			metrics: APMetrics{Radio: true},
			channelUtilizationDesc: prometheus.NewDesc(
				"wnc_ap_channel_utilization_ratio", "t", []string{"mac", "radio"}, nil),
			rxUtilizationDesc: prometheus.NewDesc(
				"wnc_ap_rx_utilization_ratio", "t", []string{"mac", "radio"}, nil),
			txUtilizationDesc: prometheus.NewDesc(
				"wnc_ap_tx_utilization_ratio", "t", []string{"mac", "radio"}, nil),
			noiseUtilizationDesc: prometheus.NewDesc(
				"wnc_ap_noise_utilization_ratio", "t", []string{"mac", "radio"}, nil),
		},
		radio: &ap.RadioOperData{WtpMAC: "aa:bb:cc:dd:ee:ff", RadioSlotID: 0, OperState: APRadioStateUp},
		rrmMap: map[string]*rrm.RRMMeasurement{
			"aa:bb:cc:dd:ee:ff:0": {Load: &rrm.Load{
				CcaUtilPercentage: 30, RxUtilPercentage: 10,
				TxUtilPercentage: 5, RxNoiseChannelUtilization: 15,
			}},
		},
	}

	values := gatherRadioValues(t, source)

	if got := values["wnc_ap_channel_utilization_ratio"]; got != 0.3 {
		t.Errorf("wnc_ap_channel_utilization_ratio = %v, want 0.3 from a leaf reporting whole percent", got)
	}
	if got := values["wnc_ap_rx_utilization_ratio"]; got != 0.1 {
		t.Errorf("wnc_ap_rx_utilization_ratio = %v, want 0.1 from a leaf reporting whole percent", got)
	}
	if got := values["wnc_ap_tx_utilization_ratio"]; got != 0.05 {
		t.Errorf("wnc_ap_tx_utilization_ratio = %v, want 0.05 from a leaf reporting whole percent", got)
	}
	if got := values["wnc_ap_noise_utilization_ratio"]; got != 0.15 {
		t.Errorf("wnc_ap_noise_utilization_ratio = %v, want 0.15 from a leaf reporting whole percent", got)
	}
}

func TestAPCollector_collectRadioMetrics_ClientCount(t *testing.T) {
	t.Parallel()

	clientsDesc := prometheus.NewDesc(
		"wnc_ap_clients", "Number of clients in the run state on this radio", []string{"mac", "radio"}, nil,
	)

	source := radioMetricsOnly{
		collector: &APCollector{
			metrics:               APMetrics{Radio: true},
			associatedClientsDesc: clientsDesc,
		},
		radio: &ap.RadioOperData{
			WtpMAC:      "aa:bb:cc:dd:ee:ff",
			RadioSlotID: 1,
			RadioType:   "dot11-5ghz-radio",
			OperState:   APRadioStateUp,
		},
		clientCountsMap: map[string]map[int]int{
			"aa:bb:cc:dd:ee:ff": {0: 7, 1: 3},
		},
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(source)

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() returned error: %v", err)
	}

	const wantClientCount = 3

	found := false
	for _, mf := range metricFamilies {
		if mf.GetName() != "wnc_ap_clients" {
			continue
		}
		found = true
		if got := mf.GetMetric()[0].GetGauge().GetValue(); got != wantClientCount {
			t.Errorf("wnc_ap_clients = %v, want %v", got, wantClientCount)
		}
	}

	if !found {
		t.Error("wnc_ap_clients was not emitted")
	}
}

func TestAPCollector_collectTrafficMetrics(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{
		WtpMAC:      "aa:bb:cc:dd:ee:ff",
		RadioSlotID: 0,
		RadioType:   "dot11-5ghz-radio",
		OperState:   APRadioStateUp,
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
		dataRxFramesTotalDesc:       prometheus.NewDesc("test_data_rx", "test", []string{"mac", "radio"}, nil),
		dataTxFramesTotalDesc:       prometheus.NewDesc("test_data_tx", "test", []string{"mac", "radio"}, nil),
		managementRxFramesTotalDesc: prometheus.NewDesc("test_mgmt_rx", "test", []string{"mac", "radio"}, nil),
		managementTxFramesTotalDesc: prometheus.NewDesc("test_mgmt_tx", "test", []string{"mac", "radio"}, nil),
		controlRxFramesTotalDesc:    prometheus.NewDesc("test_ctrl_rx", "test", []string{"mac", "radio"}, nil),
		controlTxFramesTotalDesc:    prometheus.NewDesc("test_ctrl_tx", "test", []string{"mac", "radio"}, nil),
		multicastRxFramesTotalDesc:  prometheus.NewDesc("test_mcast_rx", "test", []string{"mac", "radio"}, nil),
		multicastTxFramesTotalDesc:  prometheus.NewDesc("test_mcast_tx", "test", []string{"mac", "radio"}, nil),
		totalTxFramesTotalDesc:      prometheus.NewDesc("test_total_tx", "test", []string{"mac", "radio"}, nil),
		rtsSuccessesTotalDesc:       prometheus.NewDesc("test_rts_success", "test", []string{"mac", "radio"}, nil),
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
		OperState:   APRadioStateUp,
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
			},
		},
	}
	resetStatsMap := map[string]map[int]int{}
	rrmCoverageMap := map[string]*rrm.RRMCoverage{}
	apDot11RadarMap := map[string]*rrm.ApDot11RadarData{}

	collector := &APCollector{
		metrics:                       APMetrics{Errors: true},
		rxErrorsTotalDesc:             prometheus.NewDesc("test_rx_errors", "test", []string{"mac", "radio"}, nil),
		txRetriesTotalDesc:            prometheus.NewDesc("test_tx_retries", "test", []string{"mac", "radio"}, nil),
		transmissionFailuresTotalDesc: prometheus.NewDesc("test_ack_failures", "test", []string{"mac", "radio"}, nil),
		duplicateFramesTotalDesc:      prometheus.NewDesc("test_duplicates", "test", []string{"mac", "radio"}, nil),
		fcsErrorsTotalDesc:            prometheus.NewDesc("test_fcs_errors", "test", []string{"mac", "radio"}, nil),
		rxFragmentsTotalDesc:          prometheus.NewDesc("test_frag_rx", "test", []string{"mac", "radio"}, nil),
		txFragmentsTotalDesc:          prometheus.NewDesc("test_frag_tx", "test", []string{"mac", "radio"}, nil),
		rtsFailuresTotalDesc:          prometheus.NewDesc("test_rts_failures", "test", []string{"mac", "radio"}, nil),
		decryptionErrorsTotalDesc:     prometheus.NewDesc("test_decrypt_errors", "test", []string{"mac", "radio"}, nil),
		micErrorsTotalDesc:            prometheus.NewDesc("test_mic_errors", "test", []string{"mac", "radio"}, nil),
		coverageFailedClientsDesc:     prometheus.NewDesc("test_coverage_holes", "test", []string{"mac", "radio"}, nil),
		lastRadarOnRadioAtDesc:        prometheus.NewDesc("test_last_radar", "test", []string{"mac", "radio"}, nil),
		radioResetsTotalDesc:          prometheus.NewDesc("test_radio_resets", "test", []string{"mac", "radio"}, nil),
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
					collector.collectRadioMetrics(ch, nil, nil, nil, nil)
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

// apSnapshotValues gathers the AP collector over a caller-supplied snapshot and
// indexes the first sample of every family by name.
//
// fullFixtureSnapshot returns a fresh struct on every call, so a subtest may
// rewrite a leaf without affecting the others.
func apSnapshotValues(t *testing.T, data *wnc.WNCDataCache) map[string]float64 {
	t.Helper()

	src := fixtureSource{data: data}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewAPCollector(
		wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src),
		APMetrics{General: true, Radio: true, Errors: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	values := make(map[string]float64, len(families))
	for _, family := range families {
		metrics := family.GetMetric()
		if len(metrics) == 0 || metrics[0].GetGauge() == nil {
			continue
		}
		values[family.GetName()] = metrics[0].GetGauge().GetValue()
	}
	return values
}

// TestAPCollector_RadioAndAdminStateReportTheirOwnLeaf pins both polarities of the
// two radio state gauges. One case per polarity is not enough: with both leaves
// high, and with both low, swapping the two comparisons or replacing either with a
// constant leaves every sample unchanged. Mixing the polarities is what separates
// them. The test asserts only that a spelling other than the enabled constant
// reports zero, not that the leaf has a closed value domain.
func TestAPCollector_RadioAndAdminStateReportTheirOwnLeaf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operState  string
		adminState string
		wantRadio  float64
		wantAdmin  float64
	}{
		{"radio up while administratively disabled", APRadioStateUp, "disabled", 1, 0},
		{"radio down while administratively enabled", "radio-down", APAdminStateEnabled, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.RadioOperData[0].OperState = tt.operState
			data.RadioOperData[0].AdminState = tt.adminState

			values := apSnapshotValues(t, data)
			if got := values["wnc_ap_radio_state"]; got != tt.wantRadio {
				t.Errorf("wnc_ap_radio_state = %v, want %v from oper-state %q", got, tt.wantRadio, tt.operState)
			}
			if got := values["wnc_ap_admin_state"]; got != tt.wantAdmin {
				t.Errorf("wnc_ap_admin_state = %v, want %v from admin-state %q", got, tt.wantAdmin, tt.adminState)
			}
		})
	}
}

// TestAPCollector_ConfigStateReportsMisconfiguration pins the polarity that runs
// opposite to its siblings: radio_state, admin_state and wlan_enabled report one
// when healthy, while this one reports one when the AP is misconfigured. The
// asymmetry is deliberate and documented in the HELP string, so both directions
// are asserted here to keep a later "correction" from passing silently.
func TestAPCollector_ConfigStateReportsMisconfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		misconfigured bool
		want          float64
	}{
		{"configuration valid", false, 0},
		{"configuration invalid", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.CAPWAPData[0].TagInfo.IsApMisconfigured = tt.misconfigured

			if got := apSnapshotValues(t, data)["wnc_ap_config_state"]; got != tt.want {
				t.Errorf("wnc_ap_config_state = %v, want %v with is-ap-misconfigured %v",
					got, tt.want, tt.misconfigured)
			}
		})
	}
}

// TestAPCollector_RadarTimestampOmitsUnpopulatedLeaf pins when the radar series is
// published at all. A controller that has never seen radar renders the leaf at the
// Unix epoch, and reporting that as a detection would date every DFS event to 1970.
//
// The guard compares a calendar year rather than the epoch instant, so a genuine
// detection inside 1970 is suppressed too. That case asserts today's behavior and
// records the defect; a fix flips its expectation.
func TestAPCollector_RadarTimestampOmitsUnpopulatedLeaf(t *testing.T) {
	t.Parallel()

	const metric = "wnc_ap_last_radar_timestamp_seconds"

	tests := []struct {
		name    string
		last    time.Time
		present bool
	}{
		{"leaf absent from the payload", time.Time{}, false},
		{"rendered at the epoch", time.Unix(0, 0).UTC(), false},
		{"after the epoch but inside 1970", time.Date(1970, 6, 1, 0, 0, 0, 0, time.UTC), false},
		{"a detection this year", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.ApDot11RadarData[0].LastRadarOnRadio = tt.last

			_, ok := apSnapshotValues(t, data)[metric]
			if ok != tt.present {
				t.Errorf("%s present = %v, want %v for %s", metric, ok, tt.present, tt.last)
			}
		})
	}
}

// gatherJoinValues collects the AP collector with only the join module enabled and
// indexes every sample by metric name, then by the value of the given label. A series
// that does not carry that label is indexed under the empty string.
func gatherJoinValues(t *testing.T, data *wnc.WNCDataCache, label string) map[string]map[string]float64 {
	t.Helper()

	src := fixtureSource{data: data}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewAPCollector(
		wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src),
		APMetrics{Join: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	values := make(map[string]map[string]float64, len(families))
	for _, family := range families {
		byLabel := make(map[string]float64, len(family.GetMetric()))
		for _, metric := range family.GetMetric() {
			key := ""
			for _, pair := range metric.GetLabel() {
				if pair.GetName() == label {
					key = pair.GetValue()
				}
			}
			switch {
			case metric.GetGauge() != nil:
				byLabel[key] = metric.GetGauge().GetValue()
			case metric.GetCounter() != nil:
				byLabel[key] = metric.GetCounter().GetValue()
			}
		}
		values[family.GetName()] = byLabel
	}
	return values
}

// TestAPJoinModule_ChannelSeriesReadTheirOwnLeaf pins the DTLS series to the channel
// they are read from. The controller keeps both channels in one container with leaf
// names that differ by a single token, so folding them into one series with a label
// is exactly where a swap goes unnoticed: every count and every label name stays
// intact and only the two values exchange places.
func TestAPJoinModule_ChannelSeriesReadTheirOwnLeaf(t *testing.T) {
	t.Parallel()

	values := gatherJoinValues(t, fullFixtureSnapshot(), labelChannel)

	tests := []struct {
		name        string
		wantControl float64
		wantData    float64
	}{
		{"wnc_ap_dtls_session_requests_total", 5301, 5401},
		{"wnc_ap_dtls_session_successes_total", 5302, 5402},
		{"wnc_ap_dtls_session_failures_total", 5303, 5403},
		{"wnc_ap_dtls_decrypt_errors_total", 5304, 5404},
		{"wnc_ap_dtls_anti_replay_errors_total", 5305, 5405},
		{"wnc_ap_last_dtls_success_timestamp_seconds", 1767916800, 1768089600},
		{"wnc_ap_last_dtls_failure_timestamp_seconds", 1768003200, 1768176000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			byChannel, ok := values[tt.name]
			if !ok {
				t.Fatalf("%s has no series in the all-succeed snapshot", tt.name)
			}
			if got := byChannel[dtlsChannelControl]; got != tt.wantControl {
				t.Errorf("%s{channel=%q} = %v, want %v", tt.name, dtlsChannelControl, got, tt.wantControl)
			}
			if got := byChannel[dtlsChannelData]; got != tt.wantData {
				t.Errorf("%s{channel=%q} = %v, want %v", tt.name, dtlsChannelData, got, tt.wantData)
			}
		})
	}
}

// TestAPJoinModule_WithholdsTheEpochSentinel pins the withhold. The controller writes
// 1970-01-01T00:00:00+00:00 into a timestamp leaf for an event that has not happened,
// which parses to a real instant, so IsZero does not recognize it and a naive guard
// publishes a gauge reporting an event five decades ago.
func TestAPJoinModule_WithholdsTheEpochSentinel(t *testing.T) {
	t.Parallel()

	data := fullFixtureSnapshot()
	record := &data.JoinStats[0]
	record.ApJoinInfo.LastFailJoinAtmptTime = fixtureEpochSentinel
	record.ApJoinInfo.LastFailConfAtmptTime = fixtureEpochSentinel
	record.ApDiscoveryInfo.LastFailedDiscTime = fixtureEpochSentinel
	record.DTLSSessInfo.CtrlDTLSFailureTime = fixtureEpochSentinel
	record.DTLSSessInfo.DataDTLSFailureTime = fixtureEpochSentinel

	values := gatherJoinValues(t, data, labelChannel)

	withheld := []string{
		"wnc_ap_last_join_failure_timestamp_seconds",
		"wnc_ap_last_config_failure_timestamp_seconds",
		"wnc_ap_last_discovery_failure_timestamp_seconds",
		"wnc_ap_last_dtls_failure_timestamp_seconds",
	}
	for _, name := range withheld {
		if len(values[name]) != 0 {
			t.Errorf("%s has %d series for the epoch sentinel, want it withheld", name, len(values[name]))
		}
	}

	// Without these the assertions above would also pass on a module that publishes
	// no timestamp at all.
	present := []string{
		"wnc_ap_last_join_success_timestamp_seconds",
		"wnc_ap_last_config_success_timestamp_seconds",
		"wnc_ap_last_discovery_success_timestamp_seconds",
		"wnc_ap_last_dtls_success_timestamp_seconds",
		"wnc_ap_last_error_timestamp_seconds",
	}
	for _, name := range present {
		if len(values[name]) == 0 {
			t.Errorf("%s is absent, so the withholds above prove nothing", name)
		}
	}
}

// TestAPJoinModule_ReasonsCarryTheControllerSpelling pins each enum leaf to its own
// series. The two DTLS channels carry different spellings in the fixture, so a swap
// between them changes which label value each channel reports.
func TestAPJoinModule_ReasonsCarryTheControllerSpelling(t *testing.T) {
	t.Parallel()

	values := gatherJoinValues(t, fullFixtureSnapshot(), labelState)

	tests := []struct {
		name  string
		state string
	}{
		{"wnc_ap_last_discovery_failure_reason", "disc-fail-none"},
		{"wnc_ap_last_join_failure_reason", "jf-none"},
		{"wnc_ap_last_config_failure_reason", "cf-none"},
		{"wnc_ap_last_error_phase", "ap-con-failure-run"},
		{"wnc_ap_last_reboot_reason", "ap-reboot-reason-reboot-cmd"},
		{"wnc_ap_last_disconnect_reason", "wtp-controller-initiated-reason"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			byState := values[tt.name]
			if len(byState) != 1 {
				t.Fatalf("%s has %d series, want one per AP", tt.name, len(byState))
			}
			if got, ok := byState[tt.state]; !ok || got != 1 {
				t.Errorf("%s{state=%q} = %v (present %v), want 1", tt.name, tt.state, got, ok)
			}
		})
	}

	byState := values["wnc_ap_last_dtls_failure_reason"]
	for state, want := range map[string]bool{"dtls-hs-success": true, "dtls-hs-fragment-error": true} {
		if _, ok := byState[state]; ok != want {
			t.Errorf("wnc_ap_last_dtls_failure_reason{state=%q} present = %v, want %v", state, ok, want)
		}
	}
}

// TestAPJoinModule_FreeTextLeavesAreNotPublished keeps the two prose leaves of the
// record out of the label set. Neither has a value domain, so publishing one would
// make the label set unbounded and the series unusable in a match.
func TestAPJoinModule_FreeTextLeavesAreNotPublished(t *testing.T) {
	t.Parallel()

	src := fixtureSource{data: fullFixtureSnapshot()}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewAPCollector(
		wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src),
		APMetrics{Join: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	prose := []string{"Tag modified", "fixture decryption failure text"}
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			for _, pair := range metric.GetLabel() {
				for _, text := range prose {
					if pair.GetValue() == text {
						t.Errorf("%s carries the free-text leaf %q in label %q",
							family.GetName(), text, pair.GetName())
					}
				}
			}
		}
	}
}

// TestAPJoinModule_NamesAnAPThatLeftCAPWAP is the reason this module exists. The
// statistics list keeps a record for an AP the controller no longer lists in the AP
// inventory, so this is the only place such an AP has any series at all, and
// ap-join-info is the only container that still carries its name.
func TestAPJoinModule_NamesAnAPThatLeftCAPWAP(t *testing.T) {
	t.Parallel()

	const departedMAC = "bb:cc:dd:ee:ff:00"

	departed := newFixtureJoinStats()
	departed.WtpMAC = departedMAC
	departed.ApJoinInfo.ApName = "TEST-AP99"
	departed.ApJoinInfo.IsJoined = false

	data := fullFixtureSnapshot()
	data.JoinStats = append(data.JoinStats, departed)

	values := gatherJoinValues(t, data, labelMAC)

	if got := values["wnc_ap_joined"][departedMAC]; got != 0 {
		t.Errorf("wnc_ap_joined{mac=departed} = %v, want 0", got)
	}
	if got := values["wnc_ap_joined"][fixtureAPMAC]; got != 1 {
		t.Errorf("wnc_ap_joined{mac=joined} = %v, want 1", got)
	}
	if _, ok := values["wnc_ap_join_info"][departedMAC]; !ok {
		t.Error("wnc_ap_join_info has no series for the departed AP, so an alert on it cannot name the AP")
	}

	// The discovery counters keep advancing while the session is gone, which is what
	// makes rate(discovery) > 0 and wnc_ap_joined == 0 a signal rather than a tautology.
	if _, ok := values["wnc_ap_discovery_requests_total"][departedMAC]; !ok {
		t.Error("wnc_ap_discovery_requests_total has no series for the departed AP")
	}

	// A bare and requires identical label sets, so every series of this module has to
	// carry the mac label, and it has to be the list key the AP inventory is keyed by.
	for name, byMAC := range values {
		if _, ok := byMAC[""]; ok {
			t.Errorf("%s has a series carrying no mac label", name)
		}
		for mac := range byMAC {
			if mac != fixtureAPMAC && mac != departedMAC {
				t.Errorf("%s carries mac=%q, want the wtp-mac list key", name, mac)
			}
		}
	}
}

// TestAPJoinModule_NameSeriesIgnoresTheInfoFlag pins the name series to its own
// module. It is named for the info family so that the info cache holds it, and the
// cache wrapper is only applied when the info module is enabled, so the series has to
// be published either way.
func TestAPJoinModule_NameSeriesIgnoresTheInfoFlag(t *testing.T) {
	t.Parallel()

	values := gatherJoinValues(t, fullFixtureSnapshot(), labelName)
	if got := values["wnc_ap_join_info"][fixtureAPName]; got != 1 {
		t.Errorf("wnc_ap_join_info{name=%q} = %v with the info module disabled, want 1", fixtureAPName, got)
	}
}

// TestAPCollector_RRMProfilesMatchLeaves binds each profile label value to its own
// verdict leaf. Four booleans over one descriptor cannot be told apart by a single
// fixture — a swap between two leaves reporting the same value is invisible — so each
// case sets exactly one verdict, which makes every pairwise swap fail.
func TestAPCollector_RRMProfilesMatchLeaves(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{WtpMAC: fixtureAPMAC, RadioSlotID: 0, OperState: APRadioStateUp}

	tests := []struct {
		profile string
		data    *rrm.RadioData
	}{
		{"coverage", &rrm.RadioData{CoverageProfilePassed: true}},
		{"load", &rrm.RadioData{LoadProfPassed: true}},
		{"interference", &rrm.RadioData{InterferenceProfilePassed: true}},
		{"noise", &rrm.RadioData{NoiseProfilePassed: true}},
	}

	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			t.Parallel()

			collector := NewAPCollector(nil, nil, nil, APMetrics{Radio: true})
			registry := prometheus.NewRegistry()
			registry.MustRegister(radioMetricsOnly{
				collector: collector,
				radio:     radio,
				rrmMap:    map[string]*rrm.RRMMeasurement{},
				radioSlotMap: map[string]*rrm.RadioSlot{
					fixtureAPMAC + ":0": {WtpMAC: fixtureAPMAC, RadioSlotID: 0, RadioData: tt.data},
				},
			})

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf("Gather() error = %v, want nil", err)
			}

			got := make(map[string]float64)
			for _, family := range families {
				if family.GetName() != "wnc_ap_rrm_profile_passed" {
					continue
				}
				for _, metric := range family.GetMetric() {
					for _, pair := range metric.GetLabel() {
						if pair.GetName() == labelProfile {
							got[pair.GetValue()] = metric.GetGauge().GetValue()
						}
					}
				}
			}

			if len(got) != len(tests) {
				t.Fatalf("wnc_ap_rrm_profile_passed has %d series, want %d: %v", len(got), len(tests), got)
			}
			for _, other := range tests {
				want := 0.0
				if other.profile == tt.profile {
					want = 1
				}
				if got[other.profile] != want {
					t.Errorf("wnc_ap_rrm_profile_passed{profile=%q} = %v, want %v with only %s passing",
						other.profile, got[other.profile], want, tt.profile)
				}
			}
		})
	}
}

// TestAPCollector_RRMVerdictsAbsentWithoutTheContainer covers the one absence the slot
// list has of its own: a record present with no radio-data container. Publishing zero
// there would report four failed profiles on a radio the controller judged nothing on.
func TestAPCollector_RRMVerdictsAbsentWithoutTheContainer(t *testing.T) {
	t.Parallel()

	radio := &ap.RadioOperData{WtpMAC: fixtureAPMAC, RadioSlotID: 0, OperState: APRadioStateUp}

	tests := []struct {
		name string
		slot map[string]*rrm.RadioSlot
	}{
		{"no record for this radio", map[string]*rrm.RadioSlot{}},
		{"record without radio-data", map[string]*rrm.RadioSlot{
			fixtureAPMAC + ":0": {WtpMAC: fixtureAPMAC, RadioSlotID: 0},
		}},
		{"fetch failed", nil},
		{"record with radio-data but no dca-stats", map[string]*rrm.RadioSlot{
			fixtureAPMAC + ":0": {
				WtpMAC: fixtureAPMAC, RadioSlotID: 0,
				RadioData: &rrm.RadioData{CoverageProfilePassed: true},
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewAPCollector(nil, nil, nil, APMetrics{Radio: true})
			registry := prometheus.NewRegistry()
			registry.MustRegister(radioMetricsOnly{
				collector:    collector,
				radio:        radio,
				rrmMap:       map[string]*rrm.RRMMeasurement{},
				radioSlotMap: tt.slot,
			})

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf("Gather() error = %v, want nil", err)
			}
			// The deeper container has its own rule: the last case carries radio-data
			// without dca-stats, so the verdicts publish while the counter does not.
			withVerdicts := tt.slot != nil && tt.slot[fixtureAPMAC+":0"] != nil &&
				tt.slot[fixtureAPMAC+":0"].RadioData != nil
			for _, family := range families {
				switch family.GetName() {
				case "wnc_ap_rrm_profile_passed":
					if !withVerdicts {
						t.Errorf("wnc_ap_rrm_profile_passed has %d series, want none",
							len(family.GetMetric()))
					}
				case "wnc_ap_channel_changes_total":
					t.Errorf("wnc_ap_channel_changes_total has %d series, want none: "+
						"a zero there reads as a radio DCA has never moved",
						len(family.GetMetric()))
				}
			}
		})
	}
}

// TestAPCollector_ChannelEnergyWithholdsTheSentinels pins the two readings the energy
// leaf carries that are not measurements. Publishing either would report a channel far
// quieter than any radio can measure, in the direction that reads as a free channel.
func TestAPCollector_ChannelEnergyWithholdsTheSentinels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		energy int
		want   bool
		reason string
	}{
		{"a measurement", -32, true, "an ordinary reading is published"},
		{"the sentinel", channelEnergySentinel, false, "it is the lower bound of the leaf's own type"},
		{"zero", channelEnergyAbsent, false, "an omitted leaf decodes to it, and it is above every reading"},
	}

	radioLabels := []string{labelMAC, labelRadio}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			radio := &ap.RadioOperData{
				WtpMAC:      fixtureAPMAC,
				RadioSlotID: 0,
				OperState:   APRadioStateUp,
			}
			values := gatherRadioValues(t, radioMetricsOnly{
				collector: &APCollector{
					metrics: APMetrics{Radio: true},
					channelEnergyDesc: prometheus.NewDesc(
						"wnc_ap_channel_energy_dbm", "t", radioLabels, nil),
					channelChangesTotalDesc: prometheus.NewDesc(
						"wnc_ap_channel_changes_total", "t", radioLabels, nil),
					rrmProfilePassedDesc: prometheus.NewDesc(
						"wnc_ap_rrm_profile_passed", "t", append(radioLabels, "profile"), nil),
				},
				radio: radio,
				radioSlotMap: map[string]*rrm.RadioSlot{
					fixtureAPMAC + ":0": {RadioData: &rrm.RadioData{
						DCAStats: &rrm.DCAStats{CurrentChanEnergy: tt.energy, LastChanEnergy: -33},
					}},
				},
			})

			got, published := values["wnc_ap_channel_energy_dbm"]
			if published != tt.want {
				t.Fatalf("wnc_ap_channel_energy_dbm published = %v, want %v: %s", published, tt.want, tt.reason)
			}

			if published && got != float64(tt.energy) {
				t.Errorf("wnc_ap_channel_energy_dbm = %f, want %d", got, tt.energy)
			}
		})
	}
}

// TestAirQualityOnCurrentChannel covers every way the join can miss. The table is keyed
// by AP and band and its per-channel list is a padded array, so a reader that takes a
// fixed index, ignores the band, or trusts a zero channel reports another radio's air
// quality or a fabricated worst-case reading.
func TestAirQualityOnCurrentChannel(t *testing.T) {
	t.Parallel()

	table := []rrm.SpectrumAqTable{
		{
			WtpMAC: fixtureAPMAC,
			Band:   "dot11-2-dot-4-ghz-band",
			PerRadioAqData: &rrm.PerRadioAqData{PerChannelAqList: []rrm.PerChannelAqList{
				{ChannelNum: 0, Aqi: 0, MinAqi: 0, TotalIntfDeviceCount: 0},
				{ChannelNum: 11, Aqi: 71, MinAqi: 61, TotalIntfDeviceCount: 51},
				{ChannelNum: 6, Aqi: 96, MinAqi: 86, TotalIntfDeviceCount: 76},
			}},
		},
		{
			// The 6 GHz numbering restarts at 1, so this record carries a channel 6 too.
			WtpMAC: fixtureAPMAC,
			Band:   "dot11-6-ghz-band",
			PerRadioAqData: &rrm.PerRadioAqData{PerChannelAqList: []rrm.PerChannelAqList{
				{ChannelNum: 6, Aqi: 100, MinAqi: 90, TotalIntfDeviceCount: 80},
			}},
		},
	}

	radio := func(band string, channel int) *ap.RadioOperData {
		return &ap.RadioOperData{
			WtpMAC:            fixtureAPMAC,
			CurrentActiveBand: band,
			PhyHtCfg:          &ap.PhyHtCfg{CfgData: ap.PhyHtCfgData{CurrFreq: channel}},
		}
	}

	tests := []struct {
		name      string
		radio     *ap.RadioOperData
		wantValue int
		wantFound bool
		reason    string
	}{
		{
			"operating channel of the matching band", radio("dot11-2-dot-4-ghz-band", 6),
			96, true, "the row for channel 6 of the 2.4 GHz record",
		},
		{
			"same channel number in another band", radio("dot11-6-ghz-band", 6),
			100, true, "matching the channel alone would cross bands",
		},
		{
			"channel the band's list does not carry", radio("dot11-2-dot-4-ghz-band", 1),
			0, false, "no reading rather than the first row",
		},
		{
			"band the table has no record for", radio("dot11-5-ghz-band", 36),
			0, false, "a radio whose spectrum operation is down has no record",
		},
		{
			"monitor or sniffer radio with no primary channel", radio("dot11-2-dot-4-ghz-band", 0),
			0, false, "the zero channel would match the padding row and report the worst reading",
		},
		{
			"radio without the channel container",
			&ap.RadioOperData{WtpMAC: fixtureAPMAC, CurrentActiveBand: "dot11-2-dot-4-ghz-band"},
			0, false, "nothing says which channel to match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, found := airQualityOnCurrentChannel(table, tt.radio)
			if found != tt.wantFound {
				t.Fatalf("airQualityOnCurrentChannel() found = %v, want %v: %s",
					found, tt.wantFound, tt.reason)
			}

			if !found {
				return
			}

			// Every leaf of the row carries a distinct number, so a caller reading the
			// wrong one reports a value this assertion does not expect.
			if got.Aqi != tt.wantValue || got.MinAqi != tt.wantValue-10 ||
				got.TotalIntfDeviceCount != tt.wantValue-20 {
				t.Errorf("airQualityOnCurrentChannel() row = (%d, %d, %d), want (%d, %d, %d): %s",
					got.Aqi, got.MinAqi, got.TotalIntfDeviceCount,
					tt.wantValue, tt.wantValue-10, tt.wantValue-20, tt.reason)
			}
		})
	}
}

// TestAirQualityOnCurrentChannel_AnotherAPsRecord keeps the AP key load-bearing. Both
// records carry the same band and the same channel, so dropping the MAC test publishes
// one AP's air quality on the other's radio.
func TestAirQualityOnCurrentChannel_AnotherAPsRecord(t *testing.T) {
	t.Parallel()

	const otherMAC = "11:22:33:44:55:00"
	table := []rrm.SpectrumAqTable{
		{
			WtpMAC: otherMAC,
			Band:   "dot11-2-dot-4-ghz-band",
			PerRadioAqData: &rrm.PerRadioAqData{PerChannelAqList: []rrm.PerChannelAqList{
				{ChannelNum: 6, Aqi: 42},
			}},
		},
	}

	radio := &ap.RadioOperData{
		WtpMAC:            fixtureAPMAC,
		CurrentActiveBand: "dot11-2-dot-4-ghz-band",
		PhyHtCfg:          &ap.PhyHtCfg{CfgData: ap.PhyHtCfgData{CurrFreq: 6}},
	}

	if got, found := airQualityOnCurrentChannel(table, radio); found {
		t.Errorf("airQualityOnCurrentChannel() = (%v, true), want no reading: the only "+
			"record belongs to another AP", got)
	}
}

// TestAPCollector_StateSeriesAbsentOnAnEmptyLeaf covers the slot list carrying entries
// that are not radios. A remote-LAN port arrives with both state leaves omitted, and a
// string comparison against the up spelling reports it down — a permanently failing
// radio on every AP that has such a port. Each leaf is guarded on its own, because the
// controller omits per leaf rather than per record.
func TestAPCollector_StateSeriesAbsentOnAnEmptyLeaf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operState  string
		adminState string
		want       []string
		reason     string
	}{
		{
			"both leaves present", APRadioStateUp, "disabled",
			[]string{"test_radio_state", "test_admin_state"},
			"a real radio publishes both",
		},
		{
			"remote-LAN port with both leaves omitted", "", "",
			nil,
			"neither series exists rather than reporting a down, disabled radio",
		},
		{
			"operational state omitted alone", "", APAdminStateEnabled,
			[]string{"test_admin_state"},
			"the guards are independent",
		},
		{
			"admin state omitted alone", APRadioStateUp, "",
			[]string{"test_radio_state"},
			"the guards are independent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &APCollector{
				metrics:        APMetrics{General: true},
				radioStateDesc: prometheus.NewDesc("test_radio_state", "test", []string{"mac", "radio"}, nil),
				adminStateDesc: prometheus.NewDesc("test_admin_state", "test", []string{"mac", "radio"}, nil),
			}
			radio := &ap.RadioOperData{
				WtpMAC: fixtureAPMAC, RadioSlotID: 2,
				OperState: tt.operState, AdminState: tt.adminState,
			}

			ch := make(chan prometheus.Metric, 4)
			go func() {
				defer close(ch)
				collector.collectGeneralMetrics(ch, radio)
			}()

			var got []string
			for metric := range ch {
				_, quoted, _ := strings.Cut(metric.Desc().String(), "fqName: \"")
				name, _, _ := strings.Cut(quoted, "\"")
				got = append(got, name)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("collectGeneralMetrics() published %v, want %v: %s", got, tt.want, tt.reason)
			}
		})
	}
}

// TestAPCollector_PerRadioSeriesAbsentForANonRadioSlot pins the withhold for the entry
// of the slot list that is not a radio. The controller sends such an entry a counter
// record whose every counter is zero, and this fixture also gives it a reset entry, so
// every per-radio family would otherwise report a radio that never carries traffic.
func TestAPCollector_PerRadioSeriesAbsentForANonRadioSlot(t *testing.T) {
	t.Parallel()

	byRadio := gatherAPSeriesByRadio(t, fullFixtureSnapshot())
	slot := strconv.Itoa(fixturePseudoRadioSlot)

	// The info family is published for every entry of the slot list on purpose, so it
	// witnesses that the entry reached the collector at all.
	if !byRadio["wnc_ap_info"][slot] {
		t.Fatalf("wnc_ap_info has no series for slot %s, so the absences below prove nothing", slot)
	}

	// The three sources fail differently: the counter record exists for the entry, the
	// reset list carries an entry for it, and the client count map has none and yielded
	// Go's zero value. The reset counter is emitted before the counter-record lookup, so
	// it is the one that survives a guard placed on that lookup instead of at the top.
	perRadio := []string{
		"wnc_ap_data_rx_frames_total",
		"wnc_ap_data_tx_frames_total",
		"wnc_ap_management_rx_frames_total",
		"wnc_ap_management_tx_frames_total",
		"wnc_ap_control_rx_frames_total",
		"wnc_ap_control_tx_frames_total",
		"wnc_ap_multicast_rx_frames_total",
		"wnc_ap_multicast_tx_frames_total",
		"wnc_ap_total_tx_frames_total",
		"wnc_ap_rts_successes_total",
		"wnc_ap_rx_errors_total",
		"wnc_ap_tx_retries_total",
		"wnc_ap_transmission_failures_total",
		"wnc_ap_duplicate_frames_total",
		"wnc_ap_fcs_errors_total",
		"wnc_ap_rx_fragments_total",
		"wnc_ap_tx_fragments_total",
		"wnc_ap_rts_failures_total",
		"wnc_ap_decryption_errors_total",
		"wnc_ap_mic_errors_total",
		"wnc_ap_radio_resets_total",
		"wnc_ap_clients",
	}

	for _, name := range perRadio {
		if len(byRadio[name]) == 0 {
			t.Errorf("%s has no series at all, so its absence for slot %s proves nothing", name, slot)
			continue
		}

		if byRadio[name][slot] {
			t.Errorf("%s carries a series for slot %s, which is not a radio, want it withheld", name, slot)
		}
	}
}

// gatherAPSeriesByRadio indexes, for every family the AP collector publishes over the
// given snapshot, the slot numbers its series carry in the radio label.
func gatherAPSeriesByRadio(t *testing.T, data *wnc.WNCDataCache) map[string]map[string]bool {
	t.Helper()

	src := fixtureSource{data: data}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewAPCollector(
		wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src),
		APMetrics{Radio: true, Traffic: true, Errors: true, Info: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	byRadio := make(map[string]map[string]bool, len(families))
	for _, family := range families {
		slots := make(map[string]bool, len(family.GetMetric()))
		for _, metric := range family.GetMetric() {
			for _, pair := range metric.GetLabel() {
				if pair.GetName() == labelRadio {
					slots[pair.GetValue()] = true
				}
			}
		}
		byRadio[family.GetName()] = slots
	}

	return byRadio
}

// TestAPCollector_BandKeyedSeriesNameOnlyTheBandsTheyCanName pins both guards of the
// band-keyed rows. The fixture carries a row per named band plus the two shapes that are
// withheld: a row whose band identifier has no name, and a row reporting no channel.
func TestAPCollector_BandKeyedSeriesNameOnlyTheBandsTheyCanName(t *testing.T) {
	t.Parallel()

	byBand := gatherAPValuesByBand(t, fullFixtureSnapshot())

	for _, name := range []string{
		"wnc_rrm_worst_channel_air_quality_index_avg",
		"wnc_rrm_worst_channel_air_quality_index_min",
		"wnc_rrm_worst_channel_interferers",
		"wnc_rrm_worst_channel_number",
	} {
		bands := slices.Sorted(maps.Keys(byBand[name]))
		if want := []string{Band24GHz, Band5GHz}; !slices.Equal(bands, want) {
			t.Errorf("%s carries bands %v, want %v", name, bands, want)
		}
	}

	// The channel of the withheld row would have read as a channel, and its air quality
	// as the cleanest reading the scale has.
	if got, ok := byBand["wnc_rrm_worst_channel_number"][Band6GHz]; ok {
		t.Errorf("wnc_rrm_worst_channel_number = %f for the band reporting no channel, want it withheld", got)
	}
}

// TestAPCollector_BandKeyedSeriesAreEmittedOnceForTwoRadios pins that the band-keyed
// rows leave the per-radio loop once. Emitting them from inside it repeats one label set,
// which fails the whole scrape rather than the series.
func TestAPCollector_BandKeyedSeriesAreEmittedOnceForTwoRadios(t *testing.T) {
	t.Parallel()

	data := fullFixtureSnapshot()
	data.RadioOperData = append(data.RadioOperData, ap.RadioOperData{
		WtpMAC:      fixtureAPMAC,
		RadioSlotID: 1,
		RadioType:   "radio-80211a",
		OperState:   APRadioStateUp,
	})

	byBand := gatherAPValuesByBand(t, data)
	if got := len(byBand["wnc_rrm_worst_channel_number"]); got != 2 {
		t.Errorf("wnc_rrm_worst_channel_number carries %d series over two radios, want 2", got)
	}
}

// gatherAPValuesByBand indexes the spectrum module's families by the band label.
func gatherAPValuesByBand(t *testing.T, data *wnc.WNCDataCache) map[string]map[string]float64 {
	t.Helper()

	src := fixtureSource{data: data}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewAPCollector(
		wnc.NewAPSource(src), wnc.NewRRMSource(src), wnc.NewClientSource(src),
		APMetrics{Spectrum: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	values := make(map[string]map[string]float64, len(families))
	for _, family := range families {
		byBand := make(map[string]float64, len(family.GetMetric()))
		for _, metric := range family.GetMetric() {
			for _, pair := range metric.GetLabel() {
				if pair.GetName() == labelBand {
					byBand[pair.GetValue()] = metric.GetGauge().GetValue()
				}
			}
		}
		values[family.GetName()] = byBand
	}

	return values
}

// TestAirQualityOnCurrentChannel_RecordWithoutTheContainer covers the one branch the
// table above cannot reach: the per-radio container is a pointer, so a matching record
// can arrive without it. Dropping the nil test panics rather than misreporting, which is
// why no other case exercises it.
func TestAirQualityOnCurrentChannel_RecordWithoutTheContainer(t *testing.T) {
	t.Parallel()

	table := []rrm.SpectrumAqTable{{
		WtpMAC: fixtureAPMAC,
		Band:   "dot11-2-dot-4-ghz-band",
	}}
	radio := &ap.RadioOperData{
		WtpMAC:            fixtureAPMAC,
		CurrentActiveBand: "dot11-2-dot-4-ghz-band",
		PhyHtCfg:          &ap.PhyHtCfg{CfgData: ap.PhyHtCfgData{CurrFreq: 6}},
	}

	if got, found := airQualityOnCurrentChannel(table, radio); found {
		t.Errorf("airQualityOnCurrentChannel() = (%v, true), want no reading: the "+
			"matching record carries no per-radio container", got)
	}
}
