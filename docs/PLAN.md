# Weave CLI Development Plan

**Last updated**: 2026-09-15

**Current milestone**: continue from 45% toward the 60% adapter milestone

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
- [ ] Bring every supported vector database adapter above 50% on the path to
  the 60% project milestone.

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
