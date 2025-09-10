// Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package externalfunctions

import (
	"context"
	"fmt"
	"strings"
)

// detectTransport automatically determines the transport protocol based on the server URL format
func detectTransport(serverURL string) string {
	// WebSocket URLs
	if strings.HasPrefix(serverURL, "ws://") || strings.HasPrefix(serverURL, "wss://") {
		return "websocket"
	}

	// HTTP/SSE URLs
	if strings.HasPrefix(serverURL, "http://") || strings.HasPrefix(serverURL, "https://") {
		return "sse"
	}

	// STDIO local executables or scripts
	// Check for path separators or common executable extensions
	if strings.Contains(serverURL, "/") || strings.Contains(serverURL, "\\") ||
		strings.HasSuffix(serverURL, ".exe") || strings.HasSuffix(serverURL, ".py") ||
		strings.HasSuffix(serverURL, ".js") || strings.HasSuffix(serverURL, ".sh") {
		return "stdio"
	}

	// Default to websocket for unclear cases
	return "websocket"
}

// These functions below enable communication with MCP servers
//
// ListTools retrieves the list of available tools from an MCP server.
//
// Tags:
//   - @displayName: List MCP Tools
//
// Parameters:
//   - serverURL: MCP server URL (e.g., "ws://localhost:3000")
//   - authToken: Optional authentication token (will be sent as Bearer token)
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//
// Returns:
//   - tools: list of tools with their descriptions and parameters
//   - error: error if connection fails
func ListTools(serverURL string, authToken string, transport string) ([]interface{}, error) {
	// Create context for execution time control
	ctx := context.Background()

	// Auto-detect transport if not specified
	if transport == "" {
		transport = detectTransport(serverURL)
	}

	// Create connection configuration
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken, // Add auth token
		Transport: transport,
		Timeout:   30,
	}

	// Connect to MCP server via WebSocket
	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MCP server %s: %v", serverURL, err)
	}
	// Ensure connection closes when done
	defer conn.Close()

	// Send tools list request per MCP protocol
	response, err := sendMCPRequest(ctx, conn, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching tools: %v", err)
	}

	// Extract tools from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if tools, exists := responseMap["tools"]; exists {
			if toolsList, ok := tools.([]interface{}); ok {
				return toolsList, nil
			}
		}
	}

	return []interface{}{}, nil // Return empty list if no tools
}

// CallTool invokes a specific tool on the MCP server with given arguments.
//
// Tags:
//   - @displayName: Call MCP Tool
//
// Parameters:
//   - serverURL: MCP server URL
//   - authToken: Optional authentication token
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//   - toolName: name of the tool to call
//   - arguments: arguments for the tool as a map (e.g., {"path": "/tmp/file.txt"})
//
// Returns:
//   - result: tool execution result
//   - error: error if tool doesn't exist or execution fails
func CallTool(serverURL string, authToken string, transport string, toolName string, arguments map[string]interface{}) (interface{}, error) {
	ctx := context.Background()

	// Auto-detect transport if not specified
	if transport == "" {
		transport = detectTransport(serverURL)
	}

	// Create connection configuration
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: transport,
		Timeout:   30,
	}

	// Connect to server
	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MCP server %s: %v", serverURL, err)
	}
	defer conn.Close()

	// Prepare tool call request
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	// Send request and return result
	response, err := sendMCPRequest(ctx, conn, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("error calling tool %s: %v", toolName, err)
	}

	// Extract result from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if content, exists := responseMap["content"]; exists {
			return content, nil
		}
	}

	return response, nil
}

// ListResources retrieves the list of available resources from an MCP server.
//
// Tags:
//   - @displayName: List MCP Resources
//
// Parameters:
//   - serverURL: MCP server URL
//   - authToken: Optional authentication token
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//
// Returns:
//   - resources: list of available resources with their URIs
//   - error: error if problem occurs
func ListResources(serverURL string, authToken string, transport string) ([]interface{}, error) {
	ctx := context.Background()

	// Auto-detect transport if not specified
	if transport == "" {
		transport = detectTransport(serverURL)
	}

	// Create connection configuration
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: transport,
		Timeout:   30,
	}

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MCP server %s: %v", serverURL, err)
	}
	defer conn.Close()

	// Send resources list request
	response, err := sendMCPRequest(ctx, conn, "resources/list", nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching resources: %v", err)
	}

	// Extract resources from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if resources, exists := responseMap["resources"]; exists {
			if resourcesList, ok := resources.([]interface{}); ok {
				return resourcesList, nil
			}
		}
	}

	return []interface{}{}, nil
}

// ReadResource reads the content of a specific resource from the MCP server.
//
// Tags:
//   - @displayName: Read MCP Resource
//
// Parameters:
//   - serverURL: MCP server URL
//   - authToken: Optional authentication token
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//   - uri: URI of the resource to read (e.g., "file:///path/to/file")
//
// Returns:
//   - content: resource content
//   - error: error if resource doesn't exist or cannot be read
func ReadResource(serverURL string, authToken string, transport string, uri string) (interface{}, error) {
	ctx := context.Background()

	// Auto-detect transport if not specified
	if transport == "" {
		transport = detectTransport(serverURL)
	}

	// Create connection configuration
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: transport,
		Timeout:   30,
	}

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MCP server %s: %v", serverURL, err)
	}
	defer conn.Close()

	// Prepare resource read request
	params := map[string]interface{}{
		"uri": uri,
	}

	// Send request and return content
	response, err := sendMCPRequest(ctx, conn, "resources/read", params)
	if err != nil {
		return nil, fmt.Errorf("error reading resource %s: %v", uri, err)
	}

	// Extract content from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if contents, exists := responseMap["contents"]; exists {
			return contents, nil
		}
	}

	return response, nil
}

// ListPrompts retrieves the list of available prompt templates from an MCP server.
//
// Tags:
//   - @displayName: List MCP Prompts
//
// Parameters:
//   - serverURL: MCP server URL
//   - authToken: Optional authentication token
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//
// Returns:
//   - prompts: list of available prompt templates with their descriptions
//   - error: error if problem occurs
func ListPrompts(serverURL string, authToken string, transport string) ([]interface{}, error) {
	ctx := context.Background()

	// Auto-detect transport if not specified
	if transport == "" {
		transport = detectTransport(serverURL)
	}

	// Create connection configuration
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: transport,
		Timeout:   30,
	}

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MCP server %s: %v", serverURL, err)
	}
	defer conn.Close()

	// Send prompts list request per MCP protocol
	response, err := sendMCPRequest(ctx, conn, "prompts/list", nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching prompts: %v", err)
	}

	// Extract prompts from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if prompts, exists := responseMap["prompts"]; exists {
			if promptsList, ok := prompts.([]interface{}); ok {
				return promptsList, nil
			}
		}
	}

	return []interface{}{}, nil
}

// GetPrompt retrieves and fills a specific prompt template with given arguments.
//
// Tags:
//   - @displayName: Get MCP Prompt
//
// Parameters:
//   - serverURL: MCP server URL
//   - authToken: Optional authentication token
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//   - promptName: name of the prompt template to use
//   - arguments: arguments to fill the template
//
// Returns:
//   - prompt: filled prompt ready for use
//   - error: error if prompt doesn't exist or cannot be filled
func GetPrompt(serverURL string, authToken string, transport string, promptName string, arguments map[string]interface{}) (interface{}, error) {
	ctx := context.Background()

	// Auto-detect transport if not specified
	if transport == "" {
		transport = detectTransport(serverURL)
	}

	// Create connection configuration
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: transport,
		Timeout:   30,
	}

	// Connect to server
	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MCP server %s: %v", serverURL, err)
	}
	defer conn.Close()

	// Prepare prompt get request
	params := map[string]interface{}{
		"name": promptName,
	}

	// Add arguments if they exist
	if len(arguments) > 0 {
		params["arguments"] = arguments
	}

	// Send request and return result
	response, err := sendMCPRequest(ctx, conn, "prompts/get", params)
	if err != nil {
		return nil, fmt.Errorf("error fetching prompt %s: %v", promptName, err)
	}

	// Return full response which may contain messages array or other format
	return response, nil
}

// ListAll retrieves all available tools, resources, and prompt templates from an MCP server.
//
// Tags:
//   - @displayName: List All MCP Items
//
// Parameters:
//   - serverURL: MCP server URL
//   - authToken: Optional authentication token
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//
// Returns:
//   - result: map with keys "tools", "resources", and "prompts"
//   - error: error if problem occurs
func ListAll(serverURL string, authToken string, transport string) (map[string]interface{}, error) {
	// Reuse existing functions

	// Fetch everything using existing functions
	tools, errTools := ListTools(serverURL, authToken, transport)
	resources, errResources := ListResources(serverURL, authToken, transport)
	prompts, errPrompts := ListPrompts(serverURL, authToken, transport)

	// If anything fails, return error
	if errTools != nil {
		return nil, fmt.Errorf("error fetching tools: %v", errTools)
	}
	if errResources != nil {
		return nil, fmt.Errorf("error fetching resources: %v", errResources)
	}
	if errPrompts != nil {
		return nil, fmt.Errorf("error fetching prompts: %v", errPrompts)
	}

	// Return everything in one map
	return map[string]interface{}{
		"tools":     tools,
		"resources": resources,
		"prompts":   prompts,
	}, nil
}

// HealthCheck verifies if an MCP server is available and functional.
//
// Tags:
//   - @displayName: MCP Health Check
//
// Parameters:
//   - serverURL: MCP server URL to check
//   - authToken: Optional authentication token
//   - transport: Transport protocol ("websocket", "sse", "stdio") - auto-detected if empty
//
// Returns:
//   - available: true if server is available, false otherwise
//   - error: error if problem occurs during check
func HealthCheck(serverURL string, authToken string, transport string) (bool, error) {
	ctx := context.Background()

	// Auto-detect transport if not specified
	if transport == "" {
		transport = detectTransport(serverURL)
	}

	// Create connection configuration
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: transport,
		Timeout:   10, // Shorter timeout for health check
	}

	// Try to connect to server
	conn, err := connectToMCP(ctx, config)
	if err != nil {
		// Server not available, but that's not an "error" - just return false
		return false, nil
	}

	// If connection succeeded, server is available
	defer conn.Close()

	// Optionally: could send a ping or test request
	// Successful connection means server is healthy

	return true, nil
}

// DiscoverServer performs auto-discovery on an MCP server to determine its capabilities and requirements.
//
// Tags:
//   - @displayName: Discover MCP Server
//
// Parameters:
//   - serverURL: MCP server URL to discover
//
// Returns:
//   - discovery: DiscoverServerResponse containing server information
//   - error: error if discovery fails completely
func DiscoverServer(serverURL string) (*DiscoverServerResponse, error) {
	// Auto-detect the most likely transport
	transport := detectTransport(serverURL)

	// Initialize result with defaults
	result := &DiscoverServerResponse{
		ServerURL:            serverURL,
		Status:               "checking",
		RequiresAuth:         false,
		AvailableTransports:  []string{transport},
		HasTools:             false,
		HasResources:         false,
		HasPrompts:           false,
		RecommendedTimeout:   30,
		RecommendedTransport: transport,
	}

	// Quick health check to see if server is available
	available, err := HealthCheck(serverURL, "", transport)

	// Check if authentication is required
	if err != nil && (strings.Contains(err.Error(), "401") ||
		strings.Contains(err.Error(), "403") ||
		strings.Contains(err.Error(), "authentication")) {
		result.Status = "requires_auth"
		result.RequiresAuth = true
		result.Error = "Authentication required"
		return result, nil
	}

	// If not available
	if !available {
		// STDIO - can't test without running
		if transport == "stdio" {
			result.Status = "possible_stdio"
			result.Note = "This appears to be a local executable. Use STDIO transport."
			return result, nil
		}

		result.Status = "unavailable"
		result.Error = "Server not reachable"
		return result, nil
	}

	// Get capabilities
	result.Status = "connected"

	// Use auto-detection for all capability checks
	tools, _ := ListTools(serverURL, "", "")
	resources, _ := ListResources(serverURL, "", "")
	prompts, _ := ListPrompts(serverURL, "", "")

	// Update capability information
	result.HasTools = len(tools) > 0
	result.ToolsCount = len(tools)
	result.HasResources = len(resources) > 0
	result.ResourcesCount = len(resources)
	result.HasPrompts = len(prompts) > 0
	result.PromptsCount = len(prompts)

	return result, nil
}

// Functions connectToMCP and sendMCPRequest are defined in privatefunctions.go
// They handle connection and JSON-RPC communication with MCP server across all transports
