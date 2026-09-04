# AGENTS.md

This file is the workflow index for coding agents working in `weave-cli`.
Keep it concise and update it when repository workflows or recurring gotchas
change.

## Maintenance

- Preserve unrelated worktree changes. Commit completed work incrementally in
  small, focused commits so dependency, test, tooling, and documentation
  changes remain independently reviewable.
- Do not wait to create one large commit at the end. Finish and validate each
  coherent piece, commit it, then make one consolidated push only after the
  complete local gate passes.
- Prefer canonical documentation under `docs/` over duplicating long guidance
  here.
- Record reusable diagnostics and workflow changes, not one-off incident notes.

## Default Validation

Run the standard local gate before handing off changes:

```bash
./lint.sh
./build.sh
./test.sh --coverage
```

- `./test.sh --coverage` defaults to the unit-test lane and writes
  `coverage.out`, `coverage.html`, and `coverage.txt` in the repository root.
- Go coverage does not report branch coverage. Treat 80% statement coverage as
  the target, report the current baseline, and improve low-coverage packages
  without weakening or excluding meaningful production code.
- Unit tests must isolate filesystem and environment state with `t.TempDir`,
  `t.Setenv`, or equivalent helpers. Do not depend on credentials from `.env`
  or write into the developer's home directory.
- OCR builds require Tesseract and Leptonica. The repository scripts configure
  Homebrew include and library paths for macOS.

## Go Dependencies And Security

- Keep the Go version aligned across `go.mod` and GitHub Actions workflows.
- After dependency changes, run `go mod tidy`, `go mod verify`, the default
  validation gate, and `govulncheck ./...` when available.
- Prefer released dependency versions. If a security fix requires a major
  client migration, update imports and exercise that integration's unit tests.
- Do not treat `go vet` or `go mod verify` as substitutes for a vulnerability
  database scan.

## Integration Tests

Use focused integration lanes when changing a vector database or external
service:

```bash
./test.sh integration --weaviate
./test.sh integration --milvus
./test.sh integration --mongodb
./test.sh integration --chroma
./test.sh integration --qdrant
./test.sh integration --neo4j
./test.sh integration --pinecone
./test.sh integration --mcp
```

Cloud-backed lanes require their documented credentials. Local service lanes
should skip clearly when the service is unavailable.
