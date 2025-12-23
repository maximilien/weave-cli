// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chunking

import (
	"github.com/spf13/cobra"
)

// Cmd represents the chunking command
var Cmd = &cobra.Command{
	Use:   "chunking",
	Short: "AI-powered document chunking analysis and recommendations",
	Long: `Analyze documents and get AI-powered recommendations for optimal chunking strategies.

The chunking command helps you determine the best chunk size, overlap, and
strategy for your documents based on their structure and content type.

Examples:
  # Analyze documents and get chunking recommendations
  weave chunking suggest ./docs --collection my-docs

  # Analyze with specific requirements
  weave chunking suggest ./samples --collection tech-docs \
    --requirements "Technical documentation with code examples"

  # Output recommendations to file
  weave chunking suggest ./docs --collection articles --output chunking.yaml`,
}

func init() {
	Cmd.AddCommand(suggestCmd)
}
