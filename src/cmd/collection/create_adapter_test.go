// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package collection

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/qdrant"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
	"github.com/spf13/viper"
)

type creationFactory struct {
	vectordb.ClientFactory
	client *creationClient
}

func (f *creationFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *creationFactory) CreateClient(*vectordb.Config) (vectordb.VectorDBClient, error) {
	return f.client, nil
}

type creationClient struct {
	vectordb.VectorDBClient
	names  []string
	schema *vectordb.CollectionSchema
}

func (c *creationClient) GetDefaultSchema(_ vectordb.SchemaType, name string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{Class: name}
}
func (c *creationClient) ValidateSchema(s *vectordb.CollectionSchema) error {
	if s.Class != "Target" {
		return fmt.Errorf("unexpected class %s", s.Class)
	}
	return nil
}
func (c *creationClient) CreateCollection(_ context.Context, name string, s *vectordb.CollectionSchema) error {
	c.names = append(c.names, name)
	c.schema = s
	return nil
}

func TestCreateNamedSchemaRoutesToAdapter(t *testing.T) {
	for _, tt := range []struct {
		kind    vectordb.VectorDBType
		factory vectordb.ClientFactory
	}{
		{vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory()},
		{vectordb.VectorDBTypeChromaLocal, chroma.NewFactory()},
		{vectordb.VectorDBTypeQdrantLocal, qdrant.NewFactory()},
		{vectordb.VectorDBTypeNeo4jLocal, neo4j.NewFactory()},
		{vectordb.VectorDBTypeSupabase, supabase.NewFactory()},
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
			old := CreateCmd.Flags().Lookup("schema").Value.String()
			if err := CreateCmd.Flags().Set("schema", "WeaveDocs"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = CreateCmd.Flags().Set("schema", old) })
			runCollectionCreate(CreateCmd, []string{"Target"})
			if len(client.names) != 1 || client.names[0] != "Target" {
				t.Fatalf("created collections = %v", client.names)
			}
			if tt.kind != vectordb.VectorDBTypeSupabase && (client.schema == nil || client.schema.Vectorizer != "text-embedding-3-small") {
				t.Fatalf("schema = %+v", client.schema)
			}
		})
	}
}
