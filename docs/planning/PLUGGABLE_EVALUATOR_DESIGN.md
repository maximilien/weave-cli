# Pluggable Evaluator Architecture

**Date:** 2026-01-26
**Status:** Design
**Priority:** High

## Overview

Design a pluggable evaluator system that:
1. ✅ Keeps our existing LLM-as-judge evaluators as default
2. ✅ Adds Opik as optional provider via `--use-opik` flag
3. ✅ Allows future providers (LangSmith, custom, etc.)
4. ✅ No breaking changes to existing code

## Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                     Evaluation Runner                          │
├────────────────────────────────────────────────────────────────┤
│                                                                  │
│  func EvaluateTestCase(ctx, testCase, answer, provider)        │
│    • Runs rule-based evaluators (always)                       │
│    • Asks provider for LLM-based evaluations                   │
│    • Aggregates scores                                         │
│                                                                  │
└────────────────────────────┬───────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────┐
│                  EvaluatorProvider Interface                    │
├────────────────────────────────────────────────────────────────┤
│                                                                  │
│  type EvaluatorProvider interface {                            │
│      Name() string                                             │
│      GetAccuracyEvaluator() Evaluator                          │
│      GetFaithfulnessEvaluator() Evaluator                      │
│      GetHallucinationEvaluator() Evaluator                     │
│      GetContextRelevanceEvaluator() Evaluator                  │
│  }                                                              │
│                                                                  │
└────────────┬───────────────────────┬──────────────────────────┘
             │                       │
             ▼                       ▼
┌──────────────────────┐   ┌──────────────────────────────────┐
│  LocalProvider       │   │  OpikProvider                     │
│  (Default)           │   │  (Optional: --use-opik)          │
├──────────────────────┤   ├──────────────────────────────────┤
│ • Uses our LLM       │   │ • Uses Opik API                  │
│ • OpenAI/Claude      │   │ • Sends via OpenTelemetry        │
│ • Local scoring      │   │ • Gets scores from Opik backend  │
│ • Fast, direct       │   │ • Cost tracking automatic        │
│ • Works offline      │   │ • Dashboard visualization        │
└──────────────────────┘   └──────────────────────────────────┘

┌────────────────────────────────────────────────────────────────┐
│              Rule-Based Evaluators (Always Local)               │
├────────────────────────────────────────────────────────────────┤
│ • CitationEvaluator - regex-based, fast, no LLM               │
│ • ConceptChecker - simple string matching                      │
│ • FormatValidator - checks structure                           │
└────────────────────────────────────────────────────────────────┘
```

## Implementation

### 1. EvaluatorProvider Interface

```go
// src/pkg/evaluation/provider.go
package evaluation

import (
	"context"
)

// EvaluatorProvider creates evaluators for different evaluation types
type EvaluatorProvider interface {
	// Name returns the provider name (e.g., "local", "opik")
	Name() string

	// GetAccuracyEvaluator returns an evaluator for semantic accuracy
	GetAccuracyEvaluator() Evaluator

	// GetFaithfulnessEvaluator returns an evaluator for faithfulness/groundedness
	GetFaithfulnessEvaluator() Evaluator

	// GetHallucinationEvaluator returns an evaluator for hallucination detection
	GetHallucinationEvaluator() Evaluator

	// GetContextRelevanceEvaluator returns an evaluator for context relevance
	GetContextRelevanceEvaluator() Evaluator

	// IsAvailable checks if the provider is properly configured
	IsAvailable(ctx context.Context) bool
}

// Evaluator interface (already exists)
type Evaluator interface {
	Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error)
	Name() string
}
```

### 2. Local Provider (Our Current Implementation)

```go
// src/pkg/evaluation/provider_local.go
package evaluation

import (
	"context"
	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// LocalProvider uses our own LLM-as-judge evaluators
type LocalProvider struct {
	llmClient llm.Client
}

// NewLocalProvider creates a local evaluator provider
func NewLocalProvider(llmClient llm.Client) *LocalProvider {
	return &LocalProvider{
		llmClient: llmClient,
	}
}

func (p *LocalProvider) Name() string {
	return "local"
}

func (p *LocalProvider) GetAccuracyEvaluator() Evaluator {
	return NewAccuracyEvaluator(p.llmClient)
}

func (p *LocalProvider) GetFaithfulnessEvaluator() Evaluator {
	return NewFaithfulnessEvaluator(p.llmClient)
}

func (p *LocalProvider) GetHallucinationEvaluator() Evaluator {
	return NewHallucinationDetector(p.llmClient)
}

func (p *LocalProvider) GetContextRelevanceEvaluator() Evaluator {
	return NewContextRelevanceEvaluator(p.llmClient)
}

func (p *LocalProvider) IsAvailable(ctx context.Context) bool {
	// Local provider is always available if we have an LLM client
	return p.llmClient != nil
}
```

### 3. Opik Provider (New Implementation)

```go
// src/pkg/evaluation/provider_opik.go
package evaluation

import (
	"context"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// OpikProvider uses Opik's LLM-as-judge evaluators via OpenTelemetry
type OpikProvider struct {
	config         *llm.OpikConfig
	tracerProvider *sdktrace.TracerProvider
}

// NewOpikProvider creates an Opik evaluator provider
func NewOpikProvider(config *llm.OpikConfig) (*OpikProvider, error) {
	if !config.Enabled || config.APIKey == "" {
		return nil, fmt.Errorf("Opik is not configured (set OPIK_API_KEY)")
	}

	// Initialize OpenTelemetry tracing
	tp, err := llm.InitOpikTracing(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Opik tracing: %w", err)
	}

	return &OpikProvider{
		config:         config,
		tracerProvider: tp,
	}, nil
}

func (p *OpikProvider) Name() string {
	return "opik"
}

func (p *OpikProvider) GetAccuracyEvaluator() Evaluator {
	return &OpikAccuracyEvaluator{provider: p}
}

func (p *OpikProvider) GetFaithfulnessEvaluator() Evaluator {
	return &OpikFaithfulnessEvaluator{provider: p}
}

func (p *OpikProvider) GetHallucinationEvaluator() Evaluator {
	return &OpikHallucinationEvaluator{provider: p}
}

func (p *OpikProvider) GetContextRelevanceEvaluator() Evaluator {
	return &OpikContextRelevanceEvaluator{provider: p}
}

func (p *OpikProvider) IsAvailable(ctx context.Context) bool {
	return p.config != nil && p.config.Enabled && p.config.APIKey != ""
}

// Shutdown gracefully shuts down the Opik provider
func (p *OpikProvider) Shutdown(ctx context.Context) error {
	return llm.ShutdownOpikTracing(ctx, p.tracerProvider)
}

// OpikAccuracyEvaluator evaluates accuracy using Opik
type OpikAccuracyEvaluator struct {
	provider *OpikProvider
}

func (e *OpikAccuracyEvaluator) Name() string {
	return "accuracy"
}

func (e *OpikAccuracyEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	tracer := otel.Tracer("weave-cli-eval")
	ctx, span := tracer.Start(ctx, "evaluate-accuracy")
	defer span.End()

	// Add evaluation data to OpenTelemetry span
	span.SetAttributes(
		attribute.String("evaluator", "accuracy"),
		attribute.String("provider", "opik"),
		attribute.String("test_case.id", testCase.ID),
		attribute.String("query", testCase.Query),
		attribute.String("expected_answer", testCase.ExpectedAnswer),
		attribute.String("actual_answer", actual),
	)

	// Opik's backend will:
	// 1. Receive this trace via OpenTelemetry
	// 2. Run their LLM-as-judge evaluator
	// 3. Add evaluation scores to the trace
	// 4. Display in Opik dashboard

	// For now, we need to query Opik's API to get the score
	// TODO: Implement Opik API client to fetch evaluation score
	// This requires understanding Opik's evaluation API

	score := 0.0 // Placeholder - will fetch from Opik API

	span.SetAttributes(attribute.Float64("score", score))
	return score, nil
}

// Similar implementations for OpikFaithfulnessEvaluator, OpikHallucinationEvaluator, OpikContextRelevanceEvaluator
```

### 4. Provider Factory

```go
// src/pkg/evaluation/provider_factory.go
package evaluation

import (
	"context"
	"fmt"
	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// ProviderType represents the type of evaluator provider
type ProviderType string

const (
	ProviderTypeLocal ProviderType = "local"
	ProviderTypeOpik  ProviderType = "opik"
)

// CreateProvider creates an evaluator provider of the specified type
func CreateProvider(ctx context.Context, providerType ProviderType, llmClient llm.Client) (EvaluatorProvider, error) {
	switch providerType {
	case ProviderTypeLocal:
		return NewLocalProvider(llmClient), nil

	case ProviderTypeOpik:
		opikConfig := llm.LoadOpikConfig()
		if !opikConfig.Enabled || opikConfig.APIKey == "" {
			return nil, fmt.Errorf("Opik provider requires OPIK_API_KEY to be set")
		}
		return NewOpikProvider(opikConfig)

	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// GetAvailableProviders returns a list of available provider types
func GetAvailableProviders(ctx context.Context, llmClient llm.Client) []ProviderType {
	available := []ProviderType{ProviderTypeLocal} // Local is always available

	// Check if Opik is available
	opikConfig := llm.LoadOpikConfig()
	if opikConfig.Enabled && opikConfig.APIKey != "" {
		available = append(available, ProviderTypeOpik)
	}

	return available
}
```

### 5. Updated Evaluation Runner

```go
// src/pkg/evaluation/runner.go (updated)
package evaluation

import (
	"context"
	"fmt"
	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// EvaluateTestCase runs all evaluators on a test case using the specified provider
func EvaluateTestCase(
	ctx context.Context,
	testCase *TestCase,
	actualAnswer string,
	actualCitations []string,
	provider EvaluatorProvider,
) (*TestCaseResult, error) {
	result := &TestCaseResult{
		TestCaseID:      testCase.ID,
		Query:           testCase.Query,
		ActualAnswer:    actualAnswer,
		ActualCitations: actualCitations,
		Details:         make(map[string]interface{}),
	}

	// Add provider info
	result.Details["evaluator_provider"] = provider.Name()

	var err error

	// Rule-based evaluators (always local, always run)
	citationEval := NewCitationEvaluator()
	result.CitationScore, err = citationEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Citation evaluation failed: %v", err))
	}

	// LLM-based evaluators (use provider)
	accuracyEval := provider.GetAccuracyEvaluator()
	result.AccuracyScore, err = accuracyEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Accuracy evaluation failed: %v", err))
	}

	faithfulnessEval := provider.GetFaithfulnessEvaluator()
	result.FaithfulnessScore, err = faithfulnessEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Faithfulness evaluation failed: %v", err))
	}

	hallucinationEval := provider.GetHallucinationEvaluator()
	result.HallucinationScore, err = hallucinationEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Hallucination evaluation failed: %v", err))
	}

	contextRelEval := provider.GetContextRelevanceEvaluator()
	result.ContextRelevanceScore, err = contextRelEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Context relevance evaluation failed: %v", err))
	}

	// Determine pass/fail
	result.Passed = true

	if testCase.MinRelevanceScore > 0 && result.AccuracyScore < testCase.MinRelevanceScore {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Accuracy score %.2f below threshold %.2f", result.AccuracyScore, testCase.MinRelevanceScore))
	}

	if testCase.MustCite && result.CitationScore < 0.5 {
		result.Passed = false
		result.Errors = append(result.Errors, "Missing required citations")
	}

	if result.HallucinationScore < 0.6 {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("High hallucination risk: score %.2f", result.HallucinationScore))
	}

	return result, nil
}
```

### 6. CLI Integration

```go
// src/cmd/eval/run.go (updated)
package eval

import (
	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/spf13/cobra"
)

func NewRunCommand() *cobra.Command {
	var agentName string
	var datasetPath string
	var collection string
	var useOpik bool // NEW FLAG

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run an evaluation with a dataset",
		Long: `Run an agent evaluation using a test dataset.

Evaluator Providers:
  By default, uses local LLM-as-judge evaluators.

  --use-opik: Use Opik's evaluators and send results to Opik dashboard
              Requires OPIK_API_KEY to be set in environment

Examples:
  # Run with local evaluators (default)
  weave eval run --agent rag-agent --dataset baseline

  # Run with Opik evaluators
  weave eval run --agent rag-agent --dataset baseline --use-opik

  # Opik provides:
  - Better LLM-as-judge evaluators
  - Cost tracking
  - Rich dashboard visualization
  - Historical trends`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEvaluation(agentName, datasetPath, collection, useOpik)
		},
	}

	cmd.Flags().StringVar(&agentName, "agent", "", "Agent name (required)")
	cmd.Flags().StringVar(&datasetPath, "dataset", "", "Dataset name or path (required)")
	cmd.Flags().StringVar(&collection, "collection", "", "Override collection name")
	cmd.Flags().BoolVar(&useOpik, "use-opik", false, "Use Opik evaluators (requires OPIK_API_KEY)")

	cmd.MarkFlagRequired("agent")
	cmd.MarkFlagRequired("dataset")

	return cmd
}

func runEvaluation(agentName, datasetPath, collection string, useOpik bool) error {
	ctx := context.Background()

	// Create LLM client for local evaluators or agent execution
	llmClient, err := createLLMClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Determine provider type
	providerType := evaluation.ProviderTypeLocal
	if useOpik {
		providerType = evaluation.ProviderTypeOpik
	}

	// Create evaluator provider
	provider, err := evaluation.CreateProvider(ctx, providerType, llmClient)
	if err != nil {
		color.Red("Error: %v\n", err)
		color.Yellow("Falling back to local evaluators...\n")
		provider = evaluation.NewLocalProvider(llmClient)
	}

	color.Cyan("Using evaluator provider: %s\n\n", provider.Name())

	// Load dataset
	dataset, err := loadDataset(datasetPath)
	if err != nil {
		return fmt.Errorf("failed to load dataset: %w", err)
	}

	// Run evaluation
	results := []evaluation.TestCaseResult{}
	for _, testCase := range dataset.TestCases {
		// Execute agent query
		response, err := executeAgentQuery(ctx, agentName, testCase.Query, collection)
		if err != nil {
			color.Red("Failed to execute query for test case %s: %v\n", testCase.ID, err)
			continue
		}

		// Evaluate using provider
		result, err := evaluation.EvaluateTestCase(ctx, testCase, response.Text, response.Citations, provider)
		if err != nil {
			color.Red("Failed to evaluate test case %s: %v\n", testCase.ID, err)
			continue
		}

		results = append(results, *result)
	}

	// Display results
	displayResults(results, provider.Name())

	// Show Opik dashboard link if using Opik
	if useOpik {
		fmt.Printf("\n📊 View detailed results in Opik dashboard:\n")
		fmt.Printf("   https://www.comet.com/%s/%s\n",
			os.Getenv("OPIK_WORKSPACE"),
			os.Getenv("OPIK_PROJECT_NAME"))
	}

	// Cleanup Opik provider if needed
	if opikProvider, ok := provider.(*evaluation.OpikProvider); ok {
		defer opikProvider.Shutdown(ctx)
	}

	return nil
}
```

## CLI Usage

### Default: Local Evaluators

```bash
# Uses our LLM-as-judge evaluators
weave eval run --agent rag-agent --dataset baseline

# Output:
Using evaluator provider: local

Evaluation Run: eval-20260126-143022
Agent: rag-agent-v2.yaml
Dataset: baseline-v1.yaml (25 test cases)

Overall Metrics:
  ✓ Accuracy:       0.87  (local evaluator)
  ✓ Faithfulness:   0.92  (local evaluator)
  ✓ Citation:       0.85  (rule-based)
  ✓ Hallucination:  0.73  (local evaluator)
  ✓ Context Rel:    0.89  (local evaluator)

Pass Rate: 88% (22/25 test cases)
```

### Opik: Opik Evaluators

```bash
# Uses Opik's LLM-as-judge evaluators
export OPIK_API_KEY=your-key
export OPIK_WORKSPACE=your-workspace
export OPIK_PROJECT_NAME=weave-cli-evals

weave eval run --agent rag-agent --dataset baseline --use-opik

# Output:
Using evaluator provider: opik

Evaluation Run: eval-20260126-143022
Agent: rag-agent-v2.yaml
Dataset: baseline-v1.yaml (25 test cases)

Overall Metrics:
  ✓ Accuracy:       0.87  (opik evaluator)
  ✓ Faithfulness:   0.92  (opik evaluator)
  ✓ Citation:       0.85  (rule-based)
  ✓ Hallucination:  0.73  (opik evaluator)
  ✓ Context Rel:    0.89  (opik evaluator)

Pass Rate: 88% (22/25 test cases)
Cost: $0.45 (tracked by Opik)

📊 View detailed results in Opik dashboard:
   https://www.comet.com/your-workspace/weave-cli-evals

   Dashboard includes:
   - Detailed trace of each evaluation
   - Cost breakdown
   - Historical trends
   - Export to CSV/JSON
```

## Benefits

### For Users
- ✅ **Choice** - Use local or Opik evaluators based on needs
- ✅ **No breaking changes** - Existing code works as-is
- ✅ **Gradual adoption** - Try Opik with `--use-opik` flag
- ✅ **Offline development** - Local evaluators work without internet
- ✅ **Production ready** - Opik for production monitoring

### For Development
- ✅ **Clean abstraction** - EvaluatorProvider interface
- ✅ **Easy to extend** - Add new providers (LangSmith, etc.)
- ✅ **Testable** - Mock providers for testing
- ✅ **No duplication** - Each provider has clear purpose

### For Future
- ✅ **Pluggable** - Easy to add new providers
- ✅ **Configurable** - Provider selection via flags or config
- ✅ **Composable** - Mix providers (e.g., Opik for accuracy, local for citation)

## Implementation Plan

### Week 1: Foundation
- [x] Design pluggable architecture ✅ (this document)
- [ ] Implement EvaluatorProvider interface
- [ ] Refactor existing evaluators into LocalProvider
- [ ] Add provider factory

### Week 2: Opik Integration
- [ ] Implement OpikProvider
- [ ] Add OpenTelemetry span creation for evaluations
- [ ] Research Opik API for fetching evaluation scores
- [ ] Test Opik integration end-to-end

### Week 3: CLI & Testing
- [ ] Add `--use-opik` flag to CLI
- [ ] Update help text and examples
- [ ] Add integration tests for both providers
- [ ] Test fallback behavior (Opik unavailable → local)

### Week 4: Documentation
- [ ] Update README with provider options
- [ ] Add Opik setup guide
- [ ] Create comparison guide (when to use each)
- [ ] Add troubleshooting section

## Open Questions

1. **Opik Evaluation API** - How do we fetch evaluation scores from Opik?
   - Option A: Query Opik API after trace is sent
   - Option B: Use Opik SDK for evaluations
   - Option C: Synchronous evaluation via Opik API call

2. **Cost Tracking** - How to show costs for local evaluators?
   - Track token usage in our LLM client
   - Calculate costs based on model pricing
   - Display alongside results

3. **Mixed Providers** - Should we support using different providers for different evaluators?
   - E.g., Opik for accuracy, local for faithfulness
   - Probably overkill for MVP

4. **Provider Configuration** - Should providers be configurable via YAML?
   ```yaml
   # configs/evaluation.yaml
   evaluation:
     provider: local  # or: opik, langsmith

     providers:
       local:
         llm_model: gpt-4o
         temperature: 0.1

       opik:
         api_key: ${OPIK_API_KEY}
         workspace: ${OPIK_WORKSPACE}
   ```

## Next Steps

1. Get user approval on architecture
2. Implement EvaluatorProvider interface
3. Refactor existing evaluators into LocalProvider
4. Research Opik evaluation API
5. Implement OpikProvider
6. Add CLI flag and tests

---

**Status:** Ready for implementation pending user approval
