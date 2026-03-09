// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// IngestCheckpoint stores ingestion progress for resumability
type IngestCheckpoint struct {
	Collection     string       `json:"collection"`
	StartTime      time.Time    `json:"start_time"`
	LastUpdate     time.Time    `json:"last_update"`
	CompletedFiles []string     `json:"completed_files"`
	FailedFiles    []FailedFile `json:"failed_files"`
	PendingFiles   []string     `json:"pending_files"`
}

// LoadCheckpoint loads a checkpoint from disk
func LoadCheckpoint(path string) (*IngestCheckpoint, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // No checkpoint exists yet
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint: %w", err)
	}

	// Parse JSON
	var checkpoint IngestCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// SaveCheckpoint saves a checkpoint to disk
func SaveCheckpoint(checkpoint *IngestCheckpoint, path string) error {
	// Update last update time
	checkpoint.LastUpdate = time.Now()

	// Marshal to JSON (pretty print)
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Write to disk
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}

	return nil
}

// FilterFilesFromCheckpoint filters file list based on checkpoint state
func FilterFilesFromCheckpoint(files []string, checkpoint *IngestCheckpoint, resumeFrom string) ([]string, int, error) {
	if checkpoint == nil && resumeFrom == "" {
		return files, 0, nil // No filtering needed
	}

	completedMap := make(map[string]bool)
	if checkpoint != nil {
		for _, f := range checkpoint.CompletedFiles {
			completedMap[f] = true
		}
	}

	var filtered []string
	skippedCount := 0
	startProcessing := (resumeFrom == "")

	for _, file := range files {
		// Check if already completed (from checkpoint)
		if completedMap[file] {
			skippedCount++
			continue
		}

		// Handle --resume-from flag
		if !startProcessing {
			// Check if this is the resume file (match basename)
			if filepath.Base(file) == resumeFrom || file == resumeFrom {
				startProcessing = true
			} else {
				skippedCount++
				continue
			}
		}

		filtered = append(filtered, file)
	}

	if resumeFrom != "" && !startProcessing {
		return nil, 0, fmt.Errorf("resume file not found: %s", resumeFrom)
	}

	return filtered, skippedCount, nil
}

// InitCheckpoint creates a new checkpoint from file list
func InitCheckpoint(collection string, files []string, existingCheckpoint *IngestCheckpoint) *IngestCheckpoint {
	checkpoint := &IngestCheckpoint{
		Collection:     collection,
		StartTime:      time.Now(),
		LastUpdate:     time.Now(),
		CompletedFiles: []string{},
		FailedFiles:    []FailedFile{},
		PendingFiles:   files,
	}

	// Merge with existing checkpoint if available
	if existingCheckpoint != nil {
		checkpoint.StartTime = existingCheckpoint.StartTime
		checkpoint.CompletedFiles = existingCheckpoint.CompletedFiles
		checkpoint.FailedFiles = existingCheckpoint.FailedFiles
	}

	return checkpoint
}

// MarkFileCompleted marks a file as completed in the checkpoint
func (c *IngestCheckpoint) MarkFileCompleted(file string) {
	c.CompletedFiles = append(c.CompletedFiles, file)
	c.PendingFiles = removeFromSlice(c.PendingFiles, file)
}

// MarkFileFailed marks a file as failed in the checkpoint
func (c *IngestCheckpoint) MarkFileFailed(file string, attempts int, errorMsg string) {
	c.FailedFiles = append(c.FailedFiles, FailedFile{
		File:     file,
		Attempts: attempts,
		Error:    errorMsg,
	})
	c.PendingFiles = removeFromSlice(c.PendingFiles, file)
}

// GetFailedFile gets failed file info if exists
func (c *IngestCheckpoint) GetFailedFile(file string) *FailedFile {
	for i := range c.FailedFiles {
		if c.FailedFiles[i].File == file {
			return &c.FailedFiles[i]
		}
	}
	return nil
}

// removeFromSlice removes a string from a slice
func removeFromSlice(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
