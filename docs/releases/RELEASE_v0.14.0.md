# Weave CLI v0.14.0 — 30% Coverage Milestone

**Release date**: 2026-09-10  
**Git tag**: `v0.14.0`  
**Previous release**: `v0.13.0`

## Overview

v0.14.0 raises Weave CLI's deterministic unit-test coverage from 25.28% to
30.18%. The release focuses on user-facing command boundaries and Weaviate
protocol behavior without requiring credentials or live external services.
There are no intentional breaking CLI or configuration changes.

## Highlights

### Coverage and command reliability

- Statement coverage increased to **30.18%** (`8,307 / 27,529`).
- Function coverage increased to **49.27%** (`1,041 / 2,113`).
- Codecov project coverage is ratcheted to 30%; patch coverage remains 80%.
- `cmd/agents` reached 70.2% with isolated lifecycle, display, edit, copy,
  delete, validation, and command-contract tests.
- `cmd/embeddings` reached 93.9% with catalog, compatibility,
  recommendation, filtering, and credential tests.
- `cmd/mcp` reached 90.1% with deterministic JSON-RPC command tests.
- `cmd/vdb` reached 87.3% with configuration, filtering, helper, masking,
  and display tests.

### Weaviate protocol confidence

- The Weaviate adapter reached 44.7% statement coverage.
- Fake HTTP tests cover document create/get/delete, bulk deletion, counts,
  metadata schema discovery, and authentication.
- Query tests cover near-text, near-image, BM25, hybrid and simple fallbacks,
  filter encoding, result conversion, GraphQL failures, and transport errors.

## Coverage Campaign Results

| Milestone | Statements | Functions | Focus |
| --- | ---: | ---: | --- |
| v0.13.0 / Day 4 | 25.28% | 44.96% | MCP, stack, and Weaviate |
| Day 5 | 27.06% | 46.85% | Stack commands and documents |
| Day 6 | 28.49% | 47.52% | Weaviate queries and embeddings |
| v0.14.0 / Day 7 | 30.18% | 49.27% | Agent, MCP, and VDB commands |

## Validation

- [x] `./lint.sh`
- [x] `./build.sh`
- [x] `./test.sh --coverage`
- [x] No reachable `govulncheck` findings
- [ ] Build, Test, Lint, and Security workflows green on the release commit
- [ ] Release workflow builds Linux AMD64/ARM64, macOS AMD64/ARM64, and
  Windows AMD64 artifacts
- [ ] Published checksums verified

## Installation

Download the appropriate binary and `checksums.txt` from the
[`v0.14.0` GitHub release](https://github.com/maximilien/weave-cli/releases/tag/v0.14.0),
or install from the Homebrew tap:

```bash
brew install Maximilien-ai/weave-cli/weave-cli
```

Source builds require Go 1.26.6 plus the native dependencies installed by
`./setup.sh`:

```bash
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
git checkout v0.14.0
./setup.sh
./build.sh
```

## Compatibility

- No configuration migration is required.
- Existing MCP, agent, embedding, VDB, and query commands retain their CLI
  contracts.
- Chroma remains available in supported source builds; portable release
  binaries use the repository's unsupported-platform fallback.

For the complete change list, see [`CHANGELOG.md`](../../CHANGELOG.md).
