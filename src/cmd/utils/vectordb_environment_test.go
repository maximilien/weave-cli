// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestDeploymentClassificationAndFiltering(t *testing.T) {
	for _, test := range []struct {
		dbType config.VectorDBType
		cloud  bool
	}{
		{config.VectorDBTypeCloud, true},
		{config.VectorDBTypeMilvusCloud, true},
		{config.VectorDBTypeNeo4jCloud, true},
		{config.VectorDBTypeSupabase, true},
		{config.VectorDBTypeMongoDB, true},
		{config.VectorDBTypeLocal, false},
		{config.VectorDBTypeMock, false},
	} {
		if got := IsCloudDatabase(test.dbType); got != test.cloud {
			t.Errorf("IsCloudDatabase(%q) = %t, want %t", test.dbType, got, test.cloud)
		}
	}
	configs := []config.VectorDBConfig{{Type: config.VectorDBTypeCloud}, {Type: config.VectorDBTypeLocal}, {Type: config.VectorDBTypeMock}}
	if got := FilterByDeployment(configs, false, false); len(got) != 3 {
		t.Fatalf("unfiltered configs = %#v", got)
	}
	if got := FilterByDeployment(configs, true, false); len(got) != 1 || got[0].Type != config.VectorDBTypeCloud {
		t.Fatalf("cloud configs = %#v", got)
	}
	if got := FilterByDeployment(configs, false, true); len(got) != 2 {
		t.Fatalf("local configs = %#v", got)
	}
}

func TestMilvusEnvironmentConfig(t *testing.T) {
	clearVectorDBEnvironment(t)
	if got := tryCreateMilvusCloudConfigFromEnv(); got != nil {
		t.Fatalf("empty Milvus environment = %#v", got)
	}
	t.Setenv("MILVUS_CLOUD_ADDRESS", "cloud.example:19530")
	t.Setenv("MILVUS_CLOUD_TOKEN", "token")
	t.Setenv("MILVUS_CLOUD_DATABASE", "vectors")
	if got := tryCreateMilvusCloudConfigFromEnv(); got == nil || got.APIKey != "token" || got.Database != "vectors" {
		t.Fatalf("Milvus token config = %#v", got)
	}
	t.Setenv("MILVUS_CLOUD_TOKEN", "")
	t.Setenv("MILVUS_CLOUD_USERNAME", "user")
	t.Setenv("MILVUS_CLOUD_PASSWORD", "pass")
	if got := tryCreateMilvusCloudConfigFromEnv(); got == nil || got.Username != "user" || got.Password != "pass" {
		t.Fatalf("Milvus password config = %#v", got)
	}
	t.Setenv("MILVUS_CLOUD_PASSWORD", "")
	if got := tryCreateMilvusCloudConfigFromEnv(); got != nil {
		t.Fatalf("incomplete Milvus environment = %#v", got)
	}
	if got, err := getMilvusCloudConfig(&config.Config{}); got != nil || err == nil {
		t.Fatalf("missing Milvus config = %#v, %v", got, err)
	}
}

func TestSupabaseEnvironmentConfigs(t *testing.T) {
	clearVectorDBEnvironment(t)
	if got := tryCreateSupabaseConfigFromEnv(); got != nil {
		t.Fatalf("empty Supabase environment = %#v", got)
	}
	t.Setenv("SUPABASE_PROJECT_URL", "https://project.supabase.co")
	t.Setenv("SUPABASE_DATABASE_PASSWORD", "secret")
	t.Setenv("SUPABASE_PROJECT_API_KEY", "key")
	got := tryCreateSupabaseConfigFromEnv()
	if got == nil || !strings.Contains(got.DatabaseURL, "db.project.supabase.co") || got.DatabaseKey != "key" {
		t.Fatalf("Supabase project config = %#v", got)
	}

	clearVectorDBEnvironment(t)
	cfg := &config.Config{Databases: config.DatabasesConfig{VectorDatabases: []config.VectorDBConfig{
		{DatabaseURL: "postgresql://loaded"}, {DatabaseKey: "loaded-key"},
	}}}
	got = tryCreateSupabaseConfigFromLoadedConfig(cfg)
	if got == nil || got.DatabaseURL != "postgresql://loaded" || got.DatabaseKey != "loaded-key" {
		t.Fatalf("loaded Supabase config = %#v", got)
	}
	if selected, err := getSupabaseCloudConfig(cfg); err != nil || selected == nil {
		t.Fatalf("Supabase cloud fallback = %#v, %v", selected, err)
	}
	if selected, err := getSupabaseConfig(cfg); err != nil || selected == nil {
		t.Fatalf("Supabase fallback = %#v, %v", selected, err)
	}
}

func TestChromaEnvironmentConfigs(t *testing.T) {
	clearVectorDBEnvironment(t)
	if got := tryCreateChromaLocalConfigFromEnv(); got != nil {
		t.Fatalf("empty Chroma local environment = %#v", got)
	}
	t.Setenv("CHROMA_URL", "http://localhost:8000")
	if got := tryCreateChromaLocalConfigFromEnv(); got == nil || got.Database != "default_database" {
		t.Fatalf("Chroma local config = %#v", got)
	}
	t.Setenv("CHROMA_CLOUD_API_KEY", "cloud-key")
	t.Setenv("CHROMA_TENANT", "team")
	t.Setenv("CHROMA_DATABASE", "vectors")
	if got := tryCreateChromaCloudConfigFromEnv(); got == nil || got.Tenant != "team" || got.Database != "vectors" {
		t.Fatalf("Chroma cloud config = %#v", got)
	}
	t.Setenv("CHROMA_CLOUD_API_KEY", "")
	t.Setenv("CHROMA_API_KEY", "fallback-key")
	if got := tryCreateChromaCloudConfigFromEnv(); got == nil || got.APIKey != "fallback-key" {
		t.Fatalf("Chroma fallback key config = %#v", got)
	}
}

func TestQdrantEnvironmentConfigs(t *testing.T) {
	clearVectorDBEnvironment(t)
	if got := tryCreateQdrantLocalConfigFromEnv(); got == nil || got.URL != "localhost:6334" {
		t.Fatalf("default Qdrant local config = %#v", got)
	}
	t.Setenv("QDRANT_URL", "localhost:7334")
	if got := tryCreateQdrantLocalConfigFromEnv(); got == nil || got.URL != "localhost:7334" {
		t.Fatalf("Qdrant local config = %#v", got)
	}
	if got := tryCreateQdrantCloudConfigFromEnv(); got != nil {
		t.Fatalf("Qdrant cloud without key = %#v", got)
	}
	t.Setenv("QDRANT_API_KEY", "key")
	if got := tryCreateQdrantCloudConfigFromEnv(); got == nil || got.URL != "localhost:7334" {
		t.Fatalf("Qdrant cloud fallback config = %#v", got)
	}
	t.Setenv("QDRANT_CLOUD_URL", "cloud.example:6334")
	t.Setenv("QDRANT_CLOUD_API_KEY", "cloud-key")
	if got := tryCreateQdrantCloudConfigFromEnv(); got == nil || got.URL != "cloud.example:6334" || got.APIKey != "cloud-key" {
		t.Fatalf("Qdrant cloud config = %#v", got)
	}
}

func TestNeo4jEnvironmentConfigs(t *testing.T) {
	clearVectorDBEnvironment(t)
	if got := tryCreateNeo4jLocalConfigFromEnv(); got != nil {
		t.Fatalf("Neo4j local without password = %#v", got)
	}
	t.Setenv("NEO4J_PASSWORD", "secret")
	if got := tryCreateNeo4jLocalConfigFromEnv(); got == nil || got.URL != "bolt://localhost:7687" || got.Username != "neo4j" {
		t.Fatalf("Neo4j local config = %#v", got)
	}
	if got := tryCreateNeo4jCloudConfigFromEnv(); got != nil {
		t.Fatalf("incomplete Neo4j cloud config = %#v", got)
	}
	t.Setenv("NEO4J_CLOUD_URL", "neo4j+s://cloud.example")
	t.Setenv("NEO4J_CLOUD_USERNAME", "cloud-user")
	t.Setenv("NEO4J_CLOUD_PASSWORD", "cloud-pass")
	t.Setenv("NEO4J_CLOUD_DATABASE", "vectors")
	if got := tryCreateNeo4jCloudConfigFromEnv(); got == nil || got.Database != "vectors" {
		t.Fatalf("Neo4j cloud config = %#v", got)
	}
}

func TestOpenSearchEnvironmentConfigs(t *testing.T) {
	clearVectorDBEnvironment(t)
	if got := tryCreateOpenSearchLocalConfigFromEnv(); got == nil || got.URL != "http://localhost:9200" {
		t.Fatalf("default OpenSearch local config = %#v", got)
	}
	t.Setenv("OPENSEARCH_LOCAL_ADDRESS", "http://localhost:19200")
	if got := tryCreateOpenSearchLocalConfigFromEnv(); got == nil || got.URL != "http://localhost:19200" {
		t.Fatalf("OpenSearch local config = %#v", got)
	}
	if got := tryCreateOpenSearchCloudConfigFromEnv(); got != nil {
		t.Fatalf("OpenSearch cloud without address = %#v", got)
	}
	t.Setenv("OPENSEARCH_CLOUD_ADDRESS", "https://search.example")
	if got := tryCreateOpenSearchCloudConfigFromEnv(); got != nil {
		t.Fatalf("OpenSearch cloud without credentials = %#v", got)
	}
	t.Setenv("OPENSEARCH_CLOUD_API_KEY", "key")
	if got := tryCreateOpenSearchCloudConfigFromEnv(); got == nil || got.APIKey != "key" {
		t.Fatalf("OpenSearch API key config = %#v", got)
	}
	t.Setenv("OPENSEARCH_CLOUD_API_KEY", "")
	t.Setenv("OPENSEARCH_CLOUD_USERNAME", "user")
	t.Setenv("OPENSEARCH_CLOUD_PASSWORD", "pass")
	if got := tryCreateOpenSearchCloudConfigFromEnv(); got == nil || got.Username != "user" {
		t.Fatalf("OpenSearch password config = %#v", got)
	}
}
