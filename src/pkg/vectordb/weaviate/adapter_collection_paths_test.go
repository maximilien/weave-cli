package weaviate

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestAdapterCollectionCreateErrorPath(t *testing.T) {
	client, _ := newQueryProtocolClient(t)
	adapter := &Adapter{client: client}
	ctx := context.Background()
	for _, schema := range []*vectordb.CollectionSchema{
		{Class: "Docs", Vectorizer: "text-embedding-3-small", Properties: []vectordb.SchemaProperty{{Name: "text", DataType: []string{"text"}}}},
		{Class: "Images", Properties: []vectordb.SchemaProperty{{Name: "image", DataType: []string{"blob"}}, {Name: "image_data", DataType: []string{"blob"}}}},
	} {
		_ = adapter.CreateCollection(ctx, schema.Class, schema)
	}
}
