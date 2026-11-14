// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package collection

import (
	"fmt"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/opensearch"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/qdrant"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestCreateDefaultCollectionRoutesToAdapter(t *testing.T) {
	for _, tt := range []struct {
		kind    vectordb.VectorDBType
		factory vectordb.ClientFactory
	}{
		{vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory()},
		{vectordb.VectorDBTypeNeo4jLocal, neo4j.NewFactory()},
		{vectordb.VectorDBTypeQdrantLocal, qdrant.NewFactory()},
		{vectordb.VectorDBTypeSupabase, supabase.NewFactory()},
		{vectordb.VectorDBTypeOpenSearchLocal, opensearch.NewFactory()},
	} {
		t.Run(string(tt.kind), func(t *testing.T) {
			setupMockCollectionConfig(t)
			t.Setenv("VECTOR_DB_TYPE", "")
			if err := os.WriteFile(viper.GetString("config"), []byte(fmt.Sprintf("databases:\n  default: %s\n  vector_databases:\n    - name: fixture\n      type: %s\n      enabled: true\n", tt.kind, tt.kind)), 0600); err != nil {
				t.Fatal(err)
			}
			client := &creationClient{}
			vectordb.RegisterFactory(tt.kind, &creationFactory{client: client})
			t.Cleanup(func() { vectordb.RegisterFactory(tt.kind, tt.factory) })
			runCollectionCreate(CreateCmd, []string{"Target"})
			if len(client.names) != 1 || client.names[0] != "Target" {
				t.Fatalf("creation calls=%v", client.names)
			}
			if tt.kind != vectordb.VectorDBTypeSupabase && (client.schema == nil || client.schema.Vectorizer != "text-embedding-3-small") {
				t.Fatalf("schema=%+v", client.schema)
			}
		})
	}
}
