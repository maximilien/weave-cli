// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

// ValidationWarning represents a non-critical configuration issue
type ValidationWarning struct {
	Field      string
	Message    string
	Suggestion string
}

// ValidationResult contains all validation findings
type ValidationResult struct {
	Warnings []ValidationWarning
	Errors   []ValidationWarning
}

// ValidateConfig performs comprehensive configuration validation
// Returns warnings but doesn't block execution (errors are only for critical issues)
func ValidateConfig(cfg *Config) *ValidationResult {
	if cfg == nil {
		return &ValidationResult{
			Errors: []ValidationWarning{{
				Field:      "config",
				Message:    "Configuration is nil",
				Suggestion: "Run 'weave config create --env' to create configuration",
			}},
		}
	}

	result := &ValidationResult{
		Warnings: []ValidationWarning{},
		Errors:   []ValidationWarning{},
	}

	// Validate database configuration
	if len(cfg.Databases.VectorDatabases) == 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "databases.vector_databases",
			Message:    "No vector databases configured",
			Suggestion: "Add at least one database or set environment variables (WEAVIATE_URL, WEAVIATE_API_KEY)",
		})
		return result
	}

	// Validate default database exists
	if cfg.Databases.Default != "" {
		found := false
		for _, db := range cfg.Databases.VectorDatabases {
			if db.Name == cfg.Databases.Default {
				found = true
				break
			}
		}
		if !found {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "databases.default",
				Message:    fmt.Sprintf("Default database '%s' not found in configured databases", cfg.Databases.Default),
				Suggestion: fmt.Sprintf("Set databases.default to one of: %v", listDatabaseNames(cfg)),
			})
		}
	}

	// Validate each database
	for i, db := range cfg.Databases.VectorDatabases {
		validateDatabase(&db, i, result)
	}

	// Check for deprecated settings
	checkDeprecatedSettings(cfg, result)

	return result
}

// validateDatabase validates a single database configuration
func validateDatabase(db *VectorDBConfig, index int, result *ValidationResult) {
	prefix := fmt.Sprintf("databases.vector_databases[%d] (%s)", index, db.Name)

	// Check database name
	if db.Name == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".name",
			Message:    "Database name is required",
			Suggestion: "Add 'name' field to database configuration",
		})
	}

	// Check database type
	if db.Type == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".type",
			Message:    "Database type is required",
			Suggestion: "Set 'type' to one of: weaviate-cloud, weaviate-local, milvus-cloud, milvus-local, etc.",
		})
		return // Can't validate further without type
	}

	// Type-specific validation
	switch db.Type {
	case VectorDBTypeCloud, VectorDBTypeLocal:
		validateWeaviate(db, prefix, result)
	case VectorDBTypeMilvusLocal, VectorDBTypeMilvusCloud:
		validateMilvus(db, prefix, result)
	case VectorDBTypeQdrantLocal, VectorDBTypeQdrantCloud:
		validateQdrant(db, prefix, result)
	case VectorDBTypeChromaLocal, VectorDBTypeChromaCloud:
		validateChroma(db, prefix, result)
	case VectorDBTypeSupabase, VectorDBTypeSupabaseCloud, VectorDBTypeSupabaseLocal:
		validateSupabase(db, prefix, result)
	case VectorDBTypeMongoDB, VectorDBTypeMongoDBCloud:
		validateMongoDB(db, prefix, result)
	case VectorDBTypeNeo4jLocal, VectorDBTypeNeo4jCloud:
		validateNeo4j(db, prefix, result)
	case VectorDBTypeOpenSearchLocal, VectorDBTypeOpenSearchCloud:
		validateOpenSearch(db, prefix, result)
	case VectorDBTypeMock:
		// Mock database has minimal requirements
		if db.EmbeddingDimension <= 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      prefix + ".embedding_dimension",
				Message:    "Mock database should have embedding_dimension set",
				Suggestion: "Add 'embedding_dimension: 384' to mock database config",
			})
		}
	}

	// Common timeout validation
	if db.Timeout < 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".timeout",
			Message:    fmt.Sprintf("Invalid timeout value: %d", db.Timeout),
			Suggestion: "Set timeout to positive value (in seconds) or omit for default (10s)",
		})
	}
}

// validateWeaviate validates Weaviate-specific configuration
func validateWeaviate(db *VectorDBConfig, prefix string, result *ValidationResult) {
	// Check URL
	if db.URL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".url",
			Message:    "Weaviate URL is required",
			Suggestion: "Set 'url' field or WEAVIATE_URL environment variable",
		})
	} else if err := validateURLConfig(db.URL); err != nil {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".url",
			Message:    fmt.Sprintf("Invalid URL format: %v", err),
			Suggestion: "Ensure URL starts with http:// or https://",
		})
	}

	// Check API key for cloud
	if db.Type == VectorDBTypeCloud && db.APIKey == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".api_key",
			Message:    "Weaviate Cloud typically requires an API key",
			Suggestion: "Set 'api_key' field or WEAVIATE_API_KEY environment variable",
		})
	}

	// Suggest OpenAI key if missing
	if db.OpenAIAPIKey == "" && os.Getenv("OPENAI_API_KEY") == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".openai_api_key",
			Message:    "OpenAI API key not configured (needed for embeddings)",
			Suggestion: "Set 'openai_api_key' field or OPENAI_API_KEY environment variable",
		})
	}
}

// validateMilvus validates Milvus-specific configuration
func validateMilvus(db *VectorDBConfig, prefix string, result *ValidationResult) {
	// Check address/URL
	if db.Address == "" && db.URL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".address",
			Message:    "Milvus address or URL is required",
			Suggestion: "Set 'address' field (e.g., 'localhost:19530') or 'url' field",
		})
	}

	// Check cloud credentials
	if db.Type == VectorDBTypeMilvusCloud {
		if db.APIKey == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      prefix + ".api_key",
				Message:    "Milvus Cloud (Zilliz) requires an API key",
				Suggestion: "Set 'api_key' field for authentication",
			})
		}
	}

	// Warn about vector dimensions (but skip if weave-stack.yaml exists - dimensions are managed there)
	if db.VectorDimensions <= 0 {
		// Skip this warning if weave-stack.yaml exists (stack commands manage dimensions separately)
		if _, err := os.Stat("weave-stack.yaml"); err == nil {
			// weave-stack.yaml exists, dimensions are managed there, skip warning
			return
		}

		// Provide smart suggestions based on common embedding models
		suggestion := "Set 'vector_dimensions' to match your embedding model:\n" +
			"    • text-embedding-3-small (OpenAI): 1536\n" +
			"    • text-embedding-3-large (OpenAI): 3072\n" +
			"    • all-MiniLM-L6-v2 (sentence-transformers): 384\n" +
			"    • CLIP (multimodal): 512\n" +
			"    Or run: weave config update --env"

		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".vector_dimensions",
			Message:    "Vector dimensions not set (required for embeddings)",
			Suggestion: suggestion,
		})
	}
}

// validateQdrant validates Qdrant-specific configuration
func validateQdrant(db *VectorDBConfig, prefix string, result *ValidationResult) {
	if db.URL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".url",
			Message:    "Qdrant URL is required",
			Suggestion: "Set 'url' field (e.g., 'http://localhost:6333' for local)",
		})
	} else if err := validateURLConfig(db.URL); err != nil {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".url",
			Message:    fmt.Sprintf("Invalid URL format: %v", err),
			Suggestion: "Ensure URL starts with http:// or https://",
		})
	}

	// Check API key for cloud
	if db.Type == VectorDBTypeQdrantCloud && db.APIKey == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".api_key",
			Message:    "Qdrant Cloud requires an API key",
			Suggestion: "Set 'api_key' field for authentication",
		})
	}
}

// validateChroma validates Chroma-specific configuration
func validateChroma(db *VectorDBConfig, prefix string, result *ValidationResult) {
	// Check URL
	if db.URL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".url",
			Message:    "Chroma URL is required",
			Suggestion: "Set 'url' field (e.g., 'http://localhost:8000' for local)",
		})
	}

	// Check cloud credentials
	if db.Type == VectorDBTypeChromaCloud {
		if db.APIKey == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      prefix + ".api_key",
				Message:    "Chroma Cloud requires an API key",
				Suggestion: "Set 'api_key' field for authentication",
			})
		}
		if db.Tenant == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      prefix + ".tenant",
				Message:    "Chroma Cloud requires a tenant ID",
				Suggestion: "Set 'tenant' field to your Chroma Cloud tenant ID",
			})
		}
	}
}

// validateSupabase validates Supabase-specific configuration
func validateSupabase(db *VectorDBConfig, prefix string, result *ValidationResult) {
	if db.DatabaseURL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".database_url",
			Message:    "Supabase database_url is required",
			Suggestion: "Set 'database_url' field with PostgreSQL connection string",
		})
	} else {
		// Basic PostgreSQL URL validation
		if !strings.HasPrefix(db.DatabaseURL, "postgres://") && !strings.HasPrefix(db.DatabaseURL, "postgresql://") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      prefix + ".database_url",
				Message:    "Database URL should be a PostgreSQL connection string",
				Suggestion: "Format: postgresql://user:password@host:port/database",
			})
		}
	}
}

// validateMongoDB validates MongoDB-specific configuration
func validateMongoDB(db *VectorDBConfig, prefix string, result *ValidationResult) {
	if db.DatabaseURL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".database_url",
			Message:    "MongoDB database_url is required",
			Suggestion: "Set 'database_url' field with MongoDB Atlas connection string",
		})
	} else if !strings.HasPrefix(db.DatabaseURL, "mongodb://") && !strings.HasPrefix(db.DatabaseURL, "mongodb+srv://") {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".database_url",
			Message:    "Database URL should be a MongoDB connection string",
			Suggestion: "Format: mongodb+srv://user:password@cluster.mongodb.net/database",
		})
	}

	if db.Database == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".database",
			Message:    "Database name not specified",
			Suggestion: "Set 'database' field to your MongoDB database name",
		})
	}
}

// validateNeo4j validates Neo4j-specific configuration
func validateNeo4j(db *VectorDBConfig, prefix string, result *ValidationResult) {
	if db.URL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".url",
			Message:    "Neo4j URL is required",
			Suggestion: "Set 'url' field (e.g., 'bolt://localhost:7687' for local)",
		})
	}

	if db.Username == "" || db.Password == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      prefix + ".username/password",
			Message:    "Neo4j typically requires username and password",
			Suggestion: "Set 'username' and 'password' fields for authentication",
		})
	}
}

// validateOpenSearch validates OpenSearch-specific configuration
func validateOpenSearch(db *VectorDBConfig, prefix string, result *ValidationResult) {
	if db.URL == "" {
		result.Errors = append(result.Errors, ValidationWarning{
			Field:      prefix + ".url",
			Message:    "OpenSearch URL is required",
			Suggestion: "Set 'url' field (e.g., 'https://localhost:9200' for local)",
		})
	}
}

// validateURL validates that a string is a valid URL
func validateURLConfig(rawURL string) error {
	// Handle environment variable interpolation
	if strings.Contains(rawURL, "${") {
		return nil // Skip validation for templated URLs
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("missing scheme (http:// or https://)")
	}

	if parsed.Host == "" {
		return fmt.Errorf("missing host")
	}

	return nil
}

// checkDeprecatedSettings warns about deprecated configuration options
func checkDeprecatedSettings(cfg *Config, result *ValidationResult) {
	for i, db := range cfg.Databases.VectorDatabases {
		prefix := fmt.Sprintf("databases.vector_databases[%d]", i)

		// Check for legacy type names
		if db.Type == "supabase" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      prefix + ".type",
				Message:    "Type 'supabase' is deprecated",
				Suggestion: "Use 'supabase-cloud' or 'supabase-local' instead",
			})
		}
		if db.Type == "mongodb" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      prefix + ".type",
				Message:    "Type 'mongodb' is deprecated",
				Suggestion: "Use 'mongodb-cloud' or 'mongodb-local' instead",
			})
		}
	}
}

// listDatabaseNames returns a comma-separated list of database names
func listDatabaseNames(cfg *Config) string {
	names := make([]string, len(cfg.Databases.VectorDatabases))
	for i, db := range cfg.Databases.VectorDatabases {
		names[i] = db.Name
	}
	return strings.Join(names, ", ")
}

// PrintValidationResult prints validation warnings and errors to stderr
// Only prints if there are issues to report (quiet by default)
// Respects quiet-config setting from CLI flag, config.yaml, or environment
func PrintValidationResult(result *ValidationResult) {
	if result == nil {
		return
	}

	// Check if quiet-config is set (via --quiet-config flag or quiet_config in config.yaml)
	if viper.GetBool("quiet-config") {
		return
	}

	hasIssues := len(result.Errors) > 0 || len(result.Warnings) > 0
	if !hasIssues {
		return // Silent if everything is OK
	}

	// Print errors (critical)
	if len(result.Errors) > 0 {
		color.New(color.FgRed, color.Bold).Fprintln(os.Stderr, "\n⚠️  Configuration Errors:")
		for _, err := range result.Errors {
			fmt.Fprintf(os.Stderr, "  • %s: %s\n", color.RedString(err.Field), err.Message)
			if err.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "    💡 %s\n", color.YellowString(err.Suggestion))
			}
		}
	}

	// Print warnings (non-critical)
	if len(result.Warnings) > 0 {
		color.New(color.FgYellow, color.Bold).Fprintln(os.Stderr, "\n⚠️  Configuration Warnings:")
		for _, warn := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  • %s: %s\n", color.YellowString(warn.Field), warn.Message)
			if warn.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "    💡 %s\n", color.CyanString(warn.Suggestion))
			}
		}
	}

	// Helpful footer
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Run 'weave config show' to see current configuration")
	fmt.Fprintln(os.Stderr, "Run 'weave config update --env' to fix configuration interactively")
	fmt.Fprintln(os.Stderr)
}
