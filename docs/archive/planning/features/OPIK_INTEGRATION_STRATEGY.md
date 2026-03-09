# Opik Integration Strategy - Evaluation System

**Date:** 2026-01-26
**Status:** Proposal
**Priority:** Critical

## Executive Summary

After implementing Phase 1 & 2 of the agent evaluation system, we identified significant duplication with Opik's existing capabilities. This document proposes an integration strategy that leverages Opik's strengths while keeping our unique CLI and dataset management features.

## Problem

We built a standalone evaluation system with:
- Our own LLM-as-judge evaluators
- Local results storage
- CLI-only visualization

**However, Opik already provides:**
- LLM-as-judge evaluators
- Cloud storage with history/trends
- Rich dashboard visualization
- Cost tracking
- Production monitoring

**Result:** We're duplicating work that Opik does better, while missing features like cost tracking and trends.

## Proposed Solution

### Architecture: Hybrid Approach

```
┌──────────────────────────────────────────────────────────────┐
│                    WEAVE CLI (Local Layer)                    │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌─────────────────────┐      ┌──────────────────────┐      │
│  │  Test Datasets      │      │  Dataset Management  │      │
│  │  (YAML files)       │      │  CLI Commands        │      │
│  │                     │      │  • create            │      │
│  │  • Test cases       │      │  • list              │      │
│  │  • Ground truth     │      │  • validate          │      │
│  │  • Version control  │      └──────────────────────┘      │
│  └─────────────────────┘                                     │
│                                                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │     Evaluation Runner (weave eval run)              │    │
│  │  • Loads YAML datasets                               │    │
│  │  • Runs agent queries                                │    │
│  │  • Executes rule-based evaluators (citation)         │    │
│  │  • Creates Opik traces for LLM evaluations           │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                                │
└──────────────────────────────┼────────────────────────────────┘
                               │ OpenTelemetry traces
                               ▼
┌──────────────────────────────────────────────────────────────┐
│                     OPIK (Cloud Layer)                        │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         LLM-as-Judge Evaluators (Opik's)             │   │
│  │  • Accuracy (semantic similarity)                     │   │
│  │  • Faithfulness (groundedness)                        │   │
│  │  • Hallucination detection                            │   │
│  │  • Context relevance                                  │   │
│  │  • Moderation checks                                  │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Storage & Analytics                           │   │
│  │  • Evaluation history                                 │   │
│  │  • Trend analysis                                     │   │
│  │  • Cost tracking                                      │   │
│  │  • Performance metrics                                │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Dashboard Visualization                       │   │
│  │  • Evaluation results                                 │   │
│  │  • Comparison charts                                  │   │
│  │  • Regression alerts                                  │   │
│  │  • Export capabilities                                │   │
│  └──────────────────────────────────────────────────────┘   │
│                              │                                │
└──────────────────────────────┼────────────────────────────────┘
                               │ Fetch results
                               ▼
┌──────────────────────────────────────────────────────────────┐
│              WEAVE CLI (Display Layer)                        │
├──────────────────────────────────────────────────────────────┤
│  • Fetch results from Opik API                                │
│  • Display in CLI format (weave eval show)                    │
│  • Generate comparison reports                                │
│  • Support offline mode (local cache)                         │
└──────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### WEAVE CLI (Keep/Refactor)
- ✅ **Test dataset management** - YAML format, version control friendly
- ✅ **CLI commands** - Developer-friendly workflow
- ✅ **Rule-based evaluators** - Citation checking, simple validations
- ✅ **Agent execution** - Run queries through weave-cli agents
- ✅ **Opik trace creation** - Send evaluation data to Opik
- ✅ **Results display** - Fetch from Opik, display in CLI

#### OPIK (Use Existing)
- ✅ **LLM-as-judge evaluators** - Accuracy, Faithfulness, Hallucination, Context Relevance
- ✅ **Evaluation storage** - Cloud-based with history
- ✅ **Dashboard UI** - Rich visualization and exploration
- ✅ **Cost tracking** - Automatic token/cost calculation
- ✅ **Trend analysis** - Performance over time
- ✅ **Production monitoring** - Real-time evaluation

## Implementation Plan

### Phase 1: Remove Duplication (Week 1)

**Task 1.1: Deprecate Our LLM-as-Judge Evaluators**
```go
// BEFORE (src/pkg/evaluation/evaluators.go)
type AccuracyEvaluator struct {
    llmClient llm.Client  // Our own LLM calls
}

func (e *AccuracyEvaluator) Evaluate(...) {
    // Our own prompt to LLM
    response, _ := e.llmClient.Complete(ctx, prompt)
    return parseScore(response)
}

// AFTER (use Opik)
type AccuracyEvaluator struct {
    opikClient *opik.Client
}

func (e *AccuracyEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
    // Send to Opik for evaluation
    return e.opikClient.EvaluateAccuracy(ctx, opik.EvalRequest{
        Query:          testCase.Query,
        ExpectedAnswer: testCase.ExpectedAnswer,
        ActualAnswer:   actual,
    })
}
```

**Task 1.2: Add Opik Client**
```go
// src/pkg/evaluation/opik_client.go
package evaluation

import (
    "context"
    "github.com/maximilien/weave-cli/src/pkg/llm"
    "go.opentelemetry.io/otel"
)

type OpikClient struct {
    tracerProvider *sdktrace.TracerProvider
    projectName    string
}

func NewOpikClient(config *llm.OpikConfig) (*OpikClient, error) {
    tp, err := llm.InitOpikTracing(context.Background(), config)
    if err != nil {
        return nil, err
    }

    return &OpikClient{
        tracerProvider: tp,
        projectName:    config.ProjectName,
    }, nil
}

// Evaluate accuracy using Opik's evaluator
func (c *OpikClient) EvaluateAccuracy(ctx context.Context, req EvalRequest) (float64, error) {
    tracer := otel.Tracer("weave-cli-eval")
    ctx, span := tracer.Start(ctx, "evaluate-accuracy")
    defer span.End()

    // Add evaluation data to span
    span.SetAttributes(
        attribute.String("query", req.Query),
        attribute.String("expected", req.ExpectedAnswer),
        attribute.String("actual", req.ActualAnswer),
        attribute.String("evaluator", "accuracy"),
    )

    // Opik's backend will run LLM-as-judge evaluator
    // and return score through their API
    score, err := c.fetchEvaluationScore(ctx, span.SpanContext().TraceID())

    span.SetAttributes(attribute.Float64("score", score))
    return score, err
}
```

**Task 1.3: Keep Rule-Based Evaluators**
```go
// Citation evaluator doesn't need LLM, keep as-is
type CitationEvaluator struct{} // ✅ KEEP

func (e *CitationEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
    // Simple regex-based citation checking
    citations := extractCitations(actual)
    // ... scoring logic
    return score, nil
}
```

### Phase 2: Integrate Opik Tracing (Week 2)

**Task 2.1: Update Evaluation Runner**
```go
// src/pkg/evaluation/runner.go
func EvaluateTestCase(ctx context.Context, testCase *TestCase, actualAnswer string, actualCitations []string, opikClient *OpikClient) (*TestCaseResult, error) {
    tracer := otel.Tracer("weave-cli-eval")
    ctx, span := tracer.Start(ctx, "test-case-evaluation")
    defer span.End()

    span.SetAttributes(
        attribute.String("test_case.id", testCase.ID),
        attribute.String("query", testCase.Query),
    )

    result := &TestCaseResult{
        TestCaseID:      testCase.ID,
        Query:           testCase.Query,
        ActualAnswer:    actualAnswer,
        ActualCitations: actualCitations,
    }

    // Rule-based evaluators (local, fast)
    citationEval := NewCitationEvaluator()
    result.CitationScore, _ = citationEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)

    // LLM-based evaluators (via Opik)
    result.AccuracyScore, _ = opikClient.EvaluateAccuracy(ctx, EvalRequest{
        Query:          testCase.Query,
        ExpectedAnswer: testCase.ExpectedAnswer,
        ActualAnswer:   actualAnswer,
    })

    result.FaithfulnessScore, _ = opikClient.EvaluateFaithfulness(ctx, EvalRequest{
        Query:          testCase.Query,
        Context:        testCase.RetrievedContext,
        ActualAnswer:   actualAnswer,
    })

    result.HallucinationScore, _ = opikClient.EvaluateHallucination(ctx, EvalRequest{
        ExpectedAnswer: testCase.ExpectedAnswer,
        ActualAnswer:   actualAnswer,
        RequiredConcepts: testCase.RequiredConcepts,
    })

    result.ContextRelevanceScore, _ = opikClient.EvaluateContextRelevance(ctx, EvalRequest{
        Query:   testCase.Query,
        Context: testCase.RetrievedContext,
    })

    return result, nil
}
```

**Task 2.2: Add Config for Opik**
```yaml
# configs/evaluation.yaml
evaluation:
  # Opik integration
  opik:
    enabled: true
    api_key: ${OPIK_API_KEY}
    workspace: ${OPIK_WORKSPACE}
    project_name: "weave-cli-evals"
    endpoint: "https://www.comet.com/opik/api"

  # Local fallback (when Opik unavailable)
  fallback:
    enabled: true
    storage_dir: "evals/results"

  # Keep local dataset management
  datasets_dir: "evals/datasets"
```

### Phase 3: CLI Integration (Week 3)

**Task 3.1: Update CLI Commands**
```go
// src/cmd/eval/run.go
func runEvaluation(agentName, datasetPath string) {
    // Load Opik config
    opikConfig := loadOpikConfig()
    opikClient, err := evaluation.NewOpikClient(opikConfig)
    if err != nil {
        log.Warn("Opik unavailable, using local fallback")
        opikClient = nil
    }

    // Load dataset (from local YAML)
    dataset := loadDataset(datasetPath)

    // Run evaluation
    for _, testCase := range dataset.TestCases {
        // Execute agent query
        response := executeAgentQuery(agentName, testCase.Query)

        // Evaluate (sends to Opik if available)
        result, _ := evaluation.EvaluateTestCase(ctx, testCase, response.Text, response.Citations, opikClient)

        results = append(results, result)
    }

    // Display results (fetch from Opik or local)
    displayResults(results)

    if opikClient != nil {
        fmt.Printf("\n📊 View detailed results in Opik: %s\n", opikClient.GetDashboardURL())
    }
}
```

**Task 3.2: Add Opik URL Display**
```
Evaluation Run: eval-20260126-143022
Agent: rag-agent-v2.yaml
Dataset: baseline-v1.yaml (25 test cases)

Overall Metrics:
  ✓ Accuracy:       0.87
  ✓ Faithfulness:   0.92
  ✓ Citation:       0.85  (local evaluator)
  ✓ Hallucination:  0.73
  ✓ Context Rel:    0.89

Pass Rate: 88% (22/25 test cases)
Cost: $0.45 (tracked by Opik)

📊 View full results in Opik:
   https://www.comet.com/workspace/weave-cli-evals/eval-20260126-143022

   Dashboard includes:
   - Detailed trace of each evaluation
   - Cost breakdown
   - Historical trends
   - Export to CSV/JSON
```

### Phase 4: Documentation (Week 4)

**Task 4.1: Update README**
- Document Opik integration
- Show how to set up Opik API key
- Explain hybrid architecture

**Task 4.2: Migration Guide**
- How to migrate from standalone to Opik-integrated
- What changes for users
- How to use offline mode

## Benefits of Integration

### What We Gain
1. ✅ **Better LLM-as-judge** - Opik's evaluators are battle-tested
2. ✅ **Cost tracking** - Automatic, accurate
3. ✅ **Rich dashboards** - Better than CLI tables
4. ✅ **Historical trends** - Track performance over time
5. ✅ **Production monitoring** - Use same evaluators in prod
6. ✅ **Reduced maintenance** - Don't maintain our own LLM evaluators

### What We Keep
1. ✅ **YAML datasets** - Version control friendly
2. ✅ **CLI workflow** - Fast, scriptable, CI/CD friendly
3. ✅ **Local development** - Works offline with fallback
4. ✅ **Rule-based evaluators** - Fast citation checking
5. ✅ **Dataset management** - Creating/editing test suites

### What We Remove
1. ❌ Our LLM-as-judge implementations (~200 lines)
2. ❌ Local results storage complexity
3. ❌ Manual score parsing from LLM responses
4. ❌ No cost tracking
5. ❌ CLI-only visualization

## Migration Strategy

### For Users

**Before (Standalone):**
```bash
# Everything local
weave eval run --agent rag-agent --dataset baseline
weave eval show eval-123  # Shows local results
```

**After (Opik-integrated):**
```bash
# Set up Opik (one time)
export OPIK_API_KEY=your-key
export OPIK_WORKSPACE=your-workspace

# Same commands, better results
weave eval run --agent rag-agent --dataset baseline
# ✅ Sends to Opik for evaluation
# ✅ Shows CLI summary + Opik dashboard link
# ✅ Tracks costs automatically
# ✅ Stores results in Opik cloud

weave eval show eval-123  # Fetches from Opik
# ✅ Shows results with cost data
# ✅ Provides dashboard link for deep dive
```

**Offline Mode:**
```bash
# Works without Opik (uses local fallback)
weave eval run --agent rag-agent --dataset baseline --offline
# ⚠️ Limited to rule-based evaluators
# ⚠️ No LLM-as-judge evaluations
# ⚠️ No cost tracking
```

### Backward Compatibility

**Option 1: Gradual Migration (Recommended)**
- Keep old evaluators with deprecation warnings
- Add `--use-opik` flag to opt-in
- After 2 releases, make Opik default
- Remove old evaluators after 4 releases

**Option 2: Immediate Switch**
- Replace evaluators now
- Add fallback for offline use
- Update all documentation
- May break existing workflows

**Recommendation:** Option 1 for smooth transition

## Implementation Timeline

| Week | Tasks | Deliverable |
|------|-------|-------------|
| 1 | Remove duplication, add Opik client | Opik client working |
| 2 | Integrate Opik tracing | Evaluations sent to Opik |
| 3 | Update CLI commands | CLI shows Opik results |
| 4 | Documentation, testing | Migration complete |

**Total:** 4 weeks to full integration

## Success Metrics

- ✅ Opik integration working for all LLM-based evaluators
- ✅ Cost tracking visible in CLI and Opik dashboard
- ✅ Offline fallback working for development
- ✅ All existing tests passing
- ✅ Documentation updated
- ✅ User migration guide complete

## Open Questions

1. **Opik API access** - Do we have Opik API docs for programmatic evaluation?
2. **Offline fallback** - Should we keep simple LLM evaluators for offline use?
3. **Pricing** - What's the cost model for Opik evaluations?
4. **Rate limits** - Any limits on evaluation volume?

## Next Steps

1. **User approval** - Get sign-off on hybrid architecture
2. **Opik API research** - Understand Opik's evaluation API
3. **Prototype** - Build OpikClient with one evaluator
4. **Test** - Verify integration works end-to-end
5. **Implement** - Follow 4-week plan
6. **Document** - Update all docs and examples

---

## Recommendation

**Proceed with Hybrid Approach:**
- Use Opik for LLM-as-judge evaluations (accuracy, faithfulness, hallucination, context relevance)
- Keep YAML datasets and CLI workflow
- Keep rule-based evaluators (citation)
- Add Opik dashboard links in CLI output
- Provide offline fallback for development

**This maximizes value by:**
1. Leveraging Opik's strengths (LLM evaluation, dashboards, cost tracking)
2. Keeping our strengths (CLI workflow, YAML datasets, local development)
3. Avoiding duplication of complex LLM evaluation logic
4. Providing better user experience (dashboards + CLI)
