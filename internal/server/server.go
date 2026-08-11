// Package server provides HTTP server functionality.
package server

import (
	"html"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// New creates a new HTTP server with metrics and health endpoints. Config.Validate
// rejects every telemetryPath that http.ServeMux would panic on, apart from the root,
// which is handled below.
func New(reg *prometheus.Registry, addr, telemetryPath string) *http.Server {
	mux := http.NewServeMux()

	mux.Handle(telemetryPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics:   true,
		MaxRequestsInFlight: 10,
	}))

	mux.HandleFunc(config.HealthPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
	})

	// Serving metrics at the root is a legitimate flag value, and registering the
	// landing page as well would panic on the duplicate pattern.
	if telemetryPath != "/" {
		landing := []byte(`<html>
<head><title>Cisco WNC Exporter</title></head>
<body>
<h1>Cisco WNC Exporter</h1>
<p><a href="` + html.EscapeString(telemetryPath) + `">Metrics</a></p>
<p><a href="` + config.HealthPath + `">Health Check</a></p>
</body>
</html>`)

		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(landing)
		})
	}

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}
}
