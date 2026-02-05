---
name: OSS Feature - Ollama Model Auto-Discovery
about: Show locally installed Ollama models in interactive wizards
title: 'feat: Add Ollama model auto-discovery'
labels: enhancement, oss, dx-improvement
assignees: ''
---

## Feature Request: Ollama Model Auto-Discovery

**Priority**: ⭐ LOW-MEDIUM (nice-to-have)

### Problem

Users must manually type Ollama model names in interactive wizards:
- Easy to make typos (llama3.1:8b vs llama3:8b)
- Don't know which models are installed locally
- Can't see model sizes, last used timestamps

### Proposed Solution

Enhance `weave config agents` to show installed Ollama models:

```bash
$ weave config agents

Select LLM:
  1. OpenAI (gpt-4, gpt-3.5-turbo)
  2. Ollama - Locally Installed:
     • llama3.1:8b (4.7GB, used 2 days ago)
     • kimi-k2 (15GB, used 1 hour ago)
     • nomic-embed-text (274MB, used today)
  3. Anthropic (claude-3-5-sonnet, claude-3-opus)
  4. Custom model name
```

### Technical Approach

Call `ollama list` to discover models:

```bash
$ ollama list
NAME                    ID              SIZE      MODIFIED
llama3.1:8b            a1b2c3d4        4.7 GB    2 days ago
kimi-k2                e5f6g7h8        15 GB     1 hour ago
nomic-embed-text       i9j0k1l2        274 MB    1 minute ago
```

Parse output and present in wizard.

### Implementation Checklist

- [ ] Create `src/pkg/llm/ollama/discovery.go`
- [ ] Execute `ollama list` command
- [ ] Parse model names, sizes, timestamps
- [ ] Integrate into `weave config agents` wizard
- [ ] Graceful fallback if Ollama not installed
- [ ] Unit tests
- [ ] Integration tests
- [ ] Documentation

**Estimated time**: 4-5 hours

Related: #TBD (OSS quickstart)
