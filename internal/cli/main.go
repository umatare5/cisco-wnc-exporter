// Package cli provides the CLI implementation for the cisco-wnc-exporter.
package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
	"github.com/umatare5/cisco-wnc-exporter/internal/log"
	"github.com/umatare5/cisco-wnc-exporter/internal/server"
)

// NewApp creates a new CLI application.
func NewApp() *cli.Command {
	cmd := &cli.Command{
		Name:    "cisco-wnc-exporter",
		Usage:   "Prometheus exporter for Cisco WNC",
		Version: getVersion(),
		Flags:   registerFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := config.Parse(cmd)
			// Handle configuration parsing errors
			if err != nil {
				slog.Error("Configuration parsing failed", "error", err)
				return errors.New("configuration error")
			}

			// Setup logger
			slog.SetDefault(log.Setup(cfg.Log))

			// Handle dry-run mode early exit
			if cfg.DryRun {
				slog.Info("Configuration validation successful", "dry_run", true)
				return nil
			}

			// Start the server with the validated configuration
			return server.StartAndServe(ctx, cfg, getVersion())
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
	return cmd
}

// registerFlags defines and returns all CLI flags organized by category.
func registerFlags() []cli.Flag {
	flags := []cli.Flag{}
	flags = append(flags, registerWebFlags()...)
	flags = append(flags, registerWNCFlags()...)
	flags = append(flags, registerCollectorFlags()...)
	flags = append(flags, registerLogFlags()...)
	flags = append(flags, registerUtilityFlags()...)
	flags = append(flags, registerInternalCollectorFlags()...)
	flags = append(flags, registerAPCollectorFlags()...)
	flags = append(flags, registerClientCollectorFlags()...)
	flags = append(flags, registerWLANCollectorFlags()...)
	return flags
}

// registerWebFlags defines flags for HTTP server configuration.
func registerWebFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "web.listen-address",
			Usage: "Address to bind the HTTP server to",
			Value: config.DefaultListenAddress,
		},
		&cli.IntFlag{
			Name:  "web.listen-port",
			Usage: "Port number to bind the HTTP server to",
			Value: config.DefaultListenPort,
		},
		&cli.StringFlag{
			Name:  "web.telemetry-path",
			Usage: "Path for the metrics endpoint",
			Value: config.DefaultTelemetryPath,
		},
	}
}

// registerWNCFlags defines flags for WNC controller connection.
func registerWNCFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "wnc.controller",
			Usage:       "WNC controller hostname or IP address",
			Required:    true,
			Sources:     cli.EnvVars("WNC_HOST"),
			DefaultText: "<required>",
		},
		&cli.StringFlag{
			Name:        "wnc.access-token",
			Usage:       "WNC API access token",
			Required:    true,
			Sources:     cli.EnvVars("WNC_TOKEN"),
			DefaultText: "<required>",
		},
		&cli.DurationFlag{
			Name:  "wnc.timeout",
			Usage: "WNC API request timeout",
			Value: config.DefaultWNCTimeout,
		},
		&cli.DurationFlag{
			Name:  "wnc.cache-ttl",
			Usage: "WNC API response cache TTL",
			Value: config.DefaultWNCCacheTTL,
		},
		&cli.BoolFlag{
			Name:  "wnc.tls-skip-verify",
			Usage: "Skip TLS certificate verification",
		},
	}
}

// registerCollectorFlags defines flags for collector-wide configuration.
func registerCollectorFlags() []cli.Flag {
	return []cli.Flag{
		&cli.DurationFlag{
			Name:     "collector.cache-ttl",
			Usage:    "Cache TTL for collector metrics",
			Value:    config.DefaultCollectorCacheTTL,
			Category: "* Collector Wide Options",
		},
	}
}

// registerAPCollectorFlags defines flags for AP collector modules.
func registerAPCollectorFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "collector.ap.inventory",
			Usage:       "Enable AP inventory metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.ap.info",
			Usage:       "Enable AP info metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.StringFlag{
			Name:     "collector.ap.info-labels",
			Usage:    "Comma-separated list of AP info labels",
			Value:    config.DefaultAPInfoLabels,
			Category: "# AP Collector Options",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.BoolFlag{
			Name:        "collector.ap.state",
			Usage:       "Enable AP state metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.ap.phy",
			Usage:       "Enable AP physical layer metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.ap.rf",
			Usage:       "Enable AP RF environment metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.ap.traffic",
			Usage:       "Enable AP traffic metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.ap.errors",
			Usage:       "Enable AP error metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.ap.cpu",
			Usage:       "Enable AP CPU usage metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.ap.memory",
			Usage:       "Enable AP memory usage metrics",
			Category:    "# AP Collector Options",
			HideDefault: true,
		},
	}
}

// registerWLANCollectorFlags defines flags for WLAN collector modules.
func registerWLANCollectorFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "collector.wlan.inventory",
			Usage:       "Enable WLAN inventory collector",
			Category:    "# WLAN Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.wlan.info",
			Usage:       "Enable WLAN info metrics",
			Category:    "# WLAN Collector Options",
			HideDefault: true,
		},
		&cli.StringFlag{
			Name:     "collector.wlan.info-labels",
			Usage:    "Comma-separated list of WLAN info labels",
			Value:    config.DefaultWLANInfoLabels,
			Category: "# WLAN Collector Options",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.BoolFlag{
			Name:        "collector.wlan.state",
			Usage:       "Enable WLAN state collector",
			Category:    "# WLAN Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.wlan.traffic",
			Usage:       "Enable WLAN traffic collector",
			Category:    "# WLAN Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.wlan.security",
			Usage:       "Enable WLAN security collector",
			Category:    "# WLAN Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.wlan.networking",
			Usage:       "Enable WLAN networking collector",
			Category:    "# WLAN Collector Options",
			HideDefault: true,
		},
	}
}

// registerClientCollectorFlags defines flags for Client collector modules.
func registerClientCollectorFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "collector.client.inventory",
			Usage:       "Enable Client inventory collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.client.info",
			Usage:       "Enable Client info collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
		&cli.StringFlag{
			Name:     "collector.client.info-labels",
			Usage:    "Comma-separated list of Client info labels",
			Value:    config.DefaultClientInfoLabels,
			Category: "# Client Collector Options",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.BoolFlag{
			Name:        "collector.client.session",
			Usage:       "Enable Client session collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.client.phy",
			Usage:       "Enable Client PHY collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.client.rf",
			Usage:       "Enable Client RF collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.client.traffic",
			Usage:       "Enable Client traffic collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.client.errors",
			Usage:       "Enable Client errors collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.client.power",
			Usage:       "Enable Client power collector",
			Category:    "# Client Collector Options",
			HideDefault: true,
		},
	}
}

// registerLogFlags defines flags for logging configuration.
func registerLogFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "log.level",
			Usage: "Log level (debug, info, warn, error)",
			Value: config.DefaultLogLevel,
		},
		&cli.StringFlag{
			Name:  "log.format",
			Usage: "Log format (json, text)",
			Value: config.DefaultLogFormat,
		},
	}
}

// registerInternalCollectorFlags defines flags for internal metrics collection configuration.
func registerInternalCollectorFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "collector.internal.go-runtime",
			Usage:       "Enable Go runtime metrics collector",
			Category:    "* Internal Collector Options",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "collector.internal.process",
			Usage:       "Enable process metrics collector",
			Category:    "* Internal Collector Options",
			HideDefault: true,
		},
	}
}

// registerUtilityFlags defines utility flags.
func registerUtilityFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "Validate configuration without starting the server",
			HideDefault: true,
		},
	}
}
