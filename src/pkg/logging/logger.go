// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Level represents logging levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a log level string
func ParseLevel(level string) (Level, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN", "WARNING":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", level)
	}
}

// Logger provides structured logging with file output support
type Logger struct {
	level      Level
	writer     io.Writer
	fileWriter io.WriteCloser
	noColor    bool
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{
		level:   LevelInfo,
		writer:  os.Stderr,
		noColor: false,
	}
}

// Init initializes the global logger with options
func Init(level Level, logFile string, noColor bool) error {
	logger := &Logger{
		level:   level,
		writer:  os.Stderr,
		noColor: noColor,
	}

	// Set up file logging if specified
	if logFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logFile)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		logger.fileWriter = file
		// Write to both stderr and file
		logger.writer = io.MultiWriter(os.Stderr, file)
	}

	defaultLogger = logger
	return nil
}

// Close closes the logger and any open file handles
func Close() error {
	if defaultLogger.fileWriter != nil {
		return defaultLogger.fileWriter.Close()
	}
	return nil
}

// SetLevel sets the log level for the default logger
func SetLevel(level Level) {
	defaultLogger.level = level
}

// GetLevel returns the current log level
func GetLevel() Level {
	return defaultLogger.level
}

// log writes a log message with the given level
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	// Format: [LEVEL] timestamp message
	var levelStr string
	if l.noColor {
		levelStr = fmt.Sprintf("[%s]", level.String())
	} else {
		switch level {
		case LevelDebug:
			levelStr = color.New(color.FgCyan).Sprintf("[%s]", level.String())
		case LevelInfo:
			levelStr = color.New(color.FgGreen).Sprintf("[%s]", level.String())
		case LevelWarn:
			levelStr = color.New(color.FgYellow).Sprintf("[%s]", level.String())
		case LevelError:
			levelStr = color.New(color.FgRed).Sprintf("[%s]", level.String())
		}
	}

	fmt.Fprintf(l.writer, "%s %s %s\n", levelStr, timestamp, message)
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	defaultLogger.log(LevelDebug, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	defaultLogger.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	defaultLogger.log(LevelWarn, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	defaultLogger.log(LevelError, format, args...)
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return defaultLogger.level <= LevelDebug
}

// ContextError creates an error with rich context for debugging
type ContextError struct {
	Operation string
	VDBType   string
	Endpoint  string
	Context   map[string]interface{}
	Err       error
}

// Error implements the error interface
func (e *ContextError) Error() string {
	var parts []string

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation=%s", e.Operation))
	}
	if e.VDBType != "" {
		parts = append(parts, fmt.Sprintf("vdb=%s", e.VDBType))
	}
	if e.Endpoint != "" {
		parts = append(parts, fmt.Sprintf("endpoint=%s", e.Endpoint))
	}

	// Add additional context
	for k, v := range e.Context {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	contextStr := strings.Join(parts, " ")
	if contextStr != "" {
		return fmt.Sprintf("%s [%s]", e.Err.Error(), contextStr)
	}
	return e.Err.Error()
}

// Unwrap implements error unwrapping
func (e *ContextError) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with VDB operation context
func WrapError(err error, operation, vdbType, endpoint string) error {
	if err == nil {
		return nil
	}

	return &ContextError{
		Operation: operation,
		VDBType:   vdbType,
		Endpoint:  endpoint,
		Err:       err,
		Context:   make(map[string]interface{}),
	}
}

// WrapErrorWithContext wraps an error with additional context
func WrapErrorWithContext(err error, operation, vdbType, endpoint string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	return &ContextError{
		Operation: operation,
		VDBType:   vdbType,
		Endpoint:  endpoint,
		Err:       err,
		Context:   context,
	}
}
