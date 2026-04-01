package mcp_test

import (
	"bytes"
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/server"
	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/transport"
)

func TestMCPInitialize(t *testing.T) {
	input := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	reader := bytes.NewReader(input)
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  map[string]interface{}{},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Initialize failed: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response format")
	}

	if _, exists := result["protocolVersion"]; !exists {
		t.Fatal("Missing protocolVersion in response")
	}
}

func TestMCPToolsListAfterRegistration(t *testing.T) {
	input := []byte{}
	reader := bytes.NewReader(input)
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	// Register a tool
	mcp.RegisterTool("test_tool", func(params map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	})

	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("tools/list failed: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Invalid response format, result: %v", resp.Result)
	}

	tools, exists := result["tools"].([]interface{})
	if !exists {
		t.Fatalf("Tools key not found in result: %v", result)
	}
	if len(tools) == 0 {
		t.Fatalf("No tools returned, tools: %v, result: %v", tools, result)
	}
}

func TestMCPToolCall(t *testing.T) {
	input := []byte{}
	reader := bytes.NewReader(input)
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	mcp := server.NewMCPServer(tr)

	mcp.RegisterTool("echo", func(params map[string]interface{}) (interface{}, error) {
		return params, nil
	})

	req := server.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "echo",
			"params": map[string]interface{}{
				"message": "hello",
			},
		},
	}

	resp := mcp.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("tools/call failed: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response format")
	}

	msg, exists := result["message"].(string)
	if !exists || msg != "hello" {
		t.Fatal("Unexpected tool response")
	}
}
