# MongoDB Atlas Setup Guide

MongoDB Atlas with Vector Search provides a familiar NoSQL database with
vector search capabilities. Best for teams already using MongoDB.

## Status

🧪 **Experimental** - Functional but requires manual vector search index setup

## Prerequisites

- Weave CLI installed ([installation guide](../../README.md#installation))
- MongoDB Atlas account (free tier available: M0)
- OpenAI API key for automatic embeddings (required)

## Setup Steps

### 1. Create MongoDB Atlas Cluster

1. Go to [cloud.mongodb.com](https://cloud.mongodb.com)
2. Create a free account
3. Create a new cluster (M0 free tier works)
4. Create a database user with read/write permissions
5. Add your IP address to network access list (or allow 0.0.0.0/0 for testing)

### 2. Get Connection String

1. Click "Connect" on your cluster
2. Choose "Connect your application"
3. Copy the connection string:

```text
mongodb+srv://username:password@cluster.mongodb.net/?appName=weave-cli
```

### 3. Create Vector Search Index

**This is a critical step - semantic search won't work without it!**

1. Go to your cluster → "Search" tab
2. Click "Create Search Index"
3. Choose "JSON Editor"
4. Use this configuration:

```json
{
  "name": "vector_index",
  "type": "vectorSearch",
  "definition": {
    "fields": [
      {
        "type": "vector",
        "path": "embedding",
        "numDimensions": 1536,
        "similarity": "cosine"
      }
    ]
  }
}
```

1. Select your database and collection
2. Create index (takes 1-2 minutes to build)

See [ATLAS_SETUP.md](./ATLAS_SETUP.md) for detailed instructions with screenshots.

### 4. Configure Weave CLI

**Interactive Setup**:

```bash
weave config create --env

# When prompted, enter:
# - MONGODB_URI: mongodb+srv://username:password@cluster.mongodb.net/?appName=weave-cli
# - MONGODB_DATABASE: weave-cli (or your database name)
# - OPENAI_API_KEY: sk-...
```

**Manual Setup**:

```bash
# Add to .env or export
export MONGODB_URI="mongodb+srv://username:password@cluster.mongodb.net/?appName=weave-cli"
export MONGODB_DATABASE="weave-cli"
export OPENAI_API_KEY="sk-..."  # Required for embeddings
```

### 5. Verify Connection

```bash
weave health check --mongodb-cloud
```

## Usage Examples

### Create Collection

```bash
# Create a collection
weave cols create MyDocs --mongodb-cloud

# ⚠️ Remember to create vector search index in Atlas UI for this collection!
```

### Add Documents

```bash
# Add a single document
weave docs create MyDocs ./document.txt --mongodb-cloud

# Add PDF
weave docs create MyDocs ./document.pdf --mongodb-cloud

# Batch add
weave docs create MyDocs ./docs/*.txt --mongodb-cloud
```

### Search

```bash
# Vector search (requires vector index!)
weave cols query MyDocs "machine learning" --mongodb-cloud

# With limit
weave cols query MyDocs "AI concepts" --limit 5 --mongodb-cloud
```

### List and Manage

```bash
# List collections
weave cols ls --mongodb-cloud

# List documents
weave docs ls MyDocs --mongodb-cloud

# Delete collection
weave cols delete MyDocs --mongodb-cloud
```

## Important Notes

### Vector Search Index Required

**You MUST create a vector search index for each collection** in the
Atlas UI before semantic search will work.

- Without index: `weave docs create` works, but `weave cols query` fails
- Create index using configuration from Step 3 above
- Index must match collection name exactly

### Free Tier Limitations

MongoDB Atlas M0 (free tier):

- 512 MB storage
- Shared RAM
- Vector search supported but slower
- Consider M10+ for production

### Connection String Format

Must include:

- Protocol: `mongodb+srv://`
- Username and password
- Cluster hostname
- Optional: database name, parameters

## Troubleshooting

### Connection Issues

**"connection refused" or "timeout"**:

- Check IP whitelist in Network Access
- Verify username/password
- Ensure cluster is running (not paused)

**"authentication failed"**:

- Verify credentials in connection string
- Check user has correct permissions
- Ensure database user is created

### Search Issues

**"vector search index not found"**:

- Create vector search index in Atlas UI (Step 3)
- Wait 1-2 minutes for index to build
- Verify index name matches configuration

**"dimension mismatch"**:

- Index dimensions must match embedding model
- OpenAI ada-002: 1536 dimensions
- Update index if using different model

**"no results" on valid query**:

- Check if documents have embeddings
- Verify index is built and active
- Test with Atlas search tester

### Common Errors

**"database/collection does not exist"**:

- Collection auto-created on first document insert
- No need to pre-create in Atlas UI

**"index build failed"**:

- Check index definition JSON
- Ensure collection has documents with `embedding` field
- Try rebuilding index

## Advanced Configuration

### Custom Vector Dimensions

For models other than OpenAI ada-002:

```bash
# In config.yaml
databases:
  vectordatabases:
    - name: mongodb-cloud
      type: mongodb-cloud
      uri: ${MONGODB_URI}
      database: ${MONGODB_DATABASE}
      vector_dimensions: 768  # e.g., for sentence-transformers
```

Update your vector search index accordingly!

### Multiple Indexes

Create separate indexes for different embedding models:

```json
{
  "name": "vector_index_openai",
  "type": "vectorSearch",
  "definition": {
    "fields": [{"type": "vector", "path": "embedding_openai", "numDimensions": 1536, "similarity": "cosine"}]
  }
}
```

### Connection Options

Add to connection string:

```text
mongodb+srv://user:pass@cluster.net/?retryWrites=true&w=majority&appName=weave-cli
```

## Performance Tips

1. **Use M10+ for production**: Better performance, dedicated resources
2. **Index early**: Create vector index before adding documents
3. **Batch operations**: Use batch document creation for better throughput
4. **Monitor usage**: Check Atlas metrics for performance insights

## Resources

- [MongoDB Atlas Vector Search Docs](https://www.mongodb.com/docs/atlas/atlas-vector-search/vector-search-overview/)
- [Atlas Setup Guide with Screenshots](./ATLAS_SETUP.md)
- [MongoDB Atlas Dashboard](https://cloud.mongodb.com)
- [MongoDB Community Forum](https://www.mongodb.com/community/forums/)

## Migration

### From Other Vector DBs

```bash
# Export from source
weave docs ls MyCollection --json > export.json --source-vdb

# Import to MongoDB
# (Manual process - parse JSON and create documents)
```

## Next Steps

- See [ATLAS_SETUP.md](./ATLAS_SETUP.md) for detailed setup with screenshots
- Review vector search best practices
- Check MongoDB Atlas documentation for optimization
