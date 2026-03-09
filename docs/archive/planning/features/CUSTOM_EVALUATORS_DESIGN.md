# Custom Evaluator Definitions (Option 2)

**Date:** 2026-01-26
**Status:** 🚧 In Progress
**Priority:** High
**Estimated Time:** 4-6 hours

## Overview

Enable users to define custom evaluation metrics using YAML configuration files. This allows teams to create domain-specific evaluators tailored to their use cases without writing code.

## Use Cases

### 1. Domain-Specific Quality Checks
```yaml
name: medical_accuracy
description: Evaluates medical information accuracy and safety
prompt: |
  You are evaluating medical information for accuracy and safety.

  Query: {{.Query}}
  Response: {{.Answer}}

  Rate the medical accuracy on a scale of 0.0 (dangerous) to 1.0 (accurate).
  Consider:
  - Factual correctness
  - Safety implications
  - Current medical guidelines

  Provide only a number between 0.0 and 1.0.
threshold: 0.9
```

### 2. Brand Voice Compliance
```yaml
name: brand_voice
description: Checks if response matches brand voice guidelines
prompt: |
  Evaluate if this response matches our brand voice:
  - Professional but friendly
  - Clear and concise
  - Avoids jargon

  Response: {{.Answer}}

  Rate brand voice compliance: 0.0 (poor) to 1.0 (perfect).
threshold: 0.7
```

### 3. Regulatory Compliance
```yaml
name: gdpr_compliance
description: Checks for GDPR compliance in responses
prompt: |
  Check if this response complies with GDPR:
  - No personal data exposure
  - Proper data handling mentions
  - Privacy-aware language

  Response: {{.Answer}}

  Rate GDPR compliance: 0.0 (non-compliant) to 1.0 (compliant).
threshold: 0.95
required: true
```

## YAML Schema

### Basic Evaluator Definition

```yaml
# File: evals/evaluators/my_evaluator.yaml
name: my_evaluator
description: Short description of what this evaluates
version: 1.0.0

# Evaluation prompt template
prompt: |
  You are evaluating: {{.Description}}

  Query: {{.Query}}
  Expected Answer: {{.ExpectedAnswer}}
  Actual Answer: {{.Answer}}
  Context: {{.Context}}

  Rate on a scale of 0.0 to 1.0.
  Provide only a number.

# Scoring configuration
scoring:
  type: llm_judge      # or: regex, exact_match, contains
  threshold: 0.7       # Minimum passing score
  weight: 1.0          # Weight in overall evaluation (optional)

# Model configuration (optional - uses default if not specified)
model:
  provider: openai     # or: anthropic, local
  name: gpt-4o-mini
  temperature: 0.1

# Metadata
tags:
  - quality
  - domain-specific
author: team-name
required: false        # If true, evaluation must pass for test to pass
```

### Advanced Features

#### 1. Multiple Scoring Methods

```yaml
name: hybrid_evaluator
scoring:
  type: hybrid
  methods:
    - type: regex
      pattern: '\b(citation|reference|source)\b'
      weight: 0.3
    - type: llm_judge
      prompt: "Rate completeness..."
      weight: 0.7
  threshold: 0.8
```

#### 2. Conditional Evaluation

```yaml
name: context_dependent
conditions:
  - field: context
    operator: not_empty
    action: evaluate
  - field: context
    operator: empty
    action: skip        # Returns 1.0 if no context
scoring:
  type: llm_judge
  threshold: 0.7
```

#### 3. Multi-Step Evaluation

```yaml
name: comprehensive_check
steps:
  - name: factual_accuracy
    prompt: "Check facts..."
    weight: 0.4
  - name: completeness
    prompt: "Check completeness..."
    weight: 0.3
  - name: clarity
    prompt: "Check clarity..."
    weight: 0.3
threshold: 0.75
```

## Directory Structure

```
evals/
├── datasets/
│   ├── baseline.yaml
│   └── stress-test.yaml
├── evaluators/           # NEW
│   ├── README.md
│   ├── medical_accuracy.yaml
│   ├── brand_voice.yaml
│   ├── gdpr_compliance.yaml
│   └── examples/
│       ├── regex_example.yaml
│       ├── hybrid_example.yaml
│       └── multistep_example.yaml
└── results/
    └── ...
```

## Implementation Plan

### Phase 1: Core Schema & Loader (2 hours)

**New Files:**
1. `src/pkg/evaluation/custom_evaluator.go`
2. `src/pkg/evaluation/custom_evaluator_loader.go`
3. `src/pkg/evaluation/custom_evaluator_test.go`

**Structs:**

```go
// CustomEvaluatorDef defines a custom evaluator from YAML
type CustomEvaluatorDef struct {
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    Version     string            `yaml:"version"`
    Prompt      string            `yaml:"prompt"`
    Scoring     ScoringConfig     `yaml:"scoring"`
    Model       *ModelConfig      `yaml:"model,omitempty"`
    Tags        []string          `yaml:"tags,omitempty"`
    Author      string            `yaml:"author,omitempty"`
    Required    bool              `yaml:"required"`
}

type ScoringConfig struct {
    Type      string  `yaml:"type"`      // llm_judge, regex, exact_match, contains
    Threshold float64 `yaml:"threshold"`
    Weight    float64 `yaml:"weight"`
    Pattern   string  `yaml:"pattern,omitempty"`  // For regex type
}

type ModelConfig struct {
    Provider    string  `yaml:"provider"`
    Name        string  `yaml:"name"`
    Temperature float64 `yaml:"temperature"`
}

// CustomEvaluatorLoader loads evaluator definitions from YAML files
type CustomEvaluatorLoader struct {
    baseDir string
}

func NewCustomEvaluatorLoader(baseDir string) *CustomEvaluatorLoader
func (l *CustomEvaluatorLoader) Load(name string) (*CustomEvaluatorDef, error)
func (l *CustomEvaluatorLoader) LoadAll() ([]*CustomEvaluatorDef, error)
func (l *CustomEvaluatorLoader) Validate(def *CustomEvaluatorDef) error
```

### Phase 2: Evaluator Execution (1.5 hours)

**Executor:**

```go
// CustomEvaluator executes custom evaluator definitions
type CustomEvaluator struct {
    def       *CustomEvaluatorDef
    llmClient llm.Client
}

func NewCustomEvaluator(def *CustomEvaluatorDef, llmClient llm.Client) *CustomEvaluator

func (e *CustomEvaluator) Name() string {
    return e.def.Name
}

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

func (e *CustomEvaluator) evaluateLLMJudge(ctx context.Context, testCase *TestCase, actual string) (float64, error) {
    // Render prompt template with test case data
    prompt := e.renderPrompt(testCase, actual)

    // Call LLM
    response, err := e.llmClient.Complete(ctx, prompt)
    if err != nil {
        return 0.0, err
    }

    // Parse score
    score := parseScore(response)
    return score, nil
}

func (e *CustomEvaluator) renderPrompt(testCase *TestCase, actual string) string {
    // Use text/template to render prompt with variables
    tmpl := template.New("prompt")
    tmpl.Parse(e.def.Prompt)

    data := map[string]interface{}{
        "Query":          testCase.Query,
        "ExpectedAnswer": testCase.ExpectedAnswer,
        "Answer":         actual,
        "Context":        strings.Join(testCase.RetrievedContext, "\n"),
        "Description":    e.def.Description,
    }

    var buf bytes.Buffer
    tmpl.Execute(&buf, data)
    return buf.String()
}
```

### Phase 3: Provider Integration (1 hour)

**New Provider Type:**

```go
// CustomProvider uses user-defined custom evaluators
type CustomProvider struct {
    evaluators map[string]*CustomEvaluator
    llmClient  llm.Client
}

func NewCustomProvider(evalDir string, llmClient llm.Client) (*CustomProvider, error) {
    loader := NewCustomEvaluatorLoader(evalDir)
    defs, err := loader.LoadAll()
    if err != nil {
        return nil, err
    }

    evaluators := make(map[string]*CustomEvaluator)
    for _, def := range defs {
        evaluators[def.Name] = NewCustomEvaluator(def, llmClient)
    }

    return &CustomProvider{
        evaluators: evaluators,
        llmClient:  llmClient,
    }, nil
}

func (p *CustomProvider) GetEvaluator(name string) (Evaluator, error) {
    eval, ok := p.evaluators[name]
    if !ok {
        return nil, fmt.Errorf("custom evaluator not found: %s", name)
    }
    return eval, nil
}
```

**Dataset Enhancement:**

```yaml
# Dataset can specify which custom evaluators to use
name: medical-qa
version: 1.0.0
custom_evaluators:
  - medical_accuracy
  - gdpr_compliance
test_cases:
  - id: test-001
    query: "What is diabetes?"
    expected_answer: "..."
```

### Phase 4: CLI Integration (1 hour)

**Commands:**

```bash
# List available custom evaluators
weave eval list-evaluators

# Validate custom evaluator definition
weave eval validate-evaluator path/to/evaluator.yaml

# Run evaluation with custom evaluators
weave eval run --agent rag --dataset medical --custom-evaluators

# Create custom evaluator from template
weave eval create-evaluator medical_accuracy
```

**CLI Updates:**

```go
// In src/cmd/eval/run.go
var useCustomEvaluators bool
cmd.Flags().BoolVar(&useCustomEvaluators, "custom-evaluators", false, "Enable custom evaluators from dataset definition")

// New commands
func NewListEvaluatorsCommand() *cobra.Command
func NewValidateEvaluatorCommand() *cobra.Command
func NewCreateEvaluatorCommand() *cobra.Command
```

### Phase 5: Testing (1 hour)

**Test Files:**
1. `custom_evaluator_test.go` - Loader and validator tests
2. `custom_evaluator_executor_test.go` - Execution tests
3. `custom_evaluator_integration_test.go` - End-to-end tests

**Test Cases:**
- Load valid evaluator definition
- Validate schema compliance
- Execute LLM judge evaluator
- Execute regex evaluator
- Template rendering with test data
- Error handling (invalid YAML, missing fields)
- Integration with evaluation runner

### Phase 6: Documentation (30 min)

**Files to Create/Update:**
1. `evals/evaluators/README.md` - Guide for creating custom evaluators
2. `docs/CUSTOM_EVALUATORS.md` - Detailed documentation
3. `README.md` - Add custom evaluators section
4. Example evaluators in `evals/evaluators/examples/`

## Example: End-to-End Usage

### 1. Create Custom Evaluator

```yaml
# evals/evaluators/code_quality.yaml
name: code_quality
description: Evaluates code quality in responses
version: 1.0.0

prompt: |
  Evaluate the code quality in this response:

  Query: {{.Query}}
  Response: {{.Answer}}

  Consider:
  - Correctness
  - Readability
  - Best practices
  - Error handling

  Rate from 0.0 (poor) to 1.0 (excellent).
  Provide only a number.

scoring:
  type: llm_judge
  threshold: 0.75

tags:
  - code
  - quality
```

### 2. Reference in Dataset

```yaml
# evals/datasets/code-review.yaml
name: code-review
version: 1.0.0
custom_evaluators:
  - code_quality

test_cases:
  - id: test-001
    query: "Write a function to reverse a string"
    expected_answer: "def reverse(s): return s[::-1]"
```

### 3. Run Evaluation

```bash
weave eval run --agent code-assistant --dataset code-review --custom-evaluators

# Output:
# Provider:       local
# Custom Evals:   code_quality
#
# Test 1/5: test-001
#   Accuracy:     0.85 ✓
#   Citation:     0.60 ✓
#   Code Quality: 0.92 ✓  # Custom evaluator
#   Passed: ✓
```

## Benefits

### For Users
✅ **No Code Required** - Define evaluators in YAML
✅ **Domain-Specific** - Tailor to your use case
✅ **Version Controlled** - Evaluators in git alongside datasets
✅ **Shareable** - Easy to share evaluators across teams
✅ **Rapid Iteration** - Modify prompts without code changes

### For Teams
✅ **Standardization** - Consistent evaluation across projects
✅ **Compliance** - Enforce regulatory requirements
✅ **Quality Gates** - Automated quality checks in CI/CD
✅ **Documentation** - Self-documenting evaluation criteria

## Implementation Checklist

- [ ] Design YAML schema
- [ ] Implement evaluator loader
- [ ] Implement evaluator executor
- [ ] Add template rendering
- [ ] Support multiple scoring types
- [ ] Integrate with dataset loading
- [ ] Add CLI commands
- [ ] Write comprehensive tests
- [ ] Create example evaluators
- [ ] Update documentation
- [ ] Test end-to-end workflow

## Timeline

**Total: 6-7 hours**

- Phase 1: Core Schema & Loader (2h)
- Phase 2: Evaluator Execution (1.5h)
- Phase 3: Provider Integration (1h)
- Phase 4: CLI Integration (1h)
- Phase 5: Testing (1h)
- Phase 6: Documentation (0.5h)

## Next Steps

1. Create YAML schema and validation
2. Implement loader and parser
3. Build evaluator executor
4. Add template rendering
5. Integrate with evaluation system
6. Add tests and examples
7. Update CLI and documentation

---

**Status:** Ready to implement 🚀
