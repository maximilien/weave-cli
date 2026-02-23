# Weave Stack - PM2 Dashboard Integration

**Status**: 📋 Planning
**Priority**: HIGH (Client0 already has this)
**Goal**: Add PM2 process management for dashboard to weave-stack
**Reference**: Client0 implementation at `/Users/maximilien/github/auctionsmax-ai`

---

## Overview

Client0 (auctionsmax-ai) uses PM2 for production-grade process management of their Next.js/Express frontend. We need to integrate similar PM2 support into weave-stack for dashboard management.

### Client0 Implementation Summary

**Files**:
- `ecosystem.config.js` - PM2 configuration
- `scripts/start-frontend.sh` - PM2 startup script
- `scripts/stop-frontend.sh` - PM2 shutdown script
- `tools/ui/` - Development lifecycle scripts

**Features**:
- Auto-restart on crashes
- Memory limit monitoring (1GB)
- Log rotation and management
- Graceful shutdown
- Process monitoring with `pm2 monit`

---

## Design Decisions

### 1. Deployment Context

PM2 is primarily for **local/VM deployments**, not Kubernetes:

| Environment | Process Manager | Notes |
|-------------|----------------|-------|
| Local (dev) | PM2 | ✅ Full PM2 features |
| Local (prod) | PM2 | ✅ Production-ready |
| Kubernetes | Built-in | ❌ Use K8s native (replicas, probes) |
| VM/Bare Metal | PM2 | ✅ Ideal for single-server setups |

**Decision**: Support both PM2 (local/VM) and K8s (cloud) deployment modes.

### 2. Stack Configuration

Add PM2 configuration to dashboard config in `weave-stack.yaml`:

```yaml
dashboard:
  enabled: true
  type: "web"
  runtime: "pm2"  # or "kubernetes", "docker", "manual"

  pm2:
    app_name: "weave-dashboard"
    instances: 1
    max_memory_restart: "1G"
    autorestart: true
    watch: false
    error_log: "logs/pm2-error.log"
    out_log: "logs/pm2-out.log"
    min_uptime: "10s"
    max_restarts: 10
    kill_timeout: 5000
    env:
      NODE_ENV: "production"
      DASHBOARD_PORT: 3000
      DASHBOARD_HOST: "0.0.0.0"

  web:
    framework: "nextjs"  # or "react", "vue"
    port: 3000
    build_command: "npm run build"
    start_command: "npm start"
    dev_command: "npm run dev"
```

### 3. Command Structure

Extend `weave stack` commands to support dashboard lifecycle:

```bash
# Dashboard management
weave stack dashboard start    # Start with PM2
weave stack dashboard stop     # Stop PM2 process
weave stack dashboard restart  # Restart PM2 process
weave stack dashboard status   # Show PM2 status
weave stack dashboard logs     # Stream PM2 logs
weave stack dashboard monit    # Open PM2 monitor

# Integrated stack commands
weave stack up --with-dashboard     # Deploy VDB + Dashboard
weave stack status                  # Show VDB + Dashboard status
weave stack down                    # Stop VDB + Dashboard
```

---

## Implementation Plan

### Phase 1: Core PM2 Support (Day 4 Afternoon - Tonight)

#### Task 1: Update Stack Types (30 min)

**File**: `src/pkg/stack/types.go`

```go
// PM2Config defines PM2 process manager configuration
type PM2Config struct {
    AppName           string            `yaml:"app_name"`
    Instances         int               `yaml:"instances"`
    MaxMemoryRestart  string            `yaml:"max_memory_restart"`
    Autorestart       bool              `yaml:"autorestart"`
    Watch             bool              `yaml:"watch"`
    ErrorLog          string            `yaml:"error_log"`
    OutLog            string            `yaml:"out_log"`
    MinUptime         string            `yaml:"min_uptime"`
    MaxRestarts       int               `yaml:"max_restarts"`
    KillTimeout       int               `yaml:"kill_timeout"`
    Env               map[string]string `yaml:"env"`
}

// DashboardConfig (update existing)
type DashboardConfig struct {
    Enabled  bool          `yaml:"enabled"`
    Type     string        `yaml:"type"` // "web", "cli", "none"
    Runtime  string        `yaml:"runtime"` // "pm2", "kubernetes", "docker", "manual"
    PM2      *PM2Config    `yaml:"pm2,omitempty"`
    Web      *WebDashboard `yaml:"web,omitempty"`
}
```

#### Task 2: Create PM2 Template (30 min)

**File**: `templates/pm2/ecosystem.config.js`

```javascript
module.exports = {
  apps: [
    {
      name: '{{ .AppName }}',
      cwd: '{{ .WorkingDir }}',
      script: '{{ .ScriptPath }}',
      instances: {{ .Instances }},
      autorestart: {{ .Autorestart }},
      watch: {{ .Watch }},
      max_memory_restart: '{{ .MaxMemoryRestart }}',
      env: {
        {{- range $key, $value := .Env }}
        {{ $key }}: '{{ $value }}',
        {{- end }}
      },
      error_file: '{{ .ErrorLog }}',
      out_file: '{{ .OutLog }}',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
      merge_logs: true,
      min_uptime: '{{ .MinUptime }}',
      max_restarts: {{ .MaxRestarts }},
      kill_timeout: {{ .KillTimeout }},
    },
  ],
};
```

#### Task 3: Implement PM2 Management (1 hour)

**File**: `src/pkg/stack/pm2.go` (NEW)

```go
package stack

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "text/template"
)

// GeneratePM2Config generates ecosystem.config.js from template
func GeneratePM2Config(config *StackConfig, outputPath string) error {
    if config.Dashboard == nil || config.Dashboard.PM2 == nil {
        return fmt.Errorf("PM2 config not found in stack config")
    }

    // Load template
    tmpl, err := template.ParseFiles("templates/pm2/ecosystem.config.js")
    if err != nil {
        return fmt.Errorf("failed to load PM2 template: %w", err)
    }

    // Create output file
    f, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("failed to create PM2 config: %w", err)
    }
    defer f.Close()

    // Execute template
    data := map[string]interface{}{
        "AppName":          config.Dashboard.PM2.AppName,
        "WorkingDir":       filepath.Dir(outputPath),
        "ScriptPath":       "dist/index.js", // From config
        "Instances":        config.Dashboard.PM2.Instances,
        "Autorestart":      config.Dashboard.PM2.Autorestart,
        "Watch":            config.Dashboard.PM2.Watch,
        "MaxMemoryRestart": config.Dashboard.PM2.MaxMemoryRestart,
        "ErrorLog":         config.Dashboard.PM2.ErrorLog,
        "OutLog":           config.Dashboard.PM2.OutLog,
        "MinUptime":        config.Dashboard.PM2.MinUptime,
        "MaxRestarts":      config.Dashboard.PM2.MaxRestarts,
        "KillTimeout":      config.Dashboard.PM2.KillTimeout,
        "Env":              config.Dashboard.PM2.Env,
    }

    return tmpl.Execute(f, data)
}

// PM2Start starts the dashboard with PM2
func PM2Start(appName, configPath string) error {
    // Check if PM2 is installed
    if !commandExists("pm2") {
        return fmt.Errorf("pm2 not installed (install: npm install -g pm2)")
    }

    // Stop existing process
    _ = PM2Stop(appName) // Ignore errors

    // Start with PM2
    cmd := exec.Command("pm2", "start", configPath)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("pm2 start failed: %w\nOutput: %s", err, string(output))
    }

    return nil
}

// PM2Stop stops the PM2 process
func PM2Stop(appName string) error {
    if !commandExists("pm2") {
        return fmt.Errorf("pm2 not installed")
    }

    // Stop process
    cmd := exec.Command("pm2", "stop", appName)
    _ = cmd.Run() // Ignore errors if not running

    // Delete process
    cmd = exec.Command("pm2", "delete", appName)
    _ = cmd.Run() // Ignore errors

    return nil
}

// PM2Status returns the PM2 process status
func PM2Status(appName string) (string, error) {
    if !commandExists("pm2") {
        return "", fmt.Errorf("pm2 not installed")
    }

    cmd := exec.Command("pm2", "status", appName)
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("pm2 status failed: %w", err)
    }

    return string(output), nil
}

// PM2Logs streams logs from PM2
func PM2Logs(appName string, lines int, follow bool) error {
    if !commandExists("pm2") {
        return fmt.Errorf("pm2 not installed")
    }

    args := []string{"logs", appName, "--lines", fmt.Sprintf("%d", lines)}
    if !follow {
        args = append(args, "--nostream")
    }

    cmd := exec.Command("pm2", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

func commandExists(cmd string) bool {
    _, err := exec.LookPath(cmd)
    return err == nil
}
```

#### Task 4: Add Dashboard Commands (1 hour)

**File**: `src/cmd/stack/dashboard.go` (NEW)

```go
package stack

import (
    "fmt"

    "github.com/maximilien/weave-cli/src/cmd/utils"
    stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
    "github.com/spf13/cobra"
)

// DashboardCmd represents the dashboard command
var DashboardCmd = &cobra.Command{
    Use:   "dashboard",
    Short: "Manage stack dashboard",
    Long:  `Start, stop, and manage the weave stack dashboard.`,
}

var dashboardStartCmd = &cobra.Command{
    Use:   "start",
    Short: "Start dashboard with PM2",
    RunE:  runDashboardStart,
}

var dashboardStopCmd = &cobra.Command{
    Use:   "stop",
    Short: "Stop dashboard PM2 process",
    RunE:  runDashboardStop,
}

var dashboardStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show dashboard status",
    RunE:  runDashboardStatus,
}

var dashboardLogsCmd = &cobra.Command{
    Use:   "logs",
    Short: "Stream dashboard logs",
    RunE:  runDashboardLogs,
}

func init() {
    DashboardCmd.AddCommand(dashboardStartCmd)
    DashboardCmd.AddCommand(dashboardStopCmd)
    DashboardCmd.AddCommand(dashboardStatusCmd)
    DashboardCmd.AddCommand(dashboardLogsCmd)
}

func runDashboardStart(cmd *cobra.Command, args []string) error {
    utils.PrintHeader("Starting Dashboard")

    // Load stack config
    config, err := stackpkg.LoadStackConfig("")
    if err != nil {
        return fmt.Errorf("failed to load stack config: %w", err)
    }

    if config.Dashboard == nil || !config.Dashboard.Enabled {
        return fmt.Errorf("dashboard not enabled in weave-stack.yaml")
    }

    if config.Dashboard.Runtime != "pm2" {
        return fmt.Errorf("dashboard runtime is %s, not pm2", config.Dashboard.Runtime)
    }

    // Generate PM2 config
    pm2ConfigPath := "ecosystem.config.js"
    if err := stackpkg.GeneratePM2Config(config, pm2ConfigPath); err != nil {
        return fmt.Errorf("failed to generate PM2 config: %w", err)
    }

    utils.PrintInfo(fmt.Sprintf("Generated PM2 config: %s", pm2ConfigPath))

    // Start with PM2
    appName := config.Dashboard.PM2.AppName
    if err := stackpkg.PM2Start(appName, pm2ConfigPath); err != nil {
        return fmt.Errorf("failed to start dashboard: %w", err)
    }

    utils.PrintSuccess(fmt.Sprintf("✅ Dashboard started: %s", appName))
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println("  pm2 monit                     # Monitor processes")
    fmt.Println("  pm2 logs", appName, "         # View logs")
    fmt.Println("  weave stack dashboard stop    # Stop dashboard")

    return nil
}

func runDashboardStop(cmd *cobra.Command, args []string) error {
    utils.PrintHeader("Stopping Dashboard")

    config, err := stackpkg.LoadStackConfig("")
    if err != nil {
        return fmt.Errorf("failed to load stack config: %w", err)
    }

    if config.Dashboard == nil || config.Dashboard.PM2 == nil {
        return fmt.Errorf("PM2 config not found")
    }

    appName := config.Dashboard.PM2.AppName
    if err := stackpkg.PM2Stop(appName); err != nil {
        return fmt.Errorf("failed to stop dashboard: %w", err)
    }

    utils.PrintSuccess(fmt.Sprintf("✅ Dashboard stopped: %s", appName))
    return nil
}

func runDashboardStatus(cmd *cobra.Command, args []string) error {
    utils.PrintHeader("Dashboard Status")

    config, err := stackpkg.LoadStackConfig("")
    if err != nil {
        return fmt.Errorf("failed to load stack config: %w", err)
    }

    if config.Dashboard == nil || config.Dashboard.PM2 == nil {
        return fmt.Errorf("PM2 config not found")
    }

    appName := config.Dashboard.PM2.AppName
    status, err := stackpkg.PM2Status(appName)
    if err != nil {
        return fmt.Errorf("failed to get status: %w", err)
    }

    fmt.Println(status)
    return nil
}

func runDashboardLogs(cmd *cobra.Command, args []string) error {
    config, err := stackpkg.LoadStackConfig("")
    if err != nil {
        return fmt.Errorf("failed to load stack config: %w", err)
    }

    if config.Dashboard == nil || config.Dashboard.PM2 == nil {
        return fmt.Errorf("PM2 config not found")
    }

    appName := config.Dashboard.PM2.AppName
    return stackpkg.PM2Logs(appName, 100, true)
}
```

**Register in**: `src/cmd/stack.go`

```go
stackCmd.AddCommand(stack.DashboardCmd)
```

---

## Usage Examples

### Quick Start (Local Development)

```bash
# 1. Initialize stack with dashboard
cat > weave-stack.yaml <<EOF
version: "1.0"
name: my-stack

dashboard:
  enabled: true
  runtime: pm2
  pm2:
    app_name: weave-dashboard
    instances: 1
    max_memory_restart: "1G"
    autorestart: true
    env:
      NODE_ENV: production
      DASHBOARD_PORT: 3000
  web:
    framework: nextjs
    port: 3000
EOF

# 2. Build dashboard
cd dashboard && npm run build && cd ..

# 3. Start dashboard
weave stack dashboard start

# 4. Monitor
pm2 monit

# 5. Stop
weave stack dashboard stop
```

### Production (Kubernetes)

```yaml
dashboard:
  enabled: true
  runtime: kubernetes  # Not PM2!
  web:
    framework: nextjs
    kubernetes:
      deployment:
        replicas: 2
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
```

---

## Testing Strategy

### Unit Tests

**File**: `src/pkg/stack/pm2_test.go`

```go
func TestGeneratePM2Config(t *testing.T) {
    config := &StackConfig{
        Dashboard: &DashboardConfig{
            PM2: &PM2Config{
                AppName: "test-app",
                Instances: 2,
                // ...
            },
        },
    }

    tmpFile := filepath.Join(t.TempDir(), "ecosystem.config.js")
    err := GeneratePM2Config(config, tmpFile)
    require.NoError(t, err)

    // Verify file exists and contains expected content
    content, err := os.ReadFile(tmpFile)
    require.NoError(t, err)
    assert.Contains(t, string(content), "test-app")
}

func TestPM2Start(t *testing.T) {
    // Requires PM2 installed - skip if not available
    if !commandExists("pm2") {
        t.Skip("pm2 not installed")
    }

    // Test with mock config
    // ...
}
```

### Integration Tests

```bash
#!/bin/bash
# tests/integration/test-pm2-dashboard.sh

# 1. Create test stack
weave stack init --template with-dashboard

# 2. Build dashboard
cd dashboard && npm run build && cd ..

# 3. Start with PM2
weave stack dashboard start

# 4. Check status
pm2 list | grep weave-dashboard

# 5. Stop
weave stack dashboard stop

# 6. Verify stopped
! pm2 list | grep weave-dashboard
```

---

## Migration Path

For users with existing Client0-style PM2 setups:

1. **Keep existing ecosystem.config.js**: Weave can generate but not override
2. **Gradual adoption**: Start with `weave stack dashboard start` wrapper
3. **Backward compatible**: Support both PM2 and manual modes

---

## Next Steps

1. **Tonight (Feb 23)**: Implement Phase 1 (Tasks 1-4)
2. **Tomorrow (Feb 24)**: Test with Client0 dashboard
3. **Thursday (Feb 27)**: Kubernetes integration
4. **Friday (Feb 28)**: Documentation and demos

---

**Updated**: Feb 23, 2026 21:50 PST
