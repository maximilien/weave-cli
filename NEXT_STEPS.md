# Next Steps - Actionable Tasks

**Last Updated**: 2025-12-22 (Post REPL MCP Integration)
**Current Version**: v0.8.2-20-g14ee9b4-dirty
**Status**: 🚀 **Building New Features** - 5/7 Complete (71%)!

---

## 🎉 Latest Accomplishments (2025-12-22)

### Session: REPL Enhancement, AI Schema Suggestion, MCP Client Integration + REPL MCP

**Completed Features from Option 1:**

1. ✅ **Feature 1.2: Interactive REPL Enhancement** (3-4 hours)
   - Enhanced existing REPL with structured commands
   - Added tab completion for commands and subcommands
   - Implemented session state management (currentVDB, currentCollection)
   - Commands: /mcp, /collection, /search, /stats, /use, /status
   - Dynamic prompt updates based on active collection
   - Files: `src/pkg/repl/commands.go` (237 lines), `src/pkg/repl/completer.go` (137 lines)
   - Commit: 7024702

2. ✅ **Feature 1.7: AI Schema Suggestion** (5-6 hours)
   - AI-powered document analysis using GPT-4o
   - Document sampling for PDF, JSON, TXT, MD files
   - Schema generation with field type inference
   - YAML output with metadata headers
   - CLI: `weave schema suggest SOURCE --collection NAME --output FILE`
   - Files: `src/pkg/agents/schema_agent.go` (430 lines), `src/cmd/schema/` (2 files, 280 lines)
   - Commits: cc07c41, 6d8c23a (test fix)

3. ✅ **Feature 1.2: MCP Client Integration** (5-6 hours) 🆕
   - MCP client for calling external MCP servers
   - HTTP and stdio transport support
   - Bearer/Basic/API key authentication
   - Factory pattern with comprehensive validation
   - CLI: `weave mcp list/call/test --server URL`
   - Files: `src/pkg/mcp/` (3 files, 519 lines), `src/cmd/mcp/` (4 files, 437 lines)
   - Integration tests: 4 suites, 14 test cases, all passing
   - Commit: b89ecce

4. ✅ **Feature 1.2b: REPL MCP Integration** (1-2 hours) 🆕🆕
   - Connected MCP client to REPL for direct server calls
   - Added `/mcp list`, `/mcp call`, `/mcp status` direct implementations
   - CLI flags: `--mcp-server`, `--mcp-transport`, `--mcp-timeout`
   - Graceful fallback to LLM when MCP server not configured
   - Files: `src/pkg/repl/repl.go`, `src/pkg/repl/commands.go`, `src/cmd/root.go`
   - All integration tests passing

5. ✅ **Feature 1.1: Pipeline Commands** - ALREADY IMPLEMENTED!
   - Discovered `weave docs batch` and `weave pipeline ingest` already exist
   - Full batch ingestion with parallel processing, retry logic, progress tracking
   - 630+ lines in `src/cmd/document/batch.go` + pipeline package

**Total Implementation**: ~17.5 hours invested
**Lines Added**: ~2,700+ lines (implementation + tests)
**Progress**: 5/7 features complete (1 already existed)
**Commits**: 7+ commits (7024702, cc07c41, 887f22a, 183ceef, 6d8c23a, 928db0e, b89ecce + REPL MCP)

**Impact:**
- REPL has structured command routing alongside natural language
- REPL can directly call external MCP servers without LLM routing
- AI can suggest optimal vector DB schemas from sample documents
- weave-cli can call external MCP servers (HTTP/stdio) via CLI or REPL
- Comprehensive test coverage for all features

---

## 🎯 Session Goal: Completion & Polish

**Mission**: Finish all bugs, technical debt, and polish items. Get to a "done" state before adding new features.

**Latest Release**: v0.8.2 - UX & Test Quality Improvements
- ✅ 100% Troubleshooting Coverage (10/10 VDBs)
- ✅ 100% Batch Create Verification (10/10 VDBs)
- ✅ VDB Status Documentation Accuracy

**CI Status**: 3/4 passing (Build ✅, Test ✅, Lint ✅, Security ❌ expected - Weaviate GO-2025-4237)

---

## 📊 Quick Status: What's Left?

### ✅ ALL PHASES COMPLETE!

**Remaining Items (Optional/Blocked):**
- [ ] ❌ **Weaviate Security** (GO-2025-4237) - BLOCKED waiting for SDK compatibility
- [ ] 🔧 **Close() Signature Standardization** - Optional consistency improvement (v0.9.0)

**Completed This Session (2025-12-19):**
- [x] 📝 **Phase 2: Documentation** - MongoDB (exists), Neo4j (v0.8.1), OpenSearch AWS (new)
- [x] ✨ **Phase 3: Polish** - Test coverage, TODO audit, ARCHITECTURE.md
- [x] 🔧 **Error Message Consistency** - All 10 VDBs now have consistent error naming

**Estimated Total**: ✅ **DONE** - Only blocked Weaviate issue remains

---

## 🚀 Next Steps

**Progress**: 5/7 Option 1 features complete (1 already existed) - **71% DONE!**
**Remaining**: ~2-7 hours for additional features

### ✅ JUST COMPLETED: REPL MCP Integration (1.5 hours) 🎉

**Implementation:**
1. ✅ Added `mcpClient` field and `mcpEnabled` bool to REPL struct
2. ✅ Added MCP flags to root command (`--mcp-server`, `--mcp-transport`, `--mcp-timeout`)
3. ✅ Implemented direct MCP client calls in `/mcp list`, `/mcp call`, `/mcp status`
4. ✅ Added graceful fallback to LLM when MCP server not configured
5. ✅ Updated `/status` command to show MCP connection info
6. ✅ All tests passing (REPL integration + MCP client)

**Files Modified:**
- `src/pkg/repl/repl.go` - Added MCP client initialization and cleanup
- `src/pkg/repl/commands.go` - Direct MCP calls with fallback
- `src/cmd/root.go` - Added MCP flags (`--mcp-server`, `--mcp-transport`, `--mcp-timeout`)

**Usage:**
```bash
# Start REPL with MCP server
weave --mcp-server http://localhost:8030

# In REPL - direct MCP calls without LLM
weave> /mcp status
✓ MCP connection: active
   Server: http://localhost:8030
   Transport: http

weave> /mcp list
Available MCP Tools:
===================
1. extract_entities
   Extract named entities from text
   Parameters:
     • text (string) - Text to analyze

weave> /mcp call extract_entities text="Sample text"
→ Calling MCP tool: extract_entities
✓ Tool Result: {...}

# Without MCP server - falls back to LLM
weave> /mcp list
⚠️  No MCP server configured. Use --mcp-server flag when starting REPL
   Falling back to LLM-based MCP tools...
```

---

### Option 1: New Features (In Progress - Recommended)
**Focus**: Add value for users, expand capabilities
**Completed**:
- ✅ Pipeline Commands (already existed)
- ✅ REPL Enhancement (3-4h)
- ✅ MCP Client (5-6h)
- ✅ AI Schema Suggestion (5-6h)

**Remaining**: Progress Bars, JSON/YAML Output, Collection Statistics (3-8h total)

**Note on Pipeline Commands**:
Already implemented! Use:
- `weave docs batch --directory ./docs --collection NAME --parallel 3`
- `weave pipeline ingest ./docs --collection NAME --glob "**/*.pdf"`

Both have progress tracking, parallel workers, retry logic, and resume capability.

---

### Other Option 1 Features (Lower Priority)

#### Progress Bars & Visual Feedback (2-3 hours)
- Add progress bars to long-running operations
- Document ingestion, search, collection operations
- Use existing libraries (e.g., progressbar, spinner)

#### JSON/YAML Output Modes (1-2 hours)
- Add `--output json|yaml` flag to all commands
- Machine-readable formats for CI/CD integration
- Already partially implemented in some commands

#### Collection Statistics (2-3 hours)
- `weave stats --collection NAME` command
- Document count, average vector size, index stats
- Top keywords, metadata analysis

---

### Alternative Options (If Option 1 Complete)

#### CI/CD Integration Documentation (3-4 hours)
**Goal**: Document how to use weave-cli in CI/CD pipelines

**GitHub Actions Integration:**

- GitHub Actions workflow examples
- Argo Workflows integration
- Airflow DAG examples
- Documentation and guides

#### Option 2: VDB Expansion
- Add new vector databases (LanceDB requires CGO)
- Expand existing VDB features
- See `docs/planning/OPTION_2_VDB_EXPANSION.md`

#### Option 3: Production Hardening
- Error handling improvements
- Performance optimizations
- See `docs/planning/OPTION_3_PRODUCTION_HARDENING.md`

---

## 📚 Documentation

**Planning Docs**: `docs/planning/OPTION_*.md`
**VDB Support**: `docs/VDB_SUPPORT_MATRIX.md`
**Architecture**: `docs/ARCHITECTURE.md`

---

**Last Updated**: 2025-12-22
**Next Review**: After REPL MCP integration complete
