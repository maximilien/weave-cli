// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeAgentDefinition(t *testing.T, path, name, agentType, description string) {
	t.Helper()
	content := "name: " + name + "\n" +
		"type: " + agentType + "\n" +
		"description: " + description + "\n" +
		"version: 1.2.3\n" +
		"llm:\n  model: gpt-4o\n" +
		"system_prompt: Be helpful.\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func TestAgentLoaderCacheAndReloadLifecycle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docs.yaml")
	writeAgentDefinition(t, path, "docs", "rag", "first version")

	loader := &AgentLoader{cache: make(map[string]*CustomAgentConfig)}
	loader.AddSearchPath(dir)
	loader.AddSearchPath(dir)
	if paths := loader.GetSearchPaths(); len(paths) != 1 || paths[0] != dir {
		t.Fatalf("GetSearchPaths() = %#v", paths)
	}
	found, err := loader.FindAgentFile("docs")
	if err != nil || found != path {
		t.Fatalf("FindAgentFile() = %q, %v", found, err)
	}

	config, err := loader.LoadAgent("docs")
	if err != nil || config.Description != "first version" {
		t.Fatalf("LoadAgent() = %#v, %v", config, err)
	}
	writeAgentDefinition(t, path, "docs", "rag", "second version")
	cached, err := loader.LoadAgent("docs")
	if err != nil || cached.Description != "first version" {
		t.Fatalf("LoadAgent(cached) = %#v, %v", cached, err)
	}
	if got, ok := loader.GetCachedAgent("docs"); !ok || got != cached {
		t.Fatalf("GetCachedAgent() = %#v, %t", got, ok)
	}
	reloaded, err := loader.ReloadAgent("docs")
	if err != nil || reloaded.Description != "second version" {
		t.Fatalf("ReloadAgent() = %#v, %v", reloaded, err)
	}
	loader.ClearCache()
	if got, ok := loader.GetCachedAgent("docs"); ok || got != nil {
		t.Fatalf("GetCachedAgent(after clear) = %#v, %t", got, ok)
	}

	_, err = loader.FindAgentFile("missing")
	if err == nil || !IsAgentNotFoundError(err) || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("FindAgentFile(missing) error = %v", err)
	}
	if IsAgentNotFoundError(errors.New("other")) {
		t.Fatal("IsAgentNotFoundError() accepted ordinary error")
	}

	writeAgentDefinition(t, filepath.Join(dir, "mismatch.yaml"), "different", "rag", "mismatch")
	if _, err := loader.LoadAgent("mismatch"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("LoadAgent(mismatch) error = %v", err)
	}
}

func TestAgentRegistryDiscoveryAndValidation(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	writeAgentDefinition(t, filepath.Join(first, "docs.yaml"), "docs", "rag", "first path wins")
	writeAgentDefinition(t, filepath.Join(first, "summary.yml"), "summary", "summarize", "summary agent")
	writeAgentDefinition(t, filepath.Join(second, "docs.yaml"), "docs", "rag", "duplicate")
	if err := os.WriteFile(filepath.Join(first, "invalid.yaml"), []byte("not: valid: yaml"), 0o600); err != nil {
		t.Fatalf("WriteFile(invalid): %v", err)
	}
	if err := os.WriteFile(filepath.Join(first, "README.txt"), []byte("ignored"), 0o600); err != nil {
		t.Fatalf("WriteFile(readme): %v", err)
	}
	if err := os.Mkdir(filepath.Join(first, "nested.yaml"), 0o700); err != nil {
		t.Fatalf("Mkdir(nested): %v", err)
	}

	loader := &AgentLoader{searchPaths: []string{first, filepath.Join(first, "missing"), second}, cache: make(map[string]*CustomAgentConfig)}
	registry := NewAgentRegistry(loader)
	agents, err := registry.ListAgents()
	if err != nil || len(agents) != 2 {
		t.Fatalf("ListAgents() = %#v, %v", agents, err)
	}
	info, err := registry.GetAgentInfo("docs")
	if err != nil || info.Name != "docs" || info.Description != "first path wins" || info.Version != "1.2.3" {
		t.Fatalf("GetAgentInfo() = %#v, %v", info, err)
	}
	if !registry.AgentExists("docs") || registry.AgentExists("missing") {
		t.Fatal("AgentExists() returned incorrect result")
	}
	if err := registry.ValidateAgent("docs"); err != nil {
		t.Fatalf("ValidateAgent(): %v", err)
	}
	if err := registry.ValidateAgent("missing"); err == nil {
		t.Fatal("ValidateAgent(missing) succeeded")
	}
	ragAgents, err := registry.GetAgentsByType("rag")
	if err != nil || len(ragAgents) != 1 || ragAgents[0].Name != "docs" {
		t.Fatalf("GetAgentsByType() = %#v, %v", ragAgents, err)
	}
	if _, err := registry.GetAgentInfo("missing"); err == nil {
		t.Fatal("GetAgentInfo(missing) succeeded")
	}
}
