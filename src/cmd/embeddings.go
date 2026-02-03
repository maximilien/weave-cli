// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"github.com/maximilien/weave-cli/src/cmd/embeddings"
	"github.com/spf13/cobra"
)

// embeddingsCmd represents the embeddings command
var embeddingsCmd = &cobra.Command{
	Use:     "embeddings",
	Aliases: []string{"emb", "embed", "embeds", "vectorizer", "vectorizers"},
	Short:   "Embeddings management",
	Long: `Manage and list available embedding models.

This command provides information about embedding models available for text and image vectorization.

Note: You can also use 'vectorizer' or 'vectorizers' as aliases for this command.`,
}

func init() {
	rootCmd.AddCommand(embeddingsCmd)
	embeddingsCmd.GroupID = "data"

	// Add all embeddings subcommands
	embeddingsCmd.AddCommand(embeddings.ListCmd)
}
