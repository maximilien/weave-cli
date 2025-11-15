# MongoDB Atlas Setup Guide

This guide walks you through setting up MongoDB Atlas Vector Search for use with weave-cli.

## Prerequisites

- MongoDB Atlas account (free tier available)
- OpenAI API key for generating embeddings

## Quick Start

### 1. Create MongoDB Atlas Cluster

1. Sign up at [MongoDB Atlas](https://cloud.mongodb.com/)
2. Create a new cluster:
   - **Free Tier (M0)**: Perfect for development and testing
   - **Shared Tier (M2/M5)**: Better performance, still affordable
   - **Dedicated (M10+)**: Production workloads
3. Choose your cloud provider and region
4. Click "Create Cluster"

### 2. Configure Network Access

1. Go to **Network Access** in the left sidebar
2. Click **Add IP Address**
3. Options:
   - **Add Current IP Address**: For development
   - **Allow Access from Anywhere** (0.0.0.0/0): For testing (not recommended for production)
4. Click **Confirm**

### 3. Create Database User

1. Go to **Database Access** in the left sidebar
2. Click **Add New Database User**
3. Authentication Method: **Password**
4. Username: `weave-cli_db_user` (or your choice)
5. Password: Generate a secure password or use: `yh1de52mrDwRa3YA`
6. Database User Privileges: **Atlas admin** or **Read and write to any database**
7. Click **Add User**

### 4. Get Connection String

1. Go to **Database** in the left sidebar
2. Click **Connect** on your cluster
3. Choose **Drivers**
4. Select **Go** and version **1.13 or later**
5. Copy the connection string, it should look like:
   ```
   mongodb+srv://weave-cli_db_user:<password>@weave-cli.fy2ocan.mongodb.net/?appName=weave-cli
   ```
6. Replace `<password>` with your actual password

### 5. Create Vector Search Index

This is the most important step for vector search functionality.

#### Using Atlas UI (Recommended)

1. Go to **Database** > Your Cluster > **Browse Collections**
2. Create a database named `weave-cli` (or your choice)
3. Create a collection named `WeaveDocs`
4. Go to **Atlas Search** tab
5. Click **Create Search Index**
6. Choose **JSON Editor**
7. Index Name: `vector_index`
8. Database: `weave-cli`
9. Collection: `WeaveDocs`
10. Paste this index definition:

```json
{
  "fields": [
    {
      "type": "vector",
      "path": "embedding",
      "numDimensions": 1536,
      "similarity": "cosine"
    },
    {
      "type": "filter",
      "path": "metadata"
    }
  ]
}
```

11. Click **Create Search Index**
12. Wait for the index to become active (usually 1-2 minutes)

#### Index Configuration Explained

- **path**: `"embedding"` - The field containing vector embeddings
- **numDimensions**: `1536` - For OpenAI text-embedding-ada-002 (change to 3072 for text-embedding-3-large)
- **similarity**: `"cosine"` - Distance metric (options: cosine, euclidean, dotProduct)
- **metadata filter**: Allows pre-filtering documents by metadata before vector search

#### For Image Collection

Repeat the same process for `WeaveImages` collection with the same index configuration.

### 6. Configure weave-cli

#### Option A: Using .env file (Recommended)

Add to your `.env` file:

```bash
# MongoDB Atlas Configuration
MONGODB_URI="mongodb+srv://weave-cli_db_user:yh1de52mrDwRa3YA@weave-cli.fy2ocan.mongodb.net/?appName=weave-cli"
MONGODB_DATABASE="weave-cli"
VECTOR_DB_TYPE="mongodb"
```

#### Option B: Using config.yaml

```yaml
databases:
  default: mongodb
  vector_databases:
    - name: mongodb
      type: mongodb
      url: "mongodb+srv://weave-cli_db_user:yh1de52mrDwRa3YA@weave-cli.fy2ocan.mongodb.net/?appName=weave-cli"
      database: "weave-cli"
      vector_dimensions: 1536
      similarity_metric: cosine
      timeout: 10
      collections:
        - name: WeaveDocs
          type: text
        - name: WeaveImages
          type: image
```

### 7. Test Connection

```bash
# Test MongoDB connection
VECTOR_DB_TYPE=mongodb weave health check

# Create a collection
VECTOR_DB_TYPE=mongodb weave collection create WeaveDocs --type text

# List collections
VECTOR_DB_TYPE=mongodb weave collection list
```

## Vector Search Index Best Practices

### Choosing Vector Dimensions

| Embedding Model | Dimensions | Notes |
|----------------|-----------|-------|
| text-embedding-ada-002 | 1536 | Default, good balance |
| text-embedding-3-small | 512-1536 | Configurable, faster |
| text-embedding-3-large | 256-3072 | Configurable, best quality |
| CLIP (images) | 512-768 | Depends on model |

### Choosing Similarity Metric

- **cosine**: Best for normalized embeddings (recommended for OpenAI)
- **euclidean**: Good for unnormalized vectors
- **dotProduct**: Fastest, requires normalized vectors

### Index Limitations

**Free Tier (M0)**:
- Max 3 search indexes per cluster
- Max 10M vector dimensions per index
- Suitable for development and small projects

**Shared Tier (M2/M5)**:
- More search indexes allowed
- Better performance
- Good for small production apps

**Dedicated (M10+)**:
- No significant limitations
- Best performance
- Recommended for production

## Troubleshooting

### Connection Issues

**Error**: `connection refused` or `timeout`
- **Solution**: Check Network Access settings, ensure your IP is whitelisted

**Error**: `authentication failed`
- **Solution**: Verify username and password, ensure user has correct privileges

### Vector Search Not Working

**Error**: `index not found` or `$vectorSearch failed`
- **Solution**: Ensure vector search index is created and active
- **Check**: Index name must be `vector_index` on the `embedding` field

**Error**: `dimensions mismatch`
- **Solution**: Ensure index `numDimensions` matches your embedding size (1536 for ada-002)

### Performance Issues

**Slow queries**:
- Upgrade to M2+ tier for better performance
- Reduce `numCandidates` in search queries
- Add metadata filters to reduce search space

## Advanced Configuration

### Multiple Collections

Create indexes for each collection that needs vector search:

```bash
# For each collection, create a vector search index with:
# - Index name: vector_index
# - Path: embedding
# - Dimensions: matching your embedding model
# - Similarity: cosine (recommended)
```

### Custom Similarity Functions

```json
{
  "fields": [
    {
      "type": "vector",
      "path": "embedding",
      "numDimensions": 1536,
      "similarity": "dotProduct"  // or "euclidean"
    }
  ]
}
```

### Metadata Filtering

Add filter paths to enable pre-filtering:

```json
{
  "fields": [
    {
      "type": "vector",
      "path": "embedding",
      "numDimensions": 1536,
      "similarity": "cosine"
    },
    {
      "type": "filter",
      "path": "metadata.category"
    },
    {
      "type": "filter",
      "path": "metadata.language"
    }
  ]
}
```

## Migration from Weaviate

To migrate from Weaviate to MongoDB:

1. Export data from Weaviate
2. Set up MongoDB Atlas and create indexes
3. Update configuration to use MongoDB
4. Import data using `weave document create`

## Resources

- [MongoDB Atlas Vector Search Documentation](https://www.mongodb.com/docs/atlas/atlas-vector-search/vector-search-overview/)
- [Creating Vector Search Indexes](https://www.mongodb.com/docs/atlas/atlas-vector-search/create-index/)
- [Vector Search Tutorial](https://www.mongodb.com/docs/atlas/atlas-vector-search/tutorials/)
- [MongoDB Atlas Free Tier](https://www.mongodb.com/cloud/atlas/pricing)

## Support

If you encounter issues:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review MongoDB Atlas logs in the UI
3. File an issue on [GitHub](https://github.com/maximilien/weave-cli/issues)
