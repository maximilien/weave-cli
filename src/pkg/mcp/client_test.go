// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
