package collector

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// stubStatsProvider returns a fixed RefreshStats.
type stubStatsProvider struct {
	stats wnc.RefreshStats
}

func (s stubStatsProvider) Stats() wnc.RefreshStats { return s.stats }

// refreshSample is one gathered series reduced to what these assertions need.
type refreshSample struct {
	labels map[string]string
	value  float64
}

// gatherRefresh registers the collector and indexes the samples by metric name.
// A metric with no series does not appear at all, so a missing key means absent.
func gatherRefresh(t *testing.T, stats wnc.RefreshStats) map[string][]refreshSample {
	t.Helper()

	registry := prometheus.NewRegistry()
	registry.MustRegister(NewRefreshCollector(stubStatsProvider{stats: stats}))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	byName := make(map[string][]refreshSample, len(families))
	for _, family := range families {
		samples := make([]refreshSample, 0, len(family.GetMetric()))
		for _, metric := range family.GetMetric() {
			value := metric.GetGauge().GetValue()
			if metric.Counter != nil {
				value = metric.GetCounter().GetValue()
			}

			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			samples = append(samples, refreshSample{labels: labels, value: value})
		}
		byName[family.GetName()] = samples
	}
	return byName
}

func TestRefreshCollector_Describe(t *testing.T) {
	t.Parallel()

	ch := make(chan *prometheus.Desc, 10)
	NewRefreshCollector(stubStatsProvider{}).Describe(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	if count != 5 {
		t.Errorf("Describe() sent %d descriptors, want 5", count)
	}
}

// TestRefreshCollector_ColdStart pins what the first scrape has to look like: the
// refresh has not completed, yet wnc_up and every error series must already exist
// so an alert can fire on them.
func TestRefreshCollector_ColdStart(t *testing.T) {
	t.Parallel()

	samples := gatherRefresh(t, wnc.RefreshStats{
		Errors: map[string]int{"ap_capwap_data": 0, "client_traffic_stats": 0},
	})

	up, ok := samples["wnc_up"]
	if !ok {
		t.Fatal("wnc_up is absent, want it present so a failing refresh is detectable")
	}
	if up[0].value != 0 {
		t.Errorf("wnc_up = %v, want 0", up[0].value)
	}

	if _, ok := samples["wnc_refresh_duration_seconds"]; ok {
		t.Error("wnc_refresh_duration_seconds is present before the first attempt completed, " +
			"want it absent rather than asserting an instant refresh")
	}
	if _, ok := samples["wnc_refresh_success_timestamp_seconds"]; ok {
		t.Error("wnc_refresh_success_timestamp_seconds is present with no snapshot, want it absent")
	}

	errorSeries, ok := samples["wnc_refresh_errors_total"]
	if !ok {
		t.Fatal("wnc_refresh_errors_total is absent on the first scrape, want the seeded series")
	}
	if len(errorSeries) != 2 {
		t.Errorf("wnc_refresh_errors_total has %d series, want 2", len(errorSeries))
	}

	if _, ok := samples["wnc_refresh_items"]; ok {
		t.Error("wnc_refresh_items is present with nothing fetched, want it absent")
	}
}

func TestRefreshCollector_AfterSuccessfulRefresh(t *testing.T) {
	t.Parallel()

	refreshedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	samples := gatherRefresh(t, wnc.RefreshStats{
		Up:          true,
		Attempted:   true,
		RefreshedAt: refreshedAt,
		Duration:    1500 * time.Millisecond,
		Errors:      map[string]int{"ap_capwap_data": 0},
		Items:       map[string]int{"ap_capwap_data": 4},
	})

	if got := samples["wnc_up"][0].value; got != 1 {
		t.Errorf("wnc_up = %v, want 1", got)
	}

	duration, ok := samples["wnc_refresh_duration_seconds"]
	if !ok {
		t.Fatal("wnc_refresh_duration_seconds is absent after a completed refresh")
	}
	if duration[0].value != 1.5 {
		t.Errorf("wnc_refresh_duration_seconds = %v, want 1.5", duration[0].value)
	}

	timestamp, ok := samples["wnc_refresh_success_timestamp_seconds"]
	if !ok {
		t.Fatal("wnc_refresh_success_timestamp_seconds is absent after a completed refresh")
	}
	want := float64(refreshedAt.UnixNano()) / float64(time.Second)
	if timestamp[0].value != want {
		t.Errorf("wnc_refresh_success_timestamp_seconds = %v, want %v", timestamp[0].value, want)
	}

	items, ok := samples["wnc_refresh_items"]
	if !ok {
		t.Fatal("wnc_refresh_items is absent after a successful fetch")
	}
	if items[0].value != 4 {
		t.Errorf("wnc_refresh_items = %v, want 4", items[0].value)
	}
	if got := items[0].labels[labelData]; got != "ap_capwap_data" {
		t.Errorf("wnc_refresh_items{%s} = %q, want ap_capwap_data", labelData, got)
	}
}

// TestRefreshCollector_ItemsAbsentForFailedDataType keeps the refresh health
// family from reproducing the fabricated zero this design removes elsewhere.
func TestRefreshCollector_ItemsAbsentForFailedDataType(t *testing.T) {
	t.Parallel()

	samples := gatherRefresh(t, wnc.RefreshStats{
		Up:        true,
		Attempted: true,
		Errors:    map[string]int{"ap_capwap_data": 0, "client_traffic_stats": 2},
		Items:     map[string]int{"ap_capwap_data": 4},
	})

	items := samples["wnc_refresh_items"]
	if len(items) != 1 {
		t.Fatalf("wnc_refresh_items has %d series, want 1: a failed data type must not report zero",
			len(items))
	}
	if got := items[0].labels[labelData]; got != "ap_capwap_data" {
		t.Errorf("wnc_refresh_items series is for %q, want ap_capwap_data", got)
	}

	for _, sample := range samples["wnc_refresh_errors_total"] {
		if sample.labels[labelData] != "client_traffic_stats" {
			continue
		}
		if sample.value != 2 {
			t.Errorf("wnc_refresh_errors_total{data=client_traffic_stats} = %v, want 2", sample.value)
		}
	}
}

// TestRefreshCollector_TotalFailureKeepsFreshnessObservable covers the state the
// grace period serves stale data in: wnc_up drops to zero while the freshness
// series stays, which is what the staleness alert fires on.
func TestRefreshCollector_TotalFailureKeepsFreshnessObservable(t *testing.T) {
	t.Parallel()

	samples := gatherRefresh(t, wnc.RefreshStats{
		Up:          false,
		Attempted:   true,
		RefreshedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Duration:    2 * time.Second,
		Errors:      map[string]int{"ap_capwap_data": 3},
	})

	if got := samples["wnc_up"][0].value; got != 0 {
		t.Errorf("wnc_up = %v, want 0", got)
	}
	if _, ok := samples["wnc_refresh_success_timestamp_seconds"]; !ok {
		t.Error("wnc_refresh_success_timestamp_seconds is absent while a stale snapshot is served, " +
			"want it present so staleness stays alertable")
	}
	if _, ok := samples["wnc_refresh_duration_seconds"]; !ok {
		t.Error("wnc_refresh_duration_seconds is absent after a failed attempt, want the attempt duration")
	}
}
