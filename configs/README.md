# Weave CLI Configuration Examples

This directory contains example configuration files for Weave CLI.

## Configuration Files

### AI Agent Configuration

- **`weave-agents.yaml`** - **NEW!** AI agent configuration for:
  - Schema suggestions (`weave schema suggest`)
  - Chunking recommendations (`weave chunking suggest`)
  - REPL mode agents
  - LLM models and parameters

### Vector Database Configurations

Example configurations for 10+ vector databases. See [VDB Support Matrix](../docs/VDB_SUPPORT_MATRIX.md).

## Quick Start

```bash
# 1. Interactive setup (recommended)
weave config create --env

# 2. Test connection
weave health check

# 3. Use AI features
weave schema suggest ./docs --collection MyDocs
weave chunking suggest ./docs --collection MyDocs
```

## Agent Configuration

Copy `weave-agents.yaml` to customize AI behavior:

```bash
# Local (project-specific)
cp configs/weave-agents.yaml ./weave-agents.yaml

# Global (user-wide)
cp configs/weave-agents.yaml ~/.weave-cli/weave-agents.yaml
```

## More Information

- **User Guide**: [../docs/USER_GUIDE.md](../docs/USER_GUIDE.md)
- **VDB Setup**: [../docs/](../docs/) (per-database directories)
