// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/stack"
)

// CheckStack verifies the weave stack status: cluster info, pod status, and VDB reachability.
func CheckStack() []CheckResult {
	var results []CheckResult

	if !stack.StackExists() {
		results = append(results, CheckResult{
			Section: SectionStack,
			Name:    "Stack config",
			Status:  StatusSkip,
			Message: "weave-stack.yaml not found",
		})
		return results
	}

	results = append(results, CheckResult{
		Section: SectionStack,
		Name:    "Stack config",
		Status:  StatusOK,
		Message: "weave-stack.yaml found",
	})

	// Load cluster info
	clusterInfo, err := stack.LoadClusterInfo()
	if err != nil {
		results = append(results, CheckResult{
			Section: SectionStack,
			Name:    "Cluster info",
			Status:  StatusFail,
			Message: fmt.Sprintf("failed to load: %v", err),
			Fix:     "Run: weave stack up",
		})
		return results
	}

	if clusterInfo == nil {
		results = append(results, CheckResult{
			Section: SectionStack,
			Name:    "Cluster info",
			Status:  StatusWarn,
			Message: "no active cluster found",
			Fix:     "Run: weave stack up",
		})
		return results
	}

	// Report cluster details
	results = append(results, CheckResult{
		Section: SectionStack,
		Name:    "Cluster",
		Status:  StatusOK,
		Message: fmt.Sprintf("%s (%s), status: %s", clusterInfo.Name, clusterInfo.Provider, clusterInfo.Status),
	})

	if clusterInfo.Status != "active" {
		results = append(results, CheckResult{
			Section: SectionStack,
			Name:    "Cluster status",
			Status:  StatusWarn,
			Message: fmt.Sprintf("cluster status is %q, expected \"active\"", clusterInfo.Status),
			Fix:     "Run: weave stack up",
		})
	}

	// Check cluster health via provider-specific status
	statusMsg, err := stack.GetClusterStatus(clusterInfo)
	if err != nil {
		results = append(results, CheckResult{
			Section: SectionStack,
			Name:    "Cluster health",
			Status:  StatusFail,
			Message: fmt.Sprintf("failed to query cluster status: %v", err),
			Fix:     "Ensure kubectl is configured and the cluster is running",
		})
	} else {
		results = append(results, CheckResult{
			Section: SectionStack,
			Name:    "Cluster health",
			Status:  StatusOK,
			Message: statusMsg,
		})
	}

	return results
}
