# Integration Tests

Integration tests that require real vector database connections.

## Prerequisites

These tests require environment variables for connecting to vector databases:

### Weaviate Cloud

```bash
export WEAVIATE_CLOUD_API_KEY="your-api-key"
export WEAVIATE_CLOUD_URL="https://your-cluster.weaviate.network"
```

## Running Tests

### Run all integration tests

```bash
# With credentials set
go test ./tests/integration/... -v

# Skip if credentials not set (default behavior)
go test ./tests/integration/... -v
# Output: SKIP: Skipping integration test: WEAVIATE_CLOUD_API_KEY not set
```

### Run specific test

```bash
go test ./tests/integration/... -v -run TestTopKImagesFlag
```

### Run with timeout

```bash
go test ./tests/integration/... -v -timeout 5m
```

## Tests

### TestTopKImagesFlag

Tests the `--top_k_images` flag for multi-modal queries.

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
