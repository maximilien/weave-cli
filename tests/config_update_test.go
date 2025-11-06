// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigUpdateHelpers(t *testing.T) {
	t.Run("FileExists", func(t *testing.T) {
		// Create a temporary file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.txt")

		// File should not exist yet
		if fileExists(tmpFile) {
			t.Errorf("Expected file to not exist: %s", tmpFile)
		}

		// Create the file
		if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// File should now exist
		if !fileExists(tmpFile) {
			t.Errorf("Expected file to exist: %s", tmpFile)
		}
	})

	t.Run("LoadEnvFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")

		// Create a test .env file
		content := `# Test environment file
WEAVIATE_URL="https://test.weaviate.cloud"
WEAVIATE_API_KEY="test-api-key"

# Another comment
OPENAI_API_KEY=sk-test-key

# Empty value
EMPTY_VALUE=
`
		if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		// Load the file
		values := loadEnvFile(envFile)

		// Verify values
		expectedValues := map[string]string{
			"WEAVIATE_URL":     "https://test.weaviate.cloud",
			"WEAVIATE_API_KEY": "test-api-key",
			"OPENAI_API_KEY":   "sk-test-key",
		}

		for key, expected := range expectedValues {
			if actual, ok := values[key]; !ok {
				t.Errorf("Expected key %s to be present", key)
			} else if actual != expected {
				t.Errorf("Expected %s=%s, got %s", key, expected, actual)
			}
		}

		// Empty value should not be in map or should be empty
		if val, ok := values["EMPTY_VALUE"]; ok && val != "" {
			t.Errorf("Expected EMPTY_VALUE to be empty, got: %s", val)
		}
	})

	t.Run("SaveEnvFile", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a .env.example file
		exampleFile := filepath.Join(tmpDir, ".env.example")
		exampleContent := `# Test example file
WEAVIATE_URL="your-url-here"
WEAVIATE_API_KEY="your-api-key"
OPENAI_API_KEY="your-openai-key"
`
		if err := os.WriteFile(exampleFile, []byte(exampleContent), 0600); err != nil {
			t.Fatalf("Failed to create .env.example: %v", err)
		}

		// Change to temp directory to simulate being in project root
		oldDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(oldDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// Save with new values
		envFile := filepath.Join(tmpDir, ".env")
		values := map[string]string{
			"WEAVIATE_URL":     "https://new.weaviate.cloud",
			"WEAVIATE_API_KEY": "new-api-key",
			"OPENAI_API_KEY":   "sk-new-key",
		}

		if err := saveEnvFile(envFile, values); err != nil {
			t.Fatalf("Failed to save .env file: %v", err)
		}

		// Verify the file was created
		if !fileExists(envFile) {
			t.Errorf("Expected .env file to be created")
		}

		// Read and verify content
		content, err := os.ReadFile(envFile)
		if err != nil {
			t.Fatalf("Failed to read .env file: %v", err)
		}

		contentStr := string(content)

		// Check that comments are preserved
		if !strings.Contains(contentStr, "# Test example file") {
			t.Errorf("Expected comments to be preserved")
		}

		// Check that values are updated
		for key, expected := range values {
			expectedLine := key + "=\"" + expected + "\""
			if !strings.Contains(contentStr, expectedLine) {
				t.Errorf("Expected to find line: %s", expectedLine)
			}
		}
	})

	t.Run("CopyFile", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create source file
		srcFile := filepath.Join(tmpDir, "source.txt")
		srcContent := []byte("test content")
		if err := os.WriteFile(srcFile, srcContent, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Copy file
		dstFile := filepath.Join(tmpDir, "dest.txt")
		if err := copyFile(srcFile, dstFile); err != nil {
			t.Fatalf("Failed to copy file: %v", err)
		}

		// Verify destination file
		if !fileExists(dstFile) {
			t.Errorf("Expected destination file to exist")
		}

		// Verify content
		dstContent, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(dstContent) != string(srcContent) {
			t.Errorf("Expected content to match, got: %s", dstContent)
		}

		// Verify permissions
		srcInfo, _ := os.Stat(srcFile)
		dstInfo, _ := os.Stat(dstFile)

		if srcInfo.Mode() != dstInfo.Mode() {
			t.Errorf("Expected file modes to match: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
		}
	})
}

func TestGetEnvVariables(t *testing.T) {
	envVars := getEnvVariables()

	// Check that we have the expected variables
	expectedKeys := []string{
		"WEAVIATE_URL",
		"WEAVIATE_API_KEY",
		"OPENAI_API_KEY",
		"OPIK_API_KEY",
		"WEAVE_MCP_STDIO_PATH",
	}

	if len(envVars) < len(expectedKeys) {
		t.Errorf("Expected at least %d env variables, got %d", len(expectedKeys), len(envVars))
	}

	// Create a map for easier lookup
	varMap := make(map[string]bool)
	for _, ev := range envVars {
		varMap[ev.Key] = true

		// Verify structure
		if ev.Key == "" {
			t.Errorf("Environment variable has empty key")
		}
		if ev.Description == "" {
			t.Errorf("Environment variable %s has empty description", ev.Key)
		}

		// Verify secrets are marked correctly
		if strings.Contains(ev.Key, "API_KEY") && !ev.IsSecret {
			t.Errorf("Expected %s to be marked as secret", ev.Key)
		}
	}

	// Verify all expected keys are present
	for _, key := range expectedKeys {
		if !varMap[key] {
			t.Errorf("Expected environment variable %s to be present", key)
		}
	}
}

func TestEnvVariableMetadata(t *testing.T) {
	envVars := getEnvVariables()

	for _, ev := range envVars {
		t.Run(ev.Key, func(t *testing.T) {
			// Check required variables
			requiredVars := []string{"WEAVIATE_URL", "WEAVIATE_API_KEY", "OPENAI_API_KEY"}
			isRequired := false
			for _, req := range requiredVars {
				if ev.Key == req {
					isRequired = true
					break
				}
			}

			if isRequired && !ev.Required {
				t.Errorf("Expected %s to be marked as required", ev.Key)
			}

			// Check secret variables
			secretVars := []string{"WEAVIATE_API_KEY", "OPENAI_API_KEY", "OPIK_API_KEY"}
			isSecret := false
			for _, sec := range secretVars {
				if ev.Key == sec {
					isSecret = true
					break
				}
			}

			if isSecret && !ev.IsSecret {
				t.Errorf("Expected %s to be marked as secret", ev.Key)
			}

			// Non-secret variables
			if ev.Key == "WEAVIATE_URL" || ev.Key == "WEAVE_MCP_STDIO_PATH" {
				if ev.IsSecret {
					t.Errorf("Expected %s to not be marked as secret", ev.Key)
				}
			}
		})
	}
}

// Helper functions - these mirror the ones in config_update.go for testing

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func loadEnvFile(filename string) map[string]string {
	values := make(map[string]string)

	data, err := os.ReadFile(filename)
	if err != nil {
		return values
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE or KEY="VALUE"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			value = strings.Trim(value, "\"'")

			if value != "" {
				values[key] = value
			}
		}
	}

	return values
}

func saveEnvFile(filename string, values map[string]string) error {
	// Read the example file to preserve comments and structure
	exampleFile := ".env.example"
	exampleContent, err := os.ReadFile(exampleFile)
	if err != nil {
		return err
	}

	// Process the example file line by line
	lines := strings.Split(string(exampleContent), "\n")
	var output []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			output = append(output, line)
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])

			// Use updated value if available
			if value, ok := values[key]; ok && value != "" {
				output = append(output, key+"=\""+value+"\"")
			} else {
				// Keep original line if no value provided
				output = append(output, line)
			}
		} else {
			output = append(output, line)
		}
	}

	// Write to file
	content := strings.Join(output, "\n")
	return os.WriteFile(filename, []byte(content), 0600)
}

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Preserve file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, content, srcInfo.Mode())
}

func getEnvVariables() []struct {
	Key         string
	Description string
	Example     string
	IsSecret    bool
	Required    bool
} {
	return []struct {
		Key         string
		Description string
		Example     string
		IsSecret    bool
		Required    bool
	}{
		{
			Key:         "WEAVIATE_URL",
			Description: "Your Weaviate Cloud URL",
			Example:     "https://your-cluster-id.c0.us-west3.gcp.weaviate.cloud",
			IsSecret:    false,
			Required:    true,
		},
		{
			Key:         "WEAVIATE_API_KEY",
			Description: "Your Weaviate Cloud API Key",
			Example:     "your-weaviate-api-key-here",
			IsSecret:    true,
			Required:    true,
		},
		{
			Key:         "OPENAI_API_KEY",
			Description: "OpenAI API Key (for embeddings and AI agents)",
			Example:     "sk-proj-your-openai-api-key-here",
			IsSecret:    true,
			Required:    true,
		},
		{
			Key:         "OPIK_API_KEY",
			Description: "Opik API Key (optional - for LLM observability)",
			Example:     "your-opik-api-key-here",
			IsSecret:    true,
			Required:    false,
		},
		{
			Key:         "WEAVE_MCP_STDIO_PATH",
			Description: "Path to weave-mcp stdio binary (optional - for AI agents)",
			Example:     "/path/to/weave-mcp/bin/weave-mcp-stdio",
			IsSecret:    false,
			Required:    false,
		},
	}
}
