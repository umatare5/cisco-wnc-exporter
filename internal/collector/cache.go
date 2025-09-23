// Package collector provides enhanced collectors with info metrics caching.
package collector

import (
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

// CacheCollector wraps collectors to cache metrics for improved performance.
type CacheCollector struct {
	base  prometheus.Collector
	cache *MetricsCache
	name  string
}

// NewCacheCollector creates a new collector with metrics caching.
func NewCacheCollector(base prometheus.Collector, name string, cacheTTL time.Duration) *CacheCollector {
	return &CacheCollector{
		base:  base,
		cache: cache.New[[]prometheus.Metric](cacheTTL, name+" metrics cache"),
		name:  name,
	}
}

// Describe implements prometheus.Collector interface by delegating to base collector.
func (c *CacheCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch)
}

// Collect implements prometheus.Collector interface with full metrics caching.
// All metrics from the base collector are cached together as a single unit.
func (c *CacheCollector) Collect(ch chan<- prometheus.Metric) {
	metrics, err := c.cache.Get(func() ([]prometheus.Metric, error) {
		// Collect all metrics from base collector
		baseCh := make(chan prometheus.Metric, MetricChannelBuffer)
		go func() {
			defer close(baseCh)
			c.base.Collect(baseCh)
		}()

		// Collect all metrics into slice
		var metrics []prometheus.Metric
		for metric := range baseCh {
			metrics = append(metrics, metric)
		}

		return metrics, nil
	})
	if err != nil {
		return // On cache error, do not emit metrics
	}

	// Serve cached metrics
	for _, metric := range metrics {
		ch <- metric
	}
}
