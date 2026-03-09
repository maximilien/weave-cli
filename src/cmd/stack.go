// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"github.com/maximilien/weave-cli/src/cmd/stack"
	"github.com/spf13/cobra"
)

// stackCmd represents the stack command
var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Kubernetes stack management",
	Long: `Manage complete RAG stacks with Kubernetes.

Weave Stack deploys and manages full RAG systems including:
- Vector databases (Milvus, Qdrant, Weaviate)
- Data ingestion pipelines
- Evaluations
- Dashboards (Next.js/TypeScript)

All orchestrated via Kubernetes (Kind, Minikube, EKS, GKE) with Helm charts.

Example workflow:
  weave stack init              # Create weave-stack.yaml
  weave stack up --runtime kind # Deploy to local K8s
  weave stack status            # Check component health
  weave stack down              # Stop all services`,
}

func init() {
	rootCmd.AddCommand(stackCmd)
	stackCmd.GroupID = "data"

	// Add all stack subcommands
	stackCmd.AddCommand(stack.InitCmd)
	stackCmd.AddCommand(stack.ValidateCmd)
	stackCmd.AddCommand(stack.UpCmd)
	stackCmd.AddCommand(stack.DownCmd)
	stackCmd.AddCommand(stack.StatusCmd)
	stackCmd.AddCommand(stack.LogsCmd)
	stackCmd.AddCommand(stack.KubectlCmd)
	stackCmd.AddCommand(stack.PortForwardCmd)
	stackCmd.AddCommand(stack.DashboardCmd)
	stackCmd.AddCommand(stack.IngestCmd)
	stackCmd.AddCommand(stack.BackupCmd)
	stackCmd.AddCommand(stack.CollectionsCmd)
	// More commands will be added in subsequent phases
}
