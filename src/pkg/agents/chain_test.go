// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAgent for testing
type MockAgent struct {
	name          string
	shouldError   bool
	errorMessage  string
	outputValue   interface{}
	executeCount  int
	lastInput     interface{}
	executionTime time.Duration
}

func (m *MockAgent) Name() string {
	return m.name
}

func (m *MockAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	m.executeCount++
	m.lastInput = input

	// Simulate execution time
	if m.executionTime > 0 {
		time.Sleep(m.executionTime)
	}

	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMessage)
	}

	return m.outputValue, nil
}

func TestAgentChain_BasicExecution(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", outputValue: "output1"}
	agent2 := &MockAgent{name: "agent2", outputValue: "output2"}
	agent3 := &MockAgent{name: "agent3", outputValue: "output3"}

	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, nil)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "initial-input")

	require.NoError(t, err)
	assert.Equal(t, "output3", response.FinalOutput)
	assert.Equal(t, 3, response.AgentsExecuted)
	assert.Len(t, response.AllOutputs, 3)

	// Verify all agents executed
	assert.Equal(t, 1, agent1.executeCount)
	assert.Equal(t, 1, agent2.executeCount)
	assert.Equal(t, 1, agent3.executeCount)
}

func TestAgentChain_ContextPassing(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", outputValue: "step1"}
	agent2 := &MockAgent{name: "agent2", outputValue: "step2"}
	agent3 := &MockAgent{name: "agent3", outputValue: "step3"}

	config := &ChainConfig{
		PassContext: true,
	}
	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, config)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "initial")

	require.NoError(t, err)

	// Verify context passing
	assert.Equal(t, "initial", agent1.lastInput)
	assert.Equal(t, "step1", agent2.lastInput) // agent2 receives agent1's output
	assert.Equal(t, "step2", agent3.lastInput) // agent3 receives agent2's output
	assert.Equal(t, "step3", response.FinalOutput)
}

func TestAgentChain_StopOnSuccess(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", shouldError: true, errorMessage: "agent1 failed"}
	agent2 := &MockAgent{name: "agent2", outputValue: "success"}
	agent3 := &MockAgent{name: "agent3", outputValue: "should-not-execute"}

	config := &ChainConfig{
		StopOnSuccess: true,
		FailFast:      false,
	}
	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, config)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.NoError(t, err)
	assert.Equal(t, "success", response.FinalOutput)
	assert.Equal(t, 2, response.AgentsExecuted) // Only agent1 and agent2 executed

	// Verify agent3 was not executed
	assert.Equal(t, 0, agent3.executeCount)
}

func TestAgentChain_FailFast(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", outputValue: "output1"}
	agent2 := &MockAgent{name: "agent2", shouldError: true, errorMessage: "critical error"}
	agent3 := &MockAgent{name: "agent3", outputValue: "should-not-execute"}

	config := &ChainConfig{
		FailFast: true,
	}
	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, config)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent2 failed")
	assert.Contains(t, err.Error(), "critical error")
	assert.Equal(t, 2, response.AgentsExecuted)

	// Verify agent3 was not executed
	assert.Equal(t, 0, agent3.executeCount)
}

func TestAgentChain_ContinueOnError(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", shouldError: true, errorMessage: "agent1 failed"}
	agent2 := &MockAgent{name: "agent2", shouldError: true, errorMessage: "agent2 failed"}
	agent3 := &MockAgent{name: "agent3", outputValue: "final-success"}

	config := &ChainConfig{
		FailFast: false, // Continue on errors
	}
	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, config)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.NoError(t, err)
	assert.Equal(t, "final-success", response.FinalOutput)
	assert.Equal(t, 3, response.AgentsExecuted)

	// Verify errors were recorded
	assert.Error(t, response.AllOutputs[0].Error)
	assert.Error(t, response.AllOutputs[1].Error)
	assert.NoError(t, response.AllOutputs[2].Error)
}

func TestAgentChain_AllAgentsFail(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", shouldError: true, errorMessage: "error1"}
	agent2 := &MockAgent{name: "agent2", shouldError: true, errorMessage: "error2"}
	agent3 := &MockAgent{name: "agent3", shouldError: true, errorMessage: "error3"}

	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, nil)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no agent in chain produced valid output")
	assert.Equal(t, 3, response.AgentsExecuted)
}

func TestAgentChain_EmptyChain(t *testing.T) {
	chain := NewAgentChain([]Agent{}, nil)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no agents in chain")
	assert.Nil(t, response)
}

func TestAgentChain_SingleAgent(t *testing.T) {
	agent := &MockAgent{name: "single-agent", outputValue: "result"}
	chain := NewAgentChain([]Agent{agent}, nil)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.NoError(t, err)
	assert.Equal(t, "result", response.FinalOutput)
	assert.Equal(t, 1, response.AgentsExecuted)
}

func TestAgentChain_ExecutionTime(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", outputValue: "out1", executionTime: 10 * time.Millisecond}
	agent2 := &MockAgent{name: "agent2", outputValue: "out2", executionTime: 10 * time.Millisecond}

	chain := NewAgentChain([]Agent{agent1, agent2}, nil)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.NoError(t, err)
	assert.True(t, response.ExecutionTime >= 20*time.Millisecond)
	assert.True(t, response.AllOutputs[0].ExecutionTime >= 10*time.Millisecond)
	assert.True(t, response.AllOutputs[1].ExecutionTime >= 10*time.Millisecond)
}

func TestAgentChain_GetAgentNames(t *testing.T) {
	agent1 := &MockAgent{name: "rag-agent"}
	agent2 := &MockAgent{name: "search-agent"}
	agent3 := &MockAgent{name: "fallback-agent"}

	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, nil)

	names := chain.GetAgentNames()

	assert.Equal(t, []string{"rag-agent", "search-agent", "fallback-agent"}, names)
}

func TestChainConfig_DefaultValues(t *testing.T) {
	agent := &MockAgent{name: "agent", outputValue: "output"}
	chain := NewAgentChain([]Agent{agent}, nil)

	// Verify default config is applied
	assert.NotNil(t, chain.config)
	assert.False(t, chain.config.StopOnSuccess)
	assert.False(t, chain.config.FailFast)
	assert.True(t, chain.config.PassContext)
}

func TestChainResponse_AllOutputsRecorded(t *testing.T) {
	agent1 := &MockAgent{name: "agent1", outputValue: "out1"}
	agent2 := &MockAgent{name: "agent2", shouldError: true, errorMessage: "error2"}
	agent3 := &MockAgent{name: "agent3", outputValue: "out3"}

	config := &ChainConfig{FailFast: false}
	chain := NewAgentChain([]Agent{agent1, agent2, agent3}, config)

	ctx := context.Background()
	response, err := chain.Execute(ctx, "input")

	require.NoError(t, err)

	// Verify all outputs are recorded
	assert.Len(t, response.AllOutputs, 3)

	assert.Equal(t, "agent1", response.AllOutputs[0].AgentName)
	assert.Equal(t, "out1", response.AllOutputs[0].Output)
	assert.NoError(t, response.AllOutputs[0].Error)

	assert.Equal(t, "agent2", response.AllOutputs[1].AgentName)
	assert.Nil(t, response.AllOutputs[1].Output)
	assert.Error(t, response.AllOutputs[1].Error)

	assert.Equal(t, "agent3", response.AllOutputs[2].AgentName)
	assert.Equal(t, "out3", response.AllOutputs[2].Output)
	assert.NoError(t, response.AllOutputs[2].Error)
}
