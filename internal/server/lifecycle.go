// Package server provides HTTP server lifecycle management for Prometheus exporters.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-wnc-exporter/internal/collector"
	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// LifecycleManager manages HTTP server startup and graceful shutdown.
type LifecycleManager struct {
	server *http.Server
	cfg    *config.Config
}

// NewLifecycleManager creates a new server lifecycle manager.
func NewLifecycleManager(registry *prometheus.Registry, cfg *config.Config) *LifecycleManager {
	addr := net.JoinHostPort(cfg.Web.ListenAddress, strconv.Itoa(cfg.Web.ListenPort))
	server := New(registry, addr)

	return &LifecycleManager{
		server: server,
		cfg:    cfg,
	}
}

// StartAndServe creates collectors, sets up the server, and starts serving.
// It handles the complete server lifecycle from setup to shutdown.
func StartAndServe(ctx context.Context, cfg *config.Config, version string) error {
	slog.Info("Starting cisco-wnc-exporter",
		"version", version,
		"listen_address", cfg.Web.ListenAddress,
		"listen_port", cfg.Web.ListenPort)

	// Create and setup collector manager
	collectorMgr := collector.NewCollector(cfg)
	collectorMgr.Setup(version)

	// Create and run server lifecycle manager
	serverMgr := NewLifecycleManager(collectorMgr.Registry(), cfg)
	return serverMgr.Run(ctx)
}

// Run starts the HTTP server and handles graceful shutdown.
// It blocks until the server is shut down or an error occurs.
func (lm *LifecycleManager) Run(ctx context.Context) error {
	// Setup graceful shutdown context
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "addr", lm.server.Addr)
		if err := lm.server.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		slog.Info("Shutdown signal received")
	case err := <-errCh:
		return err
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := lm.server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Failed to shutdown HTTP server gracefully", "error", err)
		return err
	}

	slog.Info("HTTP server shutdown complete")
	return nil
}
