// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package collection

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
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

type listingFactory struct {
	vectordb.ClientFactory
	created []vectordb.VectorDBType
	fail    bool
}

func (*listingFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *listingFactory) CreateClient(cfg *vectordb.Config) (vectordb.VectorDBClient, error) {
	f.created = append(f.created, cfg.Type)
	return &listingClient{fail: f.fail}, nil
}

type listingClient struct {
	vectordb.VectorDBClient
	fail bool
}

func (c *listingClient) ListCollections(context.Context) ([]vectordb.CollectionInfo, error) {
	if c.fail {
		return nil, errors.New("fixture listing unavailable")
	}
	return []vectordb.CollectionInfo{{Name: "Alpha", Description: "fixture collection"}}, nil
}
func (*listingClient) GetCollectionCount(context.Context, string) (int64, error) { return 3, nil }

func captureListingCommand(t *testing.T, run func()) string {
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

func TestCollectionListDatabaseSelection(t *testing.T) {
	for _, mode := range []string{"default", "all summary", "all details", "all JSON", "legacy JSON", "cloud", "local", "no cloud", "no local", "explicit", "failed summary"} {
		t.Run(mode, func(t *testing.T) {
			setupMockCollectionConfig(t)
			t.Setenv("VECTOR_DB_TYPE", "")
			for _, key := range []string{"SUPABASE_DATABASE_URL", "SUPABASE_DATABASE_KEY", "DATABASE_URL", "SUPABASE_PROJECT_URL", "SUPABASE_DATABASE_PASSWORD", "SUPABASE_PROJECT_API_KEY", "SUPABASE_ANON_KEY", "SUPABASE_KEY"} {
				t.Setenv(key, "")
			}
			local := "    - name: local-fixture\n      type: milvus-local\n      enabled: true\n"
			cloud := "    - name: cloud-fixture\n      type: milvus-cloud\n      enabled: true\n"
			if mode == "no cloud" {
				cloud = ""
			}
			if mode == "no local" {
				local = ""
			}
			if err := os.WriteFile(viper.GetString("config"), []byte("databases:\n  default: local-fixture\n  vector_databases:\n"+local+cloud), 0600); err != nil {
				t.Fatal(err)
			}
			factory := &listingFactory{fail: mode == "failed summary"}
			for _, kind := range []vectordb.VectorDBType{vectordb.VectorDBTypeMilvusLocal, vectordb.VectorDBTypeMilvusCloud} {
				vectordb.RegisterFactory(kind, factory)
				t.Cleanup(func() { vectordb.RegisterFactory(kind, milvus.NewFactory()) })
			}
			command := &cobra.Command{}
			for _, flag := range []string{"virtual", "summary", "details", "cloud", "local", "all", "json"} {
				command.Flags().Bool(flag, false, "")
			}
			command.Flags().Int("limit", 7, "")
			command.Flags().String("output", "text", "")
			set := func(key, value string) {
				if err := command.Flags().Set(key, value); err != nil {
					t.Fatal(err)
				}
			}
			want := []vectordb.VectorDBType{vectordb.VectorDBTypeMilvusLocal}
			var args []string
			switch mode {
			case "all summary", "all details", "all JSON", "legacy JSON", "failed summary":
				set("all", "true")
				want = append(want, vectordb.VectorDBTypeMilvusCloud)
				if mode == "all details" {
					set("details", "true")
				}
				if mode == "all JSON" {
					set("output", "json")
				}
				if mode == "legacy JSON" {
					set("json", "true")
				}
			case "cloud":
				set("cloud", "true")
				want = []vectordb.VectorDBType{vectordb.VectorDBTypeMilvusCloud}
			case "local":
				set("local", "true")
			case "no cloud":
				set("cloud", "true")
				want = nil
			case "no local":
				set("local", "true")
				want = nil
			case "explicit":
				args = []string{"cloud-fixture"}
				want = []vectordb.VectorDBType{vectordb.VectorDBTypeMilvusCloud}
			}
			out := captureListingCommand(t, func() { runCollectionList(command, args) })
			if !reflect.DeepEqual(factory.created, want) {
				t.Fatalf("created=%v want=%v output=%s", factory.created, want, out)
			}
			if strings.Contains(mode, "summary") {
				if !strings.Contains(out, "Total: 2 databases") {
					t.Fatalf("missing summary: %s", out)
				}
			} else if len(want) > 0 && !strings.Contains(out, "Alpha") {
				t.Fatalf("missing collection: %s", out)
			}
		})
	}
}

func TestCollectionListingAndCountAdapterRoutes(t *testing.T) {
	for _, adapter := range []struct {
		kind vectordb.VectorDBType
		real vectordb.ClientFactory
	}{
		{vectordb.VectorDBTypeSupabase, supabase.NewFactory()}, {vectordb.VectorDBTypeMongoDB, mongodb.NewFactory()},
		{vectordb.VectorDBTypeChromaLocal, chroma.NewFactory()}, {vectordb.VectorDBTypeQdrantLocal, qdrant.NewFactory()}, {vectordb.VectorDBTypeNeo4jLocal, neo4j.NewFactory()},
	} {
		t.Run(string(adapter.kind), func(t *testing.T) {
			setupMockCollectionConfig(t)
			factory := &listingFactory{}
			vectordb.RegisterFactory(adapter.kind, factory)
			t.Cleanup(func() { vectordb.RegisterFactory(adapter.kind, adapter.real) })
			cfg := &config.VectorDBConfig{Type: config.VectorDBType(adapter.kind)}
			out := captureListingCommand(t, func() { listCollectionsForDatabase(context.Background(), cfg, 1, false, true, false) })
			if len(factory.created) != 1 || factory.created[0] != adapter.kind || !strings.Contains(out, "Alpha") {
				t.Fatalf("route=%v output=%s", factory.created, out)
			}
			if adapter.kind == vectordb.VectorDBTypeQdrantLocal {
				t.Setenv("VECTOR_DB_TYPE", "")
				if err := os.WriteFile(viper.GetString("config"), []byte(fmt.Sprintf("databases:\n  default: fixture\n  vector_databases:\n    - name: fixture\n      type: %s\n      enabled: true\n", adapter.kind)), 0600); err != nil {
					t.Fatal(err)
				}
				out = captureListingCommand(t, func() { runCollectionCount(&cobra.Command{}, nil) })
				if !strings.Contains(out, "Total collections: 1") {
					t.Fatalf("count output=%s", out)
				}
			}
		})
	}
}
