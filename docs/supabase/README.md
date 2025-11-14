# Supabase Integration Documentation

This directory contains all documentation specific to Supabase PGVector integration.

> **⚠️ EXPERIMENTAL**: Supabase support is currently experimental and under active
> development. Some features may not work as expected. Please report issues at
> <https://github.com/maximilien/weave-cli/issues>

## Quick Start

```bash
# Set Supabase configuration
export SUPABASE_DATABASE_URL="postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
export SUPABASE_DATABASE_KEY="your-supabase-anon-key"

# Configure weave to use Supabase
weave config create --database-type supabase

# Verify Supabase connection
weave health check
```

## Documentation

- **[TESTING.md](TESTING.md)** - Complete guide for running Supabase integration tests
- **[NAME_FIX.md](NAME_FIX.md)** - Collection name preservation implementation details
- **[BM25_IMPROVEMENT.md](BM25_IMPROVEMENT.md)** - BM25 search enhancement plan
- **[TODO.md](TODO.md)** - Roadmap of remaining improvements

## Features

### ✅ Fully Supported

- Collection CRUD operations
- Document CRUD operations
- Semantic search (vector similarity)
- BM25 full-text search (with `ts_rank_cd` normalization)
- Hybrid search (semantic + BM25)
- Metadata filtering
- Batch operations
- Collection name preservation (mixed case, underscores, hyphens, numbers)

### ⚠️ Limitations

- **Embedding Providers**: Currently only OpenAI embeddings are supported
  - Weaviate supports: OpenAI, Cohere, Hugging Face, Google PaLM, AWS Bedrock, Jina AI
  - Supabase: Only OpenAI (see [TODO.md](TODO.md) for roadmap)
- **IPv6 Requirement**: Direct database connections require IPv6 (use connection pooler for IPv4)

### 🔄 Planned

See [TODO.md](TODO.md) for the complete roadmap of upcoming features and improvements.

## Architecture

Supabase integration uses:

- **PostgreSQL** - Primary database
- **pgvector Extension** - Vector similarity search
- **Full-Text Search** - Built-in PostgreSQL FTS with `tsvector` and `tsquery`
- **JSONB** - Flexible metadata storage

## See Also

- [VDB Support Matrix](../VDB_SUPPORT.md) - Feature comparison across all vector databases
- [User Guide](../USER_GUIDE.md) - General weave-cli usage
- [Vector DB Abstraction](../guides/VECTOR_DB_ABSTRACTION.md) - Multi-database architecture
