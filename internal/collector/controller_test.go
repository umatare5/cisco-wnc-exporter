package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// gatherControllerValues collects the controller collector over the given snapshot and
// indexes every sample by metric name, then by the reason label.
func gatherControllerValues(t *testing.T, data *wnc.WNCDataCache) map[string]map[string]float64 {
	t.Helper()

	src := fixtureSource{data: data}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewControllerCollector(
		wnc.NewControllerSource(src), ControllerMetrics{General: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	values := make(map[string]map[string]float64, len(families))
	for _, family := range families {
		byReason := make(map[string]float64, len(family.GetMetric()))
		for _, metric := range family.GetMetric() {
			key := ""
			for _, pair := range metric.GetLabel() {
				if pair.GetName() == labelReason {
					key = pair.GetValue()
				}
			}
			switch {
			case metric.GetGauge() != nil:
				byReason[key] = metric.GetGauge().GetValue()
			case metric.GetCounter() != nil:
				byReason[key] = metric.GetCounter().GetValue()
			}
		}
		values[family.GetName()] = byReason
	}
	return values
}

// TestControllerCollector_Describe pins the descriptor count per module, so a series
// added to the collector without a module guard shows up here.
func TestControllerCollector_Describe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		metrics     ControllerMetrics
		expectDescs int
	}{
		{"No modules enabled", ControllerMetrics{}, 0},
		{"General module only", ControllerMetrics{General: true}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewControllerCollector(nil, tt.metrics)

			ch := make(chan *prometheus.Desc, 10)
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

// TestControllerCollector_DeleteReasonsAreOneSeriesPerLeaf pins the per-reason loop and
// the zero. A leaf reading zero is published rather than withheld: withholding it would
// make its first non-zero read look like a counter created at that moment, which
// Prometheus cannot tell from a reset.
func TestControllerCollector_DeleteReasonsAreOneSeriesPerLeaf(t *testing.T) {
	t.Parallel()

	data := fullFixtureSnapshot()
	data.ClientDeleteReasons["never-happened"] = 0

	values := gatherControllerValues(t, data)
	byReason := values["wnc_controller_client_deletes_total"]

	want := map[string]float64{
		fixtureDeleteReason:      6101,
		fixtureOtherDeleteReason: 6102,
		"never-happened":         0,
	}
	if len(byReason) != len(want) {
		t.Errorf("wnc_controller_client_deletes_total has %d series, want %d: %v",
			len(byReason), len(want), byReason)
	}
	for reason, wantValue := range want {
		if got, ok := byReason[reason]; !ok || got != wantValue {
			t.Errorf("wnc_controller_client_deletes_total{reason=%q} = %v (present %v), want %v",
				reason, got, ok, wantValue)
		}
	}
}

// TestControllerCollector_BootTimeWithheldWhenUnusable keeps a 1970 boot out of the
// series. It is the epoch every new counter is read against, so a fabricated zero there
// reports the controller as five decades up and makes a reset check silently wrong.
func TestControllerCollector_BootTimeWithheldWhenUnusable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		bootTime    string
		wantPresent bool
	}{
		{"absent leaf", "", false},
		{"unparsable leaf", "2026-01-13", false},
		{"epoch sentinel", "1970-01-01T00:00:00+00:00", false},
		{"usable leaf", fixtureBootTime, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.ControllerBootTime = tt.bootTime

			values := gatherControllerValues(t, data)
			if _, ok := values["wnc_controller_boot_time_seconds"]; ok != tt.wantPresent {
				t.Errorf("wnc_controller_boot_time_seconds present = %v for %q, want %v",
					ok, tt.bootTime, tt.wantPresent)
			}

			// The sibling family must survive, so the withhold is scoped to one series.
			if len(values["wnc_controller_client_deletes_total"]) == 0 {
				t.Error("wnc_controller_client_deletes_total is absent, so the assertion above proves nothing")
			}
		})
	}
}

// TestControllerCollector_SeriesCarryNoIdentifyingLabel pins the label sets. Both series
// describe the controller, and the delete reasons are one controller-wide object with no
// per-AP, per-WLAN or per-client key, so attributing one to a device would be a claim
// the data does not support.
func TestControllerCollector_SeriesCarryNoIdentifyingLabel(t *testing.T) {
	t.Parallel()

	src := fixtureSource{data: fullFixtureSnapshot()}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewControllerCollector(
		wnc.NewControllerSource(src), ControllerMetrics{General: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}
	if len(families) == 0 {
		t.Fatal("no family was gathered, so the label assertions below prove nothing")
	}

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			for _, pair := range metric.GetLabel() {
				if pair.GetName() != labelReason {
					t.Errorf("%s carries the label %q, want only %q",
						family.GetName(), pair.GetName(), labelReason)
				}
			}
		}
	}
}

// TestControllerCollector_RoamCounterWithheldWhenLeafAbsent pins the per-leaf guard.
// Ten of the container's thirteen leaves are deliberately unpublished, so a release
// that stopped sending one of the three that are would otherwise read as a counter
// that had fallen to zero.
func TestControllerCollector_RoamCounterWithheldWhenLeafAbsent(t *testing.T) {
	t.Parallel()

	data := fullFixtureSnapshot()
	delete(data.ClientRoamingStats, leafAPAuthDot11iFastRoam)

	values := gatherControllerValues(t, data)

	if len(values["wnc_controller_client_ap_auth_dot11i_fast_roams_total"]) != 0 {
		t.Error("wnc_controller_client_ap_auth_dot11i_fast_roams_total is present for an absent leaf")
	}

	// The siblings must survive, so the withhold is per leaf rather than per container.
	for _, name := range []string{
		"wnc_controller_client_ap_auth_roams_total",
		"wnc_controller_client_ap_auth_dot11i_slow_roams_total",
	} {
		if len(values[name]) == 0 {
			t.Errorf("%s is absent, so the assertion above proves nothing", name)
		}
	}
}
