// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// CollectionSpec represents a collection with optional VDB specification
type CollectionSpec struct {
	Name   string // Collection name
	VDBKey string // VDB key (e.g., "weaviate-local", "milvus-cloud"), empty if not specified
}

// ParseCollectionSpec parses a collection specification in the format:
//   - "CollectionName" → {Name: "CollectionName", VDBKey: ""}
//   - "CollectionName:vdb-key" → {Name: "CollectionName", VDBKey: "vdb-key"}
func ParseCollectionSpec(spec string) (CollectionSpec, error) {
	if spec == "" {
		return CollectionSpec{}, fmt.Errorf("collection spec cannot be empty")
	}

	parts := strings.SplitN(spec, ":", 2)

	result := CollectionSpec{
		Name: parts[0],
	}

	if len(parts) == 2 {
		result.VDBKey = parts[1]
	}

	// Validate collection name
	if result.Name == "" {
		return CollectionSpec{}, fmt.Errorf("collection name cannot be empty")
	}

	return result, nil
}

// ParseCollectionSpecs parses multiple collection specifications
func ParseCollectionSpecs(specs []string) ([]CollectionSpec, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("no collection specs provided")
	}

	result := make([]CollectionSpec, len(specs))
	for i, spec := range specs {
		parsed, err := ParseCollectionSpec(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid collection spec %q: %w", spec, err)
		}
		result[i] = parsed
	}

	return result, nil
}

// VDBConfigResolver resolves VDB configurations for collection specs
type VDBConfigResolver struct {
	config        *config.Config
	defaultConfig *config.VectorDBConfig
	resolvedCache map[string]*config.VectorDBConfig
}

// NewVDBConfigResolver creates a new VDB config resolver
func NewVDBConfigResolver(cfg *config.Config, defaultConfig *config.VectorDBConfig) *VDBConfigResolver {
	return &VDBConfigResolver{
		config:        cfg,
		defaultConfig: defaultConfig,
		resolvedCache: make(map[string]*config.VectorDBConfig),
	}
}

// ResolveConfigs resolves VDB configs for collection specs
// Returns map[collectionName]*config.VectorDBConfig
func (r *VDBConfigResolver) ResolveConfigs(specs []CollectionSpec) (map[string]*config.VectorDBConfig, error) {
	result := make(map[string]*config.VectorDBConfig)

	for _, spec := range specs {
		var vdbConfig *config.VectorDBConfig
		var err error

		if spec.VDBKey != "" {
			// Explicit VDB specified - resolve by key
			vdbConfig, err = r.resolveByKey(spec.VDBKey)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve VDB %q for collection %q: %w",
					spec.VDBKey, spec.Name, err)
			}
		} else {
			// No explicit VDB - use default
			if r.defaultConfig == nil {
				return nil, fmt.Errorf("collection %q has no VDB specified and no default VDB configured",
					spec.Name)
			}
			vdbConfig = r.defaultConfig
		}

		result[spec.Name] = vdbConfig
	}

	return result, nil
}

// resolveByKey resolves a VDB config by its key (e.g., "weaviate-local", "milvus-cloud")
func (r *VDBConfigResolver) resolveByKey(key string) (*config.VectorDBConfig, error) {
	// Check cache first
	if cached, ok := r.resolvedCache[key]; ok {
		return cached, nil
	}

	// Map common key formats to VectorDB config
	if r.config == nil {
		return nil, fmt.Errorf("no config available to resolve VDB key %q", key)
	}

	// Try to find matching VDB in config
	for _, vdb := range r.config.Databases.VectorDatabases {
		// Match by type or name
		vdbKey := getVDBKey(&vdb)
		if vdbKey == key || vdb.Name == key {
			vdbCopy := vdb
			r.resolvedCache[key] = &vdbCopy
			return &vdbCopy, nil
		}
	}

	return nil, fmt.Errorf("no VDB configuration found for key %q", key)
}

// getVDBKey returns a canonical key for a VDB config
func getVDBKey(vdb *config.VectorDBConfig) string {
	// Return the type as the key (e.g., "weaviate-cloud", "milvus-local")
	return string(vdb.Type)
}

// HasExplicitVDBs returns true if any collection spec has an explicit VDB
func HasExplicitVDBs(specs []CollectionSpec) bool {
	for _, spec := range specs {
		if spec.VDBKey != "" {
			return true
		}
	}
	return false
}

// GetUniqueVDBKeys returns the unique VDB keys from collection specs
func GetUniqueVDBKeys(specs []CollectionSpec) []string {
	seen := make(map[string]bool)
	var result []string

	for _, spec := range specs {
		if spec.VDBKey != "" && !seen[spec.VDBKey] {
			seen[spec.VDBKey] = true
			result = append(result, spec.VDBKey)
		}
	}

	return result
}
