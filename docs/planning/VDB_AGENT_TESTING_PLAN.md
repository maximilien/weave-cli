# RAG Agent Testing Plan for All Vector Databases

**Status:** Ready for Testing
**Created:** 2026-01-12
**Purpose:** Verify RAG agents work correctly with all 10+ vector databases

---

## Overview

Now that Multi-VDB agent support is implemented (v0.9.1), we need to verify that RAG agents work correctly with all supported vector databases. This document provides a systematic testing plan.

## Implementation Status

✅ **Completed:**
- Generic `QueryCollectionWithAgent()` function
- Mock database agent support
- Weaviate agent support (existing)
- Progress indicator with JSON Lines format
- All code passes tests, builds, and linting

⏭️ **Ready for Testing:**
- Chroma (local + cloud)
- Qdrant (local + cloud)
- Milvus (local + cloud)
- Neo4j (local + Aura)
- Supabase (local + cloud)
- MongoDB (local + Atlas)
- Pinecone (cloud)

---

## Testing Strategy

### Phase 1: Local VDB Testing (Priority: High)

Test with locally running vector databases first to ensure basic functionality.

#### 1.1 Chroma Local

**Setup:**
```bash
# Start Chroma locally
docker run -p 8000:8000 chromadb/chroma:latest
```

**Test Commands:**
```bash
# Create test collection
weave cols create TestDocs --db chroma-local

# Add test documents
weave docs add TestDocs "Machine learning is a subset of AI." \
  --metadata "source=test1" --db chroma-local

weave docs add TestDocs "Neural networks are used in deep learning." \
  --metadata "source=test2" --db chroma-local

# Test basic agent query
weave cols query TestDocs "What is machine learning?" \
  --agent rag-agent --db chroma-local

# Test with progress
weave cols query TestDocs "What is machine learning?" \
  --agent rag-agent --progress --db chroma-local

# Test with JSON output
weave cols query TestDocs "What is machine learning?" \
  --agent rag-agent --json --db chroma-local

# Test with JSON + progress (JSON Lines format)
weave cols query TestDocs "What is machine learning?" \
  --agent rag-agent --json --progress --db chroma-local

# Test all three agents
weave cols query TestDocs "machine learning" --agent rag-agent --db chroma-local
weave cols query TestDocs "machine learning" --agent qa-agent --db chroma-local
weave cols query TestDocs "machine learning" --agent summarize-agent --db chroma-local
```

**Expected Results:**
- ✅ Agent loads and executes successfully
- ✅ Progress messages display correctly
- ✅ JSON output is valid
- ✅ JSON Lines format works with --json --progress
- ✅ Citations include source documents
- ✅ All three agents work correctly

#### 1.2 Qdrant Local

**Setup:**
```bash
# Start Qdrant locally
docker run -p 6333:6333 qdrant/qdrant:latest
```

**Test Commands:**
```bash
weave cols create TestDocs --db qdrant-local
weave docs add TestDocs "Test content" --db qdrant-local
weave cols query TestDocs "test" --agent rag-agent --db qdrant-local
weave cols query TestDocs "test" --agent rag-agent --json --progress --db qdrant-local
```

#### 1.3 Milvus Local

**Setup:**
```bash
# Start Milvus locally
docker-compose up -d  # Using Milvus docker-compose.yml
```

**Test Commands:**
```bash
weave cols create TestDocs --db milvus-local
weave docs add TestDocs "Test content" --db milvus-local
weave cols query TestDocs "test" --agent rag-agent --db milvus-local
weave cols query TestDocs "test" --agent qa-agent --progress --db milvus-local
```

#### 1.4 Neo4j Local

**Setup:**
```bash
# Start Neo4j locally
docker run -p 7474:7474 -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/password \
  neo4j:latest
```

**Test Commands:**
```bash
weave cols create TestDocs --db neo4j-local
weave docs add TestDocs "Test content" --db neo4j-local
weave cols query TestDocs "test" --agent rag-agent --db neo4j-local
weave cols query TestDocs "test" --agent summarize-agent --json --db neo4j-local
```

#### 1.5 Supabase Local

**Setup:**
```bash
# Start Supabase locally using Supabase CLI
supabase start
```

**Test Commands:**
```bash
weave cols create TestDocs --db supabase-local
weave docs add TestDocs "Test content" --db supabase-local
weave cols query TestDocs "test" --agent rag-agent --db supabase-local
weave cols query TestDocs "test" --agent rag-agent --progress --db supabase-local
```

#### 1.6 MongoDB Local

**Setup:**
```bash
# Start MongoDB locally
docker run -p 27017:27017 mongodb/mongodb-atlas-local:latest
```

**Test Commands:**
```bash
weave cols create TestDocs --db mongodb-local
weave docs add TestDocs "Test content" --db mongodb-local
weave cols query TestDocs "test" --agent rag-agent --db mongodb-local
weave cols query TestDocs "test" --agent qa-agent --json --progress --db mongodb-local
```

### Phase 2: Cloud VDB Testing (Priority: Medium)

Test with cloud-hosted vector databases to ensure API compatibility.

#### 2.1 Weaviate Cloud

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db weaviate-cloud
weave cols query Docs "test query" --agent rag-agent --json --progress --db weaviate-cloud
```

**Expected:** ✅ Should work (already tested in v0.9.0)

#### 2.2 Chroma Cloud

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db chroma-cloud
weave cols query Docs "test query" --agent qa-agent --progress --db chroma-cloud
```

#### 2.3 Qdrant Cloud

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db qdrant-cloud
weave cols query Docs "test query" --agent summarize-agent --json --db qdrant-cloud
```

#### 2.4 Milvus Cloud (Zilliz)

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db milvus-cloud
weave cols query Docs "test query" --agent rag-agent --progress --db milvus-cloud
```

#### 2.5 Neo4j Aura

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db neo4j-aura
weave cols query Docs "test query" --agent qa-agent --json --progress --db neo4j-aura
```

#### 2.6 Supabase Cloud

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db supabase-cloud
weave cols query Docs "test query" --agent summarize-agent --progress --db supabase-cloud
```

#### 2.7 MongoDB Atlas

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db mongodb-atlas
weave cols query Docs "test query" --agent rag-agent --json --progress --db mongodb-atlas
```

#### 2.8 Pinecone

**Test Commands:**
```bash
weave cols query Docs "test query" --agent rag-agent --db pinecone
weave cols query Docs "test query" --agent qa-agent --progress --db pinecone
```

### Phase 3: Edge Cases & Error Handling

#### 3.1 Empty Results

**Test:**
```bash
weave cols query TestDocs "zxqwertasdfg" --agent rag-agent
```

**Expected:** Agent handles no results gracefully

#### 3.2 Large Result Sets

**Test:**
```bash
weave cols query LargeDocs "common term" --agent summarize-agent --top_k 20
```

**Expected:** Agent handles many documents efficiently

#### 3.3 Missing API Keys

**Test:**
```bash
unset OPENAI_API_KEY
weave cols query TestDocs "test" --agent rag-agent
```

**Expected:** Clear error message about missing OPENAI_API_KEY

#### 3.4 Invalid Agent Name

**Test:**
```bash
weave cols query TestDocs "test" --agent nonexistent-agent
```

**Expected:** Clear error with available agents listed

#### 3.5 Network Failures

**Test:** Stop VDB mid-query

**Expected:** Clear error message, no hanging

### Phase 4: Performance Testing

#### 4.1 Response Time

**Test:**
```bash
time weave cols query TestDocs "test query" --agent rag-agent
```

**Expected:** < 5 seconds for typical query (LLM dependent)

#### 4.2 Progress Accuracy

**Test:**
```bash
weave cols query TestDocs "test" --agent rag-agent --progress
```

**Expected:** Progress messages appear in logical order with accurate timing

#### 4.3 Memory Usage

**Test:**
```bash
# Monitor memory while running agent query
weave cols query LargeDocs "test" --agent rag-agent --top_k 50
```

**Expected:** < 200MB memory usage

---

## Test Checklist

### Core Functionality

- [ ] **Chroma Local** - Basic agent query works
- [ ] **Chroma Local** - Progress indicator works
- [ ] **Chroma Local** - JSON output works
- [ ] **Chroma Local** - JSON Lines (--json --progress) works
- [ ] **Chroma Local** - All 3 agents work (rag, qa, summarize)
- [ ] **Qdrant Local** - All agent features work
- [ ] **Milvus Local** - All agent features work
- [ ] **Neo4j Local** - All agent features work
- [ ] **Supabase Local** - All agent features work
- [ ] **MongoDB Local** - All agent features work
- [ ] **Weaviate Cloud** - All agent features work (already tested)
- [ ] **Pinecone** - All agent features work

### Cloud Testing (When Available)

- [ ] **Chroma Cloud** - Basic functionality
- [ ] **Qdrant Cloud** - Basic functionality
- [ ] **Milvus Cloud** - Basic functionality
- [ ] **Neo4j Aura** - Basic functionality
- [ ] **Supabase Cloud** - Basic functionality
- [ ] **MongoDB Atlas** - Basic functionality

### Edge Cases

- [ ] Empty query results handled gracefully
- [ ] Large result sets (20+ documents) work
- [ ] Missing API key shows clear error
- [ ] Invalid agent name shows helpful error
- [ ] Network failures handled gracefully

### Performance

- [ ] Response time < 5s for typical queries
- [ ] Progress messages accurate and timely
- [ ] Memory usage reasonable (< 200MB)

---

## Test Automation

### Automated Test Script

```bash
#!/bin/bash
# test-agents-all-vdbs.sh

echo "Testing RAG Agents with All Vector Databases"
echo "=============================================="

AGENTS=("rag-agent" "qa-agent" "summarize-agent")
VDB_LOCAL=("chroma-local" "qdrant-local" "milvus-local" "neo4j-local" "supabase-local" "mongodb-local")
TEST_QUERY="What is machine learning?"

for vdb in "${VDB_LOCAL[@]}"; do
  echo ""
  echo "Testing $vdb..."

  for agent in "${AGENTS[@]}"; do
    echo "  - Testing $agent..."

    # Test basic query
    if weave cols query TestDocs "$TEST_QUERY" --agent "$agent" --db "$vdb" > /dev/null 2>&1; then
      echo "    ✅ Basic query passed"
    else
      echo "    ❌ Basic query failed"
    fi

    # Test with progress
    if weave cols query TestDocs "$TEST_QUERY" --agent "$agent" --progress --db "$vdb" > /dev/null 2>&1; then
      echo "    ✅ Progress passed"
    else
      echo "    ❌ Progress failed"
    fi

    # Test with JSON
    if weave cols query TestDocs "$TEST_QUERY" --agent "$agent" --json --db "$vdb" | jq . > /dev/null 2>&1; then
      echo "    ✅ JSON output passed"
    else
      echo "    ❌ JSON output failed"
    fi
  done
done

echo ""
echo "Testing complete!"
```

---

## Success Criteria

### Minimum Requirements

- ✅ All 3 agents work with Weaviate (already verified)
- ⏭️ All 3 agents work with at least 5 other VDBs
- ⏭️ Progress indicator works with all tested VDBs
- ⏭️ JSON output works with all tested VDBs
- ⏭️ JSON Lines format works with all tested VDBs

### Stretch Goals

- All 3 agents work with all 10+ supported VDBs
- Automated test suite for all VDB+agent combinations
- Performance benchmarks documented
- Edge cases all handled gracefully

---

## Documentation Updates

After testing completes:

1. Update `configs/agents/README.md` with confirmed VDB compatibility
2. Update `CHANGELOG.md` with test results
3. Update `docs/USER_GUIDE.md` with multi-VDB agent examples
4. Create troubleshooting guide for common issues

---

## Issues & Resolutions

| VDB | Issue | Resolution | Status |
|-----|-------|------------|--------|
| Chroma | TBD | TBD | Pending |
| Qdrant | TBD | TBD | Pending |
| Milvus | TBD | TBD | Pending |
| Neo4j | TBD | TBD | Pending |
| Supabase | TBD | TBD | Pending |
| MongoDB | TBD | TBD | Pending |
| Pinecone | TBD | TBD | Pending |

---

## Next Steps

1. **Set up local test environment** with Docker containers for each VDB
2. **Run Phase 1 tests** (local VDBs) and document results
3. **Fix any issues** discovered during testing
4. **Run Phase 2 tests** (cloud VDBs when available)
5. **Run Phase 3 & 4 tests** (edge cases and performance)
6. **Update documentation** with confirmed compatibility
7. **Create automated test suite** for CI/CD

---

## Related Documents

- [Multi-VDB Agent Support Planning](./AGENT_VDB_SUPPORT_AND_PROGRESS.md)
- [RAG Agent Feature Planning](./RAG_AGENT_FEATURE.md)
- [Agent Configuration Guide](../../configs/agents/README.md)
