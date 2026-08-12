package collector

import "testing"

// gatherFixtureValues indexes the first sample of every gauge and counter family
// the all-succeed snapshot publishes, by family name.
//
// The registry route is what makes a descriptor's wiring observable: a helper
// tested on its own still passes when its result is handed to another family's
// descriptor, and TestAPCollector_MetricNames pins only a subset of the name to
// field pairs.
func gatherFixtureValues(t *testing.T) map[string]float64 {
	t.Helper()

	families, err := fixtureRegistry(t, "").Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	values := make(map[string]float64, len(families))
	for _, family := range families {
		metrics := family.GetMetric()
		if len(metrics) == 0 {
			continue
		}
		switch {
		case metrics[0].GetGauge() != nil:
			values[family.GetName()] = metrics[0].GetGauge().GetValue()
		case metrics[0].GetCounter() != nil:
			values[family.GetName()] = metrics[0].GetCounter().GetValue()
		}
	}
	return values
}

// TestAllCollectors_SystemAndRadioValuesMatchLeaves pins the system and RF gauges
// to their own leaves. Each pair below is emitted from adjacent lines that differ
// by one identifier, so a swap keeps every count and every label intact.
func TestAllCollectors_SystemAndRadioValuesMatchLeaves(t *testing.T) {
	t.Parallel()

	values := gatherFixtureValues(t)

	tests := []struct {
		name string
		want float64
	}{
		// cpu-usage and memory-usage are whole numbers that the collector divides
		// by one hundred, so the expectation also pins the scale.
		{"wnc_ap_cpu_utilization_ratio", 0.12},
		{"wnc_ap_memory_utilization_ratio", 0.34},

		{"wnc_ap_tx_power_dbm", 14},
		{"wnc_ap_tx_power_max_dbm", 20},

		{"wnc_client_rssi_dbm", -55},
		{"wnc_client_snr_decibels", 35},
	}

	assertValues(t, values, tests)
}

// TestAllCollectors_GaugeValuesMatchLeaves pins the remaining non-boolean gauges.
// The uptime gauges are left out on purpose: they are derived from time.Since and
// carry a different value on every run.
func TestAllCollectors_GaugeValuesMatchLeaves(t *testing.T) {
	t.Parallel()

	values := gatherFixtureValues(t)

	tests := []struct {
		name string
		want float64
	}{
		{"wnc_ap_channel_number", fixtureChannel},
		{"wnc_ap_channel_width_mhz", 40},

		// The four load leaves are whole numbers the collector divides by one
		// hundred, and each is distinct so that reading a sibling changes the ratio.
		{"wnc_ap_channel_utilization_ratio", 0.30},
		{"wnc_ap_rx_utilization_ratio", 0.10},
		{"wnc_ap_tx_utilization_ratio", 0.05},
		{"wnc_ap_noise_utilization_ratio", 0.22},

		// The noise list carries an entry for another channel first, so matching
		// the operating channel rather than an index is part of the contract.
		{"wnc_ap_noise_floor_dbm", -90},

		{"wnc_ap_clients", 1},
		{"wnc_ap_coverage_failed_clients", 7},
		// Both reset entries of this radio are totalled, not overwritten.
		{"wnc_ap_radio_reset_total", 8},
		// Seconds, so a switch to UnixMilli changes the value rather than the type.
		{"wnc_ap_last_radar_timestamp_seconds", 1767225600},

		{"wnc_wlan_clients", 1},
		{"wnc_wlan_session_timeout_seconds", 1800},

		{"wnc_client_speed_mbps", 866},
		{"wnc_client_spatial_streams", 2},
		{"wnc_client_mcs_index", 9},
		{"wnc_client_power_save_state", 1},
		// Milliseconds in the leaf, seconds in the series.
		{"wnc_client_state_transition_seconds", 0.12},
	}

	assertValues(t, values, tests)
}

func assertValues(t *testing.T, values map[string]float64, tests []struct {
	name string
	want float64
},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := values[tt.name]
			if !ok {
				t.Errorf("%s has no series in the all-succeed snapshot", tt.name)
				return
			}
			if got != tt.want {
				t.Errorf("%s = %v, want %v from its own leaf", tt.name, got, tt.want)
			}
		})
	}
}
