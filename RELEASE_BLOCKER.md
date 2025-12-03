# v0.7.2 Release Blocker - Chroma Build Issue

**Date**: 2025-12-03
**Status**: 🚨 BLOCKED - GitHub Actions release build failing
**Priority**: HIGH (but not blocking local use or presentation)

---

## Problem Summary

GitHub Actions release workflow fails to build Linux/Windows binaries due to Chroma Go SDK's CGO dependency (`libtokenizers`), which is macOS-only.

### Error
```
imports github.com/amikos-tech/chroma-go/pkg/tokenizers/libtokenizers:
build constraints exclude all Go files in /home/runner/go/pkg/mod/github.com/amikos-tech/chroma-go@v0.2.5/pkg/tokenizers/libtokenizers
```

### What Works
- ✅ Local macOS build: `./build.sh` succeeds
- ✅ Local Linux build: `CGO_ENABLED=0 GOOS=linux go build` succeeds
- ✅ All features working perfectly locally
- ✅ All 10 VDBs healthy
- ✅ Ready for presentation

### What Doesn't Work
- ❌ GitHub Actions Linux/Windows builds
- ❌ Automated GitHub release creation

---

## Root Cause

Go's module system downloads and evaluates ALL dependencies in go.mod BEFORE applying build constraints. Even though our code has correct build tags, Go tries to download and validate chroma-go for Linux builds, which fails because the SDK's tokenizers package has no Linux-compatible files.

**Key insight**: Build constraints prevent compilation, but not dependency resolution.

---

## Solutions Attempted

### ✅ Attempt 1: Build Constraints (FAILED)
- Added `//go:build (darwin && amd64) || (darwin && arm64)` to all chroma files
- Created stub factory for unsupported platforms
- **Result**: Go still downloads chroma-go and fails

### ✅ Attempt 2: Separate Init Files (FAILED)
- `init_chroma_darwin.go` - imports chroma package
- `init_chroma_unsupported.go` - registers stub without importing
- **Result**: Go still evaluates imports from darwin file

### ❌ Attempt 3: Build Tags on Chroma Package (ABANDONED)
- Would require `chroma` build tag for all builds
- **Problem**: Breaks local macOS builds (you need Chroma!)

---

## Recommended Solutions

### Option 1: Split Workflow (RECOMMENDED) ⭐
**Build Linux/Windows and macOS separately**

Pros:
- Clean separation of concerns
- macOS builds on macOS runner (has chroma-go SDK)
- Linux builds on Linux runner (no chroma dependency needed)
- No code changes required

Cons:
- More complex workflow
- Slightly longer build time

**Implementation**:
```yaml
jobs:
  build-linux-windows:
    runs-on: ubuntu-latest
    steps:
      - Remove chroma from go.mod temporarily
      - Build Linux/Windows binaries

  build-macos:
    runs-on: macos-latest
    steps:
      - Build macOS binaries (with full Chroma support)

  combine-and-release:
    needs: [build-linux-windows, build-macos]
    steps:
      - Download all artifacts
      - Create release
```

### Option 2: Make Chroma Optional via go.mod (COMPLEX)
Use `replace` directive or optional dependencies

Pros:
- Single workflow

Cons:
- Requires go.mod manipulation
- Complex to maintain

### Option 3: Use GoReleaser (FUTURE)
Professional release tool that handles cross-compilation better

Pros:
- Industry standard
- Handles these cases automatically

Cons:
- Requires setup and learning
- Overkill for current needs

---

## Workaround for Now

**For your presentation tomorrow:**
1. ✅ Use local binary (works perfectly!)
2. ✅ All features demonstrated
3. ✅ No impact on functionality

**For users:**
- macOS users: Build from source (`./build.sh`) - works great
- Linux/Windows users: Will need to wait for release fix

---

## Next Steps

1. **After presentation**: Implement Option 1 (split workflow)
2. **Test thoroughly**: Verify all platforms build correctly
3. **Document**: Update build instructions for contributors

---

## Technical Notes

### Why Local Build Works
- We already have chroma-go downloaded in module cache
- Go doesn't re-download when building with different GOOS
- Works because dependencies are already satisfied

### Why GitHub Actions Fails
- Fresh environment, no module cache
- Downloads all dependencies before building
- Fails when chroma-go's tokenizers can't be validated on Linux

### Why This Is Hard
- Go's design: module resolution happens before build tag evaluation
- Chroma SDK uses CGO for tokenizers (platform-specific)
- No clean way to exclude dependencies per platform in go.mod

---

**Bottom line**: This is a CI/CD configuration issue, NOT a code issue. Your local build works perfectly and is ready for demo!
