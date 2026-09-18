// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package supabase

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type supabaseSQLState struct {
	mu                sync.Mutex
	execs             []string
	queries           []string
	execError         error
	queryError        error
	pingError         error
	rowsAffected      int64
	rowsAffectedError error
}

type supabaseConnector struct{ state *supabaseSQLState }

func (c *supabaseConnector) Connect(context.Context) (driver.Conn, error) {
	return &supabaseConn{state: c.state}, nil
}
func (c *supabaseConnector) Driver() driver.Driver { return supabaseDriver{} }

type supabaseDriver struct{}

func (supabaseDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type supabaseConn struct{ state *supabaseSQLState }

func (c *supabaseConn) Prepare(query string) (driver.Stmt, error) {
	return &supabaseStmt{state: c.state, query: query}, nil
}
func (c *supabaseConn) PrepareContext(_ context.Context, query string) (driver.Stmt, error) {
	return c.Prepare(query)
}
func (c *supabaseConn) Close() error              { return nil }
func (c *supabaseConn) Begin() (driver.Tx, error) { return &supabaseTx{}, nil }
func (c *supabaseConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &supabaseTx{}, nil
}
func (c *supabaseConn) Ping(context.Context) error { return c.state.pingError }

func (c *supabaseConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.execs = append(c.state.execs, query)
	if c.state.execError != nil {
		return nil, c.state.execError
	}
	affected := c.state.rowsAffected
	if affected == 0 {
		affected = 1
	}
	return supabaseResult{affected: affected, err: c.state.rowsAffectedError}, nil
}

func (c *supabaseConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	c.state.mu.Lock()
	c.state.queries = append(c.state.queries, query)
	err := c.state.queryError
	c.state.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return supabaseRowsForQuery(query), nil
}

type supabaseStmt struct {
	state *supabaseSQLState
	query string
}

func (s *supabaseStmt) Close() error  { return nil }
func (s *supabaseStmt) NumInput() int { return -1 }
func (s *supabaseStmt) Exec([]driver.Value) (driver.Result, error) {
	return (&supabaseConn{state: s.state}).ExecContext(context.Background(), s.query, nil)
}
func (s *supabaseStmt) ExecContext(ctx context.Context, _ []driver.NamedValue) (driver.Result, error) {
	return (&supabaseConn{state: s.state}).ExecContext(ctx, s.query, nil)
}
func (s *supabaseStmt) Query([]driver.Value) (driver.Rows, error) {
	return (&supabaseConn{state: s.state}).QueryContext(context.Background(), s.query, nil)
}

type supabaseTx struct{}

func (*supabaseTx) Commit() error   { return nil }
func (*supabaseTx) Rollback() error { return nil }

type supabaseResult struct {
	affected int64
	err      error
}

func (r supabaseResult) LastInsertId() (int64, error) { return 0, nil }
func (r supabaseResult) RowsAffected() (int64, error) { return r.affected, r.err }

type supabaseRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *supabaseRows) Columns() []string { return r.columns }
func (r *supabaseRows) Close() error      { return nil }
func (r *supabaseRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func supabaseRowsForQuery(query string) driver.Rows {
	switch {
	case strings.Contains(query, "FROM information_schema.tables t"):
		return &supabaseRows{
			columns: []string{"table_name", "description", "count"},
			values:  [][]driver.Value{{"collection_docs", "fixture collection", int64(2)}},
		}
	case strings.Contains(query, "SELECT COUNT(*) FROM information_schema.tables"):
		return &supabaseRows{columns: []string{"count"}, values: [][]driver.Value{{int64(1)}}}
	case strings.Contains(query, "FROM information_schema.columns"):
		return &supabaseRows{
			columns: []string{"column_name", "data_type", "is_nullable", "column_default"},
			values: [][]driver.Value{
				{"id", "text", "NO", nil},
				{"category", "character varying", "YES", nil},
				{"priority", "integer", "YES", nil},
			},
		}
	case strings.Contains(query, "SELECT vector_dimensions"):
		return &supabaseRows{columns: []string{"vector_dimensions"}, values: [][]driver.Value{{int64(3)}}}
	case strings.Contains(query, "SELECT COUNT(*) FROM \"collection_"):
		return &supabaseRows{columns: []string{"count"}, values: [][]driver.Value{{int64(2)}}}
	case strings.Contains(query, "as score"):
		return &supabaseRows{
			columns: []string{"id", "content", "text", "image", "image_data", "url", "metadata", "score"},
			values:  [][]driver.Value{{"doc-1", "hello world", "hello", "", "", "https://example.test", `{"kind":"guide"}`, float64(0.8)}},
		}
	case strings.Contains(query, "SELECT id, content, text, image, image_data, url, metadata"):
		return &supabaseRows{
			columns: []string{"id", "content", "text", "image", "image_data", "url", "metadata"},
			values:  [][]driver.Value{{"doc-1", "hello world", "hello", "", "", "https://example.test", `{"kind":"guide"}`}},
		}
	default:
		return &supabaseRows{columns: []string{"value"}}
	}
}

func newSupabaseSQLFixture(t *testing.T) (*Adapter, *supabaseSQLState) {
	t.Helper()
	state := &supabaseSQLState{rowsAffected: 1}
	db := sql.OpenDB(&supabaseConnector{state: state})
	t.Cleanup(func() { _ = db.Close() })
	return &Adapter{
		db: db,
		config: &vectordb.Config{
			Type:             vectordb.VectorDBTypeSupabaseLocal,
			Timeout:          1,
			VectorDimensions: 3,
			SimilarityMetric: "cosine",
		},
	}, state
}

func TestSupabaseCollectionAndSchemaProtocol(t *testing.T) {
	adapter, state := newSupabaseSQLFixture(t)
	ctx := context.Background()
	schema := &vectordb.CollectionSchema{Class: "docs", Properties: []vectordb.SchemaProperty{
		{Name: "content", DataType: []string{"text"}},
		{Name: "category", DataType: []string{"text"}},
		{Name: "priority", DataType: []string{"integer"}},
	}}

	if err := adapter.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if err := adapter.CreateCollection(ctx, "docs", schema); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	if len(state.execs) < 6 || !strings.Contains(strings.Join(state.execs, "\n"), "category TEXT") {
		t.Fatalf("CreateCollection() execs = %#v", state.execs)
	}
	exists, err := adapter.CollectionExists(ctx, "docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists() = %v, %v", exists, err)
	}
	count, err := adapter.GetCollectionCount(ctx, "docs")
	if err != nil || count != 2 {
		t.Fatalf("GetCollectionCount() = %d, %v", count, err)
	}
	gotSchema, err := adapter.GetSchema(ctx, "docs")
	if err != nil || len(gotSchema.Properties) != 2 || gotSchema.Properties[0].Name != "category" {
		t.Fatalf("GetSchema() = %#v, %v", gotSchema, err)
	}
	if err := adapter.UpdateSchema(ctx, "docs", &vectordb.CollectionSchema{Properties: []vectordb.SchemaProperty{
		{Name: "category", DataType: []string{"text"}},
		{Name: "published", DataType: []string{"boolean"}},
	}}); err != nil {
		t.Fatalf("UpdateSchema() error = %v", err)
	}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) != 1 || collections[0].Name != "docs" || collections[0].Count != 2 {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}
	dimensions, err := adapter.getCollectionDimensions(ctx, "docs")
	if err != nil || dimensions != 3 {
		t.Fatalf("getCollectionDimensions() = %d, %v", dimensions, err)
	}
	if err := adapter.DeleteCollection(ctx, "docs"); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}

	for pgType, want := range map[string]string{
		"varchar": "text", "bigint": "int", "double precision": "float", "bool": "boolean",
		"timestamp": "date", "uuid": "uuid", "jsonb": "json", "bytea": "text",
	} {
		if got := adapter.convertPostgreSQLToDataType(pgType); got != want {
			t.Errorf("convertPostgreSQLToDataType(%q) = %q, want %q", pgType, got, want)
		}
	}
}

func TestSupabaseDocumentProtocol(t *testing.T) {
	adapter, _ := newSupabaseSQLFixture(t)
	ctx := context.Background()
	doc := &vectordb.Document{
		ID: "doc-1", Content: "hello world", Text: "hello", URL: "https://example.test",
		Metadata: map[string]interface{}{"kind": "guide"},
	}

	if err := adapter.CreateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{doc, {ID: "doc-2", Content: "second"}}); err != nil {
		t.Fatalf("CreateDocuments() error = %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("empty CreateDocuments() error = %v", err)
	}
	got, err := adapter.GetDocument(ctx, "docs", "doc-1")
	if err != nil || got.ID != "doc-1" || got.Metadata["kind"] != "guide" {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	if err := adapter.UpdateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	if err := adapter.DeleteDocument(ctx, "docs", "doc-1"); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", []string{"doc-1", "doc-2"}); err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("empty DeleteDocuments() error = %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide", "rank": 2}); err != nil {
		t.Fatalf("DeleteDocumentsByMetadata() error = %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", nil); err == nil {
		t.Fatal("DeleteDocumentsByMetadata(nil) expected error")
	}
	if err := adapter.DeleteAllDocuments(ctx, "docs"); err != nil {
		t.Fatalf("DeleteAllDocuments() error = %v", err)
	}
	docs, err := adapter.ListDocuments(ctx, "docs", 10, 0)
	if err != nil || len(docs) != 1 || docs[0].Content != "hello world" {
		t.Fatalf("ListDocuments() = %#v, %v", docs, err)
	}
	if got := floatsToVector([]float64{0.1, 0.2}); got != "[0.100000,0.200000]" {
		t.Fatalf("floatsToVector() = %q", got)
	}
	if got := floatsToVector(nil); got != "[]" {
		t.Fatalf("floatsToVector(nil) = %q", got)
	}
}

func TestSupabaseQueryProtocol(t *testing.T) {
	adapter, _ := newSupabaseSQLFixture(t)
	ctx := context.Background()
	options := &vectordb.QueryOptions{TopK: 3}

	semantic, err := adapter.SearchSemantic(ctx, "docs", "hello", options)
	if err != nil || len(semantic) != 1 || semantic[0].Score != 0.8 {
		t.Fatalf("SearchSemantic() = %#v, %v", semantic, err)
	}
	bm25, err := adapter.SearchBM25(ctx, "docs", "hello", options)
	if err != nil || len(bm25) != 1 {
		t.Fatalf("SearchBM25() = %#v, %v", bm25, err)
	}
	metadata, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, options)
	if err != nil || len(metadata) != 1 || metadata[0].Score != 1 {
		t.Fatalf("SearchByMetadata() = %#v, %v", metadata, err)
	}
	empty, err := adapter.SearchByMetadata(ctx, "docs", nil, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty SearchByMetadata() = %#v, %v", empty, err)
	}
	hybrid, err := adapter.SearchHybrid(ctx, "docs", "hello", options)
	if err != nil || len(hybrid) != 1 || hybrid[0].Score <= 0 {
		t.Fatalf("SearchHybrid() = %#v, %v", hybrid, err)
	}
	vector, err := adapter.searchByVectorSimilarity(ctx, "docs", []float64{0.1, 0.2, 0.3}, options)
	if err != nil || len(vector) != 1 {
		t.Fatalf("searchByVectorSimilarity() = %#v, %v", vector, err)
	}

	merged := adapter.mergeSearchResults(
		[]*vectordb.QueryResult{
			{Document: vectordb.Document{ID: "a"}, Score: 0.5},
			{Document: vectordb.Document{ID: "b"}, Score: 0.9},
		},
		[]*vectordb.QueryResult{
			{Document: vectordb.Document{ID: "a"}, Score: 1},
			{Document: vectordb.Document{ID: "c"}, Score: 0.2},
		},
		&vectordb.QueryOptions{TopK: 2},
	)
	if len(merged) != 2 || merged[0].Document.ID != "a" {
		t.Fatalf("mergeSearchResults() = %#v", merged)
	}
}

func TestSupabaseProtocolErrorsAndFallbacks(t *testing.T) {
	ctx := context.Background()
	adapter, state := newSupabaseSQLFixture(t)

	state.queryError = errors.New("fixture query failure")
	for name, call := range map[string]func() error{
		"exists": func() error { _, err := adapter.CollectionExists(ctx, "docs"); return err },
		"list":   func() error { _, err := adapter.ListCollections(ctx); return err },
		"get":    func() error { _, err := adapter.GetDocument(ctx, "docs", "doc-1"); return err },
		"search": func() error { _, err := adapter.SearchBM25(ctx, "docs", "hello", nil); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("expected query error")
			}
		})
	}
	state.queryError = nil

	state.execError = errors.New("fixture exec failure")
	for name, call := range map[string]func() error{
		"vector extension":  func() error { return adapter.ensureVectorExtension(ctx) },
		"metadata table":    func() error { return adapter.ensureMetadataTable(ctx) },
		"delete collection": func() error { return adapter.DeleteCollection(ctx, "docs") },
		"create document": func() error {
			return adapter.CreateDocument(ctx, "docs", &vectordb.Document{ID: "doc-1"})
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("expected exec error")
			}
		})
	}
	state.execError = nil

	state.queryError = errors.New("metadata missing")
	dimensions, err := adapter.getCollectionDimensions(ctx, "docs")
	if err != nil || dimensions != 3 {
		t.Fatalf("configured dimension fallback = %d, %v", dimensions, err)
	}
	adapter.config.VectorDimensions = 0
	dimensions, err = adapter.getCollectionDimensions(ctx, "docs")
	if err != nil || dimensions != 1536 {
		t.Fatalf("default dimension fallback = %d, %v", dimensions, err)
	}
	state.queryError = nil

	state.rowsAffectedError = errors.New("rows affected failure")
	if err := adapter.DeleteDocument(ctx, "docs", "doc-1"); err == nil {
		t.Fatal("DeleteDocument() expected rows-affected error")
	}
	state.rowsAffectedError = nil
	state.rowsAffected = -1
	if err := adapter.UpdateDocument(ctx, "docs", &vectordb.Document{ID: "missing"}); err != nil {
		t.Fatalf("UpdateDocument() with fixture result error = %v", err)
	}

	for _, message := range []string{"connection refused", "deadline timeout", "password authentication failed", "plain failure"} {
		adapter, state := newSupabaseSQLFixture(t)
		state.pingError = errors.New(message)
		err := adapter.Health(ctx)
		if err == nil || !strings.Contains(err.Error(), strings.Fields(message)[0]) {
			t.Fatalf("Health(%q) error = %v", message, err)
		}
	}

	noDB := &Adapter{config: &vectordb.Config{Type: vectordb.VectorDBTypeSupabaseLocal}}
	if err := noDB.Health(ctx); err == nil {
		t.Fatal("Health() without DB expected error")
	}
}
