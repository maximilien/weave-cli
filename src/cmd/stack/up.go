// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

var (
	upRuntime             string
	upSkipClusterCreation bool
)

// UpCmd represents the stack up command
var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy the stack to Kubernetes",
	Long: `Deploy the complete RAG stack to Kubernetes.

This command:
1. Creates a K8s cluster (if needed)
2. Generates Helm charts from weave-stack.yaml
3. Deploys infrastructure (VectorDB, etc.)
4. Waits for services to be ready

Example:
  weave stack up --runtime kind
  weave stack up --runtime minikube
  weave stack up --skip-cluster-creation  # Use existing cluster`,
	RunE: runUp,
}

func init() {
	UpCmd.Flags().StringVar(&upRuntime, "runtime", "", "Kubernetes runtime (kind, minikube)")
	UpCmd.Flags().BoolVar(&upSkipClusterCreation, "skip-cluster-creation", false, "Skip cluster creation (use existing cluster)")
}

func runUp(cmd *cobra.Command, args []string) error {
	utils.PrintInfo("Deploying weave stack...")
	fmt.Println()

	// Load stack config
	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	// Determine runtime
	runtime := upRuntime
	if runtime == "" {
		runtime = config.Runtime.Kubernetes.Provider
	}
	if runtime == "" {
		return fmt.Errorf("no runtime specified (use --runtime or set in weave-stack.yaml)")
	}

	// Override provider in config
	config.Runtime.Kubernetes.Provider = runtime

	// Step 1: Create cluster (if needed)
	var clusterInfo *stackpkg.ClusterInfo
	if !upSkipClusterCreation {
		utils.PrintInfo(fmt.Sprintf("Creating %s cluster...", runtime))

		switch runtime {
		case "kind":
			clusterInfo, err = stackpkg.CreateKindCluster(config)
		case "minikube":
			clusterInfo, err = stackpkg.CreateMinikubeCluster(config)
		default:
			return fmt.Errorf("unsupported runtime: %s (supported: kind, minikube)", runtime)
		}

		if err != nil {
			return fmt.Errorf("failed to create cluster: %w", err)
		}

		utils.PrintSuccess(fmt.Sprintf("✅ Cluster created: %s", clusterInfo.Name))
		utils.PrintSuccess(fmt.Sprintf("   Context: %s", clusterInfo.Context))
		utils.PrintSuccess(fmt.Sprintf("   Runtime: %s", clusterInfo.ContainerRuntime))
	} else {
		// Load existing cluster info
		clusterInfo, err = stackpkg.LoadClusterInfo()
		if err != nil {
			return fmt.Errorf("failed to load cluster info: %w", err)
		}
		if clusterInfo == nil {
			return fmt.Errorf("no cluster info found (remove --skip-cluster-creation to create new cluster)")
		}

		utils.PrintInfo(fmt.Sprintf("Using existing cluster: %s", clusterInfo.Name))
	}

	fmt.Println()

	// TODO: Step 2: Generate Helm charts (Phase 1 Day 3)
	utils.PrintInfo("⏳ Helm chart generation (coming in Phase 1 Day 3)")

	// TODO: Step 3: Deploy infrastructure (Phase 1 Day 4)
	utils.PrintInfo("⏳ Infrastructure deployment (coming in Phase 1 Day 4)")

	// TODO: Step 4: Wait for services (Phase 1 Day 4)
	utils.PrintInfo("⏳ Service health checks (coming in Phase 1 Day 4)")

	fmt.Println()
	utils.PrintSuccess("🎉 Stack deployment initiated!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println(fmt.Sprintf("  1. Check cluster: kubectl --context %s get pods", clusterInfo.Context))
	fmt.Println("  2. Monitor status: weave stack status")
	fmt.Println("  3. View logs: kubectl --context", clusterInfo.Context, "logs <pod-name>")

	return nil
}
