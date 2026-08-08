package collector

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// panickingCollector emits one metric and then panics, mimicking a collector that
// dereferences an absent WNC field halfway through Collect.
type panickingCollector struct {
	desc *prometheus.Desc
}

func (c panickingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c panickingCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, 1)
	panic("collector failure")
}

func TestSafeCollector_Describe(t *testing.T) {
	t.Parallel()

	desc := prometheus.NewDesc("test_metric", "test", nil, nil)
	safe := NewSafeCollector(panickingCollector{desc: desc}, "Test")

	ch := make(chan *prometheus.Desc, 10)
	go func() {
		defer close(ch)
		safe.Describe(ch)
	}()

	descCount := 0
	for range ch {
		descCount++
	}

	if descCount != 1 {
		t.Errorf("SafeCollector.Describe() emitted %d descriptors, want 1", descCount)
	}
}

func TestSafeCollector_Collect(t *testing.T) {
	t.Parallel()

	desc := prometheus.NewDesc("test_metric", "test", nil, nil)
	safe := NewSafeCollector(panickingCollector{desc: desc}, "Test")

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		safe.Collect(ch)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount != 1 {
		t.Errorf("SafeCollector.Collect() emitted %d metrics, want 1", metricCount)
	}
}

// TestSafeCollector_WithInfoCache covers the registration order used by registerAPCollector,
// where the wrapped collector runs inside the goroutine InfoCacheCollector spawns.
func TestSafeCollector_WithInfoCache(t *testing.T) {
	t.Parallel()

	desc := prometheus.NewDesc("test_metric", "test", nil, nil)
	safe := NewSafeCollector(panickingCollector{desc: desc}, "Test")
	cached := NewInfoCacheCollector(safe, "Test", time.Minute)

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		cached.Collect(ch)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount != 1 {
		t.Errorf("InfoCacheCollector.Collect() emitted %d metrics, want 1", metricCount)
	}
}
