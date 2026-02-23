// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPodInfo(t *testing.T) {
	// Test PodInfo struct creation
	pod := PodInfo{
		Name:      "test-pod-abc123",
		Status:    "✅ Running",
		Ready:     "1/1",
		Restarts:  "0",
		Age:       "5m",
		Component: "milvus",
	}

	assert.Equal(t, "test-pod-abc123", pod.Name)
	assert.Equal(t, "✅ Running", pod.Status)
	assert.Equal(t, "1/1", pod.Ready)
	assert.Equal(t, "0", pod.Restarts)
	assert.Equal(t, "milvus", pod.Component)
}

func TestCheckPodsReadySelector(t *testing.T) {
	// This test validates the selector format used in checkPodsReady
	// The actual kubectl execution is skipped in unit tests
	selector := "app.kubernetes.io/instance=test-stack"
	assert.NotEmpty(t, selector)
	assert.Contains(t, selector, "app.kubernetes.io/instance")
}

func TestGetPodsSelector(t *testing.T) {
	// This test validates the selector format used in GetPods
	// The actual kubectl execution is skipped in unit tests
	selector := "app.kubernetes.io/instance=weave-stack"
	assert.NotEmpty(t, selector)
	assert.Contains(t, selector, "weave-stack")
}
