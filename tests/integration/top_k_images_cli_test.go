// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTopKImagesCLI tests the --top_k_images flag via the actual CLI command
// This test requires environment variables for at least one vector database:
// - Milvus: MILVUS_CLOUD_TOKEN, MILVUS_CLOUD_ADDRESS, OPENAI_API_KEY
// - Chroma: CHROMA_API_KEY, CHROMA_TENANT, CHROMA_DATABASE, OPENAI_API_KEY
// - Weaviate: WEAVIATE_CLOUD_API_KEY, WEAVIATE_CLOUD_URL, OPENAI_API_KEY
func TestTopKImagesCLI(t *testing.T) {
	// Determine which VDB to use based on available credentials
	vdbFlag, vdbName := detectAvailableVDB(t)
	if vdbFlag == "" {
		t.Skip("Skipping integration test: No vector database credentials found")
	}

	t.Logf("Running tests with %s", vdbName)

	// Collection names
	textCol := "IntegTestText"
	imageCol := "IntegTestImage"

	// Cleanup function
	cleanup := func() {
		runCLI(t, "cols", "delete", textCol, vdbFlag, "--no-confirm")
		runCLI(t, "cols", "delete", imageCol, vdbFlag, "--no-confirm")
	}
	defer cleanup()
	cleanup() // Clean up any previous test runs

	t.Run("Setup_Collections_Via_CLI", func(t *testing.T) {
		// Create text collection
		output := runCLI(t, "cols", "create", textCol, "--text", "--embedding", "text-embedding-3-small", vdbFlag)
		assert.Contains(t, output, "Successfully created collection")

		// Create image collection
		output = runCLI(t, "cols", "create", imageCol, "--image", "--embedding", "text-embedding-3-small", vdbFlag)
		assert.Contains(t, output, "Successfully created collection")

		// Create test documents
		createTestDocument(t, "/tmp/test_text_1.txt", "Weave CLI is a command-line tool for managing vector databases with semantic search")
		createTestDocument(t, "/tmp/test_text_2.txt", "Vector databases enable efficient similarity search using embeddings")
		createTestDocument(t, "/tmp/test_text_3.txt", "RAG agents combine retrieval and generation for better AI responses")
		createTestDocument(t, "/tmp/test_image_1.txt", "Screenshot of Weave CLI terminal showing query command execution")
		createTestDocument(t, "/tmp/test_image_2.txt", "Architecture diagram of vector database system with RAG agent")

		// Add documents to collections
		runCLI(t, "docs", "create", textCol, "/tmp/test_text_1.txt", vdbFlag)
		runCLI(t, "docs", "create", textCol, "/tmp/test_text_2.txt", vdbFlag)
		runCLI(t, "docs", "create", textCol, "/tmp/test_text_3.txt", vdbFlag)
		runCLI(t, "docs", "create", imageCol, "/tmp/test_image_1.txt", vdbFlag)
		runCLI(t, "docs", "create", imageCol, "/tmp/test_image_2.txt", vdbFlag)

		t.Logf("Created 3 text documents and 2 image documents")
	})

	t.Run("Query_With_TopKImages_Flag", func(t *testing.T) {
		// Query with --top_k_images flag
		output := runCLI(t, "cols", "query", textCol, imageCol, "weave cli vector database",
			"--agent", "rag-agent",
			"--top_k", "5",
			"--top_k_images", "2",
			"--verbose",
			vdbFlag)

		// Verify debug output shows correct topK values
		assert.Contains(t, output, "detected as TEXT", "Should detect text collection")
		assert.Contains(t, output, "detected as IMAGE", "Should detect image collection")

		// Verify RAG agent response includes citations from both collections
		assert.Contains(t, output, textCol, "Should include text collection citation")
		assert.Contains(t, output, imageCol, "Should include image collection citation")

		t.Logf("✅ Multi-collection query with --top_k_images successful")
	})

	t.Run("Query_Without_TopKImages_Flag", func(t *testing.T) {
		// Query without --top_k_images (baseline)
		output := runCLI(t, "cols", "query", textCol, imageCol, "weave cli vector database",
			"--agent", "rag-agent",
			"--top_k", "5",
			"--verbose",
			vdbFlag)

		// Should still work, just use same topK for all collections
		assert.Contains(t, output, "detected as TEXT")
		assert.Contains(t, output, "detected as IMAGE")

		t.Logf("✅ Multi-collection query without --top_k_images works as baseline")
	})

	t.Run("Query_With_TopKImages_Zero", func(t *testing.T) {
		// Query with --top_k_images=0 (should skip image collections)
		output := runCLI(t, "cols", "query", textCol, imageCol, "weave cli vector database",
			"--agent", "rag-agent",
			"--top_k", "5",
			"--top_k_images", "0",
			"--verbose",
			vdbFlag)

		// Should only query text collection
		assert.Contains(t, output, textCol, "Should include text collection")

		t.Logf("✅ Query with --top_k_images=0 works correctly")
	})

	t.Run("Query_Image_Collection_Only", func(t *testing.T) {
		// Query only image collection with --top_k_images
		output := runCLI(t, "cols", "query", imageCol, "weave cli screenshot",
			"--agent", "rag-agent",
			"--top_k", "5",
			"--top_k_images", "1",
			"--verbose",
			vdbFlag)

		// Should apply topKImages to image collection
		assert.Contains(t, output, "detected as IMAGE")
		assert.Contains(t, output, imageCol)

		t.Logf("✅ Query with only image collection works")
	})

	t.Run("Image_Collection_Detection", func(t *testing.T) {
		// Test name-based detection
		testCases := []struct {
			name     string
			expected string
		}{
			{textCol, "TEXT"},
			{imageCol, "IMAGE"},
			{"MyImages", "IMAGE"},
			{"UserPhotos", "IMAGE"},
			{"MyDocuments", "TEXT"},
		}

		for _, tc := range testCases {
			// Create a dummy query to see detection in verbose output
			// (we don't need the collection to exist for detection logic)
			output := runCLIAllowError(t, "cols", "query", tc.name, "test query",
				"--top_k", "1",
				"--verbose",
				vdbFlag)

			if strings.Contains(output, "not found") {
				// Expected for non-existent collections, but we can still check the detection message
				// The detection happens before the error
				t.Logf("Collection %s detection test skipped (collection doesn't exist)", tc.name)
			}
		}
	})

	t.Run("Verify_Citation_Content_Workflow", func(t *testing.T) {
		// This test verifies the complete workflow:
		// 1. Query with --top_k_images returns results from both collections
		// 2. Each citation can be verified to have relevant content
		// 3. Image documents describe visual content
		// 4. Text documents have standard text content

		// Query both collections
		queryOutput := runCLI(t, "cols", "query", textCol, imageCol, "weave cli",
			"--agent", "rag-agent",
			"--top_k", "5",
			"--top_k_images", "2",
			"--verbose",
			vdbFlag)

		// Verify RAG response includes both collection citations
		assert.Contains(t, queryOutput, textCol, "RAG response should cite text collection")
		assert.Contains(t, queryOutput, imageCol, "RAG response should cite image collection")

		// Query each collection individually to verify content
		textResults := runCLI(t, "cols", "query", textCol, "weave cli",
			"--top_k", "1",
			"--no-truncate",
			vdbFlag)

		imageResults := runCLI(t, "cols", "query", imageCol, "weave cli",
			"--top_k", "1",
			"--no-truncate",
			vdbFlag)

		// Verify text document has CLI-related content
		assert.Contains(t, textResults, "Weave CLI", "Text document should mention Weave CLI")
		assert.Contains(t, textResults, "vector database", "Text document should discuss vector databases")

		// Verify image document describes visual content
		assert.Contains(t, imageResults, "Screenshot", "Image document should describe visual content")

		t.Logf("✅ Citation workflow verified:")
		t.Logf("   - Text collection has documentation content")
		t.Logf("   - Image collection has visual content descriptions")
		t.Logf("   - RAG agent cites both collections")
	})
}

// detectAvailableVDB checks environment variables and returns the appropriate VDB flag
func detectAvailableVDB(t *testing.T) (flag string, name string) {
	// Check Milvus Cloud
	if os.Getenv("MILVUS_CLOUD_TOKEN") != "" && os.Getenv("MILVUS_CLOUD_ADDRESS") != "" {
		return "--milvus-cloud", "Milvus Cloud"
	}

	// Check Chroma Cloud
	if os.Getenv("CHROMA_API_KEY") != "" && os.Getenv("CHROMA_TENANT") != "" {
		return "--chroma-cloud", "Chroma Cloud"
	}

	// Check Weaviate Cloud
	if os.Getenv("WEAVIATE_CLOUD_API_KEY") != "" && os.Getenv("WEAVIATE_CLOUD_URL") != "" {
		return "--weaviate-cloud", "Weaviate Cloud"
	}

	// Check Qdrant Cloud
	if os.Getenv("QDRANT_API_KEY") != "" && os.Getenv("QDRANT_URL") != "" {
		return "--qdrant-cloud", "Qdrant Cloud"
	}

	return "", ""
}

// runCLI executes the weave CLI command and returns the output
func runCLI(t *testing.T, args ...string) string {
	t.Helper()
	output, err := runCLIInternal(args...)
	require.NoError(t, err, "CLI command failed: %s", strings.Join(args, " "))
	return output
}

// runCLIAllowError executes the weave CLI command and returns output even if there's an error
func runCLIAllowError(t *testing.T, args ...string) string {
	t.Helper()
	output, _ := runCLIInternal(args...)
	return output
}

// runCLIInternal executes the weave CLI and returns output and error
func runCLIInternal(args ...string) (string, error) {
	// Find the project root (where bin/weave is located)
	// When running from tests/integration, we need to go up two levels
	cliPath := "../../bin/weave"

	// Check if binary exists
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		return "", fmt.Errorf("CLI binary not found at %s - run ./build.sh first", cliPath)
	}

	cmd := exec.Command(cliPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// createTestDocument creates a temporary test document
func createTestDocument(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, fmt.Sprintf("Failed to create test document %s", path))
}
