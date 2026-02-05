---
name: OSS Feature - Auto-Detect Embedding Dimensions
about: Automatically infer vector dimensions from embedding model name
title: 'feat: Auto-detect embedding dimensions from model name'
labels: enhancement, oss, dx-improvement
assignees: ''
---

## Feature Request: Auto-Detect Embedding Dimensions

**Priority**: ⭐⭐ HIGH

### Problem

Users must manually specify vector dimensions when creating collections. This causes:
- **Configuration errors**: Wrong dimensions → collection creation fails
- **Poor UX**: Users must look up dimensions for each model
- **Barrier for new users**: Not obvious where to find dimension info

**Current workflow** (error-prone):
```bash
# User must know that all-mpnet-base-v2 → 768 dimensions
weave collection create MyCol \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --dimensions 768  # ← Manual, error-prone
```

**Common mistakes**:
```bash
# Wrong dimensions → fails
weave collection create MyCol \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --dimensions 384  # ❌ Wrong! Should be 768
```

### Proposed Solution

Auto-detect dimensions from well-known embedding models:

```bash
# Dimensions auto-detected from model name
weave collection create MyCol \
  --embedding sentence-transformers/all-mpnet-base-v2
# ✓ Auto-detects 768 dimensions

# Manual override still supported
weave collection create MyCol \
  --embedding custom-model \
  --dimensions 1024  # ← Explicit override
```

### Model Registry

**sentence-transformers**:
```
all-mpnet-base-v2           → 768
all-MiniLM-L6-v2            → 384
all-MiniLM-L12-v2           → 384
paraphrase-multilingual     → 768
multi-qa-mpnet-base         → 768
```

**OpenAI**:
```
text-embedding-3-small      → 1536
text-embedding-3-large      → 3072
text-embedding-ada-002      → 1536
```

**Ollama**:
```
nomic-embed-text            → 768
mxbai-embed-large           → 1024
```

**Cohere**:
```
embed-english-v3.0          → 1024
embed-multilingual-v3.0     → 1024
```

### Technical Approach

**Implementation**:
```go
// src/pkg/embeddings/model_registry.go

type ModelInfo struct {
    Name       string
    Dimensions int
    Provider   string
    MaxTokens  int
}

var knownModels = map[string]ModelInfo{
    "sentence-transformers/all-mpnet-base-v2": {
        Dimensions: 768,
        Provider:   "sentence-transformers",
        MaxTokens:  512,
    },
    "text-embedding-3-large": {
        Dimensions: 3072,
        Provider:   "openai",
        MaxTokens:  8191,
    },
    // ... more models
}

func GetModelDimensions(modelName string) (int, error)
func GetModelInfo(modelName string) (*ModelInfo, error)
```

**Integration**:
```go
// In collection creation
if dimensions == 0 {
    auto, err := embeddings.GetModelDimensions(embeddingModel)
    if err == nil {
        dimensions = auto
        fmt.Printf("Auto-detected %d dimensions for %s\n", dimensions, embeddingModel)
    } else {
        return fmt.Errorf("could not auto-detect dimensions, please specify --dimensions")
    }
}
```

### User Experience

**Before** (manual):
```bash
$ weave collection create MyCol --embedding sentence-transformers/all-mpnet-base-v2
Error: --dimensions required

$ weave collection create MyCol \
    --embedding sentence-transformers/all-mpnet-base-v2 \
    --dimensions 768  # User had to Google this
✓ Collection created
```

**After** (auto-detected):
```bash
$ weave collection create MyCol --embedding sentence-transformers/all-mpnet-base-v2
Auto-detected 768 dimensions for sentence-transformers/all-mpnet-base-v2
✓ Collection created
```

### Implementation Checklist

- [ ] Create `src/pkg/embeddings/model_registry.go`
- [ ] Add 15+ well-known models to registry
- [ ] `GetModelDimensions(modelName string) (int, error)` function
- [ ] `GetModelInfo(modelName string) (*ModelInfo, error)` function
- [ ] Integrate into `weave collection create`
- [ ] Integrate into `weave collection re-embed` (see #TBD)
- [ ] Add `--auto-detect-dimensions` flag (default: true)
- [ ] Unit tests for registry
- [ ] Integration tests
- [ ] Documentation
- [ ] Update examples in README

### Success Metrics

- ✅ Reduce config errors by ~80% (from user testing)
- ✅ Improve new user onboarding (no dimension lookup required)
- ✅ Support 15+ popular embedding models
- ✅ Allow manual override for custom models

### Future Enhancements

**Dynamic detection** (future):
- Query OpenAI API for model info
- Query HuggingFace for sentence-transformers config
- Auto-update registry from external sources

**Validation**:
- Warn if user specifies wrong dimensions
- Suggest correct dimensions if mismatch detected

### Related Features

- #TBD - Batch re-embedding command
- #TBD - Embedding comparison reports

### Community Impact

**Who benefits**:
- New users (easier onboarding)
- OSS users (sentence-transformers, Ollama models covered)
- Anyone testing multiple embedding models
- Reduces GitHub issues about dimension errors

**Estimated impact**: 50% reduction in configuration-related support questions
