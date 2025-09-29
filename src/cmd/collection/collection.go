// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"github.com/spf13/cobra"
)

// CollectionCmd represents the collection command
var CollectionCmd = &cobra.Command{
	Use:     "collection",
	Aliases: []string{"col", "cols"},
	Short:   "Collection management",
	Long: `Manage vector database collections.

This command provides subcommands to list, view, and delete collections.`,
}

func init() {
	// Subcommands are added in their respective files
}
