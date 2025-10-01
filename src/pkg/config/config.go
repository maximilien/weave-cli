// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"
)

// VectorDBType represents the type of vector database
type VectorDBType string

const (
	VectorDBTypeCloud VectorDBType = "weaviate-cloud"
	VectorDBTypeLocal VectorDBType = "weaviate-local"
	VectorDBTypeMock  VectorDBType = "mock"
)

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
	OpenAIAPIKey       string       `yaml:"openai_api_key,omitempty"`
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
type Config struct {
	Databases  DatabasesConfig `yaml:"databases"`
	SchemasDir string          `yaml:"schemas_dir,omitempty"`
}

// LoadConfig loads configuration from files and environment variables
// LoadConfigOptions holds options for loading configuration
type LoadConfigOptions struct {
	ConfigFile     string
	EnvFile        string
	VectorDBType   string
	WeaviateAPIKey string
	WeaviateURL    string
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
	// 3. .env file
	// 4. Shell environment (lowest priority)

	// Load from --env file if specified
	if opts.EnvFile != "" {
		if err := godotenv.Load(opts.EnvFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %s: %w", opts.EnvFile, err)
		}
	} else {
		// Try to load .env from current directory
		_ = godotenv.Load() // .env file is optional, so we continue without it
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

	// Set up viper
	if opts.ConfigFile != "" {
		viper.SetConfigFile(opts.ConfigFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Get raw config data
	var rawConfig map[string]interface{}
	if err := viper.Unmarshal(&rawConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
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

	return &config, nil
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
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		return envFile
	}

	// Check common locations
	locations := []string{".env", "./.env"}
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			absPath, _ := filepath.Abs(loc)
			return absPath
		}
	}

	return ""
}

// GetDefaultDatabase returns the default vector database configuration
func (c *Config) GetDefaultDatabase() (*VectorDBConfig, error) {
	if len(c.Databases.VectorDatabases) == 0 {
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
func (c *Config) GetDatabase(name string) (*VectorDBConfig, error) {
	if len(c.Databases.VectorDatabases) == 0 {
		return nil, fmt.Errorf("no vector databases configured")
	}

	for i := range c.Databases.VectorDatabases {
		if c.Databases.VectorDatabases[i].Name == name {
			return &c.Databases.VectorDatabases[i], nil
		}
	}

	return nil, fmt.Errorf("database '%s' not found", name)
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
