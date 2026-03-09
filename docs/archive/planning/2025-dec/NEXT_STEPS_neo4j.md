# Neo4j Integration Plan - v0.8.0

## Overview

**Goal**: Add Neo4j as a vector database option with Graph + Vector hybrid capabilities
**Timeline**: 5-7 days
**Target Version**: v0.8.0
**Type**: Minor release - New VDB integration

## Neo4j Capabilities

### Vector Search Features
- **Vector Indexes**: Fast approximate k-NN search using HNSW algorithm
- **Similarity Functions**: Cosine and Euclidean distance
- **Dimensions**: 1-4096 (flexible vector sizes)
- **Quantization**: Optional size reduction for performance
- **Graph Context**: Combine vector similarity with graph relationships

### Unique Advantages
- **Graph + Vector Hybrid**: Query both structure and semantics
- **GraphRAG Support**: Perfect for knowledge graph-enhanced RAG
- **Relationship Context**: Vector similarity + graph traversal
- **Cypher Queries**: Powerful query language for complex patterns

## Technical Stack

### Go SDK
- **Package**: `github.com/neo4j/neo4j-go-driver/v5`
- **Protocol**: Bolt (binary protocol)
- **Versions**: Supports Neo4j 4.4, 5.x, and 2025.x
- **Query Method**: Cypher queries via `ExecuteQuery()` or sessions

### Vector Operations via Cypher
- Create index: `CREATE VECTOR INDEX index_name FOR (n:Label) ON (n.property)`
- Query: `CALL db.index.vector.queryNodes(index, k, vector) YIELD node, score`
- Functions: `vector.similarity.cosine()`, `vector.similarity.euclidean()`

## Implementation Plan

### Phase 1: Core Infrastructure (Day 1-2)

**1.1 Package Structure**
```
src/pkg/vectordb/neo4j/
├── client.go         # Neo4j driver wrapper
├── config.go         # Configuration types
├── adapter.go        # VectorDB interface implementation
├── collection.go     # Collection/Index management
├── document.go       # Document CRUD operations
├── query.go          # Vector search queries
└── graph.go          # Graph-specific operations (future)
```

**1.2 Configuration**
```yaml
neo4j-local:
  type: neo4j-local
  uri: bolt://localhost:7687
  username: neo4j
  password: ${NEO4J_PASSWORD}
  database: neo4j

neo4j-cloud:
  type: neo4j-cloud
  uri: ${NEO4J_URI}
  username: ${NEO4J_USERNAME}
  password: ${NEO4J_PASSWORD}
  database: ${NEO4J_DATABASE}
```

**1.3 VectorDB Types**
```go
// Add to src/pkg/config/config.go
VectorDBTypeNeo4jLocal VectorDBType = "neo4j-local"
VectorDBTypeNeo4jCloud VectorDBType = "neo4j-cloud"
```

### Phase 2: Client Implementation (Day 2-3)

**2.1 Client Structure**
```go
type Client struct {
    driver   neo4j.DriverWithContext
    database string
    config   *Config
}

type Config struct {
    URI      string
    Username string
    Password string
    Database string
    MaxConnections int
    Timeout  time.Duration
}
```

**2.2 Core Methods**
- `NewClient(config *Config) (*Client, error)`
- `Close() error`
- `Health(ctx context.Context) error`
- `ExecuteQuery(ctx context.Context, query string, params map[string]any) (neo4j.ResultWithContext, error)`

**2.3 Vector Index Operations**
```go
// Create vector index for a label property
func (c *Client) CreateVectorIndex(ctx context.Context, indexName, label, property string, dimensions int, similarity string) error

// Query vector index
func (c *Client) QueryVectorIndex(ctx context.Context, indexName string, vector []float32, topK int) ([]*Document, error)
```

### Phase 3: Collection Operations (Day 3-4)

**3.1 Collection Mapping**
Neo4j uses labels + vector indexes instead of collections:
- Collection → Node Label + Vector Index
- Example: `WeaveDocs` collection → `WeaveDocs` label + `WeaveDocs_vector_idx` index

**3.2 Operations**
```go
// CreateCollection creates a label and vector index
func (c *Client) CreateCollection(ctx context.Context, name string, dimensions int) error

// ListCollections returns all vector indexes
func (c *Client) ListCollections(ctx context.Context) ([]*CollectionInfo, error)

// DeleteCollection removes vector index and optionally nodes
func (c *Client) DeleteCollection(ctx context.Context, name string) error

// CollectionExists checks if vector index exists
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error)

// GetCollectionCount returns number of nodes with label
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error)
```

### Phase 4: Document Operations (Day 4-5)

**4.1 Document Structure**
```go
type Document struct {
    ID       string                 // Node ID or property
    Content  string                 // Text content
    Vector   []float32              // Embedding vector
    Metadata map[string]interface{} // Properties
}
```

**4.2 Document Node Pattern**
```cypher
// Create document node with vector
CREATE (d:WeaveDocs {
  id: $id,
  content: $content,
  embedding: $vector,
  filename: $filename,
  date_added: datetime()
})
```

**4.3 CRUD Operations**
```go
// CreateDocument creates node with vector
func (c *Client) CreateDocument(ctx context.Context, collectionName string, doc *Document) error

// GetDocument retrieves node by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, id string) (*Document, error)

// UpdateDocument updates node properties
func (c *Client) UpdateDocument(ctx context.Context, collectionName string, doc *Document) error

// DeleteDocument removes node
func (c *Client) DeleteDocument(ctx context.Context, collectionName, id string) error

// ListDocuments returns all nodes with label
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit int) ([]*Document, error)

// DeleteAllDocuments removes all nodes with label
func (c *Client) DeleteAllDocuments(ctx context.Context, collectionName string) error
```

### Phase 5: Vector Search (Day 5)

**5.1 Search Implementation**
```cypher
// Vector similarity search
CALL db.index.vector.queryNodes($indexName, $topK, $vector)
YIELD node, score
WHERE node.filename IS NOT NULL
RETURN node.id AS id,
       node.content AS content,
       score,
       properties(node) AS metadata
ORDER BY score DESC
LIMIT $topK
```

**5.2 Search Methods**
```go
// SearchDocuments performs vector similarity search
func (c *Client) SearchDocuments(ctx context.Context, collectionName string, vector []float32, topK int) ([]*SearchResult, error)

// SearchWithFilter combines vector search with property filters
func (c *Client) SearchWithFilter(ctx context.Context, collectionName string, vector []float32, topK int, filter map[string]interface{}) ([]*SearchResult, error)
```

### Phase 6: Testing (Day 6)

**6.1 Integration Tests**
```
tests/neo4j_integration_test.go
- TestNeo4jIntegration/Health
- TestNeo4jIntegration/CreateCollection
- TestNeo4jIntegration/CollectionExists
- TestNeo4jIntegration/ListCollections
- TestNeo4jIntegration/CreateDocument
- TestNeo4jIntegration/GetDocument
- TestNeo4jIntegration/UpdateDocument
- TestNeo4jIntegration/BatchCreateDocuments
- TestNeo4jIntegration/SearchDocuments
- TestNeo4jIntegration/GetCollectionCount
- TestNeo4jIntegration/ListDocuments
- TestNeo4jIntegration/DeleteDocument
- TestNeo4jIntegration/DeleteAllDocuments
- TestNeo4jIntegration/DeleteCollection
```

**6.2 Test Setup**
- Docker container: `neo4j:5-community` or `neo4j:5-enterprise`
- Ports: 7474 (HTTP), 7687 (Bolt)
- Environment: NEO4J_AUTH=neo4j/testpassword

### Phase 7: CLI Integration (Day 6-7)

**7.1 Management Scripts**
```bash
tools/vdb/local/neo4j.sh
- start: Start Neo4j container
- stop: Stop container
- status: Check status
- logs: View logs
- clean: Remove container and data
```

**7.2 Health Checks**
```bash
tools/vdb/health.sh
- Add Neo4j local check
- Add Neo4j cloud check
```

**7.3 CLI Flags**
Add `--neo4j-local` and `--neo4j-cloud` flags to:
- All collection commands
- All document commands
- Query/search commands
- Health check command

**7.4 Utils Functions**
```go
// src/cmd/utils/
- CreateNeo4jClient()
- ListNeo4jCollections()
- ShowNeo4jCollection()
- CreateNeo4jDocument()
- QueryNeo4jCollection()
```

### Phase 8: Documentation (Day 7)

**8.1 Release Notes**
```
docs/releases/RELEASE_v0.8.0.md
- Neo4j integration overview
- Graph + Vector hybrid features
- Migration guide
- Examples and use cases
```

**8.2 README Updates**
- Add Neo4j to supported databases
- Add Neo4j examples
- Document graph query capabilities

## Technical Considerations

### Vector Index Creation
```cypher
CREATE VECTOR INDEX collection_name_vector_idx
FOR (n:CollectionName)
ON (n.embedding)
OPTIONS {
  indexConfig: {
    `vector.dimensions`: 1536,
    `vector.similarity_function`: 'cosine'
  }
}
```

### Metadata Handling
- Store metadata as node properties
- All properties are indexed by default
- Use property filters in vector search queries

### ID Management
- Use node internal ID or custom ID property
- Store original ID in `id` property
- Return original ID in all operations

### Embedding Generation
- Same OpenAI integration as other VDBs
- Auto-generate embeddings for text content
- Support custom embedding models

### Error Handling
- Driver connection errors
- Cypher syntax errors
- Constraint violations
- Vector dimension mismatches

## Estimated Timeline

### Day 1: Research & Setup
- ✅ Research Neo4j vector capabilities
- ✅ Research Go SDK
- ⏳ Set up Neo4j local instance
- ⏳ Test basic Cypher queries

### Day 2: Core Client
- Create package structure
- Implement client.go
- Implement config.go
- Basic connection and health check

### Day 3: Collections
- Implement collection.go
- Vector index CRUD operations
- Label management
- Collection info retrieval

### Day 4: Documents
- Implement document.go
- Node CRUD operations
- Metadata handling
- Batch operations

### Day 5: Vector Search
- Implement query.go
- Vector similarity search
- Filtered search
- Result ranking

### Day 6: Testing & Integration
- Write integration tests
- Test all operations
- CLI flag wiring
- Utils functions

### Day 7: Documentation & Release
- Create management scripts
- Write release notes
- Update README
- Tag v0.8.0

## Success Criteria

- [ ] All VectorDB interface methods implemented
- [ ] 14+ integration tests passing
- [ ] CLI commands work with --neo4j-local/--neo4j-cloud
- [ ] Management scripts functional
- [ ] Health checks working
- [ ] Documentation complete
- [ ] Build and lint passing

## Future Enhancements (Post v0.8.0)

### Graph Queries (v0.8.1)
- Relationship-aware search
- Graph traversal queries
- Path-based retrieval
- Community detection

### GraphRAG (v0.8.2)
- Knowledge graph construction
- Entity relationship extraction
- Subgraph retrieval
- Hybrid graph + vector queries

### Advanced Features (v0.8.3+)
- Full-text search integration
- Property graph projections
- Graph algorithms (PageRank, Centrality)
- Multi-hop reasoning

## Dependencies

```go
// go.mod additions
require (
    github.com/neo4j/neo4j-go-driver/v5 v5.x.x
)
```

## References

- Neo4j Go Driver: https://github.com/neo4j/neo4j-go-driver
- Vector Search Docs: https://neo4j.com/docs/cypher-manual/current/indexes/semantic-indexes/vector-indexes/
- Go Manual: https://neo4j.com/docs/go-manual/current/
- Vector Functions: https://neo4j.com/docs/cypher-manual/current/functions/vector/
