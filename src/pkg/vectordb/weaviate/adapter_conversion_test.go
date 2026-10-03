package weaviate

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestAdapterDocumentConversions(t *testing.T) {
	a := &Adapter{}
	if a.convertDocument(nil) != nil || a.convertDocumentFromWeaviate(nil) != nil || a.convertDocumentsFromWeaviate(nil) != nil {
		t.Fatal("nil conversions should remain nil")
	}
	doc := &vectordb.Document{ID: "id", Text: "text", Content: "content", Metadata: map[string]interface{}{"k": "v"}}
	converted := a.convertDocument(doc)
	if converted.ID != doc.ID || converted.Text != doc.Text || converted.Content != doc.Content {
		t.Fatalf("conversion mismatch: %#v", converted)
	}
	back := a.convertDocumentFromWeaviate(converted)
	if back.ID != doc.ID || back.Metadata["k"] != "v" {
		t.Fatalf("reverse conversion mismatch: %#v", back)
	}
	all := a.convertDocumentsFromWeaviate([]*Document{converted})
	if len(all) != 1 || all[0].ID != doc.ID {
		t.Fatalf("slice conversion mismatch: %#v", all)
	}
	if opts := a.convertQueryOptions(nil); opts == nil {
		t.Fatal("nil query options conversion")
	}
	q := a.convertQueryOptions(&vectordb.QueryOptions{TopK: 3, Distance: 0.4, SearchMetadata: true, NoTruncate: true, UseBM25: true})
	if q.TopK != 3 || q.Distance != 0.4 || !q.SearchMetadata || !q.NoTruncate || !q.UseBM25 {
		t.Fatalf("query conversion mismatch: %#v", q)
	}
	schema := a.convertSchemaFromWeaviate(&CollectionSchema{Class: "Docs", Vectorizer: "none", Properties: []SchemaProperty{{Name: "meta", DataType: []string{"object"}, NestedProperties: []SchemaProperty{{Name: "author"}}}}})
	if schema.Class != "Docs" || len(schema.Properties) != 1 || len(schema.Properties[0].NestedProperties) != 1 {
		t.Fatalf("schema conversion mismatch: %#v", schema)
	}
}
