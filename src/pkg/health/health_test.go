// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubHealthClient struct {
	err error
}

func (c stubHealthClient) Health(context.Context) error { return c.err }

type blockingHealthClient struct{}

func (blockingHealthClient) Health(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestNewHealthChecker(t *testing.T) {
	checker := NewHealthChecker("1.2.3", 0)
	if checker.version != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", checker.version)
	}
	if checker.timeout != 5*time.Second {
		t.Fatalf("timeout = %s, want 5s", checker.timeout)
	}
	if checker.databases == nil {
		t.Fatal("database map was not initialized")
	}

	custom := NewHealthChecker("dev", 25*time.Millisecond)
	if custom.timeout != 25*time.Millisecond {
		t.Fatalf("custom timeout = %s, want 25ms", custom.timeout)
	}
}

func TestCheckHealthWithoutDatabases(t *testing.T) {
	checker := NewHealthChecker("test", time.Second)
	before := time.Now().Unix()
	status := checker.CheckHealth(context.Background())
	after := time.Now().Unix()

	if status.Status != "healthy" {
		t.Fatalf("status = %q, want healthy", status.Status)
	}
	if status.Version != "test" {
		t.Fatalf("version = %q, want test", status.Version)
	}
	if status.Databases["none"] != "no databases configured" {
		t.Fatalf("unexpected database status: %#v", status.Databases)
	}
	if status.Timestamp < before || status.Timestamp > after {
		t.Fatalf("timestamp %d is outside [%d, %d]", status.Timestamp, before, after)
	}
}

func TestCheckHealthAggregatesDatabaseResults(t *testing.T) {
	checker := NewHealthChecker("test", time.Second)
	checker.RegisterDatabase("healthy", stubHealthClient{})
	checker.RegisterDatabase("broken", stubHealthClient{err: errors.New("offline")})

	status := checker.CheckHealth(context.Background())
	if status.Status != "degraded" {
		t.Fatalf("status = %q, want degraded", status.Status)
	}
	if status.Databases["healthy"] != "healthy" {
		t.Errorf("healthy database = %q", status.Databases["healthy"])
	}
	if status.Databases["broken"] != "unhealthy" {
		t.Errorf("broken database = %q", status.Databases["broken"])
	}
}

func TestCheckHealthEnforcesTimeout(t *testing.T) {
	checker := NewHealthChecker("test", time.Millisecond)
	checker.RegisterDatabase("slow", blockingHealthClient{})

	status := checker.CheckHealth(context.Background())
	if status.Status != "degraded" || status.Databases["slow"] != "unhealthy" {
		t.Fatalf("unexpected timeout status: %#v", status)
	}
}

func TestHealthHandlers(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		handler    func(*HealthChecker) http.HandlerFunc
		degraded   bool
		method     string
		wantStatus int
	}{
		{name: "health healthy", path: "/healthz", handler: (*HealthChecker).HealthzHandler, method: http.MethodGet, wantStatus: http.StatusOK},
		{name: "health degraded", path: "/healthz", handler: (*HealthChecker).HealthzHandler, degraded: true, method: http.MethodGet, wantStatus: http.StatusServiceUnavailable},
		{name: "health rejects post", path: "/healthz", handler: (*HealthChecker).HealthzHandler, method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed},
		{name: "ready healthy", path: "/readyz", handler: (*HealthChecker).ReadyzHandler, method: http.MethodGet, wantStatus: http.StatusOK},
		{name: "ready degraded", path: "/readyz", handler: (*HealthChecker).ReadyzHandler, degraded: true, method: http.MethodGet, wantStatus: http.StatusServiceUnavailable},
		{name: "ready rejects post", path: "/readyz", handler: (*HealthChecker).ReadyzHandler, method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewHealthChecker("test", time.Second)
			if tt.degraded {
				checker.RegisterDatabase("broken", stubHealthClient{err: errors.New("offline")})
			}

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)
			tt.handler(checker).ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if tt.method != http.MethodGet {
				return
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}

			var status HealthStatus
			if err := json.NewDecoder(recorder.Body).Decode(&status); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if status.Version != "test" {
				t.Fatalf("version = %q, want test", status.Version)
			}
		})
	}
}
