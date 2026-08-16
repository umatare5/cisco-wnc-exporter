// Package collector provides enhanced collectors with info metrics caching.
package collector

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-wnc-exporter/internal/cache"
)

const (
	// MetricChannelBuffer defines the buffer size for metric channels.
	MetricChannelBuffer = 100
)

// MetricsCache provides caching for specific metrics to reduce cardinality explosion from frequent label changes.
type MetricsCache = cache.Cache[[]prometheus.Metric]

// InfoCacheCollector wraps collectors to cache only info metrics for improved performance.
type InfoCacheCollector struct {
	base      prometheus.Collector
	infoCache *MetricsCache
	name      string
}

// NewInfoCacheCollector creates a new collector with info-only metrics caching.
func NewInfoCacheCollector(base prometheus.Collector, name string, cacheTTL time.Duration) *InfoCacheCollector {
	return &InfoCacheCollector{
		base:      base,
		infoCache: cache.New[[]prometheus.Metric](cacheTTL, name+" info metrics cache"),
		name:      name,
	}
}

// Describe implements prometheus.Collector interface by delegating to base collector.
func (c *InfoCacheCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch)
}

// Collect implements prometheus.Collector interface with info-only metrics caching.
// Info metrics are cached, all other metrics are served in real-time.
func (c *InfoCacheCollector) Collect(ch chan<- prometheus.Metric) {
	// Collect all metrics from base collector first
	baseCh := make(chan prometheus.Metric, MetricChannelBuffer)
	go func() {
		defer close(baseCh)
		c.base.Collect(baseCh)
	}()

	// Separate info and non-info metrics
	var nonInfoMetrics []prometheus.Metric
	var infoMetrics []prometheus.Metric

	for metric := range baseCh {
		if isInfoMetric(metric) {
			infoMetrics = append(infoMetrics, metric)
		} else {
			nonInfoMetrics = append(nonInfoMetrics, metric)
		}
	}

	// Serve non-info metrics immediately (real-time)
	for _, metric := range nonInfoMetrics {
		ch <- metric
	}

	// Cache and serve info metrics
	if len(infoMetrics) > 0 {
		cachedInfoMetrics, err := c.infoCache.Get(func() ([]prometheus.Metric, error) {
			return infoMetrics, nil
		})
		if err == nil {
			for _, metric := range cachedInfoMetrics {
				ch <- metric
			}
		}
	}
}

// isInfoMetric determines if a metric belongs to an _info family, the only class this
// cache holds. Desc exposes no accessor for the name, so the name is read back out of
// Desc.String, whose single return path renders it first and quoted; a descriptor
// carrying an error renders an empty name there rather than another shape.
func isInfoMetric(metric prometheus.Metric) bool {
	_, quoted, _ := strings.Cut(metric.Desc().String(), `fqName: "`)
	fqName, _, _ := strings.Cut(quoted, `"`)

	return strings.HasSuffix(fqName, "_info")
}
