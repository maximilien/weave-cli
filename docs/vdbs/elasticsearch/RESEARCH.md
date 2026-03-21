# Elasticsearch Integration Research

**Date**: 2025-12-12
**Status**: Research Complete - Ready to Implement
**Decision**: Proceed with integration for v0.8.0

---

## Executive Summary

Elasticsearch has a mature, official Go SDK (`github.com/elastic/go-elasticsearch` v9.2.1) with **pure Go implementation (no CGO)**. It provides comprehensive vector search capabilities via `dense_vector` fields and kNN search with HNSW indexing. Recommendation is to proceed with integration as VDB #10.

---

## Research Findings

### Official Go SDK Status

**Repository**: `github.com/elastic/go-elasticsearch`
- **Latest Version**: v9.2.1
- **Stability**: Production-ready, follows Go language policy
- **Maturity**: v8.x and v9.x both actively supported
- **License**: Apache-2.0
- **Maintainer**: Elastic (official)
- **Activity**: Actively maintained, regular updates

### Features Available

✅ **Fully Supported Features**:
- Vector similarity search (`dense_vector` field type)
- Hybrid search (BM25 + kNN combined)
- BM25 full-text search (native)
- Batch operations (BulkIndexer)
- Metadata filtering (comprehensive query DSL)
- Collection/index management
- Schema management (index mappings)
- Auto-indexing (HNSW for dims >= 384)

### Technical Requirements

✅ **Pure Go - NO CGO Required**:
```bash
# Installation (simple Go modules)
go get github.com/elastic/go-elasticsearch/v9
```

**Dependencies**:
- Go 1.11+ with module support
- `elastictransport` (official, pure Go)

**Supported Platforms**:
- ✅ macOS (Intel + ARM)
- ✅ Linux (Intel + ARM)
- ✅ Windows (Intel + ARM)
- ✅ All Go-supported platforms

### API Structure

**Two Client Types**:
1. **TypedClient** - Recommended for new code
   - Type-safe API
   - Full IDE autocomplete support
   - Compile-time checking

2. **Client** (Functional Options API)
   - Flexible, lower-level access
   - Dynamic queries

### Vector Search Implementation

**Index Configuration**:
```go
// Create index with dense_vector field
PUT /my-vector-index
{
  "mappings": {
    "properties": {
      "vector_field": {
        "type": "dense_vector",
        "dims": 1536,
        "index": true,
        "similarity": "cosine"
      },
      "text": { "type": "text" },
      "metadata": { "type": "object" }
    }
  }
}
```

**kNN Search Pattern**:
```go
import (
    "github.com/elastic/go-elasticsearch/v9"
    "github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

// Client initialization
client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
    CloudID: cloudID,  // For Elastic Cloud
    APIKey:  apiKey,   // Or username/password
})

// kNN vector search
res, err := client.Search().
    Index("my-vector-index").
    Knn(types.KnnSearch{
        Field:         "vector_field",
        K:             &k,              // Number of results
        NumCandidates: &numCandidates,  // Search space
        QueryVector:   queryVector,     // Float64 slice
    }).Do(ctx)
```

**Hybrid Search (BM25 + kNN)**:
```go
res, err := client.Search().
    Index("my-vector-index").
    Query(&types.Query{
        Match: map[string]types.MatchQuery{
            "text": {Query: searchText},
        },
    }).
    Knn(types.KnnSearch{
        Field:       "vector_field",
        QueryVector: queryVector,
        K:           &k,
    }).Do(ctx)
```

### HNSW Indexing

**Automatic Indexing**:
- Vectors with `dims >= 384`: Indexed as `bbq_hnsw` (quantized)
- Vectors with `dims < 384`: Indexed as `int8_hnsw`
- No manual configuration required for basic use

**Customization Available**:
- Index parameters (m, ef_construction)
- Similarity metrics (cosine, dot_product, l2_norm)
- Quantization options

---

## Decision Analysis

### Pros of Adding Elasticsearch

1. **✅ Pure Go** - No CGO, perfect cross-platform support
2. **Enterprise-Ready** - Massive install base, battle-tested
3. **Excellent Hybrid Search** - Native BM25 + vector search
4. **Mature SDK** - v9.2.1, stable API, comprehensive docs
5. **Full Feature Set** - All VectorDBClient interface methods supported
6. **Cloud + Local** - Supports both Elastic Cloud and self-hosted
7. **Strong Ecosystem** - Kibana, enterprise features, integrations
8. **Active Development** - Regular updates, long-term support
9. **Portfolio Balance** - Maintains high pure-Go ratio (9/10 = 90%)

### Cons of Adding Elasticsearch

1. **Resource Requirements** - Needs more RAM/CPU than embedded DBs
2. **Complexity** - Full search engine, may be overkill for simple use cases
3. **Setup** - Requires running Elasticsearch server (not embedded)
4. **Cost** - Elastic Cloud pricing for managed service

### Impact on Project

**Current State**:
- 9 VDBs supported
- 1 with CGO requirement (Chroma - macOS only)
- 8 pure Go (cross-platform friendly)

**After Elasticsearch**:
- 10 VDBs supported
- 1 with CGO requirement (Chroma)
- 9 pure Go
- **90% pure Go ratio** ✅

---

## Recommendation

### ✅ Proceed with Elasticsearch Integration

**Rationale**:
1. **Pure Go** - No build complications, excellent cross-platform support
2. **Widely Deployed** - Large existing install base, enterprise adoption
3. **Feature Complete** - All required capabilities (vector, hybrid, BM25, batch)
4. **Production Ready** - Mature SDK, stable API, comprehensive testing
5. **Strategic Value** - Fills enterprise/search-focused use case

### Target Version

**v0.8.0** - Next feature release

### Estimated Effort

**Total: 4-5 days**
1. VectorDBClient implementation - 2 days
2. Collection/document operations - 1 day
3. Integration tests - 1 day
4. Documentation - 1 day

---

## Implementation Plan

### Phase 1: Core Client (Day 1-2)

**Files to Create**:
- `src/pkg/vectordb/elasticsearch/client.go` - VectorDBClient interface
- `src/pkg/vectordb/elasticsearch/config.go` - Configuration
- `src/pkg/vectordb/elasticsearch/factory.go` - Factory registration

**Key Methods**:
```go
type ElasticsearchClient struct {
    client *elasticsearch.TypedClient
    config *Config
}

func (c *ElasticsearchClient) Connect(ctx context.Context) error
func (c *ElasticsearchClient) Ping(ctx context.Context) error
func (c *ElasticsearchClient) Close() error
```

### Phase 2: Collection Operations (Day 2)

**Files to Create**:
- `src/pkg/vectordb/elasticsearch/collection.go`

**Methods**:
- `CreateCollection()` - Create index with mappings
- `DeleteCollection()` - Delete index
- `ListCollections()` - List indices
- `GetCollection()` - Get index info

### Phase 3: Document Operations (Day 3)

**Files to Create**:
- `src/pkg/vectordb/elasticsearch/document.go`
- `src/pkg/vectordb/elasticsearch/query.go`

**Methods**:
- `AddDocument()` - Index single document
- `AddDocuments()` - Bulk indexing
- `UpdateDocument()` - Update by ID
- `DeleteDocument()` - Delete by ID
- `GetDocument()` - Fetch by ID

### Phase 4: Search Operations (Day 3)

**Methods**:
- `VectorSearch()` - kNN search
- `HybridSearch()` - BM25 + kNN
- `BM25Search()` - Full-text search
- `MetadataSearch()` - Filter queries

### Phase 5: Testing (Day 4)

**Files to Create**:
- `tests/elasticsearch_integration_test.go`
- `docs/vdbs/elasticsearch/SETUP.md`
- `docs/vdbs/elasticsearch/LOCAL_SETUP.md`
- `docs/vdbs/elasticsearch/CLOUD_SETUP.md`

**Test Coverage**:
- Health checks
- Collection operations
- Document CRUD
- Batch operations
- Vector search
- Hybrid search
- BM25 search
- Metadata filtering

### Phase 6: Documentation (Day 5)

**Files to Create/Update**:
- `docs/vdbs/elasticsearch/README.md` - Overview
- `docs/VDB_SUPPORT_MATRIX.md` - Add Elasticsearch row
- `README.md` - Add Elasticsearch to main table
- `CHANGELOG.md` - v0.8.0 entry

---

## Configuration Requirements

**Environment Variables**:
```bash
# Elastic Cloud
ELASTICSEARCH_CLOUD_ID=cluster-name:base64-encoded-id
ELASTICSEARCH_API_KEY=base64-encoded-key

# Self-hosted
ELASTICSEARCH_ADDRESSES=http://localhost:9200,http://localhost:9201
ELASTICSEARCH_USERNAME=elastic
ELASTICSEARCH_PASSWORD=changeme

# Optional
ELASTICSEARCH_CERT_FINGERPRINT=sha256-fingerprint
```

**Config Structure**:
```go
type Config struct {
    CloudID          string
    APIKey           string
    Addresses        []string
    Username         string
    Password         string
    CertFingerprint  string
    MaxRetries       int
    EnableDebugLog   bool
}
```

---

## VectorDBClient Interface Mapping

| Interface Method | Elasticsearch API | Notes |
|-----------------|-------------------|-------|
| `Connect()` | `NewTypedClient()` | CloudID or addresses |
| `Ping()` | `client.Ping()` | Health check |
| `CreateCollection()` | `client.Indices().Create()` | With mappings |
| `DeleteCollection()` | `client.Indices().Delete()` | By name |
| `ListCollections()` | `client.Indices().Get()` | Filter by prefix |
| `AddDocument()` | `client.Index()` | Single doc |
| `AddDocuments()` | `esutil.BulkIndexer` | Batch insert |
| `GetDocument()` | `client.Get()` | By ID |
| `UpdateDocument()` | `client.Update()` | By ID |
| `DeleteDocument()` | `client.Delete()` | By ID |
| `VectorSearch()` | `client.Search().Knn()` | kNN query |
| `HybridSearch()` | `client.Search().Query().Knn()` | Combined |
| `BM25Search()` | `client.Search().Query(Match)` | Full-text |
| `MetadataSearch()` | `client.Search().Query(Bool)` | Filters |

**Coverage**: 100% of VectorDBClient interface ✅

---

## Testing Strategy

### Local Testing
1. Run Elasticsearch via Docker
2. Create test indices with vector fields
3. Run integration test suite

### Cloud Testing
1. Use Elastic Cloud free trial (14 days)
2. Configure with CloudID and API key
3. Validate all operations

### CI/CD
1. Add Elasticsearch to GitHub Actions
2. Use Docker service container
3. Run integration tests on PR

---

## Documentation Links

- **Elasticsearch Go SDK**: https://github.com/elastic/go-elasticsearch
- **Go Pkg Docs**: https://pkg.go.dev/github.com/elastic/go-elasticsearch/v9
- **Vector Search Guide**: https://www.elastic.co/search-labs/blog/perform-vector-search-with-the-elasticsearch-go-client
- **Dense Vector Docs**: https://www.elastic.co/docs/reference/elasticsearch/mapping-reference/dense-vector
- **kNN Search Docs**: https://www.elastic.co/docs/solutions/search/vector/knn
- **Official Docs**: https://www.elastic.co/docs/reference/elasticsearch/clients/go

---

## Next Steps

1. ✅ Document Elasticsearch research (this file)
2. ➡️ **Create integration architecture plan**
3. ➡️ **Implement VectorDBClient interface**
4. ➡️ **Add integration tests**
5. ➡️ **Document setup guides**

**Priority**: High - Target v0.8.0 release (4-5 day effort)

---

**Research Completed**: 2025-12-12
**Decision**: Proceed with implementation
**Next Review**: After v0.8.0 release for user feedback
