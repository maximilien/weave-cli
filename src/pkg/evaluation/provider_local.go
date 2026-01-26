// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// LocalProvider uses our own LLM-as-judge evaluators.
// This is the default provider that uses OpenAI/Claude/etc. directly.
type LocalProvider struct {
	llmClient llm.Client
}

// NewLocalProvider creates a local evaluator provider
func NewLocalProvider(llmClient llm.Client) *LocalProvider {
	return &LocalProvider{
		llmClient: llmClient,
	}
}

// Name returns "local"
func (p *LocalProvider) Name() string {
	return "local"
}

// GetAccuracyEvaluator returns our AccuracyEvaluator
func (p *LocalProvider) GetAccuracyEvaluator() Evaluator {
	return NewAccuracyEvaluator(p.llmClient)
}

// GetFaithfulnessEvaluator returns our FaithfulnessEvaluator
func (p *LocalProvider) GetFaithfulnessEvaluator() Evaluator {
	return NewFaithfulnessEvaluator(p.llmClient)
}

// GetHallucinationEvaluator returns our HallucinationDetector
func (p *LocalProvider) GetHallucinationEvaluator() Evaluator {
	return NewHallucinationDetector(p.llmClient)
}

// GetContextRelevanceEvaluator returns our ContextRelevanceEvaluator
func (p *LocalProvider) GetContextRelevanceEvaluator() Evaluator {
	return NewContextRelevanceEvaluator(p.llmClient)
}

// IsAvailable checks if local provider is available
func (p *LocalProvider) IsAvailable(ctx context.Context) bool {
	// Local provider is always available if we have an LLM client
	return p.llmClient != nil
}
