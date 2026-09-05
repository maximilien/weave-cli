// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package vectordb

import (
	"errors"
	"strings"
	"testing"
)

func TestVectorDBError(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewErrorWithCause(ErrorTypeConnection, "database unavailable", cause).
		WithCollection("docs").
		WithDocument("doc-1").
		WithDetail("attempt", 2)

	if !strings.Contains(err.Error(), "caused by: connection refused") {
		t.Fatalf("error string = %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatal("VectorDBError does not unwrap its cause")
	}
	if !errors.Is(err, &VectorDBError{Type: ErrorTypeConnection}) {
		t.Fatal("VectorDBError does not match the same error type")
	}
	if errors.Is(err, &VectorDBError{Type: ErrorTypeTimeout}) || err.Is(cause) {
		t.Fatal("VectorDBError matched an unrelated error")
	}
	if err.Collection != "docs" || err.Document != "doc-1" || err.Details["attempt"] != 2 {
		t.Fatalf("error context = %#v", err)
	}

	withoutCause := NewError(ErrorTypeInvalidQuery, "empty query")
	withoutCause.Details = nil
	withoutCause.WithDetail("field", "query")
	if withoutCause.Error() != "invalid_query: empty query" || withoutCause.Details["field"] != "query" {
		t.Fatalf("unexpected error without cause: %#v", withoutCause)
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	cause := errors.New("cause")
	tests := []struct {
		name       string
		err        *VectorDBError
		wantType   ErrorType
		wantText   string
		wantCause  bool
		classifier func(error) bool
	}{
		{name: "connection", err: ErrConnectionFailed("offline", cause), wantType: ErrorTypeConnection, wantText: "offline", wantCause: true, classifier: IsConnectionError},
		{name: "authentication", err: ErrAuthenticationFailed("denied"), wantType: ErrorTypeAuthentication, wantText: "denied", classifier: IsAuthenticationError},
		{name: "not found", err: ErrNotFound("collection", "docs"), wantType: ErrorTypeNotFound, wantText: "collection 'docs' not found", classifier: IsNotFoundError},
		{name: "already exists", err: ErrAlreadyExists("collection", "docs"), wantType: ErrorTypeAlreadyExists, wantText: "collection 'docs' already exists", classifier: IsAlreadyExistsError},
		{name: "invalid config", err: ErrInvalidConfig("bad config"), wantType: ErrorTypeInvalidConfig, wantText: "bad config"},
		{name: "invalid schema", err: ErrInvalidSchema("bad schema"), wantType: ErrorTypeInvalidSchema, wantText: "bad schema"},
		{name: "invalid query", err: ErrInvalidQuery("bad query"), wantType: ErrorTypeInvalidQuery, wantText: "bad query"},
		{name: "timeout", err: ErrTimeout("search"), wantType: ErrorTypeTimeout, wantText: "operation 'search' timed out", classifier: IsTimeoutError},
		{name: "quota", err: ErrQuotaExceeded("limit"), wantType: ErrorTypeQuotaExceeded, wantText: "limit"},
		{name: "internal", err: ErrInternal("failed", cause), wantType: ErrorTypeInternal, wantText: "failed", wantCause: true},
		{name: "unsupported", err: ErrUnsupported("rerank"), wantType: ErrorTypeUnsupported, wantText: "operation 'rerank' is not supported", classifier: IsUnsupportedError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Type != tt.wantType || !strings.Contains(tt.err.Error(), tt.wantText) {
				t.Fatalf("error = %#v, want type %q containing %q", tt.err, tt.wantType, tt.wantText)
			}
			if tt.wantCause && !errors.Is(tt.err, cause) {
				t.Fatal("constructor did not retain its cause")
			}
			if !IsErrorType(tt.err, tt.wantType) {
				t.Fatal("IsErrorType returned false")
			}
			if tt.classifier != nil && !tt.classifier(tt.err) {
				t.Fatal("specialized classifier returned false")
			}
		})
	}

	plain := errors.New("plain")
	if IsErrorType(plain, ErrorTypeInternal) || IsConnectionError(plain) ||
		IsAuthenticationError(plain) || IsNotFoundError(plain) ||
		IsAlreadyExistsError(plain) || IsTimeoutError(plain) ||
		IsUnsupportedError(plain) {
		t.Fatal("plain error matched a vector database error type")
	}
}
