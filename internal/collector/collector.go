// Package collector provides registry management and collector registration.
package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

// Collector manages Prometheus collectors and registry.
type Collector struct {
	registry         *prometheus.Registry
	cfg              *config.Config
	sharedDataSource wnc.DataSource
}

// Float64Metric represents a metric with float64 value.
type Float64Metric struct {
	Desc  *prometheus.Desc
	Value float64
}

// StringMetric represents a metric with string value that needs conversion.
type StringMetric struct {
	Desc  *prometheus.Desc
	Value string
}

// NewCollector creates a new collector manager.
func NewCollector(cfg *config.Config) *Collector {
	sharedDataSource := wnc.NewDataSource(cfg.WNC, cfg.Collectors)

	return &Collector{
		registry:         prometheus.NewRegistry(),
		cfg:              cfg,
		sharedDataSource: sharedDataSource,
	}
}

// Registry returns the Prometheus registry managed by this collector.
func (c *Collector) Registry() *prometheus.Registry {
	return c.registry
}

// Setup configures and registers all collectors based on configuration.
func (c *Collector) Setup(version string) {
	c.RegisterBuildInfo(version)
	c.RegisterSystemCollectors()
	c.RegisterServiceCollectors()
}

// RegisterBuildInfo registers the build information metric.
func (c *Collector) RegisterBuildInfo(version string) {
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wnc_build_info",
			Help: "Build information for the WNC exporter.",
		},
		[]string{"version"},
	)
	buildInfo.WithLabelValues(version).Set(1)
	c.registry.MustRegister(buildInfo)
}

// RegisterSystemCollectors registers Go and process collectors conditionally.
func (c *Collector) RegisterSystemCollectors() {
	if c.cfg.InternalCollector.EnableGoCollector {
		c.registry.MustRegister(collectors.NewGoCollector())
		slog.Debug("Registered Go collector")
	}
	if c.cfg.InternalCollector.EnableProcessCollector {
		c.registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		slog.Debug("Registered process collector")
	}
}

// RegisterServiceCollectors registers all service-specific collectors based on configuration.
func (c *Collector) RegisterServiceCollectors() {
	registered := false

	// Register AP collector if any AP module is enabled
	if IsEnabled(
		c.cfg.Collectors.AP.General,
		c.cfg.Collectors.AP.Radio,
		c.cfg.Collectors.AP.Traffic,
		c.cfg.Collectors.AP.Errors,
		c.cfg.Collectors.AP.Join,
		c.cfg.Collectors.AP.Info,
	) {
		apSource := wnc.NewAPSource(c.sharedDataSource)
		rrmSource := wnc.NewRRMSource(c.sharedDataSource)
		clientSource := wnc.NewClientSource(c.sharedDataSource)
		c.registerAPCollector(apSource, rrmSource, clientSource)
		registered = true
	} else {
		slog.Debug("Skipped AP collector registration - all modules disabled")
	}

	// Register WLAN collector if any WLAN module is enabled
	if IsEnabled(
		c.cfg.Collectors.WLAN.General,
		c.cfg.Collectors.WLAN.Traffic,
		c.cfg.Collectors.WLAN.Config,
		c.cfg.Collectors.WLAN.Info,
	) {
		wlanSource := wnc.NewWLANSource(c.sharedDataSource)
		clientSource := wnc.NewClientSource(c.sharedDataSource)
		c.registerWLANCollector(wlanSource, clientSource)
		registered = true
	} else {
		slog.Debug("Skipped WLAN collector registration - all modules disabled")
	}

	// Register Client collector if any Client module is enabled
	if IsEnabled(
		c.cfg.Collectors.Client.General,
		c.cfg.Collectors.Client.Radio,
		c.cfg.Collectors.Client.Traffic,
		c.cfg.Collectors.Client.Errors,
		c.cfg.Collectors.Client.Info,
	) {
		clientSource := wnc.NewClientSource(c.sharedDataSource)
		c.registerClientCollector(clientSource)
		registered = true
	} else {
		slog.Debug("Skipped Client collector registration - all modules disabled")
	}

	// Refresh health is only meaningful once something actually queries the WNC.
	if registered {
		c.registerRefreshCollector()
	}
}

// registerRefreshCollector registers the WNC data refresh health collector.
func (c *Collector) registerRefreshCollector() {
	stats, ok := c.sharedDataSource.(wnc.StatsProvider)
	if !ok {
		slog.Debug("Skipped refresh collector registration - data source reports no stats")
		return
	}

	c.registry.MustRegister(NewSafeCollector(NewRefreshCollector(stats), "Refresh"))
	slog.Debug("Registered refresh collector")
}

// registerAPCollector registers the AP collector with its modules.
func (c *Collector) registerAPCollector(apSource wnc.APSource, rrmSource wnc.RRMSource, clientSource wnc.ClientSource) {
	baseCollector := NewAPCollector(apSource, rrmSource, clientSource, APMetrics{
		General:    c.cfg.Collectors.AP.General,
		Radio:      c.cfg.Collectors.AP.Radio,
		Traffic:    c.cfg.Collectors.AP.Traffic,
		Errors:     c.cfg.Collectors.AP.Errors,
		Join:       c.cfg.Collectors.AP.Join,
		Info:       c.cfg.Collectors.AP.Info,
		InfoLabels: c.cfg.Collectors.AP.InfoLabels,
	})

	// Recover panics next to the base collector so the goroutine InfoCacheCollector spawns is covered.
	var collector prometheus.Collector = NewSafeCollector(baseCollector, "AP")

	// Apply caching for info metrics only when info metrics are enabled.
	if c.cfg.Collectors.AP.Info {
		collector = NewInfoCacheCollector(collector, "AP", c.cfg.Collectors.InfoCacheTTL)
	}

	c.registry.MustRegister(collector)
	slog.Debug("Registered AP collector")
}

// registerWLANCollector registers the WLAN collector with its modules.
func (c *Collector) registerWLANCollector(wlanSource wnc.WLANSource, clientSource wnc.ClientSource) {
	baseCollector := NewWLANCollector(wlanSource, clientSource, WLANMetrics{
		General:    c.cfg.Collectors.WLAN.General,
		Traffic:    c.cfg.Collectors.WLAN.Traffic,
		Config:     c.cfg.Collectors.WLAN.Config,
		Info:       c.cfg.Collectors.WLAN.Info,
		InfoLabels: c.cfg.Collectors.WLAN.InfoLabels,
	})

	// Recover panics next to the base collector so the goroutine InfoCacheCollector spawns is covered.
	var collector prometheus.Collector = NewSafeCollector(baseCollector, "WLAN")

	// Apply caching for info metrics only when info metrics are enabled.
	if c.cfg.Collectors.WLAN.Info {
		collector = NewInfoCacheCollector(collector, "WLAN", c.cfg.Collectors.InfoCacheTTL)
	}

	c.registry.MustRegister(collector)
	slog.Debug("Registered WLAN collector")
}

// registerClientCollector registers the Client collector with its modules.
func (c *Collector) registerClientCollector(clientSource wnc.ClientSource) {
	baseCollector := NewClientCollector(clientSource, ClientMetrics{
		General:    c.cfg.Collectors.Client.General,
		Radio:      c.cfg.Collectors.Client.Radio,
		Traffic:    c.cfg.Collectors.Client.Traffic,
		Errors:     c.cfg.Collectors.Client.Errors,
		Info:       c.cfg.Collectors.Client.Info,
		InfoLabels: c.cfg.Collectors.Client.InfoLabels,
	})

	// Recover panics next to the base collector so the goroutine InfoCacheCollector spawns is covered.
	var collector prometheus.Collector = NewSafeCollector(baseCollector, "Client")

	// Apply caching for info metrics only when info metrics are enabled.
	if c.cfg.Collectors.Client.Info {
		collector = NewInfoCacheCollector(collector, "Client", c.cfg.Collectors.InfoCacheTTL)
	}

	c.registry.MustRegister(collector)
	slog.Debug("Registered Client collector")
}
