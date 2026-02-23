// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectContainerRuntime(t *testing.T) {
	runtime, err := DetectContainerRuntime()

	// Should find at least one runtime (podman or docker)
	// This test will pass if either is installed
	if err != nil {
		t.Logf("No container runtime found: %v", err)
		assert.Equal(t, RuntimeNone, runtime)
	} else {
		t.Logf("Detected runtime: %s", runtime)
		assert.True(t, runtime == RuntimePodman || runtime == RuntimeDocker)
	}
}

func TestIsRuntimeAvailable(t *testing.T) {
	tests := []struct {
		name    string
		runtime string
		// We can't assert specific values since it depends on the system
		// but we can verify the function works
	}{
		{"check podman", "podman"},
		{"check docker", "docker"},
		{"check nonexistent", "nonexistent-runtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := isRuntimeAvailable(tt.runtime)
			t.Logf("%s available: %v", tt.runtime, available)

			// Verify it matches exec.Command result
			cmd := exec.Command(tt.runtime, "--version")
			err := cmd.Run()
			expectedAvailable := (err == nil)
			assert.Equal(t, expectedAvailable, available)
		})
	}
}

func TestGetRuntimeVersion(t *testing.T) {
	// Test with detected runtime
	runtime, err := DetectContainerRuntime()
	if err != nil {
		t.Skip("No container runtime available")
	}

	version, err := GetRuntimeVersion(runtime)
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
	t.Logf("%s version: %s", runtime, version)
}

func TestValidateRuntime(t *testing.T) {
	tests := []struct {
		name    string
		runtime ContainerRuntime
		wantErr bool
	}{
		{
			name:    "none runtime",
			runtime: RuntimeNone,
			wantErr: true,
		},
		{
			name:    "invalid runtime",
			runtime: "invalid-runtime",
			wantErr: true,
		},
	}

	// Add test for available runtime
	detected, err := DetectContainerRuntime()
	if err == nil {
		tests = append(tests, struct {
			name    string
			runtime ContainerRuntime
			wantErr bool
		}{
			name:    "available runtime",
			runtime: detected,
			wantErr: false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRuntime(tt.runtime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRuntimeCommand(t *testing.T) {
	tests := []struct {
		name      string
		preferred string
		wantErr   bool
	}{
		{
			name:      "auto-detect",
			preferred: "",
			wantErr:   false, // Should succeed if any runtime available
		},
		{
			name:      "invalid preference",
			preferred: "invalid-runtime",
			wantErr:   true,
		},
	}

	// Add test for valid preference if runtime is available
	detected, err := DetectContainerRuntime()
	if err == nil {
		tests = append(tests, struct {
			name      string
			preferred string
			wantErr   bool
		}{
			name:      "valid preference",
			preferred: string(detected),
			wantErr:   false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime, err := GetRuntimeCommand(tt.preferred)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				if err != nil {
					// No runtime available on system
					t.Skipf("No container runtime available: %v", err)
				}
				assert.NoError(t, err)
				assert.NotEqual(t, RuntimeNone, runtime)
				t.Logf("Got runtime: %s", runtime)
			}
		})
	}
}
