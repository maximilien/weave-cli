// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vdb

import (
	"github.com/spf13/cobra"
)

// VDBCmd represents the vdb command
var VDBCmd = &cobra.Command{
	Use:     "vdb",
	Aliases: []string{"db", "database"},
	Short:   "Vector database management",
	Long: `Manage vector database connections and configurations.

This command provides subcommands for working with vector databases:
- list: Show all configured databases
- health: Check database health status
- info: Show detailed information about a database

Examples:
  weave vdb list                    # List all configured databases
  weave vdb health                  # Check health of all databases
  weave vdb health weaviate-cloud   # Check health of specific database
  weave vdb info weaviate-cloud     # Show database details`,
}

func init() {
	// Subcommands are added in their respective files
}
