# Argo Workflows Integration Guide

This guide shows how to use `weave docs batch` in Argo Workflows for automated document ingestion into vector databases.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Basic Workflow](#basic-workflow)
- [Workflow Examples](#workflow-examples)
- [ConfigMaps and Secrets](#configmaps-and-secrets)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Overview

Argo Workflows is a container-native workflow engine for Kubernetes. It's ideal for:
- Scheduled document ingestion
- Event-driven processing
- Parallel batch operations
- Complex DAG-based workflows

## Prerequisites

1. Argo Workflows installed in Kubernetes cluster
2. kubectl configured with cluster access
3. weave-cli container image (or build custom image)
4. VDB credentials stored as Kubernetes secrets

### Install Argo Workflows

```bash
kubectl create namespace argo
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/latest/download/install.yaml
```

## Basic Workflow

### Simple Document Ingestion Workflow

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: weave-ingest-
  namespace: argo
spec:
  entrypoint: ingest-documents

  templates:
  - name: ingest-documents
    container:
      image: ghcr.io/maximilien/weave-cli:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          weave docs batch \
            --directory /workspace/docs \
            --collection documentation \
            --parallel 3 \
            --json > /tmp/batch-report.json

          # Check exit code
          EXIT_CODE=$?
          echo "Exit code: $EXIT_CODE"

          cat /tmp/batch-report.json
          exit $EXIT_CODE

      volumeMounts:
        - name: workspace
          mountPath: /workspace

      env:
        - name: VECTOR_DB_TYPE
          value: "qdrant-cloud"
        - name: QDRANT_API_KEY
          valueFrom:
            secretKeyRef:
              name: vdb-credentials
              key: qdrant-api-key
        - name: QDRANT_URL
          valueFrom:
            secretKeyRef:
              name: vdb-credentials
              key: qdrant-url
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: vdb-credentials
              key: openai-api-key

  volumes:
    - name: workspace
      git:
        repo: https://github.com/your-org/docs-repo.git
        revision: main
```

## Workflow Examples

### Example 1: DAG with Multiple Collections

Process multiple document collections in parallel:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: weave-multi-collection-
spec:
  entrypoint: multi-collection-dag

  templates:
  - name: multi-collection-dag
    dag:
      tasks:
      - name: ingest-api-docs
        template: ingest-collection
        arguments:
          parameters:
          - name: directory
            value: "/workspace/docs/api"
          - name: collection
            value: "api-docs"

      - name: ingest-tutorials
        template: ingest-collection
        arguments:
          parameters:
          - name: directory
            value: "/workspace/docs/tutorials"
          - name: collection
            value: "tutorials"

      - name: ingest-guides
        template: ingest-collection
        arguments:
          parameters:
          - name: directory
            value: "/workspace/docs/guides"
          - name: collection
            value: "guides"

      - name: summary
        template: create-summary
        dependencies: [ingest-api-docs, ingest-tutorials, ingest-guides]

  - name: ingest-collection
    inputs:
      parameters:
      - name: directory
      - name: collection
    container:
      image: ghcr.io/maximilien/weave-cli:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          weave docs batch \
            --directory {{inputs.parameters.directory}} \
            --collection {{inputs.parameters.collection}} \
            --parallel 3 \
            --json > /tmp/{{inputs.parameters.collection}}-report.json

          EXIT_CODE=$?
          cat /tmp/{{inputs.parameters.collection}}-report.json
          exit $EXIT_CODE

      envFrom:
        - configMapRef:
            name: weave-config
        - secretRef:
            name: vdb-credentials

      volumeMounts:
        - name: workspace
          mountPath: /workspace

  - name: create-summary
    container:
      image: alpine:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          echo "All collections processed successfully!"

  volumes:
    - name: workspace
      git:
        repo: https://github.com/your-org/docs-repo.git
        revision: main
```

### Example 2: Incremental Updates with CronWorkflow

Schedule incremental updates daily:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: CronWorkflow
metadata:
  name: weave-daily-update
  namespace: argo
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  timezone: "America/New_York"

  workflowSpec:
    entrypoint: incremental-update

    templates:
    - name: incremental-update
      container:
        image: ghcr.io/maximilien/weave-cli:latest
        command: ["/bin/sh", "-c"]
        args:
          - |
            # Process only files modified in last 24 hours
            weave docs batch \
              --directory /workspace/docs \
              --collection documentation \
              --since 24h \
              --parallel 3 \
              --json > /tmp/batch-report.json

            EXIT_CODE=$?

            # Parse and display results
            FILES_PROCESSED=$(jq -r '.files.processed' /tmp/batch-report.json)
            FILES_FAILED=$(jq -r '.files.failed' /tmp/batch-report.json)

            echo "Processed: $FILES_PROCESSED files"
            echo "Failed: $FILES_FAILED files"
            echo "Exit code: $EXIT_CODE"

            cat /tmp/batch-report.json
            exit $EXIT_CODE

        envFrom:
          - configMapRef:
              name: weave-config
          - secretRef:
              name: vdb-credentials

        volumeMounts:
          - name: workspace
            mountPath: /workspace

    volumes:
      - name: workspace
        git:
          repo: https://github.com/your-org/docs-repo.git
          revision: main
```

### Example 3: Event-Driven Workflow with Sensors

Trigger ingestion on repository updates:

```yaml
# EventSource
apiVersion: argoproj.io/v1alpha1
kind: EventSource
metadata:
  name: github-events
  namespace: argo-events
spec:
  github:
    docs-repo:
      repositories:
        - owner: your-org
          names:
            - docs-repo
      webhook:
        endpoint: /push
        port: "12000"
        method: POST
        url: https://your-domain.com
      events:
        - push
      apiToken:
        name: github-token
        key: token
      webhookSecret:
        name: github-webhook-secret
        key: secret

---
# Sensor
apiVersion: argoproj.io/v1alpha1
kind: Sensor
metadata:
  name: docs-ingestion-sensor
  namespace: argo-events
spec:
  dependencies:
    - name: github-push
      eventSourceName: github-events
      eventName: docs-repo
      filters:
        data:
          - path: body.ref
            type: string
            value:
              - refs/heads/main
          - path: body.commits.#.modified
            type: string
            comparator: "~"
            value: ["docs/.*"]

  triggers:
    - template:
        name: weave-ingestion-trigger
        argoWorkflow:
          operation: submit
          source:
            resource:
              apiVersion: argoproj.io/v1alpha1
              kind: Workflow
              metadata:
                generateName: weave-ingest-
              spec:
                entrypoint: ingest-documents
                templates:
                - name: ingest-documents
                  container:
                    image: ghcr.io/maximilien/weave-cli:latest
                    command: ["/bin/sh", "-c"]
                    args:
                      - |
                        weave docs batch \
                          --directory /workspace/docs \
                          --collection documentation \
                          --parallel 3 \
                          --json > /tmp/batch-report.json

                        EXIT_CODE=$?
                        cat /tmp/batch-report.json
                        exit $EXIT_CODE
                    envFrom:
                      - configMapRef:
                          name: weave-config
                      - secretRef:
                          name: vdb-credentials
                    volumeMounts:
                      - name: workspace
                        mountPath: /workspace
                volumes:
                  - name: workspace
                    git:
                      repo: https://github.com/your-org/docs-repo.git
                      revision: main
```

### Example 4: Workflow with Retry and Error Handling

Robust workflow with retries and failure handling:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: weave-robust-
spec:
  entrypoint: robust-ingest

  # Retry failed steps up to 3 times
  retryStrategy:
    limit: "3"
    retryPolicy: "OnFailure"
    backoff:
      duration: "1m"
      factor: 2
      maxDuration: "10m"

  templates:
  - name: robust-ingest
    steps:
    - - name: validate-env
        template: check-environment

    - - name: ingest-docs
        template: ingest-collection

    - - name: handle-failure
        template: notify-failure
        when: "{{steps.ingest-docs.status}} == Failed"

    - - name: handle-partial
        template: notify-partial
        when: "{{steps.ingest-docs.outputs.exitCode}} == 1"

  - name: check-environment
    container:
      image: ghcr.io/maximilien/weave-cli:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          # Verify environment variables are set
          if [ -z "$QDRANT_API_KEY" ]; then
            echo "Error: QDRANT_API_KEY not set"
            exit 1
          fi

          if [ -z "$OPENAI_API_KEY" ]; then
            echo "Error: OPENAI_API_KEY not set"
            exit 1
          fi

          echo "Environment validated successfully"

      envFrom:
        - secretRef:
            name: vdb-credentials

  - name: ingest-collection
    container:
      image: ghcr.io/maximilien/weave-cli:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          weave docs batch \
            --directory /workspace/docs \
            --collection documentation \
            --parallel 3 \
            --json > /tmp/batch-report.json

          EXIT_CODE=$?

          echo "Exit code: $EXIT_CODE"
          cat /tmp/batch-report.json

          # Store exit code for conditional steps
          echo $EXIT_CODE > /tmp/exit-code.txt

          exit $EXIT_CODE

      envFrom:
        - configMapRef:
            name: weave-config
        - secretRef:
            name: vdb-credentials

      volumeMounts:
        - name: workspace
          mountPath: /workspace

    outputs:
      parameters:
      - name: exitCode
        valueFrom:
          path: /tmp/exit-code.txt

  - name: notify-failure
    container:
      image: curlimages/curl:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
            -H 'Content-Type: application/json' \
            -d '{"text":"Document ingestion failed completely!"}'

  - name: notify-partial
    container:
      image: curlimages/curl:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
            -H 'Content-Type: application/json' \
            -d '{"text":"Document ingestion completed with some failures"}'

  volumes:
    - name: workspace
      git:
        repo: https://github.com/your-org/docs-repo.git
        revision: main
```

## ConfigMaps and Secrets

### Create ConfigMap for weave-cli Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: weave-config
  namespace: argo
data:
  VECTOR_DB_TYPE: "qdrant-cloud"
  EMBEDDING_MODEL: "text-embedding-3-small"
  EMBEDDING_DIMENSIONS: "1536"
```

```bash
kubectl apply -f weave-config.yaml
```

### Create Secret for VDB Credentials

```bash
kubectl create secret generic vdb-credentials \
  --from-literal=qdrant-api-key="your-qdrant-api-key" \
  --from-literal=qdrant-url="https://your-qdrant-url.cloud" \
  --from-literal=openai-api-key="your-openai-api-key" \
  -n argo
```

## Error Handling

### Exit Code Interpretation

weave-cli returns specific exit codes:

| Exit Code | Status | Action |
|-----------|--------|--------|
| 0 | Success | Continue workflow |
| 1 | Partial Failure | Log warning, continue |
| 2 | Complete Failure | Fail workflow, notify |

### Conditional Execution Based on Exit Code

```yaml
templates:
- name: conditional-handling
  steps:
  - - name: ingest
      template: ingest-collection

  - - name: on-success
      template: success-handler
      when: "{{steps.ingest.exitCode}} == 0"

  - - name: on-partial
      template: partial-handler
      when: "{{steps.ingest.exitCode}} == 1"

  - - name: on-failure
      template: failure-handler
      when: "{{steps.ingest.exitCode}} == 2"
```

## Best Practices

### 1. Use Volume Claims for Large Repositories

For large document repositories, use PersistentVolumeClaims:

```yaml
volumes:
  - name: workspace
    persistentVolumeClaim:
      claimName: docs-pvc
```

### 2. Set Resource Limits

Prevent resource exhaustion:

```yaml
container:
  image: ghcr.io/maximilien/weave-cli:latest
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "2000m"
```

### 3. Use Parallelism for Multiple Collections

Process multiple collections concurrently:

```yaml
spec:
  parallelism: 3  # Run up to 3 tasks in parallel
```

### 4. Implement Health Checks

Add liveness and readiness probes:

```yaml
container:
  livenessProbe:
    exec:
      command:
      - /bin/sh
      - -c
      - "ps aux | grep weave"
    initialDelaySeconds: 30
    periodSeconds: 10
```

### 5. Archive Workflow Results

Store batch reports for audit:

```yaml
templates:
- name: archive-results
  archiveLocation:
    archiveLogs: true
  outputs:
    artifacts:
    - name: batch-report
      path: /tmp/batch-report.json
      s3:
        bucket: my-workflow-artifacts
        key: "{{workflow.name}}/batch-report.json"
```

### 6. Monitor with Prometheus

Export metrics for monitoring:

```yaml
spec:
  metrics:
    prometheus:
      - name: workflow_duration
        help: "Duration of workflow execution"
        histogram:
          buckets: [10, 30, 60, 120, 300]
        gauge:
          value: "{{workflow.duration}}"
```

## Troubleshooting

### Issue: Git Clone Timeout

**Solution**: Increase timeout or use shallow clone:

```yaml
volumes:
  - name: workspace
    git:
      repo: https://github.com/your-org/docs-repo.git
      revision: main
      depth: 1  # Shallow clone
```

### Issue: Out of Memory

**Solution**: Increase memory limits or reduce parallelism:

```yaml
resources:
  limits:
    memory: "4Gi"
```

### Issue: Workflow Stuck

**Solution**: Add activeDeadlineSeconds:

```yaml
spec:
  activeDeadlineSeconds: 3600  # 1 hour timeout
```

## Related Documentation

- [GitHub Actions Integration](./GITHUB_ACTIONS.md)
- [Airflow Integration](./AIRFLOW.md)
- [Argo Workflows Documentation](https://argoproj.github.io/argo-workflows/)
- [VDB Support Matrix](../VDB_SUPPORT_MATRIX.md)
