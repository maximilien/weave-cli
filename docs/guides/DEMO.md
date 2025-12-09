# Weave CLI Demo Script

**GitHub**: <https://github.com/maximilien/weave-cli>

---

## Quick Overview

Weave CLI is a fast, AI-powered command-line tool for managing vector databases
(VDBs). Built in Go for performance and ease of use (single binary).

**Key Features**:

- 🤖 **AI-Powered** - Natural language REPL interface with GPT-4o multi-agent
  system
- ⚡ **Fast & Easy** - Written in Go with simple CLI and real-time progress
  feedback
- 🌐 **Flexible** - Support for 8 vector databases (Weaviate, Milvus, Supabase,
  MongoDB, Chroma, Qdrant, Neo4j, OpenSearch)
- 📦 **Batch Processing** - Parallel processing of entire directories
- 📄 **PDF Support** - Intelligent text extraction and image processing
- 🔍 **Semantic Search** - Vector-based similarity search with natural language

---

## Demo Flow

### 1. Help & Getting Started

```bash
# Main help
weave -h

# Collections help
weave cols -h

# Documents help
weave docs -h

# Configuration help
weave config -h
```

---

### 2. Configuration Management

```bash
# Show current configuration
weave config show

# List all configured databases
weave config list

# List with detailed information
weave config list --details

# Filter by deployment type (NEW in v0.7.2)
weave config list --cloud          # Show only cloud databases
weave config list --local          # Show only local databases

# Sort databases (NEW in v0.7.2)
weave config list --sort-by name       # Alphabetical (default)
weave config list --sort-by type       # By database type
weave config list --sort-by deployment # By deployment type

# List collection schemas
weave config list-schemas

# Show specific schema
weave config show-schema WeaveDocs
weave config show-schema WeaveDocs --yaml
```

---

### 3. Health Checks

```bash
# Check health of all configured databases
weave health check

# Summary view with progressive output (NEW in v0.7.2)
weave health check -S              # Shorthand
weave health check --summary       # Long form

# Filter health checks (NEW in v0.7.2)
weave health check --cloud         # Check only cloud databases
weave health check --local         # Check only local databases

# Sort health check results (NEW in v0.7.2)
weave health check --sort-by name  # Alphabetical (default)
weave health check --sort-by type  # By database type

# Check specific database
weave health check weaviate
```

---

### 4. Collections Management

#### List Collections

```bash
# List collections from default database
weave cols ls

# List from specific databases
weave cols ls --weaviate           # Weaviate only
weave cols ls --weaviate-cloud     # Weaviate cloud only
weave cols ls --weaviate-local     # Weaviate local only
weave cols ls --supabase           # Supabase only
weave cols ls --mongodb            # MongoDB Atlas only
weave cols ls --milvus-local       # Milvus local only
weave cols ls --milvus-cloud       # Milvus cloud (Zilliz) only
weave cols ls --chroma-local       # Chroma local only
weave cols ls --chroma-cloud       # Chroma cloud only
weave cols ls --qdrant-local       # Qdrant local only
weave cols ls --neo4j-local        # Neo4j local only
weave cols ls --opensearch-local   # OpenSearch local only (NEW in v0.7.3)
weave cols ls --opensearch-cloud   # OpenSearch cloud (AWS) only (NEW in v0.7.3)
weave cols ls --all                # All configured databases

# Summary view (NEW in v0.7.2)
weave cols ls -S                   # Summary table (default for multiple VDBs)
weave cols ls --summary            # Explicit summary flag

# Filter by deployment (NEW in v0.7.2)
weave cols ls --cloud              # Collections from cloud databases only
weave cols ls --local              # Collections from local databases only
```

#### Create Collections

```bash
# Create with default embedding
weave cols create DemoDocs --text --json-metadata --weaviate-cloud

# Create with specific embedding
weave cols create DemoDocs --embedding text-embedding-3-small
weave cols create DemoDocs -e text-embedding-ada-002

# Create image collection
weave cols create DemoImages --image --json-metadata --weaviate-cloud
```

#### Show Collection Details

```bash
# Show collection schema
weave cols show DemoDocs --schema --expand-metadata --weaviate-cloud

# Show with JSON output
weave cols show DemoImages --schema --expand-metadata --json --weaviate-cloud

# Count collections
weave cols count
```

---

### 5. Documents Management

#### Text Documents

```bash
# List documents
weave docs ls DemoDocs

# List with summary (NEW)
weave docs ls DemoDocs -w -S

# Create documents
weave docs create DemoDocs ./README.md
weave docs create DemoDocs ./docs/PRESENTATION.md

# Create with specific embedding
weave docs create DemoDocs document.txt --embedding text-embedding-3-small
weave docs create DemoDocs report.pdf --embedding text-embedding-ada-002

# Show document details
weave docs show DemoDocs <ID> --schema --expand-metadata

# Delete documents
weave docs del DemoDocs --name "README.md"
```

#### Image Documents

```bash
# List image documents
weave docs ls DemoImages
weave docs ls DemoImages -w -S

# Create image documents
weave docs create DemoImages ./tests/images/dog.png

# Extract images from PDF
weave docs create DemoDocs ./tests/fixtures/ragme-io.pdf --image-col DemoImages

# Show image document
weave docs show DemoImages <ID> --schema --expand-metadata

# Delete image document
weave docs del DemoImages --name "./tests/images/dog.png"
```

#### Batch Processing

```bash
# Batch process documents with parallel workers
weave docs batch --directory ./docs --collection MyCollection --parallel 3

# Convert CMYK PDFs to RGB
weave docs pdf-convert document.pdf --rgb

# Text-only PDF extraction (faster, no images)
weave docs create MyCollection document.pdf --skip-all-images

# Count documents
weave docs count DemoDocs

# Delete all documents
weave docs delete-all DemoDocs
```

---

### 6. Query Collections

```bash
# Semantic search
weave cols query DemoDocs "golang"
weave cols q DemoDocs "vector database"

# With top_k limit
weave cols query DemoDocs "RAGme.io" --top_k 3

# Natural language queries with AI agents
weave q "find all empty collections"
weave query "create TestDocs and add README.md" --dry-run
```

---

### 7. Embeddings

```bash
# List available embeddings
weave embeddings list
weave emb ls

# List with verbose output
weave emb ls --verbose
```

---

### 8. Advanced Features

#### Multi-Database Operations

```bash
# Query multiple databases at once
weave cols query MyCollection "search" --weaviate --supabase

# Read operations work with specific or all databases
weave cols ls --weaviate                     # All Weaviate databases
weave cols ls --supabase                     # Supabase only
weave cols ls --all                          # All configured databases (default)
```

#### Timeout Configuration

```bash
# Configure timeout for slow connections
weave cols ls --timeout 30s
weave health check --timeout 60s
```

#### Delete Operations

```bash
# Delete collection schema
weave cols delete-schema DemoDocs

# Delete entire collection
weave cols delete MyCollection --force
```

---

### 9. AI Agent Mode (REPL)

```bash
# Start AI-powered REPL mode
weave

# Example interactions:
> show me all my collections
> create TestDocs collection
> add README.md to TestDocs
> list my empty collections
> find documents about "vector search" in WeaveDocs
```

#### Single Query Mode

```bash
# Execute one AI query at a time
weave query "show me all my collections"
weave q "create TestDocs and add README.md"
```

---

## Demo Commands Summary

### Quick Start Commands

```bash
# 1. Help
weave -h

# 2. Configuration
weave config show
weave config list --details

# 3. Health Check
weave health check -S

# 4. Collections
weave cols ls --all
weave cols create DemoDocs --text
weave cols show DemoDocs

# 5. Documents
weave docs create DemoDocs README.md
weave docs ls DemoDocs
weave docs count DemoDocs

# 6. Query
weave cols q DemoDocs "vector database"

# 7. Embeddings
weave emb ls

# 8. AI Mode
weave
> list my collections
```

---

## New Features in v0.7.3

### OpenSearch Support (NEW)

```bash
# List OpenSearch collections
weave cols ls --opensearch-local      # Local OpenSearch
weave cols ls --opensearch-cloud      # AWS OpenSearch Service

# Create collection in OpenSearch
weave cols create MyDocs --opensearch-local

# Health check OpenSearch
weave health check --opensearch-local
```

### Sorting Features (v0.7.2)

```bash
# Sort configuration list
weave config list --sort-by name         # Alphabetical (default)
weave config list --sort-by type         # By database type
weave config list --sort-by deployment   # By deployment type

# Sort health check results
weave health check --sort-by name        # Alphabetical (default)
weave health check --sort-by type        # By database type
```

### Filtering Features (v0.7.2)

```bash
# Filter by deployment type
weave config list --cloud                # Show only cloud databases
weave config list --local                # Show only local databases
weave health check --cloud               # Check only cloud databases
weave health check --local               # Check only local databases
weave cols ls --cloud                    # Collections from cloud databases
weave cols ls --local                    # Collections from local databases
```

### Summary Tables with Progressive Output (v0.7.2)

```bash
# Collections summary
weave cols ls -S                         # Summary table (default for multiple VDBs)

# Health check summary
weave health check -S                    # Progressive output (results appear immediately)

# Features:
# • Progressive Output: Results appear as they're retrieved/checked
# • Status Indicators: ✓ OK (green) or ✗ FAIL (red)
# • Footer Statistics: Total count, healthy count, failures
# • Auto-Selection: Summary for multiple VDBs, detailed for single VDB
```

---

## Database Support

| Database | Type | Status | CLI Flags |
|----------|------|--------|-----------|
| **Weaviate** | Cloud/Local | ✅ Stable | `--weaviate`, `--weaviate-cloud`, `--weaviate-local` |
| **Milvus** | Cloud/Local | 🟢 Beta | `--milvus`, `--milvus-cloud`, `--milvus-local` |
| **Supabase** | Cloud/Local | 🟡 Alpha | `--supabase`, `--supabase-cloud`, `--supabase-local` |
| **MongoDB Atlas** | Cloud | 🧪 Experimental | `--mongodb`, `--mongodb-cloud` |
| **Chroma** | Cloud/Local | ✅ Stable | `--chroma`, `--chroma-cloud`, `--chroma-local` |
| **Qdrant** | Cloud/Local | 🧪 Experimental | `--qdrant`, `--qdrant-cloud`, `--qdrant-local` |
| **Neo4j** | Cloud/Local | 🧪 Experimental | `--neo4j`, `--neo4j-cloud`, `--neo4j-local` |
| **OpenSearch** | Cloud/Local | 🧪 Experimental | `--opensearch`, `--opensearch-cloud`, `--opensearch-local` |

---

## Video Demos

- **[Full Demo (5 min)](https://asciinema.org/a/LrKzmThBfDbTPISZzr8biP4dt)** -
  Complete feature walkthrough
- **[Quick Demo (2 min)](https://asciinema.org/a/HiAU7h1iJvZ2QdJe70ae3Cc0b)** -
  Quick overview
- **[REPL Demo](https://asciinema.org/a/U504HN4FSeMsOA0qS0os0NWUE)** -
  AI-powered natural language interface

---

## Interactive Demo Scripts

Run these scripts locally for hands-on demonstrations:

```bash
# Full demo (5 minutes)
./demos/full-demo.sh

# Quick demo (2 minutes)
./demos/quick-demo.sh

# Configuration demo
./demos/config-demo.sh

# REPL/AI interface demo
./demos/repl-demo.sh

# Supabase-specific demo
./demos/supabase-demo.sh
```

---

## Resources

- **[GitHub Repository](https://github.com/maximilien/weave-cli)**
- **[User Guide](USER_GUIDE.md)** - Complete feature documentation
- **[Changelog](CHANGELOG.md)** - Version history and updates
- **[VDB Support Matrix](VDB_SUPPORT.md)** - Database feature comparison
- **[Vector DB Abstraction](VECTOR_DB_ABSTRACTION.md)** - Architecture details

---

## Tips for Thursday Demo

### Recommended Flow (15-20 minutes)

1. **Introduction (2 min)**
   - Show `weave -h`
   - Quick overview of features

2. **Configuration & Health (3 min)**
   - `weave config list --details --sort-by type`
   - `weave health check -S --cloud`
   - Highlight 8 database support

3. **Collections Demo (5 min)**
   - `weave cols ls --all -S`
   - `weave cols create DemoDocs --embedding text-embedding-3-small`
   - `weave cols show DemoDocs --schema`

4. **Documents Demo (5 min)**
   - `weave docs create DemoDocs README.md`
   - `weave docs ls DemoDocs -w -S`
   - `weave cols q DemoDocs "vector database" --top_k 3`

5. **Advanced Features (3 min)**
   - Show multi-database: `weave cols ls --weaviate --supabase`
   - Show batch processing: `weave docs batch --help`
   - Show AI mode: `weave` (enter REPL)

6. **Q&A (2 min)**

### Key Talking Points

- **Performance**: Single Go binary, fast startup, real-time feedback
- **Flexibility**: 8 vector databases supported, easy to switch
- **AI-Powered**: Natural language REPL with GPT-4o
- **Production-Ready**: Comprehensive testing (7/8 suites passing)
- **Developer-Friendly**: Simple CLI, extensive documentation

### Commands to Highlight

```bash
# Show off new sorting/filtering
weave config list --cloud --sort-by name
weave health check -S --local

# Show off multi-database support
weave cols ls --all -S

# Show off semantic search
weave cols q DemoDocs "How do I use embeddings?"

# Show off AI mode
weave
> list my empty collections
```

---

## Notes

- All demo commands are safe to run multiple times
- Demo collections can be cleaned up with `weave cols delete <name> --force`
- For Thursday demo, focus on stability (Weaviate, Chroma, Milvus)
- Highlight new v0.7.2/v0.7.3 features (sorting, filtering, OpenSearch)
- Prepare backup slides showing test results (7/8 passing)
