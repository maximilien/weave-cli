# Next Steps - Batch Re-Embedding Implementation

## Current Status ✅

**Phase 3A Complete** (Feb 5, 2026)
- ✅ Technical specification (`BATCH_REEMBEDDING_SPEC.md`)
- ✅ `CollectionReader` - paginated document reading
- ✅ `ProgressTracker` - real-time progress with ETA
- ✅ 14 comprehensive tests (all passing)
- ✅ Committed to main

**Foundation Ready:**
- Model registry with auto-detection (v0.9.16)
- VectorDB interface with `ListDocuments()` and `CreateDocuments()`
- Progress tracking with speed and ETA calculation

---

## Weekend Plan (2-3 hours/day)

### Saturday Session (2-3 hours)

**Phase 3B: Embedding Pipeline** (1.5-2 hours)
- [ ] Create `src/pkg/reembedding/pipeline.go`
  - `EmbeddingPipeline` struct
  - Integration with model registry (auto-detect dimensions)
  - Support for OpenAI, sentence-transformers, Ollama
  - Batch embedding generation
- [ ] Create `src/pkg/reembedding/pipeline_test.go`
  - Test with mock embedding providers
  - Test dimension validation
  - Test batch processing
- [ ] Run tests: `go test ./src/pkg/reembedding/... -v`

**Expected Output:**
```go
// src/pkg/reembedding/pipeline.go
type EmbeddingPipeline struct {
    embeddingModel string
    provider       string
    dimensions     int
    batchSize      int
}

func NewEmbeddingPipeline(modelName string) (*EmbeddingPipeline, error) {
    // Auto-detect dimensions using model registry
    dims, err := embeddings.GetModelDimensions(modelName)
    if err != nil {
        return nil, err
    }
    // ...
}

func (p *EmbeddingPipeline) ProcessBatch(ctx context.Context, docs []*vectordb.Document) error {
    // Re-generate embeddings for batch
    // Update docs in-place with new vectors
    // ...
}
```

**Deliverable:** Commit "feat: add embedding pipeline for batch re-embedding"

---

### Sunday Session (2-3 hours)

**Phase 3C: CLI Command** (1.5-2 hours)
- [ ] Create `src/cmd/collection/re_embed.go`
  - Cobra command definition
  - Flags: `--new-embedding`, `--output`, `--batch-size`
  - Orchestration logic (reader → pipeline → writer)
  - Progress display integration
- [ ] Register command in `src/cmd/collection/collection.go`
- [ ] Test manually with mock database:
  ```bash
  # Create test collection
  weave collection create TestCol --mock

  # Re-embed (should work with mock)
  weave collection re-embed TestCol \
    --new-embedding sentence-transformers/all-mpnet-base-v2 \
    --output TestCol_OSS
  ```

**Expected CLI:**
```bash
weave collection re-embed SOURCE_COLLECTION \
  --new-embedding MODEL \
  --output TARGET_COLLECTION \
  [--batch-size 100]

# Example:
weave collection re-embed Client0_PDFs \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output Client0_PDFs_OSS
```

**Deliverable:** Commit "feat: add 'weave collection re-embed' CLI command"

---

## Monday Follow-up (if time)

**Phase 3D: Integration Testing** (1 hour)
- [ ] Test with real Milvus Local database
- [ ] Create small test collection (50-100 docs)
- [ ] Re-embed with different model
- [ ] Verify document count matches
- [ ] Verify dimensions changed correctly
- [ ] Performance benchmark

**Phase 3E: Documentation** (30 min)
- [ ] Update `CHANGELOG.md` for v0.9.17
- [ ] Update README with re-embedding feature
- [ ] Create release notes

**Tag Release:**
```bash
git tag -a v0.9.17 -m "Release v0.9.17 - Batch Re-Embedding"
git push origin v0.9.17
```

---

## Implementation Details

### Embedding Pipeline Architecture

**Key Integration Points:**
1. **Model Registry** (already built in v0.9.16)
   ```go
   dims, err := embeddings.GetModelDimensions(modelName)
   info, err := embeddings.GetModelInfo(modelName)
   ```

2. **LLM Package** (for embedding generation)
   ```go
   // Reuse existing embedding logic from document ingestion
   // See: src/pkg/llm/client.go
   ```

3. **Batch Processing**
   ```go
   for reader.HasMore() {
       docs, _ := reader.ReadBatch(ctx)          // Read 100 docs
       pipeline.ProcessBatch(ctx, docs)          // Re-embed
       client.CreateDocuments(ctx, target, docs) // Write
       tracker.Update(len(docs))                 // Progress
       tracker.Display()
   }
   ```

### CLI Command Flow

```go
func runReEmbed(cmd *cobra.Command, args []string) {
    // 1. Parse flags
    sourceCollection := args[0]
    newEmbedding, _ := cmd.Flags().GetString("new-embedding")
    outputCollection, _ := cmd.Flags().GetString("output")

    // 2. Auto-detect dimensions (validate model)
    dims, err := embeddings.GetModelDimensions(newEmbedding)
    utils.PrintInfo(fmt.Sprintf("📐 Auto-detected: %d dimensions", dims))

    // 3. Create VDB client
    client, _ := utils.CreateVectorDBClient(dbConfig)

    // 4. Validate source exists
    exists, _ := client.CollectionExists(ctx, sourceCollection)

    // 5. Create output collection with new dimensions
    schema := client.GetDefaultSchema(vectordb.SchemaTypeText, outputCollection)
    schema.Vectorizer = newEmbedding
    client.CreateCollection(ctx, outputCollection, schema)

    // 6. Initialize components
    reader := reembedding.NewCollectionReader(client, sourceCollection, 100)
    pipeline := reembedding.NewEmbeddingPipeline(newEmbedding)
    tracker := reembedding.NewProgressTracker(reader.totalDocs)

    // 7. Process batches
    for reader.HasMore() {
        docs, _ := reader.ReadBatch(ctx)
        pipeline.ProcessBatch(ctx, docs)
        client.CreateDocuments(ctx, outputCollection, docs)
        tracker.Update(len(docs))
        tracker.Display()
    }

    // 8. Success
    utils.PrintSuccess("Re-embedding complete!")
}
```

---

## Testing Strategy

### Unit Tests (as you build)
- `pipeline_test.go`: Test embedding generation with mocks
- Mock providers: OpenAI, sentence-transformers, Ollama
- Test error handling (unknown model, API failures)

### Integration Tests (Monday)
1. Create test collection with 100 docs
2. Re-embed with different model
3. Verify:
   - Document count: 100 → 100 ✓
   - Dimensions: 1536 → 768 ✓
   - Metadata preserved ✓
   - IDs preserved ✓

### Performance Benchmark
- Measure speed (docs/min)
- Validate ETA accuracy
- Memory usage check

---

## Success Criteria

**Functionality:**
- ✅ Re-embeds collections without re-ingesting
- ✅ Preserves all metadata and document IDs
- ✅ Supports all models in registry (17+)
- ✅ Handles errors gracefully

**Performance:**
- ✅ 200+ docs/min throughput
- ✅ <500MB memory usage
- ✅ Accurate progress tracking with ETA

**UX:**
- ✅ Real-time progress bar
- ✅ Clear error messages
- ✅ Auto-detect dimensions (no manual config)
- ✅ Works with OSS models (sentence-transformers, Ollama)

---

## Client0 Impact

**Problem Solved:**
- Current: 5+ hours to test different embedding models (full re-ingestion)
- With re-embedding: ~15 minutes (20x faster)

**Validation Timeline:**
- Week 1: Test sentence-transformers/all-mpnet-base-v2 (primary)
- Week 2: Test alternatives (all-minilm-l6-v2, nomic-embed-text)
- Week 3: Compare results, select final model

**Time Saved:**
- 3-5 re-embedding cycles × 5 hours = 15-25 hours saved
- Enables rapid iteration during 3-week OSS validation

---

## Notes

**Already Available:**
- Model registry with 17+ models (v0.9.16)
- VectorDB interface with batch operations
- Progress tracking components
- CollectionReader with pagination

**Not Yet Built:**
- Embedding pipeline (Saturday)
- CLI command (Sunday)
- Integration tests (Monday)

**Time Estimate:**
- Saturday: 2-3 hours (pipeline + tests)
- Sunday: 2-3 hours (CLI + manual testing)
- Monday: 1.5 hours (integration + docs + release)

**Total:** 5.5-7.5 hours to ship v0.9.17

---

## Quick Reference

**File Structure:**
```
src/pkg/reembedding/
├── reader.go ✅
├── reader_test.go ✅
├── progress.go ✅
├── progress_test.go ✅
├── pipeline.go (Saturday)
├── pipeline_test.go (Saturday)
└── reembedding.go (optional orchestrator)

src/cmd/collection/
├── re_embed.go (Sunday)
└── collection.go (update to register command)

docs/planning/
├── BATCH_REEMBEDDING_SPEC.md ✅
└── NEXT_STEPS.md ✅ (this file)
```

**Commands to Run:**
```bash
# Saturday - Test pipeline
go test ./src/pkg/reembedding/... -v

# Sunday - Test CLI
./bin/weave collection re-embed --help
./bin/weave collection re-embed TestCol \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output TestCol_OSS

# Monday - Integration test
./test.sh  # Run all tests
./build.sh # Build release binary
git tag -a v0.9.17 -m "Release v0.9.17"
```

---

**Ready for weekend implementation!** 🚀

Focus: Build pipeline (Sat) → Build CLI (Sun) → Ship v0.9.17 (Mon)
