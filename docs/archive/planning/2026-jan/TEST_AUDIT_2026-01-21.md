# Test Coverage Audit — 2026-01-21

## Current Status
- Command: `go test -coverprofile=coverage.out -covermode=atomic ./src/...`
- Overall coverage: 8.6%
- Coverage report: coverage.out (run `go tool cover -func=coverage.out`)

## Target
- Raise overall coverage to ≥ 60%
- Focus on fast, reliable unit tests; keep integration tests isolated behind tags

## High-Impact Areas (0% or very low coverage)
- CLI commands
  - [root.go](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/root.go)
  - [agents](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/agents/agents.go)
  - [chunking](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/chunking/chunking.go)
  - [collection](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/collection)
  - [embeddings](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/embeddings)
  - [mcp](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/mcp)
  - [pipeline](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/pipeline)
  - [query](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/query)
- Core packages
  - [executor](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/executor/executor.go)
  - [image](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/image)
  - [llm](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/llm)
  - [mcp](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/mcp)
  - [mcpinstaller](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/mcpinstaller/installer.go)
  - [mock](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/mock/client.go)
  - [output](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/output/formatter.go)
  - [progress](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/progress/reporter.go)
  - [repl](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/repl)
  - [version](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/version/version.go)
- Vector DB layer (unit-testable portions)
  - [vectordb/timeout.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/timeout.go)
  - Converters/validators in each provider package (Weaviate, Supabase, Milvus, etc.)

## Packages with existing but low coverage (raise to 50%+ quickly)
- [cmd/config](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/config) — 2.9%
- [cmd/document](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/document) — 11.1%
- [cmd/schema](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/schema) — 24.2%
- [cmd/stats](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/stats) — 26.4%
- [cmd/utils](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/utils) — 4.2%
- [pkg/agents](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/agents) — 28.2%
- [pkg/config](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/config) — 17.8%
- [pkg/pdf](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/pdf) — 28.0%
- [pkg/pipeline](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/pipeline) — 16.4%
- Vector DB providers (10–21%)
  - [chroma](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/chroma)
  - [elasticsearch](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/elasticsearch)
  - [milvus](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/milvus)
  - [mongodb](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/mongodb)
  - [neo4j](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/neo4j)
  - [opensearch](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/opensearch)
  - [pinecone](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/pinecone)
  - [qdrant](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/qdrant)
  - [supabase](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/supabase)
  - [weaviate](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/weaviate)

## Task List (prioritized)
- Add unit tests for CLI root command behaviors and flags
  - Exercise `Execute`, grouped usage, tips control in [root.go](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/root.go#L145-L176)
- Add tests for embeddings list formatting and icons
  - Validate type icons and compatibility output in [list.go](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/embeddings/list.go#L171-L247)
- Add tests for cmd/utils formatting and error helpers
  - Cover printing/truncation and config helpers in [utils](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/utils)
- Expand cmd/schema tests to cover suggest and validation paths
  - See [suggest.go](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/schema/suggest.go)
- Add tests for cmd/document create/list/update flows with mock DB
  - Target [document](file:///Users/maximilien/github/maximilien/weave-cli/src/cmd/document)
- Add unit tests for pkg/version
  - Cover `Get`, `String`, `IsRelease` in [version.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/version/version.go#L30-L82)
- Add tests for pkg/output formatter
  - Table/JSON formatting in [formatter.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/output/formatter.go)
- Add tests for pkg/progress reporter
  - Start/advance/complete events in [reporter.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/progress/reporter.go)
- Add tests for pkg/repl commands and completer using mocks
  - [repl](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/repl)
- Add tests for pkg/mcp and mcpinstaller error handling (no external IO)
  - [mcp](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/mcp), [installer](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/mcpinstaller/installer.go)
- Add tests for pkg/image helpers
  - EXIF and storage path in [image](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/image)
- Add tests for pkg/llm client configuration and fallbacks
  - [llm](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/llm/client.go)
- Add tests for pkg/executor basic orchestration
  - [executor.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/executor/executor.go)
- Vector DB layer — pure functions first (no network)
  - Timeout helpers in [timeout.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/timeout.go#L33-L88)
  - Converters and validators:
    - Weaviate converters in [adapter.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/weaviate/adapter.go)
    - Supabase helpers in [collections.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/supabase/collections.go) and [schema.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/supabase/schema.go)
    - Similar for Milvus/MongoDB/Neo4j/OpenSearch/Qdrant/Pinecone packages
- Raise factory.ValidateConfig coverage across providers
  - E.g., [weaviate/factory.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/weaviate/factory.go#L35), [supabase/factory.go](file:///Users/maximilien/github/maximilien/weave-cli/src/pkg/vectordb/supabase/factory.go#L36)

## Execution Strategy (to reach ≥60%)
- Phase 1: Pure helpers and formatting (+15–20%)
  - pkg/version, pkg/output, pkg/progress, cmd/utils, embeddings list
- Phase 2: CLI commands with mock DB (+15–20%)
  - cmd/document, cmd/schema, root flags/behaviors
- Phase 3: Vector DB pure logic (+15–20%)
  - converters, validators, timeout helpers, factory validation
- Phase 4: REPL and MCP/LLM non-network logic (+5–10%)
  - repl command paths, llm client config, mcpinstaller

## Coverage Command Recipes
- Unit-only coverage across src:
  - `go test -coverprofile=coverage.out -covermode=atomic ./src/...`
  - `go tool cover -func=coverage.out`
- HTML report:
  - `go tool cover -html=coverage.out -o coverage.html`
- Optional: use `./test.sh coverage` for curated unit + mock integration

## Notes
- Keep integration tests under `// +build integration` or `-tags=integration`
- Prefer mocks/stubs for DB clients; avoid network in unit tests
- Aim for small, isolated tests for converters/validators and command output

