// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package llm

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan creates a span with Opik-friendly attributes for monitoring dashboards.
func StartSpan(ctx context.Context, tracerName, name, spanType string, input interface{}, metadata map[string]interface{}, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, name)

	baseAttrs := []attribute.KeyValue{
		attribute.String("type", spanType),
		attribute.String("opik.span.type", spanType),
	}

	if input != nil {
		serialized := serializeTraceValue(input)
		baseAttrs = append(baseAttrs,
			attribute.String("input", serialized),
			attribute.String("opik.input", serialized),
		)
	}

	if metadata != nil {
		serialized := serializeTraceValue(metadata)
		baseAttrs = append(baseAttrs,
			attribute.String("metadata", serialized),
			attribute.String("opik.metadata", serialized),
		)
	}

	baseAttrs = append(baseAttrs, attrs...)
	span.SetAttributes(baseAttrs...)

	return ctx, span
}

// FinishSpan records span output and errors using stringified JSON payloads so they are visible in Opik.
func FinishSpan(span trace.Span, output interface{}, err error, attrs ...attribute.KeyValue) {
	if span == nil {
		return
	}

	if output != nil {
		serialized := serializeTraceValue(output)
		attrs = append(attrs,
			attribute.String("output", serialized),
			attribute.String("opik.output", serialized),
		)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		attrs = append(attrs, attribute.String("error.message", err.Error()))
	}

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	span.End()
}

func serializeTraceValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "<unserializable>"
	}

	return string(data)
}
