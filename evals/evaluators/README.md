# Custom Evaluators

Custom evaluators allow you to define domain-specific evaluation metrics using YAML configuration files. No code required!

## Quick Start

### 1. Create an Evaluator

```yaml
# evals/evaluators/my_evaluator.yaml
name: my_evaluator
description: Checks for specific quality criteria
version: 1.0.0

prompt: |
  Evaluate this response for quality:

  Query: {{.Query}}
  Expected: {{.ExpectedAnswer}}
  Actual: {{.Answer}}

  Rate from 0.0 (poor) to 1.0 (excellent).
  Provide only a number.

scoring:
  type: llm_judge
  threshold: 0.7

tags:
  - quality
author: your-team
```

### 2. Use in Dataset

```yaml
# evals/datasets/my_dataset.yaml
name: my-dataset
version: 1.0.0
custom_evaluators:
  - my_evaluator  # Reference your custom evaluator

test_cases:
  - id: test-001
    query: "What is AI?"
    expected_answer: "Artificial Intelligence"
```

### 3. Run Evaluation

```bash
weave eval run --agent rag-agent --dataset my-dataset --custom-evaluators
```

## YAML Schema

### Required Fields

- `name` - Unique evaluator identifier
- `prompt` - Evaluation prompt template (for llm_judge type)
- `scoring.type` - Scoring method
- `scoring.threshold` - Minimum passing score (0.0 - 1.0)

### Scoring Types

#### 1. LLM Judge (llm_judge)

Uses an LLM to evaluate responses.

```yaml
scoring:
  type: llm_judge
  threshold: 0.7
```

**Template Variables:**
- `{{.Query}}` - Test case query
- `{{.ExpectedAnswer}}` - Expected answer
- `{{.Answer}}` - Actual response
- `{{.Context}}` - Retrieved context
- `{{.Description}}` - Evaluator description
- `{{.Name}}` - Evaluator name

#### 2. Regex (regex)

Checks if response matches a pattern.

```yaml
scoring:
  type: regex
  pattern: '\b(citation|reference|source)\b'
  threshold: 0.5
```

Returns 1.0 if matches, 0.0 if not.

#### 3. Exact Match (exact_match)

Checks for exact match with expected answer.

```yaml
scoring:
  type: exact_match
  threshold: 0.9
```

Normalizes whitespace automatically.

#### 4. Contains (contains)

Checks if response contains expected answer.

```yaml
scoring:
  type: contains
  threshold: 0.5
```

Case-insensitive substring match.

### Optional Fields

```yaml
version: "1.0.0"           # Evaluator version
description: "..."         # Human-readable description
tags: [quality, domain]    # Tags for organization
author: "team-name"        # Evaluator author
required: false            # If true, must pass for test to pass
scoring:
  weight: 1.0             # Weight in overall evaluation
model:                     # Override default LLM
  provider: openai
  name: gpt-4o-mini
  temperature: 0.1
```

## Examples

See `examples/` directory for more examples:

- `examples/regex_example.yaml` - Pattern matching
- `examples/exact_match_example.yaml` - Exact matching
- `examples/contains_example.yaml` - Substring matching
- `examples/domain_specific.yaml` - Domain-specific checks

## Best Practices

### 1. Clear Prompts

```yaml
# Good: Specific criteria
prompt: |
  Check if the response:
  1. Answers the question
  2. Provides context
  3. Includes citations

  Rate 0.0-1.0 based on these criteria.

# Avoid: Vague criteria
prompt: "Is this good? Rate it."
```

### 2. Appropriate Thresholds

```yaml
# Strict quality gate
scoring:
  threshold: 0.9

# Moderate bar
scoring:
  threshold: 0.7

# Permissive check
scoring:
  threshold: 0.5
```

### 3. Meaningful Names

```yaml
# Good
name: medical_safety_check
name: brand_voice_compliance
name: citation_presence

# Avoid
name: check1
name: evaluator
```

### 4. Use Tags

```yaml
tags:
  - compliance      # Regulatory checks
  - quality         # Quality metrics
  - domain-medical  # Domain-specific
  - safety          # Safety checks
```

## Common Use Cases

### Medical/Healthcare

```yaml
name: medical_accuracy
description: Evaluates medical information accuracy
prompt: |
  Evaluate this medical response:

  {{.Answer}}

  Check for:
  - Factual correctness
  - Safety implications
  - Current guidelines compliance

  Rate: 0.0 (dangerous) to 1.0 (safe & accurate)
scoring:
  type: llm_judge
  threshold: 0.9
required: true
```

### Brand Voice

```yaml
name: brand_voice
description: Checks brand voice compliance
prompt: |
  Does this match our brand voice?
  - Professional but friendly
  - Clear and concise
  - Avoids jargon

  Response: {{.Answer}}

  Rate: 0.0 (poor) to 1.0 (perfect)
scoring:
  type: llm_judge
  threshold: 0.7
```

### Citation Check

```yaml
name: has_citations
description: Ensures response includes citations
scoring:
  type: regex
  pattern: '\[[0-9]+\]'  # Matches [1], [2], etc.
  threshold: 0.9
required: true
```

### Legal Compliance

```yaml
name: gdpr_compliance
description: Checks for GDPR compliance
prompt: |
  Evaluate GDPR compliance:
  - No personal data exposure
  - Privacy-aware language
  - Proper data handling mentions

  Response: {{.Answer}}

  Rate: 0.0 (non-compliant) to 1.0 (compliant)
scoring:
  type: llm_judge
  threshold: 0.95
required: true
```

## Troubleshooting

### Evaluator Not Found

```bash
Error: custom evaluator not found: my_evaluator
```

**Fix:** Check file exists at `evals/evaluators/my_evaluator.yaml`

### Invalid YAML

```bash
Error: failed to parse evaluator YAML
```

**Fix:** Validate YAML syntax with `weave eval validate-evaluator my_evaluator.yaml`

### Invalid Scoring Type

```bash
Error: invalid scoring type: my_type
```

**Fix:** Use one of: `llm_judge`, `regex`, `exact_match`, `contains`

### Threshold Out of Range

```bash
Error: threshold must be between 0.0 and 1.0
```

**Fix:** Set `threshold` between 0.0 and 1.0

## CLI Commands

```bash
# List available custom evaluators
weave eval list-evaluators

# Validate evaluator definition
weave eval validate-evaluator my_evaluator.yaml

# Create evaluator from template
weave eval create-evaluator my_evaluator

# Run evaluation with custom evaluators
weave eval run --agent rag --dataset test --custom-evaluators
```

## Directory Structure

```
evals/
├── datasets/
│   ├── baseline.yaml
│   └── my-dataset.yaml
├── evaluators/              # Your custom evaluators
│   ├── README.md           # This file
│   ├── my_evaluator.yaml
│   ├── brand_voice.yaml
│   └── examples/
│       ├── regex_example.yaml
│       ├── exact_match_example.yaml
│       └── domain_specific.yaml
└── results/
    └── ...
```

## Next Steps

1. Browse `examples/` for inspiration
2. Create your first custom evaluator
3. Test with a small dataset
4. Iterate based on results
5. Share evaluators with your team

---

**Need help?** See [Custom Evaluators Guide](../../docs/CUSTOM_EVALUATORS.md)
