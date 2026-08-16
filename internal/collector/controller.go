// Package collector provides collectors for cisco-wnc-exporter.
// This file holds the controller-wide collector.
package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// ControllerMetrics represents which controller metrics are enabled.
type ControllerMetrics struct {
	General bool
}

// ControllerCollector implements prometheus.Collector for controller-wide metrics.
// Its series describe the controller itself rather than an AP, a client or a WLAN, so
// none of them carries an identifying label.
type ControllerCollector struct {
	metrics ControllerMetrics
	src     wnc.ControllerSource

	bootTimeDesc      *prometheus.Desc
	clientDeletesDesc *prometheus.Desc
}

// NewControllerCollector creates a new controller collector.
func NewControllerCollector(src wnc.ControllerSource, metrics ControllerMetrics) *ControllerCollector {
	collector := &ControllerCollector{
		src:     src,
		metrics: metrics,
	}

	if metrics.General {
		collector.bootTimeDesc = prometheus.NewDesc(
			"wnc_controller_boot_time_seconds",
			"Unix time the controller last booted. Withheld rather than reported as 0 when "+
				"the controller does not carry the leaf, so a counter reset check has no epoch "+
				"instead of a false one. The leaf moves by a second between reads, so compare it "+
				"against a threshold rather than with changes() or an equality",
			nil, nil,
		)
		collector.clientDeletesDesc = prometheus.NewDesc(
			"wnc_controller_client_deletes_total",
			"Client deletions the controller counted for this reason, one series per reason "+
				"leaf it reports, spelled as the controller spells it. The epoch is the "+
				"controller boot reported by wnc_controller_boot_time_seconds",
			[]string{labelReason}, nil,
		)
	}

	return collector
}

// Describe implements prometheus.Collector.
func (c *ControllerCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.General {
		ch <- c.bootTimeDesc
		ch <- c.clientDeletesDesc
	}
}

// Collect implements prometheus.Collector by retrieving controller data from WNC.
func (c *ControllerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if !c.metrics.General {
		return
	}

	c.collectBootTime(ctx, ch)
	c.collectClientDeletes(ctx, ch)
}

// collectBootTime publishes the boot instant, and publishes nothing when the leaf is
// absent, unparsable or the epoch. A zero there would report a boot in 1970, which
// makes an uptime derived from it five decades long.
func (c *ControllerCollector) collectBootTime(ctx context.Context, ch chan<- prometheus.Metric) {
	bootTime, err := c.src.GetBootTime(ctx)
	if err != nil {
		slog.Debug("Failed to get controller boot time", "error", err)
		return
	}

	at, err := time.Parse(time.RFC3339, bootTime)
	if err != nil {
		slog.Debug("Controller reported no usable boot time")
		return
	}

	emitTimestamp(ch, c.bootTimeDesc, at)
}

// collectClientDeletes publishes one counter per reason leaf, including the leaves
// reading zero: withholding those would make a leaf's first non-zero read look like a
// counter that had just been created, which Prometheus cannot tell from a reset.
func (c *ControllerCollector) collectClientDeletes(ctx context.Context, ch chan<- prometheus.Metric) {
	reasons, err := c.src.GetClientDeleteReasons(ctx)
	if err != nil {
		slog.Debug("Failed to get controller client delete reasons", "error", err)
		return
	}

	for reason, value := range reasons {
		ch <- prometheus.MustNewConstMetric(c.clientDeletesDesc, prometheus.CounterValue, value, reason)
	}
}
