// Package main implements a mock MCP server for testing GatewayOps.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// MCP Request/Response types
type MCPToolCallRequest struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

type MCPToolCallResponse struct {
	Result interface{} `json:"result"`
}

type MCPToolsListResponse struct {
	Tools []MCPTool `json:"tools"`
}

type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type MCPResourcesListResponse struct {
	Resources []MCPResource `json:"resources"`
}

type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

type MCPResourceReadRequest struct {
	URI string `json:"uri"`
}

type MCPResourceReadResponse struct {
	Contents []MCPResourceContent `json:"contents"`
}

type MCPResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

// Mock tools
var mockTools = []MCPTool{
	{
		Name:        "read_file",
		Description: "Read the contents of a file",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Path to the file"}
			},
			"required": ["path"]
		}`),
	},
	{
		Name:        "write_file",
		Description: "Write contents to a file",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Path to the file"},
				"content": {"type": "string", "description": "Content to write"}
			},
			"required": ["path", "content"]
		}`),
	},
	{
		Name:        "list_directory",
		Description: "List contents of a directory",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Path to the directory"}
			},
			"required": ["path"]
		}`),
	},
	{
		Name:        "search",
		Description: "Search for text in files",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {"type": "string", "description": "Search query"},
				"path": {"type": "string", "description": "Directory to search in"}
			},
			"required": ["query"]
		}`),
	},
	{
		Name:        "execute_command",
		Description: "Execute a shell command",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"command": {"type": "string", "description": "Command to execute"}
			},
			"required": ["command"]
		}`),
	},
}

// Mock resources
var mockResources = []MCPResource{
	{
		URI:         "file:///workspace",
		Name:        "Workspace",
		Description: "Project workspace directory",
		MimeType:    "inode/directory",
	},
	{
		URI:         "file:///workspace/README.md",
		Name:        "README",
		Description: "Project documentation",
		MimeType:    "text/markdown",
	},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// MCP Tools endpoints
	mux.HandleFunc("/tools/list", handleToolsList)
	mux.HandleFunc("/tools/call", handleToolsCall)

	// MCP Resources endpoints
	mux.HandleFunc("/resources/list", handleResourcesList)
	mux.HandleFunc("/resources/read", handleResourcesRead)

	// MCP Prompts endpoints
	mux.HandleFunc("/prompts/list", handlePromptsList)
	mux.HandleFunc("/prompts/get", handlePromptsGet)

	log.Printf("Mock MCP server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPToolsListResponse{Tools: mockTools})
}

func handleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Simulate some latency
	time.Sleep(50 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")

	// Handle different tools
	switch req.Tool {
	case "read_file":
		path, _ := req.Arguments["path"].(string)
		json.NewEncoder(w).Encode(MCPToolCallResponse{
			Result: map[string]interface{}{
				"content": fmt.Sprintf("# Mock file content for %s\n\nThis is simulated file content from the mock MCP server.", path),
				"type":    "text",
			},
		})

	case "write_file":
		path, _ := req.Arguments["path"].(string)
		json.NewEncoder(w).Encode(MCPToolCallResponse{
			Result: map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Successfully wrote to %s", path),
			},
		})

	case "list_directory":
		json.NewEncoder(w).Encode(MCPToolCallResponse{
			Result: map[string]interface{}{
				"entries": []map[string]interface{}{
					{"name": "src", "type": "directory"},
					{"name": "README.md", "type": "file"},
					{"name": "package.json", "type": "file"},
					{"name": "go.mod", "type": "file"},
				},
			},
		})

	case "search":
		query, _ := req.Arguments["query"].(string)
		json.NewEncoder(w).Encode(MCPToolCallResponse{
			Result: map[string]interface{}{
				"matches": []map[string]interface{}{
					{"file": "src/main.go", "line": 42, "content": fmt.Sprintf("// %s found here", query)},
					{"file": "README.md", "line": 10, "content": fmt.Sprintf("Documentation about %s", query)},
				},
			},
		})

	case "execute_command":
		command, _ := req.Arguments["command"].(string)
		json.NewEncoder(w).Encode(MCPToolCallResponse{
			Result: map[string]interface{}{
				"stdout":    fmt.Sprintf("Mock output for: %s", command),
				"stderr":    "",
				"exit_code": 0,
			},
		})

	default:
		http.Error(w, fmt.Sprintf("Unknown tool: %s", req.Tool), http.StatusBadRequest)
	}
}

func handleResourcesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResourcesListResponse{Resources: mockResources})
}

func handleResourcesRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPResourceReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResourceReadResponse{
		Contents: []MCPResourceContent{
			{
				URI:      req.URI,
				MimeType: "text/plain",
				Text:     fmt.Sprintf("Mock content for resource: %s", req.URI),
			},
		},
	})
}

func handlePromptsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prompts": []map[string]interface{}{
			{
				"name":        "code_review",
				"description": "Generate a code review prompt",
				"arguments": []map[string]interface{}{
					{"name": "language", "description": "Programming language", "required": true},
				},
			},
		},
	})
}

func handlePromptsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"description": "Code review prompt",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": map[string]interface{}{
					"type": "text",
					"text": "Please review the following code for best practices and potential issues.",
				},
			},
		},
	})
}
