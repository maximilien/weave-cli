# Archived Documentation

This directory contains completed planning documents and status files that have been implemented and released.

## Archived Files

### Multi-Modal RAG Support (v0.9.4)

**Completed**: 2026-01-19
**Release**: v0.9.4

Files:
- `STATUS_TOP_K_IMAGES.md` - Feature status and implementation tracking
- `MULTIMODAL_RAG_SUPPORT.md` - Planning and implementation plan

**Summary**:
Implemented full multi-modal RAG support with `--top_k_images` flag, image collection creation with embeddings, and comprehensive integration tests across all VDBs.

**Key Achievements**:
- ✅ Fixed critical bug in image collection creation
- ✅ Implemented `--top_k_images` flag
- ✅ Added schema type detection
- ✅ Created 3 comprehensive integration test suites
- ✅ Multi-VDB support (Milvus, Weaviate, Chroma, Qdrant)

---

## Archive Guidelines

Documents should be moved here when:
1. Feature is fully implemented and released
2. Documentation is superseded by README.md or other main docs
3. Status tracking is complete (all tasks done)

Documents to keep in main docs/:
- Active planning documents
- Current roadmaps
- Ongoing status files
- Reference documentation

---

**Last Updated**: 2026-01-19
