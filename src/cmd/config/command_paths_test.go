// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgconfig "github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/viper"
)

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

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
