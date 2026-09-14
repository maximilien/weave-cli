// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validationFields(items []ValidationWarning) string {
	fields := make([]string, 0, len(items))
	for _, item := range items {
		fields = append(fields, item.Field)
	}
	return strings.Join(fields, "\n")
}

func TestValidateConfigMissingConfiguration(t *testing.T) {
	nilResult := ValidateConfig(nil)
	if got := validationFields(nilResult.Errors); !strings.Contains(got, "config") {
		t.Fatalf("expected nil config error, got %q", got)
	}

	emptyResult := ValidateConfig(&Config{})
	if got := validationFields(emptyResult.Warnings); !strings.Contains(got, "databases.vector_databases") {
		t.Fatalf("expected empty database warning, got %q", got)
	}
}

func TestValidateConfigDatabaseMatrix(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	cfg := &Config{Databases: DatabasesConfig{
		Default: "missing-default",
		VectorDatabases: []VectorDBConfig{
			{},
			{Name: "weaviate-cloud", Type: VectorDBTypeCloud, URL: "missing-scheme", Timeout: -1},
			{Name: "milvus-local", Type: VectorDBTypeMilvusLocal},
			{Name: "milvus-cloud", Type: VectorDBTypeMilvusCloud, Address: "cloud:19530"},
			{Name: "qdrant-local", Type: VectorDBTypeQdrantLocal, URL: "https:///missing-host"},
			{Name: "qdrant-cloud", Type: VectorDBTypeQdrantCloud, URL: "https://qdrant.example"},
			{Name: "chroma-local", Type: VectorDBTypeChromaLocal},
			{Name: "chroma-cloud", Type: VectorDBTypeChromaCloud},
			{Name: "supabase", Type: VectorDBTypeSupabase, DatabaseURL: "https://example.test"},
			{Name: "supabase-local", Type: VectorDBTypeSupabaseLocal},
			{Name: "mongodb", Type: VectorDBTypeMongoDB, DatabaseURL: "https://example.test"},
			{Name: "mongodb-cloud", Type: VectorDBTypeMongoDBCloud},
			{Name: "neo4j", Type: VectorDBTypeNeo4jCloud},
			{Name: "opensearch", Type: VectorDBTypeOpenSearchLocal},
			{Name: "redis-local", Type: VectorDBTypeRedisLocal},
			{Name: "redis-cloud", Type: VectorDBTypeRedisCloud, URL: "redis://example.test:6379"},
			{Name: "mock", Type: VectorDBTypeMock},
		},
	}}

	result := ValidateConfig(cfg)
	errors := validationFields(result.Errors)
	warnings := validationFields(result.Warnings)
	for _, field := range []string{".name", ".type", ".address", ".url", ".database_url", ".api_key"} {
		if !strings.Contains(errors, field) {
			t.Errorf("expected error containing %q, got:\n%s", field, errors)
		}
	}
	for _, field := range []string{"databases.default", ".timeout", ".openai_api_key", ".vector_dimensions", ".tenant", ".database", ".username/password", ".embedding_dimension"} {
		if !strings.Contains(warnings, field) {
			t.Errorf("expected warning containing %q, got:\n%s", field, warnings)
		}
	}
	for _, deprecated := range []string{"Type 'supabase' is deprecated", "Type 'mongodb' is deprecated"} {
		found := false
		for _, warning := range result.Warnings {
			if warning.Message == deprecated {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected warning %q", deprecated)
		}
	}
}

func TestValidateConfigAcceptsCompleteDatabases(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "configured")
	dbs := []VectorDBConfig{
		{Name: "weaviate", Type: VectorDBTypeLocal, URL: "http://localhost:8080"},
		{Name: "milvus", Type: VectorDBTypeMilvusLocal, Address: "localhost:19530", VectorDimensions: 384},
		{Name: "qdrant", Type: VectorDBTypeQdrantCloud, URL: "${QDRANT_URL}", APIKey: "secret"},
		{Name: "chroma", Type: VectorDBTypeChromaCloud, URL: "https://chroma.example", APIKey: "secret", Tenant: "tenant"},
		{Name: "supabase", Type: VectorDBTypeSupabaseCloud, DatabaseURL: "postgresql://user:pass@example.test/db"},
		{Name: "mongodb", Type: VectorDBTypeMongoDBCloud, DatabaseURL: "mongodb+srv://example.test/db", Database: "db"},
		{Name: "neo4j", Type: VectorDBTypeNeo4jLocal, URL: "bolt://localhost:7687", Username: "neo4j", Password: "secret"},
		{Name: "opensearch", Type: VectorDBTypeOpenSearchCloud, URL: "https://search.example"},
		{Name: "redis", Type: VectorDBTypeRedisCloud, Address: "example.test:6379", Password: "secret"},
		{Name: "mock", Type: VectorDBTypeMock, EmbeddingDimension: 384},
	}
	result := ValidateConfig(&Config{Databases: DatabasesConfig{Default: "weaviate", VectorDatabases: dbs}})
	if len(result.Errors) != 0 || len(result.Warnings) != 0 {
		t.Fatalf("expected complete configuration to validate, got errors=%v warnings=%v", result.Errors, result.Warnings)
	}
}

func TestValidateURLConfig(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr string
	}{
		{name: "http", value: "http://localhost:8080"},
		{name: "template", value: "${SERVICE_URL}"},
		{name: "missing scheme", value: "example.test/path", wantErr: "missing scheme"},
		{name: "missing host", value: "https:///path", wantErr: "missing host"},
		{name: "parse error", value: "://", wantErr: "missing protocol scheme"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURLConfig(tt.value)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("expected %q error, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestConfigPathPrecedenceAndCreation(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)
	oldWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDir) })

	globalDir, err := GetGlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if globalDir != filepath.Join(home, GlobalConfigDirName) {
		t.Fatalf("unexpected global directory: %s", globalDir)
	}
	created, err := EnsureGlobalConfigDir()
	if err != nil || !created {
		t.Fatalf("expected global directory creation, created=%v err=%v", created, err)
	}
	created, err = EnsureGlobalConfigDir()
	if err != nil || created {
		t.Fatalf("expected existing directory, created=%v err=%v", created, err)
	}

	paths, err := FindConfigPaths()
	if err != nil || paths.Location != "local" || paths.ConfigPath != ConfigFileName {
		t.Fatalf("unexpected default paths: %#v, %v", paths, err)
	}
	if HasLocalConfig() {
		t.Fatal("did not expect local configuration")
	}
	hasGlobal, err := HasGlobalConfig()
	if err != nil || hasGlobal {
		t.Fatalf("did not expect global configuration, has=%v err=%v", hasGlobal, err)
	}

	globalEnv, err := GetGlobalEnvPath()
	if err != nil {
		t.Fatal(err)
	}
	globalConfig, err := GetGlobalConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(globalEnv, []byte("KEY=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	paths, err = FindConfigPaths()
	if err != nil || paths.Location != "global" || paths.EnvPath != globalEnv || paths.ConfigPath != globalConfig {
		t.Fatalf("unexpected global paths: %#v, %v", paths, err)
	}
	hasGlobal, err = HasGlobalConfig()
	if err != nil || !hasGlobal {
		t.Fatalf("expected global configuration, has=%v err=%v", hasGlobal, err)
	}

	if err := os.WriteFile(ConfigFileName, []byte("databases: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	paths, err = FindConfigPaths()
	if err != nil || paths.Location != "local" || paths.ConfigPath != ConfigFileName {
		t.Fatalf("local configuration should take precedence: %#v, %v", paths, err)
	}
	if !HasLocalConfig() || !fileExists(ConfigFileName) {
		t.Fatal("expected local configuration")
	}
}
