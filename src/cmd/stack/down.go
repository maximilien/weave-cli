// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

// DownCmd represents the stack down command
var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove the stack",
	Long: `Stop all stack services and remove the Kubernetes cluster.

This command:
1. Uninstalls Helm releases
2. Deletes the K8s cluster
3. Cleans up state files

Example:
  weave stack down`,
	RunE: runDown,
}

func runDown(cmd *cobra.Command, args []string) error {
	utils.PrintInfo("Stopping weave stack...")
	fmt.Println()

	// Load cluster info
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		utils.PrintWarning("No active cluster found")
		return nil
	}

	utils.PrintInfo(fmt.Sprintf("Deleting %s cluster: %s", clusterInfo.Provider, clusterInfo.Name))

	// Delete cluster
	if err := stackpkg.DeleteCluster(clusterInfo); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("✅ Cluster deleted: %s", clusterInfo.Name))

	// TODO: Clean up state files
	// For now, we keep cluster.json for reference

	fmt.Println()
	utils.PrintSuccess("🎉 Stack stopped successfully!")

	return nil
}
