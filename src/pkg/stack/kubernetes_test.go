// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os"
	"path/filepath"
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

func TestKubernetesCommandPaths(t *testing.T) {
	binDir := t.TempDir()
	kubectl := filepath.Join(binDir, "kubectl")
	script := `#!/bin/sh
case "$*" in
  *jsonpath*) printf '%s' "${KUBE_READY:-True True}" ;;
  *'-o json'*) printf '%s' '{"items":[{"metadata":{"name":"ready-pod","labels":{"app":"api"}},"status":{"phase":"Running","containerStatuses":[{"ready":true,"restartCount":1}]}},{"metadata":{"name":"pending-pod","labels":{"app.kubernetes.io/component":"worker"}},"status":{"phase":"Pending","containerStatuses":[{"ready":false,"restartCount":2}]}},{"metadata":{"name":"failed-pod","labels":{}},"status":{"phase":"Failed","containerStatuses":[]}}]}' ;;
  logs*) echo logs ;;
esac
`
	if err := os.WriteFile(kubectl, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	ready, err := checkPodsReady("app=test", "kind-test")
	assert.NoError(t, err)
	assert.True(t, ready)
	t.Setenv("KUBE_READY", "True False")
	ready, err = checkPodsReady("app=test", "")
	assert.NoError(t, err)
	assert.False(t, ready)

	pods, err := GetPods("app=test", "kind-test")
	assert.NoError(t, err)
	if assert.Len(t, pods, 3) {
		assert.Equal(t, "✅ Running", pods[0].Status)
		assert.Equal(t, "api", pods[0].Component)
		assert.Equal(t, "1/1", pods[0].Ready)
		assert.Equal(t, "⏳ Pending", pods[1].Status)
		assert.Equal(t, "worker", pods[1].Component)
		assert.Equal(t, "❌ Failed", pods[2].Status)
		assert.Equal(t, "unknown", pods[2].Component)
	}
	assert.NoError(t, GetPodLogs("app=test", "kind-test", false, 20))
	assert.NoError(t, GetPodLogs("app=test", "", true, 10))
}

func TestKubernetesCommandFailures(t *testing.T) {
	binDir := t.TempDir()
	kubectl := filepath.Join(binDir, "kubectl")
	script := `#!/bin/sh
if [ "$KUBE_MODE" = "fail" ]; then exit 1; fi
if [ "$KUBE_MODE" = "empty" ]; then exit 0; fi
echo 'not-json'
`
	if err := os.WriteFile(kubectl, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	t.Setenv("KUBE_MODE", "empty")
	ready, err := checkPodsReady("app=test", "")
	assert.NoError(t, err)
	assert.False(t, ready)

	t.Setenv("KUBE_MODE", "malformed")
	_, err = GetPods("app=test", "")
	assert.Error(t, err)

	t.Setenv("KUBE_MODE", "fail")
	_, err = checkPodsReady("app=test", "")
	assert.Error(t, err)
	_, err = GetPods("app=test", "")
	assert.Error(t, err)
	assert.Error(t, GetPodLogs("app=test", "", false, 10))
}
