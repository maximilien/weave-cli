// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Client is the interface for LLM clients
type Client interface {
	// Complete generates a text completion
	Complete(ctx context.Context, prompt string, opts ...Option) (string, error)

	// CompleteStructured generates structured output based on schema
	CompleteStructured(ctx context.Context, prompt string, schema interface{}, opts ...Option) (interface{}, error)

	// GetMetrics returns accumulated metrics
	GetMetrics() *Metrics
}

// Option is a function that configures completion options
type Option func(*CompletionOptions)

// CompletionOptions holds options for completions
type CompletionOptions struct {
	Model       string
	Temperature float64
	MaxTokens   int
	SystemMsg   string
}

// DefaultCompletionOptions returns default options
func DefaultCompletionOptions() *CompletionOptions {
	return &CompletionOptions{
		Model:       "gpt-4o", // Latest and best OpenAI model
		Temperature: 0.1,
		MaxTokens:   4096,
	}
}

// WithModel sets the model to use
func WithModel(model string) Option {
	return func(o *CompletionOptions) {
		o.Model = model
	}
}

// WithTemperature sets the temperature
func WithTemperature(temp float64) Option {
	return func(o *CompletionOptions) {
		o.Temperature = temp
	}
}

// WithMaxTokens sets max tokens
func WithMaxTokens(tokens int) Option {
	return func(o *CompletionOptions) {
		o.MaxTokens = tokens
	}
}

// WithSystemMessage sets system message
func WithSystemMessage(msg string) Option {
	return func(o *CompletionOptions) {
		o.SystemMsg = msg
	}
}

// Metrics tracks LLM usage metrics
type Metrics struct {
	Invocations      int
	TotalTokens      int
	PromptTokens     int
	CompletionTokens int
	TotalCost        float64
}

// OpenAIClient implements the Client interface using OpenAI
type OpenAIClient struct {
	client  *openai.Client
	metrics *Metrics
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string) (*OpenAIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &OpenAIClient{
		client: &client,
		metrics: &Metrics{
			Invocations:      0,
			TotalTokens:      0,
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalCost:        0.0,
		},
	}, nil
}

// NewOpenAIClientWithHTTP creates a new OpenAI client with custom HTTP client
func NewOpenAIClientWithHTTP(apiKey string, httpClient *http.Client) (*OpenAIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
	)

	return &OpenAIClient{
		client: &client,
		metrics: &Metrics{
			Invocations:      0,
			TotalTokens:      0,
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalCost:        0.0,
		},
	}, nil
}

// Complete generates a text completion
func (c *OpenAIClient) Complete(ctx context.Context, prompt string, opts ...Option) (string, error) {
	options := DefaultCompletionOptions()
	for _, opt := range opts {
		opt(options)
	}

	messages := []openai.ChatCompletionMessageParamUnion{}

	if options.SystemMsg != "" {
		messages = append(messages, openai.SystemMessage(options.SystemMsg))
	}

	messages = append(messages, openai.UserMessage(prompt))

	params := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       openai.ChatModel(options.Model),
		Temperature: openai.Float(options.Temperature),
		MaxTokens:   openai.Int(int64(options.MaxTokens)),
	}

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create completion: %w", err)
	}

	// Update metrics
	c.metrics.Invocations++
	c.metrics.TotalTokens += int(completion.Usage.TotalTokens)
	c.metrics.PromptTokens += int(completion.Usage.PromptTokens)
	c.metrics.CompletionTokens += int(completion.Usage.CompletionTokens)
	c.metrics.TotalCost += calculateCost(options.Model, int(completion.Usage.PromptTokens), int(completion.Usage.CompletionTokens))

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return completion.Choices[0].Message.Content, nil
}

// CompleteStructured generates structured output based on schema
func (c *OpenAIClient) CompleteStructured(ctx context.Context, prompt string, schema interface{}, opts ...Option) (interface{}, error) {
	options := DefaultCompletionOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Add JSON formatting instructions
	systemMsg := "You must respond with valid JSON only. Do not include any text outside the JSON structure. Do not wrap the JSON in markdown code fences."
	if options.SystemMsg != "" {
		systemMsg = options.SystemMsg + "\n\n" + systemMsg
	}

	enhancedPrompt := prompt + "\n\nRespond with valid JSON only, without markdown code fences."

	response, err := c.Complete(ctx, enhancedPrompt, append(opts, WithSystemMessage(systemMsg))...)
	if err != nil {
		return nil, err
	}

	// Clean up response - remove markdown code fences if present
	response = cleanJSONResponse(response)

	// Parse JSON response
	if err := json.Unmarshal([]byte(response), schema); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return schema, nil
}

// cleanJSONResponse removes markdown code fences and extra whitespace
func cleanJSONResponse(response string) string {
	// Remove markdown code fences
	if len(response) > 0 && response[0] == '`' {
		// Find the first { and last }
		start := -1
		end := -1
		for i, ch := range response {
			if ch == '{' && start == -1 {
				start = i
			}
			if ch == '}' {
				end = i
			}
		}
		if start != -1 && end != -1 && end > start {
			response = response[start : end+1]
		}
	}
	return response
}

// GetMetrics returns accumulated metrics
func (c *OpenAIClient) GetMetrics() *Metrics {
	return c.metrics
}

// calculateCost calculates the cost of an API call based on model and tokens
func calculateCost(model string, promptTokens, completionTokens int) float64 {
	// Pricing as of January 2025 (per 1M tokens)
	var promptCost, completionCost float64

	switch model {
	case "gpt-4o", "gpt-4o-2024-11-20":
		promptCost = 2.50     // $2.50/1M tokens
		completionCost = 10.0 // $10/1M tokens
	case "gpt-4-turbo-preview", "gpt-4-turbo":
		promptCost = 10.0     // $10/1M tokens
		completionCost = 30.0 // $30/1M tokens
	case "gpt-4":
		promptCost = 30.0     // $30/1M tokens
		completionCost = 60.0 // $60/1M tokens
	case "gpt-3.5-turbo":
		promptCost = 0.5     // $0.50/1M tokens
		completionCost = 1.5 // $1.50/1M tokens
	default:
		// Default to gpt-4o pricing
		promptCost = 2.50
		completionCost = 10.0
	}

	return (float64(promptTokens)/1000000)*promptCost + (float64(completionTokens)/1000000)*completionCost
}
