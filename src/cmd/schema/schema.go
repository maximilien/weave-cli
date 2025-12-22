// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package schema

import (
	"github.com/spf13/cobra"
)

// SchemaCmd represents the schema command group
var SchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage collection schemas",
	Long: `Manage collection schemas for vector databases.

The schema command group provides tools for analyzing documents,
suggesting optimal schemas, and managing collection configurations.`,
}

func init() {
	SchemaCmd.AddCommand(suggestCmd)
}
