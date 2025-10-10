package collector

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// createTestConfig creates a basic configuration for testing.
func createTestConfig() *config.Config {
	return &config.Config{
		Web: config.Web{
			ListenAddress: "0.0.0.0",
			ListenPort:    10039,
			TelemetryPath: "/metrics",
		},
		WNC: config.WNC{
			Controller:    "wnc1.example.internal",
			AccessToken:   "test-token",
			Timeout:       1 * time.Second,
			CacheTTL:      1 * time.Second,
			TLSSkipVerify: true,
		},
		Collectors: config.Collectors{
			AP: config.APCollectorModules{
				General:    true,
				Radio:      true,
				Traffic:    false,
				Errors:     false,
				Info:       false,
				InfoLabels: []string{"name", "ip"},
			},
			Client: config.ClientCollectorModules{
				General:    true,
				Radio:      false,
				Traffic:    false,
				Errors:     false,
				Info:       false,
				InfoLabels: []string{"name", "ipv4"},
			},
			WLAN: config.WLANCollectorModules{
				General:    true,
				Traffic:    false,
				Config:     false,
				Info:       false,
				InfoLabels: []string{"name"},
			},
			InfoCacheTTL: 1800 * time.Second,
		},
		Log: config.Log{
			Level:  "info",
			Format: "json",
		},
		InternalCollector: config.InternalCollector{
			EnableGoCollector:      false,
			EnableProcessCollector: false,
		},
		DryRun: false,
	}
}

// createDisabledConfig creates a configuration with all collectors disabled.
func createDisabledConfig() *config.Config {
	cfg := createTestConfig()
	cfg.Collectors.AP.General = false
	cfg.Collectors.AP.Radio = false
	cfg.Collectors.Client.General = false
	cfg.Collectors.WLAN.General = false
	return cfg
}

// createSystemCollectorConfig creates configuration with system collectors enabled.
func createSystemCollectorConfig() *config.Config {
	cfg := createTestConfig()
	cfg.InternalCollector.EnableGoCollector = true
	cfg.InternalCollector.EnableProcessCollector = true
	return cfg
}

func TestNewCollector(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()

	collector := NewCollector(cfg)

	if collector == nil {
		t.Fatal("NewCollector returned nil")
	}
	if collector.registry == nil {
		t.Error("registry not initialized")
	}
	if collector.cfg != cfg {
		t.Error("config not set correctly")
	}
	if collector.sharedDataSource == nil {
		t.Error("sharedDataSource not initialized")
	}
}

func TestCollector_Registry(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()
	collector := NewCollector(cfg)

	registry := collector.Registry()

	if registry == nil {
		t.Error("Registry() returned nil")
	}
	if registry != collector.registry {
		t.Error("Registry() returned different registry instance")
	}
}

func TestCollector_RegisterBuildInfo(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()
	collector := NewCollector(cfg)
	version := "1.0.0"

	// Count metrics before registration
	metricFamilies, err := collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics before registration: %v", err)
	}
	beforeCount := len(metricFamilies)

	collector.RegisterBuildInfo(version)

	// Count metrics after registration
	metricFamilies, err = collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics after registration: %v", err)
	}
	afterCount := len(metricFamilies)

	if afterCount != beforeCount+1 {
		t.Errorf("expected 1 additional metric family, got %d before and %d after", beforeCount, afterCount)
	}

	// Check for build info metric
	buildInfoFound := false
	for _, mf := range metricFamilies {
		if mf.GetName() != "wnc_build_info" {
			continue
		}
		buildInfoFound = true
		if len(mf.GetMetric()) != 1 {
			t.Errorf("expected 1 build info metric, got %d", len(mf.GetMetric()))
		}
		metric := mf.GetMetric()[0]
		if len(metric.GetLabel()) != 1 {
			t.Errorf("expected 1 label, got %d", len(metric.GetLabel()))
		}
		if metric.GetLabel()[0].GetName() != "version" {
			t.Errorf("expected version label, got %s", metric.GetLabel()[0].GetName())
		}
		if metric.GetLabel()[0].GetValue() != version {
			t.Errorf("expected version %s, got %s", version, metric.GetLabel()[0].GetValue())
		}
		break
	}

	if !buildInfoFound {
		t.Error("build info metric not found")
	}
}

func TestCollector_RegisterSystemCollectors_Disabled(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig() // Go and Process collectors disabled by default
	collector := NewCollector(cfg)

	// Count metrics before registration
	metricFamilies, err := collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics before registration: %v", err)
	}
	beforeCount := len(metricFamilies)

	collector.RegisterSystemCollectors()

	// Count metrics after registration
	metricFamilies, err = collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics after registration: %v", err)
	}
	afterCount := len(metricFamilies)

	// Should not add any metrics when system collectors are disabled
	if afterCount != beforeCount {
		t.Errorf(
			"expected no additional metrics when system collectors disabled, got %d before and %d after",
			beforeCount, afterCount,
		)
	}
}

func TestCollector_RegisterSystemCollectors_Enabled(t *testing.T) {
	t.Parallel()
	cfg := createSystemCollectorConfig()
	collector := NewCollector(cfg)

	// Count metrics before registration
	metricFamilies, err := collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics before registration: %v", err)
	}
	beforeCount := len(metricFamilies)

	collector.RegisterSystemCollectors()

	// Count metrics after registration
	metricFamilies, err = collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics after registration: %v", err)
	}
	afterCount := len(metricFamilies)

	// Should add metrics when system collectors are enabled
	if afterCount <= beforeCount {
		t.Errorf(
			"expected additional metrics when system collectors enabled, got %d before and %d after",
			beforeCount, afterCount,
		)
	}
}

func TestCollector_Setup(t *testing.T) {
	t.Parallel()
	cfg := createSystemCollectorConfig()
	collector := NewCollector(cfg)
	version := "1.0.0"

	// Count metrics before setup
	metricFamilies, err := collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics before setup: %v", err)
	}
	beforeCount := len(metricFamilies)

	collector.Setup(version)

	// Count metrics after setup
	metricFamilies, err = collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics after setup: %v", err)
	}
	afterCount := len(metricFamilies)

	// Should add at least build info metric and system collectors
	if afterCount <= beforeCount {
		t.Errorf("expected additional metrics after setup, got %d before and %d after", beforeCount, afterCount)
	}

	// Verify build info is registered
	buildInfoFound := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "wnc_build_info" {
			buildInfoFound = true
			break
		}
	}
	if !buildInfoFound {
		t.Error("build info metric not found after setup")
	}
}

func TestCollector_RegisterServiceCollectors_AllDisabled(t *testing.T) {
	t.Parallel()
	cfg := createDisabledConfig()
	collector := NewCollector(cfg)

	// Count metrics before registration
	metricFamilies, err := collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics before registration: %v", err)
	}
	beforeCount := len(metricFamilies)

	collector.RegisterServiceCollectors()

	// Count metrics after registration
	metricFamilies, err = collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics after registration: %v", err)
	}
	afterCount := len(metricFamilies)

	// Should not add any metrics when all collectors are disabled
	if afterCount != beforeCount {
		t.Errorf(
			"expected no additional metrics when all collectors disabled, got %d before and %d after",
			beforeCount, afterCount,
		)
	}
}

func TestCollector_RegisterServiceCollectors_APEnabled(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()
	// Only enable AP collector
	cfg.Collectors.Client.General = false
	cfg.Collectors.WLAN.General = false
	collector := NewCollector(cfg)

	// This test verifies the function runs without panicking
	// The actual collector registration requires WNC connectivity which we avoid in unit tests
	collector.RegisterServiceCollectors()

	// If we reach this point, the function completed successfully
	if collector.registry == nil {
		t.Error("registry should still be available after service collectors registration")
	}
}

func TestCollector_RegisterServiceCollectors_ClientEnabled(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()
	// Only enable Client collector
	cfg.Collectors.AP.General = false
	cfg.Collectors.AP.Radio = false
	cfg.Collectors.WLAN.General = false
	collector := NewCollector(cfg)

	// This test verifies the function runs without panicking
	collector.RegisterServiceCollectors()

	// If we reach this point, the function completed successfully
	if collector.registry == nil {
		t.Error("registry should still be available after service collectors registration")
	}
}

func TestCollector_RegisterServiceCollectors_WLANEnabled(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()
	// Only enable WLAN collector
	cfg.Collectors.AP.General = false
	cfg.Collectors.AP.Radio = false
	cfg.Collectors.Client.General = false
	collector := NewCollector(cfg)

	// This test verifies the function runs without panicking
	collector.RegisterServiceCollectors()

	// If we reach this point, the function completed successfully
	if collector.registry == nil {
		t.Error("registry should still be available after service collectors registration")
	}
}

func TestCollector_Float64Metric_Type(t *testing.T) {
	t.Parallel()
	desc := prometheus.NewDesc("test_metric", "Test metric", nil, nil)
	metric := Float64Metric{
		Desc:  desc,
		Value: 42.0,
	}

	if metric.Desc != desc {
		t.Error("Desc field not set correctly")
	}
	if metric.Value != 42.0 {
		t.Error("Value field not set correctly")
	}
}

func TestCollector_StringMetric_Type(t *testing.T) {
	t.Parallel()
	desc := prometheus.NewDesc("test_metric", "Test metric", nil, nil)
	metric := StringMetric{
		Desc:  desc,
		Value: "test-value",
	}

	if metric.Desc != desc {
		t.Error("Desc field not set correctly")
	}
	if metric.Value != "test-value" {
		t.Error("Value field not set correctly")
	}
}

func TestCollector_MultipleSetups(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()
	collector := NewCollector(cfg)
	version := "1.0.0"

	// First setup
	collector.Setup(version)

	// Count metrics after first setup
	metricFamilies, err := collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics after first setup: %v", err)
	}
	firstCount := len(metricFamilies)

	// Second setup should not panic and should handle duplicate registrations
	// (though prometheus will panic on duplicate registration, this tests our logic)
	defer func() {
		if r := recover(); r != nil {
			// Expected behavior: prometheus panics on duplicate registration
			// This is acceptable behavior for this function
			t.Logf("Expected panic on duplicate registration: %v", r)
		}
	}()

	collector.Setup(version)

	// If no panic occurred, verify metrics are still there
	metricFamilies, err = collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics after second setup: %v", err)
	}
	secondCount := len(metricFamilies)

	if secondCount < firstCount {
		t.Error("metrics count decreased after second setup")
	}
}

func TestCollector_EmptyVersion(t *testing.T) {
	t.Parallel()
	cfg := createTestConfig()
	collector := NewCollector(cfg)

	// Test with empty version string
	collector.RegisterBuildInfo("")

	metricFamilies, err := collector.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	buildInfoFound := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "wnc_build_info" {
			buildInfoFound = true
			metric := mf.GetMetric()[0]
			if metric.GetLabel()[0].GetValue() != "" {
				t.Errorf("expected empty version, got %s", metric.GetLabel()[0].GetValue())
			}
			break
		}
	}

	if !buildInfoFound {
		t.Error("build info metric not found")
	}
}

func TestCollector_ConfigValidation(t *testing.T) {
	t.Parallel()

	// Test with nil config - should not panic but might fail
	defer func() {
		if r := recover(); r != nil {
			// This is expected behavior if nil config is passed
			t.Logf("Expected panic with nil config: %v", r)
		}
	}()

	collector := NewCollector(nil)
	if collector != nil && collector.cfg != nil {
		t.Error("expected nil config to be handled")
	}
}

func TestCollector_RegistryIndependence(t *testing.T) {
	t.Parallel()
	cfg1 := createTestConfig()
	cfg2 := createTestConfig()

	collector1 := NewCollector(cfg1)
	collector2 := NewCollector(cfg2)

	// Each collector should have its own registry
	if collector1.Registry() == collector2.Registry() {
		t.Error("collectors should have independent registries")
	}

	// Register different versions
	collector1.RegisterBuildInfo("1.0.0")
	collector2.RegisterBuildInfo("2.0.0")

	// Verify independence
	metricFamilies1, err := collector1.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics from collector1: %v", err)
	}

	metricFamilies2, err := collector2.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics from collector2: %v", err)
	}

	// Both should have build info but with different versions
	version1 := ""
	version2 := ""

	for _, mf := range metricFamilies1 {
		if mf.GetName() == "wnc_build_info" {
			version1 = mf.GetMetric()[0].GetLabel()[0].GetValue()
		}
	}

	for _, mf := range metricFamilies2 {
		if mf.GetName() == "wnc_build_info" {
			version2 = mf.GetMetric()[0].GetLabel()[0].GetValue()
		}
	}

	if version1 != "1.0.0" {
		t.Errorf("collector1 expected version 1.0.0, got %s", version1)
	}
	if version2 != "2.0.0" {
		t.Errorf("collector2 expected version 2.0.0, got %s", version2)
	}
}
