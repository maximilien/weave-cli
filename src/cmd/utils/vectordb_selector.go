// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// VectorDBSelection represents the selected vector databases
type VectorDBSelection struct {
	Configs []config.VectorDBConfig
	Types   []string
}

// OperationType defines the type of operation being performed
type OperationType string

const (
	OperationTypeRead   OperationType = "read"   // ls, show, count - can use multiple DBs
	OperationTypeWrite  OperationType = "write"  // create, update - requires single DB
	OperationTypeDelete OperationType = "delete" // delete, da, ds - requires single DB
)

// ValidateDatabaseSelection validates that the database selection is appropriate for the operation type
func ValidateDatabaseSelection(selection *VectorDBSelection, opType OperationType, operationName string) error {
	if selection == nil || len(selection.Configs) == 0 {
		return fmt.Errorf("no database selected")
	}

	// For write and delete operations, require exactly one database
	// Only enforce this when multiple databases are configured (convenience for single DB setups)
	if opType == OperationTypeWrite || opType == OperationTypeDelete {
		if len(selection.Configs) > 1 {
			return fmt.Errorf(
				"%s operation requires a single database. "+
					"Please specify --weaviate, --supabase, --mongodb, --milvus-local, --milvus-cloud, --chroma-local, --chroma-cloud, or --mock (found %d databases). "+
					"Use 'weave config list' to see configured databases",
				operationName, len(selection.Configs))
		}
		// If only one database configured, allow it (convenience)
	}

	return nil
}

// SelectSingleDatabase intelligently selects a single database from multiple options
// Returns the selected database config, or nil if selection cannot be made
// This function:
// 1. Returns the only config if there's just one
// 2. Uses VECTOR_DB_TYPE from env/config as default
// 3. For Weaviate databases, tries to find the collection in available DBs
// 4. Returns nil if multiple different database types and no default
func SelectSingleDatabase(ctx context.Context, configs []config.VectorDBConfig, cfg *config.Config, collectionName string) *config.VectorDBConfig {
	// If only one config, use it
	if len(configs) == 1 {
		return &configs[0]
	}

	// Try to use the default database type from config/env
	defaultDB := SelectDefaultDatabase(configs, cfg)
	if defaultDB != nil {
		PrintInfo(fmt.Sprintf("Using default database: %s (%s)", defaultDB.Name, defaultDB.Type))
		fmt.Println()
		return defaultDB
	}

	// If all are Weaviate databases and we have a collection name, try each
	if collectionName != "" && AreAllWeaviateDatabases(configs) {
		PrintInfo(fmt.Sprintf("Trying %d Weaviate databases...", len(configs)))
		fmt.Println()
		dbConfig := TryWeaviateDatabasesForCollection(ctx, configs, collectionName)
		if dbConfig != nil {
			return dbConfig
		}
	}

	// Cannot automatically select
	return nil
}

// HandleSingleDatabaseSelection is a helper that combines smart database selection with error handling
// Use this in commands that need exactly one database
// Returns the selected database config or exits the program with an error message
func HandleSingleDatabaseSelection(ctx context.Context, selection *VectorDBSelection, cfg *config.Config, collectionName, commandExample string) *config.VectorDBConfig {
	dbConfig := SelectSingleDatabase(ctx, selection.Configs, cfg, collectionName)
	if dbConfig == nil {
		// Could not automatically select - show helpful error
		PrintError(fmt.Sprintf("This operation requires a single database, but found %d databases:", len(selection.Configs)))
		for _, cfg := range selection.Configs {
			PrintInfo(fmt.Sprintf("  • %s (%s)", cfg.Name, cfg.Type))
		}
		fmt.Println()
		PrintInfo("Please use --vector-db-type (or --vdb) to specify which database:")
		for _, cfg := range selection.Configs {
			PrintInfo(fmt.Sprintf("  %s --vector-db-type %s", commandExample, cfg.Type))
		}
		os.Exit(1)
	}
	return dbConfig
}

// GetSelectedVectorDBs determines which vector databases to use based on command flags
func GetSelectedVectorDBs(cmd *cobra.Command, cfg *config.Config) (*VectorDBSelection, error) {
	// Get flags
	useWeaviate, _ := cmd.Flags().GetBool("weaviate")
	useSupabase, _ := cmd.Flags().GetBool("supabase")
	useMongoDB, _ := cmd.Flags().GetBool("mongodb")
	useMilvusLocal, _ := cmd.Flags().GetBool("milvus-local")
	useMilvusCloud, _ := cmd.Flags().GetBool("milvus-cloud")
	useChromaLocal, _ := cmd.Flags().GetBool("chroma-local")
	useChromaCloud, _ := cmd.Flags().GetBool("chroma-cloud")
	useMock, _ := cmd.Flags().GetBool("mock")
	useAll, _ := cmd.Flags().GetBool("all")

	var configs []config.VectorDBConfig
	var types []string

	// Check if any specific database flags are set
	hasSpecificFlags := useWeaviate || useSupabase || useMongoDB || useMilvusLocal || useMilvusCloud || useChromaLocal || useChromaCloud || useMock

	// If --all flag is used OR no specific flags are set, return all configured databases
	if useAll || !hasSpecificFlags {
		return getAllConfiguredDatabases(cfg)
	}

	// Handle specific database type flags
	if useWeaviate {
		weaviateConfigs, err := getWeaviateConfigs(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get Weaviate configuration: %w", err)
		}
		configs = append(configs, weaviateConfigs...)
		for _, config := range weaviateConfigs {
			types = append(types, string(config.Type))
		}
	}

	if useSupabase {
		supabaseConfig, err := getSupabaseConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get Supabase configuration: %w", err)
		}
		configs = append(configs, *supabaseConfig)
		types = append(types, string(supabaseConfig.Type))
	}

	if useMongoDB {
		mongoDBConfig, err := getMongoDBConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get MongoDB configuration: %w", err)
		}
		configs = append(configs, *mongoDBConfig)
		types = append(types, string(mongoDBConfig.Type))
	}

	if useMilvusLocal {
		milvusLocalConfig, err := getMilvusLocalConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get Milvus Local configuration: %w", err)
		}
		configs = append(configs, *milvusLocalConfig)
		types = append(types, string(milvusLocalConfig.Type))
	}

	if useMilvusCloud {
		milvusCloudConfig, err := getMilvusCloudConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get Milvus Cloud configuration: %w", err)
		}
		configs = append(configs, *milvusCloudConfig)
		types = append(types, string(milvusCloudConfig.Type))
	}

	if useChromaLocal {
		chromaLocalConfig, err := getChromaLocalConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get Chroma Local configuration: %w", err)
		}
		configs = append(configs, *chromaLocalConfig)
		types = append(types, string(chromaLocalConfig.Type))
	}

	if useChromaCloud {
		chromaCloudConfig, err := getChromaCloudConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get Chroma Cloud configuration: %w", err)
		}
		configs = append(configs, *chromaCloudConfig)
		types = append(types, string(chromaCloudConfig.Type))
	}

	if useMock {
		mockConfig, err := getMockConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get mock configuration: %w", err)
		}
		configs = append(configs, *mockConfig)
		types = append(types, string(mockConfig.Type))
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no valid vector database configurations found")
	}

	return &VectorDBSelection{
		Configs: configs,
		Types:   types,
	}, nil
}

// getAllConfiguredDatabases returns all configured vector databases
func getAllConfiguredDatabases(cfg *config.Config) (*VectorDBSelection, error) {
	var configs []config.VectorDBConfig
	var types []string

	// Get all configured databases from config.yaml
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		configs = append(configs, dbConfig)
		types = append(types, string(dbConfig.Type))
	}

	// Also check for Supabase from environment variables if not already in config
	hasSupabase := false
	for _, dbConfig := range configs {
		if dbConfig.Type == config.VectorDBTypeSupabase {
			hasSupabase = true
			break
		}
	}

	if !hasSupabase {
		if supabaseConfig := tryCreateSupabaseConfigFromLoadedConfig(cfg); supabaseConfig != nil {
			configs = append(configs, *supabaseConfig)
			types = append(types, string(supabaseConfig.Type))
		} else if supabaseConfig := tryCreateSupabaseConfigFromEnv(); supabaseConfig != nil {
			configs = append(configs, *supabaseConfig)
			types = append(types, string(supabaseConfig.Type))
		}
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no vector database configurations found")
	}

	return &VectorDBSelection{
		Configs: configs,
		Types:   types,
	}, nil
}

// getWeaviateConfigs returns Weaviate configurations (both cloud and local)
func getWeaviateConfigs(cfg *config.Config) ([]config.VectorDBConfig, error) {
	var configs []config.VectorDBConfig

	// Check all configured databases for Weaviate types
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeCloud || dbConfig.Type == config.VectorDBTypeLocal {
			configs = append(configs, dbConfig)
		}
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no Weaviate configuration found")
	}

	return configs, nil
}

// getSupabaseConfig returns Supabase configuration
func getSupabaseConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Supabase type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeSupabase {
			return &dbConfig, nil
		}
	}

	// If no explicit Supabase config found, check if we can create one from the loaded configuration
	// This allows using --supabase flag even when Supabase isn't the default database
	if supabaseConfig := tryCreateSupabaseConfigFromLoadedConfig(cfg); supabaseConfig != nil {
		return supabaseConfig, nil
	}

	// Fallback to environment variables (for cases where config isn't loaded yet)
	if supabaseConfig := tryCreateSupabaseConfigFromEnv(); supabaseConfig != nil {
		return supabaseConfig, nil
	}

	return nil, fmt.Errorf("no Supabase configuration found")
}

// getMongoDBConfig returns MongoDB configuration
func getMongoDBConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for MongoDB type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeMongoDB {
			return &dbConfig, nil
		}
	}

	return nil, fmt.Errorf("no MongoDB configuration found")
}

// getMilvusLocalConfig returns Milvus Local configuration
func getMilvusLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Milvus Local type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeMilvusLocal {
			return &dbConfig, nil
		}
	}

	return nil, fmt.Errorf("no Milvus Local configuration found")
}

// getMilvusCloudConfig returns Milvus Cloud configuration
func getMilvusCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Milvus Cloud type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeMilvusCloud {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if milvusConfig := tryCreateMilvusCloudConfigFromEnv(); milvusConfig != nil {
		return milvusConfig, nil
	}

	return nil, fmt.Errorf("no Milvus Cloud configuration found")
}

// tryCreateMilvusCloudConfigFromEnv attempts to create a Milvus Cloud config from environment variables
func tryCreateMilvusCloudConfigFromEnv() *config.VectorDBConfig {
	address := os.Getenv("MILVUS_CLOUD_ADDRESS")
	username := os.Getenv("MILVUS_CLOUD_USERNAME")
	password := os.Getenv("MILVUS_CLOUD_PASSWORD")
	token := os.Getenv("MILVUS_CLOUD_TOKEN")

	// Require at least address
	if address == "" {
		return nil
	}

	// Prefer token if available (Zilliz serverless)
	if token != "" {
		return &config.VectorDBConfig{
			Name:             "milvus-cloud",
			Type:             config.VectorDBTypeMilvusCloud,
			Address:          address,
			APIKey:           token,
			Database:         getEnvOrDefault("MILVUS_CLOUD_DATABASE", "default"),
			VectorDimensions: 1536,
			SimilarityMetric: "COSINE",
			Timeout:          60, // Longer timeout for cloud connections
		}
	}

	// Fall back to username/password
	if username != "" && password != "" {
		return &config.VectorDBConfig{
			Name:             "milvus-cloud",
			Type:             config.VectorDBTypeMilvusCloud,
			Address:          address,
			Username:         username,
			Password:         password,
			Database:         getEnvOrDefault("MILVUS_CLOUD_DATABASE", "default"),
			VectorDimensions: 1536,
			SimilarityMetric: "COSINE",
			Timeout:          60, // Longer timeout for cloud connections
		}
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getMockConfig returns mock configuration
func getMockConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Mock type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeMock {
			return &dbConfig, nil
		}
	}

	// If no mock database is configured, create a default one
	return &config.VectorDBConfig{
		Name:    "mock",
		Type:    config.VectorDBTypeMock,
		Enabled: true,
		Timeout: 10, // Default timeout
	}, nil
}

// GetVectorDBTypeFromConfig determines the vector DB type from a VectorDBConfig
func GetVectorDBTypeFromConfig(dbConfig *config.VectorDBConfig) string {
	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		return "weaviate-cloud"
	case config.VectorDBTypeLocal:
		return "weaviate-local"
	case config.VectorDBTypeMock:
		return "mock"
	case config.VectorDBTypeSupabase:
		return "supabase"
	default:
		return string(dbConfig.Type)
	}
}

// tryCreateSupabaseConfigFromLoadedConfig attempts to create a Supabase config from the loaded configuration
// This handles cases where Supabase credentials are in .env files or other config sources
func tryCreateSupabaseConfigFromLoadedConfig(cfg *config.Config) *config.VectorDBConfig {
	// Check if the main config has Supabase credentials loaded
	// This would happen if SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY are in .env files

	// The config system might have loaded these into various fields, let's check them all
	var databaseURL, databaseKey string

	// Check if they're loaded in any of the configured databases
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.DatabaseURL != "" {
			databaseURL = dbConfig.DatabaseURL
		}
		if dbConfig.DatabaseKey != "" {
			databaseKey = dbConfig.DatabaseKey
		}
		// If we found both, we can stop looking
		if databaseURL != "" && databaseKey != "" {
			break
		}
	}

	// Also check environment variables that might have been loaded by viper
	if databaseURL == "" {
		databaseURL = os.Getenv("SUPABASE_DATABASE_URL")
		if databaseURL == "" {
			databaseURL = os.Getenv("DATABASE_URL")
		}
	}
	if databaseKey == "" {
		databaseKey = os.Getenv("SUPABASE_DATABASE_KEY")
		if databaseKey == "" {
			databaseKey = os.Getenv("SUPABASE_ANON_KEY")
		}
		if databaseKey == "" {
			databaseKey = os.Getenv("SUPABASE_KEY")
		}
	}

	// Need both URL and key to create a valid Supabase config
	if databaseURL == "" || databaseKey == "" {
		return nil
	}

	return &config.VectorDBConfig{
		Name:        "supabase-from-config",
		Type:        config.VectorDBTypeSupabase,
		DatabaseURL: databaseURL,
		DatabaseKey: databaseKey,
		Enabled:     true,
		Timeout:     30, // Default timeout
	}
}

// SelectDefaultDatabase selects the default database from a list based on VECTOR_DB_TYPE
func SelectDefaultDatabase(configs []config.VectorDBConfig, cfg *config.Config) *config.VectorDBConfig {
	// Get the default database type from environment or config
	defaultType := os.Getenv("VECTOR_DB_TYPE")
	if defaultType == "" && cfg != nil {
		defaultType = string(cfg.Databases.Default)
	}

	if defaultType == "" {
		return nil
	}

	// Find a config that matches the default type
	for i := range configs {
		if string(configs[i].Type) == defaultType {
			return &configs[i]
		}
	}

	return nil
}

// AreAllWeaviateDatabases checks if all configs are Weaviate databases
func AreAllWeaviateDatabases(configs []config.VectorDBConfig) bool {
	for _, cfg := range configs {
		if cfg.Type != config.VectorDBTypeCloud && cfg.Type != config.VectorDBTypeLocal {
			return false
		}
	}
	return true
}

// TryWeaviateDatabasesForCollection tries to find a collection in Weaviate databases
func TryWeaviateDatabasesForCollection(ctx context.Context, configs []config.VectorDBConfig, collectionName string) *config.VectorDBConfig {
	for i := range configs {
		// Try to check if collection exists in this database
		client, err := CreateWeaviateClient(&configs[i])
		if err != nil {
			continue
		}

		collections, err := client.ListCollections(ctx)
		if err != nil {
			continue
		}

		// Check if collection exists
		for _, col := range collections {
			if col == collectionName {
				PrintSuccess(fmt.Sprintf("Found collection '%s' in %s (%s)", collectionName, configs[i].Name, configs[i].Type))
				fmt.Println()
				return &configs[i]
			}
		}
	}

	return nil
}

// tryCreateSupabaseConfigFromEnv attempts to create a Supabase config from environment variables
func tryCreateSupabaseConfigFromEnv() *config.VectorDBConfig {
	databaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	databaseKey := os.Getenv("SUPABASE_DATABASE_KEY")

	// Also check alternative environment variable names
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
		// Try to construct from project URL and password
		if databaseURL == "" {
			projectURL := os.Getenv("SUPABASE_PROJECT_URL")
			password := os.Getenv("SUPABASE_DATABASE_PASSWORD")
			if projectURL != "" && password != "" {
				// Convert https://project.supabase.co to postgresql://postgres:password@db.project.supabase.co:5432/postgres
				if strings.HasPrefix(projectURL, "https://") {
					projectID := strings.TrimPrefix(projectURL, "https://")
					projectID = strings.TrimSuffix(projectID, ".supabase.co")
					databaseURL = fmt.Sprintf("postgresql://postgres:%s@db.%s.supabase.co:5432/postgres", password, projectID)
				}
			}
		}
	}

	if databaseKey == "" {
		databaseKey = os.Getenv("SUPABASE_PROJECT_API_KEY")
		if databaseKey == "" {
			databaseKey = os.Getenv("SUPABASE_ANON_KEY")
		}
		if databaseKey == "" {
			databaseKey = os.Getenv("SUPABASE_KEY")
		}
	}

	// Need both URL and key to create a valid Supabase config
	if databaseURL == "" || databaseKey == "" {
		return nil
	}

	return &config.VectorDBConfig{
		Name:        "supabase-env",
		Type:        config.VectorDBTypeSupabase,
		DatabaseURL: databaseURL,
		DatabaseKey: databaseKey,
		Enabled:     true,
		Timeout:     30, // Default timeout
	}
}
