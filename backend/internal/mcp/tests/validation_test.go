package mcp_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/handlers"
	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/server"
	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/transport"
)

// TestMCPProtocolCompliance validates JSON-RPC 2.0 format and protocol version
func TestMCPProtocolCompliance(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected map[string]interface{}
	}{
		{
			name:   "Initialize Protocol Version",
			method: "initialize",
			expected: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]interface{}{
					"name":    "pgAnalytics MCP Server",
					"version": "0.1.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte{})
			writer := &bytes.Buffer{}

			tr := transport.NewStdioTransport(reader, writer)
			mcp := server.NewMCPServer(tr)

			req := server.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  tt.method,
				Params:  map[string]interface{}{},
			}

			resp := mcp.HandleRequest(req)

			// Verify JSON-RPC 2.0 format
			if resp.JSONRPC != "2.0" {
				t.Errorf("Invalid JSONRPC version: got %s, want 2.0", resp.JSONRPC)
			}

			// Verify ID matches
			if resp.ID != req.ID {
				t.Errorf("Response ID mismatch: got %d, want %d", resp.ID, req.ID)
			}

			// Verify protocol version
			result, ok := resp.Result.(map[string]interface{})
			if !ok {
				t.Fatal("Response result is not a map")
			}

			if _, exists := result["protocolVersion"]; !exists {
				t.Fatal("Missing protocolVersion in response")
			}

			if result["protocolVersion"] != "2024-11-05" {
				t.Errorf("Invalid protocol version: got %v, want 2024-11-05", result["protocolVersion"])
			}
		})
	}
}

// TestToolValidation validates all 4 tools are registered with proper schemas
func TestToolValidation(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	// Register all 4 tools
	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)
	mcp.RegisterTool("anomaly_detect", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.DetectAnomalies(params)
	})

	expectedTools := []string{"table_stats", "query_analysis", "index_suggest", "anomaly_detect"}

	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	resp := mcp.HandleRequest(req)

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Response result is not a map")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("Tools field is not a list")
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	registeredTools := make(map[string]bool)
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			t.Fatal("Tool entry is not a map")
		}

		name, ok := toolMap["name"].(string)
		if !ok {
			t.Fatal("Tool name is not a string")
		}

		registeredTools[name] = true

		// Verify all registered tools have description
		if _, exists := toolMap["description"]; !exists {
			t.Errorf("Tool %s missing description", name)
		}
	}

	for _, expected := range expectedTools {
		if !registeredTools[expected] {
			t.Errorf("Expected tool %s not registered", expected)
		}
	}
}

// TestTableStatsToolSchema validates table_stats tool parameters
func TestTableStatsToolSchema(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterTool("table_stats", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.TableStats(params)
	})

	// Test valid parameters
	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "table_stats",
			"params": map[string]interface{}{
				"table_name": "users",
			},
		},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("table_stats call failed: %v", resp.Error)
	}

	// Verify response structure - result should be a list (may be []TableStats or []interface{})
	// For JSON marshaling, it will be []interface{} with maps inside
	result := resp.Result
	if result == nil {
		t.Fatal("table_stats returned nil result")
	}

	// The result will be interface{} containing actual data
	// We need to handle both raw data and JSON-marshaled data
	var statMap map[string]interface{}
	switch v := result.(type) {
	case []handlers.TableStats:
		if len(v) == 0 {
			t.Fatal("table_stats returned empty list")
		}
		// Convert to map for validation
		data, _ := json.Marshal(v[0])
		json.Unmarshal(data, &statMap)
	case []interface{}:
		if len(v) == 0 {
			t.Fatal("table_stats returned empty list")
		}
		var ok bool
		statMap, ok = v[0].(map[string]interface{})
		if !ok {
			t.Fatal("TableStats entry is not a map")
		}
	default:
		t.Fatalf("table_stats result has unexpected type: %T", result)
	}

	requiredFields := []string{"table_name", "row_count", "size_bytes", "index_count"}
	for _, field := range requiredFields {
		if _, exists := statMap[field]; !exists {
			t.Errorf("TableStats missing required field: %s", field)
		}
	}
}

// TestQueryAnalysisToolSchema validates query_analysis tool parameters
func TestQueryAnalysisToolSchema(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterTool("query_analysis", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.QueryAnalysis(params)
	})

	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "query_analysis",
			"params": map[string]interface{}{
				"query_id": "query_123",
			},
		},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("query_analysis call failed: %v", resp.Error)
	}

	// Verify response structure
	var result map[string]interface{}
	switch v := resp.Result.(type) {
	case handlers.QueryAnalysisResult:
		// Convert struct to map for validation
		data, _ := json.Marshal(v)
		json.Unmarshal(data, &result)
	case map[string]interface{}:
		result = v
	default:
		t.Fatalf("query_analysis result has unexpected type: %T", resp.Result)
	}

	requiredFields := []string{"query_id", "query_text", "execution_count", "mean_time_ms", "max_time_ms", "total_time_ms", "anomalies", "recommendations"}
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			t.Errorf("QueryAnalysisResult missing required field: %s", field)
		}
	}

	// Verify arrays are properly typed
	anomalies, ok := result["anomalies"].([]interface{})
	if !ok {
		t.Fatal("anomalies field is not a list")
	}

	recommendations, ok := result["recommendations"].([]interface{})
	if !ok {
		t.Fatal("recommendations field is not a list")
	}

	if len(anomalies) == 0 || len(recommendations) == 0 {
		t.Log("Anomalies and recommendations may be empty in mock data")
	}
}

// TestIndexSuggestToolSchema validates index_suggest tool parameters
func TestIndexSuggestToolSchema(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterTool("index_suggest", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.IndexSuggest(params)
	})

	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "index_suggest",
			"params": map[string]interface{}{
				"table_name": "users",
			},
		},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("index_suggest call failed: %v", resp.Error)
	}

	// Verify response structure
	var suggestions interface{}
	switch v := resp.Result.(type) {
	case []handlers.IndexSuggestion:
		if len(v) == 0 {
			t.Fatal("index_suggest returned empty list")
		}
		suggestions = v
	case []interface{}:
		if len(v) == 0 {
			t.Fatal("index_suggest returned empty list")
		}
		suggestions = v
	default:
		t.Fatalf("index_suggest result has unexpected type: %T", resp.Result)
	}

	// Verify IndexSuggestion structure
	var suggestion map[string]interface{}
	switch v := suggestions.(type) {
	case []handlers.IndexSuggestion:
		data, _ := json.Marshal(v[0])
		json.Unmarshal(data, &suggestion)
	case []interface{}:
		var ok bool
		suggestion, ok = v[0].(map[string]interface{})
		if !ok {
			t.Fatal("IndexSuggestion entry is not a map")
		}
	}

	requiredFields := []string{"table_name", "columns", "estimated_gain_percent", "priority"}
	for _, field := range requiredFields {
		if _, exists := suggestion[field]; !exists {
			t.Errorf("IndexSuggestion missing required field: %s", field)
		}
	}

	// Verify priority is valid
	priority, ok := suggestion["priority"].(string)
	if !ok {
		t.Fatal("priority is not a string")
	}

	validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
	if !validPriorities[priority] {
		t.Errorf("Invalid priority value: %s", priority)
	}
}

// TestAnomalyDetectToolSchema validates anomaly_detect tool parameters
func TestAnomalyDetectToolSchema(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterTool("anomaly_detect", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.DetectAnomalies(params)
	})

	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "anomaly_detect",
			"params": map[string]interface{}{
				"table_name": "users",
			},
		},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("anomaly_detect call failed: %v", resp.Error)
	}

	// Verify response structure
	var alerts interface{}
	switch v := resp.Result.(type) {
	case []handlers.AnomalyAlert:
		if len(v) == 0 {
			t.Log("No anomalies detected in mock data (expected)")
			return
		}
		alerts = v
	case []interface{}:
		if len(v) == 0 {
			t.Log("No anomalies detected in mock data (expected)")
			return
		}
		alerts = v
	default:
		t.Fatalf("anomaly_detect result has unexpected type: %T", resp.Result)
	}

	// Verify AnomalyAlert structure
	var alert map[string]interface{}
	switch v := alerts.(type) {
	case []handlers.AnomalyAlert:
		data, _ := json.Marshal(v[0])
		json.Unmarshal(data, &alert)
	case []interface{}:
		var ok bool
		alert, ok = v[0].(map[string]interface{})
		if !ok {
			t.Fatal("AnomalyAlert entry is not a map")
		}
	}

	requiredFields := []string{"metric_name", "current_value", "baseline_value", "z_score", "severity", "timestamp", "description"}
	for _, field := range requiredFields {
		if _, exists := alert[field]; !exists {
			t.Errorf("AnomalyAlert missing required field: %s", field)
		}
	}

	// Verify severity is valid
	severity, ok := alert["severity"].(string)
	if !ok {
		t.Fatal("severity is not a string")
	}

	validSeverities := map[string]bool{"high": true, "medium": true, "low": true}
	if !validSeverities[severity] {
		t.Errorf("Invalid severity value: %s", severity)
	}
}

// TestErrorValidation validates error responses follow JSON-RPC 2.0 format
func TestErrorValidation(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		params        map[string]interface{}
		expectedError bool
		errorContains string
	}{
		{
			name:          "Unknown method",
			method:        "unknown/method",
			expectedError: true,
			errorContains: "Unknown method",
		},
		{
			name:   "Missing tool name",
			method: "tools/call",
			params: map[string]interface{}{
				"params": map[string]interface{}{},
			},
			expectedError: true,
			errorContains: "Missing tool name",
		},
		{
			name:   "Tool not found",
			method: "tools/call",
			params: map[string]interface{}{
				"name": "nonexistent_tool",
			},
			expectedError: true,
			errorContains: "Tool not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte{})
			writer := &bytes.Buffer{}

			tr := transport.NewStdioTransport(reader, writer)
			mcp := server.NewMCPServer(tr)

			req := server.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  tt.method,
				Params:  tt.params,
			}

			resp := mcp.HandleRequest(req)

			// Verify JSON-RPC 2.0 error format
			if tt.expectedError {
				if resp.Error == nil {
					t.Fatal("Expected error but got none")
				}

				errStr, ok := resp.Error.(string)
				if !ok {
					t.Fatal("Error is not a string")
				}

				if !contains(errStr, tt.errorContains) {
					t.Errorf("Error message doesn't contain expected text: got %q, want to contain %q", errStr, tt.errorContains)
				}
			}

			// Verify JSONRPC and ID are still present
			if resp.JSONRPC != "2.0" {
				t.Errorf("Invalid JSONRPC version: got %s, want 2.0", resp.JSONRPC)
			}

			if resp.ID != req.ID {
				t.Errorf("Response ID mismatch: got %d, want %d", resp.ID, req.ID)
			}
		})
	}
}

// TestJSONRPCRequestValidation validates incoming request format
func TestJSONRPCRequestValidation(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		valid    bool
		errorMsg string
	}{
		{
			name:  "Valid request",
			json:  `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			valid: true,
		},
		{
			name:  "Missing id field (notification)",
			json:  `{"jsonrpc":"2.0","method":"initialize","params":{}}`,
			valid: true,
		},
		{
			name:     "Invalid json",
			json:     `{invalid json}`,
			valid:    false,
			errorMsg: "syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req server.JSONRPCRequest
			err := json.Unmarshal([]byte(tt.json), &req)

			if tt.valid && err != nil {
				t.Errorf("Expected valid request but got error: %v", err)
			}

			if !tt.valid && err == nil {
				t.Errorf("Expected invalid request but got none")
			}

			if tt.valid && req.JSONRPC != "2.0" {
				t.Errorf("Invalid JSONRPC version: got %s, want 2.0", req.JSONRPC)
			}
		})
	}
}

// TestMissingRequiredParameters tests tool parameter validation
func TestMissingRequiredParameters(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		params    map[string]interface{}
		shouldErr bool
	}{
		{
			name:      "table_stats with table_name",
			toolName:  "table_stats",
			params:    map[string]interface{}{"table_name": "users"},
			shouldErr: false,
		},
		{
			name:      "table_stats without table_name",
			toolName:  "table_stats",
			params:    map[string]interface{}{},
			shouldErr: true,
		},
		{
			name:      "query_analysis with query_id",
			toolName:  "query_analysis",
			params:    map[string]interface{}{"query_id": "query_1"},
			shouldErr: false,
		},
		{
			name:      "query_analysis without query_id",
			toolName:  "query_analysis",
			params:    map[string]interface{}{},
			shouldErr: true,
		},
		{
			name:      "anomaly_detect with table_name",
			toolName:  "anomaly_detect",
			params:    map[string]interface{}{"table_name": "users"},
			shouldErr: false,
		},
		{
			name:      "anomaly_detect without table_name",
			toolName:  "anomaly_detect",
			params:    map[string]interface{}{},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte{})
			writer := &bytes.Buffer{}

			tr := transport.NewStdioTransport(reader, writer)
			mcp := server.NewMCPServer(tr)

			handlerCtx := handlers.NewHandlerContext(nil)
			mcp.RegisterDefaultHandlers(handlerCtx)
			mcp.RegisterTool("anomaly_detect", func(params map[string]interface{}) (interface{}, error) {
				return handlerCtx.DetectAnomalies(params)
			})

			req := server.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name":   tt.toolName,
					"params": tt.params,
				},
			}

			resp := mcp.HandleRequest(req)

			if tt.shouldErr && resp.Error == nil {
				t.Errorf("Expected error for %s with params %v", tt.toolName, tt.params)
			}

			if !tt.shouldErr && resp.Error != nil {
				t.Errorf("Unexpected error for %s with params %v: %v", tt.toolName, tt.params, resp.Error)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
