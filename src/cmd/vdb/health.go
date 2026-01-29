// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vdb

import (
	"fmt"

	"github.com/spf13/cobra"
)

// HealthCmd represents the vdb health command
var HealthCmd = &cobra.Command{
	Use:   "health [DATABASE_NAME]",
	Short: "Check vector database health",
	Long: `Check the health status of vector databases.

If no database name is provided, checks all configured databases.
If a database name is provided, checks only that database.

Examples:
  weave vdb health                  # Check all databases
  weave vdb health weaviate-cloud   # Check specific database
  weave vdb health --cloud          # Check only cloud databases
  weave vdb health --local          # Check only local databases
  weave vdb health -t 5s            # Quick health check with 5s timeout`,
	Run: func(cmd *cobra.Command, args []string) {
		// This is essentially an alias to 'weave health check'
		// We can just print a friendly message directing users
		fmt.Println("\n💡 Tip: Use 'weave health check' for database health checks")
		fmt.Println("   The vdb health command will be fully implemented in the next release.")
		fmt.Println("\nFor now, try:")
		if len(args) > 0 {
			fmt.Printf("   weave health check %s\n", args[0])
		} else {
			fmt.Println("   weave health check")
		}
		fmt.Println()
	},
}

func init() {
	VDBCmd.AddCommand(HealthCmd)

	// Add flags for future compatibility
	HealthCmd.Flags().BoolP("summary", "S", false, "show summary table (default for multiple VDBs)")
	HealthCmd.Flags().Bool("details", false, "show detailed view (default for single VDB)")
	HealthCmd.Flags().Bool("cloud", false, "check only cloud databases")
	HealthCmd.Flags().Bool("local", false, "check only local databases")
}
