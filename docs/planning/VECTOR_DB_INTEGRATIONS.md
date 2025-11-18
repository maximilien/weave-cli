# Vector Database Integration Planning

Planning document for integrating Milvus, Qdrant, Chroma, Redis, MongoDB, and Pinecone
vector databases into Weave CLI.

**Status**: Planning Phase
**Target**: v0.4.0 - v0.9.0 (path to v1.0.0)
**Last Updated**: 2025-11-18

## Overview

This document outlines the planned integration of six major vector databases to
achieve v1.0.0 release:

1. **Milvus** - Open-source, cloud-native vector database
2. **Qdrant** - Open-source vector search engine with gRPC
3. **Chroma** - Open-source embedding database for AI applications
4. **Redis** - In-memory database with RediSearch vector capabilities
5. **MongoDB** - Document database with Atlas Vector Search
6. **Pinecone** - Managed vector database with serverless option

These were selected as they represent the top native vector databases outside of
Weaviate, each with strong Go SDK support and active communities.

**Priority Order**: Milvus → Qdrant → Chroma → Redis → MongoDB → Pinecone

**v1.0.0 Release Goal**: Complete all six integrations with production-ready
support, comprehensive documentation, and demos for each database.

---

## 1. Milvus Integration

### Overview

- **License**: Apache 2.0 (Open Source)
- **Governance**: LF AI & Data Foundation
- **Primary Contributor**: Zilliz
- **Go SDK**: `github.com/milvus-io/milvus-sdk-go`
- **Latest Version**: v2.6.1 (compatible with Milvus 2.6.x)
- **Target Milestone**: v0.4.0

### Key Features

#### Core Capabilities

- **Distributed Architecture**: Separates compute and storage for horizontal
  scaling
- **Hybrid Search**: Native BM25 full-text search + vector similarity search
- **Sparse + Dense Vectors**: Can store both in same collection
- **Geospatial Support**: GIS data types and functions (ST_EQUALS, etc.)
- **Array of Vectors**: Multiple vectors per entity via Struct data type
- **Security**: Mandatory auth, TLS encryption, RBAC

#### Advanced Features

- **Multiple Vector Types**:
  - Dense vectors (standard embeddings)
  - Sparse vectors (BM25, SPLADE, BGE-M3)
  - Geospatial vectors
- **Reranking**: Built-in support for multi-request result reranking
- **Struct Data Type**: Organize multiple related fields per entity
- **CDC (Change Data Capture)**: Enhanced data sync capabilities

### API Design

#### Configuration

```go
type MilvusConfig struct {
    Address     string   // e.g., "localhost:19530"
    Username    string   // Optional
    Password    string   // Optional
    Database    string   // Default: "default"
    TLS         bool     // Enable TLS
    Timeout     int      // Connection timeout (seconds)
}
```

#### Collection Schema

Milvus uses explicit schemas with strongly-typed fields:

```go
type MilvusCollection struct {
    Name        string
    Description string
    Fields      []Field  // Explicit field definitions
    VectorField Field    // Primary vector field
    AutoID      bool     // Auto-generate IDs
    ShardNum    int      // Number of shards
}
```

### Implementation Considerations

#### Pros

- ✅ **Open Source**: Apache 2.0 license, no vendor lock-in
- ✅ **Feature Rich**: BM25, geospatial, hybrid search out of box
- ✅ **Strong Go SDK**: Official SDK with good documentation
- ✅ **Scalable**: Distributed architecture for production workloads
- ✅ **Active Development**: Regular releases, LF AI governance

#### Challenges

- ⚠️ **Schema Complexity**: Requires explicit field definitions (vs. Weaviate's
  auto-schema)
- ⚠️ **Setup**: Self-hosted requires Docker/K8s; cloud option (Zilliz) is paid
- ⚠️ **Learning Curve**: More complex than Weaviate for basic use cases
- ⚠️ **Multiple Vector Types**: Need to handle dense, sparse, and geospatial
  separately

#### Migration Path

1. Create `pkg/milvus/client.go` implementing `VectorDB` interface
2. Add Milvus-specific schema handling
3. Map Weave's simple schema to Milvus explicit schema
4. Implement BM25 search mapping
5. Add Milvus to `VectorDBType` enum

### Testing Strategy

- **Local**: Docker Compose with Milvus standalone
- **CI/CD**: Use Milvus Docker image in GitHub Actions
- **Integration**: Test against Milvus 2.6.x and Zilliz Cloud

---

## 2. Pinecone Integration

### Overview

- **License**: Proprietary (Managed Service)
- **Company**: Pinecone Systems Inc.
- **Go SDK**: `github.com/pinecone-io/go-pinecone`
- **Pricing Model**: Freemium + Pay-as-you-go
- **Target Milestone**: v0.5.0

### Key Features

#### Core Capabilities

- **Serverless**: Best-in-class performance, 50x lower cost at scale
- **Managed**: Fully hosted, no infrastructure management
- **Hybrid Search**: Vector search + keyword boosting
- **Live Updates**: Real-time index updates
- **Multi-Cloud**: AWS, Azure, GCP support
- **Namespaces**: Logical partitions within indexes (100 per index on free tier)

#### Advanced Features

- **Pinecone Inference**: Hosted embedding and reranking models
- **Pinecone Assistant**: Production-grade chat/agent applications
- **Filtering**: Metadata filtering during search
- **Sparse-Dense Hybrid**: Support for both vector types

### Pricing Tiers (2025)

#### Starter (Free)

- Up to 5 indexes (100 namespaces each)
- 2 GB storage
- 2M write units/month
- 1M read units/month
- Limited to AWS us-east-1
- 1 project, 2 users max

#### Standard

- $50/month minimum + pay-as-you-go
- $15/month usage credits
- All regions (AWS, Azure, GCP)
- Unlimited projects and users

#### Enterprise

- Custom pricing
- SLA guarantees
- Dedicated support

### API Design

#### Configuration

```go
type PineconeConfig struct {
    APIKey      string   // Required
    Environment string   // e.g., "us-east-1-aws"
    ProjectID   string   // Optional for organization accounts
    Cloud       string   // aws, gcp, azure
    Region      string   // Specific region
}
```

#### Index Management

Pinecone uses "indexes" (equivalent to collections):

```go
type PineconeIndex struct {
    Name       string
    Dimension  int          // Vector dimension (e.g., 1536 for text-embedding-ada-002)
    Metric     string       // cosine, euclidean, dotproduct
    PodType    string       // For pod-based indexes
    Serverless bool         // Use serverless architecture
    Namespaces []string     // Logical partitions
}
```

### Implementation Considerations

#### Pros

- ✅ **Fully Managed**: No infrastructure to manage
- ✅ **Generous Free Tier**: 2GB storage, good for testing/demos
- ✅ **Performance**: Optimized for speed and scale
- ✅ **Simple API**: Easy to use, well-documented
- ✅ **Multi-Cloud**: Flexibility in deployment regions

#### Challenges

- ⚠️ **Proprietary**: Vendor lock-in, not open source
- ⚠️ **Cost**: Can be expensive at scale ($50/month minimum for Standard)
- ⚠️ **API Key Required**: No local/self-hosted option for testing
- ⚠️ **Namespace Limits**: 100 namespaces per index on free tier
- ⚠️ **Different Terminology**: "Index" vs "Collection", "namespace" vs
  "tenant"

#### Migration Path

1. Create `pkg/pinecone/client.go` implementing `VectorDB` interface
2. Map Weave collections → Pinecone indexes
3. Handle namespace logic (could map to collection variants)
4. Implement metadata filtering
5. Add Pinecone to `VectorDBType` enum
6. Update config to support API key authentication

### Testing Strategy

- **Free Tier**: Use Starter plan for development and CI/CD
- **Namespaces**: Test multi-tenancy via namespaces
- **Regions**: Test AWS us-east-1 initially
- **Integration**: Mock client for unit tests, real API for integration tests

---

## 3. Chroma Integration

### Overview

- **License**: Apache 2.0 (Open Source)
- **Company**: Chroma (backed by Astasia Myers, Quiet Capital)
- **Go SDK**: `github.com/amikos-tech/chroma-go`
- **Deployment**: Self-hosted, cloud-native (embedding database)
- **Latest Version**: v0.5.0+ (actively maintained)
- **Target Milestone**: v0.6.0

### Key Features

#### Core Capabilities

- **Embedding-First Design**: Built specifically for AI applications with embeddings
- **Simple API**: Pythonic design with minimal configuration
- **Collections**: First-class collection management with metadata
- **In-Memory + Persistent**: SQLite backend for persistence, optional DuckDB
- **Multi-Modal**: Text, images, and custom embeddings
- **Built-in Embeddings**: Support for OpenAI, Cohere, Sentence Transformers
- **Filtering**: Metadata filtering with where clauses

#### Advanced Features

- **Multi-Tenancy**: Database-level isolation
- **Embedding Functions**: Pluggable embedding generation
- **Distance Metrics**: L2, cosine, inner product
- **Query Results**: Returns documents + metadata + distances
- **Update/Upsert**: Flexible document updates
- **HNSW Index**: Fast approximate nearest neighbor search

### API Design

#### Configuration

```go
type ChromaConfig struct {
    URL         string   // e.g., "http://localhost:8000"
    Tenant      string   // Multi-tenancy support (default: "default_tenant")
    Database    string   // Database name (default: "default_database")
    APIKey      string   // Optional for cloud deployments
    Timeout     int      // Request timeout (seconds)
}
```

#### Collection Schema

Chroma uses a simple collection model with automatic schema:

```go
type ChromaCollection struct {
    Name            string
    Metadata        map[string]interface{}  // Collection-level metadata
    EmbeddingFunc   string                  // Optional embedding function
    DistanceMetric  string                  // l2, cosine, ip
}
```

#### Document Model

```go
type ChromaDocument struct {
    ID          string                    // Document ID
    Embedding   []float32                 // Vector embedding
    Document    string                    // Text content
    Metadata    map[string]interface{}    // Document metadata
}
```

### Implementation Considerations

#### Pros

- ✅ **Open Source**: Apache 2.0, fully open and self-hostable
- ✅ **Simple API**: Minimal configuration, easy to get started
- ✅ **AI-First**: Designed for embedding-based AI applications
- ✅ **Built-in Embeddings**: Supports multiple embedding providers
- ✅ **Lightweight**: Easy to run locally (SQLite backend)
- ✅ **Multi-Modal**: Text, images, custom embeddings
- ✅ **Active Community**: Growing ecosystem, regular updates
- ✅ **Go SDK Available**: Community-maintained Go client

#### Challenges

- ⚠️ **Community Go SDK**: Not official, maintained by community (amikos-tech)
- ⚠️ **Limited Production Use**: Newer compared to Milvus/Qdrant
- ⚠️ **No Native BM25**: Focus on vector search only
- ⚠️ **Simpler Features**: Less advanced than Milvus (no geospatial, sparse vectors)
- ⚠️ **SQLite Limitations**: Default backend may not scale to millions of vectors

#### Migration Path

1. Create `pkg/chroma/client.go` implementing `VectorDB` interface
2. Add Chroma Go SDK: `github.com/amikos-tech/chroma-go`
3. Map Weave collections → Chroma collections
4. Handle metadata filtering (where clauses)
5. Implement embedding generation (or use Chroma's built-in)
6. Add Chroma to `VectorDBType` enum
7. Support multi-tenancy (tenant/database isolation)

### Testing Strategy

- **Local**: Run Chroma server locally (Docker or binary)
- **CI/CD**: Use Chroma Docker image in GitHub Actions
- **Integration**: Test against latest Chroma release
- **Embedding**: Test with OpenAI embeddings and custom embeddings

---

## 4. Qdrant Integration

### Overview

- **License**: Apache 2.0 (Open Source)
- **Company**: Qdrant Solutions GmbH
- **Go SDK**: `github.com/qdrant/go-client`
- **Deployment**: Self-hosted or Qdrant Cloud
- **Latest Update**: November 14, 2025 (actively maintained)
- **Target Milestone**: v0.6.0

### Key Features

#### Core Capabilities

- **gRPC API**: High-performance communication
- **Payload Filtering**: JSON-serializable payloads with complex filters
- **Collections**: First-class collection management
- **Points**: Fundamental data entity (ID + vector + payload)
- **Snapshots**: Backup and restore capabilities
- **Clustering**: Distributed deployment support

#### Advanced Features

- **Multiple Vectors per Point**: Store multiple embeddings per entity
- **Quantization**: Reduce memory footprint
- **HNSW Index**: Hierarchical Navigable Small World graphs
- **Recommendations**: Find similar items based on examples
- **Batch Operations**: Efficient bulk upserts and updates

### API Design

#### Configuration

```go
type QdrantConfig struct {
    URL       string   // e.g., "localhost:6334"
    APIKey    string   // Optional (for Qdrant Cloud)
    GRPC      bool     // Use gRPC (default) vs HTTP
    TLS       bool     // Enable TLS
    Timeout   int      // Request timeout (seconds)
}
```

#### Collection Schema

```go
type QdrantCollection struct {
    Name            string
    VectorSize      int           // Dimension of vectors
    Distance        string        // Cosine, Euclid, Dot
    OnDiskPayload   bool          // Store payload on disk vs memory
    HNSWConfig      *HNSWConfig   // Optional HNSW parameters
    QuantizationConfig *QuantConfig // Optional quantization
}
```

### Implementation Considerations

#### Pros

- ✅ **Open Source**: Apache 2.0, self-hostable
- ✅ **Official Go Client**: Well-maintained gRPC client
- ✅ **Performance**: Optimized with HNSW + quantization
- ✅ **Flexible Payloads**: JSON-serializable, complex filtering
- ✅ **Cloud Option**: Managed Qdrant Cloud available
- ✅ **Active Development**: Regular updates (as of Nov 2025)

#### Challenges

- ⚠️ **Setup**: Self-hosted requires infrastructure
- ⚠️ **gRPC Dependency**: Need to handle protobuf/gRPC in build
- ⚠️ **Point-Based Model**: Different abstraction than document-based
  (Weaviate/Supabase)
- ⚠️ **Learning Curve**: HNSW tuning, quantization config for optimization

#### Migration Path

1. Create `pkg/qdrant/client.go` implementing `VectorDB` interface
2. Map Weave documents → Qdrant points
3. Handle payload serialization (JSON)
4. Implement filtering with Qdrant's filter syntax
5. Add Qdrant to `VectorDBType` enum
6. Add gRPC dependencies to go.mod

### Testing Strategy

- **Local**: Docker with Qdrant image
- **CI/CD**: Use Qdrant Docker in GitHub Actions
- **gRPC**: Test both gRPC and HTTP APIs
- **Integration**: Test against latest Qdrant release

---

## 5. Redis Integration

### Overview

- **License**: RSALv2 / SSPLv1 / AGPLv3 (tri-license, Redis 8+)
- **Company**: Redis Ltd.
- **Go SDK**: `github.com/redis/go-redis` + RediSearch commands
- **Module**: RediSearch (integrated in Redis 8+)
- **Latest Version**: Redis 8.0 (RediSearch is now built-in)
- **Target Milestone**: v0.6.0

### Key Features

#### Core Capabilities

- **In-Memory Speed**: Extremely fast vector search (sub-millisecond latency)
- **Hybrid Search**: Combine vector search with full-text, tag, geo, numeric
  filters
- **RediSearch Integration**: Vector similarity is part of Redis Query Engine
- **No Separate Install**: Starting Redis 8, RediSearch is built-in
- **HNSW + FLAT Indexing**: Choose between accuracy (HNSW) or simplicity (FLAT)
- **Distance Metrics**: L2, IP (Inner Product), Cosine

#### Advanced Features

- **KNN Search**: K-nearest neighbors with configurable K
- **Range Search**: Find all vectors within distance threshold
- **Filtering**: Pre-filter with Redis fields before vector search
- **Real-time Updates**: Add/update vectors without index rebuilding
- **Clustering**: Redis Cluster support for horizontal scaling

### API Design

#### Configuration

```go
type RedisConfig struct {
    Address  string   // e.g., "localhost:6379"
    Password string   // Optional
    DB       int      // Database number (0-15)
    TLS      bool     // Enable TLS
    Cluster  bool     // Use Redis Cluster
    Nodes    []string // Cluster nodes (if Cluster=true)
}
```

#### Index Schema

Redis uses hash-based documents with vector fields:

```go
type RedisIndex struct {
    Name         string
    Prefix       string        // Hash key prefix (e.g., "doc:")
    VectorField  string        // Field name for vector
    Algorithm    string        // FLAT or HNSW
    Distance     string        // L2, IP, COSINE
    Dimension    int           // Vector dimension
    InitialCap   int           // Initial capacity
    M            int           // HNSW: max connections per node
    EFConstruct  int           // HNSW: construction time/accuracy tradeoff
}
```

### Implementation Considerations

#### Pros

- ✅ **Blazing Fast**: In-memory, sub-millisecond vector search
- ✅ **Simple Setup**: Redis is widely deployed, easy to run locally/cloud
- ✅ **Hybrid Search**: Best-in-class full-text + vector combination
- ✅ **Real-time**: No index rebuild needed for updates
- ✅ **Familiar**: Many teams already use Redis for caching
- ✅ **Built-in (Redis 8+)**: No modules to install separately

#### Challenges

- ⚠️ **License Change**: Redis 8 uses RSALv2/SSPLv1/AGPLv3 (not pure
  open-source)
  - Note: Valkey (Redis fork) may be alternative if license is concern
- ⚠️ **Memory Cost**: All data in-memory (expensive for large datasets)
- ⚠️ **Persistence**: Need to configure RDB/AOF for durability
- ⚠️ **Different Model**: Hash-based documents vs collection-based
- ⚠️ **Go SDK**: Need to use raw Redis commands (no high-level vector client
  yet)

#### Migration Path

1. Create `pkg/redis/client.go` implementing `VectorDB` interface
2. Use `go-redis` for Redis connection
3. Implement FT.CREATE for index creation (RediSearch syntax)
4. Map Weave documents → Redis hashes with vector field
5. Implement FT.SEARCH with KNN query syntax
6. Add Redis to `VectorDBType` enum
7. Handle Redis Cluster for scale-out scenarios

### Testing Strategy

- **Local**: Redis Docker image (redis:latest or redis:8.0)
- **CI/CD**: Use Redis Docker in GitHub Actions
- **Cluster**: Test with redis-cluster Docker compose
- **Integration**: Test FLAT and HNSW indexing strategies

---

## 6. MongoDB Atlas Vector Search Integration

### Overview

**Target Milestone**: v0.7.0 (December 2025)
**Estimated Effort**: ~1-2 weeks
**License**: Server Side Public License (SSPL) v1 (commercial-friendly with restrictions)
**Go SDK**: Official MongoDB Go Driver (`go.mongodb.org/mongo-driver`)

MongoDB Atlas Vector Search brings vector capabilities to the popular document database, making it ideal for users who:
- Already use MongoDB for their application data
- Want to combine vector search with rich document queries
- Need seamless integration with existing MongoDB infrastructure
- Want managed cloud service with generous free tier (M0 cluster)

### Key Features

**Core Capabilities**:
- ✅ **Vector Search**: Cosine similarity, Euclidean distance, dot product
- ✅ **Hybrid Search**: Combine vector search with MongoDB queries
- ✅ **Pre-filtering**: Filter documents before vector search (more efficient)
- ✅ **Scalar Quantization**: Reduce memory by 96% with minimal accuracy loss
- ✅ **Binary Quantization**: Extreme compression for large-scale deployments
- ✅ **Dimension Support**: Up to 8192 dimensions (covers most embedding models)
- ✅ **Index Types**: HNSW (approximate) and exact nearest neighbor

**Advanced Capabilities**:
- ✅ **Document Model**: Native JSON documents (no schema required)
- ✅ **Aggregation Pipeline**: Combine vector search with MongoDB aggregations
- ✅ **Atlas Search**: Full-text search integration (separate from vector search)
- ✅ **Change Streams**: Real-time notifications for document changes
- ✅ **RBAC**: Database-level and collection-level access control
- ✅ **Atlas UI**: Web-based query and index management

**Limitations**:
- ⚠️ **8192 Dimension Limit**: Lower than Milvus (32768) or Qdrant (65536)
- ⚠️ **Atlas Only**: Vector search only available on Atlas (cloud/self-hosted Atlas)
- ❌ **No BM25**: Must use separate Atlas Search for full-text (not integrated)
- ⚠️ **Managed Indexes**: Index creation is asynchronous (can take time)

### API Design Mockup

```go
// pkg/mongodb/client.go
package mongodb

import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
)

type MongoDBClient struct {
    client     *mongo.Client
    database   *mongo.Database
    config     MongoDBConfig
}

type MongoDBConfig struct {
    ConnectionString string // mongodb+srv://...
    DatabaseName     string // default: "weave"
    Timeout          int    // connection timeout in seconds
}

// Vector index definition
type VectorIndexDefinition struct {
    Type       string  // "vector" for Atlas Vector Search
    Path       string  // field path (e.g., "embedding")
    Dimensions int     // embedding dimensions
    Similarity string  // "cosine", "euclidean", "dotProduct"
}

// Implement VectorDB interface
func (m *MongoDBClient) CreateCollection(ctx context.Context, name string, schema CollectionSchema) error {
    // 1. Create MongoDB collection
    err := m.database.CreateCollection(ctx, name)

    // 2. Create vector search index (asynchronous)
    indexModel := mongo.IndexModel{
        Keys: bson.D{
            {Key: "embedding", Value: "vectorSearch"},
        },
        Options: options.Index().
            SetName(name + "_vector_index").
            SetVectorSearchOptions(bson.D{
                {Key: "type", Value: "vector"},
                {Key: "numDimensions", Value: schema.VectorDimension},
                {Key: "similarity", Value: "cosine"},
            }),
    }

    _, err = m.database.Collection(name).Indexes().CreateOne(ctx, indexModel)
    return err
}

// Hybrid search: pre-filter + vector search
func (m *MongoDBClient) SearchHybrid(ctx context.Context, collection string, query string, filter map[string]interface{}, limit int) ([]Document, error) {
    pipeline := mongo.Pipeline{
        // Stage 1: Pre-filter (optional)
        bson.D{{Key: "$match", Value: filter}},

        // Stage 2: Vector search
        bson.D{{Key: "$vectorSearch", Value: bson.D{
            {Key: "index", Value: collection + "_vector_index"},
            {Key: "queryVector", Value: embedQuery(query)}, // get embedding
            {Key: "path", Value: "embedding"},
            {Key: "numCandidates", Value: limit * 10},
            {Key: "limit", Value: limit},
        }}},

        // Stage 3: Project fields
        bson.D{{Key: "$project", Value: bson.D{
            {Key: "_id", Value: 1},
            {Key: "content", Value: 1},
            {Key: "metadata", Value: 1},
            {Key: "score", Value: bson.D{{Key: "$meta", Value: "vectorSearchScore"}}},
        }}},
    }

    cursor, err := m.database.Collection(collection).Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }

    var results []Document
    err = cursor.All(ctx, &results)
    return results, err
}
```

**Example CLI Usage**:
```bash
# Configure MongoDB Atlas
export MONGODB_CONNECTION_STRING="mongodb+srv://user:pass@cluster.mongodb.net"
export MONGODB_DATABASE="weave"

# Create collection with MongoDB backend
weave cols create MyDocs --vdb mongodb --embedding text-embedding-3-small

# Search with pre-filtering
weave search semantic MyDocs "machine learning" --vdb mongodb --filter "metadata.category=tech"

# Hybrid search (if we add Atlas Search integration later)
weave search hybrid MyDocs "AI trends" --vdb mongodb --alpha 0.7
```

### Implementation Considerations

#### Pros
- ✅ **Familiar**: Many developers already know MongoDB
- ✅ **Document Model**: No rigid schema required
- ✅ **Managed**: Atlas handles infrastructure, backups, scaling
- ✅ **Free Tier**: M0 cluster (512MB storage, shared resources)
- ✅ **Pre-filtering**: Efficient filtering before vector search
- ✅ **Quantization**: Scalar and binary quantization built-in
- ✅ **Official SDK**: Well-maintained Go driver

#### Challenges
- ⚠️ **Atlas Only**: Vector search requires Atlas (not community MongoDB)
- ⚠️ **8192 Dimension Limit**: May not support future large embedding models
- ⚠️ **No BM25**: Need separate Atlas Search for full-text (not integrated)
- ⚠️ **Async Index Creation**: Index creation can take time (need to poll status)
- ⚠️ **Different Model**: Need to map Weave's collection model to MongoDB documents

#### Migration Path

1. Add MongoDB Go driver: `go.mongodb.org/mongo-driver`
2. Create `pkg/mongodb/client.go` implementing `VectorDB` interface
3. Map Weave collections → MongoDB collections with vector indexes
4. Handle embedding field explicitly (MongoDB doesn't auto-generate)
5. Implement aggregation pipeline for hybrid search
6. Add MongoDB to `VectorDBType` enum
7. Handle async index creation (poll until ready)
8. Document Atlas setup requirements

### Testing Strategy

- **Local**: Use MongoDB Atlas M0 free tier for development
- **CI/CD**: Use test cluster or MongoDB Community (limited - no vector search)
- **Integration**:
  - Test vector search with real embeddings
  - Test pre-filtering with metadata
  - Test scalar quantization
  - Verify 8192 dimension limit
- **Mocking**: Mock `mongo.Client` for unit tests

---

## Comparison Matrix

| Feature | Milvus | Qdrant | Chroma | Redis | MongoDB | Pinecone | Weaviate | Supabase |
|---------|--------|--------|--------|-------|---------|----------|----------|----------|
| **License** | Apache 2.0 | Apache 2.0 | Apache 2.0 | RSALv2/SSPL/AGPL | SSPL v1 | Proprietary | BSD-3 | PostgreSQL |
| **Hosting** | Self/Cloud | Self/Cloud | Self/Cloud | Self/Cloud | Atlas (Managed) | Managed | Self/Cloud | Self/Cloud |
| **Free Tier** | Self-hosted | Self-hosted | Self-hosted | Self-hosted | 512MB (M0) | 2GB/2M writes | Self-hosted | 500MB |
| **Go SDK** | Official | Official | Community | go-redis | Official | Official | Official | via pgx |
| **BM25 Search** | ✅ Native | ❌ | ❌ | ✅ Native (FT) | ⚠️ Atlas Search | ⚠️ Keyword boost | ⚠️ Via plugin | ✅ Native |
| **Hybrid Search** | ✅ | ❌ | ❌ | ✅ Best-in-class | ✅ Pre-filter | ✅ | ✅ | ✅ |
| **Geospatial** | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ (PostGIS) |
| **Multi-Vector** | ✅ | ✅ | ⚠️ Multi-modal | ⚠️ Via metadata | ❌ | ⚠️ Metadata | ⚠️ Cross-ref | ❌ |
| **RBAC** | ✅ | ⚠️ Limited | ⚠️ API Key only | ⚠️ ACL only | ✅ | ✅ | ✅ | ✅ (RLS) |
| **Max Dimensions** | 32768 | 65536 | No limit | No limit | **8192** | 20000 | 65535 | 2000 (pgvector) |
| **Performance** | High | High | Medium | **Extreme** (in-mem) | High | High | High | Medium |
| **Memory Cost** | Medium | Medium | **Low** (SQLite) | **High** (all in-mem) | Low | N/A (managed) | Medium | Low |
| **Complexity** | High | Medium | **Low** (simple) | **Low** (familiar) | **Low** (familiar) | Low | Medium | Medium |
| **AI-First** | ✅ | ✅ | **✅ Built-in** | ❌ | ❌ | ✅ | ✅ | ⚠️ Extension |

---

## Implementation Roadmap

### Phase 1: Milvus (v0.4.0) - ~2-3 weeks

**Rationale**: Open source, feature-rich, good learning opportunity for
handling complex schemas and BM25 integration.

**Tasks**:
- [ ] Add Milvus Go SDK dependency (`github.com/milvus-io/milvus-sdk-go`)
- [ ] Implement `MilvusClient` with `VectorDB` interface
- [ ] Handle schema mapping (Weave → Milvus explicit schemas)
- [ ] Implement collection CRUD
- [ ] Implement document CRUD
- [ ] Add BM25 search support (native)
- [ ] Add hybrid search (sparse + dense)
- [ ] Add geospatial search support
- [ ] Write unit tests
- [ ] Write integration tests (Docker-based)
- [ ] Update documentation
- [ ] Create Milvus demo script

**Success Criteria**:
- All VectorDB interface methods implemented
- BM25 and hybrid search working
- Integration tests passing with Milvus Docker

### Phase 2: Qdrant (v0.5.0) - ~1-2 weeks

**Rationale**: Open source alternative to Pinecone with advanced features, good
for users who want performance + self-hosting.

**Tasks**:
- [ ] Add Qdrant Go SDK dependency (`github.com/qdrant/go-client`)
- [ ] Add gRPC dependencies
- [ ] Implement `QdrantClient` with `VectorDB` interface
- [ ] Map documents → points (point-based model)
- [ ] Handle payload JSON serialization
- [ ] Implement CRUD operations
- [ ] Add filtering support (flexible JSON payloads)
- [ ] Add HNSW + quantization support
- [ ] Write unit tests
- [ ] Write integration tests (Docker-based)
- [ ] Update documentation
- [ ] Create Qdrant demo script

**Success Criteria**:
- All VectorDB interface methods implemented
- Point-based model working correctly
- Integration tests passing with Qdrant Docker

### Phase 3: Chroma (v0.6.0) - ~1 week

**Rationale**: Simple, AI-first embedding database with growing community,
perfect for developers building AI applications who want minimal setup.

**Tasks**:
- [ ] Add Chroma Go SDK dependency (`github.com/amikos-tech/chroma-go`)
- [ ] Implement `ChromaClient` with `VectorDB` interface
- [ ] Map Weave collections → Chroma collections
- [ ] Handle metadata filtering (where clauses)
- [ ] Implement embedding generation (OpenAI integration)
- [ ] Support multi-tenancy (tenant/database)
- [ ] Implement CRUD operations
- [ ] Write unit tests
- [ ] Write integration tests (Docker-based)
- [ ] Update documentation
- [ ] Create Chroma demo script

**Success Criteria**:
- All VectorDB interface methods implemented
- Metadata filtering working correctly
- Integration tests passing with Chroma Docker

### Phase 4: Redis (v0.7.0) - ~1-2 weeks

**Rationale**: In-memory speed for real-time applications, familiar to many
developers, best-in-class hybrid search.

**Tasks**:
- [ ] Add Redis Go SDK dependency (`github.com/redis/go-redis`)
- [ ] Implement `RedisClient` with `VectorDB` interface
- [ ] Implement FT.CREATE for vector indexes (RediSearch)
- [ ] Map collections → Redis indexes
- [ ] Map documents → Redis hashes
- [ ] Implement FT.SEARCH with KNN queries
- [ ] Add hybrid search (vector + full-text)
- [ ] Add geospatial search support
- [ ] Write unit tests
- [ ] Write integration tests (Docker-based)
- [ ] Update documentation (including license options)
- [ ] Create Redis demo script

**Success Criteria**:
- All VectorDB interface methods implemented
- Hybrid search (vector + FT) working
- Integration tests passing with Redis Docker

### Phase 5: MongoDB (v0.8.0) - ~1-2 weeks

**Rationale**: Popular document database, familiar to many developers, good for
combining vector search with rich document queries.

**Tasks**:
- [ ] Add MongoDB Go driver dependency (`go.mongodb.org/mongo-driver`)
- [ ] Implement `MongoDBClient` with `VectorDB` interface
- [ ] Implement vector index creation (Atlas Vector Search)
- [ ] Handle async index creation (poll status)
- [ ] Implement aggregation pipeline for vector search
- [ ] Add pre-filtering support
- [ ] Add scalar/binary quantization support
- [ ] Write unit tests (mocked)
- [ ] Write integration tests (Atlas M0 free tier)
- [ ] Update documentation (Atlas setup requirements)
- [ ] Create MongoDB demo script

**Success Criteria**:
- All VectorDB interface methods implemented
- Pre-filtering working correctly
- Integration tests passing with MongoDB Atlas

### Phase 6: Pinecone (v0.9.0) - ~1-2 weeks

**Rationale**: Fully managed service, zero infrastructure, good for users who
want simplicity and scale without ops overhead.

**Tasks**:
- [ ] Add Pinecone Go SDK dependency (`github.com/pinecone-io/go-pinecone`)
- [ ] Implement `PineconeClient` with `VectorDB` interface
- [ ] Handle API key authentication
- [ ] Map collections → indexes
- [ ] Implement namespace support (multi-tenancy)
- [ ] Implement CRUD operations
- [ ] Add metadata filtering
- [ ] Add hybrid search (keyword boosting)
- [ ] Write unit tests (mocked)
- [ ] Write integration tests (free tier)
- [ ] Update documentation (pricing, free tier limits)
- [ ] Create Pinecone demo script

**Success Criteria**:
- All VectorDB interface methods implemented
- Namespace support working
- Integration tests passing with Pinecone free tier

---

## Testing Requirements

### Unit Tests

Each implementation must have:

- Collection creation/deletion
- Document upsert/get/delete
- Search operations
- Error handling
- Configuration validation

### Integration Tests

- Real API connections (local Docker or free tier)
- End-to-end workflows
- Performance benchmarks
- Compatibility tests

### CI/CD

- Docker-based testing for Milvus, Qdrant, and Redis
- Free tier testing for Pinecone and MongoDB Atlas
- Automated testing on PR

---

## Documentation Updates

### Files to Update

- [ ] `docs/VDB_SUPPORT.md` - Add feature comparison
- [ ] `docs/USER_GUIDE.md` - Add setup instructions for each DB
- [ ] `docs/VECTOR_DB_ABSTRACTION.md` - Document each implementation
- [ ] `README.md` - Update database support list
- [ ] `docs/CHANGELOG.md` - Add entries for each version
- [ ] `TODOs.md` - Add milestones and tasks

### New Documentation

- [ ] `docs/milvus/README.md` - Milvus integration guide
- [ ] `docs/milvus/SETUP.md` - Setup instructions
- [ ] `docs/qdrant/README.md` - Qdrant integration guide
- [ ] `docs/qdrant/SETUP.md` - Setup instructions
- [ ] `docs/chroma/README.md` - Chroma integration guide
- [ ] `docs/chroma/SETUP.md` - Setup instructions
- [ ] `docs/redis/README.md` - Redis integration guide
- [ ] `docs/redis/LICENSE.md` - License considerations (RSALv2/SSPL/AGPL)
- [ ] `docs/mongodb/README.md` - MongoDB Atlas integration guide
- [ ] `docs/mongodb/ATLAS_SETUP.md` - Atlas setup instructions
- [ ] `docs/pinecone/README.md` - Pinecone integration guide
- [ ] `docs/pinecone/PRICING.md` - Pricing and limits

### Demo Scripts

- [ ] `demos/milvus-demo.sh` - Milvus feature demo
- [ ] `demos/qdrant-demo.sh` - Qdrant feature demo
- [ ] `demos/chroma-demo.sh` - Chroma feature demo
- [ ] `demos/redis-demo.sh` - Redis feature demo
- [ ] `demos/mongodb-demo.sh` - MongoDB Atlas feature demo
- [ ] `demos/pinecone-demo.sh` - Pinecone feature demo

---

## Risk Assessment

### High Risk

- **Pinecone Cost**: Standard tier requires $50/month minimum
  - **Mitigation**: Use free tier for testing, document costs clearly
- **Milvus Complexity**: Schema management more complex than Weaviate
  - **Mitigation**: Provide templates, auto-generate schemas where possible
- **API Changes**: Third-party SDKs may have breaking changes
  - **Mitigation**: Pin SDK versions, test before upgrading

### Medium Risk

- **gRPC Build Complexity**: Qdrant requires protobuf/gRPC
  - **Mitigation**: Use official Go client, test on multiple platforms
- **Different Abstractions**: Each DB has unique concepts
  - **Mitigation**: Strong abstraction layer, clear mapping docs

### Low Risk

- **License Compatibility**: All are compatible with MIT license
- **Community Support**: All have active communities and documentation

---

## Success Criteria

### Milvus Integration

- ✅ All VectorDB interface methods implemented
- ✅ BM25 + vector hybrid search working
- ✅ Test coverage >80%
- ✅ Documentation complete
- ✅ Demo script created

### Pinecone Integration

- ✅ All VectorDB interface methods implemented
- ✅ Namespace support working
- ✅ Free tier testing automated
- ✅ Documentation includes pricing guide
- ✅ Demo script created

### Qdrant Integration

- ✅ All VectorDB interface methods implemented
- ✅ gRPC client working
- ✅ Filtering support complete
- ✅ Test coverage >80%
- ✅ Demo script created

### Chroma Integration

- ✅ All VectorDB interface methods implemented
- ✅ Metadata filtering working
- ✅ Built-in embedding support
- ✅ Multi-tenancy support
- ✅ Test coverage >80%
- ✅ Demo script created

### Redis Integration

- ✅ All VectorDB interface methods implemented
- ✅ RediSearch FT.SEARCH working
- ✅ Hybrid search (vector + full-text) working
- ✅ Test coverage >80%
- ✅ Documentation includes license options
- ✅ Demo script created

### MongoDB Integration

- ✅ All VectorDB interface methods implemented
- ✅ Atlas Vector Search working
- ✅ Pre-filtering support complete
- ✅ Async index creation handled
- ✅ Test coverage >80%
- ✅ Documentation includes Atlas setup
- ✅ Demo script created

---

## Open Questions

1. **Multi-Vector Support**: How should Weave handle multiple vectors per
   document (Milvus, Qdrant)?
2. **Namespace Strategy**: Should we use Pinecone namespaces for multi-tenancy?
3. **Cost Warnings**: Should CLI warn when operations might incur significant
   costs (Pinecone)?
4. **Schema Auto-Generation**: Can we auto-generate Milvus schemas from simple
   Weave schemas?
5. **Embedding Defaults**: Should different DBs have different default embedding
   models?

---

## Next Steps

1. **Review this plan** with team/stakeholders
2. **Create GitHub issues** for each phase
3. **Set up development environment** for Milvus (Docker)
4. **Prototype MilvusClient** to validate interface design
5. **Update TODOs.md** with milestones

---

## References

### Milvus

- Official Docs: https://milvus.io/docs
- Go SDK: https://github.com/milvus-io/milvus-sdk-go
- GitHub: https://github.com/milvus-io/milvus

### Pinecone

- Official Docs: https://docs.pinecone.io
- Go SDK: https://github.com/pinecone-io/go-pinecone
- Pricing: https://www.pinecone.io/pricing/

### Qdrant

- Official Docs: https://qdrant.tech/documentation
- Go Client: https://github.com/qdrant/go-client
- GitHub: https://github.com/qdrant/qdrant

### Chroma

- Official Docs: https://docs.trychroma.com
- Go SDK: https://github.com/amikos-tech/chroma-go
- GitHub: https://github.com/chroma-core/chroma
- Website: https://www.trychroma.com

### Redis

- Official Docs: https://redis.io/docs/stack/search/reference/vectors/
- Go Client: https://github.com/redis/go-redis
- RediSearch: https://redis.io/docs/stack/search/
- License Options: https://redis.io/docs/about/license/

### MongoDB

- Official Docs: https://www.mongodb.com/docs/atlas/atlas-vector-search/
- Go Driver: https://github.com/mongodb/mongo-go-driver
- Atlas Vector Search Tutorial: https://www.mongodb.com/docs/atlas/atlas-vector-search/tutorials/
- Pricing: https://www.mongodb.com/pricing
