// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_GetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			name:     "Default timeout (0)",
			timeout:  0,
			expected: 30 * time.Second,
		},
		{
			name:     "Custom timeout",
			timeout:  60,
			expected: 60 * time.Second,
		},
		{
			name:     "Small timeout",
			timeout:  10,
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{
				config: &vectordb.Config{
					Timeout: tt.timeout,
				},
			}
			result := adapter.getTimeout()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdapter_GetTimeoutFor(t *testing.T) {
	tests := []struct {
		name          string
		dbType        vectordb.VectorDBType
		baseTimeout   int
		operationType vectordb.OperationType
		expected      time.Duration
	}{
		{
			name:          "Local health operation with default timeout",
			dbType:        vectordb.VectorDBTypeSupabaseLocal,
			baseTimeout:   0,
			operationType: vectordb.OperationTypeHealth,
			expected:      10 * time.Second, // Local default for health
		},
		{
			name:          "Cloud health operation with default timeout",
			dbType:        vectordb.VectorDBTypeSupabase,
			baseTimeout:   0,
			operationType: vectordb.OperationTypeHealth,
			expected:      20 * time.Second, // Cloud default for health
		},
		{
			name:          "Local collection operation with default timeout",
			dbType:        vectordb.VectorDBTypeSupabaseLocal,
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      20 * time.Second, // Local default for collection
		},
		{
			name:          "Cloud collection operation with default timeout",
			dbType:        vectordb.VectorDBTypeSupabaseCloud,
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      40 * time.Second, // Cloud default for collection
		},
		{
			name:          "Custom timeout overrides defaults",
			dbType:        vectordb.VectorDBTypeSupabaseLocal,
			baseTimeout:   45,
			operationType: vectordb.OperationTypeQuery,
			expected:      45 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{
				config: &vectordb.Config{
					Type:    tt.dbType,
					Timeout: tt.baseTimeout,
				},
			}
			result := adapter.getTimeoutFor(tt.operationType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdapter_HasDBConnection(t *testing.T) {
	t.Run("Returns true when db is not nil", func(t *testing.T) {
		adapter := &Adapter{
			db: &sql.DB{}, // Non-nil DB
		}
		assert.True(t, adapter.hasDBConnection())
	})

	t.Run("Returns false when db is nil", func(t *testing.T) {
		adapter := &Adapter{
			db: nil,
		}
		assert.False(t, adapter.hasDBConnection())
	})
}

func TestAdapter_RequireDBConnection(t *testing.T) {
	t.Run("Returns error when db is nil", func(t *testing.T) {
		adapter := &Adapter{
			db: nil,
		}
		err := adapter.requireDBConnection("TestOperation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TestOperation")
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("Returns nil when db is not nil", func(t *testing.T) {
		adapter := &Adapter{
			db: &sql.DB{},
		}
		err := adapter.requireDBConnection("TestOperation")
		assert.NoError(t, err)
	})
}

func TestAdapter_WrapError(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name        string
		err         error
		operation   string
		expectError bool
		errorType   string
	}{
		{
			name:        "Nil error returns nil",
			err:         nil,
			operation:   "test",
			expectError: false,
		},
		{
			name:        "Connection refused error",
			err:         fmt.Errorf("connection refused"),
			operation:   "connect",
			expectError: true,
			errorType:   "connection",
		},
		{
			name:        "Timeout error",
			err:         fmt.Errorf("timeout exceeded"),
			operation:   "query",
			expectError: true,
			errorType:   "connection",
		},
		{
			name:        "Authentication failed error",
			err:         fmt.Errorf("authentication failed"),
			operation:   "auth",
			expectError: true,
			errorType:   "auth",
		},
		{
			name:        "Permission denied error",
			err:         fmt.Errorf("permission denied"),
			operation:   "access",
			expectError: true,
			errorType:   "auth",
		},
		{
			name:        "Not found error",
			err:         fmt.Errorf("not found"),
			operation:   "get",
			expectError: true,
			errorType:   "notfound",
		},
		{
			name:        "Invalid error",
			err:         fmt.Errorf("invalid format"),
			operation:   "validate",
			expectError: true,
			errorType:   "config",
		},
		{
			name:        "Generic error",
			err:         fmt.Errorf("something went wrong"),
			operation:   "generic",
			expectError: true,
			errorType:   "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.wrapError(tt.err, tt.operation)
			if tt.expectError {
				assert.Error(t, result)
			} else {
				assert.NoError(t, result)
			}
		})
	}
}

func TestAdapter_GetTableName(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name           string
		collectionName string
		expected       string
	}{
		{
			name:           "Simple name",
			collectionName: "my_collection",
			expected:       "collection_my_collection",
		},
		{
			name:           "Name with spaces",
			collectionName: "my collection",
			expected:       "collection_my_collection",
		},
		{
			name:           "Name with multiple spaces",
			collectionName: "my test collection",
			expected:       "collection_my_test_collection",
		},
		{
			name:           "Name with special characters",
			collectionName: "MyCollection123",
			expected:       "collection_MyCollection123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.getTableName(tt.collectionName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "Simple identifier",
			identifier: "table_name",
			expected:   `"table_name"`,
		},
		{
			name:       "Identifier with uppercase",
			identifier: "TableName",
			expected:   `"TableName"`,
		},
		{
			name:       "Identifier with spaces",
			identifier: "table name",
			expected:   `"table name"`,
		},
		{
			name:       "Identifier with double quotes",
			identifier: `table"name`,
			expected:   `"table""name"`,
		},
		{
			name:       "Identifier with multiple double quotes",
			identifier: `table"name"test`,
			expected:   `"table""name""test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifier(tt.identifier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdapter_ConvertMetadataToJSON(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name     string
		metadata map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "Nil metadata",
			metadata: nil,
			expected: "{}",
			wantErr:  false,
		},
		{
			name:     "Empty metadata",
			metadata: map[string]interface{}{},
			expected: "{}",
			wantErr:  false,
		},
		{
			name: "Simple metadata",
			metadata: map[string]interface{}{
				"key": "value",
			},
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name: "Complex metadata",
			metadata: map[string]interface{}{
				"string": "value",
				"number": 42,
				"bool":   true,
			},
			wantErr: false, // Just check no error, order may vary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.convertMetadataToJSON(tt.metadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected != "" {
					assert.Equal(t, tt.expected, result)
				}
			}
		})
	}
}

func TestAdapter_ConvertJSONToMetadata(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name     string
		jsonStr  string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "Empty string",
			jsonStr:  "",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "Empty JSON object",
			jsonStr:  "{}",
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:    "Simple JSON",
			jsonStr: `{"key":"value"}`,
			expected: map[string]interface{}{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name:    "Complex JSON",
			jsonStr: `{"string":"value","number":42,"bool":true}`,
			expected: map[string]interface{}{
				"string": "value",
				"number": float64(42), // JSON numbers unmarshal as float64
				"bool":   true,
			},
			wantErr: false,
		},
		{
			name:     "Invalid JSON",
			jsonStr:  `{invalid}`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.convertJSONToMetadata(tt.jsonStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAdapter_ConvertDataTypeToPostgreSQL(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name      string
		dataTypes []string
		expected  string
	}{
		{
			name:      "Empty data types",
			dataTypes: []string{},
			expected:  "",
		},
		{
			name:      "Text type",
			dataTypes: []string{"text"},
			expected:  "TEXT",
		},
		{
			name:      "String type",
			dataTypes: []string{"string"},
			expected:  "TEXT",
		},
		{
			name:      "Int type",
			dataTypes: []string{"int"},
			expected:  "INTEGER",
		},
		{
			name:      "Integer type",
			dataTypes: []string{"integer"},
			expected:  "INTEGER",
		},
		{
			name:      "Float type",
			dataTypes: []string{"float"},
			expected:  "REAL",
		},
		{
			name:      "Number type",
			dataTypes: []string{"number"},
			expected:  "REAL",
		},
		{
			name:      "Boolean type",
			dataTypes: []string{"boolean"},
			expected:  "BOOLEAN",
		},
		{
			name:      "Bool type",
			dataTypes: []string{"bool"},
			expected:  "BOOLEAN",
		},
		{
			name:      "Date type",
			dataTypes: []string{"date"},
			expected:  "TIMESTAMP",
		},
		{
			name:      "Datetime type",
			dataTypes: []string{"datetime"},
			expected:  "TIMESTAMP",
		},
		{
			name:      "UUID type",
			dataTypes: []string{"uuid"},
			expected:  "UUID",
		},
		{
			name:      "JSON type",
			dataTypes: []string{"json"},
			expected:  "JSONB",
		},
		{
			name:      "Object type",
			dataTypes: []string{"object"},
			expected:  "JSONB",
		},
		{
			name:      "Unknown type defaults to TEXT",
			dataTypes: []string{"unknown"},
			expected:  "TEXT",
		},
		{
			name:      "Multiple types uses first",
			dataTypes: []string{"text", "string"},
			expected:  "TEXT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertDataTypeToPostgreSQL(tt.dataTypes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
