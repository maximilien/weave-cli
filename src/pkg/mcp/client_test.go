// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("WEAVE_MCP_TEST_HELPER") == "1" {
		mcpTestHelperProcess()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func mcpTestHelperProcess() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request MCPRequest
		if json.Unmarshal(scanner.Bytes(), &request) != nil || request.ID == nil {
			continue
		}
		var result interface{}
		switch request.Method {
		case "initialize":
			if os.Getenv("WEAVE_MCP_TEST_MODE") == "initialize-no-response" {
				return
			}
			if os.Getenv("WEAVE_MCP_TEST_MODE") == "initialize-error" {
				_ = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"jsonrpc": "2.0", "id": *request.ID,
					"error": map[string]interface{}{"code": -1, "message": "fixture rejected"},
				})
				fmt.Fprintln(os.Stderr, "fixture initialization failed")
				continue
			}
			result = map[string]interface{}{"protocolVersion": "2024-11-05"}
		case "tools/list":
			result = map[string]interface{}{"tools": []interface{}{map[string]interface{}{
				"name": "echo", "description": "Echo fixture", "inputSchema": map[string]interface{}{"type": "object"},
			}}}
		case "tools/call":
			result = map[string]interface{}{"content": []interface{}{"called"}}
		default:
			result = map[string]interface{}{}
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": *request.ID, "result": result})
		if request.Method == "initialize" {
			fmt.Fprintln(os.Stderr, "stdio helper ready")
		}
	}
}

func TestNewClientReportsInitializationProtocolError(t *testing.T) {
	t.Setenv("WEAVE_MCP_TEST_HELPER", "1")
	t.Setenv("WEAVE_MCP_TEST_MODE", "initialize-error")
	if _, err := NewClient(os.Args[0]); err == nil || !strings.Contains(err.Error(), "initialize error: fixture rejected") {
		t.Fatalf("NewClient() error = %v", err)
	}
}

func TestNewClientReportsInitializationReadError(t *testing.T) {
	t.Setenv("WEAVE_MCP_TEST_HELPER", "1")
	t.Setenv("WEAVE_MCP_TEST_MODE", "initialize-no-response")
	if _, err := NewClient(os.Args[0]); err == nil || !strings.Contains(err.Error(), "failed to read initialize response") {
		t.Fatalf("NewClient() error = %v", err)
	}
}

type bufferWriteCloser struct{ bytes.Buffer }

func (b *bufferWriteCloser) Close() error { return nil }

type errorWriteCloser struct{}

func (errorWriteCloser) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func (errorWriteCloser) Close() error              { return nil }

func TestMCPResponseUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantCode int
		wantMsg  string
		wantErr  bool
	}{
		{name: "result", payload: `{"jsonrpc":"2.0","id":1,"result":{"ok":true}}`},
		{name: "object error", payload: `{"jsonrpc":"2.0","id":1,"error":{"code":12,"message":"bad","data":"detail"}}`, wantCode: 12, wantMsg: "bad"},
		{name: "string error", payload: `{"jsonrpc":"2.0","id":1,"error":"broken"}`, wantCode: -1, wantMsg: "broken"},
		{name: "unknown error shape", payload: `{"jsonrpc":"2.0","id":1,"error":42}`},
		{name: "invalid JSON", payload: `{`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response MCPResponse
			err := json.Unmarshal([]byte(tt.payload), &response)
			if tt.wantErr {
				if err == nil {
					t.Fatal("json.Unmarshal() succeeded")
				}
				return
			}
			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if tt.wantMsg == "" {
				if response.Error != nil {
					t.Fatalf("Error = %#v", response.Error)
				}
				return
			}
			if response.Error == nil || response.Error.Code != tt.wantCode || response.Error.Message != tt.wantMsg {
				t.Fatalf("Error = %#v", response.Error)
			}
		})
	}
}

func TestLegacyClientProtocol(t *testing.T) {
	t.Run("initialize", func(t *testing.T) {
		stdin := &bufferWriteCloser{}
		client := &Client{
			stdin:     stdin,
			reader:    bufio.NewReader(strings.NewReader(`{"jsonrpc":"2.0","id":1,"result":{}}` + "\n")),
			requestID: 0,
		}
		if err := client.initialize(); err != nil {
			t.Fatalf("initialize() error = %v", err)
		}
		if !strings.Contains(stdin.String(), `"method":"initialize"`) || !strings.Contains(stdin.String(), `"method":"notifications/initialized"`) {
			t.Fatalf("requests = %s", stdin.String())
		}
	})

	t.Run("list tools", func(t *testing.T) {
		stdin := &bufferWriteCloser{}
		response := `{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"read","description":"Read","inputSchema":{"type":"object"}},"skip"]}}` + "\n"
		client := &Client{stdin: stdin, reader: bufio.NewReader(strings.NewReader(response))}
		tools, err := client.ListTools(context.Background())
		if err != nil {
			t.Fatalf("ListTools() error = %v", err)
		}
		if len(tools) != 1 || tools[0].Name != "read" || len(client.tools) != 1 {
			t.Fatalf("tools = %#v", tools)
		}
	})

	t.Run("call tool content and result", func(t *testing.T) {
		stdin := &bufferWriteCloser{}
		reader := strings.NewReader(
			`{"jsonrpc":"2.0","id":1,"result":{"content":"first"}}` + "\n" +
				`{"jsonrpc":"2.0","id":2,"result":{"value":"second"}}` + "\n",
		)
		client := &Client{stdin: stdin, reader: bufio.NewReader(reader)}
		content, err := client.CallTool(context.Background(), "read", map[string]interface{}{"id": "1"})
		if err != nil || content != "first" {
			t.Fatalf("CallTool(content) = %#v, %v", content, err)
		}
		result, err := client.CallTool(context.Background(), "read", nil)
		if err != nil || result.(map[string]interface{})["value"] != "second" {
			t.Fatalf("CallTool(result) = %#v, %v", result, err)
		}
	})
}

func TestLegacyClientErrors(t *testing.T) {
	t.Run("send marshal", func(t *testing.T) {
		client := &Client{stdin: &bufferWriteCloser{}}
		err := client.sendRequest(MCPRequest{Params: map[string]interface{}{"bad": make(chan int)}})
		if err == nil || !strings.Contains(err.Error(), "failed to marshal") {
			t.Fatalf("sendRequest() error = %v", err)
		}
	})

	t.Run("send write", func(t *testing.T) {
		client := &Client{stdin: errorWriteCloser{}}
		err := client.sendRequest(MCPRequest{Method: "ping"})
		if err == nil || !strings.Contains(err.Error(), "failed to write") {
			t.Fatalf("sendRequest() error = %v", err)
		}
	})

	tests := []struct {
		name    string
		payload string
		list    bool
		want    string
	}{
		{name: "read EOF", want: "failed to read response"},
		{name: "invalid response", payload: "bad\n", want: "failed to unmarshal response"},
		{name: "list protocol error", payload: `{"jsonrpc":"2.0","id":1,"error":"no tools"}` + "\n", list: true, want: "list tools error: no tools"},
		{name: "list invalid shape", payload: `{"jsonrpc":"2.0","id":1,"result":{"tools":"bad"}}` + "\n", list: true, want: "invalid tools response format"},
		{name: "call protocol error", payload: `{"jsonrpc":"2.0","id":1,"error":{"code":1,"message":"call denied"}}` + "\n", want: "call tool error: call denied"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{stdin: &bufferWriteCloser{}, reader: bufio.NewReader(strings.NewReader(tt.payload))}
			var err error
			if tt.list {
				_, err = client.ListTools(context.Background())
			} else if tt.name == "call protocol error" {
				_, err = client.CallTool(context.Background(), "bad", nil)
			} else {
				_, err = client.readResponse(context.Background())
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}

	t.Run("context cancellation", func(t *testing.T) {
		reader, writer := io.Pipe()
		t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
		client := &Client{reader: bufio.NewReader(reader)}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := client.readResponse(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("readResponse() error = %v", err)
		}
	})

	if _, err := NewClient(filepath.Join(t.TempDir(), "missing-server")); err == nil || !strings.Contains(err.Error(), "failed to start MCP server") {
		t.Fatalf("NewClient() error = %v", err)
	}
}

func TestLegacyClientHelpers(t *testing.T) {
	client := &Client{}
	if client.nextRequestID() != 1 || client.nextRequestID() != 2 {
		t.Fatalf("requestID = %d", client.requestID)
	}
	if getString(map[string]interface{}{"name": "tool"}, "name") != "tool" {
		t.Fatal("getString() missed string")
	}
	if getString(map[string]interface{}{"name": 1}, "name") != "" {
		t.Fatal("getString() accepted non-string")
	}

	t.Run("find go module", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test/test\n"), 0600); err != nil {
			t.Fatal(err)
		}
		nested := filepath.Join(root, "one", "two")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(nested)
		if got := findProjectRoot(); got != root {
			t.Fatalf("findProjectRoot() = %q, want %q", got, root)
		}
	})

	t.Run("find env", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".env"), nil, 0600); err != nil {
			t.Fatal(err)
		}
		t.Chdir(root)
		if got := findProjectRoot(); got != root {
			t.Fatalf("findProjectRoot() = %q, want %q", got, root)
		}
	})
}

func TestLegacyClientStderrAndClose(t *testing.T) {
	reader, writer := io.Pipe()
	client := &Client{
		stderr:     reader,
		stderrDone: make(chan struct{}),
	}
	go client.captureStderr()

	if _, err := writer.Write([]byte("server diagnostic")); err != nil {
		t.Fatalf("write stderr fixture: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}
	<-client.stderrDone
	if got := string(client.stderrBuf); got != "server diagnostic" {
		t.Fatalf("captured stderr = %q", got)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	withoutStderr := &Client{stderrDone: make(chan struct{})}
	withoutStderr.captureStderr()
	select {
	case <-withoutStderr.stderrDone:
	default:
		t.Fatal("captureStderr() did not close stderrDone when stderr is nil")
	}
}

func TestStdioClientConfigurationAndTransportValidation(t *testing.T) {
	t.Setenv("WEAVE_MCP_TEST_HELPER", "1")
	config := &Config{ServerURL: os.Args[0], Transport: TransportStdio}
	clientValue, err := NewStdioClient(config)
	if err != nil {
		t.Fatalf("NewStdioClient() error = %v", err)
	}
	client := clientValue.(*StdioClient)
	t.Cleanup(func() { _ = client.Close() })

	if _, err := NewStdioClient(&Config{Transport: TransportHTTP}); err == nil || !strings.Contains(err.Error(), "requires stdio transport") {
		t.Fatalf("NewStdioClient(http) error = %v", err)
	}
	if client.Config() != config {
		t.Fatal("Config() did not return original configuration")
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	tools, err := client.ListTools(context.Background())
	if err != nil || len(tools) != 1 || tools[0].Name != "echo" {
		t.Fatalf("ListTools() = %#v, %v", tools, err)
	}
	result, err := client.CallTool(context.Background(), "echo", map[string]interface{}{"message": "hello"})
	if err != nil || result.([]interface{})[0] != "called" {
		t.Fatalf("CallTool() = %#v, %v", result, err)
	}
}

func TestMCPClientConfigurationValidationAndDefaults(t *testing.T) {
	if _, err := NewHTTPClient(&Config{Transport: TransportStdio}); err == nil || !strings.Contains(err.Error(), "requires http transport") {
		t.Fatalf("NewHTTPClient(stdio) error = %v", err)
	}
	invalid := []struct {
		name   string
		config *Config
	}{
		{name: "nil config"},
		{name: "missing URL", config: &Config{Transport: TransportHTTP}},
		{name: "unknown transport", config: &Config{ServerURL: "http://example.invalid", Transport: "other"}},
		{name: "negative timeout", config: &Config{ServerURL: "http://example.invalid", Transport: TransportHTTP, Timeout: -1}},
		{name: "negative retries", config: &Config{ServerURL: "http://example.invalid", Transport: TransportHTTP, MaxRetries: -1}},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateConfig(test.config); err == nil {
				t.Fatal("ValidateConfig() accepted invalid configuration")
			}
		})
	}

	config := &Config{ServerURL: "http://127.0.0.1:1", Transport: TransportHTTP}
	client, err := NewMCPClient(config)
	if err != nil {
		t.Fatalf("NewMCPClient() error = %v", err)
	}
	if config.Timeout != 30*time.Second || config.MaxRetries != 3 {
		t.Fatalf("defaulted config = timeout %s, retries %d", config.Timeout, config.MaxRetries)
	}
	if _, ok := client.(*HTTPClient); !ok {
		t.Fatalf("NewMCPClient() type = %T, want *HTTPClient", client)
	}
}
