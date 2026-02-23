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
	dashboardLogsFollow bool
	dashboardLogsTail   int
)

// DashboardCmd represents the dashboard command
var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Manage stack dashboard",
	Long: `Start, stop, and manage the weave stack dashboard.

The dashboard provides a web interface for interacting with your RAG stack.
It can be managed using PM2 (local/VM) or Kubernetes (cloud).

Examples:
  weave stack dashboard start     # Start dashboard with PM2
  weave stack dashboard stop      # Stop dashboard
  weave stack dashboard status    # Show dashboard status
  weave stack dashboard logs      # Stream dashboard logs`,
}

var dashboardStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start dashboard with PM2",
	Long: `Start the dashboard using PM2 process manager.

This command:
1. Loads dashboard configuration from weave-stack.yaml
2. Generates PM2 ecosystem.config.js
3. Starts the dashboard with PM2
4. Monitors process health

Prerequisites:
- npm install -g pm2
- Dashboard built (npm run build)
- weave-stack.yaml with dashboard.pm2 configuration

Examples:
  weave stack dashboard start`,
	RunE: runDashboardStart,
}

var dashboardStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop dashboard PM2 process",
	Long: `Stop the dashboard PM2 process.

This command gracefully stops the PM2-managed dashboard process.

Examples:
  weave stack dashboard stop`,
	RunE: runDashboardStop,
}

var dashboardRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart dashboard PM2 process",
	Long: `Restart the dashboard PM2 process.

This command restarts the PM2-managed dashboard process,
useful after configuration changes or updates.

Examples:
  weave stack dashboard restart`,
	RunE: runDashboardRestart,
}

var dashboardStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show dashboard status",
	Long: `Show the status of the dashboard PM2 process.

This command displays PM2 process information including:
- Process status (online/stopped/errored)
- CPU and memory usage
- Uptime
- Restart count

Examples:
  weave stack dashboard status`,
	RunE: runDashboardStatus,
}

var dashboardLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream dashboard logs",
	Long: `Stream logs from the dashboard PM2 process.

This command displays logs from PM2, including both stdout and stderr.

Examples:
  weave stack dashboard logs              # Show last 100 lines
  weave stack dashboard logs --follow     # Follow logs in real-time
  weave stack dashboard logs --tail 50    # Show last 50 lines`,
	RunE: runDashboardLogs,
}

var dashboardMonitCmd = &cobra.Command{
	Use:   "monit",
	Short: "Open PM2 monitoring interface",
	Long: `Open the PM2 monitoring interface for the dashboard.

This command opens an interactive terminal UI showing:
- Real-time process metrics
- CPU and memory usage graphs
- Log streaming

Press Ctrl+C to exit.

Examples:
  weave stack dashboard monit`,
	RunE: runDashboardMonit,
}

func init() {
	DashboardCmd.AddCommand(dashboardStartCmd)
	DashboardCmd.AddCommand(dashboardStopCmd)
	DashboardCmd.AddCommand(dashboardRestartCmd)
	DashboardCmd.AddCommand(dashboardStatusCmd)
	DashboardCmd.AddCommand(dashboardLogsCmd)
	DashboardCmd.AddCommand(dashboardMonitCmd)

	// Logs flags
	dashboardLogsCmd.Flags().BoolVarP(&dashboardLogsFollow, "follow", "f", false, "Follow log output")
	dashboardLogsCmd.Flags().IntVar(&dashboardLogsTail, "tail", 100, "Number of lines to show")
}

func runDashboardStart(cmd *cobra.Command, args []string) error {
	utils.PrintHeader("Starting Dashboard")
	fmt.Println()

	// Load stack config
	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	// Validate dashboard config
	if config.Dashboard == nil || !config.Dashboard.Enabled {
		utils.PrintError("Dashboard not enabled in weave-stack.yaml")
		fmt.Println()
		fmt.Println("Enable dashboard by adding to weave-stack.yaml:")
		fmt.Print(`
dashboard:
  enabled: true
  runtime: pm2
  pm2:
    app_name: weave-dashboard
    script: dist/index.js
    instances: 1
    max_memory_restart: "1G"
    autorestart: true
    watch: false
    error_log: logs/pm2-error.log
    out_log: logs/pm2-out.log
    merge_logs: true
    min_uptime: "10s"
    max_restarts: 10
    kill_timeout: 5000
    env:
      NODE_ENV: production
      DASHBOARD_PORT: 3000
`)
		return fmt.Errorf("dashboard not enabled")
	}

	if config.Dashboard.Runtime != "pm2" {
		return fmt.Errorf("dashboard runtime is %s, not pm2 (only pm2 runtime supported for 'dashboard start')", config.Dashboard.Runtime)
	}

	if config.Dashboard.PM2 == nil {
		return fmt.Errorf("PM2 configuration missing in weave-stack.yaml")
	}

	// Generate PM2 config
	utils.PrintInfo("Generating PM2 configuration...")
	pm2ConfigPath := "ecosystem.config.js"
	if err := stackpkg.GeneratePM2Config(config, pm2ConfigPath); err != nil {
		return fmt.Errorf("failed to generate PM2 config: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("✅ Generated PM2 config: %s", pm2ConfigPath))
	fmt.Println()

	// Start with PM2
	utils.PrintInfo("Starting dashboard with PM2...")
	appName := config.Dashboard.PM2.AppName
	if err := stackpkg.PM2Start(appName, pm2ConfigPath); err != nil {
		return fmt.Errorf("failed to start dashboard: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("✅ Dashboard started: %s", appName))
	fmt.Println()

	// Show useful commands
	fmt.Println("Useful commands:")
	fmt.Printf("  pm2 monit                        # Monitor processes\n")
	fmt.Printf("  pm2 logs %s              # View logs\n", appName)
	fmt.Printf("  weave stack dashboard status     # Check status\n")
	fmt.Printf("  weave stack dashboard stop       # Stop dashboard\n")
	fmt.Println()

	// Show environment variables
	if len(config.Dashboard.PM2.Env) > 0 {
		fmt.Println("Environment:")
		for key, value := range config.Dashboard.PM2.Env {
			fmt.Printf("  %s=%s\n", key, value)
		}
		fmt.Println()
	}

	return nil
}

func runDashboardStop(cmd *cobra.Command, args []string) error {
	utils.PrintHeader("Stopping Dashboard")
	fmt.Println()

	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	if config.Dashboard == nil || config.Dashboard.PM2 == nil {
		return fmt.Errorf("PM2 config not found in weave-stack.yaml")
	}

	appName := config.Dashboard.PM2.AppName
	utils.PrintInfo(fmt.Sprintf("Stopping PM2 process: %s", appName))

	if err := stackpkg.PM2Stop(appName); err != nil {
		return fmt.Errorf("failed to stop dashboard: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("✅ Dashboard stopped: %s", appName))
	return nil
}

func runDashboardRestart(cmd *cobra.Command, args []string) error {
	utils.PrintHeader("Restarting Dashboard")
	fmt.Println()

	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	if config.Dashboard == nil || config.Dashboard.PM2 == nil {
		return fmt.Errorf("PM2 config not found in weave-stack.yaml")
	}

	appName := config.Dashboard.PM2.AppName
	utils.PrintInfo(fmt.Sprintf("Restarting PM2 process: %s", appName))

	if err := stackpkg.PM2Restart(appName); err != nil {
		return fmt.Errorf("failed to restart dashboard: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("✅ Dashboard restarted: %s", appName))
	return nil
}

func runDashboardStatus(cmd *cobra.Command, args []string) error {
	utils.PrintHeader("Dashboard Status")
	fmt.Println()

	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	if config.Dashboard == nil || config.Dashboard.PM2 == nil {
		return fmt.Errorf("PM2 config not found in weave-stack.yaml")
	}

	appName := config.Dashboard.PM2.AppName
	status, err := stackpkg.PM2Status(appName)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Println(status)
	return nil
}

func runDashboardLogs(cmd *cobra.Command, args []string) error {
	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	if config.Dashboard == nil || config.Dashboard.PM2 == nil {
		return fmt.Errorf("PM2 config not found in weave-stack.yaml")
	}

	appName := config.Dashboard.PM2.AppName

	if dashboardLogsFollow {
		utils.PrintInfo(fmt.Sprintf("Following logs for %s (Ctrl+C to stop)...", appName))
		fmt.Println()
	}

	return stackpkg.PM2Logs(appName, dashboardLogsTail, dashboardLogsFollow)
}

func runDashboardMonit(cmd *cobra.Command, args []string) error {
	utils.PrintInfo("Opening PM2 monitoring interface (Ctrl+C to exit)...")
	fmt.Println()

	return stackpkg.PM2Monit()
}
