// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

// PortForwardCmd represents the port-forward command
var PortForwardCmd = &cobra.Command{
	Use:   "port-forward <service> <port-mapping>",
	Short: "Forward local port to service",
	Long: `Forward a local port to a service in the stack.

This command sets up port forwarding from your local machine to a Kubernetes
service in the stack, allowing you to access the service as if it were running locally.

The port mapping format is: LOCAL_PORT:REMOTE_PORT

Examples:
  weave stack port-forward milvus 19530:19530
  weave stack port-forward milvus 8080:19530
  weave stack port-forward vectordb 19530:19530

After running this command:
  - Access Milvus at: localhost:19530
  - Press Ctrl+C to stop port forwarding

Common Services:
  milvus:     19530 (gRPC), 9091 (metrics)
  dashboard:  3000 (HTTP)`,
	Args: cobra.ExactArgs(2),
	RunE: runPortForward,
}

func runPortForward(cmd *cobra.Command, args []string) error {
	service := args[0]
	portMapping := args[1]

	// Load cluster info
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		return fmt.Errorf("no active stack found (run: weave stack up --runtime kind)")
	}

	// Determine service name
	var serviceName string
	switch service {
	case "milvus", "vectordb", "vector-db":
		serviceName = fmt.Sprintf("%s-milvus", clusterInfo.Name)
	case "dashboard", "ui", "web":
		serviceName = fmt.Sprintf("%s-dashboard", clusterInfo.Name)
	default:
		// Generic service name
		serviceName = fmt.Sprintf("%s-%s", clusterInfo.Name, service)
	}

	// Build kubectl port-forward command
	kubectlCmd := exec.Command("kubectl", "port-forward",
		fmt.Sprintf("svc/%s", serviceName),
		portMapping,
		"--context", clusterInfo.Context)

	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr

	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if kubectlCmd.Process != nil {
			_ = kubectlCmd.Process.Kill()
		}
	}()

	// Show info
	fmt.Printf("Forwarding %s to %s...\n", portMapping, serviceName)
	fmt.Printf("Context: %s\n", clusterInfo.Context)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop port forwarding")
	fmt.Println()

	// Run port forwarding
	err = kubectlCmd.Run()
	if err != nil {
		// Check if it's just a normal interrupt
		if err.Error() == "signal: killed" || err.Error() == "exit status 130" {
			fmt.Println("\nPort forwarding stopped")
			return nil
		}
		return fmt.Errorf("port forwarding failed: %w\n\nTip: Check service exists with: weave stack kubectl -- get svc", err)
	}

	return nil
}
