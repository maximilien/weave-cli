// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"github.com/spf13/cobra"
)

// Cmd represents the config command
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `Manage Weave CLI configuration.

This command provides subcommands to view and manage configuration settings:
  • create/update - Interactively create or update .env and config.yaml files
  • update --weave-mcp - Download and install weave-mcp binary for REPL mode
  • sync - Sync local configuration files to global ~/.weave-cli directory
  • show - Display current configuration
  • list - List configured databases and schemas`,
}

func init() {
	// Register all subcommands
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(listSchemasCmd)
	Cmd.AddCommand(showSchemaCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(syncCmd)

	// Add flags for show-schema command
	showSchemaCmd.Flags().Bool("yaml", false, "Output schema as YAML")
	showSchemaCmd.Flags().Bool("json", false, "Output schema as JSON")
}
