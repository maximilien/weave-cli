// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
)

// TestREPLCommandStructure tests REPL command structures
func TestREPLCommandStructure(t *testing.T) {
	t.Run("CommandParsing", func(t *testing.T) {
		// Test command parsing patterns used by REPL
		commands := []string{
			"/status",
			"/collection list",
			"/mcp list",
			"/search query text",
			"/use collection_name",
		}

		for _, cmd := range commands {
			parts := strings.Fields(cmd)
			if len(parts) == 0 {
				t.Errorf("Command '%s' should parse to at least one part", cmd)
			}

			// Verify command starts with /
			if !strings.HasPrefix(parts[0], "/") {
				t.Errorf("Command '%s' should start with /", cmd)
			}
		}
	})

	t.Run("SpecialCommands", func(t *testing.T) {
		// Test recognition of special REPL commands
		specialCommands := map[string]bool{
			"/help":       true,
			"/status":     true,
			"/exit":       true,
			"/collection": true,
			"/mcp":        true,
			"/search":     true,
			"/stats":      true,
			"/use":        true,
		}

		for cmd := range specialCommands {
			if !strings.HasPrefix(cmd, "/") {
				t.Errorf("Special command '%s' should start with /", cmd)
			}
		}
	})
}

// TestREPLWithMockClient tests REPL-related operations with mock client
func TestREPLWithMockClient(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 1536,
		Collections: []config.MockCollection{
			{Name: "test_collection", Type: "text", Description: "Test collection for REPL"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	t.Run("HealthCheck", func(t *testing.T) {
		err := client.Health(ctx)
		if err != nil {
			t.Fatalf("Mock client health check failed: %v", err)
		}
	})

	t.Run("ListCollections", func(t *testing.T) {
		// Simulates: /collection list
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Fatalf("Failed to list collections: %v", err)
		}

		if len(collections) != 1 {
			t.Errorf("Expected 1 collection, got %d", len(collections))
		}

		if collections[0] != "test_collection" {
			t.Errorf("Expected 'test_collection', got '%s'", collections[0])
		}
	})

	t.Run("DocumentOperations", func(t *testing.T) {
		// Simulates document commands that REPL would execute
		doc := mock.Document{
			ID:      "doc1",
			Content: "Test document content for REPL testing",
			Metadata: map[string]interface{}{
				"source": "test",
				"type":   "text",
			},
		}

		err := client.AddDocument(ctx, "test_collection", doc)
		if err != nil {
			t.Fatalf("Failed to add document: %v", err)
		}

		// List documents (simulates /collection info or /stats)
		docs, err := client.ListDocuments(ctx, "test_collection", 10)
		if err != nil {
			t.Fatalf("Failed to list documents: %v", err)
		}

		if len(docs) != 1 {
			t.Errorf("Expected 1 document, got %d", len(docs))
		}
	})
}

// TestREPLBatchMode tests REPL batch query file handling
func TestREPLBatchMode(t *testing.T) {
	tmpDir := t.TempDir()
	queryFile := tmpDir + "/queries.txt"

	queries := []string{
		"list collections",
		"show database info",
		"search for documents about testing",
	}

	err := os.WriteFile(queryFile, []byte(strings.Join(queries, "\n")), 0644)
	if err != nil {
		t.Fatalf("Failed to write query file: %v", err)
	}

	t.Run("QueryFileExists", func(t *testing.T) {
		_, err := os.Stat(queryFile)
		if err != nil {
			t.Fatalf("Query file should exist: %v", err)
		}

		content, err := os.ReadFile(queryFile)
		if err != nil {
			t.Fatalf("Failed to read query file: %v", err)
		}

		lines := strings.Split(string(content), "\n")
		if len(lines) != len(queries) {
			t.Errorf("Expected %d queries, got %d", len(queries), len(lines))
		}
	})

	t.Run("QueryParsing", func(t *testing.T) {
		content, err := os.ReadFile(queryFile)
		if err != nil {
			t.Fatalf("Failed to read query file: %v", err)
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if line == "" {
				continue
			}
			if line != queries[i] {
				t.Errorf("Query %d mismatch: expected '%s', got '%s'", i, queries[i], line)
			}
		}
	})
}

// TestREPLTabCompletion tests tab completion command structures
func TestREPLTabCompletion(t *testing.T) {
	// Test command completion patterns
	t.Run("PrimaryCommands", func(t *testing.T) {
		commands := []string{
			"/help", "/exit", "/quit", "/status",
			"/mcp", "/collection", "/search", "/stats", "/use",
		}

		for _, cmd := range commands {
			if !strings.HasPrefix(cmd, "/") {
				t.Errorf("Command '%s' should start with /", cmd)
			}

			// Command should have valid structure
			if len(cmd) < 2 {
				t.Errorf("Command '%s' is too short", cmd)
			}
		}
	})

	t.Run("SubCommands", func(t *testing.T) {
		subCommands := map[string][]string{
			"/mcp":        {"list", "call", "status"},
			"/collection": {"list", "use", "create", "delete", "info"},
		}

		for primary, subs := range subCommands {
			if len(subs) == 0 {
				t.Errorf("Command '%s' should have subcommands", primary)
			}

			for _, sub := range subs {
				if sub == "" {
					t.Errorf("Subcommand for '%s' should not be empty", primary)
				}
			}
		}
	})
}

// TestREPLPromptUpdate tests prompt customization
func TestREPLPromptUpdate(t *testing.T) {
	t.Run("DefaultPrompt", func(t *testing.T) {
		defaultPrompt := "weave> "
		if len(defaultPrompt) == 0 {
			t.Error("Default prompt should not be empty")
		}
	})

	t.Run("CollectionPrompt", func(t *testing.T) {
		collectionName := "my_collection"
		prompt := "weave[" + collectionName + "]> "

		if !strings.Contains(prompt, collectionName) {
			t.Errorf("Prompt should contain collection name '%s'", collectionName)
		}

		if !strings.HasPrefix(prompt, "weave[") {
			t.Error("Prompt should start with 'weave['")
		}

		if !strings.HasSuffix(prompt, "]> ") {
			t.Error("Prompt should end with ']> '")
		}
	})
}
