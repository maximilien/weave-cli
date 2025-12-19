# Weave CLI Architecture

**Last Updated**: 2025-12-19
**Version**: v0.8.2

## Overview

Weave CLI is a command-line tool for managing vector databases with RAG (Retrieval-Augmented Generation) capabilities. It provides a unified interface across 10 different vector database systems with AI agent-based query execution.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  (Cobra Commands: vector-db, health, collections, etc.)     │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────────┐
│                    Agent Layer                               │
│  • WeaveAgent (orchestration)                               │
│  • QueryAgent, PlanningAgent, OutputAgent, etc.             │
│  • LLM Integration (OpenAI, Claude)                         │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────────┐
│              Vector DB Abstraction Layer                     │
│                VectorDBClient Interface                      │
└──────────────────┬──────────────────────────────────────────┘
                   │
      ┌────────────┴────────────┬──────────────┬──────────────┐
      │                         │              │              │
┌─────▼─────┐ ┌────────▼───────┐ ┌────▼──────┐ ┌────▼──────┐
│ Weaviate  │ │    Qdrant      │ │  Milvus   │ │   ...     │
│  Adapter  │ │    Adapter     │ │  Adapter  │ │ (10 VDBs) │
└───────────┘ └────────────────┘ └───────────┘ └───────────┘
```

## Core Components

### 1. CLI Layer (`src/cmd/`)

**Purpose**: User interface and command orchestration

**Structure**:
```
src/cmd/
├── root.go              # Root command and global flags
├── vectordb/           # VDB management commands
│   ├── health.go       # Health check command
│   ├── collections.go  # Collection CRUD
│   ├── documents.go    # Document operations
│   └── search.go       # Search operations
├── config/             # Configuration commands
└── version/            # Version command
```

**Technology**: Cobra CLI framework

### 2. Agent Layer (`src/pkg/agents/`)

**Purpose**: AI-powered query planning and execution

**Key Agents**:
- **WeaveAgent**: Main orchestrator, coordinates other agents
- **QueryAgent**: Translates natural language to VDB queries
- **PlanningAgent**: Creates execution plans for complex queries
- **OutputAgent**: Formats results for display
- **BashAgent**: Executes system commands
- **EvalAgent**: Evaluates query results

**LLM Integration**: OpenAI (GPT-4), Claude via `src/pkg/llm/`

### 3. Vector DB Abstraction (`src/pkg/vectordb/`)

**Purpose**: Unified interface across all vector databases

#### Core Interface

```go
type VectorDBClient interface {
    // Health & Connection
    Health(ctx context.Context) error
    Close() error

    // Collections
    CreateCollection(ctx, name string, schema *CollectionSchema) error
    DeleteCollection(ctx, name string) error
    ListCollections(ctx) ([]*Collection, error)
    CollectionExists(ctx, name string) (bool, error)
    GetCollectionCount(ctx, name string) (int64, error)

    // Documents
    CreateDocument(ctx, collection string, doc *Document) error
    CreateDocuments(ctx, collection string, docs []*Document) error
    GetDocument(ctx, collection, id string) (*Document, error)
    UpdateDocument(ctx, collection string, doc *Document) error
    DeleteDocument(ctx, collection, id string) error
    DeleteDocuments(ctx, collection string, ids []string) error
    DeleteDocumentsByMetadata(ctx, collection string, metadata map[string]interface{}) error
    ListDocuments(ctx, collection string, limit, offset int) ([]*Document, error)

    // Search
    SearchSemantic(ctx, collection, query string, opts *QueryOptions) ([]*SearchResult, error)
    SearchBM25(ctx, collection, query string, opts *QueryOptions) ([]*SearchResult, error)
    SearchHybrid(ctx, collection, query string, opts *QueryOptions) ([]*SearchResult, error)
    SearchByMetadata(ctx, collection string, metadata map[string]interface{}, opts *QueryOptions) ([]*SearchResult, error)

    // Schema
    GetSchema(ctx, collection string) (*CollectionSchema, error)
    UpdateSchema(ctx, collection string, schema *CollectionSchema) error
    GetDefaultSchema(schemaType SchemaType, collection string) *CollectionSchema
    ValidateSchema(schema *CollectionSchema) error
}
```

#### VDB Adapter Pattern

Each VDB implements the interface via adapter pattern:

```
src/pkg/vectordb/
├── interfaces.go         # VectorDBClient interface
├── factory.go           # Factory for creating clients
├── errors.go            # Standard error types
├── timeout.go           # Operation-specific timeouts
├── weaviate/
│   ├── adapter.go       # Weaviate adapter (implements VectorDBClient)
│   ├── collections.go   # Collection operations
│   ├── documents.go     # Document operations
│   ├── queries.go       # Search operations
│   ├── schema.go        # Schema management
│   └── factory.go       # Factory registration
├── qdrant/
│   ├── adapter.go       # Qdrant adapter
│   ├── client.go        # Qdrant client wrapper
│   ├── collection.go    # Collection ops
│   ├── document.go      # Document ops
│   ├── query.go         # Search ops
│   └── factory.go       # Factory registration
└── [8 more VDBs...]
```

**Supported VDBs** (10 total):
- ✅ Stable: Weaviate, Qdrant, Milvus, Chroma, Supabase, Neo4j, MongoDB
- 🟢 Beta: Pinecone, Elasticsearch, OpenSearch

### 4. Configuration System (`src/pkg/config/`)

**Purpose**: Manage VDB connection configurations

**Configuration Sources** (priority order):
1. Command-line flags
2. Environment variables
3. Config files (`configs/config.*.yaml`)
4. Defaults

**Config Structure**:
```yaml
vector_db:
  type: qdrant-cloud
  url: https://xyz.qdrant.io
  api_key: ${QDRANT_API_KEY}
  timeout: 60
  vector_dimensions: 1536
  similarity_metric: cosine
```

### 5. Document Processing (`src/pkg/pdf/`, `src/pkg/image/`)

**Purpose**: Extract text and metadata from documents

**Capabilities**:
- **PDF Processing**: Text extraction, image extraction from PDFs
- **Image Processing**: OCR, EXIF metadata extraction
- **Storage**: Configurable paths for extracted images

### 6. MCP Integration (`src/pkg/mcp/`)

**Purpose**: Model Context Protocol client for external integrations

**Features**:
- Connect to MCP servers
- Tool discovery and invocation
- Resource management

## Key Design Patterns

### 1. Adapter Pattern

Each VDB has a dedicated adapter implementing `VectorDBClient`:

```go
type Adapter struct {
    client    *nativeClient  // VDB-specific SDK client
    config    *Config
    llmClient *llm.OpenAIClient  // For embeddings
}

func (a *Adapter) Health(ctx context.Context) error {
    // VDB-specific health check
}
```

### 2. Factory Pattern

VDB clients created via factory with automatic registration:

```go
func CreateClient(config *Config) (VectorDBClient, error) {
    factory, exists := factories[config.Type]
    if !exists {
        return nil, ErrUnsupportedVectorDB(config.Type)
    }
    return factory(config)
}
```

### 3. Error Wrapping

Consistent error handling with context:

```go
func (a *Adapter) CreateCollection(...) error {
    if err := validateInput(); err != nil {
        return ErrInvalidInput(err)
    }
    if err := client.Create(); err != nil {
        return fmt.Errorf("Qdrant: failed to create collection: %w", err)
    }
    return nil
}
```

### 4. Timeout Strategy

Operation-specific timeouts based on deployment type:

```go
func (a *Adapter) CreateDocument(ctx context.Context, ...) error {
    ctx, cancel := context.WithTimeout(ctx,
        a.getTimeoutFor(vectordb.OperationTypeDocument))
    defer cancel()
    // ... operation
}
```

**Timeout Values**:
- Health: 10s (local) / 20s (cloud)
- Document: 15s / 30s
- Collection: 20s / 40s
- Query: 20s / 40s
- Bulk: 120s / 300s

## Data Flow

### 1. Query Execution Flow

```
User Command
    ↓
CLI Parser (Cobra)
    ↓
WeaveAgent
    ├→ QueryAgent (parse natural language)
    ├→ PlanningAgent (create execution plan)
    └→ VectorDBClient
         ├→ Adapter (VDB-specific)
         └→ Native VDB SDK
              ↓
         Vector Database
              ↓
         Results
              ↓
    OutputAgent (format)
              ↓
    Display to User
```

### 2. Document Ingestion Flow

```
Document Upload
    ↓
PDF/Image Processor
    ├→ Text Extraction
    ├→ Image Extraction (if PDF)
    └→ EXIF Metadata (if image)
         ↓
LLM Client (OpenAI)
    └→ Generate Embeddings
         ↓
VectorDBClient
    └→ CreateDocument(s)
         ↓
Vector Database
```

## Testing Strategy

### Integration Tests

Located in `tests/` directory:
- One file per VDB: `tests/weaviate_integration_test.go`, etc.
- Comprehensive test suites (10-16 tests each)
- Tests skip when VDB not available
- **Coverage**: Qdrant 50.6% (representative)

**Test Categories**:
1. Health checks
2. Collection CRUD
3. Document CRUD
4. Batch operations
5. Semantic search
6. BM25 search
7. Hybrid search
8. Metadata filtering
9. Schema operations

### Unit Tests

- Minimal unit tests in package directories
- Most testing via integration tests with real VDBs
- See `docs/TEST_COVERAGE.md` for detailed analysis

## Configuration Management

### VDB Types

```go
const (
    VectorDBTypeWeaviateCloud   = "weaviate-cloud"
    VectorDBTypeWeaviateLocal   = "weaviate-local"
    VectorDBTypeQdrantCloud     = "qdrant-cloud"
    VectorDBTypeQdrantLocal     = "qdrant-local"
    VectorDBTypeMilvusCloud     = "milvus-cloud"
    VectorDBTypeMilvusLocal     = "milvus-local"
    // ... 20 total types (10 VDBs × 2 deployment types)
)
```

### Environment Variables

```bash
# VDB Connection
WEAVIATE_URL=http://localhost:8080
WEAVIATE_API_KEY=secret

# LLM Integration
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...

# Opik Observability
OPIK_API_KEY=...
OPIK_WORKSPACE=default
```

## Error Handling

### Standard Error Types

```go
// Connection errors
func ErrConnectionFailed(msg string, err error) error

// Authentication errors
func ErrAuthenticationFailed(msg string) error

// Resource errors
func ErrNotFound(resource, id string) error

// Unsupported operations
func ErrUnsupported(operation string) error

// Configuration errors
func ErrInvalidConfig(msg string) error
```

### Troubleshooting Hints

All 10 VDBs provide helpful error messages (v0.8.2):

```go
func (a *Adapter) Health(ctx context.Context) error {
    if err := a.client.Ping(); err != nil {
        errMsg := err.Error()
        if strings.Contains(errMsg, "connection refused") {
            return fmt.Errorf("Qdrant: health check failed: %w\n\n"+
                "Connection refused. Common causes:\n"+
                "  1. Qdrant not running (docker run...)\n"+
                "  2. Wrong URL/port in configuration\n"+
                "  3. For Qdrant Cloud: verify API key\n"+
                "  → Check connection at https://cloud.qdrant.io", err)
        }
        return err
    }
}
```

## Performance Considerations

### Timeout Optimization

- Operation-specific timeouts prevent false failures
- Cloud deployments get 2x timeout (network latency)
- Bulk operations get 5-10x timeout (large batches)

### Batch Operations

All VDBs support batch document creation:
- `CreateDocuments([]*Document)` - Single API call
- Tests verify actual batch creation (v0.8.2)
- Critical for pipelining workflows

### Connection Pooling

- HTTP-based VDBs: Reuse http.Client
- gRPC-based VDBs: Connection pooling via SDK
- PostgreSQL-based (Supabase): sql.DB connection pool

## Extensibility

### Adding a New VDB

1. Create package: `src/pkg/vectordb/newvdb/`
2. Implement `VectorDBClient` interface via adapter
3. Register factory: `RegisterFactory("newvdb-local", NewAdapter)`
4. Add config type constant
5. Create integration tests
6. Add documentation: `docs/newvdb/SETUP.md`

**Example**: See Elasticsearch implementation (v0.8.0)

### Adding a New Agent

1. Create `src/pkg/agents/new_agent.go`
2. Implement agent interface
3. Register with WeaveAgent
4. Add prompts in agent file

## Dependencies

### Core Dependencies

- **CLI**: `github.com/spf13/cobra`
- **Config**: `github.com/spf13/viper`
- **LLM**: OpenAI, Anthropic SDKs
- **VDB SDKs**: 10 different SDKs (see go.mod)

### No CGO Required

- Pure Go implementation
- Exception: Chroma (CGO on some platforms, gracefully degrades)

## Security

### Known Issues

- **Weaviate GO-2025-4237**: Path traversal vulnerability
  - Impact: LOW (weave-cli doesn't use backup functionality)
  - Status: Waiting for SDK compatibility with patched version
  - Tracked: https://pkg.go.dev/vuln/GO-2025-4237

### API Key Management

- Environment variables recommended
- Config files support `${ENV_VAR}` syntax
- Never commit API keys to git

## Future Architecture Enhancements

See `docs/TODO_AUDIT.md` for planned improvements:
- AWS Signature V4 for OpenSearch
- Enhanced bulk operations
- Advanced query pagination
- Custom TLS configurations

## References

- **Main README**: `/README.md`
- **VDB Support Matrix**: `/docs/VDB_SUPPORT_MATRIX.md`
- **Test Coverage**: `/docs/TEST_COVERAGE.md`
- **TODO Audit**: `/docs/TODO_AUDIT.md`
- **Per-VDB Docs**: `/docs/{vdb}/SETUP.md`
