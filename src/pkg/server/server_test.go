// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAppliesDefaultsAndServesMetrics(t *testing.T) {
	config := &Config{Port: 9091}
	server := New(config)

	if config.ReadTimeout != 10*time.Second || config.WriteTimeout != 10*time.Second {
		t.Fatalf("default timeouts = %s/%s, want 10s/10s", config.ReadTimeout, config.WriteTimeout)
	}
	if server.port != 9091 || server.httpServer.Addr != ":9091" {
		t.Fatalf("server address = %q, port = %d", server.httpServer.Addr, server.port)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	server.GetHandler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got == "" {
		t.Fatal("metrics response has no Content-Type")
	}
}

func TestNewPreservesCustomTimeoutsAndAllowsHandlers(t *testing.T) {
	config := &Config{
		Port:         8081,
		ReadTimeout:  time.Second,
		WriteTimeout: 2 * time.Second,
	}
	server := New(config)
	if server.httpServer.ReadTimeout != time.Second || server.httpServer.WriteTimeout != 2*time.Second {
		t.Fatalf("custom timeouts were not preserved")
	}

	server.GetHandler().HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	server.GetHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("custom handler status = %d, want 204", recorder.Code)
	}
}

func TestStopBeforeStart(t *testing.T) {
	server := New(&Config{})
	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() before Start() returned %v", err)
	}
}
