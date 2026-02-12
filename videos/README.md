# Weave CLI Demo Videos

This directory contains asciinema recordings of Weave CLI demonstrations.

## Available Recordings

### Core Demos
- `weave-cli-full-demo.cast` - Complete 5-minute demo showcasing all features
- `weave-cli-quick-demo.cast` - Quick 2-minute demo for rapid overview
- `weave-cli-repl-demo.cast` - AI-powered REPL mode demonstration
- `weave-cli-config-demo.cast` - Configuration management demo
- `weave-cli-supabase-demo.cast` - Supabase integration demo

### OSS Embedding Demos (v0.9.19+) 🆕
- `scripts/oss-embeddings-basic.sh` - Basic OSS embedding workflow (~2 min)
- `scripts/oss-embeddings-reembed.sh` - Re-embedding demo (20x faster) (~3 min)
- `scripts/oss-embeddings-compare.sh` - Model comparison and selection (~3 min)

### Metadata
- `latest-demo-uploads.txt` - **Latest uploaded demo URLs** (automatically updated)

## Creating Recordings

### Recording Core Demos

Use the asciinema recording tool:

```bash
# Record full demo (5 minutes)
./tools/asciinema.sh demo

# Record quick demo (2 minutes)
./tools/asciinema.sh quick

# List available recordings
./tools/asciinema.sh list

# Upload to asciinema.org (saves URL to latest-demo-uploads.txt)
./tools/asciinema.sh upload

# View latest uploaded demo URLs
cat videos/latest-demo-uploads.txt
```

### Recording OSS Embedding Demos (v0.9.19+)

Record OSS embedding workflow demonstrations:

```bash
# Prerequisites
pip install sentence-transformers  # One-time setup
weave health check --milvus-local   # Verify Milvus is running

# Record Basic OSS Workflow (~2 min)
asciinema rec videos/oss-embeddings-basic.cast -c "bash videos/scripts/oss-embeddings-basic.sh"

# Record Re-embedding Demo (~3 min)
asciinema rec videos/oss-embeddings-reembed.cast -c "bash videos/scripts/oss-embeddings-reembed.sh"

# Record Model Comparison Demo (~3 min)
asciinema rec videos/oss-embeddings-compare.cast -c "bash videos/scripts/oss-embeddings-compare.sh"

# Upload all OSS demos
asciinema upload videos/oss-embeddings-basic.cast
asciinema upload videos/oss-embeddings-reembed.cast
asciinema upload videos/oss-embeddings-compare.cast
```

**What Each Demo Shows:**

1. **oss-embeddings-basic.sh** - Getting started with OSS embeddings
   - Install sentence-transformers
   - List available models
   - Create collection with OSS embedding
   - Query with auto-detection
   - Verify collection metadata

2. **oss-embeddings-reembed.sh** - Fast model switching (20x speedup)
   - Create OpenAI collection
   - Re-embed to OSS model (200+ docs/min)
   - Query both collections
   - Compare quality (90%+ retention typical)
   - Cost analysis ($240/year savings)

3. **oss-embeddings-compare.sh** - Model selection guide
   - Re-embed to multiple models (mpnet, MiniLM)
   - Query all collections with same query
   - Generate quality comparison reports
   - Recommendations for model selection

## Prerequisites

1. **Install asciinema**:

   ```bash
   ./tools/asciinema.sh install
   ```

2. **Configure Weaviate**: Ensure your Weaviate Cloud instance is configured

3. **Demo Data**: Ensure you have:
   - `docs/README.md` and other markdown files
   - `images/` directory with sample images (optional)

## Playing Recordings

```bash
# Play a recording locally
asciinema play videos/weave-cli-full-demo.cast

# Upload and get shareable URL
asciinema upload videos/weave-cli-full-demo.cast
```

## Demo Script

The demo follows the script in `docs/DEMO.md` with 10 pages:

1. **Health Check & Configuration** - Verify setup
2. **Create Collections** - Text and image collections
3. **List Collections** - Show available collections
4. **Create Documents** - Add sample documents
5. **Show Documents & Schema** - Detailed document view
6. **List Documents** - Simple and virtual document listing
7. **Delete Documents** - Document deletion with confirmation
8. **Cleanup Operations** - Delete all and schema cleanup
9. **Getting Weave CLI** - Download and build instructions
10. **Thank You** - Credits and references

## Recording Tips

- **Timing**: Each command has appropriate delays for readability
- **Error Handling**: Commands include fallbacks for missing files
- **Cleanup**: Demo collections are cleaned up automatically
- **Duration**: Full demo ~5 minutes, quick demo ~2 minutes

## Sharing

Once uploaded to asciinema.org, the URL is automatically saved to
`latest-demo-uploads.txt` and displayed in the terminal.

**Latest Demo URLs**: Check `videos/latest-demo-uploads.txt` for the most
recent upload links.

Shareable URLs look like: `https://asciinema.org/a/abc123`

Perfect for:

- Documentation
- Presentations
- Social media
- GitHub README
- Project showcases

## Latest Uploads

The `latest-demo-uploads.txt` file maintains:

- Latest URL for quick demo
- Latest URL for full demo
- Upload timestamps and filenames
- Automatically updated on each upload
