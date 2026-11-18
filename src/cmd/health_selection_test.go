// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package cmd

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type healthSelectionFactory struct {
	vectordb.ClientFactory
	checked []vectordb.VectorDBType
}

func (f *healthSelectionFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *healthSelectionFactory) CreateClient(cfg *vectordb.Config) (vectordb.VectorDBClient, error) {
	f.checked = append(f.checked, cfg.Type)
	return &healthSelectionClient{}, nil
}

type healthSelectionClient struct{ vectordb.VectorDBClient }

func (*healthSelectionClient) Health(context.Context) error { return nil }
func (*healthSelectionClient) ListCollections(context.Context) ([]vectordb.CollectionInfo, error) {
	return []vectordb.CollectionInfo{{Name: "Zeta"}, {Name: "Alpha"}}, nil
}

func TestHealthCommandSelectionAndOutputModes(t *testing.T) {
	for _, mode := range []string{"default", "all JSON", "all details", "all summary", "cloud", "local", "no cloud", "no local", "explicit JSON"} {
		t.Run(mode, func(t *testing.T) {
			setupHealthConfig(t)
			t.Setenv("VECTOR_DB_TYPE", "")
			t.Setenv("SUPABASE_URL", "")
			t.Setenv("SUPABASE_KEY", "")
			local := "    - name: local-fixture\n      type: weaviate-local\n      enabled: true\n      url: https://local.fixture\n"
			cloud := "    - name: cloud-fixture\n      type: weaviate-cloud\n      enabled: true\n      url: https://cloud.fixture\n"
			if mode == "no cloud" {
				cloud = ""
			}
			if mode == "no local" {
				local = ""
			}
			if err := os.WriteFile(viper.GetString("config"), []byte("databases:\n  default: local-fixture\n  vector_databases:\n"+local+cloud), 0600); err != nil {
				t.Fatal(err)
			}
			factory := &healthSelectionFactory{}
			for _, kind := range []vectordb.VectorDBType{vectordb.VectorDBTypeWeaviateLocal, vectordb.VectorDBTypeWeaviateCloud} {
				vectordb.RegisterFactory(kind, factory)
				t.Cleanup(func() { vectordb.RegisterFactory(kind, weaviate.NewFactory()) })
			}
			command := &cobra.Command{}
			for _, flag := range []string{"json", "summary", "details", "cloud", "local", "all"} {
				command.Flags().Bool(flag, false, "")
			}
			command.Flags().String("sort-by", "type", "")
			set := func(flag string) {
				if err := command.Flags().Set(flag, "true"); err != nil {
					t.Fatal(err)
				}
			}
			want := []vectordb.VectorDBType{vectordb.VectorDBTypeWeaviateLocal}
			var args []string
			switch mode {
			case "all JSON", "all details", "all summary":
				set("all")
				want = append(want, vectordb.VectorDBTypeWeaviateCloud)
				if mode == "all JSON" {
					set("json")
				}
				if mode == "all details" {
					set("details")
				}
				if mode == "all summary" {
					set("summary")
					want = []vectordb.VectorDBType{vectordb.VectorDBTypeWeaviateCloud, vectordb.VectorDBTypeWeaviateLocal}
				}
			case "cloud":
				set("cloud")
				want = []vectordb.VectorDBType{vectordb.VectorDBTypeWeaviateCloud}
			case "local":
				set("local")
			case "no cloud":
				set("cloud")
				want = nil
			case "no local":
				set("local")
				want = nil
			case "explicit JSON":
				set("json")
				args = []string{"cloud-fixture"}
				want = []vectordb.VectorDBType{vectordb.VectorDBTypeWeaviateCloud}
			}
			out := captureHealthOutput(t, func() { runHealthCheck(command, args) })
			if !reflect.DeepEqual(factory.checked, want) {
				t.Fatalf("checked types=%v want=%v output=%s", factory.checked, want, out)
			}
			if strings.Contains(mode, "JSON") {
				var response struct {
					Results []HealthCheckResult `json:"results"`
				}
				if err := json.Unmarshal([]byte(out), &response); err != nil {
					t.Fatalf("invalid JSON: %v %s", err, out)
				}
				if len(response.Results) != len(want) {
					t.Fatalf("JSON results=%+v", response)
				}
				for _, result := range response.Results {
					if !result.Healthy || result.CollectionsCount != 2 || !reflect.DeepEqual(result.Collections, []string{"Alpha", "Zeta"}) {
						t.Fatalf("health result=%+v", result)
					}
				}
			}
		})
	}
}
