# VDB Lifecycle Management - Implementation Plan

**Feature Request**: User wants CLI commands to manage local VDB instances
**Status**: Planning
**Priority**: High (user-requested)
**Effort**: 4-6 hours
**Target**: v0.9.13 or v0.10.0

---

## 🎯 User Requirements

```bash
weave vdb start --weaviate-local   # Start local Weaviate
weave vdb stop --milvus-local      # Stop local Milvus
weave vdb status --chroma-local    # Check Chroma status
weave vdb ls                       # List all configured local VDBs
weave config vdb                   # Customize VDB scripts/config (optional)
```

---

## 📊 Current State Analysis

### ✅ Infrastructure Already Exists

1. **VDB Management Scripts** (`tools/vdb/local/`):
   - `weaviate.sh` - Weaviate lifecycle management
   - `milvus.sh` - Milvus lifecycle management
   - `qdrant.sh` - Qdrant lifecycle management
   - `neo4j.sh` - Neo4j lifecycle management
   - `postgres.sh` - PostgreSQL/Supabase lifecycle management
   - `opensearch.sh` - OpenSearch lifecycle management
   - `elasticsearch.sh` - Elasticsearch lifecycle management
   - `manager.sh` - Multi-VDB orchestrator

2. **Container Abstraction** (`tools/vdb/container/`):
   - `detect.sh` - Auto-detects podman/docker (podman preferred)
   - `run.sh` - Unified container command interface
   - ✅ **No Docker/Podman complexity - already abstracted!**

3. **Existing Commands**:
   - Each script supports: `start`, `stop`, `status`, `logs`, `clean`
   - Already handles port conflicts, storage dirs, env vars
   - Already supports OPENAI_API_KEY from .env

### ⚠️ What's Missing

- CLI integration (no `weave vdb` command yet)
- Config file integration for custom ports/storage
- Cross-platform path handling in Go
- VDB discovery from config.yaml
- Unified status reporting
- Error handling and user feedback
- **Script distribution for production/installed binaries** ⚠️ Critical!

---

## 📦 Script Distribution Strategy

### ❓ Problem: Where Do Scripts Live in Production?

**Development Mode** (repo checkout):
```
weave-cli/
├── bin/weave
└── tools/vdb/local/*.sh  ✅ Scripts exist
```

**Production Mode** (installed binary):
```
/usr/local/bin/weave  or  ~/bin/weave
└── ??? Where are the scripts? ❌
```

**Users won't have `tools/vdb/local/` if they only download the binary!**

---

### ✅ Solution: Embed + Extract to ~/.weave-cli/

**Approach**: Use Go's `embed` package to bundle scripts in binary, extract on first use

**Script Location Hierarchy**:
1. **User-customized scripts**: `~/.weave-cli/vdb/` (highest priority)
2. **Embedded in binary**: Extracted to `~/.weave-cli/vdb/` on first use
3. **Repo scripts**: `tools/vdb/local/` (development only)

**Directory Structure**:
```
~/.weave-cli/
├── config.yaml          # Existing
├── .env                 # Existing
├── agents/              # Existing (custom agents)
│   ├── my-agent.yaml
│   └── custom-rag.yaml
└── vdb/                 # NEW! (VDB scripts)
    ├── weaviate.sh
    ├── milvus.sh
    ├── qdrant.sh
    ├── neo4j.sh
    ├── postgres.sh
    ├── opensearch.sh
    ├── elasticsearch.sh
    └── container/
        ├── detect.sh
        └── run.sh
```

---

### 🔧 Implementation Details

#### 1. Embed Scripts in Binary

```go
package vdb

import _ "embed"

// Embed all VDB scripts
//go:embed scripts/weaviate.sh
var weaviateScript string

//go:embed scripts/milvus.sh
var milvusScript string

//go:embed scripts/qdrant.sh
var qdrantScript string

//go:embed scripts/container/detect.sh
var containerDetectScript string

//go:embed scripts/container/run.sh
var containerRunScript string

// Map of all embedded scripts
var embeddedScripts = map[string]string{
    "weaviate.sh":         weaviateScript,
    "milvus.sh":           milvusScript,
    "qdrant.sh":           qdrantScript,
    "neo4j.sh":            neo4jScript,
    "postgres.sh":         postgresScript,
    "opensearch.sh":       opensearchScript,
    "elasticsearch.sh":    elasticsearchScript,
    "container/detect.sh": containerDetectScript,
    "container/run.sh":    containerRunScript,
}
```

#### 2. Extract on First Use

```go
// InitializeVDBScripts extracts embedded scripts to ~/.weave-cli/vdb/
func InitializeVDBScripts() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("failed to get home directory: %w", err)
    }

    vdbDir := filepath.Join(homeDir, ".weave-cli", "vdb")

    // Create directory if it doesn't exist
    if err := os.MkdirAll(vdbDir, 0755); err != nil {
        return fmt.Errorf("failed to create vdb directory: %w", err)
    }

    // Extract each embedded script
    for name, content := range embeddedScripts {
        scriptPath := filepath.Join(vdbDir, name)

        // Skip if user has customized version
        if _, err := os.Stat(scriptPath); err == nil {
            logging.Debug("Skipping %s (user-customized version exists)", name)
            continue
        }

        // Create subdirectories if needed
        scriptDir := filepath.Dir(scriptPath)
        if err := os.MkdirAll(scriptDir, 0755); err != nil {
            return fmt.Errorf("failed to create script directory: %w", err)
        }

        // Write script with executable permissions
        if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
            return fmt.Errorf("failed to write script %s: %w", name, err)
        }

        logging.Debug("Extracted VDB script: %s", scriptPath)
    }

    logging.Info("VDB scripts initialized in %s", vdbDir)
    return nil
}
```

#### 3. Script Resolution

```go
// GetScriptPath resolves the path to a VDB script
// Priority: 1) ~/.weave-cli/vdb/  2) tools/vdb/local/  (dev only)
func GetScriptPath(scriptName string) (string, error) {
    // 1. Check user directory (customized scripts)
    homeDir, err := os.UserHomeDir()
    if err == nil {
        userScript := filepath.Join(homeDir, ".weave-cli", "vdb", scriptName)
        if _, err := os.Stat(userScript); err == nil {
            return userScript, nil
        }
    }

    // 2. Check repo directory (development mode)
    repoScript := filepath.Join("tools", "vdb", "local", scriptName)
    if _, err := os.Stat(repoScript); err == nil {
        logging.Debug("Using repo script: %s", repoScript)
        return repoScript, nil
    }

    return "", fmt.Errorf("VDB script not found: %s", scriptName)
}
```

#### 4. Lazy Initialization

```go
// VDBManager with lazy script initialization
type VDBManager struct {
    containerRuntime string
    scriptsInitialized bool
}

func (m *VDBManager) ensureScripts() error {
    if m.scriptsInitialized {
        return nil
    }

    if err := InitializeVDBScripts(); err != nil {
        return err
    }

    m.scriptsInitialized = true
    return nil
}

func (m *VDBManager) Start(ctx context.Context, vdbType string) error {
    // Initialize scripts on first use
    if err := m.ensureScripts(); err != nil {
        return logging.WrapError(err, "Start", vdbType, "")
    }

    // ... rest of start logic
}
```

---

### 🔄 Script Updates & Customization

#### User Workflow: Customize Script

```bash
# 1. Extract scripts (automatic on first use)
weave vdb init  # Optional: force re-extract

# 2. Customize script
vi ~/.weave-cli/vdb/weaviate.sh
# Change port, storage, env vars, etc.

# 3. Use customized version
weave vdb start --weaviate-local
# Uses ~/.weave-cli/vdb/weaviate.sh (user version)
```

#### Update Scripts from New Release

```bash
# Force re-extract embedded scripts (overwrites user customizations)
weave vdb reset
# Warning: This will overwrite customized scripts!

# Or selectively reset one VDB
weave vdb reset --weaviate-local
```

#### Show Script Locations

```bash
weave vdb which --weaviate-local
# Output: Using script: /Users/max/.weave-cli/vdb/weaviate.sh (user-customized)

weave vdb which --all
# Output:
# weaviate-local:    /Users/max/.weave-cli/vdb/weaviate.sh (user-customized)
# milvus-local:      embedded (version 0.9.12)
# qdrant-local:      embedded (version 0.9.12)
```

---

### 📋 New Commands for Script Management

```bash
# Initialize/extract embedded scripts
weave vdb init [--force]

# Reset scripts to embedded versions
weave vdb reset [--<vdb-type>] [--all]

# Show script locations
weave vdb which [--<vdb-type>] [--all]

# Edit a VDB script
weave vdb edit --weaviate-local
# Opens ~/.weave-cli/vdb/weaviate.sh in $EDITOR

# Validate scripts (check syntax, permissions)
weave vdb validate [--<vdb-type>] [--all]
```

---

### 🎯 Benefits of This Approach

1. **Works in Production**: Binary contains everything needed
2. **Customizable**: Users can modify scripts in `~/.weave-cli/vdb/`
3. **Updateable**: New binary versions bring updated scripts
4. **Safe**: User customizations preserved (never overwritten without `--force`)
5. **Discoverable**: `weave vdb which` shows what's being used
6. **Consistent**: Follows existing pattern (`~/.weave-cli/agents/`, `~/.weave-cli/config.yaml`)
7. **Portable**: Works on any OS (Windows, macOS, Linux)

---

### 🚧 Edge Cases

#### Case 1: User Has Old Scripts

**Scenario**: User has old scripts in `~/.weave-cli/vdb/`, new binary has updates

**Solution**:
```bash
weave vdb start --weaviate-local
# Warning: ⚠️  Using user-customized script from 2025-12-01
#          New version available in binary (2026-01-28)
#          Run 'weave vdb reset --weaviate-local' to update

# Or show diff
weave vdb diff --weaviate-local
# Shows differences between user script and embedded version
```

#### Case 2: Script Corruption

**Scenario**: User accidentally breaks their customized script

**Solution**:
```bash
weave vdb validate --weaviate-local
# Error: Script has syntax errors at line 42

weave vdb reset --weaviate-local
# Restored weaviate.sh to embedded version (v0.9.12)
```

#### Case 3: Windows Path Issues

**Scenario**: Windows uses different path separators and shell

**Solution**:
- Use Go's `filepath` package for all paths
- Scripts are shell scripts → require WSL/Git Bash/Cygwin on Windows
- Detect OS and provide helpful error:
  ```
  ❌ VDB lifecycle management requires a Unix shell on Windows

  Please install one of:
    - WSL (Windows Subsystem for Linux) - recommended
    - Git Bash (comes with Git for Windows)
    - Cygwin

  After installing, run: weave vdb start --weaviate-local
  ```

---

## 🏗️ Proposed Architecture

### Command Structure

**Option A: `weave vdb` (Recommended)**
```bash
weave vdb start --weaviate-local
weave vdb stop --milvus-local
weave vdb status --qdrant-local
weave vdb ls
weave vdb restart --chroma-local
weave vdb logs --neo4j-local
weave vdb clean --weaviate-local  # Stop + remove volumes
```

**Pros**:
- Short and memorable (`vdb`)
- Follows existing pattern (`weave cols`, `weave docs`)
- No namespace conflicts

**Option B: `weave vector-db`**
```bash
weave vector-db start --weaviate-local
weave vector-db stop --milvus-local
```

**Cons**:
- Longer to type
- Hyphenated commands are less common

**Decision**: **Go with Option A (`weave vdb`)**

---

## 📋 Command Specifications

### 1. `weave vdb start`

**Syntax**:
```bash
weave vdb start [--<vdb-type>] [flags]
```

**Examples**:
```bash
weave vdb start --weaviate-local
weave vdb start --milvus-local --port 19530
weave vdb start --qdrant-local --storage /custom/path
```

**Flags**:
- `--weaviate-local` - Start Weaviate
- `--milvus-local` - Start Milvus
- `--qdrant-local` - Start Qdrant
- `--chroma-local` - Start Chroma
- `--neo4j-local` - Start Neo4j
- `--supabase-local` - Start PostgreSQL (for Supabase)
- `--opensearch-local` - Start OpenSearch
- `--elasticsearch-local` - Start Elasticsearch
- `--all` - Start all configured local VDBs
- `--port <port>` - Custom port (optional)
- `--storage <path>` - Custom storage path (optional)
- `--wait` - Wait for health check after start (default: true)
- `--timeout <duration>` - Health check timeout (default: 30s)

**Behavior**:
1. Check if VDB already running (idempotent)
2. Detect docker/podman runtime
3. Load config from config.yaml (ports, storage, env vars)
4. Execute start script via shell command
5. Wait for health check (verify VDB is ready)
6. Print success message with connection details

**Error Handling**:
- Container runtime not found → Helpful install instructions
- Port already in use → Suggest alternative port or stop conflicting service
- Health check fails → Show logs and troubleshooting steps
- Script not found → Warn about unsupported VDB type

---

### 2. `weave vdb stop`

**Syntax**:
```bash
weave vdb stop [--<vdb-type>] [flags]
```

**Examples**:
```bash
weave vdb stop --weaviate-local
weave vdb stop --all
```

**Flags**:
- Same VDB type flags as start
- `--all` - Stop all running local VDBs
- `--force` - Force stop (kill instead of graceful shutdown)

**Behavior**:
1. Check if VDB is running
2. Execute stop script
3. Wait for clean shutdown (with timeout)
4. Print success message

---

### 3. `weave vdb status`

**Syntax**:
```bash
weave vdb status [--<vdb-type>] [flags]
```

**Examples**:
```bash
weave vdb status --weaviate-local
weave vdb status --all
weave vdb status  # No flag = show all
```

**Output Format**:
```
VDB Status Report
================

Weaviate Local
  Status: ✅ Running
  Container: weaviate (1d2h ago)
  Ports: 8080 (HTTP), 50051 (gRPC)
  Storage: /Users/max/weave-cli/local/storage/weaviate_storage (1.2 GB)
  Health: ✅ Healthy (v1.24.5)
  Endpoint: http://localhost:8080

Milvus Local
  Status: ❌ Stopped
  Last seen: 2 days ago
  Storage: /Users/max/weave-cli/local/storage/milvus_storage (450 MB)

Qdrant Local
  Status: ⚠️ Unhealthy (container running but not responding)
  Container: qdrant (3h ago)
  Ports: 6333 (HTTP), 6334 (gRPC)
  Error: Connection timeout after 5s
```

**Flags**:
- `--json` - Output in JSON format
- `--short` - Compact one-line per VDB

---

### 4. `weave vdb ls`

**Syntax**:
```bash
weave vdb ls [flags]
```

**Output**:
```
Configured Local VDBs
=====================

NAME                  TYPE            STATUS      PORT    STORAGE
weaviate-local        Weaviate        Running     8080    1.2 GB
milvus-local          Milvus          Stopped     19530   450 MB
qdrant-local          Qdrant          Running     6333    230 MB
chroma-local          Chroma          Not Conf    8000    -
neo4j-local           Neo4j           Running     7474    1.8 GB
supabase-local        PostgreSQL      Stopped     5432    512 MB

Total: 6 VDBs (3 running, 3 stopped)
```

**Flags**:
- `--running` - Show only running VDBs
- `--stopped` - Show only stopped VDBs
- `--json` - JSON output

**Behavior**:
1. Read config.yaml to find all local VDB configurations
2. Check container status for each
3. Get storage sizes
4. Format output table

---

### 5. `weave vdb restart`

**Syntax**:
```bash
weave vdb restart [--<vdb-type>]
```

**Behavior**:
- Execute `stop` then `start`
- Useful for applying config changes

---

### 6. `weave vdb logs`

**Syntax**:
```bash
weave vdb logs [--<vdb-type>] [flags]
```

**Flags**:
- `--follow` / `-f` - Follow logs (tail -f style)
- `--lines <n>` - Show last N lines (default: 50)
- `--since <duration>` - Show logs since duration (e.g., "5m", "1h")

**Examples**:
```bash
weave vdb logs --weaviate-local
weave vdb logs --milvus-local --follow
weave vdb logs --qdrant-local --lines 100
```

---

### 7. `weave vdb clean`

**Syntax**:
```bash
weave vdb clean [--<vdb-type>] [flags]
```

**Behavior**:
- Stop container
- Remove container
- Remove volumes (with confirmation)
- ⚠️ **Destructive operation** - requires `--yes` or confirmation

**Flags**:
- `--yes` / `-y` - Skip confirmation
- `--keep-data` - Remove container but keep volumes

**Examples**:
```bash
weave vdb clean --weaviate-local
weave vdb clean --weaviate-local --yes
weave vdb clean --all --keep-data
```

---

## 🔧 Implementation Details

### File Structure

```
src/cmd/vdb/
  ├── vdb.go           # Root command
  ├── start.go         # Start subcommand
  ├── stop.go          # Stop subcommand
  ├── status.go        # Status subcommand
  ├── list.go          # List subcommand
  ├── restart.go       # Restart subcommand
  ├── logs.go          # Logs subcommand
  └── clean.go         # Clean subcommand

src/pkg/vdb/
  ├── lifecycle.go     # VDB lifecycle management
  ├── docker.go        # Docker/Podman abstraction
  ├── status.go        # Status checking and health
  ├── config.go        # VDB config from config.yaml
  └── scripts.go       # Script execution helpers

src/pkg/vdb/lifecycle/
  ├── weaviate.go      # Weaviate-specific lifecycle
  ├── milvus.go        # Milvus-specific lifecycle
  ├── qdrant.go        # Qdrant-specific lifecycle
  └── ...              # Other VDBs
```

### Core Types

```go
// VDBLifecycle manages the lifecycle of a local VDB
type VDBLifecycle struct {
    Name          string
    Type          VDBType
    ContainerName string
    Ports         map[string]int  // e.g., {"http": 8080, "grpc": 50051}
    StoragePath   string
    ScriptPath    string
    EnvVars       map[string]string
}

// VDBStatus represents the current status of a VDB
type VDBStatus struct {
    Name          string
    Type          VDBType
    Running       bool
    Healthy       bool
    ContainerID   string
    Uptime        time.Duration
    Ports         map[string]int
    StorageSize   int64
    Version       string
    ErrorMessage  string
}

// VDBManager orchestrates VDB lifecycle operations
type VDBManager struct {
    containerRuntime string  // "docker" or "podman"
    scriptDir        string  // path to tools/vdb/local
    config           *config.Config
    logger           *logging.Logger
}
```

### Key Functions

```go
// Start a local VDB instance
func (m *VDBManager) Start(ctx context.Context, vdbType string, opts StartOptions) error

// Stop a local VDB instance
func (m *VDBManager) Stop(ctx context.Context, vdbType string, opts StopOptions) error

// Get status of a VDB
func (m *VDBManager) Status(ctx context.Context, vdbType string) (*VDBStatus, error)

// List all configured local VDBs
func (m *VDBManager) List(ctx context.Context) ([]*VDBStatus, error)

// Execute a VDB management script
func (m *VDBManager) executeScript(scriptPath, command string, envVars map[string]string) error

// Check if VDB is running
func (m *VDBManager) IsRunning(ctx context.Context, containerName string) (bool, error)

// Wait for VDB health check
func (m *VDBManager) WaitForHealth(ctx context.Context, vdbType string, timeout time.Duration) error
```

---

## ⚙️ Config File Integration

### Existing config.yaml

```yaml
databases:
  vector_databases:
    - name: weaviate-local
      type: weaviate-local
      url: http://localhost:8080
      openai_api_key: ${OPENAI_API_KEY}
      timeout: 10
```

### Enhanced config.yaml (Optional)

```yaml
databases:
  vector_databases:
    - name: weaviate-local
      type: weaviate-local
      url: http://localhost:8080
      openai_api_key: ${OPENAI_API_KEY}
      timeout: 10

      # New: Lifecycle management settings
      lifecycle:
        enabled: true
        ports:
          http: 8080
          grpc: 50051
        storage_path: local/storage/weaviate_storage
        auto_start: false  # Start on first use
        health_check_timeout: 30s
        env_vars:
          QUERY_DEFAULTS_LIMIT: "25"
          DEFAULT_VECTORIZER_MODULE: "text2vec-openai"
```

### Migration Strategy

**Phase 1 (MVP)**: Use defaults from existing scripts, no config changes needed
**Phase 2 (Later)**: Add lifecycle section to config.yaml for customization

---

## 🧪 Testing Strategy

### Unit Tests

1. **VDBManager Tests** (`src/pkg/vdb/lifecycle_test.go`)
   - Test Start() with mock script execution
   - Test Stop() with mock script execution
   - Test Status() with mock container queries
   - Test List() with mock config
   - Test error handling (script not found, container runtime missing)

2. **Docker Abstraction Tests** (`src/pkg/vdb/docker_test.go`)
   - Test runtime detection (podman/docker/none)
   - Test container status queries
   - Test port conflict detection

### Integration Tests

3. **End-to-End Tests** (`tests/vdb_lifecycle_test.sh`)
   - Start Weaviate → verify running → stop → verify stopped
   - Start multiple VDBs → list → stop all
   - Test port conflicts
   - Test health checks
   - Test error scenarios

4. **Manual Testing Checklist**
   - [ ] `weave vdb start --weaviate-local` on macOS (Docker)
   - [ ] `weave vdb start --weaviate-local` on macOS (Podman)
   - [ ] `weave vdb start --weaviate-local` on Linux (Docker)
   - [ ] `weave vdb start --weaviate-local` on Linux (Podman)
   - [ ] `weave vdb status` with mixed running/stopped VDBs
   - [ ] `weave vdb ls` with all local VDBs configured
   - [ ] `weave vdb logs --weaviate-local --follow`
   - [ ] `weave vdb clean --weaviate-local` with confirmation
   - [ ] Error: No container runtime installed
   - [ ] Error: Port already in use
   - [ ] Error: Script not found for unsupported VDB

---

## 📅 Implementation Plan

### Phase 1: MVP (4-6 hours)

**Goal**: Basic start/stop/status for Weaviate, Milvus, Qdrant

**Tasks**:
1. ✅ **Planning** (1 hour) - This document
2. **Core Infrastructure** (2 hours)
   - [ ] Create `src/cmd/vdb/` command structure
   - [ ] Implement `VDBManager` in `src/pkg/vdb/lifecycle.go`
   - [ ] Implement container runtime detection
   - [ ] Implement script execution wrapper
   - [ ] Add logging and error context

3. **Commands** (1.5 hours)
   - [ ] Implement `weave vdb start`
   - [ ] Implement `weave vdb stop`
   - [ ] Implement `weave vdb status`
   - [ ] Implement `weave vdb ls`

4. **Testing** (1 hour)
   - [ ] Unit tests for VDBManager
   - [ ] Integration test: start → status → stop
   - [ ] Manual testing on macOS

5. **Documentation** (30 min)
   - [ ] Update README.md with VDB management section
   - [ ] Create docs/guides/VDB_LIFECYCLE.md

### Phase 2: Enhanced Features (2-3 hours) - Optional

**Goal**: Logs, restart, clean, config customization

**Tasks**:
- [ ] Implement `weave vdb logs`
- [ ] Implement `weave vdb restart`
- [ ] Implement `weave vdb clean`
- [ ] Add config.yaml lifecycle section support
- [ ] Add `--port` and `--storage` flags
- [ ] Enhanced status output (storage size, version)
- [ ] JSON output mode

### Phase 3: Polish (1-2 hours) - Optional

**Goal**: Production-ready UX

**Tasks**:
- [ ] Add `weave config vdb` for script customization
- [ ] Auto-start on first query (if `auto_start: true`)
- [ ] Progress indicators for start/stop
- [ ] Colorized output
- [ ] Add to quick start guide
- [ ] Video demo

---

## 🚧 Edge Cases & Considerations

### Port Conflicts

**Scenario**: User tries to start Weaviate but port 8080 already in use

**Solution**:
```
❌ Failed to start Weaviate: Port 8080 already in use

Troubleshooting:
  1. Check what's using port 8080:
     lsof -i :8080  (macOS/Linux)

  2. Stop the conflicting service or use a custom port:
     weave vdb start --weaviate-local --port 8081

  3. Update config.yaml to use port 8081 permanently
```

### Container Runtime Missing

**Scenario**: Neither Docker nor Podman installed

**Solution**:
```
❌ No container runtime found. Local VDBs require Docker or Podman.

Install Podman (recommended):
  macOS:   brew install podman && podman machine init && podman machine start
  Linux:   sudo apt install podman  (Debian/Ubuntu)

Or install Docker:
  macOS:   brew install --cask docker
  Linux:   https://docs.docker.com/engine/install/

After installing, run: weave vdb start --weaviate-local
```

### Health Check Timeout

**Scenario**: VDB container starts but never becomes healthy

**Solution**:
```
⚠️  Weaviate started but health check timed out after 30s

Container is running but not responding. Check logs:
  weave vdb logs --weaviate-local

Common causes:
  1. Slow startup (try --timeout 60s)
  2. Missing OPENAI_API_KEY (check .env file)
  3. Container resource limits

Status: Container running, health unknown
```

### Multiple Instances

**Scenario**: User wants two Weaviate instances (different ports)

**Solution** (Phase 2):
```yaml
# config.yaml
databases:
  vector_databases:
    - name: weaviate-local-dev
      type: weaviate-local
      url: http://localhost:8080
      lifecycle:
        ports:
          http: 8080

    - name: weaviate-local-test
      type: weaviate-local
      url: http://localhost:8081
      lifecycle:
        ports:
          http: 8081
```

```bash
weave vdb start --weaviate-local-dev
weave vdb start --weaviate-local-test
```

---

## 🎯 Success Criteria

### MVP (Phase 1)

- [x] Planning document created and reviewed
- [ ] `weave vdb start` works for Weaviate, Milvus, Qdrant
- [ ] `weave vdb stop` works for all started VDBs
- [ ] `weave vdb status` shows accurate status
- [ ] `weave vdb ls` lists all configured local VDBs
- [ ] Works on macOS with Docker
- [ ] Works on macOS with Podman
- [ ] Unit tests passing
- [ ] Integration test passing
- [ ] Documentation updated

### Phase 2 (Enhanced)

- [ ] `weave vdb logs` shows container logs
- [ ] `weave vdb restart` restarts VDB
- [ ] `weave vdb clean` removes containers and volumes
- [ ] Config customization via config.yaml
- [ ] Custom ports and storage paths
- [ ] JSON output mode
- [ ] Works on Linux

### Customer Satisfaction

- [ ] Customer can easily start local VDBs without memorizing Docker commands
- [ ] Status command provides clear overview of all VDBs
- [ ] Error messages are actionable and helpful
- [ ] Integration with existing config system is seamless
- [ ] No breaking changes to existing functionality

---

## 📝 Open Questions

1. **Auto-start behavior**: Should `weave cols ls --weaviate-local` auto-start Weaviate if not running?
   - **Recommendation**: No for MVP, add in Phase 2 with `auto_start: true` config option

2. **Container image versions**: Should we pin specific VDB versions or use `:latest`?
   - **Current**: Scripts use `:latest`
   - **Recommendation**: Keep `:latest` for MVP, add version pinning in config later

3. **Storage cleanup**: Should `weave vdb clean` require confirmation by default?
   - **Recommendation**: Yes, destructive operations should confirm unless `--yes` flag

4. **Cross-platform paths**: How to handle Windows paths?
   - **Recommendation**: Use Go's `filepath` package, test on Windows in Phase 2

5. **Config migration**: Should we auto-migrate config.yaml to add lifecycle sections?
   - **Recommendation**: No, use smart defaults. Users can opt-in to customization

---

## 🔗 Related Documentation

- [Vector DB Abstraction Guide](docs/guides/VECTOR_DB_ABSTRACTION.md)
- [VDB Support Matrix](docs/VDB_SUPPORT_MATRIX.md)
- [Configuration Guide](docs/guides/CONFIGURATION.md)
- [Existing Scripts](tools/vdb/local/)

---

**Created**: 2026-01-28
**Last Updated**: 2026-01-28
**Status**: Ready for Implementation
**Assigned**: TBD
**Target Release**: v0.9.13 or v0.10.0
