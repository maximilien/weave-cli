# Agent Evaluation System - UPDATED Proposal

**Date:** 2026-01-22 (Updated)  
**Changes:** Agent management + Scope clarification

## 🔄 **Key Updates Based on Feedback**

### 1. Agent Management Commands (Phase 0 - Prerequisite)

**Current State:**
- ✅ `weave agents list` - List available agents (exists)
- ✅ `weave agents show <name>` - Show agent details (exists)
- ✅ `weave agents validate <file>` - Validate agent YAML (exists)

**Missing Commands (to add):**
```bash
# Create new agent from template
weave agents create <name> --type rag|qa|summarize [--interactive]

# Delete agent
weave agents delete <name>

# Edit agent (opens in $EDITOR)
weave agents edit <name>

# Copy/duplicate agent
weave agents copy <source> <target>
```

**Implementation Details:**

```go
// weave agents create
func NewCreateCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create NAME",
        Short: "Create a new agent configuration",
        Long: `Create a new agent configuration from template.

Available agent types:
  rag       - General-purpose RAG agent with citations
  qa        - Precise question-answering agent
  summarize - Document summarization agent
  custom    - Start with minimal template

Examples:
  weave agents create my-medical-agent --type rag
  weave agents create legal-qa --type qa --interactive`,
        Args: cobra.ExactArgs(1),
        Run:  runCreateAgent,
    }
    
    cmd.Flags().String("type", "rag", "Agent type: rag, qa, summarize, custom")
    cmd.Flags().Bool("interactive", false, "Interactive wizard for configuration")
    cmd.Flags().String("output", "configs/agents", "Output directory")
    
    return cmd
}

// weave agents delete
func NewDeleteCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "delete NAME",
        Short: "Delete an agent configuration",
        Long: `Delete an agent configuration file.

Examples:
  weave agents delete my-old-agent
  weave agents delete test-agent --force`,
        Args: cobra.ExactArgs(1),
        Run:  runDeleteAgent,
    }
    
    cmd.Flags().Bool("force", false, "Skip confirmation prompt")
    
    return cmd
}
```

**Why This Matters:**
- Users need to manage agents before evaluating them
- Create variants of agents for A/B testing
- Delete old/experimental agents

---

### 2. Clear Differentiation from Opik/Arize Phoenix

**What Opik/Arize Do Well (we should integrate with, not duplicate):**

| Feature | Opik | Arize Phoenix | Our Approach |
|---------|------|---------------|--------------|
| **Production Monitoring** | ✅ Excellent | ✅ Excellent | → Use Opik integration |
| **LLM-as-Judge** | ✅ Built-in | ✅ Built-in | → Use Opik for agent eval |
| **Conversation Tracing** | ✅ OpenTelemetry | ✅ OpenTelemetry | → Use Opik integration |
| **Dashboards/Visualization** | ✅ Production-ready | ✅ Enterprise-grade | → Link to Opik dashboards |
| **Team Collaboration** | ✅ Multi-user | ✅ Multi-user | → Not our focus |
| **Prompt Optimization** | ✅ Automated | ❌ | → Could integrate Opik |
| **Hallucination Detection** | ✅ Built-in | ✅ Built-in | → Use Opik evaluators |

**What Weave CLI Uniquely Provides (our focus):**

| Feature | Weave CLI | Opik | Arize | Unique Value |
|---------|-----------|------|-------|--------------|
| **Chunking Quality Eval** | ✅ **Core** | ❌ | ❌ | Semantic boundaries, density |
| **Embedding Comparison** | ✅ **Core** | ❌ | ❌ | Model cost-quality tradeoffs |
| **VDB Comparison** | ✅ **Core** | ❌ | ❌ | Same content across VDBs |
| **Retrieval Metrics** | ✅ **Core** | ⚠️ Basic | ⚠️ Basic | Precision@K, Recall@K for docs |
| **Pipeline Versioning** | ✅ **Core** | ❌ | ❌ | Full RAG config tracking |
| **Local/Offline Eval** | ✅ **Core** | ⚠️ Limited | ⚠️ Limited | No cloud needed, CI/CD ready |
| **CLI-First UX** | ✅ **Core** | ⚠️ SDK-first | ⚠️ SDK-first | Developer-friendly commands |
| **Git-Friendly Results** | ✅ **Core** | ❌ | ❌ | Markdown reports, diffs |

**REVISED SCOPE - Focus on Pipeline Evaluation:**

```
┌─────────────────────────────────────────────────────────────┐
│          WEAVE CLI EVALUATION SCOPE (Unique Value)           │
└─────────────────────────────────────────────────────────────┘

PRIMARY FOCUS (80% of effort):
1. ✅ Chunking Strategy Evaluation
   - Semantic vs fixed vs paragraph
   - Chunk size optimization
   - Overlap effectiveness
   - Boundary quality

2. ✅ Embedding Model Comparison
   - OpenAI small vs large
   - Cohere vs local models
   - Dimension impact (384, 1536, 3072)
   - Cost-quality tradeoffs

3. ✅ VDB Performance Comparison
   - Weaviate vs Qdrant vs Milvus vs Chroma
   - Search latency benchmarks
   - Cost comparisons
   - Migration decision support

4. ✅ Retrieval Quality Metrics
   - Precision@K, Recall@K (document level)
   - MRR, NDCG for ranking
   - Context relevance
   - Search parameter tuning

5. ✅ Pipeline Configuration Management
   - Version control for full pipeline
   - Component-level regression detection
   - Config diffing and comparison

SECONDARY FOCUS (20% of effort):
6. ⚠️ Basic Agent Evaluation (leverage Opik)
   - Accuracy (semantic similarity)
   - Citation validation
   - Basic hallucination detection
   - **→ For advanced: Use Opik integration**

NOT IN SCOPE (use Opik/Arize instead):
❌ Advanced hallucination detection (use Opik)
❌ Bias/toxicity evaluation (use Opik)
❌ PII redaction (use Opik guardrails)
❌ Automated prompt optimization (use Opik optimizer)
❌ Production dashboards (use Opik/Arize)
❌ Team collaboration features (use Opik/Arize)
❌ Multi-turn conversation tracking (use Opik)
```

**Integration Strategy:**

```yaml
# configs/evaluation.yaml

# Weave's local/quick evaluation
local_evaluation:
  enabled: true
  
  # FOCUS: Pipeline-specific metrics
  pipeline_metrics:
    - chunking_quality
    - embedding_comparison
    - vdb_performance
    - retrieval_precision
  
  # Basic agent metrics (local, fast)
  agent_metrics:
    - accuracy_semantic
    - citation_format
    
# Opik integration for advanced agent evaluation
opik_integration:
  enabled: true
  
  # What we send to Opik
  send_to_opik:
    - agent_traces
    - production_queries
    - evaluation_runs
  
  # What we use from Opik
  use_opik_for:
    - hallucination_detection  # Opik's LLM-as-judge
    - factual_accuracy        # Opik's evaluators
    - response_coherence      # Opik's evaluators
    - production_monitoring   # Opik's dashboards
    
  # Opik configuration
  endpoint: ${OTEL_EXPORTER_OTLP_ENDPOINT}
  api_key: ${OPIK_API_KEY}
  workspace: ${OPIK_WORKSPACE}
  project: weave-cli-evals
```

**User Workflow:**

```bash
# 1. Quick local pipeline evaluation (Weave CLI)
weave eval pipeline run \
  --config pipelines/v1.yaml \
  --dataset evals/datasets/baseline.yaml
# → Fast, local, no cloud needed
# → Focuses on chunking, VDB, retrieval

# 2. Compare chunking strategies (Weave CLI - unique)
weave eval compare-chunking \
  --strategies "semantic,fixed-512,paragraph" \
  --collection medical-docs
# → Weave-specific: chunking quality metrics

# 3. Compare VDBs (Weave CLI - unique)
weave eval compare-vdbs \
  --vdbs "weaviate,qdrant,milvus" \
  --collection medical-docs
# → Weave-specific: VDB performance comparison

# 4. Advanced agent evaluation (→ Opik)
weave eval agent run \
  --agent rag-agent \
  --dataset evals/datasets/baseline.yaml \
  --send-to-opik
# → Sends to Opik for advanced metrics
# → Uses Opik's hallucination, bias, toxicity evaluators
# → View results in Opik dashboard

# 5. Production monitoring (→ Opik)
# Opik automatically traces production queries
# View dashboards at opik.comet.com
```

---

### 3. Updated CLI Commands (Clarified Scope)

**Agent Management (Phase 0):**
```bash
weave agents create <name> --type rag|qa|summarize
weave agents list [--format json|yaml|text]
weave agents show <name> [--format json|yaml|text]
weave agents validate <file>
weave agents delete <name> [--force]
weave agents edit <name>
weave agents copy <source> <target>
```

**Pipeline Evaluation (Phase 5-6 - Our Unique Value):**
```bash
# Full pipeline evaluation
weave eval pipeline run --config <file> --dataset <file>

# Component comparisons (unique to Weave)
weave eval compare-chunking --strategies "s1,s2,s3" --collection <name>
weave eval compare-embeddings --models "m1,m2,m3" --collection <name>
weave eval compare-vdbs --vdbs "v1,v2,v3" --collection <name>
weave eval compare-pipelines --pipeline-a <file> --pipeline-b <file>

# Ingestion quality checks
weave eval check-ingestion --collection <name> --sample-docs 100
```

**Basic Agent Evaluation (Phase 1-4 - Opik integration):**
```bash
# Local quick evaluation
weave eval agent run --agent <name> --dataset <file>

# Send to Opik for advanced evaluation
weave eval agent run --agent <name> --dataset <file> --send-to-opik

# Compare agents (local basic metrics)
weave eval compare --agent-a <file> --agent-b <file>

# Compare agents (with Opik advanced metrics)
weave eval compare --agent-a <file> --agent-b <file> --use-opik
```

**Results & Reporting:**
```bash
weave eval show <run-id>
weave eval list [--agent <name>] [--pipeline <name>]
weave eval report <run-id> --format markdown|html|json
weave eval diff <run-id-1> <run-id-2>  # Git-friendly diffs
```

---

## Updated Implementation Plan

### **Phase 0: Agent Management (Week 0.5)** 🆕
**Goal:** Complete agent CRUD operations

- [ ] Implement `weave agents create`
  - Template-based creation
  - Interactive wizard
  - Type selection (rag, qa, summarize)
  
- [ ] Implement `weave agents delete`
  - Safe deletion with confirmation
  - Force flag for automation
  
- [ ] Implement `weave agents edit`
  - Opens in $EDITOR
  - Validates after edit
  
- [ ] Implement `weave agents copy`
  - Duplicate for experimentation

**Deliverable:** Full agent lifecycle management

---

### Phase 1-4: Basic Agent Evaluation (Weeks 1-6)
**REVISED SCOPE:** Focus on local evaluation + Opik integration

**What we build:**
- ✅ Accuracy (semantic similarity - local)
- ✅ Citation validation (local)
- ✅ Basic hallucination detection (local rule-based)
- ✅ Opik integration for advanced metrics
- ✅ Dataset management
- ✅ Basic CLI commands

**What we delegate to Opik:**
- ❌ Advanced hallucination (use Opik LLM-as-judge)
- ❌ Factual accuracy (use Opik evaluators)
- ❌ Bias/toxicity (use Opik evaluators)
- ❌ Production dashboards (use Opik UI)

---

### **Phase 5-6: RAG Pipeline Evaluation (Weeks 7-10)** ⭐ PRIMARY FOCUS
**REVISED:** This is our unique value proposition

**Chunking Evaluation (Week 7):**
- [ ] Chunk coherence evaluator
- [ ] Boundary quality checker
- [ ] Information density calculator
- [ ] Strategy comparison (semantic vs fixed vs paragraph)
- [ ] Chunk size optimizer

**Embedding Comparison (Week 7-8):**
- [ ] Model comparison framework
- [ ] Cost calculator (API pricing)
- [ ] Dimension impact analysis
- [ ] Performance benchmarking

**VDB Comparison (Week 8):**
- [ ] Multi-VDB test harness
- [ ] Performance comparison (latency, cost)
- [ ] Search quality metrics
- [ ] Migration decision support

**Retrieval Metrics (Week 9):**
- [ ] Precision@K, Recall@K
- [ ] MRR, NDCG calculators
- [ ] Context relevance evaluator
- [ ] Search parameter tuning

**Pipeline Versioning (Week 9-10):**
- [ ] Pipeline config schema
- [ ] Version control integration
- [ ] Config diffing tool
- [ ] Regression detection

**CLI Implementation:**
- [ ] `weave eval pipeline run`
- [ ] `weave eval compare-chunking`
- [ ] `weave eval compare-embeddings`
- [ ] `weave eval compare-vdbs`
- [ ] `weave eval compare-pipelines`
- [ ] `weave eval check-ingestion`

**Deliverable:** Complete RAG pipeline optimization system

---

## Success Metrics (Revised)

### Weave CLI Unique Value
- Users can compare 3+ chunking strategies quantitatively
- VDB migration decisions backed by data
- Embedding model selection optimized for cost-quality
- <5 min local pipeline evaluation (no cloud needed)
- Git-friendly evaluation reports (markdown diffs)

### Integration Success
- Seamless Opik integration for advanced agent metrics
- Production traces automatically sent to Opik
- Users can choose: local (fast) or Opik (comprehensive)
- No duplication of Opik/Arize features

---

## Summary of Changes

### ✅ Added (Based on Feedback):

1. **Phase 0: Agent Management**
   - `weave agents create/delete/edit/copy`
   - Prerequisite for evaluation

2. **Clear Differentiation**
   - Focus 80% on pipeline evaluation (unique value)
   - Focus 20% on basic agent eval + Opik integration
   - Explicit "Not In Scope" list (use Opik instead)

3. **Revised Scope**
   - Primary: Chunking, VDB, embedding, retrieval
   - Secondary: Basic agent eval (delegate advanced to Opik)
   - Integration: Seamless Opik handoff for production/advanced metrics

### 📊 Comparison Table:

| Capability | Weave CLI | Opik | Who Owns |
|------------|-----------|------|----------|
| Chunking Quality | ✅ Core | ❌ | Weave |
| Embedding Comparison | ✅ Core | ❌ | Weave |
| VDB Comparison | ✅ Core | ❌ | Weave |
| Retrieval Metrics | ✅ Core | ⚠️ Basic | Weave |
| Basic Agent Eval | ✅ Local | ❌ | Weave |
| Advanced Hallucination | ⚠️ Delegate | ✅ Core | Opik |
| Production Monitoring | ⚠️ Delegate | ✅ Core | Opik |
| Prompt Optimization | ⚠️ Delegate | ✅ Core | Opik |
| Bias/Toxicity | ⚠️ Delegate | ✅ Core | Opik |
| Team Dashboards | ⚠️ Delegate | ✅ Core | Opik |

**Timeline Impact:** +0.5 weeks for Phase 0 (agent management)
**Total Timeline:** ~10.5 weeks for complete system

---

**Ready for implementation when approved!** 🚀
