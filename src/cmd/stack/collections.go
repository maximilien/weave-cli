// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/spf13/cobra"
)

var (
	collectionsForce bool
	collectionsKeep  []string
)

// CollectionsCmd represents the collections command
var CollectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Manage stack collections",
	Long: `Manage collections in the deployed stack.

Commands:
  reset   Delete all collections without tearing down infrastructure`,
}

// collectionsResetCmd represents the collections reset command
var collectionsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete all collections (keeps infrastructure running)",
	Long: `Delete all collections in the stack without tearing down Milvus, MinIO, or etcd.

This is much faster than 'weave stack down && weave stack up' when you only
want to clear data, not restart infrastructure.

Examples:
  # Delete all collections (with confirmation)
  weave stack collections reset

  # Delete all collections (no confirmation)
  weave stack collections reset --force

  # Delete all except specific collections
  weave stack collections reset --keep Documents --keep Images

Use Cases:
  - Development/testing iteration
  - Reset data between test runs
  - Clean slate without infrastructure restart

Time Comparison:
  weave stack down/up: ~2 minutes (stops/starts all services)
  weave stack collections reset: ~5 seconds (deletes collections only)`,
	RunE: runCollectionsReset,
}

func init() {
	CollectionsCmd.AddCommand(collectionsResetCmd)

	collectionsResetCmd.Flags().BoolVar(&collectionsForce, "force", false, "Skip confirmation prompt")
	collectionsResetCmd.Flags().StringSliceVar(&collectionsKeep, "keep", []string{}, "Collections to keep (don't delete)")
}

func runCollectionsReset(cmd *cobra.Command, args []string) error {
	utils.PrintInfo("Loading stack information...")

	// Load cluster info to verify stack is running
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		return fmt.Errorf("no active stack found\n\nRun: weave stack up --runtime kind")
	}

	utils.PrintInfo(fmt.Sprintf("Stack: %s", clusterInfo.Name))
	utils.PrintInfo(fmt.Sprintf("Context: %s", clusterInfo.Context))
	fmt.Println()

	// Start port forwarding to Milvus
	utils.PrintInfo("Connecting to Milvus...")

	portForwardCtx, err := stackpkg.StartMilvusPortForward(clusterInfo)
	if err != nil {
		return fmt.Errorf("failed to start port forwarding: %w", err)
	}
	defer portForwardCtx.Stop()

	utils.PrintSuccess("✅ Connected to Milvus")
	fmt.Println()

	// Create Milvus client using factory
	milvusAddress := fmt.Sprintf("localhost:%d", portForwardCtx.LocalPort)
	vdbConfig := &vectordb.Config{
		Type:    vectordb.VectorDBTypeMilvusLocal,
		Address: milvusAddress,
		Timeout: 30,
	}

	// Use the global registry to create client
	registry := vectordb.NewRegistry()
	client, err := registry.CreateClient(vdbConfig)
	if err != nil {
		return fmt.Errorf("failed to create Milvus client: %w", err)
	}

	// List all collections
	utils.PrintInfo("Fetching collections...")
	collections, err := client.ListCollections(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if len(collections) == 0 {
		utils.PrintWarning("No collections found")
		return nil
	}

	// Filter collections to delete (exclude --keep collections)
	collectionsToDelete := []string{}
	collectionsKept := []string{}

	keepMap := make(map[string]bool)
	for _, keep := range collectionsKeep {
		keepMap[keep] = true
	}

	for _, col := range collections {
		if keepMap[col.Name] {
			collectionsKept = append(collectionsKept, col.Name)
		} else {
			collectionsToDelete = append(collectionsToDelete, col.Name)
		}
	}

	// Show what will be deleted
	if len(collectionsToDelete) == 0 {
		utils.PrintWarning("No collections to delete (all collections are in --keep list)")
		return nil
	}

	fmt.Printf("📋 Collections to delete: %d\n", len(collectionsToDelete))
	for _, col := range collectionsToDelete {
		fmt.Printf("  • %s\n", col)
	}
	fmt.Println()

	if len(collectionsKept) > 0 {
		fmt.Printf("✅ Collections to keep: %d\n", len(collectionsKept))
		for _, col := range collectionsKept {
			fmt.Printf("  • %s\n", col)
		}
		fmt.Println()
	}

	// Confirmation prompt (unless --force)
	if !collectionsForce {
		confirmed := utils.ConfirmAction(fmt.Sprintf("Delete %d collections?", len(collectionsToDelete)))
		if !confirmed {
			utils.PrintWarning("Reset cancelled")
			return nil
		}
		fmt.Println()
	}

	// Delete collections
	utils.PrintInfo(fmt.Sprintf("Deleting %d collections...", len(collectionsToDelete)))
	fmt.Println()

	successCount := 0
	failedCount := 0
	for i, col := range collectionsToDelete {
		fmt.Printf("[%d/%d] Deleting collection: %s... ", i+1, len(collectionsToDelete), col)

		err := client.DeleteCollection(context.Background(), col)
		if err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			failedCount++
		} else {
			fmt.Printf("✅ DELETED\n")
			successCount++
		}
	}

	fmt.Println()

	// Summary
	if failedCount > 0 {
		utils.PrintWarning(fmt.Sprintf("Reset completed with errors: %d deleted, %d failed", successCount, failedCount))
		return fmt.Errorf("some collections failed to delete")
	}

	utils.PrintSuccess(fmt.Sprintf("🎉 Successfully deleted %d collections!", successCount))
	utils.PrintInfo("Infrastructure (Milvus, MinIO, etcd) still running")
	utils.PrintInfo("Collections will be recreated automatically on next ingest")

	return nil
}
