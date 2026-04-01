package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/transport"
)

type MCPServer struct {
	transport   *transport.StdioTransport
	mu          sync.Mutex
	tools       map[string]ToolHandler
	resources   map[string]ResourceHandler
	initialized bool
}

type ToolHandler func(params map[string]interface{}) (interface{}, error)
type ResourceHandler func() (interface{}, error)

type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func NewMCPServer(t *transport.StdioTransport) *MCPServer {
	return &MCPServer{
		transport: t,
		tools:     make(map[string]ToolHandler),
		resources: make(map[string]ResourceHandler),
	}
}

func (s *MCPServer) RegisterTool(name string, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[name] = handler
}

func (s *MCPServer) RegisterResource(name string, handler ResourceHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[name] = handler
}

func (s *MCPServer) Initialize() error {
	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()
	return nil
}

func (s *MCPServer) HandleRequest(req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(req)
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   "Unknown method: " + req.Method,
		}
	}
}

func (s *MCPServer) handleInitialize(req JSONRPCRequest) JSONRPCResponse {
	s.Initialize()
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "pgAnalytics MCP Server",
			"version": "0.1.0",
		},
	}
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *MCPServer) handleToolsList(req JSONRPCRequest) JSONRPCResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	tools := make([]map[string]interface{}, 0)
	for name := range s.tools {
		tools = append(tools, map[string]interface{}{
			"name":        name,
			"description": "pgAnalytics tool: " + name,
		})
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *MCPServer) handleToolCall(req JSONRPCRequest) JSONRPCResponse {
	toolName, ok := req.Params["name"].(string)
	if !ok {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   "Missing tool name",
		}
	}

	s.mu.Lock()
	handler, exists := s.tools[toolName]
	s.mu.Unlock()

	if !exists {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   "Tool not found: " + toolName,
		}
	}

	params, ok := req.Params["params"].(map[string]interface{})
	if !ok {
		params = make(map[string]interface{})
	}

	result, err := handler(params)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   err.Error(),
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *MCPServer) Run() error {
	defer s.transport.Close()
	decoder := json.NewDecoder(s.transport.GetReader())
	for {
		var req JSONRPCRequest
		err := decoder.Decode(&req)
		if err != nil {
			log.Printf("Decode error: %v", err)
			break
		}

		resp := s.HandleRequest(req)
		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Failed to marshal response: %v", err)
			continue
		}
		if err := s.transport.Write(respBytes); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
	return nil
}
