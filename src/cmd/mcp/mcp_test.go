// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkgmcp "github.com/maximilien/weave-cli/src/pkg/mcp"
)

func TestMCPCommandContractsAndValidation(t *testing.T) {
	if MCPCmd.Use != "mcp" || len(MCPCmd.Commands()) != 3 {
		t.Fatalf("MCPCmd = %q with %d subcommands", MCPCmd.Use, len(MCPCmd.Commands()))
	}
	if err := callCmd.ValidateArgs([]string{"tool"}); err != nil {
		t.Fatalf("call command rejected tool name: %v", err)
	}
	if err := callCmd.ValidateArgs(nil); err == nil {
		t.Fatal("call command accepted missing tool name")
	}

	serverURL = ""
	if err := runMCPList(nil, nil); err == nil || !strings.Contains(err.Error(), "--server is required") {
		t.Fatalf("runMCPList() error = %v", err)
	}
	if err := runMCPTest(nil, nil); err == nil || !strings.Contains(err.Error(), "--server is required") {
		t.Fatalf("runMCPTest() error = %v", err)
	}
	if err := runMCPCall(nil, []string{"tool"}); err == nil || !strings.Contains(err.Error(), "--server is required") {
		t.Fatalf("runMCPCall() error = %v", err)
	}

	serverURL = "unused"
	transport = "invalid"
	if err := runMCPList(nil, nil); err == nil || !strings.Contains(err.Error(), "failed to create MCP client") {
		t.Fatalf("runMCPList(invalid transport) error = %v", err)
	}
	callArgs = []string{"missing-separator"}
	if err := runMCPCall(nil, []string{"tool"}); err == nil || !strings.Contains(err.Error(), "invalid argument format") {
		t.Fatalf("runMCPCall(invalid argument) error = %v", err)
	}
}

func TestMCPHTTPCommands(t *testing.T) {
	server := newCommandMCPServer(t)
	serverURL = server.URL
	transport = "http"
	timeout = 2
	t.Cleanup(func() {
		serverURL = ""
		transport = "http"
		timeout = 30
		callArgs = nil
		callOutputFormat = "text"
		listOutputFormat = "table"
	})

	listOutputFormat = "table"
	if err := runMCPList(nil, nil); err != nil {
		t.Fatalf("runMCPList(table) error = %v", err)
	}
	listOutputFormat = "json"
	if err := runMCPList(nil, nil); err != nil {
		t.Fatalf("runMCPList(json) error = %v", err)
	}

	callArgs = []string{"count=3", "enabled=true", "query=hello"}
	callOutputFormat = "text"
	if err := runMCPCall(nil, []string{"search"}); err != nil {
		t.Fatalf("runMCPCall(text) error = %v", err)
	}
	callOutputFormat = "json"
	if err := runMCPCall(nil, []string{"search"}); err != nil {
		t.Fatalf("runMCPCall(json) error = %v", err)
	}

	if err := runMCPTest(nil, nil); err != nil {
		t.Fatalf("runMCPTest() error = %v", err)
	}
}

func TestMCPDisplayVariants(t *testing.T) {
	displayToolsTable(nil)
	displayToolsTable([]pkgmcp.Tool{{
		Name:        "search",
		Description: "Search documents",
		InputSchema: map[string]interface{}{
			"properties": map[string]interface{}{
				"query":   map[string]interface{}{"type": "string", "description": "search text"},
				"limit":   map[string]interface{}{},
				"ignored": "not-a-property",
			},
		},
	}})

	displayToolResult("plain text")
	displayToolResult([]interface{}{
		map[string]interface{}{"type": "text", "text": "answer"},
		map[string]interface{}{"type": "image", "data": "encoded"},
		"ignored",
	})
	displayToolResult(map[string]interface{}{"text": "answer", "count": 2})
	displayToolResult(make(chan int))
}

func newCommandMCPServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp/v1/rpc" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		var request pkgmcp.MCPRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}

		if request.Method == "notifications/initialized" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		result := map[string]interface{}{}
		switch request.Method {
		case "initialize":
			result["serverInfo"] = map[string]interface{}{"name": "test"}
		case "tools/list":
			result["tools"] = []interface{}{map[string]interface{}{
				"name":        "search",
				"description": "Search documents",
				"inputSchema": map[string]interface{}{"properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}},
			}}
		case "tools/call":
			result["content"] = []interface{}{map[string]interface{}{"type": "text", "text": "done"}}
		default:
			t.Errorf("unexpected method %q", request.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(pkgmcp.MCPResponse{JSONRPC: "2.0", ID: 1, Result: result}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
