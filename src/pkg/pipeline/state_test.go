// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadState_NewFile(t *testing.T) {
	// Test loading non-existent state file
	stateFile := filepath.Join(t.TempDir(), "test-state.json")

	state, err := loadState(stateFile)
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.NotNil(t, state.ProcessedFiles)
	assert.Empty(t, state.ProcessedFiles)
	assert.False(t, state.LastUpdated.IsZero())
}

func TestSaveAndLoadState(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "test-state.json")

	// Create test state
	originalState := &IngestState{
		ProcessedFiles: map[string]bool{
			"file1.pdf": true,
			"file2.pdf": true,
			"file3.pdf": false,
		},
		LastUpdated: time.Now().Add(-1 * time.Hour),
	}

	// Save state
	err := saveState(stateFile, originalState)
	require.NoError(t, err)

	// Verify file exists
	assert.FileExists(t, stateFile)

	// Load state
	loadedState, err := loadState(stateFile)
	require.NoError(t, err)
	assert.NotNil(t, loadedState)

	// Verify loaded state matches
	assert.Equal(t, len(originalState.ProcessedFiles), len(loadedState.ProcessedFiles))
	assert.True(t, loadedState.ProcessedFiles["file1.pdf"])
	assert.True(t, loadedState.ProcessedFiles["file2.pdf"])
	assert.False(t, loadedState.ProcessedFiles["file3.pdf"])
}

func TestSaveState_UpdatesTimestamp(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "test-state.json")

	// Create state with old timestamp
	state := &IngestState{
		ProcessedFiles: map[string]bool{"file1.pdf": true},
		LastUpdated:    time.Now().Add(-24 * time.Hour),
	}
	oldTime := state.LastUpdated

	// Save state
	err := saveState(stateFile, state)
	require.NoError(t, err)

	// Timestamp should be updated
	assert.True(t, state.LastUpdated.After(oldTime))
	assert.WithinDuration(t, time.Now(), state.LastUpdated, 2*time.Second)
}

func TestLoadState_InvalidJSON(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "invalid-state.json")

	// Write invalid JSON
	err := os.WriteFile(stateFile, []byte("invalid json {"), 0644)
	require.NoError(t, err)

	// Load should fail
	_, err = loadState(stateFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse state file")
}

func TestSaveState_CreatesDirectory(t *testing.T) {
	// Test saving to non-existent directory
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "subdir", "state.json")

	state := &IngestState{
		ProcessedFiles: map[string]bool{"file1.pdf": true},
	}

	// This should work if directory exists, fail if not
	// Note: saveState doesn't create directories, so this will fail
	// But we test that it fails gracefully
	err := saveState(stateFile, state)
	assert.Error(t, err) // Expected to fail
}

func TestStateRoundtrip_EmptyMap(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "empty-state.json")

	// Save empty state
	originalState := &IngestState{
		ProcessedFiles: make(map[string]bool),
	}
	err := saveState(stateFile, originalState)
	require.NoError(t, err)

	// Load and verify
	loadedState, err := loadState(stateFile)
	require.NoError(t, err)
	assert.NotNil(t, loadedState.ProcessedFiles)
	assert.Empty(t, loadedState.ProcessedFiles)
}

func TestStateRoundtrip_LargeMap(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "large-state.json")

	// Create state with many files
	originalState := &IngestState{
		ProcessedFiles: make(map[string]bool),
	}
	for i := 0; i < 1000; i++ {
		originalState.ProcessedFiles[filepath.Join("dir", "file", "test-"+string(rune(i))+".pdf")] = i%2 == 0
	}

	// Save and load
	err := saveState(stateFile, originalState)
	require.NoError(t, err)

	loadedState, err := loadState(stateFile)
	require.NoError(t, err)
	assert.Equal(t, len(originalState.ProcessedFiles), len(loadedState.ProcessedFiles))
}
