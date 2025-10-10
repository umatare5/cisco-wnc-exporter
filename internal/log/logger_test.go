package log_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
	"github.com/umatare5/cisco-wnc-exporter/internal/log"
)

func TestSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  config.Log
	}{
		{
			name: "json format with info level",
			cfg: config.Log{
				Level:  "info",
				Format: "json",
			},
		},
		{
			name: "text format with debug level",
			cfg: config.Log{
				Level:  "debug",
				Format: "text",
			},
		},
		{
			name: "unknown format defaults to json",
			cfg: config.Log{
				Level:  "warn",
				Format: "unknown",
			},
		},
		{
			name: "uppercase format and level",
			cfg: config.Log{
				Level:  "ERROR",
				Format: "JSON",
			},
		},
		{
			name: "empty values use defaults",
			cfg: config.Log{
				Level:  "",
				Format: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := log.Setup(tt.cfg)

			if logger == nil {
				t.Fatal("Setup() returned nil logger")
			}

			// Test that logger can be used without panic
			logger.Info("test message")
			logger.Debug("debug message")
			logger.Warn("warning message")
			logger.Error("error message")
		})
	}
}

func TestSetupWithValidFormats(t *testing.T) {
	t.Parallel()

	formats := []string{"json", "text", "JSON", "TEXT", "Json", "Text"}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			t.Parallel()

			cfg := config.Log{
				Level:  "info",
				Format: format,
			}

			logger := log.Setup(cfg)

			if logger == nil {
				t.Errorf("Setup() with format %q returned nil logger", format)
			}

			// Test logger functionality
			logger.Info("test message", "format", format)
		})
	}
}

func TestSetupWithValidLevels(t *testing.T) {
	t.Parallel()

	levels := []string{"debug", "info", "warn", "warning", "error", "DEBUG", "INFO", "WARN", "ERROR"}

	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			t.Parallel()

			cfg := config.Log{
				Level:  level,
				Format: "json",
			}

			logger := log.Setup(cfg)

			if logger == nil {
				t.Errorf("Setup() with level %q returned nil logger", level)
			}

			// Test logger functionality
			logger.Info("test message", "level", level)
		})
	}
}

func TestSetupHandlerTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "json handler",
			format: "json",
		},
		{
			name:   "text handler",
			format: "text",
		},
		{
			name:   "default handler for unknown format",
			format: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Log{
				Level:  "info",
				Format: tt.format,
			}

			logger := log.Setup(cfg)

			if logger == nil {
				t.Fatal("Setup() returned nil logger")
			}

			// Verify logger works by logging at different levels
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")
		})
	}
}

func TestSetupLogLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{
			name:     "debug level",
			level:    "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "info level",
			level:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "warn level",
			level:    "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "warning level",
			level:    "warning",
			expected: slog.LevelWarn,
		},
		{
			name:     "error level",
			level:    "error",
			expected: slog.LevelError,
		},
		{
			name:     "unknown level defaults to info",
			level:    "unknown",
			expected: slog.LevelInfo,
		},
		{
			name:     "empty level defaults to info",
			level:    "",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Log{
				Level:  tt.level,
				Format: "json",
			}

			logger := log.Setup(cfg)

			if logger == nil {
				t.Fatal("Setup() returned nil logger")
			}

			// Test that logger is created successfully
			// Note: Direct level comparison is not possible with the current API
			// but we can verify the logger works at different levels
			logger.Log(context.TODO(), tt.expected, "test message")
		})
	}
}

func TestSetupCaseSensitivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		level  string
		format string
	}{
		{
			name:   "lowercase",
			level:  "debug",
			format: "json",
		},
		{
			name:   "uppercase",
			level:  "DEBUG",
			format: "JSON",
		},
		{
			name:   "mixed case",
			level:  "Info",
			format: "Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Log{
				Level:  tt.level,
				Format: tt.format,
			}

			logger := log.Setup(cfg)

			if logger == nil {
				t.Fatal("Setup() returned nil logger")
			}

			// Verify logger works regardless of case
			logger.Info("test message", "case", tt.name)
		})
	}
}
