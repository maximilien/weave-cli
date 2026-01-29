// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"
)

// VectorDBType represents the type of vector database
type VectorDBType string

const (
	VectorDBTypeCloud           VectorDBType = "weaviate-cloud"
	VectorDBTypeLocal           VectorDBType = "weaviate-local"
	VectorDBTypeMock            VectorDBType = "mock"
	VectorDBTypeSupabase        VectorDBType = "supabase" // Legacy
	VectorDBTypeSupabaseCloud   VectorDBType = "supabase-cloud"
	VectorDBTypeSupabaseLocal   VectorDBType = "supabase-local"
	VectorDBTypeMongoDB         VectorDBType = "mongodb" // Legacy
	VectorDBTypeMongoDBCloud    VectorDBType = "mongodb-cloud"
	VectorDBTypeMilvusLocal     VectorDBType = "milvus-local"
	VectorDBTypeMilvusCloud     VectorDBType = "milvus-cloud"
	VectorDBTypeChromaLocal     VectorDBType = "chroma-local"
	VectorDBTypeChromaCloud     VectorDBType = "chroma-cloud"
	VectorDBTypeQdrantLocal     VectorDBType = "qdrant-local"
	VectorDBTypeQdrantCloud     VectorDBType = "qdrant-cloud"
	VectorDBTypeNeo4jLocal      VectorDBType = "neo4j-local"
	VectorDBTypeNeo4jCloud      VectorDBType = "neo4j-cloud"
	VectorDBTypeOpenSearchLocal VectorDBType = "opensearch-local"
	VectorDBTypeOpenSearchCloud VectorDBType = "opensearch-cloud"
)

// Default configuration values
const (
	DefaultTextCollection  = "WeaveDocs"
	DefaultImageCollection = "WeaveImages"
	DefaultDatabaseType    = "weaviate-cloud"
	DefaultLocalURL        = "http://localhost:8080"
	DefaultBatchWorkers    = 3
	DefaultRetryAttempts   = 3
	DefaultLLMModel        = "gpt-4o"
	DefaultLLMTimeout      = 300
	DefaultBatchChunkSize  = 500
	DefaultImageQuality    = 85
	DefaultMaxImageSize    = 2048
	DefaultEmbeddingDim    = 384
	DefaultTimeout         = 10 // Default timeout in seconds for vector DB operations
)

// Package-level variable to track which env file was loaded
var loadedEnvFile string

// Collection represents a collection configuration
type Collection struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description,omitempty"`
}

// MockCollection represents a mock collection (for backward compatibility)
type MockCollection struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// MockConfig holds mock database configuration (for backward compatibility)
type MockConfig struct {
	Enabled            bool             `yaml:"enabled"`
	SimulateEmbeddings bool             `yaml:"simulate_embeddings"`
	EmbeddingDimension int              `yaml:"embedding_dimension"`
	Collections        []MockCollection `yaml:"collections"`
}

// VectorDBConfig holds vector database configuration
type VectorDBConfig struct {
	Name               string       `yaml:"name"`
	Type               VectorDBType `yaml:"type"`
	URL                string       `yaml:"url,omitempty"`
	APIKey             string       `yaml:"api_key,omitempty"`
	DatabaseURL        string       `yaml:"database_url,omitempty"`      // For Supabase
	DatabaseKey        string       `yaml:"database_key,omitempty"`      // For Supabase
	Database           string       `yaml:"database,omitempty"`          // For MongoDB, Milvus, Chroma
	Tenant             string       `yaml:"tenant,omitempty"`            // For Chroma Cloud
	VectorDimensions   int          `yaml:"vector_dimensions,omitempty"` // For MongoDB, Milvus
	SimilarityMetric   string       `yaml:"similarity_metric,omitempty"` // For MongoDB, Milvus
	Address            string       `yaml:"address,omitempty"`           // For Milvus
	Username           string       `yaml:"username,omitempty"`          // For Milvus Cloud
	Password           string       `yaml:"password,omitempty"`          // For Milvus Cloud
	OpenAIAPIKey       string       `yaml:"openai_api_key,omitempty"`
	Timeout            int          `yaml:"timeout,omitempty"` // Timeout in seconds for DB operations
	Enabled            bool         `yaml:"enabled,omitempty"`
	SimulateEmbeddings bool         `yaml:"simulate_embeddings,omitempty"`
	EmbeddingDimension int          `yaml:"embedding_dimension,omitempty"`
	Collections        []Collection `yaml:"collections"`
}

// SchemaDefinition represents a named schema that can be used to create collections
type SchemaDefinition struct {
	Name     string                 `yaml:"name"`
	Schema   map[string]interface{} `yaml:"schema"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// DatabasesConfig holds multiple databases configuration
type DatabasesConfig struct {
	Default         string             `yaml:"default"`
	VectorDatabases []VectorDBConfig   `yaml:"vector_databases"`
	Schemas         []SchemaDefinition `yaml:"schemas,omitempty"`
}

// Config holds the complete application configuration
// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string `yaml:"level,omitempty"` // debug, info, warn, error
	File  string `yaml:"file,omitempty"`  // log file path (empty = stderr)
}

type Config struct {
	Databases  DatabasesConfig `yaml:"databases"`
	SchemasDir string          `yaml:"schemas_dir,omitempty"`
	Logging    LoggingConfig   `yaml:"logging,omitempty"`
}

// LoadConfig loads configuration from files and environment variables
// LoadConfigOptions holds options for loading configuration
type LoadConfigOptions struct {
	ConfigFile     string
	EnvFile        string
	VectorDBType   string
	WeaviateAPIKey string
	WeaviateURL    string
	Timeout        string // Timeout as duration string (e.g., "5s", "10s") or empty for default
}

func LoadConfig(configFile, envFile string) (*Config, error) {
	return LoadConfigWithOptions(LoadConfigOptions{
		ConfigFile: configFile,
		EnvFile:    envFile,
	})
}

func LoadConfigWithOptions(opts LoadConfigOptions) (*Config, error) {
	// Load environment variables with priority order:
	// 1. Command flags (highest priority)
	// 2. --env file
	// 3. Local .env file (current directory)
	// 4. Global .env file (~/.weave-cli/.env)
	// 5. Shell environment (lowest priority)

	// Load from --env file if specified
	if opts.EnvFile != "" {
		if err := godotenv.Load(opts.EnvFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %s: %w", opts.EnvFile, err)
		}
		loadedEnvFile = opts.EnvFile
	} else {
		// Find config paths with proper precedence
		configPaths, err := FindConfigPaths()
		if err == nil && configPaths.EnvPath != "" {
			// Try to load .env from the discovered path
			if err := godotenv.Load(configPaths.EnvPath); err == nil {
				// Only set loadedEnvFile if the file was successfully loaded
				if fileExists(configPaths.EnvPath) {
					loadedEnvFile = configPaths.EnvPath
				}
			}
		}
	}

	// Override with command-line flags (highest priority)
	if opts.VectorDBType != "" {
		os.Setenv("VECTOR_DB_TYPE", opts.VectorDBType)
	}
	if opts.WeaviateAPIKey != "" {
		os.Setenv("WEAVIATE_API_KEY", opts.WeaviateAPIKey)
	}
	if opts.WeaviateURL != "" {
		os.Setenv("WEAVIATE_URL", opts.WeaviateURL)
	}
	if opts.Timeout != "" {
		// Parse duration string to seconds for environment variable
		timeoutSeconds := parseTimeoutToSeconds(opts.Timeout)
		os.Setenv("WEAVIATE_TIMEOUT", fmt.Sprintf("%d", timeoutSeconds))
	}

	// Set up viper
	if opts.ConfigFile != "" {
		viper.SetConfigFile(opts.ConfigFile)
	} else {
		// Find config paths with proper precedence
		configPaths, err := FindConfigPaths()
		if err == nil && fileExists(configPaths.ConfigPath) {
			viper.SetConfigFile(configPaths.ConfigPath)
		} else {
			// Fallback to default behavior (search in current directory)
			viper.AddConfigPath(".")
			viper.SetConfigType("yaml")
			viper.SetConfigName("config")
		}
	}

	viper.AutomaticEnv()

	// Read config file
	var rawConfig map[string]interface{}
	if err := viper.ReadInConfig(); err != nil {
		// If config file doesn't exist, create a default config from environment variables
		if strings.Contains(err.Error(), "Config File") && strings.Contains(err.Error(), "Not Found") {
			rawConfig = createDefaultConfigFromEnv()
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Get raw config data from file
		if err := viper.Unmarshal(&rawConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Interpolate environment variables
	interpolatedConfig, err := interpolateEnvVars(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to interpolate environment variables: %w", err)
	}

	// Convert to YAML and back to struct
	yamlData, err := yaml.Marshal(interpolatedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interpolated config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal interpolated config: %w", err)
	}

	// Load schemas from directory if schemas_dir is specified
	if config.SchemasDir != "" {
		if err := config.loadSchemasFromDirectory(); err != nil {
			return nil, fmt.Errorf("failed to load schemas from directory: %w", err)
		}
	}

	// Validate configuration and print warnings (non-blocking)
	validationResult := ValidateConfig(&config)
	if os.Getenv("WEAVE_SKIP_CONFIG_VALIDATION") != "true" {
		PrintValidationResult(validationResult)
	}

	return &config, nil
}

// autoDetectDatabaseType detects the database type from environment
func autoDetectDatabaseType() string {
	// Check if explicitly set
	if dbType := os.Getenv("VECTOR_DB_TYPE"); dbType != "" {
		return dbType
	}

	// Auto-detect based on available credentials
	if os.Getenv("WEAVIATE_URL") != "" {
		// If URL contains localhost, it's likely local
		if strings.Contains(strings.ToLower(os.Getenv("WEAVIATE_URL")), "localhost") {
			return string(VectorDBTypeLocal)
		}
		// Otherwise it's cloud
		return string(VectorDBTypeCloud)
	}

	// Check for Supabase credentials (multiple possible variable names)
	supabaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_DATABASE_KEY")

	// Also check alternative names
	if supabaseURL == "" && os.Getenv("SUPABASE_PROJECT_URL") != "" && os.Getenv("SUPABASE_DATABASE_PASSWORD") != "" {
		supabaseURL = "constructed" // We can construct it
	}
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_PROJECT_API_KEY")
	}

	if supabaseURL != "" && supabaseKey != "" {
		return string(VectorDBTypeSupabase)
	}

	// Default to mock for testing if no credentials
	return string(VectorDBTypeMock)
}

// createDefaultConfigFromEnv creates a default configuration from environment variables
func createDefaultConfigFromEnv() map[string]interface{} {
	// Get environment variables with auto-detection
	vectorDBType := autoDetectDatabaseType()

	weaviateURL := os.Getenv("WEAVIATE_URL")
	weaviateAPIKey := os.Getenv("WEAVIATE_API_KEY")
	// Check for Supabase environment variables (multiple possible names)
	supabaseDatabaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	if supabaseDatabaseURL == "" {
		// Try to construct from project URL and password
		projectURL := os.Getenv("SUPABASE_PROJECT_URL")
		password := os.Getenv("SUPABASE_DATABASE_PASSWORD")
		if projectURL != "" && password != "" {
			// Convert https://project.supabase.co to postgresql://postgres:password@db.project.supabase.co:5432/postgres
			if strings.HasPrefix(projectURL, "https://") {
				projectID := strings.TrimPrefix(projectURL, "https://")
				projectID = strings.TrimSuffix(projectID, ".supabase.co")
				supabaseDatabaseURL = fmt.Sprintf("postgresql://postgres:%s@db.%s.supabase.co:5432/postgres", password, projectID)
			}
		}
	}

	supabaseDatabaseKey := os.Getenv("SUPABASE_DATABASE_KEY")
	if supabaseDatabaseKey == "" {
		supabaseDatabaseKey = os.Getenv("SUPABASE_PROJECT_API_KEY")
	}
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")

	// Get timeout from environment or use default
	timeout := DefaultTimeout
	if timeoutEnv := os.Getenv("WEAVIATE_TIMEOUT"); timeoutEnv != "" {
		if t, err := fmt.Sscanf(timeoutEnv, "%d", &timeout); err == nil && t == 1 {
			// Successfully parsed timeout
		} else {
			timeout = DefaultTimeout
		}
	}

	// Get collection names from environment or use defaults
	textCollection := os.Getenv("WEAVIATE_COLLECTION")
	if textCollection == "" {
		textCollection = DefaultTextCollection
	}

	imageCollection := os.Getenv("WEAVIATE_COLLECTION_IMAGES")
	if imageCollection == "" {
		imageCollection = DefaultImageCollection
	}

	// Create default collections
	validCollections := []map[string]interface{}{
		{
			"name":        textCollection,
			"type":        "text",
			"description": "Main text documents collection",
		},
		{
			"name":        imageCollection,
			"type":        "image",
			"description": "Image documents collection",
		},
	}

	// Create vector databases based on type
	var vectorDatabases []map[string]interface{}

	if vectorDBType == "mock" {
		// Mock database
		vectorDatabases = append(vectorDatabases, map[string]interface{}{
			"name":                "mock",
			"type":                "mock",
			"enabled":             true,
			"simulate_embeddings": true,
			"embedding_dimension": 384,
			"timeout":             timeout,
			"collections":         validCollections,
		})
	} else if vectorDBType == "weaviate-local" {
		// Local Weaviate
		weaviateLocalConfig := map[string]interface{}{
			"name":        "weaviate-local",
			"type":        "weaviate-local",
			"url":         "http://localhost:8080",
			"timeout":     timeout,
			"collections": validCollections,
		}
		if openaiAPIKey != "" {
			weaviateLocalConfig["openai_api_key"] = openaiAPIKey
		}
		vectorDatabases = append(vectorDatabases, weaviateLocalConfig)
	} else if vectorDBType == "supabase-local" {
		// Local Supabase (PostgreSQL + pgvector)
		supabaseLocalConfig := map[string]interface{}{
			"name":         "supabase-local",
			"type":         "supabase-local",
			"database_url": "postgresql://postgres:postgres@localhost:5432/weave",
			"timeout":      timeout,
			"collections":  validCollections,
		}
		if openaiAPIKey != "" {
			supabaseLocalConfig["openai_api_key"] = openaiAPIKey
		}
		vectorDatabases = append(vectorDatabases, supabaseLocalConfig)
	} else if vectorDBType == "supabase" {
		// Supabase
		supabaseConfig := map[string]interface{}{
			"name":        "supabase",
			"type":        "supabase",
			"timeout":     timeout,
			"collections": validCollections,
		}

		if supabaseDatabaseURL != "" {
			supabaseConfig["database_url"] = supabaseDatabaseURL
		}
		if supabaseDatabaseKey != "" {
			supabaseConfig["database_key"] = supabaseDatabaseKey
		}
		if openaiAPIKey != "" {
			supabaseConfig["openai_api_key"] = openaiAPIKey
		}

		vectorDatabases = append(vectorDatabases, supabaseConfig)
	} else {
		// Cloud Weaviate (default)
		weaviateConfig := map[string]interface{}{
			"name":        "weaviate-cloud",
			"type":        "weaviate-cloud",
			"timeout":     timeout,
			"collections": validCollections,
		}

		if weaviateURL != "" {
			weaviateConfig["url"] = weaviateURL
		}
		if weaviateAPIKey != "" {
			weaviateConfig["api_key"] = weaviateAPIKey
		}
		if openaiAPIKey != "" {
			weaviateConfig["openai_api_key"] = openaiAPIKey
		}

		vectorDatabases = append(vectorDatabases, weaviateConfig)
	}

	// Create the complete config structure
	return map[string]interface{}{
		"databases": map[string]interface{}{
			"default":          vectorDBType,
			"vector_databases": vectorDatabases,
		},
		"schemas_dir": "./schemas",
	}
}

// interpolateEnvVars recursively interpolates environment variables in the config
func interpolateEnvVars(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case string:
		return InterpolateString(v), nil
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			interpolated, err := interpolateEnvVars(value)
			if err != nil {
				return nil, err
			}
			result[key] = interpolated
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			interpolated, err := interpolateEnvVars(value)
			if err != nil {
				return nil, err
			}
			result[i] = interpolated
		}
		return result, nil
	default:
		return v, nil
	}
}

// InterpolateString interpolates environment variables in a string
func InterpolateString(s string) string {
	// Handle ${VAR:-default} syntax
	if strings.Contains(s, "${") && strings.Contains(s, "}") {
		// Simple implementation for ${VAR:-default} pattern
		start := strings.Index(s, "${")
		end := strings.Index(s[start:], "}")
		if end == -1 {
			return s
		}
		end += start

		varExpr := s[start+2 : end]
		varName := varExpr
		defaultValue := ""

		// Check for default value syntax
		if strings.Contains(varExpr, ":-") {
			parts := strings.SplitN(varExpr, ":-", 2)
			varName = parts[0]
			if len(parts) > 1 {
				defaultValue = parts[1]
			}
		}

		envValue := os.Getenv(varName)
		if envValue == "" {
			envValue = defaultValue
		}

		return s[:start] + envValue + InterpolateString(s[end+1:])
	}

	return s
}

// GetConfigFile returns the path to the config file being used
func GetConfigFile() string {
	return viper.ConfigFileUsed()
}

// GetEnvFile returns the path to the env file being used
func GetEnvFile() string {
	// Return the loaded env file if available
	if loadedEnvFile != "" {
		absPath, _ := filepath.Abs(loadedEnvFile)
		return absPath
	}

	// Fallback: check ENV_FILE environment variable
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		return envFile
	}

	// Fallback: use FindConfigPaths to locate the env file
	configPaths, err := FindConfigPaths()
	if err == nil && configPaths.EnvPath != "" && fileExists(configPaths.EnvPath) {
		absPath, _ := filepath.Abs(configPaths.EnvPath)
		return absPath
	}

	return ""
}

// parseTimeoutToSeconds parses a timeout string (duration or integer) to seconds
// Supports formats: "5s", "10s", "30s", "1m", "5" (integer seconds)
// Returns default timeout (10s) if parsing fails
func parseTimeoutToSeconds(timeoutStr string) int {
	if timeoutStr == "" {
		return DefaultTimeout
	}

	// Try parsing as duration string first (e.g., "5s", "10s", "1m")
	if duration, err := time.ParseDuration(timeoutStr); err == nil {
		seconds := int(duration.Seconds())
		if seconds <= 0 {
			return DefaultTimeout
		}
		return seconds
	}

	// Try parsing as integer seconds (for backward compatibility)
	var timeout int
	if _, err := fmt.Sscanf(timeoutStr, "%d", &timeout); err == nil && timeout > 0 {
		return timeout
	}

	// If parsing fails, return default
	return DefaultTimeout
}

// GetDefaultDatabase returns the default vector database configuration
func (c *Config) GetDefaultDatabase() (*VectorDBConfig, error) {
	if len(c.Databases.VectorDatabases) == 0 {
		// Check for missing environment variables and return a detailed error
		if configErr := CheckRequiredEnvVars(); configErr != nil {
			return nil, configErr
		}
		return nil, fmt.Errorf("no vector databases configured")
	}

	// Get the default database name
	defaultName := c.Databases.Default
	if defaultName == "" {
		// If no default specified, use the first database
		return &c.Databases.VectorDatabases[0], nil
	}

	// Find the database with the default name
	for i := range c.Databases.VectorDatabases {
		if c.Databases.VectorDatabases[i].Name == defaultName {
			return &c.Databases.VectorDatabases[i], nil
		}
	}

	// If default not found, return the first one
	return &c.Databases.VectorDatabases[0], nil
}

// GetDatabase returns a specific vector database configuration by name
// Supports shortcut names that resolve to -cloud variants:
//
//	weaviate -> weaviate-cloud
//	milvus -> milvus-cloud
//	chroma -> chroma-cloud
//	neo4j -> neo4j-cloud
//	qdrant -> qdrant-cloud
func (c *Config) GetDatabase(name string) (*VectorDBConfig, error) {
	if len(c.Databases.VectorDatabases) == 0 {
		// Check for missing environment variables and return a detailed error
		if configErr := CheckRequiredEnvVars(); configErr != nil {
			return nil, configErr
		}
		return nil, fmt.Errorf("no vector databases configured")
	}

	// First, try exact match
	for i := range c.Databases.VectorDatabases {
		if c.Databases.VectorDatabases[i].Name == name {
			return &c.Databases.VectorDatabases[i], nil
		}
	}

	// If no exact match, try shortcut resolution (name -> name-cloud)
	cloudName := name + "-cloud"
	for i := range c.Databases.VectorDatabases {
		if c.Databases.VectorDatabases[i].Name == cloudName {
			return &c.Databases.VectorDatabases[i], nil
		}
	}

	return nil, fmt.Errorf("database '%s' not found (also tried '%s')", name, cloudName)
}

// ListDatabases returns a list of all configured database names
func (c *Config) ListDatabases() []string {
	if len(c.Databases.VectorDatabases) == 0 {
		return []string{}
	}

	names := make([]string, len(c.Databases.VectorDatabases))
	for i, db := range c.Databases.VectorDatabases {
		names[i] = db.Name
	}

	return names
}

// GetDatabaseNames returns a map of database names to their types
func (c *Config) GetDatabaseNames() map[string]VectorDBType {
	if len(c.Databases.VectorDatabases) == 0 {
		return map[string]VectorDBType{}
	}

	names := make(map[string]VectorDBType)
	for _, db := range c.Databases.VectorDatabases {
		names[db.Name] = db.Type
	}

	return names
}

// GetSchema returns a specific schema definition by name
func (c *Config) GetSchema(name string) (*SchemaDefinition, error) {
	if len(c.Databases.Schemas) == 0 {
		return nil, fmt.Errorf("no schemas configured")
	}

	for i := range c.Databases.Schemas {
		if c.Databases.Schemas[i].Name == name {
			return &c.Databases.Schemas[i], nil
		}
	}

	return nil, fmt.Errorf("schema '%s' not found in config.yaml", name)
}

// ListSchemas returns a list of all configured schema names
func (c *Config) ListSchemas() []string {
	if len(c.Databases.Schemas) == 0 {
		return []string{}
	}

	names := make([]string, len(c.Databases.Schemas))
	for i, schema := range c.Databases.Schemas {
		names[i] = schema.Name
	}

	return names
}

// GetAllSchemas returns all configured schema definitions
func (c *Config) GetAllSchemas() []SchemaDefinition {
	return c.Databases.Schemas
}

// loadSchemasFromDirectory loads schema files from the schemas directory
// Schemas defined in config.yaml take precedence over directory schemas with same name
func (c *Config) loadSchemasFromDirectory() error {
	// Check if directory exists
	if _, err := os.Stat(c.SchemasDir); os.IsNotExist(err) {
		// Directory doesn't exist, skip loading
		return nil
	}

	// Read all YAML files in the directory
	files, err := filepath.Glob(filepath.Join(c.SchemasDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to glob schema files: %w", err)
	}

	// Also check for .yml extension
	ymlFiles, err := filepath.Glob(filepath.Join(c.SchemasDir, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to glob schema files: %w", err)
	}
	files = append(files, ymlFiles...)

	// Build a map of existing schema names for precedence checking
	existingSchemas := make(map[string]bool)
	for _, schema := range c.Databases.Schemas {
		existingSchemas[schema.Name] = true
	}

	// Load each schema file
	for _, file := range files {
		schemaData, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read schema file %s: %w", file, err)
		}

		var schemaDef SchemaDefinition
		if err := yaml.Unmarshal(schemaData, &schemaDef); err != nil {
			return fmt.Errorf("failed to parse schema file %s: %w", file, err)
		}

		// Only add if not already defined in config.yaml (precedence)
		if !existingSchemas[schemaDef.Name] {
			c.Databases.Schemas = append(c.Databases.Schemas, schemaDef)
		}
	}

	return nil
}
