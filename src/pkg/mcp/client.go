// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Client is an MCP client that communicates with weave-mcp via stdio
type Client struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	reader     *bufio.Reader
	mu         sync.Mutex
	requestID  int
	tools      []Tool
	stderrBuf  []byte
	stderrDone chan struct{}
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPRequest represents a request to the MCP server
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      *int                   `json:"id,omitempty"` // Optional for notifications
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// MCPResponse represents a response from the MCP server
type MCPResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *MCPError              `json:"error,omitempty"`
}

// MCPError represents an error from the MCP server
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewClient creates a new MCP client
func NewClient(stdioPath string) (*Client, error) {
	cmd := exec.Command(stdioPath)

	// Try to find project root by looking for .env file
	// This ensures the MCP server can find configuration files
	if workDir := findProjectRoot(); workDir != "" {
		cmd.Dir = workDir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	client := &Client{
		cmd:        cmd,
		stdin:      stdin,
		stdout:     stdout,
		stderr:     stderr,
		reader:     bufio.NewReader(stdout),
		requestID:  0,
		stderrDone: make(chan struct{}),
	}

	// Start goroutine to capture stderr
	go client.captureStderr()

	// Initialize connection
	if err := client.initialize(); err != nil {
		_ = client.Close()
		// Include stderr in error message if available
		stderrMsg := string(client.stderrBuf)
		if stderrMsg != "" {
			return nil, fmt.Errorf("failed to initialize MCP connection: %w\nweave-mcp stderr: %s", err, stderrMsg)
		}
		return nil, fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	return client, nil
}

// initialize performs the MCP handshake
func (c *Client) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send initialize request
	id := c.nextRequestID()
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "weave-cli",
				"version": "0.3.0",
			},
		},
	}

	if err := c.sendRequest(req); err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	// Read initialize response
	response, err := c.readResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to read initialize response: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("initialize error: %s", response.Error.Message)
	}

	// Send initialized notification
	notification := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	if err := c.sendRequest(notification); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	return nil
}

// ListTools lists all available tools from the MCP server
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	id := c.nextRequestID()
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  "tools/list",
	}

	if err := c.sendRequest(req); err != nil {
		return nil, fmt.Errorf("failed to send list tools request: %w", err)
	}

	response, err := c.readResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read list tools response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", response.Error.Message)
	}

	// Parse tools from response
	toolsData, ok := response.Result["tools"].([]interface{})
	if !ok {
		// Provide detailed error for debugging
		resultJSON, err := json.Marshal(response.Result)
		if err != nil {
			return nil, fmt.Errorf("invalid tools response format: expected 'tools' array, result type: %T, marshal error: %v", response.Result["tools"], err)
		}
		return nil, fmt.Errorf("invalid tools response format: expected 'tools' array, got: %s", string(resultJSON))
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

	c.tools = tools
	return tools, nil
}

// CallTool calls an MCP tool with the given arguments
func (c *Client) CallTool(ctx context.Context, tool string, args map[string]interface{}) (interface{}, error) {
	id := c.nextRequestID()
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      tool,
			"arguments": args,
		},
	}

	if err := c.sendRequest(req); err != nil {
		return nil, fmt.Errorf("failed to send call tool request: %w", err)
	}

	response, err := c.readResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read call tool response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("call tool error: %s", response.Error.Message)
	}

	// Return the content from the result
	if content, ok := response.Result["content"]; ok {
		return content, nil
	}

	return response.Result, nil
}

// captureStderr captures stderr output from the MCP server
func (c *Client) captureStderr() {
	defer close(c.stderrDone)
	if c.stderr == nil {
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := c.stderr.Read(buf)
		if n > 0 {
			c.mu.Lock()
			c.stderrBuf = append(c.stderrBuf, buf[:n]...)
			c.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

// Close closes the MCP client connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.stdout != nil {
		_ = c.stdout.Close()
	}
	if c.stderr != nil {
		_ = c.stderr.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}

	// Wait for stderr capture to finish
	if c.stderrDone != nil {
		<-c.stderrDone
	}

	return nil
}

// sendRequest sends a request to the MCP server
func (c *Client) sendRequest(req MCPRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write JSON-RPC request with newline delimiter
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// readResponse reads a response from the MCP server
func (c *Client) readResponse(ctx context.Context) (*MCPResponse, error) {
	// Create a channel for the response
	respChan := make(chan *MCPResponse, 1)
	errChan := make(chan error, 1)

	go func() {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			errChan <- fmt.Errorf("failed to read response: %w", err)
			return
		}

		var response MCPResponse
		if err := json.Unmarshal(line, &response); err != nil {
			errChan <- fmt.Errorf("failed to unmarshal response: %w", err)
			return
		}

		respChan <- &response
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case response := <-respChan:
		return response, nil
	}
}

// nextRequestID generates the next request ID
func (c *Client) nextRequestID() int {
	c.requestID++
	return c.requestID
}

// getString safely gets a string value from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// findProjectRoot tries to find the project root by looking for .env or go.mod
// This ensures the MCP server can find configuration files when started from subdirectories
func findProjectRoot() string {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up the directory tree looking for .env or go.mod
	for {
		// Check if .env exists in this directory
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return dir
		}

		// Check if go.mod exists in this directory
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return ""
}
