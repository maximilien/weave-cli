// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/maximilien/weave-cli/src/pkg/logging"
)

// Server manages the HTTP server for metrics and health endpoints
type Server struct {
	httpServer *http.Server
	port       int
}

// Config holds server configuration
type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// New creates a new server instance
func New(config *Config) *Server {
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 10 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}

	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health endpoints will be registered separately
	// This allows the health package to register its own handlers

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Server{
		httpServer: httpServer,
		port:       config.Port,
	}
}

// Start starts the HTTP server in a goroutine
func (s *Server) Start() error {
	go func() {
		logging.Info("Starting metrics server on :%d", s.port)
		logging.Info("Metrics available at http://localhost:%d/metrics", s.port)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("Metrics server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	logging.Info("Shutting down metrics server...")
	return s.httpServer.Shutdown(ctx)
}

// GetHandler returns the server's HTTP handler for registering additional endpoints
func (s *Server) GetHandler() *http.ServeMux {
	return s.httpServer.Handler.(*http.ServeMux)
}
