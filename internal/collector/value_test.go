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
