# Demo Update Plan

## Overview

Significant features have been added since the last demo recordings (Oct 15, 2025). This document tracks what needs to be updated and recorded.

## Demo Status

### Existing Demos

| Demo | Last Recorded | Current URL | Status |
|------|--------------|-------------|--------|
| Full Demo | 2025-10-15 | https://asciinema.org/a/LrKzmThBfDbTPISZzr8biP4dt | ⚠️ **Needs Update** |
| Quick Demo | 2025-10-15 | https://asciinema.org/a/HiAU7h1iJvZ2QdJe70ae3Cc0b | ⚠️ **Needs Update** |
| REPL Demo | 2025-11-05 | https://asciinema.org/a/U504HN4FSeMsOA0qS0os0NWUE | ✅ **Current** |

### New Demos

| Demo | Script Ready | Recording Status |
|------|--------------|------------------|
| Config Demo | ✅ `demos/config-demo.sh` | 📹 **Ready to Record** |
| Supabase Demo | ✅ `demos/supabase-demo.sh` | 📹 **Ready to Record** |

## Features Added Since Oct 15

### Major Features (Should be in demos)

1. **REPL Progress Improvements** ✅ (Nov 14)
   - Real-time output streaming
   - Progress spinner with elapsed time
   - Time estimation
   - Item counts
   - Status: Already in REPL demo (Nov 5)

2. **Configuration Management** (Oct-Nov)
   - Interactive `weave config create --env`
   - Global vs local configuration
   - Configuration precedence
   - Status: **New Config Demo needed**

3. **Supabase Integration** (Oct-Nov)
   - PostgreSQL + pgvector backend
   - BM25 full-text search
   - Hybrid search
   - Multi-database operations
   - Status: **New Supabase Demo needed**

4. **Embedding Support Enhancements** (Oct)
   - Multi-provider support (OpenAI, Cohere, Hugging Face)
   - `weave embeddings list` with VDB compatibility
   - JSON output support
   - Status: Should be in updated full demo

5. **Vector DB Abstraction** (Oct)
   - Pluggable database architecture
   - `--weaviate`, `--supabase`, `--all` flags
   - Unified interface across databases
   - Status: Should be in updated full demo

6. **Global Configuration Directory** (Oct)
   - `~/.weave-cli/` support
   - `weave config sync`
   - Config loaded from any directory
   - Status: Should be in config demo

7. **JSON Output Support** (Oct)
   - `--json` flag for health, docs, collections
   - Programmatic integration
   - Status: Should be in updated quick demo

## Recording Priority

### Priority 1: New Demos (Missing from current lineup)

1. **Config Demo** - `demos/config-demo.sh`
   - Duration: ~5 minutes
   - Prerequisites: None (works with any setup)
   - Command:
     ```bash
     asciinema rec videos/weave-cli-config-demo.cast
     ./demos/config-demo.sh
     # Upload and update docs
     ```

2. **Supabase Demo** - `demos/supabase-demo.sh`
   - Duration: ~8 minutes
   - Prerequisites: Supabase configured
   - Command:
     ```bash
     asciinema rec videos/weave-cli-supabase-demo.cast
     ./demos/supabase-demo.sh
     # Upload and update docs
     ```

### Priority 2: Update Existing Demos

3. **Full Demo** - Update script in `docs/guides/DEMO.md`
   - Add: Embedding model selection
   - Add: JSON output examples
   - Add: Multi-database operations
   - Add: Reference to new config/Supabase demos
   - Duration: Keep ~5 minutes
   - Command:
     ```bash
     # After updating script
     asciinema rec videos/weave-cli-full-demo.cast
     # Follow updated DEMO.md script
     # Upload and update docs
     ```

4. **Quick Demo** - Update for JSON output
   - Add: `--json` flag examples
   - Add: Quick config show
   - Duration: Keep ~2 minutes
   - Command:
     ```bash
     # After updating script
     asciinema rec videos/weave-cli-quick-demo.cast
     # Follow updated script
     # Upload and update docs
     ```

## Demo Script Updates Needed

### Full Demo (`docs/guides/DEMO.md`)

**Additions needed:**
- Page 1: Show `weave embeddings list` with compatibility
- Page 2: Show embedding model selection with `--embedding`
- Page 7: Add JSON output example: `weave cols q DemoCollection "query" --json`
- Page 10: Add links to new Config and Supabase demos

### Quick Demo (Create separate script)

**Create**: `docs/guides/QUICK_DEMO.md` with:
- Page 1: Quick config check
- Page 2: Create collection with embedding
- Page 3: Add document
- Page 4: Query with JSON output
- Page 5: Links to full demo and specialized demos

## Recording Checklist

For each demo recording:

- [ ] Review and test script
- [ ] Clean environment (remove test collections)
- [ ] Start recording: `asciinema rec videos/weave-cli-{demo-name}.cast`
- [ ] Run demo script
- [ ] Stop recording (Ctrl+D)
- [ ] Review recording: `asciinema play videos/weave-cli-{demo-name}.cast`
- [ ] Upload: `asciinema upload videos/weave-cli-{demo-name}.cast`
- [ ] Save URL to `videos/latest-demo-uploads.txt`
- [ ] Update `docs/guides/DEMO.md` with new URL
- [ ] Update `README.md` with new URL
- [ ] Commit changes

## Documentation Updates Required

After recording all demos:

1. **docs/guides/DEMO.md**
   - Update "Available Recordings" section
   - Remove "To Be Recorded" entries
   - Add config demo link
   - Add Supabase demo link

2. **README.md**
   - Update "Video Demos" section
   - Add config demo link
   - Add Supabase demo link

3. **videos/latest-demo-uploads.txt**
   - Add all new URLs
   - Update quick/full URLs

## Estimated Time

- Config demo recording: 10 minutes
- Supabase demo recording: 15 minutes (with setup verification)
- Full demo script update: 20 minutes
- Quick demo script creation: 15 minutes
- Full demo re-recording: 10 minutes
- Quick demo re-recording: 5 minutes
- Documentation updates: 15 minutes

**Total: ~90 minutes**

## Next Steps

1. ✅ Create demo scripts (config, Supabase) - DONE
2. ⏳ Record config demo
3. ⏳ Record Supabase demo
4. ⏳ Update full demo script
5. ⏳ Create quick demo script
6. ⏳ Re-record full demo
7. ⏳ Re-record quick demo
8. ⏳ Update all documentation links
9. ⏳ Verify all links work
10. ⏳ Commit and push changes

## Notes

- REPL demo (Nov 5) is current - already includes progress improvements
- Config and Supabase demos are new features not covered by existing demos
- Full/Quick demos need updates for new flags and features added in Oct-Nov
- Consider creating a "Multi-Database" demo in the future showing Weaviate + Supabase together
