// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// OpikProvider uses local LLM-as-judge evaluators and sends traces to Opik dashboard.
// This provides real evaluation scores in CLI output while enabling rich visualization
// in the Opik dashboard for production monitoring and analysis.
type OpikProvider struct {
	config         *llm.OpikConfig
	tracerProvider *sdktrace.TracerProvider
	llmClient      llm.Client
}

// NewOpikProvider creates an Opik evaluator provider
func NewOpikProvider(config *llm.OpikConfig, llmClient llm.Client) (*OpikProvider, error) {
	if !config.Enabled || config.APIKey == "" {
		return nil, fmt.Errorf("Opik is not configured (set OPIK_API_KEY)")
	}

	if llmClient == nil {
		return nil, fmt.Errorf("LLM client is required for Opik provider")
	}

	// Initialize OpenTelemetry tracing
	tp, err := llm.InitOpikTracing(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Opik tracing: %w", err)
	}

	return &OpikProvider{
		config:         config,
		tracerProvider: tp,
		llmClient:      llmClient,
	}, nil
}

// Name returns "opik"
func (p *OpikProvider) Name() string {
	return "opik"
}

// GetAccuracyEvaluator returns an Opik accuracy evaluator
func (p *OpikProvider) GetAccuracyEvaluator() Evaluator {
	return &OpikAccuracyEvaluator{provider: p}
}

// GetFaithfulnessEvaluator returns an Opik faithfulness evaluator
func (p *OpikProvider) GetFaithfulnessEvaluator() Evaluator {
	return &OpikFaithfulnessEvaluator{provider: p}
}

// GetHallucinationEvaluator returns an Opik hallucination evaluator
func (p *OpikProvider) GetHallucinationEvaluator() Evaluator {
	return &OpikHallucinationEvaluator{provider: p}
}

// GetContextRelevanceEvaluator returns an Opik context relevance evaluator
func (p *OpikProvider) GetContextRelevanceEvaluator() Evaluator {
	return &OpikContextRelevanceEvaluator{provider: p}
}

// IsAvailable checks if Opik provider is available
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

// Name returns "accuracy"
func (e *OpikAccuracyEvaluator) Name() string {
	return "accuracy"
}

// Evaluate runs local accuracy evaluation and sends trace to Opik
func (e *OpikAccuracyEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// Start trace span
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

	// Run local accuracy evaluation
	localEval := NewAccuracyEvaluator(e.provider.llmClient)
	score, err := localEval.Evaluate(ctx, testCase, actual, actualCitations)

	// Add score to trace span
	span.SetAttributes(attribute.Float64("score", score))

	// Even if evaluation failed, we still return the score
	// The error will be logged in EvaluateTestCase
	return score, err
}

// OpikFaithfulnessEvaluator evaluates faithfulness using Opik
type OpikFaithfulnessEvaluator struct {
	provider *OpikProvider
}

// Name returns "faithfulness"
func (e *OpikFaithfulnessEvaluator) Name() string {
	return "faithfulness"
}

// Evaluate runs local faithfulness evaluation and sends trace to Opik
func (e *OpikFaithfulnessEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// Start trace span
	tracer := otel.Tracer("weave-cli-eval")
	ctx, span := tracer.Start(ctx, "evaluate-faithfulness")
	defer span.End()

	span.SetAttributes(
		attribute.String("evaluator", "faithfulness"),
		attribute.String("provider", "opik"),
		attribute.String("test_case.id", testCase.ID),
		attribute.String("query", testCase.Query),
		attribute.String("actual_answer", actual),
	)

	// Add context if available
	if len(testCase.RetrievedContext) > 0 {
		for i, chunk := range testCase.RetrievedContext {
			span.SetAttributes(attribute.String(fmt.Sprintf("context.%d", i), chunk))
		}
	}

	// Run local faithfulness evaluation
	localEval := NewFaithfulnessEvaluator(e.provider.llmClient)
	score, err := localEval.Evaluate(ctx, testCase, actual, actualCitations)

	// Add score to trace span
	span.SetAttributes(attribute.Float64("score", score))

	return score, err
}

// OpikHallucinationEvaluator evaluates hallucination using Opik
type OpikHallucinationEvaluator struct {
	provider *OpikProvider
}

// Name returns "hallucination"
func (e *OpikHallucinationEvaluator) Name() string {
	return "hallucination"
}

// Evaluate runs local hallucination evaluation and sends trace to Opik
func (e *OpikHallucinationEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// Start trace span
	tracer := otel.Tracer("weave-cli-eval")
	ctx, span := tracer.Start(ctx, "evaluate-hallucination")
	defer span.End()

	span.SetAttributes(
		attribute.String("evaluator", "hallucination"),
		attribute.String("provider", "opik"),
		attribute.String("test_case.id", testCase.ID),
		attribute.String("query", testCase.Query),
		attribute.String("expected_answer", testCase.ExpectedAnswer),
		attribute.String("actual_answer", actual),
	)

	// Add required concepts if specified
	if len(testCase.RequiredConcepts) > 0 {
		for i, concept := range testCase.RequiredConcepts {
			span.SetAttributes(attribute.String(fmt.Sprintf("required_concept.%d", i), concept))
		}
	}

	// Run local hallucination evaluation
	localEval := NewHallucinationDetector(e.provider.llmClient)
	score, err := localEval.Evaluate(ctx, testCase, actual, actualCitations)

	// Add score to trace span
	span.SetAttributes(attribute.Float64("score", score))

	return score, err
}

// OpikContextRelevanceEvaluator evaluates context relevance using Opik
type OpikContextRelevanceEvaluator struct {
	provider *OpikProvider
}

// Name returns "context_relevance"
func (e *OpikContextRelevanceEvaluator) Name() string {
	return "context_relevance"
}

// Evaluate runs local context relevance evaluation and sends trace to Opik
func (e *OpikContextRelevanceEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// Start trace span
	tracer := otel.Tracer("weave-cli-eval")
	ctx, span := tracer.Start(ctx, "evaluate-context-relevance")
	defer span.End()

	span.SetAttributes(
		attribute.String("evaluator", "context_relevance"),
		attribute.String("provider", "opik"),
		attribute.String("test_case.id", testCase.ID),
		attribute.String("query", testCase.Query),
	)

	// Add retrieved context
	if len(testCase.RetrievedContext) > 0 {
		span.SetAttributes(attribute.Int("context.count", len(testCase.RetrievedContext)))
		for i, chunk := range testCase.RetrievedContext {
			span.SetAttributes(attribute.String(fmt.Sprintf("context.%d", i), chunk))
		}
	}

	// Run local context relevance evaluation
	localEval := NewContextRelevanceEvaluator(e.provider.llmClient)
	score, err := localEval.Evaluate(ctx, testCase, actual, actualCitations)

	// Add score to trace span
	span.SetAttributes(attribute.Float64("score", score))

	return score, err
}
