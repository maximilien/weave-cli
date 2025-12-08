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
	VDBTypes    []string // List of VDB types this variable applies to (empty = all)
}

// getEnvVariables returns the list of environment variables to configure
// If vdbFilter is specified, only returns variables for that VDB + common variables
func getEnvVariables(vdbFilter string) []envVariable {
	allVars := []envVariable{
		// Weaviate Configuration
		{
			Key:         "WEAVIATE_URL",
			Description: "Your Weaviate Cloud/Local URL",
			Example:     "https://your-cluster-id.c0.us-west3.gcp.weaviate.cloud",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"weaviate", "weaviate-cloud", "weaviate-local"},
		},
		{
			Key:         "WEAVIATE_API_KEY",
			Description: "Your Weaviate Cloud API Key (not needed for local)",
			Example:     "your-weaviate-api-key-here",
			IsSecret:    true,
			Required:    false,
			VDBTypes:    []string{"weaviate", "weaviate-cloud"},
		},
		// Supabase Configuration
		{
			Key:         "SUPABASE_DATABASE_URL",
			Description: "Supabase PostgreSQL connection URL (pooler recommended)",
			Example:     "postgres://postgres:password@db.project.supabase.co:6543/postgres",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"supabase", "supabase-cloud", "supabase-local"},
		},
		{
			Key:         "SUPABASE_DATABASE_KEY",
			Description: "Supabase Anon/Service Key",
			Example:     "eyJhbGc...",
			IsSecret:    true,
			Required:    false,
			VDBTypes:    []string{"supabase", "supabase-cloud", "supabase-local"},
		},
		// MongoDB Configuration
		{
			Key:         "MONGODB_URI",
			Description: "MongoDB Atlas connection string",
			Example:     "mongodb+srv://username:password@cluster.mongodb.net/?appName=weave-cli",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"mongodb", "mongodb-cloud"},
		},
		{
			Key:         "MONGODB_DATABASE",
			Description: "MongoDB database name",
			Example:     "weave-cli",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"mongodb", "mongodb-cloud"},
		},
		// Milvus Local Configuration
		{
			Key:         "MILVUS_LOCAL_ADDRESS",
			Description: "Milvus local server address",
			Example:     "localhost:19530",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"milvus-local"},
		},
		{
			Key:         "MILVUS_LOCAL_DATABASE",
			Description: "Milvus local database name",
			Example:     "default",
			IsSecret:    false,
			Required:    false,
			VDBTypes:    []string{"milvus-local"},
		},
		// Milvus Cloud (Zilliz) Configuration
		{
			Key:         "MILVUS_CLOUD_ADDRESS",
			Description: "Milvus Cloud (Zilliz) server address",
			Example:     "your-cluster.aws-us-west-2.vectordb.zillizcloud.com:19530",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"milvus-cloud"},
		},
		{
			Key:         "MILVUS_CLOUD_USERNAME",
			Description: "Milvus Cloud username",
			Example:     "db_admin",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"milvus-cloud"},
		},
		{
			Key:         "MILVUS_CLOUD_PASSWORD",
			Description: "Milvus Cloud password",
			Example:     "your-secure-password",
			IsSecret:    true,
			Required:    true,
			VDBTypes:    []string{"milvus-cloud"},
		},
		{
			Key:         "MILVUS_CLOUD_DATABASE",
			Description: "Milvus Cloud database name",
			Example:     "default",
			IsSecret:    false,
			Required:    false,
			VDBTypes:    []string{"milvus-cloud"},
		},
		// Chroma Local Configuration
		{
			Key:         "CHROMA_LOCAL_ADDRESS",
			Description: "Chroma local server address",
			Example:     "localhost:8000",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"chroma-local"},
		},
		// Chroma Cloud Configuration
		{
			Key:         "CHROMA_CLOUD_ADDRESS",
			Description: "Chroma Cloud server address",
			Example:     "api.trychroma.com",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"chroma-cloud"},
		},
		{
			Key:         "CHROMA_CLOUD_API_KEY",
			Description: "Chroma Cloud API key",
			Example:     "your-chroma-api-key",
			IsSecret:    true,
			Required:    true,
			VDBTypes:    []string{"chroma-cloud"},
		},
		// Qdrant Local Configuration
		{
			Key:         "QDRANT_LOCAL_ADDRESS",
			Description: "Qdrant local server address",
			Example:     "localhost:6333",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"qdrant-local"},
		},
		// Qdrant Cloud Configuration
		{
			Key:         "QDRANT_CLOUD_ADDRESS",
			Description: "Qdrant Cloud cluster URL",
			Example:     "xyz-example.eu-central.aws.cloud.qdrant.io",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"qdrant-cloud"},
		},
		{
			Key:         "QDRANT_CLOUD_API_KEY",
			Description: "Qdrant Cloud API key",
			Example:     "your-qdrant-api-key",
			IsSecret:    true,
			Required:    true,
			VDBTypes:    []string{"qdrant-cloud"},
		},
		// Neo4j Local Configuration
		{
			Key:         "NEO4J_LOCAL_URI",
			Description: "Neo4j local server URI",
			Example:     "bolt://localhost:7687",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"neo4j-local"},
		},
		{
			Key:         "NEO4J_LOCAL_USERNAME",
			Description: "Neo4j local username",
			Example:     "neo4j",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"neo4j-local"},
		},
		{
			Key:         "NEO4J_LOCAL_PASSWORD",
			Description: "Neo4j local password",
			Example:     "your-password",
			IsSecret:    true,
			Required:    true,
			VDBTypes:    []string{"neo4j-local"},
		},
		// Neo4j Cloud (Aura) Configuration
		{
			Key:         "NEO4J_CLOUD_URI",
			Description: "Neo4j Aura connection URI",
			Example:     "neo4j+s://xxxxx.databases.neo4j.io",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"neo4j-cloud"},
		},
		{
			Key:         "NEO4J_CLOUD_USERNAME",
			Description: "Neo4j Aura username",
			Example:     "neo4j",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"neo4j-cloud"},
		},
		{
			Key:         "NEO4J_CLOUD_PASSWORD",
			Description: "Neo4j Aura password",
			Example:     "your-secure-password",
			IsSecret:    true,
			Required:    true,
			VDBTypes:    []string{"neo4j-cloud"},
		},
		// OpenSearch Local Configuration
		{
			Key:         "OPENSEARCH_LOCAL_ADDRESS",
			Description: "OpenSearch local server address",
			Example:     "http://localhost:9200",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"opensearch-local"},
		},
		// OpenSearch Cloud Configuration
		{
			Key:         "OPENSEARCH_CLOUD_ADDRESS",
			Description: "OpenSearch Cloud/AWS endpoint URL",
			Example:     "https://your-domain.us-east-1.es.amazonaws.com",
			IsSecret:    false,
			Required:    true,
			VDBTypes:    []string{"opensearch-cloud"},
		},
		{
			Key:         "OPENSEARCH_CLOUD_USERNAME",
			Description: "OpenSearch Cloud username (if using basic auth)",
			Example:     "admin",
			IsSecret:    false,
			Required:    false,
			VDBTypes:    []string{"opensearch-cloud"},
		},
		{
			Key:         "OPENSEARCH_CLOUD_PASSWORD",
			Description: "OpenSearch Cloud password (if using basic auth)",
			Example:     "your-secure-password",
			IsSecret:    true,
			Required:    false,
			VDBTypes:    []string{"opensearch-cloud"},
		},
		{
			Key:         "OPENSEARCH_CLOUD_API_KEY",
			Description: "OpenSearch Cloud API key (alternative to username/password)",
			Example:     "your-api-key",
			IsSecret:    true,
			Required:    false,
			VDBTypes:    []string{"opensearch-cloud"},
		},
		// OpenAI Configuration (common to all)
		{
			Key:         "OPENAI_API_KEY",
			Description: "OpenAI API Key (for embeddings and AI agents)",
			Example:     "sk-proj-your-openai-api-key-here",
			IsSecret:    true,
			Required:    true,
			VDBTypes:    []string{}, // Empty = applies to all
		},
		// Optional Services (common to all)
		{
			Key:         "OPIK_API_KEY",
			Description: "Opik API Key (optional - for LLM observability)",
			Example:     "your-opik-api-key-here",
			IsSecret:    true,
			Required:    false,
			VDBTypes:    []string{}, // Empty = applies to all
		},
		{
			Key:         "WEAVE_MCP_STDIO_PATH",
			Description: "Path to weave-mcp stdio binary (optional - for AI agents)",
			Example:     "/path/to/weave-mcp/bin/weave-mcp-stdio",
			IsSecret:    false,
			Required:    false,
			VDBTypes:    []string{}, // Empty = applies to all
		},
	}

	// If no filter, return all variables
	if vdbFilter == "" {
		return allVars
	}

	// Filter variables by VDB type
	var filtered []envVariable
	for _, v := range allVars {
		// Include if VDBTypes is empty (common variable) or matches filter
		if len(v.VDBTypes) == 0 || containsString(v.VDBTypes, vdbFilter) {
			filtered = append(filtered, v)
		}
	}

	return filtered
}

// containsString checks if a string slice contains a specific string
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
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
