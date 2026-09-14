# Weave CLI v0.15.0 — 40% Coverage Milestone

**Release date**: 2026-09-13  
**Git tag**: `v0.15.0`  
**Previous release**: `v0.14.0`

## Overview

v0.15.0 raises Weave CLI's deterministic unit-test coverage from 30.18% to
40.06%. The release deepens command boundaries, vector database protocols,
pipeline processing, and agent orchestration without intentional breaking CLI
or configuration changes.

## Highlights

### Coverage and orchestration reliability

- Statement coverage increased to **40.06%** (`11,028 / 27,531`).
- Function coverage increased to **61.68%** (`1,304 / 2,114`).
- Codecov project coverage is ratcheted to 40%; patch coverage remains 80%.
- `pkg/pipeline` reached 84.9% with concurrent multi-format processing, resume,
  dry-run, batching, progress, embedding-failure, and database-failure tests.
- `pkg/agents` reached 81.7% with output, reporting, reasoning, command safety,
  chunking, configuration, and custom-agent discovery coverage.

### Database protocol confidence

- OpenSearch reached 69.7% and Elasticsearch reached 73.3% statement coverage.
- Loopback protocol suites exercise health, collection, document, bulk, schema,
  and query behavior without credentials or live services.
- Existing Weaviate document and query protocol coverage remains at 44.7%.

### Command and tooling quality

- Pipeline and serve command boundaries, document and schema displays, vector
  database selection, and batch-processing helpers have isolated tests.
- Native OCR flags no longer duplicate Tesseract and Leptonica linker entries
  on macOS.
- `weave serve` startup and shutdown failures propagate through Cobra.

## Coverage Campaign Results

| Milestone | Statements | Functions | Focus |
| --- | ---: | ---: | --- |
| v0.14.0 / Day 7 | 30.18% | 49.27% | Agent, MCP, and VDB commands |
| Day 8 | 35.04% | 53.50% | Commands, schemas, and batch utilities |
| Day 9 | 38.22% | 58.37% | OpenSearch, Elasticsearch, and agents |
| v0.15.0 / Day 10 | 40.06% | 61.68% | Pipeline and agent orchestration |

## Validation

- [x] `./lint.sh`
- [x] `./build.sh`
- [x] `./test.sh --coverage`
- [x] No reachable `govulncheck` findings
- [x] Build, Test, Lint, and Security workflows green on the release commit
- [x] Release workflow builds Linux AMD64/ARM64, macOS AMD64/ARM64, and
  Windows AMD64 artifacts
- [x] Published checksums verified

## Installation

Download the appropriate binary and `checksums.txt` from the
[`v0.15.0` GitHub release](https://github.com/maximilien/weave-cli/releases/tag/v0.15.0),
or install from the Homebrew tap:

```bash
brew install Maximilien-ai/weave-cli/weave-cli
```

Source builds require Go 1.26.6 plus the native dependencies installed by
`./setup.sh`:

```bash
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
git checkout v0.15.0
./setup.sh
./build.sh
```

## Compatibility

- No configuration migration is required.
- Existing MCP, agent, embedding, pipeline, VDB, and query commands retain
  their CLI contracts.
- Chroma remains available in supported source builds; portable release
  binaries use the repository's unsupported-platform fallback.

For the complete change list, see [`CHANGELOG.md`](../../CHANGELOG.md).
