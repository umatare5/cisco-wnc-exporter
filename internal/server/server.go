// Package server provides HTTP server functionality for the cisco-wnc-exporter.
// It serves Prometheus metrics and health check endpoints.
package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// New creates a new HTTP server with /metrics, /healthz, and /readyz endpoints.
// It uses a custom Prometheus registry and enables OpenMetrics format.
func New(reg *prometheus.Registry, addr string) *http.Server {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint with custom registry
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics:   true,
		MaxRequestsInFlight: 10, // Limit concurrent metric requests
	}))

	// Health check endpoints
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n")) // Ignore write error for health check
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n")) // Ignore write error for readiness check
	})

	// Root endpoint with basic information
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html>
<head><title>Cisco WNC Exporter</title></head>
<body>
<h1>Cisco WNC Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="/healthz">Health Check</a></p>
<p><a href="/readyz">Readiness Check</a></p>
</body>
</html>`))
	})

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second, // Prevent Slowloris attacks
	}
}
