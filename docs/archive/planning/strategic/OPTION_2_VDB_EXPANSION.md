# Option 2: VDB Expansion - Detailed Implementation Plan

**Status**: Planning
**Priority**: ⭐⭐ High (after Option 1)
**Total Effort**: 8-12 hours per VDB
**Target**: Week 3+

---

## Overview

Add support for additional vector databases to expand weave-cli's coverage beyond the current 10 VDBs. Each new VDB follows the established adapter pattern and requires implementation, testing, and documentation.

**Current VDBs (10):**
- ✅ Stable (7): Weaviate, Qdrant, Milvus, Chroma, Supabase, Neo4j, MongoDB
- 🟢 Beta (3): Pinecone, Elasticsearch, OpenSearch

**Candidate VDBs:**
1. **Vespa** (Recommended - Priority 1)
2. **Marqo** (Medium Priority)
3. **Nuclia** (Lower Priority)
4. **LanceDB** (Blocked by CGO)

---

## Candidate 1: Vespa (Recommended)

### Why Vespa?

**Strengths:**
- Full-featured search platform (vector + text + structured data)
- Native vector search with HNSW and brute-force algorithms
- Built-in BM25 and full-text search
- Native hybrid search support
- Real-time updates and deletes
- Horizontal scalability (production-ready)
- Official Go SDK (no CGO required)
- Both local and cloud deployment
- Strong documentation and community
- Enterprise adoption (Yahoo, Spotify, Verizon)

**Use Cases:**
- Advanced search beyond just vectors
- Structured + unstructured data hybrid queries
- High-scale production deployments
- Complex ranking and filtering requirements

### Technical Analysis

**SDK:**
```
github.com/vespa-engine/vespa/client/go/vespa
```

**Features:**
- REST API + Go client
- Document API for CRUD operations
- Query API with YQL (Vespa Query Language)
- Schema management via application packages
- Feeding API for bulk operations
- No CGO dependencies ✅

**API Examples:**

```go
// Creating a Vespa client
import "github.com/vespa-engine/vespa/client/go/vespa"

client, err := vespa.NewClient(vespa.Options{
    Target: "https://my-app.vespa-cloud.com",
    APIKey: os.Getenv("VESPA_API_KEY"),
})

// Document operations
doc := vespa.Document{
    ID: "id:namespace:doctype::1",
    Fields: map[string]interface{}{
        "text": "Machine learning best practices",
        "embedding": []float32{0.1, 0.2, 0.3, ...},
        "category": "tutorial",
    },
}

// Feed document
response, err := client.Feed(doc)

// Query with YQL
query := `
    select * from documents where
    ({targetHits:10}nearestNeighbor(embedding, query_embedding))
`
results, err := client.Query(query, vespa.QueryOptions{
    Ranking: "semantic",
    Hits: 10,
})
```

### Implementation Plan

#### Phase 1: Adapter Implementation (4 hours)

**Files to Create:**
```
src/pkg/vectordb/vespa/
├── adapter.go          # Main adapter implementing VectorDBClient
├── client.go           # Vespa client wrapper
├── collection.go       # Collection operations
├── document.go         # Document CRUD operations
├── query.go            # Search operations
├── schema.go           # Schema management
├── factory.go          # Factory registration
└── types.go            # Vespa-specific types

configs/
├── config.vespa-local.yaml
└── config.vespa-cloud.yaml
```

**Core Adapter Structure:**

```go
package vespa

import (
    "context"
    "fmt"

    "github.com/vespa-engine/vespa/client/go/vespa"
    "github.com/maximilien/weave-cli/src/pkg/vectordb"
    "github.com/maximilien/weave-cli/src/pkg/llm"
)

type Adapter struct {
    client    *vespa.Client
    config    *vectordb.Config
    llmClient *llm.OpenAIClient
}

func NewAdapter(config *vectordb.Config) (vectordb.VectorDBClient, error) {
    // Create Vespa client
    vespaClient, err := vespa.NewClient(vespa.Options{
        Target:  config.URL,
        APIKey:  config.APIKey,
        Timeout: config.Timeout,
    })
    if err != nil {
        return nil, fmt.Errorf("Vespa: failed to create client: %w", err)
    }

    // Create LLM client for embeddings
    llmClient, err := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
    if err != nil {
        // Optional - continue without LLM
    }

    return &Adapter{
        client:    vespaClient,
        config:    config,
        llmClient: llmClient,
    }, nil
}

func (a *Adapter) Health(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeHealth))
    defer cancel()

    // Ping Vespa service
    if err := a.client.Ping(ctx); err != nil {
        errMsg := err.Error()
        if strings.Contains(errMsg, "connection refused") {
            return fmt.Errorf("Vespa: health check failed: %w\n\nConnection refused. Common causes:\n"+
                "  1. Vespa not running (docker run vespaengine/vespa)\n"+
                "  2. Wrong URL in configuration\n"+
                "  3. For Vespa Cloud: verify API key\n"+
                "  → Check connection at https://cloud.vespa.ai", err)
        }
        return fmt.Errorf("Vespa: health check failed: %w", err)
    }

    return nil
}

func (a *Adapter) CreateDocument(ctx context.Context, collection string, doc *vectordb.Document) error {
    ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeDocument))
    defer cancel()

    // Generate embedding if needed
    if len(doc.Vector) == 0 && doc.Text != "" {
        embedding, err := a.llmClient.GenerateEmbedding(ctx, doc.Text, "")
        if err != nil {
            return fmt.Errorf("Vespa: failed to generate embedding: %w", err)
        }
        doc.Vector = embedding
    }

    // Convert to Vespa document
    vespaDoc := vespa.Document{
        ID: fmt.Sprintf("id:weave:%s::%s", collection, doc.ID),
        Fields: map[string]interface{}{
            "text":      doc.Text,
            "embedding": doc.Vector,
            "metadata":  doc.Metadata,
        },
    }

    // Feed to Vespa
    if err := a.client.Feed(ctx, vespaDoc); err != nil {
        return fmt.Errorf("Vespa: failed to create document: %w", err)
    }

    return nil
}

func (a *Adapter) SearchSemantic(ctx context.Context, collection string, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
    ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
    defer cancel()

    // Generate query embedding
    embedding, err := a.llmClient.GenerateEmbedding(ctx, query, "")
    if err != nil {
        return nil, fmt.Errorf("Vespa: failed to generate query embedding: %w", err)
    }

    topK := 10
    if opts != nil && opts.TopK > 0 {
        topK = opts.TopK
    }

    // Build YQL query for vector search
    yql := fmt.Sprintf(`
        select * from %s where
        ({targetHits:%d}nearestNeighbor(embedding, query_embedding))
    `, collection, topK)

    // Execute query
    results, err := a.client.Query(ctx, yql, vespa.QueryOptions{
        Ranking: "semantic",
        Hits:    topK,
        Tensors: map[string][]float32{
            "query_embedding": embedding,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("Vespa: semantic search failed: %w", err)
    }

    // Convert results
    return a.convertResults(results), nil
}
```

**Key Implementation Details:**

1. **Collection Management:**
   - Vespa uses "document types" instead of collections
   - Schema defined via application packages
   - Need to handle schema deployment

2. **Document IDs:**
   - Vespa uses structured IDs: `id:namespace:doctype::userkey`
   - Map collection name to document type

3. **Search:**
   - YQL (Vespa Query Language) for queries
   - nearestNeighbor operator for vector search
   - Ranking profiles for scoring

4. **Hybrid Search:**
   - Native support via YQL
   - Combine nearestNeighbor + text matching
   - Custom ranking profiles

#### Phase 2: Integration Tests (3 hours)

**Test Suite Structure:**

```go
// tests/vespa_integration_test.go
package tests

import (
    "context"
    "testing"

    "github.com/maximilien/weave-cli/src/pkg/vectordb"
    "github.com/maximilien/weave-cli/src/pkg/vectordb/vespa"
)

func TestVespaIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Vespa integration tests in short mode")
    }

    // Check if Vespa is available
    config := loadVespaConfig(t)
    client, err := vespa.NewAdapter(config)
    if err != nil {
        t.Skipf("Vespa not available: %v", err)
    }
    defer client.Close()

    ctx := context.Background()

    t.Run("Health", func(t *testing.T) {
        testVespaHealth(t, ctx, client)
    })

    t.Run("CollectionOperations", func(t *testing.T) {
        testVespaCollectionOperations(t, ctx, client)
    })

    t.Run("DocumentCRUD", func(t *testing.T) {
        testVespaDocumentCRUD(t, ctx, client)
    })

    t.Run("BatchOperations", func(t *testing.T) {
        testVespaBatchOperations(t, ctx, client)
    })

    t.Run("SemanticSearch", func(t *testing.T) {
        testVespaSemanticSearch(t, ctx, client)
    })

    t.Run("HybridSearch", func(t *testing.T) {
        testVespaHybridSearch(t, ctx, client)
    })

    t.Run("BM25Search", func(t *testing.T) {
        testVespaBM25Search(t, ctx, client)
    })

    t.Run("MetadataFiltering", func(t *testing.T) {
        testVespaMetadataFiltering(t, ctx, client)
    })

    // Add more tests...
}
```

**Test Cases (16 tests minimum):**
1. Health check
2. Create collection
3. List collections
4. Delete collection
5. Collection exists
6. Get collection count
7. Create document
8. Get document
9. Update document
10. Delete document
11. List documents
12. Batch create documents
13. Semantic search
14. BM25 search
15. Hybrid search
16. Metadata filtering

#### Phase 3: Documentation (3 hours)

**Files to Create:**

```
docs/vespa/
├── README.md              # Overview and features
├── SETUP.md               # General setup guide
├── LOCAL_SETUP.md         # Docker setup for local Vespa
├── CLOUD_SETUP.md         # Vespa Cloud setup
└── ADVANCED.md            # YQL, ranking profiles, schemas
```

**README.md Structure:**

```markdown
# Vespa Vector Database Support

Vespa is a full-featured search platform that combines vector search,
full-text search, and structured data queries in a single system.

## Features

✅ **Vector Search** - HNSW and brute-force algorithms
✅ **Full-Text Search** - BM25 ranking
✅ **Hybrid Search** - Combined vector + text
✅ **Structured Queries** - Filter by fields
✅ **Real-Time Updates** - Immediate indexing
✅ **Horizontal Scaling** - Production-ready

## Quick Start

### Local Deployment (Docker)
```bash
docker run -d -p 8080:8080 vespaengine/vespa
```

### Cloud Deployment (Vespa Cloud)
1. Sign up at https://cloud.vespa.ai
2. Create application
3. Get API credentials

### Configure weave-cli
```bash
export VESPA_URL="https://my-app.vespa-cloud.com"
export VESPA_API_KEY="your-api-key"
export VECTOR_DB_TYPE="vespa-cloud"
```

### Create Collection & Search
```bash
weave cols create my_docs --vespa-cloud
weave docs create my_docs document.txt --vespa-cloud
weave search my_docs "machine learning" --vespa-cloud
```

## Advanced Features

- **YQL Queries** - Powerful query language
- **Ranking Profiles** - Custom scoring functions
- **Schema Management** - Flexible document schemas
- **Streaming Search** - Process data in streams

## Resources

- [Vespa Documentation](https://docs.vespa.ai)
- [Vespa Cloud](https://cloud.vespa.ai)
- [Sample Applications](https://github.com/vespa-engine/sample-apps)
```

**LOCAL_SETUP.md:**

Detailed Docker setup instructions, application package creation, feeding data, querying.

**CLOUD_SETUP.md:**

Vespa Cloud account setup, application deployment, API credentials, monitoring.

#### Phase 4: VDB Support Matrix Update (0.5 hour)

Update `docs/VDB_SUPPORT_MATRIX.md`:

```markdown
| Vespa | Local + Cloud | 🟢 Beta | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | [SETUP](vespa/SETUP.md) |
```

### Effort Summary

| Phase | Task | Time |
|-------|------|------|
| 1 | Adapter Implementation | 4 hours |
| 2 | Integration Tests | 3 hours |
| 3 | Documentation | 3 hours |
| 4 | Support Matrix Update | 0.5 hour |
| **Total** | **Vespa Integration** | **10-12 hours** |

### Success Criteria

- ✅ Vespa adapter passes all 16 integration tests
- ✅ Health check works for local and cloud
- ✅ Semantic, BM25, and hybrid search work
- ✅ Documentation covers setup and advanced usage
- ✅ CI pipeline includes Vespa tests

---

## Candidate 2: Marqo

### Why Marqo?

**Strengths:**
- Multi-modal search (text + images)
- Built-in embedding models (no separate LLM API needed)
- Automatic batching and optimization
- Simple REST API
- Docker deployment

**Weaknesses:**
- No official Go SDK (need HTTP client wrapper)
- Less mature than others
- Smaller community

### Technical Analysis

**API:**
- REST API only
- Need to implement HTTP client wrapper

**Estimated Effort:** 8-10 hours
- HTTP client wrapper: 2 hours
- Adapter implementation: 3 hours
- Integration tests: 2 hours
- Documentation: 2 hours

**Priority:** Medium (defer until multi-modal search is requested)

---

## Candidate 3: Nuclia

### Why Nuclia?

**Strengths:**
- Knowledge graph + vector search hybrid
- Automatic content extraction (PDFs, images, audio)
- Multi-language support

**Weaknesses:**
- Cloud-only (no local deployment)
- REST API only (no Go SDK)
- Niche use case

**Estimated Effort:** 10-12 hours

**Priority:** Low (defer until knowledge graph feature is requested)

---

## Candidate 4: LanceDB

### Status: BLOCKED

**Issue:** CGO dependency in official SDK
**See:** `docs/lancedb/RESEARCH.md`
**Decision:** Defer until CGO-free SDK available or accept CGO for specific platforms

---

## Implementation Order

**Recommendation:**

1. **Week 3:** Vespa (10-12 hours) - Highest value, official SDK
2. **Week 4+:** Marqo or Nuclia (on demand) - Based on user requests

**Rationale:**
- Vespa provides most comprehensive feature set
- Official Go SDK ensures maintainability
- Both local and cloud deployment
- Enterprise-grade scaling
- Strong documentation

---

## VDB Expansion Checklist

For each new VDB, follow this checklist:

### Pre-Implementation
- [ ] Research SDK availability and quality
- [ ] Check CGO dependencies
- [ ] Review API documentation
- [ ] Test local/cloud deployment options
- [ ] Identify unique features

### Implementation
- [ ] Create adapter package (`src/pkg/vectordb/{vdb}/`)
- [ ] Implement VectorDBClient interface
- [ ] Add factory registration
- [ ] Create config files (local + cloud)
- [ ] Add timeout handling
- [ ] Add error messages with VDB prefix
- [ ] Implement troubleshooting hints

### Testing
- [ ] Create integration test file (`tests/{vdb}_integration_test.go`)
- [ ] Implement 16 standard tests
- [ ] Test with local deployment
- [ ] Test with cloud deployment (if applicable)
- [ ] Verify batch operations
- [ ] Test all search types

### Documentation
- [ ] Create `docs/{vdb}/README.md`
- [ ] Create `docs/{vdb}/SETUP.md`
- [ ] Create `docs/{vdb}/LOCAL_SETUP.md`
- [ ] Create `docs/{vdb}/CLOUD_SETUP.md` (if cloud available)
- [ ] Update `docs/VDB_SUPPORT_MATRIX.md`
- [ ] Add to README.md main list
- [ ] Create example configurations

### Finalization
- [ ] Run full test suite
- [ ] Update CHANGELOG.md
- [ ] Create PR with all changes
- [ ] Update version number
- [ ] Tag release

---

## Total Timeline

**Week 3: Vespa Integration**
- Days 1-2: Adapter implementation and client wrapper
- Day 3: Integration tests
- Days 4-5: Documentation and examples

**Future Weeks: Additional VDBs (On Demand)**
- Follow same pattern: 8-12 hours per VDB
- Prioritize based on user requests
