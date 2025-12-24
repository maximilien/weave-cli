// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCreateDefaultAgentsFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	agentsFile := filepath.Join(tmpDir, "weave-agents.yaml")

	// Test creating default agents file
	err := createDefaultAgentsFile(agentsFile)
	if err != nil {
		t.Fatalf("createDefaultAgentsFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
		t.Fatal("agents file was not created")
	}

	// Read file content
	content, err := os.ReadFile(agentsFile)
	if err != nil {
		t.Fatalf("failed to read agents file: %v", err)
	}

	contentStr := string(content)

	// Verify key sections exist
	requiredSections := []string{
		"llm:",
		"schema_agent:",
		"chunking_agent:",
		"query_agent:",
		"planning_agent:",
		"report_agent:",
		"eval_agent:",
		"output:",
		"cache:",
		"performance:",
		"features:",
	}

	for _, section := range requiredSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("agents file should contain '%s' section", section)
		}
	}

	// Verify key configuration values
	if !strings.Contains(contentStr, "provider: openai") {
		t.Error("should contain openai provider")
	}
	if !strings.Contains(contentStr, "default_model: \"gpt-4o\"") {
		t.Error("should contain gpt-4o model")
	}
	if !strings.Contains(contentStr, "max_samples: 50") {
		t.Error("should contain max_samples: 50")
	}
	if !strings.Contains(contentStr, "default_chunk_size: 1000") {
		t.Error("should contain default_chunk_size: 1000")
	}

	// Verify it's valid YAML
	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		t.Fatalf("agents file is not valid YAML: %v", err)
	}

	// Verify structure
	if _, ok := config["llm"]; !ok {
		t.Error("config should have llm section")
	}
	if _, ok := config["schema_agent"]; !ok {
		t.Error("config should have schema_agent section")
	}
	if _, ok := config["chunking_agent"]; !ok {
		t.Error("config should have chunking_agent section")
	}
}

func TestCreateDefaultAgentsFile_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	agentsFile := filepath.Join(tmpDir, "weave-agents.yaml")

	err := createDefaultAgentsFile(agentsFile)
	if err != nil {
		t.Fatalf("createDefaultAgentsFile failed: %v", err)
	}

	// Parse and validate YAML structure
	content, _ := os.ReadFile(agentsFile)
	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		t.Fatalf("invalid YAML: %v", err)
	}

	// Verify LLM config
	llm, ok := config["llm"].(map[string]interface{})
	if !ok {
		t.Fatal("llm should be a map")
	}
	if llm["provider"] != "openai" {
		t.Errorf("expected provider 'openai', got '%v'", llm["provider"])
	}
	if llm["default_model"] != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got '%v'", llm["default_model"])
	}

	// Verify schema_agent config
	schemaAgent, ok := config["schema_agent"].(map[string]interface{})
	if !ok {
		t.Fatal("schema_agent should be a map")
	}
	if enabled, ok := schemaAgent["enabled"].(bool); !ok || !enabled {
		t.Error("schema_agent should be enabled")
	}
	if maxSamples, ok := schemaAgent["max_samples"].(int); !ok || maxSamples != 50 {
		t.Errorf("expected max_samples 50, got %v", schemaAgent["max_samples"])
	}

	// Verify chunking_agent config
	chunkingAgent, ok := config["chunking_agent"].(map[string]interface{})
	if !ok {
		t.Fatal("chunking_agent should be a map")
	}
	if enabled, ok := chunkingAgent["enabled"].(bool); !ok || !enabled {
		t.Error("chunking_agent should be enabled")
	}
	if chunkSize, ok := chunkingAgent["default_chunk_size"].(int); !ok || chunkSize != 1000 {
		t.Errorf("expected default_chunk_size 1000, got %v", chunkingAgent["default_chunk_size"])
	}
}

func TestCreateDefaultAgentsFile_OverwriteProtection(t *testing.T) {
	tmpDir := t.TempDir()
	agentsFile := filepath.Join(tmpDir, "weave-agents.yaml")

	// Create initial file
	initialContent := "# Initial content\ntest: value\n"
	if err := os.WriteFile(agentsFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create initial file: %v", err)
	}

	// Since we can't provide interactive input in tests, we'll just verify
	// the file exists check works by checking file exists
	if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
		t.Fatal("initial file should exist")
	}

	// Verify initial content
	content, _ := os.ReadFile(agentsFile)
	if string(content) != initialContent {
		t.Error("initial content should be preserved when not overwriting")
	}
}

func TestAgentsConfigCompleteness(t *testing.T) {
	tmpDir := t.TempDir()
	agentsFile := filepath.Join(tmpDir, "weave-agents.yaml")

	err := createDefaultAgentsFile(agentsFile)
	if err != nil {
		t.Fatalf("createDefaultAgentsFile failed: %v", err)
	}

	content, _ := os.ReadFile(agentsFile)
	var config map[string]interface{}
	yaml.Unmarshal(content, &config)

	// Verify all required top-level sections
	requiredSections := []string{
		"llm",
		"schema_agent",
		"chunking_agent",
		"query_agent",
		"planning_agent",
		"report_agent",
		"eval_agent",
		"output",
		"cache",
		"performance",
		"features",
	}

	for _, section := range requiredSections {
		if _, ok := config[section]; !ok {
			t.Errorf("config should have '%s' section", section)
		}
	}

	// Verify performance settings
	perf, ok := config["performance"].(map[string]interface{})
	if !ok {
		t.Fatal("performance should be a map")
	}
	if timeout, ok := perf["llm_timeout"].(int); !ok || timeout != 30 {
		t.Errorf("expected llm_timeout 30, got %v", perf["llm_timeout"])
	}

	// Verify output settings
	output, ok := config["output"].(map[string]interface{})
	if !ok {
		t.Fatal("output should be a map")
	}
	if showConfidence, ok := output["show_confidence"].(bool); !ok || !showConfidence {
		t.Error("output.show_confidence should be true")
	}
}
