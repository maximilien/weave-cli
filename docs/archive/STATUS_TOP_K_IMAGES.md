# ⚠️ TOP PRIORITY: --top_k_images Feature Status

**Status:** In Progress - Image Collection Creation Blocked
**Priority:** TOP PRIORITY for next session
**Last Updated:** 2026-01-16 16:15 PST

## 🎯 Goal

Enable `--top_k_images` flag to guarantee image results in multi-modal RAG queries by using separate topK values for image vs text collections.

**Target Use Case:**
```bash
weave cols query WeaveDocs WeaveImages "weave cli screenshot" \
  --agent rag-agent --top_k 5 --top_k_images 2 --verbose
```

Expected: 5 text results from WeaveDocs + 2 image results from WeaveImages

## ✅ What's Working

### 1. Flag Implementation
- ✅ `--top_k_images` flag added to CLI
- ✅ Detection logic correctly identifies image vs text collections (name-based)
- ✅ Query logic applies different topK values per collection type
- ✅ Debug output only shows with `--verbose` (fixed)
- ✅ All code compiles and passes linter

### 2. Integration Tests
- ✅ Created `tests/integration/top_k_images_test.go`
- ✅ Tests compile successfully
- ✅ Test validates detection logic, query splitting, and result merging
- ✅ Documentation in `tests/integration/README.md`

### 3. Scripts & Documentation
- ✅ `scripts/test-images-with-embeddings.sh` - Quick test with sample data
- ✅ `scripts/reingest-images-with-embeddings.sh` - Re-ingest existing collections
- ✅ Both scripts show commands for easy modification
- ✅ CHANGELOG.md updated with feature description

## ❌ What's NOT Working

### BLOCKER: Image Collection Creation with Embeddings

**Current Error:**
```bash
$ ./bin/weave cols create TestWeaveImages --image --embedding text-embedding-3-small --weaviate-cloud

❌ Failed to create collection: wrong OpenAI model name, available model names are:
   [ada babbage curie davinci text-embedding-3-small text-embedding-3-large]
```

**The Problem:**
Even though we're passing `text-embedding-3-small` (a valid model name), Weaviate is rejecting it.

**Root Cause (Suspected):**
The code path through `CreateGenericCollectionWithSchemaType()` → `Adapter.CreateCollection()` → `Client.CreateCollection()` may not be properly passing the embedding model to the Weaviate schema creation.

## 🔍 Bugs Fixed Today

1. **Duplicate metadata property** - Fixed in `client_collections.go:371`
2. **Wrong schema type for --image** - Fixed in `create.go` and `collection.go`
3. **Unnecessary metadata flags** - Removed requirement for `--flat-metadata`/`--json-metadata`
4. **Debug spam without --verbose** - Added verbose parameter to detection logic
5. **Multi2vec-CLIP dependency** - Simplified to text2vec-openai only
6. **Markdown linting issues** - Fixed in `tests/integration/README.md`

## 🔧 Code Changes Summary

### Modified Files
1. `src/cmd/collection/create.go` - Fixed --image schema type, removed metadata flag requirement
2. `src/cmd/utils/collection.go` - Added CreateGenericCollectionWithSchemaType()
3. `src/cmd/utils/agent_query.go` - Added verbose parameter to detection logic
4. `src/pkg/vectordb/weaviate/client_collections.go` - Fixed duplicate metadata, removed CLIP
5. `src/pkg/vectordb/weaviate/client_queries.go` - Added TopKImages field
6. `scripts/test-images-with-embeddings.sh` - Updated with correct flags
7. `scripts/reingest-images-with-embeddings.sh` - Updated with correct flags
8. `tests/integration/top_k_images_test.go` - Created integration test
9. `tests/integration/README.md` - Created test documentation

## 🚨 Next Steps (TOP PRIORITY)

### Step 1: Debug Image Collection Creation
**Goal:** Understand why `text-embedding-3-small` is being rejected

**Investigation Points:**
1. Check how embedding model flows from CLI → CreateGenericCollectionWithSchemaType → Weaviate API
2. Verify `schema.Vectorizer` is set correctly in `collection.go:162`
3. Check if `client_collections.go:createCollectionViaREST()` is using the right field
4. Compare with working text collection creation

**Debug Commands:**
```bash
# Test text collection (should work)
./bin/weave cols create TestText --text --embedding text-embedding-3-small --weaviate-cloud --no-confirm

# Test image collection (currently fails)
./bin/weave cols create TestImage --image --embedding text-embedding-3-small --weaviate-cloud --no-confirm
```

**Expected Fix Location:**
- Likely in `src/pkg/vectordb/weaviate/client_collections.go` around line 170
- The vectorConfig setup for image collections may not be using the embedding model correctly

### Step 2: Test Image Collection with Embeddings
**Goal:** Successfully create an image collection WITH embeddings

**Test Command:**
```bash
./scripts/test-images-with-embeddings.sh
```

**Success Criteria:**
- Collection created successfully
- 4 test images added with searchable content
- Semantic search returns relevant results

### Step 3: Verify --top_k_images Works End-to-End
**Goal:** Confirm flag guarantees image results

**Test Commands:**
```bash
# Without --top_k_images (may get 0 images)
./bin/weave cols query WeaveDocs TestWeaveImages "weave cli screenshot" \
  --agent rag-agent --top_k 5 --weaviate-cloud

# With --top_k_images (MUST get 2 images)
./bin/weave cols query WeaveDocs TestWeaveImages "weave cli screenshot" \
  --agent rag-agent --top_k 5 --top_k_images 2 --verbose --weaviate-cloud
```

**Success Criteria:**
- Second query returns exactly 2 image results + up to 5 text results
- Debug output shows correct detection and topK application
- RAG agent response includes image citations with emoji

### Step 4: Run Integration Tests
**Goal:** Verify all tests pass

**Test Command:**
```bash
export WEAVIATE_CLOUD_API_KEY="..."
export WEAVIATE_CLOUD_URL="..."
export OPENAI_API_KEY="..."
go test -tags=integration ./tests/integration/... -v -run TestTopKImagesFlag -timeout 10m
```

## 📋 Test Cases to Verify

### Manual Test Cases
- [ ] Create text collection with embeddings
- [ ] Create image collection with embeddings  ← **BLOCKED HERE**
- [ ] Add documents to image collection
- [ ] Query single image collection (semantic search works)
- [ ] Query text + image without --top_k_images (baseline)
- [ ] Query text + image WITH --top_k_images (guaranteed images)
- [ ] Verify --verbose shows debug output
- [ ] Verify without --verbose is silent

### Integration Test Cases
- [ ] TestTopKImagesFlag/Setup_Collections
- [ ] TestTopKImagesFlag/Query_Without_TopKImages
- [ ] TestTopKImagesFlag/Query_With_TopKImages
- [ ] TestTopKImagesFlag/Image_Collection_Detection

## 🐛 Known Issues

1. **Image collection creation fails** - Model name validation error (BLOCKER)
2. **WeaveImages has no embeddings** - Original collection has 305 images but no vectors
3. **Multi2vec-CLIP not available** - Weaviate cluster doesn't have CLIP module (handled by using text2vec-openai)

## 📚 Key Files Reference

### Implementation
- `src/cmd/collection/query.go:96` - Flag definition
- `src/cmd/utils/agent_query.go:23` - Detection logic
- `src/cmd/utils/agent_query.go:115` - TopK application
- `src/pkg/vectordb/weaviate/client_queries.go:25` - QueryOptions struct

### Tests
- `tests/integration/top_k_images_test.go` - Integration test
- `scripts/test-images-with-embeddings.sh` - Quick test script
- `scripts/reingest-images-with-embeddings.sh` - Re-ingestion helper

### Documentation
- `tests/integration/README.md` - How to run tests
- `CHANGELOG.md:1` - Feature description

## 🎯 Definition of Done

- [ ] Image collections can be created with embeddings (text-embedding-3-small)
- [ ] Test script completes successfully showing image results
- [ ] Integration tests pass with real Weaviate credentials
- [ ] Manual query with --top_k_images returns guaranteed image results
- [ ] Documentation complete with working examples
- [ ] All code passes linter and builds cleanly

## 💡 Additional Context

### Why This Matters
Users with multi-modal collections (text + images) were getting ALL text results and ZERO images because text embeddings typically have higher similarity scores. The `--top_k_images` flag solves this by forcing a minimum number of image results.

### Original Issue
```bash
$ weave cols query WeaveDocs WeaveImages "weave cli screenshot" \
    --agent rag-agent --top_k 5

# Result: 5 text docs, 0 images (even though WeaveImages has relevant images)
```

### Target Solution
```bash
$ weave cols query WeaveDocs WeaveImages "weave cli screenshot" \
    --agent rag-agent --top_k 5 --top_k_images 2

# Result: 5 text docs + 2 images (guaranteed mix of content types)
```

## 📞 Contact Points

- Integration test: `tests/integration/top_k_images_test.go`
- Test script: `scripts/test-images-with-embeddings.sh`
- Detection logic: `src/cmd/utils/agent_query.go:23`
- Schema creation: `src/pkg/vectordb/weaviate/client_collections.go:145`

---

**Ready to resume:** Start with debugging why `text-embedding-3-small` is rejected in image collection creation.
