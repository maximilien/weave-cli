# Production Hardening Plan - Started 2026-02-03

**Status**: In Progress
**Priority**: High (AuctionsMax production readiness)
**Total Effort**: 20-30 hours (phased approach)
**Current Phase**: Phase 1 - Observability (5-7 hours)

---

## Overview

Make weave-cli enterprise-ready with comprehensive observability, performance
optimizations, security enhancements, and operational improvements.

**Context**: After successful AuctionsMax demo and customer feedback, production
deployment is the likely next step. Enterprise customers require observable,
performant, and secure infrastructure.

---

## Phase 1: Observability (5-7 hours) ⚡ CURRENT

### 1.1 Structured Logging with zerolog (2 hours)

**Goal**: Replace fmt.Printf with structured JSON logging

**Implementation**:

```go
import "github.com/rs/zerolog/log"

// Configure logger in main.go
func initLogger(format string, level string, file string) {
    // Set log level
    switch level {
    case "debug":
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    case "info":
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    case "warn":
        zerolog.SetGlobalLevel(zerolog.WarnLevel)
    case "error":
        zerolog.SetGlobalLevel(zerolog.ErrorLevel)
    default:
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }

    // Set output format
    if format == "json" {
        log.Logger = log.Output(os.Stdout)
    } else {
        // Human-friendly console output
        log.Logger = log.Output(zerolog.ConsoleWriter{
            Out: os.Stderr,
            TimeFormat: time.RFC3339,
        })
    }

    // Optional file output
    if file != "" {
        f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err == nil {
            log.Logger = log.Output(zerolog.MultiLevelWriter(
                zerolog.ConsoleWriter{Out: os.Stderr},
                f,
            ))
        }
    }
}

// Usage in VDB operations
log.Info().
    Str("vdb_type", "qdrant-cloud").
    Str("collection", "documents").
    Int("count", 1234).
    Dur("duration", elapsed).
    Msg("Listed collection")

// Error logging with rich context
log.Error().
    Err(err).
    Str("vdb_type", vdbType).
    Str("operation", "CreateDocument").
    Str("collection", collectionName).
    Str("document_id", docID).
    Msg("Operation failed")
```

**CLI Flags**:
- `--log-level` (debug, info, warn, error) - default: info
- `--log-format` (json, console) - default: console
- `--log-file` (path) - optional file output

**Files to Modify**:
- `src/cmd/root.go` - Add global flags and initialization
- `src/pkg/logging/logger.go` - New logging package
- All VDB adapters - Replace fmt.Printf with log calls

**Benefits**:
- Structured logs for easy parsing/analysis
- Correlation IDs for tracing requests
- Log levels for filtering
- File output for persistence
- Production-ready logging

---

### 1.2 Prometheus Metrics (1.5 hours)

**Goal**: Expose /metrics endpoint for monitoring

**Implementation**:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "weave_request_duration_seconds",
            Help: "Duration of VDB requests",
            Buckets: []float64{.001, .005, .01, .05, .1, .5, 1, 5, 10},
        },
        []string{"vdb_type", "operation", "status"},
    )

    documentCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "weave_documents_total",
            Help: "Total documents processed",
        },
        []string{"vdb_type", "operation"},
    )

    errorCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "weave_errors_total",
            Help: "Total errors encountered",
        },
        []string{"vdb_type", "operation", "error_type"},
    )

    activeConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "weave_active_connections",
            Help: "Number of active VDB connections",
        },
        []string{"vdb_type"},
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
    prometheus.MustRegister(documentCount)
    prometheus.MustRegister(errorCount)
    prometheus.MustRegister(activeConnections)
}

// Usage in VDB operations
timer := prometheus.NewTimer(requestDuration.WithLabelValues(vdbType, "CreateDocument", "success"))
defer timer.ObserveDuration()

documentCount.WithLabelValues(vdbType, "create").Inc()
```

**HTTP Endpoint**:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

func ServeMetrics(port int) {
    http.Handle("/metrics", promhttp.Handler())
    http.Handle("/healthz", healthHandler())
    log.Info().Int("port", port).Msg("Starting metrics server")
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```

**CLI Flags**:
- `--metrics` - Enable metrics endpoint (default: false)
- `--metrics-port` - Port for metrics server (default: 9090)

**Metrics Exposed**:
- `weave_request_duration_seconds` - Histogram of request durations
- `weave_documents_total` - Counter of documents processed
- `weave_errors_total` - Counter of errors by type
- `weave_active_connections` - Gauge of active connections

**Benefits**:
- Real-time monitoring
- Alerting on performance/errors
- Capacity planning
- Performance optimization insights

---

### 1.3 Health Check Endpoint (1 hour)

**Goal**: HTTP endpoint for Kubernetes liveness/readiness probes

**Implementation**:

```go
func healthHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        health := make(map[string]interface{})
        health["status"] = "healthy"
        health["timestamp"] = time.Now().Unix()
        health["version"] = version.Version

        // Check VDB connections
        vdbs := make(map[string]string)
        for _, cfg := range config.GetConfiguredDatabases() {
            client, err := getOrCreateClient(cfg)
            if err != nil {
                vdbs[cfg.Name] = "error: " + err.Error()
                health["status"] = "degraded"
                continue
            }

            if err := client.Health(ctx); err != nil {
                vdbs[cfg.Name] = "unhealthy"
                health["status"] = "degraded"
            } else {
                vdbs[cfg.Name] = "healthy"
            }
        }
        health["databases"] = vdbs

        w.Header().Set("Content-Type", "application/json")
        if health["status"] == "healthy" {
            w.WriteHeader(http.StatusOK)
        } else {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        json.NewEncoder(w).Encode(health)
    }
}
```

**Endpoints**:
- `GET /healthz` - Overall health (200 = healthy, 503 = unhealthy)
- `GET /readyz` - Readiness check (checks VDB connectivity)

**Response Example**:

```json
{
  "status": "healthy",
  "timestamp": 1738612800,
  "version": "0.9.14.2",
  "databases": {
    "weaviate-cloud": "healthy",
    "milvus-local": "healthy",
    "qdrant-cloud": "healthy"
  }
}
```

**Kubernetes Integration**:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 9090
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /readyz
    port: 9090
  initialDelaySeconds: 5
  periodSeconds: 10
```

**Benefits**:
- Kubernetes-native health checks
- Automatic restart on failure
- Load balancer integration
- Deployment confidence

---

## Phase 2: Performance (6-8 hours) 🔄 NEXT

### 2.1 Connection Pooling (1.5 hours)

**Goal**: Reuse VDB connections instead of creating new ones

**Implementation**:

```go
type ConnectionPool struct {
    clients map[string]vectordb.VectorDBClient
    mu      sync.RWMutex
}

var globalPool = &ConnectionPool{
    clients: make(map[string]vectordb.VectorDBClient),
}

func (p *ConnectionPool) Get(cfg *config.VectorDBConfig) (vectordb.VectorDBClient, error) {
    key := fmt.Sprintf("%s:%s", cfg.Type, cfg.Name)

    // Try to get existing client
    p.mu.RLock()
    if client, exists := p.clients[key]; exists {
        p.mu.RUnlock()
        return client, nil
    }
    p.mu.RUnlock()

    // Create new client
    p.mu.Lock()
    defer p.mu.Unlock()

    // Double-check after acquiring write lock
    if client, exists := p.clients[key]; exists {
        return client, nil
    }

    client, err := vectordb.CreateClient(cfg)
    if err != nil {
        return nil, err
    }

    p.clients[key] = client
    return client, nil
}
```

**Impact**: 50-70% reduction in operation latency for repeated operations

---

### 2.2 Embedding Cache (2 hours)

**Goal**: Cache embeddings to avoid redundant LLM API calls

**Implementation**:

```go
import "github.com/patrickmn/go-cache"

type EmbeddingCache struct {
    cache *cache.Cache
}

func NewEmbeddingCache() *EmbeddingCache {
    return &EmbeddingCache{
        cache: cache.New(1*time.Hour, 10*time.Minute),
    }
}

func (c *EmbeddingCache) GetOrGenerate(ctx context.Context, text string, model string) ([]float32, error) {
    key := hashKey(text, model)

    // Check cache first
    if val, found := c.cache.Get(key); found {
        return val.([]float32), nil
    }

    // Generate embedding
    embedding, err := llm.GenerateEmbedding(ctx, text, model)
    if err != nil {
        return nil, err
    }

    // Cache it
    c.cache.Set(key, embedding, cache.DefaultExpiration)
    return embedding, nil
}
```

**Impact**: 90%+ cost reduction for repeated queries

---

### 2.3 Concurrent Query Execution (1.5 hours)

**Goal**: Execute queries in parallel when possible

---

## Phase 3: Security (5-7 hours) 🔒 FUTURE

### 3.1 OAuth2/OIDC Authentication (2 hours)
### 3.2 Secrets Management Integration (2 hours)
### 3.3 TLS Certificate Pinning (1.5 hours)

---

## Phase 4: Operations (4-6 hours) ⚙️ FUTURE

### 4.1 Graceful Shutdown (1 hour)
### 4.2 Retry with Exponential Backoff (1 hour)
### 4.3 Circuit Breaker (1.5 hours)
### 4.4 Rate Limiting (1 hour)

---

## Success Metrics

**Phase 1 Complete When**:
- ✅ All VDB operations use structured logging
- ✅ Prometheus metrics exposed on /metrics
- ✅ Health endpoint responds with JSON status
- ✅ Integration tests pass with new logging
- ✅ Documentation updated

**Production Readiness Indicators**:
- Observable: Can debug issues from logs/metrics
- Performant: <100ms p50 latency for common operations
- Reliable: Health checks work with k8s probes
- Secure: Secrets not logged or exposed

---

## Timeline

**Day 1 (Today - 2026-02-03)**:
- ✅ Document plan
- Structured logging implementation (2h)
- Prometheus metrics (1.5h)

**Day 2**:
- Health check endpoint (1h)
- Testing and integration (1h)
- Documentation (1h)

**Week 2**:
- Connection pooling
- Embedding cache
- Concurrent query execution

---

## Dependencies

**Go Packages**:
- `github.com/rs/zerolog` - Structured logging
- `github.com/prometheus/client_golang` - Metrics
- `github.com/patrickmn/go-cache` - Caching (Phase 2)

**Installation**:

```bash
go get github.com/rs/zerolog
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

---

## References

- [Option 3 Detailed Plan](OPTION_3_PRODUCTION_HARDENING.md)
- [zerolog Documentation](https://github.com/rs/zerolog)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [12-Factor App Logging](https://12factor.net/logs)
