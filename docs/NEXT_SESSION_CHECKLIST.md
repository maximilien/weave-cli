# 🚨 NEXT SESSION: --top_k_images Feature (TOP PRIORITY)

## Quick Start

**Read First:** `STATUS_TOP_K_IMAGES.md` for full context

**Current Blocker:** Image collection creation fails with model name error

**Quick Test:**
```bash
./bin/weave cols create TestImg --image --embedding text-embedding-3-small --weaviate-cloud --no-confirm
```

## Immediate Next Steps

### 1️⃣ Debug Model Name Issue (15-30 min)
**Goal:** Understand why valid model name is rejected

**Steps:**
1. Add debug logging to `src/pkg/vectordb/weaviate/client_collections.go:170`
   ```go
   fmt.Fprintf(os.Stderr, "[DEBUG] Creating image collection with model: %s\n", embeddingModel)
   ```

2. Compare working text vs broken image collection:
   ```bash
   # Working
   ./bin/weave cols create TestText --text --embedding text-embedding-3-small --weaviate-cloud --no-confirm

   # Broken
   ./bin/weave cols create TestImg --image --embedding text-embedding-3-small --weaviate-cloud --no-confirm
   ```

3. Check if schema.Vectorizer is being set correctly in both cases

4. Inspect the actual JSON payload sent to Weaviate (add logging in createCollectionViaREST)

**Likely Fix Location:** `src/pkg/vectordb/weaviate/client_collections.go:167-176`

### 2️⃣ Fix and Test (30-45 min)
**Goal:** Image collection creation works

**Test:**
```bash
./scripts/test-images-with-embeddings.sh
```

**Expected Output:**
- Collection created ✅
- 4 images added ✅
- Semantic search returns results ✅
- --top_k_images returns 2 images ✅

### 3️⃣ Integration Tests (15 min)
**Goal:** All tests pass

**Command:**
```bash
export WEAVIATE_CLOUD_API_KEY="..."
export WEAVIATE_CLOUD_URL="..."
export OPENAI_API_KEY="..."
go test -tags=integration ./tests/integration/... -v -run TestTopKImagesFlag
```

### 4️⃣ Final Verification (10 min)
**Goal:** End-to-end workflow works

**Test:**
```bash
# Re-create WeaveImages with embeddings
./scripts/reingest-images-with-embeddings.sh

# Test multi-modal query
./bin/weave cols query WeaveDocs WeaveImages "weave cli screenshot" \
  --agent rag-agent --top_k 5 --top_k_images 2 --verbose --weaviate-cloud
```

**Success:** Should see 2 images in results with 📸 emoji

## Files to Check

### Primary Investigation
- `src/pkg/vectordb/weaviate/client_collections.go:145-210` - Schema creation
- `src/cmd/utils/collection.go:151-176` - CreateGenericCollectionWithSchemaType
- `src/pkg/vectordb/weaviate/collections.go:14-36` - Adapter.CreateCollection

### If Still Stuck
- Compare with working `CreateSupabaseCollection()` implementation
- Check how Weaviate SDK expects vectorConfig vs vectorizer fields
- Review Weaviate v1.25+ named vectors documentation

## Quick Wins If Blocked

If image creation still blocked, you can:

1. **Document the issue** - Add to STATUS_TOP_K_IMAGES.md
2. **Test with existing collections** - Use WeaveDocs (text) only
3. **Mock test** - Create mock image collection for testing detection logic
4. **Alternative approach** - Use schema YAML file instead of programmatic creation

## Success Criteria

- [ ] `./bin/weave cols create TestImg --image --embedding text-embedding-3-small --weaviate-cloud` works
- [ ] `./scripts/test-images-with-embeddings.sh` completes successfully
- [ ] Query with `--top_k_images 2` returns 2 image results
- [ ] Integration tests pass
- [ ] All code passes linter

## Estimated Time

- Debug: 15-30 minutes
- Fix: 15-30 minutes
- Test: 30 minutes
- **Total: 1-1.5 hours**

## Questions to Answer

1. Why does text collection creation work but image collection fails?
2. Is the embedding model being passed through all layers correctly?
3. Does Weaviate API expect different fields for image vs text vectorizers?
4. Should we use vectorizer vs vectorConfig for image collections?

---

**Start here:** Add debug logging to see what model name is actually being sent to Weaviate
