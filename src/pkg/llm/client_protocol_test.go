// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

type llmRoundTripFunc func(*http.Request) (*http.Response, error)

func (f llmRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func llmResponse(status int, body string) *http.Response {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestOpenAIClientCompleteProtocolAndMetrics(t *testing.T) {
	var requestBody string
	httpClient := &http.Client{Transport: llmRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		requestBody = string(data)
		return llmResponse(http.StatusOK, `{
			"id":"chatcmpl-test","object":"chat.completion","created":1,"model":"gpt-4o",
			"choices":[{"index":0,"message":{"role":"assistant","content":"hello world"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":10,"completion_tokens":4,"total_tokens":14}
		}`), nil
	})}
	client, err := NewOpenAIClientWithHTTP("test-key", httpClient)
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.Complete(context.Background(), "say hello",
		WithModel("gpt-4o"), WithTemperature(0.25), WithMaxTokens(128), WithSystemMessage("be concise"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello world" {
		t.Fatalf("unexpected completion: %q", got)
	}
	for _, want := range []string{`"content":"be concise"`, `"content":"say hello"`, `"temperature":0.25`, `"max_tokens":128`} {
		if !strings.Contains(requestBody, want) {
			t.Errorf("request missing %s: %s", want, requestBody)
		}
	}
	metrics := client.GetMetrics()
	if metrics.Invocations != 1 || metrics.PromptTokens != 10 || metrics.CompletionTokens != 4 || metrics.TotalTokens != 14 {
		t.Fatalf("unexpected metrics: %#v", metrics)
	}
	if metrics.TotalCost <= 0 {
		t.Fatalf("expected positive cost, got %f", metrics.TotalCost)
	}
}

func TestOpenAIClientCompleteFailures(t *testing.T) {
	t.Run("transport", func(t *testing.T) {
		client, err := NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: llmRoundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network unavailable")
		})})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Complete(context.Background(), "hello"); err == nil || !strings.Contains(err.Error(), "network unavailable") {
			t.Fatalf("expected transport error, got %v", err)
		}
	})

	t.Run("no choices", func(t *testing.T) {
		client, err := NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: llmRoundTripFunc(func(*http.Request) (*http.Response, error) {
			return llmResponse(http.StatusOK, `{"id":"empty","object":"chat.completion","created":1,"model":"gpt-4o","choices":[],"usage":{"prompt_tokens":2,"completion_tokens":0,"total_tokens":2}}`), nil
		})})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Complete(context.Background(), "hello"); err == nil || !strings.Contains(err.Error(), "no completion choices") {
			t.Fatalf("expected empty choices error, got %v", err)
		}
		if client.GetMetrics().Invocations != 1 {
			t.Fatalf("expected successful API invocation to be counted: %#v", client.GetMetrics())
		}
	})
}

func TestOpenAIClientCompleteStructured(t *testing.T) {
	var requestBody string
	client, err := NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: llmRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		requestBody = string(data)
		return llmResponse(http.StatusOK, "{\"id\":\"json\",\"object\":\"chat.completion\",\"created\":1,\"model\":\"gpt-4o\",\"choices\":[{\"index\":0,\"message\":{\"role\":\"assistant\",\"content\":\"```json\\n{\\\"answer\\\":42}\\n```\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":3,\"total_tokens\":6}}"), nil
	})})
	if err != nil {
		t.Fatal(err)
	}

	target := &struct {
		Answer int `json:"answer"`
	}{}
	got, err := client.CompleteStructured(context.Background(), "calculate", target, WithSystemMessage("use integers"))
	if err != nil {
		t.Fatal(err)
	}
	if got != target || target.Answer != 42 {
		t.Fatalf("unexpected structured result: %#v", got)
	}
	for _, want := range []string{"use integers", "valid JSON only", "without markdown code fences"} {
		if !strings.Contains(requestBody, want) {
			t.Errorf("structured request missing %q: %s", want, requestBody)
		}
	}

	badClient, err := NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: llmRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return llmResponse(http.StatusOK, `{"id":"bad","object":"chat.completion","created":1,"model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"not-json"},"finish_reason":"stop"}]}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := badClient.CompleteStructured(context.Background(), "calculate", &map[string]interface{}{}); err == nil || !strings.Contains(err.Error(), "failed to parse") {
		t.Fatalf("expected JSON parse error, got %v", err)
	}
}

func TestOpenAIClientGenerateEmbedding(t *testing.T) {
	var bodies []string
	client, err := NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: llmRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		bodies = append(bodies, string(data))
		return llmResponse(http.StatusOK, `{"object":"list","data":[{"object":"embedding","embedding":[0.25,-0.5,1.0],"index":0}],"model":"text-embedding-3-small","usage":{"prompt_tokens":2,"total_tokens":2}}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}

	vector, err := client.GenerateEmbedding(context.Background(), "hello", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(vector) != 3 || vector[0] != 0.25 || vector[1] != -0.5 || vector[2] != 1 {
		t.Fatalf("unexpected embedding: %#v", vector)
	}
	if !strings.Contains(bodies[0], `"model":"text-embedding-3-small"`) || !strings.Contains(bodies[0], `"input":"hello"`) {
		t.Fatalf("unexpected embedding request: %s", bodies[0])
	}

	emptyClient, err := NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: llmRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return llmResponse(http.StatusOK, `{"object":"list","data":[],"model":"custom","usage":{"prompt_tokens":0,"total_tokens":0}}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := emptyClient.GenerateEmbedding(context.Background(), "hello", "custom"); err == nil || !strings.Contains(err.Error(), "no embedding data") {
		t.Fatalf("expected empty embedding error, got %v", err)
	}

	errorClient, err := NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: llmRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("embedding transport failed")
	})})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := errorClient.GenerateEmbedding(context.Background(), "hello", "custom"); err == nil || !strings.Contains(err.Error(), "embedding transport failed") {
		t.Fatalf("expected embedding transport error, got %v", err)
	}
}

func TestTracingLifecycleAndSerialization(t *testing.T) {
	ctx, span := StartSpan(context.Background(), "test-tracer", "test-span", "llm",
		map[string]interface{}{"prompt": "hello"}, map[string]interface{}{"model": "test"}, attribute.String("custom", "value"))
	if ctx == nil || span == nil {
		t.Fatal("expected context and span")
	}
	FinishSpan(span, map[string]interface{}{"answer": "world"}, errors.New("test failure"), attribute.Int("tokens", 2))
	FinishSpan(nil, nil, nil)

	if got := serializeTraceValue(nil); got != "" {
		t.Fatalf("unexpected nil serialization: %q", got)
	}
	if got := serializeTraceValue("plain"); got != "plain" {
		t.Fatalf("unexpected string serialization: %q", got)
	}
	if got := serializeTraceValue([]byte("bytes")); got != "bytes" {
		t.Fatalf("unexpected byte serialization: %q", got)
	}
	if got := serializeTraceValue(map[string]int{"value": 7}); got != `{"value":7}` {
		t.Fatalf("unexpected JSON serialization: %q", got)
	}
	if got := serializeTraceValue(make(chan int)); got != "<unserializable>" {
		t.Fatalf("unexpected error serialization: %q", got)
	}
}
