// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vectordb

import "time"

// OperationType represents the type of database operation for timeout calculation
type OperationType int

const (
	// OperationTypeHealth represents health check operations
	OperationTypeHealth OperationType = iota
	// OperationTypeCollection represents collection management operations (create, delete, list)
	OperationTypeCollection
	// OperationTypeDocument represents single document operations (create, get, update, delete)
	OperationTypeDocument
	// OperationTypeBulk represents bulk document operations (create many, delete many)
	OperationTypeBulk
	// OperationTypeQuery represents query/search operations
	OperationTypeQuery
	// OperationTypeSchema represents schema operations (get, update, validate)
	OperationTypeSchema
)

// TimeoutStrategy defines timeout values for different operation types and deployment types
type TimeoutStrategy struct {
	IsCloud bool
}

// GetTimeout returns the appropriate timeout duration for the given operation type
// based on deployment type (local vs cloud) and operation complexity
func (ts *TimeoutStrategy) GetTimeout(opType OperationType, baseTimeout int) time.Duration {
	// If a specific timeout is configured, use it as the base
	if baseTimeout > 0 {
		return time.Duration(baseTimeout) * time.Second
	}

	// Use smart defaults based on operation type and deployment type
	switch opType {
	case OperationTypeHealth:
		if ts.IsCloud {
			return 20 * time.Second // Cloud: allow for network latency
		}
		return 10 * time.Second // Local: expect fast response

	case OperationTypeDocument:
		if ts.IsCloud {
			return 30 * time.Second // Cloud: network latency + processing
		}
		return 15 * time.Second // Local: quick single-doc operations

	case OperationTypeCollection:
		if ts.IsCloud {
			return 40 * time.Second // Cloud: schema operations can be slow
		}
		return 20 * time.Second // Local: faster local operations

	case OperationTypeQuery:
		if ts.IsCloud {
			return 40 * time.Second // Cloud: complex queries + network
		}
		return 20 * time.Second // Local: vector search is fast

	case OperationTypeSchema:
		if ts.IsCloud {
			return 30 * time.Second // Cloud: schema operations
		}
		return 15 * time.Second // Local: schema retrieval

	case OperationTypeBulk:
		if ts.IsCloud {
			return 300 * time.Second // Cloud: 5 minutes for large batches
		}
		return 120 * time.Second // Local: 2 minutes for bulk operations

	default:
		// Fallback to conservative default
		if ts.IsCloud {
			return 60 * time.Second
		}
		return 30 * time.Second
	}
}

// GetTimeoutForOperation is a convenience function that creates a TimeoutStrategy
// and returns the timeout for a specific operation
func GetTimeoutForOperation(opType OperationType, isCloud bool, baseTimeout int) time.Duration {
	strategy := &TimeoutStrategy{IsCloud: isCloud}
	return strategy.GetTimeout(opType, baseTimeout)
}
