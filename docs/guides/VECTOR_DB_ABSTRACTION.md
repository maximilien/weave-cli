# Vector Database Abstraction Layer

## Overview

The Weave CLI uses a vector database abstraction layer that allows support for multiple vector database backends. This abstraction provides a unified interface for all vector database operations, making it easy to add support for new vector databases like Supabase PGVector and Milvus.

## Architecture

### Core Components

1. **Interfaces** (`src/pkg/vectordb/interfaces.go`)
   - Defines the `VectorDBClient` interface that all vector database implementations must satisfy
   - Includes sub-interfaces: `CollectionOperations`, `DocumentOperations`, `QueryOperations`, `SchemaOperations`

2. **Factory Pattern** (`src/pkg/vectordb/factory.go`)
   - `ClientFactory` interface for creating vector database clients
   - `Registry` for managing multiple factory implementations
   - Global registry for easy access

3. **Error Handling** (`src/pkg/vectordb/errors.go`)
   - Standardized error types for vector database operations
   - Error categorization (connection, authentication, not found, etc.)

4. **Adapters** (`src/pkg/vectordb/{database}/`)
   - Database-specific implementations of the `VectorDBClient` interface
   - Each adapter wraps the native client library

## Supported Databases

### Weaviate
- **Types**: `weaviate-cloud`, `weaviate-local`
- **Adapter**: `src/pkg/vectordb/weaviate/adapter.go`
- **Factory**: `src/pkg/vectordb/weaviate/factory.go`
- **Status**: ? Fully implemented

### Mock
- **Type**: `mock`
- **Adapter**: `src/pkg/vectordb/mock/adapter.go`
- **Factory**: `src/pkg/vectordb/mock/factory.go`
- **Status**: ? Fully implemented

### Supabase PGVector
- **Type**: `supabase`
- **Adapter**: `src/pkg/vectordb/supabase/adapter.go`
- **Factory**: `src/pkg/vectordb/supabase/factory.go`
- **Status**: ? Fully implemented

### Milvus
- **Type**: `milvus`
- **Status**: ?? Planned

## Interface Design

### VectorDBClient

The main interface that all vector database implementations must satisfy:

```go
type VectorDBClient interface {
    // Health checks
    Health(ctx context.Context) error

    // Collection operations
    CollectionOperations

    // Document operations
    DocumentOperations

    // Query operations
    QueryOperations

    // Schema operations
    SchemaOperations
}
```

### CollectionOperations

```go
type CollectionOperations interface {
    CreateCollection(ctx context.Context, name string, schema *CollectionSchema) error
    DeleteCollection(ctx context.Context, name string) error
    ListCollections(ctx context.Context) ([]CollectionInfo, error)
    CollectionExists(ctx context.Context, name string) (bool, error)
    GetCollectionCount(ctx context.Context, name string) (int64, error)
}
```

### DocumentOperations

```go
type DocumentOperations interface {
    CreateDocument(ctx context.Context, collectionName string, document *Document) error
    CreateDocuments(ctx context.Context, collectionName string, documents []*Document) error
    GetDocument(ctx context.Context, collectionName, documentID string) (*Document, error)
    UpdateDocument(ctx context.Context, collectionName string, document *Document) error
    DeleteDocument(ctx context.Context, collectionName, documentID string) error
    DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error
    DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error
    ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*Document, error)
}
```

### QueryOperations

```go
type QueryOperations interface {
    SearchSemantic(ctx context.Context, collectionName, query string, options *QueryOptions) ([]*QueryResult, error)
    SearchBM25(ctx context.Context, collectionName, query string, options *QueryOptions) ([]*QueryResult, error)
    SearchHybrid(ctx context.Context, collectionName, query string, options *QueryOptions) ([]*QueryResult, error)
    SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *QueryOptions) ([]*QueryResult, error)
}
```

### SchemaOperations

```go
type SchemaOperations interface {
    GetSchema(ctx context.Context, collectionName string) (*CollectionSchema, error)
    UpdateSchema(ctx context.Context, collectionName string, schema *CollectionSchema) error
    GetDefaultSchema(schemaType SchemaType, collectionName string) *CollectionSchema
    ValidateSchema(schema *CollectionSchema) error
}
```

## Usage

### Creating a Client

```go
import (
    "github.com/maximilien/weave-cli/src/pkg/config"
    "github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// From VectorDBConfig
cfg := &config.VectorDBConfig{
    Type:   config.VectorDBTypeCloud,
    URL:    "https://your-instance.weaviate.cloud",
    APIKey: "your-api-key",
}

client, err := vectordb.CreateClientFromVectorDBConfig(cfg)
if err != nil {
    // Handle error
}

// Or directly from vectordb.Config
vdbConfig := &vectordb.Config{
    Type:   vectordb.VectorDBTypeWeaviateCloud,
    URL:    "https://your-instance.weaviate.cloud",
    APIKey: "your-api-key",
}

client, err := vectordb.CreateClient(vdbConfig)
if err != nil {
    // Handle error
}
```

### Using the Client

```go
// Health check
if err := client.Health(ctx); err != nil {
    // Handle error
}

// Create collection
schema := &vectordb.CollectionSchema{
    Class:      "MyCollection",
    Vectorizer: "text2vec-openai",
    Properties: []vectordb.SchemaProperty{
        {
            Name:     "content",
            DataType: []string{"text"},
        },
    },
}

if err := client.CreateCollection(ctx, "MyCollection", schema); err != nil {
    // Handle error
}

// Create document
doc := &vectordb.Document{
    ID:      "doc-1",
    Content: "Hello, world!",
    Metadata: map[string]interface{}{
        "title": "Test Document",
    },
}

if err := client.CreateDocument(ctx, "MyCollection", doc); err != nil {
    // Handle error
}

// Search
results, err := client.SearchSemantic(ctx, "MyCollection", "hello", &vectordb.QueryOptions{
    TopK: 10,
})
if err != nil {
    // Handle error
}

for _, result := range results {
    fmt.Printf("Score: %f, Content: %s\n", result.Score, result.Document.Content)
}
```

## Supabase PGVector Implementation

### Overview

The Supabase implementation provides full vector database functionality using PostgreSQL with the pgvector extension. This implementation leverages Supabase's managed PostgreSQL service to provide scalable vector storage and search capabilities.

### Features

- **Full CRUD Operations**: Create, read, update, and delete documents
- **Vector Search**: Semantic search using pgvector similarity functions
- **Full-Text Search**: BM25-style search using PostgreSQL's built-in full-text search
- **Hybrid Search**: Combines vector and keyword search results
- **Metadata Filtering**: Search and filter documents by metadata fields
- **Schema Management**: Dynamic schema creation and validation
- **Batch Operations**: Efficient batch document creation and deletion
- **JSONB Support**: Rich metadata storage using PostgreSQL's JSONB type

### Configuration

To use Supabase as your vector database, configure it as follows:

```go
config := &vectordb.Config{
    Type:        vectordb.VectorDBTypeSupabase,
    DatabaseURL: "postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres",
    DatabaseKey: "your-supabase-anon-key",
    Timeout:     30,
}

client, err := vectordb.CreateClient(config)
if err != nil {
    log.Fatal(err)
}
```

### Environment Variables

For integration testing and production use, set these environment variables:

```bash
export SUPABASE_DATABASE_URL="postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
export SUPABASE_DATABASE_KEY="your-supabase-anon-key"
```

### Database Setup

Before using the Supabase implementation, ensure your Supabase project has the pgvector extension enabled:

1. Go to your Supabase project dashboard
2. Navigate to the SQL Editor
3. Run the following SQL command:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### Implementation Details

#### Architecture

The Supabase implementation consists of several key components:

- **Adapter** (`adapter.go`): Main implementation of the VectorDBClient interface
- **Collections** (`collections.go`): Collection management and schema operations
- **Documents** (`documents.go`): Document CRUD operations
- **Queries** (`queries.go`): Search and query operations
- **Schema** (`schema.go`): Schema validation and management
- **Factory** (`factory.go`): Factory pattern implementation and registration

#### Table Structure

Each collection is stored as a PostgreSQL table with the following standard columns:

- `id` (TEXT PRIMARY KEY): Document identifier
- `content` (TEXT): Main document content
- `text` (TEXT): Additional text content
- `image` (TEXT): Image URL or reference
- `image_data` (TEXT): Base64 encoded image data
- `url` (TEXT): Associated URL
- `metadata` (JSONB): Document metadata as JSON
- `embedding` (VECTOR(1536)): Vector embeddings for semantic search

#### Search Capabilities

1. **Semantic Search**: Uses vector similarity with pgvector
2. **BM25 Search**: PostgreSQL full-text search with ranking
3. **Hybrid Search**: Combines semantic and keyword search results
4. **Metadata Search**: Filters documents by metadata fields using JSONB operators

#### Error Handling

The implementation provides comprehensive error handling with proper categorization:

- Connection errors for database connectivity issues
- Authentication errors for invalid credentials
- Not found errors for missing resources
- Invalid configuration errors for setup issues

### Testing

The Supabase implementation includes comprehensive test coverage:

#### Unit Tests

Run unit tests that don't require a database connection:

```bash
go test ./tests/supabase_test.go -v
```

#### Integration Tests

Run integration tests with a real Supabase instance:

```bash
export SUPABASE_DATABASE_URL="your-database-url"
export SUPABASE_DATABASE_KEY="your-database-key"
go test ./tests/supabase_integration_test.go -v
```

### Performance Considerations

- **Indexing**: Automatic creation of vector and JSONB indexes for optimal performance
- **Batch Operations**: Use batch document creation for better throughput
- **Connection Pooling**: Leverages PostgreSQL connection pooling
- **Query Optimization**: Optimized queries for vector similarity search

### Limitations

- **Vector Dimensions**: Currently configured for 1536-dimensional vectors (OpenAI embeddings)
- **Extension Dependency**: Requires pgvector extension to be installed
- **PostgreSQL Version**: Requires PostgreSQL 11+ with pgvector support

## Adding a New Vector Database

To add support for a new vector database (e.g., Supabase PGVector or Milvus):

### 1. Create the Adapter

Create a new directory `src/pkg/vectordb/{database}/` and implement:

- `adapter.go` - Implements the `VectorDBClient` interface
- `factory.go` - Implements the `ClientFactory` interface

### 2. Implement Required Methods

The adapter must implement all methods from:
- `CollectionOperations`
- `DocumentOperations`
- `QueryOperations`
- `SchemaOperations`

### 3. Register the Factory

In `factory.go`, register the factory in the `init()` function:

```go
func init() {
    factory := NewFactory()
    vectordb.RegisterFactory(vectordb.VectorDBTypeSupabase, factory)
}
```

### 4. Add Configuration Support

Update `src/pkg/vectordb/factory.go` to include the new database type:

```go
const (
    VectorDBTypeSupabase VectorDBType = "supabase"
    // ...
)
```

Add configuration fields to `vectordb.Config` if needed:

```go
type Config struct {
    // ... existing fields
    
    // Supabase-specific configuration
    DatabaseURL string `yaml:"database_url,omitempty"`
    DatabaseKey string `yaml:"database_key,omitempty"`
}
```

### 5. Update Configuration Loading

Ensure `config.VectorDBConfig` supports the new database type and can convert to `vectordb.Config`.

### 6. Add Tests

Create tests for the new adapter:
- `src/pkg/vectordb/{database}/adapter_test.go`
- Integration tests in `tests/`

## Error Handling

The abstraction layer provides standardized error types:

```go
// Connection errors
if vectordb.IsConnectionError(err) {
    // Handle connection issue
}

// Authentication errors
if vectordb.IsAuthenticationError(err) {
    // Handle auth issue
}

// Not found errors
if vectordb.IsNotFoundError(err) {
    // Handle not found
}
```

## Benefits

1. **Unified Interface**: All vector databases use the same interface
2. **Easy Extension**: Adding new databases is straightforward
3. **Type Safety**: Compile-time checking ensures all methods are implemented
4. **Error Consistency**: Standardized error handling across all databases
5. **Testability**: Easy to mock for testing

## Migration Status

The abstraction layer is implemented and Weaviate adapter is complete. The codebase is being migrated to use the abstraction layer instead of direct Weaviate client calls.

### Completed
- ? Interface definitions
- ? Factory pattern
- ? Error handling
- ? Weaviate adapter
- ? Mock adapter
- ? Supabase PGVector adapter
- ? Basic client creation utilities

### In Progress
- ?? Migration of utility functions to use VectorDBClient
- ?? Migration of command handlers to use VectorDBClient

### Planned
- ?? Milvus adapter
- ?? Complete migration from direct client calls

## Future Enhancements

1. **Connection Pooling**: Add connection pooling support
2. **Retry Logic**: Standardized retry logic for transient failures
3. **Metrics**: Standardized metrics collection
4. **Tracing**: Distributed tracing support
5. **Batch Operations**: Optimized batch operations
