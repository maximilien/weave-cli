// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os"
	"os/exec"

	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

// KubectlCmd represents the kubectl passthrough command
var KubectlCmd = &cobra.Command{
	Use:   "kubectl -- <kubectl-args>",
	Short: "Run kubectl commands with stack context",
	Long: `Run kubectl commands with the stack's Kubernetes context automatically set.

This command is a convenience wrapper that automatically adds the --context flag
pointing to your stack's Kubernetes cluster, so you don't have to specify it manually.

Examples:
  weave stack kubectl -- get pods
  weave stack kubectl -- describe pod weave-stack-milvus-xyz
  weave stack kubectl -- get services
  weave stack kubectl -- logs weave-stack-milvus-xyz --follow

Note: The "--" separator is required to separate stack flags from kubectl flags.`,
	RunE:                  runKubectl,
	DisableFlagParsing:    true, // Don't parse kubectl flags
	DisableFlagsInUseLine: true,
}

func runKubectl(cmd *cobra.Command, args []string) error {
	// Load cluster info
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		return fmt.Errorf("no active stack found (run: weave stack up --runtime kind)")
	}

	// Find "--" separator
	var kubectlArgs []string
	foundSeparator := false
	for i, arg := range args {
		if arg == "--" {
			foundSeparator = true
			// Take everything after "--"
			if i+1 < len(args) {
				kubectlArgs = args[i+1:]
			}
			break
		}
	}

	if !foundSeparator || len(kubectlArgs) == 0 {
		return fmt.Errorf(`no kubectl arguments provided

Usage: weave stack kubectl -- <kubectl-args>

Examples:
  weave stack kubectl -- get pods
  weave stack kubectl -- describe svc weave-stack-milvus

Note: The "--" separator is required`)
	}

	// Prepend --context flag
	kubectlArgsWithContext := append([]string{"--context", clusterInfo.Context}, kubectlArgs...)

	// Run kubectl
	kubectlCmd := exec.Command("kubectl", kubectlArgsWithContext...)
	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr
	kubectlCmd.Stdin = os.Stdin

	return kubectlCmd.Run()
}
