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
./bin/weave docs create WeaveDocs docs/README.md docs/CHANGELOG.md
```
**Expected Output:**
```
✅ Successfully created document: README.md (3 chunks)
✅ Successfully created document: CHANGELOG.md (5 chunks)
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
./bin/weave docs show WeaveDocs README.md
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
**Expected Output:**
```
📄 Collection: WeaveDocs
  Documents: 2
  Schema: ragmedocs
  Fields: content, metadata, filename
```

---

## Page 6: List Documents

### Simple Document List
```bash
./bin/weave docs ls WeaveDocs
```
**Expected Output:**
```
📋 Documents in WeaveDocs:
  README.md (3 chunks)
  CHANGELOG.md (5 chunks)
```

### Virtual Document List with Summary
```bash
./bin/weave docs ls WeaveDocs -w -S
```
**Expected Output:**
```
📋 Virtual Documents:
📄 README.md - 3 chunks
📄 CHANGELOG.md - 5 chunks

📋 Summary:
1. README.md - 3 chunks
2. CHANGELOG.md - 5 chunks
```

---

## Page 7: Delete Documents

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

## Page 8: Cleanup Operations

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

## Page 9: Getting Weave CLI

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

## Page 10: Thank You

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