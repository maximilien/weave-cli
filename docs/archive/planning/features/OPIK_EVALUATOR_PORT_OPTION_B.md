# Option B: Port Opik's Evaluation Logic to Go

**Date:** 2026-01-26
**Status:** 📋 Future Option (Not Implemented)
**Complexity:** High
**Priority:** Low (Option A is preferred)

## Overview

This document describes **Option B** for Opik integration: porting Opik's Python evaluation logic to Go. This option was **not selected** in favor of Option A (using local evaluators with trace sending), but is documented here for future reference.

## Context

After researching Opik's architecture, we discovered:
- Opik evaluators run **client-side** using Python SDK
- There is **no hosted API** for evaluation scores
- Opik uses LiteLLM to call LLM providers (OpenAI, Claude, etc.)
- Results are sent to Opik platform via OpenTelemetry traces

## Option B: What It Involves

### Goal
Port Opik's evaluation prompts and scoring logic from Python to Go, running them locally while sending traces to Opik.

### Implementation Approach

#### 1. Study Opik's Source Code
**Files to Review:**
- `opik/evaluation/metrics/llm_judges/hallucination/metric.py`
- `opik/evaluation/metrics/llm_judges/answer_relevance/metric.py`
- `opik/evaluation/metrics/llm_judges/context_precision/metric.py`
- `opik/evaluation/metrics/llm_judges/moderation/metric.py`

**Extract:**
- Evaluation prompts
- Scoring algorithms
- Threshold logic
- Prompt templates

#### 2. Port Evaluation Logic
**Create Go Implementations:**

```go
// File: src/pkg/evaluation/opik_logic/hallucination.go
package opik_logic

const OpikHallucinationPrompt = `
You are an impartial judge evaluating whether an AI assistant's response contains hallucinations.

Context:
{{.Context}}

Question:
{{.Query}}

Response:
{{.Answer}}

Rate the hallucination risk on a scale of 0.0 (severe hallucination) to 1.0 (no hallucination).
Provide your rating as a single number.
`

type OpikHallucinationEvaluator struct {
    llmClient llm.Client
}

func (e *OpikHallucinationEvaluator) Evaluate(ctx context.Context, query, answer string, context []string) (float64, error) {
    // Format prompt using Opik's template
    prompt := formatTemplate(OpikHallucinationPrompt, map[string]interface{}{
        "Context": strings.Join(context, "\n"),
        "Query":   query,
        "Answer":  answer,
    })

    // Call LLM using Opik's approach
    response, err := e.llmClient.Complete(ctx, prompt)
    if err != nil {
        return 0.0, err
    }

    // Parse score using Opik's parsing logic
    score := parseOpikScore(response)

    return score, nil
}
```

#### 3. Maintain Opik Prompt Compatibility
**Challenge:** Opik's prompts may change over time

**Approach:**
- Version prompt templates
- Add prompt versioning to traces
- Document prompt sources and dates
- Create update process for new Opik versions

```go
const (
    OpikPromptVersion = "v1.2.0"  // Track Opik SDK version
    OpikPromptDate    = "2026-01-26"
)
```

#### 4. Test Against Opik's Results
**Validation Strategy:**
- Run same inputs through both implementations
- Compare scores and ensure close match
- Document acceptable variance threshold
- Create regression test suite

### Files to Create

1. **src/pkg/evaluation/opik_logic/**
   - `hallucination.go` - Hallucination detection
   - `answer_relevance.go` - Answer relevance
   - `context_precision.go` - Context precision
   - `context_recall.go` - Context recall
   - `moderation.go` - Content moderation
   - `prompts.go` - All prompt templates
   - `parser.go` - Score parsing logic

2. **src/pkg/evaluation/opik_logic_test.go**
   - Unit tests for each evaluator
   - Comparison tests vs expected outputs
   - Regression tests

3. **docs/planning/OPIK_PROMPT_SOURCES.md**
   - Document all prompts and their sources
   - Version history
   - Update procedures

## Pros and Cons

### Pros ✅
- **Consistent with Opik's approach** - Uses their exact evaluation logic
- **Potentially better prompts** - Opik's prompts may be more refined
- **Official Opik methodology** - Can cite Opik as evaluation source
- **Community alignment** - Uses same logic as other Opik users

### Cons ❌
- **High maintenance burden** - Need to track Opik's prompt changes
- **Complex implementation** - Requires porting significant logic
- **Dependency on Opik's decisions** - Tied to their prompt design
- **Version synchronization** - Must keep prompts in sync
- **Testing complexity** - Hard to validate against moving target
- **No functional advantage** - Local evaluators already work well

## Why Option A Was Chosen

**Option A (Hybrid Approach):**
- ✅ Uses our existing, tested evaluators
- ✅ Still sends traces to Opik dashboard
- ✅ Low maintenance burden
- ✅ Real scores in CLI output
- ✅ Full Opik dashboard benefits
- ✅ Simple implementation

**Decision:** Option A provides all the benefits with minimal complexity.

## When Option B Might Be Reconsidered

**Scenarios where Option B could make sense:**

1. **Opik Provides API**
   - If Opik releases a hosted evaluation API
   - Then we could call their service directly
   - **Action:** Revisit architecture, implement API client

2. **Community Standard Emerges**
   - If Opik's prompts become industry standard
   - If compliance requires "official" Opik evaluations
   - **Action:** Port prompts for specific use cases

3. **Evaluation Quality Gap**
   - If user feedback indicates Opik's evaluators are significantly better
   - If benchmarks show measurable quality difference
   - **Action:** Compare evaluation quality, selectively port

4. **Certification Requirements**
   - If customers require "Opik-certified" evaluations
   - If regulatory compliance mandates specific methodology
   - **Action:** Implement parallel evaluation paths

## Implementation Plan (If Pursued)

### Phase 1: Research & Extraction (2-3 days)
- [ ] Clone Opik repository
- [ ] Study evaluation metric implementations
- [ ] Extract all prompt templates
- [ ] Document scoring algorithms
- [ ] Identify dependencies and edge cases

### Phase 2: Port Core Logic (3-4 days)
- [ ] Create opik_logic package
- [ ] Port hallucination evaluator
- [ ] Port answer relevance evaluator
- [ ] Port context precision evaluator
- [ ] Port context recall evaluator
- [ ] Port moderation evaluator

### Phase 3: Testing & Validation (2-3 days)
- [ ] Create test dataset
- [ ] Run through both implementations
- [ ] Compare scores statistically
- [ ] Document variance analysis
- [ ] Create regression test suite

### Phase 4: Integration (1-2 days)
- [ ] Update OpikProvider to use ported logic
- [ ] Add configuration flag for logic choice
- [ ] Update CLI to expose choice
- [ ] Document trade-offs

### Phase 5: Maintenance Setup (1 day)
- [ ] Create Opik version tracking system
- [ ] Set up prompt diff alerts
- [ ] Document update procedure
- [ ] Assign maintenance owner

**Total Estimated Time:** 10-14 days

## Cost-Benefit Analysis

### Option A (Hybrid - IMPLEMENTED)
- **Development Time:** 2-3 hours
- **Maintenance:** Low (existing evaluators)
- **Risk:** Low
- **Benefit:** High (real scores + traces)

### Option B (Port Logic - NOT IMPLEMENTED)
- **Development Time:** 10-14 days
- **Maintenance:** Medium-High (tracking updates)
- **Risk:** Medium (sync issues, quality parity)
- **Benefit:** Medium (Opik-exact logic)

**ROI Analysis:** Option A provides 90% of the benefit with 5% of the effort.

## Technical Debt Considerations

If Option B is ever implemented:

1. **Version Lock-in**
   - Need clear versioning strategy
   - Document which Opik version we're compatible with
   - Plan for major version migrations

2. **Prompt Drift**
   - Opik's prompts will evolve
   - Our ports will lag behind
   - Need automated drift detection

3. **Testing Burden**
   - Must test against Opik's outputs
   - Regression suite grows over time
   - Performance implications

4. **Documentation Overhead**
   - Must document all deviations
   - Explain why scores differ (if they do)
   - Support questions about methodology

## Alternative: Hybrid Approach

**Best of Both Worlds:**
- Use our evaluators by default (Option A)
- Optionally use Opik-ported logic with flag
- Let users choose based on needs

```bash
# Default: our evaluators + Opik traces
weave eval run --agent rag --dataset test --use-opik

# Future: Opik logic + Opik traces
weave eval run --agent rag --dataset test --use-opik --opik-native-logic

# Local only: our evaluators, no traces
weave eval run --agent rag --dataset test
```

## Conclusion

**Option B is documented but not recommended** at this time.

**Reasons:**
1. Option A provides equivalent functionality
2. Significantly lower implementation cost
3. Easier to maintain long-term
4. No Opik API exists to justify complexity
5. Community feedback hasn't indicated need

**Future Trigger:**
If Opik releases hosted evaluation API or evaluation quality gap is demonstrated, revisit this option.

---

**Status:** Documented for future consideration, not currently planned for implementation.
