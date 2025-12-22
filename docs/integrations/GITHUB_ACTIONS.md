# GitHub Actions Integration Guide

This guide shows how to use `weave docs batch` in GitHub Actions workflows for automated document ingestion into vector databases.

## Table of Contents

- [Basic Setup](#basic-setup)
- [Environment Variables](#environment-variables)
- [Workflow Examples](#workflow-examples)
- [Exit Code Handling](#exit-code-handling)
- [JSON Output Parsing](#json-output-parsing)
- [Best Practices](#best-practices)

## Basic Setup

### Prerequisites

1. Install weave-cli in your workflow
2. Configure VDB credentials as GitHub Secrets
3. Set up appropriate triggers (push, schedule, etc.)

### Required Secrets

Store these as GitHub repository secrets:

```yaml
# For Qdrant Cloud
QDRANT_API_KEY: ${{ secrets.QDRANT_API_KEY }}
QDRANT_URL: ${{ secrets.QDRANT_URL }}

# For OpenAI embeddings
OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

# For other VDBs, see VDB_SUPPORT_MATRIX.md
```

## Environment Variables

Create a `.env` file or set environment variables directly in the workflow:

```yaml
env:
  VECTOR_DB_TYPE: qdrant-cloud
  QDRANT_API_KEY: ${{ secrets.QDRANT_API_KEY }}
  QDRANT_URL: ${{ secrets.QDRANT_URL }}
  OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
  EMBEDDING_MODEL: text-embedding-3-small
```

## Workflow Examples

### Example 1: Basic Document Ingestion on Push

Ingest documentation whenever docs are updated:

```yaml
name: Ingest Documentation

on:
  push:
    paths:
      - 'docs/**'
    branches:
      - main

jobs:
  ingest:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download weave-cli
        run: |
          curl -L https://github.com/maximilien/weave-cli/releases/latest/download/weave-linux-amd64 -o weave
          chmod +x weave

      - name: Ingest documents
        env:
          VECTOR_DB_TYPE: qdrant-cloud
          QDRANT_API_KEY: ${{ secrets.QDRANT_API_KEY }}
          QDRANT_URL: ${{ secrets.QDRANT_URL }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          ./weave docs batch \
            --directory ./docs \
            --collection documentation \
            --parallel 3 \
            --json > batch-report.json

      - name: Check exit code
        if: failure()
        run: |
          echo "Batch ingestion failed!"
          exit 1
```

### Example 2: Incremental Updates (Only Recent Files)

Process only files modified in the last 24 hours:

```yaml
name: Incremental Documentation Update

on:
  schedule:
    # Run daily at 2 AM UTC
    - cron: '0 2 * * *'

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download weave-cli
        run: |
          curl -L https://github.com/maximilien/weave-cli/releases/latest/download/weave-linux-amd64 -o weave
          chmod +x weave

      - name: Ingest recent documents
        env:
          VECTOR_DB_TYPE: qdrant-cloud
          QDRANT_API_KEY: ${{ secrets.QDRANT_API_KEY }}
          QDRANT_URL: ${{ secrets.QDRANT_URL }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          ./weave docs batch \
            --directory ./docs \
            --collection documentation \
            --since 24h \
            --parallel 3 \
            --json > batch-report.json

      - name: Upload batch report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: batch-report
          path: batch-report.json
```

### Example 3: Multi-Environment with Matrix Strategy

Deploy to multiple environments:

```yaml
name: Multi-Environment Ingestion

on:
  push:
    branches:
      - main

jobs:
  ingest:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        environment: [dev, staging, prod]
        include:
          - environment: dev
            collection: docs-dev
          - environment: staging
            collection: docs-staging
          - environment: prod
            collection: docs-prod

    environment: ${{ matrix.environment }}

    steps:
      - uses: actions/checkout@v4

      - name: Download weave-cli
        run: |
          curl -L https://github.com/maximilien/weave-cli/releases/latest/download/weave-linux-amd64 -o weave
          chmod +x weave

      - name: Ingest to ${{ matrix.environment }}
        env:
          VECTOR_DB_TYPE: qdrant-cloud
          QDRANT_API_KEY: ${{ secrets.QDRANT_API_KEY }}
          QDRANT_URL: ${{ secrets.QDRANT_URL }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          ./weave docs batch \
            --directory ./docs \
            --collection ${{ matrix.collection }} \
            --parallel 3 \
            --json > batch-report-${{ matrix.environment }}.json

      - name: Parse JSON report
        run: |
          STATUS=$(jq -r '.status' batch-report-${{ matrix.environment }}.json)
          EXIT_CODE=$(jq -r '.exit_code' batch-report-${{ matrix.environment }}.json)

          echo "Status: $STATUS"
          echo "Exit Code: $EXIT_CODE"

          if [ "$EXIT_CODE" != "0" ]; then
            echo "::error::Batch ingestion failed for ${{ matrix.environment }}"
            exit $EXIT_CODE
          fi
```

### Example 4: PR Preview Collections

Create temporary collections for pull request previews:

```yaml
name: PR Preview Ingestion

on:
  pull_request:
    paths:
      - 'docs/**'

jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download weave-cli
        run: |
          curl -L https://github.com/maximilien/weave-cli/releases/latest/download/weave-linux-amd64 -o weave
          chmod +x weave

      - name: Ingest to PR preview collection
        env:
          VECTOR_DB_TYPE: qdrant-cloud
          QDRANT_API_KEY: ${{ secrets.QDRANT_API_KEY }}
          QDRANT_URL: ${{ secrets.QDRANT_URL }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          COLLECTION="docs-pr-${{ github.event.pull_request.number }}"

          ./weave docs batch \
            --directory ./docs \
            --collection "$COLLECTION" \
            --parallel 3 \
            --json > batch-report.json

      - name: Comment PR with results
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const report = JSON.parse(fs.readFileSync('batch-report.json', 'utf8'));

            const body = `## 📊 Document Ingestion Report

            - **Status**: ${report.status}
            - **Files Processed**: ${report.files.processed}
            - **Files Failed**: ${report.files.failed}
            - **Documents Created**: ${report.documents.created}
            - **Duration**: ${report.duration_seconds.toFixed(2)}s
            - **Throughput**: ${report.performance.throughput_files_per_sec.toFixed(2)} files/sec

            **Collection**: \`docs-pr-${{ github.event.pull_request.number }}\`
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```

## Exit Code Handling

Weave CLI batch command returns specific exit codes for CI/CD integration:

| Exit Code | Status | Meaning |
|-----------|--------|---------|
| 0 | Success | All documents processed successfully |
| 1 | Partial Failure | Some documents failed (<50% failure rate) |
| 2 | Complete Failure | >50% documents failed or critical error |

### Example: Conditional Steps Based on Exit Code

```yaml
- name: Ingest documents
  id: ingest
  continue-on-error: true
  run: |
    ./weave docs batch --dir ./docs --collection NAME --json > report.json
    echo "exit_code=$?" >> $GITHUB_OUTPUT

- name: Handle partial failure
  if: steps.ingest.outputs.exit_code == '1'
  run: |
    echo "::warning::Some documents failed to ingest, but majority succeeded"
    # Send notification, create issue, etc.

- name: Handle complete failure
  if: steps.ingest.outputs.exit_code == '2'
  run: |
    echo "::error::Document ingestion failed critically"
    exit 1
```

## JSON Output Parsing

Parse JSON output for detailed reporting:

```yaml
- name: Parse and report metrics
  run: |
    # Extract key metrics
    STATUS=$(jq -r '.status' batch-report.json)
    FILES_PROCESSED=$(jq -r '.files.processed' batch-report.json)
    FILES_FAILED=$(jq -r '.files.failed' batch-report.json)
    DOCS_CREATED=$(jq -r '.documents.created' batch-report.json)
    THROUGHPUT=$(jq -r '.performance.throughput_files_per_sec' batch-report.json)

    # Create GitHub step summary
    echo "## Batch Ingestion Results" >> $GITHUB_STEP_SUMMARY
    echo "" >> $GITHUB_STEP_SUMMARY
    echo "- **Status**: $STATUS" >> $GITHUB_STEP_SUMMARY
    echo "- **Files Processed**: $FILES_PROCESSED" >> $GITHUB_STEP_SUMMARY
    echo "- **Files Failed**: $FILES_FAILED" >> $GITHUB_STEP_SUMMARY
    echo "- **Documents Created**: $DOCS_CREATED" >> $GITHUB_STEP_SUMMARY
    echo "- **Throughput**: ${THROUGHPUT} files/sec" >> $GITHUB_STEP_SUMMARY

    # List errors if any
    if [ "$FILES_FAILED" != "0" ]; then
      echo "" >> $GITHUB_STEP_SUMMARY
      echo "### Errors:" >> $GITHUB_STEP_SUMMARY
      jq -r '.errors[] | "- **\(.file)**: \(.error)"' batch-report.json >> $GITHUB_STEP_SUMMARY
    fi
```

## Best Practices

### 1. Use Secrets for Credentials

Never hardcode API keys or credentials:

```yaml
# ❌ BAD
env:
  QDRANT_API_KEY: "my-api-key-12345"

# ✅ GOOD
env:
  QDRANT_API_KEY: ${{ secrets.QDRANT_API_KEY }}
```

### 2. Optimize Parallel Processing

Adjust parallelism based on runner resources:

```yaml
- name: Ingest with optimal parallelism
  run: |
    # GitHub hosted runners have 2 cores
    ./weave docs batch --dir ./docs --collection NAME --parallel 2
```

### 3. Use Incremental Updates for Large Repos

For repositories with many documents, use `--since` to process only recent changes:

```yaml
- name: Daily incremental update
  run: |
    ./weave docs batch \
      --dir ./docs \
      --collection NAME \
      --since 24h \
      --parallel 3
```

### 4. Store Reports as Artifacts

Always save batch reports for debugging:

```yaml
- name: Upload batch report
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: batch-report-${{ github.run_id }}
    path: batch-report.json
    retention-days: 30
```

### 5. Add Retry Logic for Transient Failures

```yaml
- name: Ingest with retry
  uses: nick-fields/retry@v2
  with:
    timeout_minutes: 30
    max_attempts: 3
    retry_on: error
    command: |
      ./weave docs batch \
        --dir ./docs \
        --collection NAME \
        --parallel 3 \
        --json > batch-report.json
```

### 6. Notify on Failures

```yaml
- name: Notify on failure
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    text: 'Document ingestion failed!'
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## Troubleshooting

### Issue: Out of Memory

**Solution**: Reduce parallelism or use smaller batch sizes:

```yaml
run: |
  ./weave docs batch \
    --dir ./docs \
    --collection NAME \
    --parallel 1 \
    --batch-size 5
```

### Issue: Timeout on Large Repositories

**Solution**: Use `--since` flag or split into multiple jobs:

```yaml
strategy:
  matrix:
    folder: [docs/api, docs/guides, docs/tutorials]
steps:
  - run: |
      ./weave docs batch \
        --dir ${{ matrix.folder }} \
        --collection NAME
```

### Issue: API Rate Limits

**Solution**: Add delays between batches or reduce parallelism:

```yaml
run: |
  ./weave docs batch \
    --dir ./docs \
    --collection NAME \
    --parallel 1
```

## Related Documentation

- [VDB Support Matrix](../VDB_SUPPORT_MATRIX.md)
- [Argo Workflows Integration](./ARGO_WORKFLOWS.md)
- [Airflow Integration](./AIRFLOW.md)
- [Batch Command Reference](../../README.md#batch-processing)
