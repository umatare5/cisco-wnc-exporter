// Package collector provides registry management and collector registration functionality.
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
	sharedDataSource wnc.DataSource // Shared data source for caching
}

// NewCollector creates a new collector manager with a custom Prometheus registry and SharedDataSource.
func NewCollector(cfg *config.Config) *Collector {
	sharedDataSource := wnc.NewDataSource(cfg.WNC)

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
	// Register AP collector if any AP module is enabled
	if isEnabled(
		c.cfg.Collectors.AP.Inventory,
		c.cfg.Collectors.AP.Info,
		c.cfg.Collectors.AP.State,
		c.cfg.Collectors.AP.Phy,
		c.cfg.Collectors.AP.RF,
		c.cfg.Collectors.AP.Traffic,
		c.cfg.Collectors.AP.Errors,
		c.cfg.Collectors.AP.CPU,
		c.cfg.Collectors.AP.Memory,
	) {
		apSource := wnc.NewAPSource(c.sharedDataSource)
		rrmSource := wnc.NewRRMSource(c.sharedDataSource)
		clientSource := wnc.NewClientSource(c.sharedDataSource)
		c.registerAPCollector(apSource, rrmSource, clientSource)
	} else {
		slog.Debug("Skipped AP collector registration - all modules disabled")
	}

	// Register WLAN collector if any WLAN module is enabled
	if isEnabled(
		c.cfg.Collectors.WLAN.Inventory,
		c.cfg.Collectors.WLAN.Info,
		c.cfg.Collectors.WLAN.State,
		c.cfg.Collectors.WLAN.Traffic,
	) {
		wlanSource := wnc.NewWLANSource(c.sharedDataSource)
		clientSource := wnc.NewClientSource(c.sharedDataSource)
		c.registerWLANCollector(wlanSource, clientSource)
	} else {
		slog.Debug("Skipped WLAN collector registration (no modules enabled)")
	}

	// Register Client collector if any Client module is enabled
	if isEnabled(
		c.cfg.Collectors.Client.Inventory,
		c.cfg.Collectors.Client.Info,
		c.cfg.Collectors.Client.Session,
		c.cfg.Collectors.Client.Phy,
		c.cfg.Collectors.Client.RF,
		c.cfg.Collectors.Client.Traffic,
		c.cfg.Collectors.Client.Errors,
		c.cfg.Collectors.Client.Power,
	) {
		clientSource := wnc.NewClientSource(c.sharedDataSource)
		c.registerClientCollector(clientSource)
	} else {
		slog.Debug("Skipped Client collector registration (no modules enabled)")
	}
}

// registerAPCollector registers the AP collector with its modules.
func (c *Collector) registerAPCollector(apSource wnc.APSource, rrmSource wnc.RRMSource, clientSource wnc.ClientSource) {
	baseCollector := NewAPCollector(apSource, rrmSource, clientSource, APMetrics{
		Inventory:  c.cfg.Collectors.AP.Inventory,
		Info:       c.cfg.Collectors.AP.Info,
		InfoLabels: c.cfg.Collectors.AP.InfoLabels,
		State:      c.cfg.Collectors.AP.State,
		Phy:        c.cfg.Collectors.AP.Phy,
		RF:         c.cfg.Collectors.AP.RF,
		Traffic:    c.cfg.Collectors.AP.Traffic,
		Errors:     c.cfg.Collectors.AP.Errors,
		CPU:        c.cfg.Collectors.AP.CPU,
		Memory:     c.cfg.Collectors.AP.Memory,
	})

	// Apply caching when info metrics are enabled.
	var collector prometheus.Collector = baseCollector
	if c.cfg.Collectors.AP.Info {
		collector = NewCacheCollector(baseCollector, "AP", c.cfg.CacheTTL)
	}

	c.registry.MustRegister(collector)
	slog.Debug("Registered AP collector")
}

// registerWLANCollector registers the WLAN collector with its modules.
func (c *Collector) registerWLANCollector(wlanSource wnc.WLANSource, clientSource wnc.ClientSource) {
	baseCollector := NewWLANCollector(wlanSource, clientSource, WLANMetrics{
		Inventory:  c.cfg.Collectors.WLAN.Inventory,
		Info:       c.cfg.Collectors.WLAN.Info,
		InfoLabels: c.cfg.Collectors.WLAN.InfoLabels,
		State:      c.cfg.Collectors.WLAN.State,
		Traffic:    c.cfg.Collectors.WLAN.Traffic,
		Security:   c.cfg.Collectors.WLAN.Security,
		Networking: c.cfg.Collectors.WLAN.Networking,
	})

	// Apply caching when info metrics are enabled.
	var collector prometheus.Collector = baseCollector
	if c.cfg.Collectors.WLAN.Info {
		collector = NewCacheCollector(baseCollector, "WLAN", c.cfg.CacheTTL)
	}

	c.registry.MustRegister(collector)
	slog.Debug("Registered WLAN collector")
}

// registerClientCollector registers the Client collector with its modules.
func (c *Collector) registerClientCollector(clientSource wnc.ClientSource) {
	baseCollector := NewClientCollector(clientSource, ClientMetrics{
		Inventory:  c.cfg.Collectors.Client.Inventory,
		Info:       c.cfg.Collectors.Client.Info,
		InfoLabels: c.cfg.Collectors.Client.InfoLabels,
		Session:    c.cfg.Collectors.Client.Session,
		Phy:        c.cfg.Collectors.Client.Phy,
		RF:         c.cfg.Collectors.Client.RF,
		Traffic:    c.cfg.Collectors.Client.Traffic,
		Errors:     c.cfg.Collectors.Client.Errors,
		Power:      c.cfg.Collectors.Client.Power,
	})

	// Apply caching when info metrics are enabled.
	var collector prometheus.Collector = baseCollector
	if c.cfg.Collectors.Client.Info {
		collector = NewCacheCollector(baseCollector, "Client", c.cfg.CacheTTL)
	}

	c.registry.MustRegister(collector)
	slog.Debug("Registered Client collector")
}

// isEnabled returns true if any of the provided boolean values is true.
func isEnabled(flags ...bool) bool {
	for _, enabled := range flags {
		if enabled {
			return true
		}
	}
	return false
}
