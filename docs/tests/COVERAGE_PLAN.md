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

The current source count makes these approximate requirements:

| Target | Covered statements needed | Increase from baseline |
| --- | ---: | ---: |
| 20% | 5,505 | 1,067 |
| 25% | 6,881 | 2,443 |
| 30% | 8,257 | 3,819 |

These counts are planning estimates. Use the percentage reported by
`./test.sh --coverage` because the denominator will change with production
code.

### Day 1 — Small Reusable Packages (2026-09-05)

- Add deterministic tests for `pkg/health`, `pkg/metrics`, `pkg/server`, and
  `pkg/mcpinstaller`.
- Begin `pkg/storage` tests around paths, metadata, thumbnails, and failures.
- Avoid live services, credentials, fixed home-directory paths, and sleeps.
- Exit target: at least 18% overall, with no package above 80% regressing.

### Day 2 — Execution Boundaries (2026-09-06)

- Cover `pkg/executor` planning, dispatch, error propagation, and cancellation.
- Cover testable `pkg/repl` parsing and command-routing behavior.
- Introduce narrow interfaces or injected functions only where they improve
  production boundaries; do not add test-only production hooks.
- Exit target: reach and stabilize 20% overall.

### Day 3 — Commands and Evaluation (2026-09-07)

- Test `cmd/eval`, `cmd/config`, and root command argument/flag validation.
- Use temporary configuration directories and fake providers.
- Prioritize invalid input, missing configuration, and provider failure paths.
- Exit target: at least 22.5% overall.

### Day 4 — MCP, Stack, and Adapters (2026-09-08)

- Add protocol and response-conversion tests for `pkg/mcp`.
- Exercise `pkg/stack` rendering and orchestration without starting services.
- Add fake HTTP/gRPC coverage to the lowest vector database adapters.
- Exit target: reach and stabilize 25% overall.

### Day 5 — Consolidate and Start 30% (2026-09-09)

- Fill branch and error-path gaps exposed by the coverage report.
- Remove flakes and duplicated fixtures introduced during the campaign.
- Select the next high-yield slices in `cmd/stack`, `cmd/eval`, `pkg/repl`, and
  the Weaviate adapter for the 30% campaign.
- Exit target: keep 25% green and publish the measured path to 30%.

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
