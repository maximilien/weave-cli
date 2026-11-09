// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/mongodb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
)

type deleteFixtureFactory struct {
	vectordb.ClientFactory
	client *deleteFixtureClient
}

func (f *deleteFixtureFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *deleteFixtureFactory) CreateClient(*vectordb.Config) (vectordb.VectorDBClient, error) {
	return f.client, nil
}

type deleteFixtureClient struct {
	vectordb.VectorDBClient
	all     []string
	ids     []string
	filters []map[string]interface{}
	fail    bool
}

func (c *deleteFixtureClient) DeleteAllDocuments(_ context.Context, name string) error {
	c.all = append(c.all, name)
	return nil
}
func (c *deleteFixtureClient) DeleteDocument(_ context.Context, name, id string) error {
	c.ids = append(c.ids, name+"/"+id)
	return errors.New("fixture ID missing")
}
func (c *deleteFixtureClient) DeleteDocumentsByMetadata(_ context.Context, name string, filter map[string]interface{}) error {
	c.filters = append(c.filters, filter)
	if c.fail {
		return errors.New("fixture deletion rejected")
	}
	return nil
}
func (c *deleteFixtureClient) GetCollectionCount(context.Context, string) (int64, error) {
	return 0, errors.New("fixture count rejected")
}

func TestDeleteAllDocumentsRoutesToAdapter(t *testing.T) {
	for _, tt := range []struct {
		kind    vectordb.VectorDBType
		factory vectordb.ClientFactory
	}{
		{vectordb.VectorDBTypeSupabase, supabase.NewFactory()},
		{vectordb.VectorDBTypeMongoDB, mongodb.NewFactory()},
		{vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory()},
	} {
		t.Run(string(tt.kind), func(t *testing.T) {
			c := &deleteFixtureClient{}
			vectordb.RegisterFactory(tt.kind, &deleteFixtureFactory{client: c})
			t.Cleanup(func() { vectordb.RegisterFactory(tt.kind, tt.factory) })
			DeleteAllDocuments(context.Background(), &config.VectorDBConfig{Type: config.VectorDBType(tt.kind)}, "Chosen")
			if !reflect.DeepEqual(c.all, []string{"Chosen"}) {
				t.Fatalf("deleted collections = %v", c.all)
			}
		})
	}
}

func TestDeleteDocumentsFilenameFallbackAndFailures(t *testing.T) {
	for _, mode := range []string{"filename fallback", "failed fallback", "failed name", "failed metadata"} {
		t.Run(mode, func(t *testing.T) {
			c := &deleteFixtureClient{fail: mode != "filename fallback"}
			vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, &deleteFixtureFactory{client: c})
			t.Cleanup(func() { vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory()) })
			cfg := &config.VectorDBConfig{Type: config.VectorDBTypeMilvusLocal}
			var ids, filters []string
			name := ""
			want := map[string]interface{}{"original_filename": "chosen.txt"}
			switch mode {
			case "filename fallback", "failed fallback":
				ids = []string{"chosen.txt"}
			case "failed name":
				name = "chosen.txt"
			case "failed metadata":
				filters = []string{"status=ready=now"}
				want = map[string]interface{}{"status": "ready=now"}
			}
			out, diag := captureCollectionListing(t, func() { DeleteDocuments(context.Background(), cfg, "Chosen", ids, filters, false, "", name) })
			if len(c.filters) != 1 || !reflect.DeepEqual(c.filters[0], want) {
				t.Fatalf("deletion filters = %#v", c.filters)
			}
			if len(ids) > 0 && !reflect.DeepEqual(c.ids, []string{"Chosen/chosen.txt"}) {
				t.Fatalf("ID deletions = %v", c.ids)
			}
			if c.fail && (!strings.Contains(out+diag, "Failed to") || strings.Contains(out+diag, "deleted successfully")) {
				t.Fatalf("failure output = %s %s", out, diag)
			}
			if !c.fail && !strings.Contains(out+diag, "deleted successfully") {
				t.Fatalf("success output = %s %s", out, diag)
			}
			if _, err := CountDocuments(context.Background(), cfg, "Chosen"); err == nil || !strings.Contains(err.Error(), "failed to count documents") {
				t.Fatalf("count error = %v", err)
			}
		})
	}
}
