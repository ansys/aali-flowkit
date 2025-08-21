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
	"net/url"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

// ListTools retrieves available tools from the MCP server using the standard tools/list method.
//
// Tags:
//   - @displayName: List MCP Tools
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//
// Returns:
//   - tools: a list of available tools with their metadata
//   - error: any error that occurred during the process
func ListTools(serverURL string, authToken string, transport TransportType, timeout int) ([]interface{}, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", config.ServerURL, err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after list tools request")

	// Send standard tools/list request
	response, err := sendMCPRequest(ctx, conn, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP tools: %v", err)
	}

	// Extract tools from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if tools, exists := responseMap["tools"]; exists {
			if toolsList, ok := tools.([]interface{}); ok {
				return toolsList, nil
			}
			return nil, fmt.Errorf("tools field is not an array")
		}
		return []interface{}{}, nil // Return empty list if no tools
	}

	return nil, fmt.Errorf("unexpected response format")
}

// ListResources retrieves available resources from the MCP server using the standard resources/list method.
//
// Tags:
//   - @displayName: List MCP Resources
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//
// Returns:
//   - resources: a list of available resources with their metadata
//   - error: any error that occurred during the process
func ListResources(serverURL string, authToken string, transport TransportType, timeout int) ([]interface{}, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", config.ServerURL, err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after list resources request")

	// Send standard resources/list request
	response, err := sendMCPRequest(ctx, conn, "resources/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP resources: %v", err)
	}

	// Extract resources from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if resources, exists := responseMap["resources"]; exists {
			if resourcesList, ok := resources.([]interface{}); ok {
				return resourcesList, nil
			}
			return nil, fmt.Errorf("resources field is not an array")
		}
		return []interface{}{}, nil // Return empty list if no resources
	}

	return nil, fmt.Errorf("unexpected response format")
}

// ListPrompts retrieves available prompts from the MCP server using the standard prompts/list method.
//
// Tags:
//   - @displayName: List MCP Prompts
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//
// Returns:
//   - prompts: a list of available prompts with their metadata
//   - error: any error that occurred during the process
func ListPrompts(serverURL string, authToken string, transport TransportType, timeout int) ([]interface{}, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", config.ServerURL, err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after list prompts request")

	// Send standard prompts/list request
	response, err := sendMCPRequest(ctx, conn, "prompts/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP prompts: %v", err)
	}

	// Extract prompts from response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if prompts, exists := responseMap["prompts"]; exists {
			if promptsList, ok := prompts.([]interface{}); ok {
				return promptsList, nil
			}
			return nil, fmt.Errorf("prompts field is not an array")
		}
		return []interface{}{}, nil // Return empty list if no prompts
	}

	return nil, fmt.Errorf("unexpected response format")
}

// ListAll retrieves all tools, resources, and prompts from the MCP server.
// This is a convenience function that calls the individual list methods.
//
// Tags:
//   - @displayName: List All MCP
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//
// Returns:
//   - result: a map with lists of tools, resources, and prompts
//   - error: any error that occurred during the process
func ListAll(serverURL string, authToken string, transport TransportType, timeout int) (map[string][]interface{}, error) {
	result := make(map[string][]interface{})

	// List tools
	tools, err := ListTools(serverURL, authToken, transport, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	result["tools"] = tools

	// List resources
	resources, err := ListResources(serverURL, authToken, transport, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	result["resources"] = resources

	// List prompts
	prompts, err := ListPrompts(serverURL, authToken, transport, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}
	result["prompts"] = prompts

	return result, nil
}

// CallTool executes a specific tool via the MCP server using the standard tools/call method.
//
// Tags:
//   - @displayName: Call MCP Tool
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//   - toolName: the name of the tool to execute
//   - arguments: a map of arguments to pass to the tool
//
// Returns:
//   - result: the response from the tool execution
//   - error: any error that occurred during execution
func CallTool(serverURL string, authToken string, transport TransportType, timeout int, toolName string, arguments map[string]interface{}) (interface{}, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", config.ServerURL, err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after tool call")

	// Prepare parameters for tools/call
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	// Send standard tools/call request
	result, err := sendMCPRequest(ctx, conn, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("failed to call MCP tool '%s': %v", toolName, err)
	}
	return result, nil
}

// ReadResource retrieves a resource from the MCP server using the standard resources/read method.
//
// Tags:
//   - @displayName: Read MCP Resource
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//   - resourceURI: the URI of the resource to retrieve
//
// Returns:
//   - result: the retrieved resource contents
//   - error: any error that occurred during the request
func ReadResource(serverURL string, authToken string, transport TransportType, timeout int, resourceURI string) (interface{}, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", config.ServerURL, err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after resource read")

	// Prepare parameters for resources/read
	params := map[string]interface{}{
		"uri": resourceURI,
	}

	// Send standard resources/read request
	result, err := sendMCPRequest(ctx, conn, "resources/read", params)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP resource '%s': %v", resourceURI, err)
	}
	return result, nil
}

// GetPrompt retrieves a prompt from the MCP server using the standard prompts/get method.
//
// Tags:
//   - @displayName: Get MCP Prompt
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//   - promptName: the name of the prompt to retrieve
//   - arguments: optional arguments for the prompt (can be nil)
//
// Returns:
//   - prompt: the retrieved prompt content
//   - error: any error that occurred during the request
func GetPrompt(serverURL string, authToken string, transport TransportType, timeout int, promptName string, arguments map[string]interface{}) (interface{}, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", config.ServerURL, err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after prompt get")

	// Prepare parameters for prompts/get
	params := map[string]interface{}{
		"name": promptName,
	}

	// Add arguments if provided
	if arguments != nil {
		params["arguments"] = arguments
	}

	// Send standard prompts/get request
	result, err := sendMCPRequest(ctx, conn, "prompts/get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP prompt '%s': %v", promptName, err)
	}
	return result, nil
}

// InitializeMCP initializes a connection with an MCP server and performs the required handshake.
// This should be called before other MCP operations to ensure proper protocol compliance.
//
// Tags:
//   - @displayName: Initialize MCP
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//
// Returns:
//   - serverInfo: information about the server capabilities and version
//   - error: any error that occurred during initialization
func InitializeMCP(serverURL string, authToken string, transport TransportType, timeout int) (map[string]interface{}, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	// Validate URL based on transport type
	if config.Transport == "websocket" {
		if valid, errMsg := ValidateMCPServerURL(config.ServerURL); !valid {
			return nil, fmt.Errorf("invalid MCP server URL: %s", errMsg)
		}
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. Error: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after initialization")

	// Prepare initialization request per MCP specification
	initParams := map[string]interface{}{
		"protocolVersion": "0.1.0",
		"capabilities": map[string]interface{}{
			"tools":     map[string]interface{}{},
			"resources": map[string]interface{}{},
			"prompts":   map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "FlowKit",
			"version": "1.0.0",
		},
	}

	// Send initialize request
	response, err := sendMCPRequest(ctx, conn, "initialize", initParams)
	if err != nil {
		return nil, fmt.Errorf("MCP initialization failed: %v", err)
	}

	// Send initialized notification (fire and forget, no response expected)
	_ = sendMCPNotification(ctx, conn, "notifications/initialized", nil)

	// Return server capabilities
	if responseMap, ok := response.(map[string]interface{}); ok {
		return responseMap, nil
	}

	return nil, fmt.Errorf("unexpected initialization response format")
}

// ValidateMCPServerURL validates that a URL is appropriate for MCP connections.
// This function validates based on the transport type in the config.
//
// Tags:
//   - @displayName: Validate MCP Server URL
//
// Parameters:
//   - serverURL: the URL to validate
//
// Returns:
//   - valid: true if the URL is valid for MCP
//   - message: error message if invalid, empty if valid
func ValidateMCPServerURL(serverURL string) (bool, string) {
	if serverURL == "" {
		return false, "server URL cannot be empty"
	}

	// For now, we only support WebSocket. SSE and STDIO validation will be added later.
	// Check for WebSocket protocol
	if !strings.HasPrefix(serverURL, "ws://") && !strings.HasPrefix(serverURL, "wss://") {
		return false, "MCP server URL must start with ws:// (insecure) or wss:// (secure)"
	}

	// Parse URL to ensure it's valid
	u, err := url.Parse(serverURL)
	if err != nil {
		return false, fmt.Sprintf("invalid URL format: %v", err)
	}

	// Check that host is specified
	if u.Host == "" {
		return false, "server URL must include a host (e.g., ws://localhost:3000)"
	}

	return true, ""
}

// HealthCheck verifies that the MCP server is accessible and responsive.
// This implements Gautam's requested "health or status" functionality.
// It performs a quick connection test without executing any operations.
//
// Tags:
//   - @displayName: MCP Health Check
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 5 for health check)
//
// Returns:
//   - healthy: true if the server is accessible, false otherwise
//   - error: any error that occurred during the health check (with sanitized token)
func HealthCheck(serverURL string, authToken string, transport TransportType, timeout int) (bool, error) {
	// Create config from individual parameters
	config := MCPConfig{
		ServerURL: serverURL,
		AuthToken: authToken,
		Transport: string(transport),
		Timeout:   timeout,
	}

	// Set defaults if not provided
	if config.Transport == "" {
		config.Transport = string(TransportWebSocket)
	}
	// Use shorter timeout for health check
	if config.Timeout == 0 {
		config.Timeout = 5
	}

	// Create a short timeout context for health check
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// Attempt to connect to the MCP server
	conn, err := connectToMCP(ctx, config)
	if err != nil {
		// Sanitize error to prevent token leakage
		return false, sanitizeError(err, config.GetAuthToken())
	}

	// Close connection immediately after successful connect
	defer conn.Close(websocket.StatusNormalClosure, "health check complete")

	// If we got here, the server is healthy
	return true, nil
}

// Status provides detailed status information about the MCP server connection.
// This includes health status, transport type, and authentication status.
//
// Tags:
//   - @displayName: MCP Status
//
// Parameters:
//   - serverURL: The URL of the MCP server (required)
//   - authToken: Optional authentication token for the MCP server
//   - transport: Transport protocol: websocket (default), sse, or stdio
//   - timeout: Connection timeout in seconds (default: 30)
//
// Returns:
//   - status: a map containing health status, transport type, and authentication info
//   - error: any error that occurred during the status check (with sanitized token)
func Status(serverURL string, authToken string, transport TransportType, timeout int) (map[string]interface{}, error) {
	// Perform health check
	healthy, healthErr := HealthCheck(serverURL, authToken, transport, timeout)

	// Set default transport if not specified
	if transport == "" {
		transport = TransportWebSocket
	}

	// Build status response
	status := map[string]interface{}{
		"serverURL":     serverURL,
		"healthy":       healthy,
		"transport":     string(transport),
		"authenticated": authToken != "",
	}

	// Add error message if health check failed
	if healthErr != nil {
		status["error"] = healthErr.Error()
	}

	// Add timestamp
	status["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	return status, healthErr
}
