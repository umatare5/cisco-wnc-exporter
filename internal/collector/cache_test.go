package collector

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// testCollector is a simple collector for testing purposes.
type testCollector struct {
	desc    *prometheus.Desc
	metrics []prometheus.Metric
}

func newTestCollector(metricName string) *testCollector {
	desc := prometheus.NewDesc(metricName, "Test metric", nil, nil)
	metric := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1.0)
	return &testCollector{
		desc:    desc,
		metrics: []prometheus.Metric{metric},
	}
}

func (c *testCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *testCollector) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func TestNewInfoCacheCollector(t *testing.T) {
	t.Parallel()
	base := newTestCollector("test_metric")
	cacheTTL := 5 * time.Second
	name := "test"

	collector := NewInfoCacheCollector(base, name, cacheTTL)

	if collector == nil {
		t.Fatal("NewInfoCacheCollector returned nil")
	}
	if collector.base != base {
		t.Error("base collector not set correctly")
	}
	if collector.name != name {
		t.Error("name not set correctly")
	}
	if collector.infoCache == nil {
		t.Error("infoCache not initialized")
	}
}

func TestInfoCacheCollector_Describe(t *testing.T) {
	t.Parallel()
	base := newTestCollector("test_metric")
	collector := NewInfoCacheCollector(base, "test", 5*time.Second)

	ch := make(chan *prometheus.Desc, 10)
	collector.Describe(ch)
	close(ch)

	// Should receive exactly one descriptor from the base collector
	count := 0
	for range ch {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 descriptor, got %d", count)
	}
}

func TestInfoCacheCollector_Collect_NonInfoMetrics(t *testing.T) {
	t.Parallel()
	// Test with non-info metric
	base := newTestCollector("test_metric")
	collector := NewInfoCacheCollector(base, "test", 5*time.Second)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should receive exactly one metric
	count := 0
	for range ch {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 metric, got %d", count)
	}
}

func TestInfoCacheCollector_Collect_InfoMetrics(t *testing.T) {
	t.Parallel()
	// Test with info metric
	base := newTestCollector("test_info")
	collector := NewInfoCacheCollector(base, "test", 5*time.Second)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should receive exactly one metric (cached info metric)
	count := 0
	for range ch {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 metric, got %d", count)
	}
}

func TestInfoCacheCollector_Collect_MixedMetrics(t *testing.T) {
	t.Parallel()
	// Create a collector with both info and non-info metrics
	collector := &testCollector{
		desc: prometheus.NewDesc("mixed_test", "Mixed test", nil, nil),
		metrics: []prometheus.Metric{
			prometheus.MustNewConstMetric(
				prometheus.NewDesc("test_info", "Info metric", nil, nil),
				prometheus.GaugeValue, 1.0,
			),
			prometheus.MustNewConstMetric(
				prometheus.NewDesc("test_count", "Count metric", nil, nil),
				prometheus.CounterValue, 42.0,
			),
		},
	}

	cacheCollector := NewInfoCacheCollector(collector, "test", 5*time.Second)

	ch := make(chan prometheus.Metric, 10)
	cacheCollector.Collect(ch)
	close(ch)

	// Should receive both metrics
	count := 0
	for range ch {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 metrics, got %d", count)
	}
}

func TestIsInfoMetric(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		metric   prometheus.Metric
		expected bool
	}{
		{
			name: "info metric",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_ap_info", "AP info", nil, nil),
				prometheus.GaugeValue,
				1.0,
			),
			expected: true,
		},
		{
			name: "client info metric",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_client_info", "Client info", nil, nil),
				prometheus.GaugeValue,
				1.0,
			),
			expected: true,
		},
		{
			name: "wlan info metric",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_wlan_info", "WLAN info", nil, nil),
				prometheus.GaugeValue,
				1.0,
			),
			expected: true,
		},
		{
			name: "non-info metric",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_ap_count", "AP count", nil, nil),
				prometheus.GaugeValue,
				1.0,
			),
			expected: false,
		},
		{
			name: "counter metric",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_bytes_total", "Total bytes", nil, nil),
				prometheus.CounterValue,
				1000.0,
			),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isInfoMetric(tt.metric)
			if result != tt.expected {
				t.Errorf("isInfoMetric() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestInfoCacheCollector_CollectMultipleTimes(t *testing.T) {
	t.Parallel()
	// Test that caching works correctly across multiple collects
	base := newTestCollector("test_info")
	collector := NewInfoCacheCollector(base, "test", 1*time.Hour) // Long TTL

	// First collect
	ch1 := make(chan prometheus.Metric, 10)
	collector.Collect(ch1)
	close(ch1)

	count1 := 0
	for range ch1 {
		count1++
	}

	// Second collect (should use cache)
	ch2 := make(chan prometheus.Metric, 10)
	collector.Collect(ch2)
	close(ch2)

	count2 := 0
	for range ch2 {
		count2++
	}

	if count1 != count2 {
		t.Errorf("cached collect returned different count: first=%d, second=%d", count1, count2)
	}
	if count1 != 1 {
		t.Errorf("expected 1 metric, got %d", count1)
	}
}

func TestMetricChannelBuffer(t *testing.T) {
	t.Parallel()
	// Test that the buffer constant is reasonable
	if MetricChannelBuffer <= 0 {
		t.Error("MetricChannelBuffer should be positive")
	}
	if MetricChannelBuffer < 10 {
		t.Error("MetricChannelBuffer should be large enough for typical use")
	}
}

func TestIsInfoMetric_DescString(t *testing.T) {
	t.Parallel()
	// Test that isInfoMetric correctly parses the descriptor string
	desc := prometheus.NewDesc(
		"test_metric_info",
		"Test info metric",
		[]string{"label1"},
		prometheus.Labels{"const": "value"},
	)
	metric := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1.0, "label_value")

	if !isInfoMetric(metric) {
		t.Error("expected metric with '_info' in name to be identified as info metric")
	}

	// Test edge case: info at the beginning (should NOT match since it's not "_info")
	desc2 := prometheus.NewDesc("info_test_metric", "Info test metric", nil, nil)
	metric2 := prometheus.MustNewConstMetric(desc2, prometheus.GaugeValue, 1.0)

	if isInfoMetric(metric2) {
		t.Error("expected metric with 'info_' at beginning to NOT be identified as info metric (requires '_info')")
	}

	// Test case sensitivity. The name ends in the upper-case form deliberately: one
	// merely containing it fails the suffix test whatever the case, so it would pass
	// against a predicate that folded case before comparing.
	desc3 := prometheus.NewDesc("test_metric_INFO", "Test INFO metric", nil, nil)
	metric3 := prometheus.MustNewConstMetric(desc3, prometheus.GaugeValue, 1.0)

	result := isInfoMetric(metric3)
	if result {
		t.Error("expected case-sensitive matching for info detection")
	}
}

func TestInfoCacheCollector_Collect_EmptyMetrics(t *testing.T) {
	t.Parallel()
	// Test behavior with a collector that returns no metrics
	emptyCollector := &testCollector{
		desc:    prometheus.NewDesc("empty_test", "Empty test", nil, nil),
		metrics: []prometheus.Metric{}, // No metrics
	}

	cacheCollector := NewInfoCacheCollector(emptyCollector, "empty", 5*time.Second)

	ch := make(chan prometheus.Metric, 10)
	cacheCollector.Collect(ch)
	close(ch)

	// Should receive no metrics
	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 metrics, got %d", count)
	}
}

func TestInfoCacheCollector_Collect_CacheError(t *testing.T) {
	t.Parallel()
	// Test behavior when cache returns an error (though current implementation doesn't generate errors)
	base := newTestCollector("test_info")
	collector := NewInfoCacheCollector(base, "test", 0) // Zero TTL to test cache behavior

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should still handle the collection (current implementation doesn't fail on cache errors)
	count := 0
	for range ch {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 metric even with zero TTL, got %d", count)
	}
}

// TestIsInfoMetric_KeysOnlyOnTheNameSuffix holds the predicate to the fully qualified
// name. Desc.String embeds the HELP text and the variable label names either side of
// the name, so a substring test reaches prose an author writes freely: a counter whose
// HELP merely names an _info family would land in a cache sized for label churn and
// then report one value for a whole TTL, which turns rate() into a staircase. The HELP
// text and the label names below are the hypothetical ones, not the published ones.
func TestIsInfoMetric_KeysOnlyOnTheNameSuffix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		metric   prometheus.Metric
		expected bool
	}{
		{
			name: "help names an info family",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"wnc_ap_uptime_seconds",
					"Seconds since the AP booted. Join wnc_ap_info for the AP name.",
					nil, nil,
				),
				prometheus.GaugeValue,
				1.0,
			),
			expected: false,
		},
		{
			name: "variable label name ends in _info",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_ap_channel_number", "Radio channel", []string{"radio_info"}, nil),
				prometheus.GaugeValue,
				36.0,
				"0",
			),
			expected: false,
		},
		{
			name: "name carries _info between two words",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_ap_info_labels_total", "Info labels published", nil, nil),
				prometheus.CounterValue,
				7.0,
			),
			expected: false,
		},
		{
			name: "name ends in _info",
			metric: prometheus.MustNewConstMetric(
				prometheus.NewDesc("wnc_ap_join_info", "AP join info", nil, nil),
				prometheus.GaugeValue,
				1.0,
			),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isInfoMetric(tt.metric); got != tt.expected {
				t.Errorf("isInfoMetric(%s) = %v, expected %v", metricFamilyName(t, tt.metric), got, tt.expected)
			}
		})
	}
}

// TestIsInfoMetric_MatchesOnlyTheInfoFamilies keeps every other family out of the info
// cache across the whole published surface, where the table above states the rule over
// four names. Nothing else in the suite would notice a family joining the cache.
func TestIsInfoMetric_MatchesOnlyTheInfoFamilies(t *testing.T) {
	t.Parallel()

	metrics := make(chan prometheus.Metric, MetricChannelBuffer*10)
	go func() {
		defer close(metrics)
		for _, collector := range fixtureCollectors(t, fullFixtureSnapshot()) {
			collector.Collect(metrics)
		}
	}()

	seen := 0
	for metric := range metrics {
		seen++
		name := metricFamilyName(t, metric)
		if got, want := isInfoMetric(metric), strings.HasSuffix(name, "_info"); got != want {
			t.Errorf("isInfoMetric(%s) = %v, want %v", name, got, want)
		}
	}
	if seen == 0 {
		t.Fatal("no metric was collected, so the classification above proves nothing")
	}
}

// metricFamilyName reads the fully qualified name back out of the descriptor, which
// exposes no accessor for it.
func metricFamilyName(t *testing.T, metric prometheus.Metric) string {
	t.Helper()

	_, rest, found := strings.Cut(metric.Desc().String(), `fqName: "`)
	if !found {
		t.Fatalf("descriptor %q carries no fqName", metric.Desc().String())
	}
	name, _, _ := strings.Cut(rest, `"`)
	return name
}
