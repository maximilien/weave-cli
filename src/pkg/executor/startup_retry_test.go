// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
)

func TestMain(m *testing.M) {
	if os.Getenv("EXECUTOR_MCP_FIXTURE") == "1" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			var request struct {
				ID     *int   `json:"id"`
				Method string `json:"method"`
			}
			if json.Unmarshal(scanner.Bytes(), &request) != nil || request.ID == nil {
				continue
			}
			response := map[string]interface{}{"jsonrpc": "2.0", "id": *request.ID}
			switch request.Method {
			case "initialize":
				response["result"] = map[string]interface{}{"protocolVersion": "2024-11-05"}
			case "tools/list":
				if os.Getenv("EXECUTOR_MCP_REJECT") == "1" {
					response["error"] = map[string]interface{}{"code": -1, "message": "fixture rejected tools"}
				} else {
					response["result"] = map[string]interface{}{"tools": []interface{}{map[string]interface{}{"name": "fixture_echo", "description": "Fixture tool", "inputSchema": map[string]interface{}{"type": "object"}}}}
				}
			case "tools/call":
				response["result"] = map[string]interface{}{"content": []interface{}{map[string]interface{}{"type": "text", "text": "fixture result"}}}
			}
			_ = json.NewEncoder(os.Stdout).Encode(response)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestExecutorStartupWithLocalMCP(t *testing.T) {
	for _, mode := range []string{"explicit path", "environment path", "rejected tools", "missing executable"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("OPIK_ENABLED", "false")
			t.Setenv("OPENAI_API_KEY", "fixture-key")
			t.Setenv("EXECUTOR_MCP_FIXTURE", "1")
			t.Setenv("EXECUTOR_MCP_REJECT", "0")
			t.Setenv("WEAVE_MCP_STDIO_PATH", os.Args[0])
			cfg := &Config{Quiet: true, NoColor: true, Verbose: true}
			if mode == "explicit path" {
				cfg.StdioPath = os.Args[0]
				t.Setenv("WEAVE_MCP_STDIO_PATH", "ignored")
			}
			if mode == "rejected tools" {
				t.Setenv("EXECUTOR_MCP_REJECT", "1")
			}
			if mode == "missing executable" {
				cfg.StdioPath = root + "/missing"
			}
			executor, err := NewExecutor(cfg)
			if mode == "rejected tools" || mode == "missing executable" {
				want := "failed to list MCP tools"
				if mode == "missing executable" {
					want = "failed to create MCP client"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("startup error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := executor.Close(); err != nil {
					t.Errorf("close: %v", err)
				}
			})
			if executor.queryAgent == nil || executor.planningAgent == nil || executor.reportAgent == nil || executor.evalAgent == nil || executor.bashAgent == nil {
				t.Fatal("startup left agents uninitialized")
			}
			output, err := executor.executeWeaveStep(context.Background(), &agents.ExecutionStep{Command: "fixture_echo"})
			if err != nil || output != "fixture result" {
				t.Fatalf("tool result = %#v, %v", output, err)
			}
		})
	}
}

func TestExecutorRetriesRecoverOrExhaust(t *testing.T) {
	for _, recover := range []bool{true, false} {
		t.Run(map[bool]string{true: "recovery", false: "exhausted"}[recover], func(t *testing.T) {
			executor := newTestExecutor()
			executor.config.MaxRetries = 2
			calls := 0
			executor.weaveAgent = &fakeAgent{execute: func(context.Context, interface{}) (interface{}, error) {
				calls++
				if !recover || calls < 2 {
					return nil, errors.New("temporary fixture failure")
				}
				return &agents.WeaveAgentResult{Success: true, Output: "recovered"}, nil
			}}
			reports, err := executor.executePlan(context.Background(), &agents.ExecutionPlan{Steps: []agents.ExecutionStep{{Type: "weave", Command: "fixture_echo"}, {Type: "weave", Command: "dependent", DependsOn: []int{0}}}})
			if err != nil || len(reports) != 2 {
				t.Fatalf("reports=%+v error=%v", reports, err)
			}
			if recover {
				if calls != 3 || !reports[0].Success || reports[0].Output != "recovered" || !reports[1].Success {
					t.Fatalf("recovery calls=%d reports=%+v", calls, reports)
				}
			} else {
				if calls != 3 || reports[0].Success || reports[1].Error != "dependency failed" {
					t.Fatalf("exhaustion calls=%d reports=%+v", calls, reports)
				}
			}
		})
	}
}
