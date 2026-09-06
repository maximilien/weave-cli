// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package repl

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/mcp"
)

type fakeMCPClient struct {
	config     *mcp.Config
	tools      []mcp.Tool
	listErr    error
	callResult interface{}
	callErr    error
	pingErr    error
	closed     bool
	calledTool string
	calledArgs map[string]interface{}
}

func (c *fakeMCPClient) Connect(context.Context) error { return nil }
func (c *fakeMCPClient) Close() error {
	c.closed = true
	return nil
}
func (c *fakeMCPClient) Ping(context.Context) error { return c.pingErr }
func (c *fakeMCPClient) ListTools(context.Context) ([]mcp.Tool, error) {
	return c.tools, c.listErr
}
func (c *fakeMCPClient) CallTool(_ context.Context, tool string, args map[string]interface{}) (interface{}, error) {
	c.calledTool = tool
	c.calledArgs = args
	return c.callResult, c.callErr
}
func (c *fakeMCPClient) Config() *mcp.Config { return c.config }

func TestCollectionAndSearchCommands(t *testing.T) {
	executor := &fakeQueryExecutor{}
	repl := &REPL{executor: executor}

	repl.handleUseCommand(nil)
	repl.handleUseCommand([]string{"docs"})
	if repl.currentCollection != "docs" {
		t.Fatalf("current collection = %q", repl.currentCollection)
	}
	if got := repl.getCurrentCollectionOrPrompt(); got != "collection docs" {
		t.Fatalf("current collection prompt = %q", got)
	}

	repl.handleSearchCommand([]string{"vector", "search"})
	repl.handleStatsCommand(nil)
	repl.collectionList()
	repl.collectionInfo()
	repl.collectionCreate([]string{"archive"})
	repl.collectionDelete([]string{"old"})
	repl.collectionUse([]string{"new-docs"})

	want := []string{
		"search for 'vector search' in collection docs",
		"show statistics for collection docs",
		"list all collections",
		"show detailed information about collection docs",
		"create a new text collection named archive",
		"delete collection old",
	}
	if !reflect.DeepEqual(executor.queries, want) {
		t.Fatalf("executed queries = %#v, want %#v", executor.queries, want)
	}

	repl.handleSearchCommand(nil)
	repl.collectionUse(nil)
	repl.collectionCreate(nil)
	repl.collectionDelete(nil)
	repl.currentCollection = ""
	repl.handleStatsCommand(nil)
	repl.collectionInfo()
	if got := repl.getCurrentCollectionOrPrompt(); got != "all collections" {
		t.Fatalf("empty collection prompt = %q", got)
	}
}

func TestCollectionCommandRouting(t *testing.T) {
	executor := &fakeQueryExecutor{}
	repl := &REPL{executor: executor}
	commands := []string{
		"/collection",
		"/collection ls",
		"/collection create docs",
		"/collection delete old",
		"/collection use active",
		"/collection info",
		"/collection help",
		"/collection unknown",
		"/search embeddings",
		"/stats",
		"/use final",
	}
	for _, command := range commands {
		if !repl.handleSpecialCommand(command) {
			t.Errorf("handleSpecialCommand(%q) was not handled", command)
		}
	}
	if repl.currentCollection != "final" {
		t.Fatalf("current collection = %q", repl.currentCollection)
	}
}

func TestMCPFallbackCommands(t *testing.T) {
	executor := &fakeQueryExecutor{}
	repl := &REPL{executor: executor}

	repl.handleMCPCommand(nil)
	repl.handleMCPCommand([]string{"help"})
	repl.handleMCPCommand([]string{"unknown"})
	repl.handleMCPCommand([]string{"list"})
	repl.handleMCPCommand([]string{"call"})
	repl.handleMCPCommand([]string{"call", "search", "query=hello"})
	repl.handleMCPCommand([]string{"status"})

	want := []string{
		"list all available MCP tools",
		"call MCP tool search with arguments: query=hello",
	}
	if !reflect.DeepEqual(executor.queries, want) {
		t.Fatalf("fallback queries = %#v, want %#v", executor.queries, want)
	}
}

func TestDirectMCPCommands(t *testing.T) {
	client := &fakeMCPClient{
		config: &mcp.Config{ServerURL: "http://mcp.test", Transport: mcp.TransportHTTP},
		tools: []mcp.Tool{{
			Name:        "search",
			Description: "Search documents",
			InputSchema: map[string]interface{}{
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search text"},
					"limit": map[string]interface{}{},
					"other": "ignored",
				},
			},
		}},
		callResult: map[string]interface{}{"matches": float64(2)},
	}
	repl := &REPL{mcpClient: client, mcpEnabled: true}

	repl.mcpList()
	repl.mcpCall([]string{"search", `limit=2`, "query=hello", "ignored"})
	repl.mcpStatus()
	repl.displayStatus()
	repl.closeMCPClient()

	if client.calledTool != "search" {
		t.Fatalf("called tool = %q", client.calledTool)
	}
	wantArgs := map[string]interface{}{"limit": float64(2), "query": "hello"}
	if !reflect.DeepEqual(client.calledArgs, wantArgs) {
		t.Fatalf("called args = %#v, want %#v", client.calledArgs, wantArgs)
	}
	if !client.closed {
		t.Fatal("MCP client was not closed")
	}
}

func TestDirectMCPFailuresAndEmptyTools(t *testing.T) {
	client := &fakeMCPClient{
		config:  &mcp.Config{ServerURL: "http://mcp.test", Transport: mcp.TransportHTTP},
		listErr: errors.New("list failed"),
		callErr: errors.New("call failed"),
		pingErr: errors.New("ping failed"),
	}
	repl := &REPL{mcpClient: client, mcpEnabled: true}
	repl.mcpList()
	repl.mcpCall([]string{"search"})
	repl.mcpStatus()

	client.listErr = nil
	repl.mcpList()
	repl.displayToolsList(nil)
}

func TestMCPConfigurationValidation(t *testing.T) {
	repl := &REPL{}
	err := repl.initMCPClient(Options{MCPServerURL: "http://mcp.test", MCPTransport: "invalid"})
	if err == nil || !strings.Contains(err.Error(), "invalid MCP config") {
		t.Fatalf("initMCPClient() error = %v", err)
	}

	err = repl.initMCPClient(Options{MCPServerURL: "", MCPTransport: "http", MCPTimeout: -1})
	if err == nil || !strings.Contains(err.Error(), "invalid MCP config") {
		t.Fatalf("initMCPClient() timeout error = %v", err)
	}
}
