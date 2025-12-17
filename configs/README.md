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

### config.chroma-local.yaml

Configuration for Chroma running locally.

- Default URL: `http://localhost:8000`
- No authentication required for local setup
- Requires Docker/Podman: `podman run -d -p 8000:8000 chromadb/chroma:0.6.2`

### config.chroma-cloud.yaml

Configuration for Chroma Cloud.

- Requires Chroma Cloud account
- Set `CHROMA_CLOUD_URL` and `CHROMA_CLOUD_API_KEY`

### config.qdrant-local.yaml

Configuration for Qdrant running locally.

- Default URL: `http://localhost:6334` (gRPC)
- No authentication required for local setup
- Requires Docker/Podman: `podman run -d -p 6333:6333 -p 6334:6334 qdrant/qdrant`

### config.qdrant-cloud.yaml

Configuration for Qdrant Cloud.

- Requires Qdrant Cloud account
- Set `QDRANT_URL` and `QDRANT_API_KEY`

### config.neo4j-local.yaml

Configuration for Neo4j running locally.

- Default URL: `bolt://localhost:7687`
- Requires authentication (username/password)
- Requires Docker/Podman: `podman run -d -p 7474:7474 -p 7687:7687 -e
  NEO4J_AUTH=neo4j/password neo4j:latest`

### config.neo4j-cloud.yaml

Configuration for Neo4j Cloud (Aura).

- Requires Neo4j Aura account
- Set `NEO4J_CLOUD_URL`, `NEO4J_CLOUD_USERNAME`, and `NEO4J_CLOUD_PASSWORD`

### config.elasticsearch-local.yaml

Configuration for Elasticsearch running locally.

- Default URL: `http://localhost:9200`
- No authentication required for local setup
- Requires Docker/Podman:

  ```bash
  docker run -d -p 9200:9200 \
    -e "discovery.type=single-node" \
    -e "xpack.security.enabled=false" \
    docker.elastic.co/elasticsearch/elasticsearch:8.11.0
  ```

### config.elasticsearch-cloud.yaml

Configuration for Elasticsearch Cloud (Elastic Cloud).

- Requires Elastic Cloud account
- Set `ELASTICSEARCH_CLOUD_ADDRESS`, `ELASTICSEARCH_CLOUD_API_KEY`
- Alternative: Use `ELASTICSEARCH_CLOUD_USERNAME` and `ELASTICSEARCH_CLOUD_PASSWORD`

### config.opensearch-local.yaml

Configuration for OpenSearch running locally.

- Default URL: `http://localhost:9200`
- No authentication required for local setup
- Requires Docker/Podman:
  `podman run -d -p 9200:9200 -p 9600:9600 -e
  "discovery.type=single-node" opensearchproject/opensearch:latest`

### config.opensearch-cloud.yaml

Configuration for OpenSearch Cloud (AWS OpenSearch Service).

- Requires AWS OpenSearch Service domain
- Set `OPENSEARCH_CLOUD_ADDRESS`, `OPENSEARCH_CLOUD_USERNAME`, and `OPENSEARCH_CLOUD_PASSWORD`
- Alternative: Use `OPENSEARCH_CLOUD_API_KEY` instead of username/password

### config.pinecone.yaml

Configuration for Pinecone (cloud-only serverless).

- Requires Pinecone account at <https://app.pinecone.io>
- Set `PINECONE_API_KEY`
- Note: No local deployment option - Pinecone is cloud-only

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

# Chroma Local
CHROMA_URL="http://localhost:8000"

# Chroma Cloud
CHROMA_CLOUD_URL="https://your-instance.chroma.cloud"
CHROMA_CLOUD_API_KEY="your-api-key"

# Qdrant Local
QDRANT_HOST="localhost"
QDRANT_GRPC_PORT="6334"

# Qdrant Cloud
QDRANT_URL="https://your-cluster.cloud.qdrant.io:6334"
QDRANT_API_KEY="your-api-key"

# Neo4j Local
NEO4J_URL="bolt://localhost:7687"
NEO4J_USERNAME="neo4j"
NEO4J_PASSWORD="yourpassword"

# Neo4j Cloud (Aura)
NEO4J_CLOUD_URL="neo4j+s://xxxxx.databases.neo4j.io"
NEO4J_CLOUD_USERNAME="neo4j"
NEO4J_CLOUD_PASSWORD="your-password"

# Elasticsearch Local
ELASTICSEARCH_LOCAL_ADDRESS="http://localhost:9200"

# Elasticsearch Cloud (Elastic Cloud)
ELASTICSEARCH_CLOUD_ADDRESS="https://my-deployment.es.us-central1.gcp.cloud.es.io:9243"
ELASTICSEARCH_CLOUD_API_KEY="your-api-key"
# Alternative: ELASTICSEARCH_CLOUD_USERNAME="elastic" and ELASTICSEARCH_CLOUD_PASSWORD="your-password"

# OpenSearch Local
OPENSEARCH_LOCAL_ADDRESS="http://localhost:9200"

# OpenSearch Cloud (AWS OpenSearch Service)
OPENSEARCH_CLOUD_ADDRESS="https://search-mydomain.us-east-1.es.amazonaws.com"
OPENSEARCH_CLOUD_USERNAME="admin"
OPENSEARCH_CLOUD_PASSWORD="your-password"
# Alternative: OPENSEARCH_CLOUD_API_KEY="your-api-key"

# Pinecone (cloud-only)
PINECONE_API_KEY="your-pinecone-api-key"

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
weave cols ls --weaviate            # Use Weaviate only
weave cols ls --milvus-local        # Use Milvus local only
weave cols ls --chroma-local        # Use Chroma local only
weave cols ls --qdrant-local        # Use Qdrant local only
weave cols ls --neo4j-local         # Use Neo4j local only
weave cols ls --elasticsearch-local # Use Elasticsearch local only
weave cols ls --opensearch-local    # Use OpenSearch local only
weave cols ls --pinecone            # Use Pinecone only
weave cols ls --all                 # Use all configured databases
```

## See Also

- [User Guide](../docs/USER_GUIDE.md) - Complete documentation
- [VDB Support Matrix](../docs/VDB_SUPPORT.md) - Database comparison
- Database-specific documentation in `docs/` directory
