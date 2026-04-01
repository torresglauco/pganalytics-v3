package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

type QueryAnalysisResponse struct {
	QueryID       int64    `json:"query_id"`
	Query         string   `json:"query"`
	ExecutionTime float64  `json:"execution_time_ms"`
	Recommendations []string `json:"recommendations"`
}

type IndexRecommendation struct {
	TableName   string   `json:"table_name"`
	Columns     []string `json:"columns"`
	Impact      float64  `json:"impact_score"`
	Size        int64    `json:"estimated_size"`
	CreationSQL string   `json:"creation_sql"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Do(method, endpoint string, body interface{}) ([]byte, error) {
	url := c.BaseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) GetQueryAnalysis(queryID int64) (*QueryAnalysisResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/queries/%d/analysis", queryID)

	resp, err := c.Do("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result QueryAnalysisResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) SuggestIndexes(tableName string) ([]IndexRecommendation, error) {
	endpoint := fmt.Sprintf("/api/v1/indexes/suggest?table=%s", tableName)

	resp, err := c.Do("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var results []IndexRecommendation
	if err := json.Unmarshal(resp, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return results, nil
}
