# Multi-VDB Support Plan for Weave Stack

**Status**: Planning (Phase 2+)
**Priority**: Medium
**Created**: Feb 26, 2026

---

## Overview

Weave Stack currently defaults to Milvus but needs comprehensive support for all vector databases that `weave` CLI supports, including both local and cloud variants.

**Goal**: Make Weave Stack truly VDB-agnostic with full support for:
- Milvus (local/cloud)
- Qdrant (local/cloud)
- Weaviate (local/cloud)
- Chroma (local/cloud)
- MongoDB Atlas (cloud)
- Supabase (local/cloud)
- Neo4j (local/cloud)
- OpenSearch (local/cloud)

---

## Current State (v0.10.2)

### What Works
- ✅ Milvus local deployment via Kind/Minikube
- ✅ User can edit `infrastructure.vectordb.type` in weave-stack.yaml
- ✅ Helm templates exist for Milvus
- ✅ Port-forward works for Milvus (hardcoded port 19530)

### What Needs Work
- ⚠️ No Helm templates for other VDBs (Qdrant, Weaviate, Chroma)
- ⚠️ Port-forward hardcoded to `milvus` service name
- ⚠️ Ingestion pipeline assumes Milvus configuration
- ⚠️ No cloud VDB support (all assume local deployment)
- ⚠️ Templates only show Milvus examples

---

## User Feedback

From CLIENT0_GETTING_STARTED.md review:
> "Client0 uses milvus but other clients and even Client 0 may want to use other VDBs. Can we convert this into STACK_GETTING_STARTED.md (or similar) that shows how to use a different VDB (both local and cloud notes)."

From WEAVE_STACK_QUICKSTART.md review:
> "Could we replace the hardcoded milvus to vdb and discover which VDB in yaml? In otherwords abstract the VDB since we want weave stack to work for all supported VDB"

> "Love the templates but we need to plan to support all the VDBs that weave support. Including local and cloud versions when available."

---

## Design Principles

1. **VDB Discovery**: Commands should auto-detect VDB from weave-stack.yaml
2. **Unified Interface**: Same commands work regardless of VDB choice
3. **Smart Defaults**: Each VDB has sensible defaults (ports, resources, config)
4. **Cloud-Ready**: Support both local (Kind/Minikube) and cloud (EKS/GKE) for each VDB
5. **Template Flexibility**: All 4 templates (quickstart, production, multimodal, oss) work with any VDB

---

## Implementation Plan

### Phase 1: VDB Abstraction Layer (Week 1)

**Goal**: Abstract hardcoded "milvus" references

#### Code Changes

1. **Add VDB Discovery** (`src/pkg/stack/vdb.go`):
```go
// Get VDB info from weave-stack.yaml
func GetVectorDBInfo(config *StackConfig) (*VectorDBInfo, error) {
    vdbType := config.Infrastructure.VectorDB.Type

    return &VectorDBInfo{
        Type:        vdbType,
        ServiceName: getServiceName(vdbType),  // e.g., "milvus", "qdrant"
        Port:        getDefaultPort(vdbType),   // e.g., 19530, 6333
        Protocol:    getProtocol(vdbType),      // e.g., "grpc", "http"
    }, nil
}
```

2. **Update Port-Forward** (`src/cmd/stack/port_forward.go`):
```go
// Before: weave stack port-forward milvus 19530:19530
// After:  weave stack port-forward vectordb 19530:19530
//         (auto-detects "milvus" from config, uses 19530)

// Also support explicit service name:
//   weave stack port-forward milvus 19530:19530
```

3. **Update Logs Command** (`src/cmd/stack/logs.go`):
```go
// weave stack logs vectordb
// Auto-resolves to correct service based on config
```

4. **Update Status Command** (`src/cmd/stack/status.go`):
```go
// Show VDB type and status
VectorDB: milvus (running)
Port:     19530
```

#### Testing
- ✅ Test with Milvus (no regression)
- ✅ Verify abstraction works
- ✅ Update integration tests

### Phase 2: Helm Templates for All VDBs (Week 2)

**Goal**: Add Helm deployment templates for each VDB

#### Templates to Create

1. **Qdrant** (`templates/helm/weave-stack/templates/qdrant-deployment.yaml`):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-qdrant
spec:
  template:
    spec:
      containers:
      - name: qdrant
        image: qdrant/qdrant:{{ .Values.vectordb.version }}
        ports:
        - containerPort: 6333  # REST API
        - containerPort: 6334  # gRPC
```

2. **Weaviate** (`templates/helm/weave-stack/templates/weaviate-deployment.yaml`):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-weaviate
spec:
  template:
    spec:
      containers:
      - name: weaviate
        image: semitechnologies/weaviate:{{ .Values.vectordb.version }}
        ports:
        - containerPort: 8080  # REST API
```

3. **Chroma** (`templates/helm/weave-stack/templates/chroma-deployment.yaml`):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-chroma
spec:
  template:
    spec:
      containers:
      - name: chroma
        image: chromadb/chroma:{{ .Values.vectordb.version }}
        ports:
        - containerPort: 8000  # HTTP API
```

#### Conditional Rendering

Update `templates/helm/weave-stack/templates/` with conditionals:
```yaml
{{- if eq .Values.vectordb.type "milvus" }}
  {{- include "milvus.deployment" . }}
{{- else if eq .Values.vectordb.type "qdrant" }}
  {{- include "qdrant.deployment" . }}
{{- else if eq .Values.vectordb.type "weaviate" }}
  {{- include "weaviate.deployment" . }}
{{- else if eq .Values.vectordb.type "chroma" }}
  {{- include "chroma.deployment" . }}
{{- end }}
```

#### Testing
- Test each VDB deploys successfully
- Verify port-forward works for each
- Test ingestion pipeline with each VDB

### Phase 3: Cloud VDB Support (Week 3)

**Goal**: Support cloud-hosted VDB services

#### Implementation

1. **Detect Local vs Cloud** (`src/pkg/stack/types.go`):
```yaml
infrastructure:
  vectordb:
    type: milvus
    deployment: local  # or "cloud"

    # Local deployment config
    local:
      version: 2.3.0
      resources: {...}

    # Cloud connection config
    cloud:
      endpoint: "https://your-cluster.zilliz.com"
      api_key_env: "MILVUS_API_KEY"
```

2. **Skip Deployment for Cloud VDBs**:
```go
if config.Infrastructure.VectorDB.Deployment == "cloud" {
    // Don't deploy Helm chart
    // Just configure connection info
    return configureCloudVDB(config)
}
```

3. **Cloud VDB Examples**:
- **Milvus Cloud (Zilliz)**: HTTPS endpoint + API key
- **Qdrant Cloud**: HTTPS endpoint + API key
- **Weaviate Cloud (WCS)**: HTTPS endpoint + API key
- **MongoDB Atlas**: Connection string
- **Supabase**: PostgreSQL connection + pgvector

#### Testing
- Test with actual cloud VDB accounts
- Verify ingestion works with cloud endpoints
- Document setup for each cloud provider

### Phase 4: Documentation & Examples (Week 4)

**Goal**: Complete documentation for all VDB options

#### Documentation Updates

1. **Create STACK_GETTING_STARTED.md** (general version):
   - VDB-agnostic quick start
   - Section for each VDB (local + cloud)
   - Configuration examples
   - Migration guides

2. **Update CLIENT0_GETTING_STARTED.md**:
   - Keep as Milvus-specific
   - Add note: "Using different VDB? See STACK_GETTING_STARTED.md"
   - Archive after Client0 migration

3. **Update WEAVE_STACK_QUICKSTART.md**:
   - Add VDB comparison table
   - When to choose each VDB
   - Performance characteristics

4. **Create VDB-Specific Guides**:
   - `docs/guides/vdb/MILVUS.md`
   - `docs/guides/vdb/QDRANT.md`
   - `docs/guides/vdb/WEAVIATE.md`
   - `docs/guides/vdb/CHROMA.md`

#### Examples to Add

1. **Example: Switch from Milvus to Qdrant**:
```bash
# Edit weave-stack.yaml
infrastructure:
  vectordb:
    type: qdrant  # Changed from milvus
    version: "1.7.4"

# Redeploy
weave stack down
weave stack up --runtime kind
```

2. **Example: Use Qdrant Cloud**:
```yaml
infrastructure:
  vectordb:
    type: qdrant
    deployment: cloud
    cloud:
      endpoint: "https://your-cluster.qdrant.io"
      api_key_env: "QDRANT_API_KEY"
```

3. **Example: Multimodal with Weaviate**:
```bash
weave stack init --template multimodal --runtime kind

# Edit weave-stack.yaml to use Weaviate
# Weaviate excels at multimodal search

weave stack up --runtime kind
```

---

## VDB Feature Matrix

| VDB | Local | Cloud | Text | Images | Filtering | Highlights |
|-----|-------|-------|------|--------|-----------|------------|
| **Milvus** | ✅ | ✅ (Zilliz) | ✅ | ✅ | Advanced | Fast, scalable, default |
| **Qdrant** | ✅ | ✅ | ✅ | ✅ | Advanced | Rust, fast, filtering |
| **Weaviate** | ✅ | ✅ (WCS) | ✅ | ✅ | GraphQL | Semantic, multimodal |
| **Chroma** | ✅ | ✅ | ✅ | ⚠️ | Basic | Simple, Python-first |
| **MongoDB** | ✅ | ✅ (Atlas) | ✅ | ⚠️ | Advanced | Familiar, integrated |
| **Supabase** | ✅ | ✅ | ✅ | ⚠️ | SQL | PostgreSQL, pgvector |

---

## Default Ports Reference

| VDB | Port(s) | Protocol |
|-----|---------|----------|
| Milvus | 19530 | gRPC |
| Qdrant | 6333 (REST), 6334 (gRPC) | HTTP/gRPC |
| Weaviate | 8080 | HTTP/GraphQL |
| Chroma | 8000 | HTTP |
| MongoDB | 27017 | MongoDB Wire Protocol |
| Supabase | 5432 | PostgreSQL |

---

## Resource Defaults by VDB

### Milvus
```yaml
resources:
  requests:
    memory: "2Gi"
    cpu: "1"
  limits:
    memory: "4Gi"
    cpu: "2"
```

### Qdrant
```yaml
resources:
  requests:
    memory: "1Gi"
    cpu: "0.5"
  limits:
    memory: "2Gi"
    cpu: "1"
```

### Weaviate
```yaml
resources:
  requests:
    memory: "2Gi"
    cpu: "1"
  limits:
    memory: "4Gi"
    cpu: "2"
```

### Chroma
```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "0.25"
  limits:
    memory: "1Gi"
    cpu: "0.5"
```

---

## Migration Strategy

### For Existing Users (Client0)

1. **No Breaking Changes**: Milvus remains default
2. **Opt-In**: Users choose to switch VDBs
3. **Clear Docs**: Migration guides for each VDB
4. **Testing**: Provide test scripts for each VDB

### For New Users

1. **Guided Choice**: CLI prompts for VDB selection
2. **Smart Defaults**: Recommend VDB based on use case
3. **Easy Switch**: Can change VDB by editing config

---

## Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| Phase 1: VDB Abstraction | 1 week | Abstract port-forward, logs, status |
| Phase 2: Helm Templates | 1 week | Templates for Qdrant, Weaviate, Chroma |
| Phase 3: Cloud VDB | 1 week | Support cloud endpoints |
| Phase 4: Documentation | 1 week | Complete docs and examples |

**Total**: 4 weeks (can run parallel with Phase 2 cloud work)

---

## Success Criteria

- ✅ User can deploy any VDB with `weave stack up`
- ✅ Commands abstract VDB type (port-forward vectordb)
- ✅ All 4 templates work with any VDB
- ✅ Cloud VDB connections work
- ✅ Complete documentation for each VDB
- ✅ Migration guides available
- ✅ No regression for Milvus users

---

## Open Questions

1. **VDB Selection UI**: Should `weave stack init` prompt for VDB choice?
2. **VDB Migration**: Tool to migrate data between VDBs?
3. **Performance Testing**: Benchmark each VDB with same dataset?
4. **Cost Analysis**: Document cloud VDB pricing?

---

## Related Issues

- User feedback: "Client0 may want to use other VDBs"
- User feedback: "Abstract the VDB since we want weave stack to work for all supported VDB"
- User feedback: "Support all the VDBs that weave support"

---

**Next Steps**:
1. Review and approve this plan
2. Add to Phase 2 or Phase 3 roadmap
3. Create GitHub issues for each phase
4. Prioritize based on user demand

**Status**: Ready for review
