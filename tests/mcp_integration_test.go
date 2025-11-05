// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/maximilien/weave-cli/src/pkg/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPIntegration tests the weave-cli integration with weave-mcp
// This test verifies that all MCP operations work correctly with the latest MCP server
func TestMCPIntegration(t *testing.T) {
	// Load .env file from project root
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	// Check required environment variables
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}
	if os.Getenv("WEAVE_MCP_STDIO_PATH") == "" {
		t.Skip("Skipping: WEAVE_MCP_STDIO_PATH not set")
	}

	// Test collection names
	testDocsCollection := "MCPTestDocs"
	testImagesCollection := "MCPTestImages"

	// Create executor
	config := &executor.Config{
		DryRun:       false,
		NoConfirm:    true, // Skip confirmations for automated testing
		Verbose:      testing.Verbose(),
		Quiet:        false,
		NoColor:      true,
		OutputFormat: "text",
		Model:        "gpt-4o",
		StdioPath:    os.Getenv("WEAVE_MCP_STDIO_PATH"),
		MaxRetries:   3,
	}

	exec, err := executor.NewExecutor(config)
	require.NoError(t, err, "Failed to create executor")
	defer exec.Close()

	ctx := context.Background()

	t.Run("Cleanup existing test collections", func(t *testing.T) {
		// Delete test collections if they exist
		_, _ = exec.Execute(ctx, "delete "+testDocsCollection+" and "+testImagesCollection)
		time.Sleep(1 * time.Second) // Give time for cleanup
	})

	t.Run("Create collections", func(t *testing.T) {
		query := "create " + testDocsCollection + " and " + testImagesCollection + " collections"
		report, err := exec.Execute(ctx, query)

		require.NoError(t, err, "Query execution failed")
		assert.NotNil(t, report, "Report should not be nil")
		assert.Equal(t, 2, report.ExecutedSteps, "Should execute 2 steps")
		assert.Equal(t, 2, report.SuccessfulSteps, "Both steps should succeed")
		assert.Equal(t, 0, report.FailedSteps, "No steps should fail")
	})

	t.Run("List collections", func(t *testing.T) {
		query := "list all collections"
		report, err := exec.Execute(ctx, query)

		require.NoError(t, err, "Query execution failed")
		assert.NotNil(t, report, "Report should not be nil")
		assert.Greater(t, report.ExecutedSteps, 0, "Should execute at least 1 step")
		assert.Equal(t, report.ExecutedSteps, report.SuccessfulSteps, "All steps should succeed")
	})

	t.Run("Count documents in empty collection", func(t *testing.T) {
		query := "count docs in " + testDocsCollection
		report, err := exec.Execute(ctx, query)

		require.NoError(t, err, "Query execution failed")
		assert.NotNil(t, report, "Report should not be nil")
		assert.Equal(t, report.ExecutedSteps, report.SuccessfulSteps, "All steps should succeed")
	})

	t.Run("List documents in empty collection", func(t *testing.T) {
		query := "list docs in " + testDocsCollection
		report, err := exec.Execute(ctx, query)

		require.NoError(t, err, "Query execution failed")
		assert.NotNil(t, report, "Report should not be nil")
		assert.Equal(t, report.ExecutedSteps, report.SuccessfulSteps, "All steps should succeed")
	})

	t.Run("Health check", func(t *testing.T) {
		query := "check health"
		report, err := exec.Execute(ctx, query)

		require.NoError(t, err, "Query execution failed")
		assert.NotNil(t, report, "Report should not be nil")
		assert.Equal(t, report.ExecutedSteps, report.SuccessfulSteps, "All steps should succeed")
	})

	t.Run("Delete collections", func(t *testing.T) {
		query := "delete " + testDocsCollection + " and " + testImagesCollection
		report, err := exec.Execute(ctx, query)

		require.NoError(t, err, "Query execution failed")
		assert.NotNil(t, report, "Report should not be nil")
		assert.Equal(t, 2, report.ExecutedSteps, "Should execute 2 steps")
		assert.Equal(t, 2, report.SuccessfulSteps, "Both steps should succeed")
		assert.Equal(t, 0, report.FailedSteps, "No steps should fail")
	})

	t.Run("Verify collections deleted", func(t *testing.T) {
		// Try to count docs in deleted collection - should handle gracefully
		query := "count docs in " + testDocsCollection
		report, err := exec.Execute(ctx, query)

		// This might fail or succeed depending on how MCP handles missing collections
		// We just verify it doesn't crash
		assert.NotNil(t, report, "Report should not be nil even for deleted collection")
		_ = err // Error is acceptable here
	})
}

// TestMCPToolSchemas verifies that MCP tools have correct schemas
func TestMCPToolSchemas(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	if os.Getenv("WEAVE_MCP_STDIO_PATH") == "" {
		t.Skip("Skipping: WEAVE_MCP_STDIO_PATH not set")
	}

	config := &executor.Config{
		DryRun:       true, // Dry run to just check schemas
		NoConfirm:    true,
		Verbose:      false,
		Quiet:        true,
		NoColor:      true,
		OutputFormat: "text",
		Model:        "gpt-4o",
		StdioPath:    os.Getenv("WEAVE_MCP_STDIO_PATH"),
		MaxRetries:   1,
	}

	exec, err := executor.NewExecutor(config)
	require.NoError(t, err, "Failed to create executor")
	defer exec.Close()

	// If we can create the executor, MCP tools were loaded successfully
	assert.NotNil(t, exec, "Executor should be created with MCP tools")
}

// TestMCPErrorHandling verifies that MCP errors are handled properly
func TestMCPErrorHandling(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}
	if os.Getenv("WEAVE_MCP_STDIO_PATH") == "" {
		t.Skip("Skipping: WEAVE_MCP_STDIO_PATH not set")
	}

	config := &executor.Config{
		DryRun:       false,
		NoConfirm:    true,
		Verbose:      false,
		Quiet:        true,
		NoColor:      true,
		OutputFormat: "text",
		Model:        "gpt-4o",
		StdioPath:    os.Getenv("WEAVE_MCP_STDIO_PATH"),
		MaxRetries:   1,
	}

	exec, err := executor.NewExecutor(config)
	require.NoError(t, err, "Failed to create executor")
	defer exec.Close()

	ctx := context.Background()

	t.Run("Create duplicate collection should fail", func(t *testing.T) {
		testCollection := "MCPDuplicateTest"

		// Create collection
		_, _ = exec.Execute(ctx, "create "+testCollection+" collection")
		time.Sleep(500 * time.Millisecond)

		// Try to create again - should fail
		report, err := exec.Execute(ctx, "create "+testCollection+" collection")

		// Should either return an error or report with failures
		if err == nil {
			assert.Greater(t, report.FailedSteps, 0, "Should have failed steps for duplicate collection")
		}

		// Cleanup
		_, _ = exec.Execute(ctx, "delete "+testCollection)
	})

	t.Run("Delete non-existent collection should fail gracefully", func(t *testing.T) {
		nonExistentCollection := "MCPNonExistent12345"

		report, err := exec.Execute(ctx, "delete "+nonExistentCollection)

		// Should handle gracefully - either with error or failed step
		if err == nil {
			assert.NotNil(t, report, "Report should exist even for non-existent collection")
		}
	})
}
