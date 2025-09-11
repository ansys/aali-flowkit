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

	"github.com/ansys/aali-sharedtypes/pkg/logging"
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
func ListTools(serverURL string, authToken string, transport string) []interface{} {
	logging.Log.Debugf(&logging.ContextMap{}, "ListTools called with serverURL: %s, transport: %s", serverURL, transport)
	
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
		logging.Log.Errorf(&logging.ContextMap{}, "Unable to connect to MCP server %s: %v", serverURL, err)
		panic(fmt.Sprintf("Unable to connect to MCP server %s: %v", serverURL, err))
	}
	// Ensure connection closes when done
	defer conn.Close()

	// Send tools list request per MCP protocol
	response, err := sendMCPRequest(ctx, conn, "tools/list", nil)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error fetching tools: %v", err)
		panic(fmt.Sprintf("Error fetching tools: %v", err))
	}

	// Extract tools from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if tools, exists := responseMap["tools"]; exists {
			if toolsList, ok := tools.([]interface{}); ok {
				return toolsList
			}
		}
	}

	return []interface{}{} // Return empty list if no tools
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
func CallTool(serverURL string, authToken string, transport string, toolName string, arguments map[string]interface{}) interface{} {
	logging.Log.Debugf(&logging.ContextMap{}, "CallTool called with serverURL: %s, toolName: %s", serverURL, toolName)
	
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
		logging.Log.Errorf(&logging.ContextMap{}, "Unable to connect to MCP server %s: %v", serverURL, err)
		panic(fmt.Sprintf("Unable to connect to MCP server %s: %v", serverURL, err))
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
		logging.Log.Errorf(&logging.ContextMap{}, "Error calling tool %s: %v", toolName, err)
		panic(fmt.Sprintf("Error calling tool %s: %v", toolName, err))
	}

	// Extract result from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if content, exists := responseMap["content"]; exists {
			return content
		}
	}

	return response
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
func ListResources(serverURL string, authToken string, transport string) []interface{} {
	logging.Log.Debugf(&logging.ContextMap{}, "ListResources called with serverURL: %s, transport: %s", serverURL, transport)
	
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
		logging.Log.Errorf(&logging.ContextMap{}, "Unable to connect to MCP server %s: %v", serverURL, err)
		panic(fmt.Sprintf("Unable to connect to MCP server %s: %v", serverURL, err))
	}
	defer conn.Close()

	// Send resources list request
	response, err := sendMCPRequest(ctx, conn, "resources/list", nil)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error fetching resources: %v", err)
		panic(fmt.Sprintf("Error fetching resources: %v", err))
	}

	// Extract resources from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if resources, exists := responseMap["resources"]; exists {
			if resourcesList, ok := resources.([]interface{}); ok {
				return resourcesList
			}
		}
	}

	return []interface{}{}
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
func ReadResource(serverURL string, authToken string, transport string, uri string) interface{} {
	logging.Log.Debugf(&logging.ContextMap{}, "ReadResource called with serverURL: %s, uri: %s", serverURL, uri)
	
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
		logging.Log.Errorf(&logging.ContextMap{}, "Unable to connect to MCP server %s: %v", serverURL, err)
		panic(fmt.Sprintf("Unable to connect to MCP server %s: %v", serverURL, err))
	}
	defer conn.Close()

	// Prepare resource read request
	params := map[string]interface{}{
		"uri": uri,
	}

	// Send request and return content
	response, err := sendMCPRequest(ctx, conn, "resources/read", params)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error reading resource %s: %v", uri, err)
		panic(fmt.Sprintf("Error reading resource %s: %v", uri, err))
	}

	// Extract content from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if contents, exists := responseMap["contents"]; exists {
			return contents
		}
	}

	return response
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
func ListPrompts(serverURL string, authToken string, transport string) []interface{} {
	logging.Log.Debugf(&logging.ContextMap{}, "ListPrompts called with serverURL: %s, transport: %s", serverURL, transport)
	
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
		logging.Log.Errorf(&logging.ContextMap{}, "Unable to connect to MCP server %s: %v", serverURL, err)
		panic(fmt.Sprintf("Unable to connect to MCP server %s: %v", serverURL, err))
	}
	defer conn.Close()

	// Send prompts list request per MCP protocol
	response, err := sendMCPRequest(ctx, conn, "prompts/list", nil)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error fetching prompts: %v", err)
		panic(fmt.Sprintf("Error fetching prompts: %v", err))
	}

	// Extract prompts from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if prompts, exists := responseMap["prompts"]; exists {
			if promptsList, ok := prompts.([]interface{}); ok {
				return promptsList
			}
		}
	}

	return []interface{}{}
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
func GetPrompt(serverURL string, authToken string, transport string, promptName string, arguments map[string]interface{}) interface{} {
	logging.Log.Debugf(&logging.ContextMap{}, "GetPrompt called with serverURL: %s, promptName: %s", serverURL, promptName)
	
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
		logging.Log.Errorf(&logging.ContextMap{}, "Unable to connect to MCP server %s: %v", serverURL, err)
		panic(fmt.Sprintf("Unable to connect to MCP server %s: %v", serverURL, err))
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
		logging.Log.Errorf(&logging.ContextMap{}, "Error fetching prompt %s: %v", promptName, err)
		panic(fmt.Sprintf("Error fetching prompt %s: %v", promptName, err))
	}

	// Return full response which may contain messages array or other format
	return response
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
func ListAll(serverURL string, authToken string, transport string) map[string]interface{} {
	logging.Log.Debugf(&logging.ContextMap{}, "ListAll called with serverURL: %s, transport: %s", serverURL, transport)
	
	// Reuse existing functions

	// Fetch everything using existing functions
	// These will panic if there's an error
	tools := ListTools(serverURL, authToken, transport)
	resources := ListResources(serverURL, authToken, transport)
	prompts := ListPrompts(serverURL, authToken, transport)

	// Return everything in one map
	return map[string]interface{}{
		"tools":     tools,
		"resources": resources,
		"prompts":   prompts,
	}
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
func HealthCheck(serverURL string, authToken string, transport string) bool {
	logging.Log.Debugf(&logging.ContextMap{}, "HealthCheck called with serverURL: %s, transport: %s", serverURL, transport)
	
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
		return false
	}

	// If connection succeeded, server is available
	defer conn.Close()

	// Optionally: could send a ping or test request
	// Successful connection means server is healthy

	return true
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
func DiscoverServer(serverURL string) DiscoverServerResponse {
	logging.Log.Debugf(&logging.ContextMap{}, "DiscoverServer called with serverURL: %s", serverURL)
	
	// Auto-detect the most likely transport
	transport := detectTransport(serverURL)

	// Initialize result with defaults
	result := DiscoverServerResponse{
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
	available := HealthCheck(serverURL, "", transport)

	// Note: Since HealthCheck now returns bool only, we can't detect auth errors this way anymore
	// This is a limitation of removing error returns

	// If not available
	if !available {
		// STDIO - can't test without running
		if transport == "stdio" {
			result.Status = "possible_stdio"
			result.Note = "This appears to be a local executable. Use STDIO transport."
			return result
		}

		result.Status = "unavailable"
		result.Error = "Server not reachable"
		return result
	}

	// Get capabilities
	result.Status = "connected"

	// Use auto-detection for all capability checks
	// Wrap in defer/recover to catch panics from list functions
	var tools []interface{}
	var resources []interface{}
	var prompts []interface{}
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If listing fails, just leave empty
			}
		}()
		tools = ListTools(serverURL, "", "")
	}()
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If listing fails, just leave empty
			}
		}()
		resources = ListResources(serverURL, "", "")
	}()
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If listing fails, just leave empty
			}
		}()
		prompts = ListPrompts(serverURL, "", "")
	}()

	// Update capability information
	result.HasTools = len(tools) > 0
	result.ToolsCount = len(tools)
	result.HasResources = len(resources) > 0
	result.ResourcesCount = len(resources)
	result.HasPrompts = len(prompts) > 0
	result.PromptsCount = len(prompts)

	return result
}

// Functions connectToMCP and sendMCPRequest are defined in privatefunctions.go
// They handle connection and JSON-RPC communication with MCP server across all transports
