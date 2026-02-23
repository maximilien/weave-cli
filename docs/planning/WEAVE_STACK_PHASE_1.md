# Weave Stack - Phase 1 Implementation

**Status**: 🚀 In Progress
**Goal**: `weave stack up --runtime kind` deploys Milvus to local K8s with podman
**Timeline**: Week 1 (Feb 24 - Mar 2, 2026)
**Target**: Prototype by end of week

---

## Objectives

**Primary Goal**:
Get a minimal working stack deployed to Kind cluster using podman:
- Kind cluster creation with podman driver
- Helm chart generation from `weave-stack.yaml`
- Milvus deployment to K8s
- Basic health checks

**Success Criteria**:
```bash
weave stack init
weave stack up --runtime kind
weave stack status  # Shows Milvus pod: RUNNING
kubectl get pods    # Verifies Milvus is healthy
```

---

## Day-by-Day Breakdown

### Day 1 (Mon Feb 24): Project Structure + Core Types

**Tasks**:
- [ ] Create `src/cmd/stack/` directory structure
- [ ] Define core types in `src/pkg/stack/types.go`
- [ ] Implement `weave stack init` (generate minimal YAML)
- [ ] Basic YAML parsing

**Deliverables**:
```go
// src/pkg/stack/types.go
type StackConfig struct {
    Version string `yaml:"version"`
    Name    string `yaml:"name"`
    Runtime RuntimeConfig `yaml:"runtime"`
    Infrastructure InfrastructureConfig `yaml:"infrastructure"`
}

type RuntimeConfig struct {
    Kubernetes KubernetesConfig `yaml:"kubernetes"`
    ContainerRuntime string `yaml:"container_runtime"` // "podman", "docker"
}

type KubernetesConfig struct {
    Provider string `yaml:"provider"` // "kind", "minikube", "eks", "gke"
    Kind     *KindConfig `yaml:"kind,omitempty"`
}
```

**Command**:
```bash
weave stack init --template quickstart

# Generates:
# - weave-stack.yaml (minimal config)
# - kubernetes/Chart.yaml (Helm skeleton)
# - .gitignore
```

**Testing**:
- Parse generated YAML → verify no errors
- Validate required fields present

---

### Day 2 (Tue Feb 25): Kind Cluster Management

**Tasks**:
- [ ] Implement `weave stack cluster create --runtime kind`
- [ ] Detect podman vs docker
- [ ] Create Kind cluster with podman provider
- [ ] Store cluster info in `.weave-state/cluster.json`

**Deliverables**:
```go
// src/pkg/stack/cluster.go
func CreateKindCluster(config *StackConfig) error {
    // 1. Detect podman or docker
    runtime := detectContainerRuntime()

    // 2. Generate kind config YAML
    kindConfig := generateKindConfig(config)

    // 3. Run: kind create cluster --name weave-stack --config kind.yaml
    //    Set KIND_EXPERIMENTAL_PROVIDER=podman if podman detected

    // 4. Save cluster info to .weave-state/cluster.json
}

func detectContainerRuntime() string {
    // Try: podman version
    // Fallback: docker version
}
```

**Command**:
```bash
weave stack cluster create --runtime kind

# Output:
# 🔍 Detecting container runtime...
# ✅ Found: podman (version 4.x)
#
# 🐳 Creating Kind cluster 'weave-stack'...
# ✅ Cluster created successfully
#
# 📝 Saved cluster info to .weave-state/cluster.json
#
# Next: weave stack up
```

**Testing**:
- Verify Kind cluster exists: `kind get clusters`
- Verify podman containers running: `podman ps`
- Verify kubectl works: `kubectl cluster-info`

---

### Day 3 (Wed Feb 26): Helm Chart Generation

**Tasks**:
- [ ] Create base Helm chart in `templates/helm/weave-stack/`
- [ ] Generate `values.yaml` from `weave-stack.yaml`
- [ ] Implement Milvus deployment template
- [ ] Add PVC for Milvus data

**Deliverables**:
```yaml
# templates/helm/weave-stack/Chart.yaml
apiVersion: v2
name: weave-stack
description: RAG Stack for weave-cli
version: 0.1.0
appVersion: "1.0"

# templates/helm/weave-stack/values.yaml (generated from weave-stack.yaml)
vectordb:
  type: milvus
  version: "2.3.0"
  replicas: 1
  resources:
    requests:
      memory: "8Gi"
      cpu: "2"
    limits:
      memory: "12Gi"
      cpu: "4"
  storage:
    class: "standard"
    size: "50Gi"

# templates/helm/weave-stack/templates/vectordb.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-milvus
spec:
  replicas: {{ .Values.vectordb.replicas }}
  selector:
    matchLabels:
      app: milvus
  template:
    metadata:
      labels:
        app: milvus
    spec:
      containers:
      - name: milvus
        image: milvusdb/milvus:{{ .Values.vectordb.version }}
        resources:
          requests:
            memory: {{ .Values.vectordb.resources.requests.memory }}
            cpu: {{ .Values.vectordb.resources.requests.cpu }}
          limits:
            memory: {{ .Values.vectordb.resources.limits.memory }}
            cpu: {{ .Values.vectordb.resources.limits.cpu }}
        ports:
        - containerPort: 19530
        livenessProbe:
          httpGet:
            path: /healthz
            port: 9091
          initialDelaySeconds: 30
          periodSeconds: 30
        volumeMounts:
        - name: milvus-data
          mountPath: /var/lib/milvus
      volumes:
      - name: milvus-data
        persistentVolumeClaim:
          claimName: {{ .Release.Name }}-milvus-pvc
```

**Code**:
```go
// src/pkg/stack/helm.go
func GenerateHelmChart(config *StackConfig, outputDir string) error {
    // 1. Create kubernetes/ directory
    // 2. Copy templates/helm/weave-stack/ to kubernetes/
    // 3. Generate values.yaml from StackConfig
    // 4. Write to kubernetes/values.yaml
}

func GenerateHelmValues(config *StackConfig) (map[string]interface{}, error) {
    return map[string]interface{}{
        "vectordb": map[string]interface{}{
            "type":     config.Infrastructure.VectorDB.Type,
            "version":  config.Infrastructure.VectorDB.Version,
            "replicas": config.Infrastructure.VectorDB.Kubernetes.Deployment.Replicas,
            "resources": config.Infrastructure.VectorDB.Resources,
            "storage": map[string]interface{}{
                "class": config.Infrastructure.VectorDB.Kubernetes.Deployment.StorageClass,
                "size":  config.Infrastructure.VectorDB.Kubernetes.Deployment.PVCSize,
            },
        },
    }, nil
}
```

**Testing**:
- Render Helm chart: `helm template kubernetes/`
- Verify valid K8s manifests
- Check values substitution works

---

### Day 4 (Thu Feb 27): Deployment + Health Checks

**Tasks**:
- [ ] Implement `weave stack up --runtime kind`
- [ ] Helm install to Kind cluster
- [ ] Wait for pod readiness
- [ ] Implement `weave stack status`

**Deliverables**:
```go
// src/cmd/stack/up.go
func StackUp(runtime string, skipIngestion bool) error {
    // 1. Validate weave-stack.yaml exists
    config := loadStackConfig()

    // 2. Ensure cluster exists (or create)
    if !clusterExists(config) {
        createCluster(config)
    }

    // 3. Generate Helm chart
    generateHelmChart(config, "kubernetes/")

    // 4. Helm install
    // helm install weave-stack kubernetes/ --wait --timeout 5m

    // 5. Wait for pods to be Ready
    waitForPods("app=milvus", 5*time.Minute)

    // 6. Show status
    showStatus()
}

// src/cmd/stack/status.go
func StackStatus(watch bool) error {
    // kubectl get pods -l app.kubernetes.io/instance=weave-stack
    // Parse output, show table:
    // Component   Status    Health   Uptime
    // milvus      RUNNING   ✅       2m 34s
}
```

**Commands**:
```bash
weave stack up --runtime kind

# Output:
# ════════════════════════════════════════
# 🚀 Weave Stack Deployment
# ════════════════════════════════════════
#
# [1/4] 🔍 Checking cluster...
#   ✅ Kind cluster 'weave-stack' exists
#
# [2/4] 📦 Generating Helm chart...
#   ✅ Generated kubernetes/values.yaml
#   ✅ Rendered 3 K8s manifests
#
# [3/4] 🚀 Deploying to Kubernetes...
#   helm install weave-stack kubernetes/ --wait
#   ⏳ Waiting for pods to be Ready... (2m 15s)
#   ✅ All pods Running
#
# [4/4] 📊 Status:
#   Component   Status     Health
#   ─────────────────────────────
#   milvus      RUNNING    ✅
#
# ════════════════════════════════════════
# 🎉 Stack deployed successfully!
#
# Next steps:
#   - Check status: weave stack status
#   - View logs: weave stack logs milvus
#   - Access Milvus: kubectl port-forward svc/weave-stack-milvus 19530:19530
```

**Testing**:
- Deploy to Kind → verify Milvus pod Running
- Check health: `curl http://localhost:9091/healthz` (via port-forward)
- Verify PVC created and bound

---

### Day 5 (Fri Mar 1): kubectl Integration + Port Forwarding

**Tasks**:
- [ ] Implement `weave stack kubectl -- <args>`
- [ ] Implement `weave stack port-forward <service> <port>`
- [ ] Implement `weave stack logs <service>`
- [ ] Add basic error handling

**Deliverables**:
```go
// src/cmd/stack/kubectl.go
func KubectlPassthrough(args []string) error {
    // Run: kubectl <args> --context kind-weave-stack
    // Pass through stdout/stderr
}

// src/cmd/stack/port_forward.go
func PortForward(service string, portMapping string) error {
    // kubectl port-forward svc/<service> <portMapping>
    // Keep running in foreground
}

// src/cmd/stack/logs.go
func Logs(service string, follow bool) error {
    // kubectl logs -l app=<service> --follow=<follow>
}
```

**Commands**:
```bash
# Direct kubectl access
weave stack kubectl -- get pods
weave stack kubectl -- describe svc weave-stack-milvus

# Port forwarding (access Milvus locally)
weave stack port-forward milvus 19530:19530
# Now accessible at localhost:19530

# View logs
weave stack logs milvus --follow
```

**Testing**:
- Verify kubectl commands work
- Port-forward → test connection from weave-cli
- Logs show Milvus startup messages

---

### Weekend (Sat-Sun Mar 1-2): Polish + Documentation

**Tasks**:
- [ ] Error handling improvements
- [ ] Add `weave stack down` (helm uninstall)
- [ ] Add `weave stack validate` (YAML validation)
- [ ] Write Phase 1 summary doc
- [ ] Create demo video/GIF

**Polish**:
- Better error messages
- Progress indicators (spinners)
- Color-coded output
- `--help` text for all commands

**Documentation**:
- Update README with Phase 1 status
- Add examples to docs/
- Document troubleshooting (podman issues, Kind issues)

---

## File Structure

```
src/
├── cmd/
│   └── stack/
│       ├── init.go          # weave stack init
│       ├── cluster.go       # weave stack cluster create/delete
│       ├── up.go            # weave stack up
│       ├── down.go          # weave stack down
│       ├── status.go        # weave stack status
│       ├── validate.go      # weave stack validate
│       ├── kubectl.go       # weave stack kubectl
│       ├── port_forward.go  # weave stack port-forward
│       └── logs.go          # weave stack logs
│
├── pkg/
│   └── stack/
│       ├── types.go         # Core types (StackConfig, etc.)
│       ├── config.go        # YAML parsing
│       ├── cluster.go       # Kind/Minikube cluster mgmt
│       ├── helm.go          # Helm chart generation
│       ├── kubernetes.go    # kubectl interactions
│       └── runtime.go       # Podman/Docker detection
│
└── templates/
    └── helm/
        └── weave-stack/
            ├── Chart.yaml
            ├── values.yaml
            └── templates/
                ├── vectordb.yaml
                ├── pvc.yaml
                └── service.yaml
```

---

## Testing Strategy

### Unit Tests
- YAML parsing (`config_test.go`)
- Helm values generation (`helm_test.go`)
- Runtime detection (`runtime_test.go`)

### Integration Tests
```bash
# scripts/test-stack-phase1.sh
#!/bin/bash
set -e

echo "Phase 1 Integration Test"

# 1. Init
weave stack init --template quickstart
test -f weave-stack.yaml

# 2. Create cluster
weave stack cluster create --runtime kind
kind get clusters | grep weave-stack

# 3. Deploy
weave stack up --runtime kind --skip-ingestion
kubectl get pods | grep milvus | grep Running

# 4. Status
weave stack status | grep milvus | grep RUNNING

# 5. Cleanup
weave stack down
weave stack cluster delete

echo "✅ Phase 1 integration test passed!"
```

### Manual Testing Checklist
- [ ] podman detected correctly
- [ ] Docker fallback works
- [ ] Kind cluster creates successfully
- [ ] Helm chart renders without errors
- [ ] Milvus pod becomes Ready within 5 minutes
- [ ] PVC bound correctly
- [ ] Health check passes
- [ ] Port-forward works
- [ ] kubectl passthrough works
- [ ] Logs streaming works
- [ ] `weave stack down` cleans up correctly

---

## Dependencies

**External Tools** (auto-detected, print instructions if missing):
- `kubectl` - Kubernetes CLI
- `helm` - Kubernetes package manager
- `kind` - Kubernetes in Docker
- `podman` or `docker` - Container runtime

**Go Libraries**:
- `gopkg.in/yaml.v3` - YAML parsing
- `k8s.io/client-go` - Kubernetes Go client (optional, can use kubectl CLI)
- `helm.sh/helm/v3` - Helm Go SDK (optional, can use helm CLI)

**Installation Check**:
```go
// src/pkg/stack/dependencies.go
func CheckDependencies() error {
    required := []string{"kubectl", "helm", "kind"}

    for _, tool := range required {
        if !commandExists(tool) {
            return fmt.Errorf("%s not found. Install: https://...", tool)
        }
    }

    // Check podman or docker
    if !commandExists("podman") && !commandExists("docker") {
        return fmt.Errorf("neither podman nor docker found")
    }

    return nil
}
```

---

## Known Issues / Limitations

**Phase 1 Scope**:
- ✅ Deploy Milvus to Kind
- ❌ No ingestion (Phase 2)
- ❌ No dashboard (Phase 5)
- ❌ No evaluations (Phase 6)
- ❌ Milvus only (no Qdrant/Weaviate yet)
- ❌ No MinIO (local PVC only)

**podman Limitations**:
- Kind with podman is experimental
- May require `KIND_EXPERIMENTAL_PROVIDER=podman`
- Rootless podman may have networking issues

**Workarounds**:
- If podman fails, auto-fallback to docker
- Print clear error messages with links to docs

---

## Success Metrics

**Quantitative**:
- [ ] `weave stack up` completes in <3 minutes
- [ ] Milvus pod reaches Ready state
- [ ] Zero manual kubectl/helm commands needed
- [ ] Works on macOS + Linux

**Qualitative**:
- [ ] Commands feel intuitive (similar to `docker-compose up`)
- [ ] Error messages are helpful
- [ ] Progress is visible

---

## Next Steps (Phase 2 Preview)

After Phase 1 is complete:
1. **Ingestion as K8s Job** - Deploy data ingestion workload
2. **Checkpoint ConfigMap** - Resume on failure
3. **Monitor Job progress** - Real-time feedback

**Phase 2 Goal**: `weave stack up` deploys VDB + ingests 3 PDFs

---

## Daily Standup Format

**Each day, update here**:

### Day 1 Status (Mon Feb 24):
- [ ] Core types defined
- [ ] `weave stack init` working
- [ ] YAML parsing tested

### Day 2 Status (Tue Feb 25):
- [ ] Kind cluster creation working
- [ ] podman detection working
- [ ] Cluster info saved

### Day 3 Status (Wed Feb 26):
- [ ] Helm chart generated
- [ ] Milvus template created
- [ ] Values substitution working

### Day 4 Status (Thu Feb 27):
- [ ] `weave stack up` working
- [ ] Milvus pod Running
- [ ] Status command working

### Day 5 Status (Fri Mar 1):
- [ ] kubectl passthrough working
- [ ] Port forwarding working
- [ ] Logs streaming working

---

**Let's build this! LFG! 🚀**
