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

The first coverage campaign began on 2026-09-05. Its targets were **20%**,
**25%**, and **30%** statement coverage. The v0.15.0 campaign adds **35%** and
**38%** checkpoints before **40%**; these milestones do not replace the
long-term 80% goal.

At the end of Day 15, the current source count makes these approximate
requirements:

| Target | Covered statements needed | Day 15 status |
| --- | ---: | ---: |
| 20% | 5,506 | Achieved (+2,801) |
| 25% | 6,883 | Achieved (+1,424) |
| 30% | 8,259 | Achieved (+48) |
| 35% | 9,636 | Achieved (+11) |
| 38% | 10,462 | Achieved (+61) |
| 40% | 11,013 | Achieved (+15) |
| 42% | 11,564 | Achieved (+22) |
| 45% | 12,385 | Achieved (+78) |
| 46% | 12,660 | Achieved (+243) |
| 50% | 13,761 | Achieved (+16) |
| 54% | 14,872 | Achieved (+88) |

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

- Statements/lines: **27.06%** (`7,450 / 27,529`), an increase of 1.78
  percentage points and 490 covered statements from Day 4.
- Functions exercised: **46.85%** (`990 / 2,113`), an increase of 1.89
  percentage points and 40 exercised functions from Day 4.
- Raised `cmd/stack` from 0% to 33.8% with isolated tests for all stack
  templates and runtime defaults, generated files, configuration validation,
  Cobra contracts, dashboard modes, and missing-stack or missing-config
  failures.
- Raised the Weaviate adapter from 16.7% to 30.8% with fake HTTP coverage for
  document creation, retrieval, deletion, bulk deletion, aggregation counts,
  metadata schema discovery, authentication, malformed responses, and server
  errors.
- No test starts Kubernetes, PM2, Milvus, or another external service. File
  generation and state checks run only in temporary working directories.
- The 25% ratchet remains green with 567 covered statements of margin. Reaching
  30% requires approximately 809 additional covered statements at the
  current source count.

### Day 6 — Queries and Command Catalogs (2026-09-10)

- Continue fake HTTP coverage for Weaviate query fallback, filtering, and
  result-conversion paths.
- Begin the zero-coverage command packages with the deterministic embeddings
  catalog and display paths.
- Exit target: exceed 28% overall and leave a measured final slice to 30%.

#### Day 6 Results

- Statements/lines: **28.49%** (`7,844 / 27,529`), an increase of 1.43
  percentage points and 394 covered statements from Day 5.
- Functions exercised: **47.52%** (`1,004 / 2,113`), an increase of 0.67
  percentage points and 14 exercised functions from Day 5.
- Raised the Weaviate adapter from 30.8% to 44.7% with fake HTTP coverage for
  near-text, near-image, BM25, hybrid and simple fallbacks, filter encoding,
  result conversion, malformed schemas, GraphQL failures, and transport
  errors.
- Raised `cmd/embeddings` from 0% to 93.9% with catalog integrity, API-key,
  database filtering, compatibility display, and collection recommendation
  coverage.
- Tests remain isolated from live services, credentials, and developer home
  state. The 25% ratchet now has 961 covered statements of margin.
- Reaching 30% requires approximately 415 additional covered statements at
  the current source count.

### Day 7 — Command Boundaries and 30% (2026-09-10)

- Promote the isolated agent lifecycle suite into the default unit lane and
  cover command construction, display, editing, copying, and deletion.
- Cover MCP validation, JSON-RPC command flows, argument parsing, and result
  rendering with a deterministic local HTTP server.
- Cover VDB command contracts, helpers, filtering, and configuration display.
- Exit target: reach and stabilize 30% overall.

#### Day 7 Results

- Statements/lines: **30.18%** (`8,307 / 27,529`), an increase of 1.69
  percentage points and 463 covered statements from Day 6.
- Functions exercised: **49.27%** (`1,041 / 2,113`), an increase of 1.75
  percentage points and 37 exercised functions from Day 6.
- Raised `cmd/agents` from 0% to 70.2%, `cmd/mcp` from 0% to 90.1%, and
  `cmd/vdb` from 0% to 87.3%.
- All promoted and new unit tests use temporary directories, isolated
  configuration, or a local fake server; none requires credentials or a live
  external service.
- The 30% milestone is achieved with 48 covered statements of margin, so the
  Codecov project target advances to 30% for v0.14.0.

### Day 8 — Command Completion and 35% (2026-09-11)

- Cover the remaining pipeline and serve command boundaries.
- Exercise document display, aggregation, schema conversion/export, vector
  database selection, and batch-processing helpers without live services.
- Remove duplicate native OCR linker flags from local quality scripts.
- Exit target: reach and stabilize 35% overall.

#### Day 8 Results

- Statements/lines: **35.04%** (`9,647 / 27,531`), an increase of 4.86
  percentage points and 1,340 covered statements from Day 7.
- Functions exercised: **53.50%** (`1,131 / 2,114`), an increase of 4.23
  percentage points and 90 exercised functions from Day 7.
- Raised `cmd/pipeline` from 0% to 58.1%, `cmd/serve` from 0% to 80.0%,
  `cmd/utils` from 5.4% to 29.8%, and `cmd/document` from 11.1% to 23.4%.
- Tests use temporary directories, isolated environment variables, and a
  loopback server lifecycle; none requires credentials or a live service.
- Native OCR flags now contribute search paths only because gosseract already
  supplies the libraries, eliminating duplicate-library linker warnings.
- The Codecov project target advances to 35%; patch coverage remains 80%.

### Day 9 — Adapter Protocols and Agent Orchestration (2026-09-12)

- Add deterministic protocol coverage to OpenSearch and Elasticsearch.
- Deepen output, reporting, and command-safety coverage in `pkg/agents`.
- Exit target: reach and stabilize 38% overall.

#### Day 9 Results

- Statements/lines: **38.22%** (`10,523 / 27,531`), an increase of 3.18
  percentage points and 876 covered statements from Day 8.
- Functions exercised: **58.37%** (`1,234 / 2,114`), an increase of 4.87
  percentage points and 103 exercised functions from Day 8.
- Raised the OpenSearch adapter from 8.3% to 69.7% and the Elasticsearch
  adapter from 10.3% to 73.3% with loopback protocol tests for collection,
  document, bulk, schema, health, and query behavior.
- Raised `pkg/agents` from 38.4% to 62.1% with output formatting, report
  generation and fallback, Bash command validation, and progress-buffer tests.
- Tests require no credentials or live database services. The Codecov project
  target advances to 38%; patch coverage remains 80%.

### Day 10 — Core Orchestration and 40% (2026-09-13)

- Cover concurrent pipeline processing, resume state, dry runs, batching, and
  failure reporting.
- Exercise query, planning, evaluation, Bash, chunking, configuration, loader,
  and registry behavior in `pkg/agents`.
- Exit target: reach and stabilize 40% overall for v0.15.0.

#### Day 10 Results

- Statements/lines: **40.06%** (`11,028 / 27,531`), an increase of 1.84
  percentage points and 505 covered statements from Day 9.
- Functions exercised: **61.68%** (`1,304 / 2,114`), an increase of 3.31
  percentage points and 70 exercised functions from Day 9.
- Raised `pkg/pipeline` from 20.9% to 84.9% with deterministic processing,
  batching, resume, dry-run, progress, and failure-path tests.
- Raised `pkg/agents` from 62.1% to 81.7% with reasoning-agent contracts,
  safe Bash execution, chunking analysis, configuration resolution, and
  loader/registry lifecycle tests.
- Tests use temporary directories, fake HTTP transports, and recording clients;
  none requires credentials or a live service.
- The Codecov project target advances to 40%; patch coverage remains 80%.

### Day 11 — Validation and Reporting Boundaries (2026-09-14)

- Cover configuration validation, environment-independent path precedence,
  and LLM HTTP protocols.
- Exercise backup inventory, OCR fallback, collection comparison, collection
  statistics, and schema recommendation output.
- Exit target: exceed 42% overall and ratchet the project threshold to 42%.

#### Day 11 Results

- Statements/lines: **42.08%** (`11,586 / 27,531`), an increase of 2.02
  percentage points and 558 covered statements from Day 10.
- Functions exercised: **64.29%** (`1,359 / 2,114`), an increase of 2.61
  percentage points and 55 exercised functions from Day 10.
- Raised `pkg/llm` from 37.2% to 86.7%, `pkg/config` from 35.2% to 46.0%,
  `pkg/image` from 31.9% to 53.6%, `cmd/backup` from 7.9% to 32.0%,
  `cmd/collection` from 8.0% to 18.2%, `cmd/stats` from 26.4% to 59.1%, and
  `cmd/schema` from 24.2% to 74.2%.
- Tests use fake HTTP transports, temporary directories, mock databases, and
  captured output; none requires credentials or a live service.
- The Codecov project target advances to 42%; patch coverage remains 80%.

### Day 12 — Chroma Protocols and Configuration Paths (2026-09-15)

- Add deterministic collection, document, batch, metadata, schema, and search
  protocol coverage for Chroma.
- Exercise configuration loading, environment defaults, schema merging, fix
  application, command display, template, shell-completion, and sync paths.
- Exit target: reach and stabilize 45% overall.

#### Day 12 Results

- Statements/lines: **45.29%** (`12,463 / 27,521`), an increase of 3.21
  percentage points and 877 covered statements from Day 11.
- Functions exercised: **66.86%** (`1,414 / 2,115`), an increase of 2.57
  percentage points and 55 exercised functions from Day 11.
- Raised the Chroma adapter from 10.6% to 81.9%, `pkg/config` from 46.0% to
  60.8%, and `cmd/config` from 15.9% to 40.0%.
- Fixed Chroma semantic result conversion for the pinned SDK and made
  primitive metadata filtering consistent across create, update, and batch
  writes.
- Tests use loopback HTTP servers and temporary directories; none requires
  credentials, developer home state, or a live service.
- The Codecov project target advances to 45%; patch coverage remains 80%.

### Day 13 — Redis and MongoDB Adapter Protocols (2026-09-16)

- Exercise Redis collection, document, query, health, response conversion,
  pipelining, and server-error paths through an in-process RESP fixture.
- Exercise MongoDB collection, index, document, query, schema metadata,
  cursor decoding, and command-error paths through the driver's mock wire
  deployment.
- Exit target: move both adapters above 50% and ratchet overall coverage.

#### Day 13 Results

- Statements/lines: **46.88%** (`12,903 / 27,521`), an increase of 1.59
  percentage points and 440 covered statements from Day 12.
- Functions exercised: **69.65%** (`1,473 / 2,115`), an increase of 2.79
  percentage points and 59 exercised functions from Day 12.
- Raised the Redis adapter from 13.7% to 70.4% and the MongoDB adapter from
  10.9% to 59.4%.
- Tests use in-memory protocol connections and the MongoDB driver's mock
  deployment; none requires credentials or a live service.
- The Codecov project target advances to 46%; patch coverage remains 80%.

### Day 14 — Adapter Protocol Expansion and 50% (2026-09-17)

- Exercise Milvus collection, document, query, schema, health, conversion,
  and failure paths through an in-process SDK fixture.
- Exercise Qdrant collection, point, scroll, search, filter, conversion, and
  health paths through deterministic generated-client implementations.
- Cover the mock adapter lifecycle and Pinecone control plane without
  credentials or live services.
- Exit target: reach and stabilize 50% overall.

#### Day 14 Results

- Statements/lines: **50.06%** (`13,777 / 27,521`), an increase of 3.18
  percentage points and 874 covered statements from Day 13.
- Functions exercised: **73.66%** (`1,558 / 2,115`), an increase of 4.01
  percentage points and 85 exercised functions from Day 13.
- Raised Milvus from 11.1% to 61.3%, the mock adapter from 24.1% to 92.9%,
  Qdrant from 20.6% to 68.4%, and Pinecone from 9.6% to 35.6%.
- Tests use in-process SDK clients, generated gRPC client interfaces, and an
  in-memory HTTP transport; none requires credentials or a live service.
- The Codecov project target advances to 50%; patch coverage remains 80%.

### Day 15 — Complete the Adapter Floor (2026-09-18)

- Exercise Supabase SQL collection, document, query, schema, transaction, and
  failure paths with a deterministic `database/sql` driver.
- Exercise Neo4j Cypher behavior through a narrow query executor boundary.
- Cover Weaviate adapter query, fallback, collection, schema, validation, and
  error-conversion paths through its fake HTTP protocol server.
- Exercise Pinecone document and metadata operations through a narrow data
  plane boundary over the public SDK types.
- Exit target: bring every supported vector database adapter above 50%.

#### Day 15 Results

- Statements/lines: **54.32%** (`14,960 / 27,540`), an increase of 4.26
  percentage points and 1,183 covered statements from Day 14.
- Functions exercised: **77.21%** (`1,636 / 2,119`), an increase of 3.55
  percentage points and 78 exercised functions from Day 14.
- Raised Supabase from 21.4% to 80.7%, Neo4j from 15.1% to 83.0%, Weaviate
  from 44.7% to 52.9%, and Pinecone from 35.6% to 70.8%.
- Every supported vector database adapter now exceeds 50% statement coverage.
- Tests use deterministic SQL, HTTP, SDK, and query-executor fixtures; none
  requires credentials or a live service.
- The Codecov project target advances to 54%; patch coverage remains 80%.

### Path from 54.32% to 80%

1. Raise low-coverage command packages, beginning with backup, collection,
   configuration, document, schema, and stats behavior.
2. Deepen adapter error and pagination coverage while preserving the new 50%
   package floor.
3. Deepen `pkg/config`, `pkg/evaluation`, `pkg/llm`, `pkg/pdf`, and `pkg/stack`
   around validation, cancellation, retries, and partial failures.

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

After an achieved checkpoint is stable locally and on `main`, raise Codecov's
project target to that value in a separate commit. Never lower the target or
exclude meaningful production code to make a gate pass. Keep patch coverage at
80% so new and changed behavior remains well tested throughout the campaign.

## Long-Term Milestones

- **40%**: cover core orchestration happy paths and validation failures.
- **60%**: bring every supported vector database adapter above 50%.
- **80%**: close command and integration-boundary gaps without test-only
  production hooks or broad coverage exclusions.
