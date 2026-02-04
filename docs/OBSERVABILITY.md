# Observability Guide

**Production-ready observability for weave-cli with structured logging, Prometheus metrics, and health endpoints.**

---

## Table of Contents

- [Overview](#overview)
- [Structured Logging](#structured-logging)
- [Prometheus Metrics](#prometheus-metrics)
- [Health Endpoints](#health-endpoints)
- [Kubernetes Integration](#kubernetes-integration)
- [Monitoring Stack Examples](#monitoring-stack-examples)

---

## Overview

Weave-cli provides three pillars of observability for production deployments:

1. **Structured Logging**: JSON format for log aggregation (ELK, Datadog, Splunk)
2. **Prometheus Metrics**: Operation telemetry for monitoring and alerting
3. **Health Endpoints**: Kubernetes-native liveness and readiness probes

All features are opt-in and designed for zero-overhead when disabled.

---

## Structured Logging

### Quick Start

```bash
# JSON format for production log aggregation
weave --log-format json --log-level debug docs ls MyDocs

# Text format for CLI usage (default, colored)
weave --log-format text --log-level info docs ls MyDocs

# Write logs to file
weave --log-file /var/log/weave.log docs ls MyDocs
```

### Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `debug` | Verbose debugging information | Development, troubleshooting |
| `info` | General informational messages | Production default |
| `warn` | Warning messages | Production |
| `error` | Error messages only | Minimal logging |

### Log Formats

**JSON Format** (production):
```json
{
  "level": "info",
  "time": "2026-02-04T10:00:00Z",
  "message": "Document created successfully",
  "vdb_type": "milvus",
  "collection": "AuctionListings",
  "document_id": "auction-12345"
}
```

**Text Format** (CLI):
```
[INFO] 2026-02-04 10:00:00 Document created successfully
```

### Structured Fields

Weave-cli automatically adds context fields to logs:

| Field | Description | Example |
|-------|-------------|---------|
| `vdb_type` | Vector database type | `"milvus"`, `"qdrant"`, `"weaviate"` |
| `operation` | VDB operation name | `"ListCollections"`, `"CreateDocument"` |
| `collection` | Collection name | `"AuctionListings"` |
| `document_id` | Document ID | `"auction-12345"` |

### Log Aggregation

**ELK Stack (Elasticsearch + Logstash + Kibana)**:
```bash
# Stream JSON logs to Logstash
weave --log-format json --log-level info docs ls MyDocs | logstash -f logstash.conf
```

**Datadog**:
```bash
# Write to file and use Datadog agent
weave --log-format json --log-file /var/log/weave/app.log docs ls MyDocs
# Configure Datadog agent to tail /var/log/weave/app.log
```

**Splunk**:
```bash
# Use Splunk Universal Forwarder
weave --log-format json --log-file /opt/splunk/weave.log docs ls MyDocs
```

---

## Prometheus Metrics

### Quick Start

```bash
# Enable metrics server on default port 9090
weave --metrics docs ls MyDocs

# Custom port
weave --metrics --metrics-port 8080 docs ls MyDocs

# Access metrics
curl http://localhost:9090/metrics
```

### Available Metrics

#### 1. Request Duration (Histogram)

Tracks VDB operation latency with 9 buckets (.001s to 10s):

```prometheus
weave_request_duration_seconds_bucket{vdb_type="milvus",operation="ListCollections",status="success",le="0.1"} 150
weave_request_duration_seconds_sum{vdb_type="milvus",operation="ListCollections",status="success"} 12.5
weave_request_duration_seconds_count{vdb_type="milvus",operation="ListCollections",status="success"} 150
```

**Labels**:
- `vdb_type`: Vector database type (milvus, qdrant, weaviate, etc.)
- `operation`: Operation name (ListCollections, CreateDocument, etc.)
- `status`: Operation status (success, error)

**Use Cases**:
- P50/P95/P99 latency tracking
- SLA monitoring
- Performance regression detection

#### 2. Document Count (Counter)

Tracks total documents processed:

```prometheus
weave_documents_total{vdb_type="milvus",operation="create"} 1523
weave_documents_total{vdb_type="milvus",operation="list"} 450
```

**Labels**:
- `vdb_type`: Vector database type
- `operation`: Operation type (create, list, update, delete)

**Use Cases**:
- Throughput monitoring
- Usage analytics
- Billing/metering

#### 3. Error Count (Counter)

Tracks errors by type:

```prometheus
weave_errors_total{vdb_type="milvus",operation="CreateDocument",error_type="timeout"} 5
weave_errors_total{vdb_type="qdrant",operation="Search",error_type="auth"} 2
```

**Labels**:
- `vdb_type`: Vector database type
- `operation`: Operation name
- `error_type`: Error category (timeout, auth, network, unknown)

**Use Cases**:
- Error rate alerting
- Reliability tracking
- Incident investigation

#### 4. Active Connections (Gauge)

Tracks active VDB connections:

```prometheus
weave_active_connections{vdb_type="milvus"} 5
weave_active_connections{vdb_type="qdrant"} 3
```

**Labels**:
- `vdb_type`: Vector database type

**Use Cases**:
- Connection pool monitoring
- Resource usage tracking
- Capacity planning

### Example Queries

**P95 latency for document creation**:
```promql
histogram_quantile(0.95,
  rate(weave_request_duration_seconds_bucket{operation="CreateDocument"}[5m])
)
```

**Error rate by VDB**:
```promql
rate(weave_errors_total[5m]) / rate(weave_request_duration_seconds_count[5m])
```

**Documents created per second**:
```promql
rate(weave_documents_total{operation="create"}[1m])
```

---

## Health Endpoints

### Quick Start

```bash
# Start with metrics enabled
weave --metrics docs ls MyDocs

# Health endpoints available at:
# http://localhost:9090/healthz  - Liveness probe
# http://localhost:9090/readyz   - Readiness probe
```

### /healthz (Liveness Probe)

**Purpose**: Overall health check for Kubernetes liveness probes

**Response** (HTTP 200 - Healthy):
```json
{
  "status": "healthy",
  "timestamp": 1738612800,
  "version": "0.9.15",
  "databases": {}
}
```

**Response** (HTTP 503 - Degraded):
```json
{
  "status": "degraded",
  "timestamp": 1738612800,
  "version": "0.9.15",
  "databases": {
    "milvus-local": "unhealthy",
    "qdrant-cloud": "healthy"
  }
}
```

**Status Codes**:
- `200 OK`: At least one database is healthy OR no databases configured
- `503 Service Unavailable`: All databases are unhealthy (degraded state)

### /readyz (Readiness Probe)

**Purpose**: Readiness check for Kubernetes readiness probes (stricter than liveness)

**Response** (HTTP 200 - Ready):
```json
{
  "status": "healthy",
  "timestamp": 1738612800,
  "version": "0.9.15",
  "databases": {
    "milvus-local": "healthy"
  }
}
```

**Response** (HTTP 503 - Not Ready):
```json
{
  "status": "degraded",
  "timestamp": 1738612800,
  "version": "0.9.15",
  "databases": {
    "milvus-local": "unhealthy"
  }
}
```

**Status Codes**:
- `200 OK`: ALL databases are healthy
- `503 Service Unavailable`: ANY database is unhealthy

**Difference from /healthz**:
- `/healthz`: Returns 200 if at least one DB is healthy (tolerates partial degradation)
- `/readyz`: Returns 200 only if ALL DBs are healthy (strict requirement)

---

## Kubernetes Integration

### Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: weave-cli
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: weave-cli
  template:
    metadata:
      labels:
        app: weave-cli
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: weave
        image: weave-cli:0.9.15
        args:
        - --metrics
        - --metrics-port=9090
        - --log-format=json
        - --log-level=info
        ports:
        - containerPort: 9090
          name: metrics
          protocol: TCP

        # Liveness probe: restart if unhealthy for 90s
        livenessProbe:
          httpGet:
            path: /healthz
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 5
          failureThreshold: 3

        # Readiness probe: remove from service if degraded
        readinessProbe:
          httpGet:
            path: /readyz
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 2

        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi

        env:
        - name: MILVUS_URL
          valueFrom:
            secretKeyRef:
              name: weave-secrets
              key: milvus-url
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: weave-secrets
              key: openai-api-key
```

### Service for Metrics

```yaml
apiVersion: v1
kind: Service
metadata:
  name: weave-metrics
  namespace: production
  labels:
    app: weave-cli
spec:
  type: ClusterIP
  ports:
  - port: 9090
    targetPort: 9090
    name: metrics
  selector:
    app: weave-cli
```

### ServiceMonitor (Prometheus Operator)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: weave-cli
  namespace: production
spec:
  selector:
    matchLabels:
      app: weave-cli
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

---

## Monitoring Stack Examples

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'weave-cli'
    static_configs:
      - targets: ['localhost:9090']
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'weave_.*'
        action: keep
```

### Grafana Dashboard

**Panel 1: Request Latency (P50, P95, P99)**
```promql
# P50
histogram_quantile(0.50, rate(weave_request_duration_seconds_bucket[5m]))

# P95
histogram_quantile(0.95, rate(weave_request_duration_seconds_bucket[5m]))

# P99
histogram_quantile(0.99, rate(weave_request_duration_seconds_bucket[5m]))
```

**Panel 2: Throughput (Requests/sec)**
```promql
sum(rate(weave_request_duration_seconds_count[1m])) by (vdb_type, operation)
```

**Panel 3: Error Rate (%)**
```promql
100 * (
  sum(rate(weave_errors_total[5m])) by (vdb_type)
  /
  sum(rate(weave_request_duration_seconds_count[5m])) by (vdb_type)
)
```

**Panel 4: Active Connections**
```promql
sum(weave_active_connections) by (vdb_type)
```

### Alerting Rules

```yaml
# alerts.yml
groups:
  - name: weave-cli
    interval: 30s
    rules:
      # High error rate
      - alert: WeaveHighErrorRate
        expr: |
          (
            sum(rate(weave_errors_total[5m])) by (vdb_type)
            /
            sum(rate(weave_request_duration_seconds_count[5m])) by (vdb_type)
          ) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate for {{ $labels.vdb_type }}"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"

      # High latency (P95 > 1s)
      - alert: WeaveHighLatency
        expr: |
          histogram_quantile(0.95,
            rate(weave_request_duration_seconds_bucket{operation="CreateDocument"}[5m])
          ) > 1.0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High P95 latency for document creation"
          description: "P95 latency is {{ $value }}s over the last 10 minutes"

      # Health check failing
      - alert: WeaveUnhealthy
        expr: up{job="weave-cli"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Weave CLI is down"
          description: "Health check has been failing for 2 minutes"
```

---

## Best Practices

### Production Deployment

1. **Always use JSON logging** for log aggregation:
   ```bash
   weave --log-format json --log-file /var/log/weave/app.log
   ```

2. **Enable metrics** for monitoring:
   ```bash
   weave --metrics --metrics-port 9090
   ```

3. **Set appropriate log levels**:
   - Development: `debug`
   - Staging: `info`
   - Production: `info` or `warn`

4. **Configure health probes**:
   - Liveness: `/healthz` with 30s interval
   - Readiness: `/readyz` with 10s interval

### Performance Impact

- **JSON logging**: ~5% overhead vs text format
- **Metrics collection**: ~2% overhead (histograms are efficient)
- **Health endpoints**: Negligible (only when probed)

**Recommendation**: Enable all features in production; the observability benefits far outweigh the minimal overhead.

---

## Troubleshooting

### Logs not appearing in JSON format

**Issue**: Logs still showing in text format despite `--log-format json`

**Solution**: Check that no other flags override the format:
```bash
weave --log-format json --log-level debug docs ls MyDocs 2>&1 | grep '^{'
```

### Metrics endpoint not accessible

**Issue**: `curl http://localhost:9090/metrics` fails

**Possible causes**:
1. Metrics server not enabled: Add `--metrics` flag
2. Port conflict: Use different port with `--metrics-port 8080`
3. Process exited: Metrics server only runs while command executes

**For long-running service**:
```bash
# Keep process alive
weave --metrics serve
```

### Health check always returns 503

**Issue**: `/readyz` always returns 503

**Possible causes**:
1. All VDB connections are down
2. No databases configured (expected behavior)
3. Strict readiness check (use `/healthz` for liveness)

**Debug**:
```bash
curl -s http://localhost:9090/healthz | jq .
# Check "databases" field for unhealthy entries
```

---

## References

- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Kubernetes Health Checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [12-Factor App Logging](https://12factor.net/logs)
- [OpenTelemetry](https://opentelemetry.io/)
