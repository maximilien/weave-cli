# Weave CLI Architecture

**Last Updated**: 2026-02-12
**Version**: v0.9.19+

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

### 7. Embedding Provider Architecture (`src/pkg/reembedding/providers/`) [v0.9.19+]

**Purpose**: Pluggable embedding model support for OSS and proprietary providers

The embedding provider architecture enables re-embedding collections with different models (OpenAI, sentence-transformers, Ollama) without re-processing source documents. This provides 20x performance improvement over full re-ingestion.

#### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│               Re-embedding Command Layer                     │
│        (src/cmd/collection/re_embed.go)                     │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│              Embedding Pipeline                              │
│         (src/pkg/reembedding/pipeline.go)                   │
│  • Batch processing (100-500 docs/batch)                    │
│  • Progress tracking                                         │
│  • Error handling                                            │
└────────────────┬────────────────────────────────────────────┘
                 │
         ┌───────▼────────┐
         │    Provider     │
         │     Factory     │
         │ CreateProvider()│
         │  (providers/)   │
         └───────┬────────┘
                 │
     ┌───────────┼───────────┬─────────────┐
     │           │           │             │
┌────▼────┐ ┌───▼────┐ ┌───▼────┐  ┌────▼────┐
│ OpenAI  │ │sentence│ │ Ollama │  │  Future │
│Provider │ │transf. │ │Provider│  │Providers│
│  (API)  │ │(Python)│ │ (HTTP) │  │         │
└────┬────┘ └───┬────┘ └───┬────┘  └─────────┘
     │          │           │
     └──────────┼───────────┘
                │
         ┌──────▼──────┐
         │  Document   │
         │ .Embedding  │
         │ []float64   │
         │ .Text       │
         └──────┬──────┘
                │
         ┌──────▼──────┐
         │ VDB Adapter │
         │Check first: │
         │len(emb) > 0?│
         │  Use/Skip   │
         └──────┬──────┘
                │
         ┌──────▼──────┐
         │Vector Database│
         │ (any VDB)    │
         └──────────────┘
```

#### Provider Interface

All embedding providers implement this interface:

```go
type EmbeddingProvider interface {
    GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
    GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error)
    IsAvailable(ctx context.Context) error
    GetDimensions() int
    GetProvider() string
}
```

#### Supported Providers

**1. OpenAI Provider** (`providers/openai.go`)
- **Models**: text-embedding-3-small (1536 dims), text-embedding-3-large (3072 dims)
- **Protocol**: REST API
- **Requirements**: OPENAI_API_KEY environment variable
- **Performance**: 200+ docs/min
- **Cost**: $0.02 per 1M tokens

**2. Sentence-Transformers Provider** (`providers/sentence_transformers.go`)
- **Models**:
  - all-mpnet-base-v2 (768 dims) - highest quality
  - all-MiniLM-L6-v2 (384 dims) - fastest
- **Protocol**: Python subprocess (sentence-transformers library)
- **Requirements**: Python 3.8+, sentence-transformers installed
- **Performance**: 150+ docs/min
- **Cost**: $0 (open-source)
- **Quality**: 92-95% of OpenAI quality (often better!)

**3. Ollama Provider** (`providers/ollama.go`)
- **Models**:
  - nomic-embed-text (768 dims)
  - mxbai-embed-large (1024 dims)
- **Protocol**: HTTP API (localhost:11434)
- **Requirements**: Ollama installed and running locally
- **Performance**: 180+ docs/min
- **Cost**: $0 (open-source, local)
- **Quality**: 90-93% of OpenAI quality

#### Provider Factory

The factory pattern enables automatic provider selection based on model name:

```go
// providers/factory.go
func CreateProvider(ctx context.Context, modelName string) (EmbeddingProvider, error) {
    // Auto-detect provider from model name
    if strings.HasPrefix(modelName, "text-embedding-") {
        return NewOpenAIProvider(ctx, modelName)
    }
    if strings.HasPrefix(modelName, "sentence-transformers/") {
        return NewSentenceTransformersProvider(ctx, modelName)
    }
    if modelName == "nomic-embed-text" || modelName == "mxbai-embed-large" {
        return NewOllamaProvider(ctx, modelName)
    }
    return nil, fmt.Errorf("unknown model: %s", modelName)
}
```

#### Re-embedding Pipeline

The pipeline orchestrates the re-embedding process:

```go
// src/pkg/reembedding/pipeline.go
type EmbeddingPipeline struct {
    provider   EmbeddingProvider
    dimensions int
}

func (p *EmbeddingPipeline) ProcessBatch(ctx context.Context, docs []*Document) error {
    // Extract text from documents
    texts := extractTexts(docs)

    // Generate embeddings in batch
    embeddings, err := p.provider.GenerateEmbeddings(ctx, texts)
    if err != nil {
        return err
    }

    // Attach embeddings to documents
    for i, doc := range docs {
        doc.Embedding = embeddings[i]
    }

    return nil
}
```

#### Pre-generated Embedding Pattern

**Critical Performance Optimization**: VDB adapters check for pre-generated embeddings and skip regeneration.

**Example** (`src/pkg/vectordb/milvus/document.go:155`):
```go
func (a *Adapter) CreateDocuments(ctx context.Context, collection string, docs []*Document) error {
    for _, doc := range docs {
        // If document already has embedding, skip generation
        if len(doc.Embedding) > 0 {
            embeddings = append(embeddings, doc.Embedding)
            continue
        }

        // Otherwise generate (legacy path for non-re-embedded docs)
        embedding, err := a.llmClient.GenerateEmbedding(ctx, doc.Text, "")
        embeddings = append(embeddings, embedding)
    }
}
```

This pattern enables:
- **20x faster re-embedding** (vs full document re-ingestion)
- **Batch processing efficiency** (100-500 docs/batch)
- **Provider independence** (any embedding model works with any VDB)

#### Query Embedding Matching [v0.9.19+]

**Critical Fix**: Queries now automatically use the collection's embedding model to ensure dimension matching.

**Problem Solved**: Previously, queries always used OpenAI (1536 dims), causing dimension mismatch errors when collections were re-embedded with OSS models (768 dims).

**Solution** (`src/pkg/vectordb/milvus/query.go`):
```go
func (a *Adapter) SearchSemantic(ctx context.Context, collection, query string, ...) {
    // Get collection's embedding model from metadata
    schema, _ := a.Client.GetSchema(ctx, collection)
    embeddingModel := schema.Vectorizer

    // Use appropriate provider
    if isOpenAI(embeddingModel) {
        queryEmbedding = a.llmClient.GenerateEmbedding(ctx, query, "")
    } else {
        // Use provider factory for OSS models
        provider := providers.CreateProvider(ctx, embeddingModel)
        queryEmbedding = provider.GenerateEmbedding(ctx, query)
    }

    // Query with matching dimensions ✅
}
```

**Collection Metadata Storage**: Vectorizer stored in collection description field:
```go
// Format: "Collection X for vector search | vectorizer=MODEL_NAME"
description := fmt.Sprintf("%s | vectorizer=%s", baseDescription, schema.Vectorizer)
```

#### Performance Metrics (Client0 Production Results)

**Test Collection**: 426 auction documents

| Metric                  | OpenAI          | sentence-transformers | Winner      |
|-------------------------|-----------------|----------------------|-------------|
| **Re-embed Time**       | ~5+ hours       | 85 seconds           | ✅ OSS 240x |
| **Re-embed Speed**      | N/A             | 308 docs/min         | ✅ OSS      |
| **Quality Score**       | 0.606 avg       | 0.673 avg (+11%)     | ✅ OSS      |
| **Dimensions**          | 1536            | 768 (50% smaller)    | ✅ OSS      |
| **Cost per 1M tokens**  | $0.02           | $0.00                | ✅ OSS      |
| **Query Latency**       | 1.5s            | 7.6s                 | ⚠️ OpenAI   |

**Key Finding**: OSS embeddings provide better quality with zero cost, 20x faster re-embedding, and 50% smaller vectors. Query latency trade-off (5x slower) acceptable for most use cases.

#### Usage Example

```bash
# List available embedding models
weave embeddings list

# Re-embed with sentence-transformers (OSS)
weave collection re-embed MyCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output MyCollection_OSS \
  --batch-size 100

# Result: 426/426 documents in 85 seconds ✅

# Query uses collection's embedding model automatically
weave query search MyCollection_OSS "vintage camera" --top-k 5

# Result: 5 results with 0.673 avg quality ✅
```

#### Files Reference

**Core Re-embedding**:
- `src/pkg/reembedding/pipeline.go` - Pipeline orchestration
- `src/pkg/reembedding/reader.go` - Batch document reading
- `src/pkg/reembedding/progress.go` - Progress tracking

**Provider Implementations**:
- `src/pkg/reembedding/providers/factory.go` - Provider factory
- `src/pkg/reembedding/providers/openai.go` - OpenAI provider
- `src/pkg/reembedding/providers/sentence_transformers.go` - sentence-transformers
- `src/pkg/reembedding/providers/ollama.go` - Ollama provider
- `src/pkg/reembedding/providers/interfaces.go` - Provider interface

**VDB Integration**:
- `src/pkg/vectordb/milvus/document.go` - Pre-generated embedding check
- `src/pkg/vectordb/milvus/query.go` - Query embedding matching
- `src/pkg/vectordb/milvus/collection.go` - Vectorizer metadata storage

**Command Layer**:
- `src/cmd/collection/re_embed.go` - Re-embed command
- `src/cmd/embeddings/list.go` - List models command

#### Benefits

✅ **Cost Savings**: $240/year per million documents (OpenAI → OSS)
✅ **Performance**: 20x faster re-embedding vs full re-ingestion
✅ **Quality**: Often better than OpenAI (Client0: +11% improvement)
✅ **Flexibility**: Test different models quickly (minutes, not hours)
✅ **Privacy**: OSS models run locally, data never leaves infrastructure
✅ **No Vendor Lock-in**: Switch embedding models freely
✅ **Provider Independence**: Works with all 10 VDBs
✅ **Automatic Query Matching**: Queries use collection's model automatically

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

## Embedding Provider Architecture

### Overview

The embedding pipeline uses a **factory pattern** to support multiple embedding providers, enabling flexible and cost-effective document vectorization. All providers generate embeddings independently, which are then stored in vector databases.

**Key Benefit**: Pre-generated embeddings enable 20x faster re-embedding without re-ingestion.

### Provider Interface

All embedding providers implement the `EmbeddingProvider` interface:

```go
type EmbeddingProvider interface {
    // GenerateEmbedding creates a vector for a single text
    GenerateEmbedding(ctx context.Context, text string) ([]float64, error)

    // GenerateEmbeddings creates vectors for multiple texts (batch)
    GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error)

    // IsAvailable checks if provider dependencies are installed
    IsAvailable(ctx context.Context) error

    // GetDimensions returns the embedding vector size
    GetDimensions() int
}
```

**Location**: `src/pkg/reembedding/providers/`

### Embedding Pipeline Flow

```
┌──────────────────────────────────────────────────────────┐
│               Embedding Pipeline                         │
│          (reembedding/pipeline.go)                       │
│   • Loads documents from VDB                             │
│   • Batches for efficiency                               │
│   • Generates embeddings via provider                    │
│   • Updates documents with new vectors                   │
└────────────────────┬─────────────────────────────────────┘
                     │
         ┌───────────▼────────────┐
         │   Provider Factory      │
         │  CreateProvider(model)  │
         │   • Auto-detects type   │
         │   • Validates available │
         └───────────┬─────────────┘
                     │
     ┌───────────────┼───────────────┬──────────────┐
     │               │               │              │
┌────▼─────┐  ┌─────▼──────┐  ┌────▼──────┐  ┌───▼─────┐
│  OpenAI  │  │ sentence-  │  │  Ollama   │  │  Mock   │
│ Provider │  │transformers│  │ Provider  │  │Provider │
│  (API)   │  │  (Python)  │  │  (HTTP)   │  │ (Test)  │
└────┬─────┘  └─────┬──────┘  └────┬──────┘  └───┬─────┘
     │              │               │              │
     │   Generates embedding vectors ([]float64)  │
     │              │               │              │
     └──────────────┴───────────────┴──────────────┘
                     │
         ┌───────────▼────────────┐
         │      Document          │
         │   .Embedding field     │
         │    []float64 vector    │
         │ (384/768/1536 dims)    │
         └───────────┬────────────┘
                     │
         ┌───────────▼────────────┐
         │    VDB Adapter         │
         │  Check: len(emb) > 0?  │
         │  YES: Use pre-generated│
         │  NO:  Generate new     │
         └────────────────────────┘
```

### Supported Providers

#### 1. OpenAI Provider (API-based)

**Models**:
- `text-embedding-3-small` (1536 dims) - Fast, cost-effective
- `text-embedding-3-large` (3072 dims) - Highest quality
- `text-embedding-ada-002` (1536 dims) - Legacy

**Requirements**:
- `OPENAI_API_KEY` environment variable
- Internet connection

**Performance**: ~200 docs/min, $0.02 per 1M tokens

**Implementation**: `src/pkg/reembedding/providers/openai_provider.go`

#### 2. sentence-transformers Provider (Python subprocess)

**Models**:
- `sentence-transformers/all-mpnet-base-v2` (768 dims) - Best quality
- `sentence-transformers/all-MiniLM-L6-v2` (384 dims) - Fast, lightweight

**Requirements**:
- Python 3.8+
- `pip install sentence-transformers`

**Performance**: ~150 docs/min (CPU), faster with GPU

**Cost**: $0 (open source)

**Implementation**: `src/pkg/reembedding/providers/sentence_transformers_provider.go`

**How it works**:
1. Spawns Python subprocess per batch
2. Passes texts via stdin (JSON)
3. Receives embeddings via stdout
4. Handles errors gracefully

#### 3. Ollama Provider (HTTP API)

**Models**:
- `nomic-embed-text` (768 dims) - Text embeddings
- `mxbai-embed-large` (1024 dims) - High quality

**Requirements**:
- Ollama installed and running (`ollama serve`)
- Model pulled: `ollama pull nomic-embed-text`

**Performance**: ~180 docs/min

**Cost**: $0 (runs locally)

**Implementation**: `src/pkg/reembedding/providers/ollama_provider.go`

**How it works**:
1. HTTP POST to `http://localhost:11434/api/embeddings`
2. Concurrent requests supported by Ollama
3. Local processing, no external dependencies

#### 4. Mock Provider (Testing)

**Purpose**: Unit testing without real embedding services

**Implementation**: Returns zero vectors or fixed test vectors

### Pre-generated Embeddings

**Key Innovation**: Documents can carry pre-generated embeddings, enabling fast provider switching.

#### How VDB Adapters Check for Pre-generated Embeddings

**Example from Milvus** (`src/pkg/vectordb/milvus/document.go:155`):

```go
func (c *Client) CreateDocument(ctx context.Context, collectionName string, doc *vectordb.Document) error {
    var embedding []float32

    // Check for pre-generated embedding
    if len(doc.Embedding) > 0 {
        // Use pre-generated embedding
        embedding = convertFloat64ToFloat32(doc.Embedding)
    } else {
        // Generate new embedding using collection's model
        embedding64, err := c.generateEmbedding(ctx, collectionName, doc.Content)
        if err != nil {
            return err
        }
        embedding = convertFloat64ToFloat32(embedding64)
    }

    // Insert into Milvus with embedding
    return c.insertDocument(ctx, collectionName, doc, embedding)
}
```

**All VDB adapters** (Weaviate, Qdrant, Chroma, etc.) follow this pattern:
1. Check `len(doc.Embedding) > 0`
2. If yes: Use pre-generated embedding
3. If no: Generate using collection's configured model

### Re-embedding Workflow

**Command**: `weave collection reembed`

**Process**:
1. Load all documents from collection
2. Create embedding provider (new model)
3. Generate embeddings in batches
4. Update documents with new vectors
5. Validate dimensions match

**Performance**: ~20x faster than re-ingestion (no PDF parsing, OCR, etc.)

**Example**:
```bash
# Re-embed collection with OSS model
weave collection reembed MyCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --batch-size 100 \
  --milvus-local

# Compare quality
weave collection compare MyCollection_OpenAI MyCollection_OSS \
  --report comparison.md
```

### Provider Selection Logic

**Factory auto-detection** (`src/pkg/reembedding/providers/factory.go`):

```go
func CreateProvider(ctx context.Context, modelName string) (EmbeddingProvider, error) {
    // OpenAI models
    if strings.HasPrefix(modelName, "text-embedding-") {
        return NewOpenAIProvider(modelName)
    }

    // sentence-transformers models
    if strings.HasPrefix(modelName, "sentence-transformers/") {
        return NewSentenceTransformersProvider(modelName)
    }

    // Ollama models (check if available)
    if provider, err := NewOllamaProvider(modelName); err == nil {
        return provider, nil
    }

    return nil, fmt.Errorf("unknown embedding model: %s", modelName)
}
```

### Error Handling

**Graceful degradation**:
- Python not installed → Clear error message with install instructions
- Ollama not running → Fallback to other providers or error
- API key missing → Helpful error with env var name

**Example error messages**:
```
❌ sentence-transformers provider not available
   Python package 'sentence-transformers' not found

   Install with: pip install sentence-transformers
   Or use: --embedding text-embedding-3-small
```

### Performance Optimization

**Batch Processing**:
- Default batch size: 100 documents
- sentence-transformers: Processes entire batch in one subprocess
- Ollama: Concurrent HTTP requests
- OpenAI: API batching with rate limiting

**Memory Management**:
- Streaming for large collections
- Garbage collection between batches
- Progress tracking with time estimates

### Testing

**Unit Tests**: `src/pkg/reembedding/providers/*_test.go`
- Mock provider tests
- Factory creation tests
- Error handling tests

**Integration Tests**: Require actual providers installed
- Skip tests if provider unavailable
- Marked with build tags

**Example test**:
```go
func TestSentenceTransformersProvider(t *testing.T) {
    if !isSentenceTransformersAvailable() {
        t.Skip("sentence-transformers not installed")
    }

    provider := NewSentenceTransformersProvider("sentence-transformers/all-MiniLM-L6-v2")
    embedding, err := provider.GenerateEmbedding(context.Background(), "test text")

    assert.NoError(t, err)
    assert.Equal(t, 384, len(embedding)) // Correct dimensions
}
```

### Cost Comparison

**1 Million Documents Re-embedded Monthly**:

| Provider | Monthly Cost | Annual Cost | Quality vs OpenAI |
|----------|--------------|-------------|-------------------|
| OpenAI text-embedding-3-small | $20 | $240 | 100% (baseline) |
| sentence-transformers all-mpnet-base-v2 | $0 | $0 | 92-95% |
| Ollama nomic-embed-text | $0 | $0 | 90-93% |

**Annual Savings**: $240/year with 90%+ quality retention

### Future Enhancements

- Hugging Face Inference API support
- Cohere embeddings integration
- Custom model fine-tuning
- Embedding caching layer
- Dimension reduction (PCA, UMAP)

---

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
