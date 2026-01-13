// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestIsCloudEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "Zilliz cloud endpoint",
			address:  "https://cloud.zilliz.com:19530",
			expected: true,
		},
		{
			name:     "Zillizcloud.com endpoint",
			address:  "https://xyz.zillizcloud.com:443",
			expected: true,
		},
		{
			name:     "Serverless endpoint",
			address:  "https://serverless.zilliz.com",
			expected: true,
		},
		{
			name:     "Local endpoint",
			address:  "localhost:19530",
			expected: false,
		},
		{
			name:     "Local IP",
			address:  "127.0.0.1:19530",
			expected: false,
		},
		{
			name:     "Private IP",
			address:  "192.168.1.100:19530",
			expected: false,
		},
		{
			name:     "Case insensitive cloud",
			address:  "HTTPS://CLOUD.ZILLIZ.COM:19530",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCloudEndpoint(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_GetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			name:     "Default timeout (0)",
			timeout:  0,
			expected: 10 * time.Second,
		},
		{
			name:     "Custom timeout",
			timeout:  30,
			expected: 30 * time.Second,
		},
		{
			name:     "Small timeout",
			timeout:  5,
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					Timeout: tt.timeout,
				},
			}
			result := client.getTimeout()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_GetTimeoutFor(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		baseTimeout   int
		operationType vectordb.OperationType
		expected      time.Duration
	}{
		{
			name:          "Local collection operation with default timeout",
			address:       "localhost:19530",
			baseTimeout:   0, // Use smart defaults
			operationType: vectordb.OperationTypeCollection,
			expected:      20 * time.Second,
		},
		{
			name:          "Cloud collection operation with default timeout",
			address:       "https://cloud.zilliz.com:19530",
			baseTimeout:   0, // Use smart defaults
			operationType: vectordb.OperationTypeCollection,
			expected:      40 * time.Second,
		},
		{
			name:          "Local query operation with default timeout",
			address:       "localhost:19530",
			baseTimeout:   0, // Use smart defaults
			operationType: vectordb.OperationTypeQuery,
			expected:      20 * time.Second,
		},
		{
			name:          "Cloud query operation with default timeout",
			address:       "https://cloud.zilliz.com:19530",
			baseTimeout:   0, // Use smart defaults
			operationType: vectordb.OperationTypeQuery,
			expected:      40 * time.Second,
		},
		{
			name:          "Custom timeout overrides defaults",
			address:       "localhost:19530",
			baseTimeout:   30,
			operationType: vectordb.OperationTypeQuery,
			expected:      30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					Address: tt.address,
					Timeout: tt.baseTimeout,
				},
			}
			result := client.getTimeoutFor(tt.operationType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	t.Skip("Requires Milvus connection - needs mock-based approach")

	// Test that NewClient applies defaults correctly
	config := &Config{
		Address: "localhost:19530",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Skip("Skipping test that requires Milvus connection")
	}
	defer client.Close()

	// Verify defaults were applied
	assert.Equal(t, "default", client.config.Database)
	assert.Equal(t, 1536, client.config.VectorDimensions)
	assert.Equal(t, "L2", client.config.SimilarityMetric)
	assert.Equal(t, 10, client.config.Timeout)
}

func TestClient_GetMetricType(t *testing.T) {
	tests := []struct {
		name             string
		similarityMetric string
		expectedStr      string
	}{
		{
			name:             "L2 metric",
			similarityMetric: "L2",
			expectedStr:      "L2",
		},
		{
			name:             "IP metric",
			similarityMetric: "IP",
			expectedStr:      "IP",
		},
		{
			name:             "COSINE metric",
			similarityMetric: "COSINE",
			expectedStr:      "COSINE",
		},
		{
			name:             "Empty defaults to L2",
			similarityMetric: "",
			expectedStr:      "L2",
		},
		{
			name:             "Unknown defaults to L2",
			similarityMetric: "unknown",
			expectedStr:      "L2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					SimilarityMetric: tt.similarityMetric,
				},
			}
			result := client.getMetricType()
			// Verify the metric type is correct (check the string value)
			assert.Equal(t, tt.expectedStr, string(result))
		})
	}
}

func TestClient_ToMilvusDocument(t *testing.T) {
	client := &Client{
		config: &Config{},
	}

	doc := &vectordb.Document{
		ID:        "test-123",
		Text:      "Test document",
		Content:   "This is test content",
		Image:     "image-id",
		ImageData: "base64-data",
		URL:       "https://example.com",
		Metadata: map[string]interface{}{
			"category": "test",
			"count":    42,
		},
	}

	milvusDoc := client.toMilvusDocument(doc)

	assert.Equal(t, doc.ID, milvusDoc.DocumentID)
	assert.Equal(t, doc.Text, milvusDoc.Text)
	assert.Equal(t, doc.Content, milvusDoc.Content)
	assert.Equal(t, doc.Image, milvusDoc.Image)
	assert.Equal(t, doc.ImageData, milvusDoc.ImageData)
	assert.Equal(t, doc.URL, milvusDoc.URL)
	assert.Equal(t, doc.Metadata, milvusDoc.Metadata)
	assert.Nil(t, milvusDoc.Embedding) // Embeddings populated separately
	assert.Greater(t, milvusDoc.CreatedAt, int64(0))
	assert.Greater(t, milvusDoc.UpdatedAt, int64(0))
}

func TestClient_FromMilvusDocument(t *testing.T) {
	client := &Client{
		config: &Config{},
	}

	metadata := map[string]interface{}{
		"category": "test",
		"count":    42,
	}

	doc := client.fromMilvusDocument(
		"test-123",
		"Test text",
		"Test content",
		"image-id",
		"base64-data",
		"https://example.com",
		metadata,
	)

	assert.Equal(t, "test-123", doc.ID)
	assert.Equal(t, "Test text", doc.Text)
	assert.Equal(t, "Test content", doc.Content)
	assert.Equal(t, "image-id", doc.Image)
	assert.Equal(t, "base64-data", doc.ImageData)
	assert.Equal(t, "https://example.com", doc.URL)
	assert.Equal(t, metadata, doc.Metadata)
}

func TestMustMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Nil value returns empty JSON",
			input:    nil,
			expected: "{}",
		},
		{
			name:     "Simple map",
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "Nested structure",
			input:    map[string]interface{}{"outer": map[string]interface{}{"inner": "value"}},
			expected: `{"outer":{"inner":"value"}}`,
		},
		{
			name:     "Empty map",
			input:    map[string]interface{}{},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mustMarshalJSON(tt.input)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestMustUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected map[string]interface{}
	}{
		{
			name:     "Empty byte slice returns empty map",
			input:    []byte{},
			expected: map[string]interface{}{},
		},
		{
			name:     "Simple JSON",
			input:    []byte(`{"key":"value"}`),
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "Nested JSON",
			input:    []byte(`{"outer":{"inner":"value"}}`),
			expected: map[string]interface{}{"outer": map[string]interface{}{"inner": "value"}},
		},
		{
			name:     "Invalid JSON returns empty map",
			input:    []byte(`{invalid json}`),
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mustUnmarshalJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCombineResultsRRF(t *testing.T) {
	tests := []struct {
		name          string
		vectorResults []*vectordb.QueryResult
		bm25Results   []*vectordb.QueryResult
		topK          int
		expected      int // Number of results expected
	}{
		{
			name: "Combine vector and BM25 results",
			vectorResults: []*vectordb.QueryResult{
				{Document: vectordb.Document{ID: "doc1"}, Score: 0.9},
				{Document: vectordb.Document{ID: "doc2"}, Score: 0.8},
			},
			bm25Results: []*vectordb.QueryResult{
				{Document: vectordb.Document{ID: "doc2"}, Score: 0.7},
				{Document: vectordb.Document{ID: "doc3"}, Score: 0.6},
			},
			topK:     5,
			expected: 3, // doc1, doc2, doc3
		},
		{
			name: "Limit results to topK",
			vectorResults: []*vectordb.QueryResult{
				{Document: vectordb.Document{ID: "doc1"}, Score: 0.9},
				{Document: vectordb.Document{ID: "doc2"}, Score: 0.8},
				{Document: vectordb.Document{ID: "doc3"}, Score: 0.7},
			},
			bm25Results: []*vectordb.QueryResult{
				{Document: vectordb.Document{ID: "doc4"}, Score: 0.6},
			},
			topK:     2,
			expected: 2, // Only top 2 results
		},
		{
			name:          "Empty results",
			vectorResults: []*vectordb.QueryResult{},
			bm25Results:   []*vectordb.QueryResult{},
			topK:          5,
			expected:      0,
		},
		{
			name: "Only vector results",
			vectorResults: []*vectordb.QueryResult{
				{Document: vectordb.Document{ID: "doc1"}, Score: 0.9},
			},
			bm25Results: []*vectordb.QueryResult{},
			topK:        5,
			expected:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combineResultsRRF(tt.vectorResults, tt.bm25Results, tt.topK)
			assert.Len(t, result, tt.expected)

			// Verify scores are in descending order
			for i := 1; i < len(result); i++ {
				assert.GreaterOrEqual(t, result[i-1].Score, result[i].Score,
					"Results should be sorted by score in descending order")
			}
		})
	}
}
