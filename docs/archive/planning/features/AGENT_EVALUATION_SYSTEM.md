# Agent Evaluation System - Design Proposal

**Date:** 2026-01-22  
**Status:** Proposal  
**Priority:** High  
**Target Version:** v1.0

## Executive Summary

This proposal outlines a comprehensive agent evaluation system for Weave CLI that enables users to:
1. Evaluate the quality and accuracy of agent responses
2. Compare different agent configurations (prompts, models, parameters)
3. Track agent performance over time
4. Integrate with industry-standard evaluation frameworks (Opik, LangSmith, etc.)
5. Create custom evaluation metrics for domain-specific needs

## Problem Statement

### Current State
- Agents are configured via YAML files with various prompts, models, and parameters
- No systematic way to evaluate if agent responses are accurate or improving
- Users manually test agents by running queries and subjectively judging outputs
- No comparison mechanism between different agent configurations
- Limited visibility into agent performance metrics

### User Needs
- **Agent Tuning:** Users need to iterate on agent YAML configs (prompts, temperature, etc.) and know if changes improve or degrade performance
- **Quality Assurance:** Ensure agents provide accurate, relevant responses
- **Version Comparison:** Compare new agent versions against baselines
- **Production Monitoring:** Track agent performance in real-world usage
- **Custom Metrics:** Domain-specific evaluation criteria (e.g., medical accuracy, legal compliance)

## Proposed Solution

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Weave CLI Agent System                    │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐      ┌───────────┐ │
│  │ Agent Config │─────▶│    Agent     │─────▶│  Response │ │
│  │  (YAML)      │      │   Executor   │      │           │ │
│  └──────────────┘      └──────────────┘      └─────┬─────┘ │
│                                                      │       │
└──────────────────────────────────────────────────────┼───────┘
                                                       │
                                                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  Evaluation System (NEW)                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Evaluation Dataset                        │  │
│  │  • Test queries + expected answers                     │  │
│  │  • Ground truth data                                   │  │
│  │  • User-provided examples                             │  │
│  │  • Synthetic test generation                          │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                 │
│                            ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           Evaluation Metrics Engine                   │  │
│  │  ┌────────────────┐  ┌────────────────┐             │  │
│  │  │  Built-in      │  │   Custom       │             │  │
│  │  │  Evaluators    │  │  Evaluators    │             │  │
│  │  │                │  │                │             │  │
│  │  │ • Accuracy     │  │ • User Python  │             │  │
│  │  │ • Hallucination│  │ • Domain-spec  │             │  │
│  │  │ • Relevance    │  │ • Business KPI │             │  │
│  │  │ • Citation     │  │                │             │  │
│  │  │ • Coherence    │  │                │             │  │
│  │  │ • Completeness │  │                │             │  │
│  │  └────────────────┘  └────────────────┘             │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                 │
│                            ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         External Integration Layer                    │  │
│  │  ┌──────────┐  ┌───────────┐  ┌──────────────┐     │  │
│  │  │  Opik    │  │LangSmith  │  │  Custom      │     │  │
│  │  │  (OpenTel│  │           │  │  Tracking    │     │  │
│  │  │  Traces) │  │           │  │              │     │  │
│  │  └──────────┘  └───────────┘  └──────────────┘     │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                 │
│                            ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Results Storage & Comparison                  │  │
│  │  • Eval runs (timestamped)                           │  │
│  │  • Metric scores per agent version                   │  │
│  │  • Comparison reports                                │  │
│  │  • Regression detection                              │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Components

#### 1. Evaluation Dataset Management

**Definition:** Collections of test cases for agent evaluation

```yaml
# evals/datasets/rag-agent-baseline.yaml
name: rag-agent-baseline-v1
description: Baseline test cases for RAG agent
version: "1.0.0"

test_cases:
  - id: tc-001
    query: "What is the capital of France?"
    expected_answer: "Paris"
    metadata:
      category: factual
      difficulty: easy
    
  - id: tc-002
    query: "Explain how photosynthesis works"
    expected_answer: |
      Photosynthesis is the process by which plants convert
      light energy into chemical energy...
    reference_sources:
      - source_id: biology-101
        expected_citation: true
    metadata:
      category: explanatory
      difficulty: medium
      requires_synthesis: true
    
  - id: tc-003
    query: "What are the main differences between RNA and DNA?"
    expected_answer: "DNA contains deoxyribose sugar while RNA contains ribose..."
    evaluation_criteria:
      - completeness: must mention 3+ key differences
      - citation_required: true
      - hallucination_check: strict
```

**Features:**
- YAML-based test case definitions
- Support for various answer types (factual, explanatory, comparative)
- Metadata for categorization and filtering
- Reference sources for citation validation
- Custom evaluation criteria per test case

#### 2. Built-in Evaluators

**a) Accuracy Evaluator**
```go
type AccuracyEvaluator struct {
    LLMClient llm.Client
}

// Evaluates semantic similarity between expected and actual answers
func (e *AccuracyEvaluator) Evaluate(ctx context.Context, testCase TestCase, response AgentResponse) (Score, error)
```

**Metrics:**
- Semantic similarity (0-1) using embeddings
- Exact match (boolean)
- Contains key facts (partial credit)

**b) Hallucination Evaluator**
```go
type HallucinationEvaluator struct {
    LLMClient llm.Client  // Uses LLM-as-judge
}

// Detects if response contains information not in source documents
func (e *HallucinationEvaluator) Evaluate(ctx context.Context, testCase TestCase, response AgentResponse) (Score, error)
```

**Detection methods:**
- Source attribution check
- Fact verification against sources
- LLM-as-judge for subtle hallucinations

**c) Citation Evaluator**
```go
type CitationEvaluator struct{}

// Validates citation format, completeness, and accuracy
func (e *CitationEvaluator) Evaluate(ctx context.Context, testCase TestCase, response AgentResponse) (Score, error)
```

**Checks:**
- All facts are cited
- Citations reference actual sources
- Citation format is correct
- No orphaned citations

**d) Relevance Evaluator**
```go
type RelevanceEvaluator struct {
    LLMClient llm.Client
}

// Measures if response addresses the query
func (e *RelevanceEvaluator) Evaluate(ctx context.Context, testCase TestCase, response AgentResponse) (Score, error)
```

**e) Coherence Evaluator**
```go
type CoherenceEvaluator struct {
    LLMClient llm.Client
}

// Evaluates response structure, flow, and readability
func (e *CoherenceEvaluator) Evaluate(ctx context.Context, testCase TestCase, response AgentResponse) (Score, error)
```

**f) Completeness Evaluator**
```go
type CompletenessEvaluator struct {
    LLMClient llm.Client
}

// Checks if response covers all aspects of the query
func (e *CompletenessEvaluator) Evaluate(ctx context.Context, testCase TestCase, response AgentResponse) (Score, error)
```

#### 3. Custom Evaluators

**Python-based Custom Evaluators:**
```yaml
# evals/evaluators/medical-accuracy.yaml
name: medical-accuracy
type: custom
language: python
script: evals/scripts/medical_accuracy.py

parameters:
  strict_mode: true
  approved_sources_only: true
  require_disclaimers: true
```

```python
# evals/scripts/medical_accuracy.py
def evaluate(test_case, response, context):
    """
    Custom medical accuracy evaluator
    
    Args:
        test_case: TestCase object with query and expected answer
        response: AgentResponse with actual answer
        context: Additional context (sources, metadata)
    
    Returns:
        EvalResult with score and explanation
    """
    score = 0.0
    issues = []
    
    # Check for medical disclaimers
    if not has_disclaimer(response.text):
        issues.append("Missing medical disclaimer")
        score -= 0.2
    
    # Verify against approved medical sources
    if not all_sources_approved(response.sources):
        issues.append("Uses non-approved sources")
        score -= 0.3
    
    # Check for outdated information
    if has_outdated_info(response.text, context):
        issues.append("Contains potentially outdated information")
        score -= 0.5
    
    return EvalResult(
        score=max(0.0, 1.0 + score),
        passed=len(issues) == 0,
        issues=issues,
        metadata={"evaluator": "medical-accuracy-v1"}
    )
```

#### 4. Opik Integration

**Leverage existing Opik integration:**

```go
// src/pkg/agents/eval_integration.go
type OpikEvaluator struct {
    tracerProvider *sdktrace.TracerProvider
    projectName    string
    workspace      string
}

func (e *OpikEvaluator) TraceEvaluation(ctx context.Context, evalRun EvalRun) error {
    // Create Opik trace for evaluation run
    tracer := otel.Tracer("weave-cli-eval")
    ctx, span := tracer.Start(ctx, "agent-evaluation")
    defer span.End()
    
    // Add evaluation metadata
    span.SetAttributes(
        attribute.String("agent.name", evalRun.AgentName),
        attribute.String("agent.version", evalRun.AgentVersion),
        attribute.String("dataset.name", evalRun.DatasetName),
        attribute.Int("test_cases.count", len(evalRun.TestCases)),
        attribute.Float64("overall.score", evalRun.OverallScore),
    )
    
    // Trace individual test case evaluations
    for _, tc := range evalRun.TestCases {
        e.traceTestCase(ctx, tc)
    }
    
    return nil
}

func (e *OpikEvaluator) traceTestCase(ctx context.Context, tc TestCaseResult) {
    tracer := otel.Tracer("weave-cli-eval")
    _, span := tracer.Start(ctx, "test-case")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("test_case.id", tc.ID),
        attribute.String("query", tc.Query),
        attribute.Float64("accuracy.score", tc.Scores["accuracy"]),
        attribute.Float64("hallucination.score", tc.Scores["hallucination"]),
        attribute.Float64("relevance.score", tc.Scores["relevance"]),
        attribute.Bool("passed", tc.Passed),
    )
}
```

**Benefits:**
- Opik's LLM-as-judge evaluators
- Production trace monitoring
- Evaluation history and trends
- Comparison dashboards
- Cost tracking per evaluation

#### 5. CLI Commands

**New `weave eval` command:**

```bash
# Run evaluation on agent
weave eval run --agent rag-agent --dataset evals/datasets/baseline.yaml

# Run evaluation comparing two agents
weave eval compare \
  --agent-a configs/agents/rag-agent-v1.yaml \
  --agent-b configs/agents/rag-agent-v2.yaml \
  --dataset evals/datasets/baseline.yaml

# Show evaluation results
weave eval show --run-id eval-123456

# List all evaluation runs
weave eval list --agent rag-agent

# Generate evaluation report
weave eval report --run-id eval-123456 --format markdown > report.md

# Create new test dataset
weave eval dataset create --name medical-qa --interactive

# Validate agent against production data
weave eval validate --agent rag-agent --production-sample 100
```

**Example output:**
```
Evaluation Run: eval-20260122-143022
Agent: rag-agent-v2.yaml
Dataset: baseline-v1.yaml (25 test cases)
Duration: 2m 34s

Overall Metrics:
  ✓ Accuracy:       0.87  (target: 0.80)  ⬆ +0.12 vs v1
  ✓ Hallucination:  0.92  (target: 0.90)  ⬆ +0.08 vs v1
  ✓ Relevance:      0.89  (target: 0.85)  ⬆ +0.05 vs v1
  ✗ Citation:       0.73  (target: 0.80)  ⬇ -0.02 vs v1
  ✓ Coherence:      0.85  (target: 0.75)  ⬆ +0.10 vs v1
  ✓ Completeness:   0.81  (target: 0.80)  ⬆ +0.03 vs v1

Pass Rate: 88% (22/25 test cases)

Failed Test Cases:
  tc-007: Low citation score (0.40)
    Issue: Missing sources for 3 key facts
  tc-015: Hallucination detected (0.60)
    Issue: Added information not in sources
  tc-023: Low accuracy (0.55)
    Issue: Incomplete answer, missed 2 key points

Cost: $0.45 (LLM tokens)

Recommendation: ✓ Agent v2 shows improvement in 5/6 metrics
  Action: Review citation configuration before production deployment

View full report: weave eval show eval-20260122-143022
```

### Configuration

#### Agent YAML Extension

Add evaluation configuration to agent YAMLs:

```yaml
# configs/agents/rag-agent.yaml
name: rag-agent
version: "2.0.0"
# ... existing config ...

# NEW: Evaluation configuration
evaluation:
  enabled: true
  
  # Default evaluation metrics for this agent
  default_metrics:
    - accuracy
    - hallucination
    - relevance
    - citation
    - coherence
  
  # Thresholds for production deployment
  thresholds:
    accuracy: 0.80
    hallucination: 0.90
    relevance: 0.85
    citation: 0.80
    coherence: 0.75
  
  # Custom evaluators for this agent type
  custom_evaluators:
    - name: medical-accuracy
      enabled: true
      weight: 2.0  # Double weight for critical metrics
    
  # Opik integration
  opik:
    enabled: true
    track_all_queries: false  # Only track eval runs
    project_name: "weave-rag-agent"
  
  # Regression detection
  regression_detection:
    enabled: true
    baseline_version: "v1.8.0"
    alert_on_degradation: true
    degradation_threshold: 0.05  # Alert if any metric drops >5%
```

#### Evaluation Config

```yaml
# configs/evaluation.yaml
evaluation:
  # Global settings
  llm_as_judge:
    provider: openai
    model: gpt-4o
    temperature: 0.1
  
  # Dataset locations
  datasets_dir: evals/datasets
  
  # Results storage
  results_dir: evals/results
  retention_days: 90
  
  # Opik configuration
  opik:
    enabled: true
    endpoint: ${OTEL_EXPORTER_OTLP_ENDPOINT}
    api_key: ${OPIK_API_KEY}
    workspace: ${OPIK_WORKSPACE}
    default_project: weave-cli-evals
  
  # Performance
  parallel_evaluation: true
  max_concurrent_evaluators: 5
  timeout_per_test_case: 30s
  
  # Cost controls
  budget:
    max_cost_per_run: 5.00  # USD
    warn_above: 2.00
```

## Implementation Plan

### Phase 1: Foundation (Week 1-2)
**Goal:** Basic evaluation infrastructure

- [ ] Create evaluation data structures
  - TestCase, EvalRun, EvalResult types
  - Dataset YAML parser
  - Results storage schema

- [ ] Implement core evaluators
  - Accuracy evaluator (semantic similarity)
  - Basic hallucination detector
  - Citation validator

- [ ] Basic CLI commands
  - `weave eval run`
  - `weave eval show`
  - `weave eval list`

**Deliverable:** Can run basic evals on agents and view results

### Phase 2: Opik Integration (Week 3)
**Goal:** External tracking and monitoring

- [ ] Opik trace integration
  - Trace evaluation runs
  - Send metrics to Opik
  - Dashboard configuration

- [ ] Enhanced evaluators
  - LLM-as-judge implementation
  - Relevance evaluator
  - Coherence evaluator

- [ ] Comparison functionality
  - `weave eval compare` command
  - Delta calculations
  - Regression detection

**Deliverable:** Full Opik integration with comparison features

### Phase 3: Custom Evaluators (Week 4)
**Goal:** User extensibility

- [ ] Custom evaluator framework
  - Python script execution
  - Evaluator plugin system
  - Parameter passing

- [ ] Dataset management
  - Dataset creation wizard
  - Import/export formats
  - Version control integration

- [ ] Reporting
  - Markdown reports
  - HTML dashboards
  - CI/CD integration

**Deliverable:** Users can create custom evaluators and generate reports

### Phase 4: Production Features (Week 5-6)
**Goal:** Production-ready evaluation

- [ ] Production sampling
  - Random production trace sampling
  - Continuous evaluation
  - Anomaly detection

- [ ] Advanced features
  - A/B testing framework
  - Automatic prompt optimization
  - Cost-quality tradeoff analysis

- [ ] Documentation & Examples
  - User guide
  - Example evaluators
  - Best practices

**Deliverable:** Production-ready evaluation system

## Success Metrics

### User Success
- Users can quantitatively compare agent versions
- Evaluation runs complete in <5 minutes for 100 test cases
- Users create and deploy custom evaluators
- Regression detection prevents quality degradation

### System Success
- >90% test case evaluation success rate
- <$1 average cost per evaluation run
- <10s latency per test case evaluation
- Integration with Opik for all traces

### Business Success
- Faster agent iteration cycles
- Reduced manual testing effort
- Higher agent quality in production
- Better visibility into agent performance

## Risk Mitigation

### Technical Risks

**Risk:** LLM-as-judge evaluations are slow and expensive
- **Mitigation:** Implement caching, parallel execution, cost budgets
- **Mitigation:** Provide cheaper heuristic evaluators as alternatives
- **Mitigation:** Sample-based evaluation for large datasets

**Risk:** Custom evaluators create security vulnerabilities
- **Mitigation:** Sandbox Python execution
- **Mitigation:** Code review for built-in evaluators
- **Mitigation:** Restrict file system access

**Risk:** Opik integration failures
- **Mitigation:** Graceful degradation when Opik unavailable
- **Mitigation:** Local-only evaluation mode
- **Mitigation:** Retry logic and error handling

### User Experience Risks

**Risk:** Complex configuration overwhelms users
- **Mitigation:** Sensible defaults for all agents
- **Mitigation:** Wizard for dataset creation
- **Mitigation:** Example evaluators and datasets

**Risk:** Evaluation results are hard to interpret
- **Mitigation:** Clear, actionable reports
- **Mitigation:** Visual comparisons
- **Mitigation:** Specific recommendations

## Future Enhancements

### V2 Features
- Automatic test case generation from production logs
- Multi-agent evaluation scenarios
- Evaluation-driven prompt optimization
- Integration with other platforms (LangSmith, Weights & Biases)
- Continuous evaluation pipelines
- Evaluation metric leaderboards

### Research Directions
- Adversarial testing for robustness
- Bias detection in agent responses
- Cross-lingual evaluation
- Multi-modal evaluation (text + images)

## Open Questions

1. **Dataset Management:** Should we provide pre-built evaluation datasets for common domains?
2. **Cost Optimization:** What's the right balance between evaluation quality and cost?
3. **Real-time vs Batch:** Should we support real-time evaluation during queries?
4. **Metrics Standardization:** Should we align with emerging industry standards (e.g., HELM, BIG-Bench)?

## Appendix

### Example Evaluation Flow

```mermaid
graph TD
    A[User edits agent YAML] --> B[Run evaluation]
    B --> C[Load test dataset]
    C --> D[Execute each test case]
    D --> E[Run built-in evaluators]
    D --> F[Run custom evaluators]
    E --> G[Aggregate scores]
    F --> G
    G --> H[Send to Opik]
    G --> I[Store local results]
    H --> J[View in Opik dashboard]
    I --> K[Generate report]
    K --> L[Compare with baseline]
    L --> M{Regression?}
    M -->|Yes| N[Alert user]
    M -->|No| O[✓ Ready for deployment]
```

### Related Work

- **OpenAI Evals:** https://github.com/openai/evals
- **LangChain Evaluators:** https://python.langchain.com/docs/guides/evaluation
- **HELM Benchmark:** https://crfm.stanford.edu/helm/
- **Opik Documentation:** https://www.comet.com/docs/opik/

### References

- [Opik GitHub](https://github.com/comet-ml/opik)
- [LLM-as-judge Best Practices](https://www.comet.com/site/products/opik/)
- [RAG Evaluation Patterns](https://www.comet.com/docs/opik/cookbook/rag)

---

## AMENDMENT: Full RAG Pipeline Evaluation

**Date:** 2026-01-22  
**Status:** Extended Proposal

### Motivation

The agent is only one part of the RAG pipeline. The quality of agent responses depends heavily on:
1. **Document Chunking** - Strategy, size, overlap affect retrieval
2. **Embedding/Vectorization** - Model choice impacts semantic search quality
3. **Vector Database** - Different VDBs have different search characteristics
4. **Collection Schema** - Field configuration affects filtering and ranking

**The same agent with different chunking or VDB will produce different results.**

### Extended Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    FULL RAG PIPELINE EVALUATION                   │
└──────────────────────────────────────────────────────────────────┘

┌──────────────┐     ┌─────────────┐     ┌─────────────┐     ┌──────────┐
│   Document   │────▶│  Chunking   │────▶│ Vectorizer  │────▶│   VDB    │
│  Ingestion   │     │  Strategy   │     │  (Embedder) │     │ Storage  │
└──────────────┘     └─────────────┘     └─────────────┘     └────┬─────┘
                           │                    │                   │
                           │                    │                   │
                     [EVAL POINT 1]        [EVAL POINT 2]      [EVAL POINT 3]
                     - Chunk quality       - Embedding          - Index quality
                     - Overlap             - Dimension          - Search accuracy
                     - Boundaries          - Model choice       - Distance metric

                                                                     │
                                                                     ▼
┌──────────────┐     ┌─────────────┐     ┌─────────────┐     ┌────────────┐
│    Query     │────▶│   Vector    │────▶│  Retrieved  │────▶│   Agent    │
│              │     │   Search    │     │   Chunks    │     │  Response  │
└──────────────┘     └─────────────┘     └─────────────┘     └────────────┘
                           │                    │                   │
                           │                    │                   │
                     [EVAL POINT 4]        [EVAL POINT 5]      [EVAL POINT 6]
                     - Search params       - Retrieval          - Generation
                     - Filters             - Precision/Recall   - Accuracy
                     - TopK               - Relevance          - Hallucination
```

### New Evaluation Points

#### 1. Chunking Quality Evaluation

**Metrics:**
- **Chunk Coherence** - Do chunks maintain semantic completeness?
- **Boundary Quality** - Are chunks split at natural boundaries?
- **Information Density** - How much useful information per chunk?
- **Overlap Effectiveness** - Does overlap improve retrieval?

**Evaluator:**
```go
type ChunkingEvaluator struct {
    LLMClient llm.Client
}

func (e *ChunkingEvaluator) Evaluate(ctx context.Context, document Document, chunks []Chunk) (*ChunkingMetrics, error) {
    return &ChunkingMetrics{
        AverageCoherence:     e.evaluateCoherence(chunks),
        BoundaryQuality:      e.evaluateBoundaries(chunks),
        InformationDensity:   e.calculateDensity(chunks),
        OptimalChunkSize:     e.recommendChunkSize(chunks),
        SemanticCompleteness: e.checkCompleteness(chunks),
    }, nil
}
```

**Test Dataset Extension:**
```yaml
# evals/datasets/rag-pipeline-test.yaml
test_cases:
  - id: tc-001
    query: "What is the capital of France?"
    expected_answer: "Paris"
    
    # NEW: Collection/VDB context
    collection_context:
      collection_name: "geography-facts"
      chunking_strategy: "semantic"
      chunk_size: 512
      chunk_overlap: 50
      embedding_model: "text-embedding-3-small"
      vector_db: "weaviate-cloud"
      
    # NEW: Expected retrieval behavior  
    retrieval_expectations:
      min_relevant_chunks: 1
      expected_chunks:
        - chunk_id: "doc-france-001-chunk-3"
          expected_rank: 1
          min_score: 0.85
      max_irrelevant_chunks: 0
```

#### 2. Retrieval Quality Evaluation

**Metrics:**
- **Precision@K** - What % of retrieved chunks are relevant?
- **Recall@K** - What % of relevant chunks were retrieved?
- **MRR (Mean Reciprocal Rank)** - How quickly is the first relevant chunk found?
- **NDCG (Normalized Discounted Cumulative Gain)** - Ranking quality
- **Context Relevance** - How relevant are chunks to the query?

**Evaluator:**
```go
type RetrievalEvaluator struct {
    VectorDB vectordb.VectorDBClient
}

func (e *RetrievalEvaluator) Evaluate(ctx context.Context, query string, retrievedChunks []Chunk, groundTruth []string) (*RetrievalMetrics, error) {
    return &RetrievalMetrics{
        Precision:        e.calculatePrecision(retrievedChunks, groundTruth),
        Recall:           e.calculateRecall(retrievedChunks, groundTruth),
        MRR:              e.calculateMRR(retrievedChunks, groundTruth),
        NDCG:             e.calculateNDCG(retrievedChunks, groundTruth),
        ContextRelevance: e.evaluateContextRelevance(query, retrievedChunks),
        
        // NEW: Per-chunk metrics
        ChunkScores: e.scoreEachChunk(retrievedChunks, groundTruth),
    }, nil
}
```

#### 3. VDB Comparison Evaluation

**Compare same content across different VDBs:**

```bash
# Compare VDB performance on same queries
weave eval compare-vdbs \
  --vdbs "weaviate-cloud,qdrant-cloud,milvus-cloud" \
  --collection geography-facts \
  --dataset evals/datasets/baseline.yaml
```

**Output:**
```
VDB Comparison: geography-facts collection
Dataset: baseline.yaml (25 queries)

Retrieval Quality:
                  Precision@5  Recall@5  MRR    NDCG   Avg Latency  Cost
Weaviate Cloud    0.92        0.85      0.88   0.91   45ms         $0.02
Qdrant Cloud      0.89        0.82      0.85   0.88   38ms         $0.01
Milvus Cloud      0.90        0.84      0.87   0.89   52ms         $0.03

Agent Response Quality (with retrieved chunks):
                  Accuracy  Hallucination  Citation  Overall
Weaviate Cloud    0.87     0.92           0.85      0.88
Qdrant Cloud      0.84     0.89           0.82      0.85
Milvus Cloud      0.86     0.91           0.84      0.87

Recommendation: Weaviate Cloud provides best overall quality
  Alternative: Qdrant Cloud for cost-sensitive deployments (20% faster, 50% cheaper)
```

#### 4. Pipeline Configuration Tracking

**Full pipeline version tracking:**

```yaml
# NEW: Pipeline configuration manifest
pipeline_config:
  version: "1.2.0"
  name: "medical-qa-pipeline"
  
  # Document processing
  ingestion:
    chunking:
      strategy: "semantic"  # or: fixed, sentence, paragraph
      chunk_size: 512
      chunk_overlap: 50
      splitter: "langchain-recursive"
      
    preprocessing:
      - type: "strip-headers"
      - type: "normalize-whitespace"
      - type: "remove-footnotes"
  
  # Vectorization
  embedding:
    model: "text-embedding-3-small"
    dimension: 1536
    provider: "openai"
    batch_size: 100
    
  # Vector storage
  vector_db:
    type: "weaviate-cloud"
    collection: "medical-docs"
    distance_metric: "cosine"
    index_type: "hnsw"
    ef_construction: 128
    max_connections: 64
    
  # Retrieval
  search:
    top_k: 5
    distance_threshold: 0.3
    reranking: true
    reranker_model: "cross-encoder/ms-marco-MiniLM-L-12-v2"
    
  # Generation
  agent:
    config: "configs/agents/rag-agent.yaml"
    version: "2.0.0"
```

**Track pipeline config in evaluation:**
```go
type EvalRun struct {
    // Existing fields...
    AgentName    string
    AgentVersion string
    
    // NEW: Pipeline context
    PipelineConfig  PipelineConfig
    CollectionName  string
    VectorDBType    string
    ChunkingConfig  ChunkingConfig
    EmbeddingModel  string
    
    // NEW: Retrieval metrics
    RetrievalMetrics *RetrievalMetrics
    ChunkingMetrics  *ChunkingMetrics
    
    // NEW: Full trace
    RetrievedChunks  []Chunk
    SearchLatency    time.Duration
}
```

### Extended CLI Commands

```bash
# Evaluate full pipeline
weave eval pipeline run \
  --config pipelines/medical-qa.yaml \
  --dataset evals/datasets/medical-baseline.yaml

# Compare chunking strategies
weave eval compare-chunking \
  --strategies "semantic,fixed-512,fixed-1024,paragraph" \
  --collection medical-docs \
  --dataset evals/datasets/baseline.yaml

# Compare embedding models
weave eval compare-embeddings \
  --models "text-embedding-3-small,text-embedding-3-large,all-MiniLM-L6-v2" \
  --collection medical-docs \
  --dataset evals/datasets/baseline.yaml

# Compare VDBs (same content, different VDBs)
weave eval compare-vdbs \
  --vdbs "weaviate-cloud,qdrant-cloud,milvus-cloud" \
  --collection medical-docs \
  --dataset evals/datasets/baseline.yaml

# Full pipeline comparison
weave eval compare-pipelines \
  --pipeline-a pipelines/v1.yaml \
  --pipeline-b pipelines/v2.yaml \
  --dataset evals/datasets/baseline.yaml

# Ingestion quality check
weave eval check-ingestion \
  --collection medical-docs \
  --sample-docs 100 \
  --check-chunking \
  --check-embeddings \
  --check-schema
```

### Extended Output Example

```
Pipeline Evaluation Run: eval-pipeline-20260122-150000
Pipeline: medical-qa-pipeline v1.2.0
Dataset: medical-baseline.yaml (50 test cases)
Duration: 4m 12s

┌─────────────────────────────────────────────────────────┐
│              PIPELINE CONFIGURATION                      │
├─────────────────────────────────────────────────────────┤
│ Chunking:      semantic, 512 tokens, 50 overlap        │
│ Embedding:     text-embedding-3-small (1536d)          │
│ Vector DB:     Weaviate Cloud (cosine, HNSW)           │
│ Collection:    medical-docs                             │
│ Agent:         rag-agent v2.0.0                         │
└─────────────────────────────────────────────────────────┘

INGESTION QUALITY:
  ✓ Chunk Coherence:         0.89  (good semantic boundaries)
  ✓ Information Density:     0.85  (optimal chunk size)
  ⚠ Overlap Effectiveness:   0.72  (consider increasing to 75)

RETRIEVAL QUALITY:
  ✓ Precision@5:             0.88  ⬆ +0.05 vs v1.1
  ✓ Recall@5:                0.82  ⬆ +0.03 vs v1.1
  ✓ MRR:                     0.86  ⬆ +0.04 vs v1.1
  ✓ NDCG:                    0.87  ⬆ +0.02 vs v1.1
  ✓ Context Relevance:       0.90  ⬆ +0.06 vs v1.1
  ✓ Search Latency:          42ms  ⬇ -8ms vs v1.1

AGENT RESPONSE QUALITY:
  ✓ Accuracy:                0.87  ⬆ +0.12 vs v1.1
  ✓ Hallucination:           0.92  ⬆ +0.08 vs v1.1
  ✓ Relevance:               0.89  ⬆ +0.05 vs v1.1
  ✗ Citation:                0.73  ⬇ -0.02 vs v1.1
  ✓ Coherence:               0.85  ⬆ +0.10 vs v1.1
  ✓ Completeness:            0.81  ⬆ +0.03 vs v1.1

COST & PERFORMANCE:
  Embedding Cost:            $0.12
  Agent LLM Cost:            $0.48
  Total Cost:                $0.60  ⬆ +$0.15 vs v1.1
  
  Search Latency:            42ms
  Generation Latency:        1.2s
  Total Latency:             1.24s  ⬇ -0.3s vs v1.1

Pass Rate: 88% (44/50 test cases)

Failed Cases Analysis:
  3 failures: Low retrieval precision (wrong chunks retrieved)
  2 failures: Citation issues (correct facts, missing citations)
  1 failure: Hallucination (added medical disclaimer not in sources)

Improvement Recommendations:
  1. ✓ Retrieval improved significantly with semantic chunking
  2. ⚠ Increase chunk overlap to 75 tokens for better context
  3. ⚠ Review citation prompt - facts are correct but citations missing
  4. ⚠ Consider reranking model to improve precision further
  5. ✓ Overall: Ready for production deployment

Component Breakdown:
┌──────────────┬─────────┬──────────────┬─────────┐
│ Component    │ Quality │ vs Baseline  │ Impact  │
├──────────────┼─────────┼──────────────┼─────────┤
│ Chunking     │ 0.89    │ +0.12        │ High    │
│ Retrieval    │ 0.88    │ +0.05        │ High    │
│ Agent Gen    │ 0.85    │ +0.08        │ Medium  │
└──────────────┴─────────┴──────────────┴─────────┘

View detailed trace: weave eval show eval-pipeline-20260122-150000
Compare with baselines: weave eval compare --run-id eval-pipeline-20260122-150000
```

### Updated Implementation Plan

#### Phase 1: Foundation (Week 1-2)
*Same as before*

#### Phase 2: Opik Integration (Week 3)
*Same as before*

#### Phase 3: Custom Evaluators (Week 4)
*Same as before*

#### Phase 4: Production Features (Week 5-6)
*Same as before*

#### **Phase 5: RAG Pipeline Evaluation (Week 7-8)** ⭐ NEW

**Goal:** Full pipeline evaluation and comparison

- [ ] Chunking evaluators
  - Coherence evaluator
  - Boundary quality checker
  - Information density calculator
  - Chunk size optimizer

- [ ] Retrieval metrics
  - Precision@K, Recall@K
  - MRR, NDCG calculators
  - Context relevance evaluator
  - Latency tracking

- [ ] VDB comparison
  - Multi-VDB test harness
  - Same-content migration tool
  - Performance comparison reports

- [ ] Pipeline versioning
  - Pipeline config schema
  - Version control integration
  - Config diffing

- [ ] CLI commands
  - `weave eval pipeline`
  - `weave eval compare-chunking`
  - `weave eval compare-vdbs`
  - `weave eval compare-embeddings`

**Deliverable:** Full RAG pipeline evaluation and optimization

#### **Phase 6: Advanced Pipeline Features (Week 9-10)** ⭐ NEW

**Goal:** Advanced optimization and experimentation

- [ ] Automatic optimization
  - Chunking parameter tuning
  - Embedding model selection
  - VDB configuration optimization

- [ ] A/B testing framework
  - Pipeline variant testing
  - Traffic splitting
  - Statistical significance testing

- [ ] Ingestion quality checks
  - Pre-ingestion validation
  - Post-ingestion verification
  - Schema compliance checking

- [ ] Integration features
  - CI/CD pipeline integration
  - Automated regression testing
  - Performance benchmarking

**Deliverable:** Production-grade pipeline optimization system

### Extended Success Metrics

#### Pipeline Metrics (NEW)
- Identify optimal chunking strategy for collection type
- 20% improvement in retrieval precision through optimization
- <50ms search latency across all VDBs
- Cost reduction through VDB selection

#### User Success (Extended)
- Compare chunking strategies quantitatively
- Choose optimal VDB for their use case
- Optimize entire pipeline, not just agent
- Trace quality issues to specific pipeline components

### Risk Mitigation (Extended)

**Risk:** VDB comparison requires data migration
- **Mitigation:** Shared test collections across VDBs
- **Mitigation:** Synthetic data generation for testing
- **Mitigation:** Sample-based comparison (not full migration)

**Risk:** Retrieval metrics require ground truth labels
- **Mitigation:** LLM-assisted labeling for test datasets
- **Mitigation:** User-provided relevance judgments
- **Mitigation:** Automatic relevance estimation

**Risk:** Pipeline evaluation is complex
- **Mitigation:** Wizards for pipeline configuration
- **Mitigation:** Sensible defaults for all components
- **Mitigation:** Clear documentation and examples

### Future Enhancements (Extended)

- **Hybrid search evaluation** (vector + keyword)
- **Multi-stage retrieval** (retrieval → reranking → filtering)
- **Cross-lingual evaluation** (different embedding languages)
- **Multi-modal RAG** (text + images + structured data)
- **Adaptive chunking** (dynamic chunk size based on content)
- **Query rewriting** evaluation (query expansion, reformulation)

---

## Summary of Amendments

This amendment extends the agent evaluation system to cover **the full RAG pipeline**:

1. ✅ **Chunking Evaluation** - Quality, coherence, boundaries
2. ✅ **Retrieval Metrics** - Precision, Recall, MRR, NDCG
3. ✅ **VDB Comparison** - Same queries across different VDBs
4. ✅ **Pipeline Versioning** - Track full pipeline configuration
5. ✅ **Component Tracing** - Identify which component causes issues

**Key Benefit:** Users can now optimize the **entire RAG system**, not just the final agent prompt.

**Timeline Impact:** Adds 3-4 weeks to implementation (Phases 5-6)

**Total Timeline:** ~10 weeks for complete system
