// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vectordb

import (
	"fmt"
)

// ErrorType represents the type of vector database error
type ErrorType string

const (
	ErrorTypeConnection     ErrorType = "connection"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeAlreadyExists  ErrorType = "already_exists"
	ErrorTypeInvalidConfig  ErrorType = "invalid_config"
	ErrorTypeInvalidSchema  ErrorType = "invalid_schema"
	ErrorTypeInvalidQuery   ErrorType = "invalid_query"
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeQuotaExceeded  ErrorType = "quota_exceeded"
	ErrorTypeInternal       ErrorType = "internal"
	ErrorTypeUnsupported    ErrorType = "unsupported"
)

// VectorDBError represents a vector database error with additional context
type VectorDBError struct {
	Type       ErrorType
	Message    string
	Cause      error
	Collection string
	Document   string
	Details    map[string]interface{}
}

// Error implements the error interface
func (e *VectorDBError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *VectorDBError) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type
func (e *VectorDBError) Is(target error) bool {
	if targetErr, ok := target.(*VectorDBError); ok {
		return e.Type == targetErr.Type
	}
	return false
}

// NewError creates a new VectorDBError
func NewError(errorType ErrorType, message string) *VectorDBError {
	return &VectorDBError{
		Type:    errorType,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewErrorWithCause creates a new VectorDBError with a cause
func NewErrorWithCause(errorType ErrorType, message string, cause error) *VectorDBError {
	return &VectorDBError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// WithCollection adds collection context to the error
func (e *VectorDBError) WithCollection(collection string) *VectorDBError {
	e.Collection = collection
	return e
}

// WithDocument adds document context to the error
func (e *VectorDBError) WithDocument(document string) *VectorDBError {
	e.Document = document
	return e
}

// WithDetail adds a detail to the error
func (e *VectorDBError) WithDetail(key string, value interface{}) *VectorDBError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Common error constructors

// ErrConnectionFailed creates a connection error
func ErrConnectionFailed(message string, cause error) *VectorDBError {
	return NewErrorWithCause(ErrorTypeConnection, message, cause)
}

// ErrAuthenticationFailed creates an authentication error
func ErrAuthenticationFailed(message string) *VectorDBError {
	return NewError(ErrorTypeAuthentication, message)
}

// ErrNotFound creates a not found error
func ErrNotFound(resource, name string) *VectorDBError {
	return NewError(ErrorTypeNotFound, fmt.Sprintf("%s '%s' not found", resource, name))
}

// ErrAlreadyExists creates an already exists error
func ErrAlreadyExists(resource, name string) *VectorDBError {
	return NewError(ErrorTypeAlreadyExists, fmt.Sprintf("%s '%s' already exists", resource, name))
}

// ErrInvalidConfig creates an invalid configuration error
func ErrInvalidConfig(message string) *VectorDBError {
	return NewError(ErrorTypeInvalidConfig, message)
}

// ErrInvalidSchema creates an invalid schema error
func ErrInvalidSchema(message string) *VectorDBError {
	return NewError(ErrorTypeInvalidSchema, message)
}

// ErrInvalidQuery creates an invalid query error
func ErrInvalidQuery(message string) *VectorDBError {
	return NewError(ErrorTypeInvalidQuery, message)
}

// ErrTimeout creates a timeout error
func ErrTimeout(operation string) *VectorDBError {
	return NewError(ErrorTypeTimeout, fmt.Sprintf("operation '%s' timed out", operation))
}

// ErrQuotaExceeded creates a quota exceeded error
func ErrQuotaExceeded(message string) *VectorDBError {
	return NewError(ErrorTypeQuotaExceeded, message)
}

// ErrInternal creates an internal error
func ErrInternal(message string, cause error) *VectorDBError {
	return NewErrorWithCause(ErrorTypeInternal, message, cause)
}

// ErrUnsupported creates an unsupported operation error
func ErrUnsupported(operation string) *VectorDBError {
	return NewError(ErrorTypeUnsupported, fmt.Sprintf("operation '%s' is not supported", operation))
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	if vdbErr, ok := err.(*VectorDBError); ok {
		return vdbErr.Type == errorType
	}
	return false
}

// IsConnectionError checks if an error is a connection error
func IsConnectionError(err error) bool {
	return IsErrorType(err, ErrorTypeConnection)
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	return IsErrorType(err, ErrorTypeAuthentication)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return IsErrorType(err, ErrorTypeNotFound)
}

// IsAlreadyExistsError checks if an error is an already exists error
func IsAlreadyExistsError(err error) bool {
	return IsErrorType(err, ErrorTypeAlreadyExists)
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	return IsErrorType(err, ErrorTypeTimeout)
}

// IsUnsupportedError checks if an error is an unsupported operation error
func IsUnsupportedError(err error) bool {
	return IsErrorType(err, ErrorTypeUnsupported)
}
