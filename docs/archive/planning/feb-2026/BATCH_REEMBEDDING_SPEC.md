# Batch Re-Embedding Technical Specification
*Phase 3 - Highest ROI Feature for Client0*

## Executive Summary

**Problem**: Re-testing different embedding models requires full re-ingestion (5+ hours for 11 PDFs)
**Solution**: Re-embed existing text chunks without re-ingesting documents
**Impact**: 5 hours → 15 minutes (20x faster), saves 10-20 hours during Client0 validation

---

## Architecture Overview

```
┌─────────────────┐
│ Source          │
│ Collection      │ ← Read existing documents (paginated)
│ (3,518 docs)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Document        │
│ Reader          │ ← Batch fetch (100 docs/batch)
│ (Paginated)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Text            │
│ Extractor       │ ← Extract text content only
│                 │   (skip images, preserve metadata)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Embedding       │
│ Generator       │ ← Generate new embeddings
│ (New Model)     │   (sentence-transformers, Ollama, etc.)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Document        │
│ Builder         │ ← Reconstruct with new vectors
│                 │   (preserve IDs, metadata)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Batch           │
│ Upserter        │ ← Insert to target collection
│                 │   (100 docs/batch)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Target          │
│ Collection      │ ← New collection with new embeddings
│ (3,518 docs)    │
└─────────────────┘

Progress: [██████████████░░░░░░] 70% (2,463/3,518 docs)
Speed: 234 docs/min | ETA: 4m 30s
```

---

## Reusable Components (From Existing Codebase)

### ✅ Already Built

1. **VectorDB Interface** (`src/pkg/vectordb/interfaces.go`)
   ```go
   // ListDocuments - paginated document reading
   ListDocuments(ctx, collectionName string, limit, offset int) ([]*Document, error)

   // CreateDocuments - batch document creation
   CreateDocuments(ctx, collectionName string, documents []*Document) error

   // CollectionExists - validation
   CollectionExists(ctx, collectionName string) (bool, error)

   // GetCollectionCount - progress tracking
   GetCollectionCount(ctx, collectionName string) (int64, error)
   ```

2. **Document Structure** (`src/pkg/vectordb/interfaces.go`)
   ```go
   type Document struct {
       ID        string
       Text      string
       Content   string
       Metadata  map[string]interface{}
   }
   ```

3. **LLM Package** (`src/pkg/llm`)
   - Embedding generation for OpenAI, Anthropic, etc.
   - Already handles batching and rate limiting

4. **Progress Tracking** (`src/pkg/progress`)
   - Progress bars, ETA calculation
   - Already used in document ingestion

5. **Model Registry** (`src/pkg/embeddings`) - **NEW! Just built!**
   - Auto-detect dimensions
   - Validate embedding models
   - OSS model support

---

## New Components Needed

### 1. Collection Reader (`src/pkg/reembedding/reader.go`)

**Purpose**: Read documents from source collection in paginated batches

```go
package reembedding

type CollectionReader struct {
    client         vectordb.VectorDBClient
    collectionName string
    batchSize      int
    offset         int
    totalDocs      int64
}

func NewCollectionReader(client vectordb.VectorDBClient, collectionName string) (*CollectionReader, error)

// ReadBatch reads next batch of documents
func (r *CollectionReader) ReadBatch(ctx context.Context) ([]*vectordb.Document, error)

// HasMore returns true if more documents available
func (r *CollectionReader) HasMore() bool

// Progress returns current progress (docs read, total docs)
func (r *CollectionReader) Progress() (int, int64)
```

**Implementation**:
```go
func (r *CollectionReader) ReadBatch(ctx context.Context) ([]*vectordb.Document, error) {
    docs, err := r.client.ListDocuments(ctx, r.collectionName, r.batchSize, r.offset)
    if err != nil {
        return nil, fmt.Errorf("failed to read batch: %w", err)
    }

    r.offset += len(docs)
    return docs, nil
}
```

**Batch Size**: 100 documents (tunable via flag)

---

### 2. Embedding Pipeline (`src/pkg/reembedding/pipeline.go`)

**Purpose**: Re-generate embeddings for documents using new model

```go
package reembedding

type EmbeddingPipeline struct {
    embeddingModel string
    provider       string // "openai", "ollama", "sentence-transformers"
    dimensions     int    // auto-detected from model registry!
    batchSize      int
}

func NewEmbeddingPipeline(modelName string) (*EmbeddingPipeline, error)

// ProcessBatch generates embeddings for batch of documents
func (p *EmbeddingPipeline) ProcessBatch(ctx context.Context, docs []*vectordb.Document) error

// GetDimensions returns vector dimensions (from model registry)
func (p *EmbeddingPipeline) GetDimensions() int
```

**Integration with Model Registry**:
```go
func NewEmbeddingPipeline(modelName string) (*EmbeddingPipeline, error) {
    // Auto-detect dimensions using our new model registry!
    dims, err := embeddings.GetModelDimensions(modelName)
    if err != nil {
        return nil, fmt.Errorf("unknown model %s: %w", modelName, err)
    }

    info, _ := embeddings.GetModelInfo(modelName)

    return &EmbeddingPipeline{
        embeddingModel: modelName,
        provider:       info.Provider,
        dimensions:     dims,
        batchSize:      10, // Adjust based on provider
    }, nil
}
```

---

### 3. Progress Tracker (`src/pkg/reembedding/progress.go`)

**Purpose**: Real-time progress updates with ETA

```go
package reembedding

type ProgressTracker struct {
    total       int64
    processed   int64
    startTime   time.Time
    lastUpdate  time.Time
}

func NewProgressTracker(total int64) *ProgressTracker

// Update increments progress and displays
func (p *ProgressTracker) Update(count int)

// Display shows progress bar with ETA
func (p *ProgressTracker) Display()

// GetETA returns estimated time remaining
func (p *ProgressTracker) GetETA() time.Duration
```

**Output Example**:
```
Re-embedding SourceCollection → TargetCollection
Embedding model: sentence-transformers/all-mpnet-base-v2
Dimensions: 768 (auto-detected, OSS)

Progress: [██████████████░░░░░░] 70% (2,463/3,518 docs)
Speed: 234 docs/min
Estimated time remaining: 4m 30s
```

---

### 4. Re-Embedding Orchestrator (`src/cmd/collection/re_embed.go`)

**Purpose**: CLI command and orchestration logic

```go
var reEmbedCmd = &cobra.Command{
    Use:   "re-embed SOURCE_COLLECTION",
    Short: "Re-embed collection with different embedding model",
    Long: `Re-embed an existing collection using a different embedding model.

This reads existing text chunks and generates new embeddings without
re-ingesting documents. Much faster than full re-ingestion.`,
    Example: `  # Re-embed with sentence-transformers (OSS)
  weave collection re-embed MyCollection \
    --new-embedding sentence-transformers/all-mpnet-base-v2 \
    --output MyCollection_OSS

  # Re-embed with OpenAI
  weave collection re-embed MyCollection \
    --new-embedding text-embedding-3-large \
    --output MyCollection_Large

  # Re-embed with Ollama (local)
  weave collection re-embed MyCollection \
    --new-embedding nomic-embed-text \
    --output MyCollection_Nomic`,
    Args: cobra.ExactArgs(1),
    Run:  runReEmbed,
}

func init() {
    reEmbedCmd.Flags().StringP("new-embedding", "e", "", "New embedding model (REQUIRED)")
    reEmbedCmd.Flags().StringP("output", "o", "", "Output collection name (REQUIRED)")
    reEmbedCmd.Flags().IntP("batch-size", "b", 100, "Batch size for processing")
    reEmbedCmd.Flags().Bool("skip-existing", false, "Skip if output collection exists")
    reEmbedCmd.MarkFlagRequired("new-embedding")
    reEmbedCmd.MarkFlagRequired("output")
}
```

---

## Implementation Flow

```go
func runReEmbed(cmd *cobra.Command, args []string) {
    sourceCollection := args[0]
    newEmbeddingModel, _ := cmd.Flags().GetString("new-embedding")
    outputCollection, _ := cmd.Flags().GetString("output")
    batchSize, _ := cmd.Flags().GetInt("batch-size")

    // 1. Load config and create client
    cfg, _ := utils.LoadConfigWithInteractiveHelp()
    dbConfig := utils.HandleSingleDatabaseSelection(ctx, selection, cfg, ...)
    client, _ := utils.CreateVectorDBClient(dbConfig)

    // 2. Validate source collection exists
    exists, _ := client.CollectionExists(ctx, sourceCollection)
    if !exists {
        utils.PrintError("Source collection not found")
        os.Exit(1)
    }

    // 3. Auto-detect dimensions for new model (using our registry!)
    dims, err := embeddings.GetModelDimensions(newEmbeddingModel)
    if err != nil {
        utils.PrintError("Unknown embedding model: " + newEmbeddingModel)
        utils.PrintInfo("Run 'weave embeddings list' to see supported models")
        os.Exit(1)
    }

    info, _ := embeddings.GetModelInfo(newEmbeddingModel)
    utils.PrintInfo(fmt.Sprintf("📐 Auto-detected: %d dimensions for %s (OSS: %v)",
        dims, newEmbeddingModel, info.OSS))

    // 4. Get total document count
    totalDocs, _ := client.GetCollectionCount(ctx, sourceCollection)
    utils.PrintInfo(fmt.Sprintf("Total documents to re-embed: %d", totalDocs))

    // 5. Create output collection with new dimensions
    schema := client.GetDefaultSchema(vectordb.SchemaTypeText, outputCollection)
    schema.Vectorizer = newEmbeddingModel
    // Dimensions set automatically by VDB client using our registry!
    client.CreateCollection(ctx, outputCollection, schema)

    // 6. Initialize components
    reader := reembedding.NewCollectionReader(client, sourceCollection)
    pipeline := reembedding.NewEmbeddingPipeline(newEmbeddingModel)
    tracker := reembedding.NewProgressTracker(totalDocs)

    // 7. Process in batches
    for reader.HasMore() {
        // Read batch
        docs, _ := reader.ReadBatch(ctx)

        // Generate new embeddings
        pipeline.ProcessBatch(ctx, docs)

        // Write to output collection
        client.CreateDocuments(ctx, outputCollection, docs)

        // Update progress
        tracker.Update(len(docs))
        tracker.Display()
    }

    // 8. Success
    utils.PrintSuccess(fmt.Sprintf("Re-embedded %d documents to %s", totalDocs, outputCollection))
}
```

---

## Error Handling & Edge Cases

### 1. Source Collection Missing
```go
exists, _ := client.CollectionExists(ctx, sourceCollection)
if !exists {
    return fmt.Errorf("source collection '%s' not found", sourceCollection)
}
```

### 2. Output Collection Already Exists
```go
exists, _ := client.CollectionExists(ctx, outputCollection)
if exists && !skipExisting {
    return fmt.Errorf("output collection '%s' already exists (use --skip-existing to override)", outputCollection)
}
```

### 3. Unknown Embedding Model
```go
dims, err := embeddings.GetModelDimensions(newEmbeddingModel)
if err != nil {
    return fmt.Errorf("unknown embedding model: %s\nRun 'weave embeddings list' to see supported models", newEmbeddingModel)
}
```

### 4. Batch Processing Failure
```go
for reader.HasMore() {
    docs, err := reader.ReadBatch(ctx)
    if err != nil {
        // Log error, skip batch, continue
        utils.PrintWarning(fmt.Sprintf("Skipped batch at offset %d: %v", reader.offset, err))
        continue
    }

    // Process...
}
```

### 5. Resume on Failure (Future Enhancement)
```go
// Save checkpoint every N batches
if batchesProcessed % 10 == 0 {
    checkpoint.Save(reader.offset)
}
```

---

## Performance Estimates

### Client0 Use Case (11 PDFs → 3,518 docs)

**Current (Full Re-Ingestion)**:
- PDF parsing: 2 hours
- Chunking: 1 hour
- Embedding: 1.5 hours
- Indexing: 30 minutes
- **Total**: ~5 hours

**With Batch Re-Embedding**:
- Reading docs: 30 seconds (paginated, 100/batch = 36 batches)
- Embedding: 12 minutes (3,518 docs ÷ 234 docs/min)
- Indexing: 2 minutes (batch upsert)
- **Total**: ~15 minutes

**Speedup**: 20x faster (5 hours → 15 minutes)

### Batch Size Tuning

| Batch Size | Read Time | Embed Time | Write Time | Total    |
|------------|-----------|------------|------------|----------|
| 50         | 1 min     | 15 min     | 4 min      | 20 min   |
| 100        | 30 sec    | 12 min     | 2 min      | **15 min** |
| 200        | 15 sec    | 10 min     | 1 min      | 11 min   |

**Optimal**: 100 docs/batch (balance between throughput and memory)

---

## Testing Strategy

### 1. Unit Tests
```go
// Test collection reader pagination
func TestCollectionReader_ReadBatch(t *testing.T)

// Test embedding pipeline
func TestEmbeddingPipeline_ProcessBatch(t *testing.T)

// Test progress tracker
func TestProgressTracker_ETA(t *testing.T)
```

### 2. Integration Tests
```go
// Test with mock VDB
func TestReEmbed_MockDatabase(t *testing.T) {
    // Create 100 docs with OpenAI embeddings
    // Re-embed with sentence-transformers
    // Verify all 100 docs re-embedded
    // Verify dimensions changed (1536 → 768)
}
```

### 3. End-to-End Test (Manual)
```bash
# Create test collection
weave docs ingest test-pdfs/ --collection TestCol --embedding text-embedding-3-small

# Re-embed with sentence-transformers
weave collection re-embed TestCol \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output TestCol_OSS

# Verify both collections work
weave query "test query" --collection TestCol
weave query "test query" --collection TestCol_OSS

# Compare results
weave embeddings compare --collections "TestCol,TestCol_OSS" --queries "test1,test2"
```

---

## File Structure

```
src/
├── pkg/
│   ├── reembedding/
│   │   ├── reader.go           # Collection reader (paginated)
│   │   ├── reader_test.go
│   │   ├── pipeline.go         # Embedding pipeline
│   │   ├── pipeline_test.go
│   │   ├── progress.go         # Progress tracker
│   │   ├── progress_test.go
│   │   └── reembedding.go      # Main orchestrator
│   │
│   └── embeddings/
│       └── model_registry.go   # ✅ Already built!
│
└── cmd/
    └── collection/
        ├── re_embed.go         # CLI command
        └── re_embed_test.go
```

---

## Implementation Timeline

### Phase 3A: Core Components (4-5 hours)
- `reader.go`: Collection reader with pagination
- `pipeline.go`: Embedding generation pipeline
- `progress.go`: Progress tracking UI
- Unit tests for each component

### Phase 3B: CLI Integration (2-3 hours)
- `re_embed.go`: CLI command
- Integration with existing commands
- Error handling and validation
- Help text and examples

### Phase 3C: Testing & Polish (2 hours)
- Integration tests with mock VDB
- Manual testing with real data
- Performance tuning (batch sizes)
- Documentation

**Total**: 6-8 hours (can split across multiple sessions)

---

## Success Metrics

✅ **Functionality**:
- Re-embeds 3,518 docs in <20 minutes
- Preserves all metadata and document IDs
- Supports all embedding models in registry
- Handles errors gracefully

✅ **Performance**:
- 200+ docs/min throughput
- <500MB memory usage
- Paginated (no loading all docs in memory)

✅ **UX**:
- Real-time progress bar with ETA
- Clear error messages
- Auto-detect dimensions (no manual config)
- OSS-friendly (sentence-transformers, Ollama)

---

## Next Steps

**Immediate (this session if time)**:
1. Create package structure
2. Implement `CollectionReader`
3. Add basic tests

**Next Session**:
4. Implement `EmbeddingPipeline`
5. Implement `ProgressTracker`
6. Wire up CLI command

**Following Session**:
7. Integration tests
8. Manual testing with Client0 dataset
9. Performance tuning
10. Ship v0.9.17!

---

**Ready to start implementation!** 🚀
