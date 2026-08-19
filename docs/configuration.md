# Configuration

A verbatim `cisco-wnc-exporter --help` transcript: every flag, its default, and the environment variable it reads where one exists.

## Flags

```bash
NAME:
   cisco-wnc-exporter - Prometheus exporter for Cisco WNC

USAGE:
   cisco-wnc-exporter [global options]

VERSION:
   0.11.0

GLOBAL OPTIONS:
   --dry-run                    Validate configuration without starting the server
   --help, -h                   show help
   --log.format string          Log format (json, text) (default: "json")
   --log.level string           Log level (debug, info, warn, error) (default: "info")
   --version, -v                print the version
   --web.listen-address string  Address to bind the HTTP server to (default: "0.0.0.0")
   --web.listen-port int        Port number to bind the HTTP server to (default: 10039)
   --web.telemetry-path string  Path for the metrics endpoint (default: "/metrics")
   --wnc.access-token string    WNC API access token [$WNC_ACCESS_TOKEN]
   --wnc.cache-ttl duration     Minimum interval between WNC data refreshes (default: 55s)
   --wnc.controller string      WNC controller hostname or IP address [$WNC_CONTROLLER]
   --wnc.timeout duration       WNC API request timeout (default: 55s)
   --wnc.tls-skip-verify        Skip TLS certificate verification

   # AP Collector Options

   --collector.ap.errors              Enable AP error metrics
   --collector.ap.general             Enable AP general metrics
   --collector.ap.info                Enable AP info metrics
   --collector.ap.info-labels string  Comma-separated list of AP info labels (default: "name,ip")
   --collector.ap.join                Enable AP CAPWAP join metrics
   --collector.ap.radio               Enable AP radio metrics
   --collector.ap.spectrum            Enable AP CleanAir spectrum metrics
   --collector.ap.traffic             Enable AP traffic metrics

   # Client Collector Options

   --collector.client.errors              Enable Client error metrics
   --collector.client.general             Enable Client general metrics
   --collector.client.info                Enable Client info metrics
   --collector.client.info-labels string  Comma-separated list of Client info labels (default: "name,ipv4")
   --collector.client.radio               Enable Client radio metrics
   --collector.client.traffic             Enable Client traffic metrics

   # Controller Collector Options

   --collector.controller.general  Enable Controller general metrics

   # WLAN Collector Options

   --collector.wlan.config              Enable WLAN config metrics
   --collector.wlan.general             Enable WLAN general metrics
   --collector.wlan.info                Enable WLAN info metrics
   --collector.wlan.info-labels string  Comma-separated list of WLAN info labels (default: "name")
   --collector.wlan.traffic             Enable WLAN traffic metrics

   * Collector Wide Options

   --collector.info-cache-ttl duration  Cache TTL for collector info metrics (default: 30m0s)

   * Internal Collector Options

   --collector.internal.go-runtime  Enable Go runtime metrics collector
   --collector.internal.process     Enable process metrics collector
```

## Notes

`--wnc.timeout` bounds a whole RESTCONF request. It does not bound the wait for the response headers or the TLS handshake, which the SDK pins at 5 seconds and exposes no option for, so raising the flag does not make the exporter wait longer for a controller that is slow to begin answering. See [Data refresh and caching](README.md#data-refresh-and-caching) for what that failure looks like in the refresh metrics.
