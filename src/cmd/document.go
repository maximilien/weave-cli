package cmd

import (
	"github.com/maximilien/weave-cli/src/cmd/document"
	"github.com/spf13/cobra"
)

// documentCmd represents the document command
var documentCmd = &cobra.Command{
	Use:     "document",
	Aliases: []string{"doc", "docs"},
	Short:   "Document management",
	Long: `Manage documents in vector database collections.

This command provides subcommands to list, show, and delete documents.`,
}

func init() {
	rootCmd.AddCommand(documentCmd)
	
	// Add all document subcommands
	documentCmd.AddCommand(document.ListCmd)
	documentCmd.AddCommand(document.ShowCmd)
	documentCmd.AddCommand(document.CountCmd)
	documentCmd.AddCommand(document.CreateCmd)
	documentCmd.AddCommand(document.DeleteCmd)
	documentCmd.AddCommand(document.DeleteAllCmd)
}
