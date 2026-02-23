# Weave Stack - Phase 1 Days 4-5

**Status**: 🚧 In Progress (Days 1-3 Complete ✅)
**Goal**: Complete helm deployment, health checks, kubectl integration
**Timeline**: Feb 24-28, 2026

---

## Overview

**Completed (Days 1-3)**:
- ✅ Core types and config parsing
- ✅ Cluster management (Kind/Minikube)
- ✅ Helm chart generation

**Remaining (Days 4-5)**:
- 🎯 Helm install to K8s
- 🎯 Pod health monitoring
- 🎯 kubectl integration
- 🎯 Port forwarding
- 🎯 Log streaming

---

## Day 4 (Thu Feb 27): Helm Deployment + Health Checks

### Morning Tasks (2-3 hours)

#### 1. Copy Helm Templates to kubernetes/ (30 min)
```go
// src/pkg/stack/helm.go
func CopyHelmTemplates(outputDir string) error {
    // Copy templates/helm/weave-stack/ to kubernetes/
    // Preserve directory structure
    // Skip values.yaml (already generated)
}
```

**Files to copy**:
- Chart.yaml
- templates/_helpers.tpl
- templates/milvus-deployment.yaml
- templates/milvus-service.yaml
- templates/vectordb-pvc.yaml

#### 2. Implement Helm Install (1 hour)
```go
// src/pkg/stack/helm.go
func HelmInstall(chartPath, releaseName, namespace string, timeout time.Duration) error {
    // helm install <releaseName> <chartPath> \
    //   --namespace <namespace> \
    //   --create-namespace \
    //   --wait \
    //   --timeout <timeout>

    cmd := exec.Command("helm", "install", releaseName, chartPath,
        "--namespace", namespace,
        "--create-namespace",
        "--wait",
        "--timeout", timeout.String())

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("helm install failed: %w\nOutput: %s", err, string(output))
    }

    return nil
}
```

**Update src/cmd/stack/up.go**:
```go
// After generating Helm chart:
utils.PrintInfo("Deploying to Kubernetes...")

// Copy templates
if err := stackpkg.CopyHelmTemplates(helmDir); err != nil {
    return fmt.Errorf("failed to copy Helm templates: %w", err)
}

// Helm install
releaseName := config.Name
namespace := "default"
timeout := 5 * time.Minute

if err := stackpkg.HelmInstall(helmDir, releaseName, namespace, timeout); err != nil {
    return fmt.Errorf("failed to install Helm chart: %w", err)
}

utils.PrintSuccess("✅ Helm chart deployed successfully!")
```

#### 3. Pod Status Monitoring (1 hour)
```go
// src/pkg/stack/kubernetes.go
func WaitForPods(selector string, timeout time.Duration) error {
    // kubectl get pods -l <selector> -o json
    // Poll until all pods are Ready
    // Show progress with spinner

    start := time.Now()
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        if time.Since(start) > timeout {
            return fmt.Errorf("timeout waiting for pods")
        }

        ready, err := checkPodsReady(selector)
        if err != nil {
            return err
        }

        if ready {
            return nil
        }

        <-ticker.C
    }
}

func checkPodsReady(selector string) (bool, error) {
    cmd := exec.Command("kubectl", "get", "pods",
        "-l", selector,
        "-o", "jsonpath={.items[*].status.conditions[?(@.type==\"Ready\")].status}")

    output, err := cmd.Output()
    if err != nil {
        return false, err
    }

    // Check if all pods report "True"
    statuses := strings.Fields(string(output))
    for _, status := range statuses {
        if status != "True" {
            return false, nil
        }
    }

    return len(statuses) > 0, nil
}
```

### Afternoon Tasks (2-3 hours)

#### 4. Update `weave stack status` (1 hour)
```go
// src/cmd/stack/status.go
func runStatus(cmd *cobra.Command, args []string) error {
    // ... existing cluster info display ...

    fmt.Println()
    utils.PrintInfo("Services:")

    // Get pods
    pods, err := stackpkg.GetPods("app.kubernetes.io/instance=" + clusterInfo.Name)
    if err != nil {
        utils.PrintWarning(fmt.Sprintf("Failed to get pods: %v", err))
        return nil
    }

    if len(pods) == 0 {
        utils.PrintInfo("  No services deployed yet")
        utils.PrintInfo("  Run: weave stack up --runtime kind")
        return nil
    }

    // Print table
    fmt.Printf("  %-20s %-15s %-10s %s\n", "Component", "Status", "Health", "Uptime")
    fmt.Printf("  %s\n", strings.Repeat("-", 60))

    for _, pod := range pods {
        fmt.Printf("  %-20s %-15s %-10s %s\n",
            pod.Name,
            pod.Status,
            pod.Health,
            pod.Uptime)
    }

    return nil
}
```

#### 5. Implement `weave stack logs` (1 hour)
```go
// src/cmd/stack/logs.go
package stack

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/spf13/cobra"
)

var (
    logsFollow bool
    logsTail   int
)

var LogsCmd = &cobra.Command{
    Use:   "logs <service>",
    Short: "Stream logs from a service",
    Args:  cobra.ExactArgs(1),
    RunE:  runLogs,
}

func init() {
    LogsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
    LogsCmd.Flags().IntVar(&logsTail, "tail", 100, "Number of lines to show")
}

func runLogs(cmd *cobra.Command, args []string) error {
    service := args[0]

    // Get cluster info
    clusterInfo, err := stackpkg.LoadClusterInfo()
    if err != nil || clusterInfo == nil {
        return fmt.Errorf("no active cluster found")
    }

    // Build kubectl logs command
    cmdArgs := []string{
        "logs",
        "-l", fmt.Sprintf("app=%s", service),
        "--context", clusterInfo.Context,
        "--tail", fmt.Sprintf("%d", logsTail),
    }

    if logsFollow {
        cmdArgs = append(cmdArgs, "--follow")
    }

    // Run kubectl logs
    kubectlCmd := exec.Command("kubectl", cmdArgs...)
    kubectlCmd.Stdout = os.Stdout
    kubectlCmd.Stderr = os.Stderr

    return kubectlCmd.Run()
}
```

**Register command**:
```go
// src/cmd/stack.go
stackCmd.AddCommand(stack.LogsCmd)
```

#### 6. Error Handling (1 hour)
- Helm install failures → show helm status, suggest debug commands
- Pod startup failures → show pod events, logs
- Timeout errors → suggest increasing timeout
- Missing dependencies → check for helm, kubectl

---

## Day 5 (Fri Mar 1): kubectl Integration + Polish

### Morning Tasks (2-3 hours)

#### 1. kubectl Passthrough (1 hour)
```go
// src/cmd/stack/kubectl.go
package stack

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/spf13/cobra"
)

var KubectlCmd = &cobra.Command{
    Use:   "kubectl -- <kubectl-args>",
    Short: "Run kubectl commands with stack context",
    Long: `Run kubectl commands with the stack's context automatically set.

Example:
  weave stack kubectl -- get pods
  weave stack kubectl -- describe svc weave-stack-milvus`,
    RunE: runKubectl,
}

func runKubectl(cmd *cobra.Command, args []string) error {
    // Get cluster info
    clusterInfo, err := stackpkg.LoadClusterInfo()
    if err != nil || clusterInfo == nil {
        return fmt.Errorf("no active cluster found")
    }

    // Build kubectl command with context
    kubectlArgs := append([]string{"--context", clusterInfo.Context}, args...)

    // Run kubectl
    kubectlCmd := exec.Command("kubectl", kubectlArgs...)
    kubectlCmd.Stdout = os.Stdout
    kubectlCmd.Stderr = os.Stderr
    kubectlCmd.Stdin = os.Stdin

    return kubectlCmd.Run()
}
```

#### 2. Port Forwarding (1 hour)
```go
// src/cmd/stack/port_forward.go
package stack

import (
    "fmt"
    "os"
    "os/exec"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
)

var PortForwardCmd = &cobra.Command{
    Use:   "port-forward <service> <port-mapping>",
    Short: "Forward local port to service",
    Args:  cobra.ExactArgs(2),
    Long: `Forward a local port to a service in the stack.

Example:
  weave stack port-forward milvus 19530:19530
  # Access Milvus at localhost:19530`,
    RunE: runPortForward,
}

func runPortForward(cmd *cobra.Command, args []string) error {
    service := args[0]
    portMapping := args[1]

    // Get cluster info
    clusterInfo, err := stackpkg.LoadClusterInfo()
    if err != nil || clusterInfo == nil {
        return fmt.Errorf("no active cluster found")
    }

    // Get service name
    serviceName := fmt.Sprintf("%s-%s", clusterInfo.Name, service)

    // Build kubectl port-forward command
    kubectlCmd := exec.Command("kubectl", "port-forward",
        fmt.Sprintf("svc/%s", serviceName),
        portMapping,
        "--context", clusterInfo.Context)

    kubectlCmd.Stdout = os.Stdout
    kubectlCmd.Stderr = os.Stderr

    // Handle Ctrl+C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        kubectlCmd.Process.Kill()
    }()

    fmt.Printf("Forwarding port %s to %s...\n", portMapping, serviceName)
    fmt.Println("Press Ctrl+C to stop")

    return kubectlCmd.Run()
}
```

### Afternoon Tasks (2-3 hours)

#### 3. Dependency Checks (1 hour)
```go
// src/pkg/stack/dependencies.go
package stack

import (
    "fmt"
    "os/exec"
)

type Dependency struct {
    Name        string
    Command     string
    Required    bool
    InstallDocs string
}

var RequiredDependencies = []Dependency{
    {
        Name:        "kubectl",
        Command:     "kubectl",
        Required:    true,
        InstallDocs: "https://kubernetes.io/docs/tasks/tools/install-kubectl/",
    },
    {
        Name:        "helm",
        Command:     "helm",
        Required:    true,
        InstallDocs: "https://helm.sh/docs/intro/install/",
    },
    {
        Name:        "kind",
        Command:     "kind",
        Required:    false, // Only for local Kind clusters
        InstallDocs: "https://kind.sigs.k8s.io/docs/user/quick-start/",
    },
}

func CheckDependencies() error {
    var missing []Dependency

    for _, dep := range RequiredDependencies {
        if !commandExists(dep.Command) {
            if dep.Required {
                missing = append(missing, dep)
            }
        }
    }

    if len(missing) > 0 {
        return fmt.Errorf("missing required dependencies: %v", missing)
    }

    return nil
}

func commandExists(cmd string) bool {
    _, err := exec.LookPath(cmd)
    return err == nil
}
```

**Add check to `weave stack up`**:
```go
// Check dependencies before deployment
if err := stackpkg.CheckDependencies(); err != nil {
    return err
}
```

#### 4. Progress Indicators (1 hour)
```go
// src/pkg/stack/progress.go
package stack

import (
    "fmt"
    "time"
)

func ShowSpinner(message string, done <-chan bool) {
    spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    i := 0

    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-done:
            fmt.Print("\r")
            return
        case <-ticker.C:
            fmt.Printf("\r%s %s", spinner[i%len(spinner)], message)
            i++
        }
    }
}
```

**Usage in helm install**:
```go
done := make(chan bool)
go stackpkg.ShowSpinner("Deploying to Kubernetes...", done)

err := stackpkg.HelmInstall(helmDir, releaseName, namespace, timeout)

close(done)
```

#### 5. Documentation (1 hour)
- Update README with Weave Stack section
- Add troubleshooting guide
- Document commands

---

## Testing Strategy

### Unit Tests
```bash
# Test helm operations
go test ./src/pkg/stack/... -run TestHelmInstall
go test ./src/pkg/stack/... -run TestCopyHelmTemplates

# Test kubernetes operations
go test ./src/pkg/stack/... -run TestWaitForPods
go test ./src/pkg/stack/... -run TestGetPods
```

### Integration Test Script
```bash
#!/bin/bash
# scripts/test-stack-phase1.sh

set -e

echo "=== Phase 1 Integration Test ==="

# 1. Init
echo "[1/6] Testing: weave stack init"
cd /tmp/test-stack-$$
weave stack init --template quickstart
test -f weave-stack.yaml

# 2. Validate
echo "[2/6] Testing: weave stack validate"
weave stack validate

# 3. Create cluster
echo "[3/6] Testing: weave stack up (cluster creation)"
weave stack up --runtime kind

# 4. Wait for pods
echo "[4/6] Testing: Pod readiness"
sleep 30
kubectl get pods | grep milvus | grep Running

# 5. Status
echo "[5/6] Testing: weave stack status"
weave stack status

# 6. Cleanup
echo "[6/6] Testing: weave stack down"
weave stack down

echo "✅ All tests passed!"
```

---

## Success Criteria

**Day 4**:
- [ ] `weave stack up --runtime kind` deploys Milvus successfully
- [ ] Milvus pod reaches Ready state within 5 minutes
- [ ] `weave stack status` shows pod health
- [ ] `weave stack logs milvus` streams logs
- [ ] All tests passing

**Day 5**:
- [ ] `weave stack kubectl -- get pods` works
- [ ] `weave stack port-forward milvus 19530:19530` forwards traffic
- [ ] Can connect to Milvus via localhost:19530
- [ ] Dependency checks warn about missing tools
- [ ] Progress indicators show during deployment

---

## Files to Create/Modify

**New Files**:
- src/pkg/stack/kubernetes.go
- src/pkg/stack/dependencies.go
- src/pkg/stack/progress.go
- src/cmd/stack/logs.go
- src/cmd/stack/kubectl.go
- src/cmd/stack/port_forward.go
- scripts/test-stack-phase1.sh

**Modified Files**:
- src/cmd/stack/up.go (add helm install)
- src/cmd/stack/status.go (show pod status)
- src/pkg/stack/helm.go (add CopyHelmTemplates, HelmInstall)
- src/cmd/stack.go (register new commands)

---

**Updated**: Feb 23, 2026 21:00 PST
