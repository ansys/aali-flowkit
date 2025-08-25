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
	"net/url"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
	"k8s.io/client-go/util/jsonpath"
)

// SendAPICall sends an API call to the specified URL with the specified headers and query parameters.
//
// Tags:
//   - @displayName: REST Call
//
// Parameters:
//   - requestType: the type of the request (GET, POST, PUT, PATCH, DELETE)
//   - urlString: the URL to send the request to
//   - headers: the headers to include in the request
//   - query: the query parameters to include in the request
//   - jsonBody: the body of the request as a JSON string
//
// Returns:
//   - success: a boolean indicating whether the request was successful
//   - returnJsonBody: the JSON body of the response as a string
func SendRestAPICall(requestType string, endpoint string, header map[string]string, query map[string]string, jsonBody string) (success bool, returnJsonBody string) {
	// verify correct request type
	if requestType != "GET" && requestType != "POST" && requestType != "PUT" && requestType != "PATCH" && requestType != "DELETE" {
		panic(fmt.Sprintf("Invalid request type: %v", requestType))
	}

	// Parse the URL and add query parameters
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		panic(fmt.Sprintf("Error parsing URL: %v", err))
	}

	q := parsedURL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	parsedURL.RawQuery = q.Encode()

	// Create the HTTP request
	var req *http.Request
	if jsonBody != "" {
		req, err = http.NewRequest(requestType, parsedURL.String(), bytes.NewBuffer([]byte(jsonBody)))
	} else {
		req, err = http.NewRequest(requestType, parsedURL.String(), nil)
	}
	if err != nil {
		panic(fmt.Sprintf("Error creating request: %v", err))
	}

	// Add headers
	for key, value := range header {
		req.Header.Add(key, value)
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Error executing request: %v", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error reading response body: %v", err))
	}

	// Check if the response code is successful (2xx)
	success = resp.StatusCode >= 200 && resp.StatusCode < 300

	return success, string(body)
}

// AssignStringToString assigns a string to another string
//
// Tags:
//   - @displayName: Assign String to String
//
// Parameters:
//   - inputString: the input string
//
// Returns:
//   - outputString: the output string
func AssignStringToString(inputString string) (outputString string) {
	return inputString
}

// PrintFeedback prints the feedback to the console in JSON format
//
// Tags:
//   - @displayName: Print Feedback
//
// Parameters:
//   - feedback: the feedback to print
func PrintFeedback(feedback sharedtypes.Feedback) {
	// create json string from feedback struct
	jsonString, err := json.Marshal(feedback)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling feedback to JSON: %v", err))
	}
	// print json string to console
	fmt.Println(string(jsonString))
}

// ExtractJSONStringField extracts a string field from a JSON string using a key path.
// The key path is a dot-separated string that specifies the path to the field in the JSON object.
//
// Tags:
//   - @displayName: Extract JSON String Field
//
// Parameters:
//   - jsonStr: the JSON string to extract the field from
//   - keyPath: the dot-separated path to the field in the JSON object
//
// Returns:
//   - the value of the field as a string
func ExtractJSONStringField(jsonStr string, keyPath string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		panic(fmt.Sprintf("Error unmarshalling JSON: %v", err))
	}

	keys := strings.Split(keyPath, ".")
	var current interface{} = data

	for _, key := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			panic(fmt.Sprintf("Expected map for key %q but got %T", key, current))
		}
		current, ok = m[key]
		if !ok {
			panic(fmt.Sprintf("Key %q not found in JSON", key))
		}
	}

	// Convert final value to string
	switch v := current.(type) {
	case string:
		return v
	default:
		// Try to marshal the value back to a JSON string
		bytes, err := json.Marshal(v)
		if err != nil {
			panic(fmt.Sprintf("Unable to convert final value to string: %v", err))
		}
		return string(bytes)
	}
}

// GenerateUUID generates a new UUID (Universally Unique Identifier).
//
// Tags:
//   - @displayName: Generate UUID
//
// Returns:
//   - a string representation of the generated UUID
func GenerateUUID() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}

// JsonPath extracts some data from an arbitrary data structure using a JSONPath pattern
//
// Tags:
//   - @displayName: JSON Path
//
// Parameters:
//   - pat (string): The JSON Path pattern
//   - data (any): The data to extract from
//   - oneResult (bool): Whether you are expecting to extract 1 result or an array of results
//     If you set oneResult=true but there are not exactle 1 result in the output, you will
//     receive an error. This should only be set if the result is guaranteed to have length 1.
//
// Returns
//   - The extracted data. If oneResult=false, this will be an array of any.
func JsonPath(pat string, data any, oneResult bool) any {
	jpath := jsonpath.New("")
	jpath.EnableJSONOutput(true)
	pat = fmt.Sprintf("{ %v }", pat)
	err := jpath.Parse(pat)
	if err != nil {
		logPanic(nil, "could not parse the provided JSONPath %q: %v", pat, err)
	}
	res, err := jpath.FindResults(data)
	if err != nil {
		logPanic(nil, "could not find JSONPath results with pattern %q in data %#v: %v", pat, data, err)
	}

	if len(res) != 1 {
		// this should be unreachable since it is hardcoded above only 1 root node (surrounded by {}) in the pattern
		logPanic(nil, "there should only ever be 1 root node but found %d", len(res))
	}

	reflectVals := res[0]
	if oneResult {
		if len(reflectVals) != 1 {
			logPanic(nil, "specified 1 result but found %d", len(reflectVals))
		}
		return reflectVals[0].Interface()
	} else {
		anyVals := make([]any, len(reflectVals))
		for i, reflectVal := range reflectVals {
			anyVals[i] = reflectVal.Interface()
		}
		return anyVals
	}
}

// StringConcat concatenates 2 strings together, with an optional separator.
//
// Tags:
//   - @displayName: Concatenate Strings
//
// Parameters
//   - a (string) the first string
//   - b (string) the second string
//   - separator (string) the separator string. If not provided, will be an empty string.
func StringConcat(a string, b string, separator string) string {
	return fmt.Sprintf("%v%v%v", a, separator, b)
}

// StringFormat formats any data as a string.
//
// Use this to turn non-string data into a string representation. This uses go's `fmt.Sprintf` under the hood.
//
// Tags:
//   - @displayName: Format data as string
//
// Parameters
//   - data (any): the data to format as a string
//   - format (string): the format specifier to use. If not provided will default to "%v".
//     See the [go fmt docs](https://pkg.go.dev/fmt) for details.
func StringFormat(data any, format string) string {
	if format == "" {
		format = "%v"
	}
	return fmt.Sprintf(format, data)
}

// FluentCodeGen sends a raw user message to the Fluent container and returns the response.
// It takes a user message and posts it to the Fluent API endpoint to generate code.
//
// Tags:
//   - @displayName: Fluent Code Gen Test
//
// Parameters:
//   - message: the raw user message to send to the container
//
// Returns:
//   - response: the response from the Fluent container as a string
func FluentCodeGenTest(message string) (response string) {
	url := "http://aali-fluent:8000/chat"
	
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
