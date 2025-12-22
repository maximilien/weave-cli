// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/mcp"
)

// TestMCPClientInterface tests the MCP client interface
func TestMCPClientInterface(t *testing.T) {
	t.Run("ValidateConfig", func(t *testing.T) {
		// Valid config
		validConfig := &mcp.Config{
			Name:      "test-server",
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			Timeout:   30 * time.Second,
		}

		err := mcp.ValidateConfig(validConfig)
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("ValidateConfigNil", func(t *testing.T) {
		err := mcp.ValidateConfig(nil)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("ValidateConfigMissingURL", func(t *testing.T) {
		config := &mcp.Config{
			Transport: mcp.TransportHTTP,
		}

		err := mcp.ValidateConfig(config)
		if err == nil {
			t.Error("Expected error for missing server URL")
		}
	})

	t.Run("ValidateConfigInvalidTransport", func(t *testing.T) {
		config := &mcp.Config{
			ServerURL: "http://localhost:8030",
			Transport: mcp.Transport("invalid"),
		}

		err := mcp.ValidateConfig(config)
		if err == nil {
			t.Error("Expected error for invalid transport")
		}
	})

	t.Run("ValidateConfigNegativeTimeout", func(t *testing.T) {
		config := &mcp.Config{
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			Timeout:   -1 * time.Second,
		}

		err := mcp.ValidateConfig(config)
		if err == nil {
			t.Error("Expected error for negative timeout")
		}
	})
}

// TestMCPClientFactory tests client creation
func TestMCPClientFactory(t *testing.T) {
	t.Run("CreateHTTPClient", func(t *testing.T) {
		config := &mcp.Config{
			Name:      "http-test",
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			Timeout:   30 * time.Second,
		}

		client, err := mcp.NewMCPClient(config)
		if err != nil {
			t.Fatalf("Failed to create HTTP client: %v", err)
		}

		if client == nil {
			t.Fatal("Client should not be nil")
		}

		defer client.Close()

		// Verify config
		if client.Config().Transport != mcp.TransportHTTP {
			t.Errorf("Expected HTTP transport, got %s", client.Config().Transport)
		}
	})

	t.Run("CreateInvalidTransport", func(t *testing.T) {
		config := &mcp.Config{
			Name:      "invalid-test",
			ServerURL: "http://localhost:8030",
			Transport: mcp.Transport("websocket"),
			Timeout:   30 * time.Second,
		}

		_, err := mcp.NewMCPClient(config)
		if err == nil {
			t.Error("Expected error for unsupported transport")
		}
	})
}

// TestMCPClientConfiguration tests configuration options
func TestMCPClientConfiguration(t *testing.T) {
	t.Run("DefaultTimeout", func(t *testing.T) {
		config := &mcp.Config{
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			// Timeout not set - should use default
		}

		client, err := mcp.NewMCPClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		if client.Config().Timeout != 30*time.Second {
			t.Errorf("Expected default timeout 30s, got %v", client.Config().Timeout)
		}
	})

	t.Run("CustomTimeout", func(t *testing.T) {
		customTimeout := 60 * time.Second
		config := &mcp.Config{
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			Timeout:   customTimeout,
		}

		client, err := mcp.NewMCPClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		if client.Config().Timeout != customTimeout {
			t.Errorf("Expected timeout %v, got %v", customTimeout, client.Config().Timeout)
		}
	})

	t.Run("DefaultMaxRetries", func(t *testing.T) {
		config := &mcp.Config{
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			// MaxRetries not set - should use default
		}

		client, err := mcp.NewMCPClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		if client.Config().MaxRetries != 3 {
			t.Errorf("Expected default max retries 3, got %d", client.Config().MaxRetries)
		}
	})

	t.Run("AuthConfiguration", func(t *testing.T) {
		config := &mcp.Config{
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			Auth: &mcp.Auth{
				Type:  "bearer",
				Token: "test-token-123",
			},
		}

		client, err := mcp.NewMCPClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		if client.Config().Auth == nil {
			t.Error("Auth configuration should not be nil")
		}

		if client.Config().Auth.Type != "bearer" {
			t.Errorf("Expected bearer auth, got %s", client.Config().Auth.Type)
		}
	})
}

// TestMCPClientErrorHandling tests error scenarios
func TestMCPClientErrorHandling(t *testing.T) {
	t.Run("InvalidServerURL", func(t *testing.T) {
		config := &mcp.Config{
			ServerURL: "http://invalid-server-that-does-not-exist:9999",
			Transport: mcp.TransportHTTP,
			Timeout:   5 * time.Second,
		}

		client, err := mcp.NewMCPClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Connection should fail
		err = client.Connect(ctx)
		if err == nil {
			t.Error("Expected connection error for invalid server")
		}
	})

	t.Run("TimeoutContext", func(t *testing.T) {
		config := &mcp.Config{
			ServerURL: "http://localhost:8030",
			Transport: mcp.TransportHTTP,
			Timeout:   30 * time.Second,
		}

		client, err := mcp.NewMCPClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		// Create already-cancelled context
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(10 * time.Millisecond) // Ensure context is expired

		// Operation should timeout
		err = client.Connect(ctx)
		if err == nil {
			t.Error("Expected timeout error for expired context")
		}
	})
}
