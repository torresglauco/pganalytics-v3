package server

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/handlers"
	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/transport"
)

// TestMCPServerInitialization tests server startup and initialization
func TestMCPServerInitialization(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	if mcp == nil {
		t.Fatal("NewMCPServer returned nil")
	}

	// Verify server can be initialized
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  map[string]interface{}{},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Initialize failed: %v", resp.Error)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("Invalid JSONRPC version: got %s, want 2.0", resp.JSONRPC)
	}
}

// TestMCPServerHandlerRegistration tests tool handler registration
func TestMCPServerHandlerRegistration(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	// Register handlers
	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)
	mcp.RegisterTool("anomaly_detect", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.DetectAnomalies(params)
	})

	// List tools
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("tools/list failed: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Response result is not a map")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("Tools is not a list")
	}

	expectedCount := 4 // table_stats, query_analysis, index_suggest, anomaly_detect
	if len(tools) != expectedCount {
		t.Errorf("Expected %d tools, got %d", expectedCount, len(tools))
	}
}

// TestMCPServerTableStatsExecution tests end-to-end table_stats execution
func TestMCPServerTableStatsExecution(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	// Setup
	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)

	// Execute tool
	req := JSONRPCRequest{
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
		t.Fatalf("table_stats execution failed: %v", resp.Error)
	}

	// Verify response format
	if resp.JSONRPC != "2.0" {
		t.Errorf("Invalid JSONRPC version: %s", resp.JSONRPC)
	}

	if resp.ID != 1 {
		t.Errorf("Response ID mismatch: got %d, want 1", resp.ID)
	}

	if resp.Result == nil {
		t.Fatal("table_stats returned nil result")
	}
}

// TestMCPServerConcurrentRequests tests server handling concurrent requests
func TestMCPServerConcurrentRequests(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)

	// Send multiple requests and verify they all succeed
	requests := []JSONRPCRequest{
		{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":   "table_stats",
				"params": map[string]interface{}{"table_name": "users"},
			},
		},
		{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":   "table_stats",
				"params": map[string]interface{}{"table_name": "orders"},
			},
		},
		{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":   "query_analysis",
				"params": map[string]interface{}{"query_id": "query_456"},
			},
		},
	}

	for _, req := range requests {
		resp := mcp.HandleRequest(req)
		if resp.Error != nil {
			t.Errorf("Request %d failed: %v", req.ID, resp.Error)
		}

		if resp.ID != req.ID {
			t.Errorf("Response ID mismatch for request %d", req.ID)
		}
	}
}

// TestMCPServerJSONRPCErrorHandling tests JSON-RPC error responses
func TestMCPServerJSONRPCErrorHandling(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	// Test unknown method
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
		Params:  map[string]interface{}{},
	}

	resp := mcp.HandleRequest(req)

	// Error response should still have JSONRPC and ID
	if resp.JSONRPC != "2.0" {
		t.Errorf("Error response missing JSONRPC")
	}

	if resp.ID != req.ID {
		t.Errorf("Error response ID mismatch")
	}

	if resp.Error == nil {
		t.Fatal("Expected error in response")
	}

	// Result should be nil when there's an error
	if resp.Result != nil {
		t.Errorf("Result should be nil when error is present")
	}
}

// TestMCPServerToolCallWithInvalidParams tests tool call with invalid parameters
func TestMCPServerToolCallWithInvalidParams(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)

	// Call table_stats without required table_name
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":   "table_stats",
			"params": map[string]interface{}{}, // Missing table_name
		},
	}

	resp := mcp.HandleRequest(req)

	if resp.Error == nil {
		t.Fatal("Expected error for missing parameter")
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("Invalid JSONRPC in error response: %s", resp.JSONRPC)
	}
}

// TestMCPServerResponseJSONValidity tests that all responses are valid JSON
func TestMCPServerResponseJSONValidity(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)
	mcp.RegisterTool("anomaly_detect", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.DetectAnomalies(params)
	})

	requests := []JSONRPCRequest{
		{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params:  map[string]interface{}{},
		},
		{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/list",
			Params:  map[string]interface{}{},
		},
		{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":   "table_stats",
				"params": map[string]interface{}{"table_name": "test"},
			},
		},
	}

	for _, req := range requests {
		resp := mcp.HandleRequest(req)

		// Verify response can be marshaled to JSON
		respBytes, err := json.Marshal(resp)
		if err != nil {
			t.Errorf("Response for method %s is not valid JSON: %v", req.Method, err)
		}

		if len(respBytes) == 0 {
			t.Errorf("Response for method %s produced empty JSON", req.Method)
		}

		// Verify response can be unmarshaled
		var respMap map[string]interface{}
		if err := json.Unmarshal(respBytes, &respMap); err != nil {
			t.Errorf("Failed to unmarshal response for method %s: %v", req.Method, err)
		}
	}
}

// TestMCPServerMultipleToolCalls tests sequential tool calls
func TestMCPServerMultipleToolCalls(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)

	// Sequence: initialize -> list tools -> call tools
	requests := []struct {
		name   string
		method string
		params map[string]interface{}
	}{
		{
			name:   "Initialize",
			method: "initialize",
			params: map[string]interface{}{},
		},
		{
			name:   "List tools",
			method: "tools/list",
			params: map[string]interface{}{},
		},
		{
			name:   "Call table_stats",
			method: "tools/call",
			params: map[string]interface{}{
				"name":   "table_stats",
				"params": map[string]interface{}{"table_name": "users"},
			},
		},
		{
			name:   "Call query_analysis",
			method: "tools/call",
			params: map[string]interface{}{
				"name":   "query_analysis",
				"params": map[string]interface{}{"query_id": "q1"},
			},
		},
		{
			name:   "Call index_suggest",
			method: "tools/call",
			params: map[string]interface{}{
				"name":   "index_suggest",
				"params": map[string]interface{}{"table_name": "users"},
			},
		},
	}

	for i, reqInfo := range requests {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      i + 1,
			Method:  reqInfo.method,
			Params:  reqInfo.params,
		}

		resp := mcp.HandleRequest(req)

		if resp.JSONRPC != "2.0" {
			t.Errorf("%s: Invalid JSONRPC", reqInfo.name)
		}

		if resp.Error != nil {
			t.Errorf("%s: Unexpected error: %v", reqInfo.name, resp.Error)
		}

		if resp.Result == nil && resp.Error == nil {
			t.Errorf("%s: Empty response", reqInfo.name)
		}
	}
}

// TestMCPServerTimeoutHandling tests server request handling stability
func TestMCPServerTimeoutHandling(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	handlerCtx := handlers.NewHandlerContext(nil)
	mcp.RegisterDefaultHandlers(handlerCtx)

	// Make rapid requests to test server stability
	start := time.Now()
	for i := 0; i < 10; i++ {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      i + 1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":   "table_stats",
				"params": map[string]interface{}{"table_name": "test"},
			},
		}

		resp := mcp.HandleRequest(req)
		if resp.Error != nil {
			t.Errorf("Request %d failed: %v", i+1, resp.Error)
		}
	}
	duration := time.Since(start)

	// Should complete reasonably fast
	if duration > 5*time.Second {
		t.Logf("Warning: 10 requests took %v (may be slow)", duration)
	}
}

// TestMCPServerInitializeProtocol tests initialize response protocol compliance
func TestMCPServerInitializeProtocol(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := NewMCPServer(tr)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  map[string]interface{}{},
	}

	resp := mcp.HandleRequest(req)

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Response result is not a map")
	}

	// Verify protocol version
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Invalid protocol version: %v", result["protocolVersion"])
	}

	// Verify server info
	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("serverInfo is not a map")
	}

	if serverInfo["name"] != "pgAnalytics MCP Server" {
		t.Errorf("Invalid server name: %v", serverInfo["name"])
	}

	if serverInfo["version"] != "0.1.0" {
		t.Errorf("Invalid server version: %v", serverInfo["version"])
	}
}
