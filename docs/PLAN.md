# Weave CLI Development Plan

**Last updated**: 2026-10-20

**Current milestone**: continue from 75.5% toward 80% in one-point daily steps

## Completed — v0.13.0

- [x] Align local builds and GitHub Actions on Go 1.26.6.
- [x] Update direct dependencies and audit reachable vulnerabilities.
- [x] Add local and CI coverage reporting.
- [x] Reach 25% statement coverage and raise the Codecov project ratchet.
- [x] Keep Build, Test, Lint, and Security green on `main`.
- [x] Prepare the changelog and release documentation.
- [x] Tag `v0.13.0` and verify all release artifacts and checksums.

## Completed — v0.14.0

- [x] Raise statement coverage from 25.28% to 30.18%.
- [x] Raise the Codecov project ratchet from 25% to 30%.
- [x] Add deterministic Weaviate query-protocol tests.
- [x] Cover the agents, embeddings, MCP, stack, and VDB command boundaries.
- [x] Keep lint, build, test, and security gates green.

## Completed — v0.15.0 at 40%

The next minor release is tied to the 40% statement-coverage milestone. The
campaign advanced from the v0.14.0 baseline of 30.18% (`8,307 / 27,529`) to
40.06% (`11,028 / 27,531`) on Day 10.

- [x] Reach 35% and raise the Codecov project ratchet to 35%.
- [x] Reach 38% and raise the Codecov project ratchet to 38%.
- [x] Reach 40% and prepare v0.15.0.
- [x] Confirm Build, Test, Lint, and Security on the release commit.
- [x] Tag v0.15.0 and verify release artifacts and checksums.

Completed targets:

1. Cover core orchestration happy paths and partial failures.
2. Add deterministic cancellation, retry, pagination, and timeout tests.
3. Establish a meaningful coverage floor across supported adapters.
4. Consolidate shared fake-server fixtures where duplication emerges.

## Active — Post-v0.15 Coverage Campaign

- [x] Reach 42% on Day 11 and raise the Codecov project ratchet to 42%.
- [x] Reach 45% on Day 12 and raise the Codecov project ratchet to 45%.
- [x] Raise Redis and MongoDB above 50% on Day 13 and ratchet project coverage
  to 46%.
- [x] Reach 50% on Day 14 by raising Milvus, the mock adapter, and Qdrant
  above 50% and adding Pinecone control-plane coverage.
- [x] Bring every supported vector database adapter above 50% on Day 15 by
  covering Supabase, Neo4j, Weaviate, and the Pinecone data plane.
- [x] Reach 55% on Day 16 by covering collection command paths and PDF context
  extraction behavior.
- [x] Reach 57% on Day 17 by covering document command, inspection, and PDF
  conversion behavior.
- [x] Reach 58% on Day 18 by covering backup and health command behavior.
- [x] Reach 60% on Day 19 by covering configuration, collection comparison,
  and stack process boundaries.
- [x] Reach 62% on Day 20 by covering shared command utilities, configuration
  prompts, stack backups, statistics output, and configuration diagnostics.
- [x] Reach 64% on Day 21 by covering Opik synchronization, remote backup
  protocols, interactive configuration fixes, doctor output, and agent and
  command initialization boundaries.
- [x] Reach 65% on Day 22 (2026-09-24), beginning the one-percentage-point
  daily coverage cadence, then raise the Codecov project target to 65%.
- [x] Reach 66% on Day 23 (2026-09-25) and raise the Codecov project target
  to 66%.
- [x] Exceed 73% statement coverage on 2026-10-14; advance the Codecov
  project ratchet by one point to 67%, with 68% as the next ratchet step.
- [x] Reach 73.5% statement coverage on 2026-10-15 with stack command and
  interactive configuration tests; advance the Codecov project ratchet to 68%.
- [x] Reach 74% statement coverage on 2026-10-17 with stack ingestion,
  collection count, configuration sync/completion, and PDF ingestion tests;
  advance the Codecov project ratchet to 69%.
- [x] Reach 74.5% statement coverage on 2026-10-18 with MCP stdio lifecycle,
  installer release parsing, stack retry/checkpoint, and configuration error
  tests; advance the Codecov project ratchet to 70%.
- [x] Reach 75.0% statement coverage on 2026-10-19 with generic PDF ingestion,
  mock document update, and Milvus query/schema/lifecycle failure tests; advance
  the Codecov project ratchet to 71%.
- [x] Reach 75.5% statement coverage on 2026-10-20 with Weaviate collection
  query and MCP installer success/failure/reinstall coverage; advance the
  Codecov project ratchet to 72%.
- [x] Reach 76% statement coverage on 2026-10-22 with EXIF metadata, dataset
  creation, document update, and Weaviate document display tests. Advance the
  Codecov project ratchet to 73% after a green `main` validation run.

- [x] Reach 77% statement coverage on 2026-10-23 with generic collection
  listing presentation and failure tests (77.15% measured). Advance the
  Codecov project ratchet to 73% after confirming the previous checkpoint's CI.

- [x] Reach 77.5% statement coverage on 2026-10-24 with document update
  selection and deletion safeguards (77.51% measured). Advance the Codecov
  project ratchet to 74% after confirming the previous checkpoint's CI.

- [x] Reach 78% statement coverage on 2026-10-25 with document operations,
  embedding diagnostics, collection selection, query aggregation, and deletion
  failure tests (78.00% measured). Advance the Codecov project ratchet to 75%.

- [x] Reach 78.25% local statement coverage on 2026-10-26 with extracted
  PDF-image metadata and YAML editing tests (78.31% measured). Advance the
  Codecov project ratchet to 76%; report Linux CI coverage separately.

## Continuous Workstreams

### Security and dependencies

- Keep Go versions aligned across `go.mod` and every workflow.
- Use released module versions and run `go mod tidy`, `go mod verify`, and
  `govulncheck ./...` after dependency changes.
- Do not substitute `go vet` or checksum verification for vulnerability
  scanning.

### Test quality

- Keep tests isolated from developer credentials, home directories, and live
  services.
- Prefer temporary directories and fake HTTP/gRPC servers.
- Beginning with Day 22, raise the daily checkpoint and Codecov project target
  exactly one percentage point at a time; overshoot is acceptable but does not
  skip the following day's rung. Keep patch coverage at 80%.

### Documentation

- Update the changelog, roadmap, release notes, and documentation index for
  every minor release.
- Keep active plans concise; move historical detail under `docs/archive/`.
- Verify commands and version requirements against repository scripts.

## Standard Gate

Before a consolidated push or release tag:

```bash
./lint.sh
./build.sh
./test.sh --coverage
```

After pushing, confirm Build, Test, Lint, and Security on the exact commit. Tag
only that green SHA, then monitor the Release workflow through publication.

See [`tests/COVERAGE_PLAN.md`](tests/COVERAGE_PLAN.md) for measurements and
[`ROADMAP.md`](ROADMAP.md) for release milestones.
