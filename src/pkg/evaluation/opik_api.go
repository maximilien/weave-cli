// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/llm"
)

const defaultOpikAPIBaseURL = "https://www.comet.com/opik/api"

type OpikAPIClient struct {
	baseURL     string
	apiKey      string
	workspace   string
	projectName string
	httpClient  *http.Client
}

type OpikDatasetSummary struct {
	ID          string
	Name        string
	VersionID   string
	VersionName string
	ItemCount   int
	URL         string
}

type OpikExperimentSummary struct {
	ID   string
	Name string
	URL  string
}

type OpikSyncResult struct {
	Dataset    OpikDatasetSummary
	Experiment OpikExperimentSummary
}

type OpikTraceRecord struct {
	ID       string
	Name     string
	ThreadID string
}

type opikDatasetResource struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	DatasetItemsCount int    `json:"dataset_items_count"`
	LatestVersion     struct {
		ID          string `json:"id"`
		VersionName string `json:"version_name"`
	} `json:"latest_version"`
}

type opikDatasetItemsPage struct {
	Content []struct {
		DatasetItemID string                 `json:"dataset_item_id"`
		Data          map[string]interface{} `json:"data"`
	} `json:"content"`
}

func NewOpikAPIClient(config *llm.OpikConfig) (*OpikAPIClient, error) {
	if config == nil || config.APIKey == "" {
		return nil, fmt.Errorf("Opik API key is required")
	}

	baseURL := strings.TrimRight(os.Getenv("OPIK_API_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = deriveOpikAPIBaseURL(config.Endpoint)
	}

	return &OpikAPIClient{
		baseURL:     baseURL,
		apiKey:      config.APIKey,
		workspace:   config.Workspace,
		projectName: config.ProjectName,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func UploadDatasetToOpik(ctx context.Context, dataset *Dataset) (*OpikDatasetSummary, error) {
	client, err := NewOpikAPIClient(llm.LoadOpikConfig())
	if err != nil {
		return nil, err
	}
	return client.UploadDataset(ctx, dataset)
}

func SyncEvaluationRunToOpik(ctx context.Context, dataset *Dataset, run *EvaluationRun) (*OpikSyncResult, error) {
	client, err := NewOpikAPIClient(llm.LoadOpikConfig())
	if err != nil {
		return nil, err
	}
	return client.SyncEvaluationRun(ctx, dataset, run)
}

func CreateTraceInOpik(ctx context.Context, traceID, name string, startTime time.Time, input, metadata map[string]interface{}) (*OpikTraceRecord, error) {
	client, err := NewOpikAPIClient(llm.LoadOpikConfig())
	if err != nil {
		return nil, err
	}
	return client.CreateTrace(ctx, traceID, name, startTime, input, metadata)
}

func UpdateTraceInOpik(ctx context.Context, traceID string, endTime time.Time, output, metadata map[string]interface{}, err error) error {
	client, clientErr := NewOpikAPIClient(llm.LoadOpikConfig())
	if clientErr != nil {
		return clientErr
	}
	return client.UpdateTrace(ctx, traceID, endTime, output, metadata, err)
}

func (c *OpikAPIClient) UploadDataset(ctx context.Context, dataset *Dataset) (*OpikDatasetSummary, error) {
	resource, err := c.ensureDataset(ctx, dataset)
	if err != nil {
		return nil, err
	}

	if err := c.upsertDatasetItems(ctx, resource.ID, dataset); err != nil {
		return nil, err
	}

	resource, err = c.getDatasetByName(ctx, dataset.Name)
	if err != nil {
		return nil, err
	}

	return c.makeDatasetSummary(resource), nil
}

func (c *OpikAPIClient) CreateTrace(ctx context.Context, traceID, name string, startTime time.Time, input, metadata map[string]interface{}) (*OpikTraceRecord, error) {
	traceUUID, err := formatTraceIDAsUUID(traceID)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"id":         traceUUID,
		"name":       name,
		"start_time": startTime.UTC().Format(time.RFC3339Nano),
		"input":      input,
		"metadata":   metadata,
		"source":     "sdk",
	}

	if err := c.doJSON(ctx, http.MethodPost, "/v1/private/traces", payload, nil); err != nil {
		return nil, err
	}

	return &OpikTraceRecord{
		ID:   traceUUID,
		Name: name,
	}, nil
}

func (c *OpikAPIClient) UpdateTrace(ctx context.Context, traceID string, endTime time.Time, output, metadata map[string]interface{}, traceErr error) error {
	traceUUID, err := formatTraceIDAsUUID(traceID)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"end_time": endTime.UTC().Format(time.RFC3339Nano),
		"output":   output,
		"metadata": metadata,
	}

	if traceErr != nil {
		payload["error_info"] = map[string]interface{}{
			"exception_type":    "error",
			"exception_message": traceErr.Error(),
		}
	}

	return c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/v1/private/traces/%s", traceUUID), payload, nil)
}

func (c *OpikAPIClient) SyncEvaluationRun(ctx context.Context, dataset *Dataset, run *EvaluationRun) (*OpikSyncResult, error) {
	datasetSummary, err := c.UploadDataset(ctx, dataset)
	if err != nil {
		return nil, err
	}

	datasetItems, err := c.getDatasetItems(ctx, datasetSummary.ID, len(dataset.TestCases)+10)
	if err != nil {
		return nil, err
	}

	itemIDs := make(map[string]string, len(datasetItems))
	for _, item := range datasetItems {
		if id, ok := item.Data["weave_test_case_id"].(string); ok {
			itemIDs[id] = item.DatasetItemID
		}
	}

	experimentUUID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Opik experiment ID: %w", err)
	}
	experimentID := experimentUUID.String()
	experimentName := fmt.Sprintf("%s-%s-%s", run.DatasetName, run.AgentName, run.ID)
	experimentMetadata := map[string]interface{}{
		"run_id":            run.ID,
		"dataset_name":      run.DatasetName,
		"agent_name":        run.AgentName,
		"collection":        run.Collection,
		"pass_rate":         run.Summary.PassRate,
		"avg_accuracy":      run.Summary.AvgAccuracy,
		"avg_citation":      run.Summary.AvgCitation,
		"avg_hallucination": run.Summary.AvgHallucination,
		"avg_context":       run.Summary.AvgContextRelevance,
		"avg_faithfulness":  run.Summary.AvgFaithfulness,
		"provider":          run.Config.Parameters["evaluator_provider"],
		"parameters":        run.Config.Parameters,
	}

	if err := c.createExperiment(ctx, experimentID, dataset.Name, experimentName, experimentMetadata, "running"); err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(run.Results))
	for _, result := range run.Results {
		datasetItemID, ok := itemIDs[result.TestCaseID]
		if !ok {
			continue
		}

		feedbackScores := []map[string]interface{}{
			{"name": "accuracy", "value": result.AccuracyScore, "source": "sdk"},
			{"name": "citation", "value": result.CitationScore, "source": "sdk"},
			{"name": "hallucination", "value": result.HallucinationScore, "source": "sdk"},
			{"name": "context_relevance", "value": result.ContextRelevanceScore, "source": "sdk"},
			{"name": "faithfulness", "value": result.FaithfulnessScore, "source": "sdk"},
		}
		for name, score := range result.CustomScores {
			feedbackScores = append(feedbackScores, map[string]interface{}{
				"name":   name,
				"value":  score,
				"source": "sdk",
			})
		}

		items = append(items, map[string]interface{}{
			"dataset_item_id": datasetItemID,
			"evaluate_task_result": map[string]interface{}{
				"prediction":       result.ActualAnswer,
				"citations":        result.ActualCitations,
				"passed":           result.Passed,
				"response_time_ms": result.ResponseTime,
				"errors":           result.Errors,
				"details":          result.Details,
			},
			"metadata": map[string]interface{}{
				"test_case_id": result.TestCaseID,
				"query":        result.Query,
			},
			"feedback_scores": feedbackScores,
		})
	}

	if err := c.recordExperimentItemsBulk(ctx, experimentID, experimentName, dataset.Name, items); err != nil {
		return nil, err
	}

	if err := c.updateExperimentStatus(ctx, experimentID, "completed"); err != nil {
		return nil, err
	}

	return &OpikSyncResult{
		Dataset: *datasetSummary,
		Experiment: OpikExperimentSummary{
			ID:   experimentID,
			Name: experimentName,
			URL:  c.projectURL(),
		},
	}, nil
}

func (c *OpikAPIClient) ensureDataset(ctx context.Context, dataset *Dataset) (*opikDatasetResource, error) {
	resource, err := c.getDatasetByName(ctx, dataset.Name)
	if err == nil {
		return resource, nil
	}

	payload := map[string]interface{}{
		"name":        dataset.Name,
		"description": dataset.Description,
		"tags":        dataset.Tags,
		"type":        "dataset",
		"visibility":  "private",
	}

	if err := c.doJSON(ctx, http.MethodPost, "/v1/private/datasets", payload, nil); err != nil {
		return nil, err
	}

	return c.getDatasetByName(ctx, dataset.Name)
}

func (c *OpikAPIClient) getDatasetByName(ctx context.Context, name string) (*opikDatasetResource, error) {
	var resource opikDatasetResource
	err := c.doJSON(ctx, http.MethodPost, "/v1/private/datasets/retrieve", map[string]interface{}{
		"dataset_name": name,
	}, &resource)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (c *OpikAPIClient) upsertDatasetItems(ctx context.Context, datasetID string, dataset *Dataset) error {
	batchGroupID := uuid.NewString()
	items := make([]map[string]interface{}, 0, len(dataset.TestCases))
	existingItems := make(map[string]string)

	if datasetID != "" {
		if currentItems, err := c.getDatasetItems(ctx, datasetID, len(dataset.TestCases)+50); err == nil {
			for _, item := range currentItems {
				if id, ok := item.Data["weave_test_case_id"].(string); ok {
					existingItems[id] = item.DatasetItemID
				}
			}
		}
	}

	for _, tc := range dataset.TestCases {
		itemID := existingItems[tc.ID]
		if itemID == "" {
			generated, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate dataset item ID: %w", err)
			}
			itemID = generated.String()
		}

		items = append(items, map[string]interface{}{
			"id":     itemID,
			"source": "manual",
			"data": map[string]interface{}{
				"weave_test_case_id":  tc.ID,
				"query":               tc.Query,
				"expected_answer":     tc.ExpectedAnswer,
				"expected_citations":  tc.ExpectedCitations,
				"required_concepts":   tc.RequiredConcepts,
				"retrieved_context":   tc.RetrievedContext,
				"collection":          tc.Collection,
				"must_cite":           tc.MustCite,
				"min_relevance_score": tc.MinRelevanceScore,
				"description":         tc.Description,
				"tags":                tc.Tags,
			},
		})
	}

	return c.doJSON(ctx, http.MethodPut, "/v1/private/datasets/items", map[string]interface{}{
		"dataset_name":   dataset.Name,
		"batch_group_id": batchGroupID,
		"items":          items,
	}, nil)
}

func (c *OpikAPIClient) getDatasetItems(ctx context.Context, datasetID string, limit int) ([]struct {
	DatasetItemID string                 `json:"dataset_item_id"`
	Data          map[string]interface{} `json:"data"`
}, error) {
	u := fmt.Sprintf("/v1/private/datasets/%s/items?page=1&size=%d", datasetID, limit)
	var page opikDatasetItemsPage
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &page); err != nil {
		return nil, err
	}
	return page.Content, nil
}

func (c *OpikAPIClient) createExperiment(ctx context.Context, id, datasetName, name string, metadata map[string]interface{}, status string) error {
	return c.doJSON(ctx, http.MethodPost, "/v1/private/experiments", map[string]interface{}{
		"id":           id,
		"dataset_name": datasetName,
		"name":         name,
		"metadata":     metadata,
		"status":       status,
		"type":         "regular",
	}, nil)
}

func (c *OpikAPIClient) updateExperimentStatus(ctx context.Context, experimentID, status string) error {
	return c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/v1/private/experiments/%s", experimentID), map[string]interface{}{
		"status": status,
	}, nil)
}

func (c *OpikAPIClient) recordExperimentItemsBulk(ctx context.Context, experimentID, experimentName, datasetName string, items []map[string]interface{}) error {
	return c.doJSON(ctx, http.MethodPut, "/v1/private/experiments/items/bulk", map[string]interface{}{
		"experiment_id":   experimentID,
		"experiment_name": experimentName,
		"dataset_name":    datasetName,
		"items":           items,
	}, nil)
}

func (c *OpikAPIClient) doJSON(ctx context.Context, method, path string, payload interface{}, out interface{}) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal Opik payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("failed to create Opik request: %w", err)
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Comet-Workspace", c.workspace)
	req.Header.Set("projectName", c.projectName)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Opik request failed: %w", err)
	}
	defer resp.Body.Close()

	data, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("failed to read Opik response: %w", readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Opik API %s %s failed: HTTP %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("failed to decode Opik response: %w", err)
		}
	}

	return nil
}

func (c *OpikAPIClient) makeDatasetSummary(resource *opikDatasetResource) *OpikDatasetSummary {
	return &OpikDatasetSummary{
		ID:          resource.ID,
		Name:        resource.Name,
		VersionID:   resource.LatestVersion.ID,
		VersionName: resource.LatestVersion.VersionName,
		ItemCount:   resource.DatasetItemsCount,
		URL:         c.projectURL(),
	}
}

func (c *OpikAPIClient) projectURL() string {
	if c.workspace == "" || c.projectName == "" {
		return "https://www.comet.com"
	}
	return fmt.Sprintf("https://www.comet.com/%s/%s", c.workspace, c.projectName)
}

func deriveOpikAPIBaseURL(endpoint string) string {
	if endpoint == "" {
		return defaultOpikAPIBaseURL
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return defaultOpikAPIBaseURL
	}

	base := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	if idx := strings.Index(parsed.Path, "/api/v1/private/otel"); idx >= 0 {
		apiPrefix := parsed.Path[:idx]
		if !strings.HasSuffix(apiPrefix, "/api") {
			apiPrefix = strings.TrimRight(apiPrefix, "/") + "/api"
		}
		return strings.TrimRight(base+apiPrefix, "/")
	}

	return defaultOpikAPIBaseURL
}

func formatTraceIDAsUUID(traceID string) (string, error) {
	cleaned := strings.ReplaceAll(traceID, "-", "")
	if len(cleaned) != 32 {
		return "", fmt.Errorf("invalid trace id length: %s", traceID)
	}

	return fmt.Sprintf("%s-%s-%s-%s-%s",
		cleaned[0:8],
		cleaned[8:12],
		cleaned[12:16],
		cleaned[16:20],
		cleaned[20:32],
	), nil
}
