// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"os"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestDatabaseEnvironmentDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		dbType string
		want   []string
	}{
		{"weaviate-cloud", []string{"WEAVIATE_URL", "WEAVIATE_API_KEY"}},
		{"supabase-cloud", []string{"SUPABASE_URL", "SUPABASE_KEY"}},
		{"milvus-local", []string{"MILVUS_ADDRESS"}},
		{"mongodb-cloud", []string{"MONGODB_URI"}},
		{"chroma-local", []string{"CHROMA_URL"}},
		{"qdrant-local", []string{"QDRANT_URL"}},
		{"neo4j-cloud", []string{"NEO4J_URI"}},
		{"opensearch-cloud", []string{"OPENSEARCH_URL"}},
		{"pinecone", []string{"PINECONE_API_KEY"}},
		{"mock", nil},
	} {
		t.Run(tc.dbType, func(t *testing.T) {
			got := envVarsForDB(config.VectorDBConfig{Type: config.VectorDBType(tc.dbType)})
			if len(got) != len(tc.want) {
				t.Fatalf("got %+v, want %v", got, tc.want)
			}
			for i, name := range tc.want {
				if got[i].name != name || got[i].required {
					t.Fatalf("got %+v, want optional %s", got[i], name)
				}
			}
		})
	}
}

func TestEmbeddingDiagnosticsWithoutModels(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := &config.Config{}
	cfg.Databases.VectorDatabases = []config.VectorDBConfig{
		{Name: "configured", VectorDimensions: 768},
		{Name: "unset"},
	}
	got := CheckEmbeddings(cfg)
	if len(got) != 1 || got[0].Status != StatusOK || got[0].Name != "configured dimensions" {
		t.Fatalf("configured dimensions: %+v", got)
	}
	if got := CheckEmbeddings(nil); len(got) != 1 || got[0].Status != StatusSkip {
		t.Fatalf("missing configuration: %+v", got)
	}
	stack := "version: '1'\nname: fixture\ninfrastructure:\n  vectordb:\n    type: milvus\n    version: '2'\ncollections:\n  - name: empty\n"
	if err := os.WriteFile("weave-stack.yaml", []byte(stack), 0600); err != nil {
		t.Fatal(err)
	}
	if got := CheckEmbeddings(nil); len(got) != 1 || got[0].Status != StatusSkip {
		t.Fatalf("stack without models: %+v", got)
	}
}
