// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"strings"
	"testing"
)

func TestWeaveAgentArgumentBoundaries(t *testing.T) {
	agent := NewWeaveAgent(nil)
	if agent.Name() != "WeaveAgent" || agent.verbose {
		t.Fatalf("unexpected new agent: %#v", agent)
	}
	agent.SetVerbose(true)
	if !agent.verbose {
		t.Fatal("SetVerbose(true) did not update the agent")
	}
	if result, err := agent.Execute(context.Background(), "invalid"); err == nil || result != nil {
		t.Fatalf("Execute(invalid) = (%#v, %v)", result, err)
	}

	if err := agent.ValidateArguments("create_collection", map[string]interface{}{"name": "Docs"}); err == nil || !strings.Contains(err.Error(), "type") {
		t.Fatalf("expected missing type error, got %v", err)
	}
	if err := agent.ValidateArguments("create_collection", map[string]interface{}{"name": "Docs", "type": "text"}); err != nil {
		t.Fatal(err)
	}
	if err := agent.ValidateArguments("unknown", nil); err != nil {
		t.Fatal(err)
	}

	args := map[string]interface{}{"collection": "Docs"}
	ctx := map[string]interface{}{"query": "fixture", "ignored": true}
	inferred := agent.InferMissingArguments("query_documents", args, ctx)
	if inferred["collection"] != "Docs" || inferred["query"] != "fixture" || len(args) != 1 {
		t.Fatalf("InferMissingArguments() = %#v, original %#v", inferred, args)
	}
	if requirements := getRequiredArguments("create_document"); len(requirements) != 3 {
		t.Fatalf("create_document requirements = %#v", requirements)
	}
}

func TestAgentUtilityBoundaries(t *testing.T) {
	bashAgent := NewBashAgent()
	installed, err := bashAgent.CheckToolInstalled(context.Background(), "sh")
	if err != nil || !installed {
		t.Fatalf("CheckToolInstalled(sh) = (%v, %v)", installed, err)
	}
	installed, err = bashAgent.CheckToolInstalled(context.Background(), "weave-tool-that-does-not-exist")
	if err != nil || installed {
		t.Fatalf("CheckToolInstalled(missing) = (%v, %v)", installed, err)
	}

	config := &CustomAgentConfig{Name: "fixture", Type: "rag"}
	ragAgent := &RAGAgent{config: config}
	if got := ragAgent.extractImageURL("prefix Image URL: https://example.test/image.png\ncaption"); got != "https://example.test/image.png" {
		t.Fatalf("extractImageURL(newline) = %q", got)
	}
	if got := ragAgent.extractImageURL("Image URL: https://example.test/end.png"); got != "https://example.test/end.png" {
		t.Fatalf("extractImageURL(end) = %q", got)
	}
	if got := ragAgent.extractImageURL("no image"); got != "" {
		t.Fatalf("extractImageURL(missing) = %q", got)
	}
	if ragAgent.GetConfig() != config || ragAgent.GetType() != "rag" || !ragAgent.IsRAGType() {
		t.Fatal("RAG configuration helpers returned unexpected values")
	}

	builder := &ContextBuilder{}
	short := builder.contentHash("  short content  ")
	long := builder.contentHash(strings.Repeat("x", 101))
	if short != "short content" || !strings.HasSuffix(long, ":101") {
		t.Fatalf("content hashes = %q, %q", short, long)
	}

	schemaAgent := NewSchemaAgent(nil)
	if schemaAgent == nil || schemaAgent.Name() != "schema-agent" {
		t.Fatalf("NewSchemaAgent(nil) = %#v", schemaAgent)
	}
}
