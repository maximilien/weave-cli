# Agent Management Guide

Complete guide to managing RAG agents in Weave CLI.

## Table of Contents

- [Overview](#overview)
- [Agent Types](#agent-types)
- [Commands](#commands)
- [Configuration Locations](#configuration-locations)
- [Quick Start](#quick-start)
- [Workflows](#workflows)
- [Customization](#customization)
- [Best Practices](#best-practices)

## Overview

Weave CLI provides a complete agent management system for creating, customizing, and managing RAG (Retrieval-Augmented Generation) agents. Agents process vector search results and generate intelligent responses with citations.

### Key Features

- **4 Built-in Templates**: RAG, QA, Summarize, Custom
- **Complete Lifecycle**: Create, edit, validate, copy, delete
- **Smart Defaults**: Automatic dev/deployed mode detection
- **Self-Contained**: No external files needed - templates embedded in binary
- **Validation**: Automatic config validation on create/edit
- **Flexible Deployment**: Works in development repos and deployed binaries

## Agent Types

### RAG Agent (General Purpose)

**Best for**: Comprehensive answers with citations from multiple sources

**Characteristics**:
- Temperature: 0.7 (balanced creativity)
- Max tokens: 2000 (detailed responses)
- Citations: Numeric format `[1]`, `[2]`
- Output: Markdown format
- Strict mode: Off (can synthesize information)

**Use cases**:
- Documentation Q&A
- Research assistance
- Knowledge base queries

**Create**:
```bash
weave agents create my-docs-agent --type rag
```

### QA Agent (Precise Answers)

**Best for**: Direct, fact-based answers strictly from sources

**Characteristics**:
- Temperature: 0.3 (lower for precision)
- Max tokens: 1000 (concise responses)
- Strict mode: On (source-only information)
- Fact checking: Enabled
- Response caching: Enabled

**Use cases**:
- Customer support (factual answers)
- Compliance queries
- Technical specifications

**Create**:
```bash
weave agents create support-agent --type qa
```

### Summarize Agent

**Best for**: Creating concise overviews of documents

**Characteristics**:
- Temperature: 0.5 (moderate)
- Max context chunks: 10 (more context for summaries)
- Output: Markdown with bullet points
- Focus: Key takeaways and main points

**Use cases**:
- Document summaries
- Meeting notes
- Report generation

**Create**:
```bash
weave agents create summarizer --type summarize
```

### Custom Agent

**Best for**: Specialized use cases requiring custom configuration

**Characteristics**:
- Minimal template
- All fields customizable
- Start from scratch

**Create**:
```bash
weave agents create my-custom-agent --type custom
```

## Commands

### `weave agents create`

Create a new agent from template.

**Syntax**:
```bash
weave agents create NAME --type TYPE [--output DIR] [--interactive]
```

**Options**:
- `--type`: Agent type (rag, qa, summarize, custom) - default: rag
- `--output`: Output directory - auto-detected based on mode
- `--interactive`: Interactive wizard for configuration

**Examples**:
```bash
# Create RAG agent (default type)
weave agents create medical-docs

# Create QA agent with specific type
weave agents create legal-qa --type qa

# Create with custom output directory
weave agents create my-agent --output ~/my-agents

# Interactive mode with prompts
weave agents create custom-agent --type custom --interactive
```

**Default Locations**:
- **Development mode** (in repo): `./configs/agents/`
- **Deployed mode** (binary): `~/.weave-cli/agents/`

### `weave agents list`

List all available agents.

**Syntax**:
```bash
weave agents list [--output FORMAT]
```

**Options**:
- `--output, -o`: Output format (text, json, yaml) - default: text

**Examples**:
```bash
# List in table format
weave agents list

# List as JSON
weave agents list --output json

# List as YAML
weave agents list -o yaml
```

**Output**:
```
NAME           TYPE  VERSION  DESCRIPTION
----           ----  -------  -----------
rag-agent      rag   1.0.0    General-purpose RAG agent with citations
qa-agent       qa    1.0.0    Precise Q&A agent for factual answers
medical-docs   rag   1.0.0    Medical documentation assistant

Total: 3 agents
```

### `weave agents show`

Show detailed information about an agent.

**Syntax**:
```bash
weave agents show NAME [--output FORMAT]
```

**Options**:
- `--output, -o`: Output format (text, json, yaml) - default: text

**Examples**:
```bash
# Show agent details
weave agents show rag-agent

# Show as JSON
weave agents show rag-agent --output json
```

**Output**:
```
Agent: rag-agent
Type: rag
Version: 1.0.0
Description: General-purpose RAG agent with citations
File Path: /Users/you/.weave-cli/agents/rag-agent.yaml

LLM Configuration:
  Provider: openai
  Model: gpt-4o
  Temperature: 0.70
  Max Tokens: 2000

Response Configuration:
  Include References: true
  Citation Format: numeric
  Max Context Chunks: 5
  Min Relevance Score: 0.30
  ...
```

### `weave agents validate`

Validate an agent configuration file.

**Syntax**:
```bash
weave agents validate FILE
```

**Examples**:
```bash
# Validate agent config
weave agents validate configs/agents/my-agent.yaml

# Validate after manual editing
weave agents validate ~/.weave-cli/agents/custom-agent.yaml
```

**Output** (success):
```
✅ Agent configuration is valid!

Agent Details:
  Name: my-agent
  Type: rag
  Version: 1.0.0
  Model: gpt-4o
```

**Output** (error):
```
❌ Validation failed: llm.temperature must be between 0.0 and 2.0
```

### `weave agents edit`

Edit an agent configuration in your default editor.

**Syntax**:
```bash
weave agents edit NAME
```

**Examples**:
```bash
# Edit agent (opens in $EDITOR, defaults to vim)
weave agents edit rag-agent

# Set custom editor
EDITOR=nano weave agents edit my-agent
```

**Behavior**:
1. Opens agent YAML in `$EDITOR` (vim, nano, code, etc.)
2. After saving, automatically validates the config
3. Reports validation errors if any

**Output**:
```
Validating edited configuration...
✅ Agent configuration is valid!
  Name: rag-agent
  Type: rag
  Model: gpt-4o
```

### `weave agents delete`

Delete an agent configuration.

**Syntax**:
```bash
weave agents delete NAME [--force]
```

**Options**:
- `--force`: Skip confirmation prompt

**Examples**:
```bash
# Delete with confirmation
weave agents delete old-agent

# Delete without confirmation
weave agents delete test-agent --force
```

**Protection**:
- Built-in agents (rag-agent, qa-agent, summarize-agent) cannot be deleted
- Prompts for confirmation unless `--force` is used

### `weave agents copy`

Copy an existing agent to create a variant.

**Syntax**:
```bash
weave agents copy SOURCE TARGET [--output DIR]
```

**Options**:
- `--output`: Output directory - default: same as create command

**Examples**:
```bash
# Copy built-in RAG agent
weave agents copy rag-agent my-rag-variant

# Copy to specific directory
weave agents copy qa-agent experimental-qa --output ~/experiments
```

**Use cases**:
- A/B testing different prompts
- Creating specialized variants
- Experimenting with temperature/parameters

## Configuration Locations

### Search Paths (Priority Order)

Weave CLI searches for agents in these directories:

1. **`./configs/agents/`** - Local (current directory)
   - For overrides and project-specific agents

2. **`~/.weave-cli/agents/`** - User home directory
   - Personal agents (deployed mode default)

3. **`/etc/weave-cli/agents/`** - System-wide (Unix-like systems)
   - Shared agents for all users

**View search paths**:
```bash
weave config show
```

### Development vs Deployed Mode

**Development Mode** (running from repo):
- Detected when `./configs/agents/rag-agent.yaml` exists
- Default create location: `./configs/agents/`
- Used for testing and development

**Deployed Mode** (installed binary):
- Running outside the repo
- Default create location: `~/.weave-cli/agents/`
- First-run initialization creates directory
- Self-contained (no external template files needed)

**Check current mode**:
```bash
# Development mode shows this default:
weave agents create --help
# → Output directory (default "configs/agents")

# Deployed mode shows this default:
weave agents create --help
# → Output directory (default "/Users/you/.weave-cli/agents")
```

## Quick Start

### 1. List Available Agents

```bash
weave agents list
```

### 2. Create Your First Agent

```bash
# Create a RAG agent for documentation
weave agents create docs-agent --type rag
```

### 3. View Agent Details

```bash
weave agents show docs-agent
```

### 4. Test the Agent

```bash
# Use with vector search
weave collections query my-docs "What is vector search?" --agent docs-agent
```

### 5. Customize the Agent

```bash
# Edit configuration
weave agents edit docs-agent

# Modify:
# - temperature (creativity vs precision)
# - max_tokens (response length)
# - system_prompt (behavior/instructions)
# - min_relevance_score (filter threshold)
```

### 6. Validate Changes

```bash
weave agents validate ~/.weave-cli/agents/docs-agent.yaml
```

## Workflows

### Workflow 1: Creating a Specialized Agent

```bash
# 1. Start with RAG template
weave agents create medical-agent --type rag

# 2. Customize for medical domain
weave agents edit medical-agent

# Edit system_prompt to add:
# "You are a medical information assistant. Always prioritize patient safety
# and recommend consulting healthcare professionals for medical decisions."

# 3. Test
weave collections query medical-docs "symptoms of flu" --agent medical-agent

# 4. Iterate based on results
weave agents edit medical-agent
```

### Workflow 2: A/B Testing Different Configurations

```bash
# 1. Create two variants
weave agents copy rag-agent variant-a
weave agents copy rag-agent variant-b

# 2. Modify variant-a (higher temperature for creative responses)
weave agents edit variant-a
# Set temperature: 0.9

# 3. Modify variant-b (lower temperature for precise responses)
weave agents edit variant-b
# Set temperature: 0.3

# 4. Compare results
weave collections query docs "explain vector databases" --agent variant-a
weave collections query docs "explain vector databases" --agent variant-b

# 5. Delete the variant that performs worse
weave agents delete variant-b
```

### Workflow 3: Domain-Specific QA Agent

```bash
# 1. Create QA agent
weave agents create legal-qa --type qa

# 2. Customize for legal domain
weave agents edit legal-qa

# Modify:
# - system_prompt: Add legal disclaimers
# - min_relevance_score: 0.7 (higher threshold for accuracy)
# - strict_mode: true (source-only answers)

# 3. Validate
weave agents validate ~/.weave-cli/agents/legal-qa.yaml

# 4. Use in production
weave collections query legal-docs "contract termination clauses" --agent legal-qa
```

## Customization

### Agent Configuration Structure

```yaml
name: my-agent
type: rag
description: Custom agent description
version: 1.0.0

llm:
  provider: openai
  model: gpt-4o
  temperature: 0.7        # 0.0 = deterministic, 2.0 = very creative
  max_tokens: 2000        # Maximum response length
  top_p: 1.0              # Nucleus sampling (0.0 - 1.0)

system_prompt: |
  Your agent's behavior instructions here.
  Define its role, style, and constraints.

response:
  include_references: true
  citation_format: numeric     # numeric, author-year, footnote
  max_context_chunks: 5        # How many sources to use
  min_relevance_score: 0.3     # Filter threshold (0.0 - 1.0)
  strict_mode: false           # true = source-only answers

output:
  format: markdown             # markdown, text, json
  show_sources: true
  truncate_sources: 500        # Characters per source (0 = no truncation)

performance:
  timeout_seconds: 60
  max_retries: 2
```

### Key Parameters to Tune

**Temperature** (`llm.temperature`):
- `0.0 - 0.3`: Precise, deterministic (QA, technical docs)
- `0.4 - 0.7`: Balanced (general RAG)
- `0.8 - 2.0`: Creative, varied (brainstorming, summaries)

**Max Context Chunks** (`response.max_context_chunks`):
- `3-5`: Focused answers
- `5-10`: Comprehensive answers
- `10+`: Summaries, deep analysis

**Min Relevance Score** (`response.min_relevance_score`):
- `0.2 - 0.4`: Broad search (more results)
- `0.5 - 0.7`: Balanced (recommended)
- `0.8+`: Strict matching (high precision)

**Strict Mode** (`response.strict_mode`):
- `false`: Can synthesize and infer (RAG)
- `true`: Source-only information (QA, compliance)

## Best Practices

### 1. Start with Built-in Templates

Don't create from scratch - copy and customize:
```bash
weave agents copy rag-agent my-specialized-agent
weave agents edit my-specialized-agent
```

### 2. Use Descriptive Names

```bash
# Good
weave agents create medical-diagnosis-assistant --type rag
weave agents create legal-contract-qa --type qa

# Avoid
weave agents create agent1 --type rag
weave agents create test --type qa
```

### 3. Version Your Agents

When making significant changes:
```bash
weave agents copy current-agent current-agent-v2
weave agents edit current-agent-v2
# Test v2, then delete v1 if successful
```

### 4. Validate After Every Edit

```bash
weave agents edit my-agent
weave agents validate ~/.weave-cli/agents/my-agent.yaml
```

### 5. Document Custom System Prompts

Add comments in your system prompts:
```yaml
system_prompt: |
  # Purpose: Medical information assistant
  # Updated: 2026-01-23
  # Changes: Added patient safety disclaimers

  You are a medical information assistant...
```

### 6. Use Source Control for Custom Agents

```bash
# Add custom agents to git
git add configs/agents/my-project-agent.yaml
git commit -m "Add project-specific RAG agent"
```

### 7. Test Before Production

```bash
# Test with sample queries
weave collections query test-docs "test query" --agent new-agent

# Compare with existing agent
weave collections query test-docs "test query" --agent rag-agent
```

### 8. Monitor Performance

After deployment, review:
- Response quality
- Citation accuracy
- Response time
- Token usage (costs)

Adjust `temperature`, `max_tokens`, `max_context_chunks` as needed.

---

## Next Steps

- Read [RAG Agent Evaluation](./planning/AGENT_EVAL_UPDATES.md) for testing agents
- See [Configuration Guide](./CONFIGURATION.md) for global settings
- Check [Main README](../README.md) for usage examples
