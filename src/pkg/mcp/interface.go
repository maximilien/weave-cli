// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"fmt"
	"time"
)

// MCPClient provides a unified interface for MCP protocol clients
// Supports both HTTP and stdio transport protocols
type MCPClient interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Tool discovery
	ListTools(ctx context.Context) ([]Tool, error)

	// Tool invocation
	CallTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error)

	// Configuration
	Config() *Config
}

// Config configures MCP client connection
type Config struct {
	Name        string            `yaml:"name" json:"name"`                                   // Server name
	ServerURL   string            `yaml:"url" json:"url"`                                     // HTTP URL or stdio command
	Transport   Transport         `yaml:"transport" json:"transport"`                         // http or stdio
	Auth        *Auth             `yaml:"auth,omitempty" json:"auth,omitempty"`               // Authentication config
	Timeout     time.Duration     `yaml:"timeout" json:"timeout"`                             // Request timeout
	MaxRetries  int               `yaml:"max_retries" json:"max_retries"`                     // Retry attempts
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`         // Custom headers (HTTP only)
	Enabled     bool              `yaml:"enabled" json:"enabled"`                             // Whether server is enabled
	Tools       []string          `yaml:"tools,omitempty" json:"tools,omitempty"`             // Available tools (optional)
	Description string            `yaml:"description,omitempty" json:"description,omitempty"` // Server description
}

// Transport represents the transport protocol
type Transport string

const (
	TransportHTTP  Transport = "http"
	TransportStdio Transport = "stdio"
)

// Auth contains authentication credentials
type Auth struct {
	Type     string            `yaml:"type" json:"type"`         // bearer, basic, api-key
	Token    string            `yaml:"token" json:"token"`       // Bearer token
	Username string            `yaml:"username" json:"username"` // Basic auth username
	Password string            `yaml:"password" json:"password"` // Basic auth password
	APIKey   string            `yaml:"api_key" json:"api_key"`   // API key
	Headers  map[string]string `yaml:"headers" json:"headers"`   // Custom auth headers
}

// NewMCPClient creates a new MCP client based on transport type
func NewMCPClient(config *Config) (MCPClient, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	switch config.Transport {
	case TransportHTTP:
		return NewHTTPClient(config)
	case TransportStdio:
		return NewStdioClient(config)
	default:
		return nil, fmt.Errorf("unsupported transport: %s", config.Transport)
	}
}

// ValidateConfig validates MCP configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}

	if config.Transport != TransportHTTP && config.Transport != TransportStdio {
		return fmt.Errorf("transport must be 'http' or 'stdio'")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	return nil
}
