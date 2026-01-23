// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package cmd

import (
	"github.com/maximilien/weave-cli/src/cmd/eval"
)

func init() {
	// Register the eval command
	evalCmd := eval.NewEvalCommand()
	evalCmd.GroupID = "ai"
	rootCmd.AddCommand(evalCmd)
}
