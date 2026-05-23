// Package server implements the MCP (Model Context Protocol) server
// for MATLAB integration, handling tool registration and request routing.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Version is the current server version.
const Version = "0.1.0"

// ServerInfo contains metadata about the MCP server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool that can be invoked by clients.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema defines the JSON schema for a tool's input parameters.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property describes a single property in an input schema.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ToolHandler is a function that handles a tool invocation.
type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// Server is the core MCP server instance.
type Server struct {
	info     ServerInfo
	tools    map[string]Tool
	handlers map[string]ToolHandler
	logger   *log.Logger
}

// New creates a new Server with the given name.
func New(name string) *Server {
	return &Server{
		info: ServerInfo{
			Name:    name,
			Version: Version,
		},
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
		logger:   log.New(os.Stderr, "[matlab-mcp] ", log.LstdFlags),
	}
}

// RegisterTool registers a tool and its handler with the server.
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name must not be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler must not be nil for tool %q", tool.Name)
	}
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
	s.logger.Printf("registered tool: %s", tool.Name)
	return nil
}

// ListTools returns all registered tools.
func (s *Server) ListTools() []Tool {
	tools := make([]Tool, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, t)
	}
	return tools
}

// CallTool invokes a registered tool by name with the given parameters.
func (s *Server) CallTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	handler, ok := s.handlers[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %q", name)
	}
	s.logger.Printf("calling tool: %s", name)
	return handler(ctx, params)
}

// Info returns the server metadata.
func (s *Server) Info() ServerInfo {
	return s.info
}

// MarshalToolsJSON serializes the registered tools to JSON.
func (s *Server) MarshalToolsJSON() ([]byte, error) {
	return json.Marshal(s.ListTools())
}
