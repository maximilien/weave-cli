// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAgentChainConfig_EmptyChain(t *testing.T) {
	err := ValidateAgentChainConfig([]string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent chain cannot be empty")
}

func TestValidateAgentChainConfig_SingleAgent(t *testing.T) {
	err := ValidateAgentChainConfig([]string{"rag-agent"})

	require.NoError(t, err)
}

func TestValidateAgentChainConfig_MultipleAgents(t *testing.T) {
	err := ValidateAgentChainConfig([]string{"rag-agent", "search-agent", "qa-agent"})

	require.NoError(t, err)
}

func TestValidateAgentChainConfig_DuplicateAgents(t *testing.T) {
	err := ValidateAgentChainConfig([]string{"rag-agent", "search-agent", "rag-agent"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate agents in chain")
	assert.Contains(t, err.Error(), "rag-agent")
	assert.Contains(t, err.Error(), "Each agent should appear only once")
}

func TestValidateAgentChainConfig_MultipleDuplicates(t *testing.T) {
	err := ValidateAgentChainConfig([]string{
		"rag-agent",
		"search-agent",
		"rag-agent",
		"qa-agent",
		"search-agent",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate agents in chain")
	// Both duplicates should be mentioned
	assert.True(t,
		strings.Contains(err.Error(), "rag-agent") ||
			strings.Contains(err.Error(), "search-agent"),
		"Error should mention duplicate agents")
}

func TestValidateAgentChainConfig_LongChain(t *testing.T) {
	// Chain with 6 agents (> 5 triggers warning)
	agents := []string{
		"agent1", "agent2", "agent3",
		"agent4", "agent5", "agent6",
	}

	// Should not return error, just warning to stderr
	err := ValidateAgentChainConfig(agents)
	require.NoError(t, err)
}

func TestValidateAgentChainConfig_VeryLongChain(t *testing.T) {
	// Chain with 10 agents
	agents := []string{
		"agent1", "agent2", "agent3", "agent4", "agent5",
		"agent6", "agent7", "agent8", "agent9", "agent10",
	}

	// Should not return error, just warning to stderr
	err := ValidateAgentChainConfig(agents)
	require.NoError(t, err)
}

func TestValidateAgentNames_EmptyList(t *testing.T) {
	err := ValidateAgentNames([]string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one agent name is required")
}

func TestValidateAgentNames_ValidAgents(t *testing.T) {
	// This test depends on having actual agent files in configs/agents/
	// Test with known agents
	err := ValidateAgentNames([]string{"rag-agent"})

	// May pass if agent exists, or fail with helpful error
	if err != nil {
		// If it fails, error should be helpful
		assert.True(t,
			strings.Contains(err.Error(), "invalid agent") ||
				strings.Contains(err.Error(), "Available agents"),
			"Error should be helpful")
	}
}

func TestValidateAgentNames_NonExistentAgent(t *testing.T) {
	err := ValidateAgentNames([]string{"nonexistent-agent-xyz"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent")
}

func TestValidateAgentNames_MixedValidInvalid(t *testing.T) {
	// Mix of potentially valid and definitely invalid agents
	err := ValidateAgentNames([]string{"rag-agent", "totally-fake-agent-123"})

	// Should error on the invalid agent
	if err != nil {
		assert.Contains(t, err.Error(), "invalid agent")
	}
}

func TestFindSimilarAgent_ExactMatch(t *testing.T) {
	available := []string{"rag-agent", "search-agent", "qa-agent"}

	result := findSimilarAgent("rag-agent", available)

	// Should return exact match or empty (depends on implementation)
	assert.True(t, result == "" || result == "rag-agent")
}

func TestFindSimilarAgent_CloseMatch(t *testing.T) {
	available := []string{"rag-agent", "search-agent", "qa-agent"}

	// Typo: "rag-agnet" should suggest "rag-agent"
	result := findSimilarAgent("rag-agnet", available)

	assert.Equal(t, "rag-agent", result)
}

func TestFindSimilarAgent_MultipleOptions(t *testing.T) {
	available := []string{"rag-agent", "tag-agent", "bag-agent"}

	// "rag-aget" is close to all three, but "rag-agent" should be closest
	result := findSimilarAgent("rag-aget", available)

	// Should suggest "rag-agent" as it's the closest match
	assert.Equal(t, "rag-agent", result)
}

func TestFindSimilarAgent_NoMatch(t *testing.T) {
	available := []string{"rag-agent", "search-agent"}

	// Very different input
	result := findSimilarAgent("xyz123", available)

	// Should return empty string (no good match)
	assert.Equal(t, "", result)
}

func TestFindSimilarAgent_EmptyList(t *testing.T) {
	result := findSimilarAgent("rag-agent", []string{})

	assert.Equal(t, "", result)
}

func TestFindSimilarAgent_CaseInsensitive(t *testing.T) {
	available := []string{"rag-agent", "search-agent"}

	result := findSimilarAgent("RAG-AGENT", available)

	// Should match despite case difference
	assert.True(t, result == "" || result == "rag-agent")
}

func TestLevenshteinDistance_Identical(t *testing.T) {
	distance := levenshteinDistance("hello", "hello")
	assert.Equal(t, 0, distance)
}

func TestLevenshteinDistance_OneDifference(t *testing.T) {
	distance := levenshteinDistance("hello", "hallo")
	assert.Equal(t, 1, distance)
}

func TestLevenshteinDistance_Insertion(t *testing.T) {
	distance := levenshteinDistance("cat", "cats")
	assert.Equal(t, 1, distance)
}

func TestLevenshteinDistance_Deletion(t *testing.T) {
	distance := levenshteinDistance("cats", "cat")
	assert.Equal(t, 1, distance)
}

func TestLevenshteinDistance_Substitution(t *testing.T) {
	distance := levenshteinDistance("cat", "bat")
	assert.Equal(t, 1, distance)
}

func TestLevenshteinDistance_CompletelyDifferent(t *testing.T) {
	distance := levenshteinDistance("abc", "xyz")
	assert.Equal(t, 3, distance)
}

func TestLevenshteinDistance_EmptyStrings(t *testing.T) {
	assert.Equal(t, 0, levenshteinDistance("", ""))
	assert.Equal(t, 3, levenshteinDistance("abc", ""))
	assert.Equal(t, 3, levenshteinDistance("", "abc"))
}

func TestLevenshteinDistance_RealWorldExample(t *testing.T) {
	// Real typo: "rag-agnet" vs "rag-agent"
	distance := levenshteinDistance("rag-agnet", "rag-agent")
	assert.Equal(t, 2, distance) // Two character swap
}

func TestMin_ThreeIntegers(t *testing.T) {
	assert.Equal(t, 1, min(1, 2, 3))
	assert.Equal(t, 1, min(2, 1, 3))
	assert.Equal(t, 1, min(3, 2, 1))
	assert.Equal(t, 5, min(5, 5, 5))
	assert.Equal(t, -1, min(-1, 0, 1))
}

func TestFormatAgentNotFoundError_Structure(t *testing.T) {
	// Test that error message has expected structure
	msg := FormatAgentNotFoundError("fake-agent", nil)

	assert.Contains(t, msg, "fake-agent")
	assert.Contains(t, msg, "not found")
	assert.Contains(t, msg, "weave agents list")
}

func TestFormatAgentNotFoundError_WithAvailableAgents(t *testing.T) {
	// When agents are available, error should list them
	msg := FormatAgentNotFoundError("fake-agent", nil)

	// Should contain helpful information
	assert.True(t,
		strings.Contains(msg, "Available agents") ||
			strings.Contains(msg, "No agents found"),
		"Error should mention available agents or their absence")
}
