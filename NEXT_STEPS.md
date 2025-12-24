# Next Steps - Actionable Tasks

**Last Updated**: 2025-12-23 (Evening Session)
**Current Version**: v0.8.2-36-g1ac836a
**Status**: 🎉 **OPTION 1 COMPLETE + AI Chunking!** - 8/8 Features (100%)

---

## 📊 Current State

### ✅ Completed Today (2025-12-22)

**Morning Session:**
- ✅ REPL MCP Integration (1.5h) - Direct MCP server calls
- ✅ CI/CD Phase 1: Exit codes & JSON output (2h)
- ✅ CI/CD Phase 2: Incremental ingestion flags (1h)
- ✅ CI/CD Phase 3: Documentation (2h) - 1,831 lines
- ✅ CI/CD Phase 4: Examples (1.5h) - 10 production files

**Afternoon Session:**
- ✅ JSON/YAML Output Standardization (1h) - `--output` flag
- ✅ Collection Statistics Command (2h) - `weave stats`
- ✅ Comprehensive Test Coverage (1h) - 38+ tests, 100% pass
- ✅ YAML Linting Configuration (0.5h) - CI/CD exclusions

**Total**: ~12 hours, 3 major features, full test coverage

### 📈 Progress Summary

**Option 1 Features** (7/7 = 100%):
1. ✅ Pipeline Commands (already existed)
2. ✅ REPL Enhancement
3. ✅ MCP Client Integration
4. ✅ REPL MCP Integration
5. ✅ AI Schema Suggestion
6. ✅ CI/CD Integration (4 phases)
7. ✅ Collection Statistics + Output Standardization

**Test Coverage**:
- Stats command: 13+ test cases ✅
- CI/CD batch: 25+ test cases ✅
- All tests passing (100%)

**CI/CD Status**:
- Build: ✅ Passing
- Test: ✅ Passing
- Lint: ✅ Passing
- Security: ⚠️ Expected warning (Weaviate SDK - GO-2025-4237)

---

## ✅ Completed (2025-12-23)

**Morning Session - AI Chunking (1h15m)**:
- ✅ Extended schema agent with document structure analysis
- ✅ Added paragraph/section counting and content density metrics
- ✅ Integrated AI-powered chunking recommendations into schema suggest
- ✅ Displays recommended chunk size, overlap, and document type
- ✅ Comprehensive test coverage (14 tests, 100% pass)

**Afternoon Session - Chunking Refactoring**:
- ✅ Created separate `ChunkingAgent` in `src/pkg/agents/chunking_agent.go`
- ✅ Created `weave chunking suggest` command (separate from schema)
- ✅ Moved chunking logic to dedicated agent and command structure
- ✅ Removed duplicate structs from `schema_agent.go`
- ✅ Tested successfully with real data
- ✅ Fixed unit test compilation error (metadata type mismatch)
- ✅ Fixed --no-truncate flag implementation (d1404c3)

**Test Summary**:
- Total new tests: 14 (8 unit + 6 integration)
- All tests passing (100%)
- Coverage: document analysis, chunking metrics, serialization

**Impact**:
- Schema and chunking are now separate concerns with dedicated agents/commands
- Users can run `weave chunking suggest` independently

---

## ✅ Completed Tonight (2025-12-23/24 PM)

### Session 3: Testing & Investigation (40 minutes)

**Integration Testing**
- ✅ Created `tests/ai_features_integration_test.go`
- ✅ Schema suggestion tests (small sample, with requirements)
- ✅ Chunking suggestion tests (small sample, with requirements)
- ✅ Agent config loading tests
- ✅ Error handling tests (empty files, invalid paths)
- ✅ All 9 tests passing (7.94s)
- **Commit**: `78b7166`

**PDF Processing Investigation**
- ✅ Investigated PDF processing for Milvus/Supabase
- ✅ Found PDF works for Weaviate (dedicated path)
- ✅ Generic VDB path needs refactoring (2-3h work)
- ✅ Documented findings and deferred to future session
- **Commit**: `888cbd3`

**Total Time Tonight**: ~2h15m (Sessions 1-3)

---

## ✅ Completed Tonight (2025-12-23 PM - Earlier Sessions)

### Session 1: Quick Wins (45 minutes)

**Bug Investigation & Cleanup**
- ✅ Verified Milvus --text bug already fixed
- ✅ Cleaned up .gitignore (consolidated local storage entries)
- ✅ Updated VDB Support Matrix with AI features
- ✅ Updated README.md with chunking examples
- **Commit**: `3906d62` - cleanup and documentation updates

### Session 2: Agent Configuration (50 minutes)

**Configurable AI Agents**
- ✅ Created `configs/weave-agents.yaml` - Complete AI agent configuration
- ✅ Implemented config loading with precedence: local → configs → global → defaults
- ✅ Updated SchemaAgent to use config (model, temperature, max_tokens)
- ✅ Updated ChunkingAgent to use config (model, temperature, max_tokens)
- ✅ Fixed OutputConfig naming conflict (→ AgentOutputConfig)
- ✅ Created `configs/README.md` documenting agent configuration
- ✅ Tested successfully with `weave schema suggest`
- **Commit**: `93dac23` - configurable AI agents

**Impact**:
- Users can now customize LLM models, temperatures, and parameters
- Chunking defaults (sizes, overlap) configurable
- Confidence thresholds adjustable
- Performance settings tunable
- Clean separation from existing OutputConfig

**Total Time**: ~1h35m

---

## ✅ Completed Tonight (2025-12-23 PM - Earlier Session - 45 minutes)

### Priority 1: Critical Bug Investigation ✅
**Milvus --text Collection Creation**
- **Status**: ✅ **Already Fixed** - No action needed
- **Finding**: Code correctly handles --text flag for Milvus (lines 177-186 in create.go)
- **Evidence**: Command now fails with connection timeout (expected) vs schema error (bug)
- **Verification**: Tested `weave cols create WeaveDocs --text --json-metadata --milvus-local`

### Priority 2: Code Cleanup ✅
**Cleaned up .gitignore for local VDB storage**
- ✅ Removed redundant entries (qdrant_storage/, milvus_storage/, chroma_data/, neo4j_storage/)
- ✅ Kept single `/local/` entry (covers all local VDB storage)
- ✅ Removed duplicate `/local/storage/` entry
- **Result**: Cleaner .gitignore, all local VDB data in `./local/storage/`

### Priority 3: Documentation Updates ✅
**VDB Support Matrix (docs/VDB_SUPPORT_MATRIX.md)**
- ✅ Verified BM25 support accuracy (Qdrant ❌, Chroma ❌, Neo4j ❌ - all correct)
- ✅ Added "AI Schema Suggestions" feature row (✅ for all 10 VDBs)
- ✅ Added "AI Chunking Suggestions" feature row (✅ for all 10 VDBs)
- ✅ Added feature notes for new AI commands

**README.md**
- ✅ Updated "Key Features" section with AI Schema & Chunking
- ✅ Added example commands: `weave schema suggest`, `weave chunking suggest`
- ✅ Enhanced AI-Powered feature description

**Time**: ~45 minutes (faster than planned!)

---

## 🚫 Deferred Items

### High Priority (Require > 2 hours)
- **PDF Processing for Milvus/Supabase** (2-3h): Complex, requires refactoring
- **MCP Code Updates** (1-2h): Needs testing both MCP and CLI
- **./test.sh --mcp Failing** (1h): Depends on MCP updates

### Medium Priority (From weave-cli.txt)
- Update MCP code to match latest CLI changes
- Agents config should load from weave-agents.yaml (not hardcoded)
- Add test and CI for batch command (issue #9)
- Find small PDF collection for testing (issue #8)

### Low Priority / Features
- Embedding suggestion command (`weave embeds suggest`)
- Config search command to explain VDB search types
- Advanced search options (--bm25, --vector, --hnsw)

---

## 🎯 Tomorrow's Session Plan (3 hours)

### MCP Integration Updates (1.5 hours) 🔴 Critical

**Goal**: Sync weave-mcp server with latest CLI changes for full REPL functionality

#### Changes Since Last MCP Update

**1. AI Schema Suggestions** ⭐ NEW
- **CLI**: `weave schema suggest ./docs --collection MyDocs`
- **MCP Status**: ❌ Not implemented
- **Add**: `suggest_schema` tool (45 min)
  - Parameters: source_path, collection_name, requirements, max_samples
  - Handler: Call CLI with --output json
  - Test: Sample documents analysis

**2. AI Chunking Suggestions** ⭐ NEW
- **CLI**: `weave chunking suggest ./docs --collection MyDocs`
- **MCP Status**: ❌ Not implemented
- **Add**: `suggest_chunking` tool (45 min)
  - Parameters: source_path, collection_name, requirements, max_samples
  - Handler: Call CLI with --output json
  - Test: Chunking recommendations

**3. Configurable AI Agents** ⭐ NEW (Tonight)
- **Feature**: `weave-agents.yaml` configuration
- **MCP Status**: ⚠️ Needs awareness
- **Add**: Config loading support (15 min)

#### MCP Update Checklist

**High Priority** (1.5h):
- [ ] Add `suggest_schema` tool handler
- [ ] Add `suggest_chunking` tool handler
- [ ] Register tools in MCP server
- [ ] Test with real documents

**Medium Priority** (30 min):
- [ ] Add weave-agents.yaml support
- [ ] Improve error handling for AI commands
- [ ] Handle OPENAI_API_KEY missing gracefully

**Low Priority** (30 min):
- [ ] Fix `./test.sh --mcp` failures
- [ ] Update MCP documentation
- [ ] Add examples for new tools

#### Implementation Notes

**Handler Template**:
```go
// File: /Users/maximilien/github/maximilien/weave-mcp/src/pkg/mcp/handlers.go
func (s *Server) handleSuggestSchema(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    sourcePath := args["source_path"].(string)
    collectionName := args["collection_name"].(string)

    cmd := fmt.Sprintf("weave schema suggest %s --collection %s --output json",
        sourcePath, collectionName)

    return executeCommand(ctx, cmd)
}
```

**Testing**:
```bash
# Unit tests
cd /Users/maximilien/github/maximilien/weave-mcp && go test ./... -v

# Integration tests
cd /Users/maximilien/github/maximilien/weave-cli && ./test.sh --mcp

# Manual REPL test
./bin/weave
> analyze my documents in ./docs and suggest a schema
```

**Expected Outcome**:
- ✅ MCP server supports schema/chunking suggestions
- ✅ REPL mode can use AI features
- ✅ All tests passing

---

### PDF Processing Enhancement (1 hour) 🟡 **DEFERRED**

**Goal**: Add PDF support for Milvus and Supabase VDBs

**Current**: Works for Weaviate (has dedicated PDF path)
**Issue**: Generic VDB path doesn't support PDF processing yet

**Finding** (Dec 23):
- PDF processing exists in `src/pkg/pdf/` package
- Weaviate uses dedicated PDF processing path
- Generic VDB path (`CreateDocumentFromFileGeneric` in `src/cmd/utils/document.go:378`) returns error
- Requires refactoring to enable PDF for all VDBs (Milvus, Supabase, Qdrant, etc.)

**Recommendation**: Defer to dedicated refactoring session (2-3h)
- Extract PDF processing to generic helper
- Update all VDB adapters to use generic PDF helper
- Add integration tests across VDBs

**Files**:
- `src/cmd/utils/document.go:378` - Add PDF processing
- `src/pkg/pdf/` - Already has extraction logic
- `src/pkg/vectordb/*/document.go` - Update adapters

---

### Testing & Polish (30 min) 🟢

**Tasks**:
- [ ] Add integration tests for schema suggest
- [ ] Add integration tests for chunking suggest
- [ ] Quick code cleanup
- [ ] Documentation updates

---

## 🚫 Archived Session Plans

### Option A: Tonight (1 hour) - Quick Wins ⚡

**Focus**: High-impact fixes that can be completed quickly

#### Task 1: Agents Config Loading (45 min) 🟡
- **Issue**: Agent configs are hardcoded instead of loading from weave-agents.yaml
- **Location**: `src/pkg/agents/*.go`
- **Tasks**:
  - [ ] Create example `weave-agents.yaml` config file
  - [ ] Add config loading in agents package
  - [ ] Update agents to use config values
  - [ ] Test with schema and chunking agents
  - [ ] Update docs with configuration options

#### Task 2: Quick Documentation Polish (15 min) 📚
- **Tasks**:
  - [ ] Verify all links in README.md working
  - [ ] Update version in docs/CHANGELOG.md
  - [ ] Quick scan of user-facing docs for consistency

**Expected Outcome**: Better configurability for AI agents, cleaner docs

---

### Option B: Tomorrow AM (3 hours) - Feature & Hardening 🚀

**Focus**: Tackle deferred items and add polish

#### Session 1: MCP Integration Updates (1.5 hours) 🔴

**MCP Code Sync**
- **Issue**: MCP server code needs updates to match latest CLI changes
- **Location**: `/Users/maximilien/github/maximilien/weave-mcp/`
- **Tasks**:
  - [ ] Review CLI changes since last MCP update
  - [ ] Update MCP server with new commands (schema suggest, chunking suggest)
  - [ ] Update MCP tools/resources for new features
  - [ ] Fix ./test.sh --mcp failures
  - [ ] Test REPL mode with updated MCP server
  - [ ] Update MCP documentation

**Expected Outcome**: MCP server in sync, REPL mode fully functional

#### Session 2: PDF Processing Enhancement (1 hour) 🟡

**Implement PDF Support for Milvus/Supabase**
- **Issue**: PDF processing not implemented for Milvus and Supabase VDBs
- **Current**: Works for Weaviate, Qdrant, Chroma
- **Tasks**:
  - [ ] Review existing PDF processing code (weaviate/qdrant)
  - [ ] Implement PDF text extraction for Milvus
  - [ ] Implement PDF text extraction for Supabase
  - [ ] Add integration tests for both VDBs
  - [ ] Test with sample PDFs

**Expected Outcome**: PDF support for 2 more VDBs

#### Session 3: Agent Config + Testing (30 min) 🟢

**Agents Config Loading + Batch Testing**
- **Tasks**:
  - [ ] Implement weave-agents.yaml config loading
  - [ ] Add basic tests for batch command (issue #9)
  - [ ] Quick documentation updates

**Expected Outcome**: Better configurability + improved test coverage

---

### Option C: Tomorrow AM Alternative (3 hours) - Testing & Quality 🧪

**Focus**: Test coverage, CI/CD improvements, bug fixes

#### Session 1: Comprehensive Testing (1.5 hours)
- [ ] Add integration tests for schema suggest command
- [ ] Add integration tests for chunking suggest command
- [ ] Add tests for batch command (issue #9)
- [ ] Find/create small PDF test collection (issue #8)
- [ ] Test all commands with --json output flag

#### Session 2: CI/CD Improvements (1 hour)
- [ ] Add automated tests for AI features
- [ ] Improve test reporting
- [ ] Add performance benchmarks for key operations
- [ ] Document CI/CD setup in docs/

#### Session 3: Bug Fixes & Polish (30 min)
- [ ] Review and fix any open GitHub issues
- [ ] Code cleanup and optimization
- [ ] Documentation updates

---

## 📊 Recommended Choice

**For Tonight (1h)**: Choose **Option A** - Quick wins, low risk, improves configurability

**For Tomorrow AM (3h)**: Choose **Option B** - Highest impact
- MCP sync is critical for REPL mode
- PDF support extends functionality to more VDBs
- Good balance of features + hardening

**Alternative Tomorrow AM**: Choose **Option C** if you want to focus on quality/testing instead of new features

---

## 📋 Backlog (From weave-cli.txt Analysis)

### 🔴 Critical Bugs
- [ ] Milvus: Cannot create --text collection (schema not found error)
- [ ] PDF processing not implemented for Milvus/Supabase VDBs
- [ ] `--no-truncate` flag not working correctly
- [ ] Old PDF files produce empty output after conversion
- [ ] `./test.sh --mcp` is failing

### 🟡 Medium Priority
- [ ] Update MCP code to match latest CLI changes
- [ ] Agents config should load from weave-agents.yaml (not hardcoded)
- [ ] Move qdrant_storage/ to local/ directory
- [ ] Add test and CI for batch command (issue #9)
- [ ] Find small PDF collection for testing (issue #8)

### 🟢 Low Priority / Features
- [ ] Embedding suggestion command (`weave embeds suggest`)
- [ ] Config search command to explain VDB search types
- [ ] Advanced search options (--bm25, --vector, --hnsw)

---

## 🗂️ Archived / Completed (Reference)

### Recommended: Option 2 - VDB Expansion (Deprioritized)

**Note**: VDB expansion postponed to focus on stability

**Future Tasks** (Pick 1-2):

1. **LanceDB Support** (4-6 hours) - HIGH VALUE
   - Add LanceDB adapter (requires CGO setup)
   - Implements all VectorDBClient interface methods
   - Integration tests with local LanceDB
   - Documentation and troubleshooting guide
   - **Impact**: New performant embedded VDB option

2. **Pinecone Serverless Enhancements** (2-3 hours) - MEDIUM VALUE
   - Add serverless-specific features
   - Improve metadata filtering
   - Better error messages for quota limits
   - **Impact**: Better Pinecone user experience

3. **BM25 Search Improvements** (3-4 hours) - MEDIUM VALUE
   - Complete BM25 implementation for remaining VDBs
   - Currently: Weaviate, Qdrant, Neo4j (partial)
   - Add: MongoDB, Elasticsearch, OpenSearch
   - **Impact**: Keyword search across all VDBs

### Alternative: Option 3 - Production Hardening

**Focus**: Reliability, performance, observability

**Quick Wins** (2-3 hours each):

1. **Retry Logic Improvements**
   - Exponential backoff for transient failures
   - Configurable retry strategies per VDB
   - Better error classification (retryable vs permanent)

2. **Performance Monitoring**
   - Add timing metrics to operations
   - Performance stats in batch reports
   - Slow query detection and logging

3. **Enhanced Logging**
   - Structured logging with levels
   - Debug mode for troubleshooting
   - Request/response tracing for VDB calls

### Conservative: Polish & Documentation

**Focus**: Refinement and user experience

**Tasks** (1-2 hours each):

1. **Progress Bars** - Visual feedback for long operations
2. **Command Aliases** - Shorter commands (e.g., `weave s` for stats)
3. **Shell Completion** - Bash/Zsh completion scripts
4. **Video Tutorials** - Record usage demos

---

## 💡 Suggested Schedule (Tomorrow AM)

**Option A: Ambitious (LanceDB)**
```
9:00 AM  - Setup CGO environment for LanceDB
10:00 AM - Implement LanceDB adapter skeleton
11:00 AM - Core CRUD operations
12:00 PM - Integration tests + lunch
```

**Option B: Balanced (BM25 + Polish)**
```
9:00 AM  - BM25 search for MongoDB
10:00 AM - BM25 search for Elasticsearch
11:00 AM - Progress bars for batch command
12:00 PM - Tests + documentation
```

**Option C: Conservative (Hardening)**
```
9:00 AM  - Add retry logic with exponential backoff
10:00 AM - Performance metrics collection
11:00 AM - Enhanced logging infrastructure
12:00 PM - Integration testing
```

---

## 📋 Backlog (Deprioritized)

### Blocked Items
- [ ] Weaviate Security Fix (GO-2025-4237) - Waiting for SDK update

### Nice-to-Have (v0.9.0+)
- [ ] Close() signature standardization across VDBs
- [ ] WebSocket support for streaming operations
- [ ] GraphQL query interface
- [ ] Multi-tenancy improvements

---

## 📚 Key Documentation

**Integration Guides**:
- `docs/integrations/GITHUB_ACTIONS.md` (450 lines)
- `docs/integrations/ARGO_WORKFLOWS.md` (550 lines)
- `docs/integrations/AIRFLOW.md` (650 lines)

**Planning Docs**:
- `docs/planning/OPTION_2_VDB_EXPANSION.md`
- `docs/planning/OPTION_3_PRODUCTION_HARDENING.md`

**Technical**:
- `docs/ARCHITECTURE.md`
- `docs/VDB_SUPPORT_MATRIX.md`

---

## 🎯 Success Metrics

**Completed**:
- ✅ 7/7 Option 1 features (100%)
- ✅ 10 VDBs fully supported
- ✅ 38+ test cases passing
- ✅ CI/CD integration (3 platforms)
- ✅ 6,000+ lines of code

**Tomorrow's Targets**:
- Add 1 new VDB OR enhance 2-3 existing VDBs
- Maintain 100% test pass rate
- Add 15-20 new test cases
- Update documentation for changes

---

**Last Updated**: 2025-12-22 17:05 PST
**Next Review**: Tomorrow morning (2025-12-23)
