// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// IsDevelopmentMode checks if we're running in development mode
// (i.e., running from the source repository)
func IsDevelopmentMode() bool {
	// Check if ./configs/agents/rag-agent.yaml exists (repo marker)
	repoAgentPath := "configs/agents/rag-agent.yaml"
	if _, err := os.Stat(repoAgentPath); err == nil {
		return true
	}
	return false
}

// GetDefaultAgentDir returns the appropriate default directory for agents
// based on whether we're in development mode or deployed
func GetDefaultAgentDir() (string, error) {
	if IsDevelopmentMode() {
		// Development mode: use ./configs/agents/
		return "configs/agents", nil
	}

	// Deployed mode: use ~/.weave-cli/agents/
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".weave-cli", "agents"), nil
}

// InitializeUserAgents ensures ~/.weave-cli/agents/ exists and contains
// built-in agents. This is called on first run of the deployed CLI.
func InitializeUserAgents() error {
	// Skip if in development mode
	if IsDevelopmentMode() {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	userAgentDir := filepath.Join(home, ".weave-cli", "agents")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(userAgentDir, 0755); err != nil {
		return fmt.Errorf("failed to create user agent directory: %w", err)
	}

	// Check if already initialized (marker file exists)
	markerFile := filepath.Join(userAgentDir, ".initialized")
	if _, err := os.Stat(markerFile); err == nil {
		// Already initialized, skip
		return nil
	}

	// Copy built-in agents from embedded data or installation directory
	builtInAgents := []string{"rag-agent.yaml", "qa-agent.yaml", "summarize-agent.yaml"}

	// First, try to find built-in agents in common installation locations
	sourceDirs := []string{
		filepath.Join(filepath.Dir(os.Args[0]), "configs", "agents"), // Next to binary
		"/usr/local/share/weave-cli/agents",                          // System-wide install
		"/opt/weave-cli/agents",                                      // Alternative install
	}

	var sourceDir string
	for _, dir := range sourceDirs {
		if _, err := os.Stat(filepath.Join(dir, "rag-agent.yaml")); err == nil {
			sourceDir = dir
			break
		}
	}

	if sourceDir == "" {
		// No built-in agents found, create marker anyway
		// User can create agents manually
		if err := os.WriteFile(markerFile, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create marker file: %w", err)
		}
		return nil
	}

	// Copy built-in agents
	for _, agentFile := range builtInAgents {
		srcPath := filepath.Join(sourceDir, agentFile)
		dstPath := filepath.Join(userAgentDir, agentFile)

		// Skip if already exists
		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", agentFile, err)
		}
	}

	// Create marker file
	if err := os.WriteFile(markerFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create marker file: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Sync to ensure data is written
	return destFile.Sync()
}

// EnsureAgentsInitialized is a convenience function that ensures
// the agent system is initialized. Safe to call multiple times.
func EnsureAgentsInitialized() error {
	return InitializeUserAgents()
}
