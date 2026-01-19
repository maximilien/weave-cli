# Integration Tests

Integration tests that require real vector database connections.

## Prerequisites

These tests require environment variables for connecting to vector databases.
The tests will automatically detect which VDB is configured and use it.

### Supported Vector Databases

**Weaviate Cloud:**

```bash
export WEAVIATE_CLOUD_API_KEY="your-api-key"
export WEAVIATE_CLOUD_URL="https://your-cluster.weaviate.network"
export OPENAI_API_KEY="your-openai-key"
```

**Milvus Cloud (Zilliz):**

```bash
export MILVUS_CLOUD_TOKEN="your-token"
export MILVUS_CLOUD_ADDRESS="your-cluster.serverless.aws.cloud.zilliz.com:443"
export OPENAI_API_KEY="your-openai-key"
```

**Chroma Cloud:**

```bash
export CHROMA_API_KEY="your-api-key"
export CHROMA_TENANT="your-tenant-id"
export CHROMA_DATABASE="your-database"
export OPENAI_API_KEY="your-openai-key"
```

**Qdrant Cloud:**

```bash
export QDRANT_API_KEY="your-api-key"
export QDRANT_URL="https://your-cluster.qdrant.io"
export OPENAI_API_KEY="your-openai-key"
```

## Running Tests

### Build CLI first

Integration tests use the actual CLI binary, so build it first:

```bash
./build.sh
```

### Run all integration tests

```bash
# With credentials set (uses -tags=integration to include integration tests)
go test -tags=integration ./tests/integration/... -v

# Skip if credentials not set (default behavior)
go test -tags=integration ./tests/integration/... -v
# Output: SKIP: Skipping integration test: No vector database credentials found
```

### Run specific test

```bash
# API-level test (uses Go API directly)
go test -tags=integration ./tests/integration/... -v -run TestTopKImagesFlag

# CLI-level test (uses actual CLI commands)
go test -tags=integration ./tests/integration/... -v -run TestTopKImagesCLI
```

### Run with timeout

```bash
go test -tags=integration ./tests/integration/... -v -timeout 10m
```

## Tests

### TestTopKImagesFlag (API-Level)

Tests the `--top_k_images` flag implementation at the Go API level using
Weaviate adapter directly for multi-modal queries.

**What it tests:**

1. **Collection Setup**: Creates text and image collections with embeddings
2. **Query Without topKImages**: Tests standard multi-collection query
3. **Query With topKImages**: Tests that image collections use separate topK
   value
4. **Image Collection Detection**: Tests name-based detection logic

**Expected Behavior:**

- Text collection: Uses `--top_k` value
- Image collection: Uses `--top_k_images` value when specified
- Results are merged and sorted by score

**Example Run:**

```bash
$ export WEAVIATE_CLOUD_API_KEY="..."
$ export WEAVIATE_CLOUD_URL="https://..."
$ go test ./tests/integration/... -v -run TestTopKImagesFlag

=== RUN   TestTopKImagesFlag
=== RUN   TestTopKImagesFlag/Setup_Collections
    top_k_images_test.go:96: Created 3 text docs in TestTopKImages_TextCol and 2 image docs in TestTopKImages_ImageCol
=== RUN   TestTopKImagesFlag/Query_Without_TopKImages
    top_k_images_test.go:106: Text collection returned 3 results
    top_k_images_test.go:111: Image collection returned 2 results
    top_k_images_test.go:115: Total results without topKImages: 5
    top_k_images_test.go:123: Image results: 2
=== RUN   TestTopKImagesFlag/Query_With_TopKImages
    top_k_images_test.go:133: Text collection returned 3 results (requested TopK=3)
    top_k_images_test.go:141: Image collection returned 2 results (requested TopK=2)
    top_k_images_test.go:149: ✅ SUCCESS: Got 2 image results with topKImages=2
=== RUN   TestTopKImagesFlag/Image_Collection_Detection
=== RUN   TestTopKImagesFlag/Image_Collection_Detection/TextCollection
=== RUN   TestTopKImagesFlag/Image_Collection_Detection/ImageCollection
=== RUN   TestTopKImagesFlag/Image_Collection_Detection/NameBasedDetection_Images
=== RUN   TestTopKImagesFlag/Image_Collection_Detection/NameBasedDetection_Photos
=== RUN   TestTopKImagesFlag/Image_Collection_Detection/NameBasedDetection_Docs
--- PASS: TestTopKImagesFlag (5.23s)
    --- PASS: TestTopKImagesFlag/Setup_Collections (2.41s)
    --- PASS: TestTopKImagesFlag/Query_Without_TopKImages (1.15s)
    --- PASS: TestTopKImagesFlag/Query_With_TopKImages (1.12s)
    --- PASS: TestTopKImagesFlag/Image_Collection_Detection (0.00s)
        --- PASS: TestTopKImagesFlag/Image_Collection_Detection/TextCollection (0.00s)
        --- PASS: TestTopKImagesFlag/Image_Collection_Detection/ImageCollection (0.00s)
        --- PASS: TestTopKImagesFlag/Image_Collection_Detection/NameBasedDetection_Images (0.00s)
        --- PASS: TestTopKImagesFlag/Image_Collection_Detection/NameBasedDetection_Photos (0.00s)
        --- PASS: TestTopKImagesFlag/Image_Collection_Detection/NameBasedDetection_Docs (0.00s)
PASS
```

### TestTopKImagesCLI (CLI-Level)

Tests the `--top_k_images` flag end-to-end via actual CLI commands. This test
runs the full CLI executable to verify the feature works in production.

**What it tests:**

1. **Collection Creation**: Creates text and image collections via
   `cols create` command
2. **Document Upload**: Adds documents via `docs create` command
3. **Multi-Collection Query**: Tests `cols query` with `--top_k_images` flag
4. **RAG Agent Integration**: Verifies RAG agent includes citations from both
   collections
5. **Edge Cases**: Tests with `topKImages=0`, image-only queries, etc.
6. **Multi-VDB Support**: Automatically detects and uses available VDB
   (Milvus, Chroma, Weaviate, Qdrant)

**Expected Behavior:**

- Creates collections and documents successfully
- RAG agent response includes citations from both text and image collections
- Debug output shows correct collection type detection
- `--top_k_images` flag controls image results separately from text
  results

**Example Run:**

```bash
$ export MILVUS_CLOUD_TOKEN="..."
$ export MILVUS_CLOUD_ADDRESS="..."
$ export OPENAI_API_KEY="..."
$ go test -tags=integration ./tests/integration/... -v -run TestTopKImagesCLI

=== RUN   TestTopKImagesCLI
    top_k_images_cli_test.go:32: Running tests with Milvus Cloud
=== RUN   TestTopKImagesCLI/Setup_Collections_Via_CLI
=== RUN   TestTopKImagesCLI/Query_With_TopKImages_Flag
    top_k_images_cli_test.go:89: ✅ Multi-collection query with --top_k_images successful
=== RUN   TestTopKImagesCLI/Query_Without_TopKImages_Flag
    top_k_images_cli_test.go:104: ✅ Multi-collection query without --top_k_images works as baseline
=== RUN   TestTopKImagesCLI/Query_With_TopKImages_Zero
    top_k_images_cli_test.go:119: ✅ Query with --top_k_images=0 works correctly
=== RUN   TestTopKImagesCLI/Query_Image_Collection_Only
    top_k_images_cli_test.go:135: ✅ Query with only image collection works
=== RUN   TestTopKImagesCLI/Image_Collection_Detection
--- PASS: TestTopKImagesCLI (25.31s)
PASS
```

### TestVerifyCitationWorkflow (Manual Verification)

Tests the complete citation workflow with existing collections across any VDB.
This test is designed for manual verification with your real collections.

**What it tests:**

1. **Multi-Collection Query**: Queries both text and image collections
2. **Citation Verification**: Confirms RAG agent cites both collections
3. **Content Validation**: Verifies text has docs, images have visual content
4. **TopK Application**: Confirms different topK values for text vs images

**Environment Variables:**

```bash
# Use custom collection names (optional)
export TEST_TEXT_COL=MyTextCol
export TEST_IMAGE_COL=MyImageCol

# Or let it auto-detect based on VDB:
# Milvus: MilvusTextCol, MilvusImageCol
# Chroma: ChromaTextCol, ChromaImageCol
# Weaviate: WeaviateTextCol, WeaviateImageCol
# Qdrant: QdrantTextCol, QdrantImageCol
```

**Example Run:**

```bash
$ go test -tags=integration ./tests/integration/... -v -run TestVerifyCitationWorkflow

=== RUN   TestVerifyCitationWorkflow
    verify_citations_test.go:37: Testing with Milvus Cloud
    verify_citations_test.go:43: Using collections: text=MilvusTextCol, image=MilvusImageCol
=== RUN   TestVerifyCitationWorkflow/Multi_Collection_Query_With_Citations
    verify_citations_test.go:59: ✅ RAG response cites text collection: MilvusTextCol
    verify_citations_test.go:62: ✅ RAG response cites image collection: MilvusImageCol
=== RUN   TestVerifyCitationWorkflow/Verify_Text_Collection_Content
    verify_citations_test.go:100: ✅ Text collection has CLI documentation content
=== RUN   TestVerifyCitationWorkflow/Verify_Image_Collection_Content
    verify_citations_test.go:123: ✅ Image collection has visual content descriptions
=== RUN   TestVerifyCitationWorkflow/Verify_TopK_Values_Applied
    verify_citations_test.go:148: ✅ Text collection used topK=5
    verify_citations_test.go:151: ✅ Image collection used topK=2 (from --top_k_images)
    verify_citations_test.go:154: ✅ Different topK values applied to text vs image collections
    verify_citations_test.go:162:
        ======================================================================
    verify_citations_test.go:163: ✅ WORKFLOW VERIFICATION COMPLETE
    verify_citations_test.go:165: Vector Database: Milvus Cloud
    verify_citations_test.go:166: Text Collection: MilvusTextCol (topK=5)
    verify_citations_test.go:167: Image Collection: MilvusImageCol (topK=2)
--- PASS: TestVerifyCitationWorkflow (10.30s)
PASS
```

## Manual Testing

For quick manual testing with real data:

```bash
# Run the test script (creates test collections)
./scripts/test-images-with-embeddings.sh

# Or test with existing collections
weave cols query WeaveDocs TestWeaveImages "weave cli screenshot" \
  --agent rag-agent --top_k 5 --top_k_images 2 --verbose
```

## Troubleshooting

### "SKIP: Skipping integration test"

This is normal if you haven't set the required environment variables.
The tests will skip automatically.

### "Failed to create collection"

Check that:

1. API key is valid
2. URL is correct (should be full HTTPS URL)
3. You have permission to create collections

### "Got 0 image results"

Check that:

1. Image collection has documents with embeddings
2. Documents have searchable content (not just base64 image data)
3. Query matches document content

Run with `--verbose` to see debug output:

```bash
weave cols query TextCol ImageCol "query" --top_k 5 \
  --top_k_images 2 --verbose
```

## CI/CD

These tests are included in the integration test suite but will skip if
credentials aren't available:

```yaml
# GitHub Actions example
env:
  WEAVIATE_CLOUD_API_KEY: ${{ secrets.WEAVIATE_CLOUD_API_KEY }}
  WEAVIATE_CLOUD_URL: ${{ secrets.WEAVIATE_CLOUD_URL }}

run: go test ./tests/integration/... -v -timeout 10m
```
