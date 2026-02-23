# Weave Stack - Work Plan

**Current Status**: Phase 1 Days 1-3 Complete ✅ | v0.9.31 Released
**Next**: Phase 1 Days 4-5 (Helm deployment, kubectl integration)

---

## Tonight (Feb 23, 2026)

**Goal**: Archive completed work, plan tomorrow

### Tasks:
- [x] Archive Phase 1 Days 1-3 work
- [x] Tag v0.9.31 release
- [x] Update WEAVE_STACK_PHASE_1.md with completion status
- [ ] Review Phase 1 Day 4 plan
- [ ] Prepare for Helm deployment implementation

**Time Estimate**: 30 minutes

---

## Tomorrow (Feb 24, 2026)

**Goal**: Complete Phase 1 Day 4 - Helm deployment and health checks

### Morning Session (2-3 hours):
1. **Implement Helm install in `weave stack up`** (1 hour)
   - Copy Helm chart templates to kubernetes/ directory
   - Run `helm install weave-stack kubernetes/ --wait --timeout 5m`
   - Handle helm command errors

2. **Wait for pod readiness** (30 min)
   - Poll `kubectl get pods` until Milvus pod is Ready
   - Show progress spinner/indicator
   - Timeout after 5 minutes

3. **Update `weave stack status`** (30 min)
   - Query pod status via kubectl
   - Show table with Component, Status, Health, Uptime
   - Parse kubectl output

4. **Testing** (30 min)
   - Manual test: `weave stack up --runtime kind`
   - Verify Milvus pod reaches Ready state
   - Test `weave stack status` output

### Afternoon Session (2-3 hours):
5. **Implement `weave stack logs`** (1 hour)
   - Stream logs from Milvus pod
   - Support `--follow` flag
   - Handle pod not found errors

6. **Error handling improvements** (1 hour)
   - Helm install failures
   - Pod startup failures
   - Clear error messages with troubleshooting hints

7. **Write tests** (1 hour)
   - Helm install logic tests
   - Pod status parsing tests
   - Integration test script

**Deliverables**:
- `weave stack up --runtime kind` deploys Milvus successfully
- Milvus pod reaches Ready state
- `weave stack status` shows component health
- `weave stack logs milvus` streams logs
- Tests passing

---

## Rest of Week (Feb 25-28, 2026)

### Wednesday (Feb 25): Phase 1 Day 5
**Goal**: kubectl integration and port forwarding

**Tasks**:
- [ ] Implement `weave stack kubectl -- <args>` passthrough
- [ ] Implement `weave stack port-forward <service> <port>`
- [ ] Test port-forwarding to Milvus (19530)
- [ ] Add dependency checks (kubectl, helm, kind installed)

**Time**: 4-5 hours

### Thursday (Feb 26): Polish & Documentation
**Goal**: Production-ready Phase 1

**Tasks**:
- [ ] Better progress indicators (spinners)
- [ ] Color-coded output
- [ ] Complete help text for all commands
- [ ] Write Phase 1 summary documentation
- [ ] Update main README with Weave Stack section

**Time**: 3-4 hours

### Friday (Feb 27): Integration Testing & Demo
**Goal**: End-to-end testing and demo

**Tasks**:
- [ ] Write integration test script (`test-stack-phase1.sh`)
- [ ] Test on clean environment
- [ ] Create demo GIF/video
- [ ] Tag Phase 1 complete (v0.10.0?)
- [ ] Plan Phase 2 kickoff

**Time**: 3-4 hours

---

## Completed Work (Archive)

### ✅ Phase 1 Day 1 (Feb 23, 2026)
**Commits**:
- `4925d14` - Core types, config parsing
- `ca5b870` - Init and validate commands

**Deliverables**:
- `weave stack init` with 4 templates
- `weave stack validate`
- Core type system (StackConfig, RuntimeConfig, etc.)
- Config parsing with defaults (podman-first)
- 12 unit tests

**Files Created**:
- src/cmd/stack.go
- src/cmd/stack/init.go
- src/cmd/stack/validate.go
- src/pkg/stack/types.go
- src/pkg/stack/config.go
- src/pkg/stack/config_test.go

### ✅ Phase 1 Day 2 (Feb 23, 2026)
**Commits**:
- `6453355` - Cluster management and up/down/status commands

**Deliverables**:
- Kind cluster creation with podman provider
- Minikube cluster creation
- Runtime detection (podman-first)
- `weave stack up/down/status` commands
- Cluster state persistence (.weave-state/cluster.json)
- 15 unit tests

**Files Created**:
- src/pkg/stack/runtime.go
- src/pkg/stack/runtime_test.go
- src/pkg/stack/cluster.go
- src/pkg/stack/cluster_test.go
- src/cmd/stack/up.go
- src/cmd/stack/down.go
- src/cmd/stack/status.go

### ✅ Phase 1 Day 3 (Feb 23, 2026)
**Commits**:
- `ca50bf2` - Helm chart generation
- `7cfc262` - YAML lint fixes

**Deliverables**:
- Helm chart structure (Chart.yaml, values.yaml, templates)
- Dynamic values.yaml generation from weave-stack.yaml
- Milvus Deployment with health checks
- Milvus Service (ClusterIP)
- PVC for data persistence
- 8 unit tests

**Files Created**:
- src/pkg/stack/helm.go
- src/pkg/stack/helm_test.go
- templates/helm/weave-stack/Chart.yaml
- templates/helm/weave-stack/values.yaml
- templates/helm/weave-stack/templates/*.yaml

**Release**: v0.9.31

---

## Quick Reference

### Current Commands:
```bash
# Working commands (v0.9.31):
weave stack init --template quickstart    # ✅
weave stack validate                      # ✅
weave stack up --runtime kind            # 🚧 (creates cluster, generates helm)
weave stack down                         # ✅
weave stack status                       # ✅ (shows cluster info)

# Coming soon (Day 4):
weave stack up --runtime kind            # Deploy Milvus to K8s
weave stack logs milvus                  # Stream logs

# Coming soon (Day 5):
weave stack kubectl -- get pods          # kubectl passthrough
weave stack port-forward milvus 19530    # Port forwarding
```

### Testing:
```bash
# Unit tests
go test ./src/pkg/stack/... -v

# Lint & build
./lint.sh && ./build.sh

# Manual test
cd /tmp/test-stack
weave stack init
weave stack validate
weave stack up --runtime kind
```

---

## Dependencies Status

**Installed**:
- ✅ Go 1.24.1
- ✅ kubectl
- ✅ kind
- ✅ podman 5.6.2
- ✅ docker (fallback)

**Needed for Day 4**:
- ✅ helm (need to verify version)

---

## Notes

- **Podman-first**: All templates default to podman, docker as fallback
- **Kind provider**: Using `KIND_EXPERIMENTAL_PROVIDER=podman`
- **Milvus version**: 2.3.0 (embedded etcd, standalone mode)
- **Health checks**: liveness (30s delay) + readiness (10s delay)
- **Storage**: PVC 50Gi default, StorageClass "standard"

---

**Updated**: Feb 23, 2026 20:55 PST
