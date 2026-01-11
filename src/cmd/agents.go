// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package cmd

import (
	"github.com/maximilien/weave-cli/src/cmd/agents"
)

func init() {
	// Register the agents command
	agentsCmd := agents.NewAgentsCommand()
	agentsCmd.GroupID = "ai"
	rootCmd.AddCommand(agentsCmd)
}
