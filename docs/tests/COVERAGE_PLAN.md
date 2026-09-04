# Go Coverage Plan

## Baseline and Goal

The unit-test baseline measured by `./test.sh --coverage` on 2026-09-04 is:

- Statements/lines: **16.15%** (`4,446 / 27,532`)
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

## Milestones

- **25%**: eliminate zero coverage in reusable `pkg/*` production packages.
- **40%**: cover core orchestration happy paths and validation failures.
- **60%**: bring every supported vector database adapter above 50%.
- **80%**: close command and integration-boundary gaps without test-only
  production hooks or broad coverage exclusions.

At each milestone, record the measured statement count here and raise CI's
ratcheted baseline only after the new level is stable.
