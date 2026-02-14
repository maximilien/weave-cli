// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package ratelimit

import (
	"context"

	"golang.org/x/time/rate"
)

// Limiter wraps rate.Limiter with provider-specific limits
type Limiter struct {
	limiter  *rate.Limiter
	enabled  bool
	provider string
}

// OpenAI rate limits (per minute)
const (
	OpenAIRPM     = 3500 // Requests per minute
	OpenAIBurst   = 10   // Burst allowance
	SentinelRPM   = 0    // Unlimited (OSS models)
	OllamaRPM     = 0    // Unlimited (local)
	HuggingFaceRP = 0    // Unlimited (local sentence-transformers)
)

// NewLimiter creates a rate limiter for the given embedding provider
func NewLimiter(embeddingModel string) *Limiter {
	// Determine provider from model name
	provider := detectProvider(embeddingModel)

	var limiter *rate.Limiter
	enabled := false

	switch provider {
	case "openai":
		// OpenAI: 3,500 RPM = ~58 RPS
		limiter = rate.NewLimiter(rate.Limit(OpenAIRPM/60.0), OpenAIBurst)
		enabled = true
	case "sentence-transformers", "ollama", "huggingface":
		// OSS models: No rate limit
		limiter = rate.NewLimiter(rate.Inf, 0)
		enabled = false
	default:
		// Unknown provider: Conservative limit
		limiter = rate.NewLimiter(rate.Limit(OpenAIRPM/60.0), OpenAIBurst)
		enabled = true
	}

	return &Limiter{
		limiter:  limiter,
		enabled:  enabled,
		provider: provider,
	}
}

// Wait blocks until the rate limiter allows an action
func (l *Limiter) Wait(ctx context.Context) error {
	if !l.enabled {
		return nil // No rate limiting for OSS models
	}
	return l.limiter.Wait(ctx)
}

// Allow checks if an action can proceed without blocking
func (l *Limiter) Allow() bool {
	if !l.enabled {
		return true
	}
	return l.limiter.Allow()
}

// Enabled returns true if rate limiting is active
func (l *Limiter) Enabled() bool {
	return l.enabled
}

// Provider returns the detected provider name
func (l *Limiter) Provider() string {
	return l.provider
}

// detectProvider determines the provider from the embedding model name
func detectProvider(embeddingModel string) string {
	if embeddingModel == "" {
		// Default to OpenAI if no model specified
		return "openai"
	}

	// Check for "sentence-transformers/" prefix (full path)
	if len(embeddingModel) >= 22 {
		if embeddingModel[:22] == "sentence-transformers/" {
			return "sentence-transformers"
		}
	}

	// Check for other known patterns
	if len(embeddingModel) >= 15 {
		prefix := embeddingModel[:15]
		if prefix == "text-embedding-" {
			return "openai"
		}
		if prefix == "all-MiniLM-L6-v" {
			return "sentence-transformers"
		}
	}

	if len(embeddingModel) >= 11 {
		prefix := embeddingModel[:11]
		if prefix == "all-mpnet-b" {
			return "sentence-transformers"
		}
	}

	if len(embeddingModel) >= 6 {
		prefix := embeddingModel[:6]
		if prefix == "nomic-" {
			return "ollama"
		}
	}

	// Conservative: Assume OpenAI-like rate limits
	return "openai"
}

// NewUnlimitedLimiter creates a limiter with no rate restrictions
func NewUnlimitedLimiter() *Limiter {
	return &Limiter{
		limiter:  rate.NewLimiter(rate.Inf, 0),
		enabled:  false,
		provider: "unlimited",
	}
}

// SetRateLimit updates the rate limit dynamically
func (l *Limiter) SetRateLimit(requestsPerMinute int) {
	if requestsPerMinute <= 0 {
		l.limiter = rate.NewLimiter(rate.Inf, 0)
		l.enabled = false
	} else {
		l.limiter = rate.NewLimiter(rate.Limit(requestsPerMinute/60.0), OpenAIBurst)
		l.enabled = true
	}
}

// Tokens returns the current number of available tokens
func (l *Limiter) Tokens() float64 {
	return l.limiter.Tokens()
}

// Burst returns the maximum burst size
func (l *Limiter) Burst() int {
	return l.limiter.Burst()
}

// Limit returns the current rate limit (requests per second)
func (l *Limiter) Limit() rate.Limit {
	return l.limiter.Limit()
}
