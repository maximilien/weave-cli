// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package vectordb

import (
	"testing"
	"time"
)

func TestTimeoutStrategy(t *testing.T) {
	tests := []struct {
		name  string
		op    OperationType
		cloud bool
		want  time.Duration
	}{
		{name: "local health", op: OperationTypeHealth, want: 10 * time.Second},
		{name: "cloud health", op: OperationTypeHealth, cloud: true, want: 20 * time.Second},
		{name: "local document", op: OperationTypeDocument, want: 15 * time.Second},
		{name: "cloud document", op: OperationTypeDocument, cloud: true, want: 30 * time.Second},
		{name: "local collection", op: OperationTypeCollection, want: 20 * time.Second},
		{name: "cloud collection", op: OperationTypeCollection, cloud: true, want: 40 * time.Second},
		{name: "local query", op: OperationTypeQuery, want: 20 * time.Second},
		{name: "cloud query", op: OperationTypeQuery, cloud: true, want: 40 * time.Second},
		{name: "local schema", op: OperationTypeSchema, want: 15 * time.Second},
		{name: "cloud schema", op: OperationTypeSchema, cloud: true, want: 30 * time.Second},
		{name: "local bulk", op: OperationTypeBulk, want: 120 * time.Second},
		{name: "cloud bulk", op: OperationTypeBulk, cloud: true, want: 300 * time.Second},
		{name: "local fallback", op: OperationType(999), want: 30 * time.Second},
		{name: "cloud fallback", op: OperationType(999), cloud: true, want: 60 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &TimeoutStrategy{IsCloud: tt.cloud}
			if got := strategy.GetTimeout(tt.op, 0); got != tt.want {
				t.Fatalf("GetTimeout() = %s, want %s", got, tt.want)
			}
			if got := GetTimeoutForOperation(tt.op, tt.cloud, 0); got != tt.want {
				t.Fatalf("GetTimeoutForOperation() = %s, want %s", got, tt.want)
			}
		})
	}

	if got := (&TimeoutStrategy{}).GetTimeout(OperationTypeBulk, 7); got != 7*time.Second {
		t.Fatalf("configured timeout = %s, want 7s", got)
	}
}
