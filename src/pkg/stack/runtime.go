// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os/exec"
	"strings"
)

// ContainerRuntime represents a container runtime
type ContainerRuntime string

const (
	RuntimePodman ContainerRuntime = "podman"
	RuntimeDocker ContainerRuntime = "docker"
	RuntimeNone   ContainerRuntime = "none"
)

// DetectContainerRuntime detects which container runtime is available
// Priority: podman > docker (OSS-first preference)
func DetectContainerRuntime() (ContainerRuntime, error) {
	// Try podman first (OSS preference)
	if isRuntimeAvailable("podman") {
		return RuntimePodman, nil
	}

	// Fall back to docker
	if isRuntimeAvailable("docker") {
		return RuntimeDocker, nil
	}

	return RuntimeNone, fmt.Errorf("no container runtime found (tried: podman, docker)")
}

// isRuntimeAvailable checks if a container runtime is available
func isRuntimeAvailable(runtime string) bool {
	cmd := exec.Command(runtime, "--version")
	err := cmd.Run()
	return err == nil
}

// GetRuntimeVersion gets the version of a container runtime
func GetRuntimeVersion(runtime ContainerRuntime) (string, error) {
	cmd := exec.Command(string(runtime), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get %s version: %w", runtime, err)
	}

	// Parse version from output
	// podman: "podman version 4.9.3"
	// docker: "Docker version 24.0.7, build afdd53b"
	version := strings.TrimSpace(string(output))
	return version, nil
}

// ValidateRuntime ensures the specified runtime is available
func ValidateRuntime(runtime ContainerRuntime) error {
	if runtime == RuntimeNone {
		return fmt.Errorf("no container runtime specified")
	}

	if !isRuntimeAvailable(string(runtime)) {
		return fmt.Errorf("%s is not available or not installed", runtime)
	}

	return nil
}

// GetRuntimeCommand returns the appropriate runtime command based on detection
func GetRuntimeCommand(preferredRuntime string) (ContainerRuntime, error) {
	// If user specified a preference, validate it
	if preferredRuntime != "" {
		runtime := ContainerRuntime(preferredRuntime)
		if err := ValidateRuntime(runtime); err != nil {
			return RuntimeNone, err
		}
		return runtime, nil
	}

	// Otherwise, auto-detect
	return DetectContainerRuntime()
}
