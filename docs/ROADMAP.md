# Weave CLI Roadmap

**Last updated**: 2026-09-11

**Current release**: v0.14.0

**Next release**: v0.15.0

This roadmap tracks active release milestones. Historical plans and completed
project notes live under [`archive/`](archive/); the detailed test campaign is
tracked in [`tests/COVERAGE_PLAN.md`](tests/COVERAGE_PLAN.md).

## v0.13.0 — Security and Quality

**Status**: Released 2026-09-08

- Upgrade the supported toolchain and CI matrix to Go 1.26.6.
- Refresh direct dependencies and migrate Weaviate to client v5 and
  OpenSearch signing to AWS SDK v2.
- Clear all reachable `govulncheck` findings.
- Add first-class `./test.sh --coverage` reporting and Codecov integration.
- Raise unit-test statement coverage from 16.13% to 25.28%.
- Add deterministic coverage for executor, REPL, evaluation commands, MCP,
  stack state, and Weaviate protocol boundaries.
- Improve Opik query tracing, dataset upload, and experiment synchronization.

See [`releases/RELEASE_v0.13.0.md`](releases/RELEASE_v0.13.0.md) for release
details.

## v0.14.0 — 30% Coverage

**Status**: Released 2026-09-10

- Raised unit-test statement coverage from 25.28% to 30.18%.
- Raised `cmd/agents`, `cmd/embeddings`, `cmd/mcp`, and `cmd/vdb` from zero to
  meaningful deterministic coverage.
- Raised Weaviate adapter coverage from 16.7% to 44.7% through document and
  query protocol tests.
- Raised the Codecov project ratchet to 30%; patch coverage remains 80%.

See [`releases/RELEASE_v0.14.0.md`](releases/RELEASE_v0.14.0.md) for release
details.

## v0.15.0 — 40% Coverage

**Target**: 40% statement coverage

**Progress**: 35.04% (`9,647 / 27,531`) after Day 8

- Deepen core orchestration coverage in `pkg/agents`, `pkg/config`,
  `pkg/pipeline`, and `pkg/stack`.
- Exercise cancellation, retries, pagination, and partial-failure behavior
  across service boundaries.
- Raise supported database adapters toward a consistent package-level floor.

## v1.0.0 — Production Stability

The long-term quality target remains 80% statement coverage without excluding
meaningful production code. A v1.0 release also requires:

- Stable CLI and configuration contracts.
- Green cross-platform builds and integration lanes.
- No reachable known vulnerabilities.
- Current installation, migration, and troubleshooting documentation.
- Release artifacts and checksums for every supported platform.

## Release History

- **v0.14.0** — 30% coverage milestone, command boundaries, and Weaviate
  protocol coverage.
- **v0.13.0** — Go 1.26.6, dependency security updates, coverage reporting,
  and the 25% quality milestone.
- **v0.12.3** — Published patch before the quality campaign.
- **v0.12.x** — Remote storage, diagnostics, performance, and reliability.
- **v0.11.x** — Backup and restore, including cross-database portability.
- **v0.9.x** — Agents, evaluation, observability, and OSS embeddings.
- **v0.8.x and earlier** — Multi-database adapter foundation.

For complete historical details, see [`CHANGELOG.md`](../CHANGELOG.md) and
[`releases/`](releases/).
