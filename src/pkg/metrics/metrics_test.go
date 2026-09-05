// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordRequest(t *testing.T) {
	RecordRequest("day1-success", "query", 250*time.Millisecond, nil)
	if got := testutil.CollectAndCount(RequestDuration, "weave_request_duration_seconds"); got != 1 {
		t.Fatalf("request duration metric families = %d, want 1", got)
	}

	RecordRequest("day1-error", "insert", time.Second, errors.New("failed"))
	if got := testutil.ToFloat64(ErrorCount.WithLabelValues("day1-error", "insert", "unknown")); got != 1 {
		t.Fatalf("error count = %v, want 1", got)
	}
}

func TestMetricHelpers(t *testing.T) {
	RecordDocument("day1-docs", "upload", 3)
	if got := testutil.ToFloat64(DocumentCount.WithLabelValues("day1-docs", "upload")); got != 3 {
		t.Fatalf("document count = %v, want 3", got)
	}

	RecordError("day1-record-error", "delete", "timeout")
	if got := testutil.ToFloat64(ErrorCount.WithLabelValues("day1-record-error", "delete", "timeout")); got != 1 {
		t.Fatalf("recorded error count = %v, want 1", got)
	}

	IncrementActiveConnections("day1-connections")
	IncrementActiveConnections("day1-connections")
	DecrementActiveConnections("day1-connections")
	if got := testutil.ToFloat64(ActiveConnections.WithLabelValues("day1-connections")); got != 1 {
		t.Fatalf("active connections = %v, want 1", got)
	}
}
