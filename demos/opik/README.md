# Weave CLI Demo — AI Alliance (v0.12.0)

Self-contained demo directory for the AI Alliance presentation.

## Contents

```
.
├── README.md                  # This file
├── DEMO.md                    # Full demo script with talking points
├── config.yaml                # Weave CLI config (weaviate-cloud + milvus-local)
├── .env                       # API keys (weaviate, openai, opik)
├── milvus.sh                  # Start/stop Milvus locally
├── configs/
│   ├── weave-agents.yaml      # Agent framework settings
│   └── agents/                # Individual agent configs
│       ├── rag-agent.yaml
│       ├── qa-agent.yaml
│       └── summarize-agent.yaml
├── evals/
│   ├── datasets/              # Evaluation datasets (baseline, simple-qa, etc.)
│   └── evaluators/            # Custom evaluators (technical_accuracy)
└── local/
    └── milvus/                # Docker/Podman compose files for Milvus
        ├── docker-compose.yml
        └── podman-compose.yml
```

## Quick Start

1. **Start Milvus** (needs Docker or Podman):

   ```bash
   ./milvus.sh start
   ```

2. **Verify setup**:

   ```bash
   weave doctor
   weave health check --cloud
   ```

3. **Run the demo** — follow `DEMO.md` step by step.

4. **After the demo**:

   ```bash
   # Cleanup collections (see DEMO.md Cleanup section)
   ./milvus.sh stop
   ```

## Prerequisites

- `weave-cli` installed (`brew install Maximilien-ai/weave-cli/weave-cli`)
- Docker or Podman (for Milvus local — no separate Milvus install needed,
  `./milvus.sh start` pulls and runs the containers automatically)
- `.env` file with valid `WEAVIATE_URL`, `WEAVIATE_API_KEY`, and `OPENAI_API_KEY`
