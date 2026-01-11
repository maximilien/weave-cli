// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AgentInfo represents metadata about an available agent
type AgentInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Version     string `json:"version"`
	FilePath    string `json:"file_path"`
}

// AgentRegistry manages discovery and validation of available agents
type AgentRegistry struct {
	loader *AgentLoader
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry(loader *AgentLoader) *AgentRegistry {
	return &AgentRegistry{
		loader: loader,
	}
}

// ListAgents discovers and returns all available agents
func (r *AgentRegistry) ListAgents() ([]AgentInfo, error) {
	var agents []AgentInfo
	seen := make(map[string]bool)

	for _, searchPath := range r.loader.GetSearchPaths() {
		// Check if directory exists
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}

		// Read directory
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			// Only process YAML files
			if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
				continue
			}

			// Extract agent name (remove extension)
			agentName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")

			// Skip if already seen (first path wins)
			if seen[agentName] {
				continue
			}
			seen[agentName] = true

			// Try to load agent to get metadata
			fullPath := filepath.Join(searchPath, name)
			config, err := LoadCustomAgentConfig(fullPath)
			if err != nil {
				// If we can't load it, skip it (might be invalid)
				continue
			}

			agents = append(agents, AgentInfo{
				Name:        config.Name,
				Type:        config.Type,
				Description: config.Description,
				Version:     config.Version,
				FilePath:    fullPath,
			})
		}
	}

	return agents, nil
}

// GetAgentInfo returns metadata about a specific agent without fully loading it
func (r *AgentRegistry) GetAgentInfo(name string) (*AgentInfo, error) {
	agentPath, err := r.loader.FindAgentFile(name)
	if err != nil {
		return nil, err
	}

	config, err := LoadCustomAgentConfig(agentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent info: %w", err)
	}

	return &AgentInfo{
		Name:        config.Name,
		Type:        config.Type,
		Description: config.Description,
		Version:     config.Version,
		FilePath:    agentPath,
	}, nil
}

// ValidateAgent validates an agent configuration
func (r *AgentRegistry) ValidateAgent(name string) error {
	_, err := r.loader.LoadAgent(name)
	return err
}

// AgentExists checks if an agent exists in the registry
func (r *AgentRegistry) AgentExists(name string) bool {
	_, err := r.loader.FindAgentFile(name)
	return err == nil
}

// GetAgentsByType returns all agents of a specific type
func (r *AgentRegistry) GetAgentsByType(agentType string) ([]AgentInfo, error) {
	allAgents, err := r.ListAgents()
	if err != nil {
		return nil, err
	}

	var filtered []AgentInfo
	for _, agent := range allAgents {
		if agent.Type == agentType {
			filtered = append(filtered, agent)
		}
	}

	return filtered, nil
}

// DefaultAgentRegistry is the global default agent registry instance
var defaultAgentRegistry *AgentRegistry

// GetDefaultAgentRegistry returns the default global agent registry
func GetDefaultAgentRegistry() *AgentRegistry {
	if defaultAgentRegistry == nil {
		defaultAgentRegistry = NewAgentRegistry(GetDefaultAgentLoader())
	}
	return defaultAgentRegistry
}

// ListAvailableAgents is a convenience function that uses the default registry
func ListAvailableAgents() ([]AgentInfo, error) {
	return GetDefaultAgentRegistry().ListAgents()
}

// ValidateAgentByName is a convenience function that uses the default registry
func ValidateAgentByName(name string) error {
	return GetDefaultAgentRegistry().ValidateAgent(name)
}
