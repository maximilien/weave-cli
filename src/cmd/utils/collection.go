// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

// isImageCollection checks if a collection name suggests it's an image collection
func isImageCollection(collectionName string) bool {
	imageKeywords := []string{"image", "img", "photo", "picture", "visual", "media"}
	name := strings.ToLower(collectionName)
	for _, keyword := range imageKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	return false
}

// isConnectionRefusedError checks if an error is a connection refused error
func isConnectionRefusedError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connect: connection refused")
}

// ParseFieldDefinitions parses field definitions from a string
func ParseFieldDefinitions(fieldsStr string) ([]weaviate.FieldDefinition, error) {
	PrintInfo("Field parsing not yet implemented in new structure")
	return nil, fmt.Errorf("field parsing not yet implemented")
}

// CreateWeaviateCollection creates a Weaviate collection with default schema
func CreateWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition) error {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	err = client.CreateCollectionWithSchema(ctx, collectionName, embeddingModel, customFields, "default")
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err)
	}

	return nil
}

// CreateSupabaseCollection creates a Supabase collection with default schema
func CreateSupabaseCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string) error {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Supabase client: %v", err)
	}

	// Create simple collection schema for Supabase
	// Supabase collections have standard fields: id, content, text, image, image_data, url, metadata, embedding
	err = client.CreateCollection(ctx, collectionName, nil)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err)
	}

	return nil
}

// CreateWeaviateCollectionFromConfigSchema creates a Weaviate collection from a named schema in config.yaml
func CreateWeaviateCollectionFromConfigSchema(ctx context.Context, cfg *config.Config, dbConfig *config.VectorDBConfig, collectionName, schemaName, metadataMode string) error {
	// Get the schema definition from config
	schemaDef, err := cfg.GetSchema(schemaName)
	if err != nil {
		return err
	}

	// Convert schema definition to CollectionSchema
	schema, err := convertSchemaDefinitionToCollectionSchema(schemaDef, collectionName, metadataMode)
	if err != nil {
		return fmt.Errorf("failed to convert schema definition: %v", err)
	}

	// Create the client
	client, err := CreateWeaviateClient(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	// Create collection from schema
	err = client.CreateCollectionFromSchema(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create collection from schema: %v", err)
	}

	return nil
}

// CreateWeaviateCollectionFromSchemaFile creates a Weaviate collection from a schema file
func CreateWeaviateCollectionFromSchemaFile(ctx context.Context, cfg *config.VectorDBConfig, collectionName, schemaFilePath string) error {
	// Load schema from file
	schemaExport, err := LoadSchemaFromYAMLFile(schemaFilePath)
	if err != nil {
		return fmt.Errorf("failed to load schema file: %v", err)
	}

	// Validate that the schema has the required data
	if schemaExport.Schema == nil {
		return fmt.Errorf("schema file does not contain valid schema information")
	}

	// Override the collection name if specified
	if collectionName != "" && collectionName != schemaExport.Name {
		PrintInfo(fmt.Sprintf("Using specified collection name '%s' instead of schema name '%s'", collectionName, schemaExport.Name))
		schemaExport.Schema.Class = collectionName
	}

	// Create the client
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	// Create collection from schema
	err = client.CreateCollectionFromSchema(ctx, schemaExport.Schema)
	if err != nil {
		return fmt.Errorf("failed to create collection from schema: %v", err)
	}

	return nil
}

// CreateGenericCollection creates a collection using the generic VectorDB interface
// Works for any database that implements the VectorDBClient interface (Milvus, etc.)
func CreateGenericCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string) error {
	// Create client using the factory
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get default schema for the database type
	schema := client.GetDefaultSchema(vectordb.SchemaTypeText, collectionName)
	if embeddingModel != "" {
		schema.Vectorizer = embeddingModel
	}

	// Validate schema
	if err := client.ValidateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Create collection
	if err := client.CreateCollection(ctx, collectionName, schema); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

// CreateMockCollection creates a mock collection
func CreateMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition) error {
	PrintInfo("Mock collection creation not yet implemented in new structure")
	return fmt.Errorf("collection creation not yet implemented")
}

// ListWeaviateCollections lists Weaviate collections
func ListWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to create client: %v", err))
		}
		return
	}

	// Add timeout to context for API operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PrintError(fmt.Sprintf("Failed to connect to Weaviate API at %s: connection timeout after 30 seconds", cfg.URL))
			fmt.Println()
			PrintInfo("Please check that:")
			PrintInfo("  1. The Weaviate server is running and accessible")
			PrintInfo("  2. The URL in your config is correct")
			PrintInfo("  3. There are no network/firewall issues")
		} else if isConnectionRefusedError(err) {
			// Connection refused - Weaviate server is not running
			PrintWarning(fmt.Sprintf("Weaviate local instance not available at %s", cfg.URL))
			if !jsonOutput {
				fmt.Println()
				PrintInfo("This usually means:")
				PrintInfo("  • Weaviate local server is not running")
				PrintInfo("  • Wrong port or URL configured")
				fmt.Println()
				PrintInfo("💡 To start Weaviate locally:")
				PrintInfo("   docker run -d -p 8080:8080 \\")
				PrintInfo("     -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true \\")
				PrintInfo("     -e PERSISTENCE_DATA_PATH=/var/lib/weaviate \\")
				PrintInfo("     cr.weaviate.io/semitechnologies/weaviate:latest")
				fmt.Println()
				PrintInfo("Or use a different database with: --weaviate (cloud) or --mock")
			}
		} else {
			// Check if this might be a configuration error
			formattedErr := config.FormatConfigError(err)
			if formattedErr != err.Error() {
				// Error was enhanced with configuration tips
				PrintError(formattedErr)

				// Check if we can prompt for fix
				if configErr := config.CheckRequiredEnvVars(); configErr != nil {
					shouldFix, promptErr := config.PromptToFixConfig(configErr)
					if promptErr == nil && shouldFix {
						if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
							PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
						}
					}
				}
			} else {
				PrintError(fmt.Sprintf("Failed to list collections: %v", err))
			}
		}
		return
	}

	if len(collections) == 0 {
		if !jsonOutput {
			PrintWarning("No collections found in the database")
		} else {
			// Output empty JSON
			fmt.Println(`{"collections": [], "total": 0}`)
		}
		return
	}

	// Sort collections alphabetically
	sort.Strings(collections)

	// Apply limit if specified
	if limit > 0 && len(collections) > limit {
		collections = collections[:limit]
	}

	if !jsonOutput {
		PrintSuccess(fmt.Sprintf("Found %d collections:", len(collections)))
		if limit > 0 && len(collections) == limit {
			fmt.Printf("(showing first %d collections)\n", limit)
		}
		fmt.Println()
	}

	// Get collection information for each collection
	type CollectionInfo struct {
		Name              string
		DocumentCount     int
		HasDocuments      bool
		IsImageCollection bool
		SchemaType        string
	}

	var collectionInfos []CollectionInfo
	for _, collection := range collections {
		count, err := client.CountDocuments(ctx, collection)
		hasDocuments := err == nil && count > 0
		isImageCollection := isImageCollection(collection)

		// Try to get schema type
		schemaType := "unknown"
		if schema, err := client.GetCollectionSchema(ctx, collection); err == nil {
			if len(schema) > 0 {
				schemaType = "text"
				// Check if it's an image collection based on schema
				// Look for specific image-related fields that indicate it's truly an image collection
				hasImageData := false
				hasImageField := false
				for _, field := range schema {
					fieldLower := strings.ToLower(field)
					if fieldLower == "image_data" || fieldLower == "imageData" {
						hasImageData = true
					}
					if fieldLower == "image" {
						hasImageField = true
					}
				}
				// Only consider it an image collection if it has both image_data and image fields
				// This distinguishes between auto-generated image fields and true image collections
				if hasImageData && hasImageField {
					schemaType = "image"
				}
			}
		}

		collectionInfos = append(collectionInfos, CollectionInfo{
			Name:              collection,
			DocumentCount:     count,
			HasDocuments:      hasDocuments,
			IsImageCollection: isImageCollection,
			SchemaType:        schemaType,
		})
	}

	// Output in JSON format if requested
	if jsonOutput {
		type JSONCollection struct {
			Name          string `json:"name"`
			DocumentCount int    `json:"document_count"`
			SchemaType    string `json:"schema_type"`
			IsImage       bool   `json:"is_image"`
		}

		type JSONOutput struct {
			Collections []JSONCollection `json:"collections"`
			Total       int              `json:"total"`
		}

		jsonCollections := make([]JSONCollection, len(collectionInfos))
		for i, info := range collectionInfos {
			jsonCollections[i] = JSONCollection{
				Name:          info.Name,
				DocumentCount: info.DocumentCount,
				SchemaType:    info.SchemaType,
				IsImage:       info.IsImageCollection,
			}
		}

		output := JSONOutput{
			Collections: jsonCollections,
			Total:       len(jsonCollections),
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}

		fmt.Println(string(jsonBytes))
		return
	}

	// Display collections in compact format
	for i, info := range collectionInfos {
		// Color coding: green for collections with documents, yellow for empty collections
		var nameColor string
		if info.HasDocuments {
			nameColor = GetStyledKeyProminent(info.Name)
		} else {
			nameColor = GetStyledKeyDimmed(info.Name)
		}

		// Document count with color
		var countStr string
		if info.HasDocuments {
			countStr = GetStyledNumber(fmt.Sprintf("%d items", info.DocumentCount))
		} else {
			countStr = GetStyledValueDimmed("empty")
		}

		// Schema type indicator
		var typeIndicator string
		switch info.SchemaType {
		case "image":
			typeIndicator = GetStyledEmoji("🖼️")
		case "text":
			typeIndicator = GetStyledEmoji("📄")
		default:
			typeIndicator = GetStyledEmoji("❓")
		}

		// Collection type indicator
		var collectionType string
		if info.IsImageCollection {
			collectionType = GetStyledValueDimmed("(image)")
		}

		// Compact single-line format
		fmt.Printf("%2d. %s %s %s %s\n",
			i+1,
			nameColor,
			countStr,
			typeIndicator,
			collectionType)

		if virtual && info.HasDocuments {
			// Show virtual structure summary for collections with documents
			showCollectionVirtualSummary(ctx, client, info.Name)
		}
	}
}

// ListSupabaseCollections lists Supabase collections
func ListSupabaseCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)
		} else {
			PrintError(fmt.Sprintf("Failed to create Supabase client: %v", err))
		}
		return
	}

	// Add timeout to context for API operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List collections
	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PrintError("Failed to connect to Supabase database: connection timeout after 30 seconds")
			fmt.Println()
			PrintInfo("Please check that:")
			PrintInfo("  1. The Supabase database URL is correct")
			PrintInfo("  2. The database key is valid")
			PrintInfo("  3. There are no network/firewall issues")
		} else {
			PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		}
		return
	}

	if len(collectionInfos) == 0 {
		if !jsonOutput {
			PrintWarning("No collections found in the Supabase database")
		} else {
			// Output empty JSON
			fmt.Println(`{"collections": [], "total": 0}`)
		}
		return
	}

	// Sort collections alphabetically by name
	sort.Slice(collectionInfos, func(i, j int) bool {
		return collectionInfos[i].Name < collectionInfos[j].Name
	})

	// Apply limit if specified
	if limit > 0 && len(collectionInfos) > limit {
		collectionInfos = collectionInfos[:limit]
	}

	if !jsonOutput {
		PrintSuccess(fmt.Sprintf("Found %d collections:", len(collectionInfos)))
		if limit > 0 && len(collectionInfos) == limit {
			PrintInfo(fmt.Sprintf("(showing first %d collections)", limit))
		}
		fmt.Println()

		// Display collections in compact format matching Weaviate/Mock listings
		for i, info := range collectionInfos {
			// Color coding: green for collections with documents, yellow for empty collections
			var nameColor string
			if info.Count > 0 {
				nameColor = GetStyledKeyProminent(info.Name)
			} else {
				nameColor = GetStyledKeyDimmed(info.Name)
			}

			// Document count with color
			var countStr string
			if info.Count > 0 {
				countStr = GetStyledNumber(fmt.Sprintf("%d items", info.Count))
			} else {
				countStr = GetStyledValueDimmed("empty")
			}

			// Schema type indicator - use generic doc icon for Supabase
			typeIndicator := GetStyledEmoji("📄")

			// Description if available
			var descriptionStr string
			if info.Description != "" {
				descriptionStr = GetStyledValueDimmed(fmt.Sprintf("(%s)", info.Description))
			}

			// Vectorizer/embedding model info
			var vectorizerStr string
			if info.Vectorizer != "" {
				vectorizerStr = fmt.Sprintf("[%s] ", GetStyledValueDimmed(info.Vectorizer))
			}

			// Compact single-line format
			fmt.Printf("%2d. %s %s %s %s%s\n",
				i+1,
				nameColor,
				countStr,
				typeIndicator,
				vectorizerStr,
				descriptionStr)
		}
	} else {
		// JSON output
		output := map[string]interface{}{
			"collections": collectionInfos,
			"total":       len(collectionInfos),
		}
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
	}
}

// ListMongoDBCollections lists MongoDB collections
func ListMongoDBCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)
		} else {
			PrintError(fmt.Sprintf("Failed to create MongoDB client: %v", err))
		}
		return
	}

	// Add timeout to context for API operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List collections
	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PrintError("Failed to connect to MongoDB database: connection timeout after 30 seconds")
			fmt.Println()
			PrintInfo("Please check that:")
			PrintInfo("  1. The MongoDB URI is correct")
			PrintInfo("  2. The database name is valid")
			PrintInfo("  3. There are no network/firewall issues")
		} else {
			PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		}
		return
	}

	if len(collectionInfos) == 0 {
		if !jsonOutput {
			PrintWarning("No collections found in the MongoDB database")
		} else {
			// Output empty JSON
			fmt.Println(`{"collections": [], "total": 0}`)
		}
		return
	}

	// Sort collections alphabetically by name
	sort.Slice(collectionInfos, func(i, j int) bool {
		return collectionInfos[i].Name < collectionInfos[j].Name
	})

	// Apply limit if specified
	if limit > 0 && len(collectionInfos) > limit {
		collectionInfos = collectionInfos[:limit]
	}

	if !jsonOutput {
		PrintSuccess(fmt.Sprintf("Found %d collections:", len(collectionInfos)))
		if limit > 0 && len(collectionInfos) == limit {
			PrintInfo(fmt.Sprintf("(showing first %d collections)", limit))
		}
		fmt.Println()

		// Display collections in compact format
		for i, info := range collectionInfos {
			// Color coding: green for collections with documents, yellow for empty collections
			var nameColor string
			if info.Count > 0 {
				nameColor = GetStyledKeyProminent(info.Name)
			} else {
				nameColor = GetStyledKeyDimmed(info.Name)
			}

			// Document count with color
			var countStr string
			if info.Count > 0 {
				countStr = GetStyledNumber(fmt.Sprintf("%d items", info.Count))
			} else {
				countStr = GetStyledValueDimmed("empty")
			}

			// Vectorizer/embedding model info
			var vectorizerStr string
			if info.Vectorizer != "" {
				vectorizerStr = fmt.Sprintf(" [%s]", GetStyledValueDimmed(info.Vectorizer))
			}

			// Display collection info
			fmt.Printf("%2d. %s %s%s 📄\n", i+1, nameColor, countStr, vectorizerStr)
		}
	} else {
		// JSON output
		type CollectionJSON struct {
			Name       string `json:"name"`
			Count      int64  `json:"count"`
			Vectorizer string `json:"vectorizer,omitempty"`
		}

		type OutputJSON struct {
			Collections []CollectionJSON `json:"collections"`
			Total       int              `json:"total"`
		}

		collections := make([]CollectionJSON, len(collectionInfos))
		for i, info := range collectionInfos {
			collections[i] = CollectionJSON{
				Name:       info.Name,
				Count:      info.Count,
				Vectorizer: info.Vectorizer,
			}
		}

		output := OutputJSON{
			Collections: collections,
			Total:       len(collections),
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
	}
}

// ListChromaCollections lists Chroma collections
func ListChromaCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)
		} else {
			PrintError(fmt.Sprintf("Failed to create Chroma client: %v", err))
		}
		return
	}

	// Add timeout to context for API operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List collections
	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PrintError("Failed to connect to Chroma: connection timeout after 30 seconds")
			fmt.Println()
			PrintInfo("Please check that:")
			PrintInfo("  1. Chroma is running (for local) or credentials are correct (for cloud)")
			PrintInfo("  2. The address is correct")
			PrintInfo("  3. There are no network/firewall issues")
		} else {
			PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		}
		return
	}

	if len(collectionInfos) == 0 {
		if !jsonOutput {
			PrintWarning("No collections found in the Chroma database")
		} else {
			// Output empty JSON
			fmt.Println(`{"collections": [], "total": 0}`)
		}
		return
	}

	// Sort collections alphabetically by name
	sort.Slice(collectionInfos, func(i, j int) bool {
		return collectionInfos[i].Name < collectionInfos[j].Name
	})

	// Apply limit if specified
	if limit > 0 && len(collectionInfos) > limit {
		collectionInfos = collectionInfos[:limit]
	}

	if !jsonOutput {
		PrintSuccess(fmt.Sprintf("Found %d collections:", len(collectionInfos)))
		if limit > 0 && len(collectionInfos) == limit {
			PrintInfo(fmt.Sprintf("(showing first %d collections)", limit))
		}
		fmt.Println()

		// Display collections in compact format
		for i, info := range collectionInfos {
			// Color coding: green for collections with documents, yellow for empty
			var nameColor string
			if info.Count > 0 {
				nameColor = GetStyledKeyProminent(info.Name)
			} else {
				nameColor = GetStyledKeyDimmed(info.Name)
			}

			// Document count with color
			var countStr string
			if info.Count > 0 {
				countStr = GetStyledNumber(fmt.Sprintf("%d items", info.Count))
			} else {
				countStr = GetStyledValueDimmed("empty")
			}

			// Vectorizer/embedding model info
			var vectorizerStr string
			if info.Vectorizer != "" {
				vectorizerStr = fmt.Sprintf(" [%s]", GetStyledValueDimmed(info.Vectorizer))
			}

			// Display collection info with vector icon
			fmt.Printf("%2d. %s %s%s 🔍\n", i+1, nameColor, countStr, vectorizerStr)
		}
	} else {
		// JSON output
		type CollectionJSON struct {
			Name       string `json:"name"`
			Count      int64  `json:"count"`
			Vectorizer string `json:"vectorizer,omitempty"`
		}

		type OutputJSON struct {
			Collections []CollectionJSON `json:"collections"`
			Total       int              `json:"total"`
		}

		collections := make([]CollectionJSON, len(collectionInfos))
		for i, info := range collectionInfos {
			collections[i] = CollectionJSON{
				Name:       info.Name,
				Count:      info.Count,
				Vectorizer: info.Vectorizer,
			}
		}

		output := OutputJSON{
			Collections: collections,
			Total:       len(collections),
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
	}
}

// ListQdrantCollections lists Qdrant collections
func ListQdrantCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)
		} else {
			PrintError(fmt.Sprintf("Failed to create Qdrant client: %v", err))
		}
		return
	}

	// Add timeout to context for API operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List collections
	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PrintError("Failed to connect to Qdrant: connection timeout after 30 seconds")
			fmt.Println()
			PrintInfo("Please check that:")
			PrintInfo("  1. Qdrant is running (for local) or credentials are correct (for cloud)")
			PrintInfo("  2. The address is correct (gRPC port 6334)")
			PrintInfo("  3. There are no network/firewall issues")
		} else {
			PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		}
		return
	}

	if len(collectionInfos) == 0 {
		if !jsonOutput {
			PrintWarning("No collections found in the Qdrant database")
		} else {
			// Output empty JSON
			fmt.Println(`{"collections": [], "total": 0}`)
		}
		return
	}

	// Sort collections alphabetically by name
	sort.Slice(collectionInfos, func(i, j int) bool {
		return collectionInfos[i].Name < collectionInfos[j].Name
	})

	// Apply limit if specified
	if limit > 0 && len(collectionInfos) > limit {
		collectionInfos = collectionInfos[:limit]
	}

	if !jsonOutput {
		PrintSuccess(fmt.Sprintf("Found %d collections:", len(collectionInfos)))
		if limit > 0 && len(collectionInfos) == limit {
			PrintInfo(fmt.Sprintf("(showing first %d collections)", limit))
		}
		fmt.Println()

		// Display collections in compact format
		for i, info := range collectionInfos {
			// Color coding: green for collections with documents, yellow for empty
			var nameColor string
			if info.Count > 0 {
				nameColor = GetStyledKeyProminent(info.Name)
			} else {
				nameColor = GetStyledKeyDimmed(info.Name)
			}

			// Document count with color
			var countStr string
			if info.Count > 0 {
				countStr = GetStyledNumber(fmt.Sprintf("%d items", info.Count))
			} else {
				countStr = GetStyledValueDimmed("empty")
			}

			fmt.Printf("%d. %s - %s\n", i+1, nameColor, countStr)
		}
	} else {
		// JSON output
		type jsonCollection struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		}

		type jsonOutput struct {
			Collections []jsonCollection `json:"collections"`
			Total       int              `json:"total"`
		}

		collections := make([]jsonCollection, len(collectionInfos))
		for i, info := range collectionInfos {
			collections[i] = jsonCollection{
				Name:  info.Name,
				Count: info.Count,
			}
		}

		output := jsonOutput{
			Collections: collections,
			Total:       len(collections),
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
	}
}

// ListMilvusCollections lists Milvus collections
func ListMilvusCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)
		} else {
			PrintError(fmt.Sprintf("Failed to create Milvus client: %v", err))
		}
		return
	}

	// Add timeout to context for API operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List collections
	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PrintError("Failed to connect to Milvus: connection timeout after 30 seconds")
			fmt.Println()
			PrintInfo("Please check that:")
			PrintInfo("  1. Milvus is running (for local) or credentials are correct (for cloud)")
			PrintInfo("  2. The address is correct")
			PrintInfo("  3. There are no network/firewall issues")
		} else {
			PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		}
		return
	}

	if len(collectionInfos) == 0 {
		if !jsonOutput {
			PrintWarning("No collections found in the Milvus database")
		} else {
			// Output empty JSON
			fmt.Println(`{"collections": [], "total": 0}`)
		}
		return
	}

	// Sort collections alphabetically by name
	sort.Slice(collectionInfos, func(i, j int) bool {
		return collectionInfos[i].Name < collectionInfos[j].Name
	})

	// Apply limit if specified
	if limit > 0 && len(collectionInfos) > limit {
		collectionInfos = collectionInfos[:limit]
	}

	if !jsonOutput {
		PrintSuccess(fmt.Sprintf("Found %d collections:", len(collectionInfos)))
		if limit > 0 && len(collectionInfos) == limit {
			PrintInfo(fmt.Sprintf("(showing first %d collections)", limit))
		}
		fmt.Println()

		// Display collections in compact format
		for i, info := range collectionInfos {
			// Color coding: green for collections with documents, yellow for empty
			var nameColor string
			if info.Count > 0 {
				nameColor = GetStyledKeyProminent(info.Name)
			} else {
				nameColor = GetStyledKeyDimmed(info.Name)
			}

			// Document count with color
			var countStr string
			if info.Count > 0 {
				countStr = GetStyledNumber(fmt.Sprintf("%d items", info.Count))
			} else {
				countStr = GetStyledValueDimmed("empty")
			}

			// Vectorizer/embedding model info
			var vectorizerStr string
			if info.Vectorizer != "" {
				vectorizerStr = fmt.Sprintf(" [%s]", GetStyledValueDimmed(info.Vectorizer))
			}

			// Display collection info with vector icon
			fmt.Printf("%2d. %s %s%s 🔍\n", i+1, nameColor, countStr, vectorizerStr)
		}
	} else {
		// JSON output
		type CollectionJSON struct {
			Name       string `json:"name"`
			Count      int64  `json:"count"`
			Vectorizer string `json:"vectorizer,omitempty"`
		}

		type OutputJSON struct {
			Collections []CollectionJSON `json:"collections"`
			Total       int              `json:"total"`
		}

		collections := make([]CollectionJSON, len(collectionInfos))
		for i, info := range collectionInfos {
			collections[i] = CollectionJSON{
				Name:       info.Name,
				Count:      info.Count,
				Vectorizer: info.Vectorizer,
			}
		}

		output := OutputJSON{
			Collections: collections,
			Total:       len(collections),
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
	}
}

// ListMockCollections lists mock collections
func ListMockCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	client := CreateMockClient(cfg)

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		if !jsonOutput {
			PrintWarning("No collections found in the database")
		} else {
			// Output empty JSON
			fmt.Println(`{"collections": [], "total": 0}`)
		}
		return
	}

	// Sort collections alphabetically
	sort.Strings(collections)

	// Apply limit if specified
	if limit > 0 && len(collections) > limit {
		collections = collections[:limit]
	}

	if !jsonOutput {
		PrintSuccess(fmt.Sprintf("Found %d collections:", len(collections)))
		if limit > 0 && len(collections) == limit {
			fmt.Printf("(showing first %d collections)\n", limit)
		}
		fmt.Println()
	}

	// Get collection information for each collection
	type CollectionInfo struct {
		Name              string
		DocumentCount     int
		HasDocuments      bool
		IsImageCollection bool
		SchemaType        string
	}

	var collectionInfos []CollectionInfo
	for _, collection := range collections {
		count, err := client.CountDocuments(ctx, collection)
		hasDocuments := err == nil && count > 0
		isImageCollection := isImageCollection(collection)

		// For mock collections, try to determine schema type from collection name or documents
		schemaType := "unknown"
		if hasDocuments {
			// Sample a few documents to determine schema type
			documents, err := client.ListDocuments(ctx, collection, 3)
			if err == nil && len(documents) > 0 {
				schemaType = "text" // Default to text
				// Check if it's an image collection based on document metadata
				// Look for specific image-related fields that indicate it's truly an image collection
				for _, doc := range documents {
					if metadata, ok := doc.Metadata["metadata"]; ok {
						if metadataStr, ok := metadata.(string); ok {
							metadataLower := strings.ToLower(metadataStr)
							// Check for specific image collection indicators
							if strings.Contains(metadataLower, "image_data") ||
								strings.Contains(metadataLower, "imagedata") ||
								strings.Contains(metadataLower, "imageformat") ||
								strings.Contains(metadataLower, "image_size") {
								schemaType = "image"
								break
							}
						}
					}
				}
			}
		}

		collectionInfos = append(collectionInfos, CollectionInfo{
			Name:              collection,
			DocumentCount:     count,
			HasDocuments:      hasDocuments,
			IsImageCollection: isImageCollection,
			SchemaType:        schemaType,
		})
	}

	// Output in JSON format if requested
	if jsonOutput {
		type JSONCollection struct {
			Name          string `json:"name"`
			DocumentCount int    `json:"document_count"`
			SchemaType    string `json:"schema_type"`
			IsImage       bool   `json:"is_image"`
		}

		type JSONOutput struct {
			Collections []JSONCollection `json:"collections"`
			Total       int              `json:"total"`
		}

		jsonCollections := make([]JSONCollection, len(collectionInfos))
		for i, info := range collectionInfos {
			jsonCollections[i] = JSONCollection{
				Name:          info.Name,
				DocumentCount: info.DocumentCount,
				SchemaType:    info.SchemaType,
				IsImage:       info.IsImageCollection,
			}
		}

		output := JSONOutput{
			Collections: jsonCollections,
			Total:       len(jsonCollections),
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}

		fmt.Println(string(jsonBytes))
		return
	}

	// Display collections in compact format
	for i, info := range collectionInfos {
		// Color coding: green for collections with documents, yellow for empty collections
		var nameColor string
		if info.HasDocuments {
			nameColor = GetStyledKeyProminent(info.Name)
		} else {
			nameColor = GetStyledKeyDimmed(info.Name)
		}

		// Document count with color
		var countStr string
		if info.HasDocuments {
			countStr = GetStyledNumber(fmt.Sprintf("%d items", info.DocumentCount))
		} else {
			countStr = GetStyledValueDimmed("empty")
		}

		// Schema type indicator
		var typeIndicator string
		switch info.SchemaType {
		case "image":
			typeIndicator = GetStyledEmoji("🖼️")
		case "text":
			typeIndicator = GetStyledEmoji("📄")
		default:
			typeIndicator = GetStyledEmoji("❓")
		}

		// Collection type indicator
		var collectionType string
		if info.IsImageCollection {
			collectionType = GetStyledValueDimmed("(image)")
		}

		// Mock indicator
		mockIndicator := GetStyledValueDimmed("(mock)")

		// Compact single-line format
		fmt.Printf("%2d. %s %s %s %s %s\n",
			i+1,
			nameColor,
			countStr,
			typeIndicator,
			collectionType,
			mockIndicator)

		if virtual && info.HasDocuments {
			// Show virtual structure summary for collections with documents
			// Note: Virtual summary not yet implemented for mock collections
			fmt.Printf("   Virtual structure: Not implemented for mock collections\n")
		}
	}
}

// ShowWeaviateCollection shows Weaviate collection details
func ShowWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool, verbose bool, showSchema bool, showMetadata bool, expandMetadata bool, outputYAML bool, outputJSON bool, yamlFile string, jsonFile string, compact bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to create client: %v", err))
		}
		return
	}

	// Check if collection exists by listing all collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		}
		return
	}

	// Check if the specific collection exists
	collectionExists := false
	for _, existingCollection := range collections {
		if existingCollection == collectionName {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		PrintError(fmt.Sprintf("Collection '%s' not found", collectionName))
		return
	}

	// Get document count using efficient method
	documentCount, err := client.CountDocuments(ctx, collectionName)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to get document count: %v", err))
		}
		return
	}

	// Display collection information with styling
	PrintStyledEmoji("📁")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection")
	fmt.Printf(": ")
	PrintStyledValueDimmed(collectionName)
	fmt.Println()

	PrintStyledKeyProminent("Database Type")
	fmt.Printf(": ")
	PrintStyledValueDimmed(string(cfg.Type))
	fmt.Println()

	PrintStyledKeyProminent("Document Count")
	fmt.Printf(": ")
	PrintStyledValueDimmed(fmt.Sprintf("%d", documentCount))
	fmt.Println()
	fmt.Println()

	// Show collection properties if available
	PrintStyledEmoji("⚙️")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection Properties")
	fmt.Println()
	PrintStyledKeyProminent("  Vector Database")
	fmt.Printf(": ")
	PrintStyledValueDimmed(string(cfg.Type))
	fmt.Println()
	PrintStyledKeyProminent("  URL")
	fmt.Printf(": ")
	PrintStyledValueDimmed(cfg.URL)
	fmt.Println()
	if cfg.APIKey != "" {
		PrintStyledKeyProminent("  API Key")
		fmt.Printf(": ")
		PrintStyledValueDimmed("[CONFIGURED]")
		fmt.Println()
	} else {
		PrintStyledKeyProminent("  API Key")
		fmt.Printf(": ")
		PrintStyledValueDimmed("[NOT CONFIGURED]")
		fmt.Println()
	}

	// Get embedding model from schema
	if schema, err := client.GetFullCollectionSchema(ctx, collectionName); err == nil && schema != nil && schema.Vectorizer != "" {
		PrintStyledKeyProminent("  Embedding Model")
		fmt.Printf(": ")
		PrintStyledValueDimmed(schema.Vectorizer)
		fmt.Println()
	}
	fmt.Println()

	if documentCount > 0 {
		// Get sample document for metadata analysis (just one document)
		sampleDocuments, err := client.ListDocuments(ctx, collectionName, 1)
		if err != nil {
			PrintWarning(fmt.Sprintf("Could not retrieve sample document: %v", err))
		} else if len(sampleDocuments) > 0 {
			// Get the actual document with full data using GetDocument
			sampleDoc, err := client.GetDocument(ctx, collectionName, sampleDocuments[0].ID)
			if err != nil {
				PrintWarning(fmt.Sprintf("Could not retrieve full sample document: %v", err))
				// Fall back to the basic document from ListDocuments
				sampleDoc = &sampleDocuments[0]
			}
			PrintStyledEmoji("📋")
			fmt.Printf(" ")
			PrintStyledKeyProminent("Sample Document Metadata")
			fmt.Println()
			if len(sampleDoc.Metadata) > 0 {
				metadataCount := 0
				for key, value := range sampleDoc.Metadata {
					if metadataCount >= shortLines {
						remainingFields := len(sampleDoc.Metadata) - shortLines
						fmt.Printf("... (truncated, %d more metadata fields)\n", remainingFields)
						break
					}

					var displayValue string
					if noTruncate {
						displayValue = fmt.Sprintf("%v", value)
					} else {
						displayValue = TruncateMetadataValue(value, 100) // Limit each value to 100 chars
					}

					// Style the key-value pair directly
					fmt.Printf("  - ")
					PrintStyledKeyProminent(key)
					fmt.Printf(": ")
					if key == "id" {
						PrintStyledID(displayValue)
					} else {
						PrintStyledValueDimmed(displayValue)
					}
					fmt.Println()
					metadataCount++
				}
			} else {
				PrintStyledKey("  No metadata available")
				fmt.Println()
			}
			fmt.Println()

			// Show sample content
			if len(sampleDoc.Content) > 0 {
				PrintStyledEmoji("📝")
				fmt.Printf(" ")
				PrintStyledKeyProminent("Sample Document Content")
				fmt.Println()

				// Check if this is image content (base64 data)
				if IsImageContent(sampleDoc.Content) {
					fmt.Printf("  ?? Image Document (Base64 encoded)\n")
					fmt.Printf("  ?? Content Size: %d characters\n", len(sampleDoc.Content))
					fmt.Printf("  ?? Preview: %s...\n", sampleDoc.Content[:min(50, len(sampleDoc.Content))])
					if len(sampleDoc.Content) > 50 {
						fmt.Printf("  ??  Full base64 data available (use --no-truncate to see all)\n")
					}
				} else {
					// Regular text content
					if noTruncate {
						fmt.Printf("%s\n", sampleDoc.Content)
					} else {
						contentLines := strings.Split(sampleDoc.Content, "\n")
						maxLines := shortLines
						if len(contentLines) > maxLines {
							for i := 0; i < maxLines; i++ {
								fmt.Printf("%s\n", contentLines[i])
							}
							fmt.Printf("  ... (%d more lines)\n", len(contentLines)-maxLines)
						} else {
							fmt.Printf("%s\n", sampleDoc.Content)
						}
					}
				}
				fmt.Println()
			} else {
				// No content, but show if it's an image document based on metadata
				if IsImageDocument(sampleDoc.Metadata) {
					PrintStyledEmoji("🖼️")
					fmt.Printf(" ")
					PrintStyledKeyProminent("Sample Document Content")
					fmt.Println()
					fmt.Printf("  ?? Image Document (no text content)\n")
					fmt.Printf("  ??  Image data stored in metadata fields\n")
					fmt.Println()
				}
			}
		}
	}

	// Handle export requests
	if outputYAML || outputJSON || yamlFile != "" || jsonFile != "" {
		// Determine if metadata should be included
		includeMetadata := showMetadata || expandMetadata || outputYAML || outputJSON || yamlFile != "" || jsonFile != ""

		// Export collection schema and metadata
		export, err := ExportCollectionSchemaAndMetadata(ctx, client, collectionName, includeMetadata, expandMetadata, compact)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to export collection data: %v", err))
			return
		}

		// Handle YAML output
		if outputYAML {
			yamlContent, err := ExportAsYAML(export)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to generate YAML: %v", err))
				return
			}
			fmt.Println()
			fmt.Println(yamlContent)
		}

		// Handle JSON output
		if outputJSON {
			jsonContent, err := ExportAsJSON(export)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to generate JSON: %v", err))
				return
			}
			fmt.Println()
			fmt.Println(jsonContent)
		}

		// Handle YAML file output
		if yamlFile != "" {
			yamlContent, err := ExportAsYAML(export)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to generate YAML: %v", err))
				return
			}
			err = WriteToFile(yamlFile, yamlContent)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to write YAML file: %v", err))
				return
			}
			PrintSuccess(fmt.Sprintf("Schema and metadata exported to YAML file: %s", yamlFile))
		}

		// Handle JSON file output
		if jsonFile != "" {
			jsonContent, err := ExportAsJSON(export)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to generate JSON: %v", err))
				return
			}
			err = WriteToFile(jsonFile, jsonContent)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to write JSON file: %v", err))
				return
			}
			PrintSuccess(fmt.Sprintf("Schema and metadata exported to JSON file: %s", jsonFile))
		}

		return
	}

	// Show schema if requested
	if showSchema {
		ShowCollectionSchema(ctx, client, collectionName)
	}

	// Show expanded metadata if requested
	if showMetadata || expandMetadata {
		ShowCollectionMetadata(ctx, client, collectionName, expandMetadata)
	}

	PrintSuccess(fmt.Sprintf("Collection '%s' summary retrieved successfully", collectionName))
}

// ShowMockCollection shows mock collection details
func ShowMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool, verbose bool, showSchema bool, showMetadata bool, expandMetadata bool, outputYAML bool, outputJSON bool, yamlFile string, jsonFile string, compact bool) {
	PrintInfo("Mock collection show not yet implemented in new structure")
}

// CountWeaviateCollections counts Weaviate collections
func CountWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %v", err)
	}

	collections, err := client.ListCollections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list collections: %v", err)
	}

	return len(collections), nil
}

// CountMockCollections counts mock collections
func CountMockCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	PrintInfo("Mock collection counting not yet implemented in new structure")
	return 0, fmt.Errorf("mock collection counting not yet implemented")
}

// DeleteWeaviateCollections deletes Weaviate collections
func DeleteWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	for _, collectionName := range collectionNames {
		err = client.DeleteCollection(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("failed to delete collection %s: %v", collectionName, err)
		}
	}

	return nil
}

// DeleteMockCollections deletes mock collections
func DeleteMockCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	PrintInfo("Mock collection deletion not yet implemented in new structure")
	return fmt.Errorf("collection deletion not yet implemented")
}

// DeleteWeaviateCollectionsByPattern deletes Weaviate collections by pattern
func DeleteWeaviateCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	// Get all collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err)
	}

	// Filter collections by pattern
	var matchingCollections []string
	for _, collection := range collections {
		if matchPattern(collection, pattern) {
			matchingCollections = append(matchingCollections, collection)
		}
	}

	// Delete matching collections
	for _, collectionName := range matchingCollections {
		err = client.DeleteCollection(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("failed to delete collection %s: %v", collectionName, err)
		}
	}

	return nil
}

// matchPattern checks if a string matches a pattern (shell glob or regex)
func matchPattern(str, pattern string) bool {
	// Check if pattern looks like regex (contains regex metacharacters)
	if isRegexPattern(pattern) {
		// Try to compile as regex
		regex, err := regexp.Compile(pattern)
		if err != nil {
			// If regex compilation fails, fall back to glob
			return matchGlobPattern(str, pattern)
		}
		return regex.MatchString(str)
	}

	// Treat as shell glob pattern
	return matchGlobPattern(str, pattern)
}

// isRegexPattern checks if a pattern looks like regex
func isRegexPattern(pattern string) bool {
	// Check for common regex metacharacters
	regexChars := []string{"^", "$", "\\", "[", "]", "(", ")", "{", "}", "|", "+", "?", "."}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}
	return false
}

// matchGlobPattern matches shell glob patterns
func matchGlobPattern(str, pattern string) bool {
	// Use filepath.Match for proper glob matching
	matched, err := filepath.Match(pattern, str)
	if err != nil {
		// If glob matching fails, fall back to simple contains
		return strings.Contains(str, strings.TrimSuffix(strings.TrimPrefix(pattern, "*"), "*"))
	}
	return matched
}

// DeleteWeaviateCollectionSchemasByPattern deletes Weaviate collection schemas by pattern
func DeleteWeaviateCollectionSchemasByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	// Get all collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err)
	}

	// Filter collections by pattern
	var matchingCollections []string
	for _, collection := range collections {
		if matchPattern(collection, pattern) {
			matchingCollections = append(matchingCollections, collection)
		}
	}

	// Delete schemas of matching collections
	for _, collectionName := range matchingCollections {
		err = client.DeleteCollectionSchema(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("failed to delete schema for collection %s: %v", collectionName, err)
		}
	}

	return nil
}

// DeleteMockCollectionsByPattern deletes mock collections by pattern
func DeleteMockCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	PrintInfo("Mock collection pattern deletion not yet implemented in new structure")
	return fmt.Errorf("collection pattern deletion not yet implemented")
}

// DeleteSupabaseCollections deletes Supabase collections
func DeleteSupabaseCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Supabase client: %v", err)
	}

	for _, collectionName := range collectionNames {
		err = client.DeleteCollection(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("failed to delete collection %s: %v", collectionName, err)
		}
	}

	return nil
}

// DeleteSupabaseCollectionsByPattern deletes Supabase collections by pattern
func DeleteSupabaseCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Supabase client: %v", err)
	}

	// Get all collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err)
	}

	// Filter collections by pattern
	var matchingCollections []string
	for _, collection := range collections {
		if matchPattern(collection.Name, pattern) {
			matchingCollections = append(matchingCollections, collection.Name)
		}
	}

	// Delete matching collections
	for _, collectionName := range matchingCollections {
		err = client.DeleteCollection(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("failed to delete collection %s: %v", collectionName, err)
		}
	}

	return nil
}

// DeleteAllWeaviateCollections deletes all documents from all Weaviate collections
func DeleteAllWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to create Weaviate client: %v", err))
		}
		return
	}

	// Get all collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		}
		return
	}

	if len(collections) == 0 {
		PrintInfo("No collections found to delete")
		return
	}

	PrintInfo(fmt.Sprintf("Found %d collections to clear", len(collections)))

	// Delete all documents from each collection
	successCount := 0
	for _, collectionName := range collections {
		PrintInfo(fmt.Sprintf("Clearing collection: %s", collectionName))

		// Delete all documents in the collection
		err := client.DeleteAllDocuments(ctx, collectionName)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to clear collection '%s': %v", collectionName, err))
			continue
		}

		successCount++
		PrintSuccess(fmt.Sprintf("? Cleared collection: %s", collectionName))
	}

	if successCount > 0 {
		PrintSuccess(fmt.Sprintf("Successfully cleared %d collections", successCount))
	} else {
		PrintError("Failed to clear any collections")
	}
}

// DeleteAllMockCollections deletes all documents from all mock collections
func DeleteAllMockCollections(ctx context.Context, cfg *config.VectorDBConfig) {
	client := CreateMockClient(cfg)

	// Get all collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		PrintInfo("No collections found to delete")
		return
	}

	PrintInfo(fmt.Sprintf("Found %d collections to clear", len(collections)))

	// Delete all documents from each collection
	successCount := 0
	for _, collectionName := range collections {
		PrintInfo(fmt.Sprintf("Clearing collection: %s", collectionName))

		// Delete all documents in the collection
		err := client.DeleteAllDocuments(ctx, collectionName)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to clear collection '%s': %v", collectionName, err))
			continue
		}

		successCount++
		PrintSuccess(fmt.Sprintf("? Cleared collection: %s", collectionName))
	}

	if successCount > 0 {
		PrintSuccess(fmt.Sprintf("Successfully cleared %d collections", successCount))
	} else {
		PrintError("Failed to clear any collections")
	}
}

// DeleteWeaviateCollectionSchema deletes Weaviate collection schema
func DeleteWeaviateCollectionSchema(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) error {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	err = client.DeleteCollectionSchema(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to delete collection schema: %v", err)
	}

	return nil
}

// showCollectionVirtualSummary shows virtual structure summary for a collection
func showCollectionVirtualSummary(ctx context.Context, client *weaviate.Client, collectionName string) {
	// This is a simplified implementation
	// In the real implementation, this would show virtual structure summary
	fmt.Printf("   Virtual structure: Not implemented yet\n")
}

// ShowCollectionSchema displays the collection schema with styling
func ShowCollectionSchema(ctx context.Context, client *weaviate.Client, collectionName string) {
	fmt.Println()
	PrintStyledEmoji("📐")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection Schema")
	fmt.Println()

	// Get collection schema from Weaviate
	schema, err := client.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get collection schema: %v", err))
		return
	}

	if schema == nil {
		PrintWarning("No schema information available")
		return
	}

	// Display schema information with styling
	PrintStyledKeyProminent("  Collection Name")
	fmt.Printf(": ")
	PrintStyledValueDimmed(schema.Class)
	fmt.Println()

	if schema.Vectorizer != "" {
		PrintStyledKeyProminent("  Vectorizer")
		fmt.Printf(": ")
		PrintStyledValueDimmed(schema.Vectorizer)
		fmt.Println()
	}

	// Note: ModuleConfig is not available in the current schema structure
	// This would need to be added to the CollectionSchema type if needed

	// Display properties
	if len(schema.Properties) > 0 {
		PrintStyledKeyProminent("  Properties")
		fmt.Printf(": ")
		fmt.Println()
		for i, prop := range schema.Properties {
			fmt.Printf("    %d. ", i+1)
			PrintStyledKeyProminent(prop.Name)
			fmt.Printf(" (")
			if len(prop.DataType) > 0 {
				PrintStyledValueDimmed(strings.Join(prop.DataType, ", "))
			}
			fmt.Printf(")")
			if prop.Description != "" {
				fmt.Printf(" - %s", prop.Description)
			}
			fmt.Println()

			// Show nested properties if available
			if len(prop.NestedProperties) > 0 {
				for j, nested := range prop.NestedProperties {
					fmt.Printf("      %d.%d. ", i+1, j+1)
					PrintStyledKey(nested.Name)
					fmt.Printf(" (")
					if len(nested.DataType) > 0 {
						PrintStyledValueDimmed(strings.Join(nested.DataType, ", "))
					}
					fmt.Printf(")")
					fmt.Println()
				}
			}
		}
	} else {
		PrintStyledKey("    No properties defined")
		fmt.Println()
	}

	fmt.Println()
}

// ShowCollectionMetadata displays collection metadata with styling
func ShowCollectionMetadata(ctx context.Context, client *weaviate.Client, collectionName string, expandMetadata bool) {
	fmt.Println()
	PrintStyledEmoji("🏷️")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection Metadata")
	fmt.Println()

	// Get all documents to analyze metadata patterns
	documents, err := client.ListDocuments(ctx, collectionName, 100) // Get up to 100 documents for analysis
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get documents for metadata analysis: %v", err))
		return
	}

	if len(documents) == 0 {
		PrintWarning("No documents found to analyze metadata")
		return
	}

	// Analyze metadata patterns
	metadataFields := make(map[string]int)
	metadataTypes := make(map[string]string)

	for _, doc := range documents {
		// Get full document to access metadata
		fullDoc, err := client.GetDocument(ctx, collectionName, doc.ID)
		if err != nil {
			continue // Skip documents that can't be retrieved
		}

		for key, value := range fullDoc.Metadata {
			metadataFields[key]++
			if metadataTypes[key] == "" {
				metadataTypes[key] = fmt.Sprintf("%T", value)
			}
		}
	}

	// Display metadata analysis
	PrintStyledKeyProminent("  Metadata Fields Analysis")
	fmt.Printf(" (from %d documents):", len(documents))
	fmt.Println()

	if len(metadataFields) == 0 {
		PrintStyledKey("    No metadata fields found")
		fmt.Println()
		return
	}

	// Sort fields by frequency
	type fieldInfo struct {
		name  string
		count int
		typ   string
	}

	var fields []fieldInfo
	for name, count := range metadataFields {
		fields = append(fields, fieldInfo{name, count, metadataTypes[name]})
	}

	// Sort by count (descending)
	for i := 0; i < len(fields)-1; i++ {
		for j := i + 1; j < len(fields); j++ {
			if fields[i].count < fields[j].count {
				fields[i], fields[j] = fields[j], fields[i]
			}
		}
	}

	// Display fields
	for i, field := range fields {
		fmt.Printf("    %d. ", i+1)
		PrintStyledKeyProminent(field.name)
		fmt.Printf(" (")
		PrintStyledValueDimmed(field.typ)
		fmt.Printf(") - ")
		PrintStyledValueDimmed(fmt.Sprintf("%d occurrences", field.count))
		fmt.Println()

		// Show sample values if expandMetadata is true
		if expandMetadata && field.count > 0 {
			// Find a sample document with this field
			for _, doc := range documents {
				fullDoc, err := client.GetDocument(ctx, collectionName, doc.ID)
				if err != nil {
					continue
				}
				if value, exists := fullDoc.Metadata[field.name]; exists {
					fmt.Printf("      Sample: ")
					sampleValue := fmt.Sprintf("%v", value)
					if len(sampleValue) > 100 {
						sampleValue = sampleValue[:100] + "..."
					}
					PrintStyledValueDimmed(sampleValue)
					fmt.Println()
					break
				}
			}
		}
	}

	fmt.Println()
}

// convertSchemaDefinitionToCollectionSchema converts a config.SchemaDefinition to weaviate.CollectionSchema
func convertSchemaDefinitionToCollectionSchema(schemaDef *config.SchemaDefinition, collectionName, metadataMode string) (*weaviate.CollectionSchema, error) {
	// Extract schema map (already map[string]interface{} type)
	schemaMap := schemaDef.Schema

	// Create the collection schema
	schema := &weaviate.CollectionSchema{}

	// Set collection name (use provided name or class name from schema)
	if collectionName != "" {
		schema.Class = collectionName
	} else if className, ok := schemaMap["class"].(string); ok {
		schema.Class = className
	} else {
		return nil, fmt.Errorf("no collection name provided and no class name in schema")
	}

	// Set vectorizer
	if vectorizer, ok := schemaMap["vectorizer"].(string); ok {
		schema.Vectorizer = vectorizer
	}

	// Start with properties from schema section
	var properties []weaviate.SchemaProperty
	if props, ok := schemaMap["properties"].([]interface{}); ok {
		for _, prop := range props {
			propMap, ok := prop.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip the metadata property if we're flattening metadata fields
			if name, ok := propMap["name"].(string); ok && name == "metadata" && metadataMode == "flat" && schemaDef.Metadata != nil {
				continue
			}

			property := weaviate.SchemaProperty{}

			// Set property name
			if name, ok := propMap["name"].(string); ok {
				property.Name = name
			}

			// Set data type
			if dataType, ok := propMap["datatype"].([]interface{}); ok {
				property.DataType = make([]string, len(dataType))
				for j, dt := range dataType {
					if dtStr, ok := dt.(string); ok {
						property.DataType[j] = dtStr
					}
				}
			}

			// Set description
			if description, ok := propMap["description"].(string); ok {
				property.Description = description
			}

			// Handle nested properties
			if nestedProps, ok := propMap["nestedProperties"].([]interface{}); ok {
				property.NestedProperties = make([]weaviate.SchemaProperty, len(nestedProps))
				for j, nested := range nestedProps {
					nestedMap, ok := nested.(map[string]interface{})
					if !ok {
						continue
					}

					nestedProp := weaviate.SchemaProperty{}

					if name, ok := nestedMap["name"].(string); ok {
						nestedProp.Name = name
					}

					if dataType, ok := nestedMap["datatype"].([]interface{}); ok {
						nestedProp.DataType = make([]string, len(dataType))
						for k, dt := range dataType {
							if dtStr, ok := dt.(string); ok {
								nestedProp.DataType[k] = dtStr
							}
						}
					}

					if description, ok := nestedMap["description"].(string); ok {
						nestedProp.Description = description
					}

					property.NestedProperties[j] = nestedProp
				}
			}

			properties = append(properties, property)
		}
	}

	// Add metadata fields as flat properties (only when metadataMode is "flat")
	if metadataMode == "flat" && schemaDef.Metadata != nil {
		// Reserved property names in Weaviate
		reservedNames := map[string]bool{
			"id":     true,
			"_id":    true,
			"_meta":  true,
			"_class": true,
		}

		for fieldName, fieldValue := range schemaDef.Metadata {
			// Skip reserved property names
			if reservedNames[fieldName] {
				continue
			}

			// Skip if this field is already defined in schema properties
			alreadyExists := false
			for _, existingProp := range properties {
				if existingProp.Name == fieldName {
					alreadyExists = true
					break
				}
			}
			if alreadyExists {
				continue
			}

			// Create property from metadata field
			property := weaviate.SchemaProperty{
				Name: fieldName,
			}

			// Determine data type from field value
			switch fieldValue.(type) {
			case string:
				property.DataType = []string{"string"}
			case bool:
				property.DataType = []string{"boolean"}
			case int, int8, int16, int32, int64, float32, float64:
				property.DataType = []string{"number"}
			case []interface{}:
				// For arrays, check if it's a simple type array or complex
				if len(fieldValue.([]interface{})) > 0 {
					switch fieldValue.([]interface{})[0].(type) {
					case string:
						property.DataType = []string{"string[]"}
					case int, int8, int16, int32, int64, float32, float64:
						property.DataType = []string{"number[]"}
					case bool:
						property.DataType = []string{"boolean[]"}
					default:
						property.DataType = []string{"string[]"}
					}
				} else {
					property.DataType = []string{"string[]"}
				}
			case map[string]interface{}:
				// Handle complex objects - check for special array definitions
				fieldMap := fieldValue.(map[string]interface{})
				if fieldType, ok := fieldMap["type"].(string); ok && fieldType == "array" {
					if items, ok := fieldMap["items"].(string); ok {
						switch items {
						case "integer", "number":
							property.DataType = []string{"number[]"}
						case "string":
							property.DataType = []string{"string[]"}
						case "boolean":
							property.DataType = []string{"boolean[]"}
						default:
							property.DataType = []string{"string[]"}
						}
					} else {
						property.DataType = []string{"string[]"}
					}
				} else {
					// Handle complex objects - convert to string for now
					property.DataType = []string{"string"}
				}
			default:
				// Default to string for unknown types
				property.DataType = []string{"string"}
			}

			// Add description based on field name
			switch fieldName {
			case "id":
				property.Description = "unique identifier for the document"
			case "added_date":
				property.Description = "date when the document was added to the collection"
			case "creation_date":
				property.Description = "date when the document was originally created"
			case "modified_date":
				property.Description = "date when the document was last modified"
			case "creator":
				property.Description = "person or system that created the document"
			case "producer":
				property.Description = "software or system that produced the document"
			case "title":
				property.Description = "title of the document"
			case "ai_summary":
				property.Description = "AI-generated summary of the document content"
			case "filename":
				property.Description = "name of the file"
			case "is_chunked":
				property.Description = "whether the document has been chunked"
			case "total_chunks":
				property.Description = "total number of chunks if document is chunked"
			case "chunk_index":
				property.Description = "index of this chunk if document is chunked"
			case "chunk_sizes":
				property.Description = "array of chunk sizes if document is chunked"
			case "original_filename":
				property.Description = "original filename before processing"
			case "storage_path":
				property.Description = "path where the document is stored"
			case "type":
				property.Description = "type of the document"
			default:
				property.Description = fmt.Sprintf("metadata field: %s", fieldName)
			}

			properties = append(properties, property)
		}
	}

	schema.Properties = properties
	return schema, nil
}

// QueryWeaviateCollection performs semantic search on a Weaviate collection
func QueryWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, queryText string, options weaviate.QueryOptions) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to create Weaviate client: %v", err))
		}
		return
	}

	// Perform the semantic search
	results, err := client.Query(ctx, collectionName, queryText, options)
	if err != nil {
		// Check if this might be a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
					}
				}
			}
		} else {
			PrintError(fmt.Sprintf("Failed to query collection '%s': %v", collectionName, err))
		}
		return
	}

	// Display results
	DisplayQueryResults(results, collectionName, queryText, options.NoTruncate, options.JSONOutput)
}

// QueryCollection performs semantic search on a collection using the vectordb abstraction (works for all DB types)
func QueryCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, queryText string, options weaviate.QueryOptions) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Convert weaviate.QueryOptions to vectordb.QueryOptions
	vdbOptions := &vectordb.QueryOptions{
		TopK:           options.TopK,
		Distance:       options.Distance,
		SearchMetadata: options.SearchMetadata,
		NoTruncate:     options.NoTruncate,
		UseBM25:        options.UseBM25,
	}

	// Perform semantic or BM25 search based on options
	var results []*vectordb.QueryResult
	if vdbOptions.UseBM25 {
		results, err = client.SearchBM25(ctx, collectionName, queryText, vdbOptions)
	} else {
		results, err = client.SearchSemantic(ctx, collectionName, queryText, vdbOptions)
	}

	if err != nil {
		PrintError(fmt.Sprintf("Failed to query collection '%s': %v", collectionName, err))
		return
	}

	// Convert vectordb.QueryResult to weaviate.QueryResult for display
	weaviateResults := make([]weaviate.QueryResult, len(results))
	for i, r := range results {
		weaviateResults[i] = weaviate.QueryResult{
			ID:       r.Document.ID,
			Content:  r.Document.Content,
			Score:    r.Score,
			Metadata: r.Document.Metadata,
		}
	}

	// Display results
	DisplayQueryResults(weaviateResults, collectionName, queryText, options.NoTruncate, options.JSONOutput)
}

// QueryMockCollection performs semantic search on a mock collection
func QueryMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, queryText string, options weaviate.QueryOptions) {
	client := CreateMockClient(cfg)

	// Perform mock semantic search
	results, err := client.Query(ctx, collectionName, queryText, options)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to query mock collection '%s': %v", collectionName, err))
		return
	}

	// Display results
	DisplayQueryResults(results, collectionName, queryText, options.NoTruncate, options.JSONOutput)
}

// Generic collection operations for VectorDB interface (Milvus, etc.)

// CountGenericCollections counts collections using the VectorDB interface
func CountGenericCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %w", err)
	}

	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list collections: %w", err)
	}

	return len(collectionInfos), nil
}

// DeleteGenericCollections deletes collections using the VectorDB interface
func DeleteGenericCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	for _, name := range collectionNames {
		if err := client.DeleteCollection(ctx, name); err != nil {
			return fmt.Errorf("failed to delete collection '%s': %w", name, err)
		}
		PrintSuccess(fmt.Sprintf("Deleted collection: %s", name))
	}

	return nil
}

// DeleteGenericCollectionsByPattern deletes collections matching a pattern using VectorDB interface
func DeleteGenericCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// List all collections
	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Filter by pattern
	var matchingCollections []string
	for _, collection := range collectionInfos {
		if matchPattern(collection.Name, pattern) {
			matchingCollections = append(matchingCollections, collection.Name)
		}
	}

	if len(matchingCollections) == 0 {
		PrintWarning(fmt.Sprintf("No collections found matching pattern: %s", pattern))
		return nil
	}

	// Delete matched collections
	for _, name := range matchingCollections {
		if err := client.DeleteCollection(ctx, name); err != nil {
			return fmt.Errorf("failed to delete collection '%s': %w", name, err)
		}
		PrintSuccess(fmt.Sprintf("Deleted collection: %s", name))
	}

	return nil
}

// ShowGenericCollection shows collection details using VectorDB interface
func ShowGenericCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool, verbose bool, showSchema bool, showMetadata bool, expandMetadata bool, outputYAML bool, outputJSON bool, yamlFile string, jsonFile string, compact bool) {
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Get collection info
	collectionInfos, err := client.ListCollections(ctx)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	var found *vectordb.CollectionInfo
	for _, info := range collectionInfos {
		if info.Name == collectionName {
			found = &info
			break
		}
	}

	if found == nil {
		PrintError(fmt.Sprintf("Collection '%s' not found", collectionName))
		return
	}

	// Display collection information with styling (matching Weaviate format)
	PrintStyledEmoji("📁")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection")
	fmt.Printf(": ")
	PrintStyledValueDimmed(collectionName)
	fmt.Println()

	PrintStyledKeyProminent("Database Type")
	fmt.Printf(": ")
	PrintStyledValueDimmed(string(cfg.Type))
	fmt.Println()

	PrintStyledKeyProminent("Document Count")
	fmt.Printf(": ")
	PrintStyledValueDimmed(fmt.Sprintf("%d", found.Count))
	fmt.Println()
	fmt.Println()

	// Show collection properties
	PrintStyledEmoji("⚙️")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection Properties")
	fmt.Println()
	PrintStyledKeyProminent("  Vector Database")
	fmt.Printf(": ")
	PrintStyledValueDimmed(string(cfg.Type))
	fmt.Println()

	// Show URL/Address based on database type
	if cfg.URL != "" {
		PrintStyledKeyProminent("  URL")
		fmt.Printf(": ")
		PrintStyledValueDimmed(cfg.URL)
		fmt.Println()
	} else if cfg.Address != "" {
		PrintStyledKeyProminent("  Address")
		fmt.Printf(": ")
		PrintStyledValueDimmed(cfg.Address)
		fmt.Println()
	}

	// Show API Key status
	if cfg.APIKey != "" || cfg.DatabaseKey != "" {
		PrintStyledKeyProminent("  API Key")
		fmt.Printf(": ")
		PrintStyledValueDimmed("[CONFIGURED]")
		fmt.Println()
	}

	// Show embedding model if available
	if found.Vectorizer != "" {
		PrintStyledKeyProminent("  Embedding Model")
		fmt.Printf(": ")
		PrintStyledValueDimmed(found.Vectorizer)
		fmt.Println()
	}
	fmt.Println()

	// Show sample document if documents exist
	if found.Count > 0 {
		// Get sample document
		sampleDocuments, err := client.ListDocuments(ctx, collectionName, 1, 0)
		if err != nil {
			PrintWarning(fmt.Sprintf("Could not retrieve sample document: %v", err))
		} else if len(sampleDocuments) > 0 {
			sampleDoc := sampleDocuments[0]

			// Show sample metadata
			PrintStyledEmoji("📋")
			fmt.Printf(" ")
			PrintStyledKeyProminent("Sample Document Metadata")
			fmt.Println()
			if len(sampleDoc.Metadata) > 0 {
				metadataCount := 0
				for key, value := range sampleDoc.Metadata {
					if metadataCount >= shortLines {
						remainingFields := len(sampleDoc.Metadata) - shortLines
						fmt.Printf("... (truncated, %d more metadata fields)\n", remainingFields)
						break
					}

					var displayValue string
					if noTruncate {
						displayValue = fmt.Sprintf("%v", value)
					} else {
						displayValue = TruncateMetadataValue(value, 100)
					}

					fmt.Printf("  - ")
					PrintStyledKeyProminent(key)
					fmt.Printf(": ")
					if key == "id" {
						PrintStyledID(displayValue)
					} else {
						PrintStyledValueDimmed(displayValue)
					}
					fmt.Println()
					metadataCount++
				}
			} else {
				PrintStyledKey("  No metadata available")
				fmt.Println()
			}
			fmt.Println()

			// Show sample content
			if len(sampleDoc.Content) > 0 {
				PrintStyledEmoji("📝")
				fmt.Printf(" ")
				PrintStyledKeyProminent("Sample Document Content")
				fmt.Println()

				if IsImageContent(sampleDoc.Content) {
					fmt.Printf("  🖼️ Image Document (Base64 encoded)\n")
					fmt.Printf("  📊 Content Size: %d characters\n", len(sampleDoc.Content))
				} else {
					if noTruncate {
						fmt.Printf("%s\n", sampleDoc.Content)
					} else {
						contentLines := strings.Split(sampleDoc.Content, "\n")
						maxLines := shortLines
						if len(contentLines) > maxLines {
							for i := 0; i < maxLines; i++ {
								fmt.Printf("%s\n", contentLines[i])
							}
							fmt.Printf("  ... (%d more lines)\n", len(contentLines)-maxLines)
						} else {
							fmt.Printf("%s\n", sampleDoc.Content)
						}
					}
				}
				fmt.Println()
			}
		}
	}

	// Success message
	PrintSuccess(fmt.Sprintf("Collection '%s' summary retrieved successfully", collectionName))

	// TODO: Implement schema display for generic collections
	if showSchema {
		PrintWarning("Schema display not yet implemented for this database type")
	}
}
