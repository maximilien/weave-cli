package weaviate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WeaveClient wraps the official Weaviate client with additional functionality
type WeaveClient struct {
	*Client
	httpClient *http.Client
	config     *Config
}

// NewWeaveClient creates a new Weave client with enhanced functionality
func NewWeaveClient(config *Config) (*WeaveClient, error) {
	// Create the official client first
	officialClient, err := NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create official client: %w", err)
	}

	// Create HTTP client for direct REST API calls
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &WeaveClient{
		Client:     officialClient,
		httpClient: httpClient,
		config:     config,
	}, nil
}

// The following methods delegate to the official client for operations that work correctly
// We only override the delete operations that are broken in the official client

// ListCollections delegates to the official client
func (wc *WeaveClient) ListCollections(ctx context.Context) ([]string, error) {
	return wc.Client.ListCollections(ctx)
}

// ListDocuments delegates to the official client
func (wc *WeaveClient) ListDocuments(ctx context.Context, collectionName string, limit int) ([]Document, error) {
	return wc.Client.ListDocuments(ctx, collectionName, limit)
}

// CountDocuments delegates to the official client
func (wc *WeaveClient) CountDocuments(ctx context.Context, collectionName string) (int, error) {
	return wc.Client.CountDocuments(ctx, collectionName)
}

// GetDocument delegates to the official client
func (wc *WeaveClient) GetDocument(ctx context.Context, collectionName, documentID string) (*Document, error) {
	return wc.Client.GetDocument(ctx, collectionName, documentID)
}

// Health delegates to the official client
func (wc *WeaveClient) Health(ctx context.Context) error {
	return wc.Client.Health(ctx)
}

// GetCollectionSchema delegates to the official client
func (wc *WeaveClient) GetCollectionSchema(ctx context.Context, collectionName string) ([]string, error) {
	return wc.Client.GetCollectionSchema(ctx, collectionName)
}

// DeleteDocument deletes a specific document by ID using REST API
func (wc *WeaveClient) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// First check if the document exists in this collection
	_, err := wc.Client.GetDocument(ctx, collectionName, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete document %s from collection %s: document not found", documentID, collectionName)
	}

	// Construct the REST API URL
	baseURL := strings.TrimSuffix(wc.config.URL, "/")
	// Ensure URL has protocol scheme
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	url := fmt.Sprintf("%s/v1/objects/%s/%s", baseURL, collectionName, documentID)

	// Create the DELETE request
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	// Add authorization header
	if wc.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+wc.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete document %s from collection %s: %w", documentID, collectionName, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}

	// If we get a 404, it means the document was not found
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("failed to delete document %s from collection %s: document not found", documentID, collectionName)
	}

	return fmt.Errorf("failed to delete document %s from collection %s: HTTP %d - %s", documentID, collectionName, resp.StatusCode, string(body))
}

// DeleteDocumentsByMetadata deletes documents matching metadata filters using REST API
func (wc *WeaveClient) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadataFilters []string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Parse metadata filters
	filters := make(map[string]string)
	for _, filter := range metadataFilters {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid metadata filter format: %s (expected key=value)", filter)
		}
		filters[parts[0]] = parts[1]
	}

	// First, query for documents matching the metadata filters
	documents, err := wc.queryDocumentsByMetadata(ctx, collectionName, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to query documents by metadata: %w", err)
	}

	if len(documents) == 0 {
		return 0, nil // No documents found matching the filters
	}

	// Delete each document individually using REST API
	deletedCount := 0
	for _, doc := range documents {
		if err := wc.DeleteDocument(ctx, collectionName, doc.ID); err != nil {
			// Log error but continue with other documents
			fmt.Printf("Warning: Failed to delete document %s: %v\n", doc.ID, err)
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}

// queryDocumentsByMetadata queries for documents matching metadata filters using GraphQL
func (wc *WeaveClient) queryDocumentsByMetadata(ctx context.Context, collectionName string, filters map[string]string) ([]Document, error) {
	// Build the where clause for metadata filtering
	var whereClauses []string
	for key, value := range filters {
		if key == "filename" {
			// For filename, we need to search within the JSON string in the metadata field
			// Use Like operator to search for the filename within the JSON string
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["metadata"]
				operator: Like
				valueString: "*filename\": \"%s\"*"
			}`, value))
		} else if key == "original_filename" {
			// For original_filename, we need to search within the JSON string in the metadata field
			// Use Like operator to search for the original_filename within the JSON string
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["metadata"]
				operator: Like
				valueString: "*original_filename\": \"%s\"*"
			}`, value))
		} else if key == "url" {
			// For URL, use Like operator to allow partial matching
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["%s"]
				operator: Like
				valueString: "*%s*"
			}`, key, value))
		} else {
			// For other fields, use direct path with exact matching
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["%s"]
				operator: Equal
				valueString: "%s"
			}`, key, value))
		}
	}

	// Combine multiple filters with AND
	var whereClause string
	if len(whereClauses) == 1 {
		whereClause = whereClauses[0]
	} else {
		whereClause = fmt.Sprintf(`{
			operator: And
			operands: [%s]
		}`, strings.Join(whereClauses, ", "))
	}

	// Create GraphQL query to get documents
	query := fmt.Sprintf(`
		query {
			Get {
				%s(where: %s) {
					_additional {
						id
					}
				}
			}
		}
	`, collectionName, whereClause)

	// Create GraphQL request payload
	payload := map[string]interface{}{
		"query": query,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL payload: %w", err)
	}

	// Construct the GraphQL endpoint URL
	baseURL := strings.TrimSuffix(wc.config.URL, "/")
	// Ensure URL has protocol scheme
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	url := fmt.Sprintf("%s/v1/graphql", baseURL)

	// Create the POST request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	// Add headers
	if wc.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+wc.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by metadata from collection %s: %w", collectionName, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query documents by metadata from collection %s: HTTP %d - %s", collectionName, resp.StatusCode, string(body))
	}

	// Parse GraphQL response
	var graphqlResp struct {
		Data   map[string]interface{} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", graphqlResp.Errors)
	}

	// Extract documents
	var documents []Document
	if data, ok := graphqlResp.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			for _, item := range collectionData {
				if docMap, ok := item.(map[string]interface{}); ok {
					doc := Document{}

					// Get the document ID from _additional
					if additional, ok := docMap["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							doc.ID = id
						}
					}

					if doc.ID != "" {
						documents = append(documents, doc)
					}
				}
			}
		}
	}

	return documents, nil
}

// GetDocumentsByMetadata gets documents matching metadata filters
func (wc *WeaveClient) GetDocumentsByMetadata(ctx context.Context, collectionName string, metadataFilters []string) ([]Document, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Parse metadata filters
	filters := make(map[string]string)
	for _, filter := range metadataFilters {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid metadata filter format: %s (expected key=value)", filter)
		}
		filters[parts[0]] = parts[1]
	}

	// Query for documents matching the metadata filters
	documents, err := wc.queryDocumentsByMetadata(ctx, collectionName, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by metadata: %w", err)
	}

	// Get full document details for each document
	var fullDocuments []Document
	for _, doc := range documents {
		fullDoc, err := wc.GetDocument(ctx, collectionName, doc.ID)
		if err != nil {
			// Log error but continue with other documents
			fmt.Printf("Warning: Failed to get document %s: %v\n", doc.ID, err)
			continue
		}
		fullDocuments = append(fullDocuments, *fullDoc)
	}

	return fullDocuments, nil
}

// DeleteCollection deletes all objects from a collection
func (wc *WeaveClient) DeleteCollection(ctx context.Context, collectionName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// First try GraphQL deletion
	err := wc.deleteCollectionViaGraphQL(ctx, collectionName)
	if err == nil {
		return nil
	}

	// If GraphQL fails (e.g., mutations disabled), try REST API
	if strings.Contains(err.Error(), "mutations") || strings.Contains(err.Error(), "Schema is not configured") {
		return wc.deleteCollectionViaREST(ctx, collectionName)
	}

	return err
}

// deleteCollectionViaGraphQL deletes all objects using GraphQL
func (wc *WeaveClient) deleteCollectionViaGraphQL(ctx context.Context, collectionName string) error {
	// Create GraphQL mutation to delete all objects in collection
	mutation := fmt.Sprintf(`
		mutation {
			delete {
				%s(where: {
					path: ["id"]
					operator: Like
					valueString: "*"
				}) {
					successful
					failed
				}
			}
		}
	`, collectionName)

	// Create GraphQL request payload
	payload := map[string]interface{}{
		"query": mutation,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL payload: %w", err)
	}

	// Construct the GraphQL endpoint URL
	baseURL := strings.TrimSuffix(wc.config.URL, "/")
	// Ensure URL has protocol scheme
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	url := fmt.Sprintf("%s/v1/graphql", baseURL)

	// Create the POST request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	// Add headers
	if wc.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+wc.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", collectionName, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete collection %s: HTTP %d - %s", collectionName, resp.StatusCode, string(body))
	}

	// Parse GraphQL response
	var graphqlResp struct {
		Data   map[string]interface{} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", graphqlResp.Errors)
	}

	// Check if deletion was successful
	if data, ok := graphqlResp.Data["delete"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].(map[string]interface{}); ok {
			if successful, ok := collectionData["successful"].(float64); ok && successful > 0 {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to delete collection %s: no objects deleted", collectionName)
}

// deleteCollectionViaREST deletes all objects using REST API
func (wc *WeaveClient) deleteCollectionViaREST(ctx context.Context, collectionName string) error {
	// First, get all objects in the collection using GraphQL query (queries are usually allowed)
	objects, err := wc.getAllObjectsInCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to get objects in collection %s: %w", collectionName, err)
	}

	if len(objects) == 0 {
		return nil // Nothing to delete
	}

	// Delete each object individually using REST API
	deletedCount := 0
	for _, obj := range objects {
		if err := wc.deleteObjectViaREST(ctx, obj.ID); err != nil {
			// Log error but continue with other objects
			fmt.Printf("Warning: Failed to delete object %s: %v\n", obj.ID, err)
			continue
		}
		deletedCount++
	}

	if deletedCount == 0 {
		return fmt.Errorf("failed to delete any objects from collection %s", collectionName)
	}

	return nil
}

// getAllObjectsInCollection gets all objects in a collection using GraphQL query
func (wc *WeaveClient) getAllObjectsInCollection(ctx context.Context, collectionName string) ([]ObjectInfo, error) {
	query := fmt.Sprintf(`
		query {
			Get {
				%s {
					_additional {
						id
					}
				}
			}
		}
	`, collectionName)

	// Create GraphQL request payload
	payload := map[string]interface{}{
		"query": query,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL payload: %w", err)
	}

	// Construct the GraphQL endpoint URL
	baseURL := strings.TrimSuffix(wc.config.URL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	url := fmt.Sprintf("%s/v1/graphql", baseURL)

	// Create the POST request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	// Add headers
	if wc.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+wc.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection %s: %w", collectionName, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query collection %s: HTTP %d - %s", collectionName, resp.StatusCode, string(body))
	}

	// Parse GraphQL response
	var graphqlResp struct {
		Data   map[string]interface{} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", graphqlResp.Errors)
	}

	// Extract objects from result
	var objects []ObjectInfo
	if data, ok := graphqlResp.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			for _, item := range collectionData {
				if obj, ok := item.(map[string]interface{}); ok {
					if additional, ok := obj["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							objects = append(objects, ObjectInfo{ID: id})
						}
					}
				}
			}
		}
	}

	return objects, nil
}

// ObjectInfo represents basic object information
type ObjectInfo struct {
	ID string
}

// deleteObjectViaREST deletes a single object using REST API
func (wc *WeaveClient) deleteObjectViaREST(ctx context.Context, objectID string) error {
	// Construct the REST API URL
	baseURL := strings.TrimSuffix(wc.config.URL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	url := fmt.Sprintf("%s/v1/objects/%s", baseURL, objectID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	if wc.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+wc.config.APIKey)
	}

	// Make the request
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete object: HTTP %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateCollection delegates to the official client
func (wc *WeaveClient) CreateCollection(ctx context.Context, collectionName, embeddingModel string, customFields []FieldDefinition) error {
	return wc.Client.CreateCollection(ctx, collectionName, embeddingModel, customFields)
}
