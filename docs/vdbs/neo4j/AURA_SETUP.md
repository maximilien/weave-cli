# Neo4j Aura Setup Guide

This guide walks you through setting up Neo4j Aura (cloud) for vector search with weave-cli.

## Prerequisites

- Neo4j Aura account (free tier available)
- OpenAI API key for generating embeddings

## Quick Start

### 1. Create Neo4j Aura Instance

1. Sign up at [Neo4j Aura](https://console.neo4j.io)
2. Click **New Instance**
3. Choose instance type:
   - **AuraDB Free**: Perfect for development and testing (no credit card required)
   - **AuraDB Professional**: Production workloads with more resources
   - **AuraDB Enterprise**: Advanced features and support
4. Configure instance:
   - **Instance Name**: `weave-cli` (or your choice)
   - **Region**: Choose closest to your location
   - **Database Version**: Latest stable version
5. Click **Create**

### 2. Save Connection Credentials

⚠️ **Important**: Save these credentials immediately - they're only shown once!

After instance creation, you'll see:
- **Connection URI**: `neo4j+s://xxxxx.databases.neo4j.io`
- **Username**: `neo4j` (default)
- **Generated Password**: Long random string

Example:
```
Connection URI: neo4j+s://abc123de.databases.neo4j.io
Username: neo4j
Password: xJ9kL2mN4pQ6rS8tV0wX2yZ4aB6cD8eF
```

**Save these credentials securely!** You'll need them to connect.

### 3. Wait for Instance to Start

1. Instance status will show **Creating** → **Running**
2. This usually takes 1-2 minutes
3. You can explore the instance in Neo4j Browser while waiting

### 4. Open Neo4j Browser

1. Click **Open** next to your instance
2. Log in with your credentials
3. This is where you'll run Cypher queries and manage data

### 5. Create Vector Index

Vector indexes enable efficient similarity search. Neo4j supports vector indexes on node properties.

#### Using Neo4j Browser (Recommended)

1. Open Neo4j Browser for your instance
2. Run this Cypher query to create a vector index:

```cypher
// Create vector index for documents
CREATE VECTOR INDEX documentEmbeddings
FOR (d:Document)
ON d.embedding
OPTIONS {
  indexConfig: {
    `vector.dimensions`: 1536,
    `vector.similarity_function`: 'cosine'
  }
}
```

3. For image embeddings (if needed):

```cypher
// Create vector index for images
CREATE VECTOR INDEX imageEmbeddings
FOR (i:Image)
ON i.embedding
OPTIONS {
  indexConfig: {
    `vector.dimensions`: 512,
    `vector.similarity_function`: 'cosine'
  }
}
```

4. Verify indexes are created:

```cypher
SHOW INDEXES
```

You should see your vector indexes in the list.

#### Index Configuration Explained

- **Node Label**: `Document` or `Image` - The type of nodes to index
- **Property**: `embedding` - The property containing vector embeddings
- **Dimensions**: `1536` - For OpenAI text-embedding-ada-002 (change to 3072 for text-embedding-3-large, or 512 for CLIP images)
- **Similarity Function**: `cosine` - Distance metric (options: cosine, euclidean)

### 6. Configure weave-cli

#### Option A: Using .env file (Recommended)

Add to your `.env` file:

```bash
# Neo4j Aura Configuration
NEO4J_CLOUD_URI="neo4j+s://abc123de.databases.neo4j.io"
NEO4J_CLOUD_USERNAME="neo4j"
NEO4J_CLOUD_PASSWORD="xJ9kL2mN4pQ6rS8tV0wX2yZ4aB6cD8eF"
VECTOR_DB_TYPE="neo4j"
```

#### Option B: Using config.yaml

```yaml
databases:
  default: neo4j
  vector_databases:
    - name: neo4j
      type: neo4j
      url: "neo4j+s://abc123de.databases.neo4j.io"
      username: "neo4j"
      password: "xJ9kL2mN4pQ6rS8tV0wX2yZ4aB6cD8eF"
      vector_dimensions: 1536
      similarity_metric: cosine
      timeout: 60  # Cloud connections need longer timeouts
```

#### Option C: Interactive Setup

```bash
# Configure only Neo4j Aura variables (smart filtering)
weave config create --env --neo4j-cloud

# Follow prompts to enter credentials
```

### 7. Test Connection

```bash
# Test Neo4j Aura connection
VECTOR_DB_TYPE=neo4j weave health check --neo4j-cloud

# Create a collection (creates Document nodes)
VECTOR_DB_TYPE=neo4j weave collection create WeaveDocs --type text --neo4j-cloud

# List collections
VECTOR_DB_TYPE=neo4j weave collection list --neo4j-cloud

# Add a document
echo "Hello from Neo4j Aura!" > test.txt
VECTOR_DB_TYPE=neo4j weave docs create WeaveDocs test.txt --neo4j-cloud

# Search with vector similarity
VECTOR_DB_TYPE=neo4j weave query semantic WeaveDocs "greeting message" --neo4j-cloud
```

## Vector Index Best Practices

### Choosing Vector Dimensions

| Embedding Model | Dimensions | Notes |
|----------------|-----------|-------|
| text-embedding-ada-002 | 1536 | Default, good balance |
| text-embedding-3-small | 512-1536 | Configurable, faster |
| text-embedding-3-large | 256-3072 | Configurable, best quality |
| CLIP (images) | 512-768 | Depends on model |

### Choosing Similarity Function

- **cosine**: Best for normalized embeddings (recommended for OpenAI)
- **euclidean**: Good for unnormalized vectors, measures absolute distance

### Index Limitations

**AuraDB Free**:
- 1 instance per account
- Limited to 200,000 nodes and 400,000 relationships
- Suitable for development and small projects
- No backups (manual export recommended)

**AuraDB Professional**:
- Multiple instances
- More storage and memory
- Automatic backups
- Good for production applications

**AuraDB Enterprise**:
- Advanced security features
- Multi-region support
- SLA guarantees
- Best for critical production workloads

## Vector Search Workflow

Neo4j stores documents as nodes with properties:

```cypher
// Example document node
(:Document {
  id: "doc1",
  text: "Original document text",
  embedding: [0.123, 0.456, ...],  // 1536 dimensions
  metadata: {...}
})
```

Vector search finds similar documents:

```cypher
// Vector similarity search
CALL db.index.vector.queryNodes('documentEmbeddings', 10, $queryEmbedding)
YIELD node, score
RETURN node.id, node.text, score
ORDER BY score DESC
```

## Troubleshooting

### Connection Issues

**Error**: `connection refused` or `timeout`
- **Cause**: Wrong URI, instance not running, or network issues
- **Solution**:
  1. Verify instance status is **Running** in Aura console
  2. Check connection URI matches exactly (including `neo4j+s://`)
  3. Ensure TLS is enabled (`neo4j+s://` not `neo4j://`)
  4. Check your network/firewall allows outbound HTTPS

**Error**: `authentication failed` or `unauthorized`
- **Cause**: Wrong username or password
- **Solution**:
  1. Verify credentials in Aura console
  2. If you lost the password, you can reset it:
     - Go to instance → **Reset Password**
     - Save the new password immediately
  3. Update your `.env` or config with new credentials

**Error**: `certificate signed by unknown authority`
- **Cause**: TLS certificate validation issue
- **Solution**: Use `neo4j+s://` (with TLS) in connection URI

### Vector Index Issues

**Error**: `index not found` or vector search returns no results
- **Solution**:
  1. Verify index exists: Run `SHOW INDEXES` in Neo4j Browser
  2. Check index name matches what weave-cli expects
  3. Ensure index is **ONLINE** (not creating or failed)

**Error**: `dimensions mismatch`
- **Solution**:
  1. Check embedding dimensions match index configuration
  2. OpenAI ada-002 uses 1536, not 512 or 3072
  3. Recreate index with correct dimensions:
     ```cypher
     DROP INDEX documentEmbeddings IF EXISTS;
     CREATE VECTOR INDEX documentEmbeddings ...
     ```

**Error**: `failed to create vector index`
- **Cause**: Syntax error or unsupported version
- **Solution**:
  1. Ensure Neo4j version 5.13+ (vector indexes require 5.13+)
  2. Check Cypher syntax is correct
  3. Verify instance is running Neo4j 5.x, not 4.x

### Performance Issues

**Slow queries**:
- Upgrade to Professional tier for more memory
- Reduce search result limit (fewer nodes to return)
- Add property indexes for metadata filtering
- Use approximate nearest neighbor (ANN) instead of exact search

**High memory usage**:
- Vector indexes consume memory
- Reduce `vector.dimensions` if possible
- Upgrade instance size in Aura console

## Advanced Configuration

### Multiple Node Labels

Create separate vector indexes for different node types:

```cypher
// Documents index
CREATE VECTOR INDEX documentEmbeddings
FOR (d:Document) ON d.embedding
OPTIONS {indexConfig: {`vector.dimensions`: 1536, `vector.similarity_function`: 'cosine'}};

// Images index
CREATE VECTOR INDEX imageEmbeddings
FOR (i:Image) ON i.embedding
OPTIONS {indexConfig: {`vector.dimensions`: 512, `vector.similarity_function`: 'cosine'}};

// Code snippets index
CREATE VECTOR INDEX codeEmbeddings
FOR (c:Code) ON c.embedding
OPTIONS {indexConfig: {`vector.dimensions`: 1536, `vector.similarity_function`: 'cosine'}};
```

### Hybrid Search (Vector + Text)

Combine vector similarity with full-text search:

```cypher
// Create full-text index
CREATE FULLTEXT INDEX documentText
FOR (d:Document) ON EACH [d.text, d.title];

// Hybrid search
CALL {
  // Vector search
  CALL db.index.vector.queryNodes('documentEmbeddings', 10, $embedding)
  YIELD node, score AS vectorScore
  RETURN node, vectorScore
}
WITH node, vectorScore
CALL {
  // Text search
  WITH node
  CALL db.index.fulltext.queryNodes('documentText', $query)
  YIELD node AS textNode, score AS textScore
  WHERE id(node) = id(textNode)
  RETURN textScore
}
RETURN node, vectorScore, textScore, (vectorScore + textScore) AS combinedScore
ORDER BY combinedScore DESC
LIMIT 10
```

### Metadata Filtering

Add property indexes for efficient metadata filtering:

```cypher
// Create property indexes
CREATE INDEX documentCategory FOR (d:Document) ON (d.category);
CREATE INDEX documentLanguage FOR (d:Document) ON (d.language);

// Vector search with metadata pre-filter
MATCH (d:Document)
WHERE d.category = 'technical' AND d.language = 'en'
WITH d
CALL db.index.vector.queryNodes('documentEmbeddings', 10, $embedding, d)
YIELD node, score
RETURN node, score
ORDER BY score DESC
```

## Migration from Other VDBs

To migrate from Weaviate, Pinecone, or other VDBs to Neo4j Aura:

1. Export data from source VDB
2. Set up Neo4j Aura instance and create vector indexes
3. Update configuration to use Neo4j
4. Transform data to Neo4j node format
5. Import using `weave document create` or bulk Cypher queries

## Instance Management

### Pause Instance

To save costs, you can pause your instance when not in use:

1. Go to Aura console
2. Click **...** next to instance → **Pause**
3. Resume when needed (no data loss)

**Note**: Free tier instances auto-pause after inactivity.

### Backup Data

**AuraDB Professional/Enterprise**: Automatic backups enabled

**AuraDB Free**: Manual export recommended
```cypher
// Export all data
CALL apoc.export.json.all("backup.json", {})
```

### Monitor Usage

Check instance metrics in Aura console:
- **Database Size**: Total nodes and relationships
- **Memory Usage**: Consumed vs available
- **Query Performance**: Slow queries and optimization tips

## Resources

- [Neo4j Aura Console](https://console.neo4j.io)
- [Neo4j Vector Index Documentation](https://neo4j.com/docs/cypher-manual/current/indexes-for-vector-search/)
- [Neo4j Aura Free Tier](https://neo4j.com/cloud/aura-free/)
- [Neo4j Graph Academy](https://graphacademy.neo4j.com/) - Free courses
- [Neo4j Community Forum](https://community.neo4j.com/)

## Support

If you encounter issues:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review Neo4j Aura logs in the console
3. Visit [Neo4j Community Forum](https://community.neo4j.com/)
4. File an issue on [GitHub](https://github.com/maximilien/weave-cli/issues)
