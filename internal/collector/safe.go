// Package collector provides panic recovery for collectors.
package collector

import (
	"log/slog"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
)

// SafeCollector wraps a collector so that a panic during Collect does not terminate the process.
// The registry drives Collect from a goroutine it owns, where an unrecovered panic takes down the
// whole exporter instead of the failing scrape alone.
type SafeCollector struct {
	base prometheus.Collector
	name string
}

// NewSafeCollector creates a collector that recovers panics raised by base.
func NewSafeCollector(base prometheus.Collector, name string) *SafeCollector {
	return &SafeCollector{
		base: base,
		name: name,
	}
}

// Describe implements prometheus.Collector interface by delegating to base collector.
func (c *SafeCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch)
}

// Collect implements prometheus.Collector interface and recovers panics from the base collector.
// Metrics emitted before the panic are retained, so the scrape returns partial data.
func (c *SafeCollector) Collect(ch chan<- prometheus.Metric) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error(
				"Collector panicked, metrics are incomplete",
				"collector", c.name,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()

	c.base.Collect(ch)
}
