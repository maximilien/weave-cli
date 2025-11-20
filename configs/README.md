# Configuration Examples

This directory contains example configuration files for different vector
database backends supported by Weave CLI.

## Usage

To use one of these example configurations:

1. Copy the desired example to the project root as `config.yaml`:

   ```bash
   cp configs/config.weaviate.yaml config.yaml
   ```

2. Edit the copied file to add your credentials and settings

3. Or combine multiple databases by merging configurations:

   ```bash
   # Start with one example
   cp configs/config.weaviate.yaml config.yaml

   # Then manually add sections from other examples
   ```

## Available Examples

### config.weaviate.yaml

Configuration for Weaviate vector database (cloud and/or local).

- **Weaviate Cloud**: Requires `WEAVIATE_URL` and `WEAVIATE_API_KEY`
- **Weaviate Local**: Typically runs at `http://localhost:8080`

### config.milvus-local.yaml

Configuration for Milvus running locally.

- Default address: `localhost:19530`
- No authentication required for local setup
- Requires Docker or Milvus standalone installation

### config.milvus-cloud.yaml

Configuration for Milvus cloud (Zilliz Cloud).

- Requires Zilliz Cloud account
- Set `MILVUS_CLOUD_ADDRESS`, `MILVUS_CLOUD_USERNAME`, `MILVUS_CLOUD_PASSWORD`

### config.supabase.yaml

Configuration for Supabase with pgvector extension.

- Requires Supabase project
- Set `SUPABASE_DATABASE_URL` and `SUPABASE_DATABASE_KEY`

### config.mongodb.yaml

Configuration for MongoDB Atlas Vector Search.

- Requires MongoDB Atlas account with vector search enabled
- Set `MONGODB_URI` and `MONGODB_DATABASE`

## Environment Variables

Most sensitive values (API keys, passwords) should be set via environment
variables in your `.env` file rather than directly in `config.yaml`.

Example `.env` file:

```bash
# Weaviate
WEAVIATE_URL="https://your-cluster.weaviate.cloud"
WEAVIATE_API_KEY="your-api-key"

# Milvus Cloud
MILVUS_CLOUD_ADDRESS="your-cluster.zillizcloud.com:19530"
MILVUS_CLOUD_USERNAME="your-username"
MILVUS_CLOUD_PASSWORD="your-password"

# Supabase
SUPABASE_DATABASE_URL="postgresql://postgres:password@db.xxx.supabase.co:5432/postgres"
SUPABASE_DATABASE_KEY="your-anon-key"

# MongoDB
MONGODB_URI="mongodb+srv://user:pass@cluster.mongodb.net/?appName=weave-cli"
MONGODB_DATABASE="weave-cli"

# OpenAI (required for embeddings)
OPENAI_API_KEY="sk-..."
```

## Interactive Configuration

Instead of manually copying files, you can use Weave CLI's interactive
configuration commands:

```bash
# Create a new .env file interactively
weave config create --env

# Update an existing .env file
weave config update --env

# Show current configuration
weave config show

# List all configured databases
weave config list
```

## Priority Order

Configuration values are resolved in this priority order (highest first):

1. Command-line flags (e.g., `--weaviate-url`)
2. Environment variables from `--env` flag
3. Environment variables from `.env` file
4. Shell environment variables
5. Values in `config.yaml`
6. Built-in defaults

## Multiple Databases

You can configure multiple databases in a single `config.yaml` file:

```yaml
databases:
  weaviate-cloud:
    type: weaviate-cloud
    # ... weaviate settings

  milvus-local:
    type: milvus-local
    # ... milvus settings

  supabase:
    type: supabase
    # ... supabase settings
```

Then use flags to select which database to use:

```bash
weave cols ls --weaviate      # Use Weaviate only
weave cols ls --milvus-local  # Use Milvus local only
weave cols ls --all           # Use all configured databases
```

## See Also

- [User Guide](../docs/USER_GUIDE.md) - Complete documentation
- [VDB Support Matrix](../docs/VDB_SUPPORT.md) - Database comparison
- Database-specific documentation in `docs/` directory
