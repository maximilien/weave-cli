// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mocks

import (
	"context"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/stretchr/testify/mock"
)

// MockLLMClient is a mock implementation of llm.Client for testing
type MockLLMClient struct {
	mock.Mock
}

// Complete mocks the Complete method
func (m *MockLLMClient) Complete(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	args := m.Called(ctx, prompt, opts)
	return args.String(0), args.Error(1)
}

// CompleteStructured mocks the CompleteStructured method
func (m *MockLLMClient) CompleteStructured(ctx context.Context, prompt string, schema interface{}, opts ...llm.Option) (interface{}, error) {
	args := m.Called(ctx, prompt, schema, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

// GetMetrics mocks the GetMetrics method
func (m *MockLLMClient) GetMetrics() *llm.Metrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*llm.Metrics)
}
