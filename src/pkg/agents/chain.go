// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"fmt"
	"time"
)

// AgentChain executes multiple agents in sequence
type AgentChain struct {
	agents []Agent
	config *ChainConfig
}

// ChainConfig configures chain execution behavior
type ChainConfig struct {
	// StopOnSuccess stops execution when an agent produces a valid response
	StopOnSuccess bool

	// FailFast stops execution on first agent error
	FailFast bool

	// PassContext passes previous agent output to next agent
	PassContext bool
}

// ChainResponse represents the response from an agent chain
type ChainResponse struct {
	// FinalOutput is the output from the last successful agent
	FinalOutput interface{}

	// AllOutputs contains outputs from all agents that executed
	AllOutputs []AgentOutput

	// ExecutionTime is total time for chain execution
	ExecutionTime time.Duration

	// AgentsExecuted is the number of agents that ran
	AgentsExecuted int
}

// AgentOutput represents output from a single agent in the chain
type AgentOutput struct {
	AgentName     string
	Output        interface{}
	Error         error
	ExecutionTime time.Duration
}

// NewAgentChain creates a new agent chain
func NewAgentChain(agents []Agent, config *ChainConfig) *AgentChain {
	if config == nil {
		config = &ChainConfig{
			StopOnSuccess: false,
			FailFast:      false,
			PassContext:   true,
		}
	}

	return &AgentChain{
		agents: agents,
		config: config,
	}
}

// Execute runs the agent chain
func (ac *AgentChain) Execute(ctx context.Context, input interface{}) (*ChainResponse, error) {
	if len(ac.agents) == 0 {
		return nil, fmt.Errorf("no agents in chain")
	}

	startTime := time.Now()
	response := &ChainResponse{
		AllOutputs: make([]AgentOutput, 0, len(ac.agents)),
	}

	var currentInput interface{} = input

	for i, agent := range ac.agents {
		agentStart := time.Now()

		// Execute agent
		output, err := agent.Execute(ctx, currentInput)
		agentDuration := time.Since(agentStart)

		// Record output
		agentOutput := AgentOutput{
			AgentName:     agent.Name(),
			Output:        output,
			Error:         err,
			ExecutionTime: agentDuration,
		}
		response.AllOutputs = append(response.AllOutputs, agentOutput)
		response.AgentsExecuted++

		// Handle errors
		if err != nil {
			if ac.config.FailFast {
				return response, fmt.Errorf("agent %s failed: %w", agent.Name(), err)
			}
			// Continue to next agent even on error
			continue
		}

		// Update final output
		response.FinalOutput = output

		// Check if we should stop
		if ac.config.StopOnSuccess && output != nil {
			// Agent produced valid output, stop here
			break
		}

		// Pass output to next agent if configured
		if ac.config.PassContext && i < len(ac.agents)-1 {
			currentInput = output
		}
	}

	response.ExecutionTime = time.Since(startTime)

	// If no agent succeeded, return error
	if response.FinalOutput == nil {
		return response, fmt.Errorf("no agent in chain produced valid output")
	}

	return response, nil
}

// GetAgentNames returns the names of all agents in the chain
func (ac *AgentChain) GetAgentNames() []string {
	names := make([]string, len(ac.agents))
	for i, agent := range ac.agents {
		names[i] = agent.Name()
	}
	return names
}
