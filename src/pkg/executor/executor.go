// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
)

// Executor orchestrates agent execution for queries
type Executor struct {
	queryAgent    *agents.QueryAgent
	planningAgent *agents.PlanningAgent
	weaveAgent    *agents.WeaveAgent
	bashAgent     *agents.BashAgent
	outputAgent   *agents.OutputAgent
	reportAgent   *agents.ReportAgent
	evalAgent     *agents.EvalAgent
	mcpClient     *mcp.Client
	llmClient     llm.Client
	config        *Config
}

// Config holds executor configuration
type Config struct {
	DryRun       bool
	NoConfirm    bool
	Verbose      bool
	Quiet        bool
	NoColor      bool
	OutputFormat string
	Model        string
	StdioPath    string
	MaxRetries   int
}

// NewExecutor creates a new Executor
func NewExecutor(config *Config) (*Executor, error) {
	// Initialize Opik tracing if enabled
	opikConfig := llm.LoadOpikConfig()
	ctx := context.Background()
	var tracerProvider interface{}
	if opikConfig.Enabled {
		tp, err := llm.InitOpikTracing(ctx, opikConfig)
		if err != nil {
			// Log warning but don't fail - Opik is optional
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize Opik tracing: %v\n", err)
		} else if tp != nil {
			tracerProvider = tp
			if config.Verbose {
				fmt.Fprintf(os.Stderr, "✓ Opik tracing enabled (workspace: %s)\n", opikConfig.Workspace)
			}
		}
	}

	// Initialize LLM client
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	var llmClient llm.Client
	var err error

	// If Opik is enabled, use HTTP client with OTEL instrumentation
	if tracerProvider != nil {
		httpClient := llm.WrapHTTPClient(nil) // Wrap default client
		llmClient, err = llm.NewOpenAIClientWithHTTP(openaiKey, httpClient)
	} else {
		llmClient, err = llm.NewOpenAIClient(openaiKey)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Initialize MCP client
	stdioPath := config.StdioPath
	if stdioPath == "" {
		stdioPath = os.Getenv("WEAVE_MCP_STDIO_PATH")
	}
	if stdioPath == "" {
		return nil, fmt.Errorf("WEAVE_MCP_STDIO_PATH must be configured")
	}

	mcpClient, err := mcp.NewClient(stdioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	// Get available MCP tools
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		_ = mcpClient.Close()
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}

	// Initialize agents
	outputConfig := agents.OutputConfig{
		NoColor:      config.NoColor,
		Verbose:      config.Verbose,
		Quiet:        config.Quiet,
		OutputFormat: config.OutputFormat,
	}

	queryAgent := agents.NewQueryAgent(llmClient)
	planningAgent := agents.NewPlanningAgentWithTools(llmClient, tools)
	weaveAgent := agents.NewWeaveAgent(mcpClient)
	bashAgent := agents.NewBashAgent()
	outputAgent := agents.NewOutputAgent(outputConfig)
	reportAgent := agents.NewReportAgent(outputAgent, llmClient)
	evalAgent := agents.NewEvalAgent(llmClient)

	// Set MCP tools on query agent for better validation
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Name
	}
	queryAgent.SetMCPTools(toolNames)

	// Set output agent on bash agent for user approvals
	bashAgent.SetOutputAgent(outputAgent)

	// Set verbose mode on weave agent for MCP debug logging
	weaveAgent.SetVerbose(config.Verbose)

	return &Executor{
		queryAgent:    queryAgent,
		planningAgent: planningAgent,
		weaveAgent:    weaveAgent,
		bashAgent:     bashAgent,
		outputAgent:   outputAgent,
		reportAgent:   reportAgent,
		evalAgent:     evalAgent,
		mcpClient:     mcpClient,
		llmClient:     llmClient,
		config:        config,
	}, nil
}

// Execute executes a query
func (e *Executor) Execute(ctx context.Context, query string) (*agents.OperationReport, error) {
	startTime := time.Now()

	// Step 1: Validate and fix query
	if !e.config.Quiet {
		fmt.Println()
		fmt.Print("🤖 Analyzing query...")
		os.Stdout.Sync()
	}
	queryInput := &agents.QueryAgentInput{Query: query}
	queryOutput, err := e.queryAgent.Execute(ctx, queryInput)
	if err != nil {
		if !e.config.Quiet {
			fmt.Println(" ✗")
		}
		return nil, fmt.Errorf("failed to analyze query: %w", err)
	}
	if !e.config.Quiet {
		fmt.Println(" ✓")
	}

	queryResult := queryOutput.(*agents.QueryAgentOutput)

	// Check if query is weave-related
	if !queryResult.IsWeaveQuery {
		e.outputAgent.PrintRejectionMessage(queryResult.Reason)
		return nil, fmt.Errorf("query is not weave-related")
	}

	// Step 2: Create execution plan
	if !e.config.Quiet {
		fmt.Print("📋 Creating execution plan...")
		os.Stdout.Sync()
	}
	planInput := &agents.PlanningAgentInput{
		FixedQuery: queryResult.FixedQuery,
		Intent:     queryResult.Intent,
	}

	planOutput, err := e.planningAgent.Execute(ctx, planInput)
	if err != nil {
		if !e.config.Quiet {
			fmt.Println(" ✗")
		}
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}
	if !e.config.Quiet {
		fmt.Println(" ✓")
	}

	plan := planOutput.(*agents.ExecutionPlan)

	// Display plan
	fmt.Println()
	e.outputAgent.PrintPlan(plan)

	// Dry run mode: just show the plan
	if e.config.DryRun {
		e.outputAgent.PrintInfo("Dry run mode: skipping execution")
		return nil, nil
	}

	// Ask for confirmation if needed
	if !e.config.NoConfirm && e.hasDestructiveOperations(plan) {
		confirmed, err := e.outputAgent.AskConfirmation("Proceed with execution?")
		if err != nil || !confirmed {
			e.outputAgent.PrintInfo("Operation cancelled")
			return nil, fmt.Errorf("operation cancelled by user")
		}
	}

	// Step 3: Execute the plan
	if !e.config.Quiet {
		fmt.Println()
		e.outputAgent.PrintInfo("⚡ Executing commands...")
		fmt.Println()
	}
	commandReports, err := e.executePlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to execute plan: %w", err)
	}

	// Step 4: Create report
	report := agents.CreateReport(queryResult.Intent, startTime, commandReports)

	// Step 5: Enhance report
	if _, err := e.reportAgent.Execute(ctx, report); err != nil {
		// Non-fatal error, continue
		e.outputAgent.PrintWarning(fmt.Sprintf("Failed to enhance report: %v", err))
	}

	// Step 6: Display report
	e.reportAgent.PrintReport(report)

	// Step 7: Evaluate and track metrics
	metrics, err := e.evalAgent.Execute(ctx, report)
	if err == nil {
		e.outputAgent.PrintMetrics(metrics.(*agents.EvaluationMetrics))
	}

	return report, nil
}

// DryRun shows the execution plan without executing
func (e *Executor) DryRun(ctx context.Context, query string) (*agents.ExecutionPlan, error) {
	// Step 1: Validate query
	queryInput := &agents.QueryAgentInput{Query: query}
	queryOutput, err := e.queryAgent.Execute(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze query: %w", err)
	}

	queryResult := queryOutput.(*agents.QueryAgentOutput)

	if !queryResult.IsWeaveQuery {
		e.outputAgent.PrintRejectionMessage(queryResult.Reason)
		return nil, fmt.Errorf("query is not weave-related")
	}

	// Step 2: Create execution plan
	planInput := &agents.PlanningAgentInput{
		FixedQuery: queryResult.FixedQuery,
		Intent:     queryResult.Intent,
	}

	planOutput, err := e.planningAgent.Execute(ctx, planInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}

	plan := planOutput.(*agents.ExecutionPlan)
	e.outputAgent.PrintPlan(plan)

	return plan, nil
}

// Close closes the executor and releases resources
func (e *Executor) Close() error {
	if e.mcpClient != nil {
		return e.mcpClient.Close()
	}
	return nil
}

// executePlan executes all steps in the plan
func (e *Executor) executePlan(ctx context.Context, plan *agents.ExecutionPlan) ([]agents.CommandReport, error) {
	reports := []agents.CommandReport{}
	stepOutputs := make(map[int]interface{}) // Store outputs for dependencies

	for i, step := range plan.Steps {
		// Check dependencies
		if len(step.DependsOn) > 0 {
			allDependenciesSuccess := true
			for _, depIdx := range step.DependsOn {
				if depIdx >= i {
					return reports, fmt.Errorf("invalid dependency: step %d depends on future step %d", i, depIdx)
				}
				if !reports[depIdx].Success && !step.Optional {
					allDependenciesSuccess = false
					break
				}
			}

			if !allDependenciesSuccess && !step.Optional {
				// Skip this step
				reports = append(reports, agents.CommandReport{
					Type:    step.Type,
					Command: step.Command,
					Success: false,
					Error:   "dependency failed",
				})
				continue
			}
		}

		// Execute step
		e.outputAgent.PrintStepProgress(i+1, &step, "running")
		report := e.executeStep(ctx, &step, stepOutputs)
		reports = append(reports, report)

		// Store output for dependent steps
		stepOutputs[i] = report.Output

		// Print completion status with duration
		if report.Success {
			e.outputAgent.PrintStepCompletion(i+1, report.Duration)
		} else {
			e.outputAgent.PrintStepError(i+1, report.Error)
		}

		// Print the command output immediately (streaming effect)
		if report.Success && report.Output != "" {
			e.outputAgent.PrintCommandOutput(report.Output)
		} else if report.Success && report.Output == "" {
			// Command succeeded but no output - this might indicate an issue
			if !e.config.Quiet {
				e.outputAgent.PrintWarning("Command completed but produced no output")
			}
		} else if !report.Success {
			e.outputAgent.PrintError(fmt.Sprintf("Failed: %s", report.Error))
		}
		fmt.Println() // Add spacing between steps

		// Stop on critical failure
		// TODO: Implement retry logic when MaxRetries > 0 and !report.Success && !step.Optional
	}

	return reports, nil
}

// executeStep executes a single step
func (e *Executor) executeStep(ctx context.Context, step *agents.ExecutionStep, stepOutputs map[int]interface{}) agents.CommandReport {
	startTime := time.Now()

	var result interface{}
	var err error

	switch step.Type {
	case "weave":
		result, err = e.executeWeaveStep(ctx, step)
	case "bash":
		result, err = e.executeBashStep(ctx, step)
	case "confirm":
		confirmed, confirmErr := e.outputAgent.AskConfirmation(step.Description)
		if confirmErr != nil || !confirmed {
			err = fmt.Errorf("user declined confirmation")
		}
	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	duration := time.Since(startTime)

	report := agents.CommandReport{
		Type:     step.Type,
		Command:  step.Command,
		Success:  err == nil,
		Duration: duration,
	}

	if err != nil {
		report.Error = err.Error()
	}

	if result != nil {
		// Format result as JSON if it's a map or struct
		switch v := result.(type) {
		case map[string]interface{}:
			if jsonBytes, err := json.MarshalIndent(v, "", "  "); err == nil {
				report.Output = string(jsonBytes)
			} else {
				report.Output = fmt.Sprintf("%v", result)
			}
		case string:
			report.Output = v
			// Check if the string output indicates an error
			if strings.HasPrefix(strings.ToLower(v), "error:") ||
				strings.HasPrefix(strings.ToLower(v), "failed to") ||
				strings.Contains(strings.ToLower(v), "error:") {
				report.Success = false
				report.Error = v
			}
		default:
			// Try to marshal as JSON
			if jsonBytes, err := json.MarshalIndent(result, "", "  "); err == nil {
				report.Output = string(jsonBytes)
			} else {
				report.Output = fmt.Sprintf("%v", result)
			}
		}
	}

	return report
}

// executeWeaveStep executes a weave MCP step
func (e *Executor) executeWeaveStep(ctx context.Context, step *agents.ExecutionStep) (interface{}, error) {
	cmd := &agents.WeaveAgentCommand{
		Tool:      step.Command,
		Arguments: step.Params,
		Timeout:   30 * time.Second,
	}

	result, err := e.weaveAgent.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}

	weaveResult := result.(*agents.WeaveAgentResult)
	if !weaveResult.Success {
		return nil, fmt.Errorf("%s", weaveResult.Error)
	}

	// Extract text content from MCP response format
	// MCP returns: [{text: "{json}", type: "text"}]
	output := e.extractMCPTextContent(weaveResult.Output)
	return output, nil
}

// extractMCPTextContent extracts the text field from MCP content array
func (e *Executor) extractMCPTextContent(output interface{}) interface{} {
	// MCP responses come as: [{text: "...", type: "text"}]
	if contentArray, ok := output.([]interface{}); ok && len(contentArray) > 0 {
		if contentMap, ok := contentArray[0].(map[string]interface{}); ok {
			if text, ok := contentMap["text"].(string); ok {
				// Try to parse as JSON
				var jsonData interface{}
				if err := json.Unmarshal([]byte(text), &jsonData); err == nil {
					return jsonData
				}
				// Return as string if not valid JSON
				return text
			}
		}
	}
	// Return original if we can't extract
	return output
}

// executeBashStep executes a bash step
func (e *Executor) executeBashStep(ctx context.Context, step *agents.ExecutionStep) (interface{}, error) {
	// Add weave binary directory to PATH if command uses weave
	command := step.Command
	if strings.Contains(command, "weave") {
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			// Only prepend if not already a simple weave command
			// This allows the command to work even when run directly
			command = fmt.Sprintf("export PATH=\"%s:$PATH\" && %s", exeDir, command)
		}
	}

	cmd := &agents.BashCommand{
		Command: command,
		Args:    step.Args,
		Timeout: 30 * time.Second,
	}

	result, err := e.bashAgent.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}

	bashResult := result.(*agents.BashResult)
	if !bashResult.Success {
		return nil, fmt.Errorf("exit code %d: %s", bashResult.ExitCode, bashResult.Stderr)
	}

	// Return stdout, but if it's empty and there's stderr, include stderr as a warning
	output := bashResult.Stdout
	if output == "" && bashResult.Stderr != "" {
		output = fmt.Sprintf("[stderr]: %s", bashResult.Stderr)
	}

	return output, nil
}

// hasDestructiveOperations checks if the plan has destructive operations
func (e *Executor) hasDestructiveOperations(plan *agents.ExecutionPlan) bool {
	for _, step := range plan.Steps {
		if step.Destructive {
			return true
		}
	}
	return false
}
