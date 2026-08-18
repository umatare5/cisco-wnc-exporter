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
		// The operating channel's row, not the padding row and not the neighboring
		// channel the fixture table also carries.
		{"wnc_ap_air_quality_index_avg", 93},
		{"wnc_ap_air_quality_index_min", 92},
		// The interferer count of that same row, not of the padding or neighboring one.
		{"wnc_ap_interferers", 41},
		// The band-keyed rows are sorted by label value, so these read the 2.4 GHz row.
		// Each leaf of that row carries a distinct number, so a descriptor reading its
		// neighbor reports a value these pins do not expect.
		{"wnc_rrm_worst_channel_air_quality_index_min", 8101},
		{"wnc_rrm_worst_channel_air_quality_index_avg", 8102},
		{"wnc_rrm_worst_channel_interferers", 8103},
		{"wnc_rrm_worst_channel_number", 11},
		// Both reset entries of this radio are totalled, not overwritten.
		{"wnc_ap_radio_resets_total", 8},
		// The change count, not one of the three energy or channel leaves beside it.
		{"wnc_ap_channel_changes_total", 34},
		// Seconds, so a switch to UnixMilli changes the value rather than the type.
		{"wnc_ap_last_radar_timestamp_seconds", 1767225600},

		{"wnc_wlan_clients", 1},
		{"wnc_wlan_session_timeout_seconds", 1800},
		{"wnc_wlan_policy_binding", 1},

		{"wnc_client_speed_mbps", 866},
		{"wnc_client_spatial_streams", 2},
		{"wnc_client_mcs_index", 9},
		{"wnc_client_power_save_state", 1},
		// Milliseconds in the leaf, seconds in the series.
		{"wnc_client_state_transition_seconds", 0.12},

		// The join module. Both value-1 gauges first, then the timestamps, which are
		// one day apart in the fixture so that reading a neighboring leaf lands on
		// another day rather than on a plausible instant.
		{"wnc_ap_joined", 1},
		{"wnc_ap_join_info", 1},
		{"wnc_ap_last_error_timestamp_seconds", 1767312000},
		{"wnc_ap_last_join_success_timestamp_seconds", 1767398400},
		{"wnc_ap_last_join_failure_timestamp_seconds", 1767484800},
		{"wnc_ap_last_config_success_timestamp_seconds", 1767571200},
		{"wnc_ap_last_config_failure_timestamp_seconds", 1767657600},
		{"wnc_ap_last_discovery_success_timestamp_seconds", 1767744000},
		{"wnc_ap_last_discovery_failure_timestamp_seconds", 1767830400},

		// The controller module. Seconds, so a switch to UnixMilli changes the value.
		{"wnc_controller_boot_time_seconds", 1768262400},
	}

	assertValues(t, values, tests)
}

// TestAllCollectors_CounterValuesMatchLeaves pins every counter read from
// radio-oper-stats and traffic-stats to its own leaf. A counter wired to a
// neighboring leaf reports a plausible rate forever, which no count and no label
// assertion can see.
func TestAllCollectors_CounterValuesMatchLeaves(t *testing.T) {
	t.Parallel()

	values := gatherFixtureValues(t)

	tests := []struct {
		name string
		want float64
	}{
		{"wnc_ap_data_rx_frames_total", 3101},
		{"wnc_ap_data_tx_frames_total", 3102},
		{"wnc_ap_management_rx_frames_total", 3103},
		{"wnc_ap_management_tx_frames_total", 3104},
		{"wnc_ap_control_rx_frames_total", 3105},
		{"wnc_ap_control_tx_frames_total", 3106},
		{"wnc_ap_multicast_rx_frames_total", 3107},
		{"wnc_ap_multicast_tx_frames_total", 3108},
		// tx-frame-count, the controller's own total, which the HELP string
		// promises is not the sum of the per-type series above.
		{"wnc_ap_total_tx_frames_total", 3109},
		{"wnc_ap_rts_successes_total", 3110},

		{"wnc_ap_rx_errors_total", 3201},
		{"wnc_ap_tx_retries_total", 3202},
		// failed-count, not ack-failure-count: docs/collector.ap.md note *5 records
		// the Cisco bug that makes the latter unusable, so a one-token revert to it
		// would leave every count and label intact.
		{"wnc_ap_transmission_failures_total", 3203},
		{"wnc_ap_duplicate_frames_total", 3204},
		{"wnc_ap_fcs_errors_total", 3205},
		{"wnc_ap_rx_fragments_total", 3206},
		{"wnc_ap_tx_fragments_total", 3207},
		{"wnc_ap_rts_failures_total", 3208},
		{"wnc_ap_decryption_errors_total", 3209},
		{"wnc_ap_mic_errors_total", 3210},

		{"wnc_client_rx_bytes_total", 4101},
		{"wnc_client_tx_bytes_total", 4102},
		{"wnc_client_rx_packets_total", 4103},
		{"wnc_client_tx_packets_total", 4104},

		{"wnc_client_policy_errors_total", 4201},
		{"wnc_client_duplicate_received_total", 4202},
		{"wnc_client_decryption_failed_total", 4203},
		{"wnc_client_mic_mismatch_total", 4204},
		{"wnc_client_mic_missing_total", 4205},
		{"wnc_client_excessive_retries_total", 4206},
		{"wnc_client_rx_group_total", 4207},
		{"wnc_client_tx_drops_total", 4208},
		{"wnc_client_data_retries_total", 4209},
		{"wnc_client_rts_retries_total", 4210},
		{"wnc_client_tx_retries_total", 4211},

		// The join module. The discovery counters live in one container and the join
		// and configuration counters in another, and the request, response and failure
		// leaves of each phase are adjacent, which is what a swap would exchange.
		{"wnc_ap_discovery_requests_total", 5201},
		{"wnc_ap_discovery_responses_total", 5202},
		{"wnc_ap_discovery_errors_total", 5203},
		{"wnc_ap_join_requests_total", 5101},
		{"wnc_ap_join_responses_total", 5102},
		{"wnc_ap_join_failures_total", 5103},
		{"wnc_ap_config_requests_total", 5104},
		{"wnc_ap_config_responses_total", 5105},
		{"wnc_ap_config_failures_total", 5106},

		// The first reason in label order of the controller module, which carries one
		// series per reason leaf.
		{"wnc_controller_client_deletes_total", 6101},

		// The three roam counters. Their leaves are adjacent in one container and the ten
		// unpublished ones carry values too, so a swap lands on a number listed nowhere.
		{"wnc_controller_client_ap_auth_roams_total", 6201},
		{"wnc_controller_client_ap_auth_dot11i_fast_roams_total", 6202},
		{"wnc_controller_client_ap_auth_dot11i_slow_roams_total", 6203},

		// Bytes in both directions, from the leaf the controller reports as a string.
		// The record's other counts each carry their own value, so reading one of them
		// lands elsewhere. Four are the onboarding phases, pinned per phase in
		// TestWLANCollector_OnboardingPhasesMatchLeaves; the run count and the
		// random-MAC count remain unpublished.
		{"wnc_wlan_data_usage_bytes_total", 7101},
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
