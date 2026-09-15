// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetConfigLoader(t *testing.T) {
	t.Helper()
	viper.Reset()
	loadedEnvFile = ""
	t.Cleanup(func() {
		viper.Reset()
		loadedEnvFile = ""
	})
	t.Setenv("WEAVE_SKIP_CONFIG_VALIDATION", "true")
}

func clearDatabaseEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"VECTOR_DB_TYPE", "WEAVIATE_URL", "WEAVIATE_API_KEY", "WEAVIATE_TIMEOUT",
		"WEAVIATE_COLLECTION", "WEAVIATE_COLLECTION_IMAGES", "OPENAI_API_KEY",
		"SUPABASE_DATABASE_URL", "SUPABASE_DATABASE_KEY", "SUPABASE_PROJECT_URL",
		"SUPABASE_DATABASE_PASSWORD", "SUPABASE_PROJECT_API_KEY", "ENV_FILE",
	} {
		t.Setenv(name, "")
	}
}

func TestLoadConfigWithOptionsFromYAML(t *testing.T) {
	resetConfigLoader(t)
	clearDatabaseEnvironment(t)
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("CUSTOM_API_KEY", "from-environment")

	schemasDir := filepath.Join(root, "schemas")
	if err := os.MkdirAll(schemasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(schemasDir, "external.yaml"), []byte("name: external\nschema:\n  class: External\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "config.yaml")
	configYAML := `databases:
  default: primary
  vector_databases:
    - name: primary
      type: weaviate-cloud
      url: ${WEAVIATE_URL:-https://default.example}
      api_key: ${CUSTOM_API_KEY}
      timeout: 15
      collections:
        - name: Docs
          type: text
  schemas:
    - name: inline
      schema:
        class: Inline
schemas_dir: ` + schemasDir + `
logging:
  level: debug
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigWithOptions(LoadConfigOptions{
		ConfigFile:     configPath,
		VectorDBType:   "weaviate-cloud",
		WeaviateURL:    "https://flag.example",
		WeaviateAPIKey: "flag-key",
		Timeout:        "2m",
	})
	if err != nil {
		t.Fatalf("LoadConfigWithOptions() error: %v", err)
	}
	database, err := cfg.GetDefaultDatabase()
	if err != nil {
		t.Fatal(err)
	}
	if database.URL != "https://flag.example" || database.APIKey != "from-environment" {
		t.Fatalf("database interpolation = %#v", database)
	}
	if os.Getenv("WEAVIATE_TIMEOUT") != "120" {
		t.Fatalf("WEAVIATE_TIMEOUT = %q", os.Getenv("WEAVIATE_TIMEOUT"))
	}
	if got := cfg.ListSchemas(); len(got) != 2 {
		t.Fatalf("ListSchemas() = %#v", got)
	}
	if GetConfigFile() != configPath {
		t.Fatalf("GetConfigFile() = %q", GetConfigFile())
	}
	if names := cfg.GetDatabaseNames(); names["primary"] != VectorDBTypeCloud {
		t.Fatalf("GetDatabaseNames() = %#v", names)
	}
}

func TestLoadConfigEnvironmentFallbackAndErrors(t *testing.T) {
	t.Run("default mock", func(t *testing.T) {
		resetConfigLoader(t)
		clearDatabaseEnvironment(t)
		root := t.TempDir()
		t.Setenv("HOME", root)
		t.Chdir(root)

		cfg, err := LoadConfig("", "")
		if err != nil {
			t.Fatalf("LoadConfig() error: %v", err)
		}
		database, err := cfg.GetDefaultDatabase()
		if err != nil {
			t.Fatal(err)
		}
		if database.Type != VectorDBTypeMock || database.Timeout != DefaultTimeout {
			t.Fatalf("default database = %#v", database)
		}
	})

	t.Run("missing env file", func(t *testing.T) {
		resetConfigLoader(t)
		_, err := LoadConfigWithOptions(LoadConfigOptions{EnvFile: filepath.Join(t.TempDir(), "missing.env")})
		if err == nil || !strings.Contains(err.Error(), "failed to load env file") {
			t.Fatalf("LoadConfigWithOptions() error = %v", err)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		resetConfigLoader(t)
		path := filepath.Join(t.TempDir(), "invalid.yaml")
		if err := os.WriteFile(path, []byte("databases: [\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfig(path, "")
		if err == nil || !strings.Contains(err.Error(), "failed to read config file") {
			t.Fatalf("LoadConfig() error = %v", err)
		}
	})
}

func TestAutoDetectDatabaseType(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "explicit", env: map[string]string{"VECTOR_DB_TYPE": "redis-local"}, want: "redis-local"},
		{name: "local weaviate", env: map[string]string{"WEAVIATE_URL": "http://localhost:8080"}, want: string(VectorDBTypeLocal)},
		{name: "cloud weaviate", env: map[string]string{"WEAVIATE_URL": "https://cluster.example"}, want: string(VectorDBTypeCloud)},
		{name: "supabase", env: map[string]string{"SUPABASE_DATABASE_URL": "postgres://db", "SUPABASE_DATABASE_KEY": "key"}, want: string(VectorDBTypeSupabase)},
		{name: "supabase aliases", env: map[string]string{"SUPABASE_PROJECT_URL": "https://project.supabase.co", "SUPABASE_DATABASE_PASSWORD": "password", "SUPABASE_PROJECT_API_KEY": "key"}, want: string(VectorDBTypeSupabase)},
		{name: "mock", env: map[string]string{}, want: string(VectorDBTypeMock)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearDatabaseEnvironment(t)
			for key, value := range test.env {
				t.Setenv(key, value)
			}
			if got := autoDetectDatabaseType(); got != test.want {
				t.Fatalf("autoDetectDatabaseType() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCreateDefaultConfigFromEnvironment(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		setup  map[string]string
	}{
		{name: "mock", dbType: "mock"},
		{name: "weaviate local", dbType: "weaviate-local", setup: map[string]string{"OPENAI_API_KEY": "openai"}},
		{name: "supabase local", dbType: "supabase-local", setup: map[string]string{"OPENAI_API_KEY": "openai"}},
		{name: "supabase direct", dbType: "supabase", setup: map[string]string{"SUPABASE_DATABASE_URL": "postgres://db", "SUPABASE_DATABASE_KEY": "key", "OPENAI_API_KEY": "openai"}},
		{name: "supabase constructed", dbType: "supabase", setup: map[string]string{"SUPABASE_PROJECT_URL": "https://project.supabase.co", "SUPABASE_DATABASE_PASSWORD": "pass", "SUPABASE_PROJECT_API_KEY": "key"}},
		{name: "weaviate cloud", dbType: "weaviate-cloud", setup: map[string]string{"WEAVIATE_URL": "https://cluster.example", "WEAVIATE_API_KEY": "key", "OPENAI_API_KEY": "openai"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearDatabaseEnvironment(t)
			t.Setenv("VECTOR_DB_TYPE", test.dbType)
			t.Setenv("WEAVIATE_TIMEOUT", "invalid")
			t.Setenv("WEAVIATE_COLLECTION", "TextDocs")
			t.Setenv("WEAVIATE_COLLECTION_IMAGES", "Pictures")
			for key, value := range test.setup {
				t.Setenv(key, value)
			}

			raw := createDefaultConfigFromEnv()
			databases := raw["databases"].(map[string]interface{})
			if databases["default"] != test.dbType {
				t.Fatalf("default = %#v", databases["default"])
			}
			configured := databases["vector_databases"].([]map[string]interface{})
			if len(configured) != 1 || configured[0]["timeout"] != DefaultTimeout {
				t.Fatalf("configured databases = %#v", configured)
			}
			collections := configured[0]["collections"].([]map[string]interface{})
			if collections[0]["name"] != "TextDocs" || collections[1]["name"] != "Pictures" {
				t.Fatalf("collections = %#v", collections)
			}
		})
	}
}

func TestInterpolateEnvVarsNestedValues(t *testing.T) {
	t.Setenv("VALUE", "configured")
	input := map[string]interface{}{
		"string": "prefix-${VALUE}-suffix",
		"list":   []interface{}{"${MISSING:-fallback}", 42, true},
		"map":    map[string]interface{}{"unchanged": "plain"},
	}
	result, err := interpolateEnvVars(input)
	if err != nil {
		t.Fatal(err)
	}
	values := result.(map[string]interface{})
	if values["string"] != "prefix-configured-suffix" {
		t.Fatalf("interpolated string = %#v", values["string"])
	}
	if values["list"].([]interface{})[0] != "fallback" {
		t.Fatalf("interpolated list = %#v", values["list"])
	}
}
