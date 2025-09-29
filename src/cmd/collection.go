package cmd

import (
	"github.com/maximilien/weave-cli/src/cmd/collection"
	"github.com/spf13/cobra"
)

// collectionCmd represents the collection command
var collectionCmd = &cobra.Command{
	Use:     "collection",
	Aliases: []string{"col", "cols"},
	Short:   "Collection management",
	Long: `Manage vector database collections.

This command provides subcommands to list, view, and delete collections.`,
}

func init() {
	rootCmd.AddCommand(collectionCmd)

	// Add all collection subcommands
	collectionCmd.AddCommand(collection.ListCmd)
	collectionCmd.AddCommand(collection.ShowCmd)
	collectionCmd.AddCommand(collection.CountCmd)
	collectionCmd.AddCommand(collection.CreateCmd)
	collectionCmd.AddCommand(collection.DeleteCmd)
	collectionCmd.AddCommand(collection.DeleteAllCmd)
	collectionCmd.AddCommand(collection.DeleteSchemaCmd)
}
