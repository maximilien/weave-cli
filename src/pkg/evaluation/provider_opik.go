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

// OpikProvider uses Opik's LLM-as-judge evaluators via OpenTelemetry.
// Evaluations are sent to Opik's backend and results are displayed in their dashboard.
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

// Evaluate sends accuracy evaluation to Opik via OpenTelemetry
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

	// TODO: Implement Opik API integration
	// Options:
	// 1. Query Opik API synchronously for evaluation score
	// 2. Use Opik SDK for evaluation (if available)
	// 3. Get trace ID and poll for results
	//
	// For now, we return a placeholder score and rely on Opik dashboard
	// to show the full evaluation results.

	score := 0.0 // Placeholder - will fetch from Opik API

	span.SetAttributes(attribute.Float64("score", score))

	return score, fmt.Errorf("Opik evaluator integration not yet implemented - see traces in Opik dashboard")
}

// OpikFaithfulnessEvaluator evaluates faithfulness using Opik
type OpikFaithfulnessEvaluator struct {
	provider *OpikProvider
}

// Name returns "faithfulness"
func (e *OpikFaithfulnessEvaluator) Name() string {
	return "faithfulness"
}

// Evaluate sends faithfulness evaluation to Opik via OpenTelemetry
func (e *OpikFaithfulnessEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
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

	score := 0.0 // Placeholder

	span.SetAttributes(attribute.Float64("score", score))

	return score, fmt.Errorf("Opik evaluator integration not yet implemented - see traces in Opik dashboard")
}

// OpikHallucinationEvaluator evaluates hallucination using Opik
type OpikHallucinationEvaluator struct {
	provider *OpikProvider
}

// Name returns "hallucination"
func (e *OpikHallucinationEvaluator) Name() string {
	return "hallucination"
}

// Evaluate sends hallucination evaluation to Opik via OpenTelemetry
func (e *OpikHallucinationEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
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

	score := 0.0 // Placeholder

	span.SetAttributes(attribute.Float64("score", score))

	return score, fmt.Errorf("Opik evaluator integration not yet implemented - see traces in Opik dashboard")
}

// OpikContextRelevanceEvaluator evaluates context relevance using Opik
type OpikContextRelevanceEvaluator struct {
	provider *OpikProvider
}

// Name returns "context_relevance"
func (e *OpikContextRelevanceEvaluator) Name() string {
	return "context_relevance"
}

// Evaluate sends context relevance evaluation to Opik via OpenTelemetry
func (e *OpikContextRelevanceEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
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

	score := 1.0 // Return 1.0 if no context (not applicable)
	if len(testCase.RetrievedContext) > 0 {
		score = 0.0 // Placeholder for actual evaluation
	}

	span.SetAttributes(attribute.Float64("score", score))

	if len(testCase.RetrievedContext) > 0 {
		return score, fmt.Errorf("Opik evaluator integration not yet implemented - see traces in Opik dashboard")
	}

	return score, nil
}
