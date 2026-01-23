// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

//go:build integration
// +build integration

package agents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"gopkg.in/yaml.v3"
)

func TestAgentLifecycle_CreateValidateDelete(t *testing.T) {
	// Create temporary directory for test agents
	tmpDir := t.TempDir()
	agentName := "test-lifecycle-agent"
	agentPath := filepath.Join(tmpDir, agentName+".yaml")

	t.Run("Create RAG agent", func(t *testing.T) {
		runCreateAgent(agentName, "rag", false, tmpDir)

		// Verify file was created
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			t.Fatalf("Expected agent file to be created at %s", agentPath)
		}
	})

	t.Run("Validate created agent", func(t *testing.T) {
		// Load and validate the created agent
		config, err := agents.LoadCustomAgentConfig(agentPath)
		if err != nil {
			t.Fatalf("Failed to load created agent: %v", err)
		}

		if config.Name != agentName {
			t.Errorf("Expected agent name '%s', got '%s'", agentName, config.Name)
		}

		if config.Type != "rag" {
			t.Errorf("Expected agent type 'rag', got '%s'", config.Type)
		}

		if err := config.Validate(); err != nil {
			t.Errorf("Created agent failed validation: %v", err)
		}
	})

	t.Run("Delete created agent", func(t *testing.T) {
		// Delete the agent file
		if err := os.Remove(agentPath); err != nil {
			t.Fatalf("Failed to delete agent: %v", err)
		}

		// Verify file was deleted
		if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
			t.Error("Expected agent file to be deleted")
		}
	})
}

func TestAgentLifecycle_CopyAndModify(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source agent
	sourceName := "test-source-agent"
	targetName := "test-target-agent"
	sourcePath := filepath.Join(tmpDir, sourceName+".yaml")
	targetPath := filepath.Join(tmpDir, targetName+".yaml")

	t.Run("Create source agent", func(t *testing.T) {
		runCreateAgent(sourceName, "qa", false, tmpDir)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			t.Fatalf("Failed to create source agent")
		}
	})

	t.Run("Copy agent manually", func(t *testing.T) {
		// Load source config
		sourceConfig, err := agents.LoadCustomAgentConfig(sourcePath)
		if err != nil {
			t.Fatalf("Failed to load source agent: %v", err)
		}

		// Create copy with new name
		sourceConfig.Name = targetName
		sourceConfig.Version = "1.0.0"
		sourceConfig.Author = ""

		// Write to target path
		data, err := yaml.Marshal(sourceConfig)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			t.Fatalf("Failed to write target file: %v", err)
		}

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Fatalf("Failed to copy agent")
		}
	})

	t.Run("Verify copied agent is independent", func(t *testing.T) {
		sourceConfig, err := agents.LoadCustomAgentConfig(sourcePath)
		if err != nil {
			t.Fatalf("Failed to load source agent: %v", err)
		}

		targetConfig, err := agents.LoadCustomAgentConfig(targetPath)
		if err != nil {
			t.Fatalf("Failed to load target agent: %v", err)
		}

		// Names should be different
		if sourceConfig.Name == targetConfig.Name {
			t.Error("Copied agent should have different name")
		}

		// Types should be the same
		if sourceConfig.Type != targetConfig.Type {
			t.Error("Copied agent should have same type as source")
		}

		// Modify target and verify source is unchanged
		targetConfig.Description = "Modified description"

		data, _ := yaml.Marshal(targetConfig)
		os.WriteFile(targetPath, data, 0644)

		// Reload source
		sourceConfigReloaded, _ := agents.LoadCustomAgentConfig(sourcePath)
		if sourceConfigReloaded.Description == "Modified description" {
			t.Error("Source agent should not be affected by target modifications")
		}
	})
}

func TestAgentCreation_AllTypes(t *testing.T) {
	tmpDir := t.TempDir()

	types := []string{"rag", "qa", "summarize", "custom"}

	for _, agentType := range types {
		t.Run("Create "+agentType+" agent", func(t *testing.T) {
			agentName := "test-" + agentType + "-agent"
			runCreateAgent(agentName, agentType, false, tmpDir)

			agentPath := filepath.Join(tmpDir, agentName+".yaml")

			// Verify creation
			if _, err := os.Stat(agentPath); os.IsNotExist(err) {
				t.Fatalf("Failed to create %s agent", agentType)
			}

			// Verify content
			config, err := agents.LoadCustomAgentConfig(agentPath)
			if err != nil {
				t.Fatalf("Failed to load %s agent: %v", agentType, err)
			}

			if config.Type != agentType {
				t.Errorf("Expected type '%s', got '%s'", agentType, config.Type)
			}

			// Validate
			if err := config.Validate(); err != nil {
				t.Errorf("%s agent failed validation: %v", agentType, err)
			}
		})
	}
}

func TestAgentCreation_DuplicateNameError(t *testing.T) {
	tmpDir := t.TempDir()
	agentName := "test-duplicate"

	// Create first agent
	runCreateAgent(agentName, "rag", false, tmpDir)

	// Attempt to create duplicate should not overwrite
	// (This test would need to capture stderr to verify the error message)
	// For now, we just verify the original file is still valid

	agentPath := filepath.Join(tmpDir, agentName+".yaml")
	config, err := agents.LoadCustomAgentConfig(agentPath)
	if err != nil {
		t.Fatalf("Original agent was corrupted: %v", err)
	}

	if config.Name != agentName {
		t.Error("Original agent name should be unchanged")
	}
}

func TestAgentDevelopmentVsDeployedMode(t *testing.T) {
	t.Run("Development mode detection", func(t *testing.T) {
		// Save current directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(originalDir)

		// Find repo root
		dir := originalDir
		for {
			testPath := filepath.Join(dir, "configs", "agents", "rag-agent.yaml")
			if _, err := os.Stat(testPath); err == nil {
				// Found repo root
				os.Chdir(dir)
				break
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				t.Skip("Not in repository, skipping development mode test")
			}
			dir = parent
		}

		if !agents.IsDevelopmentMode() {
			t.Error("Expected development mode in repository")
		}

		defaultDir, err := agents.GetDefaultAgentDir()
		if err != nil {
			t.Fatalf("Failed to get default agent dir: %v", err)
		}

		if defaultDir != "configs/agents" {
			t.Errorf("Expected 'configs/agents', got '%s'", defaultDir)
		}
	})

	t.Run("Deployed mode detection", func(t *testing.T) {
		// Save current directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(originalDir)

		// Change to temp directory
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)

		if agents.IsDevelopmentMode() {
			t.Error("Expected deployed mode in temp directory")
		}

		defaultDir, err := agents.GetDefaultAgentDir()
		if err != nil {
			t.Fatalf("Failed to get default agent dir: %v", err)
		}

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".weave-cli", "agents")

		if defaultDir != expected {
			t.Errorf("Expected '%s', got '%s'", expected, defaultDir)
		}
	})
}

func TestAgentValidation_InvalidConfigs(t *testing.T) {
	testCases := []struct {
		name        string
		modifyFunc  func(*agents.CustomAgentConfig)
		shouldError bool
	}{
		{
			name: "Valid config",
			modifyFunc: func(c *agents.CustomAgentConfig) {
				// No modifications
			},
			shouldError: false,
		},
		{
			name: "Missing name",
			modifyFunc: func(c *agents.CustomAgentConfig) {
				c.Name = ""
			},
			shouldError: true,
		},
		{
			name: "Invalid temperature",
			modifyFunc: func(c *agents.CustomAgentConfig) {
				c.LLM.Temperature = 3.0 // Too high
			},
			shouldError: true,
		},
		{
			name: "Invalid citation format",
			modifyFunc: func(c *agents.CustomAgentConfig) {
				c.Response.CitationFormat = "invalid"
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := agents.GetRAGTemplate()
			config.Name = "test-validation"
			tc.modifyFunc(config)

			err := config.Validate()
			if tc.shouldError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
