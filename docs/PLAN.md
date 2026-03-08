# Weave CLI Development Plan

**Last Updated**: 2026-03-07
**Current Version**: v0.11.2

---

## 🎯 Current Status (v0.11.2)

### Recently Shipped ✅
- **v0.11.2** (2026-03-07): Critical Weaviate bug fix
  - Fixed Weaviate backups to include vector embeddings (Issue #52)
  - Same root cause as Milvus Issue #51
  - Tested with DemoDocs (38 docs, 1024-dim vectors)
  - File size: 1.11 MB (was 324 KB without embeddings)
  - Both major VDBs (Milvus + Weaviate) now working correctly

- **v0.11.1** (2026-03-06): Critical Milvus bug fix
  - Fixed Milvus backups to include vector embeddings (Issue #51)
  - Tested with AuctionResults (127 docs, 1536-dim vectors)
  - Validation now passes

- **v0.11.0** (2026-03-06): Backup & Restore system
  - 4 commands: create, restore, validate, list
  - Portable .weavebak format with compression
  - Works with all 15+ VDBs
  - Performance: 195-272 docs/sec backup, 18 docs/sec restore

### Active Focus
- **Client0 Feedback**: Monitoring production usage with 2,636+ doc datasets
- **Performance Validation**: Gathering real-world metrics
- **Bug Fixes**: Immediate response to any issues

---

## 📅 Roadmap

### v0.12.0 - Remote Storage & Performance (2-3 weeks)

**Priority**: High
**Target**: 2026-03-20

**Features**:

1. **Remote Storage Integration** ☁️
   - Direct upload to S3/MinIO/GCS during backup
   - Download from remote storage for restore
   - Configuration via flags or config.yaml
   - **Impact**: Automated backup workflows, disaster recovery

2. **Performance Optimizations** 🚀
   - Parallel batch processing during backup
   - Adjustable batch sizes per VDB type
   - Memory optimization for large backups
   - Stream processing to reduce memory footprint
   - **Target**: 500+ docs/sec backup, 50+ docs/sec restore

3. **Backup Scheduling** ⏰
   - Built-in cron-like scheduler
   - Retention policies (keep last N backups)
   - Automatic cleanup of old backups
   - **Impact**: Hands-free backup automation

**Estimated Effort**: 40-50 hours

### v0.12.1 - Backup Encryption (1-2 weeks)

**Priority**: High (Compliance)
**Target**: 2026-04-03

**Features**:

1. **AES-256 Encryption** 🔐
   - Password-based encryption for backups
   - Key-based encryption (file or env var)
   - Encrypted backups compatible with .weavebak.gz format
   - Decrypt on restore with password/key

2. **Secure Key Management**
   - Environment variable support
   - Key file support
   - Integration with system keychain (macOS/Linux)

3. **Audit Logging**
   - Log all backup/restore operations
   - Include timestamps, user, collection, result
   - JSON format for log aggregation

**Estimated Effort**: 20-30 hours

### v0.13.0 - Advanced Features (3-4 weeks)

**Priority**: Medium
**Target**: 2026-04-24

**Features**:

1. **Incremental Backups** 🔄
   - Backup only changed documents since last backup
   - Metadata tracking (last backup timestamp)
   - Merge incremental backups into full backup
   - **Impact**: 10-100x faster daily backups

2. **Backup Comparison & Diff** 🔍
   - Compare two backups
   - Show added/removed/modified documents
   - Generate diff report
   - Merge backups (selective restore)

3. **Multi-Collection Backup**
   - Single command to backup multiple collections
   - Parallel backup execution
   - Collection filtering (glob patterns)
   - **Use Case**: Backup entire VDB instance

**Estimated Effort**: 50-60 hours

### v1.0.0 - Production Ready (Q2 2026)

**Priority**: High
**Target**: 2026-06-01

**Goals**:

- **Stability**: <5 open bugs at any time
- **Performance**: <100ms average query latency
- **Coverage**: 99%+ test coverage for core operations
- **Documentation**: Complete API reference, migration guides
- **Security**: Security audit complete, OWASP compliance
- **VDB Support**: 15+ vector databases, all production-ready

**Quality Gates**:
- All integration tests passing
- Load testing complete (10K+ documents)
- Security scan passing
- Community feedback incorporated
- Migration tools for all VDB combinations

---

## 🏃 Sprint Planning

### Sprint 1 (Mar 7-13): Client0 Feedback & Quick Wins

**Duration**: 1 week
**Focus**: Stabilization and validation

**Tasks**:
- [ ] Monitor Client0's v0.11.1/v0.11.2 usage
- [ ] Gather performance metrics from production
- [x] **DONE (Mar 7)**: Fix critical Weaviate bug (Issue #52) - released v0.11.2
- [x] **DONE (Mar 7)**: Test backup/restore with Weaviate (verified embeddings working)
- [x] **DONE (Mar 7)**: Test backup/restore with Milvus (revalidated v0.11.1 fix)
- [x] **DONE (Mar 8)**: Fix metadata dimension detection (eliminates validation warnings)
- [x] **DONE (Mar 8)**: Improve restore messaging (clearer source→target VDB)
- [ ] Test backup/restore with other VDBs (Qdrant, Supabase) - deferred
- [ ] Add performance benchmarks to docs
- [ ] Improve error messages based on feedback

**Deliverables**:
- ✅ **v0.11.2 Released**: Critical Weaviate embedding fix (Mar 7)
- ✅ **Issue #52**: Created and documented Weaviate bug
- ✅ **Testing**: Verified both Milvus and Weaviate backup/restore with embeddings
- ✅ **Weekend Improvements**: Dimension detection + restore messaging (Mar 8)
- Performance report with real-world data (pending Client0 feedback)
- Updated documentation with Client0 metrics (pending)

### Sprint 2 (Mar 14-20): Remote Storage Foundation

**Duration**: 1 week
**Focus**: S3/MinIO integration

**Tasks**:
- [ ] Design remote storage abstraction layer
- [ ] Implement S3 storage backend
- [ ] Implement MinIO storage backend (compatible with S3 API)
- [ ] Add `--storage` flag to backup create/restore
- [ ] Add remote storage config to config.yaml
- [ ] Write integration tests
- [ ] Update documentation

**Deliverables**:
- S3/MinIO backup/restore working
- Configuration examples
- Updated BACKUP_RESTORE.md guide

### Sprint 3 (Mar 21-27): Performance Optimizations

**Duration**: 1 week
**Focus**: Speed improvements

**Tasks**:
- [ ] Profile backup create with 10K+ documents
- [ ] Implement parallel batch processing
- [ ] Add adjustable batch sizes per VDB
- [ ] Optimize memory usage (streaming)
- [ ] Benchmark improvements
- [ ] Update performance metrics in docs

**Deliverables**:
- 2-3x faster backups
- Lower memory footprint
- Performance comparison report

### Sprint 4 (Mar 28 - Apr 3): Encryption

**Duration**: 1 week
**Focus**: Security and compliance

**Tasks**:
- [ ] Implement AES-256 encryption
- [ ] Add password-based encryption
- [ ] Add key-based encryption
- [ ] Integrate with system keychain
- [ ] Write security tests
- [ ] Update documentation with security best practices

**Deliverables**:
- Encrypted backups working
- Security documentation
- Compliance guide for Client0

---

## 🎯 Client0 Priorities

Based on Client0's needs (to be confirmed):

### Immediate (v0.11.x)
1. ✅ Backup/restore working with embeddings
2. ⏳ Performance validation with 2,636+ doc datasets
3. ⏳ Stability confirmation in production

### Short-term (v0.12.0)
1. **Remote Storage**: Automated S3 backups for disaster recovery
2. **Performance**: Faster backups for daily snapshots
3. **Scheduling**: Cron-like automation for hands-free operation

### Mid-term (v0.12.1)
1. **Encryption**: Compliance with data protection requirements
2. **Audit Logging**: Track all backup operations
3. **Retention Policies**: Automatic cleanup of old backups

### Questions for Client0:
- Performance: "How long did backup/restore take for your datasets?"
- Use Case: "Are you using this for disaster recovery, dev/test, or migration?"
- Pain Points: "What's the biggest limitation right now?"
- Priority: "What would make backup/restore 10x better for you?"

---

## 🔧 Technical Debt

### High Priority
- [ ] Add backup/restore integration tests for all VDBs
- [ ] Profile memory usage with large datasets
- [ ] Add retry logic for network failures
- [ ] Implement resume for interrupted backups

### Medium Priority
- [ ] Add progress streaming for large operations
- [ ] Optimize JSON serialization/deserialization
- [ ] Add compression level control
- [ ] Implement backup verification (checksum)

### Low Priority
- [ ] Add backup format versioning strategy
- [ ] Add backup metadata queries
- [ ] Implement backup deduplication
- [ ] Add backup statistics/analytics

---

## 📊 Success Metrics

### v0.12.0 Targets
- Backup speed: 500+ docs/sec (2x improvement)
- Restore speed: 50+ docs/sec (3x improvement)
- Memory usage: <500MB for 10K docs
- Remote storage: S3/MinIO fully functional
- Test coverage: 95%+ for backup module

### v1.0.0 Targets
- VDB coverage: 15+ databases, all production-ready
- Performance: <100ms average operation
- Uptime: 99.9% for production deployments
- Community: 100+ stars, 10+ contributors
- Documentation: 100% API coverage

---

## 🚀 Quick Wins (Can Do Anytime)

These are small improvements that can be done in 1-2 hours:

1. **Better Error Messages**
   - Add suggestions for common errors
   - Include troubleshooting links
   - Show progress on failures

2. **Validation Improvements**
   - Add checksum verification
   - Warn on old backup format versions
   - Check available disk space before restore

3. **Documentation**
   - Add video tutorials
   - Add more real-world examples
   - Create migration guides for each VDB pair

4. **Testing**
   - Add chaos testing (network failures, disk full)
   - Add load tests with 100K+ documents
   - Test with different embedding dimensions

---

## 📝 Notes

### Lessons Learned (v0.11.x)
1. **Vector Fields Are Special**: Both Milvus and Weaviate require explicit requests
   - Milvus: `"*"` wildcard does NOT include vector fields
   - Weaviate: Must include `_additional { vector }`
2. **Actual > Config**: Always detect actual values from data, not config defaults
3. **Testing**: Need collections with actual embeddings for E2E tests (DemoDocs: 38 docs perfect)
4. **Documentation**: Comprehensive guides are essential for adoption
5. **Compression**: 65-95% reduction makes a huge difference for large collections
6. **UX Matters**: Clear messaging (source→target VDB) as important as bug fixes

### Architecture Decisions
1. **File Format**: JSON + gzip for portability and debugging
2. **Batch Size**: 100 docs default, adjustable for performance tuning
3. **VDB Abstraction**: Use VectorDBClient interface for all operations
4. **Error Handling**: Fail fast, provide clear error messages

### Community Feedback
- (To be updated as feedback comes in)

---

## 📅 Next Week (Mar 10-14)

**Focus**: Complete Sprint 1, gather Client0 feedback, plan Sprint 2

### Monday (Mar 10)
- Check GitHub for weekend feedback
- Contact Client0 for v0.11.2 usage reports
- Update PLAN.md with Client0 input
- Decide: Release v0.11.3 or continue to Sprint 2

### Tuesday-Wednesday (Mar 11-12)
- Fix any critical bugs (<24h response)
- Create performance report with real-world data
- Update documentation with learnings
- Test other VDBs if time permits

### Thursday-Friday (Mar 13-14)
- Sprint 1 retrospective
- Sprint 2 detailed planning
- Begin remote storage design
- Set up S3/MinIO development environment

---

**Maintained by**: @maximilien
**Next Review**: 2026-03-13 (after Client0 feedback)
