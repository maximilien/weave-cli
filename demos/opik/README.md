# Weave CLI Demo — Opik (v0.12.2)

Self-contained demo directory for the Opik collaboration presentation.

## Contents

```
.
├── README.md        # This file
├── DEMO.md          # Full demo script with talking points
├── config.yaml      # VDB config (weaviate-cloud + milvus-local)
├── demo.sh          # Automated demo script (Enter-to-advance)
├── cleanup.sh       # Delete demo collections for a fresh start
└── milvus.sh        # Start/stop Milvus locally (-> ../shared/)
```

Shared resources (configs, docs, evals, milvus compose) live in `../shared/`.

## Quick Start

1. **Add your `.env`** with `WEAVIATE_URL`, `WEAVIATE_API_KEY`, `OPENAI_API_KEY`,
   and optionally `OPIK_API_KEY` / `OPIK_WORKSPACE`.

2. **Start Milvus** (needs Docker or Podman):

   ```bash
   ./milvus.sh start
   ./milvus.sh status   # wait until "running"
   ```

3. **Clean up any leftover collections** (optional, recommended before recording):

   ```bash
   ./cleanup.sh
   ```

4. **Run the demo**:

   ```bash
   ./demo.sh            # interactive, press Enter between sections
   ```

   Or follow `DEMO.md` manually command by command.

5. **After the demo**:

   ```bash
   ./cleanup.sh         # remove demo collections
   ./milvus.sh stop     # stop Milvus containers
   ```

## Prerequisites

- `weave-cli` installed (`brew install Maximilien-ai/weave-cli/weave-cli`)
- Docker or Podman (for Milvus local)
- `.env` file with valid API keys (gitignored)
