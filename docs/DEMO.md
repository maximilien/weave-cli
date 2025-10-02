# Weave CLI Demo Script

A 5-minute demonstration of Weave CLI capabilities for managing Weaviate vector databases.

## Page 1: Health Check & Configuration

### Health Check
```bash
./bin/weave health check
```
**Expected Output:**
```
✅ Weaviate connection successful
✅ Database is healthy and accessible
```

### Configuration Display
```bash
./bin/weave config show
```
**Expected Output:**
```
🔧 Configuration:
  Vector DB Type: weaviate-cloud
  Weaviate URL: https://your-instance.weaviate.cloud
  API Key: [configured]
```

### Help Command
```bash
./bin/weave --help
```
**Expected Output:**
```
Weave is a command-line tool for managing Weaviate vector databases.

Available Commands:
  collection  Collection management
  document    Document management
  config      Configuration management
  health      Health and connectivity management
```

---

## Page 2: Create Collections

### Create Text Collection
```bash
./bin/weave cols create WeaveDocs --schema-type ragmedocs --embedding-model text-embedding-3-small
```
**Expected Output:**
```
✅ Successfully created collection: WeaveDocs
📄 Schema Type: ragmedocs (text documents)
```
*Note: If collection already exists, command will show "Collection already exists"*

### Create Image Collection
```bash
./bin/weave cols create WeaveImages --schema-type ragmeimages --embedding-model text-embedding-3-small
```
**Expected Output:**
```
✅ Successfully created collection: WeaveImages
🖼️ Schema Type: ragmeimages (image documents)
```
*Note: If collection already exists, command will show "Collection already exists"*

### Show Collection Structure
```bash
./bin/weave cols show WeaveDocs
```
**Expected Output:**
```
📄 Collection: WeaveDocs
  Type: ragmedocs
  Documents: 0
  Schema: Configured for text processing
```

---

## Page 3: List Collections

### List All Collections
```bash
./bin/weave cols ls
```
**Expected Output:**
```
📋 Collections:
📄 WeaveDocs (0 docs) - ragmedocs
🖼️ WeaveImages (0 docs) - ragmeimages
```

---

## Page 4: Create Documents

### Create Text Documents
```bash
./bin/weave docs create WeaveDocs README.md docs/DEMO.md
```
**Expected Output:**
```
✅ Successfully created document: README.md (24 chunks)
✅ Successfully created document: DEMO.md (7 chunks)
```

### Create Image Documents
```bash
./bin/weave docs create WeaveImages images/screenshot1.png images/screenshot2.jpg
```
**Expected Output:**
```
✅ Successfully created document: screenshot1.png (1 chunk)
✅ Successfully created document: screenshot2.jpg (1 chunk)
```

---

## Page 5: Show Documents & Schema

### Show Document Details
```bash
./bin/weave docs show WeaveDocs --name README.md
```
**Expected Output:**
```
📄 Document: README.md
  ID: abc123-def456-ghi789
  Chunks: 3
  Content: [truncated preview]
  Metadata: {"original_filename": "README.md", "is_chunked": true}
```

### Show Collection Schema
```bash
./bin/weave cols show WeaveDocs
```

## Page 6: List Documents

### Simple Document List
```bash
./bin/weave docs ls WeaveDocs
```

### Virtual Document View with Summary
```bash
./bin/weave docs ls WeaveDocs -w -S
```
**Expected Output:**
```
✅ Found 1 virtual documents in collection 'WeaveDocs' (aggregated from 3 total documents):

1. 📄 Document: README.md
   📝 Chunks: 3
   📋 Metadata
     original_filename: README.md
     type: text
     is_chunked: true

📋 Summary
   1. README.md - 3 chunks
```

### Show Collection Schema
```bash
./bin/weave cols show WeaveDocs
```
**Expected Output:**
```
📄 Collection: WeaveDocs
  Documents: 2
  Schema: ragmedocs
  Fields: content, metadata, filename
```

---

## Page 7: Semantic Search & Query

### Basic Semantic Search
```bash
./bin/weave cols q WeaveDocs "weave-cli installation"
```
**Expected Output:**
```
✅ Semantic search results for 'weave-cli installation' in collection 'WeaveDocs':

1. 🔍 Score: 1.000
   ID: c937af68-727e-4946-8df5-f26919df7645
   Content: # Weave CLI v0.2.6
   
   A command-line tool for managing Weaviate vector databases...
   📋 Metadata: {"filename": "README.md", "type": "text"}

📊 Summary: Found 1 results
```

### Search with Custom Result Limit
```bash
./bin/weave cols q WeaveDocs "machine learning" --top_k 3
```
**Expected Output:**
```
✅ Semantic search results for 'machine learning' in collection 'WeaveDocs':

1. 🔍 Score: 1.000
   ID: doc1-chunk1
   Content: This document covers machine learning algorithms...
   📋 Metadata: {"filename": "ml_guide.txt", "type": "text"}

📊 Summary: Found 1 results
```

### Search with Metadata (NEW!)
```bash
./bin/weave cols q WeaveDocs "maximilien.org" --search-metadata
```
**Expected Output:**
```
✅ Semantic search results for 'maximilien.org' in collection 'WeaveDocs':

1. 🔍 Score: 1.000
   ID: e0b3768f-2cc9-4962-aee2-913a95e5757c
   Content: [Navigation menu content]
   📋 Metadata: {"url": "https://maximilien.org", "type": "webpage"}

📊 Summary: Found 1 results
```

### BM25 Keyword Search (NEW!)
```bash
./bin/weave cols q WeaveDocs "exact keywords" --bm25
```
**Expected Output:**
```
✅ Semantic search results for 'exact keywords' in collection 'WeaveDocs':

1. 🔍 Score: 0.850
   ID: doc1-chunk1
   Content: This document contains exact keywords for BM25 search...
   📋 Metadata: {"filename": "keywords.txt", "type": "text"}

📊 Summary: Found 1 results
```

### Query Help
```bash
./bin/weave cols q --help
```
**Expected Output:**
```
Perform semantic search on a collection using natural language queries.

Usage:
  weave collection query COLLECTION "query text" [flags]

Flags:
  -m, --search-metadata   Also search in metadata fields (default: false)
  -k, --top_k int         Number of top results to return (default: 5)
  -d, --distance float    Maximum distance threshold for results
```

---

## Page 8: Delete Documents

### Delete Single Document
```bash
./bin/weave docs delete WeaveDocs README.md
```
**Expected Output:**
```
⚠️  Are you sure you want to delete document 'README.md'? [y/N]: y
✅ Successfully deleted document: README.md
```

### Delete with Force Flag
```bash
./bin/weave docs delete WeaveImages screenshot1.png --force
```
**Expected Output:**
```
✅ Successfully deleted document: screenshot1.png
```

---

## Page 9: Cleanup Operations

### Delete All Documents
```bash
./bin/weave docs delete-all WeaveDocs --force
```
**Expected Output:**
```
✅ Successfully deleted all documents from collection: WeaveDocs
```

### Delete Collection Schema
```bash
./bin/weave cols delete-schema WeaveDocs --force
```
**Expected Output:**
```
✅ Successfully deleted schema for collection: WeaveDocs
```

---

## Page 10: Getting Weave CLI

### Download Binary
```bash
# Download latest release from GitHub
curl -L https://github.com/maximilien/weave-cli/releases/latest/download/weave-darwin-amd64 -o weave
chmod +x weave
```

### Build from Source
```bash
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh
```

### Open Source
Built with ❤️ by [github.com/maximilien](https://github.com/maximilien)

- **License**: MIT License - Free for commercial and personal use
- **Repository**: https://github.com/maximilien/weave-cli
- **Documentation**: https://github.com/maximilien/weave-cli/blob/main/README.md
- **Issues**: https://github.com/maximilien/weave-cli/issues

---

## Page 11: Thank You

### Demo Complete
```bash
echo "🎉 Demo completed successfully!"
./bin/weave --version
```
**Expected Output:**
```
🎉 Demo completed successfully!
Weave CLI 0.2.1
  Git Commit: 52b56ba
  Build Time: 2025-09-29 23:38:33
  Go Version: go1.24.1
```

### Credits
- **Weave CLI**: Vector database management made simple
- **Weaviate**: Powerful vector database platform
- **MIT License**: Open source, free for commercial use
- **Community**: Built with ❤️ by the open source community

**Thank you for watching!** 🚀

---

## Demo Notes

- **Duration**: ~5 minutes
- **Prerequisites**: Weaviate Cloud instance configured
- **Test Collections**: Uses WeaveDocs and WeaveImages for isolation
- **Cleanup**: All demo collections are cleaned up automatically
- **Recording**: Use `./tools/asciinema.sh` to record this demo