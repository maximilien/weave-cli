// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package serve

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/maximilien/weave-cli/src/pkg/health"
	"github.com/maximilien/weave-cli/src/pkg/logging"
	"github.com/maximilien/weave-cli/src/pkg/server"
	"github.com/maximilien/weave-cli/src/pkg/version"
)

var (
	metricsPort int
)

// ServeCmd represents the serve command
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start metrics and health endpoints server",
	Long: `Start a long-running HTTP server for Prometheus metrics and health endpoints.

This command starts an HTTP server that exposes:
  • /metrics  - Prometheus metrics endpoint
  • /healthz  - Liveness probe for Kubernetes
  • /readyz   - Readiness probe for Kubernetes

The server runs until interrupted (Ctrl+C) and is designed for production deployments
where you want persistent metrics collection and health monitoring.

Examples:
  # Start server on default port 9090
  weave serve

  # Start server on custom port
  weave serve --metrics-port 8080

  # With custom logging
  weave serve --log-format json --log-level info --log-file /var/log/weave.log

  # Full production setup
  weave serve --metrics-port 9090 --log-format json --log-file /var/log/weave.log

The server will gracefully shut down on SIGINT/SIGTERM signals.`,
	Run: runServe,
}

func init() {
	ServeCmd.Flags().IntVar(&metricsPort, "metrics-port", 9090, "port for metrics and health endpoints")
}

func runServe(cmd *cobra.Command, args []string) {
	logging.Info("Starting Weave metrics server...")
	logging.Info("Version: %s", version.Get().Version)

	// Create server
	srv := server.New(&server.Config{
		Port:         metricsPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Create health checker
	healthChecker := health.NewHealthChecker(version.Get().Version, 5*time.Second)

	// TODO: Register VDB clients with health checker
	// This would require loading config and initializing VDB clients
	// For now, health check will report "no databases configured"

	// Register health endpoints
	mux := srv.GetHandler()
	mux.HandleFunc("/healthz", healthChecker.HealthzHandler())
	mux.HandleFunc("/readyz", healthChecker.ReadyzHandler())

	// Start server
	if err := srv.Start(); err != nil {
		logging.Error("Failed to start metrics server: %v", err)
		os.Exit(1)
	}

	logging.Info("✅ Metrics server running on :%d", metricsPort)
	logging.Info("   • Metrics:  http://localhost:%d/metrics", metricsPort)
	logging.Info("   • Health:   http://localhost:%d/healthz", metricsPort)
	logging.Info("   • Ready:    http://localhost:%d/readyz", metricsPort)
	logging.Info("")
	logging.Info("Press Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	logging.Info("")
	logging.Info("Shutting down gracefully...")

	// Graceful shutdown with 10s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logging.Error("Error during shutdown: %v", err)
		os.Exit(1)
	}

	logging.Info("Server stopped")
}
