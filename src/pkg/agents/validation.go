// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// getAvailableAgentNames returns a list of available agent names
func getAvailableAgentNames() ([]string, error) {
	agents, err := ListAvailableAgents()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(agents))
	for i, agent := range agents {
		names[i] = agent.Name
	}
	return names, nil
}

// ValidateAgentNames validates a list of agent names and returns helpful errors
func ValidateAgentNames(agentNames []string) error {
	if len(agentNames) == 0 {
		return fmt.Errorf("at least one agent name is required")
	}

	// Get available agents
	availableAgents, err := getAvailableAgentNames()
	if err != nil {
		// If we can't list agents, proceed with loading (will fail with detailed error)
		return nil
	}

	// Check for invalid agent names
	var invalidAgents []string
	var suggestions []string

	for _, name := range agentNames {
		// Try to load the agent
		_, err := LoadAgent(name)
		if err != nil {
			invalidAgents = append(invalidAgents, name)

			// Find similar agent names
			if suggestion := findSimilarAgent(name, availableAgents); suggestion != "" {
				suggestions = append(suggestions, fmt.Sprintf("  '%s' → did you mean '%s'?", name, suggestion))
			}
		}
	}

	if len(invalidAgents) > 0 {
		errMsg := fmt.Sprintf("invalid agent name(s): %s\n\n", strings.Join(invalidAgents, ", "))

		if len(suggestions) > 0 {
			errMsg += "Suggestions:\n" + strings.Join(suggestions, "\n") + "\n\n"
		}

		errMsg += fmt.Sprintf("Available agents: %s\n", strings.Join(availableAgents, ", "))
		errMsg += "\nUse 'weave agents list' to see all available agents."

		return errors.New(errMsg)
	}

	return nil
}

// findSimilarAgent finds the most similar agent name using Levenshtein distance
func findSimilarAgent(input string, availableAgents []string) string {
	if len(availableAgents) == 0 {
		return ""
	}

	minDistance := len(input) + 10 // arbitrary high value
	var bestMatch string

	for _, agent := range availableAgents {
		distance := levenshteinDistance(strings.ToLower(input), strings.ToLower(agent))

		// Only suggest if distance is reasonable (< 50% of input length)
		threshold := len(input) / 2
		if threshold < 2 {
			threshold = 2
		}

		if distance < minDistance && distance <= threshold {
			minDistance = distance
			bestMatch = agent
		}
	}

	return bestMatch
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Calculate distances
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// ValidateAgentChainConfig validates agent chain configuration
func ValidateAgentChainConfig(agentNames []string) error {
	if len(agentNames) == 0 {
		return fmt.Errorf("agent chain cannot be empty")
	}

	// Check for duplicates
	seen := make(map[string]bool)
	var duplicates []string

	for _, name := range agentNames {
		if seen[name] {
			duplicates = append(duplicates, name)
		}
		seen[name] = true
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate agents in chain: %s\nNote: Each agent should appear only once in the chain", strings.Join(duplicates, ", "))
	}

	// Warn if chain is very long (> 5 agents)
	if len(agentNames) > 5 {
		// Not an error, just a note for the user
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Chain has %d agents. Long chains may be slow and expensive.\n", len(agentNames))
	}

	return nil
}

// FormatAgentNotFoundError improves the error message for agent not found errors
func FormatAgentNotFoundError(agentName string, err error) string {
	availableAgents, listErr := getAvailableAgentNames()

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("❌ Agent '%s' not found.\n\n", agentName))

	if listErr == nil && len(availableAgents) > 0 {
		// Try to find similar agent
		if suggestion := findSimilarAgent(agentName, availableAgents); suggestion != "" {
			msg.WriteString(fmt.Sprintf("💡 Did you mean '%s'?\n\n", suggestion))
		}

		msg.WriteString("Available agents:\n")
		for _, agent := range availableAgents {
			msg.WriteString(fmt.Sprintf("  • %s\n", agent))
		}
	} else {
		msg.WriteString("No agents found in search paths.\n")
		msg.WriteString("\nTo create agents, see: docs/guides/CUSTOM_AGENTS.md\n")
	}

	msg.WriteString("\n💡 Use 'weave agents list' to see all available agents.\n")

	return msg.String()
}
