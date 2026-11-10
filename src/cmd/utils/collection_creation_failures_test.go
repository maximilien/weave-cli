// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
)

type collectionCreationFactory struct {
	vectordb.ClientFactory
	client *collectionCreationClient
	err    error
}

func (f *collectionCreationFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *collectionCreationFactory) CreateClient(*vectordb.Config) (vectordb.VectorDBClient, error) {
	return f.client, f.err
}

type collectionCreationClient struct {
	vectordb.VectorDBClient
	invalid bool
	fail    bool
	names   []string
}

func (c *collectionCreationClient) GetDefaultSchema(_ vectordb.SchemaType, name string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{Class: name}
}
func (c *collectionCreationClient) ValidateSchema(*vectordb.CollectionSchema) error {
	if c.invalid {
		return errors.New("fixture invalid schema")
	}
	return nil
}
func (c *collectionCreationClient) CreateCollection(_ context.Context, name string, _ *vectordb.CollectionSchema) error {
	c.names = append(c.names, name)
	if c.fail {
		return errors.New("fixture create failed")
	}
	return nil
}

func TestCollectionCreationHelpersPropagateFailures(t *testing.T) {
	for _, mode := range []string{"supabase success", "supabase client failure", "supabase create failure", "generic client failure", "generic invalid schema", "generic create failure"} {
		t.Run(mode, func(t *testing.T) {
			c := &collectionCreationClient{invalid: strings.Contains(mode, "invalid schema"), fail: strings.Contains(mode, "create failure")}
			f := &collectionCreationFactory{client: c}
			if strings.Contains(mode, "client failure") {
				f.err = errors.New("fixture client failed")
			}
			vectordb.RegisterFactory(vectordb.VectorDBTypeSupabase, f)
			t.Cleanup(func() { vectordb.RegisterFactory(vectordb.VectorDBTypeSupabase, supabase.NewFactory()) })
			cfg := &config.VectorDBConfig{Type: config.VectorDBTypeSupabase}
			var err error
			if strings.HasPrefix(mode, "supabase") {
				err = CreateSupabaseCollection(context.Background(), cfg, "Target", "text-embedding-3-small")
			} else {
				err = CreateGenericCollectionWithSchemaType(context.Background(), cfg, "Target", "text-embedding-3-small", vectordb.SchemaTypeText)
			}
			if mode == "supabase success" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				want := "fixture create failed"
				if f.err != nil {
					want = "fixture client failed"
				}
				if c.invalid {
					want = "fixture invalid schema"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("creation error=%v; want %q", err, want)
				}
			}
			wantCalls := 1
			if f.err != nil || c.invalid {
				wantCalls = 0
			}
			if len(c.names) != wantCalls || (wantCalls == 1 && c.names[0] != "Target") {
				t.Fatalf("creation calls=%v", c.names)
			}
		})
	}
}
