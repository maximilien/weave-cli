// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLimiter_OpenAI(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  string
	}{
		{"text-embedding-3-small", "text-embedding-3-small", "openai"},
		{"text-embedding-3-large", "text-embedding-3-large", "openai"},
		{"text-embedding-ada-002", "text-embedding-ada-002", "openai"},
		{"empty defaults to openai", "", "openai"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewLimiter(tt.model)
			assert.Equal(t, tt.want, limiter.Provider())
			assert.True(t, limiter.Enabled(), "OpenAI should have rate limiting enabled")
		})
	}
}

func TestNewLimiter_OSS(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  string
	}{
		{"sentence-transformers/all-mpnet-base-v2", "sentence-transformers/all-mpnet-base-v2", "sentence-transformers"},
		{"sentence-transformers/all-MiniLM-L6-v2", "sentence-transformers/all-MiniLM-L6-v2", "sentence-transformers"},
		{"nomic-embed-text", "nomic-embed-text", "ollama"},
		{"all-mpnet-base-v2", "all-mpnet-base-v2", "sentence-transformers"},
		{"all-MiniLM-L6-v2", "all-MiniLM-L6-v2", "sentence-transformers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewLimiter(tt.model)
			assert.Equal(t, tt.want, limiter.Provider())
			assert.False(t, limiter.Enabled(), "OSS models should not have rate limiting")
		})
	}
}

func TestLimiter_Wait_OpenAI(t *testing.T) {
	limiter := NewLimiter("text-embedding-3-small")
	ctx := context.Background()

	// Should allow first request immediately
	start := time.Now()
	err := limiter.Wait(ctx)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 10*time.Millisecond, "First request should be immediate")

	// Should allow burst requests
	for i := 0; i < OpenAIBurst; i++ {
		err := limiter.Wait(ctx)
		require.NoError(t, err)
	}

	// Next request should be rate limited
	start = time.Now()
	err = limiter.Wait(ctx)
	elapsed = time.Since(start)

	require.NoError(t, err)
	assert.Greater(t, elapsed, 10*time.Millisecond, "Should be rate limited after burst")
}

func TestLimiter_Wait_OSS(t *testing.T) {
	limiter := NewLimiter("sentence-transformers/all-mpnet-base-v2")
	ctx := context.Background()

	// Should allow unlimited requests immediately
	for i := 0; i < 100; i++ {
		err := limiter.Wait(ctx)
		require.NoError(t, err)
	}

	// All requests should complete quickly (no cumulative delay)
	start := time.Now()
	for i := 0; i < 10; i++ {
		err := limiter.Wait(ctx)
		require.NoError(t, err)
	}
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 100*time.Millisecond, "OSS should have no rate limiting")
}

func TestLimiter_Allow(t *testing.T) {
	t.Run("OpenAI allows burst then blocks", func(t *testing.T) {
		limiter := NewLimiter("text-embedding-3-small")

		// Should allow burst
		for i := 0; i < OpenAIBurst; i++ {
			assert.True(t, limiter.Allow(), "Should allow burst requests")
		}

		// Should block after burst exhausted
		assert.False(t, limiter.Allow(), "Should block after burst")
	})

	t.Run("OSS always allows", func(t *testing.T) {
		limiter := NewLimiter("sentence-transformers/all-mpnet-base-v2")

		// Should allow unlimited
		for i := 0; i < 100; i++ {
			assert.True(t, limiter.Allow(), "OSS should always allow")
		}
	})
}

func TestNewUnlimitedLimiter(t *testing.T) {
	limiter := NewUnlimitedLimiter()

	assert.False(t, limiter.Enabled())
	assert.Equal(t, "unlimited", limiter.Provider())

	// Should allow unlimited requests
	for i := 0; i < 1000; i++ {
		assert.True(t, limiter.Allow())
	}
}

func TestLimiter_SetRateLimit(t *testing.T) {
	limiter := NewLimiter("text-embedding-3-small")

	t.Run("Set to unlimited", func(t *testing.T) {
		limiter.SetRateLimit(0)
		assert.False(t, limiter.Enabled())

		// Should allow many requests
		for i := 0; i < 100; i++ {
			assert.True(t, limiter.Allow())
		}
	})

	t.Run("Set to custom rate", func(t *testing.T) {
		limiter.SetRateLimit(60) // 60 RPM = 1 RPS
		assert.True(t, limiter.Enabled())

		// Should allow burst
		for i := 0; i < OpenAIBurst; i++ {
			assert.True(t, limiter.Allow())
		}

		// Should block after burst
		assert.False(t, limiter.Allow())
	})
}

func TestLimiter_Metrics(t *testing.T) {
	limiter := NewLimiter("text-embedding-3-small")

	t.Run("Tokens", func(t *testing.T) {
		tokens := limiter.Tokens()
		assert.Greater(t, tokens, 0.0)
	})

	t.Run("Burst", func(t *testing.T) {
		burst := limiter.Burst()
		assert.Equal(t, OpenAIBurst, burst)
	})

	t.Run("Limit", func(t *testing.T) {
		limit := limiter.Limit()
		expectedRPS := float64(OpenAIRPM) / 60.0
		assert.InDelta(t, expectedRPS, float64(limit), 0.1)
	})
}

func TestLimiter_ContextCancellation(t *testing.T) {
	limiter := NewLimiter("text-embedding-3-small")

	// Exhaust burst
	for i := 0; i < OpenAIBurst+5; i++ {
		limiter.Allow()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Wait should respect context cancellation
	err := limiter.Wait(ctx)
	assert.Error(t, err)
	// Error message from rate.Limiter includes "would exceed context deadline"
	assert.Contains(t, err.Error(), "context deadline")
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		model    string
		expected string
	}{
		{"text-embedding-3-small", "openai"},
		{"text-embedding-3-large", "openai"},
		{"text-embedding-ada-002", "openai"},
		{"sentence-transformers/all-mpnet-base-v2", "sentence-transformers"},
		{"sentence-transformers/all-MiniLM-L6-v2", "sentence-transformers"},
		{"all-mpnet-base-v2", "sentence-transformers"},
		{"all-MiniLM-L6-v2", "sentence-transformers"},
		{"nomic-embed-text", "ollama"},
		{"unknown-model", "openai"}, // Conservative default
		{"", "openai"},              // Empty defaults to OpenAI
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := detectProvider(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}
