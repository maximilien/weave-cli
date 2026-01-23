// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// AgentLoader handles loading and caching custom agents
type AgentLoader struct {
	searchPaths []string
	cache       map[string]*CustomAgentConfig
	cacheMutex  sync.RWMutex
}

// NewAgentLoader creates a new agent loader with standard search paths
func NewAgentLoader() *AgentLoader {
	loader := &AgentLoader{
		searchPaths: []string{},
		cache:       make(map[string]*CustomAgentConfig),
	}

	// Ensure agents are initialized (for deployed mode)
	// This is safe to call multiple times and is a no-op in development mode
	_ = EnsureAgentsInitialized()

	// Add standard search paths
	loader.AddStandardSearchPaths()

	return loader
}

// AddStandardSearchPaths adds the standard agent search paths
func (l *AgentLoader) AddStandardSearchPaths() {
	// 1. Current directory configs/agents/
	l.AddSearchPath("configs/agents")

	// 2. User home directory ~/.weave-cli/agents/
	if home, err := os.UserHomeDir(); err == nil {
		l.AddSearchPath(filepath.Join(home, ".weave-cli", "agents"))
	}

	// 3. System-wide /etc/weave-cli/agents/ (Unix-like systems)
	if _, err := os.Stat("/etc/weave-cli/agents"); err == nil {
		l.AddSearchPath("/etc/weave-cli/agents")
	}
}

// AddSearchPath adds a directory to the agent search paths
func (l *AgentLoader) AddSearchPath(path string) {
	// Check if path already exists
	for _, existing := range l.searchPaths {
		if existing == path {
			return
		}
	}
	l.searchPaths = append(l.searchPaths, path)
}

// GetSearchPaths returns the current search paths
func (l *AgentLoader) GetSearchPaths() []string {
	return l.searchPaths
}

// LoadAgent loads an agent configuration by name
// Searches in all configured search paths and caches the result
func (l *AgentLoader) LoadAgent(name string) (*CustomAgentConfig, error) {
	// Check cache first
	l.cacheMutex.RLock()
	if cached, ok := l.cache[name]; ok {
		l.cacheMutex.RUnlock()
		return cached, nil
	}
	l.cacheMutex.RUnlock()

	// Search for agent file
	agentPath, err := l.FindAgentFile(name)
	if err != nil {
		return nil, err
	}

	// Load agent config
	config, err := LoadCustomAgentConfig(agentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent '%s': %w", name, err)
	}

	// Verify agent name matches file
	if config.Name != name {
		return nil, fmt.Errorf("agent name in config (%s) does not match requested name (%s)", config.Name, name)
	}

	// Cache the loaded agent
	l.cacheMutex.Lock()
	l.cache[name] = config
	l.cacheMutex.Unlock()

	return config, nil
}

// FindAgentFile searches for an agent YAML file in all search paths
func (l *AgentLoader) FindAgentFile(name string) (string, error) {
	var searchedPaths []string

	for _, searchPath := range l.searchPaths {
		// Try with .yaml extension
		yamlPath := filepath.Join(searchPath, name+".yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			return yamlPath, nil
		}
		searchedPaths = append(searchedPaths, yamlPath)

		// Try with .yml extension
		ymlPath := filepath.Join(searchPath, name+".yml")
		if _, err := os.Stat(ymlPath); err == nil {
			return ymlPath, nil
		}
		searchedPaths = append(searchedPaths, ymlPath)
	}

	return "", &AgentNotFoundError{
		AgentName:     name,
		SearchedPaths: searchedPaths,
	}
}

// ReloadAgent reloads an agent from disk, bypassing the cache
func (l *AgentLoader) ReloadAgent(name string) (*CustomAgentConfig, error) {
	// Clear from cache
	l.cacheMutex.Lock()
	delete(l.cache, name)
	l.cacheMutex.Unlock()

	// Load fresh from disk
	return l.LoadAgent(name)
}

// ClearCache clears all cached agents
func (l *AgentLoader) ClearCache() {
	l.cacheMutex.Lock()
	l.cache = make(map[string]*CustomAgentConfig)
	l.cacheMutex.Unlock()
}

// GetCachedAgent returns a cached agent without loading from disk
func (l *AgentLoader) GetCachedAgent(name string) (*CustomAgentConfig, bool) {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()
	config, ok := l.cache[name]
	return config, ok
}

// AgentNotFoundError represents an error when an agent cannot be found
type AgentNotFoundError struct {
	AgentName     string
	SearchedPaths []string
}

func (e *AgentNotFoundError) Error() string {
	return fmt.Sprintf("agent '%s' not found in search paths: %v", e.AgentName, e.SearchedPaths)
}

// IsAgentNotFoundError checks if an error is an AgentNotFoundError
func IsAgentNotFoundError(err error) bool {
	_, ok := err.(*AgentNotFoundError)
	return ok
}

// DefaultAgentLoader is the global default agent loader instance
var (
	defaultAgentLoader     *AgentLoader
	defaultAgentLoaderOnce sync.Once
)

// GetDefaultAgentLoader returns the default global agent loader
func GetDefaultAgentLoader() *AgentLoader {
	defaultAgentLoaderOnce.Do(func() {
		defaultAgentLoader = NewAgentLoader()
	})
	return defaultAgentLoader
}

// LoadAgent is a convenience function that uses the default agent loader
func LoadAgent(name string) (*CustomAgentConfig, error) {
	return GetDefaultAgentLoader().LoadAgent(name)
}
