# Elasticsearch Integration Architecture

**Date**: 2025-12-12
**Status**: Architecture Planning
**Target**: v0.8.0

---

## Overview

This document outlines the architecture for integrating Elasticsearch as the 10th vector database in Weave CLI. The integration follows the established patterns used by MongoDB, Qdrant, and other VDB implementations.

---

## File Structure

```
src/pkg/vectordb/elasticsearch/
├── adapter.go      # VectorDBClient adapter (wraps Client + LLM)
├── client.go       # Core Elasticsearch client wrapper
├── config.go       # Configuration structs and validation
├── factory.go      # Factory registration (init)
├── collection.go   # Collection operations implementation
├── document.go     # Document CRUD operations
├── query.go        # Search operations (vector, BM25, hybrid)
└── schema.go       # Schema operations implementation

tests/
└── elasticsearch_integration_test.go  # Integration tests

docs/elasticsearch/
├── RESEARCH.md      # Research findings (completed)
├── ARCHITECTURE.md  # This file
├── README.md        # Overview and feature list
├── SETUP.md         # General setup guide
├── LOCAL_SETUP.md   # Docker/self-hosted setup
└── CLOUD_SETUP.md   # Elastic Cloud setup
```

---

## Component Breakdown

### 1. config.go

**Purpose**: Define Elasticsearch-specific configuration structures

```go
package elasticsearch

import (
	"fmt"
	"strings"
)

// Config holds Elasticsearch client configuration
type Config struct {
	// Connection settings (choose one method)
	// Method 1: Elastic Cloud
	CloudID string // Cloud ID from Elastic Cloud console
	APIKey  string // API key for authentication

	// Method 2: Self-hosted
	Addresses []string // List of Elasticsearch addresses

	// Authentication (for self-hosted)
	Username string
	Password string

	// Optional settings
	CertFingerprint string // SHA256 fingerprint for cert validation
	Timeout         int    // Timeout in seconds (default: 10)
	MaxRetries      int    // Max retry attempts (default: 3)
	EnableDebugLog  bool   // Enable detailed logging

	// Vector search settings
	VectorDimensions int    // Embedding vector dimensions (e.g., 1536)
	SimilarityMetric string // Similarity: "cosine", "dot_product", "l2_norm"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Must have either CloudID or Addresses
	if c.CloudID == "" && len(c.Addresses) == 0 {
		return fmt.Errorf("either CloudID or Addresses must be provided")
	}

	// CloudID requires APIKey
	if c.CloudID != "" && c.APIKey == "" {
		return fmt.Errorf("CloudID requires APIKey")
	}

	// Addresses with auth requires username/password or APIKey
	if len(c.Addresses) > 0 {
		hasAuth := c.Username != "" || c.APIKey != ""
		if !hasAuth {
			return fmt.Errorf("self-hosted Elasticsearch requires authentication")
		}
	}

	// Validate similarity metric
	if c.SimilarityMetric != "" {
		validMetrics := map[string]bool{
			"cosine":      true,
			"dot_product": true,
			"l2_norm":     true,
		}
		if !validMetrics[c.SimilarityMetric] {
			return fmt.Errorf("invalid similarity metric: %s", c.SimilarityMetric)
		}
	}

	return nil
}
```

### 2. client.go

**Purpose**: Core Elasticsearch client wrapper with health checks

```go
package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
)

// Client wraps the Elasticsearch TypedClient
type Client struct {
	client *elasticsearch.TypedClient
	config *Config
}

// NewClient creates a new Elasticsearch client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Set defaults
	if config.VectorDimensions == 0 {
		config.VectorDimensions = 1536 // OpenAI ada-002 default
	}
	if config.SimilarityMetric == "" {
		config.SimilarityMetric = "cosine"
	}
	if config.Timeout == 0 {
		config.Timeout = 10
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	// Build Elasticsearch config
	esConfig := elasticsearch.Config{}

	// Configure connection
	if config.CloudID != "" {
		esConfig.CloudID = config.CloudID
		esConfig.APIKey = config.APIKey
	} else {
		esConfig.Addresses = config.Addresses
		if config.APIKey != "" {
			esConfig.APIKey = config.APIKey
		} else {
			esConfig.Username = config.Username
			esConfig.Password = config.Password
		}
	}

	// Optional settings
	if config.CertFingerprint != "" {
		esConfig.CertificateFingerprint = config.CertFingerprint
	}
	esConfig.MaxRetries = config.MaxRetries

	// Create typed client
	client, err := elasticsearch.NewTypedClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Timeout)*time.Second)
	defer cancel()

	if _, err := client.Ping().Do(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// Health checks the health of the Elasticsearch cluster
func (c *Client) Health(ctx context.Context) error {
	timeout := time.Duration(c.config.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := c.client.Ping().Do(ctx)
	if err != nil {
		return fmt.Errorf("Elasticsearch health check failed: %w", err)
	}

	return nil
}

// Close closes the Elasticsearch client
func (c *Client) Close(ctx context.Context) error {
	// TypedClient doesn't have explicit close method
	// Connection will be closed when client is garbage collected
	return nil
}

// getTimeout returns the timeout duration
func (c *Client) getTimeout() time.Duration {
	return time.Duration(c.config.Timeout) * time.Second
}
```

### 3. adapter.go

**Purpose**: Implement VectorDBClient interface by wrapping Client with LLM support

```go
package elasticsearch

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Elasticsearch client to implement vectordb.VectorDBClient
type Adapter struct {
	*Client
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new Elasticsearch adapter from vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Map vectordb.Config to Elasticsearch Config
	esConfig := &Config{
		CloudID:          extractCloudID(config),
		APIKey:           config.APIKey,
		Addresses:        extractAddresses(config),
		Username:         config.Username,
		Password:         config.Password,
		Timeout:          config.Timeout,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	client, err := NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	// Create LLM client for embeddings (optional)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		var err error
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Warning: Failed to create OpenAI client: %v\n", err)
		}
	}

	return &Adapter{
		Client:    client,
		llmClient: llmClient,
	}, nil
}

// Helper functions to extract config values
func extractCloudID(config *vectordb.Config) string {
	// Check if URL contains Cloud ID pattern
	// Or use dedicated field if added to vectordb.Config
	return "" // Implement based on config structure
}

func extractAddresses(config *vectordb.Config) []string {
	if config.URL != "" {
		return []string{config.URL}
	}
	if config.Address != "" {
		return []string{config.Address}
	}
	return nil
}
```

### 4. factory.go

**Purpose**: Factory registration and validation

```go
package elasticsearch

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Elasticsearch
type Factory struct{}

// NewFactory creates a new Elasticsearch factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Elasticsearch client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeElasticsearchLocal,
		vectordb.VectorDBTypeElasticsearchCloud,
	}
}

// ValidateConfig validates the configuration
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	validTypes := map[vectordb.VectorDBType]bool{
		vectordb.VectorDBTypeElasticsearchLocal: true,
		vectordb.VectorDBTypeElasticsearchCloud: true,
	}
	if !validTypes[config.Type] {
		return vectordb.ErrInvalidConfig(
			fmt.Sprintf("unsupported type: %s", config.Type))
	}

	// Validate required fields
	if config.URL == "" && config.Address == "" && config.APIKey == "" {
		return vectordb.ErrInvalidConfig(
			"URL, Address, or APIKey is required")
	}

	// Validate vector dimensions
	if config.VectorDimensions < 0 {
		return vectordb.ErrInvalidConfig(
			"vector dimensions cannot be negative")
	}

	// Validate similarity metric
	if config.SimilarityMetric != "" {
		validMetrics := map[string]bool{
			"cosine":      true,
			"dot_product": true,
			"l2_norm":     true,
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(
				fmt.Sprintf("invalid similarity metric: %s",
					config.SimilarityMetric))
		}
	}

	return nil
}

// init registers the Elasticsearch factory
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeElasticsearchLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeElasticsearchCloud, factory)
}
```

### 5. collection.go

**Purpose**: Implement CollectionOperations interface

```go
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateCollection creates a new index with vector field mappings
func (c *Client) CreateCollection(ctx context.Context, name string,
	schema *vectordb.CollectionSchema) error {

	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Build index mappings with dense_vector field
	mappings := types.TypeMapping{
		Properties: map[string]types.Property{
			"text": types.NewTextProperty(),
			"content": types.NewTextProperty(),
			"vector_field": types.DenseVectorProperty{
				Type: "dense_vector",
				Dims: c.config.VectorDimensions,
				Index: true,
				Similarity: c.config.SimilarityMetric,
			},
			"metadata": types.NewObjectProperty(),
			"image": types.NewKeywordProperty(),
			"url": types.NewKeywordProperty(),
		},
	}

	// Create index
	res, err := c.client.Indices.Create(name).
		Mappings(&mappings).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if !res.Acknowledged {
		return fmt.Errorf("index creation not acknowledged")
	}

	return nil
}

// DeleteCollection deletes an index
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	_, err := c.client.Indices.Delete(name).Do(ctx)
	return err
}

// ListCollections returns list of indices
func (c *Client) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get all indices
	res, err := c.client.Indices.Get("*").Do(ctx)
	if err != nil {
		return nil, err
	}

	var collections []vectordb.CollectionInfo
	for name := range res {
		// Get document count
		count, _ := c.GetCollectionCount(ctx, name)
		collections = append(collections, vectordb.CollectionInfo{
			Name:  name,
			Count: count,
		})
	}

	return collections, nil
}

// CollectionExists checks if index exists
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	res, err := c.client.Indices.Exists(name).Do(ctx)
	if err != nil {
		return false, err
	}

	return res, nil
}

// GetCollectionCount returns document count in index
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	res, err := c.client.Count().Index(name).Do(ctx)
	if err != nil {
		return 0, err
	}

	return res.Count, nil
}
```

---

## Interface Implementation Mapping

### VectorDBClient Interface

| Method | Implementation File | Elasticsearch API |
|--------|-------------------|-------------------|
| `Health()` | `client.go` | `client.Ping()` |

### CollectionOperations

| Method | Implementation File | Elasticsearch API |
|--------|-------------------|-------------------|
| `CreateCollection()` | `collection.go` | `client.Indices.Create()` |
| `DeleteCollection()` | `collection.go` | `client.Indices.Delete()` |
| `ListCollections()` | `collection.go` | `client.Indices.Get()` |
| `CollectionExists()` | `collection.go` | `client.Indices.Exists()` |
| `GetCollectionCount()` | `collection.go` | `client.Count()` |

### DocumentOperations

| Method | Implementation File | Elasticsearch API |
|--------|-------------------|-------------------|
| `CreateDocument()` | `document.go` | `client.Index()` |
| `CreateDocuments()` | `document.go` | `esutil.BulkIndexer` |
| `GetDocument()` | `document.go` | `client.Get()` |
| `UpdateDocument()` | `document.go` | `client.Update()` |
| `DeleteDocument()` | `document.go` | `client.Delete()` |
| `DeleteDocuments()` | `document.go` | `esutil.BulkIndexer` |
| `DeleteDocumentsByMetadata()` | `document.go` | `client.DeleteByQuery()` |
| `ListDocuments()` | `document.go` | `client.Search()` |

### QueryOperations

| Method | Implementation File | Elasticsearch API |
|--------|-------------------|-------------------|
| `SearchSemantic()` | `query.go` | `client.Search().Knn()` |
| `SearchBM25()` | `query.go` | `client.Search().Query(Match)` |
| `SearchHybrid()` | `query.go` | `client.Search().Query().Knn()` |
| `SearchByMetadata()` | `query.go` | `client.Search().Query(Bool)` |

### SchemaOperations

| Method | Implementation File | Elasticsearch API |
|--------|-------------------|-------------------|
| `GetSchema()` | `schema.go` | `client.Indices.GetMapping()` |
| `UpdateSchema()` | `schema.go` | `client.Indices.PutMapping()` |
| `GetDefaultSchema()` | `schema.go` | Local generation |
| `ValidateSchema()` | `schema.go` | Local validation |

---

## Configuration Design

### Environment Variables

```bash
# Elastic Cloud (preferred for cloud deployments)
ELASTICSEARCH_CLOUD_ID=cluster-name:base64-encoded-cluster-id
ELASTICSEARCH_API_KEY=base64-encoded-api-key

# Self-hosted (for local/on-prem deployments)
ELASTICSEARCH_ADDRESSES=https://localhost:9200,https://localhost:9201
ELASTICSEARCH_USERNAME=elastic
ELASTICSEARCH_PASSWORD=changeme

# Optional
ELASTICSEARCH_CERT_FINGERPRINT=sha256:hex-encoded-fingerprint
ELASTICSEARCH_TIMEOUT=10
ELASTICSEARCH_MAX_RETRIES=3
ELASTICSEARCH_VECTOR_DIMENSIONS=1536
ELASTICSEARCH_SIMILARITY_METRIC=cosine

# Required for embeddings
OPENAI_API_KEY=sk-...
```

### VectorDB Config Integration

Add to `src/pkg/vectordb/factory.go`:

```go
const (
	// ... existing types ...
	VectorDBTypeElasticsearchLocal VectorDBType = "elasticsearch-local"
	VectorDBTypeElasticsearchCloud VectorDBType = "elasticsearch-cloud"
)
```

Add to `src/cmd/utils/vectordb_client.go`:

```go
import (
	// ... existing imports ...
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/elasticsearch"
)
```

---

## Implementation Phases

### Phase 1: Core Infrastructure (Day 1)
- [ ] Create package structure
- [ ] Implement `config.go` with validation
- [ ] Implement `client.go` with connection and health checks
- [ ] Implement `adapter.go` with VectorDBClient wrapper
- [ ] Implement `factory.go` with registration
- [ ] Add VectorDBType constants to factory.go
- [ ] Add import to vectordb_client.go

**Deliverable**: Can create client, connect to Elasticsearch, check health

### Phase 2: Collection Operations (Day 2)
- [ ] Implement `collection.go`:
  - `CreateCollection()` - Create index with vector mappings
  - `DeleteCollection()` - Delete index
  - `ListCollections()` - List all indices
  - `CollectionExists()` - Check index existence
  - `GetCollectionCount()` - Get document count

**Deliverable**: Full collection management working

### Phase 3: Document Operations (Day 3 Morning)
- [ ] Implement `document.go`:
  - `CreateDocument()` - Index single document
  - `CreateDocuments()` - Bulk indexing
  - `GetDocument()` - Fetch by ID
  - `UpdateDocument()` - Update by ID
  - `DeleteDocument()` - Delete by ID
  - `DeleteDocuments()` - Bulk delete
  - `DeleteDocumentsByMetadata()` - Delete by query
  - `ListDocuments()` - Paginated list

**Deliverable**: Full document CRUD working

### Phase 4: Query Operations (Day 3 Afternoon)
- [ ] Implement `query.go`:
  - `SearchSemantic()` - kNN vector search
  - `SearchBM25()` - Full-text search
  - `SearchHybrid()` - Combined BM25 + kNN
  - `SearchByMetadata()` - Metadata filtering

**Deliverable**: All search types working

### Phase 5: Schema Operations (Day 3 Evening)
- [ ] Implement `schema.go`:
  - `GetSchema()` - Get index mappings
  - `UpdateSchema()` - Update mappings
  - `GetDefaultSchema()` - Generate default schema
  - `ValidateSchema()` - Validate schema

**Deliverable**: Schema management complete

### Phase 6: Integration Tests (Day 4)
- [ ] Create `tests/elasticsearch_integration_test.go`
- [ ] Test health checks
- [ ] Test collection operations
- [ ] Test document CRUD
- [ ] Test batch operations
- [ ] Test vector search
- [ ] Test BM25 search
- [ ] Test hybrid search
- [ ] Test metadata filtering
- [ ] Test schema operations

**Deliverable**: Full test coverage, all tests passing

### Phase 7: Documentation (Day 5)
- [ ] Write `docs/elasticsearch/README.md` - Overview
- [ ] Write `docs/elasticsearch/SETUP.md` - General setup
- [ ] Write `docs/elasticsearch/LOCAL_SETUP.md` - Docker setup
- [ ] Write `docs/elasticsearch/CLOUD_SETUP.md` - Elastic Cloud setup
- [ ] Update `docs/VDB_SUPPORT_MATRIX.md` - Add Elasticsearch row
- [ ] Update `README.md` - Add to main table
- [ ] Update `CHANGELOG.md` - v0.8.0 entry

**Deliverable**: Complete documentation

---

## Code Structure Examples

### Document Structure

```go
// elasticsearchDocument represents the Elasticsearch document structure
type elasticsearchDocument struct {
	ID          string                 `json:"id"`
	Text        string                 `json:"text"`
	Content     string                 `json:"content"`
	Image       string                 `json:"image,omitempty"`
	ImageData   string                 `json:"image_data,omitempty"`
	URL         string                 `json:"url,omitempty"`
	VectorField []float64              `json:"vector_field"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
```

### Vector Search Example

```go
func (c *Client) SearchSemantic(ctx context.Context, collectionName, query string,
	options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {

	// Generate embedding for query
	embedding, err := c.llmClient.CreateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	// Perform kNN search
	k := options.TopK
	numCandidates := k * 10 // 10x candidates for better recall

	res, err := c.client.Search().
		Index(collectionName).
		Knn(types.KnnSearch{
			Field:         "vector_field",
			K:             &k,
			NumCandidates: &numCandidates,
			QueryVector:   embedding,
		}).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("kNN search failed: %w", err)
	}

	// Convert results to QueryResult
	var results []*vectordb.QueryResult
	for _, hit := range res.Hits.Hits {
		doc := parseDocument(hit.Source_)
		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    *hit.Score_,
		})
	}

	return results, nil
}
```

---

## Dependencies

### Go Modules

```go
require (
	github.com/elastic/go-elasticsearch/v9 v9.2.1
)
```

### Import Structure

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/esutil"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)
```

---

## Testing Strategy

### Local Testing Setup

```bash
# Run Elasticsearch via Docker
docker run -d \
  --name elasticsearch \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:8.15.0

# Set environment variables
export ELASTICSEARCH_ADDRESSES=http://localhost:9200
export ELASTICSEARCH_USERNAME=elastic
export ELASTICSEARCH_PASSWORD=changeme
export OPENAI_API_KEY=sk-...

# Run tests
go test -v -timeout=10m -tags=integration \
  -run="TestElasticsearchIntegration" ./tests
```

### Cloud Testing Setup

```bash
# Get credentials from Elastic Cloud console
export ELASTICSEARCH_CLOUD_ID=my-deployment:ABC123...
export ELASTICSEARCH_API_KEY=BASE64KEY...
export OPENAI_API_KEY=sk-...

# Run tests
go test -v -timeout=10m -tags=integration \
  -run="TestElasticsearchIntegration" ./tests
```

---

## Next Steps

1. ✅ Research complete (RESEARCH.md)
2. ✅ Architecture planned (this document)
3. ➡️ Begin Phase 1: Core Infrastructure
4. ➡️ Continue through Phases 2-7
5. ➡️ Final integration and release

**Estimated Timeline**: 5 days total (Phases 1-7)

---

**Architecture Designed**: 2025-12-12
**Ready for Implementation**: Yes
**Next Action**: Create core infrastructure files
