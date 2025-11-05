// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"time"
)

// Agent is the base interface for all agents
type Agent interface {
	// Name returns the agent's name
	Name() string

	// Execute processes input and returns output
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// ExecutionStep represents a single step in an execution plan
type ExecutionStep struct {
	Type        string                 `json:"type"` // "bash", "weave", "confirm"
	Command     string                 `json:"command"`
	Description string                 `json:"description"`
	Args        []string               `json:"args,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	DependsOn   []int                  `json:"depends_on,omitempty"` // Step indices
	Optional    bool                   `json:"optional"`
	Destructive bool                   `json:"destructive"`
}

// ExecutionPlan represents a complete execution plan
type ExecutionPlan struct {
	Steps       []ExecutionStep `json:"steps"`
	Summary     string          `json:"summary"`
	Warnings    []string        `json:"warnings,omitempty"`
	Estimations struct {
		Duration string `json:"duration"`
		Risk     string `json:"risk"` // "low", "medium", "high"
	} `json:"estimations"`
}

// QueryAgentInput is input for QueryAgent
type QueryAgentInput struct {
	Query string `json:"query"`
}

// QueryAgentOutput is output from QueryAgent
type QueryAgentOutput struct {
	IsWeaveQuery bool    `json:"is_weave_query"`
	FixedQuery   string  `json:"fixed_query"`
	Intent       string  `json:"intent"` // "list", "create", "query", "delete", etc.
	Confidence   float64 `json:"confidence"`
	Reason       string  `json:"reason"`
}

// PlanningAgentInput is input for PlanningAgent
type PlanningAgentInput struct {
	FixedQuery string `json:"fixed_query"`
	Intent     string `json:"intent"`
}

// WeaveAgentCommand represents a weave-cli command to execute
type WeaveAgentCommand struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
	Timeout   time.Duration          `json:"timeout,omitempty"`
}

// WeaveAgentResult represents result of weave command execution
type WeaveAgentResult struct {
	Success  bool          `json:"success"`
	Output   interface{}   `json:"output"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Retries  int           `json:"retries"`
}

// BashCommand represents a bash command to execute
type BashCommand struct {
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// BashResult represents result of bash command execution
type BashResult struct {
	Success  bool          `json:"success"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

// CommandReport represents a single command execution report
type CommandReport struct {
	Type     string        `json:"type"` // "bash", "weave"
	Command  string        `json:"command"`
	Success  bool          `json:"success"`
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// OperationReport represents a complete operation report
type OperationReport struct {
	QueryIntent     string          `json:"query_intent"`
	ExecutedSteps   int             `json:"executed_steps"`
	SuccessfulSteps int             `json:"successful_steps"`
	FailedSteps     int             `json:"failed_steps"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	Duration        time.Duration   `json:"duration"`
	Commands        []CommandReport `json:"commands"`
	Summary         string          `json:"summary"`
	Recommendations []string        `json:"recommendations,omitempty"`
	NextSteps       []string        `json:"next_steps,omitempty"`
}

// EvaluationMetrics represents evaluation metrics for a query
type EvaluationMetrics struct {
	QueryID          string        `json:"query_id"`
	Success          bool          `json:"success"`
	IntentMatched    bool          `json:"intent_matched"`
	LLMInvocations   int           `json:"llm_invocations"`
	TotalTokens      int           `json:"total_tokens"`
	PromptTokens     int           `json:"prompt_tokens"`
	CompletionTokens int           `json:"completion_tokens"`
	TotalCost        float64       `json:"total_cost"`
	Latency          time.Duration `json:"latency"`
	ErrorRate        float64       `json:"error_rate"`
	UserSatisfaction *float64      `json:"user_satisfaction,omitempty"`
}
