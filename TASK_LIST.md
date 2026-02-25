# Weave CLI Task List

**Current Status**: v0.10.2 - Production Ready for Client0
**Last Updated**: Feb 25, 2026 (21:00 PST)

---

## 🎯 NEXT HOUR (Tonight - 22:00 PST)

**Time Remaining**: ~60 minutes

### Priority 1: Quick Wins (30 min)

- [ ] **Audit `weave -h` and `weave stack -h` help text** (15 min)
  - Check all stack subcommands have clear help
  - Ensure examples are accurate
  - Update any outdated flag descriptions
  - Commands to check:
    ```bash
    weave -h
    weave stack -h
    weave stack init -h
    weave stack up -h
    weave stack ingest -h
    ```

- [ ] **Quick README polish** (15 min)
  - Ensure Weave Stack section is accurate
  - Add v0.10.2 highlights
  - Check all examples work
  - Link to CLIENT0_GETTING_STARTED.md

### Priority 2: Optional Polish (30 min)

- [ ] **PM2 Dashboard smoke test** (15 min)
  - Init production template
  - Generate ecosystem.config.js
  - Verify matches Client0 expectations
  - Port 3100, TypeScript, health monitoring

- [ ] **Test demo script** (15 min)
  - Run `./demos/client0-demo.sh`
  - Ensure it works end-to-end
  - Fix any issues
  - Time it (should be < 5 min)

---

## 📅 TOMORROW (Client0 Meeting Day)

**Priority**: Client0 meeting prep + start Phase 2

### Morning (Before Meeting)

**Time**: 2-3 hours before meeting

- [ ] **Review demo script** (15 min)
  - Practice the demo
  - Prepare talking points
  - Have backup plan if live demo fails

- [ ] **Prepare meeting materials** (30 min)
  - Review docs/CLIENT0_GETTING_STARTED.md
  - Review v0.10.2 release notes
  - Prepare Phase 2 roadmap overview
  - List of known issues/limitations

- [ ] **Test Client0's data** (30 min)
  - Use sample from /Users/maximilien/github/auctionsmax-ai/data/tamarkin
  - Ingest small subset (5-10 files)
  - Verify queries work
  - Note any issues

### After Client0 Meeting

**Time**: 4-6 hours

- [ ] **Incorporate Client0 feedback** (1 hour)
  - Document any issues they report
  - Create GitHub issues for bugs
  - Update documentation based on questions

- [ ] **Start Phase 2 - EKS Schema** (3-5 hours)
  - Update src/pkg/stack/types.go with EKSConfig
  - Add validation for EKS fields
  - Update templates with EKS defaults
  - Test schema validation

---

## 🗓️ FRIDAY (Feb 26)

**Goal**: EKS cluster creation working

### Morning (4 hours)

- [ ] **EKS Cluster Creation** (4 hours)
  - Implement src/pkg/stack/eks.go
  - CreateEKSCluster() function
  - Use AWS SDK or eksctl wrapper
  - Wait for cluster ready
  - Configure kubectl context

### Afternoon (4 hours)

- [ ] **EKS Helm Configuration** (2 hours)
  - Add EKS-specific Helm values
  - LoadBalancer service type
  - EBS storage class
  - IAM service accounts

- [ ] **Initial EKS Testing** (2 hours)
  - Test cluster creation
  - Verify kubectl access
  - Document setup requirements
  - Note any issues

---

## 🏖️ WEEKEND (Feb 27-28)

**Goal**: Complete EKS support, start GKE

### Saturday (6-8 hours)

#### Morning

- [ ] **EKS End-to-End Testing** (3 hours)
  - Deploy Milvus to EKS
  - Test full workflow
  - Fix any bugs
  - Document workarounds

#### Afternoon

- [ ] **EKS Documentation** (1 hour)
  - Create EKS setup guide
  - Prerequisites and IAM requirements
  - Troubleshooting section

- [ ] **GKE Schema Design** (2 hours)
  - Update types.go with GKEConfig
  - Add validation
  - Update templates

### Sunday (6-8 hours)

#### Morning

- [ ] **GKE Cluster Creation** (4 hours)
  - Implement src/pkg/stack/gke.go
  - CreateGKECluster() function
  - Use gcloud SDK wrapper
  - Configure kubectl context

#### Afternoon

- [ ] **GKE Initial Testing** (2 hours)
  - Test cluster creation
  - Verify kubectl access
  - Note any issues

- [ ] **Weekly Summary** (1 hour)
  - Document Week 1 progress
  - Update Phase 2 plan
  - Prepare for Week 2

---

## 📋 BACKLOG (Phase 2 - Weeks 2-3)

### Week 2: Production Features

#### TLS/SSL Certificates (1-2 days)
- [ ] Integrate cert-manager
- [ ] Let's Encrypt automation
- [ ] Custom certificate support
- [ ] TLS termination at ingress

#### Secrets Management (1 day)
- [ ] Kubernetes secrets for API keys
- [ ] AWS Secrets Manager (EKS)
- [ ] GCP Secret Manager (GKE)
- [ ] Encrypted .env support

#### Monitoring & Observability (2-3 days)
- [ ] Prometheus deployment
- [ ] Grafana dashboards
- [ ] Milvus health dashboard
- [ ] Ingestion metrics dashboard
- [ ] Query performance dashboard
- [ ] Alerting rules

#### Backup & Restore (1-2 days)
- [ ] Milvus backup to S3/GCS
- [ ] Scheduled backups
- [ ] Restore command
- [ ] Point-in-time recovery

### Advanced Ingestion (Parallel)

- [ ] Resume from failures
- [ ] Checkpointing
- [ ] Deduplication
- [ ] Validation before insert
- [ ] Batch metrics
- [ ] Error reporting improvements
- [ ] Dry-run mode
- [ ] Metrics export (JSON/YAML)

---

## 📖 DOCS TO REVIEW

### Before Client0 Meeting (MUST READ)

**Priority Order**:

1. **docs/CLIENT0_GETTING_STARTED.md** ⭐⭐⭐
   - Your handoff guide
   - What Client0 will read
   - Make sure it's perfect

2. **docs/planning/PRODUCTION_READINESS_AUDIT_RESULTS.md** ⭐⭐
   - What bugs we found
   - What we fixed
   - Why v0.10.2 is ready

3. **README.md - Weave Stack section** ⭐⭐
   - What users see first
   - Examples must work
   - Clear and accurate

### Background Reading (Optional)

4. **docs/guides/WEAVE_STACK_QUICKSTART.md** ⭐
   - Comprehensive guide
   - Reference material

5. **docs/planning/WEAVE_STACK_PHASE_2_PLAN.md** ⭐
   - What's coming next
   - For roadmap discussions

6. **docs/planning/MINIKUBE_VALIDATION_RESULTS.md**
   - Technical details
   - Known limitations

---

## 🎬 DEMO PREPARATION

### Before Running Demo

**Prerequisites checklist**:
- [ ] kubectl installed
- [ ] helm installed
- [ ] kind installed
- [ ] OPENAI_API_KEY set
- [ ] Previous Kind clusters deleted
- [ ] Docker or podman running

### Demo Script

**Location**: `demos/client0-demo.sh`

**What it does**:
1. Initialize quickstart stack
2. Deploy to Kind
3. Create sample data
4. Ingest documents
5. Query with semantic search
6. Clean teardown

**Time**: < 5 minutes

**Practice**: Run it 2-3 times before meeting

### Backup Plan

If live demo fails:
1. Show pre-recorded demo video (create one!)
2. Walk through code and docs
3. Show test results instead

---

## 🐛 KNOWN ISSUES TO MENTION

### Client0 Should Know

**Limitations (Document in Meeting)**:
1. Local K8s only (Kind/Minikube) - EKS/GKE coming in Phase 2
2. Minikube has environmental constraints (use Kind recommended)
3. Minimum 8GB RAM for local deployment
4. Single-node clusters by default

**NOT Issues** (Fixed in v0.10.2):
- ✅ Template path resolution
- ✅ Resource allocation
- ✅ Milvus startup

---

## 📊 SUCCESS METRICS

### By End of Week

- [ ] Client0 successfully deploys their first stack
- [ ] EKS cluster creation working
- [ ] GKE schema designed
- [ ] Phase 2 Week 1 on track

### By End of Phase 2

- [ ] `weave stack up --runtime eks` works
- [ ] `weave stack up --runtime gke` works
- [ ] TLS/SSL automated
- [ ] Monitoring dashboards live
- [ ] v0.11.0 tagged

---

## 🔗 QUICK LINKS

**Code**:
- Main: README.md
- Stack: src/cmd/stack/, src/pkg/stack/
- Templates: templates/helm/weave-stack/

**Docs**:
- Client0 guide: docs/CLIENT0_GETTING_STARTED.md
- Stack guide: docs/guides/WEAVE_STACK_QUICKSTART.md
- Phase 2 plan: docs/planning/WEAVE_STACK_PHASE_2_PLAN.md

**Planning**:
- Active docs: docs/planning/ (7 files)
- Archived: docs/archive/planning/ (40+ files)

**Tests**:
- Stack tests: `./test.sh stack`
- All tests: `./test.sh stack integration`
- Linting: `./lint.sh`

**Releases**:
- v0.10.2: Current (production ready)
- v0.10.1: Stack ingest
- v0.10.0: Phase 1 complete

---

## 💡 TIPS FOR CLIENT0 MEETING

### Do's
✅ Demo the full workflow
✅ Highlight the 3 bugs we fixed
✅ Show Getting Started guide
✅ Discuss Phase 2 roadmap
✅ Ask for feedback
✅ Take notes on pain points

### Don'ts
❌ Over-promise on timelines
❌ Skip known limitations
❌ Demo untested features
❌ Forget to ask questions

### Questions to Ask Client0

1. What's their deployment target? (Local? AWS? GCP?)
2. What's their data size/volume?
3. What embedding models do they use?
4. Do they need PM2 dashboard? (port 3100)
5. Any specific requirements we should know?

---

## 🎯 NEXT ACTIONS (Choose Your Path)

### Path A: Polish & Prep (Tonight + Tomorrow Morning)

**Best for**: Client0 meeting confidence

- Audit help text
- Polish README
- Test demo script
- Review docs
- Practice demo

**Time**: 3-4 hours total

### Path B: Start Phase 2 Early (Tonight)

**Best for**: Getting ahead on EKS

- Skip polish
- Start EKS schema now
- Get 4-6 hours head start
- Demo from docs instead

**Time**: 4-6 hours tonight

### Path C: Hybrid (Recommended)

**Best for**: Balance

- Polish help text (30 min)
- Test demo (15 min)
- Review docs (30 min)
- Start EKS schema (2-3 hours)

**Time**: 4 hours tonight

---

**Recommendation**: Path C (Hybrid)

**Why**: Client0 meeting is important (need polish) but EKS is complex (need head start)

---

**Last Updated**: Feb 25, 2026 21:00 PST
**Next Update**: After Client0 meeting
