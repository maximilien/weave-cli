// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	goredis "github.com/redis/go-redis/v9"
)

type respFixture struct {
	mu       sync.Mutex
	commands [][]string
	fail     map[string]string
}

func newProtocolClient(t *testing.T, fixture *respFixture) *Client {
	t.Helper()

	rdb := goredis.NewClient(&goredis.Options{
		Addr:            "fixture:6379",
		Protocol:        2,
		DisableIdentity: true,
		DialTimeout:     time.Second,
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		Dialer: func(context.Context, string, string) (net.Conn, error) {
			client, server := net.Pipe()
			go fixture.serve(server)
			return client, nil
		},
	})
	t.Cleanup(func() { _ = rdb.Close() })

	return &Client{
		rdb: rdb,
		config: &Config{
			Addr:             "fixture:6379",
			Timeout:          1,
			VectorDimensions: 3,
			SimilarityMetric: "COSINE",
		},
	}
}

func (f *respFixture) serve(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		command, err := readRESPCommand(reader)
		if err != nil {
			return
		}
		response := f.response(command)
		if _, err := writer.WriteString(response); err != nil {
			return
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}

func readRESPCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 3 || line[0] != '*' {
		return nil, fmt.Errorf("unexpected RESP array header %q", line)
	}
	count, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return nil, err
	}

	command := make([]string, count)
	for i := range command {
		header, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if len(header) < 3 || header[0] != '$' {
			return nil, fmt.Errorf("unexpected RESP bulk header %q", header)
		}
		length, err := strconv.Atoi(strings.TrimSpace(header[1:]))
		if err != nil {
			return nil, err
		}
		value := make([]byte, length+2)
		if _, err := io.ReadFull(reader, value); err != nil {
			return nil, err
		}
		command[i] = string(value[:length])
	}
	return command, nil
}

func (f *respFixture) response(command []string) string {
	name := strings.ToUpper(command[0])
	if name == "HELLO" {
		return "-ERR unknown command 'hello'\r\n"
	}

	f.mu.Lock()
	f.commands = append(f.commands, append([]string(nil), command...))
	failure := f.fail[name]
	f.mu.Unlock()
	if failure != "" {
		return "-ERR " + failure + "\r\n"
	}

	switch name {
	case "PING":
		return "+PONG\r\n"
	case "FT._LIST":
		return "*3\r\n$10\r\nweave:docs\r\n$13\r\nforeign:index\r\n$11\r\nweave:notes\r\n"
	case "FT.INFO":
		return "*4\r\n$8\r\nnum_docs\r\n:2\r\n$10\r\nindex_name\r\n$10\r\nweave:docs\r\n"
	case "FT.SEARCH":
		return "*3\r\n:1\r\n$16\r\nweave:docs:doc-1\r\n*8\r\n$7\r\ncontent\r\n$11\r\nhello world\r\n$8\r\nmetadata\r\n$16\r\n{\"kind\":\"guide\"}\r\n$6\r\ndoc_id\r\n$5\r\ndoc-1\r\n$14\r\n__vector_score\r\n$4\r\n0.25\r\n"
	case "HGETALL":
		return "*6\r\n$7\r\ncontent\r\n$11\r\nhello world\r\n$8\r\nmetadata\r\n$16\r\n{\"kind\":\"guide\"}\r\n$6\r\ndoc_id\r\n$5\r\ndoc-1\r\n"
	case "HSET", "DEL":
		return ":1\r\n"
	default:
		return "+OK\r\n"
	}
}

func (f *respFixture) find(name string) [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var found [][]string
	for _, command := range f.commands {
		if strings.EqualFold(command[0], name) {
			found = append(found, append([]string(nil), command...))
		}
	}
	return found
}

func TestClientProtocolOperations(t *testing.T) {
	fixture := &respFixture{}
	client := newProtocolClient(t, fixture)
	ctx := context.Background()

	if err := client.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if err := client.CreateCollection(ctx, "docs", 8, "L2"); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	names, err := client.ListCollections(ctx)
	if err != nil || len(names) != 2 || names[0] != "docs" || names[1] != "notes" {
		t.Fatalf("ListCollections() = %v, %v", names, err)
	}
	exists, err := client.CollectionExists(ctx, "docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists() = %v, %v", exists, err)
	}
	count, err := client.GetCollectionCount(ctx, "docs")
	if err != nil || count != 2 {
		t.Fatalf("GetCollectionCount() = %d, %v", count, err)
	}
	if err := client.DeleteCollection(ctx, "docs"); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}

	create := fixture.find("FT.CREATE")
	if len(create) != 1 || !strings.Contains(strings.Join(create[0], " "), "DIM 8 DISTANCE_METRIC L2") {
		t.Fatalf("FT.CREATE command = %v", create)
	}
	drop := fixture.find("FT.DROPINDEX")
	if len(drop) != 1 || drop[0][2] != "DD" {
		t.Fatalf("FT.DROPINDEX command = %v", drop)
	}
}

func TestClientDocumentAndSearchOperations(t *testing.T) {
	fixture := &respFixture{}
	client := newProtocolClient(t, fixture)
	ctx := context.Background()

	doc := &vectordb.Document{Text: "created", Metadata: map[string]interface{}{"kind": "guide"}, Embedding: []float64{1, 2, 3}}
	if err := client.CreateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}
	if doc.ID == "" {
		t.Fatal("CreateDocument() did not generate an ID")
	}
	batch := []*vectordb.Document{{ID: "a", Text: "one"}, {Text: "two"}}
	if err := client.CreateDocuments(ctx, "docs", batch); err != nil {
		t.Fatalf("CreateDocuments() error = %v", err)
	}
	if batch[1].ID == "" {
		t.Fatal("CreateDocuments() did not generate an ID")
	}

	got, err := client.GetDocument(ctx, "docs", "doc-1")
	if err != nil || got.ID != "doc-1" || got.Text != "hello world" || got.Metadata["kind"] != "guide" {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	if err := client.UpdateDocument(ctx, "docs", &vectordb.Document{ID: "doc-1", Text: "updated"}); err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	if err := client.DeleteDocument(ctx, "docs", "doc-1"); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if err := client.DeleteDocuments(ctx, "docs", []string{"a", "b"}); err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}

	docs, err := client.ListDocuments(ctx, "docs", 0, 4)
	if err != nil || len(docs) != 1 || docs[0].ID != "doc-1" {
		t.Fatalf("ListDocuments() = %#v, %v", docs, err)
	}
	semantic, err := client.SearchSemantic(ctx, "docs", []float32{1, 2, 3}, 0)
	if err != nil || len(semantic) != 1 || semantic[0].Score != 0.75 {
		t.Fatalf("SearchSemantic() = %#v, %v", semantic, err)
	}
	metadata, err := client.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, 0)
	if err != nil || len(metadata) != 1 || metadata[0].Score != 1 {
		t.Fatalf("SearchByMetadata() = %#v, %v", metadata, err)
	}
	fullText, err := client.SearchFullText(ctx, "docs", "hello.world", 0)
	if err != nil || len(fullText) != 1 || fullText[0].Document.Text != "hello world" {
		t.Fatalf("SearchFullText() = %#v, %v", fullText, err)
	}

	searches := fixture.find("FT.SEARCH")
	if len(searches) != 4 || searches[2][2] != "@metadata:(guide)" || searches[3][2] != "hello\\.world" {
		t.Fatalf("FT.SEARCH commands = %v", searches)
	}
}

func TestAdapterProtocolDelegation(t *testing.T) {
	fixture := &respFixture{}
	adapter := &Adapter{Client: newProtocolClient(t, fixture)}
	ctx := context.Background()

	if err := adapter.CreateCollection(ctx, "docs", nil); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) != 2 || collections[0].Count != 2 {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}
	if _, err := adapter.SearchBM25(ctx, "docs", "hello", &vectordb.QueryOptions{TopK: 3}); err != nil {
		t.Fatalf("SearchBM25() error = %v", err)
	}
	if _, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, nil); err != nil {
		t.Fatalf("SearchByMetadata() error = %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}); err != nil {
		t.Fatalf("DeleteDocumentsByMetadata() error = %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", nil); err == nil {
		t.Fatal("DeleteDocumentsByMetadata() expected empty-filter error")
	}
	if _, err := adapter.SearchHybrid(ctx, "docs", "hello", nil); err == nil {
		t.Fatal("SearchHybrid() expected unsupported error")
	}
	schema, err := adapter.GetSchema(ctx, "docs")
	if err != nil || schema.Class != "docs" || schema.Vectorizer != "text-embedding-3-small" {
		t.Fatalf("GetSchema() = %#v, %v", schema, err)
	}
	if err := adapter.UpdateSchema(ctx, "docs", schema); err == nil {
		t.Fatal("UpdateSchema() expected unsupported error")
	}
	if got := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "docs"); got.Class != "docs" {
		t.Fatalf("GetDefaultSchema() = %#v", got)
	}
	if err := adapter.ValidateSchema(schema); err != nil {
		t.Fatalf("ValidateSchema() error = %v", err)
	}
}

func TestClientProtocolErrors(t *testing.T) {
	tests := []struct {
		name string
		fail string
		call func(*Client) error
		want string
	}{
		{"health", "PING", func(c *Client) error { return c.Health(context.Background()) }, "health check failed"},
		{"create collection", "FT.CREATE", func(c *Client) error { return c.CreateCollection(context.Background(), "docs", 3, "COSINE") }, "failed to create index"},
		{"get document", "HGETALL", func(c *Client) error { _, err := c.GetDocument(context.Background(), "docs", "x"); return err }, "failed to get document"},
		{"semantic search", "FT.SEARCH", func(c *Client) error {
			_, err := c.SearchSemantic(context.Background(), "docs", []float32{1}, 1)
			return err
		}, "semantic search failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newProtocolClient(t, &respFixture{fail: map[string]string{tt.fail: "fixture failure"}})
			err := tt.call(client)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestRedisParsingEdgeCases(t *testing.T) {
	if got := extractNumDocs([]interface{}{"num_docs", "7"}); got != 7 {
		t.Fatalf("extractNumDocs(string) = %d", got)
	}
	if got := extractNumDocs([]interface{}{"num_docs", int64(8)}); got != 8 {
		t.Fatalf("extractNumDocs(int64) = %d", got)
	}
	if got := extractNumDocs("invalid"); got != 0 {
		t.Fatalf("extractNumDocs(invalid) = %d", got)
	}
	if got := parseSearchResults("invalid", "docs"); got != nil {
		t.Fatalf("parseSearchResults(invalid) = %#v", got)
	}
	if got := parseSearchResultsWithScore([]interface{}{}, "docs"); got != nil {
		t.Fatalf("parseSearchResultsWithScore(empty) = %#v", got)
	}
	if got := buildMetadataQuery(nil); got != "*" {
		t.Fatalf("buildMetadataQuery(nil) = %q", got)
	}
	doc := hashToDocument("fallback", map[string]string{"content": "text", "metadata": "not-json"})
	if doc.ID != "fallback" || doc.Metadata != nil {
		t.Fatalf("hashToDocument(invalid metadata) = %#v", doc)
	}
}
