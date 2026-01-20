package pipeline

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/pdf"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Processor handles document processing and ingestion
type Processor struct {
	vdbClient vectordb.VectorDBClient
	llmClient *llm.OpenAIClient
	options   *IngestOptions
	progress  *ProgressTracker
}

// NewProcessor creates a new document processor
func NewProcessor(vdbClient vectordb.VectorDBClient, llmClient *llm.OpenAIClient, options *IngestOptions, progress *ProgressTracker) *Processor {
	return &Processor{
		vdbClient: vdbClient,
		llmClient: llmClient,
		options:   options,
		progress:  progress,
	}
}

// ProcessFiles processes files concurrently and creates documents in batches
func (p *Processor) ProcessFiles(ctx context.Context, files []FileInfo) (*IngestReport, error) {
	startTime := time.Now()

	report := &IngestReport{
		Status:     "success",
		StartTime:  startTime,
		Collection: p.options.Collection,
		VDBType:    p.options.VDBType,
		BatchSize:  p.options.BatchSize,
		Workers:    p.options.Workers,
	}

	// Load state for resume capability
	processedFiles := make(map[string]bool)
	if p.options.Resume && p.options.StateFile != "" {
		state, _ := loadState(p.options.StateFile)
		if state != nil {
			processedFiles = state.ProcessedFiles
		}
	}

	// Filter out already processed files
	var filesToProcess []FileInfo
	for _, file := range files {
		if !processedFiles[file.Hash] {
			filesToProcess = append(filesToProcess, file)
		} else {
			report.FilesSkipped++
		}
	}

	report.FilesScanned = len(files)

	// Start processing progress bar
	p.progress.StartProcessing(len(filesToProcess), p.options.BatchSize, p.options.Workers)

	// Create worker pool
	resultsChan := make(chan *ProcessingResult, len(filesToProcess))
	var wg sync.WaitGroup

	// Create semaphore for limiting concurrent workers
	semaphore := make(chan struct{}, p.options.Workers)

	// Process files concurrently
	for _, file := range filesToProcess {
		wg.Add(1)
		go func(f FileInfo) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := p.processFile(ctx, f)
			resultsChan <- result

			// Update state if resume is enabled
			if p.options.Resume && p.options.StateFile != "" && result.Error == nil {
				processedFiles[f.Hash] = true
				saveState(p.options.StateFile, &IngestState{
					ProcessedFiles: processedFiles,
					LastUpdated:    time.Now(),
				})
			}
		}(file)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results and batch documents
	var allDocuments []*vectordb.Document
	for result := range resultsChan {
		// Update progress
		p.progress.UpdateProgress(1)

		if result.Error != nil {
			report.FilesFailed++
			report.Errors = append(report.Errors, ErrorDetail{
				File:      result.File.Path,
				Error:     result.Error.Error(),
				Timestamp: time.Now(),
			})
			continue
		}

		report.FilesProcessed++
		report.DocumentsCreated += result.DocumentCount
		allDocuments = append(allDocuments, result.Documents...)
	}

	// Finish processing progress
	p.progress.FinishProcessing()

	// Create documents in batches
	if !p.options.DryRun && len(allDocuments) > 0 {
		if err := p.createDocumentsInBatches(ctx, allDocuments); err != nil {
			report.Status = "partial"
			report.Errors = append(report.Errors, ErrorDetail{
				File:      "batch_creation",
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
		}
	}

	// Finalize report
	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime).Seconds()

	if report.Duration > 0 {
		report.ThroughputFiles = float64(report.FilesProcessed) / report.Duration
		report.ThroughputDocs = float64(report.DocumentsCreated) / report.Duration
	}

	if report.FilesFailed > 0 && report.FilesProcessed == 0 {
		report.Status = "failure"
	} else if report.FilesFailed > 0 {
		report.Status = "partial"
	}

	return report, nil
}

// processFile processes a single file based on its type
func (p *Processor) processFile(ctx context.Context, file FileInfo) *ProcessingResult {
	startTime := time.Now()

	var documents []*vectordb.Document
	var err error

	switch file.Type {
	case FileTypePDF:
		documents, err = p.processPDF(ctx, file)
	case FileTypeTXT, FileTypeMD:
		documents, err = p.processText(ctx, file)
	case FileTypeJSON:
		documents, err = p.processJSON(ctx, file)
	case FileTypeYAML:
		documents, err = p.processYAML(ctx, file)
	default:
		err = fmt.Errorf("unsupported file type: %s", file.Type)
	}

	return &ProcessingResult{
		File:          file,
		Documents:     documents,
		DocumentCount: len(documents),
		Error:         err,
		Duration:      time.Since(startTime),
	}
}

// processPDF extracts text and creates documents from PDF
func (p *Processor) processPDF(ctx context.Context, file FileInfo) ([]*vectordb.Document, error) {
	// Extract PDF content (text only for now, images optional) with 2000 char metadata limit
	textData, _, err := pdf.ExtractPDFContent(file.Path, 1000, true, 100, 2000, true)
	if err != nil {
		return nil, fmt.Errorf("failed to extract PDF content: %w", err)
	}

	var documents []*vectordb.Document
	for _, text := range textData {
		// Generate embedding
		embedding, err := p.llmClient.GenerateEmbedding(ctx, text.Content, "")
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %w", err)
		}

		// Convert float64 to float32
		vector := make([]float32, len(embedding))
		for i, v := range embedding {
			vector[i] = float32(v)
		}

		// Create metadata
		metadata := make(map[string]interface{})
		if p.options.Metadata != nil {
			for k, v := range p.options.Metadata {
				metadata[k] = v
			}
		}
		metadata["source_file"] = file.Path
		metadata["file_type"] = string(file.Type)
		metadata["page_number"] = text.PageNumber
		metadata["file_hash"] = file.Hash

		doc := &vectordb.Document{
			ID:       uuid.New().String(),
			Text:     text.Content,
			Content:  text.Content,
			Metadata: metadata,
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

// processText processes plain text files
func (p *Processor) processText(ctx context.Context, file FileInfo) ([]*vectordb.Document, error) {
	// Read file content
	content, err := os.ReadFile(file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)

	// Generate embedding
	embedding, err := p.llmClient.GenerateEmbedding(ctx, text, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert float64 to float32
	vector := make([]float32, len(embedding))
	for i, v := range embedding {
		vector[i] = float32(v)
	}

	// Create metadata
	metadata := make(map[string]interface{})
	if p.options.Metadata != nil {
		for k, v := range p.options.Metadata {
			metadata[k] = v
		}
	}
	metadata["source_file"] = file.Path
	metadata["file_type"] = string(file.Type)
	metadata["file_size"] = file.Size
	metadata["file_hash"] = file.Hash

	doc := &vectordb.Document{
		ID:       uuid.New().String(),
		Text:     text,
		Content:  text,
		Metadata: metadata,
	}

	return []*vectordb.Document{doc}, nil
}

// processJSON processes JSON files
func (p *Processor) processJSON(ctx context.Context, file FileInfo) ([]*vectordb.Document, error) {
	// Read file content
	content, err := os.ReadFile(file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)

	// Generate embedding
	embedding, err := p.llmClient.GenerateEmbedding(ctx, text, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert float64 to float32
	vector := make([]float32, len(embedding))
	for i, v := range embedding {
		vector[i] = float32(v)
	}

	// Create metadata
	metadata := make(map[string]interface{})
	if p.options.Metadata != nil {
		for k, v := range p.options.Metadata {
			metadata[k] = v
		}
	}
	metadata["source_file"] = file.Path
	metadata["file_type"] = string(file.Type)
	metadata["file_hash"] = file.Hash

	doc := &vectordb.Document{
		ID:       uuid.New().String(),
		Text:     text,
		Content:  text,
		Metadata: metadata,
	}

	return []*vectordb.Document{doc}, nil
}

// processYAML processes YAML files
func (p *Processor) processYAML(ctx context.Context, file FileInfo) ([]*vectordb.Document, error) {
	// Read file content
	content, err := os.ReadFile(file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)

	// Generate embedding
	embedding, err := p.llmClient.GenerateEmbedding(ctx, text, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert float64 to float32
	vector := make([]float32, len(embedding))
	for i, v := range embedding {
		vector[i] = float32(v)
	}

	// Create metadata
	metadata := make(map[string]interface{})
	if p.options.Metadata != nil {
		for k, v := range p.options.Metadata {
			metadata[k] = v
		}
	}
	metadata["source_file"] = file.Path
	metadata["file_type"] = string(file.Type)
	metadata["file_hash"] = file.Hash

	doc := &vectordb.Document{
		ID:       uuid.New().String(),
		Text:     text,
		Content:  text,
		Metadata: metadata,
	}

	return []*vectordb.Document{doc}, nil
}

// createDocumentsInBatches creates documents in VDB using batch operations
func (p *Processor) createDocumentsInBatches(ctx context.Context, documents []*vectordb.Document) error {
	batchSize := p.options.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // default
	}

	totalBatches := (len(documents) + batchSize - 1) / batchSize
	p.progress.StartBatching(len(documents), batchSize)

	batchNum := 0
	for i := 0; i < len(documents); i += batchSize {
		batchNum++
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		p.progress.UpdateBatch(batchNum, totalBatches, len(batch))

		if err := p.vdbClient.CreateDocuments(ctx, p.options.Collection, batch); err != nil {
			return fmt.Errorf("failed to create batch at index %d: %w", i, err)
		}
	}

	p.progress.FinishBatching()
	return nil
}

// loadState loads ingestion state from file
func loadState(stateFile string) (*IngestState, error) {
	// TODO: Implement state file loading (JSON format)
	return nil, nil
}

// saveState saves ingestion state to file
func saveState(stateFile string, state *IngestState) error {
	// TODO: Implement state file saving (JSON format)
	return nil
}
