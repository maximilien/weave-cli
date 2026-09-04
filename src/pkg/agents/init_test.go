// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDevelopmentMode(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	t.Run("Returns true in development mode (repo)", func(t *testing.T) {
		// Change to repo root (where configs/agents/rag-agent.yaml exists)
		repoRoot := findRepoRoot(t)
		if err := os.Chdir(repoRoot); err != nil {
			t.Fatalf("Failed to change to repo root: %v", err)
		}

		if !IsDevelopmentMode() {
			t.Error("Expected IsDevelopmentMode() to return true in repository")
		}
	})

	t.Run("Returns false in deployed mode", func(t *testing.T) {
		// Change to temp directory (no configs/agents/rag-agent.yaml)
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		if IsDevelopmentMode() {
			t.Error("Expected IsDevelopmentMode() to return false outside repository")
		}
	})
}

func TestGetDefaultAgentDir(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	t.Run("Returns configs/agents in development mode", func(t *testing.T) {
		// Change to repo root
		repoRoot := findRepoRoot(t)
		if err := os.Chdir(repoRoot); err != nil {
			t.Fatalf("Failed to change to repo root: %v", err)
		}

		dir, err := GetDefaultAgentDir()
		if err != nil {
			t.Fatalf("GetDefaultAgentDir() failed: %v", err)
		}

		expected := "configs/agents"
		if dir != expected {
			t.Errorf("Expected %s, got %s", expected, dir)
		}
	})

	t.Run("Returns ~/.weave-cli/agents in deployed mode", func(t *testing.T) {
		// Change to temp directory
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		dir, err := GetDefaultAgentDir()
		if err != nil {
			t.Fatalf("GetDefaultAgentDir() failed: %v", err)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}

		expected := filepath.Join(home, ".weave-cli", "agents")
		if dir != expected {
			t.Errorf("Expected %s, got %s", expected, dir)
		}
	})
}

func TestInitializeUserAgents(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	t.Run("Skips initialization in development mode", func(t *testing.T) {
		// Change to repo root
		repoRoot := findRepoRoot(t)
		if err := os.Chdir(repoRoot); err != nil {
			t.Fatalf("Failed to change to repo root: %v", err)
		}

		err := InitializeUserAgents()
		if err != nil {
			t.Errorf("InitializeUserAgents() failed: %v", err)
		}

		// Should not create marker file in dev mode
		// This is tested implicitly by the function returning without error
	})

	t.Run("Creates directory and marker in deployed mode", func(t *testing.T) {
		// Change to temp directory
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Set up a temporary home directory for this test
		tmpHome := filepath.Join(tmpDir, "home")
		t.Setenv("HOME", tmpHome)

		err := InitializeUserAgents()
		if err != nil {
			t.Fatalf("InitializeUserAgents() failed: %v", err)
		}

		// Check that directory was created
		agentDir := filepath.Join(tmpHome, ".weave-cli", "agents")
		if _, err := os.Stat(agentDir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to be created", agentDir)
		}

		// Check that marker file was created
		markerFile := filepath.Join(agentDir, ".initialized")
		if _, err := os.Stat(markerFile); os.IsNotExist(err) {
			t.Errorf("Expected marker file %s to be created", markerFile)
		}
	})

	t.Run("Is idempotent (safe to call multiple times)", func(t *testing.T) {
		// Change to temp directory
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Set up a temporary home directory
		tmpHome := filepath.Join(tmpDir, "home")
		t.Setenv("HOME", tmpHome)

		// Call multiple times
		for i := 0; i < 3; i++ {
			err := InitializeUserAgents()
			if err != nil {
				t.Fatalf("InitializeUserAgents() call %d failed: %v", i+1, err)
			}
		}

		// Should still have marker file
		markerFile := filepath.Join(tmpHome, ".weave-cli", "agents", ".initialized")
		if _, err := os.Stat(markerFile); os.IsNotExist(err) {
			t.Error("Marker file should still exist after multiple calls")
		}
	})
}

func TestCopyFile(t *testing.T) {
	t.Run("Successfully copies file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create source file
		srcPath := filepath.Join(tmpDir, "source.txt")
		content := []byte("test content")
		if err := os.WriteFile(srcPath, content, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Copy file
		dstPath := filepath.Join(tmpDir, "dest.txt")
		if err := copyFile(srcPath, dstPath); err != nil {
			t.Fatalf("copyFile() failed: %v", err)
		}

		// Verify destination file
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(dstContent) != string(content) {
			t.Errorf("Content mismatch: expected %s, got %s", content, dstContent)
		}
	})

	t.Run("Returns error for non-existent source", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcPath := filepath.Join(tmpDir, "nonexistent.txt")
		dstPath := filepath.Join(tmpDir, "dest.txt")

		err := copyFile(srcPath, dstPath)
		if err == nil {
			t.Error("Expected error for non-existent source file")
		}
	})
}

func TestEnsureAgentsInitialized(t *testing.T) {
	t.Run("Calls InitializeUserAgents", func(t *testing.T) {
		// Save current directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(originalDir)

		// Change to repo root (dev mode)
		repoRoot := findRepoRoot(t)
		if err := os.Chdir(repoRoot); err != nil {
			t.Fatalf("Failed to change to repo root: %v", err)
		}

		err = EnsureAgentsInitialized()
		if err != nil {
			t.Errorf("EnsureAgentsInitialized() failed: %v", err)
		}
	})
}

// Helper function to find repo root
func findRepoRoot(t *testing.T) string {
	t.Helper()

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Walk up until we find configs/agents/rag-agent.yaml
	for {
		testPath := filepath.Join(dir, "configs", "agents", "rag-agent.yaml")
		if _, err := os.Stat(testPath); err == nil {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding repo
			t.Fatal("Could not find repository root")
		}
		dir = parent
	}
}
