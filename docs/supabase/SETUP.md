# Supabase Setup Guide

Supabase (PostgreSQL with pgvector extension) is a feature-complete vector
database option for Weave CLI. It's cost-effective and integrates well with
existing PostgreSQL workflows.

## Status

🟡 **Alpha** - Functional and feature complete, recommended for development and testing

## Prerequisites

- Weave CLI installed ([installation guide](../../README.md#installation))
- Supabase project (free tier available)
- OpenAI API key for automatic embeddings (optional but recommended)

## Setup Steps

### 1. Create Supabase Project

1. Go to [supabase.com](https://supabase.com)
2. Create a free account
3. Create a new project
4. Note your project details:
   - Project URL
   - Anon/Public API key
   - Database password

### 2. Enable pgvector Extension

In your Supabase project SQL editor:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### 3. Get Database Connection Details

**Option A: Direct Database Connection** (requires IPv6):

```text
postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres
```

**Option B: Connection Pooler** (recommended, works without IPv6):

1. Go to Project Settings → Database → Connection Pooling
2. Copy the connection string:

```text
postgresql://postgres.[project].[string]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres
```

### 4. Configure Weave CLI

**Interactive Setup**:

```bash
weave config create --env

# When prompted, enter:
# - SUPABASE_DATABASE_URL: (your connection string from step 3)
# - SUPABASE_DATABASE_KEY: (your anon key)
# - OPENAI_API_KEY: sk-...
```

**Manual Setup**:

```bash
# Add to .env or export
export SUPABASE_DATABASE_URL="postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
export SUPABASE_DATABASE_KEY="your-supabase-anon-key"
export OPENAI_API_KEY="sk-..."  # Optional

# Or using pooler (recommended)
export SUPABASE_DATABASE_URL="postgresql://postgres.[project].[string]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres"
```

### 5. Verify Connection

```bash
weave health check --supabase-cloud
```

## Usage Examples

### Create Collection

```bash
# Create a collection (creates table with vector column)
weave cols create MyDocs --supabase-cloud
```

### Add Documents

```bash
# Add a single document
weave docs create MyDocs ./document.txt --supabase-cloud

# Add PDF
weave docs create MyDocs ./document.pdf --supabase-cloud

# Batch add
weave docs create MyDocs ./docs/*.txt --supabase-cloud
```

### Search

```bash
# Vector search
weave cols query MyDocs "machine learning" --supabase-cloud

# With limit
weave cols query MyDocs "AI concepts" --limit 5 --supabase-cloud
```

### List and Manage

```bash
# List collections
weave cols ls --supabase-cloud

# List documents
weave docs ls MyDocs --supabase-cloud

# Delete collection
weave cols delete MyDocs --supabase-cloud
```

## Important Notes

### IPv6 Requirement

Supabase database endpoints are IPv6-only. If your network doesn't support IPv6:

1. Use the **connection pooler** instead (see Step 3, Option B)
2. The pooler works over IPv4 and provides connection pooling benefits

### Security

- Use anon/public key for read operations
- Use service role key for admin operations (optional)
- Never commit keys to version control

### Performance

- Supabase free tier: 500MB database, 2GB bandwidth
- Paid tiers: Better performance, more storage
- Connection pooler recommended for production

## Troubleshooting

### Connection Issues

**"connection refused" or "timeout"**:

- Check if using IPv6 (try connection pooler)
- Verify database password is correct
- Check if project is paused (free tier auto-pauses)

**"extension not found"**:

```sql
-- Run in Supabase SQL editor
CREATE EXTENSION IF NOT EXISTS vector;
```

**"authentication failed"**:

- Verify DATABASE_URL has correct password
- Check DATABASE_KEY is the anon key, not service role key (unless intended)

### Common Errors

**"table does not exist"**:

- Collection not created yet
- Run: `weave cols create YourCollectionName --supabase-cloud`

**"column embedding does not exist"**:

- Schema mismatch
- Drop and recreate collection

## Advanced Configuration

### Custom Vector Dimensions

Edit `config.yaml`:

```yaml
databases:
  vectordatabases:
    - name: supabase-cloud
      type: supabase-cloud
      database_url: ${SUPABASE_DATABASE_URL}
      database_key: ${SUPABASE_DATABASE_KEY}
      vector_dimensions: 1536  # Default for OpenAI ada-002
```

### Using Service Role Key

For admin operations:

```bash
export SUPABASE_DATABASE_KEY="your-service-role-key"
```

⚠️ **Warning**: Service role key bypasses Row Level Security. Use with caution.

## Migration from Other VDBs

### From Weaviate

```bash
# Export from Weaviate
weave docs ls MyCollection --json > export.json

# Import to Supabase
# (TODO: Add import command when available)
```

## Resources

- [Supabase Documentation](https://supabase.com/docs)
- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [Supabase Dashboard](https://app.supabase.com)
- [Additional Weave/Supabase Docs](./README.md)

## Limitations

- No hybrid search (BM25 + vector) yet (in development)
- Batch operations slower than specialized vector DBs
- Query performance depends on index configuration

## Next Steps

- See [BM25_IMPROVEMENT.md](./BM25_IMPROVEMENT.md) for hybrid search plans
- Check [TESTING.md](./TESTING.md) for test coverage
- Review [README.md](./README.md) for detailed documentation
