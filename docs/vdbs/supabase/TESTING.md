# Testing Supabase Integration

Guide for running Supabase integration tests with the enhanced collection name preservation tests.

## Prerequisites

### 1. Supabase Project Setup

You need a Supabase project with:
- PostgreSQL database access
- `pgvector` extension enabled
- Service role key or anon key

### 2. Environment Variables

Set the following environment variables:

```bash
# Required for Supabase tests
export SUPABASE_DATABASE_URL="postgresql://postgres:[password]@db.[project].supabase.co:5432/postgres"
export SUPABASE_DATABASE_KEY="your-supabase-service-role-key"

# Optional: For embedding tests
export OPENAI_API_KEY="your-openai-api-key"
```

**Important: IPv6 vs IPv4**

If your network doesn't support IPv6, use the connection pooler URL instead:

```bash
export SUPABASE_DATABASE_URL="postgresql://postgres.[project].[string]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres"
```

You can find the pooler URL in: Supabase Dashboard → Project Settings → Database → Connection Pooling

### 3. Enable pgvector Extension

Run this SQL in your Supabase SQL editor:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

## Running Tests

### Run All Supabase Tests

```bash
./test.sh integration --supabase
```

This will run all Supabase integration tests including:
- Health check
- Collection name preservation (9 test cases)
- CRUD operations
- Search operations (semantic, BM25, hybrid, metadata)
- Embedding tests (OpenAI, no vectorizer)
- Schema operations
- Batch operations

### Run Specific Test

```bash
# Run only collection name preservation tests
go test -v -timeout=2m ./tests -run="TestSupabaseIntegration/CollectionNamePreservation"

# Run only a specific naming pattern test
go test -v -timeout=2m ./tests -run="TestSupabaseIntegration/CollectionNamePreservation/Mixed_case_with_underscores"
```

### Run with Verbose Output

```bash
go test -v -timeout=2m ./tests -run="TestSupabaseIntegration"
```

## Test Coverage

### Collection Name Preservation Tests

The enhanced `CollectionNamePreservation` test verifies:

**9 Naming Pattern Test Cases:**
1. `TestCollection_MixedCase` - Mixed case with underscores
2. `My_Test_Collection` - Multiple underscores
3. `Collection123_Test` - Numbers with underscores
4. `ALLCAPS_COLLECTION` - All uppercase
5. `lowercase_collection` - All lowercase
6. `CamelCaseCollection` - Camel case (no underscores)
7. `snake_case_collection` - Snake case
8. `Collection-With-Hyphens` - Hyphens preserved
9. `Mix_Case-And-Chars123` - Mixed: underscores, hyphens, numbers

**For Each Test Case, Verifies:**
- ✅ Collection creation with exact name
- ✅ Collection existence check
- ✅ Batch document creation (2 docs)
- ✅ Document retrieval by ID
- ✅ Document listing
- ✅ Collection count
- ✅ Metadata search
- ✅ Document update
- ✅ Document deletion
- ✅ Collection deletion

**Total Operations Per Test Case:** 11 operations × 9 test cases = **99 operations**

### Other Integration Tests

- **Health Check** - Database connectivity
- **CreateCollection** - Standard collection creation
- **CollectionExists** - Existence verification
- **ListCollections** - Collection enumeration
- **CreateDocument** - Single document creation
- **GetDocument** - Document retrieval
- **UpdateDocument** - Document modification
- **ListDocuments** - Document listing with pagination
- **GetCollectionCount** - Collection size
- **SearchByContent** - Content-based search
- **SearchByMetadata** - Metadata filtering
- **SearchBM25** - Full-text search (improved with ts_rank_cd)
- **SearchHybrid** - Combined semantic + BM25
- **SemanticSearchWithEmbeddings** - Vector similarity search
- **DocumentsWithDifferentEmbeddings** - OpenAI vs no vectorizer
- **GetSchema** - Schema retrieval
- **ValidateSchema** - Schema validation
- **CreateDocuments** - Batch document creation
- **DeleteDocumentsByMetadata** - Bulk deletion by metadata
- **DeleteDocument** - Single document deletion
- **DeleteCollection** - Collection removal

## Expected Output

### Successful Test Run

```
=== RUN   TestSupabaseIntegration
=== RUN   TestSupabaseIntegration/Health
✓ Health check passed
=== RUN   TestSupabaseIntegration/CollectionNamePreservation
=== RUN   TestSupabaseIntegration/CollectionNamePreservation/Mixed_case_with_underscores
✓ Created collection: TestCollection_MixedCase
✓ Collection exists check passed: TestCollection_MixedCase
✓ Created 2 documents in: TestCollection_MixedCase
✓ Document retrieval verified: TestCollection_MixedCase
✓ Document listing verified: TestCollection_MixedCase (2 docs)
✓ Collection count verified: TestCollection_MixedCase (count=2)
✓ Metadata search verified: TestCollection_MixedCase (2 results)
✓ Document update verified: TestCollection_MixedCase
✓ Document deletion verified: TestCollection_MixedCase
✅ All operations verified for collection: TestCollection_MixedCase
=== RUN   TestSupabaseIntegration/CollectionNamePreservation/Multiple_underscores
... (similar for each test case)
```

### Skipped Tests

If `OPENAI_API_KEY` is not set, embedding-related tests will be skipped:

```
=== RUN   TestSupabaseIntegration/SemanticSearchWithEmbeddings
--- SKIP: TestSupabaseIntegration/SemanticSearchWithEmbeddings (0.00s)
    supabase_integration_test.go:326: Skipping semantic search test: OPENAI_API_KEY not set
```

## Troubleshooting

### Connection Errors

**Problem:** `connection refused` or `timeout`

**Solutions:**
1. Check if your network supports IPv6
2. Use connection pooler URL (IPv4)
3. Verify firewall/VPN settings
4. Check Supabase project status

### Authentication Errors

**Problem:** `authentication failed` or `permission denied`

**Solutions:**
1. Verify `SUPABASE_DATABASE_KEY` is correct
2. Use service role key (not anon key) for full access
3. Check database password in connection URL

### pgvector Extension Missing

**Problem:** `extension "vector" does not exist`

**Solution:**
```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### Test Timeouts

**Problem:** Tests timeout after 2 minutes

**Solutions:**
1. Increase timeout: `go test -timeout=5m ...`
2. Check network latency
3. Run specific test instead of full suite

## Performance Notes

### Test Duration

- **Fast tests** (no embeddings): ~30-60 seconds
- **With embeddings**: ~2-3 minutes (due to OpenAI API calls)
- **CollectionNamePreservation**: ~20-40 seconds (9 × ~3-4 seconds each)

### Rate Limits

- OpenAI API: 3-5 requests per minute (free tier)
- Supabase: Generally no rate limits for normal testing

### Cleanup

All tests clean up after themselves:
- Collections created during tests are deleted
- Documents are removed
- No persistent state between test runs

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Supabase Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Supabase Tests
        env:
          SUPABASE_DATABASE_URL: ${{ secrets.SUPABASE_DATABASE_URL }}
          SUPABASE_DATABASE_KEY: ${{ secrets.SUPABASE_DATABASE_KEY }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: ./test.sh integration --supabase
```

## Next Steps

After tests pass:
1. Review test output for any warnings
2. Check Supabase dashboard for created/deleted collections
3. Verify performance metrics
4. Run with different Supabase regions/instances

## Related Documentation

- [Supabase TODO](TODO.md) - Remaining improvements
- [Supabase Name Fix](NAME_FIX.md) - Implementation details
- [VDB Support Matrix](../VDB_SUPPORT.md) - Feature compatibility
- [User Guide](../USER_GUIDE.md) - General usage
