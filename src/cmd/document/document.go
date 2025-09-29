// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"github.com/spf13/cobra"
)

// DocumentCmd represents the document command
var DocumentCmd = &cobra.Command{
	Use:     "document",
	Aliases: []string{"doc", "docs"},
	Short:   "Document management",
	Long: `Manage documents in vector database collections.

This command provides subcommands to list, show, and delete documents.`,
}

func init() {
	// Subcommands are added in their respective files
}
