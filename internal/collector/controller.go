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

// The roam counters the controller keys by leaf name. Only the three the controller
// actually maintains are published; the other ten leaves of the container read zero
// while these move, and two of them duplicate a WLAN configuration series.
const (
	leafAPAuthRoams          = "ap-auth-roams"
	leafAPAuthDot11iFastRoam = "ap-auth-dot11i-fast-roam"
	leafAPAuthDot11iSlowRoam = "ap-auth-dot11i-slow-roam"
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

	apAuthRoamsDesc     *prometheus.Desc
	apAuthFastRoamsDesc *prometheus.Desc
	apAuthSlowRoamsDesc *prometheus.Desc
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
		collector.apAuthRoamsDesc = prometheus.NewDesc(
			"wnc_controller_client_ap_auth_roams_total",
			"Roams the controller counted on the path where the AP authenticated the client "+
				"itself, cumulative since the controller booted. A WLAN whose policy profile "+
				"leaves central association enabled is not counted here however much it roams, "+
				"and the two dot11i counters are not a partition of this one",
			nil, nil,
		)
		collector.apAuthFastRoamsDesc = prometheus.NewDesc(
			"wnc_controller_client_ap_auth_dot11i_fast_roams_total",
			"802.11i fast roams on that same path, cumulative since the controller booted. "+
				"Zero here while the slow counter moves means a cached key is not being used",
			nil, nil,
		)
		collector.apAuthSlowRoamsDesc = prometheus.NewDesc(
			"wnc_controller_client_ap_auth_dot11i_slow_roams_total",
			"802.11i slow roams on that same path, cumulative since the controller booted, "+
				"each one a full authentication rather than a cached key",
			nil, nil,
		)
	}

	return collector
}

// Describe implements prometheus.Collector.
func (c *ControllerCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.metrics.General {
		ch <- c.bootTimeDesc
		ch <- c.clientDeletesDesc
		ch <- c.apAuthRoamsDesc
		ch <- c.apAuthFastRoamsDesc
		ch <- c.apAuthSlowRoamsDesc
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
	c.collectRoams(ctx, ch)
}

// collectRoams publishes the three roam counters the controller maintains, each only
// when the container carries its leaf: a leaf a release stops sending would otherwise
// read as a counter that had fallen to zero.
func (c *ControllerCollector) collectRoams(ctx context.Context, ch chan<- prometheus.Metric) {
	roams, err := c.src.GetClientRoamingStats(ctx)
	if err != nil {
		slog.Debug("Failed to get controller roam statistics", "error", err)
		return
	}

	for _, counter := range []struct {
		desc *prometheus.Desc
		leaf string
	}{
		{c.apAuthRoamsDesc, leafAPAuthRoams},
		{c.apAuthFastRoamsDesc, leafAPAuthDot11iFastRoam},
		{c.apAuthSlowRoamsDesc, leafAPAuthDot11iSlowRoam},
	} {
		value, ok := roams[counter.leaf]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(counter.desc, prometheus.CounterValue, value)
	}
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
