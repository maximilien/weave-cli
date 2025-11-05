// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"
)

// OpikConfig holds Opik configuration
type OpikConfig struct {
	Enabled     bool
	APIKey      string
	Endpoint    string
	Workspace   string
	ProjectName string
}

// LoadOpikConfig loads Opik configuration from environment
// Reads from OPIK_API_KEY, OPIK_WORKSPACE, OPIK_PROJECT_NAME, and OTEL_EXPORTER_OTLP_ENDPOINT
func LoadOpikConfig() *OpikConfig {
	enabled := os.Getenv("OPIK_ENABLED") == "true"
	apiKey := os.Getenv("OPIK_API_KEY")
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	workspace := os.Getenv("OPIK_WORKSPACE")
	projectName := os.Getenv("OPIK_PROJECT_NAME")

	// Default values
	if endpoint == "" {
		// Use the traces-specific endpoint for OpenTelemetry
		endpoint = "https://www.comet.com/opik/api/v1/private/otel/v1/traces"
	}
	if workspace == "" {
		workspace = "default"
	}
	if projectName == "" {
		projectName = "weave-cli-queries"
	}

	// Enable Opik if API key is provided
	if apiKey != "" {
		enabled = true
	}

	return &OpikConfig{
		Enabled:     enabled,
		APIKey:      apiKey,
		Endpoint:    endpoint,
		Workspace:   workspace,
		ProjectName: projectName,
	}
}

// InitOpikTracing initializes OpenTelemetry tracing with Opik backend
func InitOpikTracing(ctx context.Context, config *OpikConfig) (*sdktrace.TracerProvider, error) {
	if !config.Enabled || config.APIKey == "" {
		// Return no-op tracer provider if Opik is disabled
		return nil, nil
	}

	// Parse the endpoint URL
	parsedURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid OTEL endpoint URL: %w", err)
	}

	// Create OTLP HTTP exporter with custom headers for Opik
	headers := map[string]string{
		"Authorization":   config.APIKey,
		"Comet-Workspace": config.Workspace,
		"projectName":     config.ProjectName,
	}

	// Debug: Print configuration
	if os.Getenv("OPIK_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "Opik Config:\n")
		fmt.Fprintf(os.Stderr, "  Endpoint: %s\n", config.Endpoint)
		fmt.Fprintf(os.Stderr, "  Host: %s\n", parsedURL.Host)
		fmt.Fprintf(os.Stderr, "  Path: %s\n", parsedURL.Path)
		fmt.Fprintf(os.Stderr, "  Workspace: %s\n", config.Workspace)
		fmt.Fprintf(os.Stderr, "  Project: %s\n", config.ProjectName)
		fmt.Fprintf(os.Stderr, "  Headers: Authorization=<redacted>, Comet-Workspace=%s, projectName=%s\n", config.Workspace, config.ProjectName)
	}

	// Configure exporter options
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(parsedURL.Host),
		otlptracehttp.WithHeaders(headers),
		otlptracehttp.WithURLPath(parsedURL.Path),
	}

	// Use secure connection if scheme is https
	if parsedURL.Scheme == "https" {
		// By default, otlptracehttp uses secure connection
	} else if parsedURL.Scheme == "http" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("weave-cli"),
			semconv.ServiceVersionKey.String("0.2.14"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// ShutdownOpikTracing gracefully shuts down the tracer provider
func ShutdownOpikTracing(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

// WrapHTTPClient wraps an HTTP client with OpenTelemetry instrumentation
func WrapHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}

	// Wrap the transport with OTEL instrumentation
	client.Transport = otelhttp.NewTransport(client.Transport)
	return client
}
