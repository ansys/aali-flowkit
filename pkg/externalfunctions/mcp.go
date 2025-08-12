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

	"nhooyr.io/websocket"
)

// ListTools retrieves available tools from the MCP server using the standard tools/list method.
//
// Tags:
//   - @displayName: List MCP Tools
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//
// Returns:
//   - tools: a list of available tools with their metadata
//   - error: any error that occurred during the process
func ListTools(serverURL string) ([]interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", serverURL, err)
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
//   - serverURL: the WebSocket URL of the MCP server
//
// Returns:
//   - resources: a list of available resources with their metadata
//   - error: any error that occurred during the process
func ListResources(serverURL string) ([]interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", serverURL, err)
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
//   - serverURL: the WebSocket URL of the MCP server
//
// Returns:
//   - prompts: a list of available prompts with their metadata
//   - error: any error that occurred during the process
func ListPrompts(serverURL string) ([]interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", serverURL, err)
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
//   - @displayName: List All MCP Items
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//
// Returns:
//   - result: a map with lists of tools, resources, and prompts
//   - error: any error that occurred during the process
func ListAll(serverURL string) (map[string][]interface{}, error) {
	result := make(map[string][]interface{})

	// List tools
	tools, err := ListTools(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	result["tools"] = tools

	// List resources
	resources, err := ListResources(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	result["resources"] = resources

	// List prompts
	prompts, err := ListPrompts(serverURL)
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
//   - serverURL: the WebSocket URL of the MCP server
//   - toolName: the name of the tool to execute
//   - arguments: a map of arguments to pass to the tool
//
// Returns:
//   - result: the response from the tool execution
//   - error: any error that occurred during execution
func CallTool(serverURL, toolName string, arguments map[string]interface{}) (interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", serverURL, err)
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

// ExecuteTool is a compatibility wrapper for CallTool.
// Deprecated: Use CallTool instead for better alignment with MCP standard.
//
// Tags:
//   - @displayName: Execute MCP Tool (Deprecated)
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - toolName: the name of the tool to execute
//   - args: a map of arguments to pass to the tool
//
// Returns:
//   - result: the response from the tool execution as a map
//   - error: any error that occurred during execution
func ExecuteTool(serverURL, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	result, err := CallTool(serverURL, toolName, args)
	if err != nil {
		return nil, err
	}

	// Convert result to map for backward compatibility
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	// Wrap non-map results
	return map[string]interface{}{"result": result}, nil
}

// ReadResource retrieves a resource from the MCP server using the standard resources/read method.
//
// Tags:
//   - @displayName: Read MCP Resource
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - resourceURI: the URI of the resource to retrieve
//
// Returns:
//   - result: the retrieved resource contents
//   - error: any error that occurred during the request
func ReadResource(serverURL, resourceURI string) (interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", serverURL, err)
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

// GetResource is a compatibility wrapper for ReadResource.
// Deprecated: Use ReadResource instead for better alignment with MCP standard.
//
// Tags:
//   - @displayName: Get MCP Resource (Deprecated)
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - resourceName: the name/URI of the resource to retrieve
//
// Returns:
//   - result: the retrieved resource as a map
//   - error: any error that occurred during the request
func GetResource(serverURL, resourceName string) (map[string]interface{}, error) {
	result, err := ReadResource(serverURL, resourceName)
	if err != nil {
		return nil, err
	}

	// Convert result to map for backward compatibility
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	// Wrap non-map results
	return map[string]interface{}{"contents": result}, nil
}

// GetPrompt retrieves a prompt from the MCP server using the standard prompts/get method.
//
// Tags:
//   - @displayName: Get MCP Prompt
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - promptName: the name of the prompt to retrieve
//   - arguments: optional arguments for the prompt (can be nil)
//
// Returns:
//   - prompt: the retrieved prompt content
//   - error: any error that occurred during the request
func GetPrompt(serverURL, promptName string, arguments map[string]interface{}) (interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("MCP server connection failed: please check the server URL is correct and the server is running. URL: %s, Error: %v", serverURL, err)
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

// GetSystemPrompt is a compatibility wrapper for GetPrompt.
// Deprecated: Use GetPrompt instead for better alignment with MCP standard.
//
// Tags:
//   - @displayName: Get System Prompt (Deprecated)
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - promptName: the name of the system prompt to retrieve
//
// Returns:
//   - promptStr: the text of the retrieved prompt
//   - error: any error that occurred during the request
func GetSystemPrompt(serverURL, promptName string) (string, error) {
	result, err := GetPrompt(serverURL, promptName, nil)
	if err != nil {
		return "", err
	}

	// Try to extract prompt text from various response formats
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Check for "messages" field (standard MCP format)
		if messages, exists := resultMap["messages"]; exists {
			if msgList, ok := messages.([]interface{}); ok && len(msgList) > 0 {
				if msg, ok := msgList[0].(map[string]interface{}); ok {
					if content, exists := msg["content"]; exists {
						if contentStr, ok := content.(string); ok {
							return contentStr, nil
						}
					}
				}
			}
		}

		// Fallback: check for direct "prompt" field
		if prompt, exists := resultMap["prompt"]; exists {
			if promptStr, ok := prompt.(string); ok {
				return promptStr, nil
			}
		}

		// Fallback: check for "content" field
		if content, exists := resultMap["content"]; exists {
			if contentStr, ok := content.(string); ok {
				return contentStr, nil
			}
		}
	}

	// If result is directly a string
	if promptStr, ok := result.(string); ok {
		return promptStr, nil
	}

	return "", fmt.Errorf("could not extract prompt text from MCP response: unexpected format")
}

// InitializeMCP initializes a connection with an MCP server and performs the required handshake.
// This should be called before other MCP operations to ensure proper protocol compliance.
//
// Tags:
//   - @displayName: Initialize MCP Connection
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server (must start with ws:// or wss://)
//
// Returns:
//   - serverInfo: information about the server capabilities and version
//   - error: any error that occurred during initialization
func InitializeMCP(serverURL string) (map[string]interface{}, error) {
	// Validate URL first
	if valid, errMsg := ValidateMCPServerURL(serverURL); !valid {
		return nil, fmt.Errorf("invalid MCP server URL: %s", errMsg)
	}

	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
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
