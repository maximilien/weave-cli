# Neo4j Setup Guide

Neo4j is a graph database with vector search capabilities. Best for
applications that need both graph relationships and vector similarity search.

## Status

🧪 **Experimental** - Newly added, functional, limited testing

## Prerequisites

- Weave CLI installed
- Neo4j instance (local or cloud)
- OpenAI API key for embeddings

## Local Setup

### 1. Start Neo4j Locally

Using helper script:

```bash
./tools/vdb/local/neo4j.sh start
```

Or manually with Docker:

```bash
docker run -d \
  --name neo4j \
  -p 7474:7474 -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/your-password \
  neo4j:latest
```

### 2. Configure Weave

**Interactive Setup (Recommended)**:

```bash
# Configure only Neo4j local variables (smart filtering)
weave config create --env --neo4j-local

# Follow prompts to enter credentials
```

**Manual Setup**:

```bash
export NEO4J_LOCAL_URI="bolt://localhost:7687"
export NEO4J_LOCAL_USERNAME="neo4j"
export NEO4J_LOCAL_PASSWORD="your-password"
export OPENAI_API_KEY="sk-..."

weave health check --neo4j-local
```

## Cloud Setup (Neo4j Aura)

### 1. Create Aura Instance

1. Go to [console.neo4j.io](https://console.neo4j.io)
2. Create a free instance
3. Note connection URI and credentials

### 2. Configure Weave

**Interactive Setup (Recommended)**:

```bash
# Configure only Neo4j Aura variables (smart filtering)
weave config create --env --neo4j-cloud

# Follow prompts to enter credentials
```

**Manual Setup**:

```bash
export NEO4J_CLOUD_URI="neo4j+s://your-instance.databases.neo4j.io"
export NEO4J_CLOUD_USERNAME="neo4j"
export NEO4J_CLOUD_PASSWORD="your-generated-password"
export OPENAI_API_KEY="sk-..."

weave health check --neo4j-cloud
```

## Usage

```bash
# Create collection
weave cols create MyDocs --neo4j-local

# Add documents
weave docs create MyDocs ./document.txt --neo4j-local

# Vector search
weave cols query MyDocs "search query" --neo4j-local
```

## Resources

- [Neo4j README](./README.md)
- [Neo4j Documentation](https://neo4j.com/docs/)
- [Neo4j Aura](https://neo4j.com/cloud/aura/)
