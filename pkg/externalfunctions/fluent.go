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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FluentCodeGen sends a raw user message to the Fluent container and returns the response
//
// Tags:
//   - @displayName: Fluent Code Gen
//
// Parameters:
//   - message: the raw user message to send to the container
//
// Returns:
//   - response: the response from the Fluent container as a string
func FluentCodeGen(message string) (response string) {
	url := "http://localhost:9013/chat"
	
	// Create the JSON payload directly
	jsonData := fmt.Sprintf(`{"message": "%s"}`, message)
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(jsonData))
	if err != nil {
		panic(fmt.Sprintf("Error creating HTTP request: %v", err))
	}
	
	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Error executing HTTP request: %v", err))
	}
	defer resp.Body.Close()
	
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error reading response body: %v", err))
	}
	
	// Check if the response code is successful (2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		panic(fmt.Sprintf("HTTP request failed with status code %d: %s", resp.StatusCode, string(body)))
	}
	
	// Parse JSON response to extract just the response content
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		panic(fmt.Sprintf("Error parsing JSON response: %v", err))
	}
	
	// Extract the response field
	if responseField, exists := responseData["response"]; exists {
		if responseArray, ok := responseField.([]interface{}); ok && len(responseArray) > 0 {
			// Return the first item in the response array as string
			return fmt.Sprintf("%v", responseArray[0])
		}
	}
	
	// Fallback to raw response if parsing fails
	return string(body)
}