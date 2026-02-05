---
name: OSS Feature - Embedding Comparison Reports
about: Generate automated comparison reports for different embedding models
title: 'feat: Add embedding comparison report generator'
labels: enhancement, oss, validation-tool
assignees: ''
---

## Feature Request: Embedding Comparison Reports

**Priority**: ⭐⭐ MEDIUM-HIGH

### Problem

No systematic way to compare embedding quality across different models. Users need:
- Reproducible comparison metrics
- Side-by-side results visualization  
- Quantitative evaluation (precision, recall, latency)
- Reports to share with stakeholders

### Proposed Solution

```bash
weave embeddings compare \
  --collections "MyCol_OpenAI,MyCol_OSS,MyCol_Nomic" \
  --queries "query1,query2,query3" \
  --output comparison.md
```

### Generated Report Example

```markdown
# Embedding Model Comparison

## Test Configuration
- Collections: MyCol_OpenAI (1536d), MyCol_OSS (768d), MyCol_Nomic (768d)
- Queries: 3 test queries
- Date: 2026-02-05

## Results Summary

| Model | Avg Precision@5 | Avg Latency | Cost |
|-------|----------------|-------------|------|
| OpenAI (text-3-large) | 0.92 | 45ms | $0.13/1M tokens |
| sentence-transformers | 0.87 | 12ms | $0 (OSS) |
| Nomic | 0.85 | 18ms | $0 (OSS) |

## Query 1: "vintage camera Leica M3"

### OpenAI Results (score: 0.94)
1. **Doc #1234** (sim: 0.91) - "Leica M3 rangefinder camera..."
2. **Doc #5678** (sim: 0.89) - "M3 introduced in 1954..."

### OSS Results (score: 0.89)
1. **Doc #1234** (sim: 0.87) - "Leica M3 rangefinder camera..."
2. **Doc #9012** (sim: 0.85) - "Vintage Leica cameras..."
```

### Use Case

**OSS validation workflow**:
1. Create collections with different embeddings
2. Run comparison with representative queries
3. Analyze precision/recall/latency tradeoffs
4. Select best model (balancing quality vs cost)

### Implementation Checklist

- [ ] Create `src/cmd/embeddings/compare.go`
- [ ] Query executor (parallel execution)
- [ ] Metrics calculator (precision@k, recall@k, latency)
- [ ] Report generator (markdown template)
- [ ] Support multiple output formats (md, json, html)
- [ ] Cost estimation (OpenAI vs OSS)
- [ ] Unit tests
- [ ] Integration tests
- [ ] Documentation

### Success Metrics

- ✅ Generate reproducible comparison reports
- ✅ Support 5+ metrics (precision, recall, latency, cost, etc.)
- ✅ Handle multiple collections (3-10)
- ✅ Export to multiple formats (markdown, JSON, HTML)

**Estimated time**: 6-8 hours

Related: #TBD (batch re-embedding), #TBD (auto-detect dimensions)
