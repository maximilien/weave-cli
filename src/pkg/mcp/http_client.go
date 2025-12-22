// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// HTTPClient implements MCPClient for HTTP transport
type HTTPClient struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
	requestID  int
	mu         sync.Mutex
	connected  bool
}

// NewHTTPClient creates a new HTTP MCP client
func NewHTTPClient(config *Config) (MCPClient, error) {
	if config.Transport != TransportHTTP {
		return nil, fmt.Errorf("HTTP client requires http transport")
	}

	client := &HTTPClient{
		config:  config,
		baseURL: config.ServerURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	return client, nil
}

// Connect establishes HTTP connection and performs handshake
func (h *HTTPClient) Connect(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Send initialize request
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      h.nextRequestIDLocked(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "weave-cli",
				"version": "0.8.2",
			},
		},
	}

	resp, err := h.callHTTP(ctx, initReq)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	// Send initialized notification
	notif := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	if err := h.sendNotification(ctx, notif); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	h.connected = true
	return nil
}

// Close closes the HTTP client
func (h *HTTPClient) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connected = false
	return nil
}

// Ping tests the connection
func (h *HTTPClient) Ping(ctx context.Context) error {
	if !h.connected {
		return fmt.Errorf("client not connected")
	}

	// Test by listing tools
	_, err := h.ListTools(ctx)
	return err
}

// ListTools lists available tools
func (h *HTTPClient) ListTools(ctx context.Context) ([]Tool, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      h.nextRequestID(),
		Method:  "tools/list",
	}

	resp, err := h.callHTTP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	// Parse tools from response
	toolsData, ok := resp.Result["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tools response format")
	}

	var tools []Tool
	for _, t := range toolsData {
		toolMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		tool := Tool{
			Name:        getString(toolMap, "name"),
			Description: getString(toolMap, "description"),
		}

		if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
			tool.InputSchema = schema
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// CallTool calls an MCP tool
func (h *HTTPClient) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      h.nextRequestID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	resp, err := h.callHTTP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("call tool error: %s", resp.Error.Message)
	}

	// Return the content from the result
	if content, ok := resp.Result["content"]; ok {
		return content, nil
	}

	return resp.Result, nil
}

// Config returns the client configuration
func (h *HTTPClient) Config() *Config {
	return h.config
}

// callHTTP makes an HTTP request to the MCP server
func (h *HTTPClient) callHTTP(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
	// Marshal request
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	endpoint := h.baseURL + "/mcp/v1/rpc"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Add auth headers
	if h.config.Auth != nil {
		h.addAuthHeaders(httpReq)
	}

	// Add custom headers
	for k, v := range h.config.Headers {
		httpReq.Header.Set(k, v)
	}

	// Make request
	httpResp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respData))
	}

	// Parse response
	var mcpResp MCPResponse
	if err := json.Unmarshal(respData, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &mcpResp, nil
}

// sendNotification sends a notification (no response expected)
func (h *HTTPClient) sendNotification(ctx context.Context, req MCPRequest) error {
	reqData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	endpoint := h.baseURL + "/mcp/v1/rpc"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Notifications don't expect responses, just check status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// addAuthHeaders adds authentication headers to the request
func (h *HTTPClient) addAuthHeaders(req *http.Request) {
	if h.config.Auth == nil {
		return
	}

	auth := h.config.Auth

	switch auth.Type {
	case "bearer":
		if auth.Token != "" {
			req.Header.Set("Authorization", "Bearer "+auth.Token)
		}
	case "basic":
		if auth.Username != "" && auth.Password != "" {
			req.SetBasicAuth(auth.Username, auth.Password)
		}
	case "api-key":
		if auth.APIKey != "" {
			req.Header.Set("X-API-Key", auth.APIKey)
		}
	}

	// Add custom auth headers
	for k, v := range auth.Headers {
		req.Header.Set(k, v)
	}
}

// nextRequestID generates the next request ID (thread-safe)
func (h *HTTPClient) nextRequestID() *int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.nextRequestIDLocked()
}

// nextRequestIDLocked generates the next request ID (must hold lock)
func (h *HTTPClient) nextRequestIDLocked() *int {
	h.requestID++
	id := h.requestID
	return &id
}
