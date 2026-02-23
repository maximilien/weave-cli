// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

// ValidateCmd represents the stack validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate weave-stack.yaml configuration",
	Long: `Validate weave-stack.yaml for correctness.

This command checks:
- YAML syntax
- Required fields
- Valid values for runtime, vectordb type, etc.
- Resource specifications
- Collection schemas

Example:
  weave stack validate`,
	RunE: runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	utils.PrintInfo("Validating weave-stack.yaml...")
	fmt.Println()

	// Load and validate config
	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		utils.PrintError("❌ Validation failed:")
		utils.PrintError(fmt.Sprintf("   %v", err))
		return err
	}

	// Print validation summary
	utils.PrintSuccess("✅ YAML syntax: valid")
	utils.PrintSuccess(fmt.Sprintf("✅ Version: %s", config.Version))
	utils.PrintSuccess(fmt.Sprintf("✅ Name: %s", config.Name))
	utils.PrintSuccess(fmt.Sprintf("✅ Runtime: %s (%s)", config.Runtime.Kubernetes.Provider, config.Runtime.ContainerRuntime))
	utils.PrintSuccess(fmt.Sprintf("✅ VectorDB: %s v%s", config.Infrastructure.VectorDB.Type, config.Infrastructure.VectorDB.Version))

	if len(config.Collections) > 0 {
		utils.PrintSuccess(fmt.Sprintf("✅ Collections: %d defined", len(config.Collections)))
		for _, col := range config.Collections {
			fmt.Printf("   - %s (%s, %dd)\n", col.Name, col.Type, col.Schema.VectorDimensions)
		}
	}

	fmt.Println()
	utils.PrintSuccess("🎉 Configuration is valid!")

	return nil
}
