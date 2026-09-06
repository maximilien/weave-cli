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
	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Executor orchestrates agent execution for queries
type Executor struct {
	queryAgent     agents.Agent
	planningAgent  agents.Agent
	weaveAgent     agents.Agent
	bashAgent      *agents.BashAgent
	outputAgent    *agents.OutputAgent
	reportAgent    reportAgent
	evalAgent      agents.Agent
	mcpClient      *mcp.Client
	llmClient      llm.Client
	tracerProvider *sdktrace.TracerProvider
	config         *Config
}

type reportAgent interface {
	agents.Agent
	PrintReport(*agents.OperationReport)
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
	var tracerProvider *sdktrace.TracerProvider
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
		queryAgent:     queryAgent,
		planningAgent:  planningAgent,
		weaveAgent:     weaveAgent,
		bashAgent:      bashAgent,
		outputAgent:    outputAgent,
		reportAgent:    reportAgent,
		evalAgent:      evalAgent,
		mcpClient:      mcpClient,
		llmClient:      llmClient,
		tracerProvider: tracerProvider,
		config:         config,
	}, nil
}

// Execute executes a query
func (e *Executor) Execute(ctx context.Context, query string) (*agents.OperationReport, error) {
	startTime := time.Now()
	ctx, rootSpan := llm.StartSpan(ctx, "weave-cli-executor", "executor.execute", "general", map[string]interface{}{
		"query": query,
	}, map[string]interface{}{
		"component": "executor",
	})
	traceID := rootSpan.SpanContext().TraceID().String()
	traceMetadata := map[string]interface{}{
		"component": "executor",
		"source":    "weave-cli-query",
	}
	if _, err := evaluation.CreateTraceInOpik(ctx, traceID, "weave query", startTime, map[string]interface{}{
		"query": query,
	}, traceMetadata); err != nil && e.config.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to create Opik trace record: %v\n", err)
	}

	// Step 1: Validate and fix query
	if !e.config.Quiet {
		fmt.Println()
		fmt.Print("🤖 Analyzing query...")
		os.Stdout.Sync()
	}
	queryInput := &agents.QueryAgentInput{Query: query}
	queryCtx, querySpan := llm.StartSpan(ctx, "weave-cli-executor", "query-agent.execute", "general", queryInput, map[string]interface{}{
		"agent": e.queryAgent.Name(),
	})
	queryOutput, err := e.queryAgent.Execute(queryCtx, queryInput)
	llm.FinishSpan(querySpan, queryOutput, err)
	if err != nil {
		if !e.config.Quiet {
			fmt.Println(" ✗")
		}
		finalErr := fmt.Errorf("failed to analyze query: %w", err)
		_ = evaluation.UpdateTraceInOpik(ctx, traceID, time.Now(), nil, traceMetadata, finalErr)
		llm.FinishSpan(rootSpan, nil, finalErr)
		return nil, finalErr
	}
	if !e.config.Quiet {
		fmt.Println(" ✓")
	}

	queryResult := queryOutput.(*agents.QueryAgentOutput)

	// Check if query is weave-related
	if !queryResult.IsWeaveQuery {
		e.outputAgent.PrintRejectionMessage(queryResult.Reason)
		finalErr := fmt.Errorf("query is not weave-related")
		_ = evaluation.UpdateTraceInOpik(ctx, traceID, time.Now(), map[string]interface{}{
			"reason": queryResult.Reason,
		}, traceMetadata, finalErr)
		llm.FinishSpan(rootSpan, map[string]interface{}{"reason": queryResult.Reason}, finalErr)
		return nil, finalErr
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

	planCtx, planSpan := llm.StartSpan(ctx, "weave-cli-executor", "planning-agent.execute", "general", planInput, map[string]interface{}{
		"agent": e.planningAgent.Name(),
	})
	planOutput, err := e.planningAgent.Execute(planCtx, planInput)
	llm.FinishSpan(planSpan, planOutput, err)
	if err != nil {
		if !e.config.Quiet {
			fmt.Println(" ✗")
		}
		finalErr := fmt.Errorf("failed to create execution plan: %w", err)
		_ = evaluation.UpdateTraceInOpik(ctx, traceID, time.Now(), nil, traceMetadata, finalErr)
		llm.FinishSpan(rootSpan, nil, finalErr)
		return nil, finalErr
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
		_ = evaluation.UpdateTraceInOpik(ctx, traceID, time.Now(), map[string]interface{}{
			"dry_run": true,
			"plan":    plan,
		}, traceMetadata, nil)
		llm.FinishSpan(rootSpan, map[string]interface{}{"plan": plan, "dry_run": true}, nil)
		return nil, nil
	}

	// Ask for confirmation if needed
	if !e.config.NoConfirm && e.hasDestructiveOperations(plan) {
		confirmed, err := e.outputAgent.AskConfirmation("Proceed with execution?")
		if err != nil || !confirmed {
			e.outputAgent.PrintInfo("Operation cancelled")
			finalErr := fmt.Errorf("operation cancelled by user")
			_ = evaluation.UpdateTraceInOpik(ctx, traceID, time.Now(), nil, traceMetadata, finalErr)
			llm.FinishSpan(rootSpan, nil, finalErr)
			return nil, finalErr
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
		finalErr := fmt.Errorf("failed to execute plan: %w", err)
		_ = evaluation.UpdateTraceInOpik(ctx, traceID, time.Now(), nil, traceMetadata, finalErr)
		llm.FinishSpan(rootSpan, nil, finalErr)
		return nil, finalErr
	}

	// Step 4: Create report
	report := agents.CreateReport(queryResult.Intent, startTime, commandReports)

	// Step 5: Enhance report
	reportCtx, reportSpan := llm.StartSpan(ctx, "weave-cli-executor", "report-agent.execute", "general", report, map[string]interface{}{
		"agent": e.reportAgent.Name(),
	})
	if _, err := e.reportAgent.Execute(reportCtx, report); err != nil {
		// Non-fatal error, continue
		e.outputAgent.PrintWarning(fmt.Sprintf("Failed to enhance report: %v", err))
		llm.FinishSpan(reportSpan, report, err)
	} else {
		llm.FinishSpan(reportSpan, report, nil)
	}

	// Step 6: Display report
	e.reportAgent.PrintReport(report)

	// Step 7: Evaluate and track metrics
	evalCtx, evalSpan := llm.StartSpan(ctx, "weave-cli-executor", "eval-agent.execute", "general", report, map[string]interface{}{
		"agent": e.evalAgent.Name(),
	})
	metrics, err := e.evalAgent.Execute(evalCtx, report)
	llm.FinishSpan(evalSpan, metrics, err)
	if err == nil {
		e.outputAgent.PrintMetrics(metrics.(*agents.EvaluationMetrics))
	}
	traceMetadata["successful_steps"] = report.SuccessfulSteps
	traceMetadata["failed_steps"] = report.FailedSteps
	_ = evaluation.UpdateTraceInOpik(ctx, traceID, time.Now(), map[string]interface{}{
		"report":  report,
		"metrics": metrics,
	}, traceMetadata, nil)

	llm.FinishSpan(rootSpan, report, nil,
		attribute.Int("executor.successful_steps", report.SuccessfulSteps),
		attribute.Int("executor.failed_steps", report.FailedSteps),
		attribute.Float64("executor.duration_ms", float64(time.Since(startTime).Milliseconds())),
	)

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
	if e.tracerProvider != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := llm.ShutdownOpikTracing(shutdownCtx, e.tracerProvider); err != nil {
			return err
		}
	}
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

		// Implement retry logic for failed non-optional steps
		if !report.Success && !step.Optional && e.config.MaxRetries > 0 {
			retryCount := 0
			for retryCount < e.config.MaxRetries {
				retryCount++

				// Exponential backoff: 2^retry seconds (1s, 2s, 4s, etc.)
				backoffDuration := time.Duration(1<<uint(retryCount-1)) * time.Second
				fmt.Printf("\n⚠️  Step failed. Retrying in %v (attempt %d/%d)...\n",
					backoffDuration, retryCount, e.config.MaxRetries)

				retryTimer := time.NewTimer(backoffDuration)
				select {
				case <-retryTimer.C:
				case <-ctx.Done():
					if !retryTimer.Stop() {
						select {
						case <-retryTimer.C:
						default:
						}
					}
					return reports, fmt.Errorf("retry cancelled: %w", ctx.Err())
				}

				// Retry the step
				retryReport := e.executeStep(ctx, &step, stepOutputs)
				if retryReport.Success {
					fmt.Printf("✅ Retry succeeded!\n")
					report = retryReport
					reports[i] = retryReport
					stepOutputs[i] = retryReport.Output
					break
				}

				if retryCount < e.config.MaxRetries {
					fmt.Printf("❌ Retry %d failed: %s\n", retryCount, retryReport.Error)
				} else {
					fmt.Printf("❌ All retries exhausted. Step failed permanently.\n")
					report = retryReport
					reports[i] = retryReport
				}
			}
		}
	}

	return reports, nil
}

// executeStep executes a single step
func (e *Executor) executeStep(ctx context.Context, step *agents.ExecutionStep, stepOutputs map[int]interface{}) agents.CommandReport {
	startTime := time.Now()
	stepCtx, stepSpan := llm.StartSpan(ctx, "weave-cli-executor", "executor.step", "tool", step, map[string]interface{}{
		"step_type": step.Type,
	})

	var result interface{}
	var err error

	switch step.Type {
	case "weave":
		result, err = e.executeWeaveStep(stepCtx, step)
	case "bash":
		result, err = e.executeBashStep(stepCtx, step)
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

	llm.FinishSpan(stepSpan, report, err)

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
