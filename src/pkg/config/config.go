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

// WeaviateCloudConfig holds Weaviate Cloud configuration
type WeaviateCloudConfig struct {
	URL                string `yaml:"url"`
	APIKey             string `yaml:"api_key"`
	CollectionName     string `yaml:"collection_name"`
	CollectionNameTest string `yaml:"collection_name_test"`
}

// WeaviateLocalConfig holds Weaviate Local configuration
type WeaviateLocalConfig struct {
	URL                string `yaml:"url"`
	CollectionName     string `yaml:"collection_name"`
	CollectionNameTest string `yaml:"collection_name_test"`
}

// MockCollection represents a mock collection
type MockCollection struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// MockConfig holds mock database configuration
type MockConfig struct {
	Enabled            bool             `yaml:"enabled"`
	SimulateEmbeddings bool             `yaml:"simulate_embeddings"`
	EmbeddingDimension int              `yaml:"embedding_dimension"`
	Collections        []MockCollection `yaml:"collections"`
}

// VectorDBConfig holds vector database configuration
type VectorDBConfig struct {
	Type          VectorDBType        `yaml:"type"`
	WeaviateCloud WeaviateCloudConfig `yaml:"weaviate_cloud"`
	WeaviateLocal WeaviateLocalConfig `yaml:"weaviate_local"`
	Mock          MockConfig          `yaml:"mock"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	VectorDB VectorDBConfig `yaml:"vector_db"`
}

// Config holds the complete application configuration
type Config struct {
	Database DatabaseConfig `yaml:"database"`
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
