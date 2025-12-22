// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"fmt"
)

// StdioClient wraps the existing Client to implement MCPClient interface
type StdioClient struct {
	config *Config
	client *Client
}

// NewStdioClient creates a new stdio MCP client
func NewStdioClient(config *Config) (MCPClient, error) {
	if config.Transport != TransportStdio {
		return nil, fmt.Errorf("stdio client requires stdio transport")
	}

	client, err := NewClient(config.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdio client: %w", err)
	}

	return &StdioClient{
		config: config,
		client: client,
	}, nil
}

// Connect establishes connection (already done in NewClient)
func (s *StdioClient) Connect(ctx context.Context) error {
	// Connection is already established in NewClient
	return nil
}

// Close closes the connection
func (s *StdioClient) Close() error {
	return s.client.Close()
}

// Ping tests the connection
func (s *StdioClient) Ping(ctx context.Context) error {
	// Test by listing tools
	_, err := s.client.ListTools(ctx)
	return err
}

// ListTools lists available tools
func (s *StdioClient) ListTools(ctx context.Context) ([]Tool, error) {
	return s.client.ListTools(ctx)
}

// CallTool calls an MCP tool
func (s *StdioClient) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	return s.client.CallTool(ctx, toolName, args)
}

// Config returns the client configuration
func (s *StdioClient) Config() *Config {
	return s.config
}
