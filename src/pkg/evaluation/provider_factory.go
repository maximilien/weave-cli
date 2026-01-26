// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// ProviderType represents the type of evaluator provider
type ProviderType string

const (
	// ProviderTypeLocal uses our own LLM-as-judge evaluators
	ProviderTypeLocal ProviderType = "local"

	// ProviderTypeOpik uses Opik's evaluators via OpenTelemetry
	ProviderTypeOpik ProviderType = "opik"
)

// CreateProvider creates an evaluator provider of the specified type
func CreateProvider(ctx context.Context, providerType ProviderType, llmClient llm.Client) (EvaluatorProvider, error) {
	switch providerType {
	case ProviderTypeLocal:
		if llmClient == nil {
			return nil, fmt.Errorf("local provider requires LLM client")
		}
		return NewLocalProvider(llmClient), nil

	case ProviderTypeOpik:
		if llmClient == nil {
			return nil, fmt.Errorf("Opik provider requires LLM client")
		}
		opikConfig := llm.LoadOpikConfig()
		if !opikConfig.Enabled || opikConfig.APIKey == "" {
			return nil, fmt.Errorf("Opik provider requires OPIK_API_KEY to be set in environment")
		}
		return NewOpikProvider(opikConfig, llmClient)

	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// GetAvailableProviders returns a list of available provider types
func GetAvailableProviders(ctx context.Context, llmClient llm.Client) []ProviderType {
	available := []ProviderType{}

	// Check local provider
	if llmClient != nil {
		available = append(available, ProviderTypeLocal)
	}

	// Check if Opik is available
	opikConfig := llm.LoadOpikConfig()
	if opikConfig.Enabled && opikConfig.APIKey != "" {
		available = append(available, ProviderTypeOpik)
	}

	return available
}

// GetDefaultProvider returns the default provider type
func GetDefaultProvider() ProviderType {
	return ProviderTypeLocal
}
