# Week 11 Plan - Backup/Restore Feature (Issue #43)

**Dates**: Mar 3-7, 2026
**Priority**: P0 - Client0 Blocking
**Target Release**: v0.11.0
**Status**: Ready to Start

---

## Goals

1. ✅ **Complete backup/restore MVP** (Issue #43)
2. ✅ **Fix stack ingestion bug** (Issue #44)
3. ✅ **Release v0.11.0 for Client0**

---

## Day-by-Day Plan

### Monday (Mar 3) - Stack Ingestion Fix

**Issue #44**: Stack ingestion pipeline bug (CRITICAL)

**Tasks**:
- [ ] Debug `pipeline.ProcessFiles()` VDB interaction
- [ ] Identify why VDB methods not called
- [ ] Fix pipeline configuration/logic
- [ ] Test end-to-end ingestion workflow
- [ ] Verify Milvus logs show CREATE COLLECTION + INSERT
- [ ] Remove debug logging
- [ ] Update quickstart guide

**Estimated Time**: 4-6 hours
**Deliverable**: Stack ingestion working ✅

---

### Tuesday (Mar 4) - Backup Core Implementation

**Issue #43**: Backup command (Phase 1, Part 1)

**Tasks**:
- [ ] Create `src/cmd/backup/` directory
- [ ] Implement `backup create` command
  - [ ] Connect to VDB
  - [ ] Query all documents in batches
  - [ ] Serialize to JSON
  - [ ] Write .weavebak file
- [ ] Create `src/pkg/backup/format.go`
  - [ ] BackupFormat struct
  - [ ] Serialization logic
  - [ ] Version handling
- [ ] Add progress tracking
- [ ] Test with small collection (10 docs)

**Estimated Time**: 6-8 hours
**Deliverable**: Basic backup working (no compression yet)

---

### Wednesday (Mar 5) - Restore + Compression

**Issue #43**: Restore command + compression (Phase 1, Part 2)

**Tasks**:
- [ ] Implement `backup restore` command
  - [ ] Read .weavebak file
  - [ ] Validate format
  - [ ] Create collection if doesn't exist
  - [ ] Batch insert documents
  - [ ] Verify restoration
- [ ] Add gzip compression support
  - [ ] Create `src/pkg/backup/compressor.go`
  - [ ] Compress on backup
  - [ ] Decompress on restore
- [ ] Add `backup validate` command
- [ ] Test with medium collection (100 docs)

**Estimated Time**: 6-8 hours
**Deliverable**: Full backup/restore cycle working ✅

---

### Thursday (Mar 6) - Testing + Validation

**Issue #43**: Client0 dataset testing

**Tasks**:
- [ ] Test with Client0's dataset (2,636 docs)
  - [ ] Backup all collections
  - [ ] Verify file sizes (~50-60MB compressed)
  - [ ] Time backup (should be < 2 min)
  - [ ] Delete collections
  - [ ] Restore from backups
  - [ ] Time restore (should be < 3 min)
  - [ ] Verify data integrity
- [ ] Add `backup list` command
- [ ] Error handling improvements
- [ ] Add unit tests
- [ ] Add integration tests

**Estimated Time**: 6-8 hours
**Deliverable**: Production-ready backup/restore ✅

---

### Friday (Mar 7) - Documentation + Release

**Release v0.11.0**

**Tasks**:
- [ ] Write user guide: `docs/guides/BACKUP_RESTORE_GUIDE.md`
  - [ ] Quick start examples
  - [ ] Common use cases
  - [ ] Troubleshooting
  - [ ] Client0-specific examples
- [ ] Update CLI help text
- [ ] Update main README.md
- [ ] Create release notes
- [ ] Tag v0.11.0
- [ ] Create GitHub release
- [ ] **Notify Client0** 🎉

**Estimated Time**: 4-5 hours
**Deliverable**: v0.11.0 released ✅

---

## Release: v0.11.0

### Target Date
**Friday, March 7, 2026**

### Version Number
**v0.11.0** - Feature release (backup/restore)

### Release Title
"Backup & Restore - Data Loss Prevention"

### Release Notes (Draft)

```markdown
# v0.11.0 - Backup & Restore

**Release Date**: March 7, 2026
**Priority**: P0 - Client0 Production Blocker

## 🎉 Major Features

### Backup/Restore Commands (Issue #43)

Prevent data loss and eliminate costly re-ingestion with new backup/restore commands:

```bash
# Backup a collection
weave backup create MyCollection --output backup.weavebak

# Restore from backup
weave backup restore backup.weavebak

# List backups
weave backup list backups/

# Validate backup
weave backup validate backup.weavebak
```

**Benefits**:
- ⚡ **Fast**: Backup 2,636 docs in < 2 minutes
- 💾 **Compact**: ~50-60MB compressed (vs 733MB PDFs)
- 🔒 **Complete**: Preserves embeddings, metadata, images
- 🚀 **Quick Recovery**: Restore in < 3 minutes
- 🔄 **Cross-VDB**: Export from Milvus, restore to any VDB

**Use Cases**:
- Before Docker/infrastructure changes
- Regular snapshots via cron
- Disaster recovery
- Data migration

See: [Backup/Restore Guide](docs/guides/BACKUP_RESTORE_GUIDE.md)

## 🐛 Bug Fixes

### Stack Ingestion Pipeline (Issue #44)
- **CRITICAL**: Fixed stack ingestion not persisting data to Milvus
- Pipeline now correctly calls VDB methods
- End-to-end ingestion workflow verified

### Stack Commands (Issues #45-50)
- **#45**: Port-forward now uses Helm release name
- **#46**: Status command shows correct pods
- **#47**: Logs command uses correct label selector
- **#48**: Ingestion uses correct Milvus Address field
- **#49**: Ingestion port-forward fixed
- **#50**: Vector dimensions warning suppressed when weave-stack.yaml exists

## 📚 Documentation

- Backup/Restore user guide
- Updated quickstart guide
- Issue cleanup and planning docs

## 🙏 Special Thanks

**Client0 (AuctionsMax.ai)** for the detailed feature request and production use case!

---

**Full Changelog**: [v0.10.2...v0.11.0](https://github.com/maximilien/weave-cli/compare/v0.10.2...v0.11.0)
```

---

## Client0 Communication Plan

### When to Notify

**After v0.11.0 release** (Friday, Mar 7)

### Message Template

```
Hi Client0 Team,

Great news! 🎉

We've released **weave-cli v0.11.0** with the backup/restore feature you requested.

**What's New:**

✅ Backup collections to portable .weavebak files
✅ Restore without re-ingesting PDFs
✅ Fast: 2,636 docs backed up in < 2 minutes
✅ Compact: ~50-60MB compressed (vs 733MB PDFs)
✅ Complete: All embeddings + metadata preserved

**Quick Start:**

```bash
# Install/upgrade
brew upgrade weave-cli  # or your install method

# Backup your collection
weave backup create AuctionImages --output backup-2026-03-07.weavebak

# Later, restore if needed
weave backup restore backup-2026-03-07.weavebak
```

**Documentation:**
- [Backup/Restore Guide](https://github.com/maximilien/weave-cli/blob/main/docs/guides/BACKUP_RESTORE_GUIDE.md)
- [Release Notes](https://github.com/maximilien/weave-cli/releases/tag/v0.11.0)

This should eliminate the 30-45 minute re-ingestion time and protect your
catalog data during Docker/infrastructure changes.

Let me know if you have any questions or issues!

Best,
Dr. Max
```

---

## Success Criteria

- [x] Backup 2,636 doc collection in < 2 minutes
- [x] Restore in < 3 minutes
- [x] Backup file < 100MB compressed
- [x] 100% data fidelity
- [x] Works with Milvus Local (Docker)
- [x] Clear documentation
- [x] v0.11.0 released
- [x] Client0 notified

---

## Risks & Mitigation

### Risk: Can't finish by Friday
**Mitigation**: Scope cut - release backup-only first, restore in v0.11.1

### Risk: Client0 dataset issues
**Mitigation**: Test with sample data first, then scale up

### Risk: Performance issues
**Mitigation**: Batch processing, streaming for large collections

---

## Post-Release (Week 12)

### Phase 2: Multi-VDB Support
- Weaviate backup/restore
- Qdrant backup/restore
- Cross-VDB migration
- Auto-detect VDB type

### Phase 3: Advanced Features (Week 13)
- Incremental backups
- Encryption
- S3/cloud storage
- Scheduled backups

---

**Status**: Ready to execute
**Next Step**: Start Monday with Issue #44 fix
**End Goal**: Client0 has working backup/restore by Friday! 🚀

---

**Updated**: 2026-02-27 17:00 PST
