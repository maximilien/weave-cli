// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClientLifecycle(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp/v1/rpc" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var request MCPRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		methods = append(methods, request.Method)
		if request.Method != "notifications/initialized" {
			if got := r.Header.Get("Authorization"); got != "Bearer secret" {
				t.Errorf("Authorization = %q", got)
			}
			if got := r.Header.Get("X-Client"); got != "weave-test" {
				t.Errorf("X-Client = %q", got)
			}
		}

		switch request.Method {
		case "initialize":
			writeMCPResponse(t, w, map[string]interface{}{"serverInfo": map[string]interface{}{"name": "test"}})
		case "notifications/initialized":
			w.WriteHeader(http.StatusNoContent)
		case "tools/list":
			writeMCPResponse(t, w, map[string]interface{}{"tools": []interface{}{
				map[string]interface{}{
					"name":        "search",
					"description": "Search documents",
					"inputSchema": map[string]interface{}{"type": "object"},
				},
				"ignored",
			}})
		case "tools/call":
			args, _ := request.Params["arguments"].(map[string]interface{})
			if args["raw"] == true {
				writeMCPResponse(t, w, map[string]interface{}{"answer": "raw"})
				return
			}
			writeMCPResponse(t, w, map[string]interface{}{"content": []interface{}{"done"}})
		default:
			t.Errorf("unexpected method: %s", request.Method)
		}
	}))
	t.Cleanup(server.Close)

	config := &Config{
		Name:      "local",
		ServerURL: server.URL,
		Transport: TransportHTTP,
		Timeout:   time.Second,
		Auth:      &Auth{Type: "bearer", Token: "secret"},
		Headers:   map[string]string{"X-Client": "weave-test"},
	}
	clientValue, err := NewHTTPClient(config)
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}
	client := clientValue.(*HTTPClient)

	if err := client.Ping(context.Background()); err == nil || !strings.Contains(err.Error(), "not connected") {
		t.Fatalf("Ping() before Connect error = %v", err)
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if client.Config() != config {
		t.Fatal("Config() did not return original configuration")
	}
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	tools, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "search" || tools[0].Description != "Search documents" {
		t.Fatalf("ListTools() = %#v", tools)
	}
	if tools[0].InputSchema["type"] != "object" {
		t.Fatalf("InputSchema = %#v", tools[0].InputSchema)
	}

	content, err := client.CallTool(context.Background(), "search", map[string]interface{}{"query": "hello"})
	if err != nil {
		t.Fatalf("CallTool(content) error = %v", err)
	}
	if got := content.([]interface{})[0]; got != "done" {
		t.Fatalf("CallTool(content) = %#v", content)
	}
	raw, err := client.CallTool(context.Background(), "search", map[string]interface{}{"raw": true})
	if err != nil {
		t.Fatalf("CallTool(raw) error = %v", err)
	}
	if raw.(map[string]interface{})["answer"] != "raw" {
		t.Fatalf("CallTool(raw) = %#v", raw)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if client.connected {
		t.Fatal("Close() left client connected")
	}
	if len(methods) != 6 {
		t.Fatalf("methods = %v", methods)
	}
}

func TestHTTPClientProtocolErrors(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		operation string
		want      string
	}{
		{name: "HTTP status", status: http.StatusBadGateway, body: "offline", operation: "list", want: "HTTP 502: offline"},
		{name: "invalid JSON", status: http.StatusOK, body: "not-json", operation: "list", want: "failed to unmarshal response"},
		{name: "list protocol error", status: http.StatusOK, body: `{"jsonrpc":"2.0","id":1,"error":{"code":-1,"message":"denied"}}`, operation: "list", want: "list tools error: denied"},
		{name: "invalid tools", status: http.StatusOK, body: `{"jsonrpc":"2.0","id":1,"result":{"tools":"bad"}}`, operation: "list", want: "invalid tools response format"},
		{name: "call protocol error", status: http.StatusOK, body: `{"jsonrpc":"2.0","id":1,"error":"broken"}`, operation: "call", want: "call tool error: broken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			t.Cleanup(server.Close)

			client := &HTTPClient{
				config:     &Config{},
				baseURL:    server.URL,
				httpClient: server.Client(),
			}
			var err error
			if tt.operation == "call" {
				_, err = client.CallTool(context.Background(), "test", nil)
			} else {
				_, err = client.ListTools(context.Background())
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestHTTPClientConnectErrors(t *testing.T) {
	t.Run("initialize transport failure", func(t *testing.T) {
		client := &HTTPClient{config: &Config{}, baseURL: ":", httpClient: http.DefaultClient}
		err := client.Connect(context.Background())
		if err == nil || !strings.Contains(err.Error(), "failed to create HTTP request") {
			t.Fatalf("Connect() error = %v", err)
		}
	})

	t.Run("initialize protocol error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"jsonrpc":"2.0","id":1,"error":{"code":42,"message":"no handshake"}}`)
		}))
		t.Cleanup(server.Close)
		client := &HTTPClient{config: &Config{}, baseURL: server.URL, httpClient: server.Client()}
		err := client.Connect(context.Background())
		if err == nil || !strings.Contains(err.Error(), "initialize error: no handshake") {
			t.Fatalf("Connect() error = %v", err)
		}
	})

	t.Run("notification failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var request MCPRequest
			_ = json.NewDecoder(r.Body).Decode(&request)
			if request.Method == "initialize" {
				writeMCPResponse(t, w, map[string]interface{}{})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, "notification rejected")
		}))
		t.Cleanup(server.Close)
		client := &HTTPClient{config: &Config{}, baseURL: server.URL, httpClient: server.Client()}
		err := client.Connect(context.Background())
		if err == nil || !strings.Contains(err.Error(), "notification rejected") {
			t.Fatalf("Connect() error = %v", err)
		}
	})
}

func TestHTTPClientAuthHeaders(t *testing.T) {
	tests := []struct {
		name string
		auth *Auth
		key  string
		want string
	}{
		{name: "none", auth: nil, key: "Authorization", want: ""},
		{name: "bearer", auth: &Auth{Type: "bearer", Token: "token"}, key: "Authorization", want: "Bearer token"},
		{name: "basic", auth: &Auth{Type: "basic", Username: "user", Password: "pass"}, key: "Authorization", want: "Basic dXNlcjpwYXNz"},
		{name: "api key", auth: &Auth{Type: "api-key", APIKey: "key"}, key: "X-API-Key", want: "key"},
		{name: "custom", auth: &Auth{Headers: map[string]string{"X-Auth": "custom"}}, key: "X-Auth", want: "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &HTTPClient{config: &Config{Auth: tt.auth}}
			req := httptest.NewRequest(http.MethodPost, "http://example.test", nil)
			client.addAuthHeaders(req)
			if got := req.Header.Get(tt.key); got != tt.want {
				t.Fatalf("header %s = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestMCPConfigValidationAndFactory(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{name: "nil", want: "config cannot be nil"},
		{name: "missing URL", config: &Config{Transport: TransportHTTP}, want: "server URL is required"},
		{name: "bad transport", config: &Config{ServerURL: "x", Transport: "grpc"}, want: "transport must be"},
		{name: "negative timeout", config: &Config{ServerURL: "x", Transport: TransportHTTP, Timeout: -1}, want: "timeout cannot be negative"},
		{name: "negative retries", config: &Config{ServerURL: "x", Transport: TransportHTTP, MaxRetries: -1}, want: "max_retries cannot be negative"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidateConfig() error = %v, want containing %q", err, tt.want)
			}
		})
	}

	config := &Config{ServerURL: "http://example.test", Transport: TransportHTTP}
	client, err := NewMCPClient(config)
	if err != nil {
		t.Fatalf("NewMCPClient() error = %v", err)
	}
	if _, ok := client.(*HTTPClient); !ok {
		t.Fatalf("NewMCPClient() type = %T", client)
	}
	if config.Timeout != 30*time.Second || config.MaxRetries != 3 {
		t.Fatalf("defaults = timeout %s, retries %d", config.Timeout, config.MaxRetries)
	}

	if _, err := NewHTTPClient(&Config{Transport: TransportStdio}); err == nil {
		t.Fatal("NewHTTPClient() accepted stdio transport")
	}
	if _, err := NewStdioClient(&Config{Transport: TransportHTTP}); err == nil {
		t.Fatal("NewStdioClient() accepted HTTP transport")
	}
}

func writeMCPResponse(t *testing.T, w http.ResponseWriter, result map[string]interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(MCPResponse{JSONRPC: "2.0", ID: 1, Result: result}); err != nil {
		t.Errorf("encode response: %v", err)
	}
}
