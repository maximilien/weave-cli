// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgconfig "github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const commandPathConfig = `databases:
  default: local
  vector_databases:
    - name: local
      type: mock
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: Docs
          type: text
    - name: cloud
      type: qdrant-cloud
      url: https://database.example
      api_key: secret
  schemas:
    - name: direct
      schema:
        class: DirectDocs
        vectorizer: none
        properties:
          - name: title
            datatype: [text]
            description: document title
            json_schema:
              type: string
      metadata:
        source:
          type: string
          json_schema:
            type: string
        version: 1
    - name: nested
      schema:
        schema:
          class: NestedDocs
          vectorizer: text2vec-openai
`

func setupCommandPathConfig(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Chdir(root)
	path := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(path, []byte(commandPathConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Reset()
	viper.Set("config", path)
	t.Cleanup(viper.Reset)
	return root
}

func richConfig() (*pkgconfig.Config, map[string]pkgconfig.VectorDBType) {
	database := pkgconfig.VectorDBConfig{
		Name:               "cloud",
		Type:               pkgconfig.VectorDBTypeMilvusCloud,
		URL:                "https://database.example",
		APIKey:             "api-key-long",
		DatabaseURL:        "postgres://database.example/db",
		DatabaseKey:        "database-key-long",
		Database:           "default",
		Tenant:             "tenant",
		VectorDimensions:   768,
		SimilarityMetric:   "cosine",
		Address:            "database.example:19530",
		Username:           "user",
		Password:           "password-long",
		OpenAIAPIKey:       "openai-key-long",
		Timeout:            30,
		Enabled:            true,
		SimulateEmbeddings: true,
		Collections: []pkgconfig.Collection{
			{Name: "Docs", Type: "text"},
		},
	}
	local := pkgconfig.VectorDBConfig{Name: "local", Type: pkgconfig.VectorDBTypeChromaLocal}
	cfg := &pkgconfig.Config{Databases: pkgconfig.DatabasesConfig{
		Default:         "cloud",
		VectorDatabases: []pkgconfig.VectorDBConfig{database, local},
	}}
	return cfg, map[string]pkgconfig.VectorDBType{
		"default": pkgconfig.VectorDBTypeMilvusCloud,
		"cloud":   pkgconfig.VectorDBTypeMilvusCloud,
		"local":   pkgconfig.VectorDBTypeChromaLocal,
		"missing": pkgconfig.VectorDBTypeMock,
	}
}

func TestDatabaseListRendering(t *testing.T) {
	cfg, names := richConfig()
	for _, sortBy := range []string{"name", "type", "deployment"} {
		displayDatabasesTable(names, sortBy)
		displayDatabasesDetails(cfg, names, sortBy)
	}

	for _, test := range []struct {
		dbType pkgconfig.VectorDBType
		cloud  bool
	}{
		{pkgconfig.VectorDBTypeCloud, true},
		{pkgconfig.VectorDBTypeNeo4jCloud, true},
		{pkgconfig.VectorDBTypeMongoDB, true},
		{pkgconfig.VectorDBTypeSupabase, true},
		{pkgconfig.VectorDBType("pinecone"), true},
		{pkgconfig.VectorDBTypeMilvusLocal, false},
	} {
		if got := isCloudDeployment(test.dbType); got != test.cloud {
			t.Errorf("isCloudDeployment(%q) = %t", test.dbType, got)
		}
	}
	if maskSecret("short") != "****" || maskSecret("long-secret") != "long...cret" {
		t.Fatal("maskSecret() returned an unexpected value")
	}
}

func TestShellSetupPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/zsh")
	if got := detectShell(); got != "zsh" {
		t.Fatalf("detectShell() = %q", got)
	}

	for _, shell := range []string{"bash", "zsh", "fish", "powershell", "pwsh"} {
		setup := getShellSetup(shell)
		if setup == nil || setup.ConfigFile == "" || setup.CompletionLine == "" {
			t.Fatalf("getShellSetup(%q) = %#v", shell, setup)
		}
	}
	if getShellSetup("unsupported") != nil {
		t.Fatal("getShellSetup() accepted an unsupported shell")
	}

	configFile := filepath.Join(home, ".zshrc")
	setup := &ShellSetup{
		ConfigFile:     configFile,
		CompletionLine: "source <(weave completion zsh)",
	}
	updateCompletion = false
	requireNoError(t, performSetup(setup))
	requireNoError(t, performSetup(setup))
	updateCompletion = true
	requireNoError(t, performSetup(setup))
	data, err := os.ReadFile(configFile)
	if err != nil || strings.Count(string(data), "weave completion") != 1 {
		t.Fatalf("completion config = %q, %v", data, err)
	}

	withConfigStdin(t, "yes\n", func() {
		if !askPermission() {
			t.Fatal("askPermission() rejected yes")
		}
	})
	withConfigStdin(t, "no\n", func() {
		if askPermission() {
			t.Fatal("askPermission() accepted no")
		}
	})

	dryRun = true
	updateCompletion = false
	runSetupCompletion(setupCompletionCmd, nil)
	dryRun = false
	updateCompletion = false
}

func TestConfigurationTemplateFilePaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", filepath.Join(root, "home"))
	t.Chdir(root)
	if err := os.WriteFile("config.yaml.example", []byte("databases: {}\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("target", 0o755); err != nil {
		t.Fatal(err)
	}

	withConfigStdin(t, "y\n", func() {
		requireNoError(t, createConfigYAMLFileInDir("target"))
	})
	created, err := os.ReadFile(filepath.Join("target", "config.yaml"))
	if err != nil || string(created) != "databases: {}\n" {
		t.Fatalf("created config = %q, %v", created, err)
	}

	withConfigStdin(t, "n\n", func() {
		requireNoError(t, createConfigYAMLFileInDir("target"))
	})

	if err := os.Remove(filepath.Join("target", "config.yaml")); err != nil {
		t.Fatal(err)
	}
	withConfigStdin(t, "y\n", func() {
		requireNoError(t, updateConfigYAMLFileInDir("target"))
	})
	withConfigStdin(t, "y\n", func() {
		requireNoError(t, updateConfigYAMLFileInDir("target"))
	})

	if err := os.MkdirAll("configs", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("configs/weave-agents.yaml", []byte("llm: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	requireNoError(t, createAgentsFileInDir("target"))
	withConfigStdin(t, "n\n", func() {
		requireNoError(t, createAgentsFileInDir("target"))
	})
}

func TestDetermineVDBFilter(t *testing.T) {
	filters := []string{
		"weaviate-cloud", "weaviate-local", "weaviate", "supabase-cloud", "supabase-local", "supabase",
		"mongodb-cloud", "mongodb-local", "mongodb", "milvus-cloud", "milvus-local", "chroma-cloud",
		"chroma-local", "qdrant-cloud", "qdrant-local", "neo4j-cloud", "neo4j-local", "opensearch-cloud",
		"opensearch-local", "elasticsearch-cloud", "elasticsearch-local", "mock",
	}
	for _, filter := range filters {
		t.Run(filter, func(t *testing.T) {
			viper.Reset()
			viper.Set(filter, true)
			if got := determineVDBFilter(); got != filter {
				t.Fatalf("determineVDBFilter() = %q", got)
			}
		})
	}
	viper.Reset()
	if got := determineVDBFilter(); got != "" {
		t.Fatalf("determineVDBFilter() = %q", got)
	}
}

func TestConfigSyncCopiesLocalFiles(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Chdir(root)
	if err := os.WriteFile(".env", []byte("KEY=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("config.yaml", []byte("databases: {}\n"), 0o640); err != nil {
		t.Fatal(err)
	}

	syncEnv = false
	syncConfigYAML = false
	runConfigSync(syncCmd, nil)
	globalDir, err := pkgconfig.GetGlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".env", "config.yaml"} {
		if _, err := os.Stat(filepath.Join(globalDir, name)); err != nil {
			t.Errorf("synced %s: %v", name, err)
		}
	}
	syncEnv = false
	syncConfigYAML = false
}

func TestConfigReadCommands(t *testing.T) {
	setupCommandPathConfig(t)

	for _, test := range []struct {
		name    string
		details bool
		cloud   bool
		local   bool
		sortBy  string
	}{
		{name: "table", sortBy: "name"},
		{name: "details", details: true, sortBy: "type"},
		{name: "cloud", cloud: true, sortBy: "deployment"},
		{name: "local", local: true, sortBy: "name"},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("details", test.details, "")
			cmd.Flags().Bool("cloud", test.cloud, "")
			cmd.Flags().Bool("local", test.local, "")
			cmd.Flags().String("sort-by", test.sortBy, "")
			runList(cmd, nil)
		})
	}

	runListSchemas(&cobra.Command{}, nil)

	for _, format := range []string{"text", "json", "yaml"} {
		t.Run("show-"+format, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("output", format, "")
			runShow(cmd, nil)
			if format == "text" {
				runShow(cmd, []string{"local"})
			}
		})
	}
}

func TestConfigSchemaCommandFormats(t *testing.T) {
	setupCommandPathConfig(t)
	for _, test := range []struct {
		name string
		yaml bool
		json bool
	}{
		{name: "formatted"},
		{name: "yaml", yaml: true},
		{name: "json", json: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("yaml", test.yaml, "")
			cmd.Flags().Bool("json", test.json, "")
			runShowSchema(cmd, []string{"direct"})
		})
	}
	cmd := &cobra.Command{}
	cmd.Flags().Bool("yaml", false, "")
	cmd.Flags().Bool("json", false, "")
	runShowSchema(cmd, []string{"nested"})
}

func TestConfigCreateAndUpdateCommandPaths(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Chdir(root)
	if err := os.WriteFile("config.yaml.example", []byte(commandPathConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	createEnv, createConfigYAML, createGlobal = false, true, false
	withConfigStdin(t, "y\n", func() { runConfigCreate(&cobra.Command{}, nil) })
	if _, err := os.Stat("config.yaml"); err != nil {
		t.Fatalf("created config: %v", err)
	}

	updateEnv, updateConfigYAML, weaveMCP, updateGlobal = false, true, false, false
	withConfigStdin(t, "n\n", func() { runConfigUpdate(&cobra.Command{}, nil) })

	if err := os.Remove("config.yaml"); err != nil {
		t.Fatal(err)
	}
	updateGlobal = true
	withConfigStdin(t, "y\n", func() { runConfigUpdate(&cobra.Command{}, nil) })
	if _, err := os.Stat(filepath.Join(home, ".weave-cli", "config.yaml")); err != nil {
		t.Fatalf("global config: %v", err)
	}

	t.Cleanup(func() {
		createEnv, createConfigYAML, createGlobal = false, false, false
		updateEnv, updateConfigYAML, weaveMCP, updateGlobal = false, false, false, false
	})
}

func TestConfigAgentsTemplateCommand(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", filepath.Join(root, "home"))
	if err := os.Mkdir("configs", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("configs/weave-agents.yaml", []byte("llm:\n  provider: openai\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	agentsGlobal, agentsShowTemplate = false, true
	runConfigAgents(&cobra.Command{}, nil)
	agentsShowTemplate = false
	runConfigAgents(&cobra.Command{}, nil)
	if _, err := os.Stat("weave-agents.yaml"); err != nil {
		t.Fatalf("agents config: %v", err)
	}
	t.Cleanup(func() { agentsGlobal, agentsShowTemplate = false, false })
}

func TestConfigValidationPaths(t *testing.T) {
	root := setupCommandPathConfig(t)
	cfg, output, err := validateConfig(filepath.Join(root, "config.yaml"))
	if err != nil || cfg == nil || output != "" {
		t.Fatalf("validateConfig(valid) = %#v, %q, %v", cfg, output, err)
	}

	invalid := `databases:
  default: mongo
  vector_databases:
    - {name: mongo, type: mongodb}
    - {name: milvus, type: milvus-cloud}
    - {name: chroma, type: chroma-cloud}
    - {name: qdrant, type: qdrant-cloud}
`
	path := filepath.Join(root, "invalid.yaml")
	if err := os.WriteFile(path, []byte(invalid), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Reset()
	viper.Set("config", path)
	_, output, err = validateConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"database_url", "api_key", "tenant"} {
		if !strings.Contains(output, field) {
			t.Errorf("validation output missing %q: %s", field, output)
		}
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
