// Package config provides configuration parsing and validation.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	DefaultListenAddress         = "0.0.0.0"
	DefaultListenPort            = 10040
	DefaultTelemetryPath         = "/metrics"
	DefaultWNCTimeout            = 55 * time.Second
	DefaultWNCCacheTTL           = 55 * time.Second
	DefaultCollectorInfoCacheTTL = 1800 * time.Second
	DefaultLogLevel              = "info"
	DefaultLogFormat             = "json"

	DefaultAPInfoLabels     = "name,ip"
	DefaultClientInfoLabels = "name,ipv4"
	DefaultWLANInfoLabels   = "name"

	AvailableAPInfoLabels     = "name,ip,band,model,serial,sw_version,eth_mac"
	AvailableClientInfoLabels = "ap,band,wlan,name,username,ipv4,ipv6"
	AvailableWLANInfoLabels   = "name"

	RequiredAPInfoLabels     = "mac,radio"
	RequiredClientInfoLabels = "mac"
	RequiredWLANInfoLabels   = "id"
)

// Config represents the complete configuration.
type Config struct {
	Web               Web               `json:"web"`
	WNC               WNC               `json:"wnc"`
	Collectors        Collectors        `json:"collectors"`
	Log               Log               `json:"log"`
	InternalCollector InternalCollector `json:"internal_collector"`
	DryRun            bool              `json:"dry_run"`
}

// Web holds HTTP server configuration.
type Web struct {
	ListenAddress string `json:"listen_address"`
	ListenPort    int    `json:"listen_port"`
	TelemetryPath string `json:"telemetry_path"`
}

// WNC holds controller connection configuration.
type WNC struct {
	Controller    string        `json:"controller"`
	AccessToken   string        `json:"-"` // Never serialize credentials
	Timeout       time.Duration `json:"timeout"`
	CacheTTL      time.Duration `json:"cache_ttl"`
	TLSSkipVerify bool          `json:"tls_skip_verify"`
}

// Collectors holds collector module configuration.
type Collectors struct {
	AP           APCollectorModules     `json:"ap"`
	Client       ClientCollectorModules `json:"client"`
	WLAN         WLANCollectorModules   `json:"wlan"`
	InfoCacheTTL time.Duration          `json:"info_cache_ttl"`
}

// APCollectorModules represents AP collector modules.
type APCollectorModules struct {
	// General: admin_state, oper_state, radio_state, config_state, uptime, CPU, memory
	General bool `json:"general"`
	// Radio: channel, power, noise, utilization
	Radio bool `json:"radio"`
	// Traffic: clients, bytes, packets, frames
	Traffic bool `json:"traffic"`
	// Errors: errors, drops, retries, failures
	Errors bool `json:"errors"`
	// Info: info metric with labels
	Info       bool     `json:"info"`
	InfoLabels []string `json:"info_labels"`
}

// ClientCollectorModules represents Client collector modules.
type ClientCollectorModules struct {
	// General: state, uptime, power_save_state
	General bool `json:"general"`
	// Radio: protocol, mcs, streams, speed, rssi, snr
	Radio bool `json:"radio"`
	// Traffic: bytes, packets
	Traffic bool `json:"traffic"`
	// Errors: retries, drops, failures
	Errors bool `json:"errors"`
	// Info: info metric with labels
	Info       bool     `json:"info"`
	InfoLabels []string `json:"info_labels"`
}

// WLANCollectorModules represents WLAN collector modules.
type WLANCollectorModules struct {
	// General: enabled
	General bool `json:"general"`
	// Traffic: clients, bytes
	Traffic bool `json:"traffic"`
	// Config: auth, security, networking settings
	Config bool `json:"config"`
	// Info: info metric with labels
	Info       bool     `json:"info"`
	InfoLabels []string `json:"info_labels"`
}

// Log holds logging configuration.
type Log struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// InternalCollector holds internal metrics collection configuration.
type InternalCollector struct {
	EnableGoCollector      bool `json:"enable_go_collector"`
	EnableProcessCollector bool `json:"enable_process_collector"`
}

// Parse parses configuration from CLI command and environment variables.
func Parse(cmd *cli.Command) (*Config, error) {
	cfg := &Config{
		Web: Web{
			ListenAddress: cmd.String("web.listen-address"),
			ListenPort:    cmd.Int("web.listen-port"),
			TelemetryPath: cmd.String("web.telemetry-path"),
		},
		WNC: WNC{
			Controller:    cmd.String("wnc.controller"),
			AccessToken:   cmd.String("wnc.access-token"),
			Timeout:       cmd.Duration("wnc.timeout"),
			CacheTTL:      cmd.Duration("wnc.cache-ttl"),
			TLSSkipVerify: cmd.Bool("wnc.tls-skip-verify"),
		},
		Collectors: Collectors{
			AP: APCollectorModules{
				General:    cmd.Bool("collector.ap.general"),
				Radio:      cmd.Bool("collector.ap.radio"),
				Traffic:    cmd.Bool("collector.ap.traffic"),
				Errors:     cmd.Bool("collector.ap.errors"),
				Info:       cmd.Bool("collector.ap.info"),
				InfoLabels: parseAPInfoLabels(cmd.String("collector.ap.info-labels")),
			},
			Client: ClientCollectorModules{
				General:    cmd.Bool("collector.client.general"),
				Radio:      cmd.Bool("collector.client.radio"),
				Traffic:    cmd.Bool("collector.client.traffic"),
				Errors:     cmd.Bool("collector.client.errors"),
				Info:       cmd.Bool("collector.client.info"),
				InfoLabels: parseClientInfoLabels(cmd.String("collector.client.info-labels")),
			},
			WLAN: WLANCollectorModules{
				General:    cmd.Bool("collector.wlan.general"),
				Traffic:    cmd.Bool("collector.wlan.traffic"),
				Config:     cmd.Bool("collector.wlan.config"),
				Info:       cmd.Bool("collector.wlan.info"),
				InfoLabels: parseWLANInfoLabels(cmd.String("collector.wlan.info-labels")),
			},
			InfoCacheTTL: cmd.Duration("collector.info-cache-ttl"),
		},
		Log: Log{
			Level:  cmd.String("log.level"),
			Format: cmd.String("log.format"),
		},
		InternalCollector: InternalCollector{
			EnableGoCollector:      cmd.Bool("collector.internal.go-runtime"),
			EnableProcessCollector: cmd.Bool("collector.internal.process"),
		},
		DryRun: cmd.Bool("dry-run"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate performs configuration validation.
func (c *Config) Validate() error {
	validationRules := []struct {
		condition bool
		message   string
	}{
		{
			c.WNC.Controller == "", "WNC controller is required (--wnc.controller or WNC_CONTROLLER)",
		},
		{
			c.WNC.AccessToken == "", "WNC access token is required (--wnc.access-token or WNC_ACCESS_TOKEN)",
		},
		{
			c.Web.ListenPort < 1 || c.Web.ListenPort > 65535,
			fmt.Sprintf("invalid listen port: %d (must be 1-65535)", c.Web.ListenPort),
		},
		{
			c.WNC.Timeout <= 0, fmt.Sprintf("WNC timeout must be positive, got: %v", c.WNC.Timeout),
		},
		{
			c.WNC.CacheTTL <= 0, fmt.Sprintf("WNC cache TTL must be positive, got: %v", c.WNC.CacheTTL),
		},
		{
			c.Collectors.InfoCacheTTL <= 0, fmt.Sprintf("collector info cache TTL must be positive, got: %v", c.Collectors.InfoCacheTTL),
		},
		{
			c.Web.TelemetryPath == "", "telemetry path cannot be empty",
		},
		{
			!strings.HasPrefix(c.Web.TelemetryPath, "/"),
			"telemetry path must start with '/': " + c.Web.TelemetryPath,
		},
		{
			!isValidLogLevel(c.Log.Level),
			fmt.Sprintf("invalid log level: %s (must be one of: debug, info, warn, error)", c.Log.Level),
		},
		{
			!isValidLogFormat(c.Log.Format),
			fmt.Sprintf("invalid log format: %s (must be one of: json, text)", c.Log.Format),
		},
	}

	for _, rule := range validationRules {
		if rule.condition {
			return errors.New(rule.message)
		}
	}

	// Validate collector info labels
	if err := c.validateCollectorInfoLabels(); err != nil {
		return fmt.Errorf("info labels validation failed: %w", err)
	}

	return nil
}

// LogLevel returns the slog.Level for the configured log level.
func (c *Config) LogLevel() slog.Level {
	switch strings.ToLower(c.Log.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// isValidLogLevel checks if the log level is valid.
func isValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error"}
	return slices.Contains(validLevels, strings.ToLower(level))
}

// isValidLogFormat checks if the log format is valid.
func isValidLogFormat(format string) bool {
	validFormats := []string{"json", "text"}
	return contains(validFormats, strings.ToLower(format))
}

// contains checks if a slice contains a specific item.
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// parseInfoLabels parses comma-separated info-labels string.
func parseInfoLabels(labelsStr string) []string {
	labels := strings.Split(labelsStr, ",")
	for i, label := range labels {
		labels[i] = strings.TrimSpace(label)
	}
	return labels
}

// parseAPInfoLabels parses AP info labels with required labels auto-added.
func parseAPInfoLabels(labelsStr string) []string {
	if labelsStr == "" {
		labelsStr = DefaultAPInfoLabels
	}

	labels := parseInfoLabels(labelsStr)

	// Ensure required labels are always present
	for required := range strings.SplitSeq(RequiredAPInfoLabels, ",") {
		if !contains(labels, required) {
			labels = append(labels, required)
		}
	}
	return labels
}

// parseClientInfoLabels parses Client info labels with required labels auto-added.
func parseClientInfoLabels(labelsStr string) []string {
	if labelsStr == "" {
		labelsStr = DefaultClientInfoLabels
	}

	labels := parseInfoLabels(labelsStr)

	// Ensure required labels are always present
	for required := range strings.SplitSeq(RequiredClientInfoLabels, ",") {
		if !contains(labels, required) {
			labels = append(labels, required)
		}
	}
	return labels
}

// parseWLANInfoLabels parses WLAN info labels with required labels auto-added.
func parseWLANInfoLabels(labelsStr string) []string {
	if labelsStr == "" {
		labelsStr = DefaultWLANInfoLabels
	}

	labels := parseInfoLabels(labelsStr)

	// Ensure required labels are always present
	for required := range strings.SplitSeq(RequiredWLANInfoLabels, ",") {
		if !contains(labels, required) {
			labels = append(labels, required)
		}
	}
	return labels
}

// validateCollectorInfoLabels validates info labels for all collectors.
func (c *Config) validateCollectorInfoLabels() error {
	// AP collector validation
	if c.Collectors.AP.Info {
		apAvailableLabels := strings.Split("mac,radio,"+AvailableAPInfoLabels, ",")

		// Check required labels
		for required := range strings.SplitSeq(RequiredAPInfoLabels, ",") {
			if !contains(c.Collectors.AP.InfoLabels, required) {
				return fmt.Errorf("AP collector requires '%s' label in info-labels", required)
			}
		}

		// Check available labels
		for _, label := range c.Collectors.AP.InfoLabels {
			if !contains(apAvailableLabels, label) {
				return fmt.Errorf(
					"AP collector: unknown label '%s' in info-labels (available: %s)",
					label,
					strings.Join(apAvailableLabels, ", "),
				)
			}
		}
	}

	// Client collector validation
	if c.Collectors.Client.Info {
		clientAvailableLabels := strings.Split("mac,"+AvailableClientInfoLabels, ",")

		// Check required labels
		for required := range strings.SplitSeq(RequiredClientInfoLabels, ",") {
			if !contains(c.Collectors.Client.InfoLabels, required) {
				return fmt.Errorf("client collector requires '%s' label in info-labels", required)
			}
		}

		// Check available labels
		for _, label := range c.Collectors.Client.InfoLabels {
			if !contains(clientAvailableLabels, label) {
				return fmt.Errorf(
					"client collector: unknown label '%s' in info-labels (available: %s)",
					label,
					strings.Join(clientAvailableLabels, ", "),
				)
			}
		}
	}

	// WLAN collector validation
	if c.Collectors.WLAN.Info {
		wlanAvailableLabels := strings.Split("id,"+AvailableWLANInfoLabels, ",")

		// Check required labels
		for required := range strings.SplitSeq(RequiredWLANInfoLabels, ",") {
			if !contains(c.Collectors.WLAN.InfoLabels, required) {
				return fmt.Errorf("WLAN collector requires '%s' label in info-labels", required)
			}
		}

		// Check available labels
		for _, label := range c.Collectors.WLAN.InfoLabels {
			if !contains(wlanAvailableLabels, label) {
				return fmt.Errorf(
					"WLAN collector: unknown label '%s' in info-labels (available: %s)",
					label,
					strings.Join(wlanAvailableLabels, ", "),
				)
			}
		}
	}

	return nil
}
