// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package logging

import (
	"encoding/json"
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

// Format represents the log output format
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Logger provides structured logging with file output support
type Logger struct {
	level      Level
	format     Format
	writer     io.Writer
	fileWriter io.WriteCloser
	noColor    bool
	fields     map[string]interface{} // Structured fields for context
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{
		level:   LevelInfo,
		format:  FormatText,
		writer:  os.Stderr,
		noColor: false,
		fields:  make(map[string]interface{}),
	}
}

// Init initializes the global logger with options
func Init(level Level, logFile string, noColor bool) error {
	return InitWithFormat(level, FormatText, logFile, noColor)
}

// InitWithFormat initializes the global logger with format option
func InitWithFormat(level Level, format Format, logFile string, noColor bool) error {
	logger := &Logger{
		level:   level,
		format:  format,
		writer:  os.Stderr,
		noColor: noColor,
		fields:  make(map[string]interface{}),
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

	message := fmt.Sprintf(format, args...)

	if l.format == FormatJSON {
		l.logJSON(level, message)
	} else {
		l.logText(level, message)
	}
}

// logJSON writes a log message in JSON format
func (l *Logger) logJSON(level Level, message string) {
	entry := map[string]interface{}{
		"level":   strings.ToLower(level.String()),
		"time":    time.Now().Format(time.RFC3339),
		"message": message,
	}

	// Add any structured fields
	for k, v := range l.fields {
		entry[k] = v
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to text format on error
		l.logText(level, message)
		return
	}

	fmt.Fprintln(l.writer, string(data))
}

// logText writes a log message in text format
func (l *Logger) logText(level Level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

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

// WithFields creates a logger with structured fields
func WithFields(fields map[string]interface{}) *Logger {
	logger := &Logger{
		level:   defaultLogger.level,
		format:  defaultLogger.format,
		writer:  defaultLogger.writer,
		noColor: defaultLogger.noColor,
		fields:  make(map[string]interface{}),
	}

	// Copy default fields
	for k, v := range defaultLogger.fields {
		logger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		logger.fields[k] = v
	}

	return logger
}

// WithVDB creates a logger with VDB context
func WithVDB(vdbType, operation string) *Logger {
	return WithFields(map[string]interface{}{
		"vdb_type":  vdbType,
		"operation": operation,
	})
}

// WithCollection creates a logger with collection context
func WithCollection(vdbType, collection string) *Logger {
	return WithFields(map[string]interface{}{
		"vdb_type":   vdbType,
		"collection": collection,
	})
}

// WithDocument creates a logger with document context
func WithDocument(vdbType, collection, docID string) *Logger {
	return WithFields(map[string]interface{}{
		"vdb_type":    vdbType,
		"collection":  collection,
		"document_id": docID,
	})
}

// InfoWithFields logs an info message with this logger's fields
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// DebugWithFields logs a debug message with this logger's fields
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// WarnWithFields logs a warning message with this logger's fields
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// ErrorWithFields logs an error message with this logger's fields
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}
