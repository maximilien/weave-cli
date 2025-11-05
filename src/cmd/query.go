// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"github.com/maximilien/weave-cli/src/cmd/query"
)

func init() {
	// Register the query command with its 'q' alias
	rootCmd.AddCommand(query.NewQueryCommand())
}
