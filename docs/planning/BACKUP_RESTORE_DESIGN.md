# Backup/Restore Feature Design

**Issue**: #43
**Requested by**: Client0 (AuctionsMax.ai)
**Priority**: P0 - Blocking production deployment
**Status**: Design Phase
**Target**: Week 11-12 (Mar 3-14, 2026)

---

## Overview

Add `weave backup` commands to export/import vector database collections, preventing data loss and eliminating costly re-ingestion time (30-45 min for large datasets).

## Commands

```bash
# Create backup
weave backup create <collection> --output backup.weavebak [--vdb-type milvus-local]
weave backup create --all --output backups/ [--vdb-type milvus-local]

# Restore backup
weave backup restore backup.weavebak [--vdb-type milvus-local]
weave backup restore backups/ --collection <name> [--vdb-type milvus-local]

# List backups
weave backup list [directory]

# Validate backup
weave backup validate backup.weavebak
```

## File Format: `.weavebak`

### Structure (JSON + optional compression)

```json
{
  "version": "1.0.0",
  "metadata": {
    "collection": "AuctionImages",
    "vdb_type": "milvus-local",
    "embedding_model": "text-embedding-3-small",
    "vector_dimensions": 1536,
    "created_at": "2026-02-27T02:00:00Z",
    "weave_version": "0.10.3",
    "total_documents": 2636,
    "backup_size_bytes": 157286400
  },
  "schema": {
    "vector_dimensions": 1536,
    "similarity_metric": "cosine",
    "fields": [
      {"name": "content", "type": "text"},
      {"name": "embedding", "type": "vector"},
      {"name": "metadata", "type": "json"}
    ]
  },
  "documents": [
    {
      "id": "doc-123",
      "content": "auction item description",
      "embedding": [0.1, 0.2, ...],  // Full vector
      "metadata": {
        "source_file": "catalog_2025.pdf",
        "page_number": 5,
        "chunk_index": 2,
        "image_base64": "...",  // For image documents
        "image_url": "file://./storage/img123.png"
      }
    }
  ]
}
```

### Compression Options

- **None**: `.weavebak` (JSON)
- **gzip**: `.weavebak.gz` (default, ~60% reduction)
- **zstd**: `.weavebak.zst` (better compression, faster)

### File Size Estimation

Client0's dataset (2,636 docs, 1,548 images, 733MB PDFs):
- Uncompressed: ~150MB (vectors + metadata)
- gzip: ~60MB
- zstd: ~50MB

## Implementation Phases

### Phase 1: Core Backup/Restore (Week 11)

**Files to Create**:
- `src/cmd/backup/create.go` - Export collection
- `src/cmd/backup/restore.go` - Import collection
- `src/cmd/backup/list.go` - List backups
- `src/cmd/backup/validate.go` - Verify integrity
- `src/pkg/backup/format.go` - `.weavebak` serialization
- `src/pkg/backup/compressor.go` - Compression helpers

**Scope**:
- [x] Single collection export
- [x] Full restore with validation
- [x] gzip compression
- [x] Progress tracking
- [x] Works with Milvus

**Estimated Time**: 3-4 days

### Phase 2: Multi-VDB Support (Week 12)

**Scope**:
- [ ] Weaviate backup/restore
- [ ] Qdrant backup/restore
- [ ] Chroma backup/restore
- [ ] Auto-detect VDB type from backup file
- [ ] Cross-VDB migration (export from Milvus, restore to Weaviate)

**Estimated Time**: 2-3 days

### Phase 3: Advanced Features (Week 13)

**Scope**:
- [ ] Incremental backups (only changed docs)
- [ ] Backup all collections (`--all`)
- [ ] Encryption (`--encrypt`)
- [ ] Backup scheduling integration
- [ ] S3/cloud storage support

**Estimated Time**: 2-3 days

## Technical Design

### Backup Flow

```
1. Connect to VDB
2. Get collection schema
3. Query all documents (batch by 100)
4. For each document:
   - Extract: id, content, embedding, metadata
   - Serialize to JSON
5. Write to .weavebak file
6. Optional: Compress with gzip
7. Validate backup integrity
```

### Restore Flow

```
1. Read .weavebak file
2. Decompress if needed
3. Validate format version
4. Connect to VDB
5. Check if collection exists:
   - If exists: Prompt to overwrite
   - If not: Create collection with schema
6. Batch insert documents (100 per batch)
7. Verify restoration (count check)
```

### Error Handling

- **Partial backups**: Save progress, allow resume
- **Corruption**: Checksum validation
- **Version mismatch**: Warn but attempt restore
- **Disk space**: Check before backup
- **VDB errors**: Retry with exponential backoff

## API Design

### Backup Package

```go
package backup

type BackupFormat struct {
    Version  string
    Metadata BackupMetadata
    Schema   CollectionSchema
    Documents []Document
}

type BackupMetadata struct {
    Collection      string
    VDBType         string
    EmbeddingModel  string
    VectorDimensions int
    CreatedAt       time.Time
    TotalDocuments  int
    BackupSizeBytes int64
}

func Create(collection string, vdbClient vectordb.VectorDBClient, opts *CreateOptions) error
func Restore(backupFile string, vdbClient vectordb.VectorDBClient, opts *RestoreOptions) error
func Validate(backupFile string) (*ValidationResult, error)
func List(directory string) ([]BackupInfo, error)
```

### CLI Integration

```go
// src/cmd/backup/create.go
var CreateCmd = &cobra.Command{
    Use:   "create <collection> --output <file>",
    Short: "Backup a vector database collection",
    Flags: []string{
        "output",      // Output file path
        "compress",    // Compression type (gzip, zstd, none)
        "batch-size",  // Documents per batch
        "quiet",       // Suppress progress
    },
}
```

## Testing Strategy

### Unit Tests
- [ ] Format serialization/deserialization
- [ ] Compression/decompression
- [ ] Validation logic
- [ ] Error handling

### Integration Tests
- [ ] Full backup/restore cycle (Milvus)
- [ ] Large collection (1000+ docs)
- [ ] Corrupted backup handling
- [ ] Cross-VDB migration

### Manual Testing (Client0 Dataset)
- [ ] Backup 2,636 document collection
- [ ] Verify file size (~50-60MB)
- [ ] Delete collection
- [ ] Restore from backup
- [ ] Query restored data
- [ ] Compare checksums

## Documentation

### User Guide
- `docs/guides/BACKUP_RESTORE_GUIDE.md`
- Examples for common use cases
- Troubleshooting section

### CLI Help
- Clear examples in `--help` text
- Flag descriptions
- Error messages with fixes

## Success Criteria

- [x] Client0 can backup 2,636 doc collection in < 2 minutes
- [x] Restore completes in < 3 minutes
- [x] Backup file < 100MB (compressed)
- [x] 100% data fidelity (all embeddings + metadata preserved)
- [x] Works with Milvus Local (Docker)
- [x] Clear progress indication
- [x] Comprehensive error messages

## Timeline

| Week | Dates | Tasks |
|------|-------|-------|
| 11 | Mar 3-7 | Phase 1: Core backup/restore (Milvus) |
| 12 | Mar 10-14 | Phase 2: Multi-VDB support |
| 13 | Mar 17-21 | Phase 3: Advanced features |

**MVP Target**: End of Week 11 (Mar 7)
**Full Release**: End of Week 12 (Mar 14)

## Risk Mitigation

### Large Collections
**Risk**: Memory issues with huge collections
**Mitigation**: Stream processing, batch writes

### Concurrent Access
**Risk**: Collection modified during backup
**Mitigation**: Snapshot/versioning if VDB supports it

### Corrupted Backups
**Risk**: Partial writes, disk failures
**Mitigation**: Checksums, atomic writes, validation

### Version Compatibility
**Risk**: Old backups won't restore with new schema
**Mitigation**: Version field, migration logic

## Open Questions

1. **Binary format vs JSON?**
   - JSON: Human-readable, easier to debug
   - Binary: Smaller, faster
   - **Decision**: Start with JSON + compression, add binary later if needed

2. **Include source PDFs in backup?**
   - Pros: Complete self-contained backup
   - Cons: Massive file size (733MB + embeddings)
   - **Decision**: No, PDFs should be backed up separately

3. **Incremental backups priority?**
   - Useful for large collections
   - Adds complexity
   - **Decision**: Phase 3, not MVP

---

**Status**: Ready for implementation
**Next Step**: Begin Phase 1 implementation (Week 11)
