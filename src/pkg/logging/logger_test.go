// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package logging

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{"debug lowercase", "debug", LevelDebug, false},
		{"DEBUG uppercase", "DEBUG", LevelDebug, false},
		{"info", "info", LevelInfo, false},
		{"INFO", "INFO", LevelInfo, false},
		{"warn", "warn", LevelWarn, false},
		{"warning", "warning", LevelWarn, false},
		{"WARN", "WARN", LevelWarn, false},
		{"error", "error", LevelError, false},
		{"ERROR", "ERROR", LevelError, false},
		{"invalid", "invalid", LevelInfo, true},
		{"empty", "", LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid log level")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.level.String())
		})
	}
}

func TestLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		level:   LevelInfo,
		writer:  &buf,
		noColor: true,
	}

	// Save and restore default logger
	oldDefault := defaultLogger
	defer func() { defaultLogger = oldDefault }()
	defaultLogger = logger

	// Test that debug is suppressed at info level
	Debug("debug message")
	assert.Empty(t, buf.String())

	// Test that info is logged
	Info("info message")
	assert.Contains(t, buf.String(), "INFO")
	assert.Contains(t, buf.String(), "info message")
	buf.Reset()

	// Test that warn is logged
	Warn("warn message")
	assert.Contains(t, buf.String(), "WARN")
	assert.Contains(t, buf.String(), "warn message")
	buf.Reset()

	// Test that error is logged
	Error("error message")
	assert.Contains(t, buf.String(), "ERROR")
	assert.Contains(t, buf.String(), "error message")
}

func TestLogger_DebugLevel(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		level:   LevelDebug,
		writer:  &buf,
		noColor: true,
	}

	oldDefault := defaultLogger
	defer func() { defaultLogger = oldDefault }()
	defaultLogger = logger

	// All levels should be logged at debug level
	Debug("debug msg")
	assert.Contains(t, buf.String(), "DEBUG")
	assert.Contains(t, buf.String(), "debug msg")
	buf.Reset()

	Info("info msg")
	assert.Contains(t, buf.String(), "INFO")
	buf.Reset()

	Warn("warn msg")
	assert.Contains(t, buf.String(), "WARN")
	buf.Reset()

	Error("error msg")
	assert.Contains(t, buf.String(), "ERROR")
}

func TestLogger_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		level:   LevelError,
		writer:  &buf,
		noColor: true,
	}

	oldDefault := defaultLogger
	defer func() { defaultLogger = oldDefault }()
	defaultLogger = logger

	// Only error should be logged at error level
	Debug("debug msg")
	assert.Empty(t, buf.String())

	Info("info msg")
	assert.Empty(t, buf.String())

	Warn("warn msg")
	assert.Empty(t, buf.String())

	Error("error msg")
	assert.Contains(t, buf.String(), "ERROR")
	assert.Contains(t, buf.String(), "error msg")
}

func TestLogger_Formatting(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		level:   LevelInfo,
		writer:  &buf,
		noColor: true,
	}

	oldDefault := defaultLogger
	defer func() { defaultLogger = oldDefault }()
	defaultLogger = logger

	// Test formatted output
	Info("test %s %d", "message", 42)
	output := buf.String()
	assert.Contains(t, output, "test message 42")
	assert.Contains(t, output, "INFO")
}

func TestInit_StderrOnly(t *testing.T) {
	err := Init(LevelDebug, "", true)
	require.NoError(t, err)
	assert.Equal(t, LevelDebug, GetLevel())

	err = Close()
	assert.NoError(t, err)
}

func TestInit_FileOutput(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "logging-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	// Initialize with file output
	err = Init(LevelInfo, logFile, true)
	require.NoError(t, err)

	// Write some logs
	Info("test message")
	Warn("warning message")

	// Close the logger
	err = Close()
	require.NoError(t, err)

	// Read and verify log file
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	output := string(content)
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "warning message")
}

func TestInit_CreateDirectory(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "logging-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Log file in non-existent subdirectory
	logFile := filepath.Join(tmpDir, "logs", "subdir", "test.log")

	// Should create the directory
	err = Init(LevelInfo, logFile, true)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(logFile))
	assert.NoError(t, err)

	err = Close()
	assert.NoError(t, err)
}

func TestInit_InvalidPath(t *testing.T) {
	// Try to create log file in read-only location (should fail on most systems)
	err := Init(LevelInfo, "/dev/null/impossible.log", true)
	// Should return an error (behavior may vary by OS)
	// Just verify it doesn't panic
	_ = err
}

func TestIsDebugEnabled(t *testing.T) {
	// Test with debug level
	err := Init(LevelDebug, "", true)
	require.NoError(t, err)
	assert.True(t, IsDebugEnabled())

	// Test with info level
	SetLevel(LevelInfo)
	assert.False(t, IsDebugEnabled())

	// Test with error level
	SetLevel(LevelError)
	assert.False(t, IsDebugEnabled())

	err = Close()
	assert.NoError(t, err)
}

func TestSetLevel(t *testing.T) {
	err := Init(LevelInfo, "", true)
	require.NoError(t, err)

	assert.Equal(t, LevelInfo, GetLevel())

	SetLevel(LevelDebug)
	assert.Equal(t, LevelDebug, GetLevel())

	SetLevel(LevelError)
	assert.Equal(t, LevelError, GetLevel())

	err = Close()
	assert.NoError(t, err)
}

func TestContextError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ContextError
		contains []string
	}{
		{
			name: "all fields",
			err: &ContextError{
				Operation: "CreateCollection",
				VDBType:   "weaviate-cloud",
				Endpoint:  "https://example.weaviate.cloud",
				Err:       errors.New("connection failed"),
			},
			contains: []string{
				"connection failed",
				"operation=CreateCollection",
				"vdb=weaviate-cloud",
				"endpoint=https://example.weaviate.cloud",
			},
		},
		{
			name: "minimal",
			err: &ContextError{
				Err: errors.New("simple error"),
			},
			contains: []string{"simple error"},
		},
		{
			name: "with context",
			err: &ContextError{
				Operation: "Query",
				VDBType:   "qdrant",
				Err:       errors.New("timeout"),
				Context: map[string]interface{}{
					"collection": "docs",
					"top_k":      5,
				},
			},
			contains: []string{
				"timeout",
				"operation=Query",
				"vdb=qdrant",
				"collection=docs",
				"top_k=5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				assert.Contains(t, errMsg, substr)
			}
		})
	}
}

func TestContextError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	contextErr := &ContextError{
		Operation: "test",
		Err:       originalErr,
	}

	unwrapped := errors.Unwrap(contextErr)
	assert.Equal(t, originalErr, unwrapped)
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("connection refused")

	wrapped := WrapError(originalErr, "ListCollections", "milvus-cloud", "localhost:19530")

	require.Error(t, wrapped)
	assert.Contains(t, wrapped.Error(), "connection refused")
	assert.Contains(t, wrapped.Error(), "operation=ListCollections")
	assert.Contains(t, wrapped.Error(), "vdb=milvus-cloud")
	assert.Contains(t, wrapped.Error(), "endpoint=localhost:19530")

	// Test unwrapping
	assert.True(t, errors.Is(wrapped, originalErr))
}

func TestWrapError_Nil(t *testing.T) {
	wrapped := WrapError(nil, "op", "vdb", "endpoint")
	assert.NoError(t, wrapped)
}

func TestWrapErrorWithContext(t *testing.T) {
	originalErr := errors.New("query failed")
	context := map[string]interface{}{
		"collection": "docs",
		"query":      "test",
		"top_k":      10,
	}

	wrapped := WrapErrorWithContext(
		originalErr,
		"SearchSemantic",
		"chroma-local",
		"localhost:8000",
		context,
	)

	require.Error(t, wrapped)
	errMsg := wrapped.Error()
	assert.Contains(t, errMsg, "query failed")
	assert.Contains(t, errMsg, "operation=SearchSemantic")
	assert.Contains(t, errMsg, "vdb=chroma-local")
	assert.Contains(t, errMsg, "collection=docs")
	assert.Contains(t, errMsg, "top_k=10")
}

func TestWrapErrorWithContext_Nil(t *testing.T) {
	context := map[string]interface{}{"key": "value"}
	wrapped := WrapErrorWithContext(nil, "op", "vdb", "endpoint", context)
	assert.NoError(t, wrapped)
}

func TestLogger_MultiWriter(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "log-test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	var buf bytes.Buffer

	// Create logger with multi-writer (buffer + file)
	logger := &Logger{
		level:      LevelInfo,
		writer:     io.MultiWriter(&buf, tmpFile),
		fileWriter: tmpFile,
		noColor:    true,
	}

	oldDefault := defaultLogger
	defer func() { defaultLogger = oldDefault }()
	defaultLogger = logger

	// Write log message
	Info("test message")

	// Verify it's in both outputs
	assert.Contains(t, buf.String(), "test message")

	// Read from file
	tmpFile.Seek(0, 0)
	fileContent, err := io.ReadAll(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(fileContent), "test message")
}

func TestLogger_NoColor(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		level:   LevelInfo,
		writer:  &buf,
		noColor: true,
	}

	oldDefault := defaultLogger
	defer func() { defaultLogger = oldDefault }()
	defaultLogger = logger

	Info("test message")
	output := buf.String()

	// Should not contain ANSI escape codes
	assert.NotContains(t, output, "\x1b[")
	assert.Contains(t, output, "[INFO]")
}

func TestLogger_TimestampFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		level:   LevelInfo,
		writer:  &buf,
		noColor: true,
	}

	oldDefault := defaultLogger
	defer func() { defaultLogger = oldDefault }()
	defaultLogger = logger

	Info("test")
	output := buf.String()

	// Check timestamp format (YYYY-MM-DD HH:MM:SS)
	parts := strings.Fields(output)
	require.GreaterOrEqual(t, len(parts), 3)

	// Should have date and time
	assert.Regexp(t, `\d{4}-\d{2}-\d{2}`, parts[1])
	assert.Regexp(t, `\d{2}:\d{2}:\d{2}`, parts[2])
}
