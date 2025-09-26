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

// DatabasesConfig holds multiple databases configuration
type DatabasesConfig struct {
	Default         string           `yaml:"default"`
	VectorDatabases []VectorDBConfig `yaml:"vector_databases"`
}

// Config holds the complete application configuration
type Config struct {
	Databases DatabasesConfig `yaml:"databases"`
}

// LoadConfig loads configuration from files and environment variables
func LoadConfig(configFile, envFile string) (*Config, error) {
	// Load environment variables first
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %s: %w", envFile, err)
		}
	} else {
		// Try to load .env from current directory
		_ = godotenv.Load() // .env file is optional, so we continue without it
	}

	// Set up viper
	if configFile != "" {
		viper.SetConfigFile(configFile)
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
