// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

// StatusCmd represents the stack status command
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show stack status",
	Long: `Show the status of the stack and its components.

Example:
  weave stack status`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	utils.PrintHeader("Stack Status")
	fmt.Println()

	// Load cluster info
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		utils.PrintWarning("No active stack found")
		fmt.Println()
		fmt.Println("Initialize a stack with: weave stack init")
		fmt.Println("Deploy a stack with: weave stack up --runtime kind")
		return nil
	}

	// Print cluster info
	utils.PrintInfo("Cluster:")
	fmt.Printf("  Name:     %s\n", clusterInfo.Name)
	fmt.Printf("  Provider: %s\n", clusterInfo.Provider)
	fmt.Printf("  Context:  %s\n", clusterInfo.Context)
	fmt.Printf("  Runtime:  %s\n", clusterInfo.ContainerRuntime)
	fmt.Printf("  Created:  %s\n", clusterInfo.CreatedAt.Format("2006-01-02 15:04:05"))

	// Get cluster status
	status, err := stackpkg.GetClusterStatus(clusterInfo)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get cluster status: %v", err))
	} else {
		fmt.Printf("  Status:   %s\n", status)
	}

	fmt.Println()

	// TODO: Show deployed services (Phase 1 Day 4+)
	utils.PrintInfo("Services:")
	utils.PrintInfo("  ⏳ Service status tracking (coming in Phase 1 Day 4)")

	fmt.Println()

	// Show next steps
	fmt.Println("Useful commands:")
	fmt.Printf("  kubectl --context %s get pods\n", clusterInfo.Context)
	fmt.Printf("  kubectl --context %s get services\n", clusterInfo.Context)
	fmt.Println("  weave stack down  # Stop the stack")

	return nil
}
