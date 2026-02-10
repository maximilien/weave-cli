# OSS Embedding Testing Tips - Client0 Guide

**Version:** weave-cli v0.9.19
**Date:** 2026-02-10
**Feature:** OSS Embedding Providers (sentence-transformers & Ollama)

---

## Quick Testing Checklist

### Prerequisites ✅
- [ ] weave-cli v0.9.19 installed
- [ ] Python 3.8+ with pip (for sentence-transformers)
- [ ] Ollama installed (optional, for Ollama provider)
- [ ] Existing collection with documents (e.g., AuctionListings)

### 5-Minute Proof of Concept
```bash
# 1. Install sentence-transformers (one time)
pip install sentence-transformers

# 2. Re-embed existing collection with OSS model
weave collection reembed AuctionListings \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output AuctionListings_OSS \
  --milvus-local

# 3. Generate comparison report
weave collection compare \
  AuctionListings \
  AuctionListings_OSS \
  --query "vintage cameras" \
  --query "auction results" \
  --top-k 10 \
  --report comparison-report.md \
  --milvus-local

# 4. Review results
cat comparison-report.md
```

---

## Testing Tips by Provider

### sentence-transformers Provider

#### Installation
```bash
# Install Python package
pip install sentence-transformers

# Verify installation
python3 -c "from sentence_transformers import SentenceTransformer; print('✅ Ready')"
```

#### Recommended Models
1. **all-mpnet-base-v2** (768 dims) - Best quality, recommended
2. **all-MiniLM-L6-v2** (384 dims) - Fastest, lightweight
3. **all-MiniLM-L12-v2** (384 dims) - Balance of speed and quality

#### Common Issues

**Issue 1: ModuleNotFoundError**
```bash
# Error: No module named 'sentence_transformers'
# Fix:
pip install sentence-transformers
```

**Issue 2: Model Download Takes Time**
```bash
# First run downloads model (~400MB for all-mpnet-base-v2)
# Expected: 2-5 minutes on first use
# Subsequent runs are instant (model cached)
```

**Issue 3: Python Version**
```bash
# Requires Python 3.8+
python3 --version

# If too old, install newer Python:
# macOS: brew install python@3.11
# Ubuntu: sudo apt install python3.11
```

#### Performance Expectations
- **First embedding**: 2-5 min (model download)
- **Subsequent embeddings**: ~30-60 sec for 1000 docs
- **Memory**: ~2GB RAM for all-mpnet-base-v2
- **Dimensions**: 768 (better than OpenAI's 3072 for Milvus performance!)

---

### Ollama Provider

#### Installation
```bash
# Install Ollama (one time)
# macOS/Linux:
curl -fsSL https://ollama.ai/install.sh | sh

# Or download from: https://ollama.ai

# Start Ollama server
ollama serve

# Pull embedding model
ollama pull nomic-embed-text
```

#### Recommended Models
1. **nomic-embed-text** (768 dims) - Best quality for embeddings
2. **mxbai-embed-large** (1024 dims) - Higher dimensions, slower
3. **snowflake-arctic-embed** (1024 dims) - Alternative option

#### Common Issues

**Issue 1: Ollama Not Running**
```bash
# Error: Ollama server not running at http://localhost:11434
# Fix:
ollama serve

# Or start in background:
nohup ollama serve > /dev/null 2>&1 &
```

**Issue 2: Model Not Found**
```bash
# Error: model 'nomic-embed-text' not found in Ollama
# Fix:
ollama pull nomic-embed-text

# List installed models:
ollama list
```

**Issue 3: Port Conflict**
```bash
# If port 11434 already in use
# Check what's using it:
lsof -i :11434

# Kill process or change Ollama port:
OLLAMA_HOST=0.0.0.0:11435 ollama serve
```

#### Performance Expectations
- **First embedding**: 30-60 sec for 1000 docs
- **Memory**: ~1.5GB RAM for nomic-embed-text
- **Dimensions**: 768 (same as sentence-transformers)
- **Speed**: Similar to sentence-transformers

---

## Testing Workflow

### Step 1: Single Collection Test (10 minutes)

Test one collection with sentence-transformers:

```bash
# Re-embed AuctionListings
weave collection reembed AuctionListings \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output AuctionListings_OSS \
  --milvus-local \
  --verbose

# Expected output:
# ✅ Created provider: sentence-transformers
# ✅ Re-embedding 3,518 documents...
# ✅ Batch 1/36: 100 documents processed
# ...
# ✅ Collection AuctionListings_OSS created with 3,518 documents
```

### Step 2: Comparison Test (5 minutes)

Compare OpenAI vs OSS embeddings:

```bash
# Generate comparison report
weave collection compare \
  AuctionListings \
  AuctionListings_OSS \
  --query "vintage Leica cameras from 1950s" \
  --query "auction results for photography equipment" \
  --query "rare collectible cameras" \
  --query "Nikon F series cameras" \
  --query "medium format film cameras" \
  --top-k 10 \
  --report comparison-report.md \
  --milvus-local

# Review report
cat comparison-report.md
```

### Step 3: 3-Way Comparison (15 minutes)

Test all three providers:

```bash
# Re-embed with Ollama
ollama pull nomic-embed-text

weave collection reembed AuctionListings \
  --new-embedding nomic-embed-text \
  --output AuctionListings_Ollama \
  --milvus-local

# Compare all 3 stacks
weave collection compare \
  AuctionListings \
  AuctionListings_OSS \
  AuctionListings_Ollama \
  --query "vintage cameras" \
  --query "auction results" \
  --top-k 10 \
  --report 3-way-comparison.md \
  --milvus-local
```

---

## What to Look For

### Quality Metrics

1. **Relevance Scores**
   - OpenAI baseline: ~0.75-0.85 avg score
   - OSS target: >0.70 avg score (>85% retention)
   - If OSS scores < 0.65, try different model

2. **Document Overlap**
   - Compare top-10 results between providers
   - Good: 7-8 overlapping documents
   - Acceptable: 5-6 overlapping documents
   - Concerning: <5 overlapping documents

3. **Query Performance**
   - OpenAI: ~100-200ms per query
   - OSS: ~80-150ms per query (often faster!)
   - Ollama: ~100-200ms per query

### Cost Savings

- **OpenAI**: $0.02 per 1M tokens (~$0.10 for 3,518 docs)
- **sentence-transformers**: $0.00 (100% free)
- **Ollama**: $0.00 (100% free)

**Yearly savings** (assuming re-embedding quarterly):
- 4 re-embeddings × $0.10 = $0.40/year per collection
- 3 collections × $0.40 = **$1.20/year saved**
- Plus ongoing query cost savings!

---

## Troubleshooting

### Problem: Re-embedding Fails Immediately

**Symptom:**
```
Error: unknown embedding model: sentence-transformers/all-mpnet-base-v2
```

**Solution:**
Check weave-cli version:
```bash
weave -V
# Should show: Weave CLI 0.9.19

# If older version:
cd /path/to/weave-cli
git pull origin main
./build.sh
```

### Problem: Python Subprocess Error

**Symptom:**
```
Error: failed to generate embeddings: exit status 1
ModuleNotFoundError: No module named 'sentence_transformers'
```

**Solution:**
```bash
# Install sentence-transformers
pip install sentence-transformers

# Verify Python can import it
python3 -c "from sentence_transformers import SentenceTransformer; print('OK')"
```

### Problem: Ollama Connection Refused

**Symptom:**
```
Error: Ollama server not running at http://localhost:11434
```

**Solution:**
```bash
# Start Ollama server
ollama serve

# Verify server is running
curl http://localhost:11434/api/tags

# Should return JSON with available models
```

### Problem: Collection Already Exists

**Symptom:**
```
Error: collection 'AuctionListings_OSS' already exists
```

**Solution:**
```bash
# Delete old collection
weave collection delete AuctionListings_OSS --milvus-local

# Or use different output name
weave collection reembed AuctionListings \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output AuctionListings_OSS_v2 \
  --milvus-local
```

---

## Performance Benchmarks

### Small Dataset (100 documents)
- OpenAI: ~5 seconds
- sentence-transformers: ~3 seconds
- Ollama: ~4 seconds

### Medium Dataset (1,000 documents)
- OpenAI: ~45 seconds
- sentence-transformers: ~30 seconds
- Ollama: ~35 seconds

### Large Dataset (10,000 documents)
- OpenAI: ~7 minutes
- sentence-transformers: ~5 minutes
- Ollama: ~6 minutes

**Note:** First-time sentence-transformers run adds 2-5 min for model download.

---

## Success Criteria

### Minimum Acceptable Performance (Go/No-Go)
- ✅ Re-embedding completes successfully
- ✅ Search returns relevant results
- ✅ Avg relevance score >0.70 (>85% of OpenAI)
- ✅ Latency <2x OpenAI baseline
- ✅ No timeout errors from Milvus

### Target Performance (Ideal)
- ✅ Avg relevance score >0.75 (>90% of OpenAI)
- ✅ Latency similar or better than OpenAI
- ✅ 7-8 overlapping documents in top-10 results
- ✅ Consistent performance across query types

### Red Flags (Stop and Investigate)
- ❌ Avg relevance score <0.65
- ❌ Latency >5x OpenAI baseline
- ❌ Frequent timeout errors
- ❌ <5 overlapping documents in top-10 results
- ❌ Different results for same query (non-deterministic)

---

## Next Steps After Testing

### If Results Are Good (>85% retention)
1. Re-embed remaining collections (AuctionImages, AuctionResults)
2. Update production config to use OSS embeddings
3. Document decision and rationale
4. Plan migration timeline

### If Results Are Mixed (70-85% retention)
1. Test with different models:
   - Try `all-MiniLM-L12-v2` for speed
   - Try `mxbai-embed-large` for quality
2. Analyze which query types perform worse
3. Consider hybrid approach (OpenAI for critical queries, OSS for bulk)

### If Results Are Poor (<70% retention)
1. Verify test data quality
2. Check if queries are domain-specific (may need fine-tuned model)
3. Compare with fresh OpenAI re-embedding (not old collection)
4. Consider staying with OpenAI for now

---

## Command Reference

### Re-embedding Commands
```bash
# sentence-transformers (recommended)
weave collection reembed <source> \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output <destination> \
  --milvus-local

# Ollama
weave collection reembed <source> \
  --new-embedding nomic-embed-text \
  --output <destination> \
  --milvus-local

# OpenAI (for comparison)
weave collection reembed <source> \
  --new-embedding text-embedding-3-small \
  --output <destination> \
  --milvus-local
```

### Comparison Commands
```bash
# 2-way comparison
weave collection compare Collection1 Collection2 \
  --query "test query" \
  --top-k 10 \
  --report report.md \
  --milvus-local

# 3-way comparison
weave collection compare Collection1 Collection2 Collection3 \
  --query "test query" \
  --top-k 10 \
  --report report.md \
  --milvus-local

# JSON output
weave collection compare Collection1 Collection2 \
  --query "test query" \
  --top-k 10 \
  --report report.json \
  --format json \
  --milvus-local
```

### Model Discovery Commands
```bash
# List available embedding models
weave embeddings list

# Discover Ollama models
weave config agents
# (auto-detects local Ollama models)

# Check Ollama models manually
ollama list
```

---

## Support

### Quick Links
- **weave-cli Issues**: https://github.com/maximilien/weave-cli/issues
- **Client0 Status Doc**: `/Users/maximilien/github/auctionsmax-ai/docs/ISSUES-11-15-STATUS.md`
- **CHANGELOG**: https://github.com/maximilien/weave-cli/blob/main/CHANGELOG.md

### Getting Help
1. Check CHANGELOG for known issues
2. Review auctionsmax-ai docs for your specific workflow
3. Create GitHub issue with:
   - weave-cli version (`weave -V`)
   - Error message
   - Command that failed
   - Expected vs actual behavior

---

**Happy Testing! 🚀**
