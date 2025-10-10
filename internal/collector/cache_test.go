package collector

import (
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
		"test_info_metric",
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

	// Test case sensitivity
	desc3 := prometheus.NewDesc("test_INFO_metric", "Test INFO metric", nil, nil)
	metric3 := prometheus.MustNewConstMetric(desc3, prometheus.GaugeValue, 1.0)

	result := isInfoMetric(metric3)
	// The function uses strings.Contains which is case-sensitive
	// So "INFO" should not match "_info"
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
