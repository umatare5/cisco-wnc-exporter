// Package config provides configuration parsing and validation for cisco-wnc-exporter.
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

// Default configuration constants.
const (
	DefaultListenAddress     = "0.0.0.0"
	DefaultListenPort        = 10040
	DefaultTelemetryPath     = "/metrics"
	DefaultWNCTimeout        = 55 * time.Second
	DefaultWNCCacheTTL       = 55 * time.Second
	DefaultCollectorCacheTTL = 1800 * time.Second
	DefaultLogLevel          = "info"
	DefaultLogFormat         = "json"

	// Info-labels default values.
	DefaultAPInfoLabels     = "name,ip"
	DefaultClientInfoLabels = "name,ipv4"
	DefaultWLANInfoLabels   = "name"

	// Info-labels available options.
	AvailableAPInfoLabels     = "name,ip,band,model,serial,sw_version,eth_mac"
	AvailableClientInfoLabels = "ap,band,wlan,name,username,ipv4,ipv6"
	AvailableWLANInfoLabels   = "name"

	// Info-labels required fields.
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
	AP       CollectorModules `json:"ap"`
	Client   CollectorModules `json:"client"`
	WLAN     CollectorModules `json:"wlan"`
	CacheTTL time.Duration    `json:"cache_ttl"`
}

// CollectorModules represents a set of collector modules.
type CollectorModules struct {
	Inventory  bool     `json:"inventory"`
	Info       bool     `json:"info"`
	InfoLabels []string `json:"info_labels"`
	State      bool     `json:"state"`
	Phy        bool     `json:"phy"`
	RF         bool     `json:"rf"`
	Traffic    bool     `json:"traffic"`
	Errors     bool     `json:"errors"`
	CPU        bool     `json:"cpu"`
	Memory     bool     `json:"memory"`
	Session    bool     `json:"session"`
	Power      bool     `json:"power"`
	Security   bool     `json:"security"`
	Networking bool     `json:"networking"`
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
			AP: CollectorModules{
				Inventory:  cmd.Bool("collector.ap.inventory"),
				Info:       cmd.Bool("collector.ap.info"),
				InfoLabels: parseAPInfoLabels(cmd.String("collector.ap.info-labels")),
				State:      cmd.Bool("collector.ap.state"),
				Phy:        cmd.Bool("collector.ap.phy"),
				RF:         cmd.Bool("collector.ap.rf"),
				Traffic:    cmd.Bool("collector.ap.traffic"),
				Errors:     cmd.Bool("collector.ap.errors"),
				CPU:        cmd.Bool("collector.ap.cpu"),
				Memory:     cmd.Bool("collector.ap.memory"),
			},
			Client: CollectorModules{
				Inventory:  cmd.Bool("collector.client.inventory"),
				Info:       cmd.Bool("collector.client.info"),
				InfoLabels: parseClientInfoLabels(cmd.String("collector.client.info-labels")),
				Session:    cmd.Bool("collector.client.session"),
				Phy:        cmd.Bool("collector.client.phy"),
				RF:         cmd.Bool("collector.client.rf"),
				Traffic:    cmd.Bool("collector.client.traffic"),
				Errors:     cmd.Bool("collector.client.errors"),
				Power:      cmd.Bool("collector.client.power"),
			},
			WLAN: CollectorModules{
				Inventory:  cmd.Bool("collector.wlan.inventory"),
				Info:       cmd.Bool("collector.wlan.info"),
				InfoLabels: parseWLANInfoLabels(cmd.String("collector.wlan.info-labels")),
				State:      cmd.Bool("collector.wlan.state"),
				Traffic:    cmd.Bool("collector.wlan.traffic"),
				Security:   cmd.Bool("collector.wlan.security"),
				Networking: cmd.Bool("collector.wlan.networking"),
			},
			CacheTTL: cmd.Duration("collector.cache-ttl"),
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
			c.Collectors.CacheTTL <= 0, fmt.Sprintf("collector cache TTL must be positive, got: %v", c.Collectors.CacheTTL),
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
	return contains(validLevels, strings.ToLower(level))
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
	requiredLabels := strings.Split(RequiredAPInfoLabels, ",")

	// Ensure required labels are always present
	for _, required := range requiredLabels {
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
	requiredLabels := strings.Split(RequiredClientInfoLabels, ",")

	// Ensure required labels are always present
	for _, required := range requiredLabels {
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
	requiredLabels := strings.Split(RequiredWLANInfoLabels, ",")

	// Ensure required labels are always present
	for _, required := range requiredLabels {
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
		apRequiredLabels := strings.Split(RequiredAPInfoLabels, ",")
		apAvailableLabels := strings.Split("mac,radio,"+AvailableAPInfoLabels, ",")
		if err := validateInfoLabels(c.Collectors.AP.InfoLabels, apRequiredLabels, apAvailableLabels, "AP"); err != nil {
			return err
		}
	}

	// Client collector validation
	if c.Collectors.Client.Info {
		clientRequiredLabels := strings.Split(RequiredClientInfoLabels, ",")
		clientAvailableLabels := strings.Split("mac,"+AvailableClientInfoLabels, ",")
		if err := validateInfoLabels(c.Collectors.Client.InfoLabels, clientRequiredLabels, clientAvailableLabels, "Client"); err != nil {
			return err
		}
	}

	// WLAN collector validation
	if c.Collectors.WLAN.Info {
		wlanRequiredLabels := strings.Split(RequiredWLANInfoLabels, ",")
		wlanAvailableLabels := strings.Split("id,"+AvailableWLANInfoLabels, ",")
		if err := validateInfoLabels(c.Collectors.WLAN.InfoLabels, wlanRequiredLabels, wlanAvailableLabels, "WLAN"); err != nil {
			return err
		}
	}

	return nil
}

// validateInfoLabels validates info labels against required and available labels.
func validateInfoLabels(infoLabels, requiredLabels, availableLabels []string, collectorName string) error {
	// Ensure required labels are present
	for _, required := range requiredLabels {
		if !contains(infoLabels, required) {
			return fmt.Errorf("%s collector requires '%s' label in info-labels", collectorName, required)
		}
	}

	// Ensure all specified labels are available
	for _, label := range infoLabels {
		if !contains(availableLabels, label) {
			return fmt.Errorf(
				"%s collector: unknown label '%s' in info-labels (available: %s)",
				collectorName,
				label,
				strings.Join(availableLabels, ", "),
			)
		}
	}

	return nil
}
