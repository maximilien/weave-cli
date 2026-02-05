---
name: OSS Feature - OSS Quick-Start Template
about: Bootstrap OSS AI stack with single command
title: 'feat: Add OSS quick-start template generator'
labels: enhancement, oss, onboarding
assignees: ''
---

## Feature Request: OSS Quick-Start Template

**Priority**: ⭐ MEDIUM (onboarding tool)

### Problem

Setting up a fully OSS AI stack requires multiple manual steps:
- Docker Compose configuration
- Environment variables  
- Weave config files
- Container orchestration

High barrier for new OSS users (2+ hours setup time).

### Proposed Solution

```bash
weave quickstart oss \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --llm ollama:llama3.1:8b \
  --vdb milvus-local
```

**Generated structure**:
```
oss-project/
├── docker-compose.yml    # Milvus + Ollama containers
├── .env                  # OSS configuration
├── config.yaml           # Weave config  
├── oss-pipeline.yaml     # Sample RAG pipeline
└── README-OSS.md         # Getting started guide
```

### Use Case

**New OSS user workflow**:
1. Run `weave quickstart oss` (30 seconds)
2. Review generated files (2 minutes)
3. Run `docker-compose up` (5 minutes)
4. Test with sample pipeline (3 minutes)

**Total**: 10-15 minutes (vs 2+ hours manual setup)

### Implementation Checklist

- [ ] Create `src/cmd/quickstart.go`
- [ ] Template system (docker-compose, configs)
- [ ] Support: Milvus, Weaviate, Qdrant, Chroma
- [ ] Ollama model pull instructions
- [ ] sentence-transformers setup guide
- [ ] Sample RAG pipeline YAML
- [ ] Unit tests
- [ ] Integration tests
- [ ] Documentation

**Estimated time**: 8-10 hours

Related: #TBD (Ollama auto-discovery)
