// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package document

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/mongodb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/qdrant"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type countAdapterFactory struct {
	vectordb.ClientFactory
	client  *countAdapterClient
	created []vectordb.VectorDBType
}

func (*countAdapterFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *countAdapterFactory) CreateClient(cfg *vectordb.Config) (vectordb.VectorDBClient, error) {
	f.created = append(f.created, cfg.Type)
	return f.client, nil
}

type countAdapterClient struct {
	vectordb.VectorDBClient
	names  []string
	listed string
}

func (c *countAdapterClient) GetCollectionCount(_ context.Context, name string) (int64, error) {
	c.names = append(c.names, name)
	if name == "Missing" {
		return 0, errors.New("fixture count unavailable")
	}
	return 7, nil
}
func (c *countAdapterClient) ListDocuments(_ context.Context, name string, limit, offset int) ([]*vectordb.Document, error) {
	c.listed = fmt.Sprintf("%s/%d/%d", name, limit, offset)
	return nil, nil
}

func captureCountCommand(t *testing.T, run func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	previous := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = previous }()
	run()
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestDocumentCountAdapterRoutesAndPartialFailures(t *testing.T) {
	for _, adapter := range []struct {
		kind vectordb.VectorDBType
		real vectordb.ClientFactory
	}{
		{vectordb.VectorDBTypeSupabase, supabase.NewFactory()}, {vectordb.VectorDBTypeMongoDB, mongodb.NewFactory()},
		{vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory()}, {vectordb.VectorDBTypeChromaLocal, chroma.NewFactory()},
		{vectordb.VectorDBTypeQdrantLocal, qdrant.NewFactory()}, {vectordb.VectorDBTypeNeo4jLocal, neo4j.NewFactory()},
	} {
		t.Run(string(adapter.kind), func(t *testing.T) {
			setupMockDocumentConfig(t)
			t.Setenv("VECTOR_DB_TYPE", "")
			if err := os.WriteFile(viper.GetString("config"), []byte(fmt.Sprintf("databases:\n  default: fixture\n  vector_databases:\n    - name: fixture\n      type: %s\n      enabled: true\n", adapter.kind)), 0600); err != nil {
				t.Fatal(err)
			}
			client := &countAdapterClient{}
			factory := &countAdapterFactory{client: client}
			vectordb.RegisterFactory(adapter.kind, factory)
			t.Cleanup(func() { vectordb.RegisterFactory(adapter.kind, adapter.real) })
			for _, jsonOutput := range []bool{false, true} {
				client.names = nil
				factory.created = nil
				command := &cobra.Command{}
				command.Flags().Bool("json", jsonOutput, "")
				out := captureCountCommand(t, func() { runDocumentCount(command, []string{"Docs", "Missing", "Images"}) })
				if !reflect.DeepEqual(client.names, []string{"Docs", "Missing", "Images"}) {
					t.Fatalf("counted=%v", client.names)
				}
				if !reflect.DeepEqual(factory.created, []vectordb.VectorDBType{adapter.kind, adapter.kind, adapter.kind}) {
					t.Fatalf("created=%v", factory.created)
				}
				if jsonOutput {
					var results []struct {
						Collection string
						Count      int
					}
					if err := json.Unmarshal([]byte(out), &results); err != nil {
						t.Fatalf("invalid count JSON: %v %s", err, out)
					}
					if len(results) != 2 || results[0].Collection != "Docs" || results[1].Collection != "Images" || results[0].Count != 7 || results[1].Count != 7 {
						t.Fatalf("counts=%+v", results)
					}
				} else if strings.Count(out, "Total documents: 7") != 2 {
					t.Fatalf("count output=%s", out)
				}
			}
			// A multi-database listing should select the configured default.
			if adapter.kind == vectordb.VectorDBTypeMilvusLocal {
				t.Setenv("VECTOR_DB_TYPE", "milvus-local")
				f, err := os.OpenFile(viper.GetString("config"), os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.WriteString("    - name: other\n      type: mock\n      enabled: true\n")
				closeErr := f.Close()
				if err != nil {
					t.Fatal(err)
				}
				if closeErr != nil {
					t.Fatal(closeErr)
				}
				command := &cobra.Command{}
				command.Flags().Bool("all", true, "")
				command.Flags().Bool("json", true, "")
				command.Flags().Int("limit", 9, "")
				command.Flags().Int("offset", 2, "")
				out := captureCountCommand(t, func() { runDocumentList(command, []string{"Docs"}) })
				if client.listed != "Docs/9/2" {
					t.Fatalf("default listing=%s output=%s", client.listed, out)
				}
			}
		})
	}
}
