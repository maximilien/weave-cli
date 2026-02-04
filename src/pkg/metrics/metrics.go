// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestDuration tracks the duration of VDB requests
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "weave_request_duration_seconds",
			Help: "Duration of VDB requests in seconds",
			Buckets: []float64{.001, .005, .01, .05, .1, .5, 1, 5, 10},
		},
		[]string{"vdb_type", "operation", "status"},
	)

	// DocumentCount tracks the number of documents processed
	DocumentCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weave_documents_total",
			Help: "Total number of documents processed",
		},
		[]string{"vdb_type", "operation"},
	)

	// ErrorCount tracks errors encountered
	ErrorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weave_errors_total",
			Help: "Total number of errors encountered",
		},
		[]string{"vdb_type", "operation", "error_type"},
	)

	// ActiveConnections tracks active VDB connections
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "weave_active_connections",
			Help: "Number of active VDB connections",
		},
		[]string{"vdb_type"},
	)
)

// RecordRequest records a VDB request with timing
func RecordRequest(vdbType, operation string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	RequestDuration.WithLabelValues(vdbType, operation, status).Observe(duration.Seconds())

	if err != nil {
		errorType := "unknown"
		// You can enhance this to categorize errors (timeout, auth, network, etc.)
		ErrorCount.WithLabelValues(vdbType, operation, errorType).Inc()
	}
}

// RecordDocument records document operations
func RecordDocument(vdbType, operation string, count int) {
	DocumentCount.WithLabelValues(vdbType, operation).Add(float64(count))
}

// RecordError records an error
func RecordError(vdbType, operation, errorType string) {
	ErrorCount.WithLabelValues(vdbType, operation, errorType).Inc()
}

// IncrementActiveConnections increments active connection count
func IncrementActiveConnections(vdbType string) {
	ActiveConnections.WithLabelValues(vdbType).Inc()
}

// DecrementActiveConnections decrements active connection count
func DecrementActiveConnections(vdbType string) {
	ActiveConnections.WithLabelValues(vdbType).Dec()
}
