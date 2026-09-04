# Release v0.12.4 - Opik Demo & Observability Patch

**Target Release Date**: 2026-09-04
**Git Tag**: `v0.12.4`
**Status**: Release candidate pending CI and final Opik smoke test

---

## Overview

v0.12.4 is a point release focused on the Opik demo, observability flow,
dependency security, and test visibility.
It improves monitoring visibility for natural-language query runs, adds dataset
and experiment sync for evaluation demos, and cleans up demo ergonomics so the
Opik walkthrough can be recorded and shared as a coherent feature story.

This release is backward compatible and is intended to make the following flow
work cleanly:

1. Run a natural-language query
2. Inspect spans and trace-level execution in Opik
3. Upload an evaluation dataset to Opik
4. Run a single experiment in Opik
5. Run multiple experiments and compare them side-by-side

---

## What's New

### Opik Monitoring for Query Runs

Natural-language query execution now emits richer Opik telemetry across the
full executor path:

- Query analysis span
- Planning span
- Per-step execution spans
- Tool spans for bash and MCP/weave execution
- Report generation span
- Eval/metrics span
- LLM spans with prompt/response, token counts, cost, and latency

Additional work was added to ensure traces are flushed on shutdown and that
query runs create top-level Opik trace records, not just span rows.

**Impact**:
- Easier to demonstrate the full app flow in Opik
- Better debugging for latency and orchestration issues
- More complete cost/token visibility during demos and development

---

### Opik Dataset Upload and Experiment Sync

Evaluation flows can now sync their artifacts to Opik:

- `weave eval datasets upload-opik <dataset>`
- `weave eval run --use-opik`
- `weave eval benchmark --use-opik`

The sync path now:

- Uploads local dataset samples to Opik
- Reuses stable dataset item IDs to avoid duplicate samples across reruns
- Creates experiment records in Opik
- Uploads experiment items and feedback scores
- Marks experiments complete for side-by-side comparison in the Opik UI

**Impact**:
- Datasets, versions, and experiment results are visible in Opik
- Benchmark demos can compare multiple runs on the same dataset
- Rerunning a demo no longer appends duplicate dataset samples

---

### Demo and Environment Updates

The Opik demo assets were updated so the monitoring and evaluation story is
self-contained:

- `demos/opik/demo.sh`
- `demos/opik/DEMO.md`
- `.env.example`

The demo now explicitly walks through:

- Monitoring query traces in Opik
- Uploading a dataset to Opik
- Running a single experiment
- Running a multi-experiment benchmark

Environment examples now include:

- `OPIK_WORKSPACE`
- `OPIK_PROJECT_NAME`
- `OTEL_EXPORTER_OTLP_ENDPOINT`

---

## Fixes Included

### Go Toolchain and Dependency Security

- Update the supported Go toolchain and CI matrix to Go 1.26.6
- Upgrade outdated direct libraries to current compatible releases
- Migrate Weaviate from the v4 client to v5
- Migrate OpenSearch request signing to AWS SDK v2
- Resolve all reachable findings reported by `govulncheck`
- Retain `chroma-go` v0.2.5 because v0.4.1's transitive
  `chroma-go-local` artifact failed public checksum verification

### Test and Coverage Reliability

- Add `./test.sh --coverage` with terminal, HTML, text, and Codecov reports
- Establish a 16.15% statement-coverage baseline and staged 80% plan
- Require 80% patch coverage while ratcheting whole-project coverage
- Isolate tests from ambient Opik credentials and developer home directories
- Remove a probabilistic frequency assertion that caused intermittent CI
  failures

### Trace Export Reliability

- Flush Opik traces when the executor shuts down
- Preserve trace context through the reporting path

### Opik API Integration Fixes

- Correct Opik API base URL derivation for dataset and experiment sync
- Create top-level Opik trace records for query runs

### Repository Hygiene

- Ignore generated evaluation result files:
  - `evals/results/run-*.json`

---

## Key Commands

### Monitoring

```bash
./bin/weave doctor --section opik
./bin/weave query "show me all collections and count the documents in each one"
```

### Dataset Upload

```bash
./bin/weave eval datasets upload-opik baseline
```

### Single Experiment

```bash
./bin/weave eval run --agent rag-agent --dataset baseline --use-opik
```

### Multi-Experiment Benchmark

```bash
./bin/weave eval benchmark --agents rag-agent,qa-agent --dataset baseline --use-opik
```

---

## Validation Checklist

Before tagging `v0.12.4`, confirm:

- [ ] `./lint.sh` passes
- [ ] `./build.sh` passes
- [ ] `./test.sh --coverage` passes and does not regress from 16.15%
- [ ] `govulncheck ./...` reports no reachable vulnerabilities
- [ ] GitHub branch CI is green on Go 1.26.6
- [ ] Release artifacts build for Linux, macOS, and Windows

- [ ] Query runs appear in Opik `Logs -> Traces`, not only `Spans`
- [ ] Dataset upload succeeds in Opik
- [ ] Single experiment appears in Opik with results
- [ ] Benchmark creates multiple experiments for comparison
- [ ] Valid `OPENAI_API_KEY` is configured for evaluation runs
- [ ] Demo dry run completes with the rebuilt local binary

---

## Compatibility

- No breaking CLI changes
- No config migration required
- Existing evaluation commands continue to work without `--use-opik`
- Source builds now require Go 1.26.6 or later

---

## Recommended Audience

Upgrade to v0.12.4 if you want:

- A cleaner Opik demo flow
- Better trace visibility for query execution
- Opik-backed dataset and experiment walkthroughs
- Side-by-side benchmark demos in Opik
