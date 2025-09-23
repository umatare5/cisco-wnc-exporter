// Package log provides structured logging setup using Go's slog package.
package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// Setup configures and returns a slog.Logger based on the provided configuration.
// It sets up the appropriate handler (JSON or text) and log level.
func Setup(cfg config.Log) *slog.Logger {
	var handler slog.Handler

	// Configure handler based on format
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Level),
	}

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		// Fallback to JSON handler
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// parseLogLevel converts string log level to slog.Level.
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
