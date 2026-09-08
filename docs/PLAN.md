# Weave CLI Development Plan

**Last updated**: 2026-09-08

**Current milestone**: v0.13.0 release

## Immediate — Ship v0.13.0

- [x] Align local builds and GitHub Actions on Go 1.26.6.
- [x] Update direct dependencies and audit reachable vulnerabilities.
- [x] Add local and CI coverage reporting.
- [x] Reach 25% statement coverage and raise the Codecov project ratchet.
- [x] Keep Build, Test, Lint, and Security green on `main`.
- [x] Prepare the changelog and release documentation.
- [ ] Tag `v0.13.0` and verify all release artifacts and checksums.

## Next — v0.14.0 at 30%

The next minor release is tied to the 30% statement-coverage milestone.
Approximately 1,299 additional covered statements were needed after Day 4;
the exact count must be recalculated whenever production code changes.

Primary targets:

1. Command validation and dependency boundaries in `cmd/stack`, `cmd/mcp`,
   `cmd/agents`, and `cmd/vdb`.
2. Weaviate query and document HTTP protocols.
3. Low-coverage OpenSearch, Elasticsearch, Pinecone, and Chroma adapters.
4. High-impact paths in `pkg/agents`, `pkg/config`, and `pkg/pipeline`.

## Then — v0.15.0 at 40%

- Cover core orchestration happy paths and partial failures.
- Add deterministic cancellation, retry, pagination, and timeout tests.
- Establish a meaningful coverage floor across supported adapters.
- Consolidate shared fake-server fixtures where duplication emerges.

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
- Keep Codecov project coverage ratcheted at achieved milestones and patch
  coverage at 80%.

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
