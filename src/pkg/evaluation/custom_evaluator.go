// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"gopkg.in/yaml.v3"
)

// CustomEvaluatorDef defines a custom evaluator from YAML configuration
type CustomEvaluatorDef struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Version     string        `yaml:"version"`
	Prompt      string        `yaml:"prompt"`
	Scoring     ScoringConfig `yaml:"scoring"`
	Model       *ModelConfig  `yaml:"model,omitempty"`
	Tags        []string      `yaml:"tags,omitempty"`
	Author      string        `yaml:"author,omitempty"`
	Required    bool          `yaml:"required"`
}

// ScoringConfig defines how to score the evaluation
type ScoringConfig struct {
	Type      string  `yaml:"type"`      // llm_judge, regex, exact_match, contains
	Threshold float64 `yaml:"threshold"` // Minimum passing score
	Weight    float64 `yaml:"weight"`    // Weight in overall evaluation (default 1.0)
	Pattern   string  `yaml:"pattern,omitempty"` // For regex type
}

// ModelConfig specifies which LLM model to use
type ModelConfig struct {
	Provider    string  `yaml:"provider"`    // openai, anthropic
	Name        string  `yaml:"name"`        // model name
	Temperature float64 `yaml:"temperature"` // temperature for generation
}

// Validate checks if the evaluator definition is valid
func (d *CustomEvaluatorDef) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("evaluator name is required")
	}

	if d.Prompt == "" {
		return fmt.Errorf("evaluator prompt is required")
	}

	if d.Scoring.Type == "" {
		return fmt.Errorf("scoring type is required")
	}

	validTypes := map[string]bool{
		"llm_judge":   true,
		"regex":       true,
		"exact_match": true,
		"contains":    true,
	}

	if !validTypes[d.Scoring.Type] {
		return fmt.Errorf("invalid scoring type: %s (must be: llm_judge, regex, exact_match, contains)", d.Scoring.Type)
	}

	if d.Scoring.Type == "regex" && d.Scoring.Pattern == "" {
		return fmt.Errorf("regex scoring type requires pattern")
	}

	if d.Scoring.Threshold < 0.0 || d.Scoring.Threshold > 1.0 {
		return fmt.Errorf("threshold must be between 0.0 and 1.0, got: %.2f", d.Scoring.Threshold)
	}

	// Set default weight if not specified
	if d.Scoring.Weight == 0 {
		d.Scoring.Weight = 1.0
	}

	return nil
}

// CustomEvaluatorLoader loads custom evaluator definitions from YAML files
type CustomEvaluatorLoader struct {
	baseDir string
}

// NewCustomEvaluatorLoader creates a loader for custom evaluators
func NewCustomEvaluatorLoader(baseDir string) *CustomEvaluatorLoader {
	return &CustomEvaluatorLoader{
		baseDir: baseDir,
	}
}

// Load loads a single custom evaluator by name
func (l *CustomEvaluatorLoader) Load(name string) (*CustomEvaluatorDef, error) {
	// Try with .yaml extension
	path := filepath.Join(l.baseDir, name+".yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try with .yml extension
		path = filepath.Join(l.baseDir, name+".yml")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("custom evaluator not found: %s", name)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read evaluator file: %w", err)
	}

	var def CustomEvaluatorDef
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse evaluator YAML: %w", err)
	}

	if err := def.Validate(); err != nil {
		return nil, fmt.Errorf("invalid evaluator definition: %w", err)
	}

	return &def, nil
}

// LoadAll loads all custom evaluators from the directory
func (l *CustomEvaluatorLoader) LoadAll() ([]*CustomEvaluatorDef, error) {
	if _, err := os.Stat(l.baseDir); os.IsNotExist(err) {
		return []*CustomEvaluatorDef{}, nil
	}

	entries, err := os.ReadDir(l.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read evaluators directory: %w", err)
	}

	var evaluators []*CustomEvaluatorDef
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process YAML files
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		// Load evaluator
		name := strings.TrimSuffix(entry.Name(), ext)
		def, err := l.Load(name)
		if err != nil {
			// Log error but continue loading others
			fmt.Fprintf(os.Stderr, "Warning: failed to load evaluator %s: %v\n", name, err)
			continue
		}

		evaluators = append(evaluators, def)
	}

	return evaluators, nil
}

// GetDefaultCustomEvaluatorDir returns the default directory for custom evaluators
func GetDefaultCustomEvaluatorDir() string {
	// Check if we're in development mode (git repo present)
	if _, err := os.Stat(".git"); err == nil {
		return "evals/evaluators"
	}

	// Production: use ~/.weave-cli/evals/evaluators
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "evals/evaluators" // Fallback
	}

	return filepath.Join(homeDir, ".weave-cli", "evals", "evaluators")
}

// CustomEvaluator executes custom evaluator definitions
type CustomEvaluator struct {
	def       *CustomEvaluatorDef
	llmClient llm.Client
}

// NewCustomEvaluator creates a custom evaluator from a definition
func NewCustomEvaluator(def *CustomEvaluatorDef, llmClient llm.Client) *CustomEvaluator {
	return &CustomEvaluator{
		def:       def,
		llmClient: llmClient,
	}
}

// Name returns the evaluator name
func (e *CustomEvaluator) Name() string {
	return e.def.Name
}

// Evaluate executes the custom evaluator
func (e *CustomEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	switch e.def.Scoring.Type {
	case "llm_judge":
		return e.evaluateLLMJudge(ctx, testCase, actual)
	case "regex":
		return e.evaluateRegex(actual)
	case "exact_match":
		return e.evaluateExactMatch(testCase, actual)
	case "contains":
		return e.evaluateContains(testCase, actual)
	default:
		return 0.0, fmt.Errorf("unknown scoring type: %s", e.def.Scoring.Type)
	}
}

// evaluateLLMJudge uses LLM to judge the response
func (e *CustomEvaluator) evaluateLLMJudge(ctx context.Context, testCase *TestCase, actual string) (float64, error) {
	// Render prompt template
	prompt, err := e.renderPrompt(testCase, actual)
	if err != nil {
		return 0.0, fmt.Errorf("failed to render prompt: %w", err)
	}

	// Call LLM
	response, err := e.llmClient.Complete(ctx, prompt)
	if err != nil {
		return 0.0, fmt.Errorf("LLM evaluation failed: %w", err)
	}

	// Parse score from response
	score := parseScore(response)

	return score, nil
}

// evaluateRegex checks if the actual answer matches a regex pattern
func (e *CustomEvaluator) evaluateRegex(actual string) (float64, error) {
	pattern := e.def.Scoring.Pattern
	matched, err := regexp.MatchString(pattern, actual)
	if err != nil {
		return 0.0, fmt.Errorf("invalid regex pattern: %w", err)
	}

	if matched {
		return 1.0, nil
	}
	return 0.0, nil
}

// evaluateExactMatch checks for exact match with expected answer
func (e *CustomEvaluator) evaluateExactMatch(testCase *TestCase, actual string) (float64, error) {
	// Normalize whitespace
	expected := strings.TrimSpace(testCase.ExpectedAnswer)
	actual = strings.TrimSpace(actual)

	if expected == actual {
		return 1.0, nil
	}
	return 0.0, nil
}

// evaluateContains checks if actual contains expected answer
func (e *CustomEvaluator) evaluateContains(testCase *TestCase, actual string) (float64, error) {
	// Case-insensitive substring check
	expected := strings.ToLower(testCase.ExpectedAnswer)
	actual = strings.ToLower(actual)

	if strings.Contains(actual, expected) {
		return 1.0, nil
	}
	return 0.0, nil
}

// renderPrompt renders the prompt template with test case data
func (e *CustomEvaluator) renderPrompt(testCase *TestCase, actual string) (string, error) {
	tmpl, err := template.New("prompt").Parse(e.def.Prompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	data := map[string]interface{}{
		"Query":          testCase.Query,
		"ExpectedAnswer": testCase.ExpectedAnswer,
		"Answer":         actual,
		"Context":        strings.Join(testCase.RetrievedContext, "\n"),
		"Description":    e.def.Description,
		"Name":           e.def.Name,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
