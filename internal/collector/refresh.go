// Package collector provides collectors for cisco-wnc-exporter.
package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// RefreshCollector reports the health of the WNC data refresh itself. Without it
// a failed refresh yields a successful scrape carrying no series, which no alert
// can detect.
type RefreshCollector struct {
	stats wnc.StatsProvider

	upDesc               *prometheus.Desc
	durationDesc         *prometheus.Desc
	timestampDesc        *prometheus.Desc
	errorsDesc           *prometheus.Desc
	itemsDesc            *prometheus.Desc
	defaultsFallbackDesc *prometheus.Desc
}

// NewRefreshCollector creates a collector reporting WNC data refresh health.
func NewRefreshCollector(stats wnc.StatsProvider) *RefreshCollector {
	dataLabels := []string{labelData}

	return &RefreshCollector{
		stats: stats,
		upDesc: prometheus.NewDesc(
			"wnc_up",
			"Whether the last completed refresh reached the controller. A partly failed refresh still reads 1",
			nil, nil,
		),
		durationDesc: prometheus.NewDesc(
			"wnc_refresh_duration_seconds",
			"Duration of the last WNC data refresh attempt",
			nil, nil,
		),
		timestampDesc: prometheus.NewDesc(
			"wnc_refresh_success_timestamp_seconds",
			"Start time of the refresh that produced the served snapshot. Absent until a refresh publishes one",
			nil, nil,
		),
		errorsDesc: prometheus.NewDesc(
			"wnc_refresh_errors_total",
			"WNC data fetch failures per data type since process start. Only requested data types have a series",
			dataLabels, nil,
		),
		itemsDesc: prometheus.NewDesc(
			"wnc_refresh_items",
			"Items returned per data type by the last WNC data refresh. Absent where that fetch failed",
			dataLabels, nil,
		),
		defaultsFallbackDesc: prometheus.NewDesc(
			"wnc_refresh_defaults_fallback_total",
			"WLAN config fetches that fell back from the values in force to a plain read since process start",
			nil, nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *RefreshCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upDesc
	ch <- c.durationDesc
	ch <- c.timestampDesc
	ch <- c.errorsDesc
	ch <- c.itemsDesc
	ch <- c.defaultsFallbackDesc
}

// Collect implements prometheus.Collector by reporting the refresh outcome.
// It reads statistics only: the refresh is triggered by the data collectors, and
// under stale-while-revalidate the statistics always describe the last completed
// refresh regardless of the order the registry drives collectors in.
func (c *RefreshCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.stats.Stats()

	ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, boolToFloat64(stats.Up))

	// Publishing a zero duration before the first attempt completes would assert
	// that a refresh finished instantly.
	if stats.Attempted {
		ch <- prometheus.MustNewConstMetric(
			c.durationDesc, prometheus.GaugeValue, stats.Duration.Seconds())
	}
	if !stats.RefreshedAt.IsZero() {
		ch <- prometheus.MustNewConstMetric(
			c.timestampDesc, prometheus.GaugeValue,
			float64(stats.RefreshedAt.UnixNano())/float64(time.Second),
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.defaultsFallbackDesc, prometheus.CounterValue, float64(stats.DefaultsFallbacks))

	for name, count := range stats.Errors {
		ch <- prometheus.MustNewConstMetric(c.errorsDesc, prometheus.CounterValue, float64(count), name)
	}
	for name, count := range stats.Items {
		ch <- prometheus.MustNewConstMetric(c.itemsDesc, prometheus.GaugeValue, float64(count), name)
	}
}
