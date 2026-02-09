// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// SentenceTransformersProvider implements EmbeddingProvider for sentence-transformers models
type SentenceTransformersProvider struct {
	modelName string
	fullName  string // Full model name with prefix
}

// NewSentenceTransformersProvider creates a new sentence-transformers provider
func NewSentenceTransformersProvider(modelName string) (*SentenceTransformersProvider, error) {
	// Strip the sentence-transformers/ prefix if present
	shortName := strings.TrimPrefix(modelName, "sentence-transformers/")

	return &SentenceTransformersProvider{
		modelName: shortName,
		fullName:  modelName,
	}, nil
}

// GenerateEmbedding generates an embedding for a single text
func (p *SentenceTransformersProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := p.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}

	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts (batch)
func (p *SentenceTransformersProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	// Create Python script that generates embeddings
	pythonScript := `
import sys
import json
from sentence_transformers import SentenceTransformer

def main():
	try:
		# Load model
		model_name = sys.argv[1]
		model = SentenceTransformer(model_name)

		# Read texts from stdin (JSON array)
		texts_json = sys.stdin.read()
		texts = json.loads(texts_json)

		# Generate embeddings
		embeddings = model.encode(texts, show_progress_bar=False)

		# Convert to list and output as JSON
		embeddings_list = embeddings.tolist()
		print(json.dumps(embeddings_list))

	except Exception as e:
		print(json.dumps({"error": str(e)}), file=sys.stderr)
		sys.exit(1)

if __name__ == "__main__":
	main()
`

	// Prepare input JSON
	textsJSON, err := json.Marshal(texts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal texts: %w", err)
	}

	// Execute Python script with context
	cmd := exec.CommandContext(ctx, "python3", "-c", pythonScript, p.modelName)
	cmd.Stdin = strings.NewReader(string(textsJSON))

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("sentence-transformers error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute Python script: %w", err)
	}

	// Parse output JSON
	var embeddings [][]float64
	if err := json.Unmarshal(output, &embeddings); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings JSON: %w", err)
	}

	if len(embeddings) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d", len(texts), len(embeddings))
	}

	return embeddings, nil
}

// GetModelName returns the model name
func (p *SentenceTransformersProvider) GetModelName() string {
	return p.fullName
}

// GetProvider returns the provider name
func (p *SentenceTransformersProvider) GetProvider() string {
	return "sentence-transformers"
}

// IsAvailable checks if Python and sentence-transformers are installed
func (p *SentenceTransformersProvider) IsAvailable(ctx context.Context) error {
	// Check if python3 is available
	if _, err := exec.LookPath("python3"); err != nil {
		return fmt.Errorf("python3 not found in PATH\n\nTo install Python:\n  macOS: brew install python3\n  Ubuntu: sudo apt install python3\n  See: https://www.python.org/downloads/")
	}

	// Check if sentence-transformers is installed
	checkScript := `
import sys
try:
	import sentence_transformers
	print("OK")
except ImportError:
	sys.exit(1)
`

	cmd := exec.CommandContext(ctx, "python3", "-c", checkScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sentence-transformers not installed\n\nTo install:\n  pip3 install sentence-transformers\n\nOr:\n  pip3 install -U sentence-transformers\n\nSee: https://www.sbert.net/docs/installation.html")
	}

	// Check if the specific model can be loaded (will download if needed)
	// For now, skip this check to avoid auto-downloading on availability check
	// The model will be downloaded on first use

	return nil
}
