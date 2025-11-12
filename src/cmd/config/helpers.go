// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// envVariable represents an environment variable with metadata
type envVariable struct {
	Key         string
	Description string
	Example     string
	IsSecret    bool
	Required    bool
}

// getEnvVariables returns the list of environment variables to configure
func getEnvVariables() []envVariable {
	return []envVariable{
		// Weaviate Configuration
		{
			Key:         "WEAVIATE_URL",
			Description: "Your Weaviate Cloud URL (optional if using Supabase)",
			Example:     "https://your-cluster-id.c0.us-west3.gcp.weaviate.cloud",
			IsSecret:    false,
			Required:    false,
		},
		{
			Key:         "WEAVIATE_API_KEY",
			Description: "Your Weaviate Cloud API Key (optional if using Supabase)",
			Example:     "your-weaviate-api-key-here",
			IsSecret:    true,
			Required:    false,
		},
		// Supabase Configuration
		{
			Key:         "SUPABASE_DATABASE_URL",
			Description: "Supabase PostgreSQL connection URL (optional if using Weaviate)",
			Example:     "postgres://postgres:password@db.project.supabase.co:6543/postgres",
			IsSecret:    false,
			Required:    false,
		},
		{
			Key:         "SUPABASE_DATABASE_KEY",
			Description: "Supabase Anon/Service Key (optional if using Weaviate)",
			Example:     "eyJhbGc...",
			IsSecret:    true,
			Required:    false,
		},
		// OpenAI Configuration
		{
			Key:         "OPENAI_API_KEY",
			Description: "OpenAI API Key (for embeddings and AI agents)",
			Example:     "sk-proj-your-openai-api-key-here",
			IsSecret:    true,
			Required:    true,
		},
		// Optional Services
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

// File operation helpers

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func loadEnvFile(filename string) map[string]string {
	values := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return values
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

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

			values[key] = value
		}
	}

	return values
}

func saveEnvFile(filename string, values map[string]string) error {
	// Read the existing .env file to preserve existing values and structure
	// If it doesn't exist, fall back to .env.example
	sourceFile := filename
	if !fileExists(filename) {
		sourceFile = ".env.example"
	}

	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sourceFile, err)
	}

	// Process the source file line by line
	lines := strings.Split(string(sourceContent), "\n")
	var output []string
	processedKeys := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle empty lines
		if trimmed == "" {
			output = append(output, line)
			continue
		}

		// Handle commented lines - check if they contain a key we want to set
		if strings.HasPrefix(trimmed, "#") {
			// Try to parse commented KEY=VALUE
			uncommented := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			parts := strings.SplitN(uncommented, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				// If we have a value for this key, uncomment and set it
				if value, ok := values[key]; ok && value != "" {
					output = append(output, fmt.Sprintf("%s=\"%s\"", key, value))
					processedKeys[key] = true
					continue
				}
			}
			// Otherwise keep the comment
			output = append(output, line)
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			processedKeys[key] = true

			// Use updated value if available
			if value, ok := values[key]; ok && value != "" {
				output = append(output, fmt.Sprintf("%s=\"%s\"", key, value))
			} else {
				// Keep original line if no value provided
				output = append(output, line)
			}
		} else {
			output = append(output, line)
		}
	}

	// Append any new keys that weren't in the source file
	for key, value := range values {
		if !processedKeys[key] && value != "" {
			output = append(output, fmt.Sprintf("%s=\"%s\"", key, value))
		}
	}

	// Write to file
	content := strings.Join(output, "\n")
	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
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

// Input/output helpers

func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readSecret() (string, error) {
	// Check if stdin is a terminal
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Not a terminal, read normally
		return readLine()
	}

	// Read password without echo
	password, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}

	fmt.Println() // New line after password input
	return string(password), nil
}

func confirmSave(filename string) bool {
	color.New(color.FgYellow, color.Bold).Printf("💾 Save changes to %s? (Y/n): ", filename)
	response, err := readLine()
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	// Default to yes if empty
	return response == "" || response == "y" || response == "yes"
}
