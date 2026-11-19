// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

type listingTimeoutFactory struct {
	vectordb.ClientFactory
	client *listingTimeoutClient
}

func (*listingTimeoutFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *listingTimeoutFactory) CreateClient(*vectordb.Config) (vectordb.VectorDBClient, error) {
	return f.client, nil
}

type listingTimeoutClient struct {
	vectordb.VectorDBClient
	collection    string
	limit, offset int
}

func (c *listingTimeoutClient) ListDocuments(ctx context.Context, collection string, limit, offset int) ([]*vectordb.Document, error) {
	c.collection, c.limit, c.offset = collection, limit, offset
	return nil, ctx.Err()
}

func TestDocumentListingExpiredContextGuidance(t *testing.T) {
	for _, generic := range []bool{false, true} {
		name := "weaviate"
		if generic {
			name = "generic"
		}
		t.Run(name, func(t *testing.T) {
			cfg, _ := documentOperationsFixture(t, "none", false, false, false, false)
			client := &listingTimeoutClient{}
			if generic {
				vectordb.RegisterFactory(vectordb.VectorDBTypeWeaviateLocal, &listingTimeoutFactory{client: client})
				t.Cleanup(func() { vectordb.RegisterFactory(vectordb.VectorDBTypeWeaviateLocal, weaviate.NewFactory()) })
			}
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			defer cancel()
			out, diagnostics := captureCollectionListing(t, func() {
				if generic {
					ListDocuments(ctx, cfg, "Docs", 7, 2, false, 3, false, false, false, false)
				} else {
					ListWeaviateDocuments(ctx, cfg, "Docs", 7, false, 3, false, false)
				}
			})
			for _, want := range []string{"connection timeout after 30 seconds", "server is running and accessible", "config is correct", "network/firewall"} {
				if !strings.Contains(out+diagnostics, want) {
					t.Fatalf("missing %q: %s %s", want, out, diagnostics)
				}
			}
			if generic && (client.collection != "Docs" || client.limit != 7 || client.offset != 2) {
				t.Fatalf("listing arguments=%+v", client)
			}
			if strings.Contains(out+diagnostics, "First content") {
				t.Fatalf("displayed documents after timeout: %s %s", out, diagnostics)
			}
		})
	}
}
