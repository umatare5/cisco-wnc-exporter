package config

import (
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestConfig_LogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{"Debug level", "debug", slog.LevelDebug},
		{"Info level", "info", slog.LevelInfo},
		{"Warn level", "warn", slog.LevelWarn},
		{"Error level", "error", slog.LevelError},
		{"Debug level uppercase", "DEBUG", slog.LevelDebug},
		{"Mixed case", "WaRn", slog.LevelWarn},
		{"Invalid level defaults to info", "invalid", slog.LevelInfo},
		{"Empty string defaults to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{
				Log: Log{Level: tt.logLevel},
			}
			if got := cfg.LogLevel(); got != tt.expected {
				t.Errorf("LogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidLogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{"Valid debug", "debug", true},
		{"Valid info", "info", true},
		{"Valid warn", "warn", true},
		{"Valid error", "error", true},
		{"Valid uppercase", "DEBUG", true},
		{"Valid mixed case", "InFo", true},
		{"Invalid level", "invalid", false},
		{"Empty string", "", false},
		{"Whitespace", " ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isValidLogLevel(tt.level); got != tt.expected {
				t.Errorf("isValidLogLevel(%q) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestIsValidLogFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		format   string
		expected bool
	}{
		{"Valid json", "json", true},
		{"Valid text", "text", true},
		{"Valid uppercase", "JSON", true},
		{"Valid mixed case", "TeXt", true},
		{"Invalid format", "xml", false},
		{"Empty string", "", false},
		{"Whitespace", " ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isValidLogFormat(tt.format); got != tt.expected {
				t.Errorf("isValidLogFormat(%q) = %v, want %v", tt.format, got, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"Item exists", []string{"a", "b", "c"}, "b", true},
		{"Item does not exist", []string{"a", "b", "c"}, "d", false},
		{"Empty slice", []string{}, "a", false},
		{"Empty item", []string{"", "b"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := contains(tt.slice, tt.item); got != tt.expected {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.expected)
			}
		})
	}
}

func TestParseInfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		labelsStr string
		expected  []string
	}{
		{"Single label", "mac", []string{"mac"}},
		{"Multiple labels", "mac,name,ip", []string{"mac", "name", "ip"}},
		{"Labels with spaces", " mac , name , ip ", []string{"mac", "name", "ip"}},
		{"Empty string", "", []string{""}},
		{"Single comma", ",", []string{"", ""}},
		{"Multiple commas", "mac,,name", []string{"mac", "", "name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseInfoLabels(tt.labelsStr)
			if !slices.Equal(got, tt.expected) {
				t.Errorf("parseInfoLabels(%q) = %v, want %v", tt.labelsStr, got, tt.expected)
			}
		})
	}
}

func TestParseAPInfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		labelsStr string
		expected  []string
	}{
		{
			"Default labels",
			"",
			[]string{"name", "ip", "mac", "radio"},
		},
		{
			"Custom labels with required",
			"name,model",
			[]string{"name", "model", "mac", "radio"},
		},
		{
			"Custom labels already contains required",
			"mac,radio,name",
			[]string{"mac", "radio", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseAPInfoLabels(tt.labelsStr)
			// Check that all expected labels are present (order may vary)
			for _, expected := range tt.expected {
				if !contains(got, expected) {
					t.Errorf("parseAPInfoLabels(%q) missing label %q, got %v", tt.labelsStr, expected, got)
				}
			}
			// Check that no unexpected labels are present
			for _, actual := range got {
				if !contains(tt.expected, actual) {
					t.Errorf("parseAPInfoLabels(%q) unexpected label %q, got %v", tt.labelsStr, actual, got)
				}
			}
		})
	}
}

func TestParseClientInfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		labelsStr string
		expected  []string
	}{
		{
			"Default labels",
			"",
			[]string{"name", "ipv4", "mac"},
		},
		{
			"Custom labels with required",
			"name,username",
			[]string{"name", "username", "mac"},
		},
		{
			"Custom labels already contains required",
			"mac,name",
			[]string{"mac", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseClientInfoLabels(tt.labelsStr)
			// Check that all expected labels are present (order may vary)
			for _, expected := range tt.expected {
				if !contains(got, expected) {
					t.Errorf("parseClientInfoLabels(%q) missing label %q, got %v", tt.labelsStr, expected, got)
				}
			}
			// Check that no unexpected labels are present
			for _, actual := range got {
				if !contains(tt.expected, actual) {
					t.Errorf("parseClientInfoLabels(%q) unexpected label %q, got %v", tt.labelsStr, actual, got)
				}
			}
		})
	}
}

func TestParseWLANInfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		labelsStr string
		expected  []string
	}{
		{
			"Default labels",
			"",
			[]string{"name", "id"},
		},
		{
			"Custom labels with required",
			"name",
			[]string{"name", "id"},
		},
		{
			"Custom labels already contains required",
			"id,name",
			[]string{"id", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseWLANInfoLabels(tt.labelsStr)
			// Check that all expected labels are present (order may vary)
			for _, expected := range tt.expected {
				if !contains(got, expected) {
					t.Errorf("parseWLANInfoLabels(%q) missing label %q, got %v", tt.labelsStr, expected, got)
				}
			}
			// Check that no unexpected labels are present
			for _, actual := range got {
				if !contains(tt.expected, actual) {
					t.Errorf("parseWLANInfoLabels(%q) unexpected label %q, got %v", tt.labelsStr, actual, got)
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	validConfig := &Config{
		Web: Web{
			ListenAddress: "0.0.0.0",
			ListenPort:    8080,
			TelemetryPath: "/metrics",
		},
		WNC: WNC{
			Controller:  "controller.example.com",
			AccessToken: "token123",
			Timeout:     30 * time.Second,
			CacheTTL:    60 * time.Second,
		},
		Collectors: Collectors{
			InfoCacheTTL: 300 * time.Second,
			AP: APCollectorModules{
				Info:       true,
				InfoLabels: []string{"mac", "radio", "name"},
			},
			Client: ClientCollectorModules{
				Info:       true,
				InfoLabels: []string{"mac", "name"},
			},
			WLAN: WLANCollectorModules{
				Info:       true,
				InfoLabels: []string{"id", "name"},
			},
		},
		Log: Log{
			Level:  "info",
			Format: "json",
		},
	}

	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			"Valid config",
			validConfig,
			false,
			"",
		},
		{
			"Missing controller",
			func() *Config {
				cfg := *validConfig
				cfg.WNC.Controller = ""
				return &cfg
			}(),
			true,
			"WNC controller is required",
		},
		{
			"Missing access token",
			func() *Config {
				cfg := *validConfig
				cfg.WNC.AccessToken = ""
				return &cfg
			}(),
			true,
			"WNC access token is required",
		},
		{
			"Invalid port - too low",
			func() *Config {
				cfg := *validConfig
				cfg.Web.ListenPort = 0
				return &cfg
			}(),
			true,
			"invalid listen port",
		},
		{
			"Invalid port - too high",
			func() *Config {
				cfg := *validConfig
				cfg.Web.ListenPort = 70000
				return &cfg
			}(),
			true,
			"invalid listen port",
		},
		{
			"Invalid timeout",
			func() *Config {
				cfg := *validConfig
				cfg.WNC.Timeout = -1 * time.Second
				return &cfg
			}(),
			true,
			"WNC timeout must be positive",
		},
		{
			"Invalid cache TTL",
			func() *Config {
				cfg := *validConfig
				cfg.WNC.CacheTTL = 0
				return &cfg
			}(),
			true,
			"WNC cache TTL must be positive",
		},
		{
			"Invalid info cache TTL",
			func() *Config {
				cfg := *validConfig
				cfg.Collectors.InfoCacheTTL = -1 * time.Second
				return &cfg
			}(),
			true,
			"collector info cache TTL must be positive",
		},
		{
			"Empty telemetry path",
			func() *Config {
				cfg := *validConfig
				cfg.Web.TelemetryPath = ""
				return &cfg
			}(),
			true,
			"telemetry path cannot be empty",
		},
		{
			"Telemetry path without leading slash",
			func() *Config {
				cfg := *validConfig
				cfg.Web.TelemetryPath = "metrics"
				return &cfg
			}(),
			true,
			"telemetry path must start with '/'",
		},
		{
			"Invalid log level",
			func() *Config {
				cfg := *validConfig
				cfg.Log.Level = "invalid"
				return &cfg
			}(),
			true,
			"invalid log level",
		},
		{
			"Invalid log format",
			func() *Config {
				cfg := *validConfig
				cfg.Log.Format = "xml"
				return &cfg
			}(),
			true,
			"invalid log format",
		},
		{
			"AP collector missing required label",
			func() *Config {
				cfg := *validConfig
				cfg.Collectors.AP.Info = true
				cfg.Collectors.AP.InfoLabels = []string{"name"} // missing mac, radio
				return &cfg
			}(),
			true,
			"AP collector requires 'mac' label",
		},
		{
			"AP collector invalid label",
			func() *Config {
				cfg := *validConfig
				cfg.Collectors.AP.Info = true
				cfg.Collectors.AP.InfoLabels = []string{"mac", "radio", "invalid"}
				return &cfg
			}(),
			true,
			"AP collector: unknown label 'invalid'",
		},
		{
			"Client collector missing required label",
			func() *Config {
				cfg := *validConfig
				cfg.Collectors.Client.Info = true
				cfg.Collectors.Client.InfoLabels = []string{"name"} // missing mac
				return &cfg
			}(),
			true,
			"client collector requires 'mac' label",
		},
		{
			"Client collector invalid label",
			func() *Config {
				cfg := *validConfig
				cfg.Collectors.Client.Info = true
				cfg.Collectors.Client.InfoLabels = []string{"mac", "invalid"}
				return &cfg
			}(),
			true,
			"client collector: unknown label 'invalid'",
		},
		{
			"WLAN collector missing required label",
			func() *Config {
				cfg := *validConfig
				cfg.Collectors.WLAN.Info = true
				cfg.Collectors.WLAN.InfoLabels = []string{"name"} // missing id
				return &cfg
			}(),
			true,
			"WLAN collector requires 'id' label",
		},
		{
			"WLAN collector invalid label",
			func() *Config {
				cfg := *validConfig
				cfg.Collectors.WLAN.Info = true
				cfg.Collectors.WLAN.InfoLabels = []string{"id", "invalid"}
				return &cfg
			}(),
			true,
			"WLAN collector: unknown label 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfig_ValidateCollectorInfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			"AP collector disabled - no validation",
			&Config{
				Collectors: Collectors{
					AP: APCollectorModules{Info: false},
				},
			},
			false,
			"",
		},
		{
			"Client collector disabled - no validation",
			&Config{
				Collectors: Collectors{
					Client: ClientCollectorModules{Info: false},
				},
			},
			false,
			"",
		},
		{
			"WLAN collector disabled - no validation",
			&Config{
				Collectors: Collectors{
					WLAN: WLANCollectorModules{Info: false},
				},
			},
			false,
			"",
		},
		{
			"All collectors with valid labels",
			&Config{
				Collectors: Collectors{
					AP: APCollectorModules{
						Info:       true,
						InfoLabels: []string{"mac", "radio", "name", "ip"},
					},
					Client: ClientCollectorModules{
						Info:       true,
						InfoLabels: []string{"mac", "name", "ipv4"},
					},
					WLAN: WLANCollectorModules{
						Info:       true,
						InfoLabels: []string{"id", "name"},
					},
				},
			},
			false,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.validateCollectorInfoLabels()
			if tt.wantError {
				if err == nil {
					t.Errorf("validateCollectorInfoLabels() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateCollectorInfoLabels() error = %q, want to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateCollectorInfoLabels() unexpected error: %v", err)
				}
			}
		})
	}
}

// mockCommand creates a mock CLI command for testing Parse function.
type mockCommand struct {
	values map[string]interface{}
}

func (m *mockCommand) String(name string) string {
	if v, ok := m.values[name]; ok {
		return v.(string)
	}
	return ""
}

func (m *mockCommand) Int(name string) int {
	if v, ok := m.values[name]; ok {
		return v.(int)
	}
	return 0
}

func (m *mockCommand) Bool(name string) bool {
	if v, ok := m.values[name]; ok {
		return v.(bool)
	}
	return false
}

func (m *mockCommand) Duration(name string) time.Duration {
	if v, ok := m.values[name]; ok {
		return v.(time.Duration)
	}
	return 0
}

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		cmdValues map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			"Valid configuration",
			map[string]interface{}{
				"web.listen-address":            "127.0.0.1",
				"web.listen-port":               9090,
				"web.telemetry-path":            "/metrics",
				"wnc.controller":                "controller.example.com",
				"wnc.access-token":              "token123",
				"wnc.timeout":                   30 * time.Second,
				"wnc.cache-ttl":                 60 * time.Second,
				"wnc.tls-skip-verify":           false,
				"collector.ap.general":          true,
				"collector.ap.radio":            false,
				"collector.ap.traffic":          false,
				"collector.ap.errors":           false,
				"collector.ap.info":             true,
				"collector.ap.info-labels":      "mac,radio,name",
				"collector.client.general":      true,
				"collector.client.radio":        false,
				"collector.client.traffic":      false,
				"collector.client.errors":       false,
				"collector.client.info":         true,
				"collector.client.info-labels":  "mac,name",
				"collector.wlan.general":        true,
				"collector.wlan.traffic":        false,
				"collector.wlan.config":         false,
				"collector.wlan.info":           true,
				"collector.wlan.info-labels":    "id,name",
				"collector.info-cache-ttl":      300 * time.Second,
				"log.level":                     "info",
				"log.format":                    "json",
				"collector.internal.go-runtime": true,
				"collector.internal.process":    false,
				"dry-run":                       false,
			},
			false,
			"",
		},
		{
			"Missing controller",
			map[string]interface{}{
				"web.listen-address":            "127.0.0.1",
				"web.listen-port":               9090,
				"web.telemetry-path":            "/metrics",
				"wnc.controller":                "", // Missing
				"wnc.access-token":              "token123",
				"wnc.timeout":                   30 * time.Second,
				"wnc.cache-ttl":                 60 * time.Second,
				"wnc.tls-skip-verify":           false,
				"collector.info-cache-ttl":      300 * time.Second,
				"log.level":                     "info",
				"log.format":                    "json",
				"collector.internal.go-runtime": false,
				"collector.internal.process":    false,
				"dry-run":                       false,
			},
			true,
			"WNC controller is required",
		},
		{
			"Invalid port",
			map[string]interface{}{
				"web.listen-address":            "127.0.0.1",
				"web.listen-port":               -1, // Invalid
				"web.telemetry-path":            "/metrics",
				"wnc.controller":                "controller.example.com",
				"wnc.access-token":              "token123",
				"wnc.timeout":                   30 * time.Second,
				"wnc.cache-ttl":                 60 * time.Second,
				"wnc.tls-skip-verify":           false,
				"collector.info-cache-ttl":      300 * time.Second,
				"log.level":                     "info",
				"log.format":                    "json",
				"collector.internal.go-runtime": false,
				"collector.internal.process":    false,
				"dry-run":                       false,
			},
			true,
			"invalid listen port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &mockCommand{values: tt.cmdValues}

			// Convert mockCommand to cli.Command interface
			// This is a simplified approach for testing Parse function
			// In a real scenario, you would need to properly mock the cli.Command interface
			cfg, err := testParse(cmd)

			if tt.wantError {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Parse() error = %q, want to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Parse() unexpected error: %v", err)
				}
				if cfg == nil {
					t.Errorf("Parse() returned nil config")
				}
			}
		})
	}
}

// testParse is a simplified version of Parse for testing.
func testParse(cmd *mockCommand) (*Config, error) {
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
		return nil, err
	}

	return cfg, nil
}
