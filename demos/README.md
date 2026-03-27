# Weave CLI Demos

Self-contained demo environments and utility scripts for Weave CLI.

## Directory Structure

```
demos/
├── shared/              # Common resources used across demos
│   ├── configs/         # Agent configs (rag, qa, summarize)
│   ├── docs/            # Sample docs for ingestion
│   ├── evals/           # Eval datasets + evaluators
│   ├── local/milvus/    # Docker/Podman compose for Milvus
│   └── milvus.sh        # Milvus start/stop/status/clean
├── ai-alliance/         # AI Alliance presentation demo
│   ├── DEMO.md          # Step-by-step demo script
│   ├── config.yaml      # VDB config (weaviate-cloud + milvus-local)
│   ├── demo.sh          # Automated demo script
│   └── milvus.sh        # Wrapper -> ../shared/milvus.sh
├── opik/                # Opik collaboration demo
│   ├── DEMO.md          # Step-by-step demo script
│   ├── config.yaml      # VDB config (weaviate-cloud + milvus-local)
│   ├── demo.sh          # Automated demo script
│   └── milvus.sh        # Wrapper -> ../shared/milvus.sh
├── scripts/             # Utility/legacy demo scripts
│   ├── ai-alliance-demo.sh
│   ├── client0-demo.sh
│   ├── config-demo.sh
│   ├── embedding-comparison-demo.sh
│   ├── full-demo.sh
│   ├── oss-embeddings-demo.sh
│   ├── quick-demo.sh
│   ├── repl-demo.sh
│   ├── stack-demo.sh
│   └── supabase-demo.sh
└── DEMO_SCRIPT.md       # Generic manual demo script
```

## Running a Demo

Each demo directory (`ai-alliance/`, `opik/`) is self-contained.

### 1. Add your `.env`

Copy your credentials into the demo directory:

```bash
cp /path/to/.env demos/opik/.env
```

Required keys: `WEAVIATE_URL`, `WEAVIATE_API_KEY`, `OPENAI_API_KEY`.
Optional: `OPIK_API_KEY`, `OPIK_WORKSPACE`.

### 2. Start Milvus (if using milvus-local)

```bash
cd demos/opik
./milvus.sh start
```

### 3. Run the demo

```bash
# Interactive (press Enter between pages)
./demo.sh

# Or follow DEMO.md manually
```

### 4. Cleanup

```bash
./milvus.sh stop
```

## Creating a New Demo

1. Create a directory under `demos/` (e.g., `demos/my-demo/`)
2. Add `config.yaml` with your VDB configuration
3. Add `DEMO.md` with the step-by-step script
4. Create `demo.sh` referencing `../shared/` for common resources
5. Add `milvus.sh` wrapper if using Milvus local
6. `.env` files are gitignored -- users add their own

## Shared Resources

Files in `shared/` are used by all demos:

- **configs/**: Agent YAML definitions (rag-agent, qa-agent, summarize-agent)
- **docs/**: Sample documents for ingestion (ARCHITECTURE.md, ragme-io.pdf, dog.png)
- **evals/**: Evaluation datasets (baseline, technical-docs, etc.) and evaluators
- **local/milvus/**: Docker and Podman compose files for Milvus standalone
- **milvus.sh**: Start/stop/status/clean script for Milvus containers
