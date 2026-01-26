// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
)

// EvaluatorProvider creates evaluators for different evaluation types.
// This interface allows pluggable evaluator backends (local, Opik, LangSmith, etc.)
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
