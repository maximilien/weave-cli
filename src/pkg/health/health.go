// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/logging"
)

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Version   string            `json:"version"`
	Databases map[string]string `json:"databases"`
}

// HealthChecker manages health checks for configured VDBs
type HealthChecker struct {
	databases map[string]DatabaseHealthClient
	version   string
	timeout   time.Duration
}

// DatabaseHealthClient is the minimal database capability required by health checks.
type DatabaseHealthClient interface {
	Health(context.Context) error
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string, timeout time.Duration) *HealthChecker {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &HealthChecker{
		databases: make(map[string]DatabaseHealthClient),
		version:   version,
		timeout:   timeout,
	}
}

// RegisterDatabase registers a VDB client for health checking
func (hc *HealthChecker) RegisterDatabase(name string, client DatabaseHealthClient) {
	hc.databases[name] = client
}

// CheckHealth checks the health of all registered databases
func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
	ctx, cancel := context.WithTimeout(ctx, hc.timeout)
	defer cancel()

	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   hc.version,
		Databases: make(map[string]string),
	}

	if len(hc.databases) == 0 {
		status.Status = "healthy"
		status.Databases["none"] = "no databases configured"
		return status
	}

	// Check all databases concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	hasUnhealthy := false

	for name, client := range hc.databases {
		wg.Add(1)
		go func(dbName string, dbClient DatabaseHealthClient) {
			defer wg.Done()

			err := dbClient.Health(ctx)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				status.Databases[dbName] = "unhealthy"
				hasUnhealthy = true
			} else {
				status.Databases[dbName] = "healthy"
			}
		}(name, client)
	}

	wg.Wait()

	if hasUnhealthy {
		status.Status = "degraded"
	}

	return status
}

// HealthzHandler returns an HTTP handler for /healthz endpoint
func (hc *HealthChecker) HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.WithFields(map[string]interface{}{
			"endpoint": "/healthz",
			"method":   r.Method,
		})

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := hc.CheckHealth(r.Context())

		w.Header().Set("Content-Type", "application/json")

		// Return 503 if any database is unhealthy, 200 otherwise
		if status.Status == "degraded" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			logger.Error("Failed to encode health status: %v", err)
		}
	}
}

// ReadyzHandler returns an HTTP handler for /readyz endpoint
// This is similar to /healthz but can have different logic for Kubernetes readiness probes
func (hc *HealthChecker) ReadyzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.WithFields(map[string]interface{}{
			"endpoint": "/readyz",
			"method":   r.Method,
		})

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := hc.CheckHealth(r.Context())

		w.Header().Set("Content-Type", "application/json")

		// For readiness, be more strict - return 503 if ANY database is unhealthy
		if status.Status != "healthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			logger.Error("Failed to encode readiness status: %v", err)
		}
	}
}
