# Weave CLI v0.13.0 — Security and Quality

**Release date**: 2026-09-08  
**Git tag**: `v0.13.0`  
**Previous release**: `v0.12.3`

## Overview

v0.13.0 establishes a stronger quality baseline for Weave CLI. It updates the
Go toolchain and direct dependencies, clears all reachable vulnerability
findings, adds first-class coverage reporting, and raises unit-test statement
coverage from 16.13% to 25.28%.

The release also improves Opik observability and evaluation synchronization.
There are no intentional breaking CLI or configuration changes.

## Highlights

### Toolchain and dependency security

- Go 1.26.6 is used by local builds and every GitHub Actions workflow.
- Direct dependencies were refreshed to current compatible releases.
- Weaviate was migrated to the v5 Go client.
- OpenSearch request signing was migrated to AWS SDK v2.
- `govulncheck ./...` reports no reachable vulnerabilities.
- Cross-platform release builds retain an OCR-free path where native
  Tesseract and Leptonica libraries are unavailable.

### Coverage and test reliability

- `./test.sh --coverage` produces terminal, HTML, text, and raw Go coverage
  reports.
- Statement coverage increased from **16.13%** (`4,438 / 27,521`) to
  **25.28%** (`6,960 / 27,529`).
- Function coverage increased from **31.28%** (`661 / 2,113`) to
  **44.96%** (`950 / 2,113`).
- Codecov project coverage is ratcheted to 25%; patch coverage remains 80%.
- New deterministic tests cover executor cancellation and dispatch, REPL
  routing, evaluation and configuration commands, MCP JSON-RPC clients, stack
  checkpoints, and Weaviate HTTP/schema conversion.
- Tests no longer rely on ambient Opik credentials or developer home state.

### Opik observability and evaluation

- Query analysis, planning, step execution, tool calls, reporting, and
  evaluation emit richer spans.
- Traces are flushed during executor shutdown and retain their reporting
  context.
- Query runs create top-level Opik trace records.
- Evaluation datasets and experiments can synchronize to Opik.
- Stable dataset item reuse prevents duplicates across repeated demos.

## Coverage Campaign Results

| Milestone | Statements | Functions | Result |
| --- | ---: | ---: | --- |
| Baseline | 16.13% | 31.28% | Established reporting |
| Day 1 | 18.01% | 35.97% | Reusable packages |
| Day 2 | 20.18% | 38.29% | Executor and REPL |
| Day 3 | 23.27% | 42.64% | Commands and evaluation |
| Day 4 | 25.28% | 44.96% | MCP, stack, and Weaviate |

The next release milestone is v0.14.0 at 30% statement coverage. See the
[`coverage plan`](../tests/COVERAGE_PLAN.md) for the measured path forward.

## Validation

- [x] `./lint.sh`
- [x] `./build.sh`
- [x] `./test.sh --coverage`
- [x] No reachable `govulncheck` findings
- [x] Build, Test, Lint, and Security workflows green on the release commit
- [ ] Release workflow builds Linux AMD64/ARM64, macOS AMD64/ARM64, and
  Windows AMD64 artifacts
- [ ] Published checksums verified

## Installation

Download the appropriate binary and `checksums.txt` from the
[`v0.13.0` GitHub release](https://github.com/maximilien/weave-cli/releases/tag/v0.13.0),
or install from the Homebrew tap:

```bash
brew install Maximilien-ai/weave-cli/weave-cli
```

Source builds require Go 1.26.6 plus the native dependencies installed by
`./setup.sh`:

```bash
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
git checkout v0.13.0
./setup.sh
./build.sh
```

## Compatibility

- No configuration migration is required.
- Existing evaluation commands continue to work without Opik flags.
- Chroma remains available in supported source builds; portable release
  binaries use the repository's unsupported-platform fallback.

For the complete change list, see [`CHANGELOG.md`](../../CHANGELOG.md).
