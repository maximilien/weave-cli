// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// PodInfo represents information about a Kubernetes pod
type PodInfo struct {
	Name      string
	Status    string
	Ready     string
	Restarts  string
	Age       string
	Component string
}

// WaitForPods waits for pods matching selector to be Ready
func WaitForPods(selector, context string, timeout time.Duration) error {
	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for pods to be ready")
		}

		ready, err := checkPodsReady(selector, context)
		if err != nil {
			// Pod might not exist yet, continue waiting
			<-ticker.C
			continue
		}

		if ready {
			return nil
		}

		<-ticker.C
	}
}

// checkPodsReady checks if all pods matching selector are Ready
func checkPodsReady(selector, context string) (bool, error) {
	args := []string{
		"get", "pods",
		"-l", selector,
		"-o", "jsonpath={.items[*].status.conditions[?(@.type==\"Ready\")].status}",
	}

	if context != "" {
		args = append(args, "--context", context)
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Check if all pods report "True"
	statuses := strings.Fields(string(output))
	if len(statuses) == 0 {
		return false, nil
	}

	for _, status := range statuses {
		if status != "True" {
			return false, nil
		}
	}

	return true, nil
}

// GetPods gets pods matching selector
func GetPods(selector, context string) ([]PodInfo, error) {
	args := []string{
		"get", "pods",
		"-l", selector,
		"-o", "json",
	}

	if context != "" {
		args = append(args, "--context", context)
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// Parse JSON output
	var result struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					Ready        bool `json:"ready"`
					RestartCount int  `json:"restartCount"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pod JSON: %w", err)
	}

	var pods []PodInfo
	for _, item := range result.Items {
		// Calculate ready status
		readyCount := 0
		totalCount := len(item.Status.ContainerStatuses)
		restarts := 0

		for _, cs := range item.Status.ContainerStatuses {
			if cs.Ready {
				readyCount++
			}
			restarts += cs.RestartCount
		}

		ready := fmt.Sprintf("%d/%d", readyCount, totalCount)

		// Determine component from labels
		component := item.Metadata.Labels["app"]
		if component == "" {
			component = item.Metadata.Labels["app.kubernetes.io/component"]
		}
		if component == "" {
			component = "unknown"
		}

		// Health indicator
		health := "❌"
		if readyCount == totalCount && item.Status.Phase == "Running" {
			health = "✅"
		} else if item.Status.Phase == "Pending" {
			health = "⏳"
		}

		pods = append(pods, PodInfo{
			Name:      item.Metadata.Name,
			Status:    item.Status.Phase,
			Ready:     ready,
			Restarts:  fmt.Sprintf("%d", restarts),
			Age:       "N/A", // Would need to calculate from creation time
			Component: component,
		})

		// Use health in status display
		pods[len(pods)-1].Status = fmt.Sprintf("%s %s", health, item.Status.Phase)
	}

	return pods, nil
}

// GetPodLogs streams logs from pods matching selector
func GetPodLogs(selector, context string, follow bool, tail int) error {
	args := []string{
		"logs",
		"-l", selector,
		"--tail", fmt.Sprintf("%d", tail),
	}

	if follow {
		args = append(args, "--follow")
	}

	if context != "" {
		args = append(args, "--context", context)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
