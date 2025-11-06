// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/spf13/cobra"
)

// listCmd represents the config list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all configured databases",
	Long: `List all configured vector databases.

This command displays:
- All configured database names
- Database types
- Which database is the default`,
	Run: runList,
}

func runList(cobraCmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Configured Databases")
	fmt.Println()

	databaseNames := cfg.GetDatabaseNames()
	if len(databaseNames) == 0 {
		printWarning("No databases configured")
		return
	}

	for name, dbType := range databaseNames {
		isDefault := name == "default"
		if isDefault {
			color.New(color.FgGreen, color.Bold).Printf("• %s (default)\n", name)
		} else {
			fmt.Printf("• %s\n", name)
		}
		fmt.Printf("  Type: %s\n", dbType)
		fmt.Println()
	}
}
