// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/version"
)

// CheckSystem verifies CLI version, Go version, OS/arch, and external dependencies.
func CheckSystem() []CheckResult {
	var results []CheckResult

	// CLI version
	info := version.Get()
	results = append(results, CheckResult{
		Section: SectionSystem,
		Name:    "CLI version",
		Status:  StatusOK,
		Message: info.Version,
	})

	// Go version
	results = append(results, CheckResult{
		Section: SectionSystem,
		Name:    "Go version",
		Status:  StatusOK,
		Message: runtime.Version(),
	})

	// OS / Arch
	results = append(results, CheckResult{
		Section: SectionSystem,
		Name:    "OS/Arch",
		Status:  StatusOK,
		Message: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	})

	// External tool checks
	for _, tool := range []struct {
		name    string
		fix     string
		warning bool // true = warn instead of fail when missing
	}{
		{"kubectl", "Install kubectl: https://kubernetes.io/docs/tasks/tools/", true},
		{"helm", "Install helm: https://helm.sh/docs/intro/install/", true},
		{"tesseract", "Install tesseract: brew install tesseract (macOS) / apt install tesseract-ocr (Linux)", true},
	} {
		results = append(results, checkBinary(tool.name, tool.fix, tool.warning))
	}

	return results
}

func checkBinary(name, fix string, warnOnly bool) CheckResult {
	path, err := exec.LookPath(name)
	if err != nil {
		status := StatusFail
		if warnOnly {
			status = StatusWarn
		}
		return CheckResult{
			Section: SectionSystem,
			Name:    name,
			Status:  status,
			Message: "not found in PATH",
			Fix:     fix,
		}
	}

	// Try to get version
	ver := binaryVersion(name)
	msg := path
	if ver != "" {
		msg = fmt.Sprintf("%s (%s)", ver, path)
	}

	return CheckResult{
		Section: SectionSystem,
		Name:    name,
		Status:  StatusOK,
		Message: msg,
	}
}

func binaryVersion(name string) string {
	cmd := exec.Command(name, "version", "--short")
	if name == "tesseract" {
		cmd = exec.Command(name, "--version")
	} else if name == "helm" {
		cmd = exec.Command(name, "version", "--short")
	} else if name == "kubectl" {
		cmd = exec.Command(name, "version", "--client", "--short")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// Return first line only
	line := strings.TrimSpace(string(out))
	if idx := strings.IndexByte(line, '\n'); idx > 0 {
		line = line[:idx]
	}
	return line
}
