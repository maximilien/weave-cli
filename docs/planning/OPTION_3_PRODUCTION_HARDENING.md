# Option 3: Production Hardening - Detailed Implementation Plan

**Status**: Planning
**Priority**: ⭐⭐ High (enterprise focus)
**Total Effort**: 20-30 hours (phased approach recommended)
**Target**: Weeks 4-5

---

## Overview

Make weave-cli enterprise-ready with comprehensive observability, performance optimizations, security enhancements, and operational improvements.

**Target Users:**
- Enterprise deployments
- Production environments
- High-scale applications
- Security-conscious organizations

---

## Area 1: Observability (5-7 hours)

### 1.1 Structured Logging

**Goal:** Replace fmt.Printf with structured JSON logging

**Implementation:**
```go
import "github.com/rs/zerolog/log"

// Configure logger
func initLogger(format string) {
    if format == "json" {
        log.Logger = log.Output(os.Stdout)
    } else {
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
    }
    
    zerolog.SetGlobalLevel(zerolog.InfoLevel)
    if os.Getenv("DEBUG") == "true" {
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    }
}

// Usage
log.Info().
    Str("vdb_type", "qdrant-cloud").
    Str("collection", "documents").
    Int("count", 1234).
    Msg("Listed collection")

// Error with context
log.Error().
    Err(err).
    Str("vdb_type", vdbType).
    Str("operation", "CreateDocument").
    Msg("Operation failed")
```

**Effort:** 2 hours

---

### 1.2 Prometheus Metrics

**Goal:** Expose metrics endpoint for monitoring

**Implementation:**
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "weave_request_duration_seconds",
            Help: "Duration of VDB requests",
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
)

// HTTP endpoint
http.Handle("/metrics", promhttp.Handler())
go http.ListenAndServe(":9090", nil)
```

**Effort:** 1.5 hours

---

### 1.3 OpenTelemetry Tracing

**Goal:** Distributed tracing for debugging

**Implementation:**
```go
import "go.opentelemetry.io/otel"

func (a *Adapter) CreateDocument(ctx context.Context, ...) error {
    ctx, span := otel.Tracer("weave-cli").Start(ctx, "CreateDocument")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("vdb.type", "qdrant"),
        attribute.String("collection", collection),
    )
    
    // Operation logic
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    return nil
}
```

**Effort:** 1.5 hours

---

### 1.4 Health Check Endpoint

**Goal:** HTTP endpoint for liveness/readiness probes

**Implementation:**
```go
func ServeHealthCheck(port int, client vectordb.VectorDBClient) {
    http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()
        
        if err := client.Health(ctx); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "error": err.Error(),
            })
            return
        }
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
        })
    })
    
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```

**Effort:** 1 hour

---

## Area 2: Performance (6-8 hours)

### 2.1 Connection Pooling

**Goal:** Reuse connections instead of creating new ones

**Implementation:**
```go
type ConnectionPool struct {
    clients map[string]vectordb.VectorDBClient
    mu      sync.RWMutex
}

func (p *ConnectionPool) Get(vdbType string) (vectordb.VectorDBClient, error) {
    p.mu.RLock()
    if client, exists := p.clients[vdbType]; exists {
        p.mu.RUnlock()
        return client, nil
    }
    p.mu.RUnlock()
    
    p.mu.Lock()
    defer p.mu.Unlock()
    
    config := LoadVDBConfig(vdbType)
    client, err := vectordb.CreateClient(config)
    if err != nil {
        return nil, err
    }
    
    p.clients[vdbType] = client
    return client, nil
}
```

**Effort:** 1.5 hours

---

### 2.2 Embedding Cache

**Goal:** Cache embeddings to avoid redundant LLM API calls

**Implementation:**
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

func (c *EmbeddingCache) Get(text string) ([]float32, bool) {
    key := hash(text)
    if val, found := c.cache.Get(key); found {
        return val.([]float32), true
    }
    return nil, false
}

func (c *EmbeddingCache) Set(text string, embedding []float32) {
    key := hash(text)
    c.cache.Set(key, embedding, cache.DefaultExpiration)
}
```

**Effort:** 2 hours

---

### 2.3 Concurrent Query Execution

**Goal:** Execute queries in parallel when possible

**Implementation:**
```go
func SearchMultipleCollections(ctx context.Context, client vectordb.VectorDBClient, collections []string, query string) map[string][]*QueryResult {
    results := make(map[string][]*QueryResult)
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    for _, coll := range collections {
        wg.Add(1)
        go func(collection string) {
            defer wg.Done()
            
            res, err := client.SearchSemantic(ctx, collection, query, nil)
            if err != nil {
                log.Error().Err(err).Str("collection", collection).Msg("Search failed")
                return
            }
            
            mu.Lock()
            results[collection] = res
            mu.Unlock()
        }(coll)
    }
    
    wg.Wait()
    return results
}
```

**Effort:** 1.5 hours

---

## Area 3: Security (5-7 hours)

### 3.1 OAuth2/OIDC Authentication

**Goal:** Support OAuth2 for enterprise SSO

**Implementation:**
```go
import "golang.org/x/oauth2"

func NewOAuthClient(tokenURL, clientID, clientSecret string) (*http.Client, error) {
    config := &oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        Endpoint: oauth2.Endpoint{
            TokenURL: tokenURL,
        },
    }
    
    token, err := config.PasswordCredentialsToken(ctx, username, password)
    if err != nil {
        return nil, err
    }
    
    return config.Client(ctx, token), nil
}
```

**Effort:** 2 hours

---

### 3.2 Secrets Management

**Goal:** Integrate with Vault, AWS Secrets Manager

**Implementation:**
```go
type SecretsProvider interface {
    GetSecret(key string) (string, error)
}

type VaultProvider struct {
    client *vault.Client
}

func (v *VaultProvider) GetSecret(key string) (string, error) {
    secret, err := v.client.Logical().Read(key)
    if err != nil {
        return "", err
    }
    return secret.Data["value"].(string), nil
}

// AWS Secrets Manager
type AWSSecretsProvider struct {
    client *secretsmanager.SecretsManager
}
```

**Effort:** 2 hours

---

### 3.3 TLS Certificate Pinning

**Goal:** Prevent MITM attacks

**Implementation:**
```go
func CreateTLSConfig(certPins []string) *tls.Config {
    return &tls.Config{
        VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
            for _, rawCert := range rawCerts {
                cert, _ := x509.ParseCertificate(rawCert)
                fingerprint := sha256.Sum256(cert.Raw)
                pin := base64.StdEncoding.EncodeToString(fingerprint[:])
                
                for _, expectedPin := range certPins {
                    if pin == expectedPin {
                        return nil
                    }
                }
            }
            return fmt.Errorf("certificate pin verification failed")
        },
    }
}
```

**Effort:** 1.5 hours

---

## Area 4: Operations (4-6 hours)

### 4.1 Graceful Shutdown

**Goal:** Handle SIGTERM gracefully

**Implementation:**
```go
func RunWithGracefulShutdown(ctx context.Context, run func(context.Context) error) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    
    errChan := make(chan error, 1)
    go func() {
        errChan <- run(ctx)
    }()
    
    select {
    case err := <-errChan:
        return err
    case sig := <-sigChan:
        log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
        cancel()
        
        // Wait for graceful shutdown
        select {
        case err := <-errChan:
            return err
        case <-time.After(30 * time.Second):
            return fmt.Errorf("shutdown timeout")
        }
    }
}
```

**Effort:** 1 hour

---

### 4.2 Retry with Exponential Backoff

**Goal:** Resilient operations

**Implementation:**
```go
import "github.com/cenkalti/backoff/v4"

func RetryWithBackoff(operation func() error) error {
    b := backoff.NewExponentialBackOff()
    b.InitialInterval = 1 * time.Second
    b.MaxInterval = 30 * time.Second
    b.MaxElapsedTime = 5 * time.Minute
    
    return backoff.Retry(operation, b)
}
```

**Effort:** 1 hour

---

### 4.3 Circuit Breaker

**Goal:** Prevent cascade failures

**Implementation:**
```go
import "github.com/sony/gobreaker"

var cb *gobreaker.CircuitBreaker

func init() {
    cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "VectorDB",
        MaxRequests: 3,
        Interval:    time.Minute,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
    })
}

func CallWithCircuitBreaker(fn func() error) error {
    _, err := cb.Execute(func() (interface{}, error) {
        return nil, fn()
    })
    return err
}
```

**Effort:** 1.5 hours

---

### 4.4 Rate Limiting

**Goal:** Protect VDBs from overload

**Implementation:**
```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(rps int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(rps), rps*2),
    }
}

func (r *RateLimiter) Wait(ctx context.Context) error {
    return r.limiter.Wait(ctx)
}
```

**Effort:** 1 hour

---

## Total Timeline

**Week 4:**
- Days 1-2: Observability (logging, metrics, tracing, health)
- Days 3-5: Performance (connection pool, cache, concurrency)

**Week 5:**
- Days 1-3: Security (OAuth2, secrets, TLS)
- Days 4-5: Operations (shutdown, retry, circuit breaker, rate limit)

**Total: 20-28 hours across 2 weeks**
