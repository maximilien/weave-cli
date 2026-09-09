# Go Coverage Plan

## Baseline and Goal

The unit-test baseline measured by `./test.sh --coverage` on 2026-09-04 is:

- Statements/lines: **16.13%** (`4,438 / 27,521`)
- Functions exercised: **31.28%** (`661 / 2,113`)
- Branches: not reported by the Go coverage tool

The long-term target is **80% statement coverage**. Because the current gap is
large, project coverage is ratcheted from the baseline in Codecov while patch
coverage targets 80%. Do not exclude meaningful production packages merely to
raise the percentage.

## Priorities

1. Cover untested execution and data paths: `pkg/executor`, `pkg/storage`,
   `pkg/mcp`, `pkg/mcpinstaller`, `pkg/health`, and `pkg/repl`.
2. Add command tests around argument validation and dependency boundaries for
   the zero-coverage `cmd/*` packages.
3. Raise high-impact core packages to 80%: `pkg/agents`, `pkg/config`,
   `pkg/evaluation`, `pkg/llm`, `pkg/pdf`, `pkg/pipeline`, and `pkg/stack`.
4. Exercise vector database adapters through deterministic fake HTTP/gRPC
   servers, emphasizing error handling, pagination, and response conversion.
5. Keep packages already above 80% from regressing: `embeddings`, `ollama`,
   `output`, `progress`, `ratelimit`, `version`, and `worker`.

## Near-Term Ratchet

The first coverage campaign begins on 2026-09-05. Its targets are **20%**, then
**25%**, then **30%** statement coverage. The 30% target is the first stretch
goal; reaching it does not replace the long-term 80% goal.

At the end of Day 5, the current source count makes these approximate
requirements:

| Target | Covered statements needed | Remaining after Day 5 |
| --- | ---: | ---: |
| 20% | 5,506 | Achieved (+1,700) |
| 25% | 6,883 | Achieved (+323) |
| 30% | 8,259 | 1,053 |

These counts are planning estimates. Use the percentage reported by
`./test.sh --coverage` because the denominator will change with production
code.

### Day 1 — Small Reusable Packages (2026-09-05)

- Add deterministic tests for `pkg/health`, `pkg/metrics`, `pkg/server`, and
  `pkg/mcpinstaller`.
- Begin `pkg/storage` tests around paths, metadata, thumbnails, and failures.
- Avoid live services, credentials, fixed home-directory paths, and sleeps.
- Exit target: at least 18% overall, with no package above 80% regressing.

#### Day 1 Results

- Statements/lines: **18.01%** (`4,956 / 27,525`), an increase of 1.88
  percentage points and 518 covered statements from the baseline.
- Functions exercised: **35.97%** (`760 / 2,113`), an increase of 4.69
  percentage points and 99 exercised functions from the baseline.
- Package results include `pkg/vectordb` at 100%, `pkg/metrics` at 100%,
  `pkg/health` at 96.2%, `pkg/mcpinstaller` at 78.7%, `pkg/storage` at 67.3%,
  `pkg/server` at 64.7%, `cmd/query` at 63.6%, and `pkg/executor` at 36.6%.
- Added failure-path coverage that exposed and fixed acceptance of failed HTTP
  responses by MCP installer downloads.
- The 18% exit target was achieved without regressing packages already above
  80%.

### Day 2 — Execution Boundaries (2026-09-06)

- Cover `pkg/executor` planning, dispatch, error propagation, and cancellation.
- Cover testable `pkg/repl` parsing and command-routing behavior.
- Introduce narrow interfaces or injected functions only where they improve
  production boundaries; do not add test-only production hooks.
- Exit target: reach and stabilize 20% overall.

#### Day 2 Results

- Statements/lines: **20.18%** (`5,556 / 27,529`), an increase of 2.17
  percentage points and 600 covered statements from Day 1.
- Functions exercised: **38.29%** (`809 / 2,113`), an increase of 2.32
  percentage points and 49 exercised functions from Day 1.
- Raised `pkg/executor` from 36.6% to 77.0% and `pkg/repl` from 0% to 88.4%.
- Added `cmd/chunking` helper and validation coverage, bringing the package to
  73.2% and providing margin above the project milestone.
- Made executor retry backoff context-aware so cancellation stops immediately
  instead of waiting for the next retry interval.
- The 20% exit target was achieved, so the Codecov project target advances to
  20%.

### Day 3 — Commands and Evaluation (2026-09-07)

- Test `cmd/eval`, `cmd/config`, and root command argument/flag validation.
- Use temporary configuration directories and fake providers.
- Prioritize invalid input, missing configuration, and provider failure paths.
- Exit target: at least 22.5% overall.

#### Day 3 Results

- Statements/lines: **23.27%** (`6,407 / 27,529`), an increase of 3.09
  percentage points and 851 covered statements from Day 2.
- Functions exercised: **42.64%** (`901 / 2,113`), an increase of 4.35
  percentage points and 92 exercised functions from Day 2.
- Raised `cmd/eval` from 0% to 63.0%, covering command contracts, datasets,
  evaluators, reports, benchmark helpers, and validation failures.
- Raised `cmd/config` from 2.9% to 15.9% with isolated environment-file,
  input, masking, database-display, and schema-formatting tests.
- Raised the root `cmd` package from 0% to 37.8% with flag, completion,
  configuration-failure, and usage-template coverage.
- The 22.5% exit target was exceeded with enough margin to absorb modest
  source growth before Day 4. Codecov remains at 20% until the 25% milestone.

### Day 4 — MCP, Stack, and Adapters (2026-09-08)

- Add protocol and response-conversion tests for `pkg/mcp`.
- Exercise `pkg/stack` rendering and orchestration without starting services.
- Add fake HTTP/gRPC coverage to the lowest vector database adapters.
- Exit target: reach and stabilize 25% overall.

#### Day 4 Results

- Statements/lines: **25.28%** (`6,960 / 27,529`), an increase of 2.01
  percentage points and 553 covered statements from Day 3.
- Functions exercised: **44.96%** (`950 / 2,113`), an increase of 2.32
  percentage points and 49 exercised functions from Day 3.
- Raised `pkg/mcp` from 0% to 76.1% with HTTP and stdio JSON-RPC lifecycle,
  authentication, malformed-response, cancellation, and validation coverage.
- Raised `pkg/stack` from 18.5% to 33.2% through isolated checkpoint,
  dependency, error-rendering, and ingestion-helper tests without starting
  services.
- Raised the Weaviate adapter from 5.6% to 16.7% with fake HTTP coverage for
  health and schema protocols plus nested schema and query-result conversion.
- The 25% exit target was achieved with 77 covered statements of margin, so
  the Codecov project target advances to 25%.

### Day 5 — Consolidate and Start 30% (2026-09-09)

- Fill branch and error-path gaps exposed by the coverage report.
- Remove flakes and duplicated fixtures introduced during the campaign.
- Select the next high-yield slices in `cmd/stack`, `cmd/eval`, `pkg/repl`, and
  the Weaviate adapter for the 30% campaign.
- Exit target: keep 25% green and publish the measured path to 30%.

#### Day 5 Results

- Statements/lines: **26.18%** (`7,206 / 27,529`), an increase of 0.90
  percentage points and 246 covered statements from Day 4.
- Functions exercised: **46.57%** (`984 / 2,113`), an increase of 1.61
  percentage points and 34 exercised functions from Day 4.
- Raised `cmd/stack` from 0% to 33.8% with isolated tests for all stack
  templates and runtime defaults, generated files, configuration validation,
  Cobra contracts, dashboard modes, and missing-stack or missing-config
  failures.
- No test starts Kubernetes, PM2, Milvus, or another external service. File
  generation and state checks run only in temporary working directories.
- The 25% ratchet remains green with 323 covered statements of margin. Reaching
  30% requires approximately 1,053 additional covered statements at the
  current source count.

### Path from 26.18% to 30%

1. Add fake HTTP coverage for Weaviate document and query protocols, including
   pagination, malformed responses, and server failures.
2. Cover validation and dependency boundaries in the zero-coverage
   `cmd/mcp`, `cmd/agents`, `cmd/vdb`, and `cmd/embeddings` packages.
3. Add deterministic protocol coverage to the OpenSearch and Elasticsearch
   adapters, then use `pkg/agents`, `pkg/config`, or `pkg/pipeline` to close
   any remaining gap.

Recalculate the remaining statement count after every slice. Prefer behavior
and failure modes that protect users over tests written only to move the
aggregate percentage.

## Milestone Policy

For each daily slice:

1. Run focused tests while developing.
2. Commit each coherent package or behavior separately.
3. Run `./lint.sh && ./build.sh && ./test.sh --coverage` before the day's
   consolidated push.
4. Record the new statement and function baseline in this document.
5. Confirm Build, Test, Lint, and Security workflows are green on `main`.

After 20%, 25%, or 30% is stable locally and on `main`, raise Codecov's project
target to that value in a separate commit. Never lower the target or exclude
meaningful production code to make a gate pass. Keep patch coverage at 80% so
new and changed behavior remains well tested throughout the campaign.

## Long-Term Milestones

- **40%**: cover core orchestration happy paths and validation failures.
- **60%**: bring every supported vector database adapter above 50%.
- **80%**: close command and integration-boundary gaps without test-only
  production hooks or broad coverage exclusions.
