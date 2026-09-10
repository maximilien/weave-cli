// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vdb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/viper"
)

func TestVDBCommandContracts(t *testing.T) {
	if VDBCmd.Use != "vdb" || len(VDBCmd.Commands()) != 3 {
		t.Fatalf("VDBCmd = %q with %d subcommands", VDBCmd.Use, len(VDBCmd.Commands()))
	}
	if err := InfoCmd.ValidateArgs([]string{"database"}); err != nil {
		t.Fatalf("info command rejected database name: %v", err)
	}
	if err := InfoCmd.ValidateArgs(nil); err == nil {
		t.Fatal("info command accepted missing database name")
	}

	HealthCmd.Run(HealthCmd, nil)
	HealthCmd.Run(HealthCmd, []string{"cloud"})
}

func TestVDBHelpers(t *testing.T) {
	databases := []config.VectorDBConfig{
		{Name: "cloud", Type: config.VectorDBTypeCloud, URL: "https://cloud.example"},
		{Name: "local", Type: config.VectorDBTypeQdrantLocal, Address: "localhost:6334"},
		{Name: "supabase", Type: config.VectorDBTypeSupabaseCloud, DatabaseURL: "postgres://example"},
		{Name: "empty", Type: config.VectorDBTypeMock},
	}
	if got := filterCloudDatabases(databases); len(got) != 2 {
		t.Fatalf("filterCloudDatabases() returned %d databases", len(got))
	}
	if got := filterLocalDatabases(databases); len(got) != 2 {
		t.Fatalf("filterLocalDatabases() returned %d databases", len(got))
	}

	for _, dbType := range []string{"weaviate-cloud", "milvus-cloud", "mongodb-cloud", "supabase-cloud", "chroma-cloud", "qdrant-cloud", "neo4j-cloud", "opensearch-cloud", "pinecone"} {
		if !isCloudType(dbType) {
			t.Fatalf("isCloudType(%q) = false", dbType)
		}
	}
	if isCloudType("qdrant-local") {
		t.Fatal("isCloudType(qdrant-local) = true")
	}

	if got := getEndpoint(databases[0]); got != "https://cloud.example" {
		t.Fatalf("getEndpoint(URL) = %q", got)
	}
	if got := getEndpoint(databases[1]); got != "localhost:6334" {
		t.Fatalf("getEndpoint(address) = %q", got)
	}
	if got := getEndpoint(databases[2]); got != "postgres://example" {
		t.Fatalf("getEndpoint(database URL) = %q", got)
	}
	if got := getEndpoint(databases[3]); got != "N/A" {
		t.Fatalf("getEndpoint(empty) = %q", got)
	}
	if got := truncate("short", 10); got != "short" {
		t.Fatalf("truncate(short) = %q", got)
	}
	if got := truncate("a long endpoint", 8); got != "a lon..." {
		t.Fatalf("truncate(long) = %q", got)
	}
	if got := maskSecret("short"); got != "***" {
		t.Fatalf("maskSecret(short) = %q", got)
	}
	if got := maskSecret("1234567890"); got != "1234...7890" {
		t.Fatalf("maskSecret(long) = %q", got)
	}

	printSection("Section")
	printField("Field", "value")
}

func TestVDBListAndInfo(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	t.Setenv("WEAVE_SKIP_CONFIG_VALIDATION", "true")
	configPath := filepath.Join(root, "config.yaml")
	contents := `databases:
  default: cloud
  vector_databases:
    - name: cloud
      type: weaviate-cloud
      url: https://cloud.example
      api_key: 1234567890
      username: tester
      timeout: 12
      vector_dimensions: 1536
      embedding_dimension: 1536
      similarity_metric: cosine
      database: vectors
      tenant: tenant-a
      collections:
        - name: Documents
          type: text
          description: Test documents
        - name: Images
          type: image
    - name: local
      type: qdrant-local
      address: localhost:6334
`
	if err := os.WriteFile(configPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	viper.Reset()
	viper.Set("config", configPath)
	t.Cleanup(viper.Reset)

	cloudOnly, localOnly = false, false
	runVDBList(ListCmd, nil)
	cloudOnly = true
	runVDBList(ListCmd, nil)
	cloudOnly, localOnly = false, true
	runVDBList(ListCmd, nil)
	runVDBInfo(InfoCmd, []string{"cloud"})

	if !strings.Contains(getEndpoint(config.VectorDBConfig{URL: "expected"}), "expected") {
		t.Fatal("endpoint helper did not preserve URL")
	}
}
