// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"

	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   int
)

// LogsCmd represents the logs command
var LogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Stream logs from stack services",
	Long: `Stream logs from services running in the stack.

For Kubernetes deployments, this streams logs from pods.
For PM2 dashboard, use 'weave stack dashboard logs'.

Examples:
  weave stack logs milvus              # Show last 100 lines from Milvus
  weave stack logs milvus --follow     # Follow Milvus logs in real-time
  weave stack logs milvus --tail 50    # Show last 50 lines
  weave stack logs vectordb -f         # Follow vectordb logs`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogs,
}

func init() {
	LogsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	LogsCmd.Flags().IntVar(&logsTail, "tail", 100, "Number of lines to show from the end of the logs")
}

func runLogs(cmd *cobra.Command, args []string) error {
	// Load cluster info
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		return fmt.Errorf("no active stack found (run: weave stack up --runtime kind)")
	}

	// Determine service selector
	var selector string
	if len(args) > 0 {
		// Specific service requested
		service := args[0]
		// Map common service names to labels (using 'app' label which is set in deployment)
		switch service {
		case "milvus", "vectordb", "vector-db":
			selector = "app=milvus"
		default:
			// Generic selector using app label
			selector = fmt.Sprintf("app=%s", service)
		}
	} else {
		// No service specified - show all stack logs using instance label
		selector = fmt.Sprintf("app.kubernetes.io/instance=%s", clusterInfo.Name)
	}

	// Stream logs
	if logsFollow {
		fmt.Printf("Following logs (Ctrl+C to stop)...\n\n")
	}

	err = stackpkg.GetPodLogs(selector, clusterInfo.Context, logsFollow, logsTail)
	if err != nil {
		// Provide helpful error messages
		if err.Error() == "exit status 1" {
			return fmt.Errorf("no pods found matching selector: %s\n\nTip: Check pod names with: weave stack status", selector)
		}
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}
